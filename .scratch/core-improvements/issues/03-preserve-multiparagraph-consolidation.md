# 03: Preserve multi-paragraph bodies during memory consolidation

**What to build:** During memory consolidation, empty lines must be treated as paragraph separators instead of an immediate termination signal, preserving multiple paragraphs, troubleshooting sections, and code blocks within the size budget.

**Blocked by:** None (can start immediately)

**Status:** ready-for-agent

- [x] Empty lines between paragraphs in `internal/consolidate` do not prematurely break body extraction
- [x] Multi-paragraph notes retain all paragraphs up to maximum character and line budgets
- [x] `go test ./internal/consolidate` passes
