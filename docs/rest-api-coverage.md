# REST API catalog coverage

## Source and scope

Source inventory: deployed Multica server commit `d021e1bde60000399e02a3bc9144aa3499766538`, router source `server/cmd/server/router.go`. Inventory is 447 statically registered route declarations; this is **not** a count of all runtime routes because dynamic mounts are not expanded. `internal/mcp/rest-api-catalog.json` is the canonical embedded catalog; `docs/rest-api-catalog.json` is the generated documentation mirror. `implementation_scope` is the runtime allowlist. Runtime tool names derive from method+path and a short route hash.

Current static inventory: **319 included, 128 excluded, 447 registered declarations**. The runtime count is the effective `LoadEndpoints()` set after explicit catalog scope and route validation, not a claim that all API capabilities are represented. Dynamic mounts are outside these totals.

## Inclusion policy

Expose user-visible REST routes as direct method/path calls; do not implement workflows or orchestration. Exclude credentials and auth/session management; token/signing-secret operations; incoming webhooks and OAuth callbacks; WebSocket/stream transports; binary upload/download/avatar serving; health/diagnostic routes; non-`/api/` routes; and explicitly marked internal surfaces. The exact excluded declarations, handler references, and per-route reason are in the canonical catalog (`implementation_scope: exclude`, `scope_reason`). Exclusions are explicit; runtime no longer silently filters routes by broad path-word heuristics. Review catalog security exclusions before expanding scope.

## Request schemas and limits

For handlers with verified notes, `query_fields` and `request_body` record inspected field names/behavior. For the remaining included handlers, the catalog still contains legacy `Not inferred` / `Inspect linked handler source` notes: these are **unresolved documentation exceptions**, not schema claims. The source checkout cited above was not present in this workspace, so automated extraction of `URL.Query`/`Get` reads and request-struct JSON tags could not be performed reliably. MCP `body` remains an unprojected generic JSON object and does not enforce endpoint-specific validation. Source handler names and route references are retained in each catalog record to enable reproducible follow-up when source is available.

Tool-name length is tested against MCP's 64-character limit. There are no claims here that all response schemas, permissions, pagination, or body field types have been fully modeled; calls forward raw JSON and return the server response intact.

## Verification

`go test ./...` validates the effective catalog, required route families, documented required schemas, protected-route exclusions, response pass-through, and REST request handling. Candidate binary build is required before commit. No production configuration or service is touched.
