# 017 — fix_forced_text_colors is never registered; its WORKFLOW_INVALID failures are stamped 'complete'

**Filed:** 2026-07-18, from the robot-hands R1 thread
(`docs/agent_docs/docs024_key_docs_latest/robot_hands/RUNNING_NOTES_robot_hands_site_fixes.md`).
Two coupled defects, one incident.

---

## STAYS OPEN — fix committed 2026-07-18 (bugfix thread2), but INERT until a chassis image ships

**Both legs fixed in code, plus the class fix and a correction to this file's own
diagnosis — but the defect is still reproducible in production.** Per the
`/bugs_closed/` bar (CLAUDE.md, 2026-07-19: *fixed AND live*), a fix that is committed
but inert until the next image roll stays here. Move to `/bugs_closed/` only after the
image ships AND the running pod is grepped for `handlerReportedFailure`.

**Workstream docs:** `docs/agent_docs/docs024_key_docs_latest/work_item_completion_integrity/`
(the standing five — PLAN/RUNBOOK/NOTES/README_where_we_are/SUMMARY).

| Leg | Fix | Where |
|---|---|---|
| Defect 1 | `fix_forced_text_colors` registered | `actions/registry.go` — landed in HEAD via `06376bcbf` (another session swept my working tree mid-task) |
| Defect 2 | `handlerReportedFailure` blocks completion on a failed saga verdict, routing to the existing attempt machinery | `complete_work_item_verification.go` + `load_work_item_actions.go` gate 1 |
| Class | Build fails if any action registers an `ActionInputSpec` with no registry entry | `actions/registry_parity_test.go` (negative-control verified) |
| Misdiagnosis | Dead `LocalActions` map deleted; the comment + 2 live guide docs that propagated "register in TWO places" corrected | `actioncheck/`, `batch_webscrape_action.go`, 2 docs |
| Data | 54 mis-stamped rows corrected to `failed` (reversible via `result._correction`); sweep now returns 0 | live `clients_db` |

**⚠️ THIS FILE'S ROOT-CAUSE ANALYSIS FOR DEFECT 1 WAS WRONG.** It blamed drift between
two hand-maintained rosters (`registry.go` vs the DEPRECATED `actioncheck/local_actions.go`).
There is only **one** live list: `actioncheck.IsLocalAction` (`actioncheck.go:20`) delegates
to a checker `registry.go` installs at `init`; the `LocalActions` map's own lookup was
commented out (`local_actions.go:185-188`) and it had **zero live references repo-wide**.
It was dead, not drifting. The `batch_webscrape_action.go` header comment telling authors to
"add to TWO places" is what seeded the belief — and it survived long enough to misdirect the
diagnosis of the very bug it caused. Lesson recorded in 016b §9 and in `doc_notes`
(`subject_key='action_registry'`): when a deprecated-looking list sits beside a live one,
grep the symbol for live references before theorising about drift.

**Defect 2 was 27× larger than filed.** The sweep found **54** mis-stamped items across
**6 sites** (robot-hands.com, finetuning.uk, gaswholesalers.com, ai-agent-orchestration.com,
leopardessconsulting.co.uk, idea.uk) and 4 item types, back to May — not the 2 recorded below.
One of them (`render_js_snippets` vs the registry's `render_js_snippets_for_site`) is a
**seed typo**, a different cause reaching the same lie — which is why only the leg-2 guard
closes the class.

**Council-reviewed** (advisory gate): submission `319e23f6-b333-42ba-88ef-069b4426c057`.
Round 1 → REVISE. Round 2 → REVISE but **8 approve / 2 object** (guardian moved
high→medium; tooling_provenance, guidelines and debug_historian flipped to approve).
Answered across the two rounds: blast radius measured (1656 completions / 43 item types
/ 30 days, guard fires on 6, all genuine); all 8 `status='complete'` call sites audited;
unknown-verdict handling added and then upgraded to `agent_error_log` per bug_historian;
`doc_notes` persisted per tooling_provenance; doc purge per guidelines; dormantActions'
"unseeded" claim verified against `agent_definitions` per guardian.

Commits: `c82b2872c` (the fix) and `c80fffc83` (round-2 follow-up). No
`Council-Reviewed:` trailer is claimed — the verdict is REVISE.

> **CORRECTED 2026-07-19.** This paragraph previously dismissed the two residual
> objections as *"verify the author's audit independently" asks rather than identified
> defects*. **That was the wrong call.** The objections (bug_historian low, guardian
> medium) said my "only `CompleteWorkItemAction` completes from a handler reply" claim
> rested on an author-run regex audit described in prose — and they were right: I had
> read four of the eight call sites and inferred the three admin paths from their
> filenames. An unverified structural claim IS a defect, especially once it has reached a
> commit message, this handoff, a §9 guide entry and two `doc_notes`. **Now actually
> verified:** `confirm_work_item_handler.go:212`, `site_admin_handlers.go:793` and `:987`
> each construct `result` via `jsonb_build_object` from human input
> (`'resolved_by','admin'` / `'approved_by','admin'`) and never read or store a `response`
> envelope, so none can carry a failed verdict. The claim holds; the method did not.
> **Caught by:** the owner asking me to re-read CLAUDE.md, whose diagnosis section had
> been inverted the same day — see the note below.

> **⚠️ PROCESS COST, RECORDED SO NOBODY REPEATS IT.** This review cost **four** council
> runs, not two. Rounds 2's dispatches produced no `orchestration_state_audit` rows for
> minutes, which I read as silently-dropped spawns; they were merely **queued** (~16 min
> under backlog vs ~10 s when quiet). I resubmitted three times, twice on untested
> transport hypotheses (payload size, then `RESUBMIT_CORR`) — both wrong. Full pattern
> and the one query that settles it: 016b §9 "A queued orchestration is indistinguishable
> from a dropped one".

**Still to do:** ship a chassis image (Go changes are inert), then verify against the
**running pod** — `strings /app/agent-chassis | grep -c handlerReportedFailure` — before
trusting any functional re-test.

---

## Original report (below, as filed — see the correction above)

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
