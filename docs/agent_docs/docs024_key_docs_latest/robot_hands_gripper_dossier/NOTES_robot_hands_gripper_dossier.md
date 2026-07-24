# NOTES — gripper dossier pilot

*Technical log. Append-only, newest at the bottom. Missteps included.
Design of record: `DESIGN_2026-07-24_gripper_dossier_pilot.md` (consolidated
from two grounded design passes; §5 seam reconciliation wins over either
half where they disagreed).*

---

## 2026-07-24 — workstream opened; scoring action built

Parent: per_site_ai (PLAN D15/D17), features_open/013. Owner unpaused after
robot-hands-site-fixes closed (R1–R9). Pilot scope locked: chat intake
(island) → pull → report-builder workflow → deployed report page + status
sidecar → island emails link.

### Owner decisions (2026-07-24, AskUserQuestion)
Shared sender (robot-hands@contactforsales.com) · prod E2E fixtures approved
(clean up after) · soft-launch unlinked · $50/mo cap on the new Anthropic
key. Defaults accepted: §5.3 mapping table, UUID+noindex, 24h/90d retention,
max_attempts=1. Tunnel VERIFIED LIVE (tools.apis.uk → 404 from island Caddy;
/api/v1/tools/health → 502 = no engine container yet, correct). Remaining
owner action: issue the $50-capped key.

### Built: score_grippers (first task)
- `platform/orchestration/actions/score_grippers_action.go` + registry entry
  + `score_grippers_action_test.go` (all green) +
  `sql_for_agents/204_robot_hands_matchmatrix_normalized_specs.sql`
  (NOT yet applied; pre-image safe — names no actions).
- Port source: dumped the live tool template (32,228 chars) to scratchpad;
  physics ported line-for-line (MATERIALS μ/ferrous map; dyn=m·a·S;
  fJaw=dyn/(μ·n); fDir=dyn; mEq=dyn/9.81; ipRow only under a requirement;
  unknown-never-pass; conflict note; verdict thresholds incl. the 1.25×
  marginal band; rank/headroom sort; the tool's exact no-match sentence).
- Input hardening beyond the tool (server context): material accepts μ key
  OR name alias; ip accepts "IP54" form; cycle_rate→safety tier (≤10→2,
  ≤30→3, >30→4, explicit safety_factor wins) per DESIGN §5.3; malformed
  spec = hard error (routes to error_step, never a guessed default).
- **Misstep, caught by the test run**: my conflict-note test asserted
  impliedMu = (mass·a·S)/(force·n) = 0.15; the tool actually derives it from
  the PUBLISHED PAYLOAD, (payload·a·S)/(force·n) = 0.26. The code was right
  (ported faithfully); the test expectation was wrong. Lesson: when porting,
  test expectations must be computed from the reference implementation, not
  re-derived from what the formula "should" be.
- fact_block contract: opens with the mandatory sentence when match_count=0;
  marks unpublished figures "NOT PUBLISHED by the manufacturer — say so if
  mentioned; never estimate it"; carries per-candidate source_url +
  verified_date; carries the substituted formula strings (these double as
  the discriminating E2E artefact — no other page carries them).

### Verified this session (live)
- 10 products rows, all with source_url + verified_date, none with
  matchmatrix block yet (seed 204 will add).
- Island tunnel public; API path answers 502 (nothing behind Caddy yet).

### Committed + submitted (same day)
- `e19aa5d10` score_grippers + tests + registry + seed 204 (pathspec commit).
- `ede694ef2` gofmt fix (pre-commit pattern check caught the test file).
- `12fa24e6b` this workstream's DESIGN + standing docs.
- `5229d7fa1` per_site_ai session docs + features_open/013 + WRONG_CALLS +
  CAPABILITIES correction.
- Council gate: **SUBMISSION_CORR ffccb83c-1833-45cb-bb0f-8edcc874699e**
  (score_grippers change; first submission bounced client-side — `plan` must
  be an OBJECT {summary, edits[], grounded_in, risks}, not a bare edits
  array; header of 097 has the schema). If APPROVED, later platform commits
  in this lane carry `Council-Reviewed: ffccb83c-…`.

### Missteps this session (append-only, the point of this file)
1. Conflict-note test expectation derived from intuition instead of the
   reference JS (impliedMu uses published payload, not request mass) —
   caught by the first test run; code was right.
2. **A `| tail -2` pipe swallowed a `go test` build failure**, so the gofmt
   commit chained through on a red build. Lucky twice over: the failure was
   another session's uncommitted vet_med WIP (shared-tree-wont-compile
   class, verified by building a clean `git archive HEAD` — my code green),
   and the commit was formatting-only. Practice: never gate a commit on a
   piped test command; check `go build` exit directly or use pipefail.
3. First council submission rejected client-side on schema (see above) —
   cheap, by design (no credits spent).

## 2026-07-24b — batch 2: the four pipeline actions (all built, tested, committed)

- **Council verdict on batch 1 (score_grippers): APPROVED** (corr ffccb83c,
  17:07Z). Commits predate the verdict (advisory-first flow), so no trailer
  on them; verdict recorded here and resolvable by corr.
- **Batch 2 committed `2849564ec`** (pull_report_requests +
  emit_report_status_files / verify_report_prose / create_report_page /
  report_charts.go + registry). Council submission:
  **SUBMISSION_CORR 7ed137d1-361c-4f69-9361-9e4ba1dfa6bf** (pending).
- All tests green against clean `git archive HEAD` overlay (shared tree
  still carries the other session's vet_med WIP — never gate on the dirty
  tree).
- **Design deltas, declared**: ONE chart (headroom) — the payload scatter
  dropped: published payload is only comparable within payload-rated
  technologies, a cross-tech scatter would mislead. verify_report_prose
  gained a `context_field` (request strings like the mounting standard are
  legitimate prose context score_grippers never sees) and a SKU-shape check.
- **Misstep 4 (kept me honest twice)**: my "clean prose" test fixtures used
  the 2.5 kg + IP54 scenario — which is a GENUINE zero-match application
  (every candidate fails or goes unknown on unpublished IP). The gate
  rejected my own fixtures for missing the mandatory no-match sentence
  before it ever saw an LLM. Fixture bug, not gate bug — but also the first
  live proof the gate bites. Second: the SKU regex missed digit-leading
  models ("2F-140"); caught by its own test, tightened.
- **Accepted pilot residual (documented in the council submission)**: a
  fabricated plain-word product name with no digits passes the deterministic
  gate — left to the writer prompt + validate_page_content contamination
  checks. Revisit if observed.

### Next
1. Await council verdict corr 7ed137d1 (~30 min; missing row = queued).
2. Seeds 205 (report-dossier component), 206 (island config placeholder),
   207 (three agent_definitions), 208 (two scheduled_tasks, disabled) —
   205 pre-image OK; 207/208 strictly post-image.
3. Island service `cmd/gripper-intake/` (code+tests offline; live turns
   blocked only on the owner's $50-capped key).
4. Image roll (IMAGE_TAG bump; discriminating pod-grep: 'report-dossier' +
   'pull_report_requests' + positive control), then 206–208, then induced
   E2E fixtures (success / no-match / failure) per DESIGN §6.
