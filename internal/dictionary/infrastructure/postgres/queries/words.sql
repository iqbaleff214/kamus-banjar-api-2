-- name: CreateWord :one
INSERT INTO words (
    id, banjar, banjar_syllabified, dialect, word_class,
    homonym_number, is_root, root_word_id,
    status, source, source_reference, created_by,
    created_at, updated_at
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14
)
RETURNING *;

-- name: GetWordByID :one
SELECT * FROM words
WHERE id = $1 AND deleted_at IS NULL
LIMIT 1;

-- name: GetWordByIDIncludeDeleted :one
SELECT * FROM words
WHERE id = $1
LIMIT 1;

-- name: UpdateWord :one
UPDATE words
SET banjar              = $2,
    banjar_syllabified  = $3,
    word_class          = $4,
    homonym_number      = $5,
    is_root             = $6,
    root_word_id        = $7,
    status              = $8,
    source              = $9,
    source_reference    = $10,
    updated_at          = $11
WHERE id = $1 AND deleted_at IS NULL
RETURNING *;

-- name: SoftDeleteWord :exec
UPDATE words
SET deleted_at = NOW(), updated_at = NOW()
WHERE id = $1 AND deleted_at IS NULL;

-- ListWords: $1=word_class filter ('' = no filter), $2=is_root filter ('' | 'true' | 'false'),
--            $3=source filter, $4=status filter, $5=sort ('recently_added' | else alphabetical),
--            $6=limit, $7=offset
-- name: ListWords :many
SELECT * FROM words
WHERE deleted_at IS NULL
  AND ($1::text = '' OR word_class::text = $1::text)
  AND ($2::text = '' OR is_root = ($2::text = 'true'))
  AND ($3::text = '' OR source::text = $3::text)
  AND ($4::text = '' OR status::text = $4::text)
ORDER BY
    CASE WHEN $5::text = 'recently_added' THEN EXTRACT(EPOCH FROM created_at) END DESC NULLS LAST,
    banjar ASC
LIMIT $6 OFFSET $7;

-- name: CountWords :one
SELECT COUNT(*)::int FROM words
WHERE deleted_at IS NULL
  AND ($1::text = '' OR word_class::text = $1::text)
  AND ($2::text = '' OR is_root = ($2::text = 'true'))
  AND ($3::text = '' OR source::text = $3::text)
  AND ($4::text = '' OR status::text = $4::text);

-- name: SearchWords :many
SELECT DISTINCT w.* FROM words w
LEFT JOIN definitions d ON d.word_id = w.id
WHERE w.deleted_at IS NULL
  AND (
        to_tsvector('simple', w.banjar)  @@ plainto_tsquery('simple', $1::text)
     OR to_tsvector('simple', d.meaning) @@ plainto_tsquery('simple', $1::text)
     OR w.banjar ILIKE '%' || $1::text || '%'
  )
  AND ($2::text = '' OR w.word_class::text = $2::text)
  AND ($3::text = '' OR w.is_root = ($3::text = 'true'))
  AND ($4::text = '' OR w.source::text = $4::text)
  AND ($5::text = '' OR w.status::text = $5::text)
ORDER BY w.banjar ASC
LIMIT $6 OFFSET $7;

-- name: CountSearchWords :one
SELECT COUNT(DISTINCT w.id)::int FROM words w
LEFT JOIN definitions d ON d.word_id = w.id
WHERE w.deleted_at IS NULL
  AND (
        to_tsvector('simple', w.banjar)  @@ plainto_tsquery('simple', $1::text)
     OR to_tsvector('simple', d.meaning) @@ plainto_tsquery('simple', $1::text)
     OR w.banjar ILIKE '%' || $1::text || '%'
  )
  AND ($2::text = '' OR w.word_class::text = $2::text)
  AND ($3::text = '' OR w.is_root = ($3::text = 'true'))
  AND ($4::text = '' OR w.source::text = $4::text)
  AND ($5::text = '' OR w.status::text = $5::text);

-- name: UpsertWord :one
INSERT INTO words (
    id, banjar, banjar_syllabified, dialect, word_class,
    homonym_number, is_root, root_word_id,
    status, source, source_reference, created_by,
    created_at, updated_at
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14
)
ON CONFLICT (banjar, dialect, homonym_number, is_root, root_word_id) DO UPDATE
SET word_class       = EXCLUDED.word_class,
    source_reference = EXCLUDED.source_reference,
    updated_at       = EXCLUDED.updated_at
RETURNING *;
