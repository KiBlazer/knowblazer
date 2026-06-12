# Knowblazer CLI Usage Guide

Knowblazer is designed with a set of easy-to-use CLI commands. This document explains the main workflows, commands, and options.

---

## 1. Main Commands

### `knowblazer init [path]`
Initializes a new Knowblazer memory vault folder (defaults to `~/knowblazer-notes`).
It will create the folder hierarchy (`inbox/`, `daily/`, `projects/`, `experience/`, `quarantine/`) and bootstrap a `.knowblazer/config.json` file.
```bash
knowblazer init ~/my-vault
```

### `knowblazer start`
Invoked from any project directory, this command:
1. Detects or initializes the default memory folder (`~/knowblazer-notes`).
2. Infers the project configuration from the current directory name.
3. Automatically connects the AI coding CLIs it finds (writes setup markers in `CLAUDE.md` and/or `AGENTS.md` for MCP tools integration).
```bash
cd /path/to/my-project
knowblazer start
```

### `knowblazer remember "<text>"`
Saves a quick memory signal. The text is scanned for secrets:
*   If clean, it's saved in `experience/auto/` and automatically consolidated.
*   If risky, it's quarantined.
Use `--daily` to write the text directly to today's daily log instead.
```bash
knowblazer remember "Deployment to staging requires running database migrations first"
knowblazer remember "Today I fixed a memory leak in the websockets handler" --daily
```

### `knowblazer recall "<task>"`
Performs keyword-based scoring on your memory vault and generates a **Recall Pack** formatted for LLM injection.
It matches against:
1. Developer preferences (`profile/preferences.md`)
2. Current project facts (`projects/your-project.md`)
3. Relevant experiences (`experience/synthesized/` and `experience/auto/`)
4. Recent daily notes
```bash
knowblazer recall "deploy new websocket feature"
```

### `knowblazer status`
Displays the connection status and statistics of the vault, including:
*   Mapped project name
*   Configured AI integrations
*   Count of fresh, inbox, and consolidated memories
```bash
knowblazer status
```

### `knowblazer sync`
Commits and pushes changes in your private memory vault to your configured remote Git repository.
*   *Note:* It runs a safety scan on modified files first. If quarantine items are present, it warns you before pushing.
```bash
knowblazer sync
```

---

## 2. Advanced Curation & Cwd Mapping

### `knowblazer project set <name> --path <directory>`
Explicitly map a workspace directory to a specific project memory file (`projects/<name>.md`). This allows different workspaces to automatically query distinct project contexts.
```bash
knowblazer project set my-backend-service --path ~/projects/go-backend
```

### `knowblazer review list`
Lists all candidate notes currently pending in the `inbox/` directory.

### `knowblazer review promote <file> --to <destination>`
Moves a candidate note from `inbox/` or `quarantine/` into a permanent memory layer (e.g. `experience/synthesized/` or `projects/`).
```bash
knowblazer review promote inbox/2026-06-12/deploy-lesson.md --to experience/manual
```

### `knowblazer import specstory <path>`
Parses SpecStory history or conversation transcripts at the specified path and imports them into your Knowblazer inbox.

### `knowblazer backup create --output <filename>`
Compresses and archives the entire memory vault into a tarball for manual backups.
```bash
knowblazer backup create --output my-memory-backup.tgz
```
