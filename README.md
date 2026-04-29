# Knowblazer

Private Git-backed engineering memory for AI coding.

Knowblazer is an open-source tool for maintaining your own private engineering memory repo. It helps turn selected AI coding conversations, project notes, and hard-won lessons into portable Markdown memory that Claude Code, Codex, Gemini, Cursor, and other AI coding tools can reuse.

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
- a privacy-first capture, scan, promote, and recall flow

Knowblazer is not:

- a cloud memory service
- an official hosting platform for your notes
- a full AI transcript recorder
- a generic vector database
- another MCP memory server
- a team knowledge base

## Installation

Install the latest released CLI with Go:

```bash
go install github.com/knowblazer/knowblazer/cmd/knowblazer@latest
```

For local development from this repository:

```bash
make install
```

## MVP Flow

The first version focuses on one path:

```text
init a private memory repo
  ↓
capture one Markdown lesson
  ↓
scan for secrets
  ↓
store in inbox or quarantine
  ↓
promote reviewed memory
  ↓
generate a recall pack for an AI coding task
```

P0 commands:

```bash
knowblazer init ~/knowblazer-notes
knowblazer doctor --repo ~/knowblazer-notes
knowblazer capture ./deploy-lesson.md --repo ~/knowblazer-notes
knowblazer scan ~/knowblazer-notes
knowblazer promote inbox/2026-04-29/deploy-lesson.md --to experience/deployment --repo ~/knowblazer-notes
knowblazer recall --task "deploy new frontend" --project my-project --repo ~/knowblazer-notes
```

Roadmap commands:

```bash
knowblazer daily add "fixed flaky deploy" --repo ~/knowblazer-notes
knowblazer daily show --repo ~/knowblazer-notes
knowblazer project set my-project --path ~/work/my-project --repo ~/knowblazer-notes
knowblazer sync status --repo ~/knowblazer-notes
knowblazer import specstory .specstory/history --repo ~/knowblazer-notes
knowblazer adapter claude --repo ~/knowblazer-notes
knowblazer index build --repo ~/knowblazer-notes
knowblazer index search deploy rollback --repo ~/knowblazer-notes
knowblazer mcp serve --repo ~/knowblazer-notes
knowblazer review list --repo ~/knowblazer-notes
knowblazer dream --repo ~/knowblazer-notes
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
├── profile/
│   ├── preferences.md
│   └── decision-principles.md
├── projects/
├── experience/
│   ├── deployment/
│   ├── frontend/
│   ├── backend/
│   ├── ai-tools/
│   └── operations/
├── system/
│   ├── memory-policy.md
│   └── privacy-policy.md
├── quarantine/
└── recall/
```

Core rules:

- `inbox/` is for unreviewed candidate memory.
- `daily/` is for short-term working notes.
- `profile/`, `projects/`, and `experience/` are reviewed long-term memory.
- `quarantine/` is for sensitive or risky content and is never included in default recall.
- `recall/` is for generated task context packs.

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

Knowblazer currently has a local-first CLI implementation for local use. The implementation covers `init`, `doctor`, `scan`, `capture`, `promote`, `recall`, `daily`, `project`, `sync`, `import specstory`, `adapter`, `index`, `mcp serve`, `review`, `dream`, and `backup`.

Current artifacts:

- [Product requirements](docs/requirements.md)
- [MVP spec](docs/mvp-spec.md)
- [P0 implementation plan](docs/implementation-plan.md)
- [Default memory repo template](templates/default-memory-repo)

## Design Principles

- User-owned memory first.
- Markdown/Git as source of truth.
- Recall should be short, relevant, and reviewable.
- Raw AI conversations are inputs, not long-term memory by default.
- Promotion requires review.
- Privacy is a product requirement, not an optional plugin.

## Relationship To SpecStory

SpecStory is useful for capturing AI coding conversations and intent history.

Knowblazer is complementary: it focuses on promoting selected transcripts, notes, and lessons into a durable private engineering memory repo.

```text
SpecStory: raw conversation and intent archive
Knowblazer: reviewed private engineering memory
```

## License

Knowblazer is released under the MIT License. See [LICENSE](LICENSE).
