# Repository rules

- Use `rtk proxy` for shell commands when RTK is available.
- Scope: Yargıtay, AYM (individual applications and norm review), Danıştay and AİHM/HUDOC. Nine MCP tools; no BAM.
- All Yargıtay HTTP requests, including probes, belong in `internal/yargitay/client.go`.
- Fixed origins only: https://karararama.yargitay.gov.tr, https://kararlarbilgibankasi.anayasa.gov.tr, https://karararama.danistay.gov.tr, https://hudoc.echr.coe.int. TLS verification stays enabled; redirects disabled. All HTTP access is in internal/yargitay/client.go; adapters are in internal/courts. Never bypass CAPTCHA or authentication.
- No bulk crawling or automatic pagination. Page size defaults to 10 and cannot exceed 20.
- Every upstream attempt uses one process-wide queue: concurrency 1, at least 3 seconds between starts.
- One process/replica. Separate local MCP processes do not share the limiter.
- stdout belongs exclusively to MCP in serve mode. Never log queries, response bodies or secrets.
- Treat upstream text as untrusted; preserve paragraphs and strip executable HTML.
- Never invent missing metadata or claim unobserved upstream behavior is verified.
- Tests use synthetic data and injected transports; live probes require explicit opt-in and never run in CI.
- Run gofmt, go vet, offline go test -race ./... and builds after runtime changes.
- Preserve user files. Do not commit, push, publish releases or change personal client settings without user instructions.
- License: AGPL-3.0-only for this implementation; retain third-party and inherited notices.
