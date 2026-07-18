# 017 — fix_forced_text_colors is never registered; its WORKFLOW_INVALID failures are stamped 'complete'

**Filed:** 2026-07-18, from the robot-hands R1 thread
(`docs/agent_docs/docs024_key_docs_latest/robot_hands/RUNNING_NOTES_robot_hands_site_fixes.md`).
Two coupled defects, one incident. Neither is fixed.

## Symptom

`hardcoded_section_colors` work items on robot-hands "completed" twice
(2026-07-16 14:25:51, 2026-07-17 13:32:26) without touching anything. Their
`result` records the truth:

```
result->'response'->>'status' = 'failed'
result->'response'->>'error'  = 'WORKFLOW_INVALID: Invalid workflow
  configuration (caused by: step ''fix_text_colors'' with action
  ''fix_forced_text_colors'' requires a topic)'
```

Items: `b530e129-d1f3-4182-bdd6-326c354ff784`,
`e4fd567e-fdae-45ed-a254-b786d389500a` (site
`00ff3af5-dad8-4770-9f70-3edc267a3c92`), both `status='complete'`,
`handled_by='build-dispatch-loop'`.

## Defect 1 — action written but never registered (the "requires a topic" is a lie)

- `platform/orchestration/actions/fix_forced_text_colours_action.go` defines
  `FixForcedTextColorsAction` + registers an ActionInputSpec (line 56) — and
  that is ALL. The action appears in **neither**
  `actions/registry.go` **nor** `actioncheck/local_actions.go`
  (grep both for `"fix_forced_text_colors"`: zero hits).
- Workflow validation (`platform/validation/workflow.go:69,80`) consults
  `actioncheck.IsLocalAction` — the hand-maintained list in
  `actioncheck/local_actions.go`, whose own header says *"also update registry
  with new actions // DEPRICATED"*. Unknown action → "remote" → demands a
  `topic` → `WORKFLOW_INVALID` on every run of the `color-variable-fixer`
  agent (step `fix_text_colors`, config has fix_rendered/min_contrast/
  fix_templates — `topic` was never the real problem).
- Same family as the never-registered `checkpoint_for_review` found by the
  claims-verification thread, and the same two-hand-maintained-rosters drift
  class CLAUDE.md's council section warns about. Note `registry.go:1866` has
  its own registry-backed `IsLocalAction` — the validator uses the deprecated
  list instead.

## Defect 2 — a failed handler saga is stamped 'complete'

- `CompleteWorkItemAction` (`load_work_item_actions.go` ~735–800) gates
  completion on `verifyBeforeComplete` **only for item types with a
  registered verifier** — `hardcoded_section_colors` has none — and never
  inspects the handler response it is storing: a `result` whose own
  `response.status` is `'failed'` (with a WORKFLOW_INVALID error) is written
  alongside `status='complete'` in the same UPDATE.
- Consequence: the improvement loop believes the defect class is handled;
  the item_key dedup then suppresses re-detection until the next discovery
  pass re-files it — churn that looks like progress. (In the robot-hands
  case the "fix" the items described was itself misconceived — stripping a
  hardcoded DARK background on a dark site — so the no-op was accidentally
  harmless. The completion lie is not.)

## Fix candidates

1. Register the action once, in the right place: entry in
   `actions/registry.go` (Handler: FixForcedTextColorsAction, IsLocal) —
   and reconcile the validator to use the registry-backed
   `IsLocalAction` instead of the deprecated hand list (or at minimum add
   the name to both). The two-list drift is the structural defect.
2. In `CompleteWorkItemAction` (or the dispatch loop's routing before it):
   treat `result.response.status == 'failed'` / presence of
   `response.error` as a failed attempt — `attempt_count++`, status
   `failed`/retry — never `complete`. This is the durable guard; it covers
   every future unregistered/broken workflow, not just this one.
3. Optional: a verifier for `hardcoded_section_colors` (verifyBeforeComplete
   policy in `complete_work_item_verification.go`) — but (2) is the class
   fix; a per-type verifier only patches this instance.

## How to verify

- Re-run the color-variable-fixer dispatch after (1): no WORKFLOW_INVALID;
  after (2): a deliberately broken workflow leaves its item failed with
  attempt_count incremented, not complete.
- Fleet sweep for the same lie:
  `SELECT id, site_id, item_type, completed_at FROM site_work_items WHERE
  status='complete' AND result->'response'->>'status'='failed';`
  (returns the two items above today; should return nothing after (2) plus
  a data correction).
