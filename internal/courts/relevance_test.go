// SPDX-License-Identifier: AGPL-3.0-only
package courts

import (
	"context"
	"encoding/json"
	"strings"
	"testing"
)

func TestDanistayAllPhrasesAreOneBoundedRequest(t *testing.T) {
	f := &fakeFetch{raw: danistayFixture}
	c := Client{source: "danistay", fetch: f, maxPageSize: 20}
	out, e := c.Search(context.Background(), SearchRequest{Query: "takdir komisyonuna sevk", RequiredPhrases: []string{" “zamanaşımı” ", "ZAMANAŞIMI", "katma  değer vergisi"}, PageSize: 3})
	if e != nil {
		t.Fatal(e)
	}
	var p struct {
		Data struct {
			And  []string `json:"andKelimeler"`
			Page int      `json:"pageNumber"`
			Size int      `json:"pageSize"`
		}
	}
	if json.Unmarshal(f.body, &p) != nil || f.calls != 1 || p.Data.Page != 1 || p.Data.Size != 3 || len(p.Data.And) != 3 {
		t.Fatal(p, f.calls)
	}
	if p.Data.And[0] != `"takdir komisyonuna sevk"` || p.Data.And[1] != `"zamanaşımı"` || p.Data.And[2] != `"katma değer vergisi"` {
		t.Fatal(p)
	}
	if out.QueryMode != "all_phrases" || out.RelevanceChecked || out.Review.Status != "not_assessed" || out.Review.Evaluator != "calling_model" || out.Review.FullDocumentInThisResponse {
		t.Fatal("lexical search claimed legal approval", out)
	}
	if len(out.Filters.RequiredPhrases) != 2 {
		t.Fatal(out.Filters)
	}
}

func TestRequiredPhrasesFailBeforeNetwork(t *testing.T) {
	cases := []SearchRequest{
		{Query: "abc", RequiredPhrases: []string{"ab"}},
		{Query: "abc", RequiredPhrases: []string{`bad" OR anything`}},
		{Query: "abc", RequiredPhrases: []string{strings.Repeat("x", 201)}},
		{Query: "abc", RequiredPhrases: []string{"one", "two", "three", "four", "five", "six"}},
		{Query: strings.Repeat("x", 999), RequiredPhrases: []string{"more"}},
	}
	f := &fakeFetch{}
	c := Client{source: "danistay", fetch: f, maxPageSize: 20}
	for _, r := range cases {
		if _, e := c.Search(context.Background(), r); e == nil {
			t.Fatal("bad input accepted")
		}
	}
	for _, source := range []string{"aym", "aihm"} {
		c.source = source
		if _, e := c.Search(context.Background(), SearchRequest{Query: "abc", RequiredPhrases: []string{"extra"}}); e == nil {
			t.Fatal("unsupported source", source)
		}
	}
	if f.calls != 0 {
		t.Fatal("invalid input reached network")
	}
}
