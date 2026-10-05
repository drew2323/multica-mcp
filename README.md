# Multica REST MCP

A thin REST-only MCP server maintained in [drew2323/multica-mcp](https://github.com/drew2323/multica-mcp), derived from [strider2038/multica-mcp](https://github.com/strider2038/multica-mcp). Original license and attribution retained. The Go module path still uses the upstream namespace; this is not an upstream release.

## Branch and release status

Use **`rest-api-wrapper`**, tracked by [PR #1](https://github.com/drew2323/multica-mcp/pull/1). `main` still contains the earlier custom-status bridge. Do not install upstream releases or upstream `go install ...@latest` for this implementation. The PR remains draft: exhaustive endpoint schemas and live coverage are incomplete, although the branch is deployed in our private integration. No fork release binary is claimed.

## Behavior

- One catalog tool per REST method/path; arbitrary JSON body/query, repeated query values and request-local workspace override. Two explicit attachment transport tools provide bounded content read and multipart upload.
- Custom statuses are backend-authoritative: no hardcoded enum.
- No CLI calls, planning, fallback orchestration or composite workflows in registered tools.
- Embedded catalog, independent of working directory.
- Read-only mode removes mutating tools.
- 319 included endpoints; explicit exclusions for plugin authentication, internal callbacks, secrets and unsuitable transports. Not 100% API coverage.

See [endpoint catalog](docs/rest-api-catalog.json), [coverage limitations](docs/rest-api-coverage.md), [attachment content tools](docs/attachment-content-tools.md) and [domains/profiles](docs/tool-domains.md). Some input descriptions remain incomplete; raw JSON passthrough is not a complete schema.

## Build

Go 1.25+:

```bash
git clone --branch rest-api-wrapper https://github.com/drew2323/multica-mcp.git
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
- Installed binary matches candidate; historical deployed launcher tools/list 28 and custom status discovery passed. Service active/enabled; tunnel readiness healthy.

Not verified: every endpoint live, agent assignment/autopilot trigger (avoided starting work), reboot via actual reboot, post-upgrade end-to-end ChatGPT call. Tool names changed: refresh is required. Legacy source files remain but workflow tools are not registered.

## License

See [LICENSE](LICENSE); upstream attribution retained.

## Core auxiliary primitives

Core adds label definition reads and issue label add/remove; property definition reads and issue metadata/property value operations; subscriber list; attachment list and metadata read. Issue property values are available in issue detail. Explicit attachment tools now read server-approved text-previewable content (2 MiB server bound; non-UTF-8 is base64) and upload multipart files for issue/comment references (8 MiB wrapper bound). No label/property schema mutations or attachment deletion. Subscribe/unsubscribe are NOT included: REST accepts a different target user, and a thin passthrough cannot promise self-only restrictions.
