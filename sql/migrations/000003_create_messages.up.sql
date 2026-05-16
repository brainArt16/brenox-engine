CREATE TABLE messages (

    id BIGSERIAL PRIMARY KEY,

    channel_id BIGINT NOT NULL REFERENCES channels(id),

    sender_id BIGINT NOT NULL REFERENCES users(id),

    content TEXT NOT NULL,

    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);