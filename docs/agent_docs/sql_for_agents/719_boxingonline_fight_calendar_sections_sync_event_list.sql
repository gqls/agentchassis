-- 719_boxingonline_fight_calendar_sections_sync_event_list.sql
--
-- bugs_open/427 Phase B step 2 (attach), tail half. Migration 712 built the
-- `event-list` component (library-only). This session attached it to
-- boxingonline.com's live fight-calendar page NOT via the full-rebuild path
-- migration 712's own footer named (check_unresolved_sections -> needs_rebuild
-- -> page-build-handler, unverified carry-forward for the two existing
-- sections) but via the narrower, already-existing mechanism the council's
-- prior_art_librarian flagged in its gating REVISE objection on migration 712
-- (ff91e666-608d-4b26-9c41-d97d23a21437): section-editor's apply_section_edit,
-- edit_type=component_swap, dispatched directly (kafka envelope carrying
-- section-editor's own live workflow, mirroring scripts/fire-section-edit.sh).
-- It swapped the EXISTING page_components row that was slot_name
-- 'generic-text-block' (build_status='pending', rendered_html EMPTY, never
-- actually served — confirmed live before the swap by curling
-- boxingonline.ugg2.com/tools/fight-calendar/index.html and finding only
-- hero-tool on the page) onto 'event-list'. No other section was touched;
-- hero-tool is untouched. Deployed and verified: git commit 007b3a7a1
-- (tools/fight-calendar/index.html), "Sync to B2" GH Actions run 33672753667
-- uploaded the new file to portfolio-sites/boxingonline.com and purged
-- Cloudflare cache for the domain.
--
-- WHAT THIS MIGRATION DOES: `pages.sections` still names the OLD slot
-- ("generic-text-block") the swap silently orphaned. Left uncorrected, that is
-- exactly the drift check_unresolved_sections.go looks for on its own regular
-- sweep — content_components has a component named 'generic-text-block'
-- (still exists as a general library row) and no page_components row for THIS
-- page joins to it any more (the swap repointed that row's component_id to
-- event-list) — so the NEXT sweep would mark this page needs_rebuild and route
-- it through the very full-rebuild pipeline this session deliberately avoided.
-- Replacing the array entry keeps the declared manifest and the actual
-- page_components composition in agreement, closing that reopened door.
--
-- NOT DONE HERE, named rather than guessed past: `query.upcoming_events`'s
-- `items` field is not yet populating for this component. Reproduced twice
-- (dispatched page-rerender, reason=section_data_resolved, both direct kafka
-- envelopes carrying page-rerender's own live workflow): each run reports
-- rerendered:2/escalated:false and produces a byte-identical git commit to the
-- swap's own output — content_data on the event-list page_component still
-- carries the OLD generic-text-block's fields ("content"/"heading"), never
-- "items"/"headline"/"empty_text". One evidence_base fact (CIT-5b2cc9894bfc475f,
-- Canelo Alvarez vs Christian Mbilli, 2026-10-31, citation url+quote both
-- present) is genuinely future-dated and should resolve; the other 5 registered
-- event facts are past results (08-29/08-30) or historic (1998), correctly
-- excluded by the resolver's own "date.Before(today)" rule — that part is
-- working as designed. Both the resolver (queryresolve/upcoming_events.go,
-- commit da2ab0d44) and its caller (plan_sections_action.go's query.* branch)
-- are confirmed live in the deployed chassis (git_commit
-- ebf27c60377f984fd2847a1d5d88ff87ae01ebf7 — only the LATER REVISE-round
-- citation-gate commit 987ed3b3b is not yet in that build, which should make
-- the deployed resolver MORE permissive, not less). No log line from either
-- function appeared across three carefully time-boxed live-log captures
-- (kubectl logs -f started before dispatch). Root cause NOT established — this
-- needs the diagnosis loop or a fresh pair of eyes with a debugger, not another
-- guess. Filed in bugs_open/427's own status section, not just here.
--
-- Lane: docs/agent_docs/docs024_key_docs_latest/bugfix_427_event_render/
-- Rollback: restore the old array entry (commented at the foot).
\set ON_ERROR_STOP on
BEGIN;

DO $$
DECLARE
  cur jsonb;
BEGIN
  SELECT sections INTO cur FROM pages
   WHERE id = '4b74ff1f-455a-4bb2-b81d-e1d0ec824f33'
     AND site_id = 'd2aa5206-73bc-4707-a69c-2702c1eb9152'
     AND name = 'tool-fight-calendar';
  IF cur IS NULL THEN
    RAISE EXCEPTION '719 ABORT: tool-fight-calendar page not found on boxingonline.com';
  END IF;
  IF NOT (cur @> '["generic-text-block"]'::jsonb) THEN
    RAISE EXCEPTION '719 ABORT: sections no longer names generic-text-block (already synced, or drifted further) — current value: %', cur;
  END IF;
END $$;

-- Pre-check: the event-list swap must actually be live before this migration
-- removes the old array entry.
DO $$
DECLARE n int;
BEGIN
  SELECT count(*) INTO n FROM page_components pc
  JOIN content_components cc ON cc.id = pc.component_id
  WHERE pc.page_id = '4b74ff1f-455a-4bb2-b81d-e1d0ec824f33'
    AND cc.function = 'event-list'
    AND COALESCE(pc.build_status, 'pending') <> 'removed';
  IF n <> 1 THEN
    RAISE EXCEPTION '719 ABORT: expected exactly 1 live event-list page_component on tool-fight-calendar, found %', n;
  END IF;
END $$;

UPDATE pages
SET sections = (
      SELECT jsonb_agg(DISTINCT x)
      FROM jsonb_array_elements(sections) x
      WHERE x <> '"generic-text-block"'::jsonb
    ) || '["event-list"]'::jsonb,
    updated_at = NOW()
WHERE id = '4b74ff1f-455a-4bb2-b81d-e1d0ec824f33'
  AND site_id = 'd2aa5206-73bc-4707-a69c-2702c1eb9152'
  AND name = 'tool-fight-calendar'
  AND sections @> '["generic-text-block"]'::jsonb
  AND NOT (sections @> '["event-list"]'::jsonb);

-- VERIFY BEFORE COMMIT (this is a check that CAN fail the transaction, not
-- just a printed SELECT — LANDMINES' own warning about "verify" sections that
-- cannot fail applies here, so this one is a RAISE, not a comment).
DO $$
DECLARE cur jsonb;
BEGIN
  SELECT sections INTO cur FROM pages WHERE id = '4b74ff1f-455a-4bb2-b81d-e1d0ec824f33';
  IF cur @> '["generic-text-block"]'::jsonb THEN
    RAISE EXCEPTION '719 ABORT: sections still names generic-text-block after UPDATE: %', cur;
  END IF;
  IF NOT (cur @> '["event-list"]'::jsonb) THEN
    RAISE EXCEPTION '719 ABORT: sections does not name event-list after UPDATE: %', cur;
  END IF;
END $$;

COMMIT;

-- Rollback (by hand, not a separate file — this migration is a one-site,
-- one-row, reviewed-inline change):
-- UPDATE pages SET sections = '["hero-tool", "generic-text-block", "advertising"]'::jsonb
-- WHERE id = '4b74ff1f-455a-4bb2-b81d-e1d0ec824f33';
