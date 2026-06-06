ALTER TABLE calls
    ADD COLUMN mode TEXT NOT NULL DEFAULT 'voice' CHECK (mode IN ('voice', 'video'));

CREATE TABLE call_recordings (
    id BIGSERIAL PRIMARY KEY,
    call_id BIGINT NOT NULL REFERENCES calls(id) ON DELETE CASCADE,
    started_by BIGINT NOT NULL REFERENCES users(id),
    started_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    ended_at TIMESTAMPTZ,
    metadata JSONB NOT NULL DEFAULT '{}'
);

CREATE UNIQUE INDEX idx_call_recordings_active
    ON call_recordings (call_id)
    WHERE ended_at IS NULL;

CREATE INDEX idx_call_recordings_call_id
    ON call_recordings (call_id);
