-- name: InsertRefreshToken :exec
INSERT INTO refresh_tokens (id, value, issued_at, expires_at, user_id)
VALUES ($1, $2, $3, $4, $5);