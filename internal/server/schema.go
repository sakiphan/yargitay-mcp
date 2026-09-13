// SPDX-License-Identifier: AGPL-3.0-only
package server

func objectSchema(properties map[string]any, required ...string) map[string]any {
	m := map[string]any{"type": "object", "properties": properties, "additionalProperties": false}
	if len(required) > 0 {
		m["required"] = required
	}
	return m
}
func integer(lo, hi int) map[string]any {
	return map[string]any{"type": []string{"integer", "null"}, "minimum": lo, "maximum": hi}
}
func enum(values ...string) map[string]any {
	all := make([]any, 0, len(values)+1)
	for _, v := range values {
		all = append(all, v)
	}
	all = append(all, nil)
	return map[string]any{"type": []string{"string", "null"}, "enum": all}
}
func searchSchema() map[string]any {
	p := map[string]any{"query": map[string]any{"type": []string{"string", "null"}, "maxLength": 1000}, "unit": map[string]any{"type": []string{"string", "null"}, "maxLength": 100}, "sort_by": enum("decision_date", "case_number", "decision_number"), "sort_direction": enum("asc", "desc"), "page": integer(1, 1000), "page_size": integer(1, 20)}
	for _, k := range []string{"case_year", "decision_year"} {
		p[k] = integer(1900, 2100)
	}
	for _, k := range []string{"case_number_from", "case_number_to", "decision_number_from", "decision_number_to"} {
		p[k] = integer(1, 99999999)
	}
	for _, k := range []string{"decision_date_from", "decision_date_to"} {
		p[k] = map[string]any{"type": []string{"string", "null"}, "format": "date"}
	}
	return objectSchema(p)
}
func documentSchema() map[string]any {
	return objectSchema(map[string]any{"document_id": map[string]any{"type": "string", "pattern": "^[0-9]{1,32}$"}, "offset": map[string]any{"type": []string{"integer", "null"}, "minimum": 0}, "max_chars": integer(1, 100000), "output_format": enum("markdown", "text"), "expected_content_sha256": map[string]any{"type": []string{"string", "null"}, "pattern": "^[a-f0-9]{64}$"}}, "document_id")
}
