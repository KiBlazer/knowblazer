# Knowblazer Product Requirements

> Status: Historical baseline from 2026-04-29. This document describes the original private Git-backed engineering memory direction. For current dynamic-memory positioning and requirement evolution, see [Engineering positioning](../../current/positioning.md) and [Requirements history](../requirements-history.md).

Date: 2026-04-29

## 1. Background and Opportunity

AI coding tools are moving from a single IDE plugin into a daily workflow that spans multiple tools, terminals, and models. Developers may use Claude Code, Codex, Gemini, Cursor, SpecStory, and other tools at the same time, but long-term memory does not travel well between them.

The real problem is not whether an AI tool can remember things. The problem is:

> How can developers own an engineering memory system that they control, can move, can audit, and can reuse across different AI coding tools?

Now is the right time to build Knowblazer for four reasons:

- AI coding tools are diversifying, and personal engineering experience is being scattered across tools and sessions.
- Projects like SpecStory are making AI conversation capture more mature, but raw conversations are not the same as long-term engineering memory.
- Developers already understand local-first workflows such as Git, Markdown, dotfiles, and private repositories.
- The memory server and agent memory backend space is crowded, but "private Git-backed engineering memory" is still a narrower and clearer opportunity.

Knowblazer should focus on this narrow opportunity: help developers turn experience, preferences, project context, and hard-won lessons from AI coding work into their own private Git repository, and recall that memory when different AI coding tools need it.

## 2. Product Positioning

Knowblazer is:

> Private Git-backed engineering memory for AI coding.

Knowblazer is an open-source tool for maintaining your own private Git engineering memory repository, so AI coding tools can safely reuse your engineering experience.

Product boundaries must stay clear:

- Knowblazer is a tool, specification, template set, and adapter layer.
- User memory repositories are stored by default in a local directory or user-owned private Git remote.
- In the first stage, Knowblazer does not provide an official cloud service and does not host personal engineering memory.
- Markdown/Git is the primary storage. Indexing, search, MCP, and vector databases are enhancement layers.

One-line distinction:

```text
SpecStory remembers AI coding conversations.
Knowblazer turns selected conversations, notes, and lessons into a private engineering memory repo.
```

## 3. Target Users

The first stage serves individual developers only. It is not a team knowledge base or enterprise collaboration platform.

### 3.1 Initial User Profiles

**Independent developers switching between multiple AI tools**

These users work with Claude Code, Codex, Cursor, and Gemini at the same time. Each tool knows some context, but there is no shared place for stable personal engineering preferences and project experience. They need a tool-independent private memory repository.

**Multi-project maintainers**

These users maintain multiple projects and frequently switch between repositories, servers, deployment flows, and stacks. Their problem is not a lack of documentation; it is that project context, deployment pitfalls, operational constraints, and decision reasons are scattered across places.

**Privacy-sensitive engineers**

These users may handle client projects, internal systems, server details, or database connections. Compared with cloud memory services, they prefer storing engineering memory in a private Git repository, self-hosted Git, or a purely local directory.

### 3.2 Non-Target Users

The first stage is not for:

- Non-technical personal knowledge management users.
- Team knowledge base administrators.
- Enterprise compliance knowledge platforms.
- Users who want complete AI conversation cloud sync and sharing.
- Users who want a purely graphical knowledge base without Git basics.

## 4. Core Problems

Knowblazer should solve four problems in the first stage.

### 4.1 Experience Has No Stable Home

Engineering experience may be scattered across AI conversations, local analysis directories, project READMEs, Cursor rules, Claude memory, and Codex sessions. After switching machines, tools, or projects, that experience is hard to reuse.

### 4.2 AI Tools Cannot Share Long-Term Memory

Claude Code may know a preference that Codex does not. Context in Cursor project rules may not be available to Gemini CLI. Developers repeatedly explain their engineering preferences, project background, and operational constraints.

### 4.3 Automatic Saving Creates Privacy Risk

AI coding sessions may contain tokens, passwords, private keys, database connection strings, internal domains, or customer information. It is unsafe to automatically save and push every conversation.

### 4.4 Raw Conversations Are Not Long-Term Memory

Full transcripts are too long, too noisy, and too mixed. What has long-term value is selected and reviewed material: preferences, decisions, project context, pitfalls, debugging methods, and operational constraints.

## 5. MVP User Journey

The MVP is not about supporting many features. It is about letting one developer complete the first private engineering memory path within ten minutes.

Target path:

```text
install Knowblazer
  ↓
init a private memory repo
  ↓
capture one Markdown lesson
  ↓
scan for secrets
  ↓
store in inbox or quarantine
  ↓
promote one reviewed lesson
  ↓
generate one recall pack for an AI coding task
```

If this path works, Knowblazer's core hypothesis is validated.

