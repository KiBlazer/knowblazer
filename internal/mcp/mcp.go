package mcp

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"

	"github.com/knowblazer/knowblazer/internal/capture"
	"github.com/knowblazer/knowblazer/internal/index"
	"github.com/knowblazer/knowblazer/internal/recall"
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
		return Response{ID: req.ID, Result: []string{"knowblazer_recall", "knowblazer_search", "knowblazer_capture"}}
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

func stringParam(params map[string]any, name string) string {
	if params == nil {
		return ""
	}
	value, _ := params[name].(string)
	return value
}
