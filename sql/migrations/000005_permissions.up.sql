-- Role constraint and channel permission primitives.

ALTER TABLE workspace_members
    ADD CONSTRAINT workspace_members_role_check
    CHECK (role IN ('owner', 'admin', 'moderator', 'member'));

ALTER TABLE channels
    ADD COLUMN is_read_only BOOLEAN NOT NULL DEFAULT false;

CREATE TABLE channel_roles (
    id BIGSERIAL PRIMARY KEY,
    channel_id BIGINT NOT NULL REFERENCES channels(id) ON DELETE CASCADE,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    role VARCHAR(50) NOT NULL CHECK (role IN ('moderator', 'member')),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (channel_id, user_id)
);
