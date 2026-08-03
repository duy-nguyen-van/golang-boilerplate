# Agent context (golang-boilerplate)

Production-ready Go web service: **Echo**, **Uber FX**, **PostgreSQL**, **Redis**, **Atlas** migrations, **Keycloak** JWT/RBAC, **OpenTelemetry** + New Relic + Sentry. Go module: `golang-boilerplate`.

## Before you ship a change

- **Read project docs first** when changing Keycloak auth, RBAC roles, or OpenTelemetry: see `docs/` (and `rules/docs-knowledge.mdc`).
- Run `make lint` and `make tests` (see `Makefile` for scoped test targets).
- Run `go mod tidy` / `make dep` when dependencies change.
- Run `make swagger-load` when routes, handlers, or swag annotations change.
- Never commit secrets; use `cmd/server/.env.example` as a reference only.

## Where things live (by layer)

| Layer | Path |
|--------|------|
| FX wiring, HTTP server lifecycle | `cmd/server/main.go` |
| Echo routes + middleware stack | `cmd/server/routes/router.go` |
| HTTP handlers | `internal/handlers/` |
| Use cases / orchestration | `internal/services/` |
| DB access (GORM) | `internal/repositories/` |
| Persistence entities | `internal/models/` |
| JSON/API DTOs | `internal/dtos/` |
| Echo middleware | `internal/middlewares/` |
| Structured errors + Echo error middleware | `internal/errors/` |
| Shared constants (codes, roles, pagination) | `internal/constants/` |
| Configuration | `internal/config/` |
| Postgres / DB manager | `internal/db/` |
| Redis cache | `internal/cache/` |
| Keycloak, SES, Stripe, S3/GCS | `internal/integration/` |
| Outbound HTTP (Resty) | `internal/httpclient/` |
| Request-scoped helpers | `internal/request/` |
| Logging / APM / OTel | `internal/logger/`, `internal/monitoring/` |
| Pure utils | `internal/utils/` |
| Small shared packages | `pkg/` (e.g. `pkg/correlationid/`) |
| SQL migrations + checksum | `cmd/migrations/sql/`, `atlas.sum` |
| OpenAPI (swag-generated) | `docs/swagger.yaml`, `docs/swagger.json`, `docs/docs.go` |
| OTel collector config | `deploy/otel/` |

**Import direction**: see `.cursor/rules/go-layering.mdc`.

## HTTP route surface

Routes are registered in `cmd/server/routes/router.go`:

| Surface | Base path | Examples | Auth |
|---------|-----------|----------|------|
| **App API** | `/api/v1` | `/`, `/health/*`, `/users/*`, `/companies/*` | Public health; JWT + roles on users/companies |
| **Ops / docs** | `/` | `/swagger/*` (non-prod, basic auth) | Basic auth |

- swag `@BasePath` is `/api/v1`.
- CSRF is global (`X-CSRF-Token` / `csrf_token` cookie); token also exposed on responses via `ExposeCSRFToken`.

## Key patterns

### Error handling
- Prefer `internal/errors.AppError` and `errors.ErrorHandler` via `handlers.BaseHandler`
- Error codes in `internal/constants/`

### HTTP handlers
- Echo handlers embed `BaseHandler`; wire via FX constructors
- Bind DTOs, call services, return JSON via `HandleError` / `SuccessResponse`
- Add swag annotations on handler methods; regenerate with `make swagger-load`

### Services
- Business logic only; accept `context.Context` and typed params — no `echo.Context`
- May use repos, cache, integration

### Auth / RBAC
- JWT via Keycloak: `internal/integration/auth/` + `internal/middlewares/auth.go`
- Roles in `internal/constants/`; apply with `RequireRole` in `router.go`
- Docs: `docs/RBAC_GUIDE.md`, `docs/KEYCLOAK_RBAC_SETUP.md`

### Logging & observability
- **Global Zap** initialized in `cmd/server/main.go` via `logger.Init()` — use `logger.Log` / `logger.Sugar`, not injected `slog`
- Request logging: `internal/middlewares/logging.go`; Sentry/New Relic/OTel in `internal/monitoring/`
- Correlation ID: `pkg/correlationid/` via `internal/request/context.go` helpers
- OTel docs: `docs/OPENTELEMETRY.md`

### Local dev
```bash
make container-up   # dependencies
make migrate-up     # Atlas migrations
make up             # run server
make bootstrap      # all three
make otel-up        # otel-collector + jaeger
```

## Cursor configuration (`.cursor/`)

### Always-on rules
| Rule | Purpose |
|------|---------|
| `rules/project-core.mdc` | Module, stack, quality bar |
| `rules/development-rules.mdc` | YAGNI/KISS/DRY, pre-commit |
| `rules/openapi-specs.mdc` | swag sync on HTTP changes |
| `rules/security-practices.mdc` | JWT/Keycloak, CSRF, secrets, SQL |
| `rules/docs-knowledge.mdc` | Read RBAC/Keycloak/OTel docs before related changes |
| `rules/git-conventions.mdc` | Commits, branches, PRs |

### Contextual rules (by glob / request)
| Rule | When |
|------|------|
| `rules/go-layering.mdc` | Editing `**/*.go` |
| `rules/golang-patterns.mdc` | Editing `**/*.go` |
| `rules/go-testing.mdc` | Editing `**/*_test.go` |
| `rules/database-migrations.mdc` | Editing `cmd/migrations/**/*.sql` |
| `rules/code-review.mdc` | Code review / PRs |
| `rules/debugging-guide.mdc` | Debugging |
| `rules/documentation-standards.mdc` | Doc updates |
| `rules/workflow-implementation.mdc` | Feature / fix workflow |

### Skills
| Skill | Use when |
|-------|----------|
| `skills/code-review/SKILL.md` | Systematic review before PRs |
| `skills/debug/SKILL.md` | Root-cause debugging |
| `skills/system-architecture/SKILL.md` | Architecture, DDD, HA, performance |
| `skills/clean-architecture-testing/SKILL.md` | Clean Architecture, coverage ≥ 90% |

### Agents & commands
- **Agents**: `.cursor/agents/` — `code-reviewer`, `debugger`, `tester`, `git-manager`
- **Commands**: `.cursor/commands/gb:*` — `plan`, `cook`, `fix`, `test`, `review-code`
- **MCP example**: `.cursor/mcp.json.example` (copy to `mcp.json` locally; gitignored)

## Human docs

| Doc | Purpose |
|-----|---------|
| `README.md` | Setup, architecture, make targets |
| `CONTRIBUTING.md` | Contribution expectations |
| `docs/RBAC_GUIDE.md` | Roles and authorization |
| `docs/RBAC_QUICK_START.md` | Quick RBAC reference |
| `docs/KEYCLOAK_RBAC_SETUP.md` | Keycloak realm/client setup |
| `docs/OPENTELEMETRY.md` | OTel exporter, collector, Jaeger |
| `docs/swagger.yaml` | OpenAPI for `/api/v1` |
