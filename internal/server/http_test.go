// SPDX-License-Identifier: AGPL-3.0-only
package server

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func TestHTTPHealthAndMCP(t *testing.T) {
	src := &fakeSource{}
	s := New(src, "test", nil)
	ts := httptest.NewUnstartedServer(nil)
	ts.Config.Handler = HTTPHandler(s, ts.Listener.Addr().String())
	ts.Start()
	defer ts.Close()
	resp, e := http.Get(ts.URL + "/healthz")
	if e != nil {
		t.Fatal(e)
	}
	resp.Body.Close()
	if resp.StatusCode != 200 || src.calls.Load() != 0 {
		t.Fatal("health generated upstream traffic")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	cs, e := mcp.NewClient(&mcp.Implementation{Name: "test", Version: "1"}, nil).Connect(ctx, &mcp.StreamableClientTransport{Endpoint: ts.URL + "/mcp"}, nil)
	if e != nil {
		t.Fatal(e)
	}
	defer cs.Close()
	tools, e := cs.ListTools(ctx, nil)
	if e != nil || len(tools.Tools) != 3 {
		t.Fatal(tools, e)
	}
	if _, e := ListenLoopback("0.0.0.0:8000"); e == nil {
		t.Fatal("public bind accepted")
	}
}
