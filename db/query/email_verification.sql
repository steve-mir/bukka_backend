-- name: CreateEmailVerificationRequest :exec
INSERT INTO email_verification_requests (user_id, email, token, expires_at)
VALUES ($1, $2, $3, $4);

-- name: GetEmailVerificationRequestByToken :one
SELECT * FROM email_verification_requests WHERE token = $1 LIMIT 1;

-- name: UpdateEmailVerificationRequest :one
UPDATE email_verification_requests SET is_verified = $1 WHERE token = $2
RETURNING *;

-- name: CleanupVerifiedAndExpiredRequests :exec
DELETE FROM email_verification_requests WHERE is_verified = true OR expires_at < now();
