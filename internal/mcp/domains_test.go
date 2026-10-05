package mcp

import (
	"net/http"
	"strings"
	"testing"
)

func TestEndpointDomainClassification(t *testing.T) {
	cases := []struct{ path, want string }{
		{"/api/issues/{id}/comments", "comments"}, {"/api/issues/{id}/subscribers", "issues"},
		{"/api/comments/{commentId}/resolve", "comments"}, {"/api/autopilots/{id}/runs", "runs"},
		{"/api/autopilots/{id}", "autopilots"}, {"/api/issue-statuses", "statuses"},
		{"/api/issue-views/{id}", "views"}, {"/api/workspaces/{id}/plugins", "plugins"},
	}
	for _, tc := range cases {
		t.Run(tc.path, func(t *testing.T) {
			if got := endpointDomain(Endpoint{Path: tc.path}); got != tc.want {
				t.Fatalf("domain=%q want %q", got, tc.want)
			}
		})
	}
}

func TestAllIncludedEndpointsHaveDomain(t *testing.T) {
	endpoints, err := LoadEndpoints()
	if err != nil {
		t.Fatal(err)
	}
	for _, e := range endpoints {
		if endpointDomain(e) == "" || endpointDomain(e) == "misc" {
			t.Errorf("unclassified: %s %s", e.Method, e.Path)
		}
	}
}

func TestSelectRESTEndpoints(t *testing.T) {
	all, err := LoadEndpoints()
	if err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct {
		name, raw string
		want      int
		bad       bool
	}{
		{"default", "", 0, false}, {"all", "all", len(all), false}, {"custom", "projects,comments", 0, false}, {"unknown", "bogus", 0, true}, {"mixed all", "all,projects", 0, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got, err := SelectEndpoints(all, tc.raw)
			if (err != nil) != tc.bad {
				t.Fatalf("err=%v", err)
			}
			if err == nil && tc.want > 0 && len(got) != tc.want {
				t.Fatalf("got %d want %d", len(got), tc.want)
			}
			if tc.name == "default" && err == nil {
				for _, e := range got {
					d := endpointDomain(e)
					if d != "workspaces" && d != "projects" && d != "issues" && d != "comments" && d != "statuses" {
						t.Errorf("unexpected default domain %q", d)
					}
				}
			}
		})
	}
}

func TestDomainSelectionAppliedBeforeReadOnlyFilter(t *testing.T) {
	all, err := LoadEndpoints()
	if err != nil {
		t.Fatal(err)
	}
	selected, err := SelectEndpoints(all, "comments")
	if err != nil {
		t.Fatal(err)
	}
	for _, e := range selected {
		if endpointDomain(e) != "comments" {
			t.Fatal(e.Path)
		}
	}
	readOnly := 0
	for _, e := range selected {
		if e.Method == http.MethodGet || e.Method == http.MethodHead {
			readOnly++
		}
	}
	if readOnly == 0 || readOnly == len(selected) {
		t.Fatalf("unexpected read-only count %d of %d", readOnly, len(selected))
	}
}

func TestDomainSelectionTrimsAndNormalizes(t *testing.T) {
	all, _ := LoadEndpoints()
	got, err := SelectEndpoints(all, " projects, comments ")
	if err != nil {
		t.Fatal(err)
	}
	for _, e := range got {
		if !strings.Contains(" projects comments ", " "+endpointDomain(e)+" ") {
			t.Fatalf("unexpected domain %q", endpointDomain(e))
		}
	}
}
