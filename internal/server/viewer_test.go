// SPDX-License-Identifier: AGPL-3.0-only
package server

import (
	"io"
	"net/http"
	"strings"
	"testing"
	"time"
)

func TestReaderSecurityAndEscaping(t *testing.T) {
	src := &fakeSource{}
	reader, e := StartReader(src, 20)
	if e != nil {
		t.Fatal(e)
	}
	defer reader.Close()
	u := reader.URL("123")
	client := &http.Client{Timeout: 5 * time.Second}
	resp, e := client.Get(*u)
	if e != nil {
		t.Fatal(e)
	}
	body, _ := io.ReadAll(resp.Body)
	resp.Body.Close()
	if resp.StatusCode != 200 || strings.Contains(string(body), "<script>alert") || !strings.Contains(string(body), "&lt;script&gt;") || resp.Header.Get("Cache-Control") != "no-store" {
		t.Fatal(string(body), resp.StatusCode)
	}
	r, _ := http.NewRequest("GET", *u, nil)
	r.Header.Set("Origin", "https://example.com")
	resp, e = client.Do(r)
	if e != nil {
		t.Fatal(e)
	}
	resp.Body.Close()
	if resp.StatusCode != 403 {
		t.Fatal(resp.StatusCode)
	}
	r, _ = http.NewRequest("GET", *u, nil)
	r.Host = "evil.example"
	resp, e = client.Do(r)
	if e != nil {
		t.Fatal(e)
	}
	resp.Body.Close()
	if resp.StatusCode != 400 {
		t.Fatal(resp.StatusCode)
	}
	for _, id := range []string{"404", "bad"} {
		resp, e = client.Get(*reader.URL(id))
		if e != nil {
			t.Fatal(e)
		}
		resp.Body.Close()
		if resp.StatusCode != 404 {
			t.Fatal(resp.StatusCode)
		}
	}
}
