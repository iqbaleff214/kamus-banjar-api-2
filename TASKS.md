# TASKS.md — Kamus Banjar API 2

Task breakdown following the DDD bounded contexts and TDD milestones defined in
[PRD.md](PRD.md). Each task lists what to build, the Definition of Done (DoD),
and the unit/integration tests required before implementation is considered complete.

## How to Use

- Tasks are ordered — later tasks depend on earlier ones being ✅
- Follow **TDD**: write the failing test first, then the implementation
- Mark `[ ]` → `[x]` when all DoD items and tests pass
- Coverage gate: ≥ 80% on `internal/*/domain/` and `internal/*/application/`

---

## Phase 1 — Project Scaffold & Identity

### 1.1 — Project Scaffold

- [x] **1.1.1 — Initialise Go module and directory structure**
  - **Description:** Create the Go module (`go mod init`), set up the folder tree from PRD §9 (`cmd/api`, `internal/*`, `pkg/*`, `migrations/`, `scripts/`), add `.env.example` with all variables from PRD §7.6, and add a root `Makefile` with targets: `run`, `test`, `migrate-up`, `migrate-down`, `seed`, `lint`.
  - **DoD:**
    - `go build ./...` succeeds with no errors
    - `.env.example` contains every variable in PRD §7.6
    - `Makefile` has all six targets defined
    - Repository root matches the directory tree in PRD §9

- [x] **1.1.2 — Configure Fiber application entrypoint**
  - **Description:** Wire up the Fiber app in `cmd/api/main.go`. Load config from env (using a `pkg/config` package). Register a health-check route `GET /health` returning `{ "status": "ok" }`. Set Fiber options: `BodyLimit: 1MB`, `ReadTimeout: 10s`, `WriteTimeout: 10s`.
  - **DoD:**
    - `GET /health` returns HTTP 200 with `{ "status": "ok" }`
    - App fails fast with a clear error if any required env var is missing
    - Body limit of 1 MB is enforced (413 on larger payloads)
  - **Tests:**
    - `TestHealthCheck` — GET `/health` returns 200 and correct body
    - `TestBodyLimitEnforced` — POST with 2 MB body returns 413

- [x] **1.1.3 — Set up PostgreSQL connection and golang-migrate**
  - **Description:** Create `pkg/database` with a `Connect()` function returning a `*pgxpool.Pool`. Add `migrations/` directory. Write migration `000001_create_extensions.up.sql` enabling `uuid-ossp` and `pg_trgm` extensions.
  - **DoD:**
    - `make migrate-up` runs without error against a local PostgreSQL instance
    - `make migrate-down` reverts cleanly
    - Connection pool is validated on startup; app exits with error if DB is unreachable
  - **Tests:**
    - `TestDatabaseConnection` — integration test: pool pings DB successfully

- [x] **1.1.4 — Set up Redis connection**
  - **Description:** Create `pkg/cache` with a `Connect()` function returning a `*redis.Client` (using `go-redis/v9`). Validate connection on startup.
  - **DoD:**
    - App exits with a clear error if Redis is unreachable on startup
    - `PING` succeeds in integration test
  - **Tests:**
    - `TestRedisConnection` — integration test: PING returns PONG

- [x] **1.1.5 — Implement shared response envelopes and error types**
  - **Description:** Implement `pkg/httperr` with: `ErrorResponse` struct (`success: false`, `error.code`, `error.message`, `error.details`), `SuccessResponse` struct (`success: true`, `data`, `meta`), `PaginationMeta` struct, and typed error constructors for every error code in OAS §components/schemas/ErrorBody (`VALIDATION_ERROR`, `UNAUTHORIZED`, `FORBIDDEN`, `NOT_FOUND`, `CONFLICT`, `RATE_LIMITED`, `AI_UNAVAILABLE`, `INTERNAL_ERROR`).
  - **DoD:**
    - All error constructors produce JSON matching the OAS error schema exactly
    - `SuccessResponse` with pagination meta matches OAS success schema
  - **Tests:**
    - `TestErrorResponseShape` — each constructor produces correct `code` and HTTP status
    - `TestSuccessResponseShape` — `data` and `meta` are correctly nested

- [x] **1.1.6 — Implement pagination helper**
  - **Description:** Implement `pkg/pagination` with `ParseParams(c *fiber.Ctx) (page, perPage int, err error)` that reads `page` and `per_page` query params, defaults page=1 and per_page=20, enforces per_page ≤ 100, and computes the SQL `LIMIT`/`OFFSET`.
  - **DoD:**
    - `per_page > 100` is clamped or rejected with `VALIDATION_ERROR`
    - Negative or non-integer values return `VALIDATION_ERROR`
  - **Tests:**
    - `TestPaginationDefaults` — missing params return page=1, perPage=20
    - `TestPaginationMaxPerPage` — per_page=200 returns error or clamps to 100
    - `TestPaginationOffset` — page=3, per_page=20 returns offset=40

- [x] **1.1.7 — Implement JWT middleware and `pkg/auth`**
  - **Description:** Implement `pkg/auth` with: `GenerateAccessToken(userID, role string) (string, error)` (HS256, 15-min TTL), `ParseToken(token string) (*Claims, error)`, and a Fiber middleware `RequireAuth()` that reads the `Authorization: Bearer` header, validates the token, and injects `Claims` into `c.Locals`. Implement `RequireRole(roles ...string)` middleware that reads claims and returns 403 if role is not in the allowed list.
  - **DoD:**
    - Missing token → 401 with `UNAUTHORIZED`
    - Expired token → 401 with `UNAUTHORIZED`
    - Valid token with wrong role → 403 with `FORBIDDEN`
    - Valid token → handler runs with claims in context
  - **Tests:**
    - `TestGenerateAndParseToken` — round-trip: generate → parse returns same claims
    - `TestExpiredToken` — token with past expiry returns error from `ParseToken`
    - `TestRequireAuthMiddleware` — missing header returns 401
    - `TestRequireRoleMiddleware` — user token on admin route returns 403

---

### 1.2 — Identity Domain

- [x] **1.2.1 — Define `User` aggregate and value objects**
  - **Description:** Implement `internal/identity/domain/user.go` with the `User` struct (fields from PRD §4.3), the `Role` value type (`user`, `admin`), and domain methods: `NewUser(name, email, password string) (*User, error)` (validates fields, hashes password with bcrypt cost 12), `VerifyPassword(plain string) bool`, `Promote(role Role) error` (only admin role allowed), `Ban()`, `Unban()`.
  - **DoD:**
    - `NewUser` rejects empty name, invalid email, password shorter than 8 chars
    - Password is never stored in plain text
    - `Promote` rejects unknown roles
  - **Tests:**
    - `TestNewUser_ValidInput` — creates user with hashed password
    - `TestNewUser_InvalidEmail` — returns domain error
    - `TestNewUser_ShortPassword` — returns domain error
    - `TestVerifyPassword` — correct plain password returns true; wrong returns false
    - `TestPromote_InvalidRole` — returns error
    - `TestBanUnban` — toggles `is_active` correctly

