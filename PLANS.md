# Go migration

Scope: local Go MCP, Yargıtay plus AYM, Danıştay and AİHM/HUDOC; nine tools, no BAM.

1. Port bounded search/document contracts, embedded 51-unit catalog, safe HTML parser, Unicode chunks.
2. Add one process-wide upstream queue, retry cooldown, bounded in-memory TTL/LRU and request coalescing.
3. Expose three typed tools through the official Go MCP SDK, stdio by default; optional loopback HTTP and reader.
4. Add explicit client install/uninstall and an offline doctor. Never change personal settings during development.
5. Prepare verified release downloads for macOS/Linux/Windows; release publication remains a maintainer action.
6. Run synthetic offline tests, race detector, vet, cross builds and real stdio smoke tests.
7. Document AGPL scope, provenance, upstream limitations and what was actually verified.

The original Python project remains untouched. Go live behavior will not be claimed verified without an explicit live run.

Implemented: all seven phases. Nine tools, embedded catalog, loopback reader/HTTP,
bounded upstream client, explicit four-client setup, doctor, release scripts and
AGPL/third-party notices are present. Offline/race tests and six cross builds pass.
AYM and AİHM are experimental; bounded live observations and query-evaluation
limits are documented in docs/VERIFICATION.md and docs/QUERY_EVAL_REPORT.md.
Package publication, full client UI checks and Windows runtime verification
remain incomplete; source publication is separate from GitHub/npm releases.
