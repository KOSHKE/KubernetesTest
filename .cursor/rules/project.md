# Project Rules for KubernetesTest (Cursor IDE)

This directory is the single source of truth for project rules used by Cursor. It encodes project architecture, coding conventions, operational practices, and guardrails to keep changes consistent, minimal (YAGNI), and production-ready.

## Overview
- Monorepo with Go microservices and a React (TypeScript) frontend.
- Services: `api-gateway`, `user-service`, `order-service`, `inventory-service`, `payment-service`.
- Infrastructure: PostgreSQL, Redis, Kafka (KRaft), Kafka UI, Docker Compose for dev, Buf for protobuf.
- Communication: gRPC for request/response, Kafka for domain events.

## Design Principles
- Follow SOLID and clean architecture with hexagonal/DDD layering:
  - `internal/domain`: entities, value objects, domain errors
  - `internal/app`: application services and orchestration
  - `internal/ports`: interfaces (inbound/outbound)
  - `internal/infra`: adapters (DB, gRPC, Kafka, cache, clock, idgen)
- YAGNI: implement the smallest change to satisfy current requirements; avoid speculative abstraction.
- Comments must be in English. Keep them concise and explain “why”, not “how”.
- Avoid magic numbers. Inject timeouts/pool sizes via config/env.
- Deterministic builds: pin Docker images to major.minor (avoid `latest`).

## Languages & Code Style
### Go
- Use `gofmt` and idiomatic naming. Exported APIs must have explicit types; avoid `any` and unsafe casts.
- Control flow: prefer guard clauses/early returns. Avoid deep nesting (>2–3 levels).
- Errors:
  - Use domain-specific errors for business conditions (see `internal/domain/errors`).
  - Wrap errors with context: `fmt.Errorf("context: %w", err)`.
  - Don’t catch errors without meaningful handling; log with structured context.
- Concurrency:
  - Use worker pools with bounded buffers and `context.Context` cancellation.
  - Close channels and `WaitGroup` on shutdown paths.
- Logging:
  - Use `zap` (`SugaredLogger` interface: `Infow`, `Warnw`, `Errorw`).
  - Include keys like `topic`, `partition`, `offset`, `orderID`, etc.

### TypeScript (frontend)
- Strict typing; avoid `any`.
- Keep components small and pure; colocate state.
- Money/currency: operate in minor units; render via helpers in `frontend/src/utils`.

## gRPC
- Services expose gRPC servers under `internal/infra/grpc` and are wired in `internal/app/run.go`.
- Always register health service (`grpc_health_v1`), and use `GracefulStop` on shutdown.
- Clients must use context deadlines/timeouts sourced from config.

## Protobuf & Events
- Source protos in `proto/`; generated Go in `proto-go/` (Buf via Docker).
- Generate stubs with `make proto` (uses Buf Docker image).
- Event schemas under `proto/events/*` must:
  - Use versioned topics: `<boundedContext>.v<major>.<event_name>` (e.g., `orders.v1.order_created`).
  - Include `occurred_at` RFC3339 when applicable.
  - Express money as `int64` minor units and ISO 4217 `currency`.
- Breaking changes: bump topic version; keep old consumers during migrations.

## Kafka (Messaging)
- Use the shared library `libs/kafka` for all Kafka integrations. Do not implement local base consumers/publishers in services.
- Publishers (`libs/kafka`):
  - One producer per process; shared delivery channel; background goroutine logs success/error.
  - `Publish` is non-blocking; delivery is handled in background. For strict delivery semantics, add a dedicated `PublishSync(ctx)` if needed.
  - Close order: `Flush(timeout)` → `close(delivery)` → `Producer.Close()` to avoid losing delivery events and to terminate goroutine cleanly.
  - Attach logger via `WithLogger`.
- Consumers (`libs/kafka`):
  - `RunValueLoop` (value-only) и `RunMetaLoop` (с метаданными). Оба используют worker pool (4) и буфер (128).
  - Optional per-message timeout via `WithHandleTimeout(d)`; cancel called immediately after handler returns.
  - Non-blocking enqueue to worker channel; on overflow, log a warning and drop message (`dropping message due to full buffer`).
  - Always `defer Close()` inside run methods; subscribe to explicit topics; log read errors with context.
- Topics are provisioned declaratively by a one-shot init job in Compose; do not rely on auto-create.

## Databases & Caching
- Use GORM repositories in `internal/infra/repository` behind `internal/ports/repository` interfaces.
- Auto-migrations gated by env (e.g., `AUTO_MIGRATE=true`).
- No hard-coded DSNs; build DSN from config/env. Ping connections on startup.
- Redis for simple caching; namespace keys per service.

## Configuration
- Centralize env parsing in `internal/app/config.go` per service.
- Inject through env:
  - Ports, DB config, Redis URL
  - Kafka brokers and consumer settings (e.g., auto-offset-reset)
  - External service URLs and per-call timeouts
  - Worker pool sizes/timeouts where applicable

## API Gateway
- HTTP facade to gRPC services. Keep handlers thin; delegate to gRPC clients.
- Enforce CORS via env-configured origins. Use structured logging.

## Frontend
- Call API only via the gateway. Base URL from `VITE_API_URL`.
- Keep types in `frontend/src/types`; formatting/help in `frontend/src/utils`.

## Operational Practices
- Docker-only development. Use Compose with bind mounts and hot reload (Air for Go, Vite for FE).
- Make targets:
  - `make dev-up` / `make dev-rebuild` / `make dev-down`
  - `make proto` for protobuf generation
- Prefer Docker-based tooling; avoid host-specific tooling steps.
- Health checks: expose and wire `depends_on` with health/ready conditions in Compose.

## Security & Reliability
- Principle of least privilege for services and data access.
- Secure by design: validate inputs at boundaries (HTTP/gRPC), and sanitize logs.
- Avoid leaking secrets in logs; read secrets via env only.
- Use timeouts and circuit-breaker-like patterns via context and retries where appropriate.

## Version Control & Reviews
- Branches: `feat/<area>-<desc>`, `fix/<area>-<desc>`, `chore/<area>-<desc>`.
- Conventional Commit messages.
- Small PRs with clear descriptions and checklists. At least one review before merge.

## Testing
- Unit-test application services and domain logic; mock ports and adapters.
- Kafka handlers tested as pure functions with fixtures; avoid spinning brokers in unit tests.

## Conventions Specific to This Repo
- Payment status messages: default to "Payment failed" on unsuccessful outcomes; override with success message otherwise.
- Prefer domain-specific errors for business branching instead of generic `errors.Is` across layers.
- Keep Kafka consumer/producer implementations unified across services (naming, logging, buffering, lifecycle).
- Pin Docker images (e.g., `postgres:15.4`, `redis:7.2-alpine`, avoid `latest`).

## Adding a New Feature (Checklist)
1) Domain first: entities/value objects and domain errors.
2) Define/extend ports; keep application services thin.
3) Implement infra adapters (DB, gRPC, Kafka) behind ports using existing base patterns.
4) New event? Add proto in `proto/events`, run `make proto`, provision a versioned topic.
5) Wire dependencies in `internal/app/run.go`: config, health, graceful shutdown.
6) Keep edits minimal and localized (YAGNI). Avoid broad refactors unless necessary.

## AI Guardrails
- Do not change indentation style or whitespace conventions of existing files.
- Do not add new external dependencies without strong justification.
- Avoid cross-cutting refactors unless explicitly requested.
- Keep public APIs stable; prefer incremental edits.
- All code comments in English.
- Use Docker-based commands and Make targets; do not assume local toolchains.
