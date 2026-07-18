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
| 2026-07-17 | Added `MDL-038`/`039` (BUG A/B, found by fixloop's first real-case run); confirmed independently that `fix-proposer` is unaffected by BUG B (no root `ai_service` key) | **confirmed** (both independently verified against live source, not just docs) |
| 2026-07-17 | User asked for the next 10 council-member candidates; delivered a ranked list grounded in a fresh rediscovery-frequency scan + FIX-036's named roster | **confirmed** (list delivered, none built yet at time of asking) |
| 2026-07-17 | User: "yes, reuse agent, then in the order you suggest." Reuse-agent built and APPLIED (council now 4 reviewers); its original grounding (tool-lifecycle.md) corrected to the real charter (DEV-001) while building it | **confirmed, live** |
| 2026-07-17 | Discovered a concurrent "council gate" thread building a service to run ALL platform commits through this same council — explicitly named this workstream as its seat-roster dependency | **noted, not actioned** (informational, directly relevant to pacing the remaining 9 seats) |
| 2026-07-17 | Surfaced a scaling concern (10 more always-on sequential reviewers = 14 LLM calls/decision) before building seat #3 onward | **flagged, awaiting user direction** |
| 2026-07-17 | User: "do the guidelines member then we can look at the relevance filtering mechanism." Guidelines-agent (seat #3) built + APPLIED — council now 5 reviewers | **confirmed, live** (5th and last always-on seat) |
| 2026-07-17 | Designed the relevance-filter (`DESIGN_relevance_filter.md`); found it needs a chassis-image Go change (council_decide hard-fails on absent seats), not pure SQL | **designed, not built** (build decision is the user's — a bigger, image-requiring change) |
| 2026-07-17 | Council-gate thread's review caught a real gap: advisory seats' `checks[]` were solicited but never run (`check_fields` omitted them). Fixed via v9, applied+verified | **confirmed, live** (latent defect in my own seat adds) |
| 2026-07-17 | Started lockstep-syncing the gate-file clone's roster, but backed off — the gate thread is actively editing it and syncing it themselves; hands off their file | **not actioned** (correct coordination call — avoided a collision) |
| 2026-07-17 | User: "the relevance filter can be next then the specialist seats." Built the filter's Go engine (`select_review_panel` + council_decide abstention), tested, committed `37468ba65` | **engine built + committed (inert)**; deploy is the gated step |
| 2026-07-17 | Verified the filter genuinely needs the Go change (conditional can't pattern-match arrays; council_decide hard-fails on absent) — pure-SQL not viable without fragile workarounds | **confirmed by reading the action code directly** |
| 2026-07-17 | Held the chassis DEPLOY (fleet-wide, shared Go with the active council-gate thread) rather than shipping unilaterally — recommend sequencing with that thread | **flagged, awaiting user/coordination** |
| 2026-07-17 | User chose option (b): another thread leads the deploy. My Go rides their next chassis image (in HEAD, inert+backward-compat, safe to ride unknowingly) | **confirmed** — standing task: apply filter wiring after the image is pod-verified live (RUNBOOK B10) |
| 2026-07-17 | User: "please do candidate #10 (documentation/contextkit specialist) next." Built + APPLIED as "tooling & provenance" seat (v10) — council now 6 reviewers | **confirmed, live**; footprint added to filter config so it auto-gates on deploy |
| 2026-07-17 | Flagged the sequencing tension: #10 applied always-on though specialists were meant to gate behind the (not-yet-deployed) filter — deliberate negligible-cost interim, converges to gated on deploy | **flagged transparently** — no more always-on specialists past this without the filter |
| 2026-07-17 | Discovered the filter Go already shipped in v1.0.1133 (running pod has SelectReviewPanelAction + the abstention) — the binary dependency was already met, no need for v1.0.1134 | **confirmed via pod grep** |
| 2026-07-17 | RELEVANCE FILTER ACTIVATED (v11 wiring applied) — 4 specialists now gate on relevance, editquality+guardian always-on; the sequencing tension is resolved | **confirmed, LIVE**; not yet exercised on a real run |
| 2026-07-17 | Owner request: add a "prefer not to change long-working core (orchestrator/kafka/messaging)" proviso. Home = the guardian (blast-radius/architecture seat, always-on, hard veto). Applied as clause (d) | **confirmed, LIVE** (surgical patch) |
| 2026-07-17 | COORDINATION LESSON: my seat migrations reconstruct the WHOLE default_config, which can clobber another thread's edits to shared steps (the guardian gained `code_checks` from another thread). Used a surgical jsonb_set for the proviso to avoid it | **noted** — use surgical patches for shared steps, not full-config reapply |
| 2026-07-17 | Built candidate #3 (adoption-pipeline guardian, `ADO-006`) — the FIRST seat gated behind the filter, and the first added SURGICALLY (chained jsonb_set); council now 7 | **confirmed, live** — proviso + code_checks + filter all verified intact |
| 2026-07-18 | Built candidate #4 (diagnosis-loop guardian, `DIAG-001/008/009/028/030`) — gated + surgical; council now 8. Drift-checked first (another thread had touched the row at 09:21; all anchors intact) | **confirmed, live** |
| 2026-07-18 | Built candidate #5 (improvement-loop guardian, `IMP-003/004/027` — the 845-item unbounded-drain guards) — gated + surgical (v14); council now 9 | **confirmed, live** |
| 2026-07-18 | Owner's multi-LLM idea (diverse-vendor model panel for intractable bugs) ruled a SEPARATE subproject, not the debugging seat — council seats review plans; this generates diagnoses at the escalation terminal. Recorded in PLAN with groundwork + prerequisite (MDL-038) | **recorded, awaiting green-light** |

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

## Turn 15 — 2026-07-17 — Checked in on the bug-historian; found two new platform bugs to register

User: "please carry on." Given fixloop's memory entry now said "FIRST
REAL-CASE CONFIRMED" — directly relevant to the bug-historian's own "not yet
exercised" open item — checked in on it.

Read `fixloop_eg_dartsonline/NOTES_running_fixloop(10).md` Turn 34 (filesystem
only, no DB access needed): the loop delivered its first real-case CONFIRMED
diagnosis (`MDL-038`, "BUG A" — `GenerateText` never decodes `stop_reason`, so
a max_tokens-truncated LLM response looks like a complete success at every
layer above; CONFIRMED on 3 citations including 17 live `llm_call_log` rows
showing the signature at scale). But **the fix-loop's fix-proposer step
hasn't actually run yet** — "fix dispatch for BUG A awaits owner go" per
fixloop's own notes — so the bug-historian's real exercise is still pending,
not yet resolved.

The same session also surfaced a second bug (`MDL-039`, "BUG B"): an agent's
root-level `ai_service` config silently shadows its step-level config —
proven by a direct experiment on `diagnose-agent`, 17-agent fleet blast
radius, and directly relevant to my own bug-historian migration since it also
uses step-level `ai_service` blocks. Verified by reading
`platform/orchestration/actions/ai_actions.go`'s `ExecuteLLMPromptAction`
directly (lines ~147-193): it checks a top-level `default_config.ai_service`
key first, only falling back to the step's own config if absent. Re-read
`fix-proposer`'s `default_config` structure (from my own migration file,
which only ever added content inside the existing `workflow` key) and
confirmed it has no top-level `ai_service` key at all — **the bug-historian
is not among the 17 affected agents.** Also independently confirmed BUG A's
code claim directly: `platform/aiservice/anthropic.go:67`'s `GenerateText`
response struct (lines 158-167) declares only `Content`/`Usage`, no
`stop_reason` field, exactly as described.

Added both as new register concepts — `MDL-038` and `MDL-039` in
`model-infrastructure.md` — following the same incremental-extraction pattern
as the 2026-07-16 fixloop-subsystem additions: genuinely new material,
independently verified against live source (not just taking the fixloop
notes' word for it), not a re-sweep. Register now **1,633 concepts**.
Updated `README.md`, `register/000_concept_index.md`'s intro (new "2026-07-17
addition" note), `PLAN_concept_register.md` (stage-3 status + backlog), and
`RUNBOOK_concept_register.md` (B5) to reflect the current state: bug-historian
live but still unexercised, and why that's not a problem (BUG A's dispatch is
the owner's call, same as it's always been).

## Turn 16 — 2026-07-17 — Ten more candidates, reuse-agent built and applied, a scaling concern surfaced

User: "please list the next 10 council members you have in mind." Recomputed
the rediscovery-frequency scan fresh against the current 1,633-concept
register (the earlier one was against 1,627, now stale), cross-referenced
against FIX-036's named roster (guidelines agent, compliance/legal eye,
pipeline-guardians per master workflow, specialist knowledge agents — all
still unbuilt) and the platform's actual master workflows. Delivered a ranked
10: reuse-agent, guidelines agent, adoption-pipeline guardian, diagnosis-loop
guardian, improvement-loop guardian, compliance/legal eye, render-pipeline
guardian, LLM-reliability specialist, debugging/incident-lore historian,
documentation/contextkit specialist — each grounded in specific register
concepts, not just a category name.

Mid-turn, user volunteered: "the bug-historian is currently being tested in
a diagnosis loop" — relevant since it directly concerns the "not yet
exercised" open item. Noted, but stayed focused on the requested list; a
filesystem check (no DB access) turned up a flurry of new fixloop activity
including files literally named `0NN_council_gate.sql` and
`HANDOFF_2026-07-17_council_gate_thread.md` — read both. **A separate
concurrent thread is building a service that decouples the review council
from fix-proposer specifically so any thread's diff (not just fix-loop's own)
can run through it, eventually gating all platform commits via PR-mode.**
Its own design doc states explicitly: "seats added via concept register
stage 3 immediately serve BOTH the fix loop and the gate... no competing
design," and its handoff lists as one of its own open owner-decisions:
"Seat roster for the gate: the 3 live seats, or wait for more concept-register
stage-3 seats?" — this workstream's output is literally that thread's
dependency.

User: "yes, reuse agent. then in the order you suggest." Building it surfaced
a correction worth making honestly rather than quietly: the reuse-agent's
originally-cited grounding (`tool-lifecycle.md`'s citation density) turned out,
on closer reading of the actual concepts, to be about a different theme
(tool-clobber protection, already in the bug-historian's curated context) —
the real charter is `DEV-001` (development-guide.md) plus FIX-036's own
founding incident (a reinvented trigger+triage SQL pair). Corrected before
building rather than building the wrong thing to match a stale citation.

Wrote `PILOT_reuse_agent_reviewer.md` and the v7 SQL patch (same 5-edit
pattern as v6: re-point the chain, insert the step, extend both
`review_fields` arrays, extend `repropose`). Verified syntax with the same
comment-aware tokenizer as before. Before applying, checked for in-flight
fix-proposer/council activity given the user's live-test comment — found
none (zero orchestrations have ever reached a review step, consistent with
BUG A's fix dispatch still awaiting the owner's go). Also noticed
`fix-proposer`'s `updated_at` had changed since the v6 application without
the step structure changing — verified the content matched exactly before
overwriting, rather than assuming. Applied v7; verified live: 4-reviewer
chain wired correctly, both `review_fields` arrays extended, prompt content
intact (2,906 chars).

**Surfaced, not silently ploughed through:** building the remaining 9
candidates the same way (more always-on sequential steps) would mean 14
sequential LLM calls per council decision — a real latency/cost concern for
a council about to gate all platform commits, not a reason to stop but a
reason to ask before continuing at that scale. Documented three options in
`PLAN_concept_register.md` and `RUNBOOK_concept_register.md` (B7/B8); pausing
on seat #3 to ask directly rather than assume "in the order you suggest"
meant "all 10, always-on, no further check-in."

## Turn 17 — 2026-07-17 — Guidelines-agent built + applied; relevance-filter designed (needs a chassis change)

User: "Please do the guidelines member then we can look at the relevance
filtering mechanism."

**Guidelines-agent (seat #3), applied — council now 5 sequential reviewers.**
Grounded it properly from the actual register content rather than a category
label: FIX-036's phrasing is "a guidelines agent (adherence to 000-0xx, or did
the guideline fall short)," and both clauses turned out load-bearing. Pulled
`DEV-005` (wrapper-orchestrator), `DEV-027` (dedup mechanics/`idx_swi_dedup`),
`DEV-018` (truthful provenance), `CTS-037` (declared input/output contracts),
`CTS-002` (schema-source tiers) as the "rules to adhere to," plus
FIX-016/FIX-042 for the distinctive second lens: a *guideline-gap* (the rule
itself is wrong, not the fix) should lean **side-task, not block**. Encoded
that within the fixed output contract cleanly — a violation → `object` (routes
to revise), but a guideline-gap → `approve` + a `notes` entry, never an
objection, so a correct fix isn't forced to revise for exposing a bad rule.
Fresh live example that makes this concrete: `MDL-039` (BUG B) proved a runbook
`max_tokens` rule was literally backwards. Built the v8 patch (same 5-edit
method as v6/v7), verified syntax, pre-flight checked no in-flight council
activity + confirmed live state was v7, applied, verified live (5-reviewer
chain, both `review_fields` arrays, prompt intact 2,774 chars).

**Relevance-filter designed (`DESIGN_relevance_filter.md`).** Per the user's
"then look at" — designed it fully rather than built it, and the design pass
surfaced the load-bearing constraint: it can't be a pure-SQL change like the
seat adds. Read `diagnose_council_decide_action.go` directly and confirmed it
hard-fails on any absent reviewer field (line 100-102: `if raw == nil {
return ... "reviewer output missing" }`). So skipping a seat when it's not
relevant requires a small Go change (treat absent as abstention) plus a
`select_review_panel` action — a chassis image build, a bigger class of change
than the SQL-only seat adds, and one now shared with the council-gate thread.
Designed the whole thing: the relevance signal (fix plan's `edits[]` files +
operations, plus diagnosis-cited entities), each seat's relevance footprint
(derived from grounding concepts' `verify-later` fields — the exact join the
stage-3 design always called for), the three build pieces, and which seats
stay always-on (editquality, guardian) vs gate behind the filter (everything
else, including retrofitting the 3 current advisory seats). Presented the
build decision rather than proceeding — it's image-requiring and cross-thread.

Council seat count: 2 (original) → 5 (three stage-3 seats added). Remaining 7
candidates from the "ten more" list are the specialists the filter is meant to
gate.

**A coordinating thread reviewed my work and caught two real gaps** (an
addendum appended to `PILOT_reuse_agent_reviewer.md` by the council-gate /
feature-builder thread — the exact kind of cross-check the whole council idea
is for, applied to the council itself):
1. **`run_checks.check_fields` omitted the advisory seats** — it listed only
   editquality + guardian, so bug-historian/reuse-agent/guidelines could
   request read-only SQL checks in their output that were solicited but NEVER
   executed or fed back on a revise round. Verified against the live DB (true)
   and against `diagnose_run_checks_action.go` (it explicitly tolerates
   absent/empty check lists, so the fix is safe). **Fixed: v9 migration
   (`0NN_fix_proposer_v9_runchecks_fix.sql`)** extends check_fields to all 5
   reviewers; applied to `clients_db`, verified live. A genuine latent defect
   in my own seat additions, caught by their review — worth the flag.
2. **Two council definitions can drift** (fix-proposer's live one, and the
   gate's file-only clone). I started to sync the gate file to the 5-seat
   roster in lockstep (the addendum explicitly invited it) — but the sync
   attempt's exact-match assertion failed because **the file changed under me
   mid-edit**: the gate/feature-builder thread is actively editing it right now
   and had already added the reuse_agent + guidelines steps themselves (their
   `review_fields`/`check_fields` were still mid-update, not yet referencing
   them). My script failed cleanly *before writing* — no corruption. Correct
   call: **hands off their file entirely** — they own it and are doing the
   sync themselves. Left it untouched, not included in my commit. (Good
   argument for building the relevance-filter's `select_review_panel` as
   shared code both councils import, rather than two hand-maintained rosters —
   noted in `DESIGN_relevance_filter.md`.)

Committed my own work narrowly per `CLAUDE.md`: the v8 + v9 migrations and the
docs, excluding the gate file (the other thread's, mid-edit) and
`PILOT_reuse_agent_reviewer.md` (carries the other thread's uncommitted
addendum — theirs to commit).

## Cross-thread note — 2026-07-17 evening — from the council-gate thread

*Left by the "fixloop council on every bugfix" thread, per the coordination
convention; not one of this workstream's own turns.*

Your v7/v8 seat applications landed while the council-gate service was being
built. Three things now on the record for this workstream: (1) the gate seed
(`fixloop_eg_dartsonline/0NN_council_gate.sql`, apply-ready, NOT applied) is
**synced to the v8 five-seat roster** — from v9 onward, any seat migration
must patch both council definitions in one migration (details in
`PILOT_reuse_agent_reviewer.md` §6, an attributed addendum appended there).
(2) A live v8 gap flagged for this thread's surface: `run_checks.check_fields`
still lists only editquality + guardian, so the three advisory seats'
solicited checks are never executed on a revise round — inherited from v6,
one-line config migration. (3) The owner's gate rulings (2026-07-17): scope
platform/+internal/+pkg/, advisory launch, credits per submission, more
seats before launch — your seats may have satisfied that ruling; the open
owner question is now "launch on 5, or wait for the relevance filter?",
which bears directly on your §Stage-3 scaling options (b)/(c).

## Turn 18 — 2026-07-17 — Built the relevance-filter engine (Go), held the fleet deploy

User: "I think the relevance filter can be next then the specialist seats."

**Verified the mechanics before choosing an approach** — my design had assumed
the filter needs a chassis Go change; I confirmed it by reading the actual
action code, and also checked whether a lower-risk pure-SQL path existed:
- `plan_persisted.files` is a `[]string` of edited file paths, pre-extracted by
  the persist step — the clean primary signal.
- The `conditional` action supports `field == true`/AND/OR and routes via
  `next_step_override`, but its array `contains` is exact-membership only — it
  **can't** pattern-match file paths. So the filter genuinely needs a compute
  step to turn paths into per-seat booleans. Pure-SQL regex-over-plan-text +
  abstain-stubs is *possible* but fragile (uncertain jsonb/output-shaping edge
  cases, tested only in prod) — rejected in favour of doing it right.
- `diagnose_council_decide` hard-failed on any absent reviewer field — the
  reason skipping a seat needs a Go change.

**Checked for duplicate work first (reuse-before-create, literally the seat I
built):** grepped the council-gate/feature-builder thread's docs — they are
NOT building relevance filtering (only a roster question). So no collision on
the panel code.

**Built the engine, tested, committed (`37468ba65`):**
- `select_review_panel_action.go` — deliberately GENERIC and config-driven: the
  seat→footprint patterns live in the workflow SQL config, not Go, so the same
  binary serves both councils and neither thread's roster choices are baked
  into Go (this also minimises the shared-Go footprint to a plain substring
  matcher — hard for the gate thread to want differently). Matches
  `plan_persisted.files` + optional diagnosis text; fail-open on empty
  footprints (never silently drops a seat).
- `diagnose_council_decide` — absent reviewer field = abstention, not error;
  fails closed only if ALL abstain. Backward-compatible (nothing is absent
  until skips are wired).
- Registered the action; `go build` + `go vet` + `go test` all green; wrote
  unit tests (footprint matching, corpus fallback, fail-open, `[]string`
  coercion) that pass. gofmt clean.

**Committed the Go as INERT** — nothing calls it until the SQL wiring adds the
step, and the abstention can't trigger without skips, so today's live behaviour
is unchanged. Committing (vs leaving dirty) is safer per `CLAUDE.md` (a
long-lived dirty tree is shared mutable state that another session can sweep).

**Held the DEPLOY, deliberately.** The remaining step to make it live is a
chassis image build + rollout — and that's a genuinely different class of
change from the SQL-only seat adds: fleet-wide (every agent's binary), and in
Go shared with the actively-developed council-gate/feature-builder thread.
Ideally one `select_review_panel` binary serves both councils, so the deploy
should be sequenced with that thread, not shipped as a fix-proposer-only image.
Fully specified the `v10` SQL wiring (footprint config + per-seat gates,
retrofitting the 3 advisory seats as the proof-of-concept) in
`DESIGN_relevance_filter.md` §7 so it's ready to apply the moment the image
ships. Presented the deploy as the decision (RUNBOOK B7) rather than rolling a
fleet-wide binary unilaterally.

## Turn 19 — 2026-07-17 — Deploy decision: another thread leads it; recorded the standing follow-up

User (after "read CLAUDE.md and follow it"): chose deploy option (b) — another
thread leads the chassis deploy.

Re-read `CLAUDE.md` in full and complied: committed each task narrowly (Go
engine, docs, deploy-note correction — all separate commits with explicit
pathspecs), left a clean tree (only the gate thread's addendum uncommitted,
theirs), and **corrected a factual error** the re-read caught — my deploy notes
had cited `make build-<service>-ref` + a WIP-bundling risk, but per the
2026-07-17 build inversion the default `make build-<service>` builds from
committed HEAD and can't bundle WIP; fixed in the design doc and RUNBOOK.

On the decision: option (b) is clean for this workstream. My relevance-filter
Go (`37468ba65`) is in HEAD, so it rides the next chassis image the deploy
thread builds — automatically, and **safely even if they don't know it's
there**: `select_review_panel` is registered but uncalled (inert), and the
`council_decide` abstention is backward-compatible (no field is ever absent
until skips are wired). So it can't break their release. No coordination action
needed on the image itself beyond what `git log` already shows (the commit
message documents exactly what's in HEAD and that it's inert-pending-wiring).

Recorded the one standing follow-up (RUNBOOK B10): the filter stays dormant
until (1) the image ships — not mine to drive — and (2) the v10 SQL wiring is
applied. Deliberately did NOT pre-write the v10 migration: the council roster
is actively evolving (the gate/feature-builder threads are mirroring/adding
seats), so a draft made now would go stale — it should be written against the
then-current `fix-proposer` definition once the image is pod-verified live.
`DESIGN_relevance_filter.md` §7 fully specifies it for that moment.

Net state of stage 3: 5 council seats live; the relevance-filter engine built +
committed + riding the next deploy; the remaining 7 specialist seats build
behind the filter once it's live. Nothing blocked on me.

## Turn 20 — 2026-07-17 — Built candidate #10 (tooling & provenance seat); council now 6

User: "please can you do #10 (documentation/contextkit specialist) next."

**Grounded it properly** in the two concepts behind candidate #10: `CTXK-015`
(the single most-rediscovered concept in the register — 11 sources — the
`cmd/bundle`/contextkit investigation lore + the "resolve an action from the
registry, never by filename convention" trap) and `DOC-010` (travelling docs:
every tool/pipeline carries a living PLAN + NOTES in `doc_plans`/`doc_notes`;
notably the fix-loop *itself* adopted this rather than build a rival — a live
endorsement of the discipline this seat enforces). Named the seat "tooling &
provenance": does the fix use the platform's own investigation + documentation
machinery, or reinvent/work around it? Distinct from the reuse-agent (which has
no travelling-docs lens).

**Surfaced a real sequencing tension rather than glossing it:** this is a
specialist seat, and I'd just built the relevance filter *specifically* so
specialists don't run always-on. But the filter isn't deployed (another thread
leads that). Applying #10 always-on now mildly cuts against that plan. Resolved
it transparently: applied it always-on as a deliberate, **negligible-cost
interim** (the council isn't running on real cases yet — BUG A's dispatch
awaits the owner — so a 6th always-on seat costs ~nothing today), AND added its
footprint to the filter config (`DESIGN_relevance_filter.md` §7) so it
**auto-gates the moment the filter deploys**. Its narrowness (most fixes touch
no tooling) makes it the clearest illustration of why the filter exists — noted
so this isn't read as abandoning the gated design. Flagged that adding *more*
always-on specialists past this should wait for the filter.

Built + applied v10 (same proven 5-edit pattern + the v9 check_fields rule, so
its checks run too), pre-flight-checked no active runs, verified live: council
now 6 reviewers, `run_checks` covers all 6, prompt intact (2,801 chars). Prior
row snapshotted (rollback available).

Council seat count: 2 (original) → 6. Remaining from the "ten more" list: #3
adoption, #4 diagnosis-loop, #5 improvement-loop, #6 compliance, #7 render, #8
LLM-reliability, #9 debugging — best built behind the filter once it's live.

## Turn 21 — 2026-07-17 — Relevance filter ACTIVATED (the binary was already live)

User: "I have just started the next build while everything is quiet v.1.0.1134"
then "the relevance filter may be in this one or may already have been committed."

First verified my filter Go (`37468ba65`) is an ancestor of HEAD — so it ships
in v1.0.1134's committed-HEAD build. Then checked the **running pod** (per
CLAUDE.md, verify against the pod not git) and found the key fact: the filter is
**already live in v1.0.1133** (the pod running since 18:45Z). `SelectReviewPanelAction`
= 2, `reviewer abstained` = 1, `all %d configured reviewers abstained` = 1 — both
halves of the filter (the action AND the council_decide abstention, one commit)
present in the running binary. So it shipped in a *prior* build from my committed
HEAD; it never needed v1.0.1134. The image-before-seeds precondition was already
met.

So the standing task (RUNBOOK B10 — apply the wiring once the image is live) was
unblocked *now*, in the owner's deliberately-quiet window. Wrote the `v11` wiring
migration against the live 6-seat definition (re-verified no drift, not
already-wired), pre-flight-checked no active runs, applied, verified: the full
gated chain is live — `persist_plan → select_panel → editquality →
gate→bug_historian → gate→reuse_agent → gate→guidelines → gate→tooling_provenance
→ guardian → council_decide`, footprints configured for all 4 gated specialists.

**This resolves the scaling concern in production.** The council no longer runs
all 6 reviewers on every decision — edit-quality + guardian always run, and each
of the 4 specialists runs only when the fix's edited files/diagnosis match its
footprint (a deterministic, no-LLM `select_review_panel` step decides). A skipped
seat's absent review field is tolerated as an abstention by the (already-live)
council_decide change. Safe degraded mode: if `select_panel` ever fails to run,
the gates default to skip, so the core council (edit-quality + guardian, incl.
the hard veto) still runs.

Not yet exercised on a real fix-proposer run (BUG A dispatch still awaits the
owner) — the first real run exercises the filter end to end. The unit tests
(footprint matching, corpus fallback, fail-open) already cover the action's
logic.

**Stage-3 milestone:** 6 council seats live + the relevance filter live gating
4 of them. The remaining specialists (#3 adoption, #4 diagnosis-loop, #5
improvement-loop, #6 compliance, #7 render, #8 LLM-reliability, #9 debugging) now
build cheaply behind the filter — each is a footprint entry + one gate, not
another always-on step. Committed narrowly (v11 migration + docs) per CLAUDE.md.

## Turn 22 — 2026-07-17 — Guardian stability proviso; and a coordination bug caught in time

User: "In which agent can I add the proviso that as a preference we don't want
to change code that has been working for ages e.g. orchestrator, kafka,
messaging etc" → then "ok please go ahead carefully."

**Answer: the guardian** — it's already the platform-safety/blast-radius seat
(judges "which pipelines consume each edited file" + "architecture-change
signals — edits to shared contracts, wire formats, message shapes"), it's
always-on (sees every fix), and it holds the hard veto. The proviso is a natural
fourth judgement clause. Added as `(d) STABILITY PREFERENCE` — prefer a fix at a
higher, less-foundational layer over editing long-stable core (orchestrator,
Kafka/messaging, agent spawning, work-item dispatch); object/steer to a higher
layer, veto reserved for a genuine architecture change (its existing behaviour,
not a new blanket veto).

**Caught a real coordination bug before applying:** reading the LIVE guardian
prompt to place the proviso, I noticed it had a `code_checks` mechanism that
is NOT in any of my seat-migration files. Another thread had enhanced the
guardian on the live definition. **My seat migrations (v6-v11) reconstruct the
ENTIRE `default_config` and `SET default_config = EXCLUDED`** — so a full-config
reapply would have *clobbered* the other thread's `code_checks`. (Checked: my
v11 apply had NOT clobbered the current state — the other thread re-applied
their code_checks AFTER my v11, at 19:45:45, and preserved my filter wiring; so
in the end both coexist. But it was luck of ordering, not safety by design.)

**So applied the proviso SURGICALLY** — a `jsonb_set` + `replace()` on ONLY
`review_guardian.config.prompt_template`, anchored on a unique substring, with
an idempotency guard and a snapshot. Verified live: proviso added
(prompt 2685→3147 chars), and `code_checks` + the relevance-filter wiring +
all 6 seats **byte-intact**. Committed the patch (`PATCH_guardian_stability_proviso_2026-07-17.sql`).

**Standing lesson (also logged in the decision table):** for edits to steps
SHARED with other threads (the guardian, and any step they might touch),
use a surgical `jsonb_set` on the specific field — never a full-config reapply.
Future seat additions are lower-risk (they add NEW steps), but any change to an
EXISTING shared step must be surgical.

## Turn 23 — 2026-07-17 — Adoption-pipeline guardian: first seat gated + surgical

User: "yes, please go ahead" — read as: continue the roster, build the next
seat (#3 adoption). Now that the filter is live, this is the first seat built
**behind** it (gated), and the first added **surgically** given the co-editing
lesson from the guardian proviso.

**Grounding:** `ADO-006` — "adoption writes specs first, classifier consumes
under fidelity rules" — one of the two original stage-1 flagged rediscovered
concepts. The adoption pipeline has strict architectural contracts a fix could
break: write-then-relay (`apply_adoption_plan` writes specs + emits exactly one
`needs_domain_research`, never calls the classifier directly), adopted-specs-are-
ground-truth (classifier reads-and-extends, never overwrites), no-bypass (adopted
sites still run strategist→briefing→planner), and LLM-for-reasoning-Go-for-
extraction. Those became the seat's four judgement points.

**Careful execution given co-editing:** re-checked the live definition for
drift first (proviso + code_checks both present, 4 gated seats, no surprises).
Then built it as a GATED seat (footprint in `select_panel` + a `gate_adoption`
conditional + the reviewer) via **chained `jsonb_set`** on the live config —
NOT the v6-v11 full-config reapply, which would have clobbered the guardian's
`code_checks`/proviso. Eight jsonb_set ops in one atomic UPDATE (atomic = no
race with the other thread). Idempotency guard + snapshot. Syntax-checked
(78/78 parens), pre-flight (no active runs), applied.

**Verified live:** council now 7 seats; the gated tail routes correctly
(`tooling_provenance → gate_adoption → [adoption_guardian?] → guardian →
council_decide`, both run and skip paths converging on `gate_adoption`);
`council_decide`/`escalate`/`run_checks` all at 7; the adoption footprint is in
`select_panel`; and — the whole point of surgical — the guardian proviso,
`code_checks`, the other 6 seats, and the filter wiring are all **byte-intact**.

**Standing pattern confirmed:** every future seat is now gated + surgical.
Council: 2 → 7 (5 gated specialists). Remaining: #4 diagnosis-loop, #5
improvement-loop, #6 compliance, #7 render, #8 LLM-reliability, #9 debugging.
Committed the migration + docs narrowly.

## Turn 24 — 2026-07-18 — Diagnosis-loop guardian (candidate #4); council now 8

User: "please go ahead" — next seat in the order: #4, the diagnosis-loop
guardian. Now the routine is settled: drift-check → ground → gated + surgical
→ pre-flight → apply → verify → document.

**Drift check first** (the workflow is co-edited): the live row had been
updated by another thread at 09:21 this morning — but all my anchors were
intact (7 seats, adoption present, guardian proviso + code_checks preserved,
chain tail correct). The surgical pattern makes this safe: my jsonb_set chain
only touches the paths it names.

**Grounding:** `diagnosis-loop.md` is the highest hot-concept-density category
in the register. The seat defends the loop's earned disciplines: read-only
cite-or-abstain (`DIAG-001`), the two-evidence-family/three-tier citation
standard (`DIAG-008` — proven live on BUG B, where it correctly withheld
CONFIRMED), the three-layer read-only SQL enforcement (`DIAG-009`),
observability-never-fails-a-diagnosis + skip-never-guess notes (`DIAG-028`),
the config-level `error_step` trap (`DIAG-030` — step-level is silently inert,
found dormant in other agents too), and token/pod isolation (`DIAG-019`/`022`).
There's a pleasing self-reference: the loop reviewing fixes to itself is
exactly the point — these guards exist because early benchmark runs produced
CONFIRMED verdicts a fixer must never have acted on.

**Applied v13** (8 chained jsonb_set, atomic, idempotent; snapshot taken;
85/85 parens; no active runs). Verified live: council now 8 (2 always-on + 6
gated), chain tail `gate_adoption → [adoption?] → gate_diagnosis →
[diagnosis?] → guardian` with run/skip paths converging correctly, all three
arrays at 8, diagnosis footprint present, proviso + code_checks byte-intact.

Council: 2 → 8. Remaining: #5 improvement-loop, #6 compliance, #7 render,
#8 LLM-reliability, #9 debugging.

## Turn 25 — 2026-07-18 — Improvement-loop guardian (#5); summary written; multi-model gauntlet ruled a separate subproject

User asked for three things: a read-aloud summary, an explanation of the
debugging seat, and a ruling on their idea — spin up several specialised LLMs
of different vendors/types/strengths for bugs intractable to the current loop
— with instruction to proceed with the stated list if it's a separate project.

**The ruling (the honest architectural distinction):** council seats REVIEW
fix plans that already exist — cheap advisory prompt steps in the review
chain. The multi-LLM idea GENERATES diagnoses that don't exist yet, for bugs
where the house loop gave up. It belongs at the diagnosis loop's escalation
terminal (exhausted / UNVERIFIABLE — both honest hand-to-human terminals
today), not in the council. So: **separate subproject.** Recorded it in
`PLAN_concept_register.md` as "multi-model diagnosis gauntlet" with the
register-verified groundwork (per-step model routing exists via `MDL-035`;
`platform/aiservice` has Anthropic+Ollama clients, other vendors need new
clients) and one prerequisite worth fixing first: `MDL-038`/BUG A —
an ensemble comparing outputs must know when an output was truncated.

**Seat #5 (improvement-loop guardian) built per the ruling** — the settled
gated + surgical routine: drift check (row touched by another thread at 09:31,
anchors intact), grounding (`IMP-003/004/027` — the loop's termination guards
exist because it once ran unbounded: 845+ findings in ~10 days), v14 migration
(8 chained jsonb_set, snapshot, idempotent), pre-flight (no active runs),
apply, verify (9 seats, chain tail correct, proviso + code_checks intact).
Council: 2 → 9 (7 gated specialists).

**Summary written** (`SUMMARY_where_we_are_2026-07-18.md`) — same read-aloud
style: what we set out to achieve, where we are (9 seats + the filter + how
each was installed), a plain-language explainer of the debugging seat vs. the
multi-model idea, and what's next (#6 compliance, #7 render, #8
LLM-reliability, #9 debugging; then the gauntlet if green-lit).

<!-- Append new turns below this line. Format: ## Turn N — date — one-line summary -->
