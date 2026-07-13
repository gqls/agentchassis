# PLAN — F0.1 plumbing, then the dartsonline pilot as a known-answer benchmark

Written 2026-07-09; current through turn 20 (2026-07-10). Supersedes the "order of work" paragraph in the intake.

**HANDOFF NOTE (turn 21):** This plan is complete through F2.2. The revise loop is deployed and a demo run (`e08c5b01`) is in flight. See `HANDOFF_turn21_2026-07-10.md` for checkpoint queries and next steps. Companion documents: RUNBOOK_diagnosis_fix_loop(10).md (task statement, what exists, phases, boundaries), NOTES_running_fixloop(10).md (evidence trail and the reasoning that produced this plan).

## 0. What changed, and why the plan changed with it

The intake ordered: land F0.1 plumbing, then run the guides bug through it as
F0's acceptance test, with success = "a diagnosis carrying a static-tier
citation naming the code that drops or fails to route guide pages".

The mandatory pre-check ran first, as doctrine requires. **It produced that
citation by hand, in about a dozen queries and four file reads** (evidence in
NOTES(10), turn 1). The standing hypothesis was refuted; the real mechanism is
a success-labelled error terminal plus a nav query filtering on the wrong
column. The pilot can therefore no longer discover anything.

The plan below keeps the original order — plumbing first — and changes only
what the pilot *is*: from a discovery run whose answer nobody knows into a
**graded benchmark whose answer we hold**. This is a strictly stronger F0
acceptance test. Four of the five pre-registered pilot criteria were always
about the plumbing; only the fourth was about the answer, and that is the one
we can now mark objectively instead of taking on trust.

## 1. Slice F0.1 — the plumbing (do this first; unchanged in substance)

Three thin slices, each independently landable, each with criteria fixed
before the code is written.

### F0.1a — `diagnosis_artifacts` table — ✅ LANDED 2026-07-09
Applied to `clients_db` from `0NN_diagnosis_artifacts.sql` after
`verify_before_migration_diagnosis_artifacts.sql` returned clean. Columns as
designed (`correlation_id` is **text**, not uuid — the chassis does not
guarantee uuid form), plus `orchestration_id`, nullable `site_id`, `metadata`
jsonb, and the `expires_at`/`pinned` retention knob. Partial unique index on
`(correlation_id, iteration) WHERE kind='bundle'` gives retry-safety for
bundles while leaving `iteration_note` free to have several rows per iteration
(per-step notes, F0.3).

*Criteria — all pass*: applies clean; **idempotent** on re-apply; both kinds
round-trip; the `kind` CHECK rejects a third kind; the `iteration` CHECK
rejects 0; the partial unique rejects a duplicate bundle but permits a second
note. Verified additionally, because F0.1b depends on it:
`ON CONFLICT (correlation_id, iteration) WHERE kind='bundle' DO UPDATE`
infers the partial index and replaces in place. **Use that exact clause.**

### F0.1b — write-through inside the assemble action — ✅ CODE-COMPLETE 2026-07-09 (not yet deployed)
Landed in `DiagnoseAssembleBundleAction`, immediately before its existing
return. Zero workflow-shape change; zero contact with the tools chat's live
`emit → persist_note → complete` surface. New optional config: `persist_bundle`,
`iteration_field`, `site_id_field`, `bundle_retention_days`.

*Criteria*: a write failure degrades to a logged warning and **never** fails the
diagnosis — enforced on all four failure paths (nil DB, missing correlation_id,
marshal error, INSERT error). Iteration derivation unit-tested, including the
bare-`diagnose_state` trap. The production SQL was executed against the live
table with typed params: partial-index conflict inference, retry-replaces,
NULL site_id, and both retention modes all confirmed.

*Remaining*: "one `kind='bundle'` row per iteration for a real run" cannot be
observed until the chassis image is rebuilt and rolled out — the pod runs a
built binary. **F0.1b is code-complete, not live.** That proof arrives with the
benchmark run (§5, step 6).

