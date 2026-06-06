CREATE TABLE calls (
    id BIGSERIAL PRIMARY KEY,
    channel_id BIGINT NOT NULL REFERENCES channels(id) ON DELETE CASCADE,
    workspace_id BIGINT NOT NULL REFERENCES workspaces(id) ON DELETE CASCADE,
    initiator_id BIGINT NOT NULL REFERENCES users(id),
    status TEXT NOT NULL CHECK (status IN ('ringing', 'active', 'ended')),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    ended_at TIMESTAMPTZ
);

CREATE TABLE call_participants (
    id BIGSERIAL PRIMARY KEY,
    call_id BIGINT NOT NULL REFERENCES calls(id) ON DELETE CASCADE,
    user_id BIGINT NOT NULL REFERENCES users(id),
    joined_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    left_at TIMESTAMPTZ
);

CREATE INDEX idx_calls_channel_active
    ON calls (channel_id)
    WHERE status IN ('ringing', 'active');

CREATE UNIQUE INDEX idx_call_participants_active
    ON call_participants (call_id, user_id)
    WHERE left_at IS NULL;

CREATE INDEX idx_call_participants_call_id
    ON call_participants (call_id);
