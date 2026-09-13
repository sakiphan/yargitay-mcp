// SPDX-License-Identifier: AGPL-3.0-only
package courts

import (
	"context"
	"encoding/json"
	"net/url"
	"strings"
	"time"

	"github.com/sakiphan/yargitay-mcp/internal/yargitay"
)

func (c *Client) GetDecision(ctx context.Context, id string) (yargitay.Document, error) {
	if !ValidID(c.source, id) {
		return yargitay.Document{}, yargitay.ErrInvalid
	}
	var p url.Values
	var b []byte
	switch c.source {
	case "aym":
		b, _ = json.Marshal(map[string]any{"id": strings.SplitN(id, ":", 2)[1], "kararTipi": aymType(id), "size": 1})
	case "danistay":
		p = url.Values{"id": {id}, "arananKelime": {""}}
	case "aihm":
		p = url.Values{"id": {id}, "library": {"ECHR"}}
	}
	raw, at, e := c.fetch.OfficialRequest(ctx, "document", p, b, func(raw []byte) error { _, e := parseDocument(c.source, id, raw, time.Time{}); return e })
	if e != nil {
		return yargitay.Document{}, e
	}
	return parseDocument(c.source, id, raw, at)
}
func parseDocument(source, id string, raw []byte, at time.Time) (yargitay.Document, error) {
	body := string(raw)
	if source == "aym" {
		root, e := decode(raw)
		if e != nil {
			return yargitay.Document{}, e
		}
		rows, e := rowsOf(root["data"], 1)
		if e != nil {
			return yargitay.Document{}, e
		}
		if len(rows) == 0 {
			return yargitay.Document{}, yargitay.ErrNotFound
		}
		row := rows[0]
		s, e := aymSummary(row)
		if e != nil || s.ID != id {
			return yargitay.Document{}, yargitay.ErrSchema
		}
		if json.Unmarshal(row["icerik"], &body) != nil {
			return yargitay.Document{}, yargitay.ErrSchema
		}
	} else if source == "danistay" {
		var e error
		body, e = danistayDocumentHTML(raw)
		if e != nil {
			return yargitay.Document{}, e
		}
	}
	text, md, e := yargitay.CleanHTML(body)
	if e != nil {
		return yargitay.Document{}, e
	}
	// Reject empty conversions and common error/challenge pages, not legal content.
	lower := strings.ToLower(text)
	if len([]rune(text)) < 100 || len(text) < 2000 && (strings.Contains(lower, "verify you are human") || strings.Contains(lower, "access denied")) {
		return yargitay.Document{}, yargitay.ErrSchema
	}
	return yargitay.Document{ID: id, Text: text, Markdown: md, SourceURL: SourceURL(source, id), FetchedAt: at}, nil
}
