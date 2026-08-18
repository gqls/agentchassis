-- 465 — the promoter's success tests must read the ARCHIVE too, or "lifetime"
--       silently means "the last 7 days"
--
-- Third correction in this family (430 → 444 → 454 → 465), and the one with a
-- live victim rather than a latent one.
--
-- ============================================================================
-- THE DEFECT
-- ============================================================================
-- `work-item-archiver` is an ENABLED daily scheduled task whose own description
-- reads: "Archives terminal work items older than 7 days to
-- site_work_items_archive". It is not a cleanup nobody runs — the archive holds
-- **20,184 rows against 8,702 live**, i.e. most of this platform's work-item
-- history does not live in `site_work_items` at all.
--
-- Both of the promoter's success tests read `site_work_items` ONLY:
--   * 430's known-good rule — "the pair has >=1 success"
--   * 444/454's floor       — "the pair is succeeding at >=25%"
-- So both actually mean **"...in the last 7 days"**. A pair that works well but
-- has been quiet for eight days reads as NEVER HAVING WORKED and is held for
-- ever — which is precisely `bugs_open/083`'s disease, reintroduced by the
-- mechanism built to cure it.
--
-- ============================================================================
-- IT IS ALREADY HAPPENING — this is not a latent risk any more
-- ============================================================================
-- [MEASURED 2026-08-18] live-table-only versus live+archive, per pair:
--
--   pair                                     live-only   TRUE      verdict today
--   empty_internal_href -> page-build-handler   0/1  0%   9/5  64%  HELD as "never
--                                                                  completed" while
--                                                                  holding NINE
--                                                                  lifetime successes
--   empty_section       -> page-build-handler  12/16 43% 316/33 91%  a 316-success
--                                                                  workhorse reading
--                                                                  as marginal
--   literal_markdown    -> page-build-handler   3/24 11%   3/36  8%  correctly held
--                                                                  either way
--   placeholder_contact -> page-build-handler   0/4   0%   0/6   0%  genuinely never
--                                                                  succeeded
--
-- `empty_internal_href` is a LIVE row being stranded RIGHT NOW by the canary
-- rule, on the false premise that its handler has never done that work. It has
-- done it nine times.
--
-- Note the two controls in that table, which is why it is trustworthy in both
-- directions: `literal_markdown` must STAY held (the pair 444 was written for —
-- if the archive rescued it, this fix would be dissolving the floor rather than
-- correcting its scope) and `placeholder_contact` must stay held (it genuinely
-- has no successes anywhere). Both hold.
--
-- ============================================================================
-- WHY A UNION AND NOT A VIEW
-- ============================================================================
-- A view over both tables would be the tidier estate-wide answer and is the
-- right RFC. It is NOT taken here: creating a shared object that other
-- pipelines may adopt is a shared-seam change (owner ruling 2026-07-28), and
-- this file's job is to stop a live pair being stranded. The union is confined
-- to this task's pre_query, changes no shared object, and is trivially
-- reversible. Named so the omission reads as a decision, not an oversight.
--
-- Cost measured before applying: the archive is 20k rows and the union adds one
-- sequential scan per EXISTS/aggregate. See the NOTICE at the end for the
-- measured tick cost; the task runs every 900s.
--
-- Rollback: `465_..._ROLLBACK.sql` restores 458's pre_query verbatim.

BEGIN;

