# Knowblazer

[English](./README.md) | [中文](./README_ZH.md)

Local-first dynamic engineering memory for AI coding.

Knowblazer is an open-source tool for maintaining your own private engineering memory repo. It collects useful AI coding signals, scans them, stores fresh memory, consolidates them into higher-signal Markdown memory, and feeds the best current context back to Claude Code, Codex, Gemini, Cursor, and other AI coding tools.

Knowblazer does not host your memories. Your memory repo lives where you choose: local disk, GitHub private repo, GitLab private repo, self-hosted Git, NAS, or another user-owned Git remote.

## Why

AI coding work is increasingly spread across multiple tools. One tool may know your deployment habits, another may know a project constraint, and a third may have the debugging session that finally explained a recurring failure.

Raw transcripts are useful, but they are not the same as durable engineering memory. Knowblazer focuses on the layer above raw conversation capture:

- developer preferences
- project context
- reusable deployment and debugging lessons
- decision principles
- daily working notes
- privacy-aware recall packs for AI coding tasks

## What It Is

Knowblazer is:

- a CLI workflow for a private engineering memory repo
- a Markdown/Git memory structure
- a set of templates and policies
- a privacy-first capture, scan, consolidate, and recall flow

Knowblazer is not:

- a cloud memory service
- an official hosting platform for your notes
- a full AI transcript recorder
- a generic vector database
- another MCP memory server
- a team knowledge base

## Installation

### 1. Quick Install (Mac & Linux)
Install the precompiled binary with a single command (no Go runtime required):

```bash
curl -fsSL https://raw.githubusercontent.com/KiBlazer/knowblazer/main/install.sh | sh
```

### 2. From Source (Go developers)
If you prefer compiling from source, make sure you have Go installed, then run:

```bash
go install github.com/KiBlazer/knowblazer/cmd/knowblazer@latest
```

Or clone the repo locally and build:

```bash
make install
```

## Quick Start

Start from any project directory:

```bash
cd /path/to/project
knowblazer start
```

`knowblazer start` creates or reuses `~/knowblazer-notes`, infers the project from the current directory, and automatically connects supported AI coding CLIs:
*   **Claude Code**: Writes instructions to `CLAUDE.md` and configures MCP automatically.
*   **Codex CLI**: Writes instructions to `AGENTS.md` and configures MCP automatically.

For IDEs and other tools that do not support command-line MCP registration, you can generate copy-paste configuration snippets using the `adapter` command:
*   **Cursor**: Run `knowblazer adapter cursor`
*   **Antigravity (agy)**: Run `knowblazer adapter agy`
*   **Gemini CLI**: Run `knowblazer adapter gemini`

Then open whichever connected tool you use:

```bash
claude
# or
codex
```

Daily use:

```bash
knowblazer remember "Deploys need smoke tests"
knowblazer recall "deploy new frontend"
knowblazer status
knowblazer sync
```

Use `remember` for lessons worth keeping; clean lessons are written to `experience/auto/` and automatically consolidated into synthesized memory, while high-risk content is quarantined. Use `status` to check dynamic memory counts, and `sync` only when you choose to commit or push your private memory repo.

## MCP Configuration

To integrate Knowblazer's MCP server with **Cursor**, **Claude Desktop**, or other compatible tools, add the following to your MCP settings file:

```json
{
  "mcpServers": {
    "knowblazer": {
      "command": "knowblazer",
      "args": ["serve"],
      "env": {
        "KNOWBLAZER_NOTES_PATH": "/path/to/your/knowblazer-notes"
      }
    }
  }
}
```

## MVP Flow

The underlying local-first flow remains:

```text
start from a project directory
  ↓
connect Claude Code to private memory
  ↓
remember useful lessons automatically
  ↓
scan for secrets
  ↓
store fresh memory in experience/auto or risky content in quarantine
  ↓
automatically consolidate fresh memory into synthesized context
  ↓
recall task context while coding
```

Core commands:

```bash
knowblazer start
knowblazer remember "Deploys need smoke tests"
knowblazer recall "deploy new frontend"
knowblazer status
knowblazer sync
```

