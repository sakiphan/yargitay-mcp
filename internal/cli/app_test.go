// SPDX-License-Identifier: AGPL-3.0-only
package cli

import (
	"bytes"
	"errors"
	"flag"
	"strings"
	"testing"
)

func TestOfflineCLI(t *testing.T) {
	for _, command := range []string{"help", "version", "doctor"} {
		var out, errs bytes.Buffer
		if e := Run([]string{command}, &out, &errs, BuildInfo{Version: "test", Commit: "test"}); e != nil || out.Len() == 0 {
			t.Fatal(command, e)
		}
	}
	var out bytes.Buffer
	if e := Run([]string{"probe"}, &out, &out, BuildInfo{}); e == nil {
		t.Fatal("live probe without opt-in")
	}
}

func TestBuildInfoIsPerInvocation(t *testing.T) {
	for _, version := range []string{"first", "second"} {
		var out, errs bytes.Buffer
		if err := Run([]string{"version"}, &out, &errs, BuildInfo{Version: version, Commit: "revision"}); err != nil {
			t.Fatal(err)
		}
		if !strings.Contains(out.String(), "yargitay-mcp "+version+" (revision)") || errs.Len() != 0 {
			t.Fatal("build metadata or output stream changed")
		}
	}
}

func TestInvalidCommandsStayOffline(t *testing.T) {
	t.Setenv("CI", "true")
	for _, args := range [][]string{{"unknown"}, {"serve", "--transport", "invalid"}, {"serve", "extra"}, {"install", "extra"}, {"probe", "--live"}} {
		var out, errs bytes.Buffer
		if err := Run(args, &out, &errs, BuildInfo{}); err == nil || out.Len() != 0 {
			t.Fatalf("invalid command accepted or wrote to stdout: %v", args)
		}
	}
}

func TestFlagHelpDoesNotStartServer(t *testing.T) {
	var out, errs bytes.Buffer
	err := Run([]string{"serve", "--help"}, &out, &errs, BuildInfo{})
	if !errors.Is(err, flag.ErrHelp) || out.Len() != 0 || errs.Len() == 0 {
		t.Fatal("flag help contract changed", err)
	}
}
