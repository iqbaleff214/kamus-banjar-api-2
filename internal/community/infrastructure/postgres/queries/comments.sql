-- name: CreateComment :one
INSERT INTO comments (id, user_id, target_type, target_id, body, is_flagged, created_at, updated_at)
VALUES ($1, $2, $3, $4, $5, FALSE, NOW(), NOW())
RETURNING *;

-- name: GetCommentByID :one
SELECT * FROM comments WHERE id = $1 LIMIT 1;

-- name: ListCommentsByTarget :many
SELECT * FROM comments
WHERE target_type = $1 AND target_id = $2
ORDER BY created_at ASC
LIMIT $3 OFFSET $4;

-- name: CountCommentsByTarget :one
SELECT COUNT(*)::int FROM comments WHERE target_type = $1 AND target_id = $2;

-- name: UpdateComment :one
UPDATE comments
SET body       = $2,
    updated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: FlagComment :one
UPDATE comments
SET is_flagged = TRUE,
    updated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: DeleteComment :exec
DELETE FROM comments WHERE id = $1;

-- name: ListFlaggedComments :many
SELECT * FROM comments
WHERE is_flagged = TRUE
ORDER BY created_at ASC
LIMIT $1 OFFSET $2;

-- name: CountFlaggedComments :one
SELECT COUNT(*)::int FROM comments WHERE is_flagged = TRUE;
