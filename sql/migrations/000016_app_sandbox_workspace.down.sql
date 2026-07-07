ALTER TABLE app_users
    DROP CONSTRAINT IF EXISTS app_users_app_id_external_id_environment_key;

ALTER TABLE app_users
    ADD CONSTRAINT app_users_app_id_external_id_key UNIQUE (app_id, external_id);

ALTER TABLE app_users
    DROP COLUMN IF EXISTS environment;

DROP INDEX IF EXISTS idx_apps_sandbox_workspace_id;

ALTER TABLE apps
    DROP COLUMN IF EXISTS sandbox_workspace_id;
