# Product Requirements Document
# Kamus Banjar API 2

**Version:** 2.0.0  
**Date:** 2026-05-16  
**Status:** Draft  

---

## 1. Overview

### 1.1 Product Summary

Kamus Banjar API 2 is a RESTful API platform for the Banjar language dictionary (Dialek Hulu), digitized from the reference book *Kamus Bahasa Banjar Dialek Hulu-Indonesia, Edisi Pertama* (Balai Bahasa Banjarmasin, Departemen Pendidikan Nasional, 2008; ISBN 978-979-685-776-0). It is a community-driven dictionary platform that supports public read access, authenticated user contributions, AI-assisted content enrichment via OpenRouter, and admin moderation tooling.

### 1.2 Goals

- Digitize and expose Banjar Hulu dialect dictionary entries via a public REST API
- Enable community contributions with a structured approval workflow
- Integrate AI (via OpenRouter) for definition enrichment, translation suggestions, and content quality assistance
- Build a moderation layer to ensure data integrity and community health
- Serve as an open-source, long-term reference infrastructure for the Banjar language

### 1.3 Primary Data Source

The canonical data source is:

> **Kamus Bahasa Banjar Dialek Hulu-Indonesia**, Edisi Pertama  
> Balai Bahasa Banjarmasin, Departemen Pendidikan Nasional, 2008  
> ISBN: 978-979-685-776-0  
> Authors: Musdalipah, Siti Akbari, Jandiah, Wandanie Rakhman, Muhammad Yamani, H. Dede Hidayatullah, Noor Hastiah

**Dictionary characteristics (derived from source):**

| Property | Value |
|---|---|
| Direction | Banjar Hulu → Indonesian |
| Dialect | Banjar Dialek Hulu (BBDH) |
| Letters covered | A B C D G H I J K L M N P R S T U W Y |
| Letters absent from dialect | E F O Q V Z (mapped: E→I/A, F/V→P, O→U, Q→K, Z→S/J) |
| Estimated root entries | ~2,200 |
| Estimated total entries (incl. derived forms) | ~7,000 |
| Word classes used | `n`, `v`, `a`, `adv`, `p`, `pb`, `ki` |
| Entry types | Root words + derived forms (ba-, ma-, ka-, ta-, sa-, pa- prefixed) |
| Example sentences | Banjar sentence + Indonesian translation |

See [DICTIONARY_SPEC.md](DICTIONARY_SPEC.md) for detailed word class definitions, entry format spec, and seeder data notes.

### 1.4 Non-Goals

- Mobile or web frontend (API only)
- Real-time features (WebSocket, live chat)
- Multi-language dictionary beyond Banjar ↔ Indonesian (Melayu) scope in v2

---

## 2. Architecture Principles

### 2.1 Domain-Driven Design (DDD)

The system is organized around bounded contexts with explicit domain models, aggregates, value objects, domain events, and repositories.

**Core Bounded Contexts:**

| Context | Responsibility |
|---|---|
| **Dictionary** | Word entries, definitions, examples, etymologies |
| **Community** | Contributions, reviews, upvotes, bookmarks |
| **Identity** | User accounts, authentication, roles |
| **Moderation** | Approval workflows, flagging, admin actions |
| **AI** | OpenRouter integration, suggestion generation |

### 2.2 Test-Driven Development (TDD)

All domain logic, application services, and API handlers must be developed test-first:

- **Unit tests** — domain models, value objects, business rules
- **Integration tests** — repositories, database interactions, external API calls (mocked)
- **Contract/E2E tests** — API endpoints, full request-response cycle
- Minimum coverage threshold: **80%** on domain and application layers

---

## 3. User Roles & Permissions

### 3.1 Role Definitions

| Role | Description |
|---|---|
| **Guest** | Unauthenticated. Read-only public access. |
| **User** | Authenticated. Can contribute, bookmark, vote, and comment. |
| **Admin** | Full access. Direct dictionary contribution and moderation. |

### 3.2 Permission Matrix

