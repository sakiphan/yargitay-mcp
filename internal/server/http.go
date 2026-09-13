// SPDX-License-Identifier: AGPL-3.0-only
package server

import (
	"fmt"
	"net"
	"net/http"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

const SourceRepository = "https://github.com/sakiphan/yargitay-mcp"

func guarded(host string, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "no-store")
		w.Header().Set("Referrer-Policy", "no-referrer")
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "DENY")
		w.Header().Set("Content-Security-Policy", "default-src 'none'; style-src 'unsafe-inline'; frame-ancestors 'none'")
		if r.Host != host {
			http.Error(w, "Host izinli değil.", http.StatusBadRequest)
			return
		}
		if r.Header.Get("Origin") != "" {
			http.Error(w, "Origin izinli değil.", http.StatusForbidden)
			return
		}
		r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
		next.ServeHTTP(w, r)
	})
}
func HTTPHandler(s *mcp.Server, host string) http.Handler {
	mux := http.NewServeMux()
	h := mcp.NewStreamableHTTPHandler(func(*http.Request) *mcp.Server { return s }, &mcp.StreamableHTTPOptions{Stateless: true, Logger: discardLogger()})
	mux.Handle("/mcp", h)
	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"status":"ok","upstream_checked":false}`))
	})
	return guarded(host, mux)
}

func ListenLoopback(address string) (net.Listener, error) {
	host, _, e := net.SplitHostPort(address)
	if e != nil {
		return nil, fmt.Errorf("adres 127.0.0.1:PORT biçiminde olmalı")
	}
	if host != "127.0.0.1" {
		return nil, fmt.Errorf("bu yerel sürüm yalnızca 127.0.0.1 üzerinde dinler")
	}
	return net.Listen("tcp4", address)
}
