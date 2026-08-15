# HANDOFF — bugfix 270, 2026-08-15 — start here in a fresh session

**Bug:** `bugs_open/270_HANDOFF_2026-08-13_missing_structure_check_fires_on_vestigial_columns_so_every_run_orders_a_full_site_rerender.md`
**Workstream docs:** this directory. Read `PLAN_2026-08-15_missing_structure_check.md`
first (the design + reasoning), then `NOTES_bugfix_270.md` (the evidence
trail) if you need the "how do I know that" behind any claim below.

## State right now (re-check before trusting — this tree is shared, other
sessions edit it between your reads)

- `platform/orchestration/actions/discovery_checks/check_missing_structure.go`
  is REWRITTEN in the working tree (queries `site_components`, not `pages`).
  **NOT YET COMMITTED.** `git status --short` on this path should show `M`.
- `platform/orchestration/actions/discovery_checks/check_missing_structure_test.go`
  is a NEW file, also uncommitted (`??` in `git status`). 5 tests, all pass:
  `go test ./platform/orchestration/actions/discovery_checks/... -run TestMissingStructure -v`
- Both build clean: `go build ./platform/orchestration/actions/discovery_checks/...`
- **Re-run these two commands before doing anything else** — another session
  may have touched this file since this handoff was written; if `git status`
  shows anything unexpected on this path, read it before overwriting.

## The queue, most impactful first

1. **Commit the fix, narrowly.**
   ```
   git add platform/orchestration/actions/discovery_checks/check_missing_structure.go \
           platform/orchestration/actions/discovery_checks/check_missing_structure_test.go
   git commit platform/orchestration/actions/discovery_checks/check_missing_structure.go \
              platform/orchestration/actions/discovery_checks/check_missing_structure_test.go \
              -m "..."
   ```
   Remember the pathspec goes on BOTH `add` and `commit` (CLAUDE.md — a bare
   `git commit -m` sweeps whatever anyone else has staged). Do not add `-A`
   or `.`. Check the yellow commit-scope report after — it should list only
   these two files.