| Feature | Guest | User | Admin |
|---|:---:|:---:|:---:|
| Browse/search words | ✓ | ✓ | ✓ |
| View word detail | ✓ | ✓ | ✓ |
| View example sentences | ✓ | ✓ | ✓ |
| View AI-tagged definitions/examples | ✓ | ✓ | ✓ |
| **AI translate (Banjar → Indonesian)** | ✓ | ✓ | ✓ |
| Register / login | ✓ | — | — |
| Bookmark words | ✗ | ✓ | ✓ |
| Upvote / downvote word | ✗ | ✓ | ✓ |
| Submit word contribution | ✗ | ✓ | ✓ |
| Submit definition contribution | ✗ | ✓ | ✓ |
| Write comments/reviews | ✗ | ✓ | ✓ |
| Edit own contributions (pending) | ✗ | ✓ | ✓ |
| Flag content | ✗ | ✓ | ✓ |
| Add word directly (no approval) | ✗ | ✗ | ✓ |
| Edit any word directly | ✗ | ✗ | ✓ |
| Approve / reject contributions | ✗ | ✗ | ✓ |
| Delete words / definitions | ✗ | ✗ | ✓ |
| Manage users | ✗ | ✗ | ✓ |
| View moderation queue | ✗ | ✗ | ✓ |
| Trigger AI enrichment / quality check | ✗ | ✗ | ✓ |
| Review / approve AI enrichment output | ✗ | ✗ | ✓ |

---

## 4. Domain Model

### 4.1 Dictionary Context

**Aggregate: `Word`**

| Field | Type | Description |
|---|---|---|
| `id` | UUID | Primary identifier |
| `banjar` | string | Banjar word without syllable markers (e.g. `abah`) |
| `banjar_syllabified` | string? | Syllabified form from source (e.g. `a.bah`) |
| `dialect` | enum | `hulu` — only dialect in scope for v2 |
| `word_class` | enum | See word class table below |
| `homonym_number` | int | 1 for primary; 2, 3, ... for homonyms (e.g. `¹amar`, `²amar`) |
| `is_root` | bool | True = root word; false = derived form (ba-, ma-, ka-, etc.) |
| `root_word_id` | UUID? | Parent root word ID when `is_root = false` |
| `definitions` | `Definition[]` | One or more definitions |
| `examples` | `Example[]` | Usage example sentences |
| `related_words` | UUID[] | References to related `Word` IDs |
| `status` | enum | `active` \| `deprecated` |
| `source` | enum | `seeded` \| `contributed` \| `ai_generated` |
| `source_reference` | string? | Citation (for seeded entries: book title/edition) |
| `created_by` | UUID? | Admin user ID; null for system-seeded |
| `created_at` | datetime | — |
| `updated_at` | datetime | — |
| `deleted_at` | datetime? | Soft delete |

**Word Classes (from source dictionary, section 2.1)**

| Abbreviation | Full Name | Description |
|---|---|---|
| `n` | nomina | Noun |
| `v` | verba | Verb |
| `a` | adjektiva | Adjective |
| `adv` | adverbia | Adverb |
| `p` | partikel | Particle / interjection |
| `pb` | pribahasa | Proverb |
| `ki` | kiasan | Figurative / idiomatic |

**Value Object: `Definition`**

| Field | Type | Description |
|---|---|---|
| `id` | UUID | — |
| `meaning` | string | Indonesian meaning/translation |
| `sort_order` | int | Order when word has multiple definitions (1, 2, ...) |
| `source` | enum | `seeded` \| `contributed` \| `ai_generated` |
| `upvotes` | int | Community upvote count |
| `downvotes` | int | Community downvote count |

**Value Object: `Example`**

| Field | Type | Description |
|---|---|---|
| `id` | UUID | — |
| `banjar_sentence` | string | Banjar example sentence (`--` in source = word itself) |
| `indonesian_translation` | string | Indonesian translation |
| `source` | enum | `seeded` \| `contributed` \| `ai_generated` |

### 4.2 Community Context

**Aggregate: `Contribution`**

| Field | Type | Description |
|---|---|---|
| `id` | UUID | — |
| `type` | enum | `new_word` \| `new_definition` \| `new_example` \| `edit_word` |
| `contributor_id` | UUID | User who submitted |
| `target_word_id` | UUID? | Null for `new_word` type |
| `payload` | JSON | Proposed content (varies by type) |
| `status` | enum | `pending` \| `approved` \| `rejected` \| `withdrawn` |
| `reviewer_id` | UUID? | Admin who acted on it |
| `reviewer_note` | string? | Admin rejection/approval note |
| `submitted_at` | datetime | — |
| `reviewed_at` | datetime? | — |

