-- name: CreateUser :one
INSERT INTO authentications (
    id, email, username, password_hash
    )
VALUES ($1, $2, $3, $4)
RETURNING *;

-- name: GetUserByIdentifier :one
SELECT * FROM authentications
WHERE email = $1 OR username = $1 OR phone = $1
LIMIT 1;

-- name: GetUserByID :one
SELECT * FROM authentications WHERE id = $1 LIMIT 1;

-- name: GetUserByUsername :one
SELECT id FROM authentications WHERE username = $1;

-- name: CheckUsername :one
SELECT COUNT(*) FROM authentications WHERE LOWER(username) = LOWER($1);

-- name: UpdateUser :one
UPDATE authentications
SET
    username = COALESCE(sqlc.narg('username'),username),
    email = COALESCE(sqlc.narg('email'),email),
    phone = COALESCE(sqlc.narg('phone'),phone),
    password_hash = COALESCE(sqlc.narg('password_hash'),password_hash),
    is_email_verified = COALESCE(sqlc.narg('is_email_verified'),is_email_verified),
    is_suspended = COALESCE(sqlc.narg('is_suspended'),is_suspended),
    is_deleted = COALESCE(sqlc.narg('is_deleted'),is_deleted),
    updated_at = COALESCE(sqlc.narg('updated_at'),updated_at),
    deleted_at = COALESCE(sqlc.narg('deleted_at'),deleted_at),
    verified_at = COALESCE(sqlc.narg('verified_at'),verified_at),
    suspended_at = COALESCE(sqlc.narg('suspended_at'),suspended_at),
    login_attempts = COALESCE(sqlc.narg('login_attempts'),login_attempts),
    lockout_duration = COALESCE(sqlc.narg('lockout_duration'),lockout_duration),
    lockout_until = COALESCE(sqlc.narg('lockout_until'),lockout_until),
    password_last_changed = COALESCE(sqlc.narg('password_last_changed'),password_last_changed),
    is_verified = COALESCE(sqlc.narg('is_verified'),is_verified),
    is_mfa_enabled = COALESCE(sqlc.narg('is_mfa_enabled'),is_mfa_enabled)
WHERE
    id = sqlc.arg('id')
RETURNING *;


-- name: UpdateUserPassword :exec
UPDATE authentications
SET password_hash = $2, password_last_changed = now()
WHERE id = $1 OR email = $1;

-- name: DeleteUserByID :exec
DELETE FROM authentications WHERE id = $1;

-- name: BlockUser :exec
UPDATE authentications
SET is_suspended = true,
    suspended_at = NOW()
WHERE id = $1;

-- name: GetUserAndRoleByIdentifier :one
SELECT authentications.*, user_roles.role_id, users.*
FROM authentications
JOIN user_roles ON authentications.id = user_roles.user_id
LEFT JOIN users ON authentications.id = users.user_id
WHERE authentications.username = $1 OR authentications.phone = $1 OR authentications.email = $1;

-- name: GetUserIDsFromUsernames :many
SELECT id FROM authentications WHERE username = ANY($1);

-- name: GetUidsFromUsername :many
SELECT id FROM authentications WHERE LOWER(username) = ANY($1::string[]);
