-- name: InsertRefreshToken :exec
INSERT INTO refresh_tokens (id, value, issued_at, expires_at, user_id)
VALUES ($1, $2, $3, $4, $5);

-- name: GetRefreshTokenByValue :one
SELECT
    *
FROM refresh_tokens
WHERE value=$1
LIMIT 1;

-- name: Delete :exec
DELETE FROM refresh_tokens
WHERE id=$1;