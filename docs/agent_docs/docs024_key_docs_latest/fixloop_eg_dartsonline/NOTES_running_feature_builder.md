# RUNNING NOTES — feature builder (fixloop delta, docs024/fixloop_eg_dartsonline)

**What this thread is.** Builds the multi-step feature-construction capability
on top of the proven fix loop — design §1 of
`DESIGN_feature_builder_and_council_gate.md`, booted from
`HANDOFF_2026-07-17_feature_builder_thread.md`. Session name: "fixloop feature
builder". Coordinates with the concept register workstream (docs026), whose
bug-historian council seat serves this chain too.

Newest entries at the bottom. Update every turn.

---

## Turn 1 — 2026-07-17 — Boot; staged-plan schema drafted for sign-off

Read (per the handoff): design §1, `0NN_fix_proposer.sql` (v6 live shape),
`0NN_fix_implementer.sql`, `0NN_fix_implementer_orchestrator.sql`, and the
three Go seams (`diagnose_persist_fix_plan_action.go`,
`diagnose_prepare_fix_commit_action.go`, `diagnose_read_repo_files_action.go`),
plus the concept register's running notes (thread-boot instruction).

Drafted and committed `SCHEMA_staged_plan_v1.md` (`cc136c902`) — the stages[]
schema, cross-stage file discipline, encoded image-before-seed checklist,
compatibility map, the F1.2 pilot as a worked instance, and six owner
decisions D1–D6. **Key structural finding:** F1.2 (implementer ref/base as
per-run input) is a PRECONDITION of the stage loop, not a cleanup — stage 2+
must read the fix branch, and `ref` is a config literal at
`diagnose_read_repo_files_action.go:101`. No code written; sign-off requested.

## Turn 2 — 2026-07-17 — Owner signed off (all recommendations); delta 1 built

Owner: "go with all recommendations" — D1 (reuse `kind='fix_plan'` +
`plan_format` discriminator), D2 (`feat/<short-corr>` branches), D3 (caps
6/8/24/128KB), D4 (seed/checklist rules as hard validation), D5 (F1.2 pilot,
self-hosted), D6 (derived go-test end gate).

Built delta 1:

1. **Staged validation** in `diagnose_persist_fix_plan_action.go` (commit
   `4b3d50f4c`): discriminated by `plan_format`/`stages`; legacy single-plan
   path behaviourally unchanged (per-edit rules factored into shared
   `editProblems`); staged path adds stage shape rules, add-once /
   no-modify-before-create / no-create-then-delete, the gate.build⇒seed/doc
   implication, and the checklist contract (seed⇒exactly-one seed_apply;
   image_deploy strictly before any seed_apply — the wrong order is
   unexpressible). 25 new test cases + probe discriminator tests, full
   package green.
2. **Feature-designer seed** drafted: `0NN_feature_designer.sql` — a DRAFT
   FILE, deliberately not applied (the seed discipline this feature encodes,
   applied to itself). Reuses the proposer's chain wholesale: staged persist →
   3-seat council (prompts adapted for staged judging; bug-historian digest
   verbatim) → same deterministic router/verify/reframe/escalate. Intake gate:
   `site_work_items` capability_gap items whose `spec` carries BOTH
   `owner_approval` and `code_pointers` (the designer runs in-chassis with no
   repo read token, so specs must carry curated paths; no new status values —
   the dedup-index/status contract stands). Deliberate delta from v6:
   run_checks answers all THREE reviewers' checks. Models: sonnet-5 for
   design/repropose/reframe, sonnet-4-6 for reviewers. MDL-039 guard: no root
   ai_service key.

Context that moved under us mid-turn: the makefile build default inverted
(`make build-<service>` now = committed HEAD; `-tree` is the WIP escape
hatch) — schema doc references updated to the new default form.

**Where this leaves the loop:** delta 1 complete on the code side. Next:
delta 2 (implementer stage-loop — per-stage allowlist/read/commit/gate, `feat/*`
branch, derived go-test end gate, new deterministic stage router action), then
the F1.2 pilot through the full chain. The owner's future acts (not now, in
order): merge/build image carrying `4b3d50f4c`+delta 2 → apply
`0NN_feature_designer.sql` → approve a pilot spec work item.

