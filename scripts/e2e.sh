#!/usr/bin/env bash
set -euo pipefail

cd "$(dirname "$0")/.."
export FLEET_API_TOKEN="${FLEET_API_TOKEN:-01234567890123456789012345678901}"
export DATABASE_URL="${DATABASE_URL:-postgres://eflex:local-dev-only-change-me@localhost:5432/buildfleet?sslmode=disable}"

make db-up
for f in migrations/*.up.sql; do psql "$DATABASE_URL" -v ON_ERROR_STOP=1 -f "$f"; done

go run ./cmd/control-plane &
CP_PID=$!
trap 'kill $CP_PID 2>/dev/null || true' EXIT
sleep 2

NODE_CREDENTIAL="$(CONTROL_PLANE_URL=http://localhost:8080 FLEET_API_TOKEN="$FLEET_API_TOKEN" NODE_NAME=e2e-agent go run ./cmd/agent 2>&1 | sed -n 's/^NODE_CREDENTIAL=//p' | head -1)"
export NODE_CREDENTIAL
echo "Enrolled agent credential: $NODE_CREDENTIAL"

# submit a job
JOB_ID="$(curl -s -X POST http://localhost:8080/api/v1/jobs -H "Authorization: Bearer $FLEET_API_TOKEN" -H 'Content-Type: application/json' -d '{"tenant_id":"tenant-e2e","source_commit":"aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"}' | sed -n 's/.*"id":"\([^"]*\)".*/\1/p')"
echo "Created job $JOB_ID"

# dispatch
curl -s -X POST http://localhost:8080/api/v1/scheduler/dispatch -H "Authorization: Bearer $FLEET_API_TOKEN"

# run agent briefly to complete the job
timeout 10s env CONTROL_PLANE_URL=http://localhost:8080 NODE_CREDENTIAL="$NODE_CREDENTIAL" go run ./cmd/agent || true

echo "Final job status:"
curl -s "http://localhost:8080/api/v1/jobs/$JOB_ID?tenant_id=tenant-e2e" -H "Authorization: Bearer $FLEET_API_TOKEN"
echo