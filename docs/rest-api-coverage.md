# REST API coverage

Source: Multica server commit `d021e1bde60000399e02a3bc9144aa3499766538`, router `server/cmd/server/router.go`. Catalog: 447 direct declarations plus 19 shared plugin-route registrations, 466 entries total. Of these, 319 are exposed and 147 explicitly excluded. Counts describe this source inventory, not proof of all runtime behavior.

## Mapping and scope

Each included method/path has one MCP tool. Requests forward path parameters, arbitrary JSON body and repeated query values without workflow logic or status enums. Workspace override is request-local. Responses preserve REST JSON. Read-only mode removes mutating tools. No CLI invocation or composite planning tools are registered.

Minimum families: workspace detail/members; projects list/detail/create; issues list/filter/search/detail/create/update/delete/status/assignment/children; threaded comments list/create/update/delete; subscribers list/subscribe/unsubscribe; agents list/detail/tasks; issue task runs/history; autopilots CRUD/trigger/run history. Exact method/path, handler, tool-name derivation and inclusion reasons are in `internal/mcp/rest-api-catalog.json`; `docs/rest-api-catalog.json` mirrors it.

## Authentication and transport exclusions

`registerPluginActionRoutes` mounts nine JSON operations at `/v1` and nine equivalents plus a hook operation at `/api/plugin-bridge/v1`. All 19 are inventoried but excluded: public `/v1` requires plugin-specific `mpi_`/`mpc_` bearer credentials, which cannot use this wrapper's user credential; bridge routes require host installation identity and browser-session semantics. Supporting plugin credentials separately is out of the current user-scoped wrapper. Internal daemon callbacks, credentials/auth/session/signing-secret operations, streaming/binary transports and other exclusions have per-route reasons in the catalog. These are explicit exceptions to full API coverage, not silently supported routes.

## Input documentation limitations

`tools/catalog-inputs` uses Go AST to observe query reads and Decoder.Decode request targets, including referenced types. Observed field names are not a complete schema: requiredness, defaults, validation, nested structure and delegated binding helpers may remain unresolved. Some legacy notes remain generic. Generic JSON passthrough permits these REST requests, but endpoint-specific input documentation is not complete. Response schemas and permissions are not exhaustively modeled.

## Real verification

`go test -race ./...` passed. Candidate stdio initialization from an unrelated working directory and tools/list returned 319 tools. Live reads passed for workspace/members, projects, issues, statuses, agents and autopilots. Disposable issue create/read/status update with custom `ingested` passed; children/comments/subscribers/task-run reads passed; threaded comment create/edit/delete passed; issue cleanup was verified by HTTP 404. Autopilot triggering and agent assignment were not exercised live to avoid starting work. The implementation is deployed through a private stdio tunnel with core 45 tools. Attachment Markdown upload/read was verified against the deployed launcher with exact content agreement and cleanup 404; the maintainer confirmed the refreshed ChatGPT integration works. The all profile now advertises 321 tools (319 catalog operations plus two explicit attachment transport tools).
