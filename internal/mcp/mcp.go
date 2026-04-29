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
	"github.com/knowblazer/knowblazer/internal/daily"
	"github.com/knowblazer/knowblazer/internal/index"
	"github.com/knowblazer/knowblazer/internal/projectmap"
	"github.com/knowblazer/knowblazer/internal/recall"
	"github.com/knowblazer/knowblazer/internal/review"
	"github.com/knowblazer/knowblazer/internal/scan"
)

type Request struct {
	ID     any            `json:"id,omitempty"`
	Method string         `json:"method"`
	Params map[string]any `json:"params,omitempty"`
}

type Response struct {
	ID     any    `json:"id,omitempty"`
	Result any    `json:"result,omitempty"`
	Error  string `json:"error,omitempty"`
}

func Serve(repoRoot string, in io.Reader, out io.Writer) error {
	scanner := bufio.NewScanner(in)
	encoder := json.NewEncoder(out)
	for scanner.Scan() {
		var req Request
		if err := json.Unmarshal(scanner.Bytes(), &req); err != nil {
			_ = encoder.Encode(Response{Error: err.Error()})
			continue
		}
		resp := handle(repoRoot, req)
		if err := encoder.Encode(resp); err != nil {
			return err
		}
	}
	return scanner.Err()
}

func handle(repoRoot string, req Request) Response {
	switch req.Method {
	case "tools/list":
		return Response{ID: req.ID, Result: []string{"knowblazer_context", "knowblazer_remember", "knowblazer_status", "knowblazer_recall", "knowblazer_search", "knowblazer_capture"}}
	case "knowblazer_context":
		task := stringParam(req.Params, "task")
		project := stringParam(req.Params, "project")
		if project == "" {
			workspace := stringParam(req.Params, "workspace")
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
			return Response{ID: req.ID, Error: err.Error()}
		}
		return Response{ID: req.ID, Result: string(pack)}
	case "knowblazer_remember":
		text := strings.TrimSpace(stringParam(req.Params, "text"))
		if text == "" {
			return Response{ID: req.ID, Error: "text is required"}
		}
		if boolParam(req.Params, "daily") {
			path, err := daily.Add(repoRoot, daily.AddOptions{Text: text})
			if err != nil {
				return Response{ID: req.ID, Error: err.Error()}
			}
			return Response{ID: req.ID, Result: map[string]any{"path": path, "status": "daily"}}
		}
		result, err := rememberText(repoRoot, text)
		if err != nil {
			return Response{ID: req.ID, Error: err.Error()}
		}
		return Response{ID: req.ID, Result: result}
	case "knowblazer_status":
		status, err := status(repoRoot, stringParam(req.Params, "workspace"))
		if err != nil {
			return Response{ID: req.ID, Error: err.Error()}
		}
		return Response{ID: req.ID, Result: status}
	case "knowblazer_recall":
		task := stringParam(req.Params, "task")
		project := stringParam(req.Params, "project")
		pack, err := recall.Generate(repoRoot, recall.Options{Task: task, Project: project})
		if err != nil {
			return Response{ID: req.ID, Error: err.Error()}
		}
		return Response{ID: req.ID, Result: string(pack)}
	case "knowblazer_search":
		query := stringParam(req.Params, "query")
		hits, err := index.Search(repoRoot, query)
		if err != nil {
			return Response{ID: req.ID, Error: err.Error()}
		}
		return Response{ID: req.ID, Result: hits}
	case "knowblazer_capture":
		source := stringParam(req.Params, "file")
		result, err := capture.Markdown(repoRoot, source)
		if err != nil {
			return Response{ID: req.ID, Error: err.Error()}
		}
		return Response{ID: req.ID, Result: result}
	default:
		return Response{ID: req.ID, Error: fmt.Sprintf("unknown method: %s", req.Method)}
	}
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
	result, err := capture.Markdown(repoRoot, name)
	if err != nil {
		return nil, err
	}
	if result.ScanLevel == scan.High {
		return nil, fmt.Errorf("sensitive content detected; saved to quarantine: %s", result.Path)
	}
	return map[string]any{"path": result.Path, "status": "candidate", "scan_level": result.ScanLevel.String()}, nil
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
	configured := false
	if workspace != "" {
		content, err := os.ReadFile(filepath.Join(workspace, "CLAUDE.md"))
		configured = err == nil && strings.Contains(string(content), "KNOWBLAZER-CLAUDE-SETUP:START")
	}
	return map[string]any{
		"repo":              repoRoot,
		"project":           project,
		"workspace":         workspace,
		"claude_configured": configured,
		"inbox_candidates":  len(candidates),
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