- [x] **1.2.2 — Define `UserRepository` interface**
  - **Description:** Implement `internal/identity/domain/repository.go` with interface `UserRepository` containing: `Create(ctx, *User) error`, `FindByID(ctx, id UUID) (*User, error)`, `FindByEmail(ctx, email string) (*User, error)`, `Update(ctx, *User) error`.
  - **DoD:**
    - Interface defined; no implementation yet
    - All methods accept `context.Context` as first param

---

### 1.3 — Identity Infrastructure

- [x] **1.3.1 — Migration: `users` table**
  - **Description:** Write `migrations/000002_create_users.up.sql` creating the `users` table with all fields from PRD §4.3. Add unique index on `email`. Write the corresponding `.down.sql`.
  - **DoD:**
    - `make migrate-up` creates the table
    - `make migrate-down` drops it cleanly
    - Unique constraint on `email` verified by DB

- [x] **1.3.2 — sqlc: generate User queries**
  - **Description:** Write `internal/identity/infrastructure/postgres/queries/users.sql` with named queries: `CreateUser`, `GetUserByID`, `GetUserByEmail`, `UpdateUser`. Run `sqlc generate` to produce Go types. Implement `PostgresUserRepository` satisfying `UserRepository`.
  - **DoD:**
    - `sqlc generate` produces type-safe Go code with no errors
    - All four repository methods pass integration tests against a real test DB
  - **Tests (integration):**
    - `TestCreateUser` — inserts and retrieves by ID
    - `TestFindByEmail` — returns user; returns `NotFound` for unknown email
    - `TestEmailUniqueness` — second insert with same email returns conflict error
    - `TestUpdateUser` — change to `name` persists

- [x] **1.3.3 — Redis token store**
  - **Description:** Implement `internal/identity/infrastructure/redis/token_store.go` with: `StoreRefreshToken(ctx, token, userID string, ttl time.Duration) error`, `GetUserIDByRefreshToken(ctx, token string) (string, error)`, `RevokeRefreshToken(ctx, token string) error`. Key pattern: `refresh:<sha256(token)>`.
  - **DoD:**
    - Store → Get round-trip returns correct user ID
    - Revoke → Get returns `ErrTokenNotFound`
    - Expired key (TTL elapsed) → Get returns `ErrTokenNotFound`
  - **Tests (integration):**
    - `TestRefreshTokenRoundTrip`
    - `TestRevokeRefreshToken`
    - `TestExpiredRefreshToken`

- [x] **1.3.4 — SMTP email sender**
  - **Description:** Implement `pkg/mailer` with interface `Mailer { SendVerificationEmail(to, token string) error; SendPasswordResetEmail(to, token string) error }` and a `SMTPMailer` implementation reading `SMTP_*` env vars. Use Go's `net/smtp` or `gomail.v2`.
  - **DoD:**
    - Both methods send email with the token embedded in a link
    - In tests, use a `MockMailer` that captures sent messages
  - **Tests:**
    - `TestSendVerificationEmail_MockCapture` — mock records the call with correct `to` and non-empty token

---

### 1.4 — Identity Application

- [x] **1.4.1 — `RegisterUser` command**
  - **Description:** Implement `internal/identity/application/commands/register_user.go` with `RegisterUser(ctx, name, email, password, passwordConfirmation string) (*User, error)`. Validates passwords match, calls `NewUser`, calls `UserRepository.Create`, generates a verification token (random 32-byte hex), stores it in Redis with 24-hour TTL (`verify:<token>` → `user_id`), calls `Mailer.SendVerificationEmail`.
  - **DoD:**
    - Duplicate email returns `CONFLICT`
    - Password mismatch returns `VALIDATION_ERROR`
    - On success: user persisted, verification email sent, token in Redis
  - **Tests:**
    - `TestRegisterUser_Success` — all side-effects triggered
    - `TestRegisterUser_DuplicateEmail` — returns conflict error
    - `TestRegisterUser_PasswordMismatch` — returns validation error

- [x] **1.4.2 — `VerifyEmail` command**
  - **Description:** Implement `VerifyEmail(ctx, token string) error`. Looks up `verify:<token>` in Redis, sets `email_verified_at` on the user, updates in DB, deletes Redis key.
  - **DoD:**
    - Invalid/expired token returns `UNAUTHORIZED`
    - Valid token sets `email_verified_at` and invalidates the token
  - **Tests:**
    - `TestVerifyEmail_Valid`
    - `TestVerifyEmail_InvalidToken`
    - `TestVerifyEmail_AlreadyVerified` — idempotent (no error)

- [x] **1.4.3 — `LoginUser` command**
  - **Description:** Implement `LoginUser(ctx, email, password string) (accessToken, refreshToken string, err error)`. Finds user by email, verifies password, generates access token via `pkg/auth`, generates a random refresh token, stores it via `TokenStore.StoreRefreshToken` with 7-day TTL.
  - **DoD:**
    - Wrong email → `UNAUTHORIZED`
    - Wrong password → `UNAUTHORIZED`
    - Banned user → `FORBIDDEN`
    - Unverified email → `FORBIDDEN` (contribution actions gated separately, but login itself succeeds; implementation decision: allow login, gate contribution routes)
    - Success → returns both tokens
  - **Tests:**
    - `TestLoginUser_Success`
    - `TestLoginUser_WrongPassword`
    - `TestLoginUser_BannedUser`

- [x] **1.4.4 — `RefreshToken` command**
  - **Description:** Implement `RefreshToken(ctx, refreshToken string) (newAccessToken, newRefreshToken string, err error)`. Validates existing refresh token, issues new pair, revokes old refresh token (rotation).
  - **DoD:**
    - Invalid/expired refresh token → `UNAUTHORIZED`
    - Old token is revoked after refresh (cannot be reused)
  - **Tests:**
    - `TestRefreshToken_ValidRotation`
    - `TestRefreshToken_ReuseOldToken` — second use of old token returns unauthorized

- [x] **1.4.5 — `LogoutUser` command**
  - **Description:** Implement `LogoutUser(ctx, refreshToken string) error`. Calls `TokenStore.RevokeRefreshToken`.
  - **DoD:**
    - After logout, refresh token can no longer be used
  - **Tests:**
    - `TestLogoutUser_TokenRevoked`

- [x] **1.4.6 — `ForgotPassword` and `ResetPassword` commands**
  - **Description:** `ForgotPassword(ctx, email string) error` — finds user, generates reset token, stores in Redis (`reset:<token>` → `user_id`, 1-hour TTL), sends email. Always returns success (no user enumeration). `ResetPassword(ctx, token, password, passwordConfirmation string) error` — validates token, hashes new password, updates user, revokes token.
  - **DoD:**
    - Unknown email: no error returned, no email sent (silent)
    - Valid token: password updated, token invalidated
    - Expired/invalid token: `UNAUTHORIZED`
  - **Tests:**
    - `TestForgotPassword_UnknownEmail_NoLeak`
    - `TestResetPassword_ValidToken`
    - `TestResetPassword_ExpiredToken`

- [x] **1.4.7 — `UpdateProfile` and `ChangePassword` commands**
  - **Description:** `UpdateProfile(ctx, userID, name string) (*User, error)`. `ChangePassword(ctx, userID, currentPassword, newPassword string) error` — verifies current password before updating.
  - **DoD:**
    - Wrong current password → `UNAUTHORIZED`
    - Empty name → `VALIDATION_ERROR`
  - **Tests:**
    - `TestChangePassword_WrongCurrent`
    - `TestUpdateProfile_EmptyName`

---

### 1.5 — Identity HTTP

- [x] **1.5.1 — Auth routes and handlers**
  - **Description:** Implement `internal/identity/http/routes.go` registering all auth endpoints from OAS §5.6 on the Fiber app under `/api/v2/auth`. Implement handlers that call the application commands and return responses matching OAS schemas exactly.
  - **DoD:**
    - All 10 auth endpoints exist and return correct HTTP status codes
    - Request bodies that fail validation return 422 with `VALIDATION_ERROR` and `details`
    - `POST /auth/login` response includes `access_token`, `refresh_token`, `expires_in: 900`
  - **Tests (API/handler):**
    - `TestRegisterHandler_201`
    - `TestLoginHandler_200_TokensPresent`
    - `TestLoginHandler_401_WrongPassword`
    - `TestRefreshHandler_200`
    - `TestLogoutHandler_204`
    - `TestVerifyEmailHandler_204`
    - `TestForgotPasswordHandler_204`
    - `TestResetPasswordHandler_204`
    - `TestMeHandler_RequiresAuth`
    - `TestUpdateProfileHandler_200`
    - `TestChangePasswordHandler_204`

---

## Phase 2 — Dictionary

### 2.1 — Dictionary Domain

- [x] **2.1.1 — Define `WordClass`, `Dialect`, `Source` value types**
  - **Description:** Implement `internal/dictionary/domain/value_objects.go` with typed constants for `WordClass` (`n`, `v`, `a`, `adv`, `p`, `pb`, `ki`), `Dialect` (`hulu`), `Source` (`seeded`, `contributed`, `ai_generated`), and `WordStatus` (`active`, `deprecated`). Each type must have a `Validate() error` method.
  - **DoD:**
    - Unknown `WordClass` value returns a domain error
    - Unknown `Dialect` value returns a domain error
  - **Tests:**
    - `TestWordClass_Valid` — all 7 values pass
    - `TestWordClass_Invalid` — `"x"` returns error
    - `TestDialect_OnlyHulu` — `"kuala"` returns error (out of scope for v2)

- [x] **2.1.2 — Define `Definition` and `Example` value objects**
  - **Description:** Implement `internal/dictionary/domain/definition.go` and `example.go`. `Definition` has: `id`, `meaning` (max 2000 chars), `sort_order`, `source`, `upvotes`, `downvotes`. `NetScore() int` method. `Example` has: `id`, `banjar_sentence`, `indonesian_translation`, `source`.
  - **DoD:**
    - `meaning` exceeding 2000 chars returns domain error on construction
    - `NetScore()` returns `upvotes - downvotes`
  - **Tests:**
    - `TestDefinition_MeaningTooLong`
    - `TestDefinition_NetScore` — upvotes=5, downvotes=2 → NetScore=3

- [x] **2.1.3 — Define `Word` aggregate**
  - **Description:** Implement `internal/dictionary/domain/word.go` with the `Word` aggregate (all fields from PRD §4.1). Domain methods: `NewWord(banjar, wordClass, dialect string) (*Word, error)` (validates required fields), `AddDefinition(meaning string, sortOrder int) error`, `AddExample(banjar, indonesian string) error`, `Deprecate()`, `Restore()`, `SoftDelete()`. Derived field `IsDeleted() bool`.
  - **DoD:**
    - `NewWord` with empty `banjar` returns error
    - `NewWord` with invalid `word_class` returns error
    - `AddDefinition` with meaning > 2000 chars returns error
    - Deprecated word returns `status: deprecated`
  - **Tests:**
    - `TestNewWord_Valid`
    - `TestNewWord_EmptyBanjar`
    - `TestNewWord_InvalidWordClass`
    - `TestAddDefinition_TooLong`
    - `TestDeprecateAndRestore`

- [x] **2.1.4 — Define `WordRepository` interface**
  - **Description:** Implement `internal/dictionary/domain/repository.go` with `WordRepository` interface: `Create(ctx, *Word) error`, `FindByID(ctx, id) (*Word, error)`, `FindAll(ctx, filter WordFilter, page, perPage int) ([]*Word, int, error)`, `Search(ctx, query string, filter WordFilter, page, perPage int) ([]*Word, int, error)`, `Update(ctx, *Word) error`, `SoftDelete(ctx, id) error`. Define `WordFilter` struct with `WordClass`, `IsRoot *bool`, `Source`, `Status` fields.
  - **DoD:**
    - Interface compiles; no implementation yet

---

### 2.2 — Dictionary Infrastructure

- [x] **2.2.1 — Migration: `words`, `definitions`, `examples`, `word_relations` tables**
  - **Description:** Write `migrations/000003_create_dictionary.up.sql` creating:
    - `words` table with all fields from PRD §4.1, `deleted_at` for soft delete
    - `definitions` table with FK to `words`
    - `examples` table with FK to `words`
    - `word_relations` join table (`word_id`, `related_word_id`)
    - Indexes: `banjar`, `dialect`, `word_class`, `status`, `is_root`, `root_word_id`, `created_at`
    - Full-text search index: GIN index on `to_tsvector('indonesian', meaning)` on `definitions`, and `to_tsvector('simple', banjar)` on `words`
    - Unique constraint: `(banjar, dialect, homonym_number, is_root, root_word_id)`
  - **DoD:**
    - Migration applies and reverts cleanly
    - All indexes visible in `\d words` in psql
    - Unique constraint enforced by DB

- [x] **2.2.2 — sqlc: generate Word queries**
  - **Description:** Write SQL queries in `internal/dictionary/infrastructure/postgres/queries/` for: `CreateWord`, `GetWordByID`, `ListWords` (with filter + pagination), `SearchWords` (full-text on `banjar` and `definitions.meaning`), `UpdateWord`, `SoftDeleteWord`, `GetDefinitionsByWordID`, `GetExamplesByWordID`, `GetRelatedWords`. Run `sqlc generate`.
  - **DoD:**
    - `sqlc generate` succeeds with no warnings
    - `SearchWords` uses `@@` operator with `to_tsquery` and returns results ranked by relevance
  - **Tests (integration):**
    - `TestCreateAndFindWord`
    - `TestSearchWords_ByBanjar` — query "abah" returns word with banjar="abah"
    - `TestSearchWords_ByMeaning` — query "ayah" matches definition meaning
    - `TestListWords_FilterByWordClass`
    - `TestListWords_FilterByIsRoot`
    - `TestSoftDelete_HidesWord` — deleted word absent from `FindAll`
    - `TestUniqueConstraint` — duplicate (banjar, dialect, homonym_number) returns conflict

---

### 2.3 — Seed Data Import

- [x] **2.3.1 — Go seeder: import `seed_data.json` into PostgreSQL**
  - **Description:** Implement `scripts/seed/main.go`. Reads `scripts/seed/seed_data.json`, iterates entries, upserts words using the `WordRepository`. Root words are inserted first; derived forms inserted after with `root_word_id` set. Upsert key: `(banjar, dialect, homonym_number, is_root, root_word_id)`. On conflict: update `definitions`, `examples`, `word_class`, `updated_at`. Sets `source = seeded`, `created_by = NULL`. Idempotent.
  - **DoD:**
    - `make seed` runs without error on empty DB
    - `make seed` is idempotent — running twice does not duplicate entries
    - After seeding: at least 2,000 root words exist in `words` table
    - `source` is `seeded` for all imported entries
  - **Tests:**
    - `TestSeeder_Idempotent` — integration: run seeder twice, count unchanged
    - `TestSeeder_RootBeforeDerived` — derived forms have valid `root_word_id`

