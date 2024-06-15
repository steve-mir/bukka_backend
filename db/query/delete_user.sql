-- name: CreateUserDeleteRequest :exec
INSERT INTO account_recovery_requests (
    user_id, email, recovery_token, expires_at, completed_at
    )
VALUES ($1, $2, $3, $4, $5);

-- name: GetUserFromDeleteReqByToken :one
SELECT * FROM account_recovery_requests WHERE recovery_token = $1 LIMIT 1;

-- name: MarkDeleteAsUsedByToken :exec
UPDATE account_recovery_requests SET used = true, completed_at = now() WHERE recovery_token = $1;