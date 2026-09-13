// SPDX-License-Identifier: AGPL-3.0-only
// Generate the dependency-license bundle locally; no upstream HTTP requests.
package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
func run() error {
	if len(os.Args) != 2 {
		return fmt.Errorf("usage: go run ./scripts/notices OUTPUT")
	}
	goCommand := filepath.Join(runtime.GOROOT(), "bin", "go")
	if runtime.GOOS == "windows" {
		goCommand += ".exe"
	}
	out, e := exec.Command(goCommand, "list", "-m", "-json", "all").Output()
	if e != nil {
		return fmt.Errorf("go list failed: %w", e)
	}
	var b bytes.Buffer
	b.WriteString("Yargıtay MCP — bundled third-party notices\n\n")
	license, e := os.ReadFile(filepath.Join(runtime.GOROOT(), "LICENSE"))
	if e != nil {
		return e
	}
	b.WriteString("Go runtime\n")
	b.Write(license)
	b.WriteString("\n\n")
	decoder := json.NewDecoder(bytes.NewReader(out))
	for {
		var m struct {
			Path, Version, Dir string
			Main               bool
		}
		e := decoder.Decode(&m)
		if e == io.EOF {
			break
		}
		if e != nil {
			return e
		}
		if m.Main {
			continue
		}
		if m.Dir == "" {
			return fmt.Errorf("module source missing: %s; run go mod download", m.Path)
		}
		entries, e := os.ReadDir(m.Dir)
		if e != nil {
			return e
		}
		found := false
		for _, entry := range entries {
			n := strings.ToUpper(entry.Name())
			if entry.IsDir() || !(strings.HasPrefix(n, "LICENSE") || strings.HasPrefix(n, "COPYING") || n == "NOTICE" || n == "COPYRIGHT") {
				continue
			}
			data, e := os.ReadFile(filepath.Join(m.Dir, entry.Name()))
			if e != nil {
				return e
			}
			fmt.Fprintf(&b, "--- %s %s · %s ---\n", m.Path, m.Version, entry.Name())
			b.Write(data)
			b.WriteString("\n\n")
			found = true
		}
		if !found {
			return fmt.Errorf("license not found for %s", m.Path)
		}
	}
	return os.WriteFile(os.Args[1], b.Bytes(), 0644)
}
