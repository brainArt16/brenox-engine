ALTER TABLE apps
    ADD COLUMN sandbox_workspace_id BIGINT REFERENCES workspaces (id) ON DELETE RESTRICT;

DO $$
DECLARE
    r RECORD;
    ws_id BIGINT;
BEGIN
    FOR r IN
        SELECT id, name, slug, owner_id
        FROM apps
        WHERE sandbox_workspace_id IS NULL
    LOOP
        INSERT INTO workspaces (name, slug, owner_id)
        VALUES (r.name || ' Sandbox', 'app-' || r.slug || '-sandbox', r.owner_id)
        RETURNING id INTO ws_id;

        INSERT INTO workspace_members (workspace_id, user_id, role)
        VALUES (ws_id, r.owner_id, 'owner')
        ON CONFLICT DO NOTHING;

        UPDATE apps
        SET sandbox_workspace_id = ws_id
        WHERE id = r.id;
    END LOOP;
END $$;

ALTER TABLE apps
    ALTER COLUMN sandbox_workspace_id SET NOT NULL;

CREATE UNIQUE INDEX idx_apps_sandbox_workspace_id ON apps (sandbox_workspace_id);

ALTER TABLE app_users
    ADD COLUMN environment TEXT NOT NULL DEFAULT 'live';

ALTER TABLE app_users
    DROP CONSTRAINT IF EXISTS app_users_app_id_external_id_key;

ALTER TABLE app_users
    ADD CONSTRAINT app_users_app_id_external_id_environment_key
    UNIQUE (app_id, external_id, environment);