Advanced commands:

```bash
knowblazer init ~/knowblazer-notes
knowblazer setup claude --repo ~/knowblazer-notes --project my-project --path .
knowblazer setup codex --repo ~/knowblazer-notes --project my-project --path .
knowblazer doctor --repo ~/knowblazer-notes
knowblazer scan ~/knowblazer-notes
knowblazer review list --repo ~/knowblazer-notes
knowblazer review promote inbox/2026-04-29/deploy-lesson.md --to experience/deployment --repo ~/knowblazer-notes
knowblazer daily add "fixed flaky deploy" --repo ~/knowblazer-notes
knowblazer project set my-project --path ~/work/my-project --repo ~/knowblazer-notes
knowblazer import specstory .specstory/history --repo ~/knowblazer-notes
knowblazer backup create --repo ~/knowblazer-notes --output knowblazer-backup.tgz
```

Development helpers:

```bash
make test
make build
make install
make clean
```

## Memory Repo Structure

The default memory repo is intentionally human-readable:

```text
knowblazer-notes/
├── .knowblazer/
│   └── config.json
├── AI-SETUP.md
├── inbox/
├── daily/
├── projects/
├── experience/
│   └── auto/
└── quarantine/
```

Core rules:

- `inbox/` is for unreviewed candidate memory from explicit capture/import flows.
- `daily/` is for short-term working notes.
- `projects/` stores durable project facts, constraints, and operating notes.
- `experience/` stores reusable engineering lessons; `experience/auto/` is the fresh-memory buffer and `experience/synthesized/` is the automatic higher-signal consolidation layer.
- `quarantine/` is for sensitive or risky content and is never included in default recall.

## Privacy Model

Knowblazer is local-first.

P0 must work without:

- a Knowblazer account
- a Knowblazer server
- a network connection
- a Git remote
- an LLM API

Knowblazer should not automatically commit or push memory. `knowblazer sync` scans before commit or push, blocks high-risk findings, and only pushes when explicitly invoked. High-risk content should go to `quarantine/` and require manual review.

## Project Status

Knowblazer currently has a local-first CLI implementation for local use. The implementation covers `init`, `doctor`, `scan`, `capture`, `promote`, `remember`, `consolidate`, `recall`, `daily`, `project`, `sync`, `import specstory`, `adapter`, `index`, `mcp serve`, `review`, `dream`, and `backup`.

Current artifacts:

- [Documentation index](docs/README.md)
- [Engineering positioning](docs/current/positioning.md)
- [Requirements history](docs/history/requirements-history.md)
- [Default memory repo template](templates/default-memory-repo)

## Design Principles

- User-owned memory first.
- Markdown/Git as source of truth.
- Recall should be short, relevant, and dynamically adjusted.
- Raw AI conversations are inputs, not durable synthesized memory by default.
- Automatic consolidation should turn fresh memory into higher-signal context.
- Privacy is a product requirement, not an optional plugin.

## Comparison: Knowblazer vs. Nocturne Memory / Others

| Feature             | Knowblazer (Markdown/Git)   | Nocturne Memory / SQLite-based |
| :------------------ | :-------------------------- | :----------------------------- |
| **Storage Medium**  | **Plaintext Markdown files** | Relational Database (Black Box)|
| **Git Integration** | **Native** (`git diff` & branch commits) | None (Binary DB format)        |
| **Secret Scanning** | **Active Scan & Quarantine** | None                           |
| **System Footprint**| **Zero** (Local CLI execution only) | Daemon servers & dashboard web servers |
| **Deployment**      | **Single Go binary** (zero dependency)| Multi-runtime environment setup|

## Relationship To SpecStory

SpecStory is useful for capturing AI coding conversations and intent history.

Knowblazer is complementary: it turns useful signals from transcripts, notes, and lessons into a dynamic private engineering memory system.

```text
SpecStory: raw conversation and intent archive
Knowblazer: dynamic private engineering memory for AI coding
```

## License

Knowblazer is released under the MIT License. See [LICENSE](LICENSE).
