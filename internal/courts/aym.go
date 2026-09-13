// SPDX-License-Identifier: AGPL-3.0-only
package courts

import (
	"strings"

	"github.com/sakiphan/yargitay-mcp/internal/yargitay"
)

func aymType(id string) string {
	if strings.HasPrefix(id, "bb:") {
		return "BireyselBasvuru"
	}
	return "NormDenetimi"
}
func aymSummary(r row) (Summary, error) {
	s := Summary{Source: "aym"}
	id, e := required(r, "id")
	if e != nil {
		return s, e
	}
	kind, e := required(r, "kararTipi")
	if e != nil {
		return s, e
	}
	prefix := "bb:"
	if kind == "NormDenetimi" {
		prefix = "norm:"
	} else if kind != "BireyselBasvuru" {
		return s, yargitay.ErrSchema
	}
	s.ID = prefix + id
	if !ValidID("aym", s.ID) {
		return s, yargitay.ErrSchema
	}
	s.DecisionType = &kind
	e = fields(r, map[string]**string{"basvuruAdi": &s.Title, "kararVerenBirimLabel": &s.Unit, "basvuruNo": &s.ApplicationNumber, "esasNo": &s.CaseNumber, "kararNo": &s.DecisionNumber, "kararTarihi": &s.DecisionDate})
	return s, e
}
