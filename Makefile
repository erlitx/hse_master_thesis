APP_ENV ?= dev

ifeq ($(APP_ENV),prod)
	ENV_FILE := .env.prod
else
	ENV_FILE := .env.dev
endif

ifneq (,$(wildcard $(ENV_FILE)))
  include $(ENV_FILE)
  export
endif


up-dev:
	docker compose --env-file .env.dev up -d --force-recreate

up:
	docker compose  up --build -d --force-recreate
	docker compose logs -f

down:
	docker compose down

run-mcp-server: mod
	APP_ENV=dev go run ./cmd/mcp_server/main.go

run-mcp-client: mod
	APP_ENV=dev go run ./cmd/mcp_client/main.go

build-mcp-server: mod
	go build -o bin/mcp_server ./cmd/mcp_server/main.go

build-mcp-client: mod
	go build -o bin/mcp_client ./cmd/mcp_client/main.go

build-all: build-mcp-server build-mcp-client

run-staging: mod
	APP_ENV=staging go run ./cmd/etl_server/main.go

mod:
	go mod tidy

mod-update:
	go get -u all
	go mod tidy

lint:
	golangci-lint run


logs-mcp-server:	
	docker compose logs -f mcp_server --tail=200


# test:
# 	go test -v -cover ./...

