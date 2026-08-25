-- 634 — claimed-item-timeout: exclude `required_fields_missing`, so the sweep cannot
-- complete those items past the verifier that is about to be registered for them.
-- bugs_open/375 (WII-032 step 1 of 5).
--
-- ── WHAT THE PROBLEM IS, in plain terms ──────────────────────────────────────────
--
-- A work item is one recorded defect on one site. A VERIFIER re-runs that defect's own
-- predicate immediately before the item is stamped `complete`, and refuses the stamp if
-- the defect is still there.
--
-- The scheduled task `claimed-item-timeout` auto-completes an item whose claim has timed
-- out by writing `site_work_items` DIRECTLY. It goes through no completion action, so
-- NEITHER completion gate runs for it — not the verifier, not the no-change gate, not the
-- acceptance-predicate gate. Its only protection is the `item_type NOT IN (...)` list in
-- its own `pre_query`, which is what this migration amends.
--
-- ── WHY THIS MUST APPLY *BEFORE* THE GO COMMIT, AND NOT AFTER ────────────────────
--
-- bugs_open/375 is registering a verifier for `required_fields_missing`. The moment that
-- verifier is registered, this sweep becomes a completion path that bypasses it — which is
-- bugs_closed/317 reintroduced BY THE ACT of adding a guard. A false green is worse than a
-- missing one.
--
-- Both orders make `cmd/config-key-audit --live-declaration-drift` fire, so the drift is
-- NOT the discriminator between them (treating it as one is what makes the wrong order look
-- defensible). The discriminator is what the SWEEP does inside each window:
--
--   THIS ORDER (migration first): the live clause excludes the type while the Go slice does
--   not. The sweep stops auto-completing those items immediately. Drift is noisy; the estate
--   is SAFER than before. The window is pure noise.
--
--   THE OTHER ORDER (Go first): the Go slice says excluded, the live clause does not, and the
--   sweep is STILL completing items straight past the verifier — while every instrument says
--   it has already been fixed.
--
-- Identical noise; only one window contains damage. (Reasoned independently by the
-- bugs_open/395 lane, which needs the same clause amended for `content_rewrite`.)
--
-- ⚠ ANNOUNCE THE WINDOW. While this is applied and the Go slice is not, ANY unrelated
-- session running --live-declaration-drift sees a RED result that is not theirs and is not a
-- defect. A drift failure with no owner is exactly the shape that gets "helpfully fixed" by
-- somebody reverting this migration. Keep the window short (have the Go commit ready before
-- applying) and say what you are doing.
--
-- ── WHY _HOLD ────────────────────────────────────────────────────────────────────
--
-- Held from the runner deliberately (SIDECAR_RE excludes an UPPERCASE suffix), because the
-- window above must be opened at a moment somebody is watching, not whenever the next
-- unattended `--apply` happens to sweep the directory. Apply by hand:
--   kubectl -n ai-persona-system exec -i postgres-clients-0 -- \
--     psql -U clients_user -d clients_db -v ON_ERROR_STOP=1 \
--     -f - < docs/agent_docs/sql_for_agents/634_claim_timeout_exclude_required_fields_missing_HOLD.sql
--
-- ── TWO THINGS 482'S HEADER WILL TELL YOU THAT ARE NO LONGER TRUE ────────────────
--
-- 482 says "THIS FILE IS ALSO A DECLARATION, NOT ONLY A MIGRATION", because a parity test
-- once globbed `*_claimed_item_timeout_generic_evidence.sql`, took the newest match and
-- parsed the clause out of it. **That contract is GONE** — bugs_open/363 phase 1 moved the
-- declaration into `platform/livespec` (commit 873575ecf), and the only surviving mention of
-- the glob is the historical narrative in claim_timeout_exclusion_lockstep_test.go. Verified
-- 2026-08-25: no Go file parses this filename pattern. So this file is named normally, and it
-- does NOT need to restate the list for a reader that no longer exists.
--
-- What DOES read the list now: `livespec.ClaimedItemTimeoutExclusions`, rendered by
-- `ClaimedItemTimeoutExclusionClause()`, tied to this live column by
-- `livespec.Declarations["scheduled_task.claimed-item-timeout.exclusions"]` as a FragmentMatch
-- with Min:1/Max:1, audited daily by the CronJob `live-declaration-drift-check`.
--
-- ⚠ COMPOSITION, and it is why a second lane must not write its own copy of this against
-- today's clause. That fragment is the WHOLE `item_type NOT IN (...)` string. Two migrations
-- each written against the 14-type clause do NOT compose: whichever applies second will not
-- find its anchor, and if both somehow applied, the Go renderer would produce a string that
-- matches neither. bugs_open/395 owes `content_rewrite` on this same clause and has agreed to
-- anchor its amendment on THIS migration's tail.
--
-- ROLLBACK SIDECAR: 634_claim_timeout_exclude_required_fields_missing_ROLLBACK.sql

