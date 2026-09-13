# SPDX-License-Identifier: AGPL-3.0-only
.PHONY: help fmt fmt-check vet test test-npm build check

help:
	@echo "make check     Format, vet, offline Go/race and npm checks"
	@echo "make build     Build the local development binary in dist/"
	@echo "make fmt       Format Go source files"

fmt:
	gofmt -w cmd internal scripts

fmt-check:
	@test -z "$$(gofmt -l cmd internal scripts)" || { echo "Run make fmt"; exit 1; }

vet:
	go vet ./...

test:
	GOPROXY=off go test -race ./...

test-npm:
	npm run check
	npm test

build:
	go build -trimpath -o dist/yargitay-mcp ./cmd/yargitay-mcp

check: fmt-check vet test test-npm
