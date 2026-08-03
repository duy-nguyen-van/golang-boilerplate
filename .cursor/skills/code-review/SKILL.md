---
name: code-review
description: "Review code quality with technical rigor. Use before PRs, after implementing features, for security audits, or when claiming task completion."
---

# Code Review

Systematic review for golang-boilerplate (Echo, FX, Keycloak JWT, Postgres, Redis).

## Core Principle
**Technical correctness over social comfort. Be honest, specific, suggest fixes.**

## Process

### 1. Edge Case Scouting (First)
- JWT expiry, missing/wrong roles, nil/empty inputs
- DTO leaking sensitive model fields
- Cache TTL vs DB truth

### 2. Priority Order
1. Correctness
2. Security (auth, CSRF, injection, secret exposure)
3. Error handling (`AppError`, no client leaks)
4. Layer boundaries (`go-layering.mdc`)
5. Concurrency / Redis races
6. Performance (N+1, extra Redis calls)
7. Tests + OpenAPI sync (`/api/v1`)

### 3. Route Checks
- New routes under `/api/v1` in `router.go`
- `AuthMiddleware` / `RequireRole` applied correctly

### 4. Go Checks
- Errors handled; context propagated
- Services don't use `echo.Context`
- Handlers use `BaseHandler` patterns

## Feedback Format
- File + issue + WHY + fix snippet
- Critical / High / Medium / Low

See `.cursor/rules/code-review.mdc`.
