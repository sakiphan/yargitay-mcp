// SPDX-License-Identifier: AGPL-3.0-only
package yargitay

const AYMOrigin = "https://kararlarbilgibankasi.anayasa.gov.tr"
const DanistayOrigin = "https://karararama.danistay.gov.tr"
const AIHMOrigin = "https://hudoc.echr.coe.int"

func NewOfficialClient(source string, cfg Config) (*Client, error) {
	origin := map[string]string{"aym": AYMOrigin, "danistay": DanistayOrigin, "aihm": AIHMOrigin}[source]
	if origin == "" {
		return nil, ErrInvalid
	}
	c, err := NewClient(cfg)
	if err == nil {
		c.origin = origin
	}
	return c, err
}
