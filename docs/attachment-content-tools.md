# Attachment content tools

The core REST profile includes two explicit attachment transport tools in addition to its catalogued JSON tools:

- `multica_attachment_read_content`: fetches `GET /api/attachments/{id}/content`, which enforces workspace membership and server-side text-previewable file policy. The server caps reads at 2 MiB and returns 415 for unsupported types. The MCP response contains content type, byte count, and `content` with `encoding: utf-8`, or base64 in `content` with `encoding: base64` for invalid UTF-8.
- `multica_attachment_upload`: sends multipart `POST /api/upload-file` with `file` and the actual server `issue_id` or `comment_id` fields. Use `content` for text/Markdown or `content_base64` for binary bytes. Exactly one content field and an issue/comment target are required; local wrapper limit is 8 MiB. `.md` defaults to `text/markdown`.

Both use the configured bearer token and workspace headers; each call may override workspace ID. No local filesystem paths are read. Upload returns the server's JSON attachment response. Read-only mode omits upload. The JSON REST catalog counts remain independent of these two transport primitives.
