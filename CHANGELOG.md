# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

---

## [Unreleased]

### Added
- Phase 2: Dictionary bounded context
  - PostgreSQL schema: `words`, `definitions`, `examples`, `word_relations` tables
  - Full-text search via PostgreSQL GIN indexes (`plainto_tsquery`)
  - sqlc-generated type-safe queries for all dictionary operations
  - `PostgresWordRepository` with soft-delete, pagination, and filtering
  - `WordQueryService`: `GetWord`, `ListWords`, `SearchWords`
  - `WordCommandService`: `CreateWord`, `UpdateWord`, `SoftDeleteWord`
  - Public REST endpoints: `GET /api/v2/words`, `GET /api/v2/words/:id`, definitions, examples, related
  - Admin endpoints: `POST/PATCH/DELETE /api/v2/admin/words`
  - Go seeder: idempotent import of `seed_data.json` into PostgreSQL
  - Unit and handler tests for all dictionary layers

### Changed
- CI workflow: added coverage reporting and build verification job

---

## [0.1.0] — 2026-05-01

### Added
- Phase 1: Project scaffold and Identity bounded context
  - Multi-stage Dockerfile (dev with Air hot-reload, prod on scratch)
  - Docker Compose for dev (app + PostgreSQL 15 + Redis 7) and prod
  - `pkg/config`: environment-based configuration with validation
  - `pkg/auth`: JWT HS256 access tokens (15 min TTL) + Redis refresh tokens (7 days)
  - `pkg/httperr`: consistent error/success response envelope
  - `pkg/pagination`: page/per_page parsing with max-100 guard
  - `pkg/mailer`: SMTP mailer + `MockMailer` for tests
  - `pkg/database`: pgxpool connection with health check
  - `pkg/cache`: Redis connection with health check
  - Identity domain: `User` aggregate, bcrypt password hashing, role promotion, ban/unban
  - Identity infrastructure: `PostgresUserRepository` (sqlc), `RedisTokenStore`
  - Identity application: register, verify email, login, refresh, logout, forgot/reset password, update profile, change password
  - Identity HTTP: 11 endpoints under `/api/v2/auth/*` and `/api/v2/me/*`
  - Migrations: `uuid-ossp`/`pg_trgm` extensions, `users` table
  - Makefile with `run`, `test`, `migrate-up`, `migrate-down`, `seed`, `lint`, `tidy`
  - CI: GitHub Actions test + build + lint pipeline

[Unreleased]: https://github.com/iqbaleff214/kamus-banjar-api-2/compare/v0.1.0...HEAD
[0.1.0]: https://github.com/iqbaleff214/kamus-banjar-api-2/releases/tag/v0.1.0
