---
name: debug
description: "Debug systematically with root cause analysis before fixes. Use for bugs, test failures, unexpected behavior, performance issues, log analysis, CI/CD failures, database diagnostics."
---

# Debugging (backend)

## Core Principle
**NO FIXES WITHOUT ROOT CAUSE INVESTIGATION FIRST**

## When to Use
- Test failures, auth/RBAC bugs
- Cache inconsistencies
- Migration or DB errors
- OTel / observability issues
- CI failures (`make lint`, `make tests`)

## Framework

### Phase 1: Root Cause
1. Read error messages and stack traces
2. When did it start? (`git log`, recent deploy/config)
3. Scope: all users vs one endpoint vs one env
4. Form 2–3 hypotheses

### Phase 2: Evidence
- `go test -race ./internal/<package>`
- `make test-specific TEST=...`
- Manual repro against `/api/v1/...`
- Postgres: `make migrate-status`; Redis: cache keys/TTL
- Logs: global Zap `logger.Sugar`; correlation ID via `internal/request/context.go`
- Traces: Jaeger when OTel stack is up (`make otel-up`)

### Doc references
| Doc | Use for |
|-----|---------|
| [docs/RBAC_GUIDE.md](../../../docs/RBAC_GUIDE.md) | Roles, `RequireRole` |
| [docs/KEYCLOAK_RBAC_SETUP.md](../../../docs/KEYCLOAK_RBAC_SETUP.md) | JWT / Keycloak |
| [docs/OPENTELEMETRY.md](../../../docs/OPENTELEMETRY.md) | Traces, collector |

### Phase 3: Fix & Verify
- Minimal fix at root layer
- Regression test in colocated `*_test.go`
- `make tests`

## Rules
- Never guess — use evidence
- Fix root cause, not symptoms
- Always write a regression test when fixing bugs

See `.cursor/rules/debugging-guide.mdc`.
