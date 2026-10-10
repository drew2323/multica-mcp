package mcp

import (
	"fmt"
	"strings"
)

var profiles = map[string][]string{
	"core":       {"workspaces", "projects", "issues", "comments", "statuses", "labels", "properties", "attachments", "subscriptions"},
	"delivery":   {"workspaces", "projects", "issues", "comments", "comments-advanced", "issues-bulk", "issues-analytics", "issues-advanced", "statuses", "agents", "runs", "integrations", "labels", "properties", "attachments", "project-resources", "subscriptions", "quick-actions"},
	"automation": {"agents", "runs", "autopilots", "quick-actions", "wakeups"},
	"admin":      {"workspace-admin", "integrations", "plugins", "runtimes", "account", "billing", "statuses"},
}

func endpointDomain(e Endpoint) string {
	p := strings.Trim(strings.TrimPrefix(e.Path, "/api/"), "/")
	a := strings.Split(p, "/")
	first := ""
	if len(a) > 0 {
		first = a[0]
	}
	if first == "issues" && len(a) == 2 {
		switch a[1] {
		case "batch-update", "batch-delete":
			return "issues-bulk"
		case "table", "grouped", "child-progress", "limit-usage":
			return "issues-analytics"
		case "quick-create", "preview-trigger":
			return "issues-advanced"
		}
	}
	if first == "issues" && len(a) >= 3 {
		if a[1] == "table" {
			return "views"
		}
		switch a[2] {
		case "comments":
			if strings.Contains(p, "trigger-preview") || strings.Contains(p, "reactions") {
				return "comments-advanced"
			}
			return "comments"
		case "labels":
			return "labels"
		case "properties", "metadata":
			return "properties"
		case "attachments":
			return "attachments"
		case "subscribers", "subscriptions", "subscribe", "unsubscribe":
			return "subscriptions"
		case "pull-requests", "integrations":
			return "integrations"
		case "quick-actions":
			return "quick-actions"
		case "task-runs", "active-task", "execution", "usage":
			return "runs"
		case "batch-update", "batch-delete":
			return "issues-bulk"
		case "table", "grouped", "child-progress", "limit-usage":
			return "issues-analytics"
		case "preview-trigger", "quick-create", "squad-evaluated", "duplicates":
			return "issues-advanced"
		case "wakeups", "system-wakeups":
			return "wakeups"
		}
	}
	if first == "projects" && len(a) >= 3 {
		return "project-resources"
	}
	if first == "projects" && (strings.Contains(p, "resources")) {
		return "project-resources"
	}
	if first == "workspaces" && len(a) >= 2 {
		resource := a[1]
		if resource == "{id}" && len(a) >= 3 {
			resource = a[2]
		}
		switch resource {
		case "plugins":
			return "plugins"
		case "mcp-servers", "vcs", "github":
			return "integrations"
		case "runtime-profiles", "runtimes":
			return "runtimes"
		case "members":
			if e.Method == "GET" {
				return "workspaces"
			}
			return "workspace-admin"
		case "invitations", "share-links", "leave":
			return "workspace-admin"
		}
	}
	if first == "workspaces" && len(a) >= 3 {
		switch a[2] {
		case "plugins":
			return "plugins"
		case "mcp-servers", "vcs", "github":
			return "integrations"
		case "runtime-profiles":
			return "runtimes"
		case "runtimes":
			return "runtimes"
		case "members", "invitations", "share-links":
			return "workspace-admin"
		}
	}
	if first == "issue-statuses" && e.Method != "GET" && e.Method != "HEAD" {
		return "workspace-admin"
	}
	if first == "workspaces" && (e.Method == "POST" || e.Method == "PUT" || e.Method == "PATCH") && (len(a) == 1 || (len(a) > 1 && a[1] == "members")) {
		return "workspace-admin"
	}
	if first == "autopilots" && len(a) >= 3 && a[2] == "runs" {
		return "runs"
	}
	switch first {
	case "workspace":
		return "workspaces"
	case "runs":
		return "runs"
	case "issue-statuses":
		return "statuses"
	case "issue-views", "issue-view-preferences", "pins":
		return "views"
	case "issues":
		return "issues"
	case "tasks":
		return "runs"
	case "comments":
		if strings.Contains(p, "trigger-preview") || strings.Contains(p, "reactions") {
			return "comments-advanced"
		}
		return "comments"
	case "autopilots":
		return "autopilots"
	case "agents":
		return "agents"
	case "projects":
		return "projects"
	case "workspaces":
		if len(a) == 2 && a[1] == "{id}" && e.Method == "GET" {
			return "workspaces"
		}
		if len(a) > 1 && a[1] == "{id}" {
			return "workspace-admin"
		}
		return "workspaces"
	case "attachments":
		return "attachments"
	case "properties":
		return "properties"
	case "labels":
		return "labels"
	case "integrations":
		return "integrations"
	case "chat":
		return "chat"
	case "inbox", "notification-preferences":
		return "notifications"
	case "quick-actions":
		return "quick-actions"
	case "squads":
		return "squads"
	case "skills":
		return "skills"
	case "runtimes", "cloud-runtime":
		return "runtimes"
	case "system-wakeups", "issue-wakeups", "issue-wakeup-paused", "issue-wakeup-summaries":
		return "wakeups"
	case "working-agents", "agent-activity-30d", "agent-run-counts", "agent-task-snapshot", "dashboard", "assignee-frequency":
		return "dashboard"
	case "cloud-billing", "cloud-subscriptions":
		return "billing"
	case "invitations", "share-links":
		return "workspace-admin"
	case "config", "me", "feedback", "search-index":
		return "account"
	case "autopilot":
		return "autopilots"
	}
	return "misc"
}

