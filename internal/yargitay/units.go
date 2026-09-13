// SPDX-License-Identifier: AGPL-3.0-only
package yargitay

import (
	_ "embed"
	"encoding/json"
)

//go:embed units.json
var unitsJSON []byte

type Unit struct {
	Label          string  `json:"label"`
	Category       string  `json:"category"`
	UpstreamField  string  `json:"upstream_field"`
	UpstreamValue  string  `json:"upstream_value"`
	Observed       bool    `json:"observed"`
	Active         *bool   `json:"active"`
	Source         string  `json:"source"`
	LastVerifiedAt *string `json:"last_verified_at"`
}
type UnitRequest struct {
	Category *string `json:"category,omitempty" jsonschema:"all (varsayılan), boards, civil veya criminal"`
}
type UnitResponse struct {
	Units    []Unit    `json:"units"`
	Warnings []Warning `json:"warnings"`
}

func Units(category string) (UnitResponse, error) {
	if category == "" {
		category = "all"
	}
	if category != "all" && category != "boards" && category != "civil" && category != "criminal" {
		return UnitResponse{}, ErrInvalid
	}
	var all []Unit
	if err := json.Unmarshal(unitsJSON, &all); err != nil {
		return UnitResponse{}, ErrSchema
	}
	result := UnitResponse{Units: []Unit{}, Warnings: []Warning{{"activity_unverified", "Formda görünmek güncel aktif daire statüsünü doğrulamaz."}}}
	for _, u := range all {
		if category == "all" || category == u.Category {
			result.Units = append(result.Units, u)
		}
	}
	return result, nil
}
func FindUnit(label string) (Unit, bool) {
	all, _ := Units("all")
	for _, u := range all.Units {
		if label == u.Label {
			return u, true
		}
	}
	return Unit{}, false
}