---

### 2.4 — Dictionary Application

- [x] **2.4.1 — `GetWord` query**
  - **Description:** Implement `internal/dictionary/application/queries/get_word.go` with `GetWord(ctx, id string) (*Word, error)`. Returns only `status: active` words (unless caller is admin, handled at HTTP layer by using the admin repository variant).
  - **DoD:**
    - Unknown ID → `NOT_FOUND`
    - Deleted word → `NOT_FOUND`
  - **Tests:**
    - `TestGetWord_Found`
    - `TestGetWord_NotFound`
    - `TestGetWord_SoftDeleted` — returns not found

- [x] **2.4.2 — `ListWords` and `SearchWords` queries**
  - **Description:** Implement `ListWords(ctx, filter, page, perPage int) ([]*Word, PaginationMeta, error)` and `SearchWords(ctx, query string, filter, page, perPage) ([]*Word, PaginationMeta, error)`. `SearchWords` falls back to `ListWords` when `query` is empty.
  - **DoD:**
    - Results respect `word_class`, `is_root`, `source` filters
    - Sort `alphabetical` = ORDER BY banjar ASC
    - Sort `most_voted` = ORDER BY net definition vote score DESC
    - Sort `recently_added` = ORDER BY created_at DESC
    - Pagination meta is accurate (`total` reflects filtered count)
  - **Tests:**
    - `TestListWords_Pagination`
    - `TestListWords_FilterWordClass`
    - `TestSearchWords_ReturnsRelevantResults`
    - `TestSearchWords_EmptyQuery_FallsbackToList`

- [x] **2.4.3 — `CreateWord`, `UpdateWord`, `DeleteWord` commands (admin)**
  - **Description:** Implement admin commands: `CreateWord(ctx, adminID, input WordInput) (*Word, error)` (sets `source: seeded`, `created_by: adminID`), `UpdateWord(ctx, adminID, wordID, input WordInput) (*Word, error)`, `SoftDeleteWord(ctx, wordID) error`. These bypass the contribution workflow.
  - **DoD:**
    - Creating a word with duplicate banjar+dialect+homonym returns `CONFLICT`
    - Updating a soft-deleted word returns `NOT_FOUND`
  - **Tests:**
    - `TestCreateWord_DuplicateConflict`
    - `TestUpdateWord_NotFound`
    - `TestSoftDeleteWord_AlreadyDeleted` — idempotent or error (define behavior)

---

### 2.5 — Dictionary HTTP

- [x] **2.5.1 — Public dictionary routes**
  - **Description:** Implement `internal/dictionary/http/routes.go` registering public endpoints from OAS §5.2 under `/api/v2`. No auth required. Apply rate limit middleware: 60 req/min per IP (guest), 120 req/min per user (if token present).
  - **DoD:**
    - `GET /words` returns paginated list with `success: true` and `meta`
    - `GET /words?q=abah` returns words matching search
    - `GET /words?word_class=n` filters correctly
    - `GET /words/:id` returns full word with nested definitions and examples
    - `GET /words/:id/definitions` returns definitions sorted by net score DESC
    - `GET /words/:id/examples` returns examples
    - `GET /words/:id/related` returns related words summary
    - `GET /words/search?q=...` works as alias
    - Unknown ID returns 404 with `NOT_FOUND`
  - **Tests (handler):**
    - `TestListWordsHandler_200`
    - `TestListWordsHandler_SearchQuery`
    - `TestListWordsHandler_Filter`
    - `TestGetWordHandler_200`
    - `TestGetWordHandler_404`
    - `TestGetDefinitionsHandler_SortedByScore`
    - `TestGetExamplesHandler_200`
    - `TestGetRelatedWordsHandler_200`

- [x] **2.5.2 — Admin word management routes**
  - **Description:** Register admin word endpoints from OAS §5.7 under `/api/v2/admin/words`. Require `RequireAuth()` + `RequireRole("admin")` middleware on all routes.
  - **DoD:**
    - Non-admin → 403
    - Unauthenticated → 401
    - `POST /admin/words` creates and returns the word
    - `PATCH /admin/words/:id` updates and returns the word
    - `DELETE /admin/words/:id` soft-deletes; subsequent GET returns 404
  - **Tests (handler):**
    - `TestAdminCreateWordHandler_403_NonAdmin`
    - `TestAdminCreateWordHandler_201`
    - `TestAdminUpdateWordHandler_200`
    - `TestAdminDeleteWordHandler_204`

---

## Phase 3 — AI Translation

### 3.1 — OpenRouter Client

- [ ] **3.1.1 — Implement OpenRouter HTTP client**
  - **Description:** Implement `internal/ai/infrastructure/openrouter/client.go` with interface `LLMClient { Complete(ctx context.Context, req CompletionRequest) (*CompletionResponse, error) }` and an `OpenRouterClient` implementation. `CompletionRequest` has: `model`, `messages []Message`, `temperature`. Reads `OPENROUTER_API_KEY`, `OPENROUTER_BASE_URL`, `OPENROUTER_MODEL` from config. Sets `HTTP-Referer` and `X-Title` headers required by OpenRouter.
  - **DoD:**
    - Client returns `ErrAIUnavailable` (mapped to `AI_UNAVAILABLE`) on non-2xx HTTP response or network error
    - `OPENROUTER_API_KEY` missing → app startup fails with clear error
  - **Tests:**
    - `TestOpenRouterClient_Success` — mock HTTP server returns valid response
    - `TestOpenRouterClient_NetworkError` — returns `ErrAIUnavailable`
    - `TestOpenRouterClient_Non200` — 503 from server returns `ErrAIUnavailable`

---

### 3.2 — Translation Application

- [ ] **3.2.1 — `TranslateText` command**
  - **Description:** Implement `internal/ai/application/commands/translate_text.go` with `TranslateText(ctx, text, context_ string) (*TranslationResult, error)`. Builds a system prompt that instructs the model to translate Banjar Hulu → Indonesian, noting dialect-specific vocabulary. Calls `LLMClient.Complete`. Parses the response into `TranslationResult` (`original`, `translation`, `dialect`, `model`, `confidence`, `notes`). Result is NOT persisted.
  - **DoD:**
    - Input text > 1000 chars → `VALIDATION_ERROR` before calling LLM
    - Empty text → `VALIDATION_ERROR`
    - LLM unavailable → propagates `ErrAIUnavailable`
    - No DB write occurs for any outcome
  - **Tests:**
    - `TestTranslateText_InputTooLong` — returns validation error, LLM not called
    - `TestTranslateText_EmptyText` — returns validation error
    - `TestTranslateText_Success` — mock LLM returns valid response, result parsed correctly
    - `TestTranslateText_LLMUnavailable` — error propagated, nothing stored

---

### 3.3 — Translation HTTP

