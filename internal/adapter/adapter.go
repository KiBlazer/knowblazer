package adapter

import (
	"fmt"
	"strings"
)

var supported = map[string]string{
	"claude": `# Claude Code Knowblazer Adapter

Recommended one-command setup:

    knowblazer setup claude --repo "<repo>"

This configures Claude Code MCP and writes project instructions. If MCP is unavailable, fall back to:

    knowblazer recall --task "<task>" --repo "<repo>"

Use the recall pack as task context. Do not include quarantine content.
`,
	"codex": `# Codex Knowblazer Adapter

Recommended one-command setup:

    knowblazer setup codex --repo "<repo>"

This configures Codex MCP and writes project instructions to AGENTS.md. If MCP is unavailable, fall back to:

    knowblazer recall --task "<task>" --repo "<repo>"

Use the resulting Markdown as task context. Do not include quarantine content.
`,
	"gemini": `# Gemini CLI Knowblazer Adapter

Use Knowblazer recall packs as local context:

    knowblazer recall --task "<task>" --repo "<repo>"
`,
	"cursor": `# Cursor Knowblazer Adapter

Add this workflow to project rules:

    Run knowblazer recall --task "<task>" --repo "<repo>" and use the Markdown as task context.
`,
	"antigravity": `# Antigravity Knowblazer Adapter

Add this to your ~/.gemini/antigravity/mcp_config.json:

{
  "mcpServers": {
    "knowblazer": {
      "command": "knowblazer",
      "args": ["mcp", "serve", "--repo", "<repo>"]
    }
  }
}

If MCP is unavailable, you can manually generate recall packs for tasks:

    knowblazer recall --task "<task>" --repo "<repo>"
`,
}

func Generate(tool string, repoRoot string) ([]byte, error) {
	t := strings.ToLower(tool)
	if t == "agy" {
		t = "antigravity"
	}
	template, ok := supported[t]
	if !ok {
		return nil, fmt.Errorf("unsupported adapter: %s", tool)
	}
	return []byte(strings.ReplaceAll(template, "<repo>", repoRoot)), nil
}

func Tools() []string {
	return []string{"claude", "codex", "gemini", "cursor", "antigravity", "agy"}
}
