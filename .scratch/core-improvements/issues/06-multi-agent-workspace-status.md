# 06: Recognize multi-agent instructions in workspace status

**What to build:** In `knowblazer status` and `internal/mcp`, check for configuration markers across `CLAUDE.md`, `AGENTS.md`, and `GEMINI.md` to properly detect agent configuration in modern workspaces.

**Blocked by:** None (can start immediately)

**Status:** ready-for-agent

- [x] Status checks `AGENTS.md` and `GEMINI.md` in addition to `CLAUDE.md`
- [x] Returns `claude_configured: true` (or `agent_configured: true`) when `AGENTS.md` contains the setup block
- [x] `go test ./internal/mcp ./internal/doctor` passes
