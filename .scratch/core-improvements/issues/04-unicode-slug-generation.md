# 04: Generate semantic slugs for non-ASCII note titles

**What to build:** Slugs generated for captured notes must preserve non-ASCII alphanumeric characters (such as Chinese characters) instead of stripping them all and falling back to `note.md`.

**Blocked by:** None (can start immediately)

**Status:** ready-for-agent

- [x] Non-ASCII characters (letters/digits/CJK) are retained during slug generation
- [x] Titles like `# 微信小程序支付配置` generate readable filenames like `2026-xx-xx-xxxxxx-微信小程序支付配置.md`
- [x] Punctuation and symbols are converted to hyphens with consecutive hyphens squashed
- [x] `go test ./internal/capture` passes
