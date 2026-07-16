# RUNNING NOTES — Concept Register (docs026)

**What this project is about (read this first).** A three-stage programme to
extract every concept from agentchassis's `docs/` tree (~4,111 files), verify
each against the live code/DB, and eventually build an expert council agent per
concept-area for the fix-loop. The plan lives in `PLAN_concept_register.md`; the
human operator's tasks live in `RUNBOOK_concept_register.md`.

These notes are the turn-by-turn record of the discussion and decisions. Newest
entries at the bottom. Update every turn.

---

## Decision log (running summary — see turns for context)

| Date | Decision | Status |
|---|---|---|
| 2026-07-13 | Stage 1 (extract + consolidate) complete: 1,627 concepts, 107 categories, from 2,185 raw blocks | **confirmed** |
| 2026-07-14 | Stage 2 scope widened from "verify 314 partial/unknown" to also sweep all 871 `deployed` concepts for false positives | **confirmed** (user-approved) |
| 2026-07-14 | Stage 2 mechanism: multi-agent Workflow (verify → adversarial recheck pipeline) rather than sequential batches or high-value-only | **confirmed** (user-approved) |
| 2026-07-14 | MCL-002 (va001 second cluster) corrected `deployed → aspirational` — zero footprint in code/config/deployment, only in archived present-tense-plan prose | **confirmed** |
| 2026-07-14 | LoRA/Thunder evidence tension dissolved — adapter genuinely trained (MDL-029, deployed) but never wired into inference serving (FTW-003, partial); both signals were correct | **confirmed** |
| 2026-07-14 | PUB-001 confirmed a genuine duplicate of ADM-007 + ADM-008 | **confirmed, merged** (retired to pointer entry) |
| 2026-07-14 | Added 7th status `convention` to the register vocabulary, for design doctrines/practices stage 1 had defaulted to `deployed` | **confirmed** |
| 2026-07-14 | Stage 2 substantially complete: 105 corrections confirmed (97 overturned by adversarial pass), applied to register + index | **confirmed** |
| 2026-07-14 | Established this file + `RUNBOOK_concept_register.md` + `PLAN_concept_register.md` as the workstream's standing doc practice | **confirmed** (user-requested) |
| 2026-07-14 | Extended stage-2 sweep to the 102 superseded + 72 abandoned buckets (last unswept portion of the register) | **confirmed, complete** (18 corrections, 9 overturned) |
| 2026-07-14 | Stage 2 fully complete: all 1,627 concepts checked, 124 corrections total (~7.6% error rate) across all three batches | **confirmed** |
| 2026-07-14 | Stage 3 design questions (granularity/activation/freshness) resolved into a concrete recommendation grounded in the live fix-loop mechanism | **confirmed** (design only — implementation deferred to user) |
| 2026-07-16 | fixloop's triage/escalation subsystem (shipped after extraction froze) added to the register: `FIX-051`/`052`/`053`, `FIX-034` updated, `STY-049` (the missingkey=zero structural defect) | **confirmed** (independently verified against live code + a dedicated research pass) |
| 2026-07-16 | Stage-3 pilot seat: two data-driven candidates recommended (reuse-agent/tool-lifecycle; bug-historian/silent-content-loss family) | **confirmed** (recommendation only — pick + build deferred to user) |
| 2026-07-16 | Flagged to owner: fixloop's case-004 (image-landing trap) dispatch may be moot — a separate session resolved 2 of its 3 open items the same day | **flagged, not actioned** (fixloop-workstream's decision) |
| 2026-07-16 | User picked bug-historian over reuse-agent as the pilot seat to build; wrote a read-aloud summary doc + the full pilot spec (charter, curated context, prompt, exact patch) | **confirmed** (spec complete, application deferred to user) |
| 2026-07-16 | Bug-historian APPLIED to production `clients_db` (postgres-clients-0/ai-persona-system) with user's explicit named sign-off — council is now 3 reviewers | **confirmed, live** (verified via direct DB read; not yet exercised on a real run) |

---

## Turn 1 — 2026-07-13 — Stage 1 complete, handoff written

