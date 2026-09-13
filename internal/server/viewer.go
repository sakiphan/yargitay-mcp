// SPDX-License-Identifier: AGPL-3.0-only
package server

import (
	"context"
	"crypto/rand"
	_ "embed"
	"encoding/base64"
	"html/template"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/sakiphan/yargitay-mcp/internal/courts"
	"github.com/sakiphan/yargitay-mcp/internal/yargitay"
)

//go:embed templates/reader.html
var readerHTML string

var readerTemplate = template.Must(template.New("reader").Parse(strings.TrimSpace(readerHTML)))

type Reader struct {
	base string
	http *http.Server
	done chan error
}

func (r *Reader) URL(id string) *string { u := r.base + "/" + id; return &u }
func (r *Reader) CourtURL(source, id string) *string {
	if !courts.ValidID(source, id) {
		return nil
	}
	u := r.base + "/" + source + "/" + id
	return &u
}
func (r *Reader) Close() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if r.http.Shutdown(ctx) != nil {
		_ = r.http.Close()
	}
	<-r.done
}
func StartReader(source Source, maxQueue int, additional ...CourtSource) (*Reader, error) {
	providers := map[string]CourtSource{}
	for _, c := range additional {
		providers[c.Source()] = c
	}
	l, e := net.Listen("tcp4", "127.0.0.1:0")
	if e != nil {
		return nil, e
	}
	token := make([]byte, 24)
	if _, e = rand.Read(token); e != nil {
		l.Close()
		return nil, e
	}
	prefix := "/" + base64.RawURLEncoding.EncodeToString(token) + "/decisions/"
	reader := &Reader{base: "http://" + l.Addr().String() + strings.TrimSuffix(prefix, "/"), done: make(chan error, 1)}
	sem := make(chan struct{}, maxQueue)
	h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Yalnızca GET.", 405)
			return
		}
		if !strings.HasPrefix(r.URL.Path, prefix) {
			http.NotFound(w, r)
			return
		}
		select {
		case sem <- struct{}{}:
			defer func() { <-sem }()
		default:
			http.Error(w, "Kuyruk dolu.", 503)
			return
		}
		id := strings.TrimPrefix(r.URL.Path, prefix)
		var d yargitay.Document
		var e error
		court := "Yargıtay"
		if key, docID, ok := strings.Cut(id, "/"); ok {
			provider := providers[key]
			if provider == nil || !courts.ValidID(key, docID) {
				http.NotFound(w, r)
				return
			}
			d, e = provider.GetDecision(r.Context(), docID)
			court = map[string]string{"aym": "AYM", "danistay": "Danıştay", "aihm": "AİHM / HUDOC"}[key]
		} else {
			request := yargitay.DocumentRequest{DocumentID: id}
			if request.Normalize() != nil {
				http.NotFound(w, r)
				return
			}
			d, e = source.GetDecision(r.Context(), id)
		}
		if e != nil {
			code := 503
			if e == yargitay.ErrNotFound {
				code = 404
			}
			http.Error(w, e.Error(), code)
			return
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_ = readerTemplate.Execute(w, struct {
			yargitay.Document
			Court string
		}{d, court})
	})
	reader.http = &http.Server{Handler: guarded(l.Addr().String(), h), ReadHeaderTimeout: 5 * time.Second, ReadTimeout: 10 * time.Second, WriteTimeout: 310 * time.Second, IdleTimeout: 30 * time.Second, MaxHeaderBytes: 16 << 10}
	go func() { reader.done <- reader.http.Serve(l) }()
	return reader, nil
}
