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
--
-- ⚠ WHY strpos() AND NOT LIKE — a defect the council's debug_historian seat caught before this
-- was applied, and it is the kind that only bites when it matters. Every needle here is full of
-- UNDERSCORES ('needs_brand_head_assets', 'required_fields_missing'), and in a LIKE pattern `_`
-- is a SINGLE-CHARACTER WILDCARD. So `pre_query LIKE '%needs_brand_head_assets%'` does not assert
-- that the exact text is present — it asserts a pattern match. Proven against the live database
-- 2026-08-26: `SELECT 'XneedsXbrandXheadXassetsX' LIKE '%needs_brand_head_assets%'` returns TRUE
-- while `strpos('XneedsXbrandXheadXassetsX','needs_brand_head_assets') > 0` correctly returns
-- FALSE. A read-before-write anchor that can match text it was not written against is not an
-- anchor, and an idempotence guard that can false-positive would abort a legitimate first apply.
-- strpos() is exact. The occurrence count below already used replace(), which is exact too.
--
--
-- ── THE CLASS AUDIT the council's bug_historian seat asked for [MEASURED 2026-08-26] ────
--
-- The objection is right that the mechanism is GENERIC: this sweep bypasses every completion
-- gate for any type not on this list, so excluding one type is "a" fix, not "the" fix. The
-- class answer is that the class is ALREADY guarded, at build time, in both directions:
-- TestClaimTimeoutExclusionCoversBothCompletionGates enforces
--     excluded ⇔ (a registered verifier) OR (a noChangeGates entry) OR (a REFUSING
--                 acceptancePredicateGates entry)
-- and it PASSES at HEAD. So no type that any gate can currently refuse is unexcluded.
--
-- Audited against the LIVE clause rather than the Go declaration, because the objection is
-- about live exposure: **15** gate memberships (13 verifiers + noChangeGates:dark_section_audit
-- + acceptancePredicateGates:content_rewrite — an earlier draft said 16, a hand-census
-- over-count) vs 14 live exclusions, and the differences are accounted for —
--   * required_fields_missing — on NO roster yet (its verifier is deliberately unregistered),
--     so it is a pre-declared exclusion this migration adds AHEAD of its gate — the safe order;
--     the Go commit that follows closes both halves together.
--   * content_rewrite — has a gate-1c entry that does NOT yet refuse, so it is correctly
--     NOT excluded; the lockstep will demand it the moment bugs_open/395 flips that gate to
--     predicateRefuses. That is their step, anchored on this migration's tail.
-- Zero types are excluded-but-ungated, so there is no reverse-arm churn risk either.
--
-- ⚠ A note on how that audit was run, because the first attempt was wrong: re-deriving the
-- roster by grepping the three maps OVER-COUNTS, since it cannot see whether a gate-1c entry
-- actually refuses. The lockstep is the authority; a hand-rolled census of the same rosters
-- answers a slightly different question and reported content_rewrite as an open bypass.
--
-- ⚠ THE GO ENTRY MUST BE APPENDED AT THE **END** OF ClaimedItemTimeoutExclusions — POSITION IS
-- LOAD-BEARING. Verified by construction (Fable review, 2026-08-26): appended LAST, the Go
-- rendering is BYTE-IDENTICAL to the post-634 live clause and the drift auditor is clean.
-- Inserted anywhere else — alphabetically, say — every build-time test stays green (the
-- lockstep is set-based and the round-trip test is order-blind) and the daily auditor fires
-- for ever on a semantically correct clause.
--
-- ⚠ AND THE WINDOW DOES NOT CLOSE AT THE GO COMMIT. The drift auditor runs from declarations
-- compiled into the live-declaration-drift-check IMAGE (tag-pinned in its kustomization). The
-- window closes at image rebuild + tag bump + apply. Plan for days of announced red at 07:00
-- UTC, not minutes, unless that rebuild is part of the same sitting.
--
-- (Deliberately NOT fixed here: the live pre_query's own comment still cites the retired
-- "LOCKSTEP TWIN" contract and a deleted test. Correcting prose was not in the council-approved
-- scope of this migration, and widening an approved migration post-verdict is the shape the
-- council exists to catch. Recorded as owed in bugs_open/375's lane notes instead.)
--
-- ROLLBACK SIDECAR: 634_claim_timeout_exclude_required_fields_missing_ROLLBACK.sql

DO $$
DECLARE
  -- Anchor on the TAIL only, deliberately: the opening `item_type NOT IN (` prefix is
  -- shared with other predicates in this column, and anchoring on the whole clause would
  -- make this file's text a second declaration of the list (see the 482 note above).
  -- ⚠ THE CLOSING ')' IS LOAD-BEARING — a Fable adversarial review (2026-08-26) proved by
  -- construction that without it this anchor is a PREFIX, not a tail: with another lane's
  -- amendment already appended (exactly what bugs_open/395 plans), strpos still matched, the
  -- migration applied silently MID-LIST, every guard passed — and the Go renderer could then
  -- never reproduce the live clause, putting the drift auditor into a permanent ownerless red.
  -- With the ')', a moved clause ABORTS at guard 1 instead. Reproduced against the live DB
  -- before fixing: prefix-anchor matched the moved clause (t), ')'-anchor refused (f).
  old_tail text := '''needs_brand_head_assets'', ''dark_section_audit'')';
  new_tail text := '''needs_brand_head_assets'', ''dark_section_audit'', ''required_fields_missing'')';
  n int;
