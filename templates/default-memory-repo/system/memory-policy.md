# Memory Policy

Knowblazer memory is stored as Markdown in a user-owned private Git repo.

Markdown and Git are the source of truth. Indexes, caches, vector databases, MCP servers, or other integrations may be added later, but they must not replace the human-readable memory repo.

## Memory Layers

### inbox/

Raw candidate notes. Content here has not been reviewed and should not be treated as durable memory.

Default recall behavior: exclude.

### daily/

Short-term working notes. Use for today's context, temporary observations, and unresolved work.

Default recall behavior: include today and yesterday when relevant.

### profile/

Stable developer preferences, principles, and working style.

Default recall behavior: include concise relevant sections.

### projects/

Durable project context, architecture notes, deployment notes, and project-specific cautions.

Default recall behavior: include the current project's file when known.

### experience/

Reusable engineering lessons and troubleshooting notes.

Default recall behavior: include relevant files by task keywords.

### quarantine/

Sensitive or risky content. Files here require manual review.

Default recall behavior: never include.

### recall/

Generated task context packs.

Default recall behavior: exclude unless explicitly requested.

## Promotion

Long-term memory should be promoted from `inbox/` or `daily/` only after review.

Promotion targets:

- `profile/`
- `projects/`
- `experience/`

Do not promote content with unresolved secrets or sensitive data.
