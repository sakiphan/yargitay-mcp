// SPDX-License-Identifier: AGPL-3.0-only
package yargitay

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestMain(m *testing.M) {
	// Every unmocked HTTP dial in this package fails without accessing the network.
	tr := http.DefaultTransport.(*http.Transport).Clone()
	tr.DialContext = func(context.Context, string, string) (net.Conn, error) {
		return nil, errors.New("offline test: network forbidden")
	}
	http.DefaultTransport = tr
	os.Exit(m.Run())
}

type roundTrip func(*http.Request) (*http.Response, error)

func (f roundTrip) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }
func response(status int, body string) *http.Response {
	return &http.Response{StatusCode: status, Header: http.Header{"Content-Type": {"application/json"}}, Body: io.NopCloser(strings.NewReader(body))}
}
func testClient(t *testing.T, f roundTrip) *Client {
	t.Helper()
	c, e := NewClient(DefaultConfig())
	if e != nil {
		t.Fatal(e)
	}
	c.http.Transport = f
	c.limiter = &limiter{interval: 0, maxQueue: 20}
	t.Cleanup(c.Close)
	return c
}

const searchFixture = `{"data":{"recordsTotal":1,"recordsFiltered":1,"data":[{"id":"123456","daire":"9. Hukuk Dairesi","esasNo":"2024/123","kararNo":"2025/456","kararTarihi":"21.04.2025"}]}}`
const documentFixture = `{"data":"<p>Sentetik karar: çalışma.</p><script>bad()</script><p>İkinci paragraf.</p>"}`

