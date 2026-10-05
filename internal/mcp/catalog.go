package mcp

import (
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

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
	b, err := os.ReadFile("docs/rest-api-catalog.json")
	if err != nil {
		return nil, err
	}
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
		e.Path = strings.ReplaceAll(e.Path, "//", "/")
		e.Path = strings.TrimSuffix(e.Path, "/")
		if e.Path == "" {
			e.Path = "/"
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
