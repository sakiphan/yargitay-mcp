// SPDX-License-Identifier: AGPL-3.0-only
package courts

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"github.com/sakiphan/yargitay-mcp/internal/yargitay"
	"net/url"
	"strings"
	"testing"
	"time"
)

const uuid = "11111111-2222-3333-4444-555555555555"
const aymFixture = `{"total":1,"data":[{"id":"11111111-2222-3333-4444-555555555555","kararTipi":"BireyselBasvuru","basvuruAdi":"Sentetik başvuru","basvuruNo":"2020/123","kararTarihi":"2024-01-02"}]}`
const danistayFixture = `{"metadata":{"FMTY":"SUCCESS"},"data":{"recordsFiltered":1,"data":[{"id":"123","daireKurul":"Sentetik Daire","esasNo":"2020/1","kararNo":"2021/2","kararTarihi":"02.01.2024"}]}}`
const hudocFixture = `{"resultcount":1,"results":[{"columns":{"itemid":"001-123","docname":"SYNTHETIC CASE","appno":"123/20","kpdate":"2024-01-02T00:00:00","languageisocode":"ENG","doctype":"HEJUD","documentcollectionid2":"CASELAW;JUDGMENTS;CHAMBER;ENG"}}]}`

type fakeFetch struct {
	calls     int
	raw       string
	operation string
	params    url.Values
	body      []byte
}

