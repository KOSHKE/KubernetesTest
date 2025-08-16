# Microservices Order System (Dev)

Monorepo with Go microservices and React (TypeScript). Dev environment is powered by Docker Compose, hot reload via Air (Go) and Vite (frontend).

## Quick Start

1) Create `.env` from `.env.example` and adjust values if needed
2) Generate protobuf and start dev stack
```bash
make dev-up
```
3) Optional: force rebuild images after Dockerfile changes
```bash
make dev-rebuild
```
4) Stop stack
```bash
make dev-down
```

Frontend: `http://localhost:3001`
API Gateway: `http://localhost:8080`

## What’s inside

- Go services: `api-gateway`, `user-service`, `order-service`, `inventory-service`, `payment-service`
- React + TypeScript frontend (Vite dev server)
- PostgreSQL + Redis
- Air for Go hot reload (preinstalled in Dockerfile)

## Dev notes

- Source code is bind-mounted into containers for hot reload
- Protobuf stubs are generated via `make proto` (Buf in Docker)
- Env variables are managed via `.env` and injected into Docker Compose

## Make commands

```bash
make help         # list commands
make proto        # generate protobuf stubs
make proto-clean  # clean generated protobuf stubs
make dev-up       # start dev environment (no rebuild)
make dev-rebuild  # rebuild images and start dev environment
make dev-down     # stop dev environment
```

## Troubleshooting

- Changed Dockerfile but container still runs old command
  - Run `make dev-rebuild`

- Air reports duplicated tables in .air.toml
  - Ensure each `.air.toml` has only one set of sections (`root`, `build`, `log`)

- Frontend cannot reach API
  - Check `VITE_API_URL` in `.env` matches API Gateway
