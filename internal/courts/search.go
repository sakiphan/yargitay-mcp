// SPDX-License-Identifier: AGPL-3.0-only
package courts

import (
	"context"
	"encoding/json"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/sakiphan/yargitay-mcp/internal/yargitay"
)

func (c *Client) Search(ctx context.Context, r SearchRequest) (SearchResponse, error) {
	if e := r.normalize(c.source); e != nil {
		return SearchResponse{}, e
	}
	if r.PageSize > c.maxPageSize {
		return SearchResponse{}, yargitay.ErrInvalid
	}
	params, body := searchPayload(c.source, r)
	raw, at, e := c.fetch.OfficialRequest(ctx, "search", params, body, func(raw []byte) error { _, e := parseSearch(c.source, raw, r, time.Time{}); return e })
	if e != nil {
		return SearchResponse{}, e
	}
	return parseSearch(c.source, raw, r, at)
}
func searchPayload(source string, r SearchRequest) (url.Values, []byte) {
	phrase := `"` + r.Query + `"`
	var payload any
	switch source {
	case "aym":
		kind := "BireyselBasvuru"
		if r.DecisionType == "norm_review" {
			kind = "NormDenetimi"
		}
		payload = map[string]any{"kararTipi": kind, "query": phrase, "page": r.Page, "size": r.PageSize}
	case "danistay":
		and := []string{phrase}
		for _, p := range r.RequiredPhrases {
			and = append(and, `"`+p+`"`)
		}
		payload = map[string]any{"data": map[string]any{"andKelimeler": and, "orKelimeler": []string{}, "notAndKelimeler": []string{}, "notOrKelimeler": []string{}, "pageNumber": r.Page, "pageSize": r.PageSize, "siralama": "3", "siralamaDirection": "desc"}}
	case "aihm":
		collection := "JUDGMENTS"
		if r.DecisionType == "decisions" {
			collection = "DECISIONS"
		}
		return url.Values{"query": {"(contentsitename=ECHR) AND (documentcollectionid2=" + collection + ") AND (languageisocode=" + r.Language + ") AND (" + phrase + ")"}, "select": {"itemid,docname,appno,kpdate,languageisocode,doctype,documentcollectionid2,originatingbody"}, "sort": {"kpdate Descending"}, "start": {strconv.Itoa((r.Page - 1) * r.PageSize)}, "length": {strconv.Itoa(r.PageSize)}}, nil
	}
	b, _ := json.Marshal(payload)
	return nil, b
}
func parseSearch(source string, raw []byte, r SearchRequest, at time.Time) (SearchResponse, error) {
	out := SearchResponse{Source: source, Filters: r, QueryMode: "phrase", Results: []Summary{}, FetchedAt: at, Warnings: []yargitay.Warning{}}
	out.Review = yargitay.PendingRelevanceReview(false)
	if len(r.RequiredPhrases) > 0 {
		out.QueryMode = "all_phrases"
	}
	root, e := decode(raw)
	if e != nil {
		return out, e
	}
	var rows []row
	var total json.RawMessage
	switch source {
	case "aym":
		rows, e = rowsOf(root["data"], r.PageSize)
		total = root["total"]
	case "danistay":
		var data row
		data, e = decode(root["data"])
		if e == nil {
			rows, e = rowsOf(data["data"], r.PageSize)
			total = data["recordsFiltered"]
		}
	case "aihm":
		rows, e = rowsOf(root["results"], r.PageSize)
		total = root["resultcount"]
	}
	if e != nil {
		return out, e
	}
	if len(total) > 0 && string(total) != "null" {
		var n int
		if json.Unmarshal(total, &n) != nil || n < 0 {
			return out, yargitay.ErrSchema
		}
		out.Total = &n
	}
	for _, row := range rows {
		s := Summary{Source: source}
		switch source {
		case "aym":
			s, e = aymSummary(row)
			if e == nil && (r.DecisionType == "norm_review") != strings.HasPrefix(s.ID, "norm:") {
				e = yargitay.ErrSchema
			}
		case "danistay":
			if json.Unmarshal(row["id"], &s.ID) != nil {
				s.ID = string(row["id"])
			}
			e = fields(row, map[string]**string{"daireKurul": &s.Unit, "esasNo": &s.CaseNumber, "kararNo": &s.DecisionNumber, "kararTarihi": &s.DecisionDate})
		case "aihm":
			var cols map[string]json.RawMessage
			cols, e = decode(row["columns"])
			if e == nil {
				s.ID, e = required(cols, "itemid")
			}
			if e == nil {
				e = fields(cols, map[string]**string{"docname": &s.Title, "appno": &s.ApplicationNumber, "kpdate": &s.DecisionDate, "languageisocode": &s.Language, "doctype": &s.DecisionType, "originatingbody": &s.Unit})
			}
			if e == nil && (s.Language == nil || *s.Language != r.Language) {
				e = yargitay.ErrSchema
			}
			if e == nil {
				collection, err := required(cols, "documentcollectionid2")
				want := "JUDGMENTS"
				if r.DecisionType == "decisions" {
					want = "DECISIONS"
				}
				if err != nil || !strings.Contains(";"+collection+";", ";"+want+";") {
					e = yargitay.ErrSchema
				}
			}
		}
		if e != nil {
			return out, e
		}
		if !ValidID(source, s.ID) {
			return out, yargitay.ErrSchema
		}
		if s.DecisionDate != nil {
			layout := "2006-01-02"
			if source == "danistay" {
				layout = "02.01.2006"
			}
			if source == "aihm" {
				layout = "2006-01-02T15:04:05"
			}
			t, err := time.Parse(layout, *s.DecisionDate)
			if err != nil {
				return out, yargitay.ErrSchema
			}
			d := t.Format("2006-01-02")
			s.DecisionDate = &d
		}
		s.SourceURL = SourceURL(source, s.ID)
		out.Results = append(out.Results, s)
	}
	if source == "aihm" {
		out.Warnings = append(out.Warnings, yargitay.Warning{Code: "language_and_coverage", Message: "HUDOC seçilen dildeki yayımlanmış belgeleri döndürür; Türkçe kapsam eksik olabilir. Metnin çeviri/resmî dil statüsünü kontrol edin; araç çeviri yapmaz."})
	}
	return out, nil
}
