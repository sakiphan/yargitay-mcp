// SPDX-License-Identifier: AGPL-3.0-only
package yargitay

import (
	"context"
	"net/http"
	"net/url"
	"strings"
	"testing"
	"time"
)

func TestOfficialDocumentMediaAndErrorKinds(t *testing.T) {
	for _, tc := range []struct {
		origin, op, media, body string
		status                  int
		want                    error
	}{
		{DanistayOrigin, "document", "text/html;charset=UTF-8", `<script src="captcha.js"></script><p id="hiddencontent">encoded</p>`, 200, nil},
		{DanistayOrigin, "document", "application/json", `{"data":"example"}`, 200, nil},
		{DanistayOrigin, "search", "text/html", `<script src="captcha.js"></script>`, 200, ErrSchema},
		{AYMOrigin, "document", "text/html", `<script src="captcha.js"></script>`, 200, ErrSchema},
		{DanistayOrigin, "document", "text/plain", "captcha", 200, ErrSchema},
		{DanistayOrigin, "document", "text/html", "", 302, ErrRedirect},
		{DanistayOrigin, "document", "text/html", "", 403, ErrAccess},
		{DanistayOrigin, "document", "application/json", "", 404, ErrNotFound},
	} {
		t.Run(tc.origin+tc.op+tc.media+http.StatusText(tc.status), func(t *testing.T) {
			c := testClient(t, func(*http.Request) (*http.Response, error) {
				r := response(tc.status, tc.body)
				r.Header.Set("Content-Type", tc.media)
				return r, nil
			})
			c.origin = tc.origin
			checked := false
			_, _, e := c.OfficialRequest(context.Background(), tc.op, nil, nil, func([]byte) error { checked = true; return nil })
			if e != tc.want {
				t.Fatalf("got %v want %v", e, tc.want)
			}
			if checked != (tc.want == nil) {
				t.Fatal("unexpected parser dispatch")
			}
		})
	}
}

func TestOfficialBoundariesAndCache(t *testing.T) {
	for _, source := range []string{"aym", "danistay", "aihm"} {
		t.Run(source, func(t *testing.T) {
			cfg := DefaultConfig()
			c, e := NewOfficialClient(source, cfg)
			if e != nil {
				t.Fatal(e)
			}
			defer c.Close()
			if c.limiter != sharedLimiter(cfg) {
				t.Fatal("not process-wide")
			}
			tr := c.http.Transport.(*http.Transport)
			if tr.Proxy != nil || tr.TLSClientConfig != nil && tr.TLSClientConfig.InsecureSkipVerify {
				t.Fatal("unsafe transport")
			}
			if c.http.CheckRedirect(nil, nil) != http.ErrUseLastResponse {
				t.Fatal("redirects enabled")
			}
			c.limiter = &limiter{maxQueue: 20}
			calls := 0
			c.http.Transport = roundTrip(func(r *http.Request) (*http.Response, error) {
				calls++
				if r.URL.Scheme != "https" || r.URL.Host != strings.TrimPrefix(c.origin, "https://") {
					t.Fatal("wrong origin")
				}
				return response(200, `{"data":[]}`), nil
			})
			for i := 0; i < 2; i++ {
				if _, _, e := c.OfficialRequest(context.Background(), "search", url.Values{}, []byte(`{}`)); e != nil {
					t.Fatal(e)
				}
			}
			if calls != 1 {
				t.Fatal("cache miss", calls)
			}
			if _, _, e := c.OfficialRequest(context.Background(), "https://evil.example", nil, nil); e != ErrInvalid {
				t.Fatal(e)
			}
			if _, e := c.request(context.Background(), "GET", "/getDokuman", nil, nil); e != ErrInvalid {
				t.Fatal("Yargıtay route exposed on foreign origin")
			}
		})
	}
	if _, e := NewOfficialClient("https://evil.example", DefaultConfig()); e != ErrInvalid {
		t.Fatal(e)
	}
}
func TestOfficialFailureNotCachedOrRetried(t *testing.T) {
	for _, status := range []int{200, 301, 403, 429, 503} {
		t.Run(http.StatusText(status), func(t *testing.T) {
			calls := 0
			c := testClient(t, func(*http.Request) (*http.Response, error) {
				calls++
				r := response(status, `{}`)
				r.Header.Set("Retry-After", "0")
				return r, nil
			})
			c.origin = AYMOrigin
			ctx, cancel := context.WithTimeout(context.Background(), time.Second)
			defer cancel()
			for i := 0; i < 2; i++ {
				if _, _, e := c.OfficialRequest(ctx, "search", nil, []byte(`{}`), func([]byte) error { return ErrSchema }); e == nil {
					t.Fatal("error accepted")
				}
			}
			if calls != 2 {
				t.Fatal("automatic retry or cached failure", calls)
			}
		})
	}
}
