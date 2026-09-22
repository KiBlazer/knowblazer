# Spec: RAG-Inspired Retrieval Quality, Attention Optimization, and Memory Packaging Enhancements

Status: completed

## Problem Statement

When AI coding agents (Claude Code, Codex, Antigravity, Gemini) work with Knowblazer to retrieve engineering memory, developers face several subtle yet damaging retrieval and context assembly shortcomings:

1. **Length Bias and Term Frequency Distortion**: Knowblazer currently counts raw keyword occurrences without inverse document frequency (IDF) or document length normalization. Long, verbose notes (e.g. 8KB multi-issue daily notes or scratch dumps) easily outscore short, laser-focused 200-word engineering rules that contain fewer total keyword repetitions, pushing the best answers down or out of the recall pack.
2. **Low-Signal Noise Pollution**: The retrieval logic admits any file where the score is greater than zero. If a task query contains common development terms (e.g., "test", "build", "git", "config"), unrelated historical notes that casually mention these words get included in the candidate list, filling the prompt with low-confidence noise.
3. **Monolithic File Ingestion**: Knowblazer loads entire Markdown files up to 8KB at a time. If an experience or daily note contains several distinct troubleshooting cases under different headings, the entire document is packed into the 20KB budget. Irrelevant sections consume precious token capacity, causing subsequent high-value notes to be truncated or omitted.
4. **Attention Decay in Long Contexts (Lost in the Middle)**: Large language models exhibit U-shaped attention distribution when reading long prompt packs—they pay greatest attention to the start and end of the retrieved context, and tend to miss or disregard instructions buried in the middle. Knowblazer currently outputs experience notes in a simple linear sequence, placing lower-priority items at the tail and letting valuable secondary memories sit in the attention dead zone.
5. **Ambiguous Context Governance**: The recall pack renders developer preferences, project context, and past experience notes without explicit semantic boundaries. AI agents occasionally confuse past project-specific workarounds with universal engineering constraints, leading to over-generalized or contradictory code suggestions.
6. **Black-box Recall Diagnostics**: When an engineer searches or notices that a particular memory was not retrieved for a task, there is no diagnostic or explainability output to inspect why candidates were scored, filtered, or ranked in that order.

## Solution

1. **Pure Go BM25 Scoring**: Replace naive keyword counting with a zero-dependency BM25 ranking algorithm in the standard library. Incorporate term frequency saturation, inverse document frequency (IDF), and document length normalization ($k_1=1.2, b=0.75$) so short, highly concentrated notes outrank long, diluted notes.
2. **Relevance Quality Gate**: Introduce a minimum relevance score threshold. When candidate memories do not meet the confidence baseline, they are omitted entirely rather than diluting the context pack with marginal hits.
3. **Section-Level Hierarchical Ingestion**: Parse Markdown documents into structural sections based on Markdown headers (`##` / `###`) with contextual breadcrumbs (e.g., `kiblazer > Durable Constraints`). Index and recall at the section level, injecting only relevant sections and their breadcrumbs into the recall pack.
4. **U-Shaped Attention Reordering (LongContextReorder)**: Reorder retrieved experience items using a dual-ended alternate folding algorithm, placing the most critical memories at the beginning and end of the experience section to maximize LLM compliance.
5. **Dual-Tier Context Labeling**: Separate the recall pack into explicit semantic tiers: Tier 1 for non-negotiable governance (Mandatory Constraints & Preferences) and Tier 2 for empirical heuristics (Reference Experience & Past Cases), guiding the agent to treat past cases as reference material rather than absolute rules.
6. **Recall Explainability & Audit Trail**: Provide an explainability mode for recall (via CLI flag and structured metadata) that reveals matched terms, BM25 scores, document length adjustments, and ranking rationale.

## User Stories

1. As an engineer with a concise 5-line note capturing a critical build caveat, I want BM25 ranking to score my concise note higher than an 8KB daily log with casual keyword mentions, so that the AI coding assistant receives the most authoritative instruction.
2. As a developer running a task that mentions generic words like "test", I want Knowblazer to filter out weakly matching notes with a relevance threshold, so that my AI agent's prompt is not cluttered with irrelevant historical chatter.
3. As an engineer managing a multi-section project note with distinct headings (e.g., architecture, deployment, constraints), I want Knowblazer to extract only the matching section with its breadcrumb heading, so that my 20KB recall budget is filled with dense, relevant information rather than whole-file boilerplate.
4. As a developer writing coding tasks with multiple relevant past experiences, I want the retrieved memories to be arranged using U-shaped attention reordering, so that the AI model does not overlook key lessons placed in the middle of the pack.
5. As an AI coding assistant reading the recall pack, I want clear semantic demarcation between mandatory project constraints and empirical past experiences, so that I never mistake a one-off temporary hack for a durable architectural requirement.
6. As an engineer troubleshooting why a specific memory did not appear in the recall pack, I want to run `knowblazer recall "<task>" --explain` to see the candidate list, BM25 scores, term matches, and filter decisions.
7. As an AI agent using the MCP tool `knowblazer_recall` or `knowblazer_context`, I want the recalled context to remain compact and high-signal, so that my context window is utilized efficiently.
8. As a developer working across multiple projects in the same memory repository, I want section extraction to preserve the parent project and file breadcrumb, so that the AI understands the provenance and context of every recalled snippet.
9. As a terminal user running `knowblazer memory search`, I want search results to be ranked by BM25 relevance with clear match snippets, so that I can quickly pinpoint past knowledge.
10. As a system administrator on an air-gapped or offline developer machine, I want all retrieval, scoring, and reordering algorithms to run entirely within the Go standard library with zero external API calls or database daemons, so that Knowblazer remains instant and self-contained.