**Aggregate: `Vote`**

| Field | Type | Description |
|---|---|---|
| `id` | UUID | — |
| `user_id` | UUID | — |
| `target_type` | enum | `word` \| `definition` \| `example` \| `contribution` |
| `target_id` | UUID | — |
| `value` | enum | `up` \| `down` |
| `created_at` | datetime | — |

**Aggregate: `Bookmark`**

| Field | Type | Description |
|---|---|---|
| `id` | UUID | — |
| `user_id` | UUID | — |
| `word_id` | UUID | — |
| `created_at` | datetime | — |

**Aggregate: `Comment`**

| Field | Type | Description |
|---|---|---|
| `id` | UUID | — |
| `user_id` | UUID | — |
| `target_type` | enum | `word` \| `contribution` |
| `target_id` | UUID | — |
| `body` | string | Comment text (max 1000 chars) |
| `is_flagged` | bool | Flagged for moderation |
| `created_at` | datetime | — |
| `updated_at` | datetime | — |

### 4.3 Identity Context

**Aggregate: `User`**

| Field | Type | Description |
|---|---|---|
| `id` | UUID | — |
| `name` | string | Display name |
| `email` | string | Unique |
| `password_hash` | string | Bcrypt |
| `role` | enum | `user` \| `admin` |
| `is_active` | bool | Account active/banned |
| `email_verified_at` | datetime? | — |
| `created_at` | datetime | — |

### 4.4 AI Context

**Aggregate: `AIRequest`** — only for admin-triggered, async enrichment jobs (translation is stateless, not stored)

| Field | Type | Description |
|---|---|---|
| `id` | UUID | — |
| `type` | enum | `enrich_definition` \| `suggest_example` \| `suggest_related` \| `quality_check` |
| `target_word_id` | UUID | Word being enriched |
| `target_contribution_id` | UUID? | Contribution being quality-checked |
| `requested_by` | UUID | Admin user ID |
| `model` | string | OpenRouter model ID used (e.g. `mistralai/mistral-7b-instruct:free`) |
| `prompt` | string | Prompt sent to OpenRouter |
| `response` | JSON? | Raw OpenRouter response |
| `parsed_output` | JSON? | Structured extraction from response |
| `status` | enum | `pending` \| `completed` \| `failed` |
| `review_status` | enum | `unreviewed` \| `approved` \| `rejected` |
| `reviewed_by` | UUID? | Admin who reviewed |
| `reviewed_at` | datetime? | — |
| `created_at` | datetime | — |

---

## 5. API Specification

### 5.1 Base

```
Base URL: /api/v2
Content-Type: application/json
Authentication: Bearer token (JWT)
```

### 5.2 Dictionary Endpoints

```
GET    /words                        # List/search words (paginated)
GET    /words/:id                    # Get word detail
GET    /words/:id/definitions        # List definitions for a word
GET    /words/:id/examples           # List examples for a word
GET    /words/:id/related            # Get related words
GET    /words/search?q=:query        # Full-text search
```

### 5.3 Contribution Endpoints

```
POST   /contributions                # Submit a contribution (user)
GET    /contributions                # List own contributions (user) / all (admin)
GET    /contributions/:id            # Get contribution detail
PATCH  /contributions/:id/withdraw   # Withdraw pending contribution (own)
PATCH  /contributions/:id/approve    # Approve contribution (admin)
PATCH  /contributions/:id/reject     # Reject contribution (admin)
```

### 5.4 Community Endpoints

```
POST   /words/:id/votes              # Cast vote on word (user)
DELETE /words/:id/votes              # Remove vote (user)
POST   /definitions/:id/votes        # Cast vote on definition (user)
DELETE /definitions/:id/votes        # Remove vote (user)

GET    /bookmarks                    # List own bookmarks (user)
POST   /bookmarks                    # Add bookmark (user)
DELETE /bookmarks/:word_id           # Remove bookmark (user)

GET    /words/:id/comments           # List comments on word
POST   /words/:id/comments           # Post comment (user)
PATCH  /comments/:id                 # Edit own comment (user)
DELETE /comments/:id                 # Delete comment (own or admin)
POST   /comments/:id/flag            # Flag comment (user)
```

