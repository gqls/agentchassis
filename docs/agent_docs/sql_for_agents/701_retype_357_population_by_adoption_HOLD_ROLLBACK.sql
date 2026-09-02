-- 701_retype_357_population_by_adoption_HOLD_ROLLBACK.sql
--
-- Returns every row that migration 701 repaired to its pre-701 state: the
-- page_components rows go back to the shared hero identity (component_id
-- 23f95f00-…, slot_name 'hero'), pages.sections and the current-plan elements
-- go back to naming 'hero', and the adopted components are removed. It is
-- driven by page_components_backup_357b_20260902 and rolls back EVERYTHING
-- that table holds — pilot and remainder together if both ran. A partial
-- rollback is a deliberate hand-edit of the WHERE clauses, not a mode.
--
-- ⚠ WHAT THIS DOES NOT UNDO. Rolling back RESTORES THE DEFECT — the rows are
-- back in bugs_open/357's population, claiming a hero identity over a whole
-- stored tool, and 357's §3 hazard is REARMED: the next rebuild of a generic
-- page whose plan says 'hero' again re-mints a 2KB title band over the working
-- tool. Unlike lendzy 693 (whose pages could not deploy at all), these 22
-- pages serve correctly on BOTH sides of this rollback — the adopted template
-- is byte-identical to the stored HTML, so there is NO artefact-level reason
-- to hurry. Check first:
--
--   -- has any 701-filed rerender already run, or any target page rebuilt
--   -- since the repair? (b.applied_at is the repair time)
--   SELECT b.domain, b.page_name, p.last_built_at, p.deployed_at, b.applied_at,
--          swi.status AS rerender_status
--     FROM page_components_backup_357b_20260902 b
--     JOIN pages p ON p.id = b.page_id
--     LEFT JOIN site_work_items swi
--       ON swi.site_id = b.site_id
--      AND swi.item_key = 'page_rerender_' || b.page_name || '_' || b.site_id || '_assemble'
--      AND swi.source = 'bugfix_357_component_identity lane (migration 701)'
--    ORDER BY b.domain, b.page_name;
--
-- If a rerender is claimed/complete, or last_built_at postdates applied_at,
-- the served artefact was ASSEMBLED FROM the new identity: rolling the record
-- back then leaves the record disagreeing with how the artefact was produced,
-- and hands the next rebuild the re-mint disaster. Do NOT roll back blindly —
-- the reason for reverting needs re-stating against that evidence.
--
-- Apply by hand, one transaction:
--   kubectl -n ai-persona-system exec -i postgres-clients-0 -- \
--     psql -U clients_user -d clients_db -v ON_ERROR_STOP=1 \
--     < 701_retype_357_population_by_adoption_HOLD_ROLLBACK.sql
--
-- Deviations from 693's rollback, argued in notes_700_draft.md:
--   * strict state classification per row (post-state → restore; already
--     pre-state → skip; anything else → ABORT naming the row) instead of a
--     blanket predicate UPDATE — three legs moved here, so a drifted leg must
--     be seen, not silently half-restored;
--   * the unreferenced check before DELETE also covers site_components and
--     forked_from (someone may have forked FROM an adopted row), not just
--     page_components;
--   * a rollback doc_notes row is filed, so the decision trail does not end on
--     "adoption happened" when it was later undone.
-- The backup table itself is KEPT (evidence, and re-apply support).

BEGIN;

-- ---------------------------------------------------------------------------
-- GUARD R1. There must be something to roll back, and every backed-up row
-- must be in a state this file understands. States:
--   'post' — as migration 701 left it (restore it);
--   'pre'  — already restored (skip it, idempotent re-run);
--   anything else — ABORT: a later actor moved the row, and a blind restore
--   would clobber their work; resolve by hand.
-- ---------------------------------------------------------------------------
DO $$
DECLARE
  n int; r record; pc record; pg record;
  cc_id uuid; errs text[] := '{}'; state text;
