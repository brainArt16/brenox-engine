UPDATE plans
SET
    video_calls_enabled = false,
    video_minutes_limit = 0,
    max_call_participants = 5
WHERE slug = 'starter';
