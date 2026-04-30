# Knowblazer MVP Specification

> Status: Historical P0 specification from 2026-04-29. It documents the original init/capture/scan/promote/recall path. For current dynamic-memory positioning and requirement evolution, see [Engineering positioning](../../current/positioning.md) and [Requirements history](../requirements-history.md).

Date: 2026-04-29

## 1. Goal

This document turns the P0 scope from `requirements.md` into an implementable and testable MVP specification.

The MVP validates one core path:

```text
initialize private engineering memory repo
  ↓
save one Markdown lesson
  ↓
scan for sensitive information
  ↓
promote into long-term engineering memory
  ↓
generate recall context for one AI coding task
```

The MVP does not aim for maximum automation or intelligent search. It prioritizes validating whether private Git + Markdown + layered memory + safe recall is useful in practice.

## 2. MVP Commands

P0 commands are:

```bash
knowblazer init [path]
knowblazer capture <file> [--repo <path>]
knowblazer scan <path> [--repo <path>]
knowblazer promote <file> --to <target> [--repo <path>]
knowblazer recall --task <text> [--repo <path>] [--project <name>] [--output <file>]
```

`--repo` explicitly specifies the Knowblazer memory repository path. When omitted, the tool searches in this order:

1. The current directory, if it is a Knowblazer memory repo.
2. Parent directories that contain a Knowblazer repo marker.
3. The `KNOWBLAZER_REPO` environment variable.
4. The default path under the user's home directory: `~/knowblazer-notes`.

## 3. Memory Repository Structure

`init` creates this structure by default:

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

`.knowblazer/config.json` is the repository marker and base configuration file. It does not store sensitive information.

Recommended initial content:

```json
{
  "version": 1,
  "created_by": "knowblazer",
  "memory_repo": true
}
```

## 4. File Metadata

Markdown files generated or moved by Knowblazer should include simple front matter when practical, so both humans and tools can understand them.

Recommended format:

```markdown
---
title: "Deploy frontend rollback lesson"
type: "inbox"
status: "candidate"
source: "/path/to/original.md"
captured_at: "2026-04-29T12:30:00+08:00"
---

# Deploy frontend rollback lesson

...
```

The MVP does not require a complex schema, but should ensure:

- The title is readable.
- The source is traceable.
- The status is clear.
- The timestamp is explicit.

## 5. `init`

### 5.1 Usage

```bash
knowblazer init [path]
```

Example:

```bash
knowblazer init ~/knowblazer-notes
```

### 5.2 Behavior

`init` should:

- Create the memory repository root directory.
- Create the first-version minimal directory structure.
- Create `.knowblazer/config.json`.
- Create `AI-SETUP.md`.
- Create `system/memory-policy.md`.
- Create `system/privacy-policy.md`.
- Create `profile/preferences.md`.
- Create `profile/decision-principles.md`.

### 5.3 Idempotency

When re-running `init`:

- Do not overwrite files the user has modified.
- Fill in missing directories.
- Create missing template files.
- If the target directory is non-empty and is not a Knowblazer memory repo, fail or ask for confirmation. The MVP may fail by default to avoid accidental writes.

### 5.4 Output

Successful output should include:

```text
Initialized Knowblazer memory repo: <path>
Next steps:
  1. Edit AI-SETUP.md
  2. Capture a lesson with: knowblazer capture <file> --repo <path>
  3. Generate recall with: knowblazer recall --task "<task>" --repo <path>
```

### 5.5 Acceptance

- Running init in an empty directory creates the full minimal structure.
- Re-running init does not overwrite existing content.
- A non-empty ordinary directory is not silently modified.

## 6. `scan`

### 6.1 Usage

```bash
knowblazer scan <path> [--repo <path>]
```

`path` may be a file or directory.

### 6.2 Scan Scope

The MVP should at least detect these high-risk patterns:

