-- ============================================================================
-- Phase 1.5 — Smoke test queries
--
-- Site under test: 00ff3af5-dad8-4770-9f70-3edc267a3c92 (robot-hands.com)
--
-- Pre-conditions (already verified):
--   - Migration phase_1_register_imagery_checks.sql applied: checks array
--     has 14 entries including unfulfilled_image_prompt, placeholder_image_in_use,
--     image_url_404.
--   - Chassis binary registers all three checks (visible in the registered:
--     array of RunDiscoveryChecksAction logs).
--   - Site has 7 image_prompts and zero active assets — textbook gap.
--
-- Flow being verified:
--   trigger
--     → design-discovery-agent (run_discovery_checks)
--     → triage_detected_items (detected → triaged, pipeline → 'build')
--     → build-dispatch-loop (claims item, spawns handler from current_item.handler_agent)
--     → image-build-handler (call_hero_gen / call_logo_gen → store_asset → deploy)
-- ============================================================================


-- STEP 1 — Trigger discovery
-- ----------------------------------------------------------------------------
-- Run from the kafka tooling pod / shell, not in psql:
--
--   ./trigger-audit.sh design-discovery-agent 00ff3af5-dad8-4770-9f70-3edc267a3c92
--
-- Or via the kcat block from before, with site_id 00ff3af5-... and
-- AGENT_TYPE=design-discovery-agent.
--
-- Expectation: chassis logs show RunDiscoveryChecksAction firing for the
-- 14 checks, including the three new imagery checks. Action completes in
-- under ~5 seconds.


-- STEP 2 — Confirm work items were created
-- ----------------------------------------------------------------------------
SELECT id,
       item_type,
       status,
       priority,
       severity,
       handler_agent,
       LEFT(summary, 100)        AS summary,
       LEFT(spec::text, 250)     AS spec_preview,
       created_at,
       created_by,
       item_key
FROM site_work_items
WHERE site_id = '00ff3af5-dad8-4770-9f70-3edc267a3c92'
  AND created_by = 'design-discovery-agent'
  AND created_at > NOW() - INTERVAL '5 minutes'
ORDER BY created_at DESC;

