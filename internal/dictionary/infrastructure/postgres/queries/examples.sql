-- name: CreateExample :one
INSERT INTO examples (id, word_id, banjar_sentence, indonesian_translation, source, created_at, updated_at)
VALUES ($1, $2, $3, $4, $5, $6, $7)
RETURNING *;

-- name: GetExamplesByWordID :many
SELECT * FROM examples
WHERE word_id = $1
ORDER BY created_at ASC;

-- name: DeleteExamplesByWordID :exec
DELETE FROM examples WHERE word_id = $1;

-- name: UpsertExample :one
INSERT INTO examples (id, word_id, banjar_sentence, indonesian_translation, source, created_at, updated_at)
VALUES ($1, $2, $3, $4, $5, $6, $7)
ON CONFLICT (id) DO UPDATE
SET banjar_sentence        = EXCLUDED.banjar_sentence,
    indonesian_translation = EXCLUDED.indonesian_translation,
    updated_at             = EXCLUDED.updated_at
RETURNING *;
