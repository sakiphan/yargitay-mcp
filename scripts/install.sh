#!/bin/sh
# SPDX-License-Identifier: AGPL-3.0-only
set -eu

client=""
version="v0.3.1"
while [ "$#" -gt 0 ]; do
  case "$1" in
    --client) [ "$#" -ge 2 ] || exit 2; client=$2; shift 2 ;;
    --version) [ "$#" -ge 2 ] || exit 2; version=$2; shift 2 ;;
    *) echo "Kullanım: install.sh --client codex|claude|claude-desktop|gemini [--version v0.3.1]" >&2; exit 2 ;;
  esac
done
case "$client" in codex|claude|claude-desktop|gemini) ;; *) echo "--client seçilmeli." >&2; exit 2 ;; esac
printf '%s' "$version" | LC_ALL=C grep -Eq '^v[0-9]+\.[0-9]+\.[0-9]+$' || { echo "Geçersiz sürüm." >&2; exit 2; }
case "$(uname -s)" in Darwin) platform=darwin ;; Linux) platform=linux ;; *) echo "Windows için install.ps1 kullanın." >&2; exit 2 ;; esac
case "$(uname -m)" in arm64|aarch64) arch=arm64 ;; x86_64|amd64) arch=amd64 ;; *) echo "Desteklenmeyen işlemci." >&2; exit 2 ;; esac
command -v curl >/dev/null || { echo "curl gerekli." >&2; exit 2; }
if command -v sha256sum >/dev/null; then hash_cmd=sha256sum; elif command -v shasum >/dev/null; then hash_cmd=shasum; else echo "SHA-256 doğrulama aracı gerekli." >&2; exit 2; fi

asset="yargitay-mcp_${platform}_${arch}"
base="https://github.com/sakiphan/yargitay-mcp/releases/download/$version"
work=$(mktemp -d)
cleanup() { rm -f "$work/binary" "$work/checksums.txt"; rmdir "$work" 2>/dev/null || true; }
trap cleanup EXIT HUP INT TERM
curl --proto '=https' --tlsv1.2 --fail --silent --show-error --location "$base/$asset" -o "$work/binary"
curl --proto '=https' --tlsv1.2 --fail --silent --show-error --location "$base/checksums.txt" -o "$work/checksums.txt"
expected=$(awk -v file="$asset" '$2 == file {print $1}' "$work/checksums.txt")
printf '%s' "$expected" | LC_ALL=C grep -Eq '^[a-fA-F0-9]{64}$' || { echo "Geçerli checksum bulunamadı." >&2; exit 1; }
if [ "$hash_cmd" = sha256sum ]; then actual=$(sha256sum "$work/binary" | awk '{print $1}'); else actual=$(shasum -a 256 "$work/binary" | awk '{print $1}'); fi
[ "$actual" = "$expected" ] || { echo "SHA-256 uyuşmuyor; kurulum durduruldu." >&2; exit 1; }
dest=${YARGITAY_INSTALL_DIR:-"$HOME/.local/bin"}
mkdir -p "$dest"
target="$dest/yargitay-mcp"
[ ! -L "$target" ] || { echo "Hedef sembolik bağlantı; kurulum durduruldu." >&2; exit 1; }
[ ! -L "$dest/yargitay-mcp.previous" ] || { echo "Yedek hedefi sembolik bağlantı; kurulum durduruldu." >&2; exit 1; }
if [ -f "$target" ]; then cp -p "$target" "$dest/yargitay-mcp.previous"; fi
chmod 755 "$work/binary"
# Same-directory rename avoids partially replacing an executable in use.
stage=$(mktemp "$dest/.yargitay-mcp.XXXXXX")
if ! cp "$work/binary" "$stage" || ! chmod 755 "$stage" || ! mv -f "$stage" "$target"; then rm -f "$stage"; exit 1; fi
"$target" install --client "$client"
"$target" version
echo "Kuruldu: $target — istemcinizi yeniden başlatın."
