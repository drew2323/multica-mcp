package mcp

import (
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"fmt"
	_ "embed"
	"strings"
)

//go:embed rest-api-catalog.json
var catalogJSON []byte

type Endpoint struct {
	Method  string `json:"method"`
	Path    string `json:"path"`
	Scope   string `json:"implementation_scope"`
	Name    string `json:"tool_name,omitempty"`
	Handler string `json:"handler,omitempty"`
	Query   string `json:"query_fields,omitempty"`
	Body    string `json:"request_body,omitempty"`
	Reason  string `json:"scope_reason,omitempty"`
}

func LoadEndpoints() ([]Endpoint, error) {
	b := catalogJSON
	var err error
	var catalog struct {
		Endpoints []Endpoint `json:"endpoints"`
	}
	if err = json.Unmarshal(b, &catalog); err != nil {
		return nil, err
	}
	out := make([]Endpoint, 0, len(catalog.Endpoints))
	for _, e := range catalog.Endpoints {
		if e.Scope != "include" {
			continue
		}
		e.Path = normalizeRoute(e.Path)
		if !safeEndpoint(e) {
			continue
		}
		normalized := strings.TrimSuffix(strings.TrimPrefix(e.Path, "/api/"), "/")
		normalized = strings.Trim(normalized, "/")
		normalized = strings.ReplaceAll(normalized, "/", "_")
		normalized = strings.NewReplacer("{", "", "}", "", "-", "_").Replace(normalized)
		e.Name = "multica_" + strings.ToLower(e.Method) + "_" + normalized
		// Duplicated route declarations and path shapes can otherwise collide after normalization.
		h := sha1.Sum([]byte(e.Method + " " + e.Path))
		e.Name += "_" + hex.EncodeToString(h[:3])
		out = append(out, e)
	}
	return out, nil
}
func normalizeRoute(path string) string {
	path = strings.ReplaceAll(path, "//", "/")
	path = strings.TrimSuffix(path, "/")
	if path == "" { return "/" }
	return path
}

func safeEndpoint(e Endpoint) bool {
	p := strings.ToLower(e.Path + " " + e.Handler)
	for _, denied := range []string{"webhook", "token", "secret", "rotate", "signing", "mcp-server", "plugin", "runtime-profile", "share-link", "invitation", "github", "vcs/", "wakeups", "quick-actions", "cancel", "rerun", "trigger-preview", "preview-trigger", "replay", "deliveries", "reaction", "squad-evaluated", "system-wakeup", "archive", "restore", "env"} {
		if strings.Contains(p, denied) { return false }
	}
	return true
}

func ValidateEndpoints(es []Endpoint) error {
	seen := map[string]bool{}
	for _, e := range es {
		if e.Method == "" || !strings.HasPrefix(e.Path, "/api/") {
			return fmt.Errorf("invalid endpoint %s %s", e.Method, e.Path)
		}
		if seen[e.Name] {
			return fmt.Errorf("duplicate tool %s", e.Name)
		}
		seen[e.Name] = true
	}
	return nil
}
