ALTER TABLE "password_reset_requests" DROP CONSTRAINT IF EXISTS "password_reset_requests_user_id_fkey";
ALTER TABLE "two_factor_secrets" DROP CONSTRAINT IF EXISTS "two_factor_secrets_user_id_fkey";
ALTER TABLE "two_factor_revocation" DROP CONSTRAINT IF EXISTS "two_factor_revocation_user_id_fkey";
ALTER TABLE "two_factor_backup_codes" DROP CONSTRAINT IF EXISTS "two_factor_backup_codes_user_id_fkey";

ALTER TABLE "account_recovery_requests" DROP CONSTRAINT IF EXISTS "account_recovery_requests_user_id_fkey";


-- Drop indexes
DROP INDEX IF EXISTS "password_reset_requests_email_user_id_token_expires_at_idx";
DROP INDEX IF EXISTS "account_recovery_requests_user_id_idx";


-- Drop tables
DROP TABLE IF EXISTS "two_factor_backup_codes";
DROP TABLE IF EXISTS "two_factor_revocation";
DROP TABLE IF EXISTS "two_factor_secrets";
DROP TABLE IF EXISTS "password_reset_requests";

DROP TABLE IF EXISTS "account_recovery_requests";