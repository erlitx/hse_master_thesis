# Гибридный аналитический слой DWH на Model Context Protocol (Go)

---

## О проекте

Репозиторий реализует гибридный контур аналитики для корпоративного хранилища данных (DWH): витрины описываются через **dbt** (`manifest.json`), доступ к **ClickHouse** и метаданным витрин идёт только через **MCP-сервер** (Model Context Protocol), а **MCP-клиент** связывает этот контур с языковой моделью **Anthropic Claude**.

Идея работы: LLM не подключается к СУБД напрямую. Все SQL-запросы и чтение схемы проходят через зарегистрированные MCP-инструменты с политикой **read-only**, а семантика витрин подтягивается из dbt-манифеста (модели с `meta.level = bi_data_mart`).

Сборка на **Go 1.24**, HTTP на **chi**, протокол MCP — библиотека **mark3labs/mcp-go**, OLAP — **clickhouse-go/v2**.

### Два исполняемых модуля


| Бинарник     | Назначение                                                          | Порт по умолчанию                  |
| ------------ | ------------------------------------------------------------------- | ---------------------------------- |
| `mcp_server` | MCP JSON-RPC over HTTP, REST для манифеста и Superset               | `8081` (`HTTP_SERVER_PORT`)        |
| `mcp_client` | HTTP API чата с Claude; два режима работы с MCP-сервером (см. ниже) | `8082`–`8083` (`HTTP_CLIENT_PORT`) |


Оба приложения используют слоистую архитектуру: **controller → usecase → adapter**.

---

## Архитектура

В обоих сценариях **MCP-сервер всегда участвует** — к ClickHouse и метаданным dbt обращаются только его tools/resources. Разница в том, **кто оркестрирует** вызовы MCP: собственный Go-клиент или встроенный MCP-коннектор Claude.

```text
                    ┌─────────────────────────────────────────┐
                    │              mcp_server                 │
                    │   tools / resources / prompts           │
                    │         │              │                │
                    │         ▼              ▼                │
                    │   manifest.json   ClickHouse            │
                    └────────▲──────────────▲─────────────────┘
                             │              │
           JSON-RPC          │              │  MCP (native)
      tools/call, …          │              │
                             │              │
┌──────────────┐  Claude API │              │
│  mcp_client  │◄────────────┘              │
│  (gateway)   │  tool_use → наш CallTool    │
└──────┬───────┘                             │
       │                                     │
       │   HTTP /api/v1/chat/gateway         │
       ▼                                     │
  Пользователь ──────────────────────────────┤
       │                                     │
       │  HTTP /api/v1/chat/claude           │
       ▼                                     │
┌──────────────┐   Claude API + mcp_servers  │
│  mcp_client  │─────────────────────────────┘
│  (claude)    │  Claude сам ходит в MCP
└──────────────┘
```

### Два режима на стороне `mcp_client`


| Эндпоинт                    | Кто вызывает MCP-сервер                                         | Роль `mcp_client`                                                                                                                                                                                                              |
| --------------------------- | --------------------------------------------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------ |
| `POST /api/v1/chat/gateway` | **Собственный MCP-клиент** (`adapter/mcpclient`)                | Создаёт и хранит сессию (in-memory), ведёт полную историю диалога, получает от Claude блоки `tool_use`, сам выполняет `tools/call` на MCP-сервере и возвращает Claude `tool_result` (цикл до финального ответа).               |
| `POST /api/v1/chat/claude`  | **Claude напрямую** (native MCP, `mcp_servers` в запросе к API) | Проксирует диалог в Claude API; MCP-сервер указывается через `MCP_SERVER_ADDR`. Вызовы tools/resources выполняет Claude без ручного цикла `tool_use` → `tool_result` в Go-коде. Сессия по-прежнему ведётся на стороне клиента. |


**Потоки на стороне сервера**

- При старте прогревается **кеш dbt-манифеста** из пути `DBT_MANIFEST_PATH` (см. `.env.dev`).
- Регистрируются MCP **tools**, **resources** и **prompts**.
- ClickHouse: ping и read-only запросы только через инструмент `ch_query` (дополнительная проверка префикса в use-case).

---

## Структура репозитория

```text
.
├── cmd/
│   ├── mcp_server/          # Точка входа MCP-сервера
│   └── mcp_client/          # Точка входа MCP-клиента
├── config/                  # Загрузка .env и envconfig
├── internal/
│   ├── mcp_server/          # Сервер: MCP + REST
│   │   ├── app/             # Сборка зависимостей и HTTP
│   │   ├── adapter/         # clickhouse, dbt, superset, nop
│   │   ├── controller/      # HTTP (chi), mcpserver (MCP)
│   │   ├── domain/          # Доменные модели (манифест, QueryResult, …)
│   │   ├── dto/             # DTO для HTTP и парсинга manifest
│   │   └── usecase/         # Бизнес-логика
│   └── mcp_client/          # Клиент: Claude + MCP + сессии
│       ├── app/
│       ├── adapter/         # claude, mcpclient, storage
│       ├── controller/http/
│       ├── domain/
│       ├── dto/
│       └── usecase/
├── pkg/                     # httpserver, logger, clickhouse pool, render
├── utils/                   # Вспомогательные утилиты (JSON)
├── manifest.json            # dbt manifest (путь задаётся через DBT_MANIFEST_PATH)
├── docker/                  # nginx, certbot
├── Dockerfile
├── docker-compose.yml
├── Makefile
└── go.mod
```

