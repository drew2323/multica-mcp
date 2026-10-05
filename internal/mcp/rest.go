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
	Query       map[string]any    `json:"query"`
	Body        json.RawMessage   `json:"body"`
	WorkspaceID string            `json:"workspace_id"`
}

func registerRESTTools(server *mcp.Server, client *multica.Client, readOnly bool, domainSelection, profile string) error {
	endpoints, err := LoadEndpoints()
	if err != nil {
		return err
	}
	endpoints, err = SelectEndpointsForProfile(endpoints, profile, domainSelection)
	if err != nil {
		return err
	}
	for _, endpoint := range endpoints {
		e := endpoint
		if readOnly && e.Method != http.MethodGet && e.Method != http.MethodHead {
			continue
		}
		tool := &mcp.Tool{Name: e.Name, Description: restToolDescription(e), InputSchema: map[string]any{"type": "object", "properties": map[string]any{
			"path_params":  map[string]any{"type": "object", "additionalProperties": map[string]any{"type": "string"}, "description": "Values for the named {placeholders} in the route path."},
			"query":        map[string]any{"type": "object", "additionalProperties": map[string]any{"oneOf": []any{map[string]any{"type": "string"}, map[string]any{"type": "array", "items": map[string]any{"type": "string"}}}}, "description": "REST query parameters; values may be strings or arrays for repeated parameters."},
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
			query, err := parseRESTQuery(args.Query)
			if err != nil {
				return nil, nil, err
			}
			result, err := client.REST(ctx, e.Method, path, query, body, args.WorkspaceID)
			if err != nil {
				return nil, nil, err
			}
			return &mcp.CallToolResult{Content: []mcp.Content{&mcp.TextContent{Text: string(result)}}}, nil, nil
		})
	}
	return nil
}

func parseRESTQuery(values map[string]any) (map[string][]string, error) {
	query := make(map[string][]string, len(values))
	for key, value := range values {
		switch v := value.(type) {
		case string:
			query[key] = []string{v}
		case []any:
			for _, item := range v {
				s, ok := item.(string)
				if !ok {
					return nil, fmt.Errorf("query.%s: expected only strings", key)
				}
				query[key] = append(query[key], s)
			}
		default:
			return nil, fmt.Errorf("query.%s: expected a string or array of strings", key)
		}
	}
	return query, nil
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
		if value == "" || value == "." || value == ".." || strings.Contains(value, "/") || strings.Contains(value, "\\") {
			return "", fmt.Errorf("invalid path_params.%s: expected a non-empty single path segment", key)
		}
		path = strings.Replace(path, path[start:end+1], url.PathEscape(value), 1)
	}
	return path, nil
}

func restToolDescription(e Endpoint) string {
	base := fmt.Sprintf("%s %s. Supply path_params, query, body, and optional workspace_id.", e.Method, e.Path)
	if e.Description != "" {
		return base + " " + e.Description
	}
	if e.Query != "" || e.Body != "" {
		return base + " Query: " + e.Query + " Request body: " + e.Body
	}
	return base
}
