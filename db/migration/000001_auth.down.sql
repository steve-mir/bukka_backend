-- Remove foreign key constraints from user_roles
ALTER TABLE "user_roles" DROP CONSTRAINT IF EXISTS user_roles_user_id_fkey;
ALTER TABLE "user_roles" DROP CONSTRAINT IF EXISTS user_roles_role_id_fkey;

-- Drop indexes created on tables
DROP INDEX IF EXISTS authentications_email_phone_username_created_at_updated_at_idx;
DROP INDEX IF EXISTS user_roles_role_id_user_id_idx;

-- Drop the tables
DROP TABLE IF EXISTS "user_roles";
DROP TABLE IF EXISTS "roles";
DROP TABLE IF EXISTS "users";
DROP TABLE IF EXISTS "authentications";