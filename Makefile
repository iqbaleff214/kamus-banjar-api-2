MODULE  := github.com/iqbaleff214/kamus-banjar-api-2
BINARY  := ./tmp/main
MIGRATE := $(shell which migrate 2>/dev/null || echo "docker compose run --rm migrate")
DB_URL  ?= postgres://$(DB_USER):$(DB_PASS)@$(DB_HOST):$(DB_PORT)/$(DB_NAME)?sslmode=disable

include .env
export

.PHONY: run test migrate-up migrate-down seed lint tidy coverage

run:
	air -c .air.toml

test:
	go test ./... -v -race -count=1

coverage:
	go test -coverprofile=coverage.out ./internal/...
	@echo "--- coverage by package ---"
	go tool cover -func=coverage.out | grep -E "^(total|.*/(domain|application)/)"
	@echo ""
	@awk '/\/(domain|application)\// { \
		split($$3, a, "%"); pct = a[1]+0; \
		if (pct < 80) { printf "FAIL: %s has %.1f%% < 80%%\n", $$1, pct; fail=1 } \
	} END { if (fail) exit 1 }' coverage.out && echo "Coverage gate PASSED (>=80%% on domain+application)"

migrate-up:
	migrate -path migrations -database "$(DB_URL)" up

migrate-down:
	migrate -path migrations -database "$(DB_URL)" down

seed:
	go run scripts/seed/main.go

lint:
	golangci-lint run ./...

tidy:
	go mod tidy
