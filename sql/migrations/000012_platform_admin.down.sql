DROP INDEX IF EXISTS idx_users_suspended_at;
DROP INDEX IF EXISTS idx_users_platform_role;

ALTER TABLE users DROP COLUMN IF EXISTS tokens_invalidated_at;
ALTER TABLE users DROP COLUMN IF EXISTS suspended_at;
ALTER TABLE users DROP COLUMN IF EXISTS platform_role;
