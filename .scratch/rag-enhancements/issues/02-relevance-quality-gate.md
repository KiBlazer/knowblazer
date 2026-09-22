# 02: Relevance Quality Gate to Filter Low-Confidence Memories

**What to build:** Implement a confidence threshold cutoff in `recall.relevantExperience` so that candidate memories with negligible relevance are omitted from the Recall Pack, preventing low-signal noise from cluttering prompt context.

**Blocked by:** 01 (BM25 scoring algorithm)

**Status:** resolved

- [x] Configurable or baseline score threshold applied during candidate selection
- [x] Tasks matching only incidental/common words without strong signal return `_No matching memory found._`
- [x] Tests verify low-confidence candidate exclusion and graceful fallback output
- [x] `go test ./internal/recall` passes

