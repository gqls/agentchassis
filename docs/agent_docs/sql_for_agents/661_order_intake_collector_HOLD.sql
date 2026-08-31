-- 661 (_HOLD; born 660, renamed same evening: the 394 lane's 660_render_audit_coverage_cursor took the number first and is already applied): the order-intake collector — P4's scheduled half (owner GO
-- 2026-08-26, PLAN_2026-07-31_p4_order_intake).
--
-- WHAT IT DOES: every 15 minutes (owner-ruled interval, 2026-07-31 §7.1) the
-- 'order-intake-collect' task dispatches one 'order-intake-collector' run,
-- whose single step is collect_external_orders: poll the box's committed-brief
-- list over the public edge, release each brief with a PAID
-- billing_orders.external_reference into build_queue (priority 10), file
-- needs_human_review for what a machine must not decide (repeat domain, no
-- domain — the 2026-07-31 §7.2 ruling), acknowledge what was taken.
--
-- ⚠ _HOLD, image-before-seeds (652's own reasoning): this file names
-- `collect_external_orders`, which exists only from the roll carrying commit
-- c32a5121a's successor on platform/orchestration/actions. Apply BY HAND after:
--   kubectl -n ai-persona-system logs -l app=agent-chassis --tail=300 | grep -m1 'build provenance'
--   git merge-base --is-ancestor <the commit adding collect_external_orders> <the stamp>
--
-- ⚠ SHIPS DISABLED (enabled=false), and the verify block ASSERTS disabled.
-- Two things are owed before anyone flips it on:
--   1. ~~P5 seeding — seed_build_queue does not yet write the customer's
--      contact details into `sites` or seed an evidence_base aspect~~
--      BUILT 2026-08-31 (`seedCustomerIdentity` in seed_build_queue_action.go:
--      sites.email/company_name + a two-fact evidence_base register from the
--      direction's customer fields, mutation-proven; existing values and an
--      existing register always win). ⚠ Go is INERT UNTIL A ROLL — before
--      flipping the task on, ask the RUNNING BINARY, not git (a same-tag
--      rebuild ships the cached image and no ancestry check can see it;
--      council 7e3dd082, debug_historian). Per SERVICE, with a control:
--        POD=$(kubectl -n ai-persona-system get pods -l app=agent-chassis --no-headers -o custom-columns=:metadata.name | head -1)
--        kubectl -n ai-persona-system exec "$POD" -- grep -aq "seedCustomerIdentity" /proc/1/exe && echo P5-IN-POD || echo P5-ABSENT
--        kubectl -n ai-persona-system exec "$POD" -- grep -aq "seedCustomerIdentityZZZ" /proc/1/exe && echo CONTROL-BROKEN || echo control-ok
--   2. The chassis env needs WEBDESIGN_BOX_ORDERS_TOKEN (terraform
--      047-base-configs; the collector fails loudly without it).
-- Enable with:
--   UPDATE scheduled_tasks SET enabled = true, updated_at = now()
--    WHERE name = 'order-intake-collect';

BEGIN;

INSERT INTO agent_definitions (type, display_name, description, category, agent_category, status, is_active, default_config)
SELECT
  'order-intake-collector',
  'Order Intake Collector',
  'Heartbeat agent: polls the webdesign.uk box''s committed-brief list (chat-submitted briefs, BR- references), releases paid briefs into build_queue by billing_orders.external_reference, files needs_human_review for repeat-domain and no-domain paid orders, acknowledges collection back to the box. One batch (max 10) per invocation. Dispatched by the order-intake-collect scheduled task. Owner rulings: reference-not-brief join (2026-08-26); 15-min poll and reject-repeat-domains-to-a-human (2026-07-31).',
  'orchestrator',
  -- agent_category: check_ad_category admits only strategist/executor/analyst/
  -- integrator/coordinator/specialist — 'orchestrator' here aborted the first
  -- apply (2026-08-31); live orchestrators pair category='orchestrator' with
  -- agent_category='coordinator'.
  'coordinator',
  'active',
  true,
  jsonb_build_object('workflow', jsonb_build_object(
    'start_step', 'collect',
    'processing_mode', 'orchestrator',
    'timeout_seconds', 300,
    'steps', jsonb_build_object(
      'collect', jsonb_build_object(
        'action', 'collect_external_orders',
        'config', jsonb_build_object(
          'orders_url', 'https://preview.webdesign.uk/internal/orders',
          'max_orders', 10
        ),
        'next_step', 'complete',
        'description', 'Poll the box, release paid briefs to build_queue, ack what was taken',
        'output_field', 'collect_result'
      ),
      'complete', jsonb_build_object(
        'action', 'complete_workflow',
        'config', jsonb_build_object('output_fields', jsonb_build_array('collect_result')),
        'description', 'Report listed/queued/awaiting_payment/human_review/acked counts'
      )
    )
  ))
WHERE NOT EXISTS (
  SELECT 1 FROM agent_definitions WHERE type = 'order-intake-collector' AND deleted_at IS NULL
);

INSERT INTO scheduled_tasks (name, description, interval_seconds, target_agent_type, target_topic, concurrency_group, max_concurrent, enabled)
SELECT
  'order-intake-collect',
  'Collect paid customer briefs from the webdesign.uk box into build_queue (P4). 15-min interval per the 2026-07-31 owner ruling. SHIPS DISABLED: enable only after P5 seeding lands and WEBDESIGN_BOX_ORDERS_TOKEN is in the chassis env — see migration 661''s header.',
  900,
  'order-intake-collector',
  'system.agent.generic.requests',
  'order-intake-collect',
  1,
  false
WHERE NOT EXISTS (SELECT 1 FROM scheduled_tasks WHERE name = 'order-intake-collect');

-- Verify: agent active; schedule present, DISABLED (the contract of this
-- file — a 661 that arrives enabled is a defect, not a convenience), pointed
-- at the agent; and the workflow still names the action and the box URL.
DO $$
DECLARE n int; en boolean; cfg jsonb;
BEGIN
  SELECT count(*) INTO n FROM agent_definitions
   WHERE type = 'order-intake-collector' AND is_active AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;
  IF n <> 1 THEN RAISE EXCEPTION '661 verify failed: order-intake-collector not active (found %)', n; END IF;

  SELECT enabled INTO en FROM scheduled_tasks
   WHERE name = 'order-intake-collect' AND target_agent_type = 'order-intake-collector';
  IF en IS NULL THEN RAISE EXCEPTION '661 verify failed: schedule missing or mistargeted'; END IF;
  IF en THEN RAISE EXCEPTION '661 verify failed: order-intake-collect arrived ENABLED — this file ships it disabled (P5 seeding + token are owed first)'; END IF;

  SELECT default_config INTO cfg FROM agent_definitions
   WHERE type = 'order-intake-collector' AND is_active AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;
  IF cfg::text NOT LIKE '%collect_external_orders%' OR cfg::text NOT LIKE '%/internal/orders%' THEN
    RAISE EXCEPTION '661 verify failed: the workflow lost the action or the box URL';
  END IF;
END $$;

COMMIT;
