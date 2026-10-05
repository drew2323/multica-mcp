# REST tool domains

Set `MULTICA_MCP_DOMAINS` per MCP server process/connection. Only selected domains are registered in tools/list. This is static connection-level filtering, not dynamic per-agent loading. Restart the connection and refresh the client tool catalog after configuration changes.

- Unset/empty: workspaces,projects,issues,comments,statuses (114 tools).
- `projects,comments`: 20 tools.
- `issues,comments,statuses`: 61 tools; a focused issue-management connection.
- `all`: 319 tools.
- Unknown domains and combining `all` with other domains fail startup.
- Read-only mode further removes mutating tools.

Available domains and counts (write-enabled source catalog): account 12; agents 25; attachments 4; autopilots 15; billing 10; chat 7; comments 10; integrations 5; issues 46; labels 5; notifications 17; plugins 1; projects 10; properties 4; quick-actions 4; runs 5; runtimes 26; skills 14; squads 10; status 11; statuses 5; tasks 2; views 11; wakeups 17; workspaces 43.

`status` means activity/dashboard endpoints; `statuses` means configurable issue statuses. Comments and execution runs nested below issue routes have their own domains; child issues and subscribers remain in issues. Domain filtering does not change tool names, REST semantics, credentials or workspace permissions.

Verification: race-enabled tests and build passed; actual stdio tools/list from a neutral working directory returned 114 (default), 20 (projects,comments), and 319 (all). No production deployment performed.
