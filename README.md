# Multica REST MCP

A broad Multica REST API adapter for MCP clients, with selectable connection profiles—not a server limited to 45 tools. Maintained in [drew2323/multica-mcp](https://github.com/drew2323/multica-mcp), derived from [strider2038/multica-mcp](https://github.com/strider2038/multica-mcp). Original license and attribution retained. The Go module path still uses the upstream namespace; this is not an upstream release.

## Project status

The default `main` branch contains this REST-only implementation, used in our private ChatGPT integration. This is an independently maintained fork, not an upstream release. Exhaustive endpoint schemas and live coverage remain incomplete; limitations are documented below. Build from this repository: upstream releases and upstream `go install ...@latest` do not provide these capabilities. No new fork release binary is claimed.

## Behavior

- One catalog tool per REST method/path; arbitrary JSON body/query, repeated query values and request-local workspace override. Two explicit attachment transport tools provide bounded content read and multipart upload.
- Custom statuses are backend-authoritative: no hardcoded enum.
- No CLI calls, planning, fallback orchestration or composite workflows in registered tools.
- Embedded catalog, independent of working directory.
- Read-only mode removes mutating tools.
- 319 included endpoints; explicit exclusions for plugin authentication, internal callbacks, secrets and unsuitable transports. Not 100% API coverage.

See [endpoint catalog](docs/rest-api-catalog.json), [coverage limitations](docs/rest-api-coverage.md), [attachment content tools](docs/attachment-content-tools.md) and [domains/profiles](docs/tool-domains.md). Some input descriptions remain incomplete; raw JSON passthrough is not a complete schema.

## REST API scope and launch profiles

The server exposes the supported Multica REST API surface directly, rather than a small collection of predefined workflows. **`all` exposes 321 tools: 319 catalog-driven REST operations plus attachment content read and multipart upload.** The **45-tool `core` is only the default discovery profile**, not the extent of the server's capabilities.

Each supported REST method/path maps to a tool. Profiles determine which tools a particular server process advertises:

- **`core` — 45:** routine projects/issues/comments, status reads, label assignment, property values, subscriber reads, attachment metadata/content/upload.
- **`delivery` — 128:** expanded delivery work including agents, execution history, integrations, resource operations and broader issue/comment capabilities.
- **`automation` — 70:** agents, runs, autopilots, quick actions and wakeups.
- **`admin` — 102:** workspace configuration, integrations, plugins, runtimes, account, billing and status administration.
- **`all` — 321:** every included REST operation plus both attachment transport tools.

Counts are for write-enabled configuration. Profiles overlap: they are not disjoint partitions. The API inventory covers 466 route registrations, with 147 documented exclusions (plugin-specific authentication, internal callbacks, credentials/secrets and unsupported transports). Thus this is **broad REST API coverage with explicit exceptions**, not a claim that every route is supported or tested. See [coverage](docs/rest-api-coverage.md).

### Launch with a profile

Provide `MULTICA_BASE_URL`, `MULTICA_TOKEN` and workspace selection via a protected environment/launcher, then select the profile at startup:

```bash
# Broad API access
MULTICA_MCP_PROFILE=all MCP_TRANSPORT=stdio ./bin/multica-mcp

# Routine work, delivery, automation or administration
MULTICA_MCP_PROFILE=core ./bin/multica-mcp
MULTICA_MCP_PROFILE=delivery ./bin/multica-mcp
MULTICA_MCP_PROFILE=automation ./bin/multica-mcp
MULTICA_MCP_PROFILE=admin ./bin/multica-mcp

# Custom domain selection overrides the named profile
MULTICA_MCP_DOMAINS=issues,comments,attachments ./bin/multica-mcp

# Broad discovery without writes (including no upload)
MULTICA_MCP_PROFILE=all MULTICA_READ_ONLY=true ./bin/multica-mcp
```

Leave `MULTICA_MCP_DOMAINS` unset/empty when using a named profile. A local MCP client sets these variables in its server process environment; a remote/tunneled client uses the host's launcher configuration. Each connection advertises one static selection: changing it requires restarting that process and refreshing the client's tool catalog. For simultaneous roles configure separate connections. Profiles restrict discovery, not backend permissions; REST authorization still applies.

## What it provides today

- **Workspaces:** list/detail/member discovery; no implicit workspace administration in core.
- **Projects:** search, list, detail and CRUD; project resources in specialized profiles.
- **Issues:** search/list/detail/create/update/delete, assignment and arbitrary custom statuses through native update fields, hierarchy/children and general issue timeline.
- **Comments:** threaded create/read/update/delete and resolve operations.
- **Labels:** read available definitions and issue labels; add/remove issue labels in core. Definition mutations are not in core.
- **Properties:** read typed definitions and issue values, set/clear issue values and metadata. Definition mutations are not in core.
- **Followers:** list subscribers in core. Subscribe/unsubscribe are specialized because native REST also permits targeting other users.
- **Attachments:** list/detail, read text-previewable content, upload files to issues/comments. Markdown supplementary requirements can be read and written directly. No attachment deletion in core, no image interpretation or PDF text extraction.
- **Specialized profiles:** agent/execution history, autopilots, integrations, runtimes, plugins, notifications, views and configuration endpoints. Presence in the catalog is not a claim of exhaustive live validation.

### Attachment tools

- `multica_attachment_read_content`: authenticated server text-preview endpoint, maximum 2 MiB; returns `content`, `encoding`, `content_type`, `size_bytes`.
- `multica_attachment_upload`: multipart upload, maximum 8 MiB locally; supply `filename`, exactly one of `content` / `content_base64`, and an issue/comment target. `.md` defaults to `text/markdown`. No host filesystem reads.

Example upload arguments:

```json
{"filename":"requirements.md","content":"# Additional requirements\nKeep custom statuses intact.\n","issue_id":"<issue-id>"}
```

Use the attachment ID in the upload response for `multica_attachment_read_content`:

```json
{"attachment_id":"<attachment-id>"}
```

## Build

Go 1.25+:

```bash
git clone https://github.com/drew2323/multica-mcp.git
cd multica-mcp
go test -race ./...
go build -o bin/multica-mcp .
```

## Configuration

Read at process startup. Protect credentials outside Git and service unit text.

- `MULTICA_BASE_URL`: required, actual Multica origin.
- `MULTICA_TOKEN`: required user credential.
- `MULTICA_WORKSPACE_ID` / `MULTICA_WORKSPACE_SLUG`: workspace; slug takes precedence. Unset can resolve a single workspace; multiple workspaces require a choice.
- `MULTICA_MCP_PROFILE`: core (default), delivery, automation, admin, all.
- `MULTICA_MCP_DOMAINS`: explicit comma-separated domains overriding profile selection; unknown selections fail startup.
- `MULTICA_READ_ONLY`: default false; true removes non-GET/HEAD tools.
- `MCP_TRANSPORT`: stdio (default), or legacy http.
- `LOG_LEVEL`: info (default).

Write-enabled profiles: **core 45 / delivery 128 / automation 70 / admin 102 / all 321**. Exact membership in [tool-domains.md](docs/tool-domains.md). Core includes issue search/children/timeline and status discovery, includes subscriber read and excludes status administration. Profiles filter discovery, not credential authorization. Read-only mode and workspace permissions are separate.

For local stdio clients, configure your server command's environment, e.g. `MULTICA_MCP_PROFILE=core`. Use an absolute binary path and your client's protected credential mechanism. MCP arguments are `path_params`, `query`, `body` and optional `workspace_id`; obtain exact names and schemas via `tools/list`, not legacy task-tool examples.

## ChatGPT and private servers

Our deployment: `ChatGPT → OpenAI Secure MCP Tunnel → tunnel-client → stdio server → existing Multica API`.

**The profile is configured on the host**, in the launched MCP process environment. ChatGPT cannot launch local stdio and has no profile dropdown for this server. Prompts and URL parameters do not switch profiles. Restart after changing host configuration, then refresh the client tool catalog.

Concurrent profiles require separate configured processes/connections and remote registrations; not automatically provisioned. ChatGPT action controls can further restrict tools, but do not change the server profile.

Current [OpenAI instructions](https://help.openai.com/en/articles/12584461-developer-mode-and-mcp-apps-in-chatgpt):

- Enterprise/Edu: Workspace settings → Apps → app menu → Action control → Refresh. Review/enable new actions (disabled by default).
- Business published apps: recreate and republish to change tools/metadata.
- New app: Scan Tools before creation, then select it in a new chat.

UI/plan capabilities may change. Preserve the existing tunnel endpoint and authentication. No public Multica exposure is needed. Legacy HTTP (`MCP_HTTP_PORT` default 8080, optional `MCP_API_KEY`) is not our private deployment; review binding/network/auth before using it.

## Verification and remaining limits

Verified:

- `go test -race ./...`, build and actual stdio tools/list for all five profiles/counts above.
- Live MCP reads: workspace/members, projects, issues, statuses, agents, autopilots.
- Disposable unassigned issue create/read and `ingested → todo → ingested` transitions.
- Children/comments/subscribers/task-run reads; threaded comment create/update/delete.
- Cleanup independently confirmed by HTTP 404.
- Installed binary matches candidate; deployed launcher tools/list 45 and custom status discovery passed. Service active/enabled; tunnel readiness healthy.

Maintainer confirmed the refreshed ChatGPT connection works after deployment. Live Markdown upload/read was separately verified through the deployed MCP launcher with exact Czech text agreement, followed by attachment and issue cleanup (404).

Not verified: every endpoint live, agent assignment/autopilot trigger (avoided starting work), reboot via actual reboot. Tool names changed: refresh is required. Legacy source files remain but workflow tools are not registered.

## License

See [LICENSE](LICENSE); upstream attribution retained.

## Core auxiliary primitives

Core adds label definition reads and issue label add/remove; property definition reads and issue metadata/property value operations; subscriber list; attachment list and metadata read. Issue property values are available in issue detail. Explicit attachment tools now read server-approved text-previewable content (2 MiB server bound; non-UTF-8 is base64) and upload multipart files for issue/comment references (8 MiB wrapper bound). No label/property schema mutations or attachment deletion. Subscribe/unsubscribe are NOT included: REST accepts a different target user, and a thin passthrough cannot promise self-only restrictions.
