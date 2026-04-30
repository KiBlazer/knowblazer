# Knowblazer Requirements History

This file records how Knowblazer requirements evolve over time. It is not a replacement for detailed specs; it explains which documents represent which product stage.

## 2026-04-29 — P0 Private Engineering Memory Repo

Primary documents:

- [Product requirements](2026-04-29-p0/requirements.md)
- [MVP spec](2026-04-29-p0/mvp-spec.md)
- [P0 implementation plan](2026-04-29-p0/implementation-plan.md)

Positioning at this stage:

> Private Git-backed engineering memory for AI coding.

Core path:

```text
init -> capture -> scan -> promote -> recall
```

Main assumptions:

- Markdown/Git is the source of truth.
- Users explicitly capture and promote memory.
- The first version should avoid cloud services, accounts, LLM APIs, automatic commits, and automatic pushes.
- Recall should exclude raw inbox content and quarantine by default.

This stage validated the local-first storage model and the initial CLI shape.

## 2026-04-30 — Claude Code First + Automatic Long-term Memory

Primary source:

- Current README and implementation

Requirement shift:

- Claude Code became the preferred first integration path.
- `knowblazer start`, MCP tools, and `knowblazer remember` became central workflows.
- Clean/warning memory can be automatically written to `experience/auto/`.
- High-risk content goes to `quarantine/`.
- The default memory repo template became smaller and more functional.

Core path:

```text
start -> remember -> scan -> auto-store -> recall through CLI/MCP
```

This stage moved Knowblazer from explicit manual promotion toward automatic memory capture.

## 2026-04-30 — Dynamic Engineering Memory System

Primary document:

- [Engineering positioning](../current/positioning.md)

Current positioning:

> Local-first dynamic engineering memory system for AI coding.

Requirement shift:

- Knowblazer should not only store memory; it should operate a closed memory loop.
- The product should automatically collect memory, process and synthesize it, and feed high-signal context back to AI tools.
- Manual review can exist, but it must not be required for memory to become useful.
- `experience/auto/` is a fresh-memory buffer, not the final knowledge layer.
- Synthesized `projects/` and `experience/` memory should be preferred during recall.

Target loop:

```text
collect -> process -> synthesize -> recall -> learn from use -> process again
```

Open implementation implications:

- Add an automatic consolidation workflow.
- Track memory status such as fresh, consolidated, superseded, and quarantined.
- Rank memory by relevance, recency, scope, confidence, and usage feedback.
- Keep Markdown/Git as the human-readable source of truth while using indexes, MCP, and future background jobs as acceleration layers.

## Conflict Rule

If older docs mention manual promotion or larger default repo structures, treat those as historical P0 design, not current product direction.

For current direction, prefer:

1. [Engineering positioning](../current/positioning.md)
2. [Requirements history](requirements-history.md)
3. README and current code
4. Older P0 documents as historical background
