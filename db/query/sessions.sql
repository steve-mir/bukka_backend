-- Create a new session
-- name: CreateSession :one
INSERT INTO sessions (id, user_id, refresh_token, refresh_token_exp,
user_agent, updated_at, ip_address, blocked_at, invalidated_at, last_active_at, fcm_token)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11) RETURNING *;

-- name: GetSessionsByUserID :many
SELECT * FROM sessions
WHERE user_id = $1;

-- name: GetSessionsByRefreshToken :one
SELECT * FROM sessions
WHERE refresh_token = $1;

-- name: GetSessionAndUserByRefreshToken :one
SELECT s.*, u.username, u.email, u.phone, u.is_email_verified, ur.role_id
FROM sessions s
JOIN authentications u ON s.user_id = u.id
LEFT JOIN user_roles ur ON u.id = ur.user_id
WHERE s.refresh_token = $1;


-- name: GetSessionsByID :one
SELECT * FROM sessions WHERE id = $1 LIMIT 1;

-- name: RotateSessionTokens :exec
UPDATE "sessions"
SET
  "refresh_token" = $2,
  "refresh_token_exp" = $3,
  "updated_at" = now(),
  "last_active_at" = now()
WHERE "id" = $1;

-- name: RevokeSessionById :exec
UPDATE sessions SET invalidated_at = now() WHERE user_id = $1;

-- name: BlockAllUserSession :exec
UPDATE sessions SET blocked_at = now(), invalidated_at = now() WHERE user_id = $1;

-- name: DeleteSession :exec
DELETE FROM sessions
WHERE id = $1;