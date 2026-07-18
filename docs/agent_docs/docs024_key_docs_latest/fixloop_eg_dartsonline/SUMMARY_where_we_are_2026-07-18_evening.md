# Where we are — 2026-07-18 evening

*Current-state snapshot of the diagnosis→fix loop: what just shipped, what's
open, what's planned. The journey narrative is `SUMMARY_the_immune_system_2026-07-18.md`;
this is the operational "state of play" as of turn 40. Turn detail in
`NOTES_running_fixloop(10).md`.*

---

## Live now (shipped + verified)

- **Both councils on Claude Sonnet 5, reviewers at max_tokens 8000** (D1). Proven
  non-truncating; two reviewers wrote >3000 tokens, so the raise was necessary,
  not precautionary. `fix-proposer` + `council-gate` aligned. Backups taken.
- **The gate roster mirror (`099`) now detects config-value drift**, not just
  seat-name drift — the blind spot that stranded the gate on the old model. Deep
  JSON compare; tested both directions.
- **F1.2 done — the implementer base branch is a per-run input.** The stale
  `084_site_improvements_local_ai` literal is gone from all three spots
  (read/cut/PR); `read_current_files` + `create_branch` are fully per-run now,
  `prepare` completes when the `base_branch_field` image lands (safe `main`
  fallback until then).
- **The code-lookup verify tier (F2.3b(c))** with the Go-receiver-aware symbol
  match + dedup, proven converting the historian's escalation into an approval.
- **Agent-state autogather** in the diagnosis bundle for config-shaped bugs.
- **BUG A has a council-APPROVED fix plan** (`/bugs_open/008`) ready for the
  implementer; the CI-guard test to bundle in is noted there.

## Just changed / carried by the next image (inert until built)

- `base_branch_field` on `diagnose_prepare_fix_commit` (F1.2 Go half) — committed,
  rides the next chassis image; verify with `strings … | grep base_branch_field`.

## Open — correctness issues on the council (confirmed live)

- **bugs_open/016 finding 1 — the `.result}}` render fix is UNPROVEN.** The fix
  landed on `fix-proposer` at 13:15:11Z, but no fix-proposer repropose has STARTED
  since. **Correction to an earlier claim in this thread:** the D1 proof run
  (`00a20123`) started 13:11:13Z — four minutes BEFORE the fix — so it carried
  pre-fix config and its repropose DID render `<no value>` (the reviser was blind
  to review text in that run). The D1 *truncation* conclusion still stands (that
  is independent), but the review-injection was not clean in that run, and I did
  not catch it at the time. **Watch for the first fix-proposer repropose whose
  ORCHESTRATION starts after 13:15:11Z** — join `llm_call_log` to
  `orchestration_states.created_at`, test the run start not the step time. That
  run is the proof.
- **bugs_open/016 finding 2 — the reviser is half-blind.** 13 seats seeded, the
  repropose prompt threads only 6, so **7 seats' objections are invisible to the
  reviser** (adoption_guardian, compliance, debug_historian, llm_reliability,
  diagnosis_guardian, improvement_guardian, render_guardian) — 54% of the council.
  Arrived by seat growth, so it recurs on seat 14. **Decision made: read the
  `council_report` artifact's reviews array once, not per-seat prompt refs**
  (idempotent). Planned in `DESIGN_diagnosis_side_code_tier.md` §6; not yet built.
- **bugs_open/019 (another thread) interacts with D1.** The council gate voids a
  round if any reviewer overruns 8000 tokens, and substantial submissions push
  four of seven seats past half that. My D1 raise set the ceiling AT 8000 — which
  prevents truncation but sits exactly where 019's void triggers. Flagged for the
  gate thread; **worth revisiting whether 8000 is the right ceiling given 019**,
  or whether the void-on-overrun behaviour should change instead of the ceiling.

## Planned — next loop-development build

- **The diagnosis-side code tier** (`DESIGN_diagnosis_side_code_tier.md`): give
  the DIAGNOSER the code-search the council already has, by reusing the
  `diagnose_code_lookup` action via a new verdict `code_requests` field. Closes
  the "is this cause elsewhere?" gap on the diagnosis side — the same class of
  question the council tier was built for. Small (reuse; the action + index
  exist); complementary to call-graph following (breadth vs depth).
- **The 016-finding-2 reviser fix** (read the reviews array) — independent, and a
  live correctness bug, so arguably first.

## Standing owner decisions (unchanged)
- CI guard the historian keeps asking for → bundle into the BUG A PR (008).
- Whether to auto-default the implementer `BASE_BRANCH` to the diagnosis's own
  ref rather than requiring the operator to set it.
- The 8000 ceiling vs bugs_open/019 (new, above).

## The honest one-liner
The loop's tooling keeps maturing — a real model upgrade, a per-run base branch,
a drift-aware mirror — and the audits keep finding real correctness bugs in the
council's plumbing (a half-blind reviser, an unproven render fix, a
void-on-overrun ceiling). Both the new capability (diagnosis-side code tier) and
the fixes (016 finding 2) are small because the hard parts already exist; what's
left is disciplined building, and one honest correction: a run I called a clean
pass had a blind reviser I missed.
