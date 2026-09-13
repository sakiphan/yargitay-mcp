// SPDX-License-Identifier: AGPL-3.0-only
package server

import (
	"context"
	"encoding/json"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
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
	for _, marker := range []string{"SORUDAN ARAMA PLANI", "query'yi bütünüyle tırnaklar", "desteklenmeyen", "en fazla 2 arama", "zaman aşımı", "metadata araması istenmişse anahtar kelime ekleme", "initial police questioning", "verify_in_text"} {
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
			if args["page"] != float64(1) || args["page_size"] != float64(2) {
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
