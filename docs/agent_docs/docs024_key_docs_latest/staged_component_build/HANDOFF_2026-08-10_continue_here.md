# HANDOFF — 2026-08-10, fresh chat starts here: batch 8 is the TOOLS, and the pool is qualified

**Supersedes `HANDOFF_2026-08-09_continue_here.md`** for state and work-list; that file's
§0 (the shared-228 situation), §2 (the two rerender traps) and the 08-08 handoff's §3 (the
interactive-fence line) remain binding and are not repeated. Read them.

## 1. State (all verified 2026-08-10 ~10:15Z)

- **56 subjects proven end-to-end: 54 sections + 2 tools.** Batches 6/6b/7 all S6-green
  with negative controls red. Batch 7's table: NOTES `## 2026-08-09 (evening)`.
- Fleet: chassis + browser-runner **v1.0.1277** (pods up 21:35Z 08-09). Re-grep at
  session start; never carry a version forward.
- Naming-contract check: **PASS** — 54 canonical tools, 25 testable, 13 authoring
  backlog, 0 BROKEN.
- **None of the four batch-7 gates moved overnight** (re-measured): ROI CSS unfixed
  (`roi-inputs-title` still fixed-width, component row untouched since 07-25); all four
  tracker feeds still 404; no reply on the tracker CONTRIB; no movement on 228's JS
  choice.

## 2. Batch 8 — the ready tools, QUALIFIED (probed 2026-08-10)

17 active tool placements have no `subject_type='tool'` PLAN. They split three ways:

### 2a. Five clean singles — fence these first, same line as batch 7

| tool | proof page (200, 0 bad assets) | site_id |
|---|---|---|
| tool-bayesian-ranking | gamesdesign.co.uk/tools/bayesian-ranking.html | e33263f4-74f8-494f-b191-546845dbbddf |
| tool-grip-force-friction-calculator | robot-hands.com/tools/grip-force-friction-calculator/index.html | 00ff3af5-dad8-4770-9f70-3edc267a3c92 |
| tool-llm-cost-calculator | ai-agent-orchestration.com/tools/tool-llm-cost-calculator.html | 2a8ebf9c-20a2-4c39-b191-840b012371da |
| tool-matchmatrix | robot-hands.com/tools/matchmatrix/index.html | 00ff3af5-dad8-4770-9f70-3edc267a3c92 |
| tool-setup-builder | dartsonline.com/tools/setup-builder/index.html | 5fe8785b-223d-41a3-88ee-c07187622381 |

Page IDs are in the census output (re-run it; they move):
`/tmp` copies die — the query is the LEFT JOIN on `doc_plans dp ON … dp.subject_type='tool'`
over `component_level='tool' AND is_active AND forked_from IS NULL` placements.

**Tool-fence specifics that differ from the section line:**
- PLANs go to **`subject_type='tool'`** — `gen_component_plan_sql.py` hardcodes
  `'component'`; extend the manifest schema with an optional `"subject_type"` per entry
  (default `'component'`) before persisting. Small, do it first, keep it committed.
- The S6 instrument is **`./docs/leopardessconsulting/scripts/tool_acceptance_run.sh
  <site_id> <domain> <function>`** (RUNBOOK §10), not the component dispatch script.
  Read its result honestly: FAILED still reports COMPLETED; `__step_error` holds truth.
- **A calculator's fence wants `computed_values`, not a regex** — `interaction` with
  `text_matches /£/` passes an unwired tool printing £0.00 (the check type's own doc
  comment, `run_checks_action.go` ~752, says exactly this). Capture goldens from the
  live tool while known-good; `toolgolden.py --emit-criteria` is the instrument
  (leopardess lane precedent). The 120s deadline applies: profile-gate everything
  static to desktop.

### 2b. Nine loancalculator.co.uk tools — REUSE THE LANE'S GOLDENS, do not author blind

application-tracker, car-finance-pcp-hp, compare-loan-offers, consolidation-risk,
credit-health-check, early-settlement, loan-repayment (on `/index.html`),
overpayment-impact, rate-stress-test.

The `loancalculator_couk` lane already maintains a golden-values acceptance harness over
**exactly these pages**: `loancalculator_couk/toolgolden.py --compare` against
`acceptance/GOLDEN_2026-08-03b_after_orphan_retired.json` (their
`HANDOFF_2026-08-08b_continue_here.md` §6 is the worked command). The right design is
fences whose `computed_values` come FROM those goldens — one source of truth, no drift.
**Coordinate first** (ownership check at the point of WRITE, not filing — the 08-09
`WRONG_CALLS` lesson): their handoff's open items are voice/content, not fences, so this
is likely a welcome contribution — but ask, in their dir, before persisting anything.
Their landmine applies to you: **guard every served fetch with `wc -c` + a DOCTYPE
check** — a deploy-window fetch returns a B2 error blob at HTTP 200.

### 2c. Blocked/skip, with evidence (probed 2026-08-10)

- `tool-fuel-budget-forecaster` — gaswholesalers logo 404 (1 bad asset; **6+ days now**).
  Chrome-blocked class. NOTE: a 404'd `<img>` DOES log `Failed to load resource` as a
  console error (seen live on the vonc transient), while a favicon 404 does NOT (measured
  08-09) — so this page cannot hold a green `no_console_errors` baseline until the logo
  is fixed.
- `tool-gas-unit-converter` — page serves but is the known-broken unlabelled tool
  (standing defect list; repair ticket parked `wont_fix` in a human pile). Fencing it
  would certify a broken page or red forever. Owner call, not effort.
- The four batch-7 gated interactive sections — unchanged, see the 08-09 handoff §3.

## 3. The line, unchanged — plus the tool-batch order

Per subject: author fence+mutants beside the batch-7 worked examples → `try_fence.go`
(live) → `prove_fence_mutants_file.go` (every check watched red; inert-script mutant for
every driven check) → manifest → `gen_component_plan_sql.py` dry-run then `--apply` (with
the `subject_type` extension) → readback byte-diff → acceptance run → NOTES table +
README + tally. Qualify each subject first: JS binds template selectors AND the served
page loads the script AND the effect is observable and safe to drive.

## 4. Standing defect list for the owner

Unchanged from `HANDOFF_2026-08-09` §4 items 1–8 (228 fixed+proven; gaswholesalers logo
now 6+ days; hero.jpg family; finetuning card images; article-body overflow; broken tool
pages; ROI mobile overflow; tracker feeds 404). Nothing new today.

## 5. Session-start checklist (do these before any dispatch)

1. `git log --oneline -10` and re-read this file FROM DISK (co-edited tree).
2. Pod-grep chassis (`request_component_browser_run` ≥1 + a negative control) and
   browser-runner (`"non-numeric w/h in result"`, `"no element matches"`,
   `"computed_values"` → 1 each). No dispatch within 300s of a pod restart.
3. Re-run the census and the naming-contract check; the figures above WILL have moved.
4. `who-owns.py` + workstream-dir grep for every subject you are about to WRITE —
   ownership answers age in hours (08-09 lesson, twice over).
