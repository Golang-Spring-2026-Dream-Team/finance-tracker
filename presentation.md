# Finance Tracker — Software Engineering Defense Materials

## 1. Project Overview

A full-stack personal finance management system built with Go (backend) and React/TypeScript (frontend). Supports accounts, transactions, analytics, budgets, recurring payments, and multi-protocol API access (REST + gRPC).

---

## 2. Architecture & Structure

### Clean Layered Architecture
```
cmd/api/main.go          → Entry point (9 lines — delegates to internal/app)
internal/app/app.go      → Dependency injection wiring, route registration
pkg/handler/             → HTTP handlers (request parsing, response formatting)
pkg/service/             → Business logic layer (validation, orchestration)
pkg/repository/          → Data access layer (DB queries, transactions)
pkg/middleware/          → Cross-cutting concerns (auth, rate limiting, CSRF)
pkg/apperror/            → Structured error types with HTTP/gRPC mapping
pkg/auth/                → JWT, refresh tokens, crypto utilities
pkg/cache/               → Redis integration (token blocklist, refresh sessions)
proto/                   → Protocol Buffers definitions (gRPC)
db/                      → SQL migrations + sqlc type-safe queries
```

**Key principle:** Each layer depends only on the layer below it. Handlers → Services → Repositories → Database. No cross-layer shortcuts.

---

## 3. Technology Stack

| Layer | Technology |
|---|---|
| Language | Go 1.25.0 |
| HTTP Framework | Gin (gin-gonic) |
| RPC | gRPC with Protocol Buffers (buf.build tooling) |
| Database | PostgreSQL 15 (pgx/v5 connection pool) |
| Cache | Redis 7 (go-redis/v9) |
| Migrations | Hand-rolled SQL + sqlc for type-safe queries |
| Auth | JWT (golang-jwt/jwt/v5, HS256) + bcrypt |
| API Docs | Swagger/OpenAPI (swaggo), auto-generated |
| Container | Docker multi-stage build, docker-compose (4 services) |
| Frontend | React 19, TypeScript, Vite, TailwindCSS |

---

## 4. Code Metrics

| Metric | Value |
|---|---|
| Go source files | 75 |
| Test files | 21 |
| Test functions | 64 |
| Total Go lines of code | ~14,262 |
| Total test lines | ~4,121 |
| SQL migrations | 9 |
| gRPC service definitions | 1 (TransactionService with 5 RPCs) |
| CI steps | 4 (format check → tests → build) |
| Docker services | 4 (Go API, PostgreSQL, Redis, Frontend) |

---

## 5. Design Patterns Used

### 5.1 Dependency Injection via Interfaces
Services define interface types for their dependencies (e.g., `authUserRepository`, `tokenBlocklist`, `refreshSessionStore`). This makes testing straightforward — every dependency can be mocked.

### 5.2 Repository Pattern
`pkg/repository/` wraps sqlc-generated queries with transaction support (`txPool` interface, `WithTx` pattern). Financial operations (create, transfer, update, delete) are wrapped in DB transactions that atomically update balances.

### 5.3 Structured Error Handling
`pkg/apperror/apperror.go` defines named error types (`VALIDATION_ERROR`, `UNAUTHORIZED`, `INSUFFICIENT_FUNDS`, etc.) with HTTP status codes. Errors map cleanly to both HTTP responses and gRPC status codes via `MapAppErrorToGRPC`.

### 5.4 Thin Entry Point
`cmd/api/main.go` is only 9 lines. All application logic flows through `internal/app/app.go`, keeping the entry point minimal and focused.

### 5.5 Dual Transport (REST + gRPC)
Both HTTP REST and gRPC share the same `pkg/service` layer. The gRPC layer (`internal/grpc/`) is a thin adapter converting between protobuf messages and internal models.

---

## 6. Security Practices

| Practice | Implementation |
|---|---|
| **JWT Access Tokens** | Short-lived (15 minutes), HS256 signing |
| **Refresh Token Rotation** | Long-lived (30 days), stored as **hashed values** in Redis, not plaintext |
| **Atomic Token Rotation** | Lua script in Redis — old tokens cannot be replayed after rotation |
| **CSRF Protection** | Double-submit cookie pattern when `COOKIE_SAMESITE=none` |
| **Rate Limiting** | Redis sliding windows — per-IP and per-email for login endpoints |
| **Token Revocation** | Logout blocks access token in Redis and deletes refresh session |
| **Database Ownership** | All queries filtered by `user_id` — users can only access their own data |
| **Soft Delete** | `deleted_at` pattern for accounts and transactions |

---

## 7. Data Integrity

- **Numeric precision:** Monetary values use `numeric(15,4)` in PostgreSQL and are passed as strings in JSON, avoiding floating-point issues.
- **Transaction safety:** Financial operations wrapped in DB transactions ensuring balance consistency.
- **Cache invalidation:** Analytics cache uses `SCAN + DEL` pattern after any transaction mutation.
- **Nil-safe Redis clients:** All cache/repository operations guard against nil clients.

---

## 8. Testing Strategy

**64 test functions across 21 test files — covering every layer:**

| Layer | Test Files | Coverage |
|---|---|---|
| Handler | 7 files | HTTP request/response validation |
| Service | 6 files | Business logic validation |
| Repository | 1 file | Transaction handling |
| Middleware | 3 files | Auth, CSRF, rate limiting |
| Cache | 1 file | Redis token rotation |
| Auth | 1 file | JWT generation/validation |

**Testing techniques used:**
- `httptest` with gin's `CreateTestContext` for handlers
- Interface-based mocks for service tests
- `miniredis` (in-memory Redis) for cache tests
- Security-specific test scenarios (`auth_handler_security_test.go`)

---

## 9. CI/CD Pipeline

```
Push to main / feature/** / PR
    │
    ├── 1. Verify formatting (gofmt) — enforces code style
    ├── 2. Run all tests (go test ./...)
    ├── 3. Build verification (go build ./cmd/api)
    └── 4. Cache enabled for faster builds
```

**Makefile provides 15 targets:** `run`, `dev` (hot reload), `test`, `fmt`, `sqlc`, `swagger`, `gen`, `docker-up/down`, `docker-run/stop`, `migrate`, `grpc`.

---

## 10. Documentation

- **README.md** — Full stack overview, feature list, API surface, business rules, curl examples
- **docs/api-endpoints.md** — Tabular API reference with request/response formats and error codes
- **docs/db-schema.md** — Detailed table schemas with column types, constraints, and FK behavior
- **Swagger/OpenAPI** — Auto-generated, available at `/docs/index.html` at runtime
- **Finance_Tracker.postman_collection.json** — Ready-to-import Postman collection

---

## 11. Key Software Engineering Principles Demonstrated

| Principle | How It's Applied |
|---|---|
| **Separation of Concerns** | Handlers / Services / Repositories are strictly separated |
| **Single Responsibility** | Each file has one clear purpose; services are focused |
| **Dependency Inversion** | Services depend on interfaces, not concrete implementations |
| **DRY** | Shared error envelope, shared middleware, shared utilities |
| **Testability** | Interface-driven design enables comprehensive mocking |
| **Security First** | Auth, rate limiting, CSRF, and data ownership are built-in |
| **Maintainability** | Clean structure, consistent patterns, comprehensive docs |
| **Extensibility** | Dual transport (REST + gRPC) ready for microservice evolution |
| **DevOps** | Docker Compose, CI pipeline, Makefile automation |

---

## 12. Suggested Presentation Flow

1. **Problem** — Personal finance tracking needs a reliable, secure solution
2. **Architecture** — Show the layered diagram; explain separation of concerns
3. **Technology Choices** — Why Go, PostgreSQL, Redis, React, gRPC
4. **Security** — Walk through the auth flow (JWT → refresh → rotation → logout)
5. **Data Integrity** — Explain transaction safety and numeric precision
6. **Testing** — Show the 64 test functions across 4 layers
7. **CI/CD** — Show the pipeline; emphasize automated quality gates
8. **Code Quality** — 14K lines, clean structure, dual transport, comprehensive docs
9. **Conclusion** — This project demonstrates professional software engineering practices at a university level
