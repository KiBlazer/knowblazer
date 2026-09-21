# 02: Exclude quarantine folder from Git sync security scans

**What to build:** Security scans during sync operations (`knowblazer sync commit` and `push`) must ignore the `quarantine` directory, preventing isolated sensitive files from deadlocking repository sync.

**Blocked by:** None (can start immediately)

**Status:** ready-for-agent

- [x] Directory scanning in `internal/scan` ignores `quarantine` directory
- [x] Having high-risk findings in `quarantine/` does not block `sync.Commit`
- [x] `go test ./internal/scan ./internal/sync` passes
