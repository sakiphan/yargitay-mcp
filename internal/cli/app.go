// SPDX-License-Identifier: AGPL-3.0-only
package cli

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"runtime"
	"time"

	"github.com/sakiphan/yargitay-mcp/internal/courts"
	"github.com/sakiphan/yargitay-mcp/internal/setup"
	"github.com/sakiphan/yargitay-mcp/internal/yargitay"
)

// BuildInfo carries linker-provided metadata without mutable package globals.
type BuildInfo struct {
	Version string
	Commit  string
}

// Run dispatches a CLI command. It never exits the calling process.
func Run(args []string, out, errOut io.Writer, build BuildInfo) error {
	command := "serve"
	if len(args) > 0 && len(args[0]) > 0 && args[0][0] != '-' {
		command = args[0]
		args = args[1:]
	}
	switch command {
	case "version":
		fmt.Fprintf(out, "yargitay-mcp %s (%s) %s/%s\n", build.Version, build.Commit, runtime.GOOS, runtime.GOARCH)
		return nil
	case "help":
		fmt.Fprintln(out, "yargitay-mcp [serve | install --client codex|claude|claude-desktop|gemini | uninstall --client ... | doctor | probe --live | version]\nserve: --transport stdio|streamable-http --listen 127.0.0.1:8000 --viewer\ninstall/uninstall: --dry-run önizleme; mevcut farklı kayıtlar korunur.")
		return nil
	case "install", "uninstall":
		f := flag.NewFlagSet(command, flag.ContinueOnError)
		f.SetOutput(errOut)
		client := f.String("client", "", "codex, claude, claude-desktop veya gemini")
		dry := f.Bool("dry-run", false, "dosya değiştirmeden önizle")
		if e := f.Parse(args); e != nil {
			return e
		}
		if f.NArg() != 0 {
			return errors.New("beklenmeyen argüman")
		}
		home, e := os.UserHomeDir()
		if e != nil {
			return e
		}
		exe, e := os.Executable()
		if e != nil {
			return e
		}
		result, e := setup.Apply(setup.Options{Client: *client, Home: home, Executable: exe, Remove: command == "uninstall", DryRun: *dry})
		if e != nil {
			return e
		}
		fmt.Fprintf(out, "%s: %s (değişiklik: %t, önizleme: %t)\n", command, result.Path, result.Changed, *dry)
		if result.Backup != "" {
			fmt.Fprintln(out, "Yedek:", result.Backup)
		}
		if !*dry {
			fmt.Fprintln(out, "İstemciyi yeniden başlatın. Binary dosyası silinmez.")
		}
		return nil
	case "doctor":
		cfg, e := yargitay.ConfigFromEnv()
		if e != nil {
			return e
		}
		units, e := yargitay.Units("all")
		if e != nil {
			return e
		}
		fmt.Fprintf(out, "yargitay-mcp %s · %s/%s\nKatalog: %d birim\nTLS: doğrulama açık; sabit resmî kaynak\nHız sınırı: %s; kuyruk: %d; süre sınırı: %s\nUpstream kontrol edilmedi; bu komut ağ isteği yapmaz.\n", build.Version, runtime.GOOS, runtime.GOARCH, len(units.Units), cfg.Interval, cfg.MaxQueue, cfg.Deadline)
		home, e := os.UserHomeDir()
		if e != nil {
			return e
		}
		for _, client := range []string{"codex", "claude", "claude-desktop", "gemini"} {
			p, e := setup.ConfigPath(client, home)
			if e != nil {
				continue
			}
			_, e = os.Stat(p)
			fmt.Fprintf(out, "%s ayar dosyası mevcut: %t\n", client, e == nil)
		}
		return nil
	case "probe":
		f := flag.NewFlagSet(command, flag.ContinueOnError)
		f.SetOutput(errOut)
		live := f.Bool("live", false, "bir arama ve en fazla bir belge isteğine açık izin")
		source := f.String("source", "yargitay", "yargitay, aym, danistay veya aihm; tek kaynak")
		kind := f.String("decision-type", "", "AYM: individual_application/norm_review; AİHM: judgments/decisions")
		if e := f.Parse(args); e != nil {
			return e
		}
		if !*live || os.Getenv("CI") != "" {
			return errors.New("canlı probe için --live gerekir; CI ortamında yasaktır")
		}
		if f.NArg() != 0 {
			return errors.New("beklenmeyen argüman")
		}
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
		defer cancel()
		var report map[string]any
		var e error
		if *source == "yargitay" {
			if *kind != "" {
				return errors.New("Yargıtay probe karar türü almaz")
			}
			report, e = yargitay.Probe(ctx)
		} else {
			report, e = courts.Probe(ctx, *source, *kind)
		}
		_ = json.NewEncoder(out).Encode(report)
		return e
	case "serve":
		return serve(args, errOut, build.Version)
	default:
		return fmt.Errorf("bilinmeyen komut %q; help kullanın", command)
	}
}
