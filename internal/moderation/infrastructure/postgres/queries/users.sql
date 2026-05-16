-- name: ListUsersAdmin :many
SELECT id, name, email, password_hash, role, is_active, email_verified_at, created_at, updated_at
FROM users
WHERE
    ($1::text = '' OR role = $1::text)
    AND ($2::text = '' OR is_active::text = $2::text)
    AND ($3::text = '' OR name ILIKE '%' || $3 || '%' OR email ILIKE '%' || $3 || '%')
ORDER BY created_at DESC
LIMIT $4 OFFSET $5;

-- name: CountUsersAdmin :one
SELECT COUNT(*)::int FROM users
WHERE
    ($1::text = '' OR role = $1::text)
    AND ($2::text = '' OR is_active::text = $2::text)
    AND ($3::text = '' OR name ILIKE '%' || $3 || '%' OR email ILIKE '%' || $3 || '%');
