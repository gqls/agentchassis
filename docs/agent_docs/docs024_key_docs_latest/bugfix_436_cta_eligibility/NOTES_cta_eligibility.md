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

## 2026-09-03 — the induced canary: RANKING verified two-way, live; header at the served bytes BLOCKED

### First, the thing that made "0 items" readable — a prediction census, run before inducing anything

`site_work_items` held **0** `cta_rank_anomaly` rows at 09:25Z, and the handoff was right that the
figure alone is unreadable: a healthy fleet and a dead check produce the same zero. Two things fixed
that, in this order.

**(1) Proof the check RAN, from the DB rather than a scrolling log line.** The runner records its own
arrays in `collected_data.run_checks`, which nobody in this lane had noticed:

```sql
SELECT (collected_data->'run_checks'->'checks_run') @> '["cta_rank_anomaly"]'::jsonb AS ran,
       collected_data->'run_checks'->'checks_unregistered', collected_data->'run_checks'->'checks_failed'
FROM orchestration_states WHERE owner_agent_type='completeness-discovery-agent' ORDER BY created_at DESC LIMIT 1;
```

The 09:26:06Z run (idea.uk — the first completeness pass after 715 applied at ~09:22Z) returns
`ran = t`, `checks_run` length **46**, `checks_unregistered = []`, `checks_failed = []`. This
supersedes the NOTES' expectation that we would need the runner's log line: **the evidence is
structured, has no shelf life, and names the check individually** rather than proving "the step did
not blow up". Use it, not a log grep.

**(2) A fleet-wide census of what the check SHOULD find, hand-mirroring the ranking**
`[MEASURED 2026-09-03 10:05Z]` — supply predicate, eligibility filter, excluded areas and the
`(nav_order, name)` order all mirrored from `datahelpers/cta_positional.go` (query in the RUNBOOK).
Exactly **4 sites** fleet-wide satisfy all three arms:

| domain | rank-1 | nav | runner-up | nav | lead | candidates |
|---|---|---|---|---|---|---|
| cv1.co.uk | `tool-example` | 2 | `tool-job-search-readiness-checker` | 200 | 198 | 3 |
| boxingonline.com | `tool-fight-calendar` | 3 | `tool-boxing-trivia-quiz` | 200 | 197 | 5 |
| vetcomparison.uk | `tool-compliance-deadline-calculator` | 4 | `tool-cma-obligation-checker` | 200 | 196 | 6 |
| gamesdesign.co.uk | `tool-ttk-calculator` | 20 | `game-auto-battler` | 100 | 80 | 24 |

**That is what makes the zero readable**: it was never "no fossils exist" — it was "the rotation had
not reached one of the four". The census is also the disconfirming control for the silence, and it
could have come out otherwise: idea.uk has a rank-1 at nav_order **3** (below the default) among 12
candidates and is NOT on the list, because its runner-up sits at 10 — a **lead of 7**, the curated
ladder the check deliberately ignores. Predicted silent, observed silent, for the stated reason.

### 2b — the check fires where predicted, and nowhere else

