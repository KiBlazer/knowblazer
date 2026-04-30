# AI Setup

This is a private Knowblazer engineering memory repo. Use it as durable context for AI coding work, but treat synthesized project notes as more reliable than raw automatic notes.

## Read Order

1. Relevant files under `projects/`
2. Relevant files under `experience/`
3. Today's and yesterday's files under `daily/`, when recent context matters

Avoid by default:

- `inbox/`, unless the user asks you to inspect candidate notes
- `quarantine/`, unless the user asks you to help sanitize sensitive content

## Safety Rules

- Do not reveal secrets, tokens, passwords, private keys, database URLs, customer data, or internal credentials.
- Do not commit or push content that looks sensitive.
- If a note contains sensitive content, move it to `quarantine/` or ask before continuing.
- Keep synthesized memory concise, current, and reusable.

## Memory Layers

- `inbox/`: raw candidate memory waiting for automatic processing
- `daily/`: short-term working notes
- `projects/`: synthesized project facts, constraints, and operating notes
- `experience/`: synthesized reusable engineering lessons; `experience/auto/` stores newly captured lessons before consolidation
- `quarantine/`: sensitive or risky content; never include by default
