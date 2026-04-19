# MCP over HTTP (Go)

This repository ships two binaries: an **MCP server** (JSON-RPC over HTTP, plus a small REST surface) and an **MCP client** HTTP service that talks to Claude and optionally to the MCP server. Both follow a layered layout (controller → usecase → adapters).

## Project layout

```text
.
├── cmd/
│   ├── mcp_server/main.go    # MCP server + HTTP API (see HTTP_SERVER_PORT)
│   └── mcp_client/main.go    # Chat / gateway HTTP API (see HTTP_CLIENT_PORT)
├── config/
│   └── config.go             # Env-based config for both apps
├── internal/
│   ├── mcp_server/           # MCP server domain
│   │   ├── app/              # Wiring: adapters, router, HTTP server lifecycle
│   │   ├── adapter/          # clickhouse, clock, dbt (+ manifest cache/parser), nop
│   │   ├── controller/
│   │   │   ├── http/         # Chi: /health, /metrics, /api/v1/*, MCP handler
│   │   │   └── mcpserver/    # MCP tools / resources / prompts registration
│   │   ├── domain/           # e.g. DBT manifest models
│   │   ├── dto/              # Transport shapes for manifest / MCP resources
│   │   └── usecase/          # Business logic used by HTTP + MCP layers
│   └── mcp_client/           # MCP-aware client / gateway
│       ├── app/
│       ├── adapter/          # claude, mcpclient, storage (memory, postgres)
│       ├── controller/http/  # Chi: /health, /metrics, /api/v1/chat/*
│       ├── domain/
│       ├── dto/
│       └── usecase/
├── pkg/                      # Shared libraries
│   ├── httpserver/           # HTTP server wrapper
│   ├── logger/
│   ├── clickhouse/           # Pool / shared CH config
│   └── render/               # JSON / error helpers
├── utils/                    # Small shared helpers (e.g. JSON)
├── docker/
│   ├── nginx/                # Reverse proxy config for compose
│   └── certbot/              # Optional TLS material (mounted by compose)
├── manifest.json             # DBT manifest path used by mcp_server (see app wiring)
├── Dockerfile
├── docker-compose.yml        # mcp_server + nginx
├── Makefile                  # run-mcp-server, run-mcp-client, compose targets, …
├── Taskfile.yml
├── go.mod
└── go.sum
```

**Where things live**

- MCP protocol registration (tools, resources, prompts): `internal/mcp_server/controller/mcpserver`.
- REST + MCP HTTP routes for the server: `internal/mcp_server/controller/http` (MCP POST handler is mounted at `/api/v1/` and `/api/v1/mcp`).
- Client chat endpoints: `internal/mcp_client/controller/http` under `/api/v1/chat/`.

## What the MCP server exposes

### Tools

- `ch_ping` — ping ClickHouse (requires `CLICKHOUSE_DSN` / pool config)
- `ch_query` — read-only ClickHouse query (`SELECT` / `SHOW` / `DESCRIBE` / `EXPLAIN`)
- `list_dwh_models` — list DWH model names and descriptions from the cached manifest
- `get_dwh_model` — full model metadata by `model_name` from the cached manifest

### Resources

- `time://now` — current time (RFC3339Nano, UTC)
- `dwh://models` — bulk JSON list of models
- `dwh://models/{model_name}` — single model (from manifest)

### Prompts

- `greet` — greeting message (argument: `name`)

## Run locally

```bash
go mod tidy
APP_ENV=dev go run ./cmd/mcp_server
```

Client (requires `CLAUDE_API_KEY`, `MCP_SERVER_ADDR`, and `.env.dev` or equivalent):

```bash
APP_ENV=dev go run ./cmd/mcp_client
```

Makefile shortcuts:

```bash
make run-mcp-server
make run-mcp-client
```

## Environment variables (common)

**Server / MCP**

```bash
APP_NAME=mcp_server
APP_VERSION=v0.1.0
APP_ENV=dev
HTTP_SERVER_PORT=8081          # Chi + MCP handler listen port

CLICKHOUSE_DSN=clickhouse://localhost:9000/default?username=default&password=
```

The `config.Config.MCP` block (`MCP_ADDR`, `MCP_PATH`, `MCP_TRANSPORT`) is defined for compatibility; routing today is fixed under `/api/v1/` (see `internal/mcp_server/controller/http/router.go`).

**Client**

```bash
HTTP_CLIENT_PORT=8082
# Full URL for POST JSON-RPC (must include path), e.g.:
MCP_SERVER_ADDR=http://localhost:8081/api/v1/mcp
CLAUDE_API_KEY=...
CLAUDE_MODEL=claude-opus-4-6
```

Config loading: `config/config.go` loads `.env.${APP_ENV}` when not running in Docker (`DOCKER=true` skips dotenv).

## Test MCP with curl

MCP JSON-RPC is served with **POST** on `/api/v1/mcp` (or `/api/v1/`) on the server HTTP port.

Initialize:

```bash
curl -sS http://localhost:8081/api/v1/mcp \
  -H 'Content-Type: application/json' \
  -d '{"jsonrpc":"2.0","id":1,"method":"initialize","params":{}}'
```

List tools:

```bash
curl -sS http://localhost:8081/api/v1/mcp \
  -H 'Content-Type: application/json' \
  -d '{"jsonrpc":"2.0","id":2,"method":"tools/list","params":{}}'
```

Example: call `list_dwh_models`:

```bash
curl -sS http://localhost:8081/api/v1/mcp \
  -H 'Content-Type: application/json' \
  -d '{"jsonrpc":"2.0","id":3,"method":"tools/call","params":{"name":"list_dwh_models","arguments":{}}}'
```

## Docker

```bash
make up-dev    # docker compose with .env.dev
```

Services: `mcp_server` (app on 8081 inside the network) and `nginx` (80/443 with optional certbot volumes). See `docker-compose.yml` and `docker/nginx/default.conf`.

## Notes

- The server warms a **DBT manifest cache** at startup from `./manifest.json` (see `internal/mcp_server/app/app.go`).
- Prometheus metrics: `GET /metrics` on each binary’s HTTP port.
