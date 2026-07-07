ALTER TABLE plans
    DROP COLUMN IF EXISTS updated_at,
    DROP COLUMN IF EXISTS description,
    DROP COLUMN IF EXISTS sort_order,
    DROP COLUMN IF EXISTS is_highlighted,
    DROP COLUMN IF EXISTS is_active;
