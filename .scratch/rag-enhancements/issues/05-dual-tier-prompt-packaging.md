# 05: Dual-Tier Semantic Prompt Packaging

**What to build:** Restructure the Recall Pack Markdown output to explicitly divide recalled knowledge into Tier 1 (Mandatory Constraints & Developer Governance) and Tier 2 (Empirical Reference & Past Troubleshooting Cases), with clear instructional guidance for AI coding agents.

**Blocked by:** 03 (LongContextReorder), 04 (Hierarchical section extraction)

**Status:** resolved

- [x] Recall Pack structure updated with distinct Tier 1 and Tier 2 headings
- [x] Tier 1 clearly labels developer preferences, architectural principles, and project durable constraints
- [x] Tier 2 frames past experiences and daily notes as empirical reference patterns rather than rigid commands
- [x] Existing prompt budget truncation logic preserved and respects tier boundaries
- [x] Tests in `internal/recall` verify dual-tier layout and system headers
- [x] `go test ./internal/recall` passes

