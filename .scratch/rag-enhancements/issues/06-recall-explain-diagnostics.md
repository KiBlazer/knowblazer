# 06: Recall Explainability and Diagnostic Inspection

**What to build:** Add an `--explain` flag to `knowblazer recall` and support in `recall.Options` to output structured diagnostic information, detailing tokenized terms, candidate BM25 scores, threshold cutoffs, and reordering moves.

**Blocked by:** 01 (BM25 scoring algorithm), 02 (Relevance quality gate)

**Status:** resolved

- [x] `recall.Options` includes `Explain bool`
- [x] Structured diagnostic block rendered when `--explain` is provided
- [x] CLI command `knowblazer recall "<task>" --explain` wired up and tested
- [x] Tests verify explain output format and non-regression on standard recall
- [x] `go test ./internal/recall ./internal/cli` passes

