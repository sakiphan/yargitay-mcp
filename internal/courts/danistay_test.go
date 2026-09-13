// SPDX-License-Identifier: AGPL-3.0-only
package courts

import (
	"context"
	"errors"
	"github.com/sakiphan/yargitay-mcp/internal/yargitay"
	"html"
	"strings"
	"testing"
	"time"
)

func TestDanistayEncodedViewer(t *testing.T) {
	decision := "<p>Sentetik karar: " + strings.Repeat("vergi 😀 & delil. ", 12) + "</p><script>EXECUTE_BAD()</script><p>İkinci paragraf.</p>"
	raw := `<!DOCTYPE html><html><head><script src="/captcha.js"></script></head><body><nav>NOT_A_DECISION</nav><div id="content"></div><p id="hiddencontent" style="display:none">` + html.EscapeString(decision) + `</p><p id="hiddenArananKelime">QUERY_ONLY</p></body></html>`
	f := &fakeFetch{raw: raw}
	c := Client{source: "danistay", fetch: f, maxPageSize: 20}
	d, e := c.GetDecision(context.Background(), "123")
	if e != nil {
		t.Fatal(e)
	}
	if f.calls != 1 || f.params.Get("arananKelime") != "" || !f.params.Has("arananKelime") {
		t.Fatal("changed document request", f.calls, f.params)
	}
	if !strings.Contains(d.Text, "vergi 😀 & delil.") || !strings.Contains(d.Text, "\n\nİkinci paragraf.") {
		t.Fatal("lost paragraphs/entities", d.Text)
	}
	for _, bad := range []string{"NOT_A_DECISION", "QUERY_ONLY", "EXECUTE_BAD", "<p>", "captcha.js"} {
		if strings.Contains(d.Text, bad) {
			t.Fatal("viewer noise or executable content leaked", bad)
		}
	}
	if d.ID != "123" || d.SourceURL != SourceURL("danistay", "123") {
		t.Fatal("lost identity")
	}
}

func TestDanistayRejectsNonDecisions(t *testing.T) {
	for _, tc := range []struct {
		raw  string
		want error
	}{
		{`<html><script src="captcha.js"></script><p>Error page, not a decision.</p></html>`, yargitay.ErrSchema},
		{`<html><form><div class="g-recaptcha"></div></form></html>`, yargitay.ErrAccess},
		{`<html><div class="h-captcha"></div></html>`, yargitay.ErrAccess},
		{`<p id="hiddencontent"></p>`, yargitay.ErrSchema},
		{`<p id="hiddencontent">one</p><p id="hiddencontent">two</p>`, yargitay.ErrSchema},
		{`<p id="hiddencontent"><b>unexpected encoding</b></p>`, yargitay.ErrSchema},
		{`<template><p id="hiddencontent">not displayed</p></template>`, yargitay.ErrSchema},
		{`{"metadata":{"FMTY":"ERROR","FMC":"ADALET_RUNTIME_EXCEPTION"},"data":null}`, yargitay.ErrUpstream},
	} {
		if _, e := parseDocument("danistay", "123", []byte(tc.raw), time.Now()); !errors.Is(e, tc.want) {
			t.Fatalf("got %v want %v", e, tc.want)
		}
	}
}

func TestDanistaySearchNoStaleAccessWarning(t *testing.T) {
	r := SearchRequest{Query: "vergi"}
	_ = r.normalize("danistay")
	out, e := parseSearch("danistay", []byte(danistayFixture), r, time.Now())
	if e != nil {
		t.Fatal(e)
	}
	if out.RelevanceChecked || len(out.Warnings) != 0 {
		t.Fatal("stale warning or relevance claim", out)
	}
}
