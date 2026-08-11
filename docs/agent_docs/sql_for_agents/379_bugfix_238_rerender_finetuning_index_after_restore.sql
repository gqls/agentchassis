-- FILE: docs/agent_docs/sql_for_agents/379_bugfix_238_rerender_finetuning_index_after_restore.sql
--
-- bugs_open/238, step 2 of 2. Queue the re-render that carries 378's restored
-- content_data to the artefact. APPLY 378 FIRST — this file refuses if the row
-- is still in its damaged state.
--
-- THE SHAPE IS THE POINT. `spec.reason` is load-bearing: with
-- 'section_data_resolved' (or 'image_landed') the item routes through
-- rerender_page_sections, which re-renders each section from its template plus
-- its STORED content_data and merges stored ⊕ freshly-resolved — no LLM
-- anywhere. A reason-LESS page_rerender takes rerender_single_page, which
-- re-staples the stored HTML and would faithfully reproduce the five empty
-- src attributes it is meant to remove. Three reason-less items have already
-- completed on this site's pages and left them unchanged (finetuning lane §7.5).
--
-- AND NO LLM IS DELIBERATE, not incidental: a REGENERATING rebuild is exactly
-- the operation that caused bugs_open/238 in the first place. A needs_page full
-- rebuild would also rewrite the card copy, and has a poor record on this site
-- (5 failed / 4 wont_fix / 2 rejected against 20 complete).
--
-- PRE-MEASURED so the guard cannot escalate: rerender_page_sections refuses a
-- page and raises needs_page when a section lacks stored content_data or is
-- missing a required source:"llm" field. Checked 2026-08-10 against every
-- section on this page and its component schema — none is missing one, so the
-- escalation arm cannot fire. (The 11 fields this bug is about are NOT llm-
-- sourced, which is precisely why that gate never saw them: json_envelope.go
-- `if source != "llm" || !required { continue }`.)
--
-- page_id goes in the SPEC and the COLUMN both. The completed items on this
-- site carry it in the spec with a NULL column and still work, but the column
-- is what every queue query filters on — a NULL there makes the item invisible
-- to per-page coverage checks. Both, always.

\set ON_ERROR_STOP on

BEGIN;

-- ---------------------------------------------------------------------------
-- 1. Refuse unless 378 has landed and no equivalent item is already in flight.
-- ---------------------------------------------------------------------------
DO $$
DECLARE
    v_keys    int;
    v_inflight int;
BEGIN
    SELECT (SELECT count(*) FROM jsonb_object_keys(pc.content_data))
      INTO v_keys
      FROM page_components pc
     WHERE pc.id = 'e20e474f-2e22-4a60-b56d-3cc2fd0f1de7';

    IF NOT FOUND THEN
        RAISE EXCEPTION '238/379: target row not found — re-derive it (page a716cacc-..., slot case-studies-grid) before queueing';
    END IF;
    IF v_keys <> 58 THEN
        RAISE EXCEPTION '238/379: content_data has % keys, expected 58 — apply 378 first; re-rendering now would ship the damage', v_keys;
    END IF;

    -- Checking the pod does not check the queue: another session may already
    -- have a rerender in flight on this page.
    SELECT count(*) INTO v_inflight
      FROM site_work_items
     WHERE page_id = 'a716cacc-eec2-4aa6-a08b-7e6732506f41'
       AND item_type = 'page_rerender'
       AND status NOT IN ('complete', 'cancelled', 'rejected', 'wont_fix', 'failed');
    IF v_inflight > 0 THEN
        RAISE EXCEPTION '238/379: % page_rerender item(s) already open on this page — read them before adding another', v_inflight;
    END IF;
END $$;

-- ---------------------------------------------------------------------------
-- 2. The item. ON CONFLICT so a replay is a no-op rather than a duplicate.
-- ---------------------------------------------------------------------------
INSERT INTO site_work_items (
    site_id, source, pipeline, item_type, severity, summary, spec,
    page_id, priority, handler_agent, status, created_by, item_key
) VALUES (
    '1368e337-dd1d-4799-bbb3-8221a1b79bcc',
    'manual',
    'build',
    'page_rerender',
    'high',
    'Re-render index — 378 restored the 11 structural URL keys a content regeneration dropped (bugs_open/238); the stored HTML still serves five empty img src, no card links and no section CTA',
    jsonb_build_object(
        'domain',    'finetuning.uk',
        'page_id',   'a716cacc-eec2-4aa6-a08b-7e6732506f41',
        'page_name', 'index',
        'filename',  'index.html',
        -- THE LOAD-BEARING FIELD — see the header. Without it this re-staples
        -- the very HTML it exists to replace.
        'reason',    'section_data_resolved'
    ),
    'a716cacc-eec2-4aa6-a08b-7e6732506f41',
    20,                     -- ahead of the reason-less backlog (default 100)
    'page-rerender',
    'triaged',              -- directly claimable
    'bugfix-238',
    'page_rerender:index:238-structural-url-restore'
)
ON CONFLICT DO NOTHING;

DO $$
DECLARE
    v_id uuid;
BEGIN
    SELECT id INTO v_id FROM site_work_items
     WHERE item_key = 'page_rerender:index:238-structural-url-restore';
    IF v_id IS NULL THEN
        RAISE EXCEPTION '238/379: the item was not created (ON CONFLICT swallowed it?) — investigate before assuming it is queued';
    END IF;
    RAISE NOTICE '238/379: queued page_rerender % — watch it, then verify at the SERVED page, not at this row', v_id;
END $$;

COMMIT;

-- ---------------------------------------------------------------------------
-- WATCH:
--   SELECT status, attempt_count, updated_at FROM site_work_items
--    WHERE item_key = 'page_rerender:index:238-structural-url-restore';
--
-- VERIFY AT THE ARTEFACT (a `complete` status is not proof the work happened):
--   curl -s https://finetuning.uk/index.html > /tmp/ft.html
--   grep -c 'csg-card-image" src=""'          /tmp/ft.html   # want 0, is 5
--   grep -c 'src="/assets/images/case-study-' /tmp/ft.html   # want 5
--   grep -c '<a class="csg-card-link" href="' /tmp/ft.html   # want 5
--   grep -c '<a class="csg-cta-btn" href="'   /tmp/ft.html   # want 1
-- Grep the ANCHOR, not the bare class: the class name appears in the
-- component's own inline <style>, so a class grep is satisfied by CSS alone.
-- ---------------------------------------------------------------------------
