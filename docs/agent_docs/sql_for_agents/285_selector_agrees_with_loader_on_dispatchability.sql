-- 285 — find_dispatchable_site applies the SAME dispatchability test as the loader
--
-- Head-of-line blocking. Discovered 2026-08-02 while verifying 284 (the fairness
-- ORDER BY) — 284 is correct and stays; this is a SECOND, independent defect that
-- 284 did not cause but does make permanent. Config only, no image dependency.
--
-- ── THE DEFECT ────────────────────────────────────────────────────────────
-- Two different queries decide "is this item dispatchable", and they disagree.
--
--   SELECTOR — build-pipeline-trigger.find_dispatchable_site — picks the SITE:
--     status IN ('triaged','approved') AND attempt_count < max_attempts
--     AND no 'claimed' item on the site
--
--   LOADER — LoadWorkItemsAction (load_work_item_actions.go), inside
--   build-dispatch-loop, picks the ITEMS on that site. Same three, PLUS TWO:
--     AND (COALESCE(approval_mode,'auto') = 'auto' OR status = 'approved')
--     AND (depends_on IS NULL OR NOT EXISTS (unresolved dependency))
--
-- So the selector can hand the loop a site whose only "eligible" item the loader
-- refuses. The loop then loads ZERO items, reports has_items=false with
-- rows_dropped=0, notifies the scheduler and completes — successfully, having
-- done nothing. The site still has no 'claimed' item, so it is still eligible
-- next tick. It is picked again. **The queue does not advance, and nothing
-- anywhere reports an error.**
--
-- ── MEASURED, 2026-08-02 ──────────────────────────────────────────────────
-- robot-hands.com holds exactly ONE selector-eligible item: `93f2a3b7`
-- (content_rewrite, page-build-handler, priority 110, created 07-31 12:27:28).
-- It depends on `0733a7a4`, a `needs_content_page` item sitting in
-- **needs_human_review** — a state no automation will ever move to
-- complete/verified. So the item is permanently unloadable.
--
-- Fleet-wide that day: exactly 1 item in this blocked class, out of 366
-- selector-eligible items across 17 sites — and it belonged to whichever site
-- was at the head of the queue. One row stalls the fleet.
--
-- Timeline (claims, all sites):
--   08:03–08:06  robot-hands drains its last 5 CLAIMABLE items
--   08:06–09:36  **NOTHING claimed anywhere for 89 minutes** — robot-hands is
--                lowest-UUID, so the pre-284 selector picked it every tick and
--                every dispatch loop loaded 0 items
--   09:36        284 goes live; gamesdesign (starved 3d10h) is served, 5 items
--   09:41→       robot-hands is back at the head — now by AGE, its item being
--                the fleet's oldest — and the fleet stalls again
--
-- **This retires a wrong call of mine.** I twice recorded ~90-minute fleet quiet
-- spells as "comparable to known behaviour, not yet outside it". They were not
-- benign: that IS this mechanism, and it is exactly reproducible. A recurring
-- gap that matches a known range is not thereby explained.
--
-- ── WHY 284 MAKES IT PERMANENT RATHER THAN INTERMITTENT ───────────────────
-- Under the old UUID order a blocked site held the head only while it happened
-- to sort lowest, and released it when a lower-UUID site gained work. Under
-- oldest-waiting-first the blocked item's key is its `created_at`, which never
-- changes and only ages — so an unloadable item, once at the head, is at the
-- head FOR EVER. Fairness ordering is right, and it converts an intermittent
-- stall into a permanent one unless the selector agrees with the loader. Both
-- properties are needed; neither alone is sufficient.
--
-- ── THE FIX ───────────────────────────────────────────────────────────────
-- Give the selector the loader's two extra clauses verbatim, so "this site has
-- dispatchable work" means the same thing in both places.
--
-- STRICTLY NARROWING, and that is the whole safety argument: every site this
-- removes is a site where the loop would have loaded 0 items and claimed
-- nothing. No site that would have been served loses service; the only
-- behaviour removed is the wasted pick that blocks everyone behind it.
--
-- NOT mirrored, deliberately: the loader's optional `item_pipeline` /
-- `handler_agent` filters. The live build-dispatch-loop `load_items` config
-- carries neither (`{"site_id": …, "max_items": 5}`), so mirroring them would
-- invent a filter the loader does not apply — and re-introduce the dead
-- `domain='build'` clause that migration 067 removed for that exact reason.
--
-- KNOWN, NOT FIXED HERE — the dependency subquery is SITE-SCOPED in the loader
-- (`WHERE site_id = $1`), so a `depends_on` pointing at another site's item can
-- never resolve and blocks its item for ever. This migration copies that
-- behaviour deliberately: the selector's job is to AGREE with the loader, not
-- to be independently correct. Fixing the cross-site case means changing the
-- loader (Go) and is a separate change with its own blast radius. Recorded as a
-- landmine + in the register.
--
-- ALSO NOT FIXED: robot-hands' actual blockage. `0733a7a4` needs a human — that
-- is a work-item triage question, not a dispatch defect. After this migration
-- the queue simply stops waiting for it. Whoever picks that up: the item is
-- `needs_content_page` on robot-hands.com, and `93f2a3b7` unblocks the moment it
-- reaches complete/verified.
--
-- COUNCIL: not submitted — gate scope is platform/, internal/, pkg/ and refuses
-- config-only submissions client-side; no Go half. Owner authorised the
-- starvation work directly 2026-08-02; this is the same defect's second half
-- (a fair queue that picks an unservable head is not fair, it is stuck).

