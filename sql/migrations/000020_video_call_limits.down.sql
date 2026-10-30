ALTER TABLE usage_counters DROP COLUMN IF EXISTS video_minutes;

ALTER TABLE plans
    DROP COLUMN IF EXISTS video_minutes_limit,
    DROP COLUMN IF EXISTS max_call_participants;