Prior session built the full register: 34 extraction units swept ~4,111 files,
2,185 raw concept blocks consolidated into 107 category files (1,627 concepts),
master index at `register/000_concept_index.md`. Handoff written recommending
stage 2 start with the two flagged evidence tensions (multi-cluster dispatch,
LoRA/Thunder) and three likely-duplicate concepts.

## Turn 2 — 2026-07-14 — New session; verified register intact on disk

Confirmed nothing was lost between sessions: 107 register files, 1,627 index
rows (+ header), 32 extraction archives all present. Read `README.md` method
spec and computed the status distribution: 871 deployed / 274 partial / 271
aspirational / 99 superseded / 72 abandoned / 40 unknown.

## Turn 3 — 2026-07-14 — Hand-verified the handoff's priority items (batch 1)

Checked the two evidence tensions and three suspected duplicates directly
against the repo (grep/find, no agents yet):
- **Multi-cluster dispatch (MCL-001/002/003/004):** the Go code
  (`dispatch_actions.go`, `remote-job-spawner/main.go`) exists, is registered
  (`registry.go:95`), and is deployed to production (uk_001 kustomize +
  terraform) — but zero workflows/agent_definitions reference it, and the
  claimed second cluster (`va001`) has zero footprint anywhere outside archived
  prose. MCL-002 corrected `deployed → aspirational`.
- **LoRA/Thunder (MDL-029/FTW-003):** tension dissolved — the adapter is
  genuinely trained and closed out (828MB, Llama 3.3 70B, final loss 0.266,
  completed 2026-06-04) but `ollama-adapter`'s deployment only pulls stock
  models; zero code references the adapter. Both original signals were
  correct, just describing different things (trained vs. served).
- **Duplicates:** PUB-001 confirmed a genuine duplicate of ADM-007 + ADM-008.
- Wrote up full findings in `006_VERIFICATION_stage2.md` (batch 1).

**Scope finding that reshaped stage 2:** the one true status error
(MCL-002) came from the `deployed` bucket, not partial/unknown — the bucket
stage 2 was originally scoped to check. Hypothesis: a plan doc narrating its
own design in the present tense reads as evidence of completion, and
consolidation's "prefer the newest, most specific evidence" tie-break actively
selects for this failure.

## Turn 4 — 2026-07-14 — User scoping decisions

Asked the user two questions given the scope finding: (1) widen stage 2 to also
sweep all 871 `deployed` concepts, or stick to the original 314 partial/unknown?
(2) run the sweep as a multi-agent Workflow, sequential batches, or high-value-
only? User chose: **widen to deployed + partial/unknown (1,185 concepts total)**,
run as a **multi-agent Workflow**.

## Turn 5 — 2026-07-14 — Computed work units, built the verification workflow