func (f *fakeFetch) OfficialRequest(_ context.Context, op string, p url.Values, b []byte, checks ...func([]byte) error) ([]byte, time.Time, error) {
	f.calls++
	f.operation = op
	f.params = p
	f.body = b
	for _, check := range checks {
		if e := check([]byte(f.raw)); e != nil {
			return nil, time.Time{}, e
		}
	}
	return []byte(f.raw), time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC), nil
}
func TestSearchContracts(t *testing.T) {
	for source, fixture := range map[string]string{"aym": aymFixture, "danistay": danistayFixture, "aihm": hudocFixture} {
		t.Run(source, func(t *testing.T) {
			f := &fakeFetch{raw: fixture}
			c := Client{source: source, fetch: f, maxPageSize: 20}
			out, e := c.Search(context.Background(), SearchRequest{Query: "  “ifade   özgürlüğü”  ", Page: 2, PageSize: 3})
			if e != nil {
				t.Fatal(e)
			}
			if f.calls != 1 || out.RelevanceChecked || len(out.Results) != 1 || out.Filters.Query != "ifade özgürlüğü" || *out.Results[0].DecisionDate != "2024-01-02" {
				t.Fatal(out, f.calls)
			}
			switch source {
			case "aym":
				var p map[string]any
				_ = json.Unmarshal(f.body, &p)
				if p["kararTipi"] != "BireyselBasvuru" || p["page"] != float64(2) || p["query"] != `"ifade özgürlüğü"` {
					t.Fatal(p)
				}
				if out.Results[0].CaseNumber != nil {
					t.Fatal("fabricated case number")
				}
			case "danistay":
				var p struct {
					Data struct {
						And  []string `json:"andKelimeler"`
						Page int      `json:"pageNumber"`
					}
				}
				_ = json.Unmarshal(f.body, &p)
				if len(p.Data.And) != 1 || p.Data.And[0] != `"ifade özgürlüğü"` || p.Data.Page != 2 {
					t.Fatal(p)
				}
			case "aihm":
				if f.params.Get("start") != "3" || f.params.Get("length") != "3" || !strings.Contains(f.params.Get("query"), "documentcollectionid2=JUDGMENTS") {
					t.Fatal(f.params)
				}
			}
		})
	}
}
func TestInvalidBeforeNetwork(t *testing.T) {
	for _, source := range []string{"aym", "danistay", "aihm"} {
		f := &fakeFetch{}
		c := Client{source: source, fetch: f, maxPageSize: 20}
		for _, r := range []SearchRequest{{Query: ""}, {Query: "ab"}, {Query: `foo" OR bar`}, {Query: "abc", PageSize: 21}, {Query: "abc", Page: -1}, {Query: "abc", Language: "BAD"}, {Query: "abc", DecisionType: "BAD"}} {
			if _, e := c.Search(context.Background(), r); e == nil {
				t.Fatal(source, r)
			}
		}
		for _, id := range []string{"../123", "https://example.org", "001-1/../../", "", "bb:123"} {
			if _, e := c.GetDecision(context.Background(), id); e == nil {
				t.Fatal(id)
			}
		}
		if f.calls != 0 {
			t.Fatal("invalid input made request")
		}
	}
}
func TestRejectMalformedOrWrongSource(t *testing.T) {
	for _, tc := range []struct{ source, body string }{
		{"aym", `{"data":null}`}, {"aym", `{"data":[] ,"total":-1}`}, {"aym", strings.Replace(aymFixture, "BireyselBasvuru", "SiyasiParti", 1)},
		{"danistay", `{"metadata":{"FMTY":"ERROR"},"data":{"data":[]}}`}, {"danistay", `<html>captcha</html>`},
		{"aihm", strings.Replace(hudocFixture, "\"ENG\"", "\"TUR\"", 1)}, {"aihm", `{"results":null}`}, {"aihm", strings.Replace(hudocFixture, "001-123", "../123", 1)},
	} {
		r := SearchRequest{Query: "abc"}
		_ = r.normalize(tc.source)
		if _, e := parseSearch(tc.source, []byte(tc.body), r, time.Now()); e == nil {
			t.Fatal(tc)
		}
	}
}
func TestDocumentsAndLinks(t *testing.T) {
	html := "<p>" + strings.Repeat("Sentetik karar 😀. ", 12) + "</p><script>EVIL()</script><p>İkinci paragraf.</p>"
	for _, source := range []string{"aym", "danistay", "aihm"} {
		id := "001-123"
		body := []byte(html)
		if source == "aym" {
			id = "bb:" + uuid
			body, _ = json.Marshal(map[string]any{"data": []any{map[string]any{"id": uuid, "kararTipi": "BireyselBasvuru", "icerik": html}}})
		}
		if source == "danistay" {
			id = "123"
			body, _ = json.Marshal(map[string]any{"data": html})
		}
		f := &fakeFetch{raw: string(body)}
		c := Client{source: source, fetch: f, maxPageSize: 20}
		d, e := c.GetDecision(context.Background(), id)
		if e != nil {
			t.Fatal(source, e)
		}
		if strings.Contains(d.Text, "EVIL") || !strings.Contains(d.Text, "\n\n") || f.calls != 1 {
			t.Fatal(d)
		}
		maxChars := 4
		part, e := yargitay.ChunkContent(d, yargitay.DocumentRequest{DocumentID: id, MaxChars: &maxChars})
		if e != nil || part.DocumentID != id || !part.IsTruncated {
			t.Fatal(part, e)
		}
		u, e := url.Parse(d.SourceURL)
		if e != nil || u.Scheme != "https" {
			t.Fatal(d.SourceURL)
		}
		if source == "aym" {
			b, e := base64.RawURLEncoding.DecodeString(u.Query().Get("id"))
			if e != nil || string(b) != "kbb:"+uuid {
				t.Fatal(u)
			}
		}
	}
	if SourceURL("aihm", "../abc") != "" {
		t.Fatal("unsafe link")
	}
}
func TestAYMDocumentIDMismatch(t *testing.T) {
	raw := []byte(`{"data":[{"id":"11111111-2222-3333-4444-555555555555","kararTipi":"NormDenetimi","icerik":"long enough"}]}`)
	if _, e := parseDocument("aym", "bb:"+uuid, raw, time.Now()); e == nil {
		t.Fatal("wrong decision accepted")
	}
}