- Private key blocks, such as `-----BEGIN PRIVATE KEY-----`.
- Common token fields, such as `token=...`, `api_key=...`, and `secret=...`.
- Password fields, such as `password=...`.
- Database connection strings, such as `mysql://`, `postgres://`, and `mongodb://`.
- `.env`-style secret variables, such as `AWS_SECRET_ACCESS_KEY=...`.

The MVP can use rule matching first and does not need an external secret scanning service.

### 6.3 Result Levels

The MVP uses three levels:

```text
clean       no obvious risk found
warning     suspicious content found, but not necessarily a secret
high        high-risk sensitive information found
```

### 6.4 Output

Scan output should include:

- File path.
- Risk level.
- Rule name.
- Line number.
- Redacted snippet.

Do not print complete secret values.

Example:

```text
HIGH  deploy.md:12  database-url  postgres://user:****@host/db
```

### 6.5 Exit Codes

Recommended:

```text
0  clean
1  warning or high
2  command error
```

The implementation may later distinguish warning and high if needed.

### 6.6 Acceptance

- Mock private keys are detected.
- Mock database URLs are detected.
- Mock passwords are detected.
- Output redacts secrets.

## 7. `capture`

### 7.1 Usage

```bash
knowblazer capture <file> [--repo <path>]
```

### 7.2 Input Requirements

The MVP only supports Markdown files.

If the input is not Markdown, fail with:

```text
Only Markdown files are supported in MVP.
```

### 7.3 Behavior

`capture` should:

1. Read the input file.
2. Run `scan`.
3. If the result is `clean` or `warning`, copy it to `inbox/YYYY-MM-DD/`.
4. If the result is `high`, copy it to `quarantine/YYYY-MM-DD/`.
5. Write or supplement front matter.
6. Print the saved path and scan result.

### 7.4 Naming

The destination filename should be readable and stable.

Recommended rule:

```text
YYYYMMDD-HHMMSS-<slug>.md
```

`slug` can come from the source filename or first-level heading.

### 7.5 No Automatic Commit

MVP `capture` does not automatically run `git commit` or `git push`.

If a sync helper is added later, it must scan before allowing commit or push.

### 7.6 Output

For clean or warning:

```text
Captured to inbox: <repo>/inbox/2026-04-29/20260429-123000-deploy-lesson.md
Scan result: clean
```

For high risk:

```text
Sensitive content detected.
Moved to quarantine: <repo>/quarantine/2026-04-29/20260429-123000-deploy-lesson.md
Review and sanitize before promoting or committing.
```

### 7.7 Acceptance

- Markdown files enter `inbox/YYYY-MM-DD/`.
- High-risk Markdown files enter `quarantine/YYYY-MM-DD/`.
- Capture output clearly tells the user what happened next.
- Capture does not automatically commit or push.

## 8. `promote`

### 8.1 Usage

```bash
knowblazer promote <file> --to <target> [--repo <path>]
```

Examples:

```bash
knowblazer promote inbox/2026-04-29/deploy.md --to experience/deployment
knowblazer promote inbox/2026-04-29/preferences.md --to profile/preferences.md
knowblazer promote inbox/2026-04-29/project.md --to projects/kiblazer.md
```

### 8.2 Allowed Targets

The MVP only allows promotion to:

- `experience/`
- `projects/`
- `profile/`

It must not allow promotion to:

- `quarantine/`
- `recall/`
- `.knowblazer/`
- paths outside the memory repository

### 8.3 Behavior

`promote` should:

1. Confirm the source file exists.
2. Re-run `scan` on the source file.
3. Block promotion if the result is `high`.
4. Confirm the target path is in an allowed area.
5. Move or copy the file to the target path.
6. Update front matter:
   - `status: "promoted"`
   - `promoted_at: "<timestamp>"`
   - `promoted_to: "<target>"`

The MVP may use move semantics by default. A later version can add `--copy` if preserving the inbox source is needed.

### 8.4 Target Conflicts

If the target file already exists, do not overwrite it by default.

The user can be asked to choose a different name. The MVP does not need `--force`.

### 8.5 Acceptance

