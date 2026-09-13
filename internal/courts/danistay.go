// SPDX-License-Identifier: AGPL-3.0-only
package courts

import (
	"bytes"
	"encoding/json"
	"github.com/sakiphan/yargitay-mcp/internal/yargitay"
	"golang.org/x/net/html"
	"strings"
)

// The official viewer reads #hiddencontent.text() into its visible container.
// The outer HTML parser decodes its entities once; CleanHTML then sanitizes the
// resulting decision markup. Never return the surrounding viewer or scripts.
func danistayDocumentHTML(raw []byte) (string, error) {
	if json.Valid(raw) {
		root, e := decode(raw)
		if e != nil {
			return "", e
		}
		var body string
		if json.Unmarshal(root["data"], &body) != nil {
			return "", yargitay.ErrSchema
		}
		return body, nil
	}
	root, e := html.Parse(bytes.NewReader(raw))
	if e != nil {
		return "", yargitay.ErrSchema
	}
	var candidates []*html.Node
	challenge := false
	var walk func(*html.Node)
	walk = func(n *html.Node) {
		if n.Type == html.ElementNode {
			switch n.Data {
			case "head", "script", "style", "template", "noscript":
				return
			}
			for _, a := range n.Attr {
				if a.Key == "id" && a.Val == "hiddencontent" {
					candidates = append(candidates, n)
				}
				if a.Key == "class" {
					for _, v := range strings.Fields(a.Val) {
						if v == "g-recaptcha" || v == "h-captcha" {
							challenge = true
						}
					}
				}
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			walk(c)
		}
	}
	walk(root)
	if len(candidates) != 1 {
		if len(candidates) == 0 && challenge {
			return "", yargitay.ErrAccess
		}
		return "", yargitay.ErrSchema
	}
	var b strings.Builder
	for n := candidates[0].FirstChild; n != nil; n = n.NextSibling {
		if n.Type == html.CommentNode {
			continue
		}
		if n.Type != html.TextNode {
			return "", yargitay.ErrSchema
		}
		b.WriteString(n.Data)
	}
	if strings.TrimSpace(b.String()) == "" {
		return "", yargitay.ErrSchema
	}
	return b.String(), nil
}
