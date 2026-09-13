// SPDX-License-Identifier: AGPL-3.0-only
package server

import (
	"context"
	"encoding/json"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/sakiphan/yargitay-mcp/internal/yargitay"
)

type fakeSource struct{ calls atomic.Int32 }

func (s *fakeSource) Search(_ context.Context, r yargitay.SearchRequest) (yargitay.SearchResponse, error) {
	if e := r.Normalize(); e != nil {
		return yargitay.SearchResponse{}, e
	}
	s.calls.Add(1)
	return yargitay.SearchResponse{Page: 1, PageSize: 10, Results: []yargitay.Summary{{ID: "123", SourceURL: yargitay.BaseURL + "/getDokuman?id=123"}}, Warnings: []yargitay.Warning{}, Source: yargitay.BaseURL, FetchedAt: time.Now().UTC()}, nil
}
func (s *fakeSource) GetDecision(_ context.Context, id string) (yargitay.Document, error) {
	s.calls.Add(1)
	if id == "404" {
		return yargitay.Document{}, yargitay.ErrNotFound
	}
	return yargitay.Document{ID: id, Text: "Sentetik <script>alert(1)</script> karar 😀.", Markdown: "Sentetik karar 😀.", SourceURL: yargitay.BaseURL + "/getDokuman?id=" + id, FetchedAt: time.Now().UTC()}, nil
}

func TestToolsSchemaAndStructuredOutput(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	src := &fakeSource{}
	s := New(src, "test", nil)
	ct, st := mcp.NewInMemoryTransports()
	ss, e := s.Connect(ctx, st, nil)
	if e != nil {
		t.Fatal(e)
	}
	defer ss.Close()
	cs, e := mcp.NewClient(&mcp.Implementation{Name: "test", Version: "1"}, nil).Connect(ctx, ct, nil)
	if e != nil {
		t.Fatal(e)
	}
	defer cs.Close()
	for _, required := range []string{"relevance_checked=false", "numaralı kısa liste", "details veya summary", "2-3 cümlelik özet", "kullanıcı açıkça", "İlgisiz kararı"} {
		if !strings.Contains(cs.InitializeResult().Instructions, required) {
			t.Fatalf("missing response guidance: %s", required)
		}
	}
	listed, e := cs.ListTools(ctx, nil)
	if e != nil || len(listed.Tools) != 3 {
		t.Fatal(listed, e)
	}
	for _, tool := range listed.Tools {
		if tool.OutputSchema == nil || !tool.Annotations.ReadOnlyHint {
			t.Fatal(tool)
		}
	}
	u, e := cs.CallTool(ctx, &mcp.CallToolParams{Name: "list_yargitay_units", Arguments: map[string]any{}})
	if e != nil || u.IsError || src.calls.Load() != 0 {
		t.Fatal(u, e)
	}
	b, _ := json.Marshal(u.StructuredContent)
	var units yargitay.UnitResponse
	if json.Unmarshal(b, &units) != nil || len(units.Units) != 51 {
		t.Fatal(string(b))
	}
	r, e := cs.CallTool(ctx, &mcp.CallToolParams{Name: "search_yargitay_decisions", Arguments: map[string]any{"query": "çalışma"}})
	if e != nil || r.IsError {
		t.Fatal(r, e)
	}
	d, e := cs.CallTool(ctx, &mcp.CallToolParams{Name: "get_yargitay_decision", Arguments: map[string]any{"document_id": "123", "max_chars": 3}})
	if e != nil || d.IsError {
		t.Fatal(d, e)
	}
	for _, args := range []map[string]any{{"query": "abc", "page_size": 21}, {"query": "abc", "extra": "bad"}, {"query": "abc", "page": 0}, {"query": "abc", "decision_date_from": "yesterday"}} {
		r, e := cs.CallTool(ctx, &mcp.CallToolParams{Name: "search_yargitay_decisions", Arguments: args})
		if e == nil && !r.IsError {
			t.Fatal("invalid input accepted", args)
		}
	}
}
