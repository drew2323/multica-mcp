package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/strider2038/multica-mcp/internal/multica"
)

type restArgs struct {
	PathParams  map[string]string `json:"path_params"`
	Query       map[string]string `json:"query"`
	Body        json.RawMessage   `json:"body"`
	WorkspaceID string            `json:"workspace_id"`
}

func registerRESTTools(server *mcp.Server, client *multica.Client, readOnly bool) error {
	endpoints, err := LoadEndpoints()
	if err != nil {
		return err
	}
	for _, endpoint := range endpoints {
		e := endpoint
		if readOnly && e.Method != http.MethodGet && e.Method != http.MethodHead {
			continue
		}
		tool := &mcp.Tool{Name: e.Name, Description: fmt.Sprintf("%s %s. Supply path_params, query, body, and optional workspace_id.", e.Method, e.Path), InputSchema: map[string]any{"type": "object", "properties": map[string]any{
			"path_params":  map[string]any{"type": "object", "additionalProperties": map[string]any{"type": "string"}, "description": "Values for the named {placeholders} in the route path."},
			"query":        map[string]any{"type": "object", "additionalProperties": map[string]any{"type": "string"}, "description": "REST query parameters, forwarded without projection."},
			"body":         map[string]any{"type": "object", "additionalProperties": true, "description": "REST JSON request body, forwarded without projection."},
			"workspace_id": map[string]any{"type": "string", "description": "Optional per-call workspace override; otherwise configured workspace is used."},
		}}}
		mcp.AddTool(server, tool, func(ctx context.Context, _ *mcp.CallToolRequest, args restArgs) (*mcp.CallToolResult, any, error) {
			path, err := expandPath(e.Path, args.PathParams)
			if err != nil {
				return nil, nil, err
			}
			body := json.RawMessage(args.Body)
			if len(body) == 0 {
				body = nil
			}
			result, err := client.REST(ctx, e.Method, path, args.Query, body, args.WorkspaceID)
			if err != nil {
				return nil, nil, err
			}
			return &mcp.CallToolResult{Content: []mcp.Content{&mcp.TextContent{Text: string(result)}}}, nil, nil
		})
	}
	return nil
}

func expandPath(path string, params map[string]string) (string, error) {
	for strings.Contains(path, "{") {
		start := strings.IndexByte(path, '{')
		end := strings.IndexByte(path[start:], '}')
		if end < 0 {
			return "", fmt.Errorf("malformed route path %q", path)
		}
		end += start
		key := path[start+1 : end]
		value, ok := params[key]
		if !ok {
			return "", fmt.Errorf("missing path_params.%s", key)
		}
		path = strings.Replace(path, path[start:end+1], url.PathEscape(value), 1)
	}
	return path, nil
}