func TestSearchValidation(t *testing.T) {
	cases := []string{`{}`, `{"query":" "}`, `{"query":"ab"}`, `{"query":"abc","page":0}`, `{"query":"abc","page_size":21}`, `{"query":"abc","case_year":1899}`, `{"query":"abc","decision_year":2101}`, `{"query":"abc","case_number_from":0}`, `{"query":"abc","case_number_from":10,"case_number_to":2}`, `{"query":"abc","decision_number_from":10,"decision_number_to":2}`, `{"query":"abc","decision_date_from":"2025-02-30"}`, `{"query":"abc","decision_date_from":"2025-02-01","decision_date_to":"2025-01-01"}`, `{"query":"abc","unit":"ALL"}`, `{"query":"abc","sort_by":"unknown"}`, `{"query":"abc","sort_direction":"unknown"}`}
	for i, raw := range cases {
		t.Run(fmt.Sprint(i), func(t *testing.T) {
			var r SearchRequest
			if json.Unmarshal([]byte(raw), &r) != nil {
				t.Fatal("fixture")
			}
			if r.Normalize() == nil {
				t.Fatal("accepted invalid request")
			}
		})
	}
	r := SearchRequest{Query: ptr("  çalışma  "), Unit: ptr("9. Hukuk Dairesi"), DateFrom: ptr("2025-01-01")}
	if e := r.Normalize(); e != nil {
		t.Fatal(e)
	}
	d := searchPayload(r)["data"].(map[string]any)
	if d["pageSize"] != 10 || d["baslangicTarihi"] != "01.01.2025" || d["birimYrgHukukDaire"] != "9. Hukuk Dairesi" || d["arananKelime"] != `"çalışma"` {
		t.Fatal(d)
	}
}
func TestParserAndUnicodeChunks(t *testing.T) {
	text, md, e := CleanHTML(`<p>  Çağrı&nbsp;  için 😀. </p><script>bad()</script><style>bad</style><iframe>bad</iframe><p><a href="javascript:bad()">[link](https://invalid)</a><br>Son satır</p>`)
	if e != nil || strings.Contains(text, "bad") || !strings.Contains(text, "\n\n") || !strings.Contains(md, `\[link\]`) {
		t.Fatalf("bad normalization: %q %q %v", text, md, e)
	}
	d := Document{ID: "1", Text: text, Markdown: md}
	var joined strings.Builder
	offset := 0
	var hash *string
	for {
		c, e := ChunkDocument(d, DocumentRequest{DocumentID: "1", Offset: &offset, MaxChars: ptr(3), OutputFormat: ptr("text"), ExpectedHash: hash})
		if e != nil {
			t.Fatal(e)
		}
		joined.WriteString(c.Content)
		hash = &c.ContentSHA256
		if c.NextOffset == nil {
			break
		}
		offset = *c.NextOffset
	}
	if joined.String() != text {
		t.Fatal("Unicode chunks changed content")
	}
	end := len([]rune(text))
	c, e := ChunkDocument(d, DocumentRequest{DocumentID: "1", Offset: &end, OutputFormat: ptr("text")})
	if e != nil || c.Content != "" || c.NextOffset != nil {
		t.Fatal(c, e)
	}
	if _, e = ChunkDocument(d, DocumentRequest{DocumentID: "1", Offset: ptr(end + 1), OutputFormat: ptr("text")}); e == nil {
		t.Fatal("past-end accepted")
	}
	if _, e = ChunkDocument(d, DocumentRequest{DocumentID: "1", ExpectedHash: ptr(strings.Repeat("0", 64))}); !errors.Is(e, ErrChanged) {
		t.Fatal(e)
	}
	if _, _, e = CleanHTML(`<script>only</script>`); !errors.Is(e, ErrSchema) {
		t.Fatal(e)
	}
}
func TestDocumentValidation(t *testing.T) {
	for _, r := range []DocumentRequest{{DocumentID: "../1"}, {DocumentID: "abc"}, {DocumentID: "1", Offset: ptr(-1)}, {DocumentID: "1", MaxChars: ptr(0)}, {DocumentID: "1", MaxChars: ptr(100001)}, {DocumentID: "1", OutputFormat: ptr("html")}, {DocumentID: "1", ExpectedHash: ptr("bad")}} {
		if r.Normalize() == nil {
			t.Fatal("accepted", r)
		}
	}
}
func TestSchemaChanges(t *testing.T) {
	r := SearchRequest{Query: ptr("abc")}
	_ = r.Normalize()
	for i, raw := range []string{`null`, `[]`, `{}`, `{"data":null}`, `{"data":{"data":null}}`, `{"data":{"data":[],"recordsTotal":true}}`, `{"data":{"data":[],"recordsTotal":"1"}}`, `{"data":{"data":[null]}}`, `{"data":{"data":[{"id":"../1"}]}}`, `{"data":{"data":[{"id":1,"daire":3}]}}`, `{"data":{"data":[{"id":1,"kararTarihi":"bad"}]}}`} {
		t.Run(fmt.Sprint(i), func(t *testing.T) {
			if _, e := parseSearch([]byte(raw), r); e == nil {
				t.Fatal("schema accepted")
			}
		})
	}
	out, e := parseSearch([]byte(`{"new":true,"data":{"data":[{"id":1,"new":"future"}]}}`), r)
	if e != nil || out.Total != nil || out.Results[0].Unit != nil {
		t.Fatal(out, e)
	}
}
func TestClientPayloadCacheAndCoalescing(t *testing.T) {
	var calls atomic.Int32
	c := testClient(t, func(r *http.Request) (*http.Response, error) {
		calls.Add(1)
		if r.URL.Scheme != "https" || r.URL.Host != "karararama.yargitay.gov.tr" {
			t.Error("unsafe origin")
		}
		if r.URL.Path == "/aramadetaylist" {
			var body map[string]map[string]any
			_ = json.NewDecoder(r.Body).Decode(&body)
			if body["data"]["pageSize"] != float64(10) {
				t.Error(body)
			}
			return response(200, searchFixture), nil
		}
		return response(200, documentFixture), nil
	})
	var wg sync.WaitGroup
	for range 8 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if _, e := c.Search(context.Background(), SearchRequest{Query: ptr("PRIVATE query")}); e != nil {
				t.Error(e)
			}
		}()
	}
	wg.Wait()
	if calls.Load() != 1 {
		t.Fatal(calls.Load())
	}
	a, e := c.GetDecision(context.Background(), "123456")
	if e != nil {
		t.Fatal(e)
	}
	b, e := c.GetDecision(context.Background(), "123456")
	if e != nil || a.FetchedAt != b.FetchedAt || calls.Load() != 2 || strings.Contains(a.Text, "bad()") {
		t.Fatal(a, b, e, calls.Load())
	}
	if c.pending != 0 || len(c.flights) != 0 {
		t.Fatal("leaked flights")
	}
}
func TestHTTPFailures(t *testing.T) {
	for _, status := range []int{400, 401, 403, 404, 302, 429, 500, 502, 503, 504} {
		t.Run(fmt.Sprint(status), func(t *testing.T) {
			var calls int
			c := testClient(t, func(*http.Request) (*http.Response, error) {
				calls++
				r := response(status, `{}`)
				r.Header.Set("Retry-After", "0")
				return r, nil
			})
			c.config.MaxRetries = 2
			_, e := c.GetDecision(context.Background(), "1")
			if e == nil {
				t.Fatal("expected error")
			}
			want := 1
			if status == 429 || status >= 500 {
				want = 3
			}
			if calls != want {
				t.Fatal(calls, want)
			}
		})
	}
}
func TestUpstreamAccessAndMalformedBodies(t *testing.T) {
	for i, body := range []string{`{"metadata":{"FMTY":"ERROR","secret":"PRIVATE"},"data":null}`, `{"status":"EXCEPTION PRIVATE"}`, `{"data":{}}`, `{"data":null}`, `not JSON`} {
		t.Run(fmt.Sprint(i), func(t *testing.T) {
			c := testClient(t, func(*http.Request) (*http.Response, error) { return response(200, body), nil })
			_, e := c.GetDecision(context.Background(), "1")
			if e == nil || strings.Contains(e.Error(), "PRIVATE") {
				t.Fatal(e)
			}
		})
	}
	c := testClient(t, func(*http.Request) (*http.Response, error) {
		r := response(200, `captcha`)
		r.Header.Set("Content-Type", "text/html")
		return r, nil
	})
	if _, e := c.GetDecision(context.Background(), "1"); e != ErrAccess {
		t.Fatal(e)
	}
}
func TestBoundsTimeoutAndCapacity(t *testing.T) {
	c := testClient(t, func(*http.Request) (*http.Response, error) { return response(200, strings.Repeat("x", 1025)), nil })
	c.config.MaxResponseBytes = 1024
	if _, e := c.GetDecision(context.Background(), "1"); e != ErrSchema {
		t.Fatal(e)
	}
	var calls int
	c.http.Transport = roundTrip(func(*http.Request) (*http.Response, error) { calls++; return nil, errors.New("PRIVATE timeout") })
	if _, e := c.GetDecision(context.Background(), "1"); e != ErrUnavailable || calls != 1 {
		t.Fatal(e, calls)
	}
	c.config.Deadline = 10 * time.Millisecond
	c.limiter.deferFor(time.Hour)
	if _, e := c.GetDecision(context.Background(), "2"); e != ErrUnavailable {
		t.Fatal(e)
	}
	if len(c.limiter.queue) != 0 || c.pending != 0 {
		t.Fatal("cancel leaked queue")
	}
	c.config.MaxQueue = 1
	c.pending = 1
	if _, e := c.GetDecision(context.Background(), "2"); e != ErrCapacity {
		t.Fatal(e)
	}
	c.pending = 0
}
func TestCacheEvictionExpiryAndCopies(t *testing.T) {
	c := newCache(2, 6)
	c.set("a", []byte("abc"), time.Hour)
	c.set("b", []byte("def"), time.Hour)
	v := c.get("a")
	v[0] = 'x'
	if string(c.get("a")) != "abc" {
		t.Fatal("mutated cache")
	}
	c.set("c", []byte("ghi"), time.Hour)
	if c.get("b") != nil {
		t.Fatal("LRU not evicted")
	}
	c.set("huge", []byte("1234567"), time.Hour)
	if c.get("huge") != nil {
		t.Fatal("byte limit")
	}
	c.set("c", []byte("x"), time.Nanosecond)
	time.Sleep(time.Millisecond)
	if c.get("c") != nil {
		t.Fatal("expired")
	}
}
func TestLimiterSpacingSerializationAndCancellation(t *testing.T) {
	l := &limiter{interval: 15 * time.Millisecond, maxQueue: 2}
	start := time.Now()
	release, e := l.acquire(context.Background())
	if e != nil {
		t.Fatal(e)
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, e := l.acquire(ctx); e != ErrUnavailable {
		t.Fatal(e)
	}
	release()
	l.deferFor(20 * time.Millisecond)
	release, e = l.acquire(context.Background())
	if e != nil {
		t.Fatal(e)
	}
	release()
	if time.Since(start) < 20*time.Millisecond {
		t.Fatal("spacing")
	}
	if len(l.queue) != 0 {
		t.Fatal("queue leaked")
	}
}
func TestLimiterQueueFull(t *testing.T) {
	l := &limiter{maxQueue: 1}
	release, _ := l.acquire(context.Background())
	defer release()
	if _, e := l.acquire(context.Background()); e != ErrCapacity {
		t.Fatal(e)
	}
}
func TestConfigCatalogAndTLS(t *testing.T) {
	u, e := Units("all")
	if e != nil || len(u.Units) != 51 {
		t.Fatal(e, len(u.Units))
	}
	for _, u := range u.Units {
		if !u.Observed || u.Active != nil {
			t.Fatal("invented metadata")
		}
	}
	c, e := NewClient(DefaultConfig())
	if e != nil {
		t.Fatal(e)
	}
	defer c.Close()
	tr := c.http.Transport.(*http.Transport)
	if tr.Proxy != nil || (tr.TLSClientConfig != nil && tr.TLSClientConfig.InsecureSkipVerify) {
		t.Fatal("unsafe transport")
	}
	if c.http.CheckRedirect(nil, nil) != http.ErrUseLastResponse {
		t.Fatal("redirects enabled")
	}
	d, e := NewClient(DefaultConfig())
	if e != nil {
		t.Fatal(e)
	}
	defer d.Close()
	if c.limiter != d.limiter {
		t.Fatal("not process shared")
	}
	for _, pair := range [][2]string{{"YARGITAY_MIN_INTERVAL_SECONDS", "1"}, {"YARGITAY_TIMEOUT_SECONDS", "NaN"}, {"YARGITAY_MAX_PAGE_SIZE", "21"}, {"YARGITAY_BASE_URL", "https://example.com"}, {"YARGITAY_VIEWER_ENABLED", "garbage"}} {
		t.Run(pair[0], func(t *testing.T) {
			t.Setenv(pair[0], pair[1])
			if _, e := ConfigFromEnv(); e == nil {
				t.Fatal("unsafe config accepted")
			}
		})
	}
}
func TestRetryAfter(t *testing.T) {
	now := time.Now()
	for _, v := range []string{"10", now.Add(time.Minute).UTC().Format(http.TimeFormat)} {
		d, ok := retryAfter(v, now)
		if !ok || d <= 0 {
			t.Fatal(v, d)
		}
	}
	if _, ok := retryAfter("bad", now); ok {
		t.Fatal("bad retry header")
	}
}
