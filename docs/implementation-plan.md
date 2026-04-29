# Knowblazer P0 Implementation Plan

Date: 2026-04-29

## 1. Implementation Goal

The P0 goal is to implement a locally runnable CLI that completes the core path from the MVP specification:

```text
init
  ↓
capture
  ↓
scan
  ↓
promote
  ↓
recall
```

P0 does not require network access, an official account, an LLM API, automatic commits, or automatic pushes.

## 2. Recommended Technology Stack

Go is recommended.

Reasons:

- Single-binary distribution is a good fit for CLI tools.
- The standard library is sufficient for files, paths, time, regex, and testing.
- Cross-platform support is relatively low-cost.
- Go works well for local file workflows such as Git and Markdown.
- P0 does not need a complex runtime or external services.

A small number of dependencies may be used, but P0 should prefer the standard library. CLI argument parsing can start with the standard library `flag` package; Cobra or urfave/cli can be introduced later if better UX is needed.

## 3. Suggested Code Structure

Suggested structure:

```text
knowblazer/
├── cmd/
│   └── knowblazer/
│       └── main.go
├── internal/
│   ├── cli/
│   ├── repo/
│   ├── templates/
│   ├── scan/
│   ├── capture/
│   ├── promote/
│   ├── recall/
│   └── markdown/
├── templates/
│   └── default-memory-repo/
├── docs/
├── go.mod
└── README.md
```

Module responsibilities:

- `cli/`: command dispatch, argument parsing, and error output.
- `repo/`: memory repo discovery, path validation, config reading, and directory creation.
- `templates/`: embed default templates and write them to the target directory.
- `scan/`: sensitive information rules, scan results, and redacted output.
- `capture/`: read Markdown, scan, and write to inbox or quarantine.
- `promote/`: validate target paths, scan, move files, and update status.
- `recall/`: keyword matching, candidate file selection, and context pack generation.
- `markdown/`: title extraction, front matter read/write, and slug generation.

## 4. Implementation Phases

### 4.1 Phase One: Project Skeleton

Tasks:

- Initialize the Go module.
- Create the CLI entry point.
- Support `knowblazer --help`.
- Support command dispatch for `init`, `scan`, `capture`, `promote`, and `recall`.
- Add the basic test structure.

Acceptance:

- `go test ./...` passes.
- Unimplemented commands return a clear error or help output.

### 4.2 Phase Two: Repo and Init

Tasks:

- Implement `.knowblazer/config.json` recognition.
- Implement repo lookup order:
  1. Current directory.
  2. Parent directories.
  3. `KNOWBLAZER_REPO`.
  4. Default path.
- Generate the memory repository from `templates/default-memory-repo/`.
- Ensure init is idempotent and does not overwrite existing files.
- Fail by default for non-empty directories that are not Knowblazer repos.

Acceptance:

- Init in an empty directory creates the full structure.
- Re-running init does not overwrite user content.
- Init fails with a clear message for a non-empty ordinary directory.

### 4.3 Phase Three: Secret Scan

Tasks:

- Implement file and directory scanning.
- Recursively scan Markdown, text, and env-like files in directories.
- Implement rules for:
  - private key blocks
  - token/api_key/secret fields
  - password fields
  - database URLs
  - common cloud secret env variables
- Implement risk levels: `clean`, `warning`, and `high`.
- Output line number, rule name, and redacted snippet.
- Avoid printing full secrets.

Acceptance:

- Mock private keys are detected.
- Mock database URLs are detected.
- Mock passwords are detected.
- Output is redacted.
- Clean results return exit code 0, risky results return exit code 1, and command errors return exit code 2.

### 4.4 Phase Four: Markdown Utilities

Tasks:

- Extract first-level headings.
- Generate slugs from filenames.
- Generate front matter.
- Update front matter status fields.
- Keep body content readable and avoid damaging the original text.

Acceptance:

