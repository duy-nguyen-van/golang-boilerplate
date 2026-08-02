---
name: clean-architecture-testing
description: >-
  Applies Clean Architecture boundaries when writing Go unit and integration
  tests, and drives test coverage toward at least 90% for relevant packages.
  Use when implementing or reviewing tests, coverage gaps, mocking strategy,
  integration test layout, or when the user mentions clean Architecture, unit
  test, integrate test, or coverage targets for golang-boilerplate.
disable-model-invocation: true
---

# Clean architecture, unit tests, integration tests, coverage

## Requested focus (verbatim)

clean Architecture, unit test, integrate test, ensure test coverage >= 90%

## Repo alignment

Layer boundaries and import rules: **`.cursor/rules/go-layering.mdc`**. Test mechanics (table-driven, stubs, build tags): **`.cursor/rules/go-testing.mdc`**.

## Clean Architecture → how to test

| Layer | Unit test | Integration test |
|-------|-----------|-------------------|
| **Handlers** (`internal/handlers/`) | `httptest` + Echo context; **stub** services or narrow repo interfaces; assert status/body. | Rare; prefer service-level integration for business rules. |
| **Services** (`internal/services/`) | **Fake** repos/cache/auth dependencies; table-driven success + error paths; no real DB. | Optional: with test DB/container behind `integration` tag if a flow must cross real SQL. |
| **Repositories** (`internal/repositories/`) | Prefer **sqlmock** or thin tests on query builders if used; otherwise defer to integration. | **Testcontainers** / real Postgres (`//go:build integration`), see `internal/db/postgres_integration_test.go` pattern. |
| **Pure utils** (`internal/utils/`) | Full unit coverage; no I/O. | Usually unnecessary. |

**Rule**: Unit tests **must not** require Docker/Postgres unless the file is behind **`//go:build integration`**. Default `go test ./...` stays fast.

## Unit test workflow

1. Identify the **use case** under test (service method or handler action).
2. List **outbound ports** (repo interfaces, cache, HTTP clients); replace with fakes/stubs.
3. Add cases: **happy path**, validation errors, auth/forbidden, not found, downstream failure.
4. Run: `go test -race -cover ./internal/<package>` (narrow package first).

## Integration test workflow

1. Put tests in `*_test.go` with **`//go:build integration`** at the top of the file.
2. Use **testcontainers** or documented env DSN only when necessary; always `t.Parallel()` where safe; **terminate** containers in `defer`.
3. Run explicitly: `go test -tags=integration ./...` (not part of default CI unless workflow is extended).

## Coverage >= 90%

**Target**: **≥ 90% statement coverage** for each **Go package you change** in the PR (not a vague repo average).

1. Generate profile for that package:

   `go test ./internal/services -coverprofile=coverage.out -covermode=atomic`

2. Inspect: `go tool cover -func=coverage.out | tail -1` (package total) or open HTML: `go tool cover -html=coverage.out`.

3. If below 90%: add tests for **unhit branches** (errors, `if` guards, switch defaults), not trivial `return nil` only lines.

4. **Do not** “game” coverage with meaningless asserts on generated code or by excluding large swaths without team agreement.

5. Full suite (CI-like): **`make tests`**; CI also runs race/cover in `.github/workflows/ci.yml`.

## Checklist before merge

- [ ] New/changed **services/handlers** have unit tests with fakes/stubs.
- [ ] Persistence-sensitive behavior has **integration** coverage **or** justified unit-level DB mocks.
- [ ] Changed packages report **≥ 90%** coverage when measured with `-cover` on that package.
- [ ] `//go:build integration` preserved on slow tests; default `./...` remains quick.

## Additional resources

- Deeper architecture context: [../system-architecture/SKILL.md](../system-architecture/SKILL.md)
