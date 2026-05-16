-- name: CreateBookmark :one
INSERT INTO bookmarks (id, user_id, word_id, created_at)
VALUES ($1, $2, $3, NOW())
RETURNING *;

-- name: DeleteBookmark :exec
DELETE FROM bookmarks WHERE user_id = $1 AND word_id = $2;

-- name: GetBookmarkByUserAndWord :one
SELECT * FROM bookmarks WHERE user_id = $1 AND word_id = $2 LIMIT 1;

-- name: ListBookmarksByUser :many
SELECT * FROM bookmarks
WHERE user_id = $1
ORDER BY created_at DESC
LIMIT $2 OFFSET $3;

-- name: CountBookmarksByUser :one
SELECT COUNT(*)::int FROM bookmarks WHERE user_id = $1;
