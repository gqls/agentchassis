-- 377_relock_six_verbatim_tool_pages_missed_by_367_ROLLBACK.sql
--
-- Re-unlocks EXACTLY the rows 377 stamped, restoring each page's `prev_policy`.
-- Scoped to the stamp table rather than to the two domains, because another
-- thread may legitimately have flipped a page since 377 ran and a domain-wide
-- revert would silently take that with it.
--
-- ⛔ READ THIS BEFORE RUNNING IT. 377 exists because these six pages are single
-- verbatim `page_components` rows holding a live consumer-finance calculator, and
-- `rebuild_policy='owned'` is the only thing standing between them and a generic
-- rebuild that replaces the calculator with prose and ships it (TL-001, migration
-- 164). Running this rollback re-opens that door for all six at once.
--
-- The legitimate way to make one of these pages generic is per page, after it has
-- been decomposed and its tool row locked — Track B of
-- docs/agent_docs/docs024_key_docs_latest/loanandmortgagecalculator_couk/HANDOFF_2026-08-10c_continue_here.md
-- A per-page flip is a new forward migration, not this file.
--
-- If you are running this because decomposition is finished for all six, the
-- verify block below will still pass — it only checks that the policy moved. It
-- deliberately does NOT check that the pages are safe to unlock, because a
-- rollback that silently refuses is worse than one that is honest about its
-- scope. Check the shape yourself first:
--   SELECT url, sections::text FROM pages WHERE id IN
--     (SELECT page_id FROM _mig377_relocked_tool_pages);
-- Any row still reading ["ported-page"] is NOT ready.

BEGIN;

UPDATE pages p
   SET rebuild_policy = m.prev_policy,
       updated_at = now()
  FROM _mig377_relocked_tool_pages m
 WHERE p.id = m.page_id
   AND COALESCE(p.rebuild_policy, 'generic') = 'owned';

DO $$
DECLARE
    n_stamped int;
    n_wrong   int;
BEGIN
    SELECT count(*) INTO n_stamped FROM _mig377_relocked_tool_pages;
    SELECT count(*) INTO n_wrong
      FROM pages p JOIN _mig377_relocked_tool_pages m ON m.page_id = p.id
     WHERE COALESCE(p.rebuild_policy, 'generic') <> m.prev_policy;
    IF n_wrong <> 0 THEN
        RAISE EXCEPTION 'mig377 rollback: % of % stamped pages did not return to '
                        'their prev_policy', n_wrong, n_stamped;
    END IF;
    RAISE NOTICE 'mig377 rollback OK: % page(s) returned to their pre-377 policy. '
                 'Those pages are now rebuildable by the generic pipeline again.',
                 n_stamped;
END $$;

DROP TABLE IF EXISTS _mig377_relocked_tool_pages;

COMMIT;
