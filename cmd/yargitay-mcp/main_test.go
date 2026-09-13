// SPDX-License-Identifier: AGPL-3.0-only
package main

import (
	"bytes"
	"context"
	"os"
	"os/exec"
	"strings"
	"testing"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func TestStdioProcess(t *testing.T) {
	if os.Getenv("YARGITAY_TEST_HELPER") == "1" {
		os.Args = []string{os.Args[0], "serve"}
		main()
		os.Exit(0)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	cmd := exec.Command(os.Args[0], "-test.run=^TestStdioProcess$")
	cmd.Env = append(os.Environ(), "YARGITAY_TEST_HELPER=1", "YARGITAY_VIEWER_ENABLED=false")
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	cs, e := mcp.NewClient(&mcp.Implementation{Name: "stdio-test", Version: "1"}, nil).Connect(ctx, &mcp.CommandTransport{Command: cmd}, nil)
	if e != nil {
		t.Fatal(e, stderr.String())
	}
	tools, e := cs.ListTools(ctx, nil)
	if e != nil || len(tools.Tools) != 9 {
		t.Fatal(tools, e)
	}
	r, e := cs.CallTool(ctx, &mcp.CallToolParams{Name: "list_yargitay_units", Arguments: map[string]any{}})
	if e != nil || r.IsError {
		t.Fatal(r, e)
	}
	if e = cs.Close(); e != nil {
		t.Fatal(e)
	}
	if strings.Contains(stderr.String(), "PRIVATE") {
		t.Fatal("private log")
	}
}
