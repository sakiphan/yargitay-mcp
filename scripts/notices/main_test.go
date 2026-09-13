// SPDX-License-Identifier: AGPL-3.0-only
package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLicenseBundle(t *testing.T) {
	previous := os.Args
	t.Cleanup(func() { os.Args = previous })
	output := filepath.Join(t.TempDir(), "notices.txt")
	os.Args = []string{"notices", output}
	if err := run(); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(output)
	if err != nil {
		t.Fatal(err)
	}
	for _, required := range []string{"Go runtime", "github.com/modelcontextprotocol/go-sdk", "github.com/pelletier/go-toml/v2", "golang.org/x/net", "Copyright", "Redistribution"} {
		if !strings.Contains(string(data), required) {
			t.Errorf("missing dependency notice: %s", required)
		}
	}
}

func TestOutputPathRequired(t *testing.T) {
	previous := os.Args
	t.Cleanup(func() { os.Args = previous })
	os.Args = []string{"notices"}
	if run() == nil {
		t.Fatal("missing output path accepted")
	}
}
