# 04: Hierarchical Markdown Section Extraction with Heading Breadcrumbs

**What to build:** Add Markdown section splitting by `##` and `###` headers in memory documents so recall extracts only the relevant section along with its parent breadcrumb (e.g., `[project > section]`), preventing entire 8KB multi-topic documents from consuming the context budget.

**Blocked by:** 01 (BM25 scoring algorithm)

**Status:** resolved

- [x] Header-aware section splitter parses markdown into titled blocks with hierarchical breadcrumbs
- [x] Section-level matching extracts matching sections instead of dumping monolithic documents
- [x] Short notes without subheadings are preserved intact as single blocks
- [x] Tests verify section extraction and breadcrumb formatting in Recall Pack
- [x] `go test ./internal/recall` passes