**Provenance note (same turn):** this turn's three doc artifacts
(`0NN_feature_designer.sql`, this file, the schema doc's sign-off edit) were
swept into a concurrent session's bulk commit `cf3803b49` ("product specs
additions and miscellaneous runbooks and handoffs") between our add and
commit — the exact index race `CLAUDE.md` documents. Content verified intact
in that commit; per the forward-only rule, no corrective action — this note is
the record. The delta-1 Go change committed cleanly under its own message
(`4b3d50f4c`).

## Turn 3 — 2026-07-17 — E1–E5 signed off; delta 2 BUILT; standing docs established

User: keep running notes + runbook + plan at all times, write a read-aloud
summary, carry on with all recommendations. E1–E5 thereby approved.

**Delta 2 built** (commit `c19b5d097`, full package tests green):

1. `feature_stage_route` — the loop's only new control machinery. Emits each
   stage as a SINGLE-PLAN shape so the proven read/prepare actions loop
   unchanged; per-stage read ref (base for stage 1, the feat branch after);
   per-stage commit message; terminal emission carries the PR payload
   (checklist rendered as an owner task list) + go-test packages DERIVED from
   the plan's .go edits (D6). E4 enforced at seed time: a pre-existing
   `feat/*` branch is a hard refusal via a GitHub API existence check.
2. Seams, all optional-field additions with single-plan behaviour untouched:
   `diagnose_read_repo_files` gains `ref_field`; `diagnose_prepare_fix_commit`
   gains `branch_field`/`commit_message_field`/`expected_symbols_field`
   (symbols checked against produced bodies); `diagnose_build_gate` gains
   `test_packages_field` (E2); `feature-implementer` joins the
   isRepoCloningAgent spawn gate (E1).
3. Seeds DRAFTED as files (`5b131b88a`): `0NN_feature_implementer.sql` (22
   steps, graph-validated), `0NN_feature_implementer_orchestrator.sql`
   (dedicated-pod wrapper — the read-token lesson), and two trigger drafts on
   the proven 092 kcat envelope.

**Council roster moved under us mid-turn:** the concept-register thread
shipped `review_reuse_agent` (fix-proposer v7, 4 seats). Extended the
feature-designer seed to mirror it — chain editquality → bug_historian →
reuse_agent → guardian, all four seats' checks answered. Reuse is this
builder's hard rule 1, so the new seat bites hardest here. RUNBOOK A3 covers
future roster drift.

**Standing docs established** (user request): `PLAN_feature_builder.md`,
`RUNBOOK_feature_builder.md` (A1–A7: image → seeds → roster check → pilot
spec SQL → fire designer → fire implementer → close out),
`SUMMARY_feature_builder_2026-07-17.md` (read-aloud), this file continuing as
the running record. RUNBOOK A5 evolves E5 slightly, flagged as the owner's
choice: with the designer built, prefer firing the full chain on the pilot
spec and GRADING its plan against the hand-written §6 reference, over
hand-injecting the plan.

Another index-race sweep this turn: registry.go's new entry rode into
`aabd38161` (experience-loop's commit). Content verified intact; noted here,
forward-only.

**State: all delta-1+2 code committed and inert. Everything from here is
owner acts (RUNBOOK A1–A7) — nothing further to build until the pilot's
grades come back.**

## Turn 4 — 2026-07-17 — Image built+pushed; rollout HELD on live traffic (user correction); designer synced to v8

User said "let's carry on"; began RUNBOOK A1. Bumped IMAGE_TAG → v1.0.1131
(commit `202019e6c`), built agent-chassis from committed HEAD, pushed to
docker.io/aqls. **User stopped the rollout step and said: read CLAUDE.md and
follow it.** They were right, twice over: (1) acting on the stale
session-start git status — the chassis kustomization was already dirty with
another session's uncommitted newTag; (2) restarting the shared chassis
without checking what's in flight.

Checks then run (the CLAUDE.md way): the dirty kustomization diff is ONLY
newTag 1128→1130 residue matching the live deployment (v1.0.1130, 1/1
ready) — safe. But orchestration_states shows a LIVE pipeline: 1
EXECUTING_STEP (spawn_link_resolver, seconds old) + 8 AWAITING_RESPONSES,
new arrivals every 1–2 minutes since 17:28. A restart now kills the
executing step (never reaped — bugs_open/003) and drops spawns for ~300s.
**Rollout held; owner picks the moment.** v1.0.1131 sits pushed and ready.

Council roster moved AGAIN mid-day (v8, 5 seats: + review_guidelines).
Synced the designer seed: chain editquality → bug_historian → reuse_agent →
guidelines → guardian; all five seats' checks answered; guidelines prompt
carries the wrapper-orchestrator / dedup-contract / declared-contracts /
schema-tier rules + the approve-don't-object meta-rule for stale guidelines.
Graph re-validated (22 steps, all targets defined).

## Turn 5 — 2026-07-17 — LIVE: image verified (v1.0.1132), three seeds applied, pilot spec created; approval is the owner's

Owner answered the two questions: (1) another thread owns the rollout —
hands off; (2) yes, apply the seeds after the pod verifies.

The other thread's **v1.0.1132** landed while we watched — running pod
binary verified (`feature_stage_route` count 3), so A1 closed without us
touching the deployment. Applied the three seeds to
postgres-clients-0/clients_db in order (`INSERT 0 1` × 3) and verified:
step counts 22/22/3, designer chain editquality → bug_historian →
reuse_agent → guidelines → guardian, 5 review_fields, no root `ai_service`
anywhere. **The feature builder is now REGISTERED and inert in production.**

A4: created the pilot spec work item
`db066cac-c647-44bf-a3ca-e04416405b28` (capability_gap,
needs_human_review, item_key f12-ref-input-pilot, goal + 3 code_pointers).
Caught a runbook bug on the way: the site anchor is named `System
(internal)`, not `system.internal` — the drafted INSERT would have silently
inserted nothing; fixed in the runbook.

**Remaining, both the owner's: the named approval UPDATE (runbook A4) and
the credit-spending go to fire the designer (A5).** Mind the ~300s
no-dispatch window if the chassis pod has just restarted.

## Turn 6 — 2026-07-17 — New chassis image v1.0.1134; machinery re-verified live

Owner: a new chassis image was deployed. Verified the running pod
(`agent-chassis-6d85fff446-54jzc`, `v1.0.1134`, started 19:34:11Z): binary
carries `feature_stage_route` (×3) AND the delta-2 expected-symbols guard
("expected symbols not present"). All three feature agent rows still active
(an image swap doesn't touch the DB — confirmed, not assumed). At 19:43 the
pod is ~9 min old, past the ~300s dispatch-drop window, so a fire would be
safe timing-wise. Pilot work item `db066cac` still `needs_human_review`,
`owner_approval` absent.

**Nothing new to build or apply. The two remaining acts are the owner's and
cannot be delegated: the NAMED approval UPDATE (A4 — it records who
approved) and the credit-spending GO to fire the designer (A5).**

## Turn 7 — 2026-07-17 — FIRST FIRE: feature-designer live on the F1.2 pilot

Owner approved (aaa) and gave the fire go. Confirmed pod age 1530s (past the
300s window), fired `0NN_TRIGGER_feature_designer_v1.sh db066cac`.

- FEATURE correlation: `fc2cf851-ab2d-4303-a8e4-a545275f6ee3`
- Run orchestration: `69def2d7-8932-461b-8258-93c4444ce3ab`

Confirmed live it passed the spec gate: load_spec → check_spec_approved →
load_schema_hint → design — i.e. the owner_approval + code_pointers gate
computed `approved` on real data (first end-to-end proof of the intake gate).
**RESULT — PARTIAL SUCCESS, validation did its job.** The run went
load_spec → check_spec_approved → load_schema_hint → design → persist_plan →
`complete_refused` (COMPLETED, no artifact written), in ~90s. First proof the
intake gate works on real data, and the designer produced a well-formed
4-stage staged plan (s1/s2 code, s3 seed, s4 config_change) with a checklist —
but staged VALIDATION refused it, all three reasons on stage s4 (a
config_change edit):
1. config_change `file` was prose with spaces/parens (`fixloop….fix_implementer
   (live workflow definition row…)`) → tripped the repo-relative/no-whitespace
   rule (shared with single plans, where config_change targets look like
   `agent_definitions:page-build-handler`).
2. `artifact_role: "doc"` on a config_change → my contradiction rule fired.
3. `config_change` used as a post_merge_checklist act → not in the allowlist
   (image_deploy|seed_apply|verify).

Diagnosis: two are model-steer gaps (config_change file format; role), one is
the model over-decomposing — it added a config_change "activate v2" stage that
is redundant because the seed's `ON CONFLICT DO UPDATE` already activates. NOT
a schema gap: config_change edits are already surfaced in the PR body's own
section (buildStagedPRPayload), so they were never meant to be checklist acts.
No council credits spent (refused before review).

**Fix applied (prompt-steer, contract preserved):** designer seed rule 5
now states a seed upsert IS the activation (no separate activate stage);
new rule 5b constrains config_change — `file` = compact `agent_definitions:<type>`,
role code/omit never seed/doc, appears in PR body not the checklist. Seed
re-validated (balanced, graph intact). This is a seed-FILE edit — the LIVE
designer row still has the old prompt until the seed is re-applied (owner act).

**Next (owner calls, both stated):** re-apply the improved designer seed
(prod DB write) + re-fire on db066cac (credits). Optional spec sharpening:
the Go seams (ref_field etc.) already exist since c19b5d097, so F1.2's real
remaining work is the fix_implementer WORKFLOW seed — the spec's goal could
say so, to stop the designer proposing redundant s1/s2 Go edits.

## Turn 8 — 2026-07-17 — Run 2: the designer was RIGHT and validation was wrong; D4 refined

Owner said go: sharpened the spec (goal now states the Go seams exist since
`c19b5d097`; deliverable = ONE v2 workflow seed; pointers marked REFERENCE
ONLY), re-applied the fixed designer seed (snapshot `ba8f1fcd`), re-fired →
run `bcc96877` / correlation `76ed7fee`. Refused again at persist_plan —
**but the plan was right and the validation was wrong.**

The designer produced exactly what the sharpened spec asked: ONE stage
(add `0NN_fix_implementer_v2.sql`, role seed) — no redundant Go edits, no
config_change stage (turn 7's prompt fixes both worked) — and the TRUTHFUL
checklist `seed_apply → verify`. My D4 rule demanded image-before-seed
unconditionally; a seed-only feature has no image to deploy, so the rule
forced the checklist to lie. The pilot found a real validation bug on its
second fire.

**D4 refined** (code + tests green, schema doc §4 rule 2 amended): image-
before-seed stays HARD when the plan ships code edits; seed/doc-only plans
may checklist `seed_apply → verify` with no image entry; ordering still
enforced if an image entry is present. Designer prompt rule 5 updated to
match ("do not invent an image step") and re-applied live.

**Blocked on:** the validation fix is Go — inert until the next chassis
image (rollouts owned by another thread today; HEAD builds pick the commit
up automatically). Re-firing before that refuses identically. Then the
owner's per-run credit go. Score so far: 2 fires, 0 council spends, 2 real
defects found and fixed (1 prompt-steer, 1 validation) — the cheap gates
are doing exactly their job.

## Turn 9 — 2026-07-18 — Run 3: plan PERSISTED, the five-seat council ran for the first time, one steer gap left

New image v1.0.1135 verified (carries the D4 refinement); owner's
image-landed message = the agreed go; fired run 3 (correlation `f7806376`,
orchestration `a6d1d6d3`).

**The furthest run yet, and the first council exercise ever:**
1. Plan persisted (kind=fix_plan, 1 stage — the D4 seed-only fix proven live).
2. ALL FIVE seats reviewed, valid JSON, genuinely distinct catches:
   editquality OBJECT (sketch used non-existent `definition` column — it's
   `default_config`; invented `base_branch_field` key would be silently
   ignored); **bug-historian OBJECT — the charter's first real exercise,
   on-target**: applied pattern #7 (missingkey=zero) to the seed's new field
   paths + flagged the ON CONFLICT overwrite shape, demanded fail-loud;
   reuse-agent APPROVE (seams reused, nothing parallel); guidelines APPROVE
   with an input_contract hygiene NOTE (the FIX-016 meta-rule working as
   designed — did not block a correct plan); guardian OBJECT below veto
   (same schema mismatches independently + mid-flight activation ownership).
3. council_decide → revise; router → run_checks (reviewers' SQL EXECUTED) →
   repropose.
4. The REVISED plan failed validation: checklist act `image_check` invented —
   the act set is closed but the repropose/reframe prompts never restated it.
   complete_refused; round-1 artifacts survive on the correlation.

**Fix applied + live:** repropose AND reframe prompts now state the closed
act set (image_deploy|seed_apply|verify; pre-apply confirmation = a verify
entry ordered before seed_apply). Designer seed re-applied (snapshotted).

**Cross-thread drift, flagged not actioned:** the fix-proposer's council is
now THIRTEEN seats (2 always-on + 11 gated via the relevance filter,
migrations v6–v18) — my designer's always-on 5-seat chain is an architecture
generation behind, not just a roster behind. Run 3 shows 5 always-on works;
adopting the gated architecture is a seed-only change (the filter Go is in
the running image) but it is the OWNER's call, post-pilot. Also noted
precisely: run 3 exercised the bug-historian CHARTER (this workflow's copy);
the fix-proposer's own seat instance is still unexercised (their BUG A
dispatch remains their first test).

Score: 3 fires, 1 council round spent, 3 defects found (2 steer, 1
validation) — each caught by the cheapest gate that could catch it. Next
fire should carry the revision through validation into council round 2.

## Turn 10 — 2026-07-18 — Run 4 went the DISTANCE: 3 rounds, correct ESCALATION, and the council caught a real architectural danger

Run 4 (`3b084712` / `eaae17f3`) completed the whole machine for the first
time: design → persist → 5 seats → revise → **checks executed** → repropose →
persist → 5 seats → revise → repropose → persist → 5 seats → revise → round
cap (3) → `escalate` → `complete_escalated`. No errors. Seven artifacts: 3
fix_plans, 3 council_reports, 1 escalation. Escalation is a FIRST-CLASS
SUCCESS terminal — the loop refused to approve something it couldn't
satisfy, and packaged it for a human. That is the design working.

**The checks discovered the real production defect.** Reviewer SQL, executed
under containment, revealed the live `fix-implementer` row's actual config:
`read_current_files.config.ref` and `prepare.config.base_branch` and
`create_branch.config.data_literals.from_branch` are all the STALE LITERAL
`084_site_improvements_local_ai` — the exact gotcha F1.2 exists to fix,
confirmed live from the database rather than from our notes. It also
corrected the plan's structural assumption (only create_branch uses the
data_literals/data_fields split; the other two carry flat literals).

**Final round: 4 approve, 1 object.** editquality, reuse_agent, guidelines,
guardian all APPROVED the third plan. The **bug-historian** held out with
two high-severity objections — and it is RIGHT:
- The seed reconstructs the WHOLE `default_config` via upsert. Any step or
  key outside the builder's partial view is silently dropped (its own
  verification query had truncated). Occurrence-6 shape.
- The proven safer shape exists and wasn't used: surgical `jsonb_set` on
  only the changed keys, leaving unaudited steps byte-identical BY
  CONSTRUCTION.

**This independently rediscovered the concept-register thread's own hard-won
rule** (their memory, 2026-07-17: seat migrations that `SET default_config =
EXCLUDED` CLOBBER concurrent threads' edits; any change to an EXISTING shared
step MUST be surgical `jsonb_set`). Two threads, different evidence, same
conclusion. The designer wasn't careless — it followed the house seed
template, which is correct for NEW agents and dangerous for EDITING
co-edited ones.

Its third objection (are `*_field` keys overlays or ignored-when-literal-
present?) I settled directly from the deployed source: `ref_field` wins when
it resolves non-empty, else the literal stands — so the overlay assumption
holds, BUT the silent fallback to a stale literal is real, which is why the
plan correcting those literals to `main` is load-bearing.

**Rule encoded + live:** designer rule 5a — NEW definition → full INSERT;
EDITING an existing definition → surgical `jsonb_set` on only the changed
paths, never whole-column replacement, because these rows are co-edited and
your view of them is always partial. Seed re-applied (snapshotted).

Score: 4 fires, 4 council rounds, 4 defects found (3 steer, 1 validation) +
1 genuine architectural finding + 1 live production defect confirmed. Run 5
should produce a surgical seed the bug-historian can approve.

## Turn 11 — 2026-07-18 — CLAUDE.md re-read: bug 016 fixed here, and run 4's grading was WRONG

Owner asked me to re-read `CLAUDE.md` (two whole new sections since this
thread started: an advisory **council gate** for platform changes, and
diagnosis-before-debugging) and flagged `bugs_open/016`.

**016 is ours and it invalidates turn 10's conclusion.** In a prompt
TEMPLATE, `{{.X.result}}` renders `<no value>` silently — `ExtractFields` →
`UnwrapDeep` strips the `{type,result}` wrapper, so `.result` is a lookup for
a key that no longer exists. `feature-designer` had 5 such refs in
`repropose` and 2 in `reframe`. **The reviser therefore never saw a single
objection in any run.** Run `3b084712` burned all three rounds with the
bug-historian's objection unchanged each time — not because the plan was
unfixable (turn 10's reading) but because the feedback loop was severed.
Facts still improved between rounds because `{{.check_results.results_text}}`
is correct (a field ON the unwrapped value), which is exactly what made the
failure look like stubbornness. **That asymmetry — facts improve, objections
never get addressed — is the tell for auditing any other council.**

Fixed SURGICALLY (`PATCH_feature_designer_016_revise_prompts.sql`): jsonb_set
on the two prompt_template leaf paths only, snapshotted, config dot-paths
untouched — i.e. exactly the co-edited-row rule this thread's own council
taught it last turn, applied to itself the very next change. Verified 0
broken refs, `check_results` intact, 5 review_fields intact. Seed file
corrected too. Live sweep: both feature-* agents clean; the only remaining
`.result}}` in the fleet is `content-creator-hero`, another thread's to fix.
Turn 10's rule 5a stands on its own merits (the design prompt DOES render).

**Council-gate compliance gap (mine).** The new CLAUDE.md section asks
threads to run `platform/` changes past the council before committing, with a
`Council-Reviewed:` trailer. My three platform commits (`4b3d50f4c`,
`c19b5d097`, `62018e272`) carry no trailer — the coverage report lists them
UNREVIEWED. So are 29 of 30 in-scope commits fleet-wide, so this is an
un-adopted convention rather than a personal lapse, and forward-only forbids
amending trailers in. The part that genuinely merits review is **delta 2's
stage-loop machinery**: it has unit tests but has NEVER run live, and it is
the next thing to fire. Recommend submitting it before the first implementer
run. (The coverage report itself was mid-fix during this turn — an earlier
run reported 3 of 30 commits; another thread's stdin fix landed between my
two invocations. No bug to file.)

**Open, both owner calls (credits):** run 5 with a WORKING revise loop, and
whether to council-review delta 2 before the implementer's first fire.

## Turn 12 — 2026-07-18 — RUN 5 APPROVED (unanimous 5/5) — and the pilot target was fixed by hand one minute earlier

**The milestone: the feature builder produced its first council-APPROVED
staged plan.** Run `8e837814` / orch `a8b66dee`: 3 plans, 3 council rounds,
terminal `complete` — final round **editquality, bug_historian, reuse_agent,
guidelines, guardian ALL approve, zero objections**.

**016 fix PROVEN in the wild** (the reasoning-dataset thread asked to be told
when a genuinely post-fix repropose landed). Using their run-start join
(`llm_call_log` → `orchestration_states.created_at`, test the RUN not the
step): run started 15:27:33, repropose call 15:33:40, **`<no value>` count 0**,
prompt names `bug_historian`, carries objection text, 31,043 bytes. The
reviser genuinely read the council. Correction for their corpus work:
these rows are logged under `agent_type='generic'`, NOT the real agent type
(`feature-designer` returns 0 rows; all 20 `repropose` rows are `generic`) —
`agent_type`-keyed queries will miss every orchestrator-run agent; join on
`orchestration_id`.

**Plan quality — the run-4 lesson visibly absorbed.** The approved plan is
ONE seed stage using a guarded `DO $$` block: `FOR UPDATE` lock, explicit
`RAISE EXCEPTION` if the row or either expected path is absent (fails LOUD),
then three `jsonb_set` calls on specific leaf paths. No `ON CONFLICT`, no
whole-column replacement — exactly what the bug-historian demanded and rule
5a encoded. The pre-apply `verify`-before-`seed_apply` checklist shape also
came through as steered.

**BUT: DO NOT APPLY THE SEED.** The fixloop thread completed F1.2 by hand
(`0dd750bcc`, `4e9445e49`, `a2e868585` "turn 40 — F1.2 done"), patching the
live row at **15:26:40 — one minute before run 5 fired**. Their design uses
ONE per-run input (`base_branch`) driving both read-ref and branch-base, and
their 092 trigger passes only `base_branch`. The approved plan sets
`read_current_files.ref_field='input_data.ref'` — an input nothing passes —
so applying it would make reads fall back to the literal `main` instead of
following the run's base: a REGRESSION of a working fix. Pilot work item
`db066cac` closed `complete` with the reason recorded in its spec.
(My own pre-flight read was already stale by 60s — the row moved between
my check and the fire. Third co-edit collision today.)

**Structural fix applied — the reviser now reads the ARTIFACT** (their second
016 finding; the list-vs-artifact call was left to me).
`PATCH_feature_designer_017_reviser_reads_artifact.sql`: new
`load_council_report` step (query_database → latest `council_report` body),
`run_checks → load_council_report → repropose`, repropose `input_fields`
now `[spec_row, plan_persisted, council_report_row, check_results]`, and the
five per-seat prompt sections replaced by one artifact section. Verified: 0
per-seat refs, artifact ref present, checks ref intact, 5 reviewer steps and
5 `council_decide.review_fields` untouched. Now idempotent under roster
growth — seat 6 appears in the reviser automatically, because
`council_decide.review_fields` is the single list, and it is self-enforcing
(a seat missing there already fails loudly by not counting in the decision).
Applied surgically, AFTER run 5 terminated, never mid-run.

**Council gate: submission `5a65ec4c` (delta 2) — my "never dispatched" call
was WRONG.** It was QUEUED, not dropped: I checked ~2 min after firing and
saw no orchestration row, and reported a silent drop. It started at 15:42:59
and finished 15:50:18. **Lesson: absence of a row shortly after dispatch is
not evidence of a drop** — the 300s window makes early absence normal. I
retracted this to the owner.

## Turn 13 — 2026-07-18 — Council gate on delta 2: REVISE, and it found a real high-severity defect

First use of the council gate by this thread (submission `5a65ec4c`). Verdict
**revise**: 7 seats fired via the relevance filter — diagnosis_guardian
approve; editquality, bug_historian, reuse_agent, tooling_provenance,
guardian, debug_historian object (15 objections). Worth the credits on the
strength of one finding alone.

**THE REAL DEFECT (bug_historian high, + 2 medium — one root cause).** All
three routed seams used "resolve non-empty, else fall back", which collapses
*not configured* with *configured but resolved empty* — the platform's
worst-known failure shape. Concretely: `ref_field` empty → stage N silently
reads `main` instead of the branch carrying stage N-1's commits, and the
implementer rewrites files from the wrong tree; `branch_field` empty →
prepare re-derives `fix/<corr>` and commits a stage's files to a DIFFERENT
branch than the loop is building; `test_packages_field` empty → a silent
build-only gate, forfeiting the exact D6 guarantee that mode exists for.
**Fixed (`9c94cc842`): all three now error when configured-but-empty; unset
keeps the single-plan path byte-identical.** Test added locking the router's
side of the contract (it must never emit an empty branch/ref/message, and
must derive packages for a plan with .go edits). Package green.

**Three objections settled by direct verification rather than assertion:**
- guardian [high] "confirm no OTHER pipeline uses these actions": live DB
  sweep — ONLY `fix-implementer` and `feature-implementer`. Blast radius is
  exactly the two known consumers.
- guardian [med] "confirm the buildGateScript call sites": 3 test + 1
  production + the definition, all updated (my submission said "both", i.e.
  2 — wrong count, right substance).
- debug_historian [med] "commit hashes are not deploy evidence — grep the
  RUNNING pod": correct, and my rationale did cite commits. Pod-grep confirms
  `feature-implementer` IS in the live binary. Their point stands as method.

**Not yet actioned (owner's call):** editquality — registry.go registration
should be its own declared edit, not buried in a sketch; editquality+guardian
— `expected_symbols`' verbatim substring check can false-reject a correct
stage whose symbol lives in an earlier stage's file (self-identified in my
own risks, still unmitigated; the honest fix is designer-prompt guidance that
expected_symbols name only symbols the stage's OWN files introduce);
reuse_agent — asks whether `site_work_items` sequencing (parent_item_id /
depends_on / batch_id) should carry stage state instead of a new action (my
view: no — that is work-item queueing, not in-run workflow state, and
`diagnose_route` is the right precedent — but it deserves a written answer,
not a dismissal); tooling_provenance — these actions should carry travelling
PLAN+NOTES subjects.

**Correction to turn 12:** my "council gate silently never dispatched" was
WRONG — it was queued, and ran 15:42:59→15:50:18. Absence of an orchestration
row ~2 min after dispatch is normal, not evidence of a drop.

## Turn 14 — 2026-07-19 — Session close: docs squared for a fresh chat

Owner asked for a read-aloud summary and a handoff so a new chat can resume.
Written and committed (`768b549a4`):
- `SUMMARY_feature_builder_2026-07-19.md` — read-aloud; supersedes the 07-18
  one (written before the machine had ever been approved).
- `HANDOFF_2026-07-19_feature_builder_thread.md` — cold-start: what is live,
  the five hard-won lessons, explicit do-NOTs, open items with owners.
- `PLAN_feature_builder.md` — architecture now 5 seats + artifact-reading
  reviser; status table says designer PROVEN / implementer NEVER FIRED; the
  four open council-gate objections recorded.
- `RUNBOOK_feature_builder.md` — A1–A5 closed with evidence; new **B1–B4**
  aimed at one thing only: the implementer's first fire.

**State at close.** Designer half: proven, unanimous council approval (run
`8e837814`). Implementer half: live, never executed — the whole remaining
gap. F1.2 pilot closed as superseded; its approved plan must NOT be applied.
Delta 2 reviewed by the council gate (verdict revise): the high-severity
fail-loud defect is fixed (`9c94cc842`), four objections remain open.

Nothing is mid-flight; no background runs pending; no uncommitted work of
ours in the tree.

<!-- Append new turns below this line. Format: ## Turn N — date — one-line summary -->

## Turn 15 — 2026-07-20 — Fresh chat ("fixloop feature builder 2"): state re-check + B1 candidate vetting

Resumed from `HANDOFF_2026-07-19_feature_builder_thread.md`. Re-checked live
state before touching anything (the row moves — turn 12's lesson):

- **No staged-v1 plan since the handoff.** Today's `fix_plan` artifacts
  (`2eed453a`, `d35844da`, 10:28–11:01Z) carry NO `plan_format` — legacy
  fix-proposer runs from other threads. The implementer gap is still ours.
- **No `feat/*` or `fix/*` branches on the remote** (fetched + pruned). E4
  refusal will not trip on stale branches.
- **Live chassis = v1.0.1139** (`agent-chassis-645674b498-rndg9`). Pod-grep:
  `feature_stage_route` = 3 (stage loop live), **`formatGeneratedGo` = 0** —
  the bugs_open/013 fix (`fc38c6058`, today 12:03 BST: commit-prep gofmts
  generated Go via go/format) is committed but NOT in the live binary. That
  fix exists because gofmt trivia killed the fix-implementer's first run
  (70680566), and the prepare action is shared — firing our pilot on
  v1.0.1139 carries the same known burn risk. The 019 council-truncation
  degrade fix (a3b606798) is likewise inert until a roll after v1.0.1139;
  until it lands, a truncated reviewer voids whole council rounds, including
  our designer's.
- **B1 vetting.** Strongest candidate: bugs_open/023's delivery gap —
  **294 items sit at `status='needs_human_review'` and nothing consumes
  them** (live count today). The digest reads `needs_diagnosis` and
  `capability_gap`/`deferred` (fixloop_digest_action.go:307,358) but has NO
  section for `needs_human_review`. A digest section surfacing that queue is:
  genuinely wanted (023 names the gap; no open work item claims it), our own
  machinery (file untouched 4 days, no collisions), read-only/additive, and
  naturally a 2-stage staged-v1 plan (stage 1: new section file + new test
  file; stage 2: wire into the digest) — Go edit + new files = the full B1
  gate exercise. Runner-up: feature-run visibility in the digest
  (plan_format discrimination — self-hosted awareness). Vetted and DEFERRED:
  agent_type='generic' attribution (high cross-thread value but the caller
  path `ai_actions.go` carries another session's live uncommitted WIP —
  collision risk B1 forbids; arguably fix-loop jurisdiction anyway).
  Rejected: bugs_open/007 `--record-only` migrations tooling (re-armed
  today, genuinely wanted, but shell-only — leaves the Go build/test gates
  unexercised; wrong pilot shape).
- Working tree carries several other sessions' WIP (banana provider,
  ai_actions.go, council actions, fleet kustomizations, prepare-commit
  action) — left strictly alone.

Next: owner decisions — B1 target choice, image roll before firing (my
recommendation: yes, v1.0.1140 from committed HEAD arms both 013 and 019
de-riskers), delta-2 gate resubmission. The four open delta-2 objections are
ours and free — will action once the target is set.
