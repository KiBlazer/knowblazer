# 01: Expand MCP JSON-RPC payload buffer to 10MB

**What to build:** Allow the MCP server to receive and process JSON-RPC requests larger than 64KB up to 10MB without terminating the scanner, verified by handling oversized input payloads.

**Blocked by:** None (can start immediately)

**Status:** ready-for-agent

- [x] MCP scanner is configured with a 10MB maximum token capacity
- [x] JSON-RPC request with payload > 64KB is handled and returns a valid response
- [x] `go test ./internal/mcp` passes
