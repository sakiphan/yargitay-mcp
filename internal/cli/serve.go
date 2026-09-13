// SPDX-License-Identifier: AGPL-3.0-only
package cli

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/sakiphan/yargitay-mcp/internal/courts"
	"github.com/sakiphan/yargitay-mcp/internal/server"
	"github.com/sakiphan/yargitay-mcp/internal/yargitay"
)

func serve(args []string, errOut io.Writer, version string) error {
	cfg, e := yargitay.ConfigFromEnv()
	if e != nil {
		return e
	}
	f := flag.NewFlagSet("serve", flag.ContinueOnError)
	f.SetOutput(errOut)
	transport := f.String("transport", "stdio", "stdio veya streamable-http")
	address := f.String("listen", "127.0.0.1:8000", "yalnızca loopback HTTP adresi")
	viewer := f.Bool("viewer", cfg.Viewer, "yerel karar görüntüleyicisi")
	if e = f.Parse(args); e != nil {
		return e
	}
	if f.NArg() != 0 {
		return errors.New("beklenmeyen argüman")
	}
	if *transport != "stdio" && *transport != "streamable-http" {
		return errors.New("geçersiz transport")
	}
	client, e := yargitay.NewClient(cfg)
	if e != nil {
		return e
	}
	defer client.Close()
	var courtSources []server.CourtSource
	for _, name := range []string{"aym", "danistay", "aihm"} {
		c, err := courts.New(name, cfg)
		if err != nil {
			return err
		}
		defer c.Close()
		courtSources = append(courtSources, c)
	}
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	var view func(string) *string
	var courtView func(string, string) *string
	if *viewer {
		reader, e := server.StartReader(client, cfg.MaxQueue, courtSources...)
		if e != nil {
			return errors.New("yerel görüntüleyici başlatılamadı")
		}
		defer reader.Close()
		view = reader.URL
		courtView = reader.CourtURL
	}
	s := server.New(client, version, view)
	server.AddCourtTools(s, courtSources, courtView)
	if *transport == "stdio" {
		e = s.Run(ctx, &mcp.StdioTransport{})
		if ctx.Err() != nil {
			return nil
		}
		if e != nil {
			return errors.New("MCP stdio bağlantısı kapandı veya protokol hatası oluştu")
		}
		return nil
	}
	l, e := server.ListenLoopback(*address)
	if e != nil {
		return e
	}
	defer l.Close()
	h := &http.Server{Handler: server.HTTPHandler(s, l.Addr().String()), ReadHeaderTimeout: 5 * time.Second, ReadTimeout: 10 * time.Second, WriteTimeout: 310 * time.Second, IdleTimeout: 30 * time.Second, MaxHeaderBytes: 16 << 10, ErrorLog: log.New(io.Discard, "", 0)}
	done := make(chan error, 1)
	go func() { done <- h.Serve(l) }()
	fmt.Fprintln(errOut, "Yerel MCP:", "http://"+l.Addr().String()+"/mcp")
	select {
	case e := <-done:
		if errors.Is(e, http.ErrServerClosed) {
			return nil
		}
		return errors.New("yerel HTTP sunucusu durdu")
	case <-ctx.Done():
		shutdown, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if h.Shutdown(shutdown) != nil {
			_ = h.Close()
		}
		return nil
	}
}