---

## MCP-сервер: возможности

### Инструменты (tools)


| Имя               | Описание                                                                        |
| ----------------- | ------------------------------------------------------------------------------- |
| `ch_ping`         | Проверка соединения с ClickHouse                                                |
| `ch_query`        | Read-only SQL: `SELECT`, `SHOW`, `DESCRIBE`, `EXPLAIN` (до 1000 строк в ответе) |
| `list_dwh_models` | Список витрин DWH (имя и описание) из кеша манифеста                            |
| `get_dwh_model`   | Полные метаданные модели по `model_name`                                        |


### Ресурсы (resources)


| URI                         | Содержимое                                         |
| --------------------------- | -------------------------------------------------- |
| `dwh://models`              | Список имён моделей (JSON)                         |
| `dwh://models/{model_name}` | Полное описание одной модели (колонки, refs, meta) |


### Промпты (prompts)


| Имя     | Описание                                                                                         |
| ------- | ------------------------------------------------------------------------------------------------ |
| `greet` | Промпт аналитика DWH: инструкции + аналитический вопрос (`question`), ответ — SQL для ClickHouse |


### REST API (дополнительно к MCP)


| Метод  | Путь                            | Назначение                          |
| ------ | ------------------------------- | ----------------------------------- |
| `GET`  | `/health`                       | Проверка живости                    |
| `GET`  | `/metrics`                      | Метрики Prometheus                  |
| `POST` | `/api/v1/`, `/api/v1/mcp`       | MCP JSON-RPC                        |
| `GET`  | `/api/v1/dbt/manifest`          | Манифест из кеша                    |
| `POST` | `/api/v1/bitool/create_dataset` | Создание датасета в Apache Superset |


---

## MCP-клиент: HTTP API


| Метод  | Путь                   | Назначение                                                                |
| ------ | ---------------------- | ------------------------------------------------------------------------- |
| `GET`  | `/health`              | Проверка живости                                                          |
| `GET`  | `/metrics`             | Метрики Prometheus                                                        |
| `POST` | `/api/v1/chat/gateway` | Диалог: **ручная оркестрация** MCP (сессия + `tools/call` из Go-клиента)  |
| `POST` | `/api/v1/chat/claude`  | Диалог: **Claude ↔ MCP напрямую** (native connector, сессия в Go-клиенте) |


---

## Быстрый старт

### Требования

- Go **1.24+**
- Файл окружения `.env.dev` (или `.env.${APP_ENV}`)
- Для сервера: доступный ClickHouse и файл dbt-манифеста по пути `DBT_MANIFEST_PATH`
- Для клиента: ключ **Claude API** и URL запущенного MCP-сервера

### Установка зависимостей

```bash
cd Source/mcp_server_http
go mod tidy
```

### Запуск MCP-сервера

```bash
APP_ENV=dev go run ./cmd/mcp_server
# или
make run-mcp-server
```

Сервер слушает `HTTP_SERVER_PORT` (по умолчанию `8081`).

### Запуск MCP-клиента

В `.env.dev` должны быть заданы как минимум `CLAUDE_API_KEY` и `MCP_SERVER_ADDR` (полный URL с путём, например `http://localhost:8081/api/v1/mcp`).

```bash
APP_ENV=dev go run ./cmd/mcp_client
# или
make run-mcp-client
```

### Сборка бинарников

```bash
make build-all
# bin/mcp_server, bin/mcp_client
```

### Docker

```bash
make up-dev    # docker compose с .env.dev
make down
```

Сервисы: `mcp_server` (порт 8081 внутри сети) и `nginx` (80/443). Конфигурация прокси — `docker/nginx/default.conf`.

---

## Переменные окружения

Конфигурация загружается в `config/config.go`: при `DOCKER != true` подключается файл `.env.${APP_ENV}`.

### Общие


| Переменная     | Описание                        | По умолчанию |
| -------------- | ------------------------------- | ------------ |
| `APP_ENV`      | Окружение (`dev`, `staging`, …) | `dev`        |
| `APP_NAME`     | Имя приложения                  | `mcp_server` |
| `APP_VERSION`  | Версия                          | `v0.1.0`     |
| `LOGGER_LEVEL` | Уровень логов                   | `debug`      |


### MCP-сервер


