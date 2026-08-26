-- ROLLBACK for 634 — remove `required_fields_missing` from the claimed-item-timeout
-- exclusion list. bugs_open/375.
--
-- ⚠ DANGEROUS ONCE **ANY** GATE EXISTS FOR THE TYPE — not only a registered verifier. A
-- noChangeGates entry or a REFUSING acceptancePredicateGates entry would equally be bypassed
-- by the resumed sweep (Fable review 2026-08-26 widened this from the original wording).
-- ⚠ AND ITS POST-CHECK COUNTS THE BARE WORD: if pre_query ever mentions
-- required_fields_missing in a COMMENT, this rollback aborts AFTER its UPDATE, every time —
-- fail-safe (the whole DO is one statement, so the UPDATE rolls back too), but permanently
-- un-runnable until the comment is reworded.
-- ⚠ ONLY SAFE WHILE THE GO SIDE HAS NOT LANDED. Once
-- livespec.ClaimedItemTimeoutExclusions carries `required_fields_missing` AND a verifier is
-- registered for it, rolling this back re-opens the hole this migration exists to close: the
-- claimed-item-timeout sweep resumes auto-completing those items straight past the verifier,
-- silently, and every Go-side test stays green because they assert the DECLARATION, not the
-- live object. If you are rolling back after the Go commit, revert that commit FIRST.

DO $$
DECLARE
  -- ')' load-bearing here too (Fable review 2026-08-26): without it, a clause that another
  -- lane has extended past the 634 tail still prefix-matches and this rollback runs on a
  -- moved clause it promised to abort on.
  new_tail text := '''needs_brand_head_assets'', ''dark_section_audit'', ''required_fields_missing'')';
  old_tail text := '''needs_brand_head_assets'', ''dark_section_audit'')';
  n int;
BEGIN
  SELECT count(*) INTO n FROM scheduled_tasks
   WHERE name = 'claimed-item-timeout' AND strpos(pre_query, new_tail) > 0;
  IF n <> 1 THEN
    RAISE EXCEPTION 'ABORT: the 634 tail is not present (matched % rows, want 1) — nothing to roll back, or the clause has moved since.', n;
  END IF;

  UPDATE scheduled_tasks
     SET pre_query = replace(pre_query, new_tail, old_tail)
   WHERE name = 'claimed-item-timeout';

  SELECT count(*) INTO n FROM scheduled_tasks
   WHERE name = 'claimed-item-timeout' AND strpos(pre_query, 'required_fields_missing') > 0;
  IF n <> 0 THEN
    RAISE EXCEPTION 'ABORT: rollback verification failed — required_fields_missing still present.';
  END IF;

  RAISE NOTICE '634 rolled back: claimed-item-timeout excludes 14 item types again.';
END $$;
