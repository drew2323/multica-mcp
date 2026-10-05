package mcp

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
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