- [ ] **3.3.1 — `POST /ai/translate` handler**
  - **Description:** Implement `internal/ai/http/routes.go` registering `POST /api/v2/ai/translate`. Require `RequireAuth()` (user or admin). Apply Redis rate limit: 30 req/hour per user ID. Call `TranslateText` command. Return OAS `TranslationResult` response on success.
  - **DoD:**
    - Unauthenticated request → 401
    - Guest (no token) → 401
    - Text > 1000 chars → 422
    - Rate limit exceeded → 429
    - LLM unavailable → 503
    - Success → 200 with all `TranslationResult` fields present
  - **Tests (handler):**
    - `TestTranslateHandler_401_NoToken`
    - `TestTranslateHandler_422_TooLong`
    - `TestTranslateHandler_200`
    - `TestTranslateHandler_503_LLMDown`
    - `TestTranslateHandler_429_RateLimit`

- [ ] **3.3.2 — Redis rate limiter middleware**
  - **Description:** Implement `pkg/ratelimit/middleware.go` with `Limiter(key func(*fiber.Ctx) string, limit int, window time.Duration) fiber.Handler` using a Redis sliding window counter (`INCR` + `EXPIRE` or sorted set approach). Return 429 with `RATE_LIMITED` when limit exceeded. Used by `/ai/translate`, `/contributions`, `/auth/login`, and admin AI endpoints.
  - **DoD:**
    - Under limit: request passes through
    - Over limit: 429 with `{ "success": false, "error": { "code": "RATE_LIMITED" } }`
    - Window resets correctly after `window` duration
  - **Tests (integration with Redis):**
    - `TestRateLimiter_UnderLimit`
    - `TestRateLimiter_AtLimit` — Nth request passes, (N+1)th returns 429
    - `TestRateLimiter_WindowReset` — counter resets after window

---

## Phase 4 — Community

### 4.1 — Contribution Domain

- [ ] **4.1.1 — Define `Contribution` aggregate**
  - **Description:** Implement `internal/community/domain/contribution.go` with `Contribution` struct (fields from PRD §4.2). Domain methods: `NewContribution(contributorID, type_, targetWordID *string, payload JSON) (*Contribution, error)`, `Approve(reviewerID, note string) error`, `Reject(reviewerID, note string) error`, `Withdraw(callerID string) error`. Status state machine enforced in domain.
  - **DoD:**
    - `Approve` on non-`pending` contribution returns `ErrInvalidTransition`
    - `Reject` on non-`pending` returns `ErrInvalidTransition`
    - `Withdraw` by non-contributor returns `ErrForbidden`
    - `Withdraw` after approval/rejection returns `ErrInvalidTransition`
    - `new_definition`, `new_example`, `edit_word` types require `target_word_id`
  - **Tests:**
    - `TestContribution_ApprovePending` — status becomes `approved`
    - `TestContribution_ApproveAlreadyApproved` — returns error
    - `TestContribution_RejectPending`
    - `TestContribution_WithdrawByNonOwner` — returns forbidden error
    - `TestContribution_WithdrawAfterApproval` — returns error
    - `TestContribution_NewDefinition_RequiresTargetWordID`

- [ ] **4.1.2 — Define `ContributionRepository` interface**
  - **Description:** Interface with: `Create`, `FindByID`, `FindByContributor(ctx, userID, filter, page, perPage)`, `FindAll(ctx, filter, page, perPage)`, `Update`.

---

### 4.2 — Vote Domain

- [ ] **4.2.1 — Define `Vote` aggregate**
  - **Description:** Implement `internal/community/domain/vote.go`. `CastVote(userID, targetType, targetID, value string) (*Vote, error)` validates `targetType ∈ {word, definition}` and `value ∈ {up, down}`.
  - **DoD:**
    - Invalid `targetType` → domain error
    - Invalid `value` → domain error
  - **Tests:**
    - `TestCastVote_Valid`
    - `TestCastVote_InvalidTargetType`
    - `TestCastVote_InvalidValue`

- [ ] **4.2.2 — Define `VoteRepository` interface**
  - **Description:** Interface with: `Upsert(ctx, *Vote) error` (insert or update if same user+target), `Delete(ctx, userID, targetType, targetID string) error`, `FindByUserAndTarget(ctx, userID, targetType, targetID string) (*Vote, error)`.

---

### 4.3 — Bookmark Domain

- [ ] **4.3.1 — Define `Bookmark` aggregate and repository**
  - **Description:** `Bookmark` struct with `id`, `user_id`, `word_id`, `created_at`. `BookmarkRepository` interface: `Create`, `Delete(ctx, userID, wordID)`, `FindByUser(ctx, userID, page, perPage)`, `Exists(ctx, userID, wordID) bool`.
  - **DoD:**
    - Creating a duplicate bookmark is caught at repository level (unique constraint)
  - **Tests:**
    - `TestBookmark_Create`

---

### 4.4 — Comment Domain

- [ ] **4.4.1 — Define `Comment` aggregate**
  - **Description:** Implement `internal/community/domain/comment.go`. `NewComment(userID, targetType, targetID, body string) (*Comment, error)` validates `body` is not empty and ≤ 1000 chars. `Edit(callerID, newBody string) error` enforces caller is owner. `Flag()` sets `is_flagged = true`. `CommentRepository` interface: `Create`, `FindByID`, `FindByTarget(ctx, targetType, targetID, page, perPage)`, `Update`, `Delete`.
  - **DoD:**
    - Body > 1000 chars → domain error
    - `Edit` by non-owner → `ErrForbidden`
  - **Tests:**
    - `TestNewComment_BodyTooLong`
    - `TestComment_EditByNonOwner`
    - `TestComment_Flag`

---

### 4.5 — Community Infrastructure

- [ ] **4.5.1 — Migration: community tables**
  - **Description:** Write `migrations/000004_create_community.up.sql` creating: `contributions`, `votes`, `bookmarks`, `comments` tables with all fields from PRD §4.2. Add unique constraints: `votes(user_id, target_type, target_id)`, `bookmarks(user_id, word_id)`.
  - **DoD:**
    - Migration applies and reverts cleanly
    - Unique constraints enforced by DB

- [ ] **4.5.2 — sqlc: generate community queries**
  - **Description:** Write and generate queries for all four community aggregates. Implement `PostgresContributionRepository`, `PostgresVoteRepository`, `PostgresBookmarkRepository`, `PostgresCommentRepository`.
  - **DoD:**
    - `sqlc generate` succeeds
    - All repository methods pass integration tests
  - **Tests (integration):**
    - `TestVoteUpsert_ChangeDirection` — vote up then down: only one vote record exists with value `down`
    - `TestVoteDelete`
    - `TestBookmark_DuplicateConflict`
    - `TestContribution_FindByContributor`
    - `TestComment_FindByTarget`

---

### 4.6 — Community Application

- [ ] **4.6.1 — Contribution commands**
  - **Description:** `SubmitContribution(ctx, userID, type_, targetWordID *string, payload) (*Contribution, error)` — validates user email is verified (else `FORBIDDEN`), validates `targetWordID` exists for non-`new_word` types, persists. `WithdrawContribution(ctx, callerID, contributionID) error`.
  - **DoD:**
    - Unverified email → `FORBIDDEN` with message explaining verification requirement
    - Non-existent `targetWordID` → `NOT_FOUND`
    - Rate limit: caller checks are done at HTTP layer
  - **Tests:**
    - `TestSubmitContribution_UnverifiedEmail`
    - `TestSubmitContribution_InvalidTargetWord`
    - `TestWithdrawContribution_Success`

