// SPDX-License-Identifier: AGPL-3.0-only
package server

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/sakiphan/yargitay-mcp/internal/yargitay"
)

// These tests validate shipped examples and MCP wiring, not model-generated plans.
func TestQueryPlanningExamplesUseSupportedTools(t *testing.T) {
	var examples []struct {
		Source      string         `json:"source"`
		Question    string         `json:"question"`
		Primary     map[string]any `json:"primary"`
		Alternative map[string]any `json:"alternative"`
		Verify      []string       `json:"verify_in_text"`
	}
	if err := json.Unmarshal([]byte(queryExamples), &examples); err != nil || len(examples) != 4 {
		t.Fatal("invalid planning examples", err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	original := &fakeSource{}
	aym, danistay, aihm := &fakeCourt{name: "aym"}, &fakeCourt{name: "danistay"}, &fakeCourt{name: "aihm"}
	s := New(original, "test", nil)
	AddCourtTools(s, []CourtSource{aym, danistay, aihm}, nil)
	ct, st := mcp.NewInMemoryTransports()
	ss, err := s.Connect(ctx, st, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer ss.Close()
	cs, err := mcp.NewClient(&mcp.Implementation{Name: "planning-test", Version: "1"}, nil).Connect(ctx, ct, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer cs.Close()
	if original.calls.Load() != 0 || aym.calls+danistay.calls+aihm.calls != 0 {
		t.Fatal("initialization executed example searches")
	}
	for _, marker := range []string{"SORUDAN ARAMA PLANI", "query'yi bütünüyle tırnaklar", "desteklenmeyen", "page_size=10", "zaman aşımı", "metadata araması istenmişse anahtar kelime ekleme", "initial police questioning", "verify_in_text"} {
		if !strings.Contains(cs.InitializeResult().Instructions, marker) {
			t.Fatal("missing planning guidance", marker)
		}
	}
	listed, err := cs.ListTools(ctx, nil)
	if err != nil || len(listed.Tools) != 9 {
		t.Fatal("tool surface changed", err)
	}
	for _, tool := range listed.Tools {
		if strings.HasPrefix(tool.Name, "search_") && !strings.Contains(tool.Description, "ayırt edici doğal ifade") {
			t.Fatal("tool-only clients lack planning guidance", tool.Name)
		}
	}
	seen := map[string]bool{}
	for _, example := range examples {
		if seen[example.Source] || example.Question == "" || len(example.Verify) < 3 {
			t.Fatal("invalid source or review checks")
		}
		seen[example.Source] = true
		if example.Primary["query"] == example.Alternative["query"] {
			t.Fatal("duplicate alternative", example.Source)
		}
		for _, args := range []map[string]any{example.Primary, example.Alternative} {
			if args["page"] != float64(1) || args["page_size"] != float64(10) {
				t.Fatal("unbounded example")
			}
			query, ok := args["query"].(string)
			if !ok || strings.ContainsAny(query, "\"\\*") || strings.Contains(query, " AND ") || strings.Contains(query, " OR ") {
				t.Fatal("nonliteral example", query)
			}
			if example.Source != "danistay" && args["required_phrases"] != nil {
				t.Fatal("unsupported AND filter")
			}
			result, err := cs.CallTool(ctx, &mcp.CallToolParams{Name: "search_" + example.Source + "_decisions", Arguments: args})
			if err != nil || result.IsError {
				t.Fatal("shipped example rejected", example.Source, err, result)
			}
		}
		// An alternative may change wording, not the user's source/type/language/filters.
		delete(example.Primary, "query")
		delete(example.Alternative, "query")
		if !reflect.DeepEqual(example.Primary, example.Alternative) {
			t.Fatal("alternative dropped a constraint", example.Source)
		}
	}
	if original.calls.Load() != 2 || aym.calls != 2 || danistay.calls != 2 || aihm.calls != 2 {
		t.Fatal("unexpected request fan-out")
	}
}

// These assertions guard the shipped policy, not the reasoning of a live model.
func TestSearchCoverageInstructions(t *testing.T) {
	for name, instructions := range map[string]string{
		"initialize":   responseInstructions,
		"search tools": searchReviewInstructions,
	} {
		t.Run(name, func(t *testing.T) {
			for _, required := range []string{
				"page_size=10", "page_size=20", "Yargıtay", "tümünü/hepsini/ne kadar varsa",
				"istemci kontrollü sayfalama", "sıralama", "sayfa boyutu",
				"İlk uygun kararda durma", "konu taraması", "Metin2", "Metin 2", "literal",
				"total", "null", "çelişkili sayfa", "hiç yeni kimlik", "page=1000",
				"kaynak+kimliği tekrar sayma", "benzersiz", "tam incelenen", "bekleyen",
				"sonraki sayfa", "offset/format/hash", "kısmi", "tüm arşivdeki bütün",
				"CAPTCHA", "429", "şema", "eşzamanlılık 1", "3 saniye",
			} {
				if !strings.Contains(strings.ToLower(instructions), strings.ToLower(required)) {
					t.Fatalf("missing coverage guidance: %q", required)
				}
			}
			for _, obsolete := range []string{
				"page_size=2;", "en fazla 2 arama", "en fazla 3 farklı",
				"En fazla üç adayla", "İlk arama uygun kanıt sağladıysa dur",
			} {
				if strings.Contains(instructions, obsolete) {
					t.Fatalf("early-stop policy remains: %q", obsolete)
				}
			}
		})
	}
	for _, required := range []string{"konu taraması", "required_checks", "ortak hukuki soruyu zorunlu tutmayın", "Eksik metni ilgisiz saymayın"} {
		if !strings.Contains(documentReviewInstructions, required) {
			t.Fatalf("document tools lack topic-review guidance: %q", required)
		}
	}
}

type pagedPlanningSource struct{ fakeSource }

func (s *pagedPlanningSource) Search(_ context.Context, r yargitay.SearchRequest) (yargitay.SearchResponse, error) {
	if err := r.Normalize(); err != nil {
		return yargitay.SearchResponse{}, err
	}
	s.calls.Add(1)
	total := 21
	rows := []yargitay.Summary{}
	page, pageSize := *r.Page, *r.PageSize
	for i := (page-1)*pageSize + 1; i <= page*pageSize && i <= total; i++ {
		rows = append(rows, yargitay.Summary{ID: fmt.Sprint(i)})
	}
	return yargitay.SearchResponse{Filters: r, Total: &total, Page: *r.Page, PageSize: *r.PageSize, Results: rows}, nil
}

func TestExplicitPageCallsRemainSinglePageAndAllowMoreThanThreeDocuments(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	src := &pagedPlanningSource{}
	ct, st := mcp.NewInMemoryTransports()
	ss, err := New(src, "test", nil).Connect(ctx, st, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer ss.Close()
	cs, err := mcp.NewClient(&mcp.Implementation{Name: "coverage-test", Version: "1"}, nil).Connect(ctx, ct, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer cs.Close()
	for page, wantCount := range []int{20, 1, 0} {
		result, err := cs.CallTool(ctx, &mcp.CallToolParams{Name: "search_yargitay_decisions", Arguments: map[string]any{"query": "Sentetik Oyun", "page": page + 1, "page_size": 20}})
		if err != nil || result.IsError {
			t.Fatal("explicit page rejected", err, result)
		}
		raw, err := json.Marshal(result.StructuredContent)
		if err != nil {
			t.Fatal(err)
		}
		var out yargitay.SearchResponse
		if err := json.Unmarshal(raw, &out); err != nil {
			t.Fatal(err)
		}
		if out.Page != page+1 || out.PageSize != 20 || out.Total == nil || *out.Total != 21 || len(out.Results) != wantCount {
			t.Fatal("page/count information lost", out)
		}
		if out.Filters.Query == nil || *out.Filters.Query != "Sentetik Oyun" || src.calls.Load() != int32(page+1) {
			t.Fatal("filters changed or backend paginated automatically")
		}
	}
	for id := 1; id <= 4; id++ {
		result, err := cs.CallTool(ctx, &mcp.CallToolParams{Name: "get_yargitay_decision", Arguments: map[string]any{"document_id": fmt.Sprint(id)}})
		if err != nil || result.IsError {
			t.Fatal("document review capped", id, err)
		}
	}
	result, err := cs.CallTool(ctx, &mcp.CallToolParams{Name: "search_yargitay_decisions", Arguments: map[string]any{"query": "Sentetik Oyun", "page": 1001, "page_size": 20}})
	if err == nil && !result.IsError {
		t.Fatal("page safety limit bypassed")
	}
	if src.calls.Load() != 7 {
		t.Fatal("unexpected upstream fan-out", src.calls.Load())
	}
}