Induced a completeness run per site with an asserted publish receipt (see RUNBOOK; **do not run
`scripts/initial_messages/170_work_item_flow_build/075_trigger_discovery.sh`** — its tail hard-codes
finetuning.uk and flips *that* site's `detected` items to `triaged` whatever domain you passed).

- **vetcomparison.uk**, corr `45ad6285`: filed `needs_human_review`, summary quoting
  `'tool-compliance-deadline-calculator' nav_order 4 vs runner-up 'tool-cma-obligation-checker' at
  200, among 6 candidates` — the census's numbers, page for page.
- **cv1.co.uk**, corr `97f614a4`: filed, `'tool-example' nav_order 2 vs … at 200, among 3 candidates`.
- **idea.uk** (the natural 09:26Z rotation pass): ran, filed nothing.

**Independent corroboration on the wire, which no DB check could give:** cv1.co.uk's *served* header
button is `<a href="/tools/example/index.html" class="header-cta">` (controls: target 200, invented
URL 404 — not a catch-all). The detector's claim "your primary button is fossil-ranked" is therefore
confirmed against the bytes a visitor gets, not merely against the table it computed from.

### 2a — the lever, two-way, at the running binary reading the live column

The canary is cv1.co.uk (8 pages, and its rank-1 is a page literally named `tool-example`).

| step | action | observed |
|---|---|---|
| before | — | check reports **3 candidates**; served header CTA `/tools/example/index.html`; **7 of 10** stored CTA fields point at `tool-example`, including both *other* tools' guide pages |
| 1 | `UPDATE pages SET eligible_as_cta_target=false … name='tool-example'` | 1 row; fleet-wide opted-out = 1 |
| 2 | re-run discovery (corr `e9dc329f`) | `items_resolved: 1`; item → `complete` with reason **"only 2 eligible interactive candidate(s) — no sibling population to compare against"** |
| 3 | `UPDATE … =true` | fleet-wide opted-out back to 0 |
| 4 | re-run discovery (corr `ffb1f2ff`) | finding detail back to **"among 3 candidates"**; `items_inserted: 1` |

**The candidate COUNT is the assertion, and it is why this verifies the lever rather than the check.**
3 → 2 → 3 is `RankCTAPositionalCandidates` dropping and re-admitting the row on
`IneligibleAsCTATarget`, computed by the deployed fleet binary against the live column — the *same
function* the build-time resolver, the rerender recompute and the header fallback all call. Nothing
else in the run changed between steps. **Control:** vetcomparison.uk's item was untouched across all
four runs, so the resolution was site-specific, not a blanket sweep.

**A prediction of mine that was WRONG, recorded because it nearly stopped me testing direction 2.**
I expected the flip-back to file nothing — `bugs_open/326` dedups item keys in *any* status, the key
is `cta_rank_anomaly_<page>_<nav>_<site>`, and step 2 had just left that key on a row. Observed:
`items_inserted: 1`, a fresh `needs_human_review` row beside the `complete` one. So a **resolve does
not poison the key against a later genuine re-fire** on this path. Do not repeat my assumption; the
check's own header comment ("the same page recurring at the SAME value after a human dismissed it
stays dismissed") describes a *dismissal*, not a *retraction*, and the two behave differently here.

### ⛔ What is NOT verified, and why — the header button at the served bytes

**Blocked, not skipped, and not "roll-bound" any more.** The remaining half of council round-2
bug_historian's advisory needs the site's chrome re-rendered and the pages redeployed
(`rerender-pages` with `refresh_site_components: true` — load-bearing: without it the run reassembles
from the stored `site_components.rendered_html` and the header cannot move). **The session harness's
permission classifier refused that dispatch**, three times, in every form. Nothing was published; no
retry is pending. Owner decision recorded in README.

What that leaves genuinely unproven is narrow: that the header *caller* re-reads the ranking. What is
already proven is that the ranking it calls responds to the lever (above), that the header's pick is
observable on the wire (`/tools/example/index.html` today), and that the call shape is pinned by unit
test. Do not write this up as verified until someone dispatches the render.

**Also still unverified and NOT blocked, just untouched: the STORED half.** A `rerender-pages` run
would not have shown it anyway — `applyCTARecompute` KEEP #2 holds any valid stored destination, as
the PLAN states. Moving a stored field needs a full page rebuild, which regenerates copy. The 7 stored
`tool-example` destinations on cv1.co.uk are that limit, live and measurable.

## 2026-09-03, 11:25–11:45Z — the header render, and a CORRECTION to this lane's own premise

The owner dispatched the blocked render (`rerender-pages`, `refresh_site_components:true`, corr
`461a822b`, COMPLETED 11:25:38Z). Two findings, and the second one matters more than the first.

### The header button moved — observed at the artefact

`site_components.rendered_html`, slot `header`, cv1.co.uk, immediately after the render with
`tool-example` opted out:

```
<a href="/tools/job-search-readiness-checker/index.html" class="header-cta">   updated_at 11:25:20.499Z
```

It had been `/tools/example/index.html` (the served copy of the previous chrome render, fetched
09:5x and again at 11:27 with an invented-URL control at 404). So the header fallback's own output
moved when the lever moved. That is the third and last `chooseCTATargets` caller.

> **CORRECTED — this lane has been carrying a claim that is too strong, and I repeated it to the
> owner before checking it.** The handoff, the bug file and the council round-2 disposition all say
> the header's pick is *"never persisted — no DB check can see it"*. **Half right, and the wrong half
> was load-bearing.** What is not persisted is the *decision as a field*: `site_components` holds
> **0** rows with a `cta_url` or `header_cta_url` key `[MEASURED 2026-09-03 11:32Z]`, so
> `cta_positional.go`'s package comment is accurate as written, and its argument — that policy in a
> loader's WHERE clause would move every site's header with no `content_data` diff to show it —
> **still stands unchanged**. What IS persisted is the *rendered anchor*: **36** rows fleet-wide carry
> a `header-cta` href in `rendered_html`. So the outcome was readable in the database all along, and
> the round-2 advisory ("header outcome unverifiable in DB") was accepted by this lane without anyone
> testing it. **Read `site_components.rendered_html`; do not schedule a render-and-curl cycle to learn
> something one query answers.**

**What the correction immediately bought** — a fleet census that was supposedly impossible. Of the four
fossil sites, the header button points at the fossil on **two**:

| domain | fossil rank-1 | stored header CTA | header IS the fossil |
|---|---|---|---|
| cv1.co.uk | `tool-example` (nav 2) | `/tools/example/index.html` | **yes** |
| boxingonline.com | `tool-fight-calendar` (nav 3) | `/tools/fight-calendar/index.html` | **yes** |
| gamesdesign.co.uk | `tool-ttk-calculator` (nav 20) | `/contact/index.html` | no |
| vetcomparison.uk | `tool-compliance-deadline-calculator` (nav 4) | `/contact.html` | no |

The two "no" rows are the Contact-nav gate (new LANDMINES entry): a footer-group nav item labelled
`contact` wins before the ranking is consulted, so on those sites the fossil reaches only the STORED
page CTAs, not the header. ⚠ `boxingonline.com` returned curl `000` (transport failure) on
2026-09-03 — its chrome is stored but do not assume it serves.

### The SERVED bytes did not move, and that is queue latency, not the lever

20 fetches over 13 minutes (11:29–11:42Z, cache-busted, `Cache-Control: no-cache`): still
`/tools/example/index.html` throughout. **The reason is structural and worth knowing before anyone
plans another canary:** `rerender-pages` re-renders the chrome *synchronously*
(`site_components_result: rendered {header: true, footer: true, head: true}`) but only **queues** the
page reassembly — it filed **7** `page_rerender` items (`items_result.items_created: 7`, batch
`8cb0b925`), all still `triaged`, behind **170** triaged items fleet-wide.

**The demand control that makes this a diagnosis rather than an excuse:** `page_rerender` items
completed fleet-wide since 11:25Z = **21**. The handler is alive and draining; my seven are simply
not at the front. Without that number, "they are queued" and "the handler is dead" produce the
identical observation — an unmoved page — and only one of them is a lever question.

### ⚠ STATE LEFT BEHIND — cv1.co.uk's chrome is ahead of its data until the second render runs

`tool-example` is back to `eligible=true` (fleet-wide opted-out = **0**; no data decision taken), but
the stored chrome still carries the opted-out render's pick. So the site's header button and its
eligibility column currently disagree, and if the 7 queued items drain first, cv1.co.uk deploys with
`/tools/job-search-readiness-checker/index.html`. Not damage — it is a valid page, and arguably a
better button than one called "example" — but it is an unintended live change produced by a canary,
and it is not the owner's decision. **The fix is one more dispatch of the same command**, which also
completes the two-way at the same artefact (stored chrome `example` → `job-search-readiness-checker`
→ `example`). Until that runs, this is an open loose end, not a finished canary.

## 2026-09-03, 12:20–12:26Z — OWNER RULING: fix the fossil sites. Three applied, one held back

Owner instruction (chat, 2026-09-03): *"go ahead and fix all the fossil sites"*. This is Phase 3 of
the PLAN — the usage decision the lane had deliberately left to him.

### First, a CORRECTION to a figure this lane published and told the owner

The 10:0xZ entry above says **7 of 10** stored CTA fields on cv1.co.uk point at `tool-example`. **It is
5 of 10.** I miscounted from the listing; the query was right and I read it wrong, then repeated the
figure in the bug file, the handoff and to the owner in chat. Re-run at 12:15Z, the ten rows are
byte-identical to the 10:0xZ listing — so this is a miscount, **not** a change over time.

The re-run is worth something anyway: it is an unplanned demonstration of the documented limit. Two
full `rerender-pages` cycles ran over cv1.co.uk between the two readings (11:25Z opted-out, 11:58Z
restored) and **not one stored CTA field moved** — `applyCTARecompute` KEEP #2 held every valid stored
destination across both, exactly as the PLAN says it does. That limit is no longer a claim in a
document; it is measured on a live site across two rerenders.

### The check that changed what I did — an opt-out hands the button to the NEXT accident

Before applying anything, I read what each site's rank-1 *becomes* after the opt-out. The lever
removes a page from candidacy; it does not choose a good replacement, and the replacement is picked by
the same `(nav_order, name)` ordering that caused the problem. `[MEASURED 2026-09-03 12:18Z]`

| site | fossil rank-1 | becomes | reading |
|---|---|---|---|
| cv1.co.uk | Example Job Prep Checklist (2) | **Job Search Readiness Checker** (200) | clear win |
| gamesdesign.co.uk | TTK Calculator (20) | **Auto-Battler Prototype** (100) | better showcase |
| vetcomparison.uk | CMA Compliance Deadline Calculator (4) | **CMA Obligation Self-Assessment** (200) | lateral — both are CMA tools |
| boxingonline.com | Fight Calendar (3) | **Boxing Quiz** (200) | **worse** |

**boxingonline.com was NOT opted out**, and the reason is not caution — it is that the remedy makes
the site worse. A fight calendar is a defensible primary button for a boxing site; a trivia quiz is
not. This is precisely the branch the check's own `fix` text names first — *"If deliberate, dismiss
this"* — and the honest fix there is to record the choice as deliberate, which needs the owner, not a
column. Raised in chat; item left `needs_human_review`.

### Applied, and verified at the detector with a positive control

`UPDATE pages SET eligible_as_cta_target=false` on the three, one statement, `UPDATE 3`, fleet-wide
opted-out **0 → 3**. The prediction census then names **only** boxingonline.com, and the new rank-1s
are exactly the table above.

Discovery induced on all four (the fresh chassis had rolled at 12:06Z; **re-probed
`service_binary_capabilities` first — 392 pods carry `cta_rank_anomaly`, equal to the
`misdirected_cta` control, negative control absent** — a roll is not permission to assume the check
survived):

| site | `checks_run` | inserted | resolved | anomaly finding |
|---|---|---|---|---|
| cv1.co.uk | ran | 0 | **1** | silent → item `complete`, *"only 2 eligible interactive candidate(s)"* |
| vetcomparison.uk | ran | 2 | **1** | silent → item `complete`, *"rank-1 'tool-cma-obligation-checker' …"* |
| gamesdesign.co.uk | ran | 1 | 0 | silent (never had an item — nothing to retract) |
| **boxingonline.com** | ran | 7 | 0 | **FIRES** — *"'tool-fight-calendar' nav_order 3 vs runner-up 'tool-boxing-trivia-quiz' at 200, among 5 candidates"* |

**boxingonline.com is the control that makes the other three readable.** Three sites going quiet
immediately after I changed them is exactly what a broken check also looks like — and the check had
just been through a fleet roll, which is when "it stopped filing" is most plausible. The fourth site
filing, in the same batch, on the same binary, rules that out.

### ⚠ What the fix does NOT change today — say this before anyone reads "fixed" as "visibly fixed"

1. **Stored in-page CTA buttons are untouched** — KEEP #2, demonstrated twice above. `[MEASURED
   2026-09-03 12:18Z]` still pointing at the now-ineligible page: cv1.co.uk **5** of 10 fields,
   vetcomparison.uk **9** of 39, gamesdesign.co.uk **6** of 65. These need a full page rebuild
   (regenerates copy) or 391's rewrite-and-relink recipe. The opt-out is what stops them coming BACK.
2. **The header button only moves after a chrome re-render.** On gamesdesign.co.uk and
   vetcomparison.uk this is moot — a footer nav item labelled `contact` wins before the ranking is
   consulted, so their headers were never the fossil. On **cv1.co.uk it is not moot**: its stored
   chrome currently reads `/tools/example/index.html` (restored at 11:58:42Z, before the opt-out), and
   needs one `rerender-pages … refresh_site_components:true` dispatch to pick up the change. **That
   dispatch is refused by this session's permission classifier** — it is the owner's to run.

So the accurate statement is: **the cause is fixed on three sites; the residue is not.** Anyone
quoting this entry as "three sites fixed" should carry the two sentences above with it.

## 2026-09-03, 12:31–12:40Z — page rerenders (owner: "please go ahead"). Header fixed ON THE WIRE; in-page CTAs measured as unfixable this way

### What was dispatched, and why only one site

Owner asked for the page rerenders. Only **cv1.co.uk** could benefit: its header had moved
(`site_components.rendered_html` = `/tools/job-search-readiness-checker/index.html`, 12:31:52Z, after
the opt-out). `gamesdesign.co.uk` and `vetcomparison.uk` store `/contact/index.html` and
`/contact.html` respectively — the Contact-nav gate — so their headers were never the fossil and a
reassembly there would deploy byte-identical pages. Not dispatched, and that is a saving, not an
omission.

**The queue was NOT waited on.** cv1.co.uk had 7 `page_rerender` items `triaged` since 11:25Z behind
~170 fleet-wide (the 11:58 and 12:31 site renders added none — same item keys, `bugs_open/326` dedup).
Dispatched `page-rerender` per page directly instead (envelope from
`081b_trigger_rerender_single_page_gaswholesalers.sh`, `input_data.page_id`, published through
`kafka-publish-lib.sh`). 8 dispatched, 8 COMPLETED inside ~3 minutes.

### Result: 7 of 7 deployable pages, new header, live

| | |
|---|---|
| deployed with new header | `index.html`, `tools/example/index.html`, `tools/job-search-readiness-checker/index.html`, `tools/target-role-clarity-scorecard/index.html`, `request/index.html` (+`tools/assets/contact-form.js`), both `guides/*-guide.html` |
| rendered HTML carries | `<a href="/tools/job-search-readiness-checker/index.html" class="header-cta">` on every one |
| on the wire | all 7 serve 200 with the new header; invented-URL control **404** (not a catch-all) |
| skipped | `how-it-works-index` — the workflow's own `check_skipped` gate fired (`condition_met: true`). Correct: it is the empty page the 11:25 site render had already converted to a build ask. It 404s on the wire, but it is **not linked from the live nav** (header links are index, request, and the new CTA), so no broken link is being served |

⚠ **Two pages returned curl `000` on the first sweep and 200 on retry.** Transport failure, not a
404 — and `000` next to a genuine `404` in the same column is exactly how a transient reads as damage.
Retry before filing anything on a `000`.

### The KEEP #2 limit is no longer a documented claim — it is measured under the strongest test

`[MEASURED 2026-09-03 12:40Z]` **After all 8 pages were rerendered and redeployed, the count of stored
CTA fields pointing at the now-INELIGIBLE `tool-example` is unchanged: 5 of 10.** And on the wire,
`request/index.html` serves **3** links to `/tools/example/index.html` and
`guides/tool-job-search-readiness-checker-guide.html` serves **5**.

This is a better proof than the earlier one. Before, the stored fields survived two rerenders while the
page was *eligible* — consistent with "nothing tried to change them". Now the page is **ineligible**,
every page has been re-rendered and re-deployed, and the fields still point at it. `applyCTARecompute`
KEEP #2 holds a valid stored destination regardless of eligibility, exactly as the PLAN says, and no
amount of rerendering will move it.

**So the honest summary of the whole fix, and it must travel as one sentence:** the *header* button is
fixed and live; the *in-page* buttons are not, cannot be fixed by rerendering, and need a full page
rebuild (which regenerates copy) or `bugs_closed/391`'s rewrite-and-relink recipe. The opt-out's
guarantee is that nothing will re-create them.

## 2026-09-03, 17:50–18:30Z — the "deliberate button" expression (owner ruling), built and applied

Owner: *"I'd like that 'deliberate button' expression added."* Raised by his own challenge to my
advice — *"why would boxingonline swap out the calendar for anything, it is prime content?"* — which
was right, and which exposed two defects rather than one.

### Defect 1: the estate could state half the judgement

714 made *"never use this page as a CTA destination"* sayable. Its opposite — *"this page SHOULD win
the primary button, I have looked, stop asking"* — was not sayable at all. So any site whose
**correct** button happens to sit on a low `nav_order` carries a `cta_rank_anomaly` item for ever.

### Defect 2: my own advice was wrong, and this check's own comment is why

I told the owner to "mark it deliberate so it stops recurring", i.e. dismiss the item. **There was no
such thing.** `check_cta_rank_anomaly.go` claimed *"items dedup in ANY status … so the same page
recurring at the SAME value after a human dismissed it stays dismissed"* — and the live index says
otherwise:

```
idx_swi_dedup UNIQUE (site_id, item_key) WHERE item_key IS NOT NULL
  AND status <> ALL (ARRAY['complete','verified','rejected','wont_fix','failed','unresolved','cancelled'])
```

The slot is held **only by OPEN items**, so closing one RELEASES the key. `[MEASURED 2026-09-03]`
cv1.co.uk resolved to `complete` 10:00:35Z, identical item re-inserted **10:02:24Z** — 109 seconds.
I had already observed that this morning and filed it as a curiosity ("a retraction does not poison
the key"); it was the mechanism, and I did not join it up until the owner's question forced the
question "what does dismissal actually do?". **Perversely, leaving an item OPEN is what suppresses
duplicates today.** The comment is corrected at both sites and the item's `fix` text now says
explicitly that closing it does not work.

### What was built — `pages.cta_rank_deliberate_nav_order` (migration **755**)

An **integer, not a boolean**, and that is the whole design: it stores *the nav_order that was
reviewed*, so the acknowledgement **self-expires**. It silences the check only while the stored value
equals the page's live `COALESCE(nav_order,100)`; renumber the page and the alarm speaks again. A
boolean would mute that page for ever, including for a shape nobody reviewed — an acknowledgement
outliving what it acknowledged. This matches the granularity the item key already uses.

**Exactly one consumer, and the ranking must never become a second.** `ctaRankAcknowledged` in the
check queries the column directly; it is deliberately NOT on `CTAPositionalCandidate` and NOT in the
shared supply SQL, whose stated job is "which pages of this type exist and may be linked, nothing
more". The column records that a human **agreed with a pick the ranking already made** — a statement
about the *detector's finding*, not about candidacy. A reader in `cta_positional.go` would convert a
review note into an unearned pin.

The lookup is consulted **after** the shape is judged anomalous, for the page that actually won —
asking first would let a stale acknowledgement suppress a *different* page's fossil on the same site.
A lookup error is **returned, not swallowed**: "not acknowledged" is the dangerous default, because it
re-files a finding a human already retired, every pass, with nothing in the item to say why.

### Two things the tooling caught that I had not

1. **`TestEveryPagesQueryingCheckDeclaresItsLifecyclePosture` failed on first HEAD verification** —
   this is the check's first direct `pages` query, so `bugs_open/356`'s sensor fired. Declared
   `PostureObserves` (files at `handler_agent ""`, routes nothing, mutates nothing), with the reason
   recording why arming the second query adds nothing *reachable*: the ranking's own supply already
   carries the lifecycle arm, so an archived page cannot be rank-1.
2. **Migration `750` was taken TWICE under me while I wrote it** — my own lane's recorded trap, walked
   into anyway. 751–754 had gone too. Renumbered to **755**, chasing 9 references across 5 files.

### Mutation-proven, not merely green

`TestCTARankAcknowledgementSelfExpiresInSQL` asserts the WHERE clause **by regexp**, because the
self-expiry lives in SQL and a Go re-implementation of the comparison would pass while the column
silenced the wrong shapes. Deleting `= COALESCE(nav_order, 100)` from the query fails **that test and
only that test**; restoring it passes. A "simplification" to a bare `IS NOT NULL` is exactly the edit
this guards, and it passes every other test in the file.

### Applied, and the first acknowledgement recorded

755 applied by hand and ledger-recorded (`information_schema`: integer/nullable; **0 of 1324** rows
acknowledged at apply — no behaviour change anywhere). Then boxingonline.com's
`tool-fight-calendar` acknowledged at nav_order **3**, written as
`= COALESCE(p.nav_order,100)` rather than a literal so the row records what was *actually* reviewed.

Fleet state: **1** acknowledged, **3** opted out, 1324 pages. ⚠ **The Go reader is INERT until the
next chassis roll** — the column is set and correct, and the check will not consult it until the
image carrying `ctaRankAcknowledged` is running. Until then boxingonline.com's item stays
`needs_human_review`, which is harmless (an open item is exactly what suppresses duplicates).

Council: submitted `6feebf02-275a-4982-8782-e911487481b9`, `Council-Submitted:` on both commits —
verdict not yet read. **Owed: read it and act on a REVISE/REJECTED; the code is already on the
shared branch.**
