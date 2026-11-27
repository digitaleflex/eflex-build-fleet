# AGENTS.md

Go 1.25 modular monolith: Control Plane (`cmd/control-plane`), Build Agent (`cmd/agent`), placeholder health server (`cmd/build-agent`). PostgreSQL is the only datastore. Build Fleet is separate from the E-FLEX Runtime Plane.

## Commands

- `make db-up` / `make db-down` / `make db-logs`: PostgreSQL via Docker Compose.
- `make test` = `go test ./...` (unit tests; no DB required).
- `make test-race` = `go test -race ./...`.
- `make fmt` = `gofmt -w` over all `.go` files.
- `make run-control-plane` = `go run ./cmd/control-plane`.
- `make run-agent` = `go run ./cmd/build-agent`.
- Single package/test: `go test ./internal/controlplane/ -run TestName`.
- Integration tests (build tag): `go test -tags integration ./...` (requires `DATABASE_URL` and applied migrations).
- Lint/type-ish checks: `go vet ./...`; CI also runs `go mod tidy`.

CI runs `go mod tidy`, `gofmt -w`, `go test -race ./...`, `go vet ./...` on pushes to `main`, `feat/**`, `fix/**` and PRs to `main`. Keep code gofmt-clean and `go.mod` tidy before committing.

## Local setup gotchas

- `.env` is **not** auto-loaded; config reads `os.Getenv`. Export vars manually, e.g. `set -a; source .env; set +a`.
- `POSTGRES_PORT` (host port in Compose) and the port in `DATABASE_URL` must agree; default `5432`.
- PostgreSQL must be running and migrations applied **manually** before the Control Plane starts; there is no auto-migration. Apply up files in order:
  `for f in migrations/*_up.sql; do psql "$DATABASE_URL" -v ON_ERROR_STOP=1 -f "$f"; done`
- Control Plane requires `DATABASE_URL` and `FLEET_API_TOKEN` (>= 32 chars). `/healthz` is public, `/readyz` checks PostgreSQL, `/api/v1/*` requires Bearer token.
- Agent enrollment: `cmd/agent` registers a node using `FLEET_API_TOKEN`, issues a node credential, and prints `NODE_CREDENTIAL=...` once. Store it; credentials are shown only at issuance. With `NODE_CREDENTIAL` set, enrollment is skipped.
- `CONTROL_PLANE_URL` defaults to `http://localhost:8080`.

## Architecture

- Entry points: `cmd/control-plane`, `cmd/agent`, `cmd/build-agent`.
- Core logic: `internal/controlplane` (nodes, jobs, scheduler, credentials, leases, rate limits, metrics).
- Generic health server: `internal/httpserver`.
- Build execution: `internal/buildrunner`.
- Config: `internal/config`.
- API contract: `api/openapi/v1.yaml`.
- Agent protocol: `docs/agent-protocol.md` (fencing token header `X-Fencing-Token` required on `/started`, `/lease`, `/complete`; node identity from credential).
- Migrations: `migrations/*_up.sql` / `*_down.sql`.

## Conventions / constraints (from CONTRIBUTING.md)

- Do not introduce a language, service, DB, broker, framework, or engine without justification.
- Do not contradict an Accepted ADR without a replacement ADR; architecture/boundary/protocol changes require an ADR.
- Never commit secrets or real config; pin versions of structural dependencies; verify licenses.
- Keep API contracts independent of SQL models.
- No public BuildKit or agent endpoints; no runtime access for a Build Agent; never share a write cache across tenants.
- Security/recovery changes must include tests; docs describing changed behavior must update in the same change.
- Errors explicit, structured secret-free logs, important ops idempotent, persisted state, minimal permissions.