ALTER TABLE users
    ADD COLUMN platform_role VARCHAR(20) NOT NULL DEFAULT 'user'
        CHECK (platform_role IN ('user', 'support', 'admin'));

ALTER TABLE users
    ADD COLUMN suspended_at TIMESTAMPTZ;

ALTER TABLE users
    ADD COLUMN tokens_invalidated_at TIMESTAMPTZ;

CREATE INDEX idx_users_platform_role ON users (platform_role);
CREATE INDEX idx_users_suspended_at ON users (suspended_at) WHERE suspended_at IS NOT NULL;
