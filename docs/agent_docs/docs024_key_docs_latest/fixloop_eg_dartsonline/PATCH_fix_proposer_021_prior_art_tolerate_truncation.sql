-- PATCH_fix_proposer_021_prior_art_tolerate_truncation.sql
-- Close the last hole in the bugs_open/019 fix. 2026-07-20. clients_db.
--
-- WHY (found 2026-07-20 by inspecting the live rows after the 019 fix shipped):
-- a3b606798 gave every reviewer step an opt-in `tolerate_truncation` so a
-- truncated seat degrades instead of voiding the whole council round. It was
-- applied to the seats that existed at the time. `review_prior_art` — the 16th
-- seat, added as an always-on librarian by 3cf14d429 — does NOT have it, on
-- BOTH councils:
--
--   fix-proposer: 16 seats, missing flag: ['review_prior_art']
--   council-gate: 16 seats, missing flag: ['review_prior_art']
--
-- Identical on both, so this is NOT gate mirror drift — it is a genuine
-- omission on the seat itself, and the mirror faithfully copied the hole.
--
-- SEVERITY IS THE POINT: `review_prior_art` has no `gate_` step and does not
-- appear in `select_panel`, i.e. it is ALWAYS-ON — it runs on EVERY round, at
-- max_tokens 8000 like every other seat. So the one seat still able to void a
-- round is the one seat guaranteed to run. 019 reads as fixed (15/16) and is
-- not fixed in practice.
--
-- THE SHAPE, recorded because it is now the fourth instance in two days:
-- a fix applied to the members that existed, then the roster grew and the new
-- arrival did not get it. Same as bugs_open/016 finding 2 (seats grew, reviser
-- kept a hand-written list) and as the withPriorRequests/withPriorCodeRequests
-- twin (016b section 9 #26). The class is "a fix whose scope is a snapshot of a
-- growing set". `scripts/pattern-check.py` catches the Go version of this; it
-- cannot see agent config, which is where this one lived.
--
-- ORDER MATTERS (CLAUDE.md): seat fix-proposer, then mirror to the gate with
-- 099_SYNC_gate_roster.py --apply. Do NOT hand-patch council-gate — two
-- hand-maintained rosters that must stay identical is exactly the drift class
-- this council reviews for.
--
-- SURGICAL + IDEMPOTENT: one jsonb_set on one path; re-running is a no-op
-- because the value written is a fixed `true`. Snapshot first.

BEGIN;

SELECT snapshot_agent('fix-proposer', 'pre-update: 021 — review_prior_art gains tolerate_truncation (last seat missing the 019 fix)');

UPDATE agent_definitions SET default_config = jsonb_set(
  default_config, '{workflow,steps,review_prior_art,config,tolerate_truncation}', 'true'::jsonb, true)
WHERE type='fix-proposer' AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;

-- Verify inside the transaction: every review_ seat must now carry the flag.
SELECT count(*) FILTER (WHERE NOT flagged) AS seats_still_missing_flag,
       count(*)                            AS review_seats
FROM (
  SELECT COALESCE((default_config#>'{workflow,steps}'->k->'config'->>'tolerate_truncation')::boolean, false) AS flagged
  FROM agent_definitions,
       LATERAL jsonb_object_keys(default_config->'workflow'->'steps') k
  WHERE type='fix-proposer' AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL
    AND k LIKE 'review\_%'
) s;

COMMIT;
