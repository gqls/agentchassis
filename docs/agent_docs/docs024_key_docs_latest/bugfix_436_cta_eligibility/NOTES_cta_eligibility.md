# NOTES — `bugs_open/436` CTA eligibility lever — append-only, newest at the bottom

## 2026-09-02 — session start, thread resumption check

- `scripts/who-owns.py 436` → "OWNED or recently active", but the only commit is `5449f71b1`,
  the 391-closing commit that FILED 436. No `bugfix_436_*` lane dir existed; no peer session
  named for it (ListAgents 19:09 BST — 64 peers, none on 436); 391's final summary is
  `SUMMARY_2026-09-02_cta_relevance_CLOSED.md` and its handoff says "its own lane, filed when
  391 closes". **Verdict: no active thread; resumed here.**

## 2026-09-02 — bug still valid at HEAD (re-verified, not assumed)

- `chooseCTATargets` (`resolve_internal_links_action.go:651`) unchanged: rank by
  `COALESCE(nav_order,100)` then `name`, take `[0]`/`[1]`; no eligibility input.
  `git log` on the file: nothing behaviour-changing since the 391 proofs.
- No `eligible_as_cta_target` anywhere in the tree (`grep -rn`, 0 hits outside 391/436 docs).
- `\d pages` `[MEASURED 2026-09-02]`: no eligibility column; `rebuild_policy` is the policy-column
  precedent (NOT NULL + CHECK).
- Live corroboration `[MEASURED 2026-09-02]`, `site_work_items` open rows: two
  `needs_content_planning` deferred verdicts describing "two arbitrary tools … as default
  fallback CTAs on unrelated guide and news pages" (63-tool site — the all-default alphabetical
  shape), and ~10 `page_rerender` "misdirected CTA" items unresolved after 2 attempts.
- Consumer enumeration re-run at HEAD — 3 ranking callers, 4 universe callers, 3
  `BestLabelMatchForPage` callers; full table in PLAN. The header fallback
  (`render_site_components_action.go:181-196`) confirmed never-persisted and label-match-free.

## 2026-09-02 — design decisions taken while reading (evidence inline)

- **Refuse-not-drop for the label binding.** `cta_label_universe.go:144-155` records the
  measurement: filtering the best candidate out of the pool let a runner-up win on one shared
  token — 10 of 35 wrote somewhere else, most wrong. An opted-out page removed from the universe
  would reproduce that failure for every label naming it. So the page stays a candidate and
  `BestLabelMatchForPage` refuses — same NO OPINION semantics as the self-link rule.
- **Judge inherits the refusal** (`JudgeCTALabel` calls `BestLabelMatchForPage:169`) — needs a
  distinct `Silence` reason or "copy names an opted-out page" folds into `names_nothing` and the
  signal reaching detectors dies. The Silence seam was built for exactly this
  (`cta_label_agreement.go:110-115`).
