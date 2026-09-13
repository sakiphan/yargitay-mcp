// SPDX-License-Identifier: AGPL-3.0-only
package server

import (
	"context"
	"encoding/json"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/sakiphan/yargitay-mcp/internal/courts"
	"github.com/sakiphan/yargitay-mcp/internal/yargitay"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"
)

type fakeCourt struct {
	name  string
	calls int
}

func (c *fakeCourt) Source() string { return c.name }
func (c *fakeCourt) Search(_ context.Context, r courts.SearchRequest) (courts.SearchResponse, error) {
	c.calls++
	return courts.SearchResponse{Source: c.name, Filters: r, QueryMode: "phrase", Results: []courts.Summary{}, Warnings: []yargitay.Warning{}}, nil
}
func (c *fakeCourt) GetDecision(_ context.Context, id string) (yargitay.Document, error) {
	c.calls++
	return yargitay.Document{ID: id, Text: "Sentetik 😀 <script>EVIL</script> karar", Markdown: "Sentetik 😀 karar", SourceURL: courts.SourceURL(c.name, id)}, nil
}
func TestCourtToolsAndNamespaces(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	srcs := []CourtSource{&fakeCourt{name: "aym"}, &fakeCourt{name: "danistay"}, &fakeCourt{name: "aihm"}}
	original := &fakeSource{}
	reader, e := StartReader(original, 20, srcs...)
	if e != nil {
		t.Fatal(e)
	}
	defer reader.Close()
	s := New(original, "test", reader.URL)
	AddCourtTools(s, srcs, reader.CourtURL)
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
	tools, e := cs.ListTools(ctx, nil)
	if e != nil || len(tools.Tools) != 9 {
		t.Fatal(tools, e)
	}
	for _, tool := range tools.Tools {
		if tool.OutputSchema == nil || !tool.Annotations.ReadOnlyHint {
			t.Fatal(tool.Name)
		}
		if strings.HasPrefix(tool.Name, "search_") && (!strings.Contains(tool.Description, "doğrudan kullanıcıya") || strings.Contains(tool.Description, "Kısa numaralı liste ve Kararı aç bağlantısı verin")) {
			t.Fatal("candidate listing instruction leaked", tool.Name)
		}
		if strings.HasPrefix(tool.Name, "get_") && !strings.Contains(tool.Description, "nihai hüküm") {
			t.Fatal("missing operative holding check", tool.Name)
		}
	}
	for _, required := range []string{"UYGUN EMSAL SUNMA AKIŞI", "uygun / uygun değil / belirsiz", "İncelediğim adaylar arasında", "Taraf iddiası", "required_phrases", "Ek yapay zekâ API çağrısı yoktur"} {
		if !strings.Contains(cs.InitializeResult().Instructions, required) {
			t.Fatal("missing client review rule", required)
		}
	}
	for source, id := range map[string]string{"aym": "bb:11111111-2222-3333-4444-555555555555", "danistay": "123", "aihm": "001-123"} {
		r, e := cs.CallTool(ctx, &mcp.CallToolParams{Name: "get_" + source + "_decision", Arguments: map[string]any{"document_id": id, "max_chars": 4}})
		if e != nil || r.IsError {
			t.Fatal(source, r, e)
		}
		b, _ := json.Marshal(r.StructuredContent)
		var chunk courts.Chunk
		if json.Unmarshal(b, &chunk) != nil || chunk.Source != source || chunk.DocumentID != id || chunk.ViewURL == nil {
			t.Fatal(string(b))
		}
		resp, e := http.Get(*chunk.ViewURL)
		if e != nil {
			t.Fatal(e)
		}
		body, _ := io.ReadAll(resp.Body)
		resp.Body.Close()
		if resp.StatusCode != 200 || strings.Contains(string(body), "<script>EVIL") || !strings.Contains(string(body), "&lt;script&gt;") {
			t.Fatal(source, string(body))
		}
		r, e = cs.CallTool(ctx, &mcp.CallToolParams{Name: "search_" + source + "_decisions", Arguments: map[string]any{"query": "sentetik"}})
		if e != nil || r.IsError {
			t.Fatal(source, r, e)
		}
	}
	for _, args := range []map[string]any{{"query": "abc", "page_size": 21}, {"query": "abc", "language": "ENG"}, {"query": "abc", "unknown": 1}} {
		r, e := cs.CallTool(ctx, &mcp.CallToolParams{Name: "search_danistay_decisions", Arguments: args})
		if e == nil && !r.IsError {
			t.Fatal("unsupported filter accepted", args)
		}
	}
	if original.calls.Load() != 0 {
		t.Fatal("cross-source leakage")
	}
	for _, tool := range []string{"search_aym_decisions", "search_aihm_decisions", "search_yargitay_decisions"} {
		r, e := cs.CallTool(ctx, &mcp.CallToolParams{Name: tool, Arguments: map[string]any{"query": "abc", "required_phrases": []string{"extra"}}})
		if e == nil && !r.IsError {
			t.Fatal("unsupported AND filter accepted", tool)
		}
	}
	if reader.CourtURL("aym", "../123") != nil {
		t.Fatal("unsafe viewer URL")
	}
}
