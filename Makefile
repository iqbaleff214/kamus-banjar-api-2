MODULE  := github.com/iqbaleff214/kamus-banjar-api-2
DB_URL  ?= postgres://$(DB_USER):$(DB_PASS)@$(DB_HOST):$(DB_PORT)/$(DB_NAME)?sslmode=disable

include .env
export

.PHONY: \
	run up down logs ps \
	migrate-up migrate-down \
	seed \
	test coverage lint tidy \
	prod-up prod-down prod-logs prod-ps prod-seed

# ─── Local dev (Docker Compose) ───────────────────────────────────────────────

run:
	docker compose up --build

up:
	docker compose up -d --build

down:
	docker compose down

logs:
	docker compose logs -f

ps:
	docker compose ps

# ─── Migrations ───────────────────────────────────────────────────────────────

migrate-up:
	docker compose run --rm migrate

migrate-down:
	docker compose run --rm migrate down 1

# ─── Seeder ───────────────────────────────────────────────────────────────────

seed:
	go run scripts/seed/main.go

# ─── Tests / quality ──────────────────────────────────────────────────────────

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

lint:
	golangci-lint run ./...

tidy:
	go mod tidy

# ─── Production ───────────────────────────────────────────────────────────────

prod-up:
	docker compose -f docker-compose.prod.yml up -d --build

prod-down:
	docker compose -f docker-compose.prod.yml down

prod-logs:
	docker compose -f docker-compose.prod.yml logs -f

prod-ps:
	docker compose -f docker-compose.prod.yml ps

prod-seed:
	docker compose -f docker-compose.prod.yml --profile seed run --rm seed