### F0.1c — the `needs_diagnosis` envelope — ✅ LANDED 2026-07-09
`090_TRIGGER_needs_diagnosis_v1.sh` writes the durable intake record, then fires
the 084 envelope on the same `correlation_id`, so item → bundles → terminal note
all join on one key. `DISPATCH=0` records without firing. 084 is retained for
ad-hoc runs.

**Q-B corrected.** "Null-site allowed" was impossible: `site_work_items.site_id`
is NOT NULL, *and* `LoadWorkItemsAction` requires a uuid and queries
`WHERE wi.site_id = $1`, so a NULL-site item could never be loaded anyway.
Instead we reuse the existing `system.internal` pseudo-site
(`eac60db8-…`, `sites.status='system'`) that already carries platform-wide
maintenance work. **Every** `needs_diagnosis` item anchors there — even
site-specific bugs — because `build-dispatch-loop`'s `load_items` has **no
`item_pipeline` filter** and would otherwise claim diagnose items parked on a
real site. Items start at `status='detected'`, outside the loader's
`('triaged','approved')` filter, as a second guard.

*Criteria — all pass*: correct row shape with the envelope in `spec`; the
loader's exact query returns 0 rows against `system.internal`; a **negative
control** (flip to `triaged`) makes it appear, proving the status guard is what
holds it and that the dispatch hazard is real; re-running the same `SLUG` is
idempotent via `idx_swi_dedup`.

### F0.1d — automatic dispatch — ✅ LANDED 2026-07-09, SHIPPED DISABLED
`0NN_diagnose_dispatch_loop.sql`: the `diagnose-dispatch-loop` agent (image
columns copied from the `build-dispatch-loop` donor) + the
`diagnose-pipeline-trigger` scheduled task on a 60s tick, `max_concurrent=1`.

The `diagnose` namespace cannot ride the shared claim path: neither
`build-dispatch-loop.load_items` nor `build-pipeline-trigger.find_dispatchable_site`
filters by pipeline, `triage_detect_items` rewrites `pipeline` to `'build'`, and
`claim_work_item` claims only `triaged|approved`. So diagnose items use two
private statuses — `awaiting_diagnosis` → `diagnosing` — that no sweep names,
claimed atomically by the loop's own `UPDATE … SKIP LOCKED … RETURNING`. Because
that opts out of `claimed-item-timeout`, the loop reaps its own dead runs
(`reap_stuck`, 75 min).

*Criteria — all pass*: inertness matrix scores 0 against all six sweeps in both
states; positive control confirms our claim still finds it; the claim returns the
**target** site from `spec`, not the anchor; a second tick returns 0 rows; the
reap turns a 76-minute-old run terminal (`failed`), never `triaged`; and a
diagnosis in flight no longer blocks build dispatch for `system.internal`.

**The trigger ships `enabled=false` on purpose.** Enable only after the image is
live *and* §3's blinding is confirmed — otherwise the loop claims the pilot item
and runs it before we have checked it cannot read the answer:
`UPDATE scheduled_tasks SET enabled = true WHERE name = 'diagnose-pipeline-trigger';`

*Routed to the builder thread, not fixed here*: the `maintenance` pipeline is
dispatched only because `build-dispatch-loop` lacks a pipeline filter, so the
obvious one-key fix would orphan it. And `triage_detect_items`' comment asserts a
filter that does not exist. Both are the pilot bug's family — a rule enforced in
one place and not its partner.

**Gotcha to honour while touching any workflow**: `error_step` belongs *inside*
a step's `config`. Step-level `error_step` is silently ignored (001 §16).
Note that `page-build-handler`'s live definition carries `error_step` at
*both* levels on `deploy_page`, `plan_sections`, `save_sections`,
`validate_content` and `load_spec_sections` — the inner one is load-bearing,
the outer is dead. Correct adjacent instances as a noted change if we touch it.

## 2. Regenerate the bundle — ✅ DONE 2026-07-09

`z_bundles/BUNDLE_fixloop_F1.md` (306,897 B) is rebuilt **with** `-psql`: schema,
capabilities and runtime rows all present, 468 files analysed. The deficient
`BUNDLE_fixloop_F0.md` (199,579 B, code+docs only) is kept for comparison.