BEGIN
  SELECT count(*) INTO n FROM page_components_backup_357b_20260902;
  IF n = 0 THEN
    RAISE EXCEPTION '701 ROLLBACK ABORT: the backup table is empty — nothing was applied, or the wrong database';
  END IF;

  FOR r IN SELECT * FROM page_components_backup_357b_20260902 ORDER BY domain, page_name LOOP
    SELECT * INTO pc FROM page_components WHERE id = r.pc_id;
    IF NOT FOUND THEN
      errs := errs || (r.domain || '/' || r.page_name || ': page_components row ' || r.pc_id || ' is GONE');
      CONTINUE;
    END IF;
    IF md5(COALESCE(pc.rendered_html, '')) <> r.pre_md5 THEN
      errs := errs || (r.domain || '/' || r.page_name || ': stored bytes have MOVED since the backup (md5 ' || md5(COALESCE(pc.rendered_html, '')) || ' vs ' || r.pre_md5 || ') — a rebuild has rewritten this row; do not blind-restore');
      CONTINUE;
    END IF;

    SELECT id INTO cc_id FROM content_components
     WHERE name = r.new_name AND created_from = 'adopted';

    IF cc_id IS NOT NULL AND pc.component_id = cc_id AND pc.slot_name = r.new_name THEN
      state := 'post';
    ELSIF pc.component_id = r.pre_component_id AND pc.slot_name = r.pre_slot_name THEN
      state := 'pre';
    ELSE
      errs := errs || (r.domain || '/' || r.page_name || ': row is in NEITHER the post-701 nor the pre-701 state (component_id ' || COALESCE(pc.component_id::text, 'NULL') || ', slot ' || COALESCE(pc.slot_name, 'NULL') || ') — a later actor moved it; resolve by hand');
      CONTINUE;
    END IF;

    -- The pages.sections leg must be in the matching state.
    SELECT * INTO pg FROM pages WHERE id = r.page_id;
    IF NOT FOUND THEN
      errs := errs || (r.domain || '/' || r.page_name || ': pages row is GONE');
      CONTINUE;
    END IF;
    IF state = 'post' THEN
      IF pg.sections IS DISTINCT FROM (
           SELECT jsonb_agg(CASE WHEN e.value = to_jsonb('hero'::text)
                                 THEN to_jsonb(r.new_name) ELSE e.value END ORDER BY e.ord)
             FROM jsonb_array_elements(r.pre_pages_sections) WITH ORDINALITY e(value, ord)) THEN
        errs := errs || (r.domain || '/' || r.page_name || ': pages.sections is not the expected post-701 value (' || pg.sections::text || ') — something else has edited it; resolve by hand');
      END IF;
    ELSIF pg.sections IS DISTINCT FROM r.pre_pages_sections THEN
      errs := errs || (r.domain || '/' || r.page_name || ': row is pre-701 but pages.sections is not (' || pg.sections::text || ') — resolve by hand');
    END IF;

    -- The plan leg (absent for tool-ttk-calculator, whose backup row pins
    -- pre_plan_row_id NULL): the backed-up plan row must still exist and read
    -- either the new name (post) or hero (pre).
    IF r.pre_plan_row_id IS NOT NULL THEN
      PERFORM 1 FROM site_plan_sections sps
        WHERE sps.id = r.pre_plan_row_id
          AND sps.component_name IN ('hero', r.new_name);
      IF NOT FOUND THEN
        errs := errs || (r.domain || '/' || r.page_name || ': the backed-up plan row ' || r.pre_plan_row_id || ' is gone or renamed to something else — the plan has been replaced or edited since; resolve by hand');
      END IF;
    END IF;
  END LOOP;

  IF array_length(errs, 1) IS NOT NULL THEN
    RAISE EXCEPTION '701 ROLLBACK ABORT: % row(s) not safely restorable: %',
      array_length(errs, 1), array_to_string(errs, ' || ');
  END IF;
END $$;

