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
