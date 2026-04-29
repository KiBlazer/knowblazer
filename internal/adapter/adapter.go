package adapter

import (
	"fmt"
	"strings"
)

var supported = map[string]string{
	"claude": `# Claude Code Knowblazer Adapter

Before starting a task, run:

    knowblazer recall --task "<task>" --repo "<repo>"

Paste the recall pack into the coding session. Do not include quarantine content.
`,
	"codex": `# Codex Knowblazer Adapter

Generate task context with:

    knowblazer recall --task "<task>" --repo "<repo>"

Use the resulting Markdown as local project memory.
`,
	"gemini": `# Gemini CLI Knowblazer Adapter

Use Knowblazer recall packs as local context:

    knowblazer recall --task "<task>" --repo "<repo>"
`,
	"cursor": `# Cursor Knowblazer Adapter

Add this workflow to project rules:

    Run knowblazer recall --task "<task>" --repo "<repo>" and use the Markdown as task context.
`,
}

func Generate(tool string, repoRoot string) ([]byte, error) {
	template, ok := supported[strings.ToLower(tool)]
	if !ok {
		return nil, fmt.Errorf("unsupported adapter: %s", tool)
	}
	return []byte(strings.ReplaceAll(template, "<repo>", repoRoot)), nil
}

func Tools() []string {
	return []string{"claude", "codex", "gemini", "cursor"}
}
