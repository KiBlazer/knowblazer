# Knowblazer Security Model & Secrets Scanner

Knowblazer is built on a **local-first, privacy-by-default** model. Because long-term memories generated during AI coding can easily pick up API keys, passwords, and private tokens, Knowblazer implements a proactive scanner that runs before any memory is written or indexed.

---

## 1. Safety Scan & Quarantine Flow

Whenever a memory is captured via the CLI (`knowblazer remember`) or via the MCP server (`knowblazer_remember`, `knowblazer_capture`), it undergoes the following safety checks:

```
Captured Note / Text
       │
       ▼
 ┌───────────┐
 │  Scanner  │ ──(High-Risk Match)──► [ Quarantine Folder ]
 └─────┬─────┘                        (Excluded from recall, index, & sync)
       │ (Clean)
       ▼
[ Experience / Daily Note ]
(Included in indexing & recall)
```

1.  **High-Risk Scan:** The memory is matched against a set of regex patterns (see Rules below).
2.  **Quarantine:** If a high-risk finding is detected, the note is saved inside the `quarantine/` directory of your vault instead of `experience/` or `daily/`.
3.  **Default Exclusion:** Quarantined memories are strictly excluded from:
    *   Lexical search index (`index.json`)
    *   Recall pack generation (`knowblazer_recall` / `knowblazer_context`)
    *   Git push validation (`knowblazer sync` checks quarantine folder)

---

## 2. Default Scanner Rules

The scanner looks for 6 major categories of sensitive credentials. The matching rules are case-insensitive and defined as follows:

| Rule Name | Target / Pattern | Mitigation / Redaction in Status Snippet |
| :--- | :--- | :--- |
| **`private-key`** | PEM formatting private key markers (e.g. `-----BEGIN RSA PRIVATE KEY-----`) | Redacts the entire line |
| **`database-url`** | Connection URIs for MySQL, PostgreSQL, and MongoDB containing credentials | Redacts the password component (keeps host/port) |
| **`secret-field`** | Assignments to fields containing `token`, `secret`, `api_key`, `apikey`, or `password` | Redacts the value assigned to the variable |
| **`cloud-secret-env`**| Cloud credential environment variables (`AWS_SECRET_ACCESS_KEY`, `GOOGLE_APPLICATION_CREDENTIALS`, etc.) | Redacts the value assigned |
| **`bearer-token`** | Bearer tokens in headers or string definitions | Redacts the token signature |
| **`openai-style-key`**| API keys starting with `sk-` | Redacts the key signature (replaces with `sk-****`) |

---

## 3. Reviewing and Promoting Quarantined Files

If a file was quarantined false-positively (e.g. mock tokens in a test file), you can review and manually move them:

1.  Inspect the file contents in `quarantine/` manually using your editor or terminal.
2.  Redact or remove any genuine secrets.
3.  Once clean, move it to the `inbox/` or `experience/auto/` folders manually, or run the promote command:
    ```bash
    knowblazer review promote quarantine/2026-06-12/your-note.md --to experience/manual
    ```
