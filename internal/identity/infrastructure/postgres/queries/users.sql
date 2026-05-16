-- name: CreateUser :one
INSERT INTO users (id, name, email, password_hash, role, is_active, email_verified_at, created_at, updated_at)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
RETURNING *;

-- name: GetUserByID :one
SELECT * FROM users
WHERE id = $1
LIMIT 1;

-- name: GetUserByEmail :one
SELECT * FROM users
WHERE email = $1
LIMIT 1;

-- name: UpdateUser :one
UPDATE users
SET name              = $2,
    password_hash     = $3,
    role              = $4,
    is_active         = $5,
    email_verified_at = $6,
    updated_at        = $7
WHERE id = $1
RETURNING *;
