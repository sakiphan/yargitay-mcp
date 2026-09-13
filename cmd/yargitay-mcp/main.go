// SPDX-License-Identifier: AGPL-3.0-only
package main

import (
	"errors"
	"flag"
	"fmt"
	"os"

	"github.com/sakiphan/yargitay-mcp/internal/cli"
)

// Kept in main so existing release -ldflags remain compatible.
var version = "0.3.2-dev"
var commit = "unknown"

func main() {
	if e := cli.Run(os.Args[1:], os.Stdout, os.Stderr, cli.BuildInfo{Version: version, Commit: commit}); e != nil && !errors.Is(e, flag.ErrHelp) {
		fmt.Fprintln(os.Stderr, e)
		os.Exit(1)
	}
}
