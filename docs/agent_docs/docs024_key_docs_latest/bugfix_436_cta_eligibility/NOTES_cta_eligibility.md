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
