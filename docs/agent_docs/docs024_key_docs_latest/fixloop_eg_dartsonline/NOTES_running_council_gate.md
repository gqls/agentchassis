# RUNNING NOTES — Council Gate thread

Turn-by-turn record, newest at the bottom (house practice, matching
`docs026_concept_register/RUNNING_NOTES_concept_register.md`). Standing state
lives in `RUNBOOK_council_gate.md`; this file records how it got there.

---

## Turn 1 — 2026-07-17 — Decisions collected; advisory gate built end to end (NOT applied)

Cold-started from `HANDOFF_2026-07-17_council_gate_thread.md` + design §2.

**Recon before building** (schema-first rule): read the v6 fix-proposer
workflow, `diagnose_persist_fix_plan_action.go`,
`diagnose_council_decide_action.go`, `append_doc_note_action.go`,
`fixloop_digest_action.go`, and the 091 trigger envelope. Three mechanical
facts shaped the build:
1. `plan_field` resolves over collected data, where `input_data` lives — so a
   submission can carry a ready-made fix_plan JSON and the **wrapper needs
   zero new Go**; the whole gate is config + scripts, live on seed apply, no
   image build.
2. Council round-counting is **orchestration-scoped** in the deployed source,
   so resubmissions on the same correlation are judged fresh; the shared
   correlation is for the artifact trail, not the cap.
3. The digest action runs **in the pod, which has no git repo** — the
   "un-reviewed platform commits" join can only run repo-side. Built as a
   script that can persist its report to doc_notes (same awareness channel),
   not as a digest section.

**Owner decisions collected** (the handoff's four questions): scope =
`platform/` + `internal/` + `pkg/`; advisory first; credits per submission
(= per task/commit); roster = **wait for more stage-3 seats** — which makes
seat-building the launch gate.

**Built** (files only, nothing applied):
- `0NN_council_gate.sql` — 17-step orchestrator: schema hint → persist
  submission (same structural validation as proposer plans, 64KB cap) → the
  three v6 reviewer prompts with the diagnosis context swapped for the
  author's rationale → deterministic council_decide (guardian hard veto,
  all-three review_fields) → no repropose loop (the author revises; a revise
  verdict runs ALL three reviewers' checks first so objections come back
  settled with evidence — v6 only ran two reviewers' checks, an omission not
  copied) → deterministic SQL-composed verdict doc_note (no LLM in an
  awareness surface) → three terminals. Literal balance verified with a
  comment-aware tokenizer (validator agreed with the known-good v6 file).
- `097_TRIGGER_council_review_v1.sh` — submission wrapper: client-side
  validation (thin-rationale refusal — missingkey=zero would render a missing
  rationale as silent blank in the prompts, so it must fail HERE; scope
  pre-filter so docs never spend credits; ≤8 edits, ≤64KB), single-line jq -c
  payload (kcat -P line-split trap), RESUBMIT_CORR support. Dry-run tested
  against a stubbed kubectl: good/docs-only/thin-rationale all behave.
- `098_REPORT_unreviewed_commits_v1.sh` — deterministic git↔verdict join on
  the `Council-Reviewed: <correlation>` commit trailer (exact join, not
  file-overlap heuristics; MISMATCH bucket makes false claims visible;
  NO_DB degrades to TRAILER-UNVERIFIED, never silently to REVIEWED).
  **First live run: 28 in-scope commits in 3 days, 0 reviewed — the
  baseline.**
- `RUNBOOK_council_gate.md` — standing state + launch checklist.

**Seat #4 (reuse-agent) — a live cross-thread collision, resolved by
addendum, not overwrite.** The owner's roster ruling put seat-building on the
gate's critical path, so this thread set out to draft the reuse-agent spec —
and found `docs026_concept_register/PILOT_reuse_agent_reviewer.md` already on
disk, written minutes earlier by the concurrently-running concept-register
thread, complete and better-grounded (FIX-036's founding incident: a
reinvented trigger+triage SQL pair). Did NOT overwrite it. Appended an
attributed §6 addendum carrying the two facts only this thread knew: (i)
there are now TWO council definitions (fix-proposer + the gate seed) and any
seat migration must patch both or the rosters silently drift; (ii) v6's
`run_checks.check_fields` omits the bug-historian's checks — their four-edit
v7 patch would repeat that omission for the new seat; the v7 migration should
carry all four reviewers' checks (the gate seed already does); (iii) the
owner's four gate rulings, as context for their roster-scaling question. The
file was modified on disk even between this thread's read and its edit — the
concurrency is not theoretical.

**Read-aloud summary written** (user request, mid-turn):
`SUMMARY_council_gate_2026-07-17.md` — what we're doing / where we are /
what's next, plain language, house read-aloud convention. Standing practice
from here: this notes file is updated every turn.