2. **Council submission.** This is `platform/` code so the advisory gate
   applies (CLAUDE.md, "Council review of platform changes"). Build the
   submission JSON (`rationale` + `plan`, ≤8 edits, `grounded_in` evidence)
   from PLAN.md §2-4 — the rationale is essentially PLAN.md §1-2 condensed,
   the plan is the one file's diff. Run:
   ```
   ./docs/agent_docs/docs024_key_docs_latest/fixloop_eg_dartsonline/097_TRIGGER_council_review_v1.sh <submission.json>
   ```
   Save the printed `SUBMISSION_CORR`. Budget ~30 min for the round to
   actually land (dispatch queues behind the fleet — this is latency, not a
   dropped run; don't retry on that alone). If you commit before the verdict
   is back, add `Council-Submitted: <corr>` to the commit message (not
   `Council-Reviewed:` — only write that once you've actually read an
   APPROVED verdict). On REVISE, the objections come back with the
   reviewers' own checks already answered — read them into NOTES, revise,
   resubmit with `RESUBMIT_CORR=<corr>`. On REJECTED, read the guardian's
   note (`SELECT body FROM doc_notes WHERE categories ? 'council-gate' ORDER
   BY created_at DESC LIMIT 1;`) before doing anything else — PLAN.md §3
   already pre-empts the most likely objection (architecture-scope) with a
   consumer enumeration, so a REJECTED verdict here would be worth reading
   very carefully rather than assuming the pre-empt failed.

3. **File `bugs_open/279`** for the second vestigial-column reader in
   `check_decision_guards.go`. The full problem statement is already drafted
   — PLAN.md §5 has the exact paragraph to use (it was written to be filed
   verbatim, not summarised further). **Re-check the number is still free**
   before filing (`ls bugs_open/ bugs_closed/ | grep '^279'`) — as of
   2026-08-15 two OTHER sessions had already raced to claim 278 between when
   this was scoped and when it was filed, so numbers move fast on this tree.
   Separate commit from everything else here (one-commit-per-task).

4. **Update `bugs_open/270` itself** with a pointer to the fix (commit sha,
   this workstream directory) — "point at bugs, don't restate them," and a
   HANDOFF file that outlives the fix it asked for is exactly the trap
   CLAUDE.md warns about, so close the loop in the bug file, not just here.

5. **Ship and verify** (this is the part that actually can't happen until an
   image builds and the fleet rolls — not blocked on you, but don't skip it):
   - `make build-agent-chassis` (or whichever service registers
     `discovery_checks` — confirm which binary this package builds into if
     you're not sure) — builds from committed HEAD, so step 1 must be done
     first. Bump `IMAGE_TAG`.
   - Fleet roll is whole-fleet and owner-run (`make release`) — not a
     one-service apply. This step is likely NOT yours to trigger; say so and
     wait, or ask, rather than rolling a single service at its own tag.
   - Once rolled, confirm at the artefact, per service:
     `kubectl -n ai-persona-system logs -l app=<service> --tail=300 | grep -m1 'build provenance'`,
     then `git merge-base --is-ancestor <your-commit> <the stamp>`. If the
     startup line has scrolled out of range, probe the binary directly
     (LANDMINES.md has the exact recipe — never `strings`, always a
     known-present/known-absent control pair).
   - **Fleet verification** (PLAN.md §7 has the exact queries): after one
     full discovery rotation, the 17 stale `missing_structure:rerender`
     items (14 `unresolved` + 1 `detected` + 2 `deferred` as of 2026-08-15 —
     re-count, don't trust this number) should flip to `complete` with
     `result->>'resolved_by'` set, and `max(created_at)` for that key should
     stop advancing on sites whose chrome serves.

6. **Close out** — move `bugs_open/270` → `bugs_closed/` ONLY once fixed AND
   live AND the stale items have actually closed (CLAUDE.md's bar, restored
   by the owner 2026-08-12 — don't move it on "committed" or "council
   approved" alone). Update the LANDMINES.md vestigial-columns entry's
   reader-count pointer (it currently says "read by exactly one caller left
   in the tree" — that becomes wrong the moment 279 is filed, and wrong
   again in the other direction once 270's fix ships and 270's own caller is
   fixed too). Append the outcome to `NOTES_bugfix_270.md` and to this
   workstream's `README_where_we_are.md` (create one if it doesn't exist yet
   — it didn't as of this handoff; the SUMMARY doc's prose is a reasonable
   seed for its first entry if you want to fork it rather than start blank).

## Things NOT to do

- Don't touch `check_site_structural_validity.go` — it's under active
  council review by a different workstream (`portfolio_positioning`, round
  2/3 as of 2026-08-15). Its `head_essentials_missing` comment cites a now-
  stale count from bug 270; that's a note for THAT workstream to refresh,
  not a reason to edit their in-flight file.
- Don't fold the `check_decision_guards.go` finding into 270's fix — it's a
  different check with a different failure shape (wrong stored-assembly
  definition, not an always-true firing predicate) and has produced zero
  observed wrong verdicts so far. File it separately (queue item 3).
- Don't re-derive the Candidate-1-vs-Candidate-2 decision from scratch —
  PLAN.md §2-3 already did that weighing, including a marked correction to
  the original bug file's own fix sketch (`build_status='pending'` is NOT a
  safe "missing" signal — see NOTES). Re-litigating it costs a full research
  pass for no new information unless something has changed underneath it.
- Don't trust `scripts/who-owns.py 270`'s raw verdict without reading the
  cited workstream's docs — it says "OWNED or recently active" because the
  FILING session (`portfolio_positioning`) is still active on OTHER work; its
  own docs explicitly say 270 is "unowned, not on Phase B's critical path."
  This is a live false-positive shape in that tool, not a reason to stop.
