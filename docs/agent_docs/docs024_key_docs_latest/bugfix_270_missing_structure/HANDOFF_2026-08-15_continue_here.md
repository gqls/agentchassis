# HANDOFF — bugfix 270, 2026-08-15 — start here in a fresh session

**Bug:** `bugs_open/270_HANDOFF_2026-08-13_missing_structure_check_fires_on_vestigial_columns_so_every_run_orders_a_full_site_rerender.md`
(read its `## UPDATE` sections at the bottom — they're the current state, kept current)
**Sibling bug found while fixing this one:** `bugs_open/280` (a second reader
of the same vestigial columns, in `check_decision_guards.go` — different
failure shape, filed separately, not part of this fix).
**Workstream docs:** this directory. Read `PLAN_2026-08-15_missing_structure_check.md`
first (the design + reasoning — nothing there has changed and it does not
need re-deriving), then `NOTES_bugfix_270.md` (the full evidence + session
trail, append-only, newest at the bottom) if you need the "how do I know
that" behind any claim.

## State right now (re-check before trusting — this tree is shared, other
sessions edit it between your reads)

**Done, all committed:**
- The fix itself: `fdc5daec1` — `check_missing_structure.go` retyped to
  query `site_components`; `check_missing_structure_test.go` added (5 tests,
  asserts on the actual SQL text so a regression can't hide behind a green
  happy-path). Re-verify with
  `go test ./platform/orchestration/actions/discovery_checks/... -run TestMissingStructure -v`
  if you doubt it — should be 5/5 pass.
- `bugs_open/280` filed (`0aec56e01`) for the sibling `check_decision_guards.go`
  finding.
- `bugs_open/270` itself updated twice (`036d9a7e9`, `514c53052`) with the
  fix pointer and the council verdict.
- LANDMINES.md's vestigial-columns entry corrected (`017aae6e4`) — it used
  to claim "read by exactly one caller"; now notes two (270's and 280's), and
  flags that its own "zero rows = no chrome" heuristic has a stated
  exception. Synced to `doc_notes` via `scripts/landmines-sync.py --apply`
  (succeeded on retry — `postgres-clients-0` exec was flaky for large
  payloads this session; a plain `SELECT 1` always worked, larger queries
  sometimes needed `run_in_background` or a retry — if this recurs, it's
  cluster load, not a broken connection, don't chase it further).
- **Council verdict: APPROVED**, 5 advisory objections, none high-severity,
  round 1 (13 reviewers, 4 abstained). Correlation
  `524ff897-b697-4c5c-a66f-8939b0457049`. The commit already carries
  `Council-Submitted:` (written before the verdict landed, correctly) —
  **do not amend it to `Council-Reviewed:`**; forward-only forbids amends,
  and the `098` coverage report resolves and credits `Council-Submitted:`
  commits automatically once their correlation shows approved. Nothing
  further to do on this front.

**Owner decision, 2026-08-15: shipping (build + release) is explicitly
LEFT TO THE OWNER, not this session** — asked directly, answer was "Leave
it — I'll ship it." Do not build or release on this bug's behalf without a
fresh, explicit ask; the owner may already have shipped it by the time you
read this — **check the artefact (step 2 below) before assuming step 1 is
still outstanding.**

**NOT done — this is the actual remaining queue, most impactful first:**

1. **Image build + fleet roll — NOT this session's job (see owner decision
   above); check whether it's already happened before offering to do it.** The code is inert until an image is built
   from committed HEAD (`make build-agent-chassis` — confirm that's still
   the right target service for the `discovery_checks` package before
   running it; it was as of 2026-08-15) and the fleet is rolled
   (`make release`, whole-fleet, **owner-run — this is very likely not this
   session's call to trigger unilaterally**; say so and ask rather than
   rolling a single service at its own tag).
2. **Verify at the artefact, once rolled**, per service not per fleet:
   `kubectl -n ai-persona-system logs -l app=agent-chassis --tail=300 | grep -m1 'build provenance'`,
   then `git merge-base --is-ancestor fdc5daec1 <the stamp>`. If the startup
   line has scrolled out of range (`agent-chassis` is busy — this has
   happened before), probe the binary directly per LANDMINES.md's recipe
   (never `strings`; always run a known-present AND a known-absent sha as
   controls in the same breath).
3. **Fleet verification** (exact queries in PLAN.md §7): after one full
   discovery rotation, the 17 stale `missing_structure:rerender` items (14
   `unresolved` + 1 `detected` + 2 `deferred` as of 2026-08-15 — **re-count,
   this will have moved**) should flip to `complete` with
   `result->>'resolved_by'` set to something naming this fix, and
   `max(created_at)` for that key should stop advancing on sites whose
   chrome serves (pre-fix baseline to beat: 2026-08-14 16:35).
4. **Close out**, only once 1-3 are all actually true (CLAUDE.md's
   fixed-AND-live bar — "committed" and "approved" are both already true and
   neither is enough on their own):
   - `git mv bugs_open/270_... bugs_closed/270_...` (or the two-file
     add+commit dance if `git mv` behaves oddly on this tree — LANDMINES.md
     has a `git mv` + pathspec-commit trap for exactly this move, worth
     reading first: `grep -A6 "git mv.*pathspec" docs/agent_docs/docs024_key_docs_latest/LANDMINES.md`).
   - Append the outcome to `NOTES_bugfix_270.md` and `README_where_we_are.md`
     (both already exist and are append-only — add to the bottom, don't
     rewrite).
   - A fresh `SUMMARY_<date>_...md` is warranted at this point — this is a
     real milestone (shipped and verified, not just "fix written"), and the
     existing summary is explicitly current-state-only per the standing-five
     convention — don't edit it, write a new dated file.

## Things NOT to do

- Don't touch `check_site_structural_validity.go` — it's under active
  council review by a different workstream (`portfolio_positioning`, round
  2/3 as of 2026-08-15). Its `head_essentials_missing` comment cites a now-
  stale count from bug 270; that's a note for THAT workstream to refresh,
  not a reason to edit their in-flight file.
- Don't fold the `check_decision_guards.go` finding into 270 — already
  filed separately as `bugs_open/280`; leave it there.
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
- Don't re-submit to the council — already APPROVED. Re-submitting a second
  time for the same unchanged code would be pure waste.
