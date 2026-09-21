# Spec: Knowblazer Core Stability, Retrieval, and Formatting Improvements

Status: ready-for-agent

## Problem Statement

Users of Knowblazer experience several functional and stability limitations during daily AI coding sessions:

1. **MCP Connection Drops on Large Content**: When saving long debugging logs or large markdown files via `knowblazer_remember` or `knowblazer_capture`, the MCP server process silently terminates because of default input line buffer limits.
2. **Retrieval Misses for Chinese and Non-spaced Tasks**: When tasks or questions are written in Chinese or other non-whitespace-separated languages, keyword-based scoring fails to segment terms, leading to 0 matching score and completely missing relevant experience files.
3. **Quarantine Directory Blocks Git Operations**: When sensitive data is correctly isolated into `quarantine/`, running sync operations (like `knowblazer sync commit` or `push`) fails because the security scan inspects the quarantine folder and flags high-severity items, permanently deadlocking Git synchronization.
4. **Truncation of Multi-paragraph Knowledge**: When automatic memories are consolidated, content after the very first empty line is completely discarded. Crucial context such as step-by-step solutions, multi-paragraph post-mortems, and code blocks are lost.
5. **Loss of File Name Semantics for Non-ASCII Titles**: Notes with non-ASCII titles (such as Chinese, Japanese, or accented characters) have their titles stripped to empty strings, defaulting to generic names like `note.md`, making filesystem navigation opaque.
6. **Incomplete Agent Configuration Detection**: The system status command only checks for `CLAUDE.md`, ignoring other modern agent instruction files like `AGENTS.md` and `GEMINI.md`.

## Solution

1. **Expand MCP Input Buffer**: Allow large JSON-RPC messages and memory payloads up to 10MB without terminating the scanner.
2. **N-gram / Character-aware Term Splitting**: Support CJK/Unicode character sequencing in recall and search indexing, allowing Chinese and non-spaced tasks to correctly score and retrieve relevant notes.
3. **Exclude Quarantine from Git Sync Scans**: Ensure that security scans run by sync operations explicitly ignore `quarantine/`, preserving the safety barrier while unblocking version control.
4. **Paragraph-preserving Consolidation**: Remove premature termination on blank lines in bounded body extraction, relying on total character and line budgets instead.
5. **Unicode-safe Slug Generation**: Preserve alphanumeric Unicode characters in filenames so non-English titles retain human-readable slugs.
6. **Broaden Agent Configuration Detection**: Recognize `AGENTS.md` and `GEMINI.md` alongside `CLAUDE.md` in workspace status reporting.

## User Stories

1. As an engineer capturing large error stack traces and verbose logs into memory, I want the MCP server to handle payloads up to 10MB without dropping the connection.
2. As a Chinese-speaking developer, I want to describe my coding task in Chinese and have Knowblazer retrieve relevant experience notes that match keywords in the task.
3. As a developer searching the local Knowblazer index using Chinese terms, I want to find matching documents even when the query contains no spaces.
4. As an engineer whose sensitive information was safely redirected to the `quarantine` directory, I want to run `knowblazer sync commit` without the operation being blocked by files inside quarantine.
5. As a developer writing multi-paragraph technical lessons with explanations and code snippets, I want the synthesized memory to preserve all paragraphs up to the size limit rather than stopping at the first empty line.
6. As an engineer capturing markdown notes with non-ASCII titles (e.g., `# 微信小程序支付接入`), I want the resulting filename to reflect the title rather than falling back to `note.md`.
7. As a team member using `AGENTS.md` or `GEMINI.md` in my workspace, I want `knowblazer status` to recognize that my workspace has been properly configured for AI coding agents.

## Implementation Decisions

1. **MCP Buffer Expansion**:
   - The JSON-RPC scanner in the MCP module will configure a custom buffer with initial size 64KB and maximum capacity 10MB.
   - Any scanner errors will be caught gracefully and logged rather than terminating abruptly.

2. **Multilingual Token Extraction**:
   - In both recall scoring and index building, the term extraction logic will distinguish between Latin-style alphanumeric words and CJK character sequences.
   - For CJK runs, character-level unigrams and adjacent bigrams will be indexed, enabling substring and concept matching without requiring heavy external dictionaries or CGO dependencies.

3. **Quarantine Directory Bypassing in Scanner**:
   - The directory scanning logic will treat `quarantine` as an ignored directory during directory traversal, alongside `.git`, `node_modules`, `.idea`, and `.vscode`.
   - Sync pre-commit and pre-push validations will verify only the repository content that is intended for source tracking.

4. **Multi-paragraph Consolidation**:
   - The bounded body extractor in consolidation will treat empty lines as paragraph dividers rather than break signals.
   - Extraction will respect an overall maximum character budget and line count budget, appending an explicit truncation notice only when the overall budget is exceeded.

5. **Unicode Slug Generation**:
   - The slug generation algorithm will retain Unicode letters and digits while normalizing whitespace, punctuation, and symbols into standard hyphens.
   - Consecutive hyphens will be squashed and edge hyphens trimmed.

6. **Workspace Status Multi-file Inspection**:
   - Workspace configuration checking will look for marker tokens in `CLAUDE.md`, `AGENTS.md`, and `GEMINI.md`.
   - If any of these files exist and contain a valid Knowblazer setup block, the workspace is reported as configured.

## Testing Decisions

- **Good Test Criteria**: Tests must verify observable behavior (JSON-RPC responses, recall pack contents, file creation names, exit codes, and commit statuses) rather than private helper functions.
- **Modules Under Test**:
  - `internal/mcp`: Test handling of >64KB JSON-RPC messages and workspace status detection.
  - `internal/recall` and `internal/index`: Test search and recall with Chinese queries and experience notes.
  - `internal/scan` and `internal/sync`: Test that having files in `quarantine/` does not block `sync.Commit`.
  - `internal/consolidate`: Test that multi-paragraph notes with empty lines and code blocks retain all sections in synthesized memory.
  - `internal/capture`: Test slug generation with Chinese and mixed-language titles.
- **Prior Art**:
  - Existing `recall_test.go`, `index_test.go`, `consolidate_test.go`, `scan_test.go`, and `mcp_test.go` provide table-driven test patterns and filesystem sandboxes using `t.TempDir()`.

## Out of Scope

- Integrating heavy external NLP segmenters (e.g., jieba) or adding external CGO dependencies.
- Automatic semantic translation of filenames.
- Cloud-based vector embeddings or external remote model calls for consolidation.

## Further Notes

All changes remain strictly within the standard Go 1.22 library with zero third-party dependencies, preserving Knowblazer's lightweight, portable architecture.