### 5.5 AI Endpoints (Public)

```
POST   /ai/translate                 # Translate Banjar text → Indonesian (guest + user)
```

### 5.6 Identity Endpoints

```
POST   /auth/register                # Register new user
POST   /auth/login                   # Login, return JWT
POST   /auth/logout                  # Invalidate token
POST   /auth/refresh                 # Refresh token
GET    /auth/me                      # Get current user profile
PATCH  /auth/me                      # Update profile
PATCH  /auth/me/password             # Change password
POST   /auth/verify-email            # Verify email with token
POST   /auth/forgot-password         # Send password reset email
POST   /auth/reset-password          # Reset password with token
```

### 5.7 Admin Endpoints

```
GET    /admin/words                          # List all words including inactive
POST   /admin/words                          # Create word directly
PATCH  /admin/words/:id                      # Update word directly
DELETE /admin/words/:id                      # Soft-delete word

GET    /admin/users                          # List users
GET    /admin/users/:id                      # Get user detail
PATCH  /admin/users/:id/ban                  # Ban user
PATCH  /admin/users/:id/unban                # Unban user
PATCH  /admin/users/:id/role                 # Change user role

GET    /admin/moderation/queue               # Pending contributions
GET    /admin/moderation/flags               # Flagged comments
GET    /admin/moderation/stats               # Moderation statistics

POST   /admin/ai/enrich/:word_id             # Trigger definition enrichment job
POST   /admin/ai/example/:word_id            # Trigger example suggestion job
POST   /admin/ai/related/:word_id            # Trigger related word suggestion job
POST   /admin/ai/check/:contribution_id      # Run quality check on a contribution
GET    /admin/ai/requests                    # List AI request history (paginated)
GET    /admin/ai/requests/:id                # Get AI request + parsed output
PATCH  /admin/ai/requests/:id/approve        # Approve and merge AI output
PATCH  /admin/ai/requests/:id/reject         # Reject AI output
```

### 5.7 Standard Response Format

**Success:**
```json
{
  "success": true,
  "data": { ... },
  "meta": {
    "page": 1,
    "per_page": 20,
    "total": 500
  }
}
```

**Error:**
```json
{
  "success": false,
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "Human-readable message",
    "details": { ... }
  }
}
```

### 5.8 Error Codes

| Code | HTTP | Description |
|---|---|---|
| `VALIDATION_ERROR` | 422 | Request body validation failure |
| `UNAUTHORIZED` | 401 | Missing or invalid token |
| `FORBIDDEN` | 403 | Insufficient role |
| `NOT_FOUND` | 404 | Resource does not exist |
| `CONFLICT` | 409 | Duplicate resource (e.g. existing bookmark, vote) |
| `RATE_LIMITED` | 429 | Too many requests |
| `AI_UNAVAILABLE` | 503 | OpenRouter unreachable |
| `INTERNAL_ERROR` | 500 | Unexpected server error |

---

## 6. Key Features

### 6.1 Dictionary Search

- Full-text search on `banjar` and Indonesian `meaning` fields
- Filter by: `word_class` (`n`, `v`, `a`, `adv`, `p`, `pb`, `ki`), `is_root`, `source`
- Sort by: `alphabetical` (default), `most_voted`, `recently_added`
- Pagination: default 20, max 100 per page

### 6.2 Contribution Workflow

```
User submits contribution
        ↓
Status: pending
        ↓
Admin reviews in moderation queue
        ↓
    approved → content merged into dictionary
    rejected → contributor notified with reason
        ↓
User may withdraw while pending
```

- Contributors receive notification (DB flag; email in future scope) on status change
- Rejected contributions include admin note
- Admin can approve with optional edits before merge

### 6.3 Voting System

- One vote per user per target (word, definition, example)
- Vote can be changed (up → down or vice versa)
- Vote removed by sending DELETE
- Net score displayed: `upvotes - downvotes`
- Definitions sorted by net score by default

### 6.4 Bookmark System

- Users maintain personal word list
- No public visibility
- Paginated list endpoint

