SHELL := /bin/sh
.PHONY: db-up db-down db-logs test test-race fmt run-control-plane run-agent

db-up:
	docker compose up -d postgres

db-down:
	docker compose down

db-logs:
	docker compose logs -f postgres

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