Full procedure, verified end to end, is in RUNBOOK §"REGENERATING THE CONTEXT
BUNDLE". Three points that cost time to discover: contextkit is **not** at
`cmd/bundle` but at `docs/…/docs019_…/go_files/contextkit`, and you must `cd`
there because it shells out to relative `./cmd/dbcontext` and `./cmd/assembler`;
`-schema-tables` must include `site_plans`/`site_plan_pages` (omitted originally,
and where the guides evidence lives) plus `diagnosis_artifacts`; and the heading
`## Runtime evidence` is *always* a placeholder — the real runtime rows arrive as
a `-doc` under "Recent errors" and "Work-item lifecycle". Exactly one placeholder
means success; two means a gather silently failed.

## 3. The benchmark — pre-registered scoring rubric

The loop receives **only the original symptom string**, with no hint of the
answer:

> "dartsonline.com published a Guides nav link and a /guides/index.html page,
> but the page is blank and no guide pages exist — while gamesdesign.co.uk, on
> the same platform, has working guides (and games and tools), and
> gaswholesalers.com has a working news feed."

Score its emitted diagnosis against these, fixed now, before the run:

| # | Claim the loop must reach | Tier | Weight |
|---|---|---|---|
| 1 | `pages.sections` empty ⟺ page not built (the exact 5/10 partition) | live data | must |
| 2 | `check_has_ready_sections` routes sectionless pages to `complete_error` | static | must |
| 3 | `complete_error` is a `complete_workflow` — a **success** terminal — so the item is stamped `complete` | static | must |
| 4 | nav selects on `pages.status`, not `build_status` (`populate_nav_tables_action.go:242-243`) | static | must |
| 5 | the `page-build-handler` work-item `result` lacks `deploy_result` for the 10 | runtime | should |
| 6 | gamesdesign's guide pages have sections and use the same handler ⇒ handler is not the discriminator | live data | should |
| 7 | `mark_no_sections` is referenced in a comment and does not exist | static | bonus |
| 8 | the two intake paths (`WriteBuildItemsAction` vs `reconcile_site_plan`) disagree on `unavailableBuilders` | static | bonus |

| 9 | the four "guide" pages are `blog-post` at `/blog/*.html` with `site_area_id IS NULL`, so `resolveSectionIndexForType` can never bind them to `guides-index` | live data + static | bonus |

*(Claim 9 added 2026-07-09 **after** the first run, which discovered it. It is
recorded as a bonus, not a must, and it was **not** part of the pre-registered set
— saying so is the point of pre-registration.)*

**RESULT OF RUN 1 (2026-07-09, correlation `4d43d002-…`): 0 of 4 musts; verdict
`CONFIRMED` anyway; refutation credit passed; claim 9 discovered.** Full scoring
and the three engine defects it exposed are in RUNBOOK §BENCHMARK RESULT and
NOTES(10) turn 6.

