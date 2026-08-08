# NOTES — bugfix 210 (append-only, newest at the bottom)

## 2026-08-08 — pickup, validation, research

- **Ownership checks before pickup**: `who-owns.py 210` → "OWNED or recently active" naming the
  208 lane — that lane FILED 210 and closed itself 2026-08-08 ("Still genuinely open for a
  future session: bugs_open/210" in its handoff). Live-transcript symbol grep
  (`pageSectionShortfall|pageHasComponents|OrdinarySkipStillStamps|ownedPageSkipReasonPrefix`,
  6h window): one session at 38 hits = the 208 lane's own closing session; its final message
  hands 210 off. Clean.
- **Bug still valid at HEAD** (5c138b279): skip refusal scoped to `OWNED_PAGE_GUARD` prefix
  only (`v3_site_actions.go:665-666`); both scope-pinning tests present.
- **MISSTEP (twice)**: wrote SQL against `agent_error_log` assuming `created_at` (it is
  `occurred_at`) and against `orchestration_states` assuming `id` (it is `orchestration_id`) —
  both without `\d` first, the exact CLAUDE.md "schema first" rule. Cost: two error round-trips.
  Cheap check that would have caught it: `\d <table>` before the first query against any table
  this session. → logged in WRONG_CALLS.md.
- **Live-window measurement**: 0 orchestration_states rows with
  `assembled_page.skipped='true'` — recorded as WEAK (terminal snapshot = last iteration only;
  the bug file already rejected the deployed_at-vs-components proxy as confounded). Frequency
  stays unknown; the fix's own error-log row is the counter.
- **Key discovery — the park item must be a RAW insert**: `writeWorkItem`'s two-strike counts
  the prior dishonestly-`complete` build items under `needs_page:<name>` and would brand the
  escalation `unresolved` at birth → terminal → does not hold `idx_swi_dedup`'s slot → bounds
  nothing. `emitOwnedPageReviewItem` is the sanctioned raw-insert template.
- **Drift found in `loadOpenPageItems`**: excludes 5 statuses where the canonical
  `workItemTerminalStatuses` has 7 — `cancelled` and `unresolved` are treated as OPEN by the
  reconciler while the dedup index frees them. `cancelled` contradicts migration 157's stated
  intent; `unresolved`-as-blocking is load-bearing anti-churn (keep, with comment). Split-brain
  scenario that decided it: a human cancels the park → planner path resumes, reconciler wedges
  forever.
- **Third producer on the `needs_page:` key namespace found live**: `needs_tool_recreation`
  items (mortgagecalculator, `needs_page:tool-overpayment` etc., 3 terminal rows/key). Their
  future inserts are blocked while a park is open for the same page — named as a consumer to
  tell (PLAN §consumers).
- **Corroboration**: my three-consumer measurement of `assemble_page` (config text match)
  agrees with LANDMINES.md:6019's nested `jsonb_path_query` walk — differently-shaped checks,
  so this is real agreement, not two-blind-checks-agreeing.

## 2026-08-08 — implementation

- Council submitted BEFORE implementation: corr `c9647117-3a4b-48a2-b34c-1ea25f4e1f7f`
  (5 edits, validated against the fixPlan schema client-side first). Verdict pending at
  commit time → commit carries `Council-Submitted:`.
- Implemented per PLAN: `page_build_failure_guard.go` (refusal/strike/park/auto-close),
  widened branch in `UpdatePageStatusAction` (owned branch untouched above it),
  `loadOpenPageItems` type+cancelled alignment, three tests replacing the scope pin.
- **All three mutations killed their named tests** (M1 skip-branch disabled →
  OrdinarySkipRefusesStamp failed on the unexpected stamp UPDATE; M2 park call removed →
  ThirdRefusalParksThePage failed on unmet park-INSERT expectation; M3 auto-close removed →
  SuccessfulStampClosesPark failed on unmet close-UPDATE expectation). Reverted; full
  `go test ./platform/orchestration/...` green, exit status checked via PIPESTATUS
  (the `| head && echo OK` landmine appended to LANDMINES the same day by another lane).
- Register: PBP-038 added; PBP-036's now-stale "keyed to THIS skip" bullet corrected
  visibly (strike-through + date), per the stale-status landmine.
- LANDMINES: two entries appended + `landmines-sync.py --apply` run (1350 rows owned).
- Consumers told by append-only notes: mortgagecalculator lane (ACTIVE session — used
  `cat >>` append rather than a whole-file Write to be collision-safe), feature_021 lane,
  208 lane handoff pointer.
