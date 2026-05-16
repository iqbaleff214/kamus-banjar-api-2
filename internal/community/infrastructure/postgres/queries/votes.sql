-- name: UpsertVote :one
INSERT INTO votes (id, user_id, target_type, target_id, value, created_at)
VALUES ($1, $2, $3, $4, $5, NOW())
ON CONFLICT (user_id, target_type, target_id) DO UPDATE
    SET value = EXCLUDED.value
RETURNING *;

-- name: GetVoteByUserAndTarget :one
SELECT * FROM votes
WHERE user_id = $1 AND target_type = $2 AND target_id = $3
LIMIT 1;

-- name: DeleteVote :exec
DELETE FROM votes
WHERE user_id = $1 AND target_type = $2 AND target_id = $3;
