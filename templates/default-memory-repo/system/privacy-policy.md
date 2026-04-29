# Privacy Policy

This is a local policy for this Knowblazer memory repo.

Knowblazer should not upload memory to any official Knowblazer service. Remote sync, if used, must go through a Git remote chosen and controlled by the user.

## Default Behavior

- Store memory locally.
- Do not require a Knowblazer account.
- Do not require network access for core workflows.
- Do not automatically commit.
- Do not automatically push.
- Scan before capture, promotion, commit, or push.

## Sensitive Content

Treat the following as sensitive:

- API tokens
- passwords
- private keys
- database URLs
- `.env` values
- cloud provider credentials
- customer data
- internal infrastructure credentials
- unreleased proprietary information

## Quarantine Rules

Content with high-risk sensitive data should go to `quarantine/`.

Files in `quarantine/`:

- must not be included in default recall
- must not be automatically committed
- must not be automatically pushed
- require manual review before reuse

## Redaction

When summarizing or moving memory, redact sensitive values. Keep enough context to preserve the lesson, but remove credentials and private data.