- [ ] **4.6.2 — Vote commands**
  - **Description:** `CastVote(ctx, userID, targetType, targetID, value string) (*Vote, error)` — validates target exists (word or definition), calls `VoteRepository.Upsert`. `RemoveVote(ctx, userID, targetType, targetID string) error`.
  - **DoD:**
    - Non-existent target → `NOT_FOUND`
    - Casting on own word/definition: no restriction in v2
  - **Tests:**
    - `TestCastVote_TargetNotFound`
    - `TestCastVote_ChangeDirection` — second vote changes value, count reflects correctly

- [ ] **4.6.3 — Bookmark commands**
  - **Description:** `AddBookmark(ctx, userID, wordID string) (*Bookmark, error)`. `RemoveBookmark(ctx, userID, wordID string) error`. `ListBookmarks(ctx, userID, page, perPage)`.
  - **DoD:**
    - Adding duplicate bookmark → `CONFLICT`
    - Removing non-existent bookmark → `NOT_FOUND`
  - **Tests:**
    - `TestAddBookmark_Duplicate`
    - `TestRemoveBookmark_NotFound`

- [ ] **4.6.4 — Comment commands**
  - **Description:** `PostComment(ctx, userID, wordID, body string) (*Comment, error)`. `EditComment(ctx, callerID, commentID, body string) (*Comment, error)`. `DeleteComment(ctx, callerID, callerRole, commentID string) error` (admin can delete any). `FlagComment(ctx, callerID, commentID string) error`.
  - **DoD:**
    - Edit by non-owner → `FORBIDDEN`
    - Delete by non-owner non-admin → `FORBIDDEN`
    - Flagging already-flagged comment → idempotent (no error)
  - **Tests:**
    - `TestEditComment_NonOwner`
    - `TestDeleteComment_AdminCanDeleteAny`
    - `TestFlagComment_Idempotent`

---

### 4.7 — Community HTTP

- [ ] **4.7.1 — Contribution routes**
  - **Description:** Register contribution endpoints from OAS §5.3. All require `RequireAuth()`. `POST /contributions` has rate limit: 10/hour per user. `approve` and `reject` sub-routes require `RequireRole("admin")`.
  - **DoD:**
    - `POST /contributions` → 201 on success
    - `GET /contributions` → user sees own; admin sees all
    - `PATCH /contributions/:id/withdraw` → 409 if not pending
    - `PATCH /contributions/:id/approve` → 403 for non-admin
    - `PATCH /contributions/:id/reject` → requires `note` field; 422 if missing
  - **Tests (handler):**
    - `TestSubmitContributionHandler_201`
    - `TestSubmitContributionHandler_429_RateLimit`
    - `TestApproveContributionHandler_403_NonAdmin`
    - `TestRejectContributionHandler_422_MissingNote`
    - `TestWithdrawContributionHandler_409_NotPending`

- [ ] **4.7.2 — Vote, Bookmark, Comment routes**
  - **Description:** Register all community endpoints from OAS §5.4. All write operations require `RequireAuth()`. Comment reads (`GET /words/:id/comments`) are public.
  - **DoD:**
    - `POST /words/:id/votes` with same value twice: second call updates (upsert), no 409
    - `DELETE /words/:id/votes` removes vote; 404 if no vote existed
    - `POST /bookmarks` duplicate → 409
    - `DELETE /bookmarks/:word_id` → 204
    - `PATCH /comments/:id` by non-owner → 403
    - `DELETE /comments/:id` by admin → 204
    - `POST /comments/:id/flag` → 200; second flag → 200 (idempotent)
  - **Tests (handler):**
    - `TestCastVoteHandler_201`
    - `TestCastVoteHandler_UpdateDirection`
    - `TestRemoveVoteHandler_404_NoneExists`
    - `TestAddBookmarkHandler_409_Duplicate`
    - `TestPostCommentHandler_201`
    - `TestEditCommentHandler_403_NonOwner`
    - `TestDeleteCommentHandler_204_Admin`
    - `TestFlagCommentHandler_200_Idempotent`

---

## Phase 5 — Moderation

### 5.1 — Moderation Domain

- [ ] **5.1.1 — Define `AuditLog` entity**
  - **Description:** Implement `internal/moderation/domain/audit_log.go` with `AuditLog` struct: `id`, `actor_id` (admin), `action` (string enum: `approve_contribution`, `reject_contribution`, `ban_user`, `unban_user`, `change_role`, `delete_word`, `approve_ai`, `reject_ai`), `target_type`, `target_id`, `metadata` (JSON), `created_at`. `NewAuditLog(actorID, action, targetType, targetID string, metadata map[string]any) *AuditLog`.
  - **DoD:**
    - All defined action values are valid
    - Unknown action value returns domain error

- [ ] **5.1.2 — Define `ModerationRepository` interface**
  - **Description:** Interface with: `GetPendingContributions(ctx, filter, page, perPage) ([]*Contribution, int, error)`, `GetFlaggedComments(ctx, page, perPage) ([]*Comment, int, error)`, `GetStats(ctx) (*ModerationStats, error)`, `CreateAuditLog(ctx, *AuditLog) error`.

---

### 5.2 — Moderation Infrastructure

- [ ] **5.2.1 — Migration: `audit_logs` table**
  - **Description:** Write `migrations/000005_create_moderation.up.sql` creating `audit_logs` table. Index on `actor_id`, `action`, `created_at`.
  - **DoD:**
    - Migration applies and reverts cleanly

- [ ] **5.2.2 — sqlc: generate moderation queries**
  - **Description:** Queries for pending contributions (status=pending, sorted by submitted_at ASC), flagged comments, moderation stats (counts), insert audit log. Implement `PostgresModerationRepository`.
  - **Tests (integration):**
    - `TestGetPendingContributions_FilterByType`
    - `TestGetModerationStats`

---

### 5.3 — Moderation Application

- [ ] **5.3.1 — `ApproveContribution` command**
  - **Description:** Implement `ApproveContribution(ctx, adminID, contributionID, note string) error`. Loads contribution, calls `contribution.Approve()`, merges payload into dictionary (creates Word, Definition, or Example depending on contribution type), persists contribution update, writes `AuditLog`.
  - **DoD:**
    - `new_word` contribution: creates a new Word with `source: contributed`
    - `new_definition`: adds Definition to existing word
    - `new_example`: adds Example to existing word
    - `edit_word`: applies field changes to existing word
    - Audit log written for every successful approval
    - Non-pending contribution → `CONFLICT`
  - **Tests:**
    - `TestApproveContribution_NewWord_CreatesWord`
    - `TestApproveContribution_NewDefinition_AddsDefinition`
    - `TestApproveContribution_NonPending_Conflict`
    - `TestApproveContribution_WritesAuditLog`

- [ ] **5.3.2 — `RejectContribution` command**
  - **Description:** `RejectContribution(ctx, adminID, contributionID, note string) error`. Note is required. Writes audit log.
  - **DoD:**
    - Empty note → `VALIDATION_ERROR`
    - Non-pending → `CONFLICT`
    - Audit log written
  - **Tests:**
    - `TestRejectContribution_EmptyNote`
    - `TestRejectContribution_WritesAuditLog`

