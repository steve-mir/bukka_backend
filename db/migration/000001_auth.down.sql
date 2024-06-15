ALTER TABLE "user_roles" DROP CONSTRAINT IF EXISTS "user_roles_user_id_fkey";
ALTER TABLE "user_roles" DROP CONSTRAINT IF EXISTS "user_roles_role_id_fkey";

DROP INDEX IF EXISTS "authentications_email_phone_username_created_at_updated_at_idx";

DROP INDEX IF EXISTS "user_id_email_phone_username_created_at_updated_at_idx";

DROP INDEX IF EXISTS "user_roles_role_id_idx";

DROP INDEX IF EXISTS "user_profiles_user_id_following_count_follower_count_idx";

DROP INDEX IF EXISTS "users_email_username_created_at_updated_at";

DROP TABLE IF EXISTS "roles";
DROP TABLE IF EXISTS "user_roles";
DROP TABLE IF EXISTS "authentications";

-- TODO: DROP users table and its constraints
