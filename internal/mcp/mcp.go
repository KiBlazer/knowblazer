package mcp

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/knowblazer/knowblazer/internal/capture"
	"github.com/knowblazer/knowblazer/internal/consolidate"
	"github.com/knowblazer/knowblazer/internal/daily"
	"github.com/knowblazer/knowblazer/internal/index"
	"github.com/knowblazer/knowblazer/internal/projectmap"
	"github.com/knowblazer/knowblazer/internal/recall"
	"github.com/knowblazer/knowblazer/internal/review"
	"github.com/knowblazer/knowblazer/internal/scan"
)

type Request struct {
	JSONRPC string         `json:"jsonrpc,omitempty"`
	ID      any            `json:"id,omitempty"`
	Method  string         `json:"method"`
	Params  map[string]any `json:"params,omitempty"`
}

type Response struct {
	JSONRPC string `json:"jsonrpc,omitempty"`
	ID      any    `json:"id,omitempty"`
	Result  any    `json:"result,omitempty"`
	Error   string `json:"error,omitempty"`
}

func Serve(repoRoot string, in io.Reader, out io.Writer) error {
	scanner := bufio.NewScanner(in)
	encoder := json.NewEncoder(out)
	for scanner.Scan() {
		var req Request
		if err := json.Unmarshal(scanner.Bytes(), &req); err != nil {
			_ = encoder.Encode(Response{JSONRPC: "2.0", Error: err.Error()})
			continue
		}
		resp := handle(repoRoot, req)
		if resp == nil {
			continue
		}
		if err := encoder.Encode(resp); err != nil {
			return err
		}
	}
	return scanner.Err()
}

func handle(repoRoot string, req Request) *Response {
	switch req.Method {
	case "initialize":
		return ok(req.ID, map[string]any{
			"protocolVersion": "2024-11-05",
			"capabilities": map[string]any{
				"tools": map[string]any{},
			},
			"serverInfo": map[string]any{
				"name":    "knowblazer",
				"version": "0.1.0",
			},
		})
	case "notifications/initialized":
		return nil
	case "ping":
		return ok(req.ID, map[string]any{})
	case "tools/list":
		return ok(req.ID, map[string]any{"tools": toolDefinitions()})
	case "tools/call":
		return callTool(repoRoot, req)
	case "knowblazer_context":
		return directToolResult(repoRoot, req.ID, req.Method, req.Params)
	case "knowblazer_remember":
		return directToolResult(repoRoot, req.ID, req.Method, req.Params)
	case "knowblazer_status":
		return directToolResult(repoRoot, req.ID, req.Method, req.Params)
	case "knowblazer_recall":
		return directToolResult(repoRoot, req.ID, req.Method, req.Params)
	case "knowblazer_search":
		return directToolResult(repoRoot, req.ID, req.Method, req.Params)
	case "knowblazer_capture":
		return directToolResult(repoRoot, req.ID, req.Method, req.Params)
	case "knowblazer_consolidate":
		return directToolResult(repoRoot, req.ID, req.Method, req.Params)
	default:
		return fail(req.ID, fmt.Sprintf("unknown method: %s", req.Method))
	}
}

func ok(id any, result any) *Response {
	return &Response{JSONRPC: "2.0", ID: id, Result: result}
}

func fail(id any, message string) *Response {
	return &Response{JSONRPC: "2.0", ID: id, Error: message}
}

func callTool(repoRoot string, req Request) *Response {
	name := stringParam(req.Params, "name")
	args := mapParam(req.Params, "arguments")
	if name == "" {
		return fail(req.ID, "tool name is required")
	}
	result := directToolResult(repoRoot, req.ID, name, args)
	if result == nil || result.Error != "" {
		return result
	}
	text, ok := result.Result.(string)
	if !ok {
		data, err := json.MarshalIndent(result.Result, "", "  ")
		if err != nil {
			return fail(req.ID, err.Error())
		}
		text = string(data)
	}
	return okResponse(req.ID, text)
}

func okResponse(id any, text string) *Response {
	return ok(id, map[string]any{
		"content": []map[string]string{
			{"type": "text", "text": text},
		},
	})
}

func directToolResult(repoRoot string, id any, name string, params map[string]any) *Response {
	switch name {
	case "knowblazer_context":
		task := stringParam(params, "task")
		project := stringParam(params, "project")
		if project == "" {
			workspace := stringParam(params, "workspace")
			if workspace == "" {
				workspace, _ = os.Getwd()
			}
			if workspace != "" {
				if mapped, ok, err := projectmap.Resolve(repoRoot, workspace); err == nil && ok {
					project = mapped
				}
			}
		}
		pack, err := recall.Generate(repoRoot, recall.Options{Task: task, Project: project})
		if err != nil {
			return fail(id, err.Error())
		}
		return ok(id, string(pack))
	case "knowblazer_remember":
		text := strings.TrimSpace(stringParam(params, "text"))
		if text == "" {
			return fail(id, "text is required")
		}
		if boolParam(params, "daily") {
			path, err := daily.Add(repoRoot, daily.AddOptions{Text: text})
			if err != nil {
				return fail(id, err.Error())
			}
			return ok(id, map[string]any{"path": path, "status": "daily"})
		}
		result, err := rememberText(repoRoot, text)
		if err != nil {
			return fail(id, err.Error())
		}
		return ok(id, result)
	case "knowblazer_status":
		status, err := status(repoRoot, stringParam(params, "workspace"))
		if err != nil {
			return fail(id, err.Error())
		}
		return ok(id, status)
	case "knowblazer_recall":
		task := stringParam(params, "task")
		project := stringParam(params, "project")
		pack, err := recall.Generate(repoRoot, recall.Options{Task: task, Project: project})
		if err != nil {
			return fail(id, err.Error())
		}
		return ok(id, string(pack))
	case "knowblazer_search":
		query := stringParam(params, "query")
		hits, err := index.Search(repoRoot, query)
		if err != nil {
			return fail(id, err.Error())
		}
		return ok(id, hits)
	case "knowblazer_capture":
		source := stringParam(params, "file")
		result, err := capture.Markdown(repoRoot, source)
		if err != nil {
			return fail(id, err.Error())
		}
		return ok(id, result)
	case "knowblazer_consolidate":
		result, err := consolidate.Run(repoRoot)
		if err != nil {
			return fail(id, err.Error())
		}
		if result.Count == 0 {
			return ok(id, map[string]any{"status": "noop", "count": 0})
		}
		return ok(id, map[string]any{"status": "synthesized", "path": result.Path, "count": result.Count})
	default:
		return fail(id, fmt.Sprintf("unknown tool: %s", name))
	}
}