BEGIN
  -- READ BEFORE WRITE. The live column is the fact; the repo file is history. Abort if the
  -- predicate is not the shape this migration was written against, rather than half-applying
  -- to something that has moved underneath us.
  SELECT count(*) INTO n FROM scheduled_tasks
   WHERE name = 'claimed-item-timeout' AND strpos(pre_query, old_tail) > 0;
  IF n <> 1 THEN
    RAISE EXCEPTION 'ABORT: claimed-item-timeout pre_query does not carry the expected exclusion tail (matched % rows, want 1). Re-read the live clause and re-anchor; do NOT edit this file to match a moved target without re-reading why it moved.', n;
  END IF;

  -- Idempotence: refuse a second application rather than appending a duplicate entry.
  -- A duplicate would render as a clause the Go slice can never produce, so the drift
  -- auditor would fire for ever on a change that looked correct.
  SELECT count(*) INTO n FROM scheduled_tasks
   WHERE name = 'claimed-item-timeout' AND strpos(pre_query, 'required_fields_missing') > 0;
  IF n <> 0 THEN
    RAISE EXCEPTION 'ABORT: required_fields_missing is already excluded — this migration has been applied.';
  END IF;

  UPDATE scheduled_tasks
     SET pre_query = replace(pre_query, old_tail, new_tail)
   WHERE name = 'claimed-item-timeout';

  -- VERIFY, and RAISE rather than SELECT: ON_ERROR_STOP ignores a non-empty result set, so a
  -- verification block built from SELECTs cannot stop the COMMIT (LANDMINES.md, RFC_006).
  SELECT count(*) INTO n FROM scheduled_tasks
   WHERE name = 'claimed-item-timeout' AND strpos(pre_query, new_tail) > 0;
  IF n <> 1 THEN
    RAISE EXCEPTION 'ABORT: post-write verification failed — the exclusion list does not carry required_fields_missing (matched % rows, want 1).', n;
  END IF;

  -- (Guard 4 below counts the QUOTED needle, not the tail. That is equivalent to "added
  -- exactly once" only because guard 2 established the bare-word count was ZERO pre-write —
  -- stated so the equivalence is an argument, not an accident.)

  -- ⚠ replace() is GLOBAL. Assert the tail appears exactly once, or a second occurrence
  -- elsewhere in this column would have been rewritten silently.
  SELECT count(*) INTO n FROM scheduled_tasks
   WHERE name = 'claimed-item-timeout'
     AND (length(pre_query) - length(replace(pre_query, '''required_fields_missing''', ''))) / length('''required_fields_missing''') <> 1;
  IF n <> 0 THEN
    RAISE EXCEPTION 'ABORT: required_fields_missing appears more than once in pre_query — replace() hit an unintended second occurrence.';
  END IF;

  RAISE NOTICE '634 applied: claimed-item-timeout now excludes 15 item types (required_fields_missing added). THE DRIFT AUDITOR WILL NOW FIRE — expected, bugs_open/375s window. ⚠ The Go commit alone does NOT close it: the daily auditor compares this column to declarations COMPILED INTO the live-declaration-drift-check image, so the window closes at that image REBUILD + TAG BUMP + apply (its kustomization says bump in the same commit as the rebuild). Until then the 07:00 UTC job is red daily; announce it or it gets helpfully reverted.';
END $$;
