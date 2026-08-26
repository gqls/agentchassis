-- Close this lane's two dead rows for /model-directory.html.
--
-- Both were superseded on 2026-08-25 by the retry pair (item_keys
-- cta_label_relevance_retry: / cta_relink_retry:), which completed and verified at the
-- served bytes at 21:00:09Z. These two cannot ever resolve themselves:
--
--   cta_label_relevance:2c7c836c…  needs_human_review  (refused at validate_content for
--                                  an unregistered "150" that predated this lane)
--   cta_relink:2c7c836c…           triaged, depends_on = [the row above]
--
-- load_work_item_actions.go:713 refuses to dispatch a row whose depends_on is not
-- complete/verified. 'needs_human_review' is neither, so the relink is PERMANENTLY
-- undispatchable — it will sit at triaged for ever, and a future reader querying this
-- lane sees an unfinished pair next to 26 complete ones. That is the bugs_open/396
-- shape (parked rows that are undispatchable, unrefilable and carry no provenance);
-- leaving them is not honesty, it is litter that reads as unfinished work.
--
-- Cancelled, not deleted, and each carries WHY and WHAT REPLACED IT in its own row.
-- No LLM involved: this is a status write, unaffected by the credit outage.

BEGIN;

UPDATE site_work_items
SET status = 'cancelled',
    error = COALESCE(NULLIF(error,'') || ' | ', '') ||
            'SUPERSEDED 2026-08-26 by item_key cta_label_relevance_retry:2c7c836c-98e7-4600-bcbe-8a8f884abcb7, '
            'which completed 2026-08-25 20:57:36Z and was verified at the served bytes. '
            'This row was refused at validate_content for unregistered_number "150" — a claim that '
            'predated this lane and is now rewritten. See bugs_open/391.',
    updated_at = now()
WHERE item_key = 'cta_label_relevance:2c7c836c-98e7-4600-bcbe-8a8f884abcb7'
  AND status = 'needs_human_review';

UPDATE site_work_items
SET status = 'cancelled',
    error = COALESCE(NULLIF(error,'') || ' | ', '') ||
            'SUPERSEDED 2026-08-26 by item_key cta_relink_retry:2c7c836c-98e7-4600-bcbe-8a8f884abcb7, '
            'which completed 2026-08-25 21:00:09Z. This row was permanently undispatchable: its '
            'depends_on named a row that ended needs_human_review, and load_work_item_actions.go:713 '
            'only dispatches on complete/verified. See bugs_open/391.',
    updated_at = now()
WHERE item_key = 'cta_relink:2c7c836c-98e7-4600-bcbe-8a8f884abcb7'
  AND status = 'triaged';

DO $$
DECLARE n_open int; n_cancelled int;
BEGIN
  SELECT count(*) FILTER (WHERE status NOT IN ('complete','verified','cancelled')),
         count(*) FILTER (WHERE status = 'cancelled')
    INTO n_open, n_cancelled
    FROM site_work_items WHERE created_by = 'bugfix_391_cta_relevance';

  IF n_cancelled <> 2 THEN
    RAISE EXCEPTION 'expected exactly 2 cancelled rows, found % (did the keys match?)', n_cancelled;
  END IF;
  IF n_open <> 0 THEN
    RAISE EXCEPTION 'this lane still has % non-terminal row(s); expected 0', n_open;
  END IF;
  RAISE NOTICE 'lane clean: 0 open, 2 cancelled with reasons, the rest complete';
END $$;

COMMIT;
