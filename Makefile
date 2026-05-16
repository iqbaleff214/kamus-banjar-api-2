MODULE  := github.com/iqbaleff214/kamus-banjar-api-2
BINARY  := ./tmp/main
MIGRATE := $(shell which migrate 2>/dev/null || echo "docker compose run --rm migrate")
DB_URL  ?= postgres://$(DB_USER):$(DB_PASS)@$(DB_HOST):$(DB_PORT)/$(DB_NAME)?sslmode=disable

include .env
export

.PHONY: run test migrate-up migrate-down seed lint tidy

run:
	air -c .air.toml

test:
	go test ./... -v -race -count=1

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
