# Knowblazer Engineering Positioning

Date: 2026-04-30

Related docs:

- [Documentation index](../README.md)
- [Requirements history](../history/requirements-history.md)

## One-line Positioning

Knowblazer is a local-first dynamic engineering memory system for AI coding.

It continuously collects engineering signals, turns them into higher-signal memory, and feeds the best current context back to AI coding tools.

## What Knowblazer Is

Knowblazer is not just a memory folder or transcript archive. It is an automatic memory loop:

```text
coding work and AI sessions
  -> capture memory signals
  -> scan privacy and risk
  -> store fresh memory
  -> consolidate and synthesize
  -> recall the best context
  -> improve future AI work
```

The goal is to make AI coding context improve over time without requiring the user to manually organize every note.

## Core Engineering Role

Knowblazer owns the memory layer between raw AI conversations and AI tool context windows.

Its responsibilities are:

- collect useful engineering signals from AI coding work
- keep memory private and user-owned by default
- scan and quarantine sensitive or risky content
- consolidate noisy fresh memories into concise project and experience memory
- rank memory by relevance, recency, scope, and confidence
- provide compact context packs to Claude Code, Codex, Gemini, Cursor, and other tools

## Memory Lifecycle

### 1. Capture

New signals come from CLI commands, MCP calls, imported transcripts, daily notes, project notes, and future integrations.

Fresh reusable lessons should land in `experience/auto/` first. This keeps capture fast and avoids forcing a taxonomy too early.

### 2. Safety Scan

Every captured or synthesized memory must pass scanning before it becomes usable context.

High-risk content goes to `quarantine/` and must be excluded from default recall, indexing, and AI context.

### 3. Consolidation

Knowblazer should automatically process fresh memory into higher-signal memory.

Consolidation should:

- deduplicate repeated lessons
- merge related memories
- remove temporary task chatter
- preserve durable causes, constraints, and decisions
- update project-level facts in `projects/`
- update reusable lessons in `experience/`
- mark source memories as consolidated or superseded instead of deleting them immediately

Manual review can exist as a power-user tool, but it must not be required for memory to become useful.

### 4. Recall

Recall should prefer synthesized memory over raw automatic memory.

Suggested priority:

1. relevant `projects/` memory
2. synthesized `experience/` memory
3. recent `daily/` memory when useful
4. fresh `experience/auto/` memory as lower-confidence context

`inbox/` and `quarantine/` should not enter default recall.

### 5. Feedback

Usage should feed back into memory quality.

Knowblazer should track which memories are recalled, which become stale, which are repeatedly useful, and which conflict with newer facts. This feedback should influence future consolidation and ranking.

## Source Of Truth

Markdown and Git remain the source-of-truth storage layer.

Indexes, MCP servers, embeddings, background jobs, and adapters are acceleration layers. They may improve recall and synthesis, but they must not replace the human-readable memory repo.

## Privacy Boundary

Knowblazer is local-first and user-owned.

By default, it should not require:

- a Knowblazer account
- an official Knowblazer server
- a network connection
- a Git remote
- an LLM API for basic local workflows

If networked or model-assisted consolidation is added later, it must be explicit, inspectable, and compatible with the local-first privacy model.

## Non-goals

Knowblazer is not:

- a complete AI conversation archive
- a generic personal knowledge base
- a team wiki
- an enterprise document management system
- a hosted memory backend by default
- a system that requires users to manually curate every useful memory

## Product Direction

The product should optimize for a dynamic closed loop:

```text
collect -> process -> synthesize -> recall -> learn from use -> process again
```

The user should feel that Knowblazer quietly improves AI coding context over time while keeping the memory repo portable, auditable, and private.