-- Expectation: 7 rows, all status='detected':
--
--   2 routable (handler_agent = 'image-build-handler', severity = 'high'):
--     item_type=needs_logo,        item_key=unfulfilled_image_prompt:logo
--     item_type=needs_hero_image,  item_key=unfulfilled_image_prompt:hero
--
--   5 flag-only (handler_agent = '' or NULL, severity = 'medium'):
--     item_type=unfulfilled_hero_variant, distinct item_keys:
--       unfulfilled_image_prompt:hero_about
--       unfulfilled_image_prompt:hero_tools
--       unfulfilled_image_prompt:hero_matchmatrix
--       unfulfilled_image_prompt:hero_how_it_works
--       unfulfilled_image_prompt:hero_selection_guide
--
-- Each spec contains "check": "unfulfilled_image_prompt", "purpose": <p>,
-- "prompt_key": <k>, and "image_prompts": {<k>: <prompt text>}.
--
-- The other two new checks (placeholder_image_in_use, image_url_404) are
-- expected to find nothing on this site — diagnostic B earlier showed
-- uses_fallback_paths=false and no /assets/images/* references at all.


-- STEP 3 — Confirm triage promotes them
-- ----------------------------------------------------------------------------
-- The triage step runs as part of improvement-loop. If you've triggered
-- improvement-loop directly, skip ahead. Otherwise call triage manually
-- by sending a workflow request that hits triage_detected_items, OR wait
-- for the scheduled improvement-loop pass.
--
-- After triage:

SELECT item_type,
       status,
       pipeline,
       spec->>'original_pipeline' AS original_pipeline,
       handler_agent,
       severity,
       priority
FROM site_work_items
WHERE site_id = '00ff3af5-dad8-4770-9f70-3edc267a3c92'
  AND created_by = 'design-discovery-agent'
  AND created_at > NOW() - INTERVAL '10 minutes'
ORDER BY created_at DESC;

-- Expectation:
--   - All 7 rows: status='triaged', pipeline='build'
--   - spec.original_pipeline='design' (preserved by triage_detected_items)
--   - handler_agent unchanged from STEP 2
--
-- Note: triage promotes everything blindly, including the 5 flag-only items
-- with empty handler_agent. They reach 'triaged' status but build-dispatch-loop
-- skips them because there's no handler to spawn.


-- STEP 4 — Confirm build-dispatch-loop claims the routable ones
-- ----------------------------------------------------------------------------
-- build-dispatch-loop runs on the build-pipeline-trigger schedule, OR can
-- be invoked directly. Watch it picking up needs_logo and needs_hero_image:
--
--   kubectl -n ai-persona-system logs -f -l app=agent-chassis --tail=200 \
--     | grep -E 'build-dispatch-loop|image-build-handler|claim_work_item'
--
-- After dispatch begins:

SELECT item_type,
       status,
       handler_agent,
       attempt_count,
       claimed_at,
       claimed_by,
       updated_at
FROM site_work_items
WHERE site_id = '00ff3af5-dad8-4770-9f70-3edc267a3c92'
  AND created_by = 'design-discovery-agent'
ORDER BY created_at DESC;

-- Expectation:
--   - 2 routable items transition: triaged → claimed → complete
--     (claimed_at populated, attempt_count goes to 1+)
--   - 5 flag-only items remain triaged with empty handler_agent and
--     no claimed_at — dispatch ignores them by design.


-- STEP 5 — Confirm assets were generated and deployed
-- ----------------------------------------------------------------------------
SELECT a.purpose,
       a.origin_type,
       a.origin_model,
       LEFT(a.origin_prompt, 200) AS origin_prompt_preview,
       a.url,
       a.created_at,
       a.updated_at
FROM assets a
WHERE a.site_id = '00ff3af5-dad8-4770-9f70-3edc267a3c92'
  AND a.status = 'active'
ORDER BY a.created_at DESC;

-- Expectation (after the routable items complete):
--   - 2 rows with purpose 'logo' and 'hero'
--   - origin_type = 'generated'
--   - origin_model = 'sdxl'                 (Phase 0.2 wiring)
--   - origin_prompt begins with the imagery_direction prefix
--                                            (Phase 0.1 composition)
--                                            then the planner's image_prompts
--                                            value for that purpose
--
-- Sample of what origin_prompt should look like for the gripper logo
-- (~190 chars, post-truncation):
--
--   "Industrial photography of robotic grippers and end-effectors in real
--    manufacturing environments — close-up detail shots showing machined
--    surfaces, actuator mechanisms, and gripper-to-workpiece interaction.
--    A precise, technical logomark for Robot-Hands.com..."


-- STEP 6 — Confirm idempotency
-- ----------------------------------------------------------------------------
-- After STEP 5 completes (logo + hero assets exist), trigger discovery
-- again. The check should NOT produce duplicate work items for the
-- now-fulfilled purposes, but should still produce the 5 flag-only items
-- because nothing has been done about them.

-- After re-trigger:

SELECT item_type,
       COUNT(*)                                              AS total,
       COUNT(*) FILTER (WHERE status = 'detected')           AS detected_now,
       COUNT(*) FILTER (WHERE status IN ('triaged','claimed')) AS in_flight,
       COUNT(*) FILTER (WHERE status = 'complete')           AS done
FROM site_work_items
WHERE site_id = '00ff3af5-dad8-4770-9f70-3edc267a3c92'
  AND created_by = 'design-discovery-agent'
  AND item_type IN ('needs_logo','needs_hero_image','unfulfilled_hero_variant','image_url_404')
GROUP BY item_type
ORDER BY item_type;

-- Expectation:
--   - needs_logo: 1 total, 1 done — no duplicate inserted
--   - needs_hero_image: 1 total, 1 done — no duplicate inserted
--   - unfulfilled_hero_variant: 5 total, 5 detected/triaged
--                            (still flag-only, still recorded)
--
-- Dedup is enforced by the unique-on-(site_id, item_key) constraint in
-- site_work_items: the check tries to insert with the same item_key, the
-- DB rejects, the insertWorkItem helper logs a skip rather than failing.


-- ============================================================================
-- ESCAPE HATCHES
-- ============================================================================

-- E1 — directly check the discovery action's findings count from logs:
--
--   kubectl -n ai-persona-system logs -l app=agent-chassis --since=10m \
--     | grep 'RunDiscoveryChecksAction: Complete' \
--     | tail -3
--
--   Expected to show "findings:N, items_inserted:M" where N >= 7 (5 from
--   our checks + whatever the existing 11 checks produce). M may be less
--   than N if some findings collide on item_key with prior runs.


-- E2 — manually create a routable test item to bypass discovery:
--      (uncomment if STEP 4 stalls and you want to test dispatch in isolation)
--
-- INSERT INTO site_work_items (
--     site_id, source, pipeline, item_type, severity, summary, spec,
--     priority, handler_agent, status, created_by, item_key, batch_id
-- ) VALUES (
--     '00ff3af5-dad8-4770-9f70-3edc267a3c92',
--     'manual-smoke-test',
--     'build',
--     'needs_logo',
--     'high',
--     'Manual smoke test logo regeneration',
--     '{"check":"manual","purpose":"logo","image_prompts":{"logo":"A precise, technical logomark for Robot-Hands.com — a stylised robotic gripper or end-effector silhouette rendered in clean geometric lines."}}'::jsonb,
--     70,
--     'image-build-handler',
--     'triaged',
--     'manual-smoke-test',
--     'manual_smoke_test_logo_' || EXTRACT(EPOCH FROM NOW())::text,
--     gen_random_uuid()
-- );


-- E3 — see the full spec on a single work item:
--      (substitute an id from STEP 2)
--
-- SELECT spec FROM site_work_items WHERE id = '<uuid>';
