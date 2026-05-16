# Product Requirements Document
# Kamus Banjar API 2

**Version:** 2.0.0  
**Date:** 2026-05-16  
**Status:** Draft  

---

## 1. Overview

### 1.1 Product Summary

Kamus Banjar API 2 is a RESTful API platform for the Banjar language dictionary (Dialek Hulu), digitized from the reference book *Kamus Bahasa Banjar Dialek Hulu*. It is a community-driven dictionary platform that supports public read access, authenticated user contributions, AI-assisted content enrichment via OpenRouter, and admin moderation tooling.

### 1.2 Goals

- Digitize and expose Banjar Hulu dialect dictionary entries via a public REST API
- Enable community contributions with a structured approval workflow
- Integrate AI (via OpenRouter) for definition enrichment, translation suggestions, and content quality assistance
- Build a moderation layer to ensure data integrity and community health
- Serve as an open-source, long-term reference infrastructure for the Banjar language

### 1.3 Non-Goals

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
| Register / login | ✓ | — | — |
| Bookmark words | ✗ | ✓ | ✓ |
| Upvote / downvote word | ✗ | ✓ | ✓ |
| Submit word contribution | ✗ | ✓ | ✓ |
| Submit definition contribution | ✗ | ✓ | ✓ |
| Write comments/reviews | ✗ | ✓ | ✓ |
| Edit own contributions (pending) | ✗ | ✓ | ✓ |
| Add word directly (no approval) | ✗ | ✗ | ✓ |
| Edit any word directly | ✗ | ✗ | ✓ |
| Approve / reject contributions | ✗ | ✗ | ✓ |
| Delete words / definitions | ✗ | ✗ | ✓ |
| Flag/unflag content | ✗ | ✓ | ✓ |
| Manage users | ✗ | ✗ | ✓ |
| View moderation queue | ✗ | ✗ | ✓ |
| Trigger AI enrichment | ✗ | ✗ | ✓ |

---

## 4. Domain Model

### 4.1 Dictionary Context

**Aggregate: `Word`**

| Field | Type | Description |
|---|---|---|
| `id` | UUID | Primary identifier |
| `banjar` | string | Banjar word (Dialek Hulu) |
| `latin` | string | Latin script romanization |
| `dialect` | enum | `hulu` \| `kuala` |
| `word_class` | enum | Noun, verb, adjective, adverb, etc. |
| `definitions` | `Definition[]` | One or more definitions |
| `examples` | `Example[]` | Usage example sentences |
| `etymology` | string? | Origin or root word info |
| `related_words` | UUID[] | References to related `Word` IDs |
| `status` | enum | `active` \| `deprecated` |
| `created_by` | UUID | Admin user who seeded/approved |
| `created_at` | datetime | — |
| `updated_at` | datetime | — |

**Value Object: `Definition`**

| Field | Type | Description |
|---|---|---|
| `id` | UUID | — |
| `meaning` | string | Indonesian/Melayu meaning |
| `register` | enum | `formal` \| `informal` \| `archaic` \| `slang` |
| `source` | enum | `seeded` \| `contributed` \| `ai_generated` |
| `upvotes` | int | Community upvote count |
| `downvotes` | int | Community downvote count |

**Value Object: `Example`**

| Field | Type | Description |
|---|---|---|
| `id` | UUID | — |
| `banjar_sentence` | string | Banjar example sentence |
| `translation` | string | Indonesian translation |
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

**Aggregate: `AIRequest`**

| Field | Type | Description |
|---|---|---|
| `id` | UUID | — |
| `type` | enum | `enrich_definition` \| `suggest_example` \| `translate` \| `quality_check` |
| `target_word_id` | UUID | — |
| `requested_by` | UUID | Admin ID |
| `prompt` | string | Prompt sent to OpenRouter |
| `response` | JSON? | Raw OpenRouter response |
| `status` | enum | `pending` \| `completed` \| `failed` |
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

### 5.5 Identity Endpoints

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

### 5.6 Admin Endpoints

