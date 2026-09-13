# Third-party notices

The original Python project's Apache-2.0 license is retained in Apache-2.0.txt.
The Go implementation as a whole is offered under AGPL-3.0-only, preserving
applicable original notices. No third-party license is replaced by this notice.

Dependencies and their source licenses are identified in go.mod/go.sum and in
their respective source distributions. Before publishing a release, generate the
dependency notice bundle with `go run ./scripts/notices dist/THIRD_PARTY_NOTICES.txt`
after `go mod download all`; attach that file alongside the binary and LICENSE.
