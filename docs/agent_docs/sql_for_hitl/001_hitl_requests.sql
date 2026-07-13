-- Queries for finding and managing HITL requests
-- These work with the existing awaited_requests table

-- Find the pending HITL request for escalate_to_human
-- From the earlier output, we have:
--   request_id: f9375885-5a1f-4bec-b337-4d241d761aa1
--   reply_to_topic: job.d6bc7920-4d7eef75-content-reviewer-spawn_reviewer.responses
--   correlation_id: d6bc7920-6fad-47d4-a4ac-2777d193432a

-- Find all HITL-related awaited requests (by step name pattern)
SELECT
    request_id,
    orchestration_id,
    correlation_id,
    step_name,
    target_agent_type,
    responses_topic as reply_to_topic,
    status,
    timeout_at,
    timeout_at - NOW() as time_remaining
FROM awaited_requests
WHERE step_name LIKE '%human%'
   OR step_name LIKE '%hitl%'
   OR step_name LIKE '%approval%'
   OR step_name LIKE '%escalate%'
   OR step_name LIKE '%review%'
ORDER BY created_at DESC
    LIMIT 20;

-- Find pending requests that haven't expired
SELECT
    request_id,
    orchestration_id,
    correlation_id,
    step_name,
    responses_topic as reply_to_topic,
    status,
    timeout_at,
    EXTRACT(EPOCH FROM (timeout_at - NOW())) / 60 as minutes_remaining
FROM awaited_requests
WHERE status = 'waiting'
  AND timeout_at > NOW()
ORDER BY timeout_at ASC;

-- Get details for a specific correlation_id
SELECT
    request_id,
    orchestration_id,
    step_name,
    target_agent_type,
    responses_topic as reply_to_topic,
    requests_topic,
    status,
    sent_at,
    timeout_at,
    processed_at,
    claimer_pod_name
FROM awaited_requests
WHERE correlation_id = 'd6bc7920-6fad-47d4-a4ac-2777d193432a'
ORDER BY sent_at DESC;

-- Reset an expired request to allow re-processing (use with caution)
-- UPDATE awaited_requests
-- SET status = 'waiting',
--     timeout_at = NOW() + INTERVAL '1 hour'
-- WHERE request_id = 'f9375885-5a1f-4bec-b337-4d241d761aa1';

-- Check orchestration state for the content-reviewer
SELECT
    orchestration_id,
    orchestration_name,
    status,
    current_step,
    parent_orchestration_id,
    updated_at
FROM orchestration_state
WHERE orchestration_id = '4f60b9e2-28ab-482d-a7bf-3c098311661f';  -- content-reviewer

-- Get the collected data (to see what data is available for HITL)
SELECT
    orchestration_id,
    current_step,
    collected_data
FROM orchestration_state
WHERE orchestration_id = '4f60b9e2-28ab-482d-a7bf-3c098311661f';

-- ============================================================
-- TO MANUALLY CONTINUE THE HITL FLOW:
-- ============================================================
--
-- From the awaited_requests row:
--   request_id: f9375885-5a1f-4bec-b337-4d241d761aa1
--   reply_to_topic: job.d6bc7920-4d7eef75-content-reviewer-spawn_reviewer.responses
--   correlation_id: d6bc7920-6fad-47d4-a4ac-2777d193432a
--
-- Send this to Kafka (replace values from your actual request):
--
-- ./hitl_respond.sh \
--     "f9375885-5a1f-4bec-b337-4d241d761aa1" \
--     "job.d6bc7920-4d7eef75-content-reviewer-spawn_reviewer.responses" \
--     "d6bc7920-6fad-47d4-a4ac-2777d193432a" \
--     "true"
--
-- ============================================================

-- Check if the topic exists in Kafka (run in Kafka pod)
-- /opt/kafka/bin/kafka-topics.sh --bootstrap-server localhost:9092 --list | grep "job.d6bc7920"

-- Consume from the notification topic to see what was sent
-- /opt/kafka/bin/kafka-console-consumer.sh \
--     --bootstrap-server localhost:9092 \
--     --topic system.notifications.ui \
--     --from-beginning \
--     --max-messages 10