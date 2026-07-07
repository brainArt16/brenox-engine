ALTER TABLE plans
    ADD COLUMN is_active BOOLEAN NOT NULL DEFAULT true,
    ADD COLUMN is_highlighted BOOLEAN NOT NULL DEFAULT false,
    ADD COLUMN sort_order INT NOT NULL DEFAULT 0,
    ADD COLUMN description TEXT NOT NULL DEFAULT '',
    ADD COLUMN updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW();

UPDATE plans SET sort_order = 1, is_highlighted = false WHERE slug = 'starter';
UPDATE plans SET sort_order = 2, is_highlighted = true WHERE slug = 'growth';
UPDATE plans SET sort_order = 3, is_highlighted = false WHERE slug = 'scale';

UPDATE plans SET description = 'For early-stage apps getting realtime chat live' WHERE slug = 'starter';
UPDATE plans SET description = 'For production apps with growing user bases' WHERE slug = 'growth';
UPDATE plans SET description = 'For teams that need higher limits and reliability' WHERE slug = 'scale';