Computed per-category concept counts (314 partial/unknown, 871 deployed),
chunked into 145 bounded work units (68 "deep" units for partial/unknown, 77
"sweep" units for deployed, ≤45 concepts each) via a small Python script.
Wrote `stage2_workflow.js`: a pipeline of verify-agent (follows `verify-later`
pointers into the repo, classifies `kind`, proposes `verified_status`) →
adversarial-recheck (independent agent tries to refute any proposed correction
before it's accepted).

## Turn 6 — 2026-07-14 — First launch failed (args-serialization bug)

Launched the workflow with the 145 work units as `args`. Failed instantly (0
agents ran, 69ms) — `pipeline()` got a non-array; the recovery snippet showed
`args` had round-tripped as a JSON-encoded string rather than the array passed
in. Fixed by making the script defensive: `JSON.parse(args)` if it arrives as a
string, with a clear error if it's still not a non-empty array after that.

## Turn 7 — 2026-07-14 — Relaunched successfully

Resumed from the same run ID with the fix in place. Ran clean: 145 units, 347
agents total (verify + adversarial), 1,189 verdicts, ~44 minutes, 11.0M tokens,
3,101 tool calls. 105 corrections confirmed, 97 overturned by the adversarial
pass.

## Turn 8 — 2026-07-14 — Applied corrections; stage 2 write-up

Reviewed the result: confirmed the adversarial gate does real work in both
directions (e.g. overturned several `deployed → convention` proposals in
business-strategy after finding real enforcing code — a live £29 Stripe charge
driving one of them). Found the biggest correction bucket (48 of 105) was a
taxonomy gap, not a truth error: concepts tagged `deployed` that are actually
design doctrines/working practices, not code artifacts — added `convention` as
a 7th status. Applied all 105 corrections directly to `register/<category>.md`
files (status line + inline `stage2-verified` provenance note) and
`register/000_concept_index.md` (Status column) via a Python script — verified
zero misses across 43 category files. New distribution: 847 deployed / 246
partial / 290 aspirational / 102 superseded / 72 abandoned / 21 unknown / 49
convention (sums to 1,627 ✓). Updated `README.md` status vocabulary,
`000_concept_index.md` intro, and the `concept-register-workstream` memory file
to reflect stage 2's substantial completion.

## Turn 9 — 2026-07-14 — Established the doc practice (this file + RUNBOOK + PLAN)

User asked for a running docs file updated every turn, plus a runbook and a
plan. Followed this repo's house convention (matched against
`docs024_key_docs_latest/imagery/RUNNING_NOTES_imagery_best_in_class.md` /
`RUNBOOK_*` / `PLAN_*`) rather than inventing a new format. Created
`PLAN_concept_register.md` (overarching 3-stage plan + stage-3 open design
questions + backlog), `RUNBOOK_concept_register.md` (human tasks: credential
rotation, PUB-001 merge go-ahead, superseded/abandoned sweep decision, stage-3
kickoff), and this file (backfilled turns 1–9 from session context since no
running-notes file existed for this workstream before now).

## Turn 10 — 2026-07-14 — Carried on: PUB-001 merge, launched superseded/abandoned sweep, resolved stage-3 design

User said "please carry on" — worked through the RUNBOOK's open items in
readiness order:

1. **PUB-001 merge (B2):** read PUB-001, ADM-007, ADM-008 in full — they
   already cross-referenced each other from stage 1. Retired PUB-001 to a
   pointer entry in `register/public-api.md` (kept the ID and full text for
   its distinct P2 source citation; added an explicit `duplicate-of` relation
   and a stage2-verified note), updated its index row. Chose not to fold
   `public-api.md` away — single-concept category files are already normal in
   this register.

2. **Superseded/abandoned sweep (B3):** computed work units for the last
   unswept portion of the register (102 superseded + 72 abandoned = 174
   concepts, 73 bounded work units). Extended `stage2_workflow.js` in place
   with two new prompt builders — `supersededPrompt` (hunts for claimed
   replacements that don't actually exist, meaning the "old" mechanism is
   still the live one) and `abandonedPrompt` (hunts for ideas quietly
   resurrected since the abandonment note was written) — reusing the same
   VERDICT_SCHEMA/ADV_SCHEMA and verify→adversarial-recheck pipeline shape.
   Launched as task `wmezfdgyg` (run ID `wf_ef077538-cea`). Result pending —
   next turn should report it.

3. **Stage 3 design (B4):** read `FIX-036` (the council-roster-expansion
   vision) plus `FIX-014/015/020/043` in `register/fix-loop.md` to ground the
   design in the actual live mechanism rather than invent one. Found: the
   council today is two workflow steps (`review_editquality`,
   `review_guardian`) in `0NN_fix_proposer.sql`, aggregated by a deterministic
   Go action (`diagnose_council_decide`); adding a reviewer is just a new
   named step + prompt + optional `hard_veto_from` entry; both reviewers
   currently share identical context with "no per-reviewer curated corpus yet"
   (FIX-043, Q-G — explicitly the open question this register was meant to
   answer). Resolved the three open questions against that mechanism:
   relevance-filtered subset (not all 107, cost/latency and irrelevance both
   argue against a fixed roster), activation via matching the fix plan's
   touched files/tables against each category's `verify-later` footprint (not
   a symptom-keyword match), and freshness via curated-corpus-for-framing +
   mandatory live-recheck-of-the-specific-pointer for anything load-bearing
   (mirrors the stage-2 pattern, cheap per-run since only a handful of
   concepts are touched per fix). Wrote this into `PLAN_concept_register.md`
   §Stage 3. **Deliberately did not touch the live
   `fixloop_eg_dartsonline/0NN_fix_proposer.sql`** — that file belongs to the
   separately-active fixloop-workstream, and changing a production
   decision-gating workflow is a cross-workstream call for the user, not
   something to proceed on unilaterally (consistent with FIX-035's own
   "awareness before autonomy" rule).

