DROP TABLE IF EXISTS usage_counters;
DROP TABLE IF EXISTS app_subscriptions;
DROP TABLE IF EXISTS platform_settings;
DROP TABLE IF EXISTS plans;

ALTER TABLE users DROP COLUMN IF EXISTS stripe_customer_id;