### 6.5 Community Comments

- Threaded comments on words and contributions (flat, no nested replies in v2)
- Edit window: own comments only
- Flagging triggers moderation queue entry
- Admin can delete any comment

### 6.6 AI Integration (OpenRouter)

**Model configuration:** OpenRouter model ID is set via `OPENROUTER_MODEL` environment variable. Defaults to a free-tier model (e.g. `mistralai/mistral-7b-instruct:free`). Can be swapped to any OpenRouter-supported model without code changes.

#### AI Feature Tiers

| Feature | Who can use | Approval required |
|---|---|---|
| Text translation (Banjar → Indonesian) | Guest, User, Admin | No — immediate response |
| Word definition enrichment | Admin only | Yes — admin reviews before publish |
| Example sentence suggestion | Admin only | Yes — admin reviews before publish |
| Contribution quality check | Admin only | No — advisory output only |
| Related word suggestion | Admin only | Yes — admin reviews before publish |

#### 6.6.1 Translation (Public Feature)

**The primary public-facing AI feature.** Any caller (including unauthenticated guests) can submit free-form Banjar Hulu text and receive an Indonesian translation.

```
POST /ai/translate
```

Request:
```json
{
  "text": "inya kada kawa tulak ka pasar",
  "context": "informal conversation"   // optional
}
```

Response:
```json
{
  "success": true,
  "data": {
    "original": "inya kada kawa tulak ka pasar",
    "translation": "dia tidak bisa pergi ke pasar",
    "dialect": "hulu",
    "model": "mistralai/mistral-7b-instruct:free",
    "confidence": "high",
    "notes": "Uses BBDH vocabulary: kada=tidak, kawa=bisa, tulak=pergi"
  }
}
```

- Stateless — results are NOT stored (no `AIRequest` record created)
- Rate-limited: 10 req/hour per IP (guest), 30 req/hour per user
- Context window: max 1000 characters of input text
- Prompt instructs model to use the Banjar Hulu dialect specifically and output only the translation + brief lexical notes

#### 6.6.2 Dictionary Enrichment (Admin-Only, Async)

Admin triggers enrichment on a specific word. Output is stored as a pending `AIRequest` and surfaced in the admin panel for review before becoming canonical.

| `type` | Input | Output | Approval needed |
|---|---|---|---|
| `enrich_definition` | Word + existing definitions | Suggested additional Indonesian definitions | Yes |
| `suggest_example` | Word + definition | Suggested Banjar sentence + Indonesian translation | Yes |
| `suggest_related` | Word + its definitions | List of suggested related Banjar words | Yes |
| `quality_check` | Contribution payload | Structured report: accuracy score, flags, notes | No (advisory) |

**Endpoints:**
```
POST   /admin/ai/enrich/:word_id            # Trigger enrichment job
GET    /admin/ai/requests                   # List all AI requests (paginated)
GET    /admin/ai/requests/:id               # Get AI request + response
PATCH  /admin/ai/requests/:id/approve       # Approve and merge AI output
PATCH  /admin/ai/requests/:id/reject        # Reject AI output
POST   /admin/ai/contributions/:id/check    # Run quality_check on a contribution
```

#### 6.6.3 AI-Generated Content Visibility

Per product decision: **AI-generated definitions and examples are visible to all users (including guests) without admin approval**, but are clearly labeled `source: ai_generated` in the API response. Admin approval promotes them to `source: contributed` or `source: seeded` status.

Rationale: increases perceived content richness immediately while the admin works through the review queue.

### 6.7 Moderation Tools

- Pending contribution queue with filter/sort
- Flagged comment queue
- Bulk approve/reject on contributions
- User ban with reason
- Moderation audit log (who acted, when, on what)

---

## 7. Technical Requirements

### 7.1 Stack

| Layer | Technology |
|---|---|
| Language | Go |
| Framework | Fiber |
| DB Layer | sqlc |
| Database | PostgreSQL 15+ |
| Cache / Rate limiter | Redis |
| Auth | JWT — access token (15 min) + refresh token stored in Redis (7 days) |
| Email | SMTP (configurable provider via env) |
| AI | OpenRouter HTTP API |
| Testing | testify, gomock |
| Migration | golang-migrate |
| Config | Environment variables (.env) |