## Implementation Decisions

1. **Standard Library BM25 Ranking**:
   - Implement Okapi BM25 scoring directly in the Go indexing and recall modules.
   - Use corpus statistics: total document/section count $N$, document frequency $n(q_i)$ for term $q_i$, document length $|D|$, and average document length $avgdl$.
   - Standard parameters $k_1 = 1.2$ and $b = 0.75$.
   - Retain existing multilingual tokenization (Latin words and CJK unigrams/bigrams) as the input stream to BM25 term collection.

2. **Relevance Thresholding (Quality Gate)**:
   - Establish a minimum score cutoff for experience candidates.
   - If no candidates exceed the threshold, emit an explicit fallback message (`_No matching memory found._`) rather than admitting low-confidence candidates.

3. **Hierarchical Section-Aware Chunking**:
   - Enhance the Markdown parser to split documents by top-level section headers (`##` and `###`) while preserving file-level frontmatter metadata.
   - Prepend hierarchical breadcrumbs (e.g., `[project > section]`) to each extracted chunk.
   - For short documents without subheadings, treat the entire document as a single chunk.
   - When building the recall pack, assemble matched sections up to per-section and total pack budgets.

4. **U-Shaped Attention Reordering**:
   - For ranked lists of length $N > 2$, apply an alternating placement strategy:
     - Rank 1 placed at index 0 (top)
     - Rank 2 placed at index $N-1$ (bottom)
     - Rank 3 placed at index 1
     - Rank 4 placed at index $N-2$, etc.
   - This ensures the highest-signal memories are positioned at the extremes where LLM attention is strongest.

5. **Semantic Context Packaging**:
   - Refactor the Markdown output template of the Recall Pack into two distinct tiers:
     - **Tier 1 (Mandatory Constraints & Governance)**: Developer preferences, principles, and active project durable constraints.
     - **Tier 2 (Reference Experience & Past Heuristics)**: Ranked, section-sliced experience and recent daily notes.
   - Include clear system-level instructions in each tier defining how the model should interpret and apply the context.

6. **Recall Explainability Interface**:
   - Add an `Explain` boolean to recall options and a `--explain` flag to `knowblazer recall`.
   - When enabled, append or output a structured diagnostic block detailing:
     - Query tokens and terms extracted.
     - Scored candidate items with raw term counts, BM25 scores, and length penalties.
     - Applied threshold cutoffs and reordering shifts.

## Testing Decisions

- **Good Test Criteria**: Tests must verify observable behavior against realistic Markdown memory repositories (comparing ranked order, pack content, section breadcrumbs, and explainability output) using temporary directories (`t.TempDir()`), rather than asserting on internal mathematical state.
- **Seams Under Test**:
  - **Primary Seam**: `recall.Generate(repoRoot, opts)` in `internal/recall`. This is the top-level API called by both CLI `knowblazer recall` and MCP `knowblazer_context`. Testing here verifies BM25 scoring, section slicing, threshold filtering, U-shaped reordering, and prompt assembly in one unified pass.
  - **Secondary Seam**: `index.Search(repoRoot, query)` in `internal/index`. Verifies that BM25 search ranking and snippet generation correctly surface relevant documents over diluted long documents.
  - **CLI Integration Seam**: `cli.Run(args)` with `recall --explain` and `memory search <query>` to ensure exit codes, flags, and formatted outputs match contracts.
- **Prior Art**:
  - Existing tests in `internal/recall/recall_test.go` and `internal/index/index_test.go` demonstrate how to construct fixture notes in `t.TempDir()`, invoke the generator, and inspect the resulting markdown pack.

## Out of Scope

- Introducing external database engines (e.g. PostgreSQL, SQLite with vector extensions).
- Adding remote vector embedding calls or external neural rerankers during synchronous recall.
- Modifying the on-disk storage format from plain Git-backed Markdown files.
- Automated internet web search fallbacks.

## Further Notes

All improvements maintain Knowblazer's core architectural tenets: single static binary, zero external runtime dependencies, sub-50ms execution speed, and full cross-platform compatibility across Linux, macOS, and Windows.
