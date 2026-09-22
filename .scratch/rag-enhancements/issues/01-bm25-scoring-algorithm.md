# 01: Pure Go BM25 Scoring and Document Length Normalization

**What to build:** Replace naive `strings.Count` term scoring in `internal/recall` and `internal/index` with an in-memory Okapi BM25 ranking algorithm ($k_1=1.2, b=0.75$) using document length normalization and inverse document frequency (IDF).

**Blocked by:** None (can start immediately)

**Status:** resolved

- [x] BM25 score calculation function implemented with zero external dependencies
- [x] Document length normalization penalizes long diluted documents and favors compact, focused notes
- [x] Integrated into `recall.relevantExperience` and `index.Search`
- [x] Tests in `internal/recall` and `internal/index` demonstrate short relevant notes outranking long diluted notes
- [x] `go test ./internal/recall ./internal/index` passes

