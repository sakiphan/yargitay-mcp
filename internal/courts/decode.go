// SPDX-License-Identifier: AGPL-3.0-only
package courts

import (
	"encoding/json"
	"strings"

	"github.com/sakiphan/yargitay-mcp/internal/yargitay"
)

type row map[string]json.RawMessage

func decode(b []byte) (row, error) {
	var r row
	if json.Unmarshal(b, &r) != nil || r == nil {
		return nil, yargitay.ErrSchema
	}
	var meta struct{ FMTY string }
	_ = json.Unmarshal(r["metadata"], &meta)
	if meta.FMTY == "ERROR" {
		return nil, yargitay.ErrUpstream
	}
	var status string
	_ = json.Unmarshal(r["status"], &status)
	if strings.Contains(strings.ToUpper(status), "EXCEPTION") || strings.EqualFold(status, "ERROR") {
		return nil, yargitay.ErrUnavailable
	}
	if string(r["success"]) == "false" {
		return nil, yargitay.ErrUnavailable
	}
	return r, nil
}
func rowsOf(b []byte, limit int) ([]row, error) {
	var rows []row
	if json.Unmarshal(b, &rows) != nil || rows == nil || len(rows) > limit {
		return nil, yargitay.ErrSchema
	}
	for _, r := range rows {
		if r == nil {
			return nil, yargitay.ErrSchema
		}
	}
	return rows, nil
}
func str(r row, k string) (*string, error) {
	v, ok := r[k]
	if !ok || string(v) == "null" {
		return nil, nil
	}
	var s string
	if json.Unmarshal(v, &s) != nil {
		return nil, yargitay.ErrSchema
	}
	if s == "" {
		return nil, nil
	}
	return &s, nil
}
func required(r row, k string) (string, error) {
	s, e := str(r, k)
	if e != nil || s == nil {
		return "", yargitay.ErrSchema
	}
	return *s, nil
}
func fields(r row, pairs map[string]**string) error {
	for k, p := range pairs {
		var e error
		*p, e = str(r, k)
		if e != nil {
			return e
		}
	}
	return nil
}
