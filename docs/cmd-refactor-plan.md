# Engineering Plan: Refactor & Integration Testing

## Part 1 — cmd/ferrogw Refactor

Move business logic out of `cmd/ferrogw/` into `internal/` packages.
Goal: `main.go` becomes ~80 lines (Cobra wiring + `os.Exit`).

Current state: ~1,870 lines across 11 production files, all `package main`.

### New internal packages

```
internal/
├── apierror/       # OpenAI-format error responses
├── bootstrap/      # Store factories, gateway construction, startup
├── dashboard/      # Template rendering, pprof, static assets
├── handler/        # All /v1/* HTTP handlers
├── httpserver/     # Server constructor, conn tracker, router
├── middleware/     # CORS, rate-limit, proxy auth
├── proxy/          # Pass-through proxy + model scanner
├── sse/            # SSE streaming writer
└── testutil/       # Shared test helpers (containers, fixtures)
```

### Phase 1 — Self-contained moves ✅ DONE

| File | Destination | Lines |
|------|-------------|-------|
| `cors.go` | `internal/middleware/cors.go` | 50 |
| `sse.go` + test | `internal/sse/sse.go` | 151 |
| `writeOpenAIError`, `routeErrorDetails` from `http_helpers.go` | `internal/apierror/apierror.go` | ~60 |
| `store_init.go` | `internal/bootstrap/stores.go` | 114 |

### Phase 2 — HTTP server + observability ✅ DONE

| File | Destination | Lines |
|------|-------------|-------|
| `server.go` | `internal/httpserver/server.go` | 53 |
| `server_observability.go` + test | `internal/httpserver/conntracker.go` | 116 |

### Phase 3 — Proxy (performance-critical, keep benchmarks) ✅ DONE

| File | Destination | Lines |
|------|-------------|-------|
| `proxy.go` + test | `internal/proxy/proxy.go` | 320 |

### Phase 4 — Handlers ✅ DONE

| File | Destination | Lines |
|------|-------------|-------|
| `chat_request.go` | `internal/handler/chatrequest.go` | 164 |
| `embeddings.go` | `internal/handler/embeddings.go` | 39 |
| `images.go` | `internal/handler/images.go` | 39 |
| `models_handler.go` | `internal/handler/models.go` | 84 |
| `completions.go` + test | `internal/handler/completions.go` | 198 |

### Phase 5 — Router + remaining middleware ✅ DONE

| Source | Destination | Lines |
|--------|-------------|-------|
| `rateLimitMiddleware` from `http_helpers.go` | `internal/middleware/ratelimit.go` | ~40 |
| `proxyAuth` from `router_routes.go` | `internal/middleware/proxyauth.go` | ~20 |
| Template rendering + dashboard from `http_helpers.go` + `router_routes.go` | `internal/dashboard/dashboard.go` | ~100 |
| `router.go` + `router_routes.go` | `internal/httpserver/router.go` | 322 |

### Phase 6 — Bootstrap + thin main ✅ DONE

| Source | Destination | Lines |
|--------|-------------|-------|
| `buildGateway`, `registerProviders`, `loadConfig`, `printStartupBanner`, `newRateLimitStore` from `main.go` | `internal/bootstrap/bootstrap.go` | ~300 |

Result: `cmd/ferrogw/main.go` reduced to **59 lines** (Cobra wiring + plugin blank imports only).

### Refactor rules

- One phase per commit — must compile and pass `go test ./...` after each
- Move tests alongside production code
- No behavior changes — pure mechanical moves with package/export renames
- Export only what callers need (keep internal surface minimal)
- Preserve `topLevelModelScanner` benchmarks in Phase 3

---

## Part 2 — Integration Testing (testcontainers-go)

### Problem

Postgres integration tests require a manual `FERROGW_TEST_POSTGRES_DSN` env var and skip when unset. Postgres paths are untested in CI and on most developer machines.

### Solution

Use [testcontainers-go](https://github.com/testcontainers/testcontainers-go) to spin up real Postgres (and optionally Redis for v1.1.0) containers on demand during `go test`.

### Dependency

```
go get github.com/testcontainers/testcontainers-go
go get github.com/testcontainers/testcontainers-go/modules/postgres
```

### Phase 7 — Postgres persistence ✅ DONE

All integration tests live in `test/integration/` with a single `TestMain` that starts one shared Postgres container.

| Test file | What it covers |
|-----------|---------------|
| `admin_store_test.go` | Full `Store` contract against real Postgres: CRUD, validate+usage, expiration, revoke, rotate, list masking |
| `config_store_test.go` | `SQLConfigStore` + `GatewayConfigManager`: save/load roundtrip, delete, reload persists |
| `requestlog_test.go` | `SQLWriter`: write+list, pagination, delete by timestamp |
| `bootstrap_test.go` | `CreateKeyStoreFromEnv`, `CreateRequestLogReaderFromEnv`, `CreateConfigManagerFromEnv` with `backend=postgres` |

### Phase 8 — Gateway E2E (full HTTP stack)

| Package | Test file | What to cover |
|---------|-----------|---------------|
| `test/integration` | `gateway_test.go` | Start ferrogw binary + Postgres container, hit `/health`, `/v1/models`, `/admin/keys`, `/v1/chat/completions` (mock provider), verify request logs persisted in Postgres |

### Phase 9 — Redis (future, v1.1.0)

| Package | Test file | What to cover |
|---------|-----------|---------------|
| `internal/ratelimit` | `redis_integration_test.go` | Rate limit state shared across gateway instances |
| `internal/cache` | `redis_integration_test.go` | Auth token caching |

### Design

- **`test/integration/`** — separate package, tests use public API only
- **Single `TestMain`** — starts one Postgres container shared by all tests
- **`testing.Short()`** — `go test -short` skips integration tests; no build tags needed
- **Graceful Docker skip** — if Docker is unavailable, tests skip instead of failing
- **`internal/testutil/postgres.go`** — shared helper with panic recovery for environments without Docker

### Makefile target

```makefile
test-integration-containers:
	go test -v -race -timeout 120s ./test/integration/...
```

### File layout

```
internal/
└── testutil/
    └── postgres.go              # StartPostgres(), Terminate()
test/
└── integration/
    ├── main_test.go             # TestMain: container lifecycle + short skip
    ├── helpers_test.go          # truncateTable()
    ├── admin_store_test.go      # key store CRUD, validate, rotate, revoke
    ├── config_store_test.go     # config store save/load/delete, manager reload
    ├── requestlog_test.go       # request log write/list/paginate/delete
    └── bootstrap_test.go        # factory functions with backend=postgres
```