**Pass** = all four *must* claims, each with a citation that resolves to the
named file/step. **Refutation credit**: the loop should *not* assert the
standing hypothesis (that `reconcile_site_plan`'s routing table drops guide
pages). If it asserts it, that is a scored failure regardless of the rest —
the hypothesis is false and a citing loop that confirms a false hypothesis is
worse than one that abstains. **Abstention** on any *should*/*bonus* row is
neutral, not penalised: cite-or-abstain is the contract.

Also measured, independent of the answer: (a) intake landed via the documented
route; (b) every iteration's bundle is fetchable from `diagnosis_artifacts`;
(c) per-iteration notes were written; (d) iteration count and wall-clock, as
the first entry in the loop's own performance record.

**Blinding — narrower than first assumed, and now established.** The loop cannot
structurally reach this directory: `diagnose-agent` reads only the bodies of
in-scope **Go symbols** from a checkout, the analyser walks Go source, and the
live workflow has **no doc step** (verified 2026-07-09). The context bundle is for
humans, not the loop. So only two things can leak the answer, and both are on us:
(1) the **symptom string** — pass the original verbatim, it describes only what a
user could observe; (2) **`seed_scope`** — run with **none**. Seeding it with
`populate_nav_tables_action.go` or `load_work_item_actions.go` hands the loop the
answer. Absent a seed, assemble falls back to `lookup_code_symbols`' `code_results`,
which is the honest starting point.

## 3b. F0.4 — engine fixes from benchmark run 1 (added 2026-07-09, turn 7)

Run 1's failure decomposed, on evidence, into: the verdict **never sees the
original symptom after iteration 2** (bundles 3–5 carry only the drifted
hypothesis — proven by grep over the persisted bundles); the static tier is
Go-only while cause B lives in workflow JSON; retrieval scoped the right file's
wrong symbol; and the engine enforces **no tier coverage** (its only confirm
guard is citations ≥ 1, advance.go:93).

Slices, by ownership: **F0.4a** symptom anchor in assemble (Go, ours) —
✅ CODE-COMPLETE 2026-07-09; **F0.4b** follow-the-error-log workflow-step
enrichment (Go, ours) — ✅ CODE-COMPLETE, its SQL verified live against run 1's
actual `page-build-handler/complete_error` ref; **F0.4c** same-file sibling
signatures (Go, ours) — ✅ CODE-COMPLETE, unit test pins the
isLegalPage/loadPagesForNav gap; **F0.4e** tier-coverage guard (Go, ours) —
✅ CODE-COMPLETE via a shared `coerceVerdict()` (three duplicated coercion
blocks unified), CONFIRMED now requires static + state|runtime, REFUTED exempt;
**F0.4d** symptom-closure gate on CONFIRMED — ✅ BUILT 2026-07-10.
Engine half (ours): the gate lives in the shared `coerceVerdict` — CONFIRMED
without a `symptom_check`, or with any `explained:false` entry, degrades to
Unverifiable with the residue named, so the loop works the residue instead of
stopping on a half-answer; conclusion renders a "Symptom coverage:" block.
Prompt half (tools chat's surface): hard rule 8 + schema entry, applied
fetch-first with snapshot `34f4afc8` and an FYI filed in travelling_docs.
**RUN 3 RESULT (2026-07-10, corr `5120c0dc`, v1.0.1102): PASS on the primary
criterion.** 3 iterations, ~16 min, honest UNVERIFIABLE (scope-not-narrowing)
with a precise gap list naming `complete_error`/`sections=[]` as prime suspect
and "hand to a human; do NOT auto-conclude". **The tier guard fired in
production** — iteration 2's state-only CONFIRMED was coerced with F0.4e's
exact message. No gaming observed. Full record: NOTES(10) turn 12.

**F0.5 — persist data_request answers across iterations — ✅ CODE-COMPLETE
2026-07-10 (from run 3):** re-run, don't store. `diagnose_route` now forwards
the UNION of the current verdict's requests and the engine's accumulated
`SeenRequests` keys (already raw SQL, already round-tripping in state for the
spin guard) — deduped, sorted, capped at 12, prior keys re-linted read-only.
`load_runtime` re-runs them every iteration under its existing caps, so
answered evidence persists without touching collected_data size (the cd-bloat
constraint that ruled out storing answer text in state). One file changed;
five unit tests incl. the run-3 hole and a poisoned-state write statement
being refused. **RUN 4 RESULT (2026-07-10, corr `5179a2ea`, v1.0.1103): the first
full-coverage CONFIRMED** — 2 iterations, ~8 min, all three guards passed
legitimately, five-entry symptom_check rendered in the conclusion. The
blank-page chain is right and cited (must-claim 3 again, via the F0.4b step
definition); the nav clause is now VISIBLE in coverage but explained shallowly
("the nav row exists") — must-claim 4 remains unreached. F0.5 went unexercised
(no iteration 3). Full honest scoring: NOTES(10) turn 14.

**F0.6 — ✅ BUILT 2026-07-10:** `context` disposition + citation-backed
`explained` (`cites` indices, in-range required). Prompt live (snapshot
'pre-F0.6', FYI addendum filed); engine gate awaits the next image. Run 4's
verdict would not survive the new gate — intended.

**Blind spot — ✅ FIXED (it was ours, not retrieval's):** `loadPagesForNav`
was in the corpus and its file in scope all along; F0.4c's sibling cap was
first-come-first-served and alphabetically-early giants starved
populate_nav_tables_action.go in every run (cap_hit=1 in all persisted
bundles). Now fair-shared per file with a "+N more" affordance. Regression
test reproduces the starvation.

**F1.1a — ✅ BUILT 2026-07-10 (the fixer's first slice, plan-only):**
`fix-proposer` agent (live) + `diagnose_persist_fix_plan` action (awaits
image) + artifacts kinds extended with `fix_plan` (applied, verified). The
workflow REFUSES non-CONFIRMED diagnoses — the gate F1 waited three benchmark
runs for. Plans are constrained (≤8 allowlisted edits, repo-relative paths,
grounded_in quotes required, 32KB cap) and persisted as artifacts; **no code
writes, no git token**. F1.1b (branch + PR via the spawn-gated token, gofmt +
build in a spawned job) is the next slice.

**RUN 5 + FIRST FIX PLAN — both landed 2026-07-10.** Run 5: CONFIRMED under
the strictest gates, [context] marks working, fair-share put nav-generation
code into the citations for the first time (full score: NOTES turn 16).
Fix-proposer: first plan persisted on correlation `e08c5b01` after two
truncated attempts were correctly REFUSED — root cause a platform-wide
dead-config gotcha (`max_tokens` must live INSIDE `ai_service`; the verdict
step ran capped at 2048 through all five runs; both agents fixed live).
Plan judged against ground truth (NOTES turn 17): machinery PASSED; plan
quality bounded by its input — it misses causes B and C and half its edits
are no-ops.

**F1.1b(a)+(b) + F2.1 — ✅ BUILT 2026-07-10:**
- (a) validator rejects no-op edits — explicit phrases only ("no code change",
  "clarifying note/comment", …); the first plan's two semantic no-ops are the
  test fixtures.
- (b) proposer input: last TWO bundles; prompt gains rules 6 (cover every
  cited mechanism or say why not in risks) and 7 (every edit changes
  something).
- **F2.1 — the council's first slice, wired INTO fix-proposer (10 steps,
  live):** persist_plan → review_editquality (real edits? right causal path?
  missing mechanisms?) → review_guardian (blast radius, architecture-change
  signals, surface ownership — **holds the hard veto**, Q-D v1 placement in
  step config) → `diagnose_council_decide` (deterministic Go: hard veto →
  rejected, any veto → rejected, any objection → revise, else approved;
  persists kind='council_report' on the same correlation). Reviewer contract
  is the verdict-wire-style opinion Q-D asked for. Q-G v1 = role prompts +
  plan + diagnosis (no per-reviewer corpora yet); Q-E v1 = the guardian's
  signal list. 5 aggregation tests; malformed reviewer output fails closed.

**F1.1b(c) — branch + PR (DESIGN, build next):** a separate `fix-implementer`
agent gated on `council.decision == 'approved'`. Write token isolation mirrors
`isRepoCloningAgent`: a new gate injects GITHUB_WRITE_TOKEN only into
fix-implementer pods, never shared chassis. Flow: clone at explicit ref →
LLM turns plan sketches into concrete diffs → a constrained editor action
applies them with the plan's file list as a hard allowlist → branch pushed →
**gofmt + go build in a spawned golang-image Job** (the chassis image has no
toolchain) gates PR creation → PR body carries diagnosis + coverage + plan +
council report (Q-H's package). Human review is the terminal; nothing merges
itself.

**F2.1 — ✅ PROVEN LIVE 2026-07-10 (v1.0.1106).** First council run on
`e08c5b01`: plan v2 fixed cause B directly (complete_error → fail_workflow, no
no-ops — the F1.1b input/prompt changes worked); the council returned
**revise** with substantively correct objections (editquality: wrong causal
path + the real ON CONFLICT mechanism missing; guardian: shared-file blast
radius, unnamed pipeline, and the unbounded-retry safety question). It neither
rubber-stamped nor dead-ended. Full record: NOTES turn 19.

**F2.2 REVISE LOOP — ✅ BUILT 2026-07-10 (workflow v3 live; Go rides next image).**
`diagnose_council_decide` counts council_reports (the per-round durable counter)
→ `round` + `should_revise`; a revise past the cap becomes `exhausted`
(terminal, not a silent approve). Loop: council_decide → check_revise →
{repropose (objections fed back) → persist_plan → review → decide | complete}.
Cap 2 terminates; a fresh plan each round converges. Verified closed; 5-case
cap test.

**F2.3 DECISION ROUTER + VERIFY + REFRAME + ESCALATION — ✅ CODE BUILT
2026-07-10 (turn 22); rides the next image with the round-scoping fix; v4 seed
ready, apply AFTER that image.** Motivated by the two clean benchmark runs:
`8c770fd5` (guardian hard-veto → rejected at round 1 — correct, but a silent
dead-end) and `aadd532a` (3 rounds, editquality converged to approve, guardian
down to "containable by pre-deploy audit queries" — exhausted one verification
short of approval). Four pieces:
1. **Router** (thin conditionals in workflow v4): approved→complete;
   revise(rounds left)→run_checks→repropose; rejected(first)→reframe;
   rejected(again)/exhausted→escalate. Flags computed deterministically in Go
   (`applyCouncilCaps`, extracted pure + tested, 8 cases).
2. **Verify step** (`diagnose_run_checks`): reviewers attach
   `checks:[{sql,why}]`; executed under the diagnosis data_request containment
   (IsReadOnlySQL lint → READ ONLY tx → statement_timeout → EXPLAIN gate →
   capped rows — pure reuse of `runDataRequests`); results feed the next
   repropose so fact-shaped objections are settled with evidence.
3. **Reframe-once** (`should_reframe`: first rejection with rounds left): a
   veto means the SHAPE is wrong — reproposing it gets vetoed again. One
   reframe: strictly narrower (site-scoped interim allowed IF risks names the
   deferred structural fix) or an explicit needs-architecture-review plan.
   Rejected-count sourced from council_report metadata; fails CLOSED to
   escalate.
4. **Escalation** (`diagnose_escalate`, kind='escalation'): rejected/exhausted
   become a first-class SUCCESS terminal persisting the hand-off package
   (decision + diagnosis + final plan + reviews, whose notes carry the
   reviewer's recommended alternative). F1.1b(c)'s PR body will carry it.
DEPLOY ORDER: chassis image (> v1.0.1107) → v4 seed → fire. v4 max_rounds=3.

**F1.1b(c): branch + PR — ✅ COMPLETE & PROVEN 2026-07-13 (PR #1 opened &
merged). Architecture (owner decision): the write credential stays in the
GIT-ADAPTER; the fix-implementer never holds a write token; it runs as a
DEDICATED pod that gets a read-only token via the isRepoCloningAgent spawn
gate.**
- **Part 1 ✅ (git-adapter):** create_branch (idempotent), create_pull_request
  (human terminal — created, never merged), branch-aware commit with
  repo-relative paths. Deployed (adapter rebuild); write-scope smoke passed.
- **Part 2a ✅ (chassis):** `diagnose_prepare_fix_commit` — the HARD
  file-allowlist safety core + branch/commit/PR payload assembly carrying the
  Q-H package. 7-case test suite.
- **Part 2b ✅ build gate = Option B (owner):** `diagnose_build_gate` — pre-PR
  gofmt+build in a golang k8s Job; green→PR, red→no-PR+log. (Its first red
  caught a real pre-existing breakage — cmd/test-spawning — fixed 9f29efb9.)
- **Part 2c ✅ fix-implementer:** seed live; reads via `diagnose_read_repo_files`
  (contents API, read token) at the correct ref; whole-file rewrite (32k
  budget); allowlist; branch/commit/gate/PR via `git_adapter_request`.
- **Part 2d ✅ dedicated pod:** `fix-implementer-orchestrator`
  (spawn_agent→call_agent) so the read-token gate fires; fired via the 092
  trigger, never directly.
- **First end-to-end run ✅:** seeded a hand-authored CONFIRMED diagnosis of a
  REAL tiny defect (misleading fmt.Printf); real proposer→council APPROVED
  (first approval)→implementer→gate green→**PR #1**, owner-merged. Diff: 1
  file, 2 deletions, zero drive-bys.
- **Open (F1.2):** ref/base are live-set to the active working branch because
  origin/main is stale — make them a per-run INPUT so the committed seed
  needs no branch-specific edit.

The dartsonline platform fix (mark_no_sections/fail_workflow + nav
build_status) remains human-implementable any time, independent of all this.

**Run protocol:** run 2 = a+b+c+e only, identical symptom string, site data
untouched — measures whether the loop now *finds* the cause. Run 3 = d —
measures whether a loop that cannot reach an answer says so honestly instead of
confirming a drift. One variable cluster per run.

**RUN 2 RESULT (2026-07-09, corr `dd1186b9`): 0/4 → 1 pass + 2 partial + 1 fail.**
Claim 3 cited via the agent_definitions enrichment; the confirm now explains the
blank-page half of the symptom (the silent-success mechanism). The nav half was
dismissed unexplained ("not a nav issue") — the precise residue F0.4d exists to
catch. Full scoring: RUNBOOK §BENCHMARK RUN 2 and NOTES(10) turn 10.

**F1 split:** the dartsonline platform fix (mark_no_sections; nav on
build_status) is human-diagnosed and does not wait on any of this. The F1 *fixer
mechanism* waits on F0.4d, because it gates on CONFIRMED and today's CONFIRMED
admits wrong-cause confirmations.

## 4. F1 stretch — the constrained edit plan

If the benchmark passes, F1 emits an edit plan on a branch. Its target is the
**platform**, not the site (NOTES(10), DECISIONS 2026-07-09):

1. `page-build-handler` workflow: give `check_has_ready_sections` an
   `else_step` that flags rather than succeeds — build the `mark_no_sections`
   step the guard comment at `load_work_item_actions.go:756` already assumes
   exists, setting `needs_human_review`. The completion guard at `:759-766`
   then preserves it, and the whole thing works as designed.
2. `populate_nav_tables_action.go:243`: ground nav in the built set —
   `AND build_status = 'deployed'`, or an explicit join to the built pages.
   This is the "nav never links unbuilt pages" principle the roadmap work
   already identified; it wants to be a guideline amendment too (side-task,
   per Q-D).
3. Leave cause A (the planner under-populating `sections`) to the builder
   thread. It overlaps item 6/7 territory and a fix there is a bigger change
   than F1 should attempt on its first outing.

Validation before any PR: `gofmt` + `go build` in a spawned job (Q-C).
Verification after: rebuild dartsonline and assert nav contains no link whose
page is not `deployed` — a natural first job for the tools chat's Stage-6
browser-runner adapter.

## 5. Order of work

1. F0.1a migration → 2. F0.1b write-through → 3. F0.1c envelope →
4. regenerate bundle with `-psql` → 5. confirm blinding → 6. benchmark run →
7. score against §3 → 8. F1 edit plan if passed.

Steps 1–3 are unblocked and independent of everything above; they can start
now. Step 5 gates step 6 absolutely.

## 6. What this pilot has already taught the workstream

Three pilot candidates, three dissolutions by cheap pre-check. The pattern is
now strong enough to name: **on this platform, bug mechanisms tend to be
legible to schema access plus grep.** The loop's value is therefore unlikely to
be *discovery* on bugs of this shape. It is more plausibly: (a) doing this
unattended, at 3am, on a bug nobody has looked at; (b) doing it with citations
a human can audit; (c) doing it consistently across a class of bugs. The
benchmark measures exactly (a) and (b). Worth stating plainly rather than
letting the workstream's premise drift unexamined.