| Переменная            | Описание               |
| --------------------- | ---------------------- |
| `HTTP_SERVER_PORT`    | Порт HTTP              |
| `CLICKHOUSE_HOST`     | Хост ClickHouse        |
| `CLICKHOUSE_PORT`     | Порт ClickHouse        |
| `CLICKHOUSE_USER`     | Пользователь           |
| `CLICKHOUSE_PASSWORD` | Пароль                 |
| `CLICKHOUSE_DB`       | База данных            |
| `CLICKHOUSE_SECURE`   | TLS (`true`/`false`)   |
| `SUPERSET_URL`        | URL Superset           |
| `SUPERSET_USER`       | Логин Superset         |
| `SUPERSET_PASSWORD`   | Пароль Superset        |
| `DBT_MANIFEST_PATH`   | Путь к `manifest.json` |


### MCP-клиент


| Переменная          | Описание                                                  |
| ------------------- | --------------------------------------------------------- |
| `HTTP_CLIENT_PORT`  | Порт HTTP клиента                                         |
| `MCP_SERVER_ADDR`   | URL MCP JSON-RPC (**обязательно**, с путём `/api/v1/mcp`) |
| `CLAUDE_API_KEY`    | Ключ Anthropic API (**обязательно**)                      |
| `CLAUDE_MODEL`      | Модель                                                    |
| `CLAUDE_MAX_TOKENS` | Лимит токенов ответа                                      |


Пример фрагмента `.env.dev`:

```bash
APP_ENV=dev
HTTP_SERVER_PORT=8081
HTTP_CLIENT_PORT=8083
MCP_SERVER_ADDR=http://localhost:8081/api/v1/mcp
DBT_MANIFEST_PATH=./manifest.json

CLICKHOUSE_HOST=localhost
CLICKHOUSE_PORT=9000
CLICKHOUSE_USER=default
CLICKHOUSE_PASSWORD=
CLICKHOUSE_DB=default

CLAUDE_API_KEY=sk-ant-...
CLAUDE_MODEL=claude-opus-4-6
```

> **Безопасность:** не коммитьте реальные ключи и пароли. Используйте локальные `.env.`*, исключённые из git.

---

## Проверка MCP через curl

Инициализация:

```bash
curl -sS http://localhost:8081/api/v1/mcp \
  -H 'Content-Type: application/json' \
  -d '{"jsonrpc":"2.0","id":1,"method":"initialize","params":{}}'
```

Список инструментов:

```bash
curl -sS http://localhost:8081/api/v1/mcp \
  -H 'Content-Type: application/json' \
  -d '{"jsonrpc":"2.0","id":2,"method":"tools/list","params":{}}'
```

Вызов `list_dwh_models`:

```bash
curl -sS http://localhost:8081/api/v1/mcp \
  -H 'Content-Type: application/json' \
  -d '{"jsonrpc":"2.0","id":3,"method":"tools/call","params":{"name":"list_dwh_models","arguments":{}}}'
```

Ping ClickHouse:

```bash
curl -sS http://localhost:8081/api/v1/mcp \
  -H 'Content-Type: application/json' \
  -d '{"jsonrpc":"2.0","id":4,"method":"tools/call","params":{"name":"ch_ping","arguments":{}}}'
```

---

## Безопасность и ограничения

- **ClickHouse:** на уровне use-case разрешены только запросы с префиксами `SELECT`, `SHOW`, `DESCRIBE`, `EXPLAIN`. В адаптере результат обрезается (**не более 1000 строк**).
- **dbt:** в MCP попадают только модели с `meta.level = bi_data_mart`.
- Для production рекомендуется дополнительно: сетевые ACL, отдельная роль CH, квоты, аудит вызовов `ch_query`.

---

## Разработка

```bash
make mod          # go mod tidy
make lint         # golangci-lint (если установлен)
make logs-mcp-server
```

### Тесты


| Тест | Назначение |
|------|------------|
| [sql_guard_test.go](internal/mcp_server/service/sql_guard_test.go) | Read-only режим SQL: допустимые префиксы и блокировка запрещённых конструкций |
| [tool_router_test.go](internal/mcp_server/service/tool_router_test.go) | Маршрутизация MCP tool-вызовов по категориям (ClickHouse / метаданные) |
| [gateway_conversation_test.go](internal/mcp_client/usecase/gateway_conversation_test.go) | Gateway-диалог: цикл Claude ↔ MCP (`tool_use` / `tool_result`), сессия |

```bash
go test ./internal/mcp_server/service/... -v
go test ./internal/mcp_client/usecase/... -v
```

---

## Технологический стек

- [Model Context Protocol](https://modelcontextprotocol.io/)
- [mark3labs/mcp-go](https://github.com/mark3labs/mcp-go)
- [ClickHouse Go driver](https://github.com/ClickHouse/clickhouse-go)
- [chi](https://github.com/go-chi/chi)
- [zerolog](https://github.com/rs/zerolog)
- [Anthropic Claude API](https://docs.anthropic.com/)

---

## Лицензия и контакты

Проект создан в рамках **ВКР** Белогорцева Михаила Владимировича.  
По вопросам использования и доработок обращайтесь к автору работы.