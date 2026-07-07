ALTER TABLE apps
    ADD COLUMN allowed_origins TEXT[] NOT NULL DEFAULT '{}';
