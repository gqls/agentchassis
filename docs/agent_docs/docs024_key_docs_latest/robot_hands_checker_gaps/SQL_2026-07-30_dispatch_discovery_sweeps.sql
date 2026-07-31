-- Dispatch the three discovery lanes at robot-hands.com.
--
-- WHY THIS FILE EXISTS: measured 2026-07-30, NONE of the 24 enabled
-- scheduled_tasks targets completeness-, design- or quality-discovery-agent.
-- The checker layer runs only when a thread fires it by hand. robot-hands'
-- design lane had not swept since 2026-07-24, which is why the MatchMatrix
-- card image (detected correctly by content_image_missing, pass 1 complete
-- 07-25) never reached pass 2 to derive its card asset.
--
-- LANDMINE (cost a prior thread a silent no-op): scheduled_tasks.target_topic
-- DEFAULTS to 'system.agent.generic.requests', which NOTHING CONSUMES. The one
-- discovery task ever created (oneshot-discovery-aao-20260726) carries that
-- default and never ran anything. All 18 working enabled tasks use
-- 'system.agent.scheduled.requests'. Name the topic explicitly, always.
--
-- ONESHOT SEMANTICS: the scheduler selects on `last_triggered_at IS NULL`
-- (cmd/scheduler/main.go:354-374), so an enabled row with a NULL timestamp
-- fires on the next tick. Disable each row once last_triggered_at is stamped —
-- otherwise it re-fires every interval_seconds.
--
-- All three agents start at `ensure_site_record`, whose input_fields are
-- [site_id, domain], so input_data carries exactly those two keys.

\set sid '00ff3af5-dad8-4770-9f70-3edc267a3c92'

INSERT INTO scheduled_tasks
  (name, description, target_agent_type, target_topic,
   interval_seconds, timeout_seconds, enabled, fire_message, input_data)
VALUES
  ('oneshot-design-discovery-rh-20260730',
   'One-shot design discovery sweep of robot-hands.com (missing MatchMatrix card asset). Disable once fired.',
   'design-discovery-agent', 'system.agent.scheduled.requests',
   86400, 900, true, true,
   jsonb_build_object('site_id', :'sid', 'domain', 'robot-hands.com'))
ON CONFLICT (name) DO NOTHING;
