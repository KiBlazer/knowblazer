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
curl -fsSL https://raw.githubusercontent.com/KiBlazer/knowblazer/main/install.sh | bash
```

### 2. Updating Knowblazer
To check your current version and update to the latest release:

```bash
# Check version
knowblazer version

# Self-update in-place
knowblazer update
```

Or rerun the install script:
```bash
curl -fsSL https://raw.githubusercontent.com/KiBlazer/knowblazer/main/install.sh | bash
```

### 3. From Source (Go developers)
If you prefer compiling from source, make sure you have Go installed, then run:

```bash
go install github.com/KiBlazer/knowblazer/cmd/knowblazer@latest
```

Or clone the repo locally and build:

```bash
make install
```

## Quick Start

### 1. Initialize Memory & Connect AI Tools

Run `setup` from any project directory:

```bash
cd /path/to/project
knowblazer setup
```

Or enable Knowblazer memory globally across all your workspaces:

```bash
knowblazer setup --global
```

`knowblazer setup` automatically creates or reuses `~/knowblazer-notes`, infers the project from the current directory, and connects supported AI coding tools:
*   **Claude Code**: Writes instructions to `CLAUDE.md` and configures MCP automatically.
*   **Codex CLI**: Writes instructions to `AGENTS.md` and configures MCP automatically.
*   **Antigravity (agy)**: Dynamically registers the MCP server in `mcp_config.json` automatically.

Then start whichever connected tool you use:

```bash
claude
# or
codex
```

### 2. Daily Core Commands

```bash
knowblazer remember "Deploys need smoke tests"     # Save a lesson or markdown file
knowblazer remember "Shipped v0.3.0" --daily       # Append note to today's daily log
knowblazer recall "deploy new frontend"            # Generate tailored context pack
knowblazer status                                  # Check memory repo & project status
knowblazer sync                                    # Inspect, commit, push, or pull Git remote
```

Use `remember` for lessons worth keeping; clean lessons are written to `experience/auto/` and automatically consolidated into synthesized memory, while high-risk content is quarantined. Use `status` to check dynamic memory counts, and `sync` only when you choose to commit or push your private memory repo.

## Command Reference

Knowblazer features a clean, unified command taxonomy:

### Core Commands (Daily Engineering)
* `knowblazer setup [--global] [--path <dir>] [--tool <name>] [--skip-mcp]`: Initialize repo and connect AI coding tools.
* `knowblazer remember <text|file> [-d|--daily]`: Record lesson, solution, markdown file, or daily note.
* `knowblazer recall <task> [--project <name>] [--output <file>]`: Retrieve targeted context pack for coding.
* `knowblazer status [--path <dir>]`: Check active workspace mapping, instruction files, and memory statistics.
* `knowblazer sync [status|commit|push|pull] [-m <msg>]`: Git synchronization with sensitive scanning pre-checks.

### Management & Curation
* `knowblazer memory list`: List unreviewed inbox candidate memories.
* `knowblazer memory review [suggest]`: Review candidates or generate consolidation suggestions.
* `knowblazer memory promote <file> --to <target>`: Move reviewed candidate memory into experience hierarchy.
* `knowblazer memory reject <file>`: Reject and remove inbox candidate.
* `knowblazer memory consolidate`: Consolidate fresh auto memories into higher-signal Markdown summaries.
* `knowblazer memory search <query>`: BM25/keyword search across the entire memory index.
* `knowblazer memory capture <file>`: Import markdown note directly to inbox with sensitive content scan.
* `knowblazer memory import specstory <path>`: Distill historical AI coding sessions into inbox memories.
* `knowblazer memory scan <path>`: Scan file or directory for exposed secrets and tokens.
* `knowblazer daily [show|add <text>]`: View or update today's engineering daily log.
* `knowblazer project [list|set <name>|clear]`: Manage workspace-to-project mappings.
* `knowblazer doctor`: Verify repository integrity, config, and check for quarantine leaks.
* `knowblazer backup <create|restore> [--passphrase <text>]`: Encrypted archive creation and restoration.

### System & Integration
* `knowblazer mcp [serve]`: Run Model Context Protocol server over stdio for AI tools.
* `knowblazer update`: Self-update knowblazer binary to the latest GitHub release.
* `knowblazer version`: Display current version, build commit, and platform architecture.

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

## Comparison: Knowblazer vs. SQLite-based / Others

| Feature             | Knowblazer (Markdown/Git)   | SQLite-based / Others          |
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
