-- name: GetDefinitionByID :one
SELECT * FROM definitions WHERE id = $1 LIMIT 1;

-- name: CreateDefinition :one
INSERT INTO definitions (id, word_id, meaning, sort_order, source, upvotes, downvotes, created_at, updated_at)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
RETURNING *;

-- name: GetDefinitionsByWordID :many
SELECT * FROM definitions
WHERE word_id = $1
ORDER BY sort_order ASC, (upvotes - downvotes) DESC;

-- name: DeleteDefinitionsByWordID :exec
DELETE FROM definitions WHERE word_id = $1;

-- name: UpsertDefinition :one
INSERT INTO definitions (id, word_id, meaning, sort_order, source, upvotes, downvotes, created_at, updated_at)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
ON CONFLICT (id) DO UPDATE
SET meaning    = EXCLUDED.meaning,
    sort_order = EXCLUDED.sort_order,
    updated_at = EXCLUDED.updated_at
RETURNING *;