-- ---------------------------------------------------------------------------
-- 1. Withdraw the rerenders migration 701 queued, if they have not been picked
-- up. A CLAIMED or COMPLETE one is left alone deliberately: cancelling a row a
-- handler is already acting on does not stop the handler, it only makes the
-- record lie. (693's shape.)
-- ---------------------------------------------------------------------------
UPDATE site_work_items
   SET status = 'cancelled',
       updated_at = NOW()
 WHERE item_type = 'page_rerender'
   AND source = 'bugfix_357_component_identity lane (migration 701)'
   AND status = 'triaged'
   AND site_id IN (SELECT DISTINCT site_id FROM page_components_backup_357b_20260902);

-- ---------------------------------------------------------------------------
-- 2. Restore the page_components rows still in the post-701 state.
-- ---------------------------------------------------------------------------
UPDATE page_components pc
   SET component_id = b.pre_component_id,
       slot_name    = b.pre_slot_name,
       updated_at   = NOW()
  FROM page_components_backup_357b_20260902 b
  JOIN content_components cc ON cc.name = b.new_name AND cc.created_from = 'adopted'
 WHERE pc.id = b.pc_id
   AND pc.component_id = cc.id
   AND pc.slot_name = b.new_name;

-- ---------------------------------------------------------------------------
-- 3. Restore pages.sections to the exact backed-up pre-state (guard R1 has
-- proven the current value is the expected post-state wherever this fires).
-- ---------------------------------------------------------------------------
UPDATE pages p
   SET sections   = b.pre_pages_sections,
       updated_at = NOW()
  FROM page_components_backup_357b_20260902 b
 WHERE p.id = b.page_id
   AND p.sections IS DISTINCT FROM b.pre_pages_sections;

-- ---------------------------------------------------------------------------
-- 4. Restore the plan elements, keyed on the backed-up site_plan_sections id.
-- ---------------------------------------------------------------------------
UPDATE site_plan_sections sps
   SET component_name = 'hero'
  FROM page_components_backup_357b_20260902 b
 WHERE b.pre_plan_row_id IS NOT NULL
   AND sps.id = b.pre_plan_row_id
   AND sps.component_name = b.new_name;

-- ---------------------------------------------------------------------------
-- 5. Remove the adopted components — only where NOTHING references them:
-- no page_components row, no site_components row, and no component that was
-- forked FROM them (component_versions rows, had any been stamped since,
-- cascade with the delete). A survivor is caught by the verify below.
-- ---------------------------------------------------------------------------
DELETE FROM content_components cc
 WHERE cc.created_from = 'adopted'
   AND cc.name IN (SELECT new_name FROM page_components_backup_357b_20260902)
   AND NOT EXISTS (SELECT 1 FROM page_components pc WHERE pc.component_id = cc.id)
   AND NOT EXISTS (SELECT 1 FROM site_components sc WHERE sc.component_id = cc.id)
   AND NOT EXISTS (SELECT 1 FROM content_components f WHERE f.forked_from = cc.id);

-- ---------------------------------------------------------------------------
-- 6. Rollback of record, so the doc_notes trail does not end on "adoption
-- happened" when it was later undone.
-- ---------------------------------------------------------------------------
INSERT INTO doc_notes (subject_type, subject_key, site_id, body, categories, source, created_by)
SELECT 'decision', b.domain, b.site_id,
       'ROLLBACK of migration 701 (bugs_open/357 phase 3 Option B): the ' || count(*)
       || ' adopted tool component(s) on ' || b.domain || ' ('
       || string_agg(b.page_name, ', ' ORDER BY b.page_name)
       || ') were removed and the page rows, plan elements and pages.sections restored to the shared hero identity. '
       || 'The bugs_open/357 defect is DELIBERATELY back in place for these pages, and the s3 hazard is REARMED: '
       || 'a rebuild of a generic one will re-mint a hero band over the stored tool. See the '
       || 'bugfix_357_component_identity lane NOTES for why this rollback was run.',
       '["bugfix-357","tool-adoption","component-identity"]'::jsonb,
       'bugfix_357_component_identity lane',
       'bugfix_357_component_identity lane (migration 701 rollback)'
  FROM page_components_backup_357b_20260902 b
 GROUP BY b.domain, b.site_id;

-- ---------------------------------------------------------------------------
-- VERIFY, as DO/RAISE (a verify block of bare SELECTs cannot stop the COMMIT).
-- ---------------------------------------------------------------------------
DO $$
DECLARE n_backup int; restored int; sections_ok int; plans_ok int; plans_expected int;
        comps int; live int;
BEGIN
  SELECT count(*) INTO n_backup FROM page_components_backup_357b_20260902;

  SELECT count(*) INTO restored
    FROM page_components_backup_357b_20260902 b
    JOIN page_components pc ON pc.id = b.pc_id
   WHERE pc.component_id = b.pre_component_id
     AND pc.slot_name = b.pre_slot_name
     AND pc.position = b.pre_position
     AND md5(COALESCE(pc.rendered_html, '')) = b.pre_md5;
  IF restored <> n_backup THEN
    RAISE EXCEPTION '701 ROLLBACK VERIFY: expected % row(s) back at the hero identity with bytes unchanged, found %',
      n_backup, restored;
  END IF;

  SELECT count(*) INTO sections_ok
    FROM page_components_backup_357b_20260902 b
    JOIN pages p ON p.id = b.page_id
   WHERE p.sections = b.pre_pages_sections;
  IF sections_ok <> n_backup THEN
    RAISE EXCEPTION '701 ROLLBACK VERIFY: expected % pages.sections restored, found %', n_backup, sections_ok;
  END IF;

  SELECT count(*) FILTER (WHERE b.pre_plan_row_id IS NOT NULL), count(sps.id)
    INTO plans_expected, plans_ok
    FROM page_components_backup_357b_20260902 b
    LEFT JOIN site_plan_sections sps
      ON sps.id = b.pre_plan_row_id AND sps.component_name = 'hero'
   WHERE b.pre_plan_row_id IS NOT NULL;
  IF plans_ok <> plans_expected THEN
    RAISE EXCEPTION '701 ROLLBACK VERIFY: expected % plan element(s) back at hero, found %',
      plans_expected, plans_ok;
  END IF;

  SELECT count(*) INTO comps
    FROM content_components cc
   WHERE cc.name IN (SELECT new_name FROM page_components_backup_357b_20260902);
  IF comps <> 0 THEN
    RAISE EXCEPTION '701 ROLLBACK VERIFY: % adopted component(s) survive — something still references one (a page_components or site_components row, or a fork); investigate before re-running', comps;
  END IF;

  SELECT count(*) INTO live
    FROM site_work_items swi
   WHERE swi.source = 'bugfix_357_component_identity lane (migration 701)'
     AND swi.item_type = 'page_rerender'
     AND swi.status = 'triaged';
  IF live <> 0 THEN
    RAISE EXCEPTION '701 ROLLBACK VERIFY: % filed rerender(s) still triaged after the cancel sweep', live;
  END IF;

  RAISE NOTICE '701 ROLLBACK OK: % row(s) restored to the hero identity (bytes unchanged), sections and plan elements back to ''hero'', adopted components removed, untouched rerenders cancelled — the bugs_open/357 defect is back in place and its population predicate will again return these rows. The backup table page_components_backup_357b_20260902 is kept.',
    n_backup;
END $$;

COMMIT;
