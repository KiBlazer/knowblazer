# 03: U-Shaped Attention Reordering (LongContextReorder)

**What to build:** Implement dual-ended alternating placement (`LongContextReorder`) for recalled experience items to place the highest-relevance items at the beginning and end of the recalled list, mitigating LLM "Lost-in-the-Middle" degradation.

**Blocked by:** 01 (BM25 scoring algorithm)

**Status:** resolved

- [x] `LongContextReorder` function places rank 1 at start, rank 2 at end, subsequent ranks inwards
- [x] Lists with <= 2 items remain unchanged
- [x] Integrated into `relevantExperience` result ordering
- [x] Tests verify exact ordering transformation for 3, 4, and 5 candidate lists
- [x] `go test ./internal/recall` passes

