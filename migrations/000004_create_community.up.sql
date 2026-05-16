CREATE TYPE contribution_type   AS ENUM ('new_word', 'new_definition', 'new_example', 'edit_word');
CREATE TYPE contribution_status AS ENUM ('pending', 'approved', 'rejected', 'withdrawn');
CREATE TYPE vote_target_type    AS ENUM ('word', 'definition');
CREATE TYPE vote_value          AS ENUM ('up', 'down');
CREATE TYPE comment_target_type AS ENUM ('word', 'contribution');

-- ─── contributions ────────────────────────────────────────────────────────────
CREATE TABLE IF NOT EXISTS contributions (
    id              UUID                NOT NULL PRIMARY KEY DEFAULT uuid_generate_v4(),
    type            contribution_type   NOT NULL,
    contributor_id  UUID                NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    target_word_id  UUID                REFERENCES words(id) ON DELETE SET NULL,
    payload         JSONB               NOT NULL DEFAULT '{}',
    status          contribution_status NOT NULL DEFAULT 'pending',
    reviewer_id     UUID                REFERENCES users(id) ON DELETE SET NULL,
    reviewer_note   TEXT,
    submitted_at    TIMESTAMPTZ         NOT NULL DEFAULT NOW(),
    reviewed_at     TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS contributions_contributor_idx ON contributions (contributor_id);
CREATE INDEX IF NOT EXISTS contributions_status_idx      ON contributions (status);
CREATE INDEX IF NOT EXISTS contributions_submitted_idx   ON contributions (submitted_at DESC);

-- ─── votes ────────────────────────────────────────────────────────────────────
CREATE TABLE IF NOT EXISTS votes (
    id          UUID             NOT NULL PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id     UUID             NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    target_type vote_target_type NOT NULL,
    target_id   UUID             NOT NULL,
    value       vote_value       NOT NULL,
    created_at  TIMESTAMPTZ      NOT NULL DEFAULT NOW(),

    CONSTRAINT votes_unique UNIQUE (user_id, target_type, target_id)
);

CREATE INDEX IF NOT EXISTS votes_target_idx ON votes (target_type, target_id);

-- ─── bookmarks ────────────────────────────────────────────────────────────────
CREATE TABLE IF NOT EXISTS bookmarks (
    id         UUID        NOT NULL PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id    UUID        NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    word_id    UUID        NOT NULL REFERENCES words(id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT bookmarks_unique UNIQUE (user_id, word_id)
);

CREATE INDEX IF NOT EXISTS bookmarks_user_idx ON bookmarks (user_id);

-- ─── comments ─────────────────────────────────────────────────────────────────
CREATE TABLE IF NOT EXISTS comments (
    id          UUID                NOT NULL PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id     UUID                NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    target_type comment_target_type NOT NULL,
    target_id   UUID                NOT NULL,
    body        TEXT                NOT NULL,
    is_flagged  BOOLEAN             NOT NULL DEFAULT FALSE,
    created_at  TIMESTAMPTZ         NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ         NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS comments_target_idx  ON comments (target_type, target_id);
CREATE INDEX IF NOT EXISTS comments_user_idx    ON comments (user_id);
CREATE INDEX IF NOT EXISTS comments_flagged_idx ON comments (is_flagged) WHERE is_flagged = TRUE;
