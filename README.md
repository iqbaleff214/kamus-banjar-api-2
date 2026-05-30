# Kamus Banjar API

> A community-driven REST API for the Banjar Hulu dialect dictionary — digitized, searchable, and open to contributions.

[![CI](https://github.com/iqbaleff214/kamus-banjar-api-2/actions/workflows/ci.yml/badge.svg)](https://github.com/iqbaleff214/kamus-banjar-api-2/actions/workflows/ci.yml)
[![codecov](https://codecov.io/gh/iqbaleff214/kamus-banjar-api-2/graph/badge.svg)](https://codecov.io/gh/iqbaleff214/kamus-banjar-api-2)
[![Go](https://img.shields.io/badge/Go-1.25-00ADD8?logo=go&logoColor=white)](https://go.dev)
[![License: GPL v3](https://img.shields.io/badge/License-GPLv3-blue.svg)](LICENSE)
[![Version](https://img.shields.io/badge/version-2.0.0-blue)](https://github.com/iqbaleff214/kamus-banjar-api-2)

---

## Overview

**Kamus Banjar API** is a RESTful backend platform for the Banjar language dictionary (*Dialek Hulu*), digitized from the reference book *Kamus Bahasa Banjar Dialek Hulu–Indonesia* (Balai Bahasa Banjarmasin, Departemen Pendidikan Nasional, 2008). It provides public read access to ~7,000 dictionary entries, a community contribution and review workflow, AI-assisted translation via OpenRouter, and moderation tooling for admins.

The project is built with [Domain-Driven Design (DDD)](https://martinfowler.com/bliki/DomainDrivenDesign.html) and [Test-Driven Development (TDD)](https://martinfowler.com/bliki/TestDrivenDevelopment.html) principles, and is intended as a long-term open-source infrastructure for preserving and enriching the Banjar language.

### Primary Data Source

> **Kamus Bahasa Banjar Dialek Hulu–Indonesia**, Edisi Pertama  
> Balai Bahasa Banjarmasin, Departemen Pendidikan Nasional, 2008  
> ISBN: 978-979-685-776-0  
> Authors: Musdalipah, Siti Akbari, Jandiah, Wandanie Rakhman, Muhammad Yamani, H. Dede Hidayatullah, Noor Hastiah

---

## Features

| Feature | Guest | User | Admin |
|---|:---:|:---:|:---:|
| Browse & search dictionary | ✓ | ✓ | ✓ |
| View word detail, definitions, examples | ✓ | ✓ | ✓ |
| AI translation (Banjar → Indonesian) | — | ✓ | ✓ |
| Bookmark words | — | ✓ | ✓ |
| Upvote / downvote definitions | — | ✓ | ✓ |
| Submit word/definition contributions | — | ✓ | ✓ |
| Comment on words and contributions | — | ✓ | ✓ |
| Add/edit/delete words directly | — | — | ✓ |
| Approve/reject contributions | — | — | ✓ |
| Manage users & moderation queue | — | — | ✓ |
| Trigger AI enrichment jobs | — | — | ✓ |

---

## Tech Stack

| Layer | Technology |
|---|---|
| Language | Go 1.25 |
| Framework | [Fiber v2](https://gofiber.io) |
| Database | PostgreSQL 15+ |
| Query Layer | [sqlc](https://sqlc.dev) |
| Cache / Rate Limiting | Redis 7 |
| Auth | JWT (HS256) + Redis refresh tokens |
| Email | SMTP (configurable via env) |
| AI | [OpenRouter](https://openrouter.ai) |
| Migrations | [golang-migrate](https://github.com/golang-migrate/migrate) |
| Testing | [testify](https://github.com/stretchr/testify) |
| Containerization | Docker + Docker Compose |
| Reverse Proxy | [Traefik v3](https://traefik.io) (production) |

---

## Architecture

The project follows **Domain-Driven Design** organized into five bounded contexts:

```
internal/
├── dictionary/     # Word entries, definitions, examples
├── community/      # Contributions, votes, bookmarks, comments
├── identity/       # User accounts, authentication, roles
├── moderation/     # Approval queues, flagging, audit logs
└── ai/             # OpenRouter integration, enrichment jobs
```

Each bounded context is structured as:

```
<context>/
├── domain/         # Aggregates, value objects, repository interfaces
├── application/    # Commands and queries (use cases)
├── infrastructure/ # PostgreSQL/Redis implementations
└── http/           # Fiber handlers and routes
```

---

## Getting Started

### Prerequisites

- [Docker](https://docs.docker.com/get-docker/) and [Docker Compose](https://docs.docker.com/compose/)
- [Go 1.25+](https://go.dev/dl/) (for local development without Docker)
- [golang-migrate CLI](https://github.com/golang-migrate/migrate/tree/master/cmd/migrate) (for running migrations manually)

### 1. Clone the repository

```bash
git clone https://github.com/iqbaleff214/kamus-banjar-api-2.git
cd kamus-banjar-api-2
```

### 2. Configure environment

```bash
cp .env.example .env
```

Edit `.env` and fill in the required values (see [Configuration](#configuration)).

### 3. Start with Docker (recommended)

```bash
docker compose up --build
```

This starts:
- **api** — Go API with hot-reload via [Air](https://github.com/air-verse/air) at `http://localhost:3000`
- **db** — PostgreSQL 15 at `localhost:5432`
- **redis** — Redis 7 at `localhost:6379`

### 4. Run migrations

```bash
make migrate-up
```

### 5. Seed the dictionary data

```bash
make seed
```

This imports ~7,000 Banjar entries from `scripts/seed/seed_data.json` into PostgreSQL and creates a default admin user. The seeder is idempotent — safe to run multiple times.

Default admin credentials (override via env vars):

| Env var | Default |
|---|---|
| `ADMIN_EMAIL` | `admin@kamus-banjar.id` |
| `ADMIN_PASSWORD` | `Admin1234!` |
| `ADMIN_NAME` | `Admin` |

```bash
ADMIN_EMAIL=me@example.com ADMIN_PASSWORD=StrongPass1! make seed
```

### Health check

```bash
curl http://localhost:3000/health
# {"status":"ok"}
```

---

## API Reference

Base URL: `http://localhost:3000/api/v2`

All responses follow a consistent envelope:

```json
{
  "success": true,
  "data": { },
  "meta": { "page": 1, "per_page": 20, "total": 500, "total_pages": 25 }
}
```

Errors:

```json
{
  "success": false,
  "error": { "code": "NOT_FOUND", "message": "word not found" }
}
```

### Dictionary (Public)

| Method | Endpoint | Description |
|---|---|---|
| `GET` | `/words` | List words (paginated) |
| `GET` | `/words?q=:query` | Full-text search |
| `GET` | `/words?word_class=n` | Filter by word class |
| `GET` | `/words/:id` | Get word detail |
| `GET` | `/words/:id/definitions` | List definitions (sorted by net score) |
| `GET` | `/words/:id/examples` | List usage examples |
| `GET` | `/words/:id/related` | Get related words |

**Query parameters for list/search:**

| Parameter | Values | Default |
|---|---|---|
| `page` | integer | `1` |
| `per_page` | 1–100 | `20` |
| `word_class` | `n`, `v`, `a`, `adv`, `p`, `pb`, `ki` | — |
| `is_root` | `true`, `false` | — |
| `sort` | `alphabetical`, `most_voted`, `recently_added` | `alphabetical` |

### Identity

| Method | Endpoint | Description |
|---|---|---|
| `POST` | `/auth/register` | Register new account |
| `POST` | `/auth/login` | Login, receive JWT |
| `POST` | `/auth/logout` | Invalidate refresh token |
| `POST` | `/auth/refresh` | Rotate access token |
| `POST` | `/auth/verify-email` | Verify email with token |
| `POST` | `/auth/forgot-password` | Send password reset email |
| `POST` | `/auth/reset-password` | Reset password with token |
| `GET` | `/auth/me` | Get current user profile |
| `PUT` | `/auth/me` | Update profile |
| `PUT` | `/auth/me/password` | Change password |

### AI Translation (Auth Required)

```http
POST /api/v2/ai/translate
Authorization: Bearer <token>
Content-Type: application/json

{
  "text": "inya kada kawa tulak ka pasar",
  "context": "informal conversation"
}
```

```json
{
  "success": true,
  "data": {
    "original": "inya kada kawa tulak ka pasar",
    "translation": "dia tidak bisa pergi ke pasar",
    "dialect": "hulu",
    "model": "mistralai/mistral-7b-instruct:free",
    "notes": "kada=tidak, kawa=bisa, tulak=pergi"
  }
}
```

Rate limit: 30 requests/hour per authenticated user.

### Admin Word Management (Admin Only)

| Method | Endpoint | Description |
|---|---|---|
| `POST` | `/admin/words` | Create word directly |
| `PATCH` | `/admin/words/:id` | Update word |
| `DELETE` | `/admin/words/:id` | Soft-delete word |

For the full OpenAPI specification, see [`openapi.yaml`](openapi.yaml).

---

## Configuration

All configuration is via environment variables. Copy `.env.example` to `.env`:

```env
# Application
APP_PORT=3000
APP_ENV=development
APP_DOMAIN=api.example.com       # Production domain (used by Traefik)

# PostgreSQL
DB_HOST=db
DB_PORT=5432
DB_NAME=kamus_banjar
DB_USER=kamus
DB_PASS=secret

# Redis
REDIS_ADDR=redis:6379
REDIS_PASS=

# JWT — generate a strong random secret for production
JWT_SECRET=change-me-in-production-min-32-chars

# SMTP (leave empty to use mock mailer in development)
SMTP_HOST=
SMTP_PORT=587
SMTP_USER=
SMTP_PASS=
SMTP_FROM=noreply@example.com

# OpenRouter
OPENROUTER_API_KEY=
OPENROUTER_MODEL=mistralai/mistral-7b-instruct:free
OPENROUTER_BASE_URL=https://openrouter.ai/api/v1
```

---

## Development

### Makefile targets

**Local dev (Docker Compose)**

```bash
make run           # docker compose up --build (foreground)
make up            # docker compose up -d --build (detached)
make down          # docker compose down
make logs          # Follow compose logs
make ps            # Show running services
```

**Migrations**

```bash
make migrate-up    # Run all pending migrations (via migrate container)
make migrate-down  # Roll back last migration
```

**Seeder** (requires running DB, uses host Go)

```bash
make seed          # Seed dictionary data + default admin user
```

**Tests / quality**

```bash
make test          # Run all tests (race detector enabled)
make coverage      # Coverage report — enforces ≥80% on domain + application layers
make lint          # Run golangci-lint
make tidy          # go mod tidy
```

**Production**

```bash
make prod-up       # docker compose -f docker-compose.prod.yml up -d --build
make prod-down     # docker compose -f docker-compose.prod.yml down
make prod-logs     # Follow prod logs
make prod-ps       # Show prod services
make prod-seed     # Run seeder container in production
```

### Running tests

```bash
go test ./... -v -race -count=1
```

Tests are organized by layer:

| Layer | Location | Type |
|---|---|---|
| Domain | `internal/*/domain/` | Pure unit — no I/O |
| Application | `internal/*/application/` | Unit with in-memory fakes |
| HTTP handlers | `internal/*/http/` | Integration via `app.Test()` |
| Repository | `internal/*/infrastructure/postgres/` | Integration against real DB |

**Integration tests** (repository layer) require a running PostgreSQL instance. Set `TEST_DATABASE_URL` before running:

```bash
TEST_DATABASE_URL=postgres://user:pass@localhost:5432/testdb go test ./...
```

Tests skip automatically when `TEST_DATABASE_URL` is not set — unit and handler tests always run.

### Code generation

After editing SQL queries in `internal/*/infrastructure/postgres/queries/`:

```bash
sqlc generate
```

### Project structure

```
kamus-banjar-api-2/
├── cmd/api/                  # Main entrypoint
├── internal/
│   ├── dictionary/           # Dictionary bounded context
│   ├── community/            # Community bounded context
│   ├── identity/             # Identity bounded context
│   ├── moderation/           # Moderation bounded context
│   └── ai/                   # AI bounded context
├── pkg/
│   ├── auth/                 # JWT helpers
│   ├── cache/                # Redis connection
│   ├── config/               # Environment config
│   ├── database/             # PostgreSQL connection pool
│   ├── httperr/              # Error response helpers
│   ├── mailer/               # SMTP + mock mailer
│   ├── middleware/           # Request ID, logging middleware
│   ├── pagination/           # Pagination utilities
│   └── ratelimit/            # Redis-backed rate limiter
├── migrations/               # golang-migrate SQL files
├── scripts/seed/             # Dictionary data extraction + import
│   ├── extract_dictionary.py # PDF → seed_data.json (requires pdfminer.six)
│   ├── seed_data.json        # Extracted Banjar dictionary entries
│   └── main.go               # Go seeder (seed_data.json → PostgreSQL)
├── testutil/                 # Shared test helpers (DB setup, isolated schemas)
├── traefik/                  # Server-level Traefik reverse proxy (deploy once per server)
│   ├── docker-compose.yml    # Traefik stack
│   └── .env.example          # Traefik env vars
├── docs/                     # Reference PDF and specs
├── openapi.yaml              # OpenAPI 3.0 specification
├── DICTIONARY_SPEC.md        # Word class definitions, entry format, known caveats
├── PRD.md                    # Product Requirements Document
├── Dockerfile                # Multi-stage (dev / prod)
├── docker-compose.yml        # Development stack
└── docker-compose.prod.yml   # Production stack (Traefik-integrated)
```

---

## Production Deployment

The production setup uses **Traefik v3** as a server-level reverse proxy handling TLS termination (Let's Encrypt) and HTTP→HTTPS redirects. Traefik is deployed once per server; each application joins the shared `traefik_public` network.

### 1. Deploy Traefik (once per server)

```bash
cd traefik/
cp .env.example .env
# Fill in ACME_EMAIL, TRAEFIK_DOMAIN, and TRAEFIK_DASHBOARD_AUTH
docker compose up -d
```

Generate `TRAEFIK_DASHBOARD_AUTH`:

```bash
docker run --rm httpd:alpine htpasswd -nB admin
# Escape every $ as $$ in the .env value
```

### 2. Deploy the API

```bash
cp .env.example .env
# Set APP_DOMAIN, DB_*, REDIS_*, JWT_SECRET, OPENROUTER_API_KEY, etc.
docker compose -f docker-compose.prod.yml up -d
```

Migrations run automatically on startup via the `migrate` service.

### 3. Seed the dictionary data

```bash
docker compose -f docker-compose.prod.yml --profile seed run --rm seed
```

This imports ~7,000 dictionary entries and creates a default admin user. Idempotent — safe to re-run.

Override admin credentials:

```bash
ADMIN_EMAIL=you@example.com ADMIN_PASSWORD=StrongPass1! \
  docker compose -f docker-compose.prod.yml --profile seed run --rm seed
```

The API is available at `https://<APP_DOMAIN>`. Traefik automatically provisions and renews a Let's Encrypt certificate.

**Network topology:**

```
Internet → Traefik (:443) → [traefik_public] → api
                                                 └── [internal] → db, redis
```

`db` and `redis` are on an isolated internal network — not reachable from outside the host.

---

## Roadmap

| Phase | Status | Description |
|---|---|---|
| Phase 1 — Scaffold & Identity | ✅ Done | Project setup, auth, email verification, JWT |
| Phase 2 — Dictionary | ✅ Done | Schema, seed import, CRUD, full-text search, public API |
| Phase 3 — AI Translation | ✅ Done | `POST /ai/translate`, OpenRouter, rate limiting |
| Phase 4 — Community | ✅ Done | Contributions, votes, bookmarks, comments |
| Phase 5 — Moderation | ✅ Done | Approval queue, user ban, audit log |
| Phase 6 — AI Enrichment | ✅ Done | Admin async enrichment jobs, review workflow |
| Phase 7 — Hardening | ✅ Done | Swagger UI, ≥80% coverage enforcement, lint gates |

---

## Dictionary Data

The seed data covers the **Banjar Hulu** dialect only.

| Property | Value |
|---|---|
| Direction | Banjar Hulu → Indonesian |
| Estimated root entries | ~2,200 |
| Estimated total entries | ~7,000 |
| Word classes | `n` (noun), `v` (verb), `a` (adjective), `adv` (adverb), `p` (particle), `pb` (proverb), `ki` (figurative) |

The PDF extraction has known OCR artefacts. Human review of seed data is recommended before deploying to production. See [DICTIONARY_SPEC.md](DICTIONARY_SPEC.md) for full entry format specification and known limitations.

---

## Contributing

Contributions are welcome. Please note:

- This project is licensed under the [GNU General Public License v3.0 (GPL-3.0)](LICENSE). Contributions you submit are accepted under the same license.
- Any derivative work or redistributed modification must also be licensed under GPL-3.0.
- Proper credit to the original author must be retained in derivative works: **M. Iqbal Effendi**.
- If you distribute modified versions of this project, you must make the source code available.
- Open an issue before submitting a large PR to discuss scope.
- Follow existing code conventions (DDD layers, TDD, no ORM).

---

## Author

**M. Iqbal Effendi**  
[github.com/iqbaleff214](https://github.com/iqbaleff214)  
iqbaleff214@gmail.com

---

## License

This project is licensed under the **GNU General Public License v3.0 (GPL-3.0)**.

You are free to:
- **Use** — run and use this project for any purpose
- **Study** — inspect and learn from the source code
- **Modify** — change and improve the project
- **Distribute** — share copies of the original or modified project

Under the following conditions:
- **Copyleft** — Any derivative work or modified version of this project must also be licensed under GPL-3.0.
- **Source Disclosure** — Source code of modified versions must be made available when distributed.
- **Attribution** — Proper credit to the original author (**M. Iqbal Effendi**) must be retained.
- **License Notice** — A copy of the GPL-3.0 license must be included with distributions.

This project is distributed in the hope that it will be useful, but WITHOUT ANY WARRANTY; without even the implied warranty of MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.

See the full license text in [LICENSE](LICENSE).

---

*In memory of the Banjar language and its speakers — Haram Manyarah Waja Sampai Kaputing.*