- **Recompute KEEP #2 is NOT changed**: flipping the flag on an already-locked page unlocks it on
  full REBUILD only, not on recompute — same asymmetry 391 measured ("only retirement unblocks
  them"). Stated in PLAN §what-this-does-not-do so nobody reads the lever as a repair tool.
- **Detector cannot import `actions`** (`discovery_checks` is a child package; parent imports
  child) → the supply SQL + ranking must be shared through `datahelpers`, or the check mirrors by
  hand and drifts. LNK-036 is the precedent for putting the shared definition in datahelpers.
- **Runner tolerates an unknown check name** (`discovery_checks.go:201-212` logs against the
  registered list) — but enablement still goes `_HOLD` so config never names a check the running
  image lacks. [VERIFY the nil branch is non-fatal before relying on it — see TODO]

## TODO / open verification

- [ ] Read `discovery_checks.go:195-220` nil-check branch (non-fatal confirmed?).
- [ ] `cta_label_audit.go:310` — which match function does the audit use; does it need the reason?
- [ ] Latest migration number at write time (709 seen; re-check before writing 710/711 — CONFIRMED STALE: 710 was taken mid-session, renamed to 714/715 — other
      sessions add files hourly).
- [ ] Council round; RESUBMIT_CORR chain if REVISE.

## 2026-09-02 — corrections to this file's own earlier lines

> **CORRECTED 2026-09-02 (same session):** the line above saying "Runner tolerates an unknown
> check name" is **REFUTED** — its own [VERIFY] marker was discharged by reading
> `discovery_checks.go:195-218`: the runner **FAILS THE WHOLE STEP** on an unregistered name
> unless `allow_unregistered_checks:true` is set, and that flag is off by default on purpose.
> This is what makes `715_HOLD` mandatory rather than cautious. The marker did its job: the
> claim was checked before anything acted on it.

> **CORRECTED 2026-09-02 (same session):** `715`'s first draft targeted
> `{workflow,steps,run_discovery_checks,config,checks}` — a path composed from the step's purpose.
> The live row keys the step **`run_checks`**. Caught by querying the live config before commit;
> the migration's DO/RAISE guard would also have refused. WRONG_CALLS row added
> (a-config-path-composed-from-purpose-not-read-from-the-row).

## 2026-09-02 — build complete; what exists and what proved it

- `datahelpers/cta_positional.go` (supply SQL + ranking + eligibility, shared), `cta_label_universe.go`
  (+column, refusal), `cta_label_agreement.go` (`names_ineligible_page`), `label_match.go` (inverted
  field, polarity-pinned), `check_cta_rank_anomaly.go` (+test), migrations `714` + `715_HOLD`,
  register LNK-041 + index row, verifier-coverage ratchet acknowledged (review-only + self-retracting).
- Tests: `cta_positional_test.go` (both directions; header form pageName="" separately; pre-lever
  behaviour pinned), `cta_eligibility_label_test.go` (refuse-not-drop with a discriminating fixture —
  fails on a pre-filter rewrite; judge reason; polarity), detector predicate table (391 fossil fires;
  demoted/all-default/ladder/two-leaders/too-few all silent, each named as a real estate state) and
  the lever-silences-alarm pairing test through the REAL ranking.
- `go build ./...` clean. `discovery_checks` green. `datahelpers` green EXCEPT
  `TestNoHandSpelledTombstonePredicate` — **pre-existing, not this lane's**: it flags
  `check_unrendered_page_imagery.go` (committed by the 114 lane, `ed8480a25`) and
  `wire_page_hero_on_landing.go` (UNTRACKED — 114's in-flight WIP). HEAD is red on that test
  before my commit. The actions package cannot be tested in this tree at all — another session's
  untracked, half-written `invalid_banned_claim_pattern_test.go` (missing `sql` import) breaks
  package compilation; my actions-package test edits are verified at committed HEAD via
  `verify-head-builds.sh --test` instead, where untracked files are invisible.
- Existing behaviour pinned rather than trusted: all pre-lever label/judge/choose tests pass
  unchanged; the wiring test's fixtures were updated to the new column shapes (they encode
  "as the SQL returns them" as their own contract).

## 2026-09-02 — committed, verified at HEAD, 714 applied; council executing

- Commits: `215c7eead` (the change, 26 files), `20419fe0f` (ScanShortfall on the new loader — a
  thinned scan silently RE-RANKS a site, so refusal is right, not ceremony; + LNK-041 carries its
  sha), `85518e497` (the literal `scan-loss:accepted` marker — the detector reads the marker, not
  the downstream refusal), `def8126e3` (scan_swallow_baseline 3→2 for
  resolve_internal_links_action.go: deleting loadCTACandidatePages removed a counted swallow and
  the ratchet REQUIRES the baseline drop — a falling count fails the test too, so a gain cannot be
  silently given back).
- Verified at committed HEAD in a git worktree (the shared tree cannot compile the actions
  package): `actions` ok, `discovery_checks` ok, `datahelpers` fails ONLY the pre-existing
  tombstone scan (114's; session messaged, msg 4dfd9db8).
- Migration `714` hand-applied 2026-09-02 ~20:05 BST and `--record-only`'d with the verification
  note: column present, 0 rows opted out, DO/RAISE verify passed. Safe under the running (old)
  binary; closes the "new binary before column" ordering window. `715_HOLD` NOT applied — waits
  for the roll, by design.
- Council run: EXECUTING at `review_debug_historian` as of 18:47Z. Corr
  `9faa2a23-f3bc-464e-8c3a-9d3d44759cc0`.
- Sweeps INTO others' commits, for the record: first WRONG_CALLS row → `6bd26baf0`; LNK-041 index
  row → swept before my commit (visible in `git log -S LNK-041` on 000_concept_index.md). Both
  content-intact; the 710→714 renumber corrected the swept index row forward in `215c7eead`.

## 2026-09-02 — council round 1: REVISE (gating: bug_historian HIGH); round 2 resubmitted

- **The gating objection was the right kind of wrong**: "what happens when the filter leaves the
  slice with <2 or 0 items — index panic or silent no-CTA?" The panic scenario cannot occur (no
  consumer indexes unguarded: chooseCTATargets len-guards; the header guards `primary.URL != ""`;
  the objection's "header takes ordered[0] directly" repeats the BUG FILE's shorthand, not the
  code) and the degrades are designed and pre-tested — but the plan NEVER SAID SO and no test
  exercised the filter-induced near-empty case. Both halves fixed rather than defended
  (a REVISE round is cheaper than the defect it finds): `TestRankAllOptedOutIsEmptyNotPanic` +
  `TestChooseCTATargetsAllOptedOut`, commit `24b871535`.
- debug_historian's two were simply right: 715_HOLD now snapshots
  (`snapshot_agent` TWO-ARG overload — agent_definitions_backup, per LANDMINES) before the
  UPDATE, and a `715_..._ROLLBACK.sql` sidecar exists with its own snapshot + DO/RAISE verify.
- guardian/editquality/reuse_agent asks answered by query, recorded in the resubmission: 3 ranking
  call sites exactly (signatures unchanged, so a 4th could not silently break); no non-test
  `LabelMatchCandidate{...}` construction outside the constructor (zero value = eligible by
  design); runner has no non-empty-HandlerAgent requirement (check_cta_nonpage live precedent);
  no prior fossil-CTA detector anywhere in register/bugs.
- **Round 2 published** on the SAME trail: `SUBMISSION_CORR=9faa2a23-…` (RESUBMIT_CORR chain),
  run envelope `92ce9931`, run orch id `ca1f014c`. Monitor armed for the second council_report.
- 114's session fixed the tombstone test (`d1cf3aac3`) after my flag — datahelpers now FULLY
  green in-tree. The untracked, non-compiling `invalid_banned_claim_pattern_test.go` belongs to
  neither of us and still blocks in-tree actions-package tests; owner unknown.

## 2026-09-02 — council round 2: APPROVED (advisories dispositioned below)

**"approved with 3 advisory objection(s) — none high-severity"; architecture APPROVED** with a
medium noting the shared-contract trigger was met — i.e. ruled-and-passed on the submission's own
transparency, exactly the outcome the same-commit LNK-041 registration exists to earn. Trailer for
subsequent commits: `Council-Reviewed: 9faa2a23-f3bc-464e-8c3a-9d3d44759cc0` (the earlier
`Council-Submitted` commits are credited by 098 at report time; forward-only, no amends).

Advisory dispositions, so none silently evaporates:
- **editquality m (alias visibility):** misread of the sketch, and the code is right —
  `check_misdirected_cta.go` points its var at the EXPORTED `datahelpers.CTAExcludedAreas`, not at
  actions' unexported alias. No change needed; the sketch's phrasing caused it.
- **editquality l (tests bundled in one edit):** the 8-edit cap; full diff was in the commits.
  Process note only.
- **bug_historian l (header outcome unverifiable in DB):** correct and known — it is why the
  induced canary verifies the header AT THE SERVED BYTES (RUNBOOK), and why the unit pins the call
  shape. Roll-bound.
- **bug_historian l (bugs_closed/023 guard):** not bypassed — 023 made ABSENT destinations
  unrenderable (gated templates render no button). The refusal produces exactly that state when
  nothing else supplies the field, so 023's guard is the degrade this design leans ON.
- **reuse_agent m (why a third eligibility axis):** `status='archived'` is page LIFECYCLE — it
  freezes the page, drops it from validPages/listings/nav and invites retraction; `noindex` is
  crawler policy read by head rendering. The owner-approved sayable is "fully live, linked,
  indexed — but never the framework's CTA pick", which neither can express without dragging its
  own semantics along. 391's worked case REQUIRED the page to stay served.
- **guardian m (grep is not verification):** the compile is the verification for an
  exported-signature change — a missed caller cannot build, and HEAD builds+tests green in a clean
  worktree. Recorded here for the next reviewer.
- **guardian l (universe consumers and the flag):** designed in — the judge inherits the refusal,
  so no Contradicts can name an opted-out repair target; the unsatisfiable-finding loop cannot
  form. `names_ineligible_page` is the first-class replacement signal.
- **debug_historian m (HOW the roll is confirmed):** actioned — RUNBOOK §715 now probes
  `service_binary_capabilities` (`kind='discovery_check'`) with a positive control
  (misdirected_cta), the new name, and an absent-control; no shelf life, no log scroll, no image
  tag. This replaces the weaker log-grep recipe.
- **debug_historian l (moved SQL narrowing):** the two supply queries moved VERBATIM plus one
  SELECT column; the wiring tests match on the predicates themselves and pass unchanged.
- **architecture m / guardian l (shared-seam widening):** acknowledged in the submission, ruled by
  the approval; LNK-041 is the same-commit registration.

## 2026-09-03 — the roll landed; 715 applied after the registration proof

- **Roll proven at the artefact, both controls** `[MEASURED 2026-09-03 ~09:20Z]`:
  `service_binary_capabilities` kind=`discovery_check` shows `cta_rank_anomaly` on **412 pods**,
  exactly the positive control `misdirected_cta`'s 412; the negative control `no_such_check_zz`
  is absent. TWO distinct commits are running (`0d2feee2ff61…`, `7bf1ff674021…` — the
  bugs_open/249 straddle shape) and **both** are descendants of `24b871535`
  (`git merge-base --is-ancestor` each), so every pod carries the full change incl. the REVISE
  fixes.
- **715_HOLD hand-applied ~09:22Z** per RUNBOOK: snapshot NOTICE seen
  (source_id `b05773e0-…`), UPDATE 1, DO/RAISE verify passed, COMMIT; post-apply the checks array
  carries `cta_rank_anomaly` (t) and the newest `agent_definitions_backup` row for
  `completeness-discovery-agent` holds the PRE-change config (`has_old = t`, snapshot_taken_at
  2026-09-03 09:22:41Z).
- ⚠ **`--record-only` REFUSES `_HOLD` files** — "UPPERCASE-suffixed sidecar … never applied by
  the runner, so recording one is meaningless." So there is deliberately NO ledger row for 715;
  this NOTES entry + the register status are the application record (same as 643/645). Do not
  "fix" this by hand-writing an INSERT.
- **First fleet pass NOT yet observed**: `site_work_items` `cta_rank_anomaly` count = **0** at
  09:25Z. The discovery rotation must reach sites first. Remember the check's own design: on a
  HEALTHY site it files nothing and RETRACTS (Resolved/AllOfType) — so "0 items" after passes run
  is only meaningful alongside evidence the check RAN (the runner logs the enabled/registered
  arrays per run; or induce one).