Updated `PLAN_concept_register.md` (stage 2/3 status, backlog) and
`RUNBOOK_concept_register.md` (checked off B2/B3/B4, updated the standing
summary) to match.

## Turn 11 — 2026-07-14 — Superseded/abandoned sweep landed; stage 2 fully complete

Task `wmezfdgyg` completed: 73 units, 174 verdicts, 100 agents, 2.7M tokens,
~7.6 minutes. **18 corrections confirmed, 9 overturned** (~10.3% error rate,
consistent with batch 2's ~9%). Notable findings:

- A **scheduler-and-tasks cluster** (SCH-007/008/009, all `abandoned →
  deployed`): three independent confirmations of the same drift — real fixes
  landed in `cmd/scheduler/main.go` (ownership gap, starvation prevention,
  fire_message handling) but the docs were never updated to reflect it.
- A **new failure-mode class**, distinct from batches 1-2's present-tense-plan
  problem: **search-scope gaps**. DOC-064 was tagged `abandoned` because
  extraction's search was scoped to one doc subtree (`idea.uk/`) and missed a
  sibling live project folder (`adoption/docubundle/`) holding a byte-identical
  copy — evidence never found, not misread.
- Several "bundled" concepts turned out to be **half-superseded**: ADM-010,
  DES-009, STY-036, STY-039 each claimed a clean replacement but the old
  mechanism was still live and wired for part of what the concept bundled
  together.

Two entries (SYS-025, SYS-026) tripped a bug in the apply script: an italic
"merged from N independent findings" annotation line right after the header
broke the block-capture regex (it only matched consecutive `- **` bullet
lines). Patched those two by hand rather than fixing the script generally,
since only 2 of 121 corrections across both batches hit it.

Applied all 18 to the register + index. **Final status distribution: 853
deployed / 257 partial / 290 aspirational / 90 superseded / 67 abandoned / 21
unknown / 49 convention** (1,627 total). Stage 2 grand total across all three
batches: **124 corrections confirmed, 106 overturned, out of 1,627 concepts
(~7.6% error rate)** — every concept now checked at least once.

Wrote up batch 3 in `006_VERIFICATION_stage2.md` (including the scheduler
cluster and search-scope-gap observations), updated the master index intro,
`PLAN_concept_register.md` (stage 2 → COMPLETE, backlog trimmed),
`RUNBOOK_concept_register.md` (B3 checked off, A1's numbers corrected), and the
`concept-register-workstream` memory file to reflect full stage-2 completion.

**Stage 2 is done. Remaining open items are both user-facing, not agent work:**
credential rotation (RUNBOOK B1) and the stage-3 implementation decision
(RUNBOOK B4 / PLAN §Stage 3 scope boundary — a cross-workstream change to the
live fix-loop council that needs the user's sign-off, not something to
proceed on unilaterally).

## Turn 12 — 2026-07-16 — Doc sweep + fixloop coordination: 4 new concepts, a stale-case finding, pilot-seat recommendation

User asked to survey docs created/updated since the last sweep (2026-07-14) and
carry on the plan while coordinating with the fixloop ("diagnose") thread.

**Sweep:** 62 commits since 2026-07-14 across many concurrent workstreams.
Confirmed via file mtimes that `docs026_concept_register/` itself was
untouched by anyone else — no drift to reconcile in the register's own files.
Read `fixloop_eg_dartsonline/SUMMARY_where_we_are_2026-07-16.md` (the fixloop
thread's own journey doc) — it independently names this workstream
("search-tab2") as the answer to its council-widening question and states the
identical scope boundary already in `PLAN_concept_register.md`: implementing
council seats against the live workflow is reserved for owner sign-off.

**Register gap found and closed:** fixloop's triage/escalation subsystem (4
phases, all live v1.0.1117→v1.0.1123) shipped entirely after extraction froze
on 2026-07-13, so none of it existed in the register. Added `FIX-051` (triage
router), `FIX-052` (silent-check verifier), `FIX-053` (feedback close-out), and
updated `FIX-034` in place (Phase 4 digest addition). Verified every claim
independently — direct greps for registry entries, action files, function
line numbers — then cross-checked against a dedicated research agent's
independent pass, which additionally confirmed every cited commit against
`git log --oneline --all` and `git merge-base --is-ancestor`. Everything
matched; the agent added extra precision (e.g. capability-gap routing reuses
the pre-existing `WriteBuildItemsAction`, not new code; the two named
silent-check checks `nav_linked_never_built`/`deployed_zero_components`).

**Real-case investigation:** fixloop's real-case queue picked
`aaa_fails_to_mend/004` (image-landing/article-body blanking trap) as its
first dispatch target. Traced the mechanism to `call_agent.go:1152`'s
`Option("missingkey=zero")` and wrote it up as `STY-049`, cross-linking it to
a **cross-cutting failure family** discovered while researching relations:
`TL-001` (tool-lifecycle.md, tool-widget-clobber), `PBP-012`/`PBP-019`
(page-build-pipeline.md), `STY-004`/`STY-019` (styling-render-pipeline.md),
and `CLC-003` (tool-library.md) are all the same "schema says required,
renderer says silently empty" shape, independently recurring across at least 5
categories.

**Critical finding, surfaced not actioned:** `aaa_fails_to_mend/004`'s "3 open
items" framing is stale. A separate concurrent session
(`article-body-json-envelope-workstream`) resolved the underlying data loss
the same day — `005_HANDOFF...FIXED.md` (mtime 17:52, later than 004's 11:58)
shows all 17 article-body instances recovered; independently re-ran `go test
./platform/orchestration/actions/... -run TestParseLLMJSON` myself and
confirmed every test green. Only the structural `missingkey=zero` defect
remains genuinely open. This means fixloop's planned dispatch to case 004 may
no longer be its best next move — flagged in `PLAN_concept_register.md`'s new
"Coordination checkpoint" section and `RUNBOOK_concept_register.md` B6, not
decided unilaterally (that's fixloop's dispatch call, not this workstream's).

**Stage-3 pilot-seat recommendation:** computed a source-citation count across
all 1,631 concepts as a rediscovery-frequency proxy (per FIX-036's own
suggested method). `tool-lifecycle.md` is the single most rediscovered
category (5 of the top ~30 cited concepts) — candidate A, "reuse-agent," with
`DEV-001` (reuse-before-create) as its ready-made charter. The silent-content-
loss family found above is candidate B, "bug-historian" — more concretely tied
to fixloop's actual current work. Wrote both up in `PLAN_concept_register.md`
§Stage 3, recommending B first but leaving the choice to the user.

Updated all three running docs, `README.md` (concept count 1,627→1,631,
directory layout), and `register/000_concept_index.md`'s intro (new
"2026-07-16 addition" note). Did not touch the live fixloop workflow or make
any dispatch decision — both remain explicitly the owner's call.

## Turn 13 — 2026-07-16 — Read-aloud summary written; bug-historian pilot fully spec'd

User asked for (1) a summary document they could read out, and (2) to go
ahead with either the reuse-agent or bug-historian pilot, whichever I
preferred.

Wrote `SUMMARY_where_we_are_2026-07-16.md` — plain-language, narrative,
matching this repo's own convention (fixloop keeps a "calm, read-aloud"
companion to its technical journey doc; did the same here) — a
where-we-came-from / where-we-are / where-we're-going arc, no code citations,
readable aloud in a couple of minutes.

Picked **bug-historian** over reuse-agent, consistent with the earlier
recommendation: the recurring failure family (7 independent occurrences) is
durable evidence even though the specific triggering incident (the
image-landing trap) has now fully closed — a historian's whole value is
catching the *next* occurrence, and this incident is a genuinely excellent
worked example to seed it with.

Before designing the patch, read the actual live workflow file
(`fixloop_eg_dartsonline/0NN_fix_proposer.sql`) end to end, plus
`platform/orchestration/actions/diagnose_council_decide_action.go` — this
surfaced a precise mechanical fact not spelled out in the register: **any**
reviewer's `veto` verdict triggers "rejected" identically, regardless of
`hard_veto_from` (which only changes the audit label, not the outcome). This
directly shaped the design: the new seat's prompt offers only
`approve|object`, never `veto`, so it can influence a revise round but never
single-handedly reject a plan — genuinely advisory, not a second gatekeeper
by accident.

Wrote `PILOT_bug_historian_reviewer.md`: charter, the exact 7-concept curated
context digest, a full prompt template matching the existing two reviewers'
style/contract precisely, and the exact 4-edit patch (new step + 3 config
changes) to slot it into the sequential
`review_editquality → review_bug_historian → review_guardian → council_decide`
chain. **Did not touch the live SQL file or apply anything to a database** —
per the established (and fixloop-corroborated) boundary, that step waits for
explicit sign-off, which the document's §6 asks for directly.

Updated `PLAN_concept_register.md` (§Stage 3, pilot-seat section now points to
the completed spec) and `RUNBOOK_concept_register.md` (B5 updated to reflect
candidate B is spec'd and awaiting an apply decision).

## Turn 14 — 2026-07-16 — Bug-historian wired into production, with a permission gate along the way

User: "please go ahead and wire the bug historian into the workflow."

Assembled `0NN_fix_proposer_v6_bug_historian.sql` programmatically (Python
string substitution against the real, freshly-read `0NN_fix_proposer.sql`,
rather than hand-retyping ~500 lines) — five precise edits: re-point
`review_editquality`'s `next_step`, insert the new step, extend
`council_decide`'s `review_fields`, extend `repropose`'s `input_fields` and
prompt, and extend `escalate`'s `review_fields` too — the last one a gap I
caught only on a final grep sweep (the human hand-off package would otherwise
have silently omitted the third reviewer's opinion in an escalated run).
Verified syntactic validity with a proper comment-aware SQL string-literal
tokenizer (respecting `''`-escaped quotes and `--` line comments) before going
further — confirmed both the original file and my new one parse cleanly.

**Hit a permission gate applying it:** the auto-mode safety classifier blocked
even a read-only `kubectl exec` query into `postgres-clients-0` with the
reason that my instructions never specifically named that production
database/host as the target. Did not attempt to route around it — reported
back and asked the user to confirm the specific target by name, which they
did verbatim.

Applied cleanly: pre-flight read confirmed one live `fix-proposer` row
(version 1); ran the migration; `snapshot_agent` captured the pre-update row
(id `f9d90a2d-...`, confirmed present in `agent_definitions_backup`,
timestamp matching the pre-flight read exactly); transaction committed with
no errors. Post-apply verification queries confirmed: the new step exists and
is correctly wired in the sequential chain, both `review_fields` arrays
(`council_decide` and `escalate`) carry all three reviewers in order, and the
prompt content is intact (3,986 chars, correct opening text).

Discovered mid-task that a separate concurrent session had already swept my
earlier register edits (FIX-051/052/053, STY-049, PUB-001, etc.) into one of
its own commits (`6880c669e`) — exactly the risk this repo's `CLAUDE.md`
warns about. Per its own guidance ("if your work does get swept into
someone's commit: nothing is lost, forward-only still holds"), took no
corrective action — just committed the remainder of my own outstanding work
narrowly on top, following `CLAUDE.md`'s exact commit-per-task convention
(explicit pathspec on `commit`, new files `add`ed first, no bulk `add`,
excluding three other files mid-edit by a different session in the same
directory): one commit for the docs026 documentation (`27b5e5f2f`), one for
the production SQL migration itself (`187a1208e`).

Updated `PILOT_bug_historian_reviewer.md`, `PLAN_concept_register.md`, and
`RUNBOOK_concept_register.md` to mark the seat LIVE rather than merely
designed, noting what's still unverified (a real fix-loop run exercising it
end to end, not just a DB-state check).

<!-- Append new turns below this line. Format: ## Turn N — date — one-line summary -->
