DROP INDEX IF EXISTS channels_workspace_name_unique;

ALTER TABLE channels
    DROP COLUMN IF EXISTS workspace_id;

DROP TABLE IF EXISTS workspace_members;
DROP TABLE IF EXISTS workspaces;