func toolDefinitions() []map[string]any {
	return []map[string]any{
		toolDefinition("knowblazer_context", "Generate dynamic task context from Knowblazer memory.", map[string]any{
			"task":      stringSchema("Task description to generate context for."),
			"project":   stringSchema("Optional project memory name."),
			"workspace": stringSchema("Optional workspace path used to resolve project mapping."),
		}, []string{"task"}),
		toolDefinition("knowblazer_remember", "Capture a fresh reusable memory signal or daily note.", map[string]any{
			"text":  stringSchema("Lesson text to remember."),
			"daily": map[string]any{"type": "boolean", "description": "Save as a daily note instead of fresh automatic memory."},
		}, []string{"text"}),
		toolDefinition("knowblazer_status", "Check Knowblazer memory connection and dynamic memory status.", map[string]any{
			"workspace": stringSchema("Optional workspace path."),
		}, nil),
		toolDefinition("knowblazer_recall", "Generate a dynamic recall pack for a task.", map[string]any{
			"task":    stringSchema("Task description."),
			"project": stringSchema("Optional project memory name."),
		}, []string{"task"}),
		toolDefinition("knowblazer_search", "Search the Knowblazer local index.", map[string]any{
			"query": stringSchema("Search query."),
		}, []string{"query"}),
		toolDefinition("knowblazer_capture", "Capture a Markdown file into Knowblazer.", map[string]any{
			"file": stringSchema("Markdown file path to capture."),
		}, []string{"file"}),
		toolDefinition("knowblazer_consolidate", "Consolidate fresh automatic memories into synthesized memory.", map[string]any{}, nil),
	}
}

func toolDefinition(name string, description string, properties map[string]any, required []string) map[string]any {
	schema := map[string]any{
		"type":       "object",
		"properties": properties,
	}
	if len(required) > 0 {
		schema["required"] = required
	}
	return map[string]any{
		"name":        name,
		"description": description,
		"inputSchema": schema,
	}
}

func stringSchema(description string) map[string]any {
	return map[string]any{"type": "string", "description": description}
}

func rememberText(repoRoot string, text string) (map[string]any, error) {
	file, err := os.CreateTemp("", "knowblazer-remember-*.md")
	if err != nil {
		return nil, err
	}
	name := file.Name()
	defer os.Remove(name)
	if _, err := fmt.Fprintf(file, "# Memory\n\n%s\n", text); err != nil {
		file.Close()
		return nil, err
	}
	if err := file.Close(); err != nil {
		return nil, err
	}
	result, err := capture.MarkdownAuto(repoRoot, name)
	if err != nil {
		return nil, err
	}
	if result.ScanLevel == scan.High {
		return nil, fmt.Errorf("sensitive content detected; saved to quarantine: %s", result.Path)
	}
	consolidated, err := consolidate.Run(repoRoot)
	if err != nil {
		return nil, err
	}
	resultMap := map[string]any{"path": result.Path, "status": "fresh", "scan_level": result.ScanLevel.String()}
	if consolidated.Count > 0 {
		resultMap["consolidation"] = map[string]any{"status": "synthesized", "path": consolidated.Path, "count": consolidated.Count}
	}
	return resultMap, nil
}

func status(repoRoot string, workspace string) (map[string]any, error) {
	if workspace == "" {
		workspace, _ = os.Getwd()
	}
	project := ""
	if workspace != "" {
		if mapped, ok, err := projectmap.Resolve(repoRoot, workspace); err == nil && ok {
			project = mapped
		}
	}
	candidates, err := review.List(repoRoot)
	if err != nil {
		return nil, err
	}
	freshCount, err := consolidate.CountFresh(repoRoot)
	if err != nil {
		return nil, err
	}
	synthesizedCount, err := consolidate.CountSynthesized(repoRoot)
	if err != nil {
		return nil, err
	}
	configured := false
	if workspace != "" {
		content, err := os.ReadFile(filepath.Join(workspace, "CLAUDE.md"))
		configured = err == nil && strings.Contains(string(content), "KNOWBLAZER-CLAUDE-SETUP:START")
	}
	return map[string]any{
		"repo":                 repoRoot,
		"project":              project,
		"workspace":            workspace,
		"claude_configured":    configured,
		"fresh_auto_memories":  freshCount,
		"synthesized_memories": synthesizedCount,
		"inbox_candidates":     len(candidates),
	}, nil
}

func boolParam(params map[string]any, name string) bool {
	if params == nil {
		return false
	}
	value, _ := params[name].(bool)
	return value
}

func stringParam(params map[string]any, name string) string {
	if params == nil {
		return ""
	}
	value, _ := params[name].(string)
	return value
}

func mapParam(params map[string]any, name string) map[string]any {
	if params == nil {
		return nil
	}
	value, _ := params[name].(map[string]any)
	return value
}
