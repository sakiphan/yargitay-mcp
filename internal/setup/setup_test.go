// SPDX-License-Identifier: AGPL-3.0-only
package setup

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestInstallIdempotentBackupAndUninstall(t *testing.T) {
	for _, client := range []string{"codex", "claude", "gemini"} {
		t.Run(client, func(t *testing.T) {
			home := t.TempDir()
			exe, e := os.Executable()
			if e != nil {
				t.Fatal(e)
			}
			path, _ := ConfigPath(client, home)
			if e = os.MkdirAll(filepath.Dir(path), 0700); e != nil {
				t.Fatal(e)
			}
			original := []byte(`{"other":{"keep":"yes"},"mcpServers":{"existing":{"command":"keep"}}}`)
			if client == "codex" {
				original = []byte("# Keep this comment\nmodel = \"example\"\n[mcp_servers.existing]\ncommand = \"keep\"\n")
			}
			if e = os.WriteFile(path, original, 0600); e != nil {
				t.Fatal(e)
			}
			o := Options{Client: client, Home: home, Executable: exe}
			result, e := Apply(o)
			if e != nil || !result.Changed || result.Backup == "" {
				t.Fatal(result, e)
			}
			backup, _ := os.ReadFile(result.Backup)
			if !bytes.Equal(backup, original) {
				t.Fatal("backup mismatch")
			}
			installed, _ := os.ReadFile(path)
			if !strings.Contains(string(installed), "keep") {
				t.Fatal("lost unrelated setting")
			}
			again, e := Apply(o)
			if e != nil || again.Changed {
				t.Fatal("not idempotent", again, e)
			}
			o.Remove = true
			removed, e := Apply(o)
			if e != nil || !removed.Changed {
				t.Fatal(removed, e)
			}
			last, _ := os.ReadFile(path)
			if bytes.Contains(last, []byte("[mcp_servers.yargitay]")) {
				t.Fatal("still installed")
			}
			if client != "codex" {
				var root map[string]any
				_ = json.Unmarshal(last, &root)
				if _, ok := root["mcpServers"].(map[string]any)[name]; ok {
					t.Fatal("still installed")
				}
			}
		})
	}
}
func TestConflictsAndMalformedConfigsPreserved(t *testing.T) {
	for _, tc := range []struct{ client, body string }{{"codex", "[mcp_servers.yargitay]\ncommand=\"someone-else\""}, {"claude", `{"mcpServers":{"yargitay":{"command":"someone-else"}}}`}, {"gemini", `// JSONC\n{}`}, {"codex", "not valid toml"}, {"claude", `{"mcpServers":null}`}} {
		t.Run(tc.client+tc.body, func(t *testing.T) {
			home := t.TempDir()
			exe, _ := os.Executable()
			p, _ := ConfigPath(tc.client, home)
			_ = os.MkdirAll(filepath.Dir(p), 0700)
			_ = os.WriteFile(p, []byte(tc.body), 0600)
			if _, e := Apply(Options{Client: tc.client, Home: home, Executable: exe}); e == nil {
				t.Fatal("accepted conflict")
			}
			after, _ := os.ReadFile(p)
			if string(after) != tc.body {
				t.Fatal("changed conflicting file")
			}
		})
	}
}
func TestDryRunDoesNotCreateConfig(t *testing.T) {
	home := t.TempDir()
	exe, _ := os.Executable()
	r, e := Apply(Options{Client: "codex", Home: home, Executable: exe, DryRun: true})
	if e != nil || !r.Changed {
		t.Fatal(r, e)
	}
	if _, e = os.Stat(r.Path); !os.IsNotExist(e) {
		t.Fatal("dry run wrote file")
	}
}
