CREATE TYPE word_class  AS ENUM ('n', 'v', 'a', 'adv', 'p', 'pb', 'ki');
CREATE TYPE dialect     AS ENUM ('hulu');
CREATE TYPE word_source AS ENUM ('seeded', 'contributed', 'ai_generated');
CREATE TYPE word_status AS ENUM ('active', 'deprecated');

-- ─── words ────────────────────────────────────────────────────────────────────
CREATE TABLE IF NOT EXISTS words (
    id                  UUID        PRIMARY KEY DEFAULT uuid_generate_v4(),
    banjar              TEXT        NOT NULL,
    banjar_syllabified  TEXT,
    dialect             dialect     NOT NULL DEFAULT 'hulu',
    word_class          word_class  NOT NULL,
    homonym_number      SMALLINT    NOT NULL DEFAULT 1,
    is_root             BOOLEAN     NOT NULL DEFAULT TRUE,
    root_word_id        UUID        REFERENCES words(id) ON DELETE SET NULL,
    status              word_status NOT NULL DEFAULT 'active',
    source              word_source NOT NULL DEFAULT 'seeded',
    source_reference    TEXT,
    created_by          UUID        REFERENCES users(id) ON DELETE SET NULL,
    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at          TIMESTAMPTZ,

    CONSTRAINT words_unique UNIQUE NULLS NOT DISTINCT
        (banjar, dialect, homonym_number, is_root, root_word_id)
);

CREATE INDEX IF NOT EXISTS words_banjar_idx       ON words (banjar);
CREATE INDEX IF NOT EXISTS words_dialect_idx      ON words (dialect);
CREATE INDEX IF NOT EXISTS words_word_class_idx   ON words (word_class);
CREATE INDEX IF NOT EXISTS words_status_idx       ON words (status);
CREATE INDEX IF NOT EXISTS words_is_root_idx      ON words (is_root);
CREATE INDEX IF NOT EXISTS words_root_word_id_idx ON words (root_word_id);
CREATE INDEX IF NOT EXISTS words_created_at_idx   ON words (created_at DESC);
CREATE INDEX IF NOT EXISTS words_deleted_at_idx   ON words (deleted_at) WHERE deleted_at IS NULL;

-- Full-text search on banjar word
CREATE INDEX IF NOT EXISTS words_fts_idx ON words
    USING GIN (to_tsvector('simple', banjar));

-- ─── definitions ──────────────────────────────────────────────────────────────
CREATE TABLE IF NOT EXISTS definitions (
    id         UUID        PRIMARY KEY DEFAULT uuid_generate_v4(),
    word_id    UUID        NOT NULL REFERENCES words(id) ON DELETE CASCADE,
    meaning    TEXT        NOT NULL,
    sort_order SMALLINT    NOT NULL DEFAULT 1,
    source     word_source NOT NULL DEFAULT 'seeded',
    upvotes    INT         NOT NULL DEFAULT 0,
    downvotes  INT         NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS definitions_word_id_idx ON definitions (word_id);
CREATE INDEX IF NOT EXISTS definitions_sort_idx    ON definitions (word_id, sort_order ASC);

-- Full-text search on Indonesian meanings
CREATE INDEX IF NOT EXISTS definitions_fts_idx ON definitions
    USING GIN (to_tsvector('simple', meaning));

-- ─── examples ─────────────────────────────────────────────────────────────────
CREATE TABLE IF NOT EXISTS examples (
    id                     UUID        PRIMARY KEY DEFAULT uuid_generate_v4(),
    word_id                UUID        NOT NULL REFERENCES words(id) ON DELETE CASCADE,
    banjar_sentence        TEXT        NOT NULL,
    indonesian_translation TEXT        NOT NULL,
    source                 word_source NOT NULL DEFAULT 'seeded',
    created_at             TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at             TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS examples_word_id_idx ON examples (word_id);

-- ─── word_relations ───────────────────────────────────────────────────────────
CREATE TABLE IF NOT EXISTS word_relations (
    word_id         UUID NOT NULL REFERENCES words(id) ON DELETE CASCADE,
    related_word_id UUID NOT NULL REFERENCES words(id) ON DELETE CASCADE,
    PRIMARY KEY (word_id, related_word_id)
);