## 6. Scope and Priorities

The first version must stay focused. The priorities below guide implementation tradeoffs.

### 6.1 P0: Must Have

**init**

Initialize a local Knowblazer memory repository with the minimal directory structure, entry instructions, and base policy files.

**capture**

Save a local Markdown file into the memory repository. By default it goes to `inbox/YYYY-MM-DD/` and keeps source, timestamp, title, and basic metadata.

**scan**

Scan content before saving, committing, or pushing. When high-risk content is found, the default behavior is to write or move it to `quarantine/` and block automatic commit or push.

**promote**

Promote reviewed candidate memory from `inbox/` or `daily/` into long-term directories such as `experience/`, `projects/`, or `profile/`. The first version can be an explicit command or a clear manual workflow.

**recall**

Generate a short context pack from a task description. The first version can use filenames, directories, and keyword matching; semantic vector search is not required.

### 6.2 P1: Should Have

**sync helper**

Provide sync helpers based on the user's own Git remote. Knowblazer may prompt or wrap `git status`, `git commit`, and `git push`, but must not provide an official remote.

**daily**

Support daily short-term notes. First-version recall can read today's and yesterday's daily files by default.

**project mapping**

Map the current working directory to a project memory file under `projects/`, so recall can prioritize current project context.

### 6.3 P2: Defer

The following are not part of the first core delivery:

- Automatic Claude Code hook installation.
- Codex, Gemini, and Cursor adapters.
- SpecStory `.specstory/history/` import.
- MCP server.
- Vector search, SQLite indexing, or hybrid search.
- Automatic organization or OpenClaw-like dreaming background promotion.
- Web UI.
- Team sharing.
- Official cloud sync.

These can enter the later roadmap, but they must not block P0.

## 7. Memory Model

Knowblazer needs layered memory management. The purpose of layers is not complexity; it is preventing long-term memory from becoming a messy chat archive.

### 7.1 First-Version Minimal Directory

The first version creates this structure by default:

```text
knowblazer-notes/
├── AI-SETUP.md
├── inbox/
│   └── YYYY-MM-DD/
├── daily/
│   └── YYYY-MM-DD.md
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

The first version does not create `relationships/`, `templates/`, or `bin/` by default. These can be added later to avoid making the initial structure too heavy.

### 7.2 Layer Rules

`inbox/` is the raw candidate layer. AI conversation summaries, SpecStory imports, local analysis documents, and temporary manual notes start here. It is not included in recall by default.

`daily/` is the short-term working layer. It records what happened today, current task state, temporary observations, and unsorted context. Recall can read today's and yesterday's daily files.

`profile/` is the stable personal preference layer. It stores engineering preferences, decision principles, writing style, and other stable information.

`projects/` is the long-term project context layer. It stores project background, architecture, deployment methods, important directories, operational constraints, and common pitfalls.

`experience/` is the reusable engineering experience layer. It stores deployment, frontend, backend, operations, AI tools, and similar experience.

`quarantine/` is the sensitive-content isolation layer. It is never included in recall and is not automatically committed or pushed.

`recall/` is the output layer. It stores or temporarily generates task-specific context packs. It can be excluded from Git by default.

### 7.3 Promotion Mechanism

Long-term memory must not be written by AI automatically and unconditionally.

Base flow:

```text
capture / import
  ↓
scan
  ↓
inbox / daily / quarantine
  ↓
review
  ↓
promote
  ↓
profile / projects / experience
  ↓
