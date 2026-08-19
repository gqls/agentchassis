-- 482 — claimed-item-timeout: the exclusion list must cover BOTH completion gates,
-- not just the verifier registry. bugs_open/317.
--
-- WHY. The sweep auto-completes a `claimed` item past its timeout by writing the row
-- directly, so NEITHER completion gate runs — not gate 2 (the verifier registry) and
-- not gate 1b (`noChangeGates`, complete_work_item_no_change.go). Its protection is
-- this item_type exclusion list, and migration 220's own comment states the contract
-- it was written to: "the LOCKSTEP TWIN of the RegisterVerifier() calls". That was
-- correct when gate 2 was the only gate. Since 2026-08-13 there is a second gate with
-- its own opt-in roster, and a type on THAT roster with no registered verifier was not
-- excluded — so for it the sweep is a completion path no gate can see.
--
-- `dark_section_audit` is exactly that type, and it is the only one today: it carries a
-- noChangeGates entry (and, since bugs_closed/302, an `unreadableRefuses` declaration)
-- and has no verifier, deliberately — verifier_coverage_test.go classifies its family as
-- needing a browser on the completion path, which is the standing objection this estate
-- has refused three times.
--
-- ⚠ THIS FILE IS ALSO A DECLARATION, NOT ONLY A MIGRATION. The parity test reads the
-- NEWEST file matching *_claimed_item_timeout_generic_evidence.sql and parses the first
-- that clause out of it. So the new list appears below in full, and the
-- pre-write anchor deliberately omits that clause's opening prefix so it cannot be
-- parsed as the declaration. 220 is left untouched: it is applied history, and editing
-- it would make its recorded checksum a lie.
--
-- The resulting predicate fragment, which is what the test reads:
--   item_type NOT IN ('truncated_component', 'hardcoded_section_colors', 'empty_section', 'orphan_element_refs', 'content_duplication', 'page_canonical_collision', 'dead_fragment_link', 'literal_markdown', 'unbuilt_internal_link', 'revenue_shape_cta', 'missing_conversion_path', 'decision_regression', 'needs_brand_head_assets', 'dark_section_audit')
--
-- ROLLBACK SIDECAR: 482_ROLLBACK_claim_timeout_exclusion.sql

DO $$
DECLARE
  -- Anchor WITHOUT that clause's opening prefix, on purpose — see the note above.
  old_tail text := '''decision_regression'', ''needs_brand_head_assets''';
  new_tail text := '''decision_regression'', ''needs_brand_head_assets'', ''dark_section_audit''';
  n int;
BEGIN
  -- READ BEFORE WRITE. The live column is the fact; the repo file is history. Abort if
  -- the predicate is not the shape this migration was written against, rather than
  -- half-applying to something that has moved underneath us.
  SELECT count(*) INTO n FROM scheduled_tasks
   WHERE name = 'claimed-item-timeout' AND pre_query LIKE '%' || old_tail || '%';
  IF n <> 1 THEN
    RAISE EXCEPTION 'ABORT: claimed-item-timeout pre_query does not carry the expected exclusion tail (matched % rows, want 1). The predicate has changed since this migration was written — re-read the live column and re-derive the anchor.', n;
  END IF;

  -- Idempotence: refuse a second application rather than appending a duplicate entry.
  SELECT count(*) INTO n FROM scheduled_tasks
   WHERE name = 'claimed-item-timeout' AND pre_query LIKE '%dark_section_audit%';
  IF n <> 0 THEN
    RAISE EXCEPTION 'ABORT: dark_section_audit is already excluded — this migration has been applied.';
  END IF;

  UPDATE scheduled_tasks
     SET pre_query = replace(pre_query, old_tail, new_tail)
   WHERE name = 'claimed-item-timeout';

  -- VERIFY, and RAISE rather than SELECT: ON_ERROR_STOP ignores a non-empty result set,
  -- so a verification block built from SELECTs cannot stop the COMMIT (LANDMINES).
  SELECT count(*) INTO n FROM scheduled_tasks
   WHERE name = 'claimed-item-timeout'
     AND pre_query LIKE '%''needs_brand_head_assets'', ''dark_section_audit''%';
  IF n <> 1 THEN
    RAISE EXCEPTION 'ABORT: post-write verification failed — the exclusion list does not carry dark_section_audit (matched % rows).', n;
  END IF;

  RAISE NOTICE 'claimed-item-timeout: exclusion list now covers both completion gates (14 item types).';
END $$;
