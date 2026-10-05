package mcp

import (
	"crypto/sha1"
	_ "embed"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strings"
)

//go:embed rest-api-catalog.json
var catalogJSON []byte

type Endpoint struct {
	Method      string `json:"method"`
	Path        string `json:"path"`
	Scope       string `json:"implementation_scope"`
	Name        string `json:"tool_name,omitempty"`
	Handler     string `json:"handler,omitempty"`
	Query       string `json:"query_fields,omitempty"`
	Body        string `json:"request_body,omitempty"`
	Reason      string `json:"scope_reason,omitempty"`
	Description string `json:"description,omitempty"`
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
		suffix := "_" + hex.EncodeToString(h[:3])
		e.Name += suffix
		if len(e.Name) > 64 {
			e.Name = strings.TrimRight(e.Name[:64-len(suffix)], "_") + suffix
		}
		out = append(out, e)
	}
	return out, nil
}
func normalizeRoute(path string) string {
	path = strings.ReplaceAll(path, "//", "/")
	path = strings.TrimSuffix(path, "/")
	if path == "" {
		return "/"
	}
	return path
}

func safeEndpoint(e Endpoint) bool {
	// Catalog scope is the reviewed boundary. Do not apply broad path-word
	// filters here: user-visible REST routes may legitimately include these
	// terms. Security-sensitive routes must be explicitly excluded in catalog.
	return e.Scope == "include"
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
