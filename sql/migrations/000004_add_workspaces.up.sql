CREATE TABLE workspaces (
    id BIGSERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    slug VARCHAR(255) NOT NULL UNIQUE,
    owner_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE TABLE workspace_members (
    id BIGSERIAL PRIMARY KEY,
    workspace_id BIGINT NOT NULL REFERENCES workspaces(id) ON DELETE CASCADE,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    role VARCHAR(50) NOT NULL DEFAULT 'member',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE (workspace_id, user_id)
);

ALTER TABLE channels
    ADD COLUMN workspace_id BIGINT REFERENCES workspaces(id);

-- Backfill: one default workspace per channel owner.
INSERT INTO workspaces (name, slug, owner_id)
SELECT
    'Default Workspace',
    'default-' || owner_id::text,
    owner_id
FROM (
    SELECT DISTINCT owner_id
    FROM channels
) AS owners;

INSERT INTO workspace_members (workspace_id, user_id, role)
SELECT w.id, w.owner_id, 'owner'
FROM workspaces w
WHERE w.slug LIKE 'default-%';

UPDATE channels c
SET workspace_id = w.id
FROM workspaces w
WHERE w.owner_id = c.owner_id;

ALTER TABLE channels
    ALTER COLUMN workspace_id SET NOT NULL;

CREATE UNIQUE INDEX channels_workspace_name_unique
    ON channels (workspace_id, name);