- [ ] **5.3.3 — `BanUser` and `UnbanUser` commands**
  - **Description:** `BanUser(ctx, adminID, targetUserID, reason string) error` — cannot ban another admin. `UnbanUser(ctx, adminID, targetUserID string) error`. Both write audit log.
  - **DoD:**
    - Admin cannot ban another admin → `FORBIDDEN`
    - Banning already-banned user → idempotent
    - Audit log written with reason in metadata
  - **Tests:**
    - `TestBanUser_CannotBanAdmin`
    - `TestBanUser_WritesAuditLog`
    - `TestUnbanUser_NotBanned_Idempotent`

- [ ] **5.3.4 — `ChangeUserRole` command**
  - **Description:** `ChangeUserRole(ctx, adminID, targetUserID string, role Role) error`. Admin cannot demote themselves.
  - **DoD:**
    - Admin demoting themselves → `FORBIDDEN`
    - Audit log written
  - **Tests:**
    - `TestChangeUserRole_SelfDemotion_Forbidden`
    - `TestChangeUserRole_WritesAuditLog`

---

### 5.4 — Moderation HTTP

- [ ] **5.4.1 — Admin moderation and user management routes**
  - **Description:** Register all admin endpoints from OAS §5.7 under `/api/v2/admin`. All require `RequireRole("admin")`. Wire to moderation application commands.
  - **DoD:**
    - `GET /admin/moderation/queue` returns pending contributions paginated
    - `GET /admin/moderation/flags` returns flagged comments paginated
    - `GET /admin/moderation/stats` returns all four counts
    - `GET /admin/users` supports `role`, `is_active`, `q` filters
    - `PATCH /admin/users/:id/ban` → 200 with updated user
    - `PATCH /admin/users/:id/role` → 422 for invalid role value
    - All endpoints return 403 for non-admin
  - **Tests (handler):**
    - `TestModerationQueueHandler_200`
    - `TestFlaggedCommentsHandler_200`
    - `TestModerationStatsHandler_200`
    - `TestBanUserHandler_200`
    - `TestBanUserHandler_403_BanAdmin`
    - `TestChangeRoleHandler_422_InvalidRole`
    - `TestAdminUsersHandler_FilterByRole`

---

## Phase 6 — AI Enrichment

### 6.1 — AIRequest Domain

- [ ] **6.1.1 — Define `AIRequest` aggregate**
  - **Description:** Implement `internal/ai/domain/ai_request.go` with `AIRequest` struct (all fields from PRD §4.4). Domain methods: `NewAIRequest(type_, targetWordID, targetContributionID *string, requestedBy, model, prompt string) (*AIRequest, error)`, `MarkCompleted(response, parsedOutput JSON) error`, `MarkFailed(rawError JSON) error`, `Approve(reviewerID string) error`, `Reject(reviewerID string) error`.
  - **DoD:**
    - `Approve` on `quality_check` type → `ErrCannotApproveQualityCheck`
    - `Approve` on already-approved → `ErrInvalidTransition`
    - `Reject` on already-rejected → `ErrInvalidTransition`
    - `MarkCompleted` on non-pending → `ErrInvalidTransition`
  - **Tests:**
    - `TestAIRequest_ApproveQualityCheck_Error`
    - `TestAIRequest_ApproveAlreadyApproved_Error`
    - `TestAIRequest_RejectAlreadyRejected_Error`
    - `TestAIRequest_MarkCompleted_StateTransition`

- [ ] **6.1.2 — Define `AIRequestRepository` interface**
  - **Description:** Interface with: `Create`, `FindByID`, `FindAll(ctx, filter, page, perPage)`, `Update`.

---

### 6.2 — AI Enrichment Infrastructure

- [ ] **6.2.1 — Migration: `ai_requests` table**
  - **Description:** Write `migrations/000006_create_ai_requests.up.sql` creating `ai_requests` table with all fields from PRD §4.4. Index on `target_word_id`, `status`, `review_status`, `type`, `created_at`.
  - **DoD:**
    - Migration applies and reverts cleanly

- [ ] **6.2.2 — sqlc: generate AI request queries**
  - **Description:** Queries: `CreateAIRequest`, `GetAIRequestByID`, `ListAIRequests` (filter by type/status/review_status + pagination), `UpdateAIRequest`. Implement `PostgresAIRequestRepository`.
  - **Tests (integration):**
    - `TestCreateAndFindAIRequest`
    - `TestListAIRequests_FilterByReviewStatus`

---

### 6.3 — AI Enrichment Application

- [ ] **6.3.1 — `TriggerEnrichment` commands (enrich, example, related)**
  - **Description:** Implement three commands sharing the same pattern — `TriggerDefinitionEnrichment(ctx, adminID, wordID string) (*AIRequest, error)`, `TriggerExampleSuggestion(...)`, `TriggerRelatedWordSuggestion(...)`. Each: loads the word, builds a context-specific prompt including the word, its existing definitions, and Banjar dialect context, creates `AIRequest` with `status: pending`, calls `LLMClient.Complete` (potentially async — return pending request immediately, complete in goroutine), updates `AIRequest` with response or failure.
  - **DoD:**
    - Word not found → `NOT_FOUND`, no `AIRequest` created
    - LLM unavailable → `AIRequest.status = failed`, raw error stored in `response`
    - On LLM success → `AIRequest.status = completed`, `parsed_output` populated
    - Endpoint returns 202 with the pending `AIRequest` immediately
    - Rate limit: 50 req/hour per admin (enforced at HTTP layer)
  - **Tests:**
    - `TestTriggerEnrichment_WordNotFound`
    - `TestTriggerEnrichment_LLMFails_StatusFailed`
    - `TestTriggerEnrichment_LLMSucceeds_StatusCompleted`

- [ ] **6.3.2 — `TriggerQualityCheck` command**
  - **Description:** `TriggerQualityCheck(ctx, adminID, contributionID string) (*AIRequest, error)`. Loads contribution payload, builds a quality-check prompt asking the model to evaluate accuracy, consistency with BBDH vocabulary, and flag issues. Stores result in `AIRequest`. No approval flow — purely advisory.
  - **DoD:**
    - Contribution not found → `NOT_FOUND`
    - Result stored in `parsed_output` with at least `score` and `notes` fields
  - **Tests:**
    - `TestTriggerQualityCheck_ContributionNotFound`
    - `TestTriggerQualityCheck_StoresAdvisoryOutput`

- [ ] **6.3.3 — `ApproveAIRequest` command**
  - **Description:** `ApproveAIRequest(ctx, adminID, requestID string) (*AIRequest, error)`. Loads `AIRequest`, calls `aiRequest.Approve()`, merges `parsed_output` into the target word (adds definitions/examples/related_words with `source: ai_generated`), updates `AIRequest`, writes audit log.
  - **DoD:**
    - `quality_check` type → `CONFLICT` (cannot approve advisory requests)
    - Already-approved → `CONFLICT`
    - On success: word updated with `source: ai_generated` content
    - Audit log written
  - **Tests:**
    - `TestApproveAIRequest_QualityCheck_Conflict`
    - `TestApproveAIRequest_AlreadyApproved_Conflict`
    - `TestApproveAIRequest_MergesOutput`

