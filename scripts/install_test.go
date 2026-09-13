// SPDX-License-Identifier: AGPL-3.0-only
package scripts

import (
	"crypto/sha256"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestShellInstallerVerifiesDownloads(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("POSIX installer; Windows uses install.ps1")
	}
	for _, scenario := range []string{"success", "bad-hash", "download-failure"} {
		t.Run(scenario, func(t *testing.T) {
			dir := t.TempDir()
			bin := filepath.Join(dir, "tools")
			dest := filepath.Join(dir, "install")
			_ = os.MkdirAll(bin, 0700)
			payload := []byte("#!/bin/sh\nprintf 'mock-binary %s\\n' \"$*\"\n")
			payloadPath := filepath.Join(dir, "payload")
			_ = os.WriteFile(payloadPath, payload, 0600)
			digest := fmt.Sprintf("%x", sha256.Sum256(payload))
			if scenario == "bad-hash" {
				digest = strings.Repeat("0", 64)
			}
			checks := filepath.Join(dir, "checksums")
			_ = os.WriteFile(checks, []byte(digest+"  yargitay-mcp_linux_amd64\n"), 0600)
			curl := `#!/bin/sh
set -eu
url=''
out=''
while [ "$#" -gt 0 ]; do
  case "$1" in -o) out=$2; shift 2;; https:*) url=$1; shift;; *) shift;; esac
done
[ "$YARGITAY_TEST_SCENARIO" != download-failure ] || exit 22
case "$url" in
  https://github.com/sakiphan/yargitay-mcp/releases/download/v0.1.0/checksums.txt) cp "$YARGITAY_TEST_CHECKSUMS" "$out" ;;
  https://github.com/sakiphan/yargitay-mcp/releases/download/v0.1.0/yargitay-mcp_linux_amd64) cp "$YARGITAY_TEST_PAYLOAD" "$out" ;;
  *) exit 23 ;;
esac
`
			_ = os.WriteFile(filepath.Join(bin, "curl"), []byte(curl), 0700)
			_ = os.WriteFile(filepath.Join(bin, "uname"), []byte("#!/bin/sh\nif [ \"$1\" = -s ]; then echo Linux; else echo x86_64; fi\n"), 0700)
			cmd := exec.Command("sh", "install.sh", "--client", "codex", "--version", "v0.1.0")
			cmd.Env = append(os.Environ(), "PATH="+bin+string(os.PathListSeparator)+os.Getenv("PATH"), "YARGITAY_INSTALL_DIR="+dest, "YARGITAY_TEST_SCENARIO="+scenario, "YARGITAY_TEST_PAYLOAD="+payloadPath, "YARGITAY_TEST_CHECKSUMS="+checks)
			output, e := cmd.CombinedOutput()
			target := filepath.Join(dest, "yargitay-mcp")
			if scenario == "success" {
				if e != nil || !strings.Contains(string(output), "mock-binary install --client codex") {
					t.Fatal(string(output), e)
				}
				b, e := os.ReadFile(target)
				if e != nil || string(b) != string(payload) {
					t.Fatal("wrong installed payload", e)
				}
			} else {
				if e == nil {
					t.Fatal("unsafe install succeeded")
				}
				if _, e = os.Stat(target); !os.IsNotExist(e) {
					t.Fatal("failed verification wrote executable")
				}
			}
		})
	}
}
