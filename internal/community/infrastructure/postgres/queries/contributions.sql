-- name: CreateContribution :one
INSERT INTO contributions (id, type, contributor_id, target_word_id, payload, status, submitted_at)
VALUES ($1, $2, $3, $4, $5, 'pending', NOW())
RETURNING *;

-- name: GetContributionByID :one
SELECT * FROM contributions WHERE id = $1 LIMIT 1;

-- name: ListContributionsByContributor :many
SELECT * FROM contributions
WHERE contributor_id = $1
  AND ($2::text = '' OR status::text = $2::text)
  AND ($3::text = '' OR type::text = $3::text)
ORDER BY submitted_at DESC
LIMIT $4 OFFSET $5;

-- name: CountContributionsByContributor :one
SELECT COUNT(*)::int FROM contributions
WHERE contributor_id = $1
  AND ($2::text = '' OR status::text = $2::text)
  AND ($3::text = '' OR type::text = $3::text);

-- name: ListAllContributions :many
SELECT * FROM contributions
WHERE ($1::text = '' OR status::text = $1::text)
  AND ($2::text = '' OR type::text = $2::text)
ORDER BY submitted_at DESC
LIMIT $3 OFFSET $4;

-- name: CountAllContributions :one
SELECT COUNT(*)::int FROM contributions
WHERE ($1::text = '' OR status::text = $1::text)
  AND ($2::text = '' OR type::text = $2::text);

-- name: UpdateContribution :one
UPDATE contributions
SET status        = $2,
    reviewer_id   = $3,
    reviewer_note = $4,
    reviewed_at   = $5
WHERE id = $1
RETURNING *;
