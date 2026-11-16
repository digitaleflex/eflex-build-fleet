# E-FLEX Build Fleet

Build infrastructure for E-FLEX: a Control Plane schedules work and independent Build Agents execute builds. Build Fleet is separate from the Runtime Plane.

## Local development

1. Copy `.env.example` to `.env` and set a unique `FLEET_API_TOKEN` (at least 32 characters).
2. Set `POSTGRES_PORT` and the port in `DATABASE_URL` to an available local port.
3. Start PostgreSQL: `docker compose up -d postgres`.
4. Apply migrations in order from `migrations/` (up files only).
5. Run `go mod tidy`.
6. Start the Control Plane: `go run ./cmd/control-plane`.

The Control Plane requires `DATABASE_URL` and `FLEET_API_TOKEN`. `/healthz` is public liveness; `/readyz` checks PostgreSQL. All `/api/v1/*` routes require a Bearer token.

## Build node API

- `POST /api/v1/nodes`: register a node; new nodes start as `pending`.
- `GET /api/v1/nodes`: list up to 200 nodes.
- `GET /api/v1/nodes/{id}`: retrieve a node.
- `POST /api/v1/nodes/{id}/heartbeat`: record heartbeat and transition node to `ready`.

Example:

```sh
curl -X POST http://localhost:8080/api/v1/nodes \
  -H "Authorization: Bearer $FLEET_API_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"node_name":"local-builder-01","agent_version":"dev","capabilities":{"linux-amd64":{"arch":"amd64","os":"linux"}}}'
```

## Scope of this slice

This slice delivers PostgreSQL-backed node registration, listing, retrieval, heartbeat, readiness checks, and token-protected HTTP routes. It does not yet implement agent-specific credentials/mTLS, scheduling, leases, job execution, or build engines. The shared API token is a development bootstrap mechanism, not the final production identity model.

## Jobs and initial scheduler

- `POST /api/v1/jobs`: queue a job. Required fields: `tenant_id`, `source_commit` (40/64 hex chars); optional `idempotency_key`, `engine`, `engine_version`, `requested_resources`.
- `GET /api/v1/jobs?tenant_id=...`: list jobs for a tenant.
- `GET /api/v1/jobs/{id}`: retrieve a job.
- `POST /api/v1/scheduler/dispatch`: atomically claim one queued job and one healthy node, create an attempt and a 60-second lease.

The scheduler uses PostgreSQL row locks with `SKIP LOCKED`, so concurrent dispatch requests cannot claim the same queued job. This first slice does not yet match requested resources to node capabilities; capability-aware placement and lease renewal/expiry recovery are follow-up work. The API token remains an administrative bootstrap credential and must not be exposed to tenants.

## Agent job protocol

See [docs/agent-protocol.md](docs/agent-protocol.md). Agents authenticate with individual node credentials and can poll assignments, acknowledge start, renew leases, and report completion. The protocol records lifecycle transitions; it does not execute builds.