### 7.2 Database

- PostgreSQL 15+
- UUID primary keys
- Soft delete via `deleted_at` on mutable aggregates
- Indexes on: `banjar`, `dialect`, `word_class`, `status`, `is_root`, `root_word_id`, `created_at`
- Full-text search index on `banjar`, `meaning` (Indonesian translation field)
- Unique constraint: `(banjar, dialect, homonym_number, is_root, root_word_id)`

### 7.3 Authentication

- JWT access token: 15-minute TTL, signed HS256
- Refresh token: 7-day TTL, stored as opaque token in Redis (key: `refresh:<token_hash>`, value: `user_id`)
- Logout invalidates refresh token immediately (Redis DEL)
- Email verification required before contribution privileges are granted
- Verification and password-reset tokens sent via SMTP; configurable via `SMTP_HOST`, `SMTP_PORT`, `SMTP_USER`, `SMTP_PASS`, `SMTP_FROM`
- Password: bcrypt, min cost factor 12

### 7.4 Rate Limiting

Redis-backed sliding window rate limiter (key: `ratelimit:<endpoint_group>:<identifier>`).

| Endpoint Group | Limit | Key |
|---|---|---|
| `GET /words*` (guest) | 60 req/min | per IP |
| `GET /words*` (auth) | 120 req/min | per user ID |
| `POST /contributions` | 10 req/hour | per user ID |
| `POST /auth/login` | 5 req/min | per IP |
| `POST /ai/translate` (user) | 30 req/hour | per user ID |
| `POST /ai/translate` (guest) | 10 req/hour | per IP |
| `POST /admin/ai/*` | 50 req/hour | per admin ID |

### 7.5 Seeding

- Initial dictionary data extracted from `.references/kamus-bahasa-banjar-dialek-hulu.pdf`
- Extraction script: `scripts/seed/extract_dictionary.py` (requires `pdfminer.six`)
- Produces `scripts/seed/seed_data.json` with ~2,200 root entries and ~5,000 total entries
- Seeder inserts entries with `source: seeded`, `created_by: null` (system), `dialect: hulu`
- Seeding is idempotent — upsert on `(banjar, dialect, homonym_number, root_word_id)`
- PDF extraction has known OCR artefacts; human review of seed data recommended before first deploy
- See [DICTIONARY_SPEC.md](DICTIONARY_SPEC.md) for entry format, word class definitions, and known extraction limitations

### 7.6 Environment Variables

