# 05: Support CJK and unspaced search term segmentation in recall and indexing

**What to build:** In task recall scoring and local search index building, segment CJK character sequences into character unigrams and adjacent bigrams so unspaced Chinese queries match and score relevant notes.

**Blocked by:** None (can start immediately)

**Status:** ready-for-agent

- [x] CJK character sequences generate unigrams and bigrams in `internal/recall` and `internal/index`
- [x] Recall scoring correctly matches Chinese tasks against Chinese experience notes
- [x] Index search correctly returns hits for Chinese search queries
- [x] `go test ./internal/recall ./internal/index` passes