```
GET    /admin/words                  # List all words including inactive
POST   /admin/words                  # Create word directly
PATCH  /admin/words/:id              # Update word directly
DELETE /admin/words/:id              # Soft-delete word

GET    /admin/users                  # List users
GET    /admin/users/:id              # Get user detail
PATCH  /admin/users/:id/ban          # Ban user
PATCH  /admin/users/:id/unban        # Unban user
PATCH  /admin/users/:id/role         # Change user role

GET    /admin/moderation/queue       # Pending contributions
GET    /admin/moderation/flags       # Flagged comments
GET    /admin/moderation/stats       # Moderation statistics

POST   /admin/ai/enrich/:word_id     # Trigger AI enrichment for word
GET    /admin/ai/requests            # List AI request history
GET    /admin/ai/requests/:id        # Get AI request detail
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

- Full-text search on `banjar`, `latin`, and `meaning` fields
- Filter by: `word_class`, `dialect`, `register`
- Sort by: `alphabetical`, `most_voted`, `recently_added`
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

Triggered by admin only. Three use cases:

| Type | Input | Output |
|---|---|---|
| `enrich_definition` | Word + existing definitions | Suggested additional definitions |
| `suggest_example` | Word + definition | Suggested Banjar example + translation |
| `quality_check` | Contributed payload | Consistency/quality score + notes |

- All AI output is stored as `AIRequest` with raw response
- AI-generated content marked `source: ai_generated`
- Admin must explicitly approve AI output before it becomes canonical
- OpenRouter model configurable via environment variable

### 6.7 Moderation Tools

- Pending contribution queue with filter/sort
- Flagged comment queue
- Bulk approve/reject on contributions
- User ban with reason
- Moderation audit log (who acted, when, on what)

---

## 7. Technical Requirements

### 7.1 Stack (Recommended)

| Layer | Technology |
|---|---|
| Language | Go |
| Framework | net/http + chi router (or Fiber) |
| ORM / DB Layer | sqlc or GORM |
| Database | PostgreSQL |
| Auth | JWT (access + refresh token) |
| AI | OpenRouter HTTP API |
| Testing | testify, gomock |
| Migration | golang-migrate |
| Config | Environment variables (.env) |

### 7.2 Database

- PostgreSQL 15+
- UUID primary keys
- Soft delete via `deleted_at` on mutable aggregates
- Indexes on: `banjar`, `latin`, `status`, `contributor_id`, `created_at`
- Full-text search index on `banjar`, `latin`, `meaning`

### 7.3 Authentication

- JWT with short-lived access token (15 min) + refresh token (7 days)
- Refresh token stored server-side (DB or Redis) for revocation
- Email verification required for contribution privileges
- Password: bcrypt, min cost factor 12

### 7.4 Rate Limiting

| Endpoint Group | Limit |
|---|---|
| `GET /words*` (guest) | 60 req/min per IP |
| `GET /words*` (auth) | 120 req/min per user |
| `POST /contributions` | 10 req/hour per user |
| `POST /auth/login` | 5 req/min per IP |
| `POST /admin/ai/*` | 20 req/hour per admin |

### 7.5 Seeding

- Initial dictionary data seeded from `.references/kamus-bahasa-banjar-dialek-hulu.pdf`
- Seed script parses PDF and inserts words with `source: seeded` and `created_by: system`
- Seeding is idempotent (upsert by `banjar` + `dialect`)

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
│   └── validator/
├── migrations/
├── scripts/
│   └── seed/           # PDF parsing + seed logic
├── .references/
│   └── kamus-bahasa-banjar-dialek-hulu.pdf
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

**AI:**
- Graceful failure when OpenRouter unavailable
- AI output stored regardless of success/failure
- Only admin can trigger AI requests

---

## 11. Milestones

| Phase | Scope |
|---|---|
| **Phase 1** | Identity context (auth, register, JWT), base project structure, DB migrations |
| **Phase 2** | Dictionary context (seed, CRUD, search), public read API |
| **Phase 3** | Community context (contributions, votes, bookmarks, comments) |
| **Phase 4** | Moderation context (approval queue, flagging, admin tools) |
| **Phase 5** | AI context (OpenRouter integration, enrichment, quality check) |
| **Phase 6** | Hardening (rate limiting, audit log, coverage, docs) |

---

## 12. Open Questions

| # | Question | Owner |
|---|---|---|
| 1 | Which Go web framework: chi vs Fiber? | Tech lead |
| 2 | Use sqlc or GORM for DB layer? | Tech lead |
| 3 | Redis required for refresh token store, or DB table sufficient? | Tech lead |
| 4 | Which OpenRouter model default for Banjar language tasks? | Product |
| 5 | Should AI-generated content be surfaced to guests or users before admin approval? | Product |
| 6 | PDF parsing: manual data entry vs automated extraction script? | Tech lead |
| 7 | Email provider for verification/notification emails? | Ops |
| 8 | Rate limiting: in-process (memory) or Redis-backed? | Tech lead |
