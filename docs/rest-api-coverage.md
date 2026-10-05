# REST API coverage requirements

## Source and boundary

Inventory source is deployed Multica server commit `d021e1bde60000399e02a3bc9144aa3499766538`, primarily `server/cmd/server/router.go`. The machine-readable catalog enumerates registered route declarations and source line numbers. Routes outside the feature boundary remain cataloged as exclusions; this MCP server must be a thin REST mapping only (no CLI, workflow, planning, fallback, or orchestration behavior).

## Required implementation families

- Workspace discovery/detail/members: list, create (if applicable), detail, member list and mutation routes. Workspace selection/override must be per tool call, not sticky process state.
- Projects: search/list/detail/create/update/delete.
- Issues: list/search/query, detail, create/update/delete, status transition (using actual status endpoints), assignment fields, children/parent queries.
- Comments: issue comments list/create; threaded comment reply/detail/edit/delete and thread operations as supported by registered routes.
- Subscribers: list, subscribe, unsubscribe (including subtree if needed by API contract).
- Agents: list/detail/tasks; include task run/history and user-visible task messages.
- Issue execution/history: issue task-runs/timeline and associated run/message routes that are needed for retrieval.
- Autopilots: list/detail/create/update/delete, trigger, runs and run detail; related trigger CRUD when needed by supported contract.

The endpoint catalog `implementation_scope` marks coarse boundary matches. Fine-grained allowlist should be reviewed by the implementer against this family list; adjacent routes (for example wakeups, reactions, labels, operational controls, billing, plugins) are not implicitly required merely because they share a path prefix.

## Mapping/verification rules

1. Use the deployed REST handlers as source of truth for path, HTTP method, query names, JSON fields, response envelope, pagination, and permission checks. The catalog preserves handler names and router line references; `query_fields`/`request_body` intentionally remain handler-inspection requirements rather than guesses.
2. Preserve Multica identifiers, nullable fields, enums, and error semantics. Do not invent aliases or translate into workflows.
3. Expose workspace override explicitly per call (header/query/tool argument according to source handler contract); default to authenticated user's selected workspace only when the REST client contract permits it. Never mutate global client workspace state.
4. Keep writes explicit and one-to-one with REST operations. No automatic retries of non-idempotent requests unless server idempotency is verified.
5. Do not include credential-bearing, webhook, daemon/runtime callback, health/metrics/profiling, static asset, or internal transport routes in general MCP tools. Their exclusions remain in catalog.
6. Implement authorization failures and route-level role restrictions as server responses; do not emulate privileges client-side.
7. Test each included method/path with mocked HTTP transport: request serialization, query encoding, workspace override isolation, response parsing, and error propagation. No live writes are part of this inventory task.

## Inventory status

Generated from registered route literals. Exact request-field and permission documentation requires per-handler source inspection before coding; do not treat uninspected schema notes as verified. See catalog for complete endpoint inventory and inclusion/exclusion reasons.
