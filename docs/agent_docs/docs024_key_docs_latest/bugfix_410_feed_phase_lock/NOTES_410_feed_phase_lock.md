# NOTES — bugfix_410_feed_phase_lock (append-only, newest at bottom)

## 2026-08-26 ~17:30–18:00Z — pickup, verification at HEAD, fix built

- Ambiguity resolved first: TWO bugs named 410. Messaged the other `bugs_open/410` session
  (scan-loss lane, `[0bb588]`) — no overlap; messaged `idea.uk [e14ec9]` (the filer) — handoff
  confirmed, three refinements received (recorded in PLAN: (c) refutation criterion needs the
  trigger's actual fire time; I own recording (c)/(d); roll-before-apply sequencing; plus:
  316 has no active owner, and remortgagecalculator.uk's 13:43Z run must be excluded from any
  census).
- Queue check: no open `site_work_items` row covers the cadence defect → fired `090`
  17:31Z, RUN_CORRELATION_ID `15d56c13-2081-431a-ad70-9516c5fcfbc7`. [verdict pending]
- **Bug still valid at HEAD, verified first-hand:** both stamp arms unchanged
  (`dispatch_feed_sources_action.go` optimistic stamp; `feed_actions.go`
  `UpdateSourceTimestampsAction` success+failure arms); live `find_news_sites` query read
  from `agent_definitions` — byte-identical to migration 556's post-image, bare
  `cs.next_fetch_at <= NOW()`; `scheduled_tasks.content-feed-refresh` = 21600 s, enabled,
  last fired 14:46:30Z.
- **CORRECTION to the bug file's reader census:** live workflow uses `dispatch_feed_sources`
  (its own due query) — `LoadDueSourcesAction` (`feed_actions.go`, the file's cited "selector
  the orchestrator itself uses") has **zero live workflow callers** (`jsonb_each` over all
  active agent_definitions steps: only `feed-ingester.update_timestamps` uses
  `update_source_timestamps`; nothing uses `load_due_sources`). Mechanism unaffected — both
  carry the same predicate — but the fix target had to be the dispatcher.
- **New evidence, source-level lock inside admitted sites:** dartsonline's 6 h sources (last
  fetched 08:51Z, due 14:51:09–28Z) missed the 14:50:27Z dispatch by ~40–60 s while its 4 h
  sources were fetched — so multi-interval "control" sites are served per-pass at SITE level
  while their 6 h SOURCES still run 12-hourly. The source-level half of the fix is
  load-bearing.
- Detector check before changing the predicate: `cmd/config-key-audit/cappedscheduleordering.go`
  `dueColumnRe` requires `<op> NOW()` — `<= NOW() + COALESCE(...)` still matches, so the 316
  audit is not blinded by the look-ahead. (Checked the regex, not assumed.)
- Fix built: shared `feedSourceDuePredicate` const + both Go queries + migration 653_HOLD
  (556 idiom: snapshot, two DO/RAISE guards — cadence-equals-21600 and pre-image match —
  jsonb_set, post-verify).
- **Mutations executed, not claimed** (all against `go test -run Lookahead`):
  1. dispatcher reverted to bare NOW() → `TestDispatchFeedSourcesQueriesWithTheDueLookahead` RED;
  2. const halving changed 2.0→1.0 → `TestFeedDueLookaheadShape` RED;
  3. LoadDueSources reverted to bare NOW() → `TestLoadDueSourcesQueriesWithTheDueLookahead` RED;
  restored → 3/3 PASS; full actions suite `ok ... 5.411s`.
- Misstep (WRONG_CALLS row added): queried `scheduled_tasks` twice with guessed column names
  (`schedule_interval_seconds`, then `schedule`) before `\d scheduled_tasks`. The schema-first
  rule exists; typing `\d` first was cheaper than either failure.

## 2026-08-26 ~18:4xZ — council APPROVED round 1; the advisory objection re-run and discharged with a STRONGER census

- **Verdict: APPROVED, round 1, 1 advisory objection, none high** (corr `04c657d2`, decided_by
  "approved with 1 advisory objection(s)"; 8 seats abstained — relevance gating working).
  Commit `201236b2a` already carries `Council-Submitted:`, which 098 auto-credits — nothing to
  amend, forward-only holds.
- **The objection (prior_art_librarian, medium): my "LoadDueSourcesAction has zero live
  workflow callers" was asserted without its query in `grounded_in`.** Fair — and its
  hypothetical exposed a real weakness: my original census (`jsonb_each` over
  `default_config->'workflow'->'steps'`) walks TOP-LEVEL steps only, so a nested caller
  inside a loop step would have been invisible (the estate's own detectors use
  `validation.WalkSteps` "top level and nested" for exactly this reason). Re-ran as a full
  config-text search, which cannot miss nesting:
  ```sql
  SELECT type FROM agent_definitions WHERE is_active AND COALESCE(is_snapshot,false)=false
    AND deleted_at IS NULL AND default_config::text LIKE '%load_due_sources%';   -- 0 rows
  SELECT type FROM agent_definitions WHERE default_config::text LIKE '%load_due_sources%'; -- 0 rows (snapshots + inactive too)
  SELECT name FROM scheduled_tasks WHERE input_data::text LIKE '%load_due_sources%'
    OR COALESCE(pre_query,'') LIKE '%load_due_sources%';                          -- 0 rows
  ```
  `[MEASURED 2026-08-26 ~18:4xZ]` The claim was true, but the original check could not have
  proven it — lesson: **an absence census over workflow steps must search the whole config
  text (or WalkSteps), never top-level `jsonb_each`.**
- The second objection (low — "byte-identical to 556's post-image" asserted from my own live
  read): discharged by construction — migration 653's guard 2 re-checks that identity
  against the LIVE row at apply time with DO/RAISE, which is the remedy the objection asks
  for; and the pre-image was cross-checked against `556_..._ROLLBACK.sql`'s `$post$` string
  in-session before writing the guard.
