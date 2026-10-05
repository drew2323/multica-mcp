package mcp

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/strider2038/multica-mcp/internal/multica"
)

func TestRESTCallPreservesResponseAndScopesPerCall(t *testing.T) {
	var gotPath, gotWorkspace string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath, gotWorkspace = r.URL.RequestURI(), r.Header.Get("X-Workspace-ID")
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"unknown":{"kept":true},"array":[1,2]}`))
	}))
	defer srv.Close()
	c := multica.NewClient(srv.URL, "test-token", "test")
	result, err := c.REST(context.Background(), http.MethodGet, "/api/issues/search", map[string][]string{"q": {"a b", "second"}}, nil, "workspace-override")
	if err != nil {
		t.Fatal(err)
	}
	if gotPath != "/api/issues/search?q=a+b&q=second" || gotWorkspace != "workspace-override" {
		t.Fatalf("request path=%q workspace=%q", gotPath, gotWorkspace)
	}
	var got map[string]any
	if err := json.Unmarshal(result, &got); err != nil {
		t.Fatal(err)
	}
	if got["unknown"] == nil || got["array"] == nil {
		t.Fatalf("response was projected: %s", result)
	}
}

func TestExpandPathRejectsUnsafeSegments(t *testing.T) {
	for _, value := range []string{"", ".", "..", "a/b"} {
		if _, err := expandPath("/api/issues/{id}", map[string]string{"id": value}); err == nil {
			t.Errorf("expandPath accepted unsafe id %q", value)
		}
	}
}

func TestExpandPathEscapesExactlyOnce(t *testing.T) {
	got, err := expandPath("/api/issues/{id}", map[string]string{"id": "a b%2Fc"})
	if err != nil {
		t.Fatal(err)
	}
	if got != "/api/issues/a%20b%252Fc" {
		t.Fatalf("path = %q", got)
	}
}

func TestEndpointCatalogCoverageAndInputDocumentation(t *testing.T) {
	endpoints, err := LoadEndpoints()
	if err != nil {
		t.Fatal(err)
	}
	if err := ValidateEndpoints(endpoints); err != nil {
		t.Fatal(err)
	}
	byHandler := make(map[string]Endpoint, len(endpoints))
	for _, e := range endpoints {
		byHandler[e.Handler] = e
	}
	required := []string{
		"GetWorkspace", "ListMembersWithUser", "SearchProjects", "ListProjects", "CreateProject", "UpdateProject", "DeleteProject",
		"SearchIssues", "ListIssues", "QueryIssues", "CreateIssue", "GetIssue", "UpdateIssue", "ListChildIssues", "ListTasksByIssue",
		"CreateComment", "ListComments", "ListTimeline", "ListIssueSubscribers", "SubscribeToIssue", "UnsubscribeFromIssue",
		"ListAgents", "GetAgent", "ListAgentTasks", "ListAutopilots", "CreateAutopilot", "UpdateAutopilot", "TriggerAutopilot", "ListAutopilotRuns", "GetAutopilotRun",
	}
	for _, name := range required {
		if _, ok := byHandler[name]; !ok {
			t.Errorf("required endpoint handler %s missing", name)
		}
	}
	for _, name := range required {
		e := byHandler[name]
		if strings.Contains(e.Query, "Not inferred") || strings.Contains(e.Body, "Inspect linked") {
			t.Errorf("%s has placeholder input docs", name)
		}
	}
	for _, e := range endpoints {
		if strings.HasPrefix(e.Path, "/api/daemon/") {
			t.Errorf("daemon endpoint exposed: %s %s", e.Method, e.Path)
		}
		if len(e.Name) > 64 {
			t.Errorf("tool name too long (%d): %s", len(e.Name), e.Name)
		}
	}
	// Catalog runtime entries must be fully categorized, and only included
	// safe user-facing JSON endpoints may register as tools.
	var catalog struct { Endpoints []Endpoint `json:"endpoints"` }
	if err := json.Unmarshal(catalogJSON, &catalog); err != nil { t.Fatal(err) }
	included := 0
	for _, e := range catalog.Endpoints {
		if e.Scope == "include" { included++ } else if strings.TrimSpace(e.Reason) == "" {
			t.Errorf("excluded route has no reason: %s %s", e.Method, e.Path)
		}
	}
	if included < len(endpoints) {
		t.Errorf("effective tools %d exceed included catalog entries %d", len(endpoints), included)
	} else if included-len(endpoints) != 0 {
		t.Errorf("included catalog entries %d but effective tools %d", included, len(endpoints))
	}
}

func TestRESTToolDescriptionIncludesCatalogInputs(t *testing.T) {
	e := Endpoint{Method: http.MethodPost, Path: "/api/issues/query", Query: "Same filters", Body: "JSON filter object"}
	got := restToolDescription(e)
	for _, want := range []string{"Same filters", "JSON filter object", "path_params", "workspace_id"} {
		if !strings.Contains(got, want) {
			t.Errorf("description %q missing %q", got, want)
		}
	}
}
