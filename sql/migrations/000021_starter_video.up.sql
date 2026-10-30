-- Include video on free Starter with a modest monthly allotment.

UPDATE plans
SET
    video_calls_enabled = true,
    video_minutes_limit = 100,
    max_call_participants = 5
WHERE slug = 'starter';
