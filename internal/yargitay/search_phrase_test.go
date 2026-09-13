// SPDX-License-Identifier: AGPL-3.0-only
package yargitay

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"testing"
)

func TestLiteralPhraseNormalization(t *testing.T) {
	for _, input := range []string{"fazla çalışma", `"fazla çalışma"`, "“fazla çalışma”", " fazla\n çalışma "} {
		r := SearchRequest{Query: ptr(input)}
		if err := r.Normalize(); err != nil {
			t.Fatal(err)
		}
		if *r.Query != "fazla çalışma" {
			t.Fatal("unexpected normalized phrase")
		}
		if err := r.Normalize(); err != nil {
			t.Fatal("not idempotent", err)
		}
		if got := searchPayload(r)["data"].(map[string]any)["arananKelime"]; got != `"fazla çalışma"` {
			t.Fatal("unquoted or double-quoted phrase")
		}
	}
	for _, input := range []string{`""`, `"ab"`, `abc" OR *`, `abc\def`, `"abc`, `abc"`, `“abc`, `“abc” OR def`, `""abc""`} {
		r := SearchRequest{Query: ptr(input)}
		if r.Normalize() == nil {
			t.Errorf("accepted unsupported expression %q", input)
		}
	}
}

func TestSearchUsesPhraseWithoutAutomaticallyFetchingDocuments(t *testing.T) {
	calls := 0
	c := testClient(t, func(r *http.Request) (*http.Response, error) {
		calls++
		if r.URL.Path != "/aramadetaylist" {
			t.Fatal("unexpected additional request")
		}
		var body map[string]map[string]any
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatal(err)
		}
		if body["data"]["arananKelime"] != `"fazla çalışma"` {
			t.Fatal("plain text would broaden the upstream search")
		}
		return response(200, searchFixture), nil
	})
	r, err := c.Search(context.Background(), SearchRequest{Query: ptr("fazla çalışma"), PageSize: ptr(3)})
	if err != nil {
		t.Fatal(err)
	}
	if calls != 1 || r.QueryMode != "phrase" || r.RelevanceChecked || r.Filters.Query == nil || *r.Filters.Query != "fazla çalışma" || *r.Filters.PageSize != 3 {
		t.Fatal("missing scope or fabricated relevance verification")
	}
}

func TestMetadataOnlySearchDoesNotInventKeyword(t *testing.T) {
	r := SearchRequest{Unit: ptr("9. Hukuk Dairesi")}
	if err := r.Normalize(); err != nil {
		t.Fatal(err)
	}
	if searchPayload(r)["data"].(map[string]any)["arananKelime"] != "" {
		t.Fatal("invented keyword")
	}
	out, err := parseSearch([]byte(searchFixture), r)
	if err != nil || out.QueryMode != "metadata_only" || out.Filters.Query != nil || out.RelevanceChecked {
		t.Fatal("invalid metadata-only scope")
	}
}

func TestProbeDoesNotVerifyIrrelevantOrEmptySearch(t *testing.T) {
	for _, kind := range []string{"match", "irrelevant", "empty"} {
		t.Run(kind, func(t *testing.T) {
			calls := 0
			c := testClient(t, func(r *http.Request) (*http.Response, error) {
				calls++
				if r.URL.Path == "/aramadetaylist" {
					if kind == "empty" {
						return response(200, `{"data":{"data":[],"recordsTotal":0}}`), nil
					}
					return response(200, searchFixture), nil
				}
				body := "<p>Sentetik adli yardım metni.</p>"
				if kind == "match" {
					body = "<p>Sentetik FAZLA ÇALIŞMA metni.</p>"
				}
				encoded, _ := json.Marshal(map[string]string{"data": body})
				return response(200, string(encoded)), nil
			})
			report, err := probeClient(context.Background(), c)
			if (err != nil) != (kind == "irrelevant") {
				t.Fatal("wrong probe error", err)
			}
			if report["search_verified"] != (kind == "match") || report["phrase_found_in_first_document"] != (kind == "match") {
				t.Fatal("fabricated keyword verification")
			}
			want := 2
			if kind == "empty" {
				want = 1
			}
			if calls != want {
				t.Fatal("unexpected request count")
			}
			encoded, _ := json.Marshal(report)
			if strings.Contains(string(encoded), "Sentetik") {
				t.Fatal("probe leaked document body")
			}
		})
	}
}
