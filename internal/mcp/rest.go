package mcp

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"unicode/utf8"

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
	if profile == "" {
		profile = "core"
	}
	attachmentSelected := false
	for _, e := range endpoints {
		if endpointDomain(e) == "attachments" {
			attachmentSelected = true
			break
		}
	}
	if attachmentSelected {
		readTool := &mcp.Tool{Name: "multica_attachment_read_content", Description: "Read authenticated text-previewable attachment content by attachment ID. Text is returned as UTF-8; non-UTF-8 bytes are returned as base64. Server limits content to 2 MiB.", InputSchema: map[string]any{"type": "object", "properties": map[string]any{"attachment_id": map[string]any{"type": "string"}, "workspace_id": map[string]any{"type": "string"}}, "required": []string{"attachment_id"}}}
		mcp.AddTool(server, readTool, func(ctx context.Context, _ *mcp.CallToolRequest, args struct {
			AttachmentID string `json:"attachment_id"`
			WorkspaceID  string `json:"workspace_id"`
		}) (*mcp.CallToolResult, any, error) {
			data, ct, err := client.AttachmentContent(ctx, args.AttachmentID, args.WorkspaceID)
			if err != nil {
				return nil, nil, err
			}
			result := map[string]any{"content_type": ct, "size_bytes": len(data)}
			if utf8.Valid(data) {
				result["encoding"] = "utf-8"
				result["content"] = string(data)
			} else {
				result["encoding"] = "base64"
				result["content"] = base64.StdEncoding.EncodeToString(data)
			}
			b, _ := json.Marshal(result)
			return &mcp.CallToolResult{Content: []mcp.Content{&mcp.TextContent{Text: string(b)}}}, nil, nil
		})
		if !readOnly {
			uploadTool := &mcp.Tool{Name: "multica_attachment_upload", Description: "Upload an attachment to an issue or comment using text content or base64 bytes (maximum 8 MiB). Provide exactly one of content or content_base64.", InputSchema: map[string]any{"type": "object", "properties": map[string]any{"filename": map[string]any{"type": "string"}, "content_type": map[string]any{"type": "string"}, "content": map[string]any{"type": "string"}, "content_base64": map[string]any{"type": "string"}, "issue_id": map[string]any{"type": "string"}, "comment_id": map[string]any{"type": "string"}, "workspace_id": map[string]any{"type": "string"}}, "required": []string{"filename"}}}
			mcp.AddTool(server, uploadTool, func(ctx context.Context, _ *mcp.CallToolRequest, args struct {
				Filename      string `json:"filename"`
				ContentType   string `json:"content_type"`
				Content       string `json:"content"`
				ContentBase64 string `json:"content_base64"`
				IssueID       string `json:"issue_id"`
				CommentID     string `json:"comment_id"`
				WorkspaceID   string `json:"workspace_id"`
			}) (*mcp.CallToolResult, any, error) {
				if args.Filename == "" || strings.ContainsAny(args.Filename, "/\\") || args.Filename == "." || args.Filename == ".." {
					return nil, nil, fmt.Errorf("filename must be a single non-empty filename")
				}
				if (args.Content == "") == (args.ContentBase64 == "") {
					return nil, nil, fmt.Errorf("provide exactly one of content or content_base64")
				}
				data := []byte(args.Content)
				if args.ContentBase64 != "" {
					var err error
					data, err = base64.StdEncoding.DecodeString(args.ContentBase64)
					if err != nil {
						return nil, nil, fmt.Errorf("content_base64: %w", err)
					}
				}
				if len(data) > 8<<20 {
					return nil, nil, fmt.Errorf("attachment exceeds 8 MiB")
				}
				if args.IssueID == "" && args.CommentID == "" {
					return nil, nil, fmt.Errorf("issue_id or comment_id is required")
				}
				if args.ContentType == "" {
					args.ContentType = "application/octet-stream"
					if strings.HasSuffix(strings.ToLower(args.Filename), ".md") {
						args.ContentType = "text/markdown"
					}
				}
				result, err := client.UploadAttachment(ctx, args.Filename, args.ContentType, data, args.IssueID, args.CommentID, args.WorkspaceID)
				if err != nil {
					return nil, nil, err
				}
				return &mcp.CallToolResult{Content: []mcp.Content{&mcp.TextContent{Text: string(result)}}}, nil, nil
			})
		}
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