- [ ] **6.3.4 — `RejectAIRequest` command**
  - **Description:** `RejectAIRequest(ctx, adminID, requestID string) (*AIRequest, error)`. Calls `aiRequest.Reject()`, updates DB, writes audit log.
  - **DoD:**
    - Already-rejected → `CONFLICT`
    - Audit log written
  - **Tests:**
    - `TestRejectAIRequest_AlreadyRejected_Conflict`

---

### 6.4 — AI Enrichment HTTP

- [ ] **6.4.1 — Admin AI enrichment routes**
  - **Description:** Register all admin AI endpoints from OAS §5.7 under `/api/v2/admin/ai`. All require `RequireRole("admin")`. Apply rate limit: 50 req/hour per admin ID on trigger endpoints (`POST /admin/ai/enrich/:word_id`, `/example/:word_id`, `/related/:word_id`, `/check/:contribution_id`).
  - **DoD:**
    - All trigger endpoints return 202 immediately
    - `GET /admin/ai/requests` supports `type`, `status`, `review_status` filters
    - `PATCH /admin/ai/requests/:id/approve` on quality_check type → 409
    - `PATCH /admin/ai/requests/:id/approve` on already-approved → 409
    - Rate limit exceeded → 429
    - Non-admin → 403
  - **Tests (handler):**
    - `TestEnrichHandler_202`
    - `TestEnrichHandler_404_WordNotFound`
    - `TestEnrichHandler_403_NonAdmin`
    - `TestEnrichHandler_429_RateLimit`
    - `TestListAIRequestsHandler_FilterByType`
    - `TestApproveAIRequestHandler_409_QualityCheck`
    - `TestApproveAIRequestHandler_200`
    - `TestRejectAIRequestHandler_200`

---

## Phase 7 — Hardening

### 7.1 — Rate Limiting Integration

- [ ] **7.1.1 — Apply all rate limits from PRD §7.4**
  - **Description:** Audit every route group and ensure the rate limiter middleware from task 3.3.2 is applied with correct limits: `GET /words*` (60/min IP for guest, 120/min user ID for auth), `POST /contributions` (10/hour user ID), `POST /auth/login` (5/min IP), `POST /ai/translate` (30/hour user ID), `POST /admin/ai/*` (50/hour admin ID).
  - **DoD:**
    - Each rate limit verified by an automated test hitting the endpoint N+1 times
  - **Tests:**
    - `TestRateLimit_LoginEndpoint`
    - `TestRateLimit_ContributionEndpoint`
    - `TestRateLimit_TranslateEndpoint`

---

### 7.2 — Observability & Configuration

- [ ] **7.2.1 — Structured logging**
  - **Description:** Add structured JSON logging using `slog` (stdlib). Log every request (method, path, status, latency) via Fiber middleware. Log every error at ERROR level with stack context. Log AI request outcomes at INFO level.
  - **DoD:**
    - Every HTTP request produces a log line with `method`, `path`, `status`, `latency_ms`
    - Errors include `error` and `request_id` fields
    - Logs are JSON in `APP_ENV=production`, human-readable in `development`

- [ ] **7.2.2 — Request ID middleware**
  - **Description:** Add Fiber middleware that sets a `X-Request-ID` header on every response (generate UUID if not present in request).
  - **DoD:**
    - Every response has a non-empty `X-Request-ID` header
  - **Tests:**
    - `TestRequestIDMiddleware_GeneratesID`
    - `TestRequestIDMiddleware_PreservesClientID`

- [ ] **7.2.3 — Graceful shutdown**
  - **Description:** Handle `SIGINT`/`SIGTERM` in `cmd/api/main.go`. Drain active connections with 30-second timeout before exiting.
  - **DoD:**
    - `kill -TERM <pid>` causes the server to finish in-flight requests before exiting
    - Exit code is 0 on clean shutdown

---

### 7.3 — Test Coverage

- [ ] **7.3.1 — Enforce coverage gate**
  - **Description:** Add a `Makefile` target `make coverage` that runs `go test -coverprofile=coverage.out ./internal/...` and fails if coverage in `domain/` and `application/` packages is below 80%. Use `go tool cover` or a simple `awk` check.
  - **DoD:**
    - `make coverage` passes (≥ 80% on domain + application layers)
    - CI pipeline runs `make coverage` and fails the build if below threshold
  - **Tests:**
    - N/A — this task is the coverage gate itself

- [ ] **7.3.2 — Integration test database setup**
  - **Description:** Create `testutil/db.go` with a helper `SetupTestDB(t *testing.T) *pgxpool.Pool` that creates a throwaway schema per test using `t.Cleanup` to drop it. Allows parallel integration tests without interference.
  - **DoD:**
    - Integration tests can run in parallel without state leaking between them
    - Each test gets a fresh schema; `t.Cleanup` drops it after the test

---

### 7.4 — API Documentation

- [ ] **7.4.1 — Serve OpenAPI spec via Swagger UI**
  - **Description:** Serve the `openapi.yaml` file at `GET /docs/openapi.yaml`. Serve Swagger UI at `GET /docs` using `scalar` or `swagger-ui` (embed as static files using Go's `embed` package).
  - **DoD:**
    - `GET /docs/openapi.yaml` returns the YAML with `Content-Type: application/yaml`
    - `GET /docs` renders an interactive UI that lets users try endpoints

- [ ] **7.4.2 — Validate all API responses match OAS schemas**
  - **Description:** Add an optional test mode where every HTTP response is validated against the `openapi.yaml` schema using a Go OpenAPI validator (e.g. `kin-openapi`). Run these as part of API/handler tests.
  - **DoD:**
    - Any handler that returns a response shape not matching OAS causes a test failure
    - Validation covers at minimum: success response `data` shape and error envelope shape

---

### 7.5 — Deployment Configuration

- [ ] **7.5.1 — Dockerfile**
  - **Description:** Write a multi-stage `Dockerfile`: stage 1 builds the Go binary; stage 2 is a minimal `gcr.io/distroless/static` image with just the binary. `EXPOSE 8080`.
  - **DoD:**
    - `docker build` succeeds
    - `docker run` with correct env vars starts the API and responds to `GET /health`
    - Image is ≤ 30 MB

- [ ] **7.5.2 — Docker Compose for local development**
  - **Description:** Write `docker-compose.yml` with services: `api` (built from Dockerfile), `db` (PostgreSQL 15), `redis` (Redis 7), `migrate` (one-shot container running `make migrate-up`). Include health checks so `api` waits for `db` and `redis`.
  - **DoD:**
    - `docker compose up` starts all services
    - `GET /health` succeeds after services are healthy
    - `GET /words` returns data after `make seed` runs

---

## Cross-Cutting Checklist

These apply across all phases and should be verified before each phase is considered complete:

- [ ] All new routes have corresponding handler tests (HTTP layer)
- [ ] All domain methods have unit tests with edge cases
- [ ] All repository methods have integration tests against a real test DB
- [ ] All 401/403/404/409/422/429 responses are tested explicitly
- [ ] No plaintext secrets in code or committed `.env` files
- [ ] `go vet ./...` and `golangci-lint run` pass with no new warnings
- [ ] `sqlc generate` is re-run after any SQL query changes
- [ ] `make migrate-up && make migrate-down` applies and reverts cleanly for every new migration