```
# App
APP_PORT=8080
APP_ENV=production

# PostgreSQL
DB_HOST=
DB_PORT=5432
DB_NAME=kamus_banjar
DB_USER=
DB_PASS=

# Redis
REDIS_ADDR=localhost:6379
REDIS_PASS=

# JWT
JWT_SECRET=

# SMTP
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

## 8. Non-Functional Requirements

| Requirement | Target |
|---|---|
| API response time (p95) | < 200ms for read, < 500ms for write |
| Uptime | 99.5% |
| Test coverage (domain + app layer) | ≥ 80% |
| Max payload size | 1MB |
| Pagination max | 100 items |
| Word definition max length | 2000 chars |
| Comment max length | 1000 chars |

---

## 9. DDD Structure

```
kamus-banjar-api-2/
├── cmd/
│   └── api/            # Entry point
├── internal/
│   ├── dictionary/     # Bounded context: Dictionary
│   │   ├── domain/
│   │   │   ├── word.go
│   │   │   ├── definition.go
│   │   │   ├── example.go
│   │   │   └── repository.go   # Interface
│   │   ├── application/
│   │   │   ├── commands/
│   │   │   └── queries/
│   │   └── infrastructure/
│   │       └── postgres/
│   ├── community/      # Bounded context: Community
│   │   ├── domain/
│   │   ├── application/
│   │   └── infrastructure/
│   ├── identity/       # Bounded context: Identity
│   │   ├── domain/
│   │   ├── application/
│   │   └── infrastructure/
│   ├── moderation/     # Bounded context: Moderation
│   │   ├── domain/
│   │   ├── application/
│   │   └── infrastructure/
│   └── ai/             # Bounded context: AI
│       ├── domain/
│       ├── application/
│       └── infrastructure/
├── pkg/
│   ├── auth/           # JWT helpers
│   ├── httperr/        # Error types and response helpers
│   ├── pagination/
│   ├── ratelimit/      # Redis sliding window rate limiter
│   └── validator/
├── migrations/
├── scripts/
│   └── seed/           # PDF extraction + seed import
│       ├── extract_dictionary.py
│       ├── seed_data.json
│       └── main.go     # Go seeder (reads seed_data.json → PostgreSQL)
├── .references/
│   └── kamus-bahasa-banjar-dialek-hulu.pdf
├── DICTIONARY_SPEC.md
└── PRD.md
```

---

## 10. TDD Approach

### 10.1 Test Layers

```
Unit tests     → domain/ (pure business logic, no I/O)
Integration    → infrastructure/ (real DB, mocked external)
API tests      → HTTP handlers (httptest)
```

### 10.2 Red-Green-Refactor Cycle

1. Write failing test that specifies desired behavior
2. Write minimum code to pass test
3. Refactor without breaking tests
4. Repeat

### 10.3 Key Test Scenarios

**Dictionary domain:**
- Word creation with valid/invalid fields
- Definition score calculation
- Related word linking

**Contribution workflow:**
- State machine: pending → approved/rejected/withdrawn
- Cannot approve already-approved contribution
- Cannot withdraw after approval

**Voting:**
- Cannot vote twice on same target
- Vote change updates counts correctly
- Vote removal deletes record

**Identity:**
- Password hashing and verification
- Role promotion only by admin
- Email uniqueness constraint

**AI — translation (public):**
- Guest and user can call `/ai/translate`
- Input exceeding 1000 chars returns `VALIDATION_ERROR`
- Rate limit enforced per IP (guest) and per user ID (auth)
- OpenRouter unavailable → returns `AI_UNAVAILABLE`, not stored
- Translation result is never persisted

**AI — enrichment (admin async):**
- Only admin can trigger enrichment jobs
- Job stored as `AIRequest` with `status: pending` immediately
- On OpenRouter failure: `AIRequest.status = failed`, raw error stored
- Approved output merges into word's definitions/examples with `source: ai_generated`
- Cannot approve an already-approved or rejected `AIRequest`

---

## 11. Milestones

| Phase | Scope |
|---|---|
| **Phase 1** | Project scaffold (Fiber, sqlc, golang-migrate), Identity context (register, email verify, login, JWT + Redis refresh token, password reset via SMTP) |
| **Phase 2** | Dictionary context (PostgreSQL schema, seed import from `seed_data.json`, CRUD, full-text search, public read API) |
| **Phase 3** | AI — Translation (`POST /ai/translate`, OpenRouter integration, Redis rate limiting, stateless response) |
| **Phase 4** | Community context (contributions workflow, votes, bookmarks, comments, flagging) |
| **Phase 5** | Moderation context (approval queue, bulk actions, user ban, audit log) |
| **Phase 6** | AI — Enrichment (admin-async enrichment jobs: `enrich_definition`, `suggest_example`, `suggest_related`, `quality_check`, review/approve flow) |
| **Phase 7** | Hardening (API docs/Swagger, coverage enforcement ≥80%, observability, deployment config) |

---

## 12. Technical Decisions Log

Previously open questions, now resolved:

| # | Decision | Choice | Rationale |
|---|---|---|---|
| 1 | Go web framework | **Fiber** | Preferred over chi |
| 2 | DB query layer | **sqlc** | Type-safe generated queries, no ORM overhead |
| 3 | Refresh token store | **Redis** | Instant revocation, TTL management built-in |
| 4 | OpenRouter default model | **Flexible via env** | Default to a free-tier model (`OPENROUTER_MODEL`); swap without code change |
| 5 | AI content visibility before approval | **Visible to all** | AI-tagged content shown to guests/users; labeled `source: ai_generated` |
| 6 | PDF data entry | **Automated extraction** | `scripts/seed/extract_dictionary.py` → `seed_data.json` |
| 7 | Email provider | **SMTP** | Configurable via `SMTP_*` env vars; bring your own provider |
| 8 | Rate limiter backend | **Redis** | Consistent across instances, uses sliding window counters |
