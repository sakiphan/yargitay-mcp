// SPDX-License-Identifier: AGPL-3.0-only
// Explicit local client configuration only; never called while serving MCP.
package setup

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"strings"
	"time"

	"github.com/pelletier/go-toml/v2"
)

const name = "yargitay"
const begin = "# BEGIN yargitay-mcp managed configuration"
const end = "# END yargitay-mcp managed configuration"

type Options struct {
	Client, Home, Executable string
	Remove, DryRun           bool
}
type Result struct {
	Path, Backup string
	Changed      bool
}

func ConfigPath(client, home string) (string, error) {
	switch client {
	case "codex":
		return filepath.Join(home, ".codex", "config.toml"), nil
	case "claude":
		return filepath.Join(home, ".claude.json"), nil
	case "gemini":
		return filepath.Join(home, ".gemini", "settings.json"), nil
	case "claude-desktop":
		switch runtime.GOOS {
		case "darwin":
			return filepath.Join(home, "Library", "Application Support", "Claude", "claude_desktop_config.json"), nil
		case "windows":
			return filepath.Join(home, "AppData", "Roaming", "Claude", "claude_desktop_config.json"), nil
		}
		return "", errors.New("Claude Desktop için yalnızca macOS/Windows destekleniyor")
	}
	return "", errors.New("istemci codex, claude, claude-desktop veya gemini olmalı")
}
func Apply(o Options) (Result, error) {
	p, e := ConfigPath(o.Client, o.Home)
	result := Result{Path: p}
	if e != nil {
		return result, e
	}
	exe, e := filepath.Abs(o.Executable)
	if e != nil {
		return result, e
	}
	if stat, e := os.Stat(exe); e != nil || stat.IsDir() {
		return result, errors.New("kurulacak çalıştırılabilir dosya bulunamadı")
	}
	// Reject a symlink at the config path; avoid replacing unrelated linked configuration.
	if info, e := os.Lstat(p); e == nil && info.Mode()&os.ModeSymlink != 0 {
		return result, errors.New("yapılandırma sembolik bağlantı; dosya değiştirilmedi")
	}
	old, e := os.ReadFile(p)
	exists := e == nil
	if e != nil && !os.IsNotExist(e) {
		return result, e
	}
	var next []byte
	if o.Client == "codex" {
		next, e = mergeTOML(old, exe, o.Remove)
	} else {
		next, e = mergeJSON(old, exe, o.Client, o.Remove)
	}
	if e != nil {
		return result, e
	}
	if bytes.Equal(old, next) {
		return result, nil
	}
	result.Changed = true
	if o.DryRun {
		return result, nil
	}
	if e = os.MkdirAll(filepath.Dir(p), 0700); e != nil {
		return result, e
	}
	lock, e := os.OpenFile(p+".yargitay-lock", os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600)
	if e != nil {
		return result, errors.New("yapılandırma kilitli; eşzamanlı kurulumu tamamlayın")
	}
	lock.Close()
	defer os.Remove(p + ".yargitay-lock")
	// Fail if the client changed its config while we were preparing this edit.
	now, re := os.ReadFile(p)
	if (exists && re != nil) || (!exists && !os.IsNotExist(re)) || !bytes.Equal(old, now) {
		return result, errors.New("yapılandırma eşzamanlı değişti; yeniden deneyin")
	}
	if exists {
		result.Backup = p + ".yargitay-backup-" + time.Now().UTC().Format("20060102T150405.000000000")
		b, e := os.OpenFile(result.Backup, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600)
		if e != nil {
			return result, e
		}
		_, e = b.Write(old)
		ce := b.Close()
		if e != nil {
			return result, e
		}
		if ce != nil {
			return result, ce
		}
	}
	temp, e := os.CreateTemp(filepath.Dir(p), ".yargitay-config-*")
	if e != nil {
		return result, e
	}
	defer os.Remove(temp.Name())
	if _, e = temp.Write(next); e != nil {
		temp.Close()
		return result, e
	}
	if e = temp.Sync(); e != nil {
		temp.Close()
		return result, e
	}
	if e = temp.Close(); e != nil {
		return result, e
	}
	if e = os.Rename(temp.Name(), p); e != nil {
		return result, fmt.Errorf("ayar dosyası değiştirilemedi; yedek korundu: %w", e)
	}
	return result, nil
}
func configEntry(exe, client string) map[string]any {
	m := map[string]any{"command": exe, "args": []any{"serve"}, "env": map[string]any{"YARGITAY_VIEWER_ENABLED": "true"}}
	if client == "claude" {
		m["type"] = "stdio"
	}
	if client == "gemini" {
		m["timeout"] = float64(70000)
		m["trust"] = false
	}
	return m
}
func mergeJSON(old []byte, exe, client string, remove bool) ([]byte, error) {
	root := map[string]any{}
	if len(bytes.TrimSpace(old)) > 0 {
		if json.Unmarshal(old, &root) != nil || root == nil {
			return nil, errors.New("ayar dosyası geçerli JSON değil (JSONC desteklenmiyor); dosya değiştirilmedi")
		}
	}
	servers := map[string]any{}
	if v, ok := root["mcpServers"]; ok {
		var valid bool
		servers, valid = v.(map[string]any)
		if !valid {
			return nil, errors.New("mcpServers biçimi geçersiz; dosya değiştirilmedi")
		}
	}
	want := configEntry(exe, client)
	existing, found := servers[name]
	if remove {
		if !found {
			return old, nil
		}
		if !reflect.DeepEqual(existing, want) {
			return nil, errors.New("yargitay kaydı bu kurulumla eşleşmiyor; mevcut kayıt korunuyor")
		}
		delete(servers, name)
	} else {
		if found {
			if reflect.DeepEqual(existing, want) {
				return old, nil
			}
			return nil, errors.New("yargitay adlı farklı kayıt zaten var; mevcut kayıt korunuyor")
		}
		servers[name] = want
	}
	root["mcpServers"] = servers
	b, e := json.MarshalIndent(root, "", "  ")
	return append(b, '\n'), e
}
func mergeTOML(old []byte, exe string, remove bool) ([]byte, error) {
	root := map[string]any{}
	if len(bytes.TrimSpace(old)) > 0 {
		if toml.Unmarshal(old, &root) != nil {
			return nil, errors.New("geçersiz TOML; dosya değiştirilmedi")
		}
	}
	// JSON quoted paths are also valid TOML basic strings for normal filesystem paths.
	quoted, _ := json.Marshal(exe)
	block := begin + "\n[mcp_servers.yargitay]\ncommand = " + string(quoted) + "\nargs = [\"serve\"]\nstartup_timeout_sec = 20\ntool_timeout_sec = 70\n[mcp_servers.yargitay.env]\nYARGITAY_VIEWER_ENABLED = \"true\"\n" + end + "\n"
	text := string(old)
	start := strings.Index(text, begin)
	if start >= 0 {
		if strings.Count(text, begin) != 1 || strings.Count(text, end) != 1 {
			return nil, errors.New("kurulum işaretleri tutarsız; dosya korunuyor")
		}
		finish := strings.Index(text[start:], end)
		if finish < 0 {
			return nil, errors.New("kurulum işareti eksik")
		}
		finish += start + len(end)
		if finish < len(text) && text[finish] == '\n' {
			finish++
		}
		if text[start:finish] != block {
			return nil, errors.New("yargitay ayarları değiştirilmiş veya farklı dosyaya bağlı; mevcut kayıt korunuyor")
		}
		if !remove {
			return old, nil
		}
		return []byte(text[:start] + text[finish:]), nil
	}
	if servers, ok := root["mcp_servers"].(map[string]any); ok {
		if _, exists := servers[name]; exists {
			return nil, errors.New("yargitay adlı yönetilmeyen kayıt zaten var; mevcut kayıt korunuyor")
		}
	}
	if remove {
		return old, nil
	}
	if len(text) > 0 && !strings.HasSuffix(text, "\n") {
		text += "\n"
	}
	next := []byte(text + "\n" + block)
	var verify map[string]any
	if toml.Unmarshal(next, &verify) != nil {
		return nil, errors.New("yeni TOML doğrulanamadı; dosya korunuyor")
	}
	return next, nil
}