- Clean files can be promoted to `experience/`.
- High-risk files cannot be promoted.
- The target path cannot escape the memory repository.
- Existing targets are not overwritten.

## 9. `recall`

### 9.1 Usage

```bash
knowblazer recall --task <text> [--repo <path>] [--project <name>] [--output <file>]
```

Example:

```bash
knowblazer recall --task "deploy new frontend" --project kiblazer
```

### 9.2 Input

Required:

- `--task`

Optional:

- `--project`
- `--repo`
- `--output`

### 9.3 Candidate Sources

MVP recall uses these sources:

1. `profile/preferences.md`
2. `profile/decision-principles.md`
3. `projects/<project>.md`, if `--project` is provided
4. Files under `experience/**/*.md` that match task keywords
5. Today's and yesterday's `daily/*.md`

Default exclusions:

- `inbox/`
- `quarantine/`
- `recall/`

### 9.4 Matching Rules

The MVP uses simple keyword matching:

- Extract English words, numbers, and continuous Chinese text segments from `--task`.
- Filename matches have weight.
- Title matches have weight.
- Body matches contribute score.
- Prefer the highest-scoring files.

Semantic search and embeddings are not required.

### 9.5 Output Format

Output is Markdown:

```markdown
# Knowblazer Recall Pack

Task: deploy new frontend
Generated at: 2026-04-29T12:30:00+08:00

## Developer Preferences

...

## Project Context

...

## Relevant Experience

...

## Recent Daily Notes

...

## Cautions

- This pack excludes inbox and quarantine by default.
- Verify commands and secrets before running anything.
```

### 9.6 Length Control

The MVP should enforce a default length limit.

Suggested limits:

- At most 5 experience files.
- At most the first 120 lines or 8KB per file.
- Total output around 20KB by default.

Specific values can be adjusted during implementation, but the tool must avoid concatenating the whole memory repository.

### 9.7 Output Location

Default output goes to stdout.

If `--output <file>` is provided, write to that file. Users may write to `recall/`, but this is not required.

### 9.8 Acceptance

- Missing `--task` fails.
- Output includes title, task, and generated time.
- Output excludes `inbox/` and `quarantine/` by default.
- Matching experience does not concatenate the whole repository.
- `--output` writes a file.

## 10. Error Handling

MVP error messages should directly explain the problem and next step.

Typical errors:

```text
Knowblazer repo not found. Run `knowblazer init <path>` or set KNOWBLAZER_REPO.
```

```text
Target path is outside the Knowblazer repo and is not allowed.
```

```text
Sensitive content detected. File was not promoted.
```

```text
Only Markdown files are supported in MVP.
```

## 11. Safety Constraints

The MVP must follow these constraints:

- Do not upload user memory to any official Knowblazer service.
- Do not require an official account.
- Do not automatically commit.
- Do not automatically push.
- Do not print full secrets in scan output.
- Do not generate recall from `quarantine/`.
- Do not include `inbox/` in recall by default.

## 12. Implementation-Independent Requirements

This document does not require a specific programming language.

The implementation should:

- Handle paths across platforms.
- Be friendly to Git repositories.
- Not require network access for P0.
- Use temporary directories in tests.
- Not require an external LLM API.

## 13. P0 Test Checklist

Minimum tests should cover:

- `init` creates directories and templates.
- `init` is idempotent and does not overwrite existing files.
- `capture` sends clean files to inbox.
- `capture` sends high-risk files to quarantine.
- `scan` detects mock secrets.
- `scan` redacts output.
- `promote` sends clean files to long-term directories.
- `promote` blocks high-risk files.
- `promote` blocks path escape.
- `recall` outputs Markdown.
- `recall` excludes inbox and quarantine.
- `recall --output` writes a file.

## 14. Later Non-P0 Work

These are outside the MVP specification:

- Automatic Git commit/push.
- Official cloud service.
- SpecStory import.
- AI tool hook installation.
- MCP server.
- Vector search.
- Background automatic promotion.
- Web UI.

They can be specified separately after P0 is validated.
