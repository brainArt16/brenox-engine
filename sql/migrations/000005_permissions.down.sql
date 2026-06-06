DROP TABLE IF EXISTS channel_roles;

ALTER TABLE channels
    DROP COLUMN IF EXISTS is_read_only;

ALTER TABLE workspace_members
    DROP CONSTRAINT IF EXISTS workspace_members_role_check;
