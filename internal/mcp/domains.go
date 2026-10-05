package mcp

import (
	"fmt"
	"strings"
)

var defaultRESTDomains = map[string]bool{"workspaces": true, "projects": true, "issues": true, "comments": true, "statuses": true}

// endpointDomain classifies routes by their most specific resource owner.
func endpointDomain(e Endpoint) string {
	p := strings.Trim(strings.TrimPrefix(e.Path, "/api/"), "/")
	parts := strings.Split(p, "/")
	if len(parts) >= 3 && parts[0] == "issues" {
		switch parts[2] {
		case "comments":
			return "comments"
		case "task-runs", "tasks", "rerun":
			return "runs"
		case "wakeups", "system-wakeups":
			return "wakeups"
		}
	}
	if len(parts) >= 3 && parts[0] == "autopilots" && parts[2] == "runs" {
		return "runs"
	}
	if len(parts) >= 3 && parts[0] == "workspaces" && parts[2] == "plugins" {
		return "plugins"
	}
	if len(parts) >= 3 && parts[0] == "workspaces" && parts[2] == "mcp-servers" {
		return "integrations"
	}
	if len(parts) > 0 {
		switch parts[0] {
		case "workspace":
			return "workspaces"
		case "issue-statuses":
			return "statuses"
		case "issue-views", "issue-view-preferences":
			return "views"
		case "issues":
			return "issues"
		case "comments":
			return "comments"
		case "autopilots":
			return "autopilots"
		case "agents":
			return "agents"
		case "projects":
			return "projects"
		case "workspaces":
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
		case "pins":
			return "views"
		case "squads":
			return "squads"
		case "skills":
			return "skills"
		case "runtimes", "cloud-runtime":
			return "runtimes"
		case "tasks":
			return "tasks"
		case "system-wakeups", "issue-wakeups", "issue-wakeup-paused", "issue-wakeup-summaries":
			return "wakeups"
		case "working-agents", "agent-activity-30d", "agent-run-counts", "agent-task-snapshot", "dashboard", "assignee-frequency":
			return "status"
		case "cloud-billing", "cloud-subscriptions":
			return "billing"
		case "invitations", "share-links":
			return "workspaces"
		case "config", "me", "feedback", "search-index":
			return "account"
		}
	}
	return "misc"
}

// SelectEndpoints applies the configured allowlist; unset/empty selects the core set.
func SelectEndpoints(all []Endpoint, selection string) ([]Endpoint, error) {
	selection = strings.TrimSpace(selection)
	selected := map[string]bool{}
	if selection == "" {
		for d := range defaultRESTDomains {
			selected[d] = true
		}
	} else {
		for _, raw := range strings.Split(selection, ",") {
			d := strings.ToLower(strings.TrimSpace(raw))
			if d == "" {
				return nil, fmt.Errorf("MULTICA_MCP_DOMAINS contains an empty domain")
			}
			if d == "all" {
				if len(selected) > 0 || len(strings.Split(selection, ",")) > 1 {
					return nil, fmt.Errorf("domain all cannot be combined with other domains")
				}
				return append([]Endpoint(nil), all...), nil
			}
			known := false
			for _, e := range all {
				if endpointDomain(e) == d {
					known = true
					break
				}
			}
			if !known {
				return nil, fmt.Errorf("unknown REST tool domain %q", d)
			}
			selected[d] = true
		}
	}
	out := make([]Endpoint, 0, len(all))
	for _, e := range all {
		if selected[endpointDomain(e)] {
			out = append(out, e)
		}
	}
	return out, nil
}
