# REST handler input catalog extraction

Reproduce source-observed handler details from the pinned Multica checkout:

```sh
go run ./tools/catalog-inputs -source /path/to/multica/server \
  -catalog docs/rest-api-catalog.json -out /tmp/rest-api-inputs.json
```

The source reference is recorded in `docs/rest-api-catalog.json` (`source.repository` and `source.commit`). The extractor uses only Go's standard `go/parser`/`go/ast`; it resolves handler declarations under `server/internal/handler`, records literal `Get`/`Has` arguments, referenced struct JSON tags, and declaration locations. Its output is evidence, not a complete HTTP contract: AST-only traversal cannot establish branch conditions, requiredness, validation, defaults, or whether a referenced struct is a request versus response. Never infer requiredness from `omitempty` or field names. Review ambiguities against source before copying findings into endpoint descriptions.

Remaining ambiguity: among 319 in-scope catalog endpoints, 110 still have no cataloged query keys and 189 retain a generic/inspect-source request-body description. The 9 query annotations added from extraction are explicitly marked as literal AST observations; `Get`/`Has` may read headers or other values, not URL query parameters. Verify the receiver and any helper/delegated inputs before treating them as query keys. The two named catalog routes `/health/realtime` and `/ws` have no handler identifier; they remain outside the in-scope handler resolution. Structs referenced in handler bodies can be response payloads, not request payloads. No requiredness is inferred.
