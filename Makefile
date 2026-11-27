SHELL := /bin/sh
.PHONY: db-up db-down db-logs migrate test test-race test-integration fmt run-control-plane run-agent

db-up:
	docker compose up -d postgres

db-down:
	docker compose down

db-logs:
	docker compose logs -f postgres

migrate:
	for f in migrations/*.up.sql; do psql "$$DATABASE_URL" -v ON_ERROR_STOP=1 -f "$$f"; done

test-integration:
	go test -tags integration -race ./...

test:
	go test ./...

test-race:
	go test -race ./...

fmt:
	gofmt -w $$(find . -name '*.go' -not -path './vendor/*')

run-control-plane:
	go run ./cmd/control-plane

run-agent:
	go run ./cmd/build-agent