- Files without headings still get readable titles.
- Files with existing front matter are not destructively rewritten.
- Generated Markdown is human-readable.

### 4.5 Phase Five: Capture

Tasks:

- Accept only `.md` and `.markdown` files.
- Run scan before capture.
- Write clean or warning files to `inbox/YYYY-MM-DD/`.
- Write high-risk files to `quarantine/YYYY-MM-DD/`.
- Use destination filenames in the form `YYYYMMDD-HHMMSS-<slug>.md`.
- Write front matter.
- Do not automatically commit or push.

Acceptance:

- Clean Markdown enters inbox.
- High-risk Markdown enters quarantine.
- Output clearly shows the saved path and scan result.
- The source file is not modified.

### 4.6 Phase Six: Promote

Tasks:

- Support `promote <file> --to <target>`.
- Re-scan the source file.
- Block promotion for high-risk files.
- Allow targets only under `experience/`, `projects/`, or `profile/`.
- Block path escape.
- Do not overwrite existing targets by default.
- Update front matter: `status`, `promoted_at`, and `promoted_to`.

Acceptance:

- Inbox files can be promoted to `experience/deployment/`.
- High-risk files cannot be promoted.
- `--to ../../x` is blocked.
- Existing targets are not overwritten.

### 4.7 Phase Seven: Recall

Tasks:

- Support `recall --task <text>`.
- Read:
  - `profile/preferences.md`
  - `profile/decision-principles.md`
  - `projects/<project>.md`
  - matching `experience/**/*.md`
  - today's and yesterday's `daily/*.md`
- Exclude `inbox/`, `quarantine/`, and `recall/` by default.
- Implement simple keyword matching and file ranking.
- Generate a Markdown recall pack.
- Support `--output <file>`.
- Control output length.

Acceptance:

- Missing task fails.
- Output includes task and generated time.
- Output excludes inbox and quarantine by default.
- Matching experience does not concatenate the whole repository.
- `--output` writes a file.

### 4.8 Phase Eight: End-to-End Verification

Tasks:

- Simulate the full path in a temporary directory.
- Verify README command examples.
- Add fixtures for:
  - clean lesson
  - secret lesson
  - project context
  - experience note
  - daily note

Acceptance:

- `go test ./...` passes.
- Manual MVP flow succeeds.
- The tool runs without network access.

## 5. Test Strategy

P0 should primarily use unit tests and temporary-directory integration tests.

Required tests:

- Repo discovery and path safety.
- Init idempotency.
- Scan rules and redaction.
- Capture destination selection.
- Promote target restrictions.
- Recall exclusion rules.

Tests should use temporary directories, not depend on the user's real home directory, and not read or write a real Git remote.

## 6. Path Safety Requirements

All writes must stay inside the Knowblazer repo unless the user explicitly specifies a source file to read.

The implementation must prevent:

- `../` path escape.
- Writing outside the repo through symlinks.
- Promotion to `.knowblazer/`.
- Promotion to `quarantine/` or `recall/`.
- Recall reading from `quarantine/`.

## 7. Output Style

CLI output should be concise, clear, and script-friendly.

Recommended:

- Print saved paths on success.
- Print matched rules and redacted snippets for risks.
- Print next-step guidance for errors.

Avoid:

- Printing full secrets.
- Printing large blocks of unrelated explanation.
- Automatically running Git operations.

## 8. Not Implemented in P0

P0 does not implement:

- Git commit/push.
- Git remote configuration.
- SpecStory import.
- AI tool hooks.
- MCP server.
- SQLite or vector indexing.
- LLM summarization.
- Web UI.
- Background jobs.

## 9. Suggested Milestones

Suggested commit order:

1. Go module and CLI skeleton.
2. Repo discovery and init from templates.
3. Secret scanner.
4. Capture command.
5. Promote command.
6. Recall command.
7. End-to-end tests and README command alignment.

Each milestone should keep `go test ./...` passing.
