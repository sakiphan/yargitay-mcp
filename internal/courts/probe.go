// SPDX-License-Identifier: AGPL-3.0-only
package courts

import (
	"context"
	"errors"
	"github.com/sakiphan/yargitay-mcp/internal/yargitay"
	"strings"
	"unicode"
)

// Probe requires CLI --live opt-in. One search and at most one document; no retry.
func Probe(ctx context.Context, source, kind string) (map[string]any, error) {
	report := map[string]any{"source": source, "search_response_received": false, "search_verified": false, "document_verified": false, "legal_relevance_verified": false}
	cfg := yargitay.DefaultConfig()
	cfg.CacheEnabled = false
	cfg.MaxRetries = 0
	c, e := New(source, cfg)
	if e != nil {
		return report, e
	}
	defer c.Close()
	q := map[string]string{"aym": "ifade özgürlüğü", "danistay": "vergi", "aihm": "freedom of expression"}[source]
	if source == "aym" && kind == "norm_review" {
		q = "mülkiyet"
	}
	r, e := c.Search(ctx, SearchRequest{Query: q, DecisionType: kind, PageSize: 1})
	if e != nil {
		return report, e
	}
	report["search_response_received"] = true
	report["returned_rows"] = len(r.Results)
	if len(r.Results) == 0 {
		return report, errors.New("Örnek arama boş; belge ve ifade doğrulanamadı.")
	}
	d, e := c.GetDecision(ctx, r.Results[0].ID)
	if e != nil {
		return report, e
	}
	report["document_verified"] = true
	report["document_chars"] = len([]rune(d.Text))
	fold := func(s string) string {
		return strings.Join(strings.Fields(strings.ToLowerSpecial(unicode.TurkishCase, s)), " ")
	}
	found := strings.Contains(fold(d.Text), fold(q))
	report["phrase_found_in_first_document"] = found
	report["search_verified"] = found
	if !found {
		return report, errors.New("İlk belgede örnek ifade bulunamadı; ilgili emsal kabul etmeyin.")
	}
	return report, nil
}