recall
```

The first version only requires lightweight promotion: move a file, generate a target-path suggestion, or update document status. Automatic scoring, periodic organization, and OpenClaw-like dreaming can be considered later.

## 8. Data and Privacy Boundaries

Knowblazer must make data ownership a product principle, not an implementation detail.

### 8.1 Users Own Their Memory Repository

Engineering memory belongs to the user. Knowblazer does not own, host, or collect a user's personal engineering memory by default.

### 8.2 Local Storage by Default

All P0 features must run without a network connection. Without a Git remote, official account, or LLM API, users should still be able to init, capture, scan, promote, and recall.

### 8.3 User-Selected Remotes

Cross-device sync can only happen through a Git remote configured by the user. Optional remotes include GitHub private repositories, GitLab private repositories, self-hosted Gitea/Forgejo, internal company Git, NAS, or personal servers.

### 8.4 No Official Hosting

The first stage does not provide Knowblazer Cloud, official memory hosting, or an official account requirement.

If cloud service appears in the future, it must be optional and must not change the default local-first and user-owned Git repository model.

### 8.5 Scan Before Push

Any automatic commit, sync, or push flow must scan sensitive information first. If high-risk content is found, the default behavior is to block the push and tell the user to review and handle it.

## 9. Recall Strategy

The goal of recall is not searching the whole knowledge base. It is generating a short, relevant, readable context pack for the current AI coding task.

First-version recall input:

- Task description, such as `deploy new frontend`.
- Current working directory.
- Optional project name.

First-version recall candidate sources:

- Stable, short preferences in `profile/`.
- The current project's file under `projects/`.
- Task-keyword-matching files under `experience/`.
- Today's and yesterday's `daily/` files.

Default exclusions:

- `inbox/`, unless explicitly requested.
- `quarantine/`, always excluded from default recall.
- Historical `recall/` output, unless explicitly requested.

First-version recall output must be:

- Human-readable.
- Directly pasteable into Claude Code, Codex, Gemini, or Cursor.
- Free of high-risk sensitive content found by scanning.
- Length-controlled by default, so it does not dump the whole memory repository into context.

## 10. Key User Flows

### 10.1 First Use

After installing Knowblazer, a developer runs init in an empty directory and gets a readable, Git-manageable memory repository. Even without a remote repository, local features remain usable.

### 10.2 Save One Real Lesson

After finishing deployment debugging, the developer writes a Markdown summary. Knowblazer capture scans it first. If it passes, it enters `inbox/`; if it contains sensitive content, it enters `quarantine/` and automatic commit is blocked.

### 10.3 Promote One Long-Term Lesson

The developer reviews a candidate in `inbox/` and promotes the useful lesson into `experience/deployment/` or a project-specific file.

### 10.4 Recall Context for the Current Task

Before asking an AI tool to work on a task, the developer runs recall. Knowblazer generates a short context pack from profile, project, experience, and daily notes.

### 10.5 Restore on a New Machine

On a new computer, the developer clones their private Knowblazer memory repository. After running init or doctor, AI coding tools can continue using the same engineering preferences and project experience.

## 11. Acceptance Criteria

P0 must satisfy the following testable criteria.

### 11.1 init

- Running init in an empty directory creates the first-version minimal structure.
- It creates `AI-SETUP.md`, `system/memory-policy.md`, and `system/privacy-policy.md`.
- Re-running init does not overwrite existing user memory.

### 11.2 capture

- Given a Markdown file, Knowblazer saves it to `inbox/YYYY-MM-DD/`.
- The saved file preserves the original title or generates a readable title.
- The saved file includes source path and capture time.

### 11.3 scan

- Files containing mock tokens, mock private keys, mock passwords, and mock database URLs are identified as high risk.
- High-risk files are not automatically committed or pushed.
- High-risk content enters `quarantine/` or remains local for handling, and the user gets a clear prompt.

### 11.4 promote

- The user can promote an inbox memory into `experience/`, `projects/`, or `profile/`.
- After promotion, the original file state is clear so users can tell which content entered long-term memory.

### 11.5 recall

- `recall --task "<task>"` generates human-readable context.
- Output includes relevant profile, current project, matching experience, and recent daily notes.
- Output does not include `inbox/` or `quarantine/` by default.
- Output is length-controlled and does not simply concatenate the whole repository.

### 11.6 Local-First

- P0 features run without a network connection.
- P0 works without configuring an official Knowblazer account.
- P0 works without configuring a Git remote.

## 12. Risks and Counter-Metrics

These signals indicate that product direction or design may need adjustment:

- Users treat Knowblazer only as a backup script and never use recall.
- `inbox/` grows quickly, but users never review or promote.
- The directory structure confuses users.
- Scan false positives are so frequent that users disable or bypass scanning.
- Recall output is too long and AI tools still miss the point.
- Recall output is too generic and does not reduce repeated explanation.
- Users worry about privacy because they mistakenly think Knowblazer uploads data to an official service.
- Users believe SpecStory already fully covers Knowblazer's value.

## 13. Explicitly Not Doing

The first stage explicitly does not do:

- Official cloud sync.
- Official memory hosting.
- Team knowledge base collaboration.
- Web UI.
- Full AI conversation recording.
- IDE plugin.
- Generic memory server.
- Default MCP server.
- Vector database.
- Automatic understanding and organization of all historical chats.
- Automatic generation of high-quality long-term memory.

## 14. Later Roadmap

After P0 works, consider:

- Codex, Claude Code, Gemini, and Cursor adapters.
- SpecStory `.specstory/history/` import.
- Git sync helper.
- Smarter project mapping.
- Optional SQLite or vector indexing.
- MCP server.
- OpenClaw-like periodic organization, but only if it is reviewable and reversible.
- Self-hosted sync service or encrypted backup.

These extensions must not change the first principle: the user's own private Git engineering memory repository is the primary storage, and Knowblazer does not host user memory officially.

## 15. One-Sentence Summary

Knowblazer is not about saving every AI conversation or hosting user data. It helps developers maintain important engineering experience in their own private Git repository, making it portable, auditable, and recallable as long-term engineering memory.
