-- name: CreateAIRequest :exec
INSERT INTO ai_requests (
    id, type, target_word_id, target_contribution_id, requested_by,
    model, prompt, response, parsed_output, status, review_status,
    reviewed_by, reviewed_at, created_at
) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14);

-- name: GetAIRequestByID :one
SELECT id, type, target_word_id, target_contribution_id, requested_by,
       model, prompt, response, parsed_output, status, review_status,
       reviewed_by, reviewed_at, created_at
FROM ai_requests
WHERE id = $1;

-- name: UpdateAIRequest :exec
UPDATE ai_requests
SET response       = $2,
    parsed_output  = $3,
    status         = $4,
    review_status  = $5,
    reviewed_by    = $6,
    reviewed_at    = $7
WHERE id = $1;

-- name: ListAIRequestsByWord :many
SELECT id, type, target_word_id, target_contribution_id, requested_by,
       model, prompt, response, parsed_output, status, review_status,
       reviewed_by, reviewed_at, created_at
FROM ai_requests
WHERE target_word_id = $1
ORDER BY created_at DESC
LIMIT $2 OFFSET $3;

-- name: CountAIRequestsByWord :one
SELECT COUNT(*)::int FROM ai_requests WHERE target_word_id = $1;

-- name: ListAIRequestsPendingReview :many
SELECT id, type, target_word_id, target_contribution_id, requested_by,
       model, prompt, response, parsed_output, status, review_status,
       reviewed_by, reviewed_at, created_at
FROM ai_requests
WHERE status = 'completed' AND review_status = 'unreviewed'
ORDER BY created_at ASC
LIMIT $1 OFFSET $2;

-- name: CountAIRequestsPendingReview :one
SELECT COUNT(*)::int FROM ai_requests
WHERE status = 'completed' AND review_status = 'unreviewed';