BEGIN;

-- ── STEP 1 — PRE-FLIGHT ASSERTION ─────────────────────────────────────────
-- Expect exactly 1: the live row still carrying 284's query, byte-identical.
-- Anything else means the row drifted — stop and re-read it.
SELECT count(*) AS rows_to_change_expect_1
FROM agent_definitions
WHERE type = 'build-pipeline-trigger'
  AND is_active AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL
  AND md5(default_config #>> '{workflow,steps,find_dispatchable_site,config,query}')
      = 'dd9eb707e342220a19daa77ddddd8b2a';

-- ── STEP 2 — SNAPSHOT ─────────────────────────────────────────────────────
SELECT snapshot_agent('build-pipeline-trigger',
                      'selector agrees with loader on dispatchability — head-of-line block (285)');

-- ── STEP 3 — THE CHANGE ───────────────────────────────────────────────────
UPDATE agent_definitions
SET default_config = jsonb_set(
      default_config,
      '{workflow,steps,find_dispatchable_site,config,query}',
      $q$"SELECT wi.site_id::text, s.domain FROM site_work_items wi JOIN sites s ON s.id = wi.site_id WHERE wi.status IN ('triaged', 'approved') AND wi.attempt_count < wi.max_attempts AND (COALESCE(wi.approval_mode, 'auto') = 'auto' OR wi.status = 'approved') AND (wi.depends_on IS NULL OR NOT EXISTS (SELECT 1 FROM unnest(wi.depends_on) dep_id WHERE dep_id NOT IN (SELECT id FROM site_work_items WHERE site_id = wi.site_id AND status IN ('complete', 'verified')))) AND NOT EXISTS (SELECT 1 FROM site_work_items active WHERE active.site_id = wi.site_id AND active.status = 'claimed') ORDER BY wi.created_at ASC, wi.priority ASC, wi.id ASC LIMIT 1"$q$::jsonb,
      false
    ),
    updated_at = now()
WHERE type = 'build-pipeline-trigger'
  AND is_active AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL
  AND md5(default_config #>> '{workflow,steps,find_dispatchable_site,config,query}')
      = 'dd9eb707e342220a19daa77ddddd8b2a';   -- idempotent + refuses a drifted row

-- ── STEP 4 — VERIFY BEFORE COMMIT ─────────────────────────────────────────
-- Expect the query to contain BOTH 'approval_mode' and 'depends_on'.
SELECT type,
       default_config #>> '{workflow,steps,find_dispatchable_site,config,query}' LIKE '%approval_mode%' AS has_approval_clause,
       default_config #>> '{workflow,steps,find_dispatchable_site,config,query}' LIKE '%depends_on%'    AS has_depends_clause,
       default_config #>> '{workflow,steps,find_dispatchable_site,config,query}' LIKE '%created_at ASC%' AS still_fifo
FROM agent_definitions
WHERE type = 'build-pipeline-trigger'
  AND is_active AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;

COMMIT;

-- ── ROLLBACK ──
-- Preferred: restore the step-2 snapshot. Note the correct target is 284's
-- query (FIFO, no extra clauses) — NOT the pre-284 DISTINCT ON form.

-- ── VERIFY THE FIX AT THE ARTEFACT ──
-- Within a couple of ticks (scheduler interval 120s), claims should resume on a
-- site that is NOT robot-hands. Predicted first pick at the time of writing:
-- relojistas.com (oldest unblocked item, 07-31 13:59:23).
--
--   SELECT s.domain, wi.claimed_at FROM site_work_items wi
--   JOIN sites s ON s.id = wi.site_id
--   WHERE wi.claimed_at > now() - interval '10 minutes' ORDER BY wi.claimed_at;
--
-- The decisive negative control: robot-hands.com must NOT appear, and its
-- `93f2a3b7` must remain triaged — it is genuinely undispatchable and the point
-- of this change is that the queue stops waiting for it.
--
-- To confirm the loop is doing real work rather than idling (this is what the
-- defect looked like — COMPLETED, no error, nothing done):
--   SELECT collected_data->'load_items'->>'item_count',
--          collected_data->'load_items'->>'has_items'
--   FROM orchestration_states WHERE owner_agent_type='build-dispatch-loop'
--   ORDER BY created_at DESC LIMIT 5;   -- expect non-zero item_count
