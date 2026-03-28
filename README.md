# MCP Server (Go) - Clean Architecture Template

This project mirrors a typical Go service layout (cmd → config → internal/{app,controller,usecase,adapter} → pkg),
but exposes functionality via an **MCP server** using `github.com/mark3labs/mcp-go` **v0.7.0**.

## What it provides

### Tools
- `add` — add two numbers
- `echo` — echo a string
- `now` — current time in RFC3339Nano (UTC)
- `ch_ping` — ping ClickHouse (requires CLICKHOUSE_DSN)
- `ch_query` — run a read-only ClickHouse query (SELECT/SHOW/DESCRIBE/EXPLAIN)

### Resources
- `time://now` — same as `now`, but as a resource

### Prompts
- `greet` — returns a greeting message (argument: `name`)

## Run (HTTP)

```bash
go mod tidy
go run ./cmd/app
```

Environment variables (optional):

```bash
APP_NAME=mcp_server
APP_VERSION=v0.1.0
APP_ENV=dev
MCP_TRANSPORT=http
MCP_ADDR=:8081
MCP_PATH=/mcp

CLICKHOUSE_DSN=clickhouse://localhost:9000/default?username=default&password=
```

## Test with curl

Initialize:

```bash
curl -sS http://localhost:8081/mcp \
  -H 'Content-Type: application/json' \
  -d '{"jsonrpc":"2.0","id":1,"method":"initialize","params":{}}'
```

List tools:

```bash
curl -sS http://localhost:8081/mcp \
  -H 'Content-Type: application/json' \
  -d '{"jsonrpc":"2.0","id":2,"method":"tools/list","params":{}}'
```

Call add:

```bash
curl -sS http://localhost:8081/mcp \
  -H 'Content-Type: application/json' \
  -d '{"jsonrpc":"2.0","id":3,"method":"tools/call","params":{"name":"add","arguments":{"a":10,"b":20}}}'
```

## Notes

- The MCP server registration lives in `internal/controller/mcpserver`.
- Business logic lives in `internal/usecase`.
- External dependencies (like time) are behind adapters (see `internal/adapter/clock`).
