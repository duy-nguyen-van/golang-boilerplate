---
name: system-architecture
description: >-
  Guides system and software architecture for Go services on PostgreSQL using
  domain-driven layering, performance, high availability, and sustainable
  operations. Use when designing or reviewing architecture, scalability, DDD
  boundaries, HA/disaster recovery, Postgres schema strategy, performance
  budgets, observability, or when the user mentions Golang, postgres, domain
  driven design, high performance, high availability, or sustainability in an
  architectural sense for golang-boilerplate.
disable-model-invocation: true
---

# System architecture (Go · Postgres · DDD · performance · HA · sustainability)

## When this skill applies

User wants **design decisions**, **trade-off analysis**, **layering reviews**, or **non-functional requirements** (latency, throughput, uptime, operability) — not a single-line code fix unless framed as architectural impact.

## Tech stack focus

Golang, postgres, domain driven design, high performance, high availability, sustainable

## Map DDD concepts to this repo

| DDD idea | Where it lives here | Guidance |
|----------|---------------------|----------|
| **Bounded context** | Users, companies, auth/RBAC, integrations (email/payment/storage) | Prefer **explicit boundaries** between contexts: separate services/packages, avoid cross-context “god” services. |
| **Application / use cases** | `internal/services/` | Orchestration, transactions, policies, calls to repos and integrations. **No** Echo types; use `context.Context`. |
| **Domain model & rules** | `internal/models/` + pure functions in services | Keep **invariants** near the model or small helpers; avoid duplicating the same rule in handlers and repos. |
| **Infrastructure** | `internal/repositories/`, `internal/db/`, `internal/integration/`, `internal/cache/` | I/O and vendor SDKs; **interfaces** at boundaries when testing or swapping implementations. |
| **API / adapters in** | `internal/handlers/`, `internal/dtos/` | Translate HTTP ↔ DTOs ↔ service inputs; thin. |

Cross-check concrete paths with `.cursor/rules/go-layering.mdc` before suggesting new packages.

## High performance (Go)

- **Hot paths**: minimize allocations, avoid N+1 queries — use **batch loads** or **JOINs** via GORM thoughtfully.
- **Concurrency**: use `context` deadlines/cancellation; prefer **bounded** worker patterns over unbounded goroutines for background work.
- **HTTP server**: tune read/write/header timeouts; align with reverse-proxy timeouts.
- **Evidence**: recommend **pprof** / OTel traces for concrete bottlenecks instead of guessing.

## PostgreSQL

- **Schema**: explicit constraints (FK, unique, check), narrow indexes for real query patterns; document breaking changes in **Atlas** migrations (`cmd/migrations/sql/`, `make migrate-hash`).
- **Connections**: size pool to `(instances × pool per instance) < Postgres max_connections`; use **timeouts** on queries where appropriate.
- **Scale-out read path**: if read replicas are introduced later, route **read-only** queries explicitly; do not assume replica lag is zero for auth-critical reads unless designed for it.

## High availability

- **Stateless app tier**: JWT/RBAC logic should assume **multiple replicas**; ephemeral state in **Redis** (`internal/cache/`) is appropriate for shared ephemeral data — design **TTL** and **idempotency** for writes.
- **Database**: single-primary Postgres is common; HA means backups, **PITR**, runbooks, and failover strategy — mention RPO/RTO when discussing “HA” for data.
- **Graceful degradation**: fail closed for security; for non-critical paths, return structured errors and log with correlation IDs (`pkg/correlationid/`).
- **Lifecycle**: honor **graceful shutdown** (HTTP drain + DB close) — align new long-running tasks with `fx` lifecycle hooks.

## Sustainable engineering (operability & longevity)

- **Observability**: structured logs (Zap), OTel traces/metrics, errors to Sentry, APM where enabled — tie new features to **actionable** metrics/logs, not noise.
- **Complexity budget**: prefer **boring** patterns that match existing layers over new frameworks; document decisions in `docs/` only when non-obvious.
- **Security as NFR**: secrets in env, no credentials in repo; rate limits and CSRF already in middleware — extend consistently.

## Architecture response checklist

When answering an architecture question, prefer this shape:

1. **Context**: what part of the system (context boundary) is affected.
2. **Options**: 2–3 approaches with trade-offs (performance, HA, complexity, security).
3. **Recommendation**: one default with **why**.
4. **Repo alignment**: which directories/layers change; migration or rollout notes if any.

## Do not

- Invent deployment topology (K8s regions, exact replica counts) without user input — ask or state assumptions explicitly.
- Bypass layering rules (e.g. business rules only in handlers) without calling out the debt.