DO $$
DECLARE
  -- Anchor on the TAIL only, deliberately: the opening `item_type NOT IN (` prefix is
  -- shared with other predicates in this column, and anchoring on the whole clause would
  -- make this file's text a second declaration of the list (see the 482 note above).
  old_tail text := '''needs_brand_head_assets'', ''dark_section_audit''';
  new_tail text := '''needs_brand_head_assets'', ''dark_section_audit'', ''required_fields_missing''';
  n int;
BEGIN
  -- READ BEFORE WRITE. The live column is the fact; the repo file is history. Abort if the
  -- predicate is not the shape this migration was written against, rather than half-applying
  -- to something that has moved underneath us.
  SELECT count(*) INTO n FROM scheduled_tasks
   WHERE name = 'claimed-item-timeout' AND pre_query LIKE '%' || old_tail || '%';
  IF n <> 1 THEN
    RAISE EXCEPTION 'ABORT: claimed-item-timeout pre_query does not carry the expected exclusion tail (matched % rows, want 1). Re-read the live clause and re-anchor; do NOT edit this file to match a moved target without re-reading why it moved.', n;
  END IF;

  -- Idempotence: refuse a second application rather than appending a duplicate entry.
  -- A duplicate would render as a clause the Go slice can never produce, so the drift
  -- auditor would fire for ever on a change that looked correct.
  SELECT count(*) INTO n FROM scheduled_tasks
   WHERE name = 'claimed-item-timeout' AND pre_query LIKE '%required_fields_missing%';
  IF n <> 0 THEN
    RAISE EXCEPTION 'ABORT: required_fields_missing is already excluded — this migration has been applied.';
  END IF;

  UPDATE scheduled_tasks
     SET pre_query = replace(pre_query, old_tail, new_tail)
   WHERE name = 'claimed-item-timeout';

  -- VERIFY, and RAISE rather than SELECT: ON_ERROR_STOP ignores a non-empty result set, so a
  -- verification block built from SELECTs cannot stop the COMMIT (LANDMINES.md, RFC_006).
  SELECT count(*) INTO n FROM scheduled_tasks
   WHERE name = 'claimed-item-timeout'
     AND pre_query LIKE '%''dark_section_audit'', ''required_fields_missing''%';
  IF n <> 1 THEN
    RAISE EXCEPTION 'ABORT: post-write verification failed — the exclusion list does not carry required_fields_missing (matched % rows, want 1).', n;
  END IF;

  -- ⚠ replace() is GLOBAL. Assert the tail appears exactly once, or a second occurrence
  -- elsewhere in this column would have been rewritten silently.
  SELECT count(*) INTO n FROM scheduled_tasks
   WHERE name = 'claimed-item-timeout'
     AND (length(pre_query) - length(replace(pre_query, '''required_fields_missing''', ''))) / length('''required_fields_missing''') <> 1;
  IF n <> 0 THEN
    RAISE EXCEPTION 'ABORT: required_fields_missing appears more than once in pre_query — replace() hit an unintended second occurrence.';
  END IF;

  RAISE NOTICE '634 applied: claimed-item-timeout now excludes 15 item types (required_fields_missing added). THE DRIFT AUDITOR WILL NOW FIRE until livespec.ClaimedItemTimeoutExclusions carries the same entry — that is expected, it is bugs_open/375s window, and the Go commit closes it.';
END $$;
