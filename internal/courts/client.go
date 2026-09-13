// SPDX-License-Identifier: AGPL-3.0-only
package courts

import (
	"context"
	"net/url"
	"time"

	"github.com/sakiphan/yargitay-mcp/internal/yargitay"
)

type fetcher interface {
	OfficialRequest(context.Context, string, url.Values, []byte, ...func([]byte) error) ([]byte, time.Time, error)
}
type Client struct {
	source      string
	fetch       fetcher
	close       func()
	maxPageSize int
}

func New(source string, cfg yargitay.Config) (*Client, error) {
	c, e := yargitay.NewOfficialClient(source, cfg)
	if e != nil {
		return nil, e
	}
	return &Client{source, c, c.Close, cfg.MaxPageSize}, nil
}
func (c *Client) Close()         { c.close() }
func (c *Client) Source() string { return c.source }
