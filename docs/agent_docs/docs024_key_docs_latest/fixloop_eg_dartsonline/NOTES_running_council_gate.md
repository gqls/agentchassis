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

<!-- Append new turns below this line. Format: ## Turn N — date — one-line summary -->