**Deliberately not done:** applying the seed (roster ruling + named-target
permission gate), PR-mode (build order step 4, owner's explicit go), any edit
to the live fix-proposer, any overwrite of the other thread's pilot spec.

**LATE-TURN DEVELOPMENT — the roster arrived while we worked.** The
concept-register thread, in its own conversation with the owner, built and
APPLIED seats #4 and #5 to the live fix-proposer the same afternoon:
reuse-agent (v7) and guidelines-agent (v8) — the live council is now **5
sequential reviewers**, and it also produced `DESIGN_relevance_filter.md`
(seat-skipping needs a small chassis Go change: `diagnose_council_decide`
hard-fails on an absent reviewer field, so "skip" must become "abstain").
Consequences handled here: (1) synced `0NN_council_gate.sql` to the v8
roster — both new steps mirrored verbatim with the rationale-context swap,
`review_fields` and `check_fields` now carry all five; 19 steps, literal
balance re-verified; (2) flagged (RUNBOOK + pilot §6(ii)) that the live v8
`run_checks.check_fields` still lists only editquality + guardian — three
advisory seats' checks are solicited but never executed; a one-line config
fix belonging to the fixloop/concept-register surface, not this thread's;
(3) the owner's launch condition ("more seats first") is now plausibly MET —
reframed as launch-checklist step 0: launch on 5 seats, or wait for the
relevance filter? (4) updated the read-aloud summary to match reality.

**AND IT MOVED AGAIN before the turn closed** (their turns 18–19): the
run_checks gap this thread flagged was fixed by their **v9, applied live**
(gate docs/seed updated to match); their relevance-filter Go engine is built
and committed (`37468ba65`) — registered but uncalled, council_decide
abstention backward-compatible, so it rides the next chassis image inertly;
and the owner chose deploy option (b): another thread leads the chassis
deploy, no action assigned here. The v10 SQL wiring is deliberately unwritten
until the image is pod-verified — when fix-proposer adopts the filter, the
gate seed adopts it in the same migration (noted in the seed header).

**Committed** (per CLAUDE.md, explicit pathspecs, one commit per task):
`e8530898f` (the six gate files; 097's content had already been swept
verbatim into another session's bulk commit `4f581dcf9` — nothing lost,
forward-only) and `7fd6f6c64` (the PILOT §6 addendum; the running-notes
cross-note had itself been swept into their `d1129f285` between my status
check and commit — the concurrency again, again harmless).

## Turn 2 — 2026-07-17 (evening) — Owner: filter next; the image arrived on its own; seed synced to v11

Owner (here): "I have suggested in the other thread that the relevance filter
should be next, so please go in that direction." Read
`DESIGN_relevance_filter.md` + the `select_review_panel` Go (config contract:
plan_field/extra_text_fields/footprints, fail-open, pairs with
council_decide abstention; confirmed `37468ba65` in HEAD).

**The deploy resolved itself mid-turn, and the pod-verify rule earned its
keep twice.** First check: cluster running v1.0.1132, pod 17 min old — but
`strings` on the running binary found **zero** filter symbols (the tag
proves nothing; that image predated the engine commit). Prepared to bump to
v1.0.1133 and lead the deploy — the owner interrupted: a new image was
already shipping. Confirmed: **v1.0.1133 released fleet-wide by the
coordination thread**, new pod carries `select_review_panel` (4 symbol
hits), and the concept-register thread had applied **v10**
(tooling-provenance, seat #6) and **v11** (select_panel + 4 relevance
gates) to the live fix-proposer. The interrupt also prevented a tag
collision — my planned bump target was the tag that had just shipped.

**Synced the gate seed to v11** (owner's caution honoured: re-read the file
from disk before resuming — another thread edits it; found only my own
changes, mid-surgery with a gate pointing at a not-yet-existing step):
select_panel (extra_text_fields = `input_data.rationale` instead of the
diagnosis), 4 gates, review_tooling_provenance (rationale-context variant),
six-way review_fields + check_fields, headers to v11 / image ≥ v1.0.1133.
25 steps; literal balance + routing integrity + reachability all verified
(every next/then/else/error target resolves, no orphan steps). Committed
`d1ab1eacf`.

**State: the launch precondition is fully met** (6 seats + filter live on a
pod-verified image; seed mirrors it exactly). The single remaining step is
the owner's named go to apply `0NN_council_gate.sql` to clients_db, then the
smoke run per RUNBOOK checklist steps 2–5.

## Turn 3 — 2026-07-17 (night) — Named go received; live roster re-read (7 seats!); seed mirrored, APPLIED, smoke fired

Owner: "apply the council-gate seed to clients_db on postgres-clients-0 in
ai-persona-system. Be aware another thread is frequently adding more council
members." The warning was load-bearing: pre-flight read of the LIVE
fix-proposer row (DB, not files) found the council at **7 seats** — an
adoption-pipeline guardian (`review_adoption_guardian` + `gate_adoption`,
footprint key `adoption`) and a `code_lookup` step had landed since the v11
files. Mirrored the seventh seat verbatim from the live row (rationale-context
swap only), re-pointed the gate chain, went to seven-way
review_fields/check_fields, and recorded two standing rules in the seed
header: (1) before ANY re-apply, diff the seed's reviewer steps against the
live fix-proposer row first — the roster moves frequently; (2) `code_lookup`
is a DELIBERATE divergence, not drift (it answers code-shaped questions for
the blind reproposer, which the gate lacks — same reason repropose/reframe
are absent). Validated: 27 steps, routing integrity + reachability + literal
balance all green. Committed `9049ec53a`.

**APPLIED** with the owner's named target: snapshot no-op (first row),
INSERT 1, COMMIT clean. Post-apply verification all green: active
experimental row, 27 steps, seven-way arrays, five gated footprints
(adoption, bug_historian, guidelines, reuse_agent, tooling_provenance), all
seven prompts intact (2.1k–4.0k chars), start=load_schema_hint,
mode=orchestrator.

**Smoke fired** (pod 60m old, well past the 300s window): a genuine
submission — the digest gate-verdicts section (the handoff's own "awareness
channel to extend"), 2 edits, grounded and scoped. Correlation
`bd12762a-5b10-416b-a70d-90ee3067ce7d`, orchestration
`72e552df-c120-4fd2-a80a-3ad8c43d0d3d`. Expected panel behaviour: the
'render'/'_action.go'/'doc_notes' patterns should fire bug-historian,
reuse-agent and tooling-provenance; guidelines + adoption should SKIP — so
one run exercises both the run and abstention paths.

**SMOKE RESULT — PASS, with a real catch.** Completed `complete_revise` in
~100s. The panel behaved exactly as predicted: 5 seats reviewed, guidelines
+ adoption skipped by their gates (no LLM call), council_decide handled the
two abstentions cleanly. Verdict: **REVISE — objection from editquality**
(4 objections + 1 missing, all fixable and correct: struct/summary field
mismatch, an unaddressed renderDigest caller, an uncited shortID dependency,
and the doc_notes claim in the rationale with no covering edit). The
run_checks step answered 8 reviewer queries against live data and — first
real outing — **empirically refuted the plan's core partition assumption:
ALL 17 council_report rows have `source_agent='generic'`** (the persist
actions stamp params.AgentType, which resolves to 'generic' at runtime), so
the proposed `source_agent='council-gate'` filter would have rendered a
PERMANENTLY EMPTY digest section with no error — the exact silent-blank
family the bug-historian watches. The verdict note also honestly reported
"3 further check(s) dropped (max_checks=8) — coverage was capped, not
complete." The digest-section change stays on the revise trail (correlation
`bd12762a`) for whoever picks it up, with the corrected partition key as the
first revision item. **Cross-flag for the fixloop thread: council_report
source_agent is 'generic' fleet-wide — any consumer partitioning gate vs
fix-proposer artifacts needs a different key (e.g. join via the gate's own
fix_plan rows, or stamp metadata at persist time).**

The gate is COMMISSIONED: applied, verified, and proven end to end on a
genuine submission — run path, skip path, abstention, checks, doc_note,
terminal routing all exercised live.

## Turn 4 — 2026-07-18 — Re-synced to 9 seats; CLAUDE.md how-to-use note; fresh read-aloud summary

Owner asked for a read-aloud summary and a note telling threads how to use the
tool, flagging that more council members may exist by now. They did: the live
fix-proposer had grown to **9 seats** overnight — diagnosis-loop guardian
(#4) and improvement-loop guardian (#5), both gated, plus their footprints
(`diagnosis`, `improvement`). Followed the standing rule: read the live row
from the DB, mirrored both reviewer steps verbatim (rationale-context swap,
`error_step` → `complete_invalid`), added both gates into the chain
(adoption → diagnosis → improvement → guardian), extended review_fields and
check_fields to nine. Validated (31 steps, routing + reachability + literals
green), committed `0351e193e`, **re-applied to clients_db** — this time
`snapshot_agent` DID capture the prior row (`be2a7614`, in
`agent_definitions_backup`, rollback available). Verified live: 31 steps, 9
reviewers, 7 gated footprints.

**CLAUDE.md** gained a "Council review of platform changes" section — the
right home for it: every session loads that file at startup, so a how-to-use
note reaches threads that will never read this directory. Kept it short
(submit / verdicts / cost is relevance-gated / coverage report / patch BOTH
councils and diff the live row first) and pointed at the runbook for the
schema and detail.

**`SUMMARY_council_gate_2026-07-18.md`** written to supersede the 17th's
(that one predates the gate going live and now reads as stale) — what we set
out to do, what we did, where we are, where we're going, plain language.

RUNBOOK gained a live-roster section with the verification query and the
re-sync procedure, so the next thread does not have to re-derive it.

## Turn 5 — 2026-07-18 — First outside adoption; three corrections to my own tooling; the .result class closed

Owner asked whether threads need telling to re-read CLAUDE.md, and to start
watching adoption. **Answer: new sessions load it at startup and get it free;
sessions already running hold the old copy and never re-read** — so the note
reaches tomorrow's threads automatically and today's only if told.

**Adoption is real.** The watcher (5-min poll on trailer commits, gate verdict
notes, roster drift) caught, at 13:48, a **second gate submission from another
thread** — the imagery thread, correlation `098b29b8`, and on its **round 2**:
it submitted, got REVISE, revised, and resubmitted on the same correlation,
which is exactly the intended workflow. Separately `f32b208e5` carried a
`Council-Reviewed:` trailer. Two threads have now used the mechanism without
being asked to.

**Three corrections, two of them to my own tools, one by another thread:**
1. `098` was reporting **4 of 41** in-scope commits and looking healthy —
   `kubectl exec -i` inside the read-loop ate the loop's stdin. Found and fixed
   by another thread. (My own heredoc later hit the mirror image: no `-i`, psql
   got nothing, exited 0, printed nothing, changed nothing. Verify the write.)
2. `098` accused an honest commit: it resolved trailers only against
   `correlation_id`, so a fix-proposer-approved commit read as MISMATCH. Now
   resolves correlation **or** run id, by prefix, and says which matched.
3. Then the verdict it pointed at **vanished** — `091`'s documented "clear
   council_reports for a fair run" DELETE ran against that orchestration
   between two runs of the report (approved 12:03 → gone 13:29). Retired that
   advice (round counting is orchestration-scoped in code, so clearing buys
   nothing) and split **EVIDENCE GONE** from MISMATCH, because "we cannot find
   your evidence" is not "you lied".

**Roster mirroring went mechanical.** Seats hit 13 (compliance, render,
llm-reliability, debug-historian). Wrote `099_SYNC_gate_roster.py` — reads the
live fix-proposer row and mirrors every `review_*`/`gate_*` step + footprints,
swapping diagnosis context for the submitter's rationale, refusing on dangling
routes, snapshotting before write. Applied: gate at 13 seats. Another thread
then **deepened it to a step-by-step compare** after finding the name-set check
was blind to config drift (the gate was still on an older model/token setting).
Both councils now verify in sync.

**bugs_open/016 (owner flagged).** The gate is structurally immune — no
`.result}}` anywhere, and no reviser at all. Fixed `fix-proposer`'s
repropose/reframe (snapshot `f9d90a2d`), then, after the experience-loop thread
corrected its own caveat, `content-creator-hero` too (snapshot `d8b5e2c1`):
**fleet sweep now returns zero `.result}}` in any live prompt — class closed.**
Filed a second finding not fixed: `repropose` sees only 6 of 13 seats (the
newer 7 are in neither its prompt nor `input_fields`), left to the seat-owning
thread because adding sections is not idempotent under concurrent edits.

**Compliance:** created `PLAN_2026-07-17_council_gate.md` — the workstream had
RUNBOOK/NOTES/SUMMARY but no PLAN, and CLAUDE.md's new standing-four directive
requires one. Backfilled with decisions, reasons, and the corrections above.

## Turn 6 — 2026-07-20 — v1.0.1140 verified; round 2 ran (10/16 seats); I called a queue delay a dropped dispatch

Owner: "a fresh chassis build has been deployed."

**Image verified against the pod, not the tag.** v1.0.1140 carries the
`bugs_open/036` fix — `objectionEdit` 7 symbol hits, so a reviewer answering
`"3"` instead of `3` now costs one seat, not the whole paid round — plus 019's
truncation fix. Read the code to confirm the abstention path my gated seats
depend on survived the struct change: intact, now with an `abstained` counter.
Note `load_council_reviews` greps **0** in the binary and that is correct — it is
a workflow step name in DB config, not a Go symbol.

**Roster moved again while I was away: 16 seats** (new: constitution, mission,
prior_art_librarian). `099` dry run: zero drift, footprints identical — another
thread ran the mirror. The mechanical sync is doing its job without me.

**Exercised the gate on the new binary with real work** — round 2 of the digest
gate-verdicts change on `bd12762a`. This time I verified the discriminator
BEFORE submitting (round 1's lesson): correlations with a `kind='bundle'`
artifact are diagnosis-backed (2 correlations / 14 reports) vs without (37 /
110); spot-checked e505f70f=true, 098b29b8=false, bd12762a=false. Stated the
honest limit in the submission itself — "no bundle" means "not diagnosis-backed",
which lumps the gate together with feature-designer and experience-planner, not
a gate-only filter.

**Verdict: REVISE, 10 of 16 seats** (2026-07-20 19:28, run `0b8bcc1b`). Four
objected. The good result: **three seats independently caught the same real
risk** — `body::jsonb` cast unguarded against a non-JSON row (guardian,
bug_historian, debug_historian). Also caught: the edit declares one file but
changes `fixloop_digest_test.go` too; gatherer error handling unstated;
travelling-docs NOTES entry unaddressed. That is the council doing exactly what
it is for.

**MISSTEP — I called a 29-minute queue delay a dropped dispatch.** No
orchestration row appeared 13 minutes after publish, so I wrote "the dispatch was
dropped, not slow", retried it (a duplicate 16-seat round's worth of credits at
risk), and spent ~25 minutes probing Kafka and reading the topic tail for a
payload-size threshold. The run started at 19:20:36 — **29 minutes after the
18:51:57Z publish** — from the FIRST dispatch, and completed normally. My own
RUNBOOK carries that exact trap, and I had quoted it to the owner earlier the
same session. Logged in `WRONG_CALLS.md`; the standing "wait / query again before
calling an absence a failure" tally went to 4. Second, smaller misstep in the
same stretch: a poll loop matching `*completed*` against a status of `COMPLETED`
never broke and burned 10 minutes.

Two facts corrected in CLAUDE.md as a result: a run is **~30 minutes**
end-to-end, not ~2 (with the find-your-run-by-payload query attached), and the
seat count is 14-of-16 gated with a live query to check it rather than trusting
the line.

**Docs for the next chat:** wrote `HANDOFF_2026-07-20_council_gate_thread.md`
(supersedes the 07-17 one, which described a gate that did not yet exist).

<!-- Append new turns below this line. Format: ## Turn N — date — one-line summary -->

> **Turn 7 was written concurrently by two sessions — merged 2026-07-21.** This
> thread and its sibling "fixloop council on every bugfix" both picked up the
> handoff at once, both ran the same measurement, and both filed against the
> *same single* diagnosis (corr `fa333384-c3bc-4c7e-8d26-105f25ade755`; verified
> in the DB as ONE intake item `needs_diagnosis:council-approve-threshold` — so no
> wasteful duplicate run, and the run had already started when I checked: orch
> `2556c072` COMPLETED, `9d86cb45` executing `route`). My near-identical copy is
> folded into the fuller entry below to leave one record; the collision itself is
> the multi-session hazard, recorded rather than tidied away silently.

## Turn 7 — 2026-07-21 — measured the approval rate (~4.5%); filed a diagnosis on the decision rule

Picked the thread back up. Ran the probes from the handoff first: `016` second
finding still UNEXERCISED (`load_council_reviews` audit count = 0); gate and
`fix-proposer` both at **16 seats**, zero drift.

**On-mission measurement, grounded in `diagnosis_artifacts` kind=`council_report`
`metadata->>'decision'`:**

| window | revise | rejected | approved | distinct corrs | approvals |
|---|---|---|---|---|---|
| 2 days | 73 | 2 | 1 | 25 | 1 (`17be3962`, an experience PLAN — not a bugfix) |
| 7 days | 123 | 3 | 2 | ~44 | 2 |

Rounds-per-correlation (2d): threads DO iterate — correlations reach 3/4/5/6/**7**
rounds and still land REVISE. The lone 2-day approval took 5 rounds. `098` in the
same window: **0 REVIEWED, 45 UNREVIEWED** (nearly all `fix(bugs_open/NNN)`), 1
MISMATCH. So adoption is real (25 submissions, 56 gate verdict notes / 2d) but
approval is ~4.5% → trailer-coverage is structurally stuck near 0 and PR-mode is
unbuildable as designed at this rate.

Caveats recorded honestly: council_report is `source_agent='generic'` fleet-wide,
so this rate spans ALL councils, gate is a subset I can't cleanly split; and
approved reports are deletable (handoff item 4), so the approved count is a FLOOR.

**Owner chose: diagnose the low approval rate** (over accept-as-is / recalibrate /
carry-on-mid-flight).

**Quick code read → hypothesis (UNVERIFIED, under diagnosis):** `decideCouncil`
(`diagnose_council_decide_action.go:535-552`) is ordered veto→veto→**object→revise**
→all-approve. Rule 3 returns `revise` on ANY single readable `object`, and
`councilObjection.Severity` (`low|medium|high`, line 87) is **never consulted** in
the decision. So a lone low-severity nit from 1 of 16 seats blocks exactly as hard
as a high-severity flaw, and `approved` needs unanimous bare approve. This would
explain the rate mechanically — but per CLAUDE.md "confidence is not a signal", I
filed it rather than asserting it.

**Filed:** `090` → corr `fa333384-c3bc-4c7e-8d26-105f25ade755`, item
`needs_diagnosis:council-approve-threshold`, REF = current branch (origin tip
`2d529d6dc` confirmed to carry the severity-free rule at line 548),
SEED_SCOPE = `diagnose_council_decide_action.go:decideCouncil`. The symptom points
the loop at `decideCouncil` + unused `Severity`, the `council_decide` config in
`agent_definitions`, and `council_report` `body->'reviews'` — to check whether
each REVISE was decided by a lone `object` while other seats approved, and at what
severity. A REFUTE (deciding objections are substantive/high across seats) is a
success: it says the plans genuinely need the work, not the rule. Budget ~30 min.

## Turn 8 — 2026-07-22 — the loop STALLED at `route`; hand-confirmed the diagnosis from primary evidence

**The 090 run never produced a verdict.** orch `9d86cb45` FAILED, reaped by the
`>4h stale EXECUTING_STEP` reaper (bugs_open/003 F1) **stuck at step `route`** —
one bundle written (iter 1), then nothing. This is the spawn-loss / dispatch-queue
backlog (bugs_open/003, /030) biting the DIAGNOSE LOOP ITSELF: `route` spawns a
code-lookup child and awaits it; that await never returned, so a code-seeded
diagnosis stalls at route. The reaper worked as designed; the diagnosis did not.
**Second finding, worth a separate line: the canonical diagnosis path is
currently unreliable for code-seeded symptoms** — re-firing would likely re-stall
at the same step. Did NOT re-fire; gathered the evidence by hand instead.

**Hand-diagnosis — CONFIRMED. Grounded in `diagnosis_artifacts` kind='council_report'
`body::jsonb->'reviews'`, last 3 days, `metadata->>'decision'='revise'`:**
- Objector-count per REVISE round (88 rounds): **14 rounds decided by a LONE
  objector** against avg 8.8 approving seats; mode is 2–3 objectors of ~9 seats.
  So a single seat routinely overrides 8–9 approvals — the unanimity rule, live.
- **Severity of the 485 objections: medium 279 / low 165 / high 41. High is 8.5%.**
- Lone-objector rounds (14): the single blocking objection was medium in 9, low in
  4, **high in only 1**. A lone seat blocks 8–9 approvers on a non-high nit 13/14×.
- **Counterfactual (the sizing number): 59 of 88 REVISE rounds (67%) had ZERO
  high-severity objections and no veto** → would flip to approve/approve-with-notes
  under a rule where only high (or veto) blocks. **25 of 32 revised correlations**
  reached such a round. So approval is unreachable because of the RULE, not the
  plans: `decideCouncil` (`diagnose_council_decide_action.go:535-552`) returns
  revise on ANY `object` and never reads `councilObjection.Severity` (line 87).

**Confidence, marked honestly:** the STRUCTURAL defect is certain from the code
(severity field exists, decision ignores it). The rate-causation is confirmed by
primary DB evidence (queries above, reproducible). What I do NOT have: the loop's
INDEPENDENT council grading (it stalled) — so this is a hand-diagnosis, not a
council-graded verdict. And `severity` is reviewer-self-reported: if seats
under-label, some low/medium are really blocking — which is why any fix must stay
conservative (keep high + veto blocking; the design question is medium). `[ASSUMED]`
that reviewers' low/medium are mostly genuine nits — the 67% is an upper bound on
what a severity-gate would flip, not a promise.

**Fix direction (NOT yet built — fleet-wide behaviour change, owner's call):** wire
severity into `decideCouncil` — high objection or veto still blocks; low (and
possibly medium unless ≥N seats raise it) becomes approve-with-notes: recorded and
returned to the proposer, non-gating. Preserves the objections' value (still
surfaced) while making approval reachable, which unblocks trailer-coverage and
PR-mode. Affects ALL councils; should itself go through the gate once. Taking the
decision to the owner (threshold shape) before writing anything.

## Turn 9 — 2026-07-22 — widened the guidelines footprint to cover council logic; found + fixed a 099 aliasing bug that silently blocked footprint syncs

Owner: "I'd like guidelines to be in the council" (for the severity-gate change).
The `guidelines` seat is relevance-gated (`gate_guidelines` → `panel.run_guidelines
== true`); its footprint targeted contracts/schemas/agent-defs and did NOT match
council *code* like `diagnose_council_decide_action.go`, so it would not fire.
Owner chose (AskUserQuestion): **widen the footprint to cover council logic** (over
one-off / always-on).

**Matching surface (read `select_review_panel_action.go` first):** a seat fires if
any footprint token is a case-insensitive substring of any edited FILE PATH **or**
of the plan/rationale text corpus (`plan_persisted.plan_json` + extra_text_fields).
Plain substrings, not regex.

**Change (config, live immediately, source-of-truth = fix-proposer):** snapshotted
fix-proposer (`f9d90a2d`), appended `council_decide` + `select_review_panel` to
`select_panel.config.footprints.guidelines` via `jsonb_set ... || '[...]'`. Then
mirrored to the gate with `099_SYNC_gate_roster.py --apply` (NOT a hand-patch of
the gate — the standing rule). Verified: both councils identical, and
`council_decide` is a substring of `diagnose_council_decide_action.go`, so
guidelines will now fire for the dogfood submission and every future
council-logic change.

**BUG FOUND & FIXED in 099 (the reason the first --apply would have written
nothing).** The first dry-run reported `drift (none)` AFTER I changed
fix-proposer's footprint — wrong. Cause: `new_steps` was a SHALLOW copy of the
gate's carried-over step objects (line 107), so line 126
`new_steps["select_panel"]["config"]["footprints"] = fp_panel["footprints"]`
**mutated the gate object too** (aliasing). The deep drift compare (`norm(new) !=
norm(gate)`) then saw both sides change together → no drift → `changed=False` →
`if not changed: return` skips the write. Its own docstring claims it mirrors the
footprints map; the aliasing meant it silently could not (same latent class for
`council_decide.review_fields` and `run_checks.check_fields` if they were the ONLY
change). **Fix:** `copy.deepcopy` the carried-over steps. Dry-run then correctly
reported `drift: ['select_panel']`; --apply wrote; a final dry-run reports in-sync.
`[LESSON]` a mirror that mutates one side of its own before/after comparison can
report "in sync" while being out of sync — verify a sync by reading the TARGET
row, not by trusting the tool's drift line.

## Turn 10 — 2026-07-22 — severity gate LIVE on v1.0.1149; dogfood APPROVED (first ever); guidelines fired; objections harvested

**Owner deployed a new chassis build (v1.0.1149).** Verified the fix live against
the POD, not the tag: `objectionGates`/`severityGates` + literals `"advisory
objection"`/`"gating objection"` present (control `decideCouncil` present).
**Behavioural proof (pod-grep only proves deployment):** an ORGANIC round decided
`approved` — decided_by "approved with 4 advisory objection(s) — none
high-severity" (corr `cbbc7c83`, 14:07). A 4-objection round that was a guaranteed
revise yesterday now approves.

**Dogfood: APPROVED** (SUBMISSION_CORR `e0a9b843`, orch `26b87be3` →
`complete_approved`, 14:57). **First approved gate submission.** 9 seats ran, 7
abstained (relevance filter). Seat verdicts: constitution/debug_historian/
diagnosis_guardian/**guidelines**/mission/reuse_agent = approve (0 obj);
editquality = approve+1 low note; prior_art = approve+1 low note; **guardian =
object (medium + low)**. Decided_by "approved with 1 advisory objection(s) — none
high-severity" — `advisory` counts OBJECTING SEATS (guardian, verdict=object),
not individual objections. **The demonstration:** the guardian (always-on
hard-veto seat) OBJECTED medium+low and wrote "the plan is right that this is the
correct convergence point ... so I'm not vetoing"; old rule → revise, new rule →
approved. The strictest seat, exercising the exact behaviour we changed.

**GUIDELINES FIRED and approved** — the footprint widening (Turn 9) worked: a
council-logic change (`diagnose_council_decide_action.go` contains `council_decide`)
drew the guidelines seat. The owner's request is satisfied and proven on a real run.

**Objections harvested (advisory = optional, but two were fair):**
- guardian (low): no test for a Degraded review with ZERO objections (Degraded and
  empty-objections branches each covered, not their conjunction). **Added**
  `TestObjectionGates` case "degraded object with zero objections → gates". Green.
- editquality (low): the ruling says low/medium must be "recorded and returned to
  the proposer", the sketch only showed the count string. They ARE returned — the
  full `reviews[]` (objections included) is persisted in the report body and read
  by the reviser (the 016 fix). A sketch-visibility gap, not a code gap. `[NOTE]`
- prior_art (low): can't verify the "live on v1.0.1149" liveness claim from the
  plan. Fair — inherent to a retrospective dogfood.

**Trailer/coverage limit (honest):** the reviewed change (`872c830a8`) was
committed BEFORE the dogfood approval (forward-only, no amend), so it can't carry
`Council-Reviewed`. The harvested-test commit does carry the trailer — it's part of
the approved change and touches a file that was in the approved plan. The approval
is recorded against corr `e0a9b843`; 098's auto-join won't see the main commit.
This is a property of dogfooding a LIVE change, not a gap in the fix.

**`bugs_open/057`:** fixed & live, two live approve examples (organic + dogfood),
floor test-covered. Ready to close; holding only for a live GATING example (none
organic yet — councils quiet). Milestone summary: `SUMMARY_2026-07-22b_council_approval_rate_live.md`.

## Turn 11 — 2026-07-24 — PR-mode reopened → DEFERRED (strengthen advisory instead); commit-nudge + norm + 098 baseline

Owner: "please reopen" (PR-mode). Measured the precondition (does the fix make
approval reachable enough for enforcement?): approval **5% → 80%** (before fix
07-15..07-22: 3 approved / 143 revise / 4 rejected across 53 corrs; after
07-22 14:00: 8 approved / 5 revise / 2 rejected across 10 corrs), and `098`
REVIEWED **0 → 4** in 3 days (my `e37ec804f` + 3 other threads' council-approved
commits: bugs_open/056, 054, dead_controls — all "by correlation"). So PR-mode is
now buildable. **Owner chose (AskUserQuestion): strengthen advisory, DEFER
structural.** Reason PR-mode is costly here: it collides with many-sessions/one-
shared-branch — platform fixes land directly and ride the next sweep build; fix/*
PRs hold them until merge + add a council round of latency = a workflow change for
every thread. PLAN open-Q #1 marked RESOLVED.

**Delivered (loud + regular):**
- LOUD — a **commit-msg** advisory nudge (`scripts/council-coverage-nudge.sh`):
  prints when a commit touches `platform/|internal/|pkg/` code with NO
  `Council-Reviewed:` trailer; silent on reviewed/docs commits (commit-msg, not
  pre-commit, so it can read the message and check the trailer). NEVER blocks
  (trap + exit 0). Wired into `.githooks/commit-msg` AFTER the existing D2
  direction gate. **CRITICAL: preserved the direction gate verbatim** — it is a
  BLOCKING gate; verified live it still blocks a blessed doc w/o trailer (exit 1)
  and passes with one, and the nudge never blocks. Its early `exit 0`s became
  fall-throughs so the nudge runs on normal commits.
- NORM — hardened the CLAUDE.md council section: approval is reachable now (~80%),
  so "submit platform changes to the gate" is a real norm; still advisory,
  PR-mode still deferred.
- REGULAR — persisted a `098` coverage baseline to doc_notes (`PERSIST=1`,
  categories digest+council-gate).

`[LANDMINE]` the shared tree switched branch mid-session (085→086, another
session) — my commits are recorded on whatever branch was HEAD at commit time;
the hooks/CLAUDE.md are working-tree files so they are live for the shared tree
regardless of branch. Forward-only; did not fight it.

`[OPEN]` truly-automated "regular": a cloud cron agent likely lacks kubectl to the
cluster (098 needs BOTH git + DB), so a fully-scheduled run needs a runner with
both — recommend running `098 --persist` at the digest cadence; a scheduled job is
an owner call (offered).

## Turn 12 — 2026-07-24 — `057` CLOSED: the gating floor arrived organically (a different session, "bugfix 057")

Turn 10's hold ("holding only for a live GATING example — none organic yet") was
satisfied within hours and nobody looked back: the first organic gating round was
the **guardian hard veto at 07-22 17:08 (corr `1d8ef2c0` R1** — the same corr the
056-regen thread later took to an R2 APPROVED). Checked the full post-fix window
(07-22 14:00 → 07-24 15:49, 25 organic rounds):

- **13 approved** — every one "approved with N advisory objection(s) — none
  high-severity"; no bare-unanimity round needed.
- **9 revise** — every one "gating objection from <seat>" and every one carried
  ≥1 *explicitly high* objection when the reviews were checked per-round
  (editquality ×5, contracts ×4; e.g. corr `c2a9fd27` 07-23 15:56, 2 high).
- **3 rejected** — all vetoes (guardian `1d8ef2c0`, feasibility `fa4b77cd`,
  guardian `6cdbc374`).
- The case file's own discriminator over the post-fix window: **0/9** revise
  rounds without a high objection (pre-fix 59/88). One query gotcha for reuse:
  post-fix some reviews omit `objections` entirely, so wrap it
  `coalesce(rv->'objections','[]'::jsonb)` or the EXISTS silently skips rows.

That is both halves of the contract observed live — approve-on-nits AND
gate-on-high/veto — which is exactly what "verify the failing branch" wanted and
the unit tests could only promise. **Moved the case to `bugs_closed/057`**, added
the §10 index row, pointed CLAUDE.md / this dir's PLAN+RUNBOOK at the new path.
Dedup checks before acting: who-owns (this dir active but on `059`), no 057
commits since 07-22, `site_work_items` clear. No code touched; nothing for the
next image roll.

## 2026-07-29 — contributed by thread "bugsearch 5" (bugs_open/138): the trailer cannot reach the commit it approves, and that is now the DEFAULT case

Not a request and not a bug filing — this is your mechanism, so it is recorded
here for you to judge. I hit it doing exactly what the norms ask.

**What happened.** Submitted a platform change to the gate (corr
`919a05bf-c51a-440b-865e-bd07e69e1c36`), got **APPROVED**, 11 seats, 0 unreadable.
The code had already been committed (`3a59b5012`) *before* the verdict, because
the **owner ruling of 2026-07-29** retired the ordering exemption's condition (1)
on the grounds that no thread on this tree can hold a change back — HEAD is shared
and any other session's roll ships your commit. So the ruling makes
commit-then-submit the honest default.

**The consequence.** Forward-only forbids amending a trailer in, so `3a59b5012`
can never carry `Council-Reviewed:`. Verified, not assumed — `098` at a 1-day
window buckets it under **UNREVIEWED**, one of 40, against 8 REVIEWED.

**Why this is not the adoption gap already in `NOTES_running_feature_builder.md`
(~07-24).** That entry says the trailer is "an un-adopted convention rather than a
personal lapse" — a compliance observation, and true. This is different in kind:
**a fully compliant thread — submitted, approved, trailer discipline understood —
still lands in UNREVIEWED.** So the bucket now conflates two populations that mean
opposite things:

- never submitted (the thing the report exists to surface), and
- submitted, approved, and committed first (the thing the 07-29 ruling asks for).

**And the ratio moves the wrong way as compliance improves.** Every thread that
follows the ruling adds to UNREVIEWED. The report's headline number therefore
understates review coverage by an amount that grows with adoption — which is the
one direction a visibility metric must not drift, because it makes the norm look
ignored precisely when it is being followed.

**Not proposing a fix** — options all have costs that are yours to weigh (a
corr→commit mapping written at verdict time; `098` accepting a trailer on any
commit in the window that names an approved correlation; or simply documenting the
two populations as distinct in the report's own output). The honest-limits block at
the top of `098` already models exactly this kind of disclosure, which is why I
think the finding belongs to that block rather than to a code change.

One thing worth stating either way: **`8 REVIEWED / 40 UNREVIEWED` in a single day
should not be read as 17% compliance** until this is resolved.

> **CORRECTED 2026-07-29, same day, by the thread that wrote the entry above — the
> attribution was wrong, and the owner caught it.** I wrote that the 07-29 ruling
> "makes commit-then-submit the honest default". **It does not, and three separate
> checks say so:**
>
> 1. **The ruling is about a different axis.** Q2 asked whether a seam must ship
>    behind a **default-OFF switch**; the answer was no, and "review here is after
>    the fact" means *after the change is LIVE* — because HEAD is shared and any
>    session's build ships your commit. It constrains **submission** timing only:
>    *"submit to the gate before or alongside committing."* It says nothing about
>    waiting for the **verdict**, and does not forbid it.
> 2. **The commit-early norm predates it by nine days and has a different source** —
>    owner feedback of **2026-07-20**, after a thread held the `bugs_open/011` fix
>    uncommitted across four council rounds and the OWNER's own sweep commit
>    (`bca5d8255`) took it to production with the verdict still REVISE. That
>    guidance also already anticipated this state: *let the trailer "or its
>    deliberate absence" record the review status*, and **state the verdict status
>    in the message body** — which I did not do on `3a59b5012`, and which would
>    have left exactly the corr↔commit link I then went looking for.
> 3. **Measured, which is what I should have done before writing "default":** of the
>    trailered commits that day, **3 of 3 sampled were committed AFTER their
>    approval** — `f5fc3014` approved 17:17 → committed 17:19 (2 min), `49392838`
>    13:30 → 14:26, `7ba5b8c4` 13:33 → 14:24. Threads routinely wait. "Default" was
>    an inference from my own single case, stated in the voice of a finding.
>
> **What survives, and it is sharper than what I claimed.** Not a policy
> consequence — a **standing tension between two live practices**: the 07-20
> feedback says commit the moment the work is coherent and let the trailer's
> absence carry the status; the trailer mechanism only ever attaches to a commit
> made *after* approval. **A thread cannot satisfy both**, and each resolves it ad
> hoc — which is why `098`'s UNREVIEWED bucket mixes "never submitted" with
> "approved, but committed first". That is worth a decision; blaming the 07-29
> ruling for it was not.
