# E-FLEX Build Fleet — Phase 1

Premier socle exécutable du Control Plane et du Build Agent.

## Prérequis
- Go 1.25+
- Docker Engine avec Docker Compose

## Démarrage local
```bash
cp .env.example .env
docker compose up -d postgres
go test ./...
go run ./cmd/control-plane
```
Dans un second terminal :
```bash
go run ./cmd/build-agent
```

- Control Plane : http://localhost:8080/healthz
- Build Agent : http://localhost:8081/healthz

## Commandes
- `docker compose up -d postgres` : démarrer PostgreSQL
- `go test ./...` : tests
- `go test -race ./...` : tests avec race detector
- `go run ./cmd/control-plane` / `go run ./cmd/build-agent` : démarrer les services

## Structure
- `cmd/control-plane` : binaire Control Plane
- `cmd/build-agent` : binaire Build Agent
- `internal/config` : configuration
- `internal/httpserver` : health checks et arrêt gracieux
- `api/openapi/v1.yaml` : contrat HTTP versionné
- `migrations` : schéma PostgreSQL initial

## Limites actuelles
Aucun scheduler, gRPC, enrôlement, mTLS, lease applicatif ni moteur de build n'est implémenté. `/readyz` est provisoire et vérifiera PostgreSQL après intégration du pool de connexions. Les secrets de `.env.example` sont réservés au développement local.
