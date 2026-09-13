// SPDX-License-Identifier: AGPL-3.0-only
package yargitay

import (
	"crypto/sha256"
	"encoding/hex"
	"strings"
	"unicode"

	"golang.org/x/net/html"
)

func containsPhrase(text, phrase string) bool {
	normalize := func(s string) string {
		return strings.Join(strings.Fields(strings.ToLowerSpecial(unicode.TurkishCase, s)), " ")
	}
	return strings.Contains(normalize(text), normalize(phrase))
}

func CleanHTML(body string) (string, string, error) {
	root, err := html.Parse(strings.NewReader(body))
	if err != nil {
		return "", "", ErrSchema
	}
	drop := map[string]bool{"script": true, "style": true, "iframe": true, "object": true, "embed": true, "svg": true, "form": true, "template": true, "noscript": true, "img": true, "head": true}
	blocks := map[string]bool{"p": true, "div": true, "h1": true, "h2": true, "h3": true, "h4": true, "li": true, "tr": true, "blockquote": true}
	var b strings.Builder
	var walk func(*html.Node)
	walk = func(n *html.Node) {
		if n.Type == html.ElementNode && drop[n.Data] {
			return
		}
		if n.Type == html.TextNode {
			b.WriteString(n.Data)
			return
		}
		if n.Type == html.ElementNode && n.Data == "br" {
			b.WriteByte('\n')
			return
		}
		block := n.Type == html.ElementNode && blocks[n.Data]
		if block {
			b.WriteString("\n\n")
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			walk(c)
		}
		if block {
			b.WriteString("\n\n")
		}
	}
	walk(root)
	// Match Python splitlines semantics for Unicode line separators as well as CR/LF.
	raw := strings.ReplaceAll(b.String(), "\r\n", "\n")
	raw = strings.Map(func(r rune) rune {
		if strings.ContainsRune("\r\v\f\x1c\x1d\x1e\u0085\u2028\u2029", r) {
			return '\n'
		}
		return r
	}, raw)
	lines := strings.Split(raw, "\n")
	normalized := make([]string, 0, len(lines))
	blank := false
	for _, line := range lines {
		line = strings.Join(strings.Fields(line), " ")
		if line == "" {
			if !blank {
				normalized = append(normalized, "")
			}
			blank = true
		} else {
			normalized = append(normalized, line)
			blank = false
		}
	}
	text := strings.TrimSpace(strings.Join(normalized, "\n"))
	if text == "" {
		return "", "", ErrSchema
	}
	var md strings.Builder
	for _, r := range text {
		if strings.ContainsRune("\\`*_{}[]()<>#+.!|~-", r) {
			md.WriteByte('\\')
		}
		md.WriteRune(r)
	}
	return text, md.String(), nil
}

func ChunkDocument(d Document, r DocumentRequest) (Chunk, error) {
	if err := r.Normalize(); err != nil {
		return Chunk{}, err
	}
	return ChunkContent(d, r)
}

// ChunkContent is used after the source adapter has validated its own ID.
func ChunkContent(d Document, r DocumentRequest) (Chunk, error) {
	if err := r.NormalizeRange(); err != nil {
		return Chunk{}, err
	}
	text := d.Markdown
	if *r.OutputFormat == "text" {
		text = d.Text
	}
	sum := sha256.Sum256([]byte(text))
	hash := hex.EncodeToString(sum[:])
	if r.ExpectedHash != nil && *r.ExpectedHash != hash {
		return Chunk{}, ErrChanged
	}
	runes := []rune(text)
	start := *r.Offset
	if start > len(runes) {
		return Chunk{}, ErrInvalid
	}
	end := start + min(*r.MaxChars, len(runes)-start)
	c := Chunk{DocumentID: d.ID, Content: string(runes[start:end]), OutputFormat: *r.OutputFormat, Offset: start, ReturnedChars: end - start, TotalChars: len(runes), IsTruncated: end < len(runes), SourceURL: d.SourceURL, FetchedAt: d.FetchedAt, ContentSHA256: hash, Warnings: []Warning{}}
	c.Review = PendingRelevanceReview(start == 0 && end == len(runes))
	if c.IsTruncated {
		c.NextOffset = &end
	}
	return c, nil
}