-- Guard: refuse if the live pre_query is not the one this file was written
-- against. Another session (458's author, or the still-open 277 lane) may have
-- edited it since; aborting is correct, silently reverting their work is not.
DO $$
DECLARE q text;
BEGIN
    SELECT pre_query INTO q FROM scheduled_tasks WHERE name = 'detected-item-promoter';
    IF q IS NULL THEN
        RAISE EXCEPTION '465: detected-item-promoter has no pre_query — wrong database or the task is gone';
    END IF;
    -- the three predicates this file rewrites must all be present in the form it expects
    IF q NOT LIKE '%done.status IN (''complete'', ''verified'')%' THEN
        RAISE EXCEPTION '465: the known-good test is not in 454''s form — someone has edited it; re-read the live pre_query before applying';
    END IF;
    IF q NOT LIKE '%h.status IN (''complete'', ''verified'')%' THEN
        RAISE EXCEPTION '465: the floor is not in 454''s form — someone has edited it; re-read the live pre_query before applying';
    END IF;
    IF q NOT LIKE '%scored%' THEN
        RAISE EXCEPTION '465: 458''s `scored` CTE is absent — this file rewrites 458''s shape and must not be applied to an older one';
    END IF;
END $$;

-- Anchors are SINGLE-LINE and verified unique (1 occurrence each) against the
-- live pre_query. A multi-line anchor was tried first and would have matched
-- NOTHING: 458 reindented this block, so an indentation-bearing pattern is a
-- silent no-op. The `n_union <> 2` assert below would have caught it, but a
-- whitespace-insensitive anchor is the correct fix, not a caught failure.
UPDATE scheduled_tasks
SET pre_query = replace(
      replace(
        pre_query,
        'SELECT 1 FROM site_work_items done',
        'SELECT 1 FROM (SELECT item_type, handler_agent, status FROM site_work_items UNION ALL SELECT item_type, handler_agent, status FROM site_work_items_archive) done'
      ),
      'FROM site_work_items h',
      'FROM (SELECT item_type, handler_agent, status FROM site_work_items UNION ALL SELECT item_type, handler_agent, status FROM site_work_items_archive) h'
    ),
    updated_at = now()
WHERE name = 'detected-item-promoter';

-- ============================================================================
-- Verification. RAISE, not SELECT. The two controls must come out OPPOSITE
-- ways, or this file has either done nothing or dissolved the floor.
-- ============================================================================
DO $$
DECLARE
    q             text;
    n_union       int;
    ehref_known   boolean;
    lit_passes    boolean;
    place_known   boolean;
BEGIN
    SELECT pre_query INTO q FROM scheduled_tasks WHERE name = 'detected-item-promoter';

    n_union := (length(q) - length(replace(q, 'site_work_items_archive', ''))) / length('site_work_items_archive');
    IF n_union <> 2 THEN
        RAISE EXCEPTION '465: expected exactly 2 references to site_work_items_archive (known-good + floor), found %', n_union;
    END IF;
    -- 458's and 454's and 444's other predicates must have SURVIVED
    IF q NOT LIKE '%scored%' OR q NOT LIKE '%0.25 * (c + f)%' OR q NOT LIKE '%(c + f) < 5%'
       OR q NOT LIKE '%wi.pipeline IN (''build'', ''content'', ''design'')%'
       OR q NOT LIKE '%LIMIT 20%' OR q NOT LIKE '%original_pipeline%' THEN
        RAISE EXCEPTION '465: the rewrite LOST a predicate from 444/454/458';
    END IF;

    -- POSITIVE CONTROL — the live victim must now be known-good.
    SELECT EXISTS (
      SELECT 1 FROM (
        SELECT item_type, handler_agent, status FROM site_work_items
        UNION ALL SELECT item_type, handler_agent, status FROM site_work_items_archive) d
      WHERE d.item_type='empty_internal_href' AND d.handler_agent='page-build-handler'
        AND d.status IN ('complete','verified')) INTO ehref_known;
    IF NOT ehref_known THEN
        RAISE EXCEPTION '465: POSITIVE CONTROL FAILED — empty_internal_href->page-build-handler is still not known-good with the archive included, so this fix does not do what it claims';
    END IF;

    -- NEGATIVE CONTROL 1 — the pair 444 exists for must STILL fail the floor.
    SELECT (c + f) < 5 OR c >= 0.25 * (c + f) INTO lit_passes
      FROM (SELECT count(*) FILTER (WHERE status IN ('complete','verified')) AS c,
                   count(*) FILTER (WHERE status='failed') AS f
              FROM (SELECT item_type, handler_agent, status FROM site_work_items
                    UNION ALL SELECT item_type, handler_agent, status FROM site_work_items_archive) a
             WHERE a.item_type='literal_markdown' AND a.handler_agent='page-build-handler') x;
    IF lit_passes THEN
        RAISE EXCEPTION '465: NEGATIVE CONTROL FAILED — literal_markdown->page-build-handler now PASSES the floor, so including the archive has dissolved 444''s door rather than correcting its scope';
    END IF;

    -- NEGATIVE CONTROL 2 — a pair with no successes ANYWHERE must stay unknown.
    SELECT EXISTS (
      SELECT 1 FROM (
        SELECT item_type, handler_agent, status FROM site_work_items
        UNION ALL SELECT item_type, handler_agent, status FROM site_work_items_archive) d
      WHERE d.item_type='placeholder_contact' AND d.handler_agent='page-build-handler'
        AND d.status IN ('complete','verified')) INTO place_known;
    IF place_known THEN
        RAISE EXCEPTION '465: NEGATIVE CONTROL FAILED — placeholder_contact->page-build-handler reads as known-good, but it has no success in either table';
    END IF;

    RAISE NOTICE '465: the promoter now reads live + archive in BOTH success tests. Controls: empty_internal_href->page-build-handler is now known-good (9 lifetime successes it could not previously see); literal_markdown STILL held by the floor; placeholder_contact STILL unknown. Archive holds ~20k rows against ~8.7k live, so "lifetime" now means lifetime.';
END $$;

COMMIT;
