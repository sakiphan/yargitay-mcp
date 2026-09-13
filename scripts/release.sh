#!/bin/sh
# SPDX-License-Identifier: AGPL-3.0-only
set -eu
version=${1:-v0.3.1}
printf '%s' "$version" | LC_ALL=C grep -Eq '^v[0-9]+\.[0-9]+\.[0-9]+$' || exit 2
commit=unknown
if resolved_commit=$(git rev-parse --verify HEAD 2>/dev/null); then
  commit=$resolved_commit
fi
go mod download all
mkdir -p dist
for os in darwin linux windows; do
  for arch in amd64 arm64; do
    suffix=""; [ "$os" != windows ] || suffix=.exe
    CGO_ENABLED=0 GOOS="$os" GOARCH="$arch" go build -trimpath -ldflags "-s -w -X main.version=${version#v} -X main.commit=$commit" -o "dist/yargitay-mcp_${os}_${arch}${suffix}" ./cmd/yargitay-mcp
  done
done
cp LICENSE dist/LICENSE
cp NOTICE dist/NOTICE
go run ./scripts/notices dist/THIRD_PARTY_NOTICES.txt
printf '{"version":"%s","commit":"%s"}\n' "${version#v}" "$commit" > dist/release.json
(cd dist && if command -v sha256sum >/dev/null; then sha256sum yargitay-mcp_*; else shasum -a 256 yargitay-mcp_*; fi) > dist/checksums.txt
