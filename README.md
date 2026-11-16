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