func SelectEndpoints(all []Endpoint, selection string) ([]Endpoint, error) {
	return SelectEndpointsForProfile(all, "core", selection)
}
func SelectEndpointsForProfile(all []Endpoint, profile, selection string) ([]Endpoint, error) {
	profile = strings.ToLower(strings.TrimSpace(profile))
	if profile == "" {
		profile = "core"
	}
	domains, ok := profiles[profile]
	if !ok && profile != "all" {
		return nil, fmt.Errorf("unknown REST profile %q", profile)
	}
	selected := map[string]bool{}
	if strings.TrimSpace(selection) != "" {
		for _, raw := range strings.Split(selection, ",") {
			d := strings.ToLower(strings.TrimSpace(raw))
			if d == "" {
				return nil, fmt.Errorf("MULTICA_MCP_DOMAINS contains an empty domain")
			}
			if d == "all" {
				if len(strings.Split(selection, ",")) > 1 {
					return nil, fmt.Errorf("domain all cannot be combined with other domains")
				}
				return append([]Endpoint(nil), all...), nil
			}
			found := false
			for _, e := range all {
				if endpointDomain(e) == d {
					found = true
					break
				}
			}
			if !found {
				return nil, fmt.Errorf("unknown REST tool domain %q", d)
			}
			selected[d] = true
		}
	} else if profile != "all" {
		for _, d := range domains {
			selected[d] = true
		}
	} else {
		for _, e := range all {
			selected[endpointDomain(e)] = true
		}
	}
	out := make([]Endpoint, 0, len(all))
	for _, e := range all {
		d := endpointDomain(e)
		if d == "misc" {
			return nil, fmt.Errorf("unclassified REST endpoint %s %s", e.Method, e.Path)
		}
		if selected[d] && (strings.TrimSpace(selection) != "" || profile != "core" || coreEndpoint(e)) {
			out = append(out, e)
		}
	}
	return out, nil
}

func coreEndpoint(e Endpoint) bool {
	p := e.Path
	d := endpointDomain(e)
	// These are the narrow, user-facing primitives allowed into core beyond
	// issue CRUD. Match exact method and route so schema/admin mutations and
	// unrelated attachment operations never leak in by path substring.
	if d == "labels" || d == "properties" || d == "attachments" || d == "subscriptions" {
		key := e.Method + " " + p
		for _, allowed := range []string{
			"GET /api/labels", "GET /api/labels/{id}", "POST /api/labels",
			"GET /api/properties", "GET /api/properties/{id}",
			"GET /api/issues/{id}/labels", "POST /api/issues/{id}/labels", "DELETE /api/issues/{id}/labels/{labelId}",
			"GET /api/issues/{id}/metadata", "PUT /api/issues/{id}/metadata/{key}", "DELETE /api/issues/{id}/metadata/{key}",
			"PUT /api/issues/{id}/properties/{propertyId}", "DELETE /api/issues/{id}/properties/{propertyId}",
			"GET /api/issues/{id}/subscribers",
			"GET /api/issues/{id}/attachments", "GET /api/attachments/{id}",
		} {
			if key == allowed {
				return true
			}
		}
		return false
	}
	if d == "workspaces" {
		return e.Path == "/api/workspaces" || e.Path == "/api/workspaces/{id}" || e.Path == "/api/workspaces/{id}/members"
	}
	if d == "statuses" {
		return e.Method == "GET"
	}
	if d == "projects" {
		return !strings.Contains(p, "resources")
	}
	if d != "issues" {
		return true
	}
	for _, x := range []string{"/quick-create", "/preview-trigger", "/batch-", "/table/", "/grouped", "/child-progress", "/limit-usage", "/duplicates", "/squad-evaluated", "/wakeups", "/task-runs", "/active-task", "/tasks/", "/rerun", "/quick-actions/", "/usage", "/reactions", "/attachments", "/labels", "/metadata", "/properties", "/pull-requests", "/pr-auto-complete", "/subscribers", "/subscribe", "/unsubscribe"} {
		if strings.Contains(p, x) {
			return false
		}
	}
	return true
}
