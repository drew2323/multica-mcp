# REST tool domains and connection profiles

The reviewed REST catalog contains **319** enabled endpoints. Every endpoint is assigned to one semantic domain. Resource-specific routes override their parent route: issue comments, labels, properties/metadata, attachments, subscriptions, pull-request integrations and quick actions are not classified as generic issues; task-runs, active-task and execution/usage belong to `runs`. Issue timelines remain in `issues`. Workspace runtime profiles, plugins and integrations are separated.

## Profiles and counts

Counts below are catalog-derived. Profile membership is the union of these named domains, with the core profile additionally keeping only everyday operations.

| Profile | Tools | Included domains / scope |
|---|---:|---|
| `core` (default) | 43 | everyday `workspaces`, `projects`, `issues`, `comments`; `statuses` GET/HEAD only |
| `delivery` | 126 | core plus agents, runs, integrations, labels, properties, attachments, project-resources, subscriptions, quick-actions and specialized issue/comment domains |
| `automation` | 70 | agents, runs, autopilots, quick-actions, wakeups |
| `admin` | 102 | workspace-admin, integrations, plugins, runtimes, account, billing, statuses |
| `all` | 319 | all included catalog endpoints; preserves raw method/path/body/query/workspace override behavior |

`core` excludes status mutation, subscriptions, workspace administration, automation/runs, bulk and analytics operations, project resources, and advanced comments.

## Configuration precedence

`MULTICA_MCP_PROFILE` selects `core` when unset, or one of `core`, `delivery`, `automation`, `admin`, `all`. `MULTICA_MCP_DOMAINS` is a comma-separated domain allowlist; if non-empty it **overrides** profile membership. The special value `all` cannot be combined with other values. Unknown profiles/domains, empty domain elements, and any included catalog endpoint without a semantic domain cause startup selection to fail. `MULTICA_READ_ONLY=true` is applied after selection and removes all but GET/HEAD endpoints.

Core also selects narrowly allowlisted label/property operations, subscriber reads, and attachment metadata reads. Definition mutations, subscription mutations, attachment deletion/content/download/upload are excluded. Subscription routes can target others, so cannot be presented as self-only in a raw REST wrapper. Multipart upload and binary reads remain unsupported.
