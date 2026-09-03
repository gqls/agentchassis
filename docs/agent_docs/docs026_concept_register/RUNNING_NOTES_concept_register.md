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
| 2026-07-18 | User: "please go ahead and complete the council." Built + applied v15-v18 (compliance, render, LLM-reliability, debugging) sequentially — THE COUNCIL IS COMPLETE: 13 reviewers, 11 gated | **confirmed, live** — full chain verified, proviso + code_checks intact |
| 2026-07-18 | THE COUNCIL'S FIRST REAL OUTING: BUG A (MDL-038) fix-proposer run 53da3a30 ran the full 13-seat chain, 3 revise rounds, → **APPROVED** (7 voted, 6 abstained). Filter woke bug_historian+reuse+diagnosis+llm_reliability+debugging. The two purpose-built seats (llm_reliability, debug_historian) gave grounded approvals + flagged the token-write-back ordering constraint. Revise loop IMPROVED the plan: final covers both anthropic.go AND ollama.go | **confirmed** — approved fix_plan persisted; implementer (opens PR) is the next, separate step |

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

## Turn 26 — 2026-07-18 — THE COUNCIL IS COMPLETE: 13 reviewers

User: "please go ahead and complete the council." Built the final four seats
in one sitting — the routine is settled enough that the migrations were
generated from a shared template (with per-seat charters individually
grounded), each verified, then applied sequentially with a pre-flight check
between every apply.

- **v15 compliance** (#6): the severity seat — two live fabrication incidents
  (vetcomparison prices, leopardess claims incl. the poisoned writing rule
  that INSTRUCTED fabrication). Judges unevidenced claims, weakened scanners,
  poisoned specs, legal surface (`LGL-001`, `CQ-017`).
- **v16 render-guardian** (#7): fail-loud-not-silent render paths, the two
  rerender modes' skip semantics (`STY-048` — chrome changes through scoped
  mode silently miss pages), runtime-fill exemption + rendered-artifacts
  landmine (`STY-019`), var() chain (`CTS-011`), validation layers (`STY-004`).
- **v17 LLM-reliability** (#8): BUG B root-shadows-step, BUG A
  truncation-looks-like-success (the output_tokens==max_tokens signature),
  thinking-spend budgets, swap discipline (`MDL-038/039/005/006`).
- **v18 debug-historian** (#9): the deliberately-broad seat, LOOSELY gated
  (.go/platform//internal//cmd//.sql — most code fixes), carrying the largest
  lore category: needle-gate SQL surgery + the Postgres pitfall catalogue
  (`DBG-016`), informational-column blast radius (`DBG-017`), pod-not-git
  verification, repair-vs-regenerate (`DBG-065`).

One generator bug caught before applying: charter lines for v17/v18 had
pre-escaped quotes that sql_quote() double-escaped (`''''`) — found by a grep
sweep, fixed, re-verified. All four applied cleanly (UPDATE 1 each, snapshots
taken); end-to-end chain verification shows all 26 steps routing correctly,
run/skip paths converging, all three arrays at 13, guardian proviso +
code_checks byte-intact, 11 footprints in select_panel.

**Final roster:** select_panel → editquality → 11 gated specialists →
guardian (hard veto, stability proviso) → council_decide. A typical fix wakes
2-5 seats. Every candidate from the "ten more" list is now built. What remains
for stage 3 is watching the council's first real outing (BUG A's dispatch,
owner's call) — and the proposed multi-model gauntlet subproject if
green-lit.

## Turn 27 — 2026-07-18 — The council's first real outing: BUG A approved

User: "please go ahead with the BUG A fix dispatch." Ran six premise checks
first (fixloop discipline — the bug may already be fixed): BUG A still absent
from HEAD and from the running pod's binary, no competing work items, the
CONFIRMED diagnosis + evidence bundles intact, pod past the dispatch-quiet
window. A concurrent session fired the run (53da3a30) at 10:16, ~1 min before
I would have — so I did NOT re-dispatch (would have double-run the
correlation) and monitored instead.

**Outcome: APPROVED after 3 revise rounds** (completed 10:37). The complete
13-seat council's first real case. The relevance filter woke 5 specialists
(bug_historian, reuse_agent, diagnosis, llm_reliability, debugging) + the 2
always-on; `council_decide` correctly recorded 6 abstentions for the skipped
seats. The two seats built THIS WEEK for exactly this bug class earned their
place: llm_reliability confirmed no ai_service config was touched (BUG B trap
irrelevant), that the guard sits after the token write-back so llm_call_log
still records output_tokens on the error path, and that no model swap was
implicated; debug_historian cleared all four lore dimensions and flagged the
same ordering constraint. Neither raised a blocking objection — both approved
with a review note.

**The revise loop demonstrably improved the fix.** Round 1 plan touched only
anthropic.go. During reproposal a code lookup found finish_reason absent from
the whole codebase → the identical silent-truncation gap in
OllamaClient.GenerateText. The FINAL approved plan covers BOTH providers:
add StopReason (json:stop_reason) + a max_tokens hard-error guard in
anthropic.go, and a parallel DoneReason (done_reason=='length') guard in
ollama.go. Grounded in 23 real silently-truncated llm_call_log rows
(output_tokens==max_tokens fingerprint, 5 agent types, 2 models).

**Pipeline state:** approved fix_plan persisted (diagnosis_artifacts, kind
fix_plan, correlation e505f70f). The NEXT stage — fix-implementer
(092 trigger, agent_type fix-implementer-orchestrator) — is a separate,
outward-facing dispatch: whole-file LLM implementation → allowlist → fix/*
branch → commit → gofmt+build gate in a k8s Job → (green) **PR into main**.
Held for owner decision since it opens a real PR.

## Turn 28 — 2026-07-18 — BUG A fix implemented, gate-blocked on gofmt, hand-finished onto 085

User: "the deploy system uses the current branch rather than main at the
moment, yes let's dispatch the implementer." Dispatched the fix-implementer
(092, agent_type fix-implementer-orchestrator) for the approved plan — run
70680566. Pre-checks all clean (approved decision latest, nothing in flight,
agent registered, pod past quiet window, no pre-existing branch).

**The orchestrator spawned a child implementer that did the real work.** It
generated logically-CORRECT whole-file implementations of both guards (matching
the approved plan exactly), pushed branch `fix/e505f70f` (based on
084_site_improvements_local_ai — the live base_branch config, not main), and
ran the gofmt+build gate in a k8s Job (build-gate-fix-e505f70f). **The gate
FAILED — correctly — on `gofmt FAILED for platform/aiservice/anthropic.go`**
and opened NO PR. The failure was purely cosmetic: the LLM added the new
`StopReason` field but didn't re-align the sibling `Usage struct` field, plus a
trailing blank line. ollama.go was gofmt-clean. The red-build path worked as
designed: branch + log preserved, nothing merged.

**Process gap worth noting (fixloop candidate):** the implementer does not run
`gofmt -w` on its own generated files before the build gate, so trivially
-unformatted LLM output burns a whole build cycle. Running gofmt in commit_prep
before the gate would fix this class.

**Hand-finished onto 085 (user chose the deploy branch).** Confirmed the two
aiservice files are byte-identical between the 084 base and current 085, so
applied the approved guards as SURGICAL edits (not whole-file replace) to the
working tree, gofmt -w, `go vet` + `go build` green. Committed narrowly as
`f32b208e5` with Diagnosed-Via + Council-Reviewed:53da3a30 trailers. anthropic:
decode stop_reason, hard-error on "max_tokens"; ollama: decode done_reason,
hard-error on "length". Both guards sit AFTER the token write-back (the
observability constraint both purpose-built seats flagged).

**State:** fix is on the deploy branch (085), gofmt-clean, vet+build green.
NOT yet built/deployed — chassis image build + roll is another thread's lead /
owner's call; the next `make build-agent-chassis` from HEAD will include it.
Loose ends: the stale `fix/e505f70f` remote branch (084-based, gofmt-broken, no
PR) is now superseded and can be deleted.

## Turn 29 — 2026-07-20 — Constitution + mission as always-on council seats (v19)

Owner's three directions (root-cause-not-workaround; gatekeep the constitution;
everything follows the mission) resolved to ONE gap after reading the artifacts:
the platform has a written constitution (`thin_slice_constitution.md`;
`CTS-029`/`DEV-054`, deployed but only pasted as passive bundle context) and a
written mission (`028_platform_mission_and_pipeline_direction`; `BIZ-001`,
status *partial*), both of which state in their own text that changes should be
checked against them — enforced by NO council seat (verified: grep of all 13
charters; only incidental "symptom"/"workaround" mentions). Point 1 was already
constitutional: *"Fix structural problems, not symptoms."*

**Built (owner chose: two always-on seats; fix-proposer + gate scope):**
- `review_constitution` + `review_mission`, ALWAYS-ON (no gate, no footprint) —
  gating would contradict what a constitution/mission are. Advisory
  (approve|object, no veto); an objection forces a revise round.
- constitution headline = root-cause / anti-workaround (object if a fix routes
  AROUND a known/listed bug instead of fixing the cause, absent an explicit
  justified deferral) + reuse-before-recreate + schema-first + parameterised +
  no-silent-rename + tone. Defers detailed contracts to the guidelines seat.
- mission = best-site-per-domain / revenue-shapes-the-site (no consultancy
  default) / classifier-is-strategic-brain / silent-override-is-the-failure-mode;
  bug fixes usually mission-neutral → approve.

**Method (mirror-safe, per CLAUDE.md):** read 099 first to confirm it copies
`review_*` verbatim + re-asserts `editquality.next_step` + refuses on dangling
targets. Matched live seat settings (claude-sonnet-5 @ 8000, temp 0.0,
input_fields [diagnosis_row, plan_persisted, schema_hint], error_step
complete_refused). v19 surgical migration (6 chained jsonb_set), snapshot, 0
active runs, applied (UPDATE 1). Chain: editquality → constitution → mission →
gate_bug_historian. Then `099 --apply` mirrored to council-gate (snapshot). Both
councils: **15 seats**, deep-compare drift NONE, routing OK; gate got the four
transforms (input_data / input_data.rationale / complete_invalid). Committed
`6a25f3607`.

**Flagged, not done:** (1) extend to feature-designer + experience-planner
councils and the build pipeline (bigger, separate); (2) protect the
constitution/mission DOCUMENTS from drift (owner-sign-off gate) — these seats
enforce conformance TO the docs, they don't guard the docs themselves.

## Turn 30 — 2026-07-20 — Rerender trap into render-guardian; prior-art librarian (v20, 16 seats); direction reach + drift-guard PLAN

Three owner asks this turn (two arrived mid-turn):

**1. Render trap (owner-named, applied+mirrored).** "page-rerender re-deploys
the existing HTML — it does not regenerate it from content_data." Added to the
render-guardian charter + judge clause (e), phrased in the CORRECTED
bugs_closed/031 mechanics. Method note that matters: the seat's live prompt
CHANGED UNDER THIS SESSION between my needle-count and my full dump (the 031
thread's own patch landed in that window) — so the patch re-asserts both anchors
inside the UPDATE's WHERE and no-ops if the text moved again. Also: judged a
text-only additive prompt patch safe with a run mid-chain (run was at
bug_historian, before the seat; no routing/schema change; 031 thread precedent)
— the strict 0-runs rule stays for STRUCTURAL changes.

**2. Prior-art librarian (owner suggestion, v20, applied+mirrored).** New
ALWAYS-ON seat: asserted-absence / dormant-machinery — no other seat verifies
the rationale's existence claims; all seats inherit a false premise (section-
editor case; 031's adjacent face). Mechanical lookups mapped to the LIVE check
tiers: code_checks symbol/content/ls (prior art), SQL checks (agent seeded?
ever run?), and an explicit CANNOT for pod-binary liveness (names the pod-grep
instead of guessing). Always-on per owner's lean — the tell is rationale
LANGUAGE; a path footprint misses it, a keyword gate is mushy. Both councils now
**16 seats** (5 always-on: editquality, constitution, mission, prior_art,
guardian; 11 gated), 099 deep-compare drift NONE. Chain: mission →
prior_art → gate_bug_historian. The inverted-lookup dormant-machinery sweep is
recorded in the plan as a complement, NOT built.

**3. The main ask — PLAN for build-pipeline reach + doc drift guard** —
`PLAN_2026-07-20_direction_reach_and_drift_guard.md`. Grounding done first:
NO standards table exists (CTS-029 unbuilt); classifier dispatch sites read from
Go; all constitution/mission copies hash-identical today; .githooks two-tier
precedent (check-secrets = real gate). Reach: R0 inject platform-mission digest
into classifier prompts → R1 observe-only mission-review step on classifier
output (detected + triage-sweep consumer — 023's lesson) → R2 enforcement only
on measured FP rate (owner-gated) → R3 fleet discovery check (Go+image) → R4
feature/experience councils. Guard: D1 bless canonicals + DIRECTION_LEDGER →
D2 real-gate hook requiring `Direction-Approved:` trailer (owner picks
gate/advisory) → D3 three-surface integrity check (files/copies/seat prompts) →
D4 standards-table migration as the structural close. Nothing applied from the
plan; commits 6a25f... family + this turn's three.

## Turn 31 — 2026-07-20 — D1-D3 + R0+R1 IMPLEMENTED (owner approved gate + staged reach)

Owner: "I agree with your lean on D2 and R2 please go ahead." Implemented the
plan's first tranche, grounding-first throughout — and the grounding changed the
design twice (both corrected VISIBLY in the plan):

- **Consumer correction (the big one):** the plan's R1 said findings go to
  site_work_items at 'detected', consumer = triage sweep. Reading
  `triage_detect_items_action.go:91-103` refuted it — the triager is site-scoped
  and TYPE-BLIND, promoting ALL detected items into dispatch → a mission_review
  item would be swept toward a nonexistent handler (a fresh instance of the 023
  class, inside the plan meant to avoid it). As built: objections append
  doc_notes (categories mission-review; the gate's own verdict machinery —
  reuse), consumer = NEW `101_REPORT_mission_review_findings.sh`.
- **Hook-type correction:** D2 planned as pre-commit; a trailer check needs the
  commit MESSAGE, which doesn't exist at pre-commit time → built as
  `.githooks/commit-msg` (hooksPath already active in this clone).
- **CTXK-004 is stale:** `cmd/assembler` does not exist in this repo (checked
  cmd/ directly) — ledger blesses the docubundle copy and records the gap.
  Register correction candidate; also a live example of the librarian's
  asserted-absence class pointing at OUR OWN register.

**LIVE now:** R0 (mission digest in classify_and_extract, before its return-JSON
anchor), R1 (review_mission_alignment → gate_mission_note → append_mission_note,
EVERY error path → create_next_item so a review failure can never block a
build; snapshot taken; 0 active classifier runs at apply; chain verified),
D1 (DIRECTION_LEDGER.md: canonical + copies + sha256s), D2 (commit-msg REAL
GATE, tested no-op/block/trailer, constitution restored byte-identical after
the test), D3 (100_CHECK: files/copies/seat-anchors across BOTH councils —
first run ALL GREEN), 101 (R1's consumer; also prints the R2 denominator).
**The guard-creation commit itself tripped the new gate** (ledger is a blessed
path) and passed via the trailer earned by the owner's approval message — the
gate's first live firing.

**NOT started (per plan):** R2 (needs a week of R1 numbers + owner grading),
R3 (Go + image window), R4 (feature/experience councils — theirs are co-edited
machinery; coordinate before seating), D4 (standards table, council-reviewed).

## Turn 32 — 2026-07-20 — v1.0.1140 deploy verified; config survived; one wrong call logged

Owner: fresh chassis build deployed 18:58:33 BST. Verification per CLAUDE.md
(pod, never tag):

- **Pod:** agent-chassis-5567d99bd6-5snzn, image v1.0.1140, started 17:58:20Z
  (matches the owner's stamp); past the 300s dispatch-quiet window at check time.
- **Truncation family LIVE in the binary:** strings-grep = 1 each for
  `stop_reason=max_tokens`, `done_reason=length`, `stop_reason=refusal`,
  `response truncated`. The superseding TruncatedError form (carries the
  partial; 019's upstream cause) is what shipped.
- **Re-seed clobber check: config SURVIVED.** 100_CHECK ALL GREEN (both
  councils' seat anchors); classifier R0 block + R1 chain wiring intact
  (design→review→gate→append verified). Note: the classifier row was updated
  17:57:45Z — 35s before the pod — by someone else's pre-deploy touch; my
  anchors intact, whatever it was preserved them.
- **008 was ALREADY CLOSED** by another thread (19:06, their own pod
  verification, §10 row updated) while I was drafting my own "verified but
  don't close yet" update — my append recreated the moved path as an untracked
  orphan fork, caught by the commit's pathspec error. Orphan deleted; their
  closure stands; logged in WRONG_CALLS.md (cheap check: `ls bugs_open/NNN*
  bugs_closed/NNN*` before UPDATING a case, not just before filing).
- llm_call_log had 0 calls in the first 6 min post-roll [UNMEASURED beyond
  that window] — behavioural evidence will accrue organically; the closing
  thread's evidence standard already covered it.

<!-- Append new turns below this line. Format: ## Turn N — date — one-line summary -->


---

## 2026-07-27 — third missed subsystem, and the case for a coverage sensor

Contributed from the oufe.com workstream (not the register thread's own work).

**What happened.** A session working on claims verification concluded that a whole
class of defect — a site claiming its own reliability — was "invisible to every
scanner in the estate", wrote that into the standing instructions of a live
council seat, and was one owner review away from building a redundant subsystem.
It was wrong: the banned-claim scanner is a bare regex over prose and always could
have caught it. Nobody had ever written a pattern for the class.

**Why this belongs in these notes.** Every one of the four things that session
missed — the scanner's real reach, a deferred decision in a spec's
open-questions section, a precedent in a sibling file, and a struct field declared
and never read — is a *design artefact of a subsystem that is not in this
register*. The register is the instrument that should have closed exactly that
gap, and it had a hole precisely where it was needed.

**The measurement.** Extraction froze 2026-07-13. **51 of 76 workstream
directories were created after it — 67%.** Until today there was no
claims-verification entry at all.

**The structural point, which is worth more than the entry I added.** This is the
third time a subsystem has been found missing (fixloop, model-directory, now
claims), and all three were found because somebody happened to be working next to
the hole. Three coincidental detections of one failure mode is a missing detector,
not bad luck.

And it rhymes with a defect found in the claims layer the same week: a decision
deferred "until two sites have evidence bases", never revisited at eight. **A
freeze with no watcher becomes permanent, exactly as a deferral with no watcher
becomes policy.** Everything else here relies on watchers — cooldowns, staleness
sweeps, claim timeouts, citation re-verification. Frozen indexes and deferred
decisions are the two classes with none.

**What I did and did not do.** Added `register/claims-verification.md` (12
concepts, first-hand citations) and the index note. **I did not** touch any
existing entry, re-run extraction, or alter the taxonomy — those belong to this
workstream. `bugs_open/106` carries the coverage-sensor proposal, modelled on
`verifier_coverage_test.go`'s sensor-plus-ratchet, and a cheaper interim: a
`covers-through:` stamp per register file, so the register can say where it
stopped looking instead of implying completeness.

---

## 2026-08-04 — bringing the register up to date: two checks that could not fail, and a 1,339-file deletion

Session "concept register". Task from the owner, in two halves: *look up what the
concept register is, how it was collected and what to do next, and bring it up to
date*; then *commit, and delete the out-of-date near-duplicate versioned documents
— they will still be in git*.

### What "up to date" turned out to mean

Not another extraction pass. Extraction froze 2026-07-13, **129 of the 155
workstream directories on disk postdate it** (`102_CHECK_register_coverage.py`,
2026-08-04), and re-sweeping was explicitly not the answer — the register grows by
per-commit registration under CLAUDE.md's "when you BUILD a new reusable
mechanism, register it" rule, watched by the 102 coverage check. So "up to date"
means: is the machinery that keeps it current actually working? It was not, in two
places, and both failures have the same shape — **a check whose result could not
have come out otherwise.**

**Defect 1 — the master index was silently 2% short.** 34 concepts had a
`### ID —` entry in a category file and **no row in `000_concept_index.md`**: all
of `CLM-001`…`012` (the first half of the claims-verification layer), plus
`IMG-067`, `LNK-029`, `DBI-025`, `PLAN-043`…`046`, `PUB-002`…`004`, `WII-009`,
`TL-031`, `SEO-002`, `VONC-011`, `FIX-054`, `DOC-067/069/071/072`, `LCO-005/006`,
`OPP-003`. The index is what a session or a council seat consults to find out
whether something exists, so a missing row is invisible in exactly the lookup the
file exists for.

It survived roughly twenty recorded re-measurements because of *how* the headline
was re-measured: each thread counted index rows and compared them to the previous
index-row count, confirming "my row landed, plus or minus a concurrent arrival".
That check is blind to a row that was never written by anyone. **A count of a
thing cannot audit whether the thing is complete — only a comparison against an
independent source can**, which here is the category files themselves. Backfilled
all 34 rows, and put the **drift pair** in the index header so the next thread runs
it: `comm` both ways between the entry ids and the row ids. Both lists are now
empty, 1,756 ids each way. The reverse list (a row with no entry) has always been
empty — the drift is one-directional because adding a concept is two edits in two
files and only the first is load-bearing for the author.

**Defect 2 — the coverage ratchet's annotated lines suppressed nothing.**
`read_ratchet()` keyed on the whole line, so any line a session had annotated —

```
bugfix_161_register_ratifies    # one-off: SQL + repair script for one false ...
```

— never matched the bare subsystem name and was reported as NEW on every single
run. **12 of 53 lines carried an annotation; 12 of the 17 names in that day's "NEW
since the ratchet" list were already accepted backlog.** 71% noise, in the check
whose own docstring says "that is what stops a coverage report becoming
wallpaper". The annotation is the most valuable thing in that file — it is the
reasoning for why a lane is a one-off — and it was costing suppression.

Fixed in `ratchet_name()`: strip an inline `#` or ` — ` comment, tolerate a
path-form line, take the basename. Also fixed `--update-ratchet`, which
regenerated bare names and would therefore have **deleted every annotation in the
file** the first time anyone accepted a new baseline. After the fix the report
went 17 NEW → 7 NEW, and after triaging those 7 it is quiet.

### The 7 real ones, triaged

- **`bugfix_158_reply_drop_sizing`** → **registered as ADP-018**,
  `check_silent_reply_drop`: a pre-commit detector for a reply produce whose error
  is only logged, and for `DeliverReply`'s outcome assigned to `_`. It exists
  *because* widening ADP-017's adoption is an RFC (architecture seat, round
  `7478233b`) — so the lane shipped a detector that holds the line at 2 of 9 sites
  with no behaviour change. Worth copying as a shape. Its own landmine is on the
  entry: the first version keyed on `responsesTopic` and missed 3 of the 4 known
  sites, because `websearch` spells it `responseTopic`, singular.
- **`bugfix_162_fix_proposer_plan_repair`** → **not a new concept**; it switched
  the already-registered opt-in repair loop (FIX-057) on for a **second consumer**.
  Recorded on FIX-057 itself, which is where a session looking for the mechanism
  would look, along with the consumer-census query that lane had to work out
  (`orchestration_states` has no `agent_type` column, and `agent_definitions`
  hides the consumer inside a JSON step, so the obvious queries error or lie).
- **`bugfix_179_deploy_path_override`** → already handled correctly by its own
  lane: it corrected IMG-067 in the same commit that shipped the fix, per the
  platform-seams condition (2). Ratchet line, no new entry.
- **`dartsonline_traffic`, `mortgagecalculator_couk_adoption`, `loancash_couk`**
  → site lanes, out of scope by the register's own bar.
- **`bugfix_087_writer_self_plans`** → opened today, docs-only, nothing callable.

### The deletion, and what it costs

441 version families under `docs/` (`X.md`, `X(1).md` … `X(39).md`), 1,973 members.
Deleted 1,339, kept 634, ~70MB. The rule was deliberately conservative and two
parts of it are load-bearing:

- **Never delete the unnumbered base file.** It is what every by-name reference in
  code, bundle scripts and prose points at — and in some families it is the
  *newest* member, not the oldest: `docs024/005_tool_pipeline.md` is 2026-07-26
  while its `(1)` is 2026-06-22. A blanket "keep the highest N" would have deleted
  the live canonical doc and kept a six-week-old copy.
- **Keep the newest member by mtime, not the highest N.** Five families have an
  older-numbered member with a newer mtime.

Also kept: 8 files referenced from outside `docs/` (scripts, `z_bundles`,
Makefile, the direction-integrity checker), and both members of the one family
where the newest is under half the size of the largest (`RUNBOOK_component_regen_clobber`
(49) is 25KB against (30)'s 56KB — read it, and the shrink is a deliberate trim
with history moved to NOTES, not a truncation casualty; kept anyway, because the
cost of being wrong is asymmetric).

**The measured cost: 43 of the deleted files are cited in a register `sources:`
line.** Those citations now resolve only through git. That is a real loss of
one-click provenance and it is worth stating plainly rather than discovering it
later: recover with `git log --diff-filter=D -- <path>` then `git show <sha>^:<path>`.
It is bounded by the extraction's own treatment rule — earlier members of a family
were only ever read `family-delta`, for concepts *absent* from the latest — so the
citations that matter most (`family-latest`) are all still on disk.

**What I did NOT do:** no register entry was rewritten, no status re-verified, no
taxonomy change, and `register/rebuild-cascade.md` was left alone — it was another
session's uncommitted work in the tree, and a pathspec commit is how you leave
that out.

---

## 2026-08-04 (later) — the watcher, in the framework rather than in the CLI

Owner asked: add a watcher for the index drift, **in the framework rather than the
local CLI — is that possible?** Yes, and there was a live pattern to copy rather
than a design to invent.

### The answer to "is that possible"

The framework already reads repo markdown on a schedule. `bugs-open-staleness-sweep`
(DOC-071, RFC_005 §3.3) is a K8s CronJob on `postgres:16-alpine` that resolves a
pinned ref through the GitHub Trees/Contents API, analyses the files, and writes
its findings straight to `doc_notes` over psql. It has been firing unattended since
2026-08-02. `component-fallback-check` and `single-owner-carriers-check` are the
same family on daily schedules. So this is an established fleet shape, not a
one-off, and the register watcher is a **cheaper** member of it: no LLM, no DB
reads, two regexes and a set difference.

### The three placement decisions, each with a reason that outlives the build

**Not a pre-commit hook**, which is where a check like this instinctively goes. A
hook fires only for a session already committing to the register — the session most
likely to have got it right — and never for drift accumulating between them. The
failure being watched for was never "someone skipped the check": ~20 sessions ran
the documented re-measurement faithfully and **not one of them could see it**,
because they compared the row count to the previous row count.

**Not a chassis agent.** Same reason DOC-071 is not one: no generic action lists a
repo directory or reads arbitrary non-`.go` files, and the function that came
closest was closed as a hazard (`bugs_closed/145`). An agent would add a queue and
a bill without adding judgement, since there is no judgement here to add.

**Findings to `doc_notes`, not `site_work_items`.** The queue's triage would route
them toward a handler that does not exist — the "findings die with no consumer"
failure the mission-reviewer lane hit and designed against.

### What it checks, and the one that immediately earned its place

Four comparisons: entry with no index row; row with no entry; duplicate id; and
**the index's own bolded headline against the actual row count**. The fourth was
almost an afterthought and it caught something within hours.

### Two things I got wrong, both caught by the harness rather than by review

**1. My positive-control assertion was false.** I asserted that at the pre-fix ref
the headline must also have drifted — reasoning that a broken register should look
broken from every angle. It did not: the headline said 1,721 and there were exactly
1,721 rows. **The headline was honest while 34 entries were missing**, because a
row count cannot see a row nobody wrote. The assertion now tests for *agreement* at
that ref, which is a far better control — it encodes the actual lesson instead of
my assumption about it.

**2. The first HEAD run was not clean, and the drift was three hours old.** Commit
`8bafcf9d4` (another lane, closing `bugs_open/087`) added a concept and its index
row correctly, then set the headline to **1,764** — the raw `###` heading count —
while the real row count was **1,757**. Measured the same minute: documented row
regex **1,758**, loose row regex **1,765**, raw headings **1,766**. Three numbers,
all correct answers to different questions, sitting in adjacent lines of the same
header — and the wrong one was picked up by a careful session seven hours after the
last correction. Corrected in place with a dated note, and the loose command is now
labelled as **not** the headline. This is the strongest argument for the watcher
that exists: the convention had been strengthened that morning, in that very file,
and it still went wrong before the day was out.

### Proof discipline, because a clean run proves nothing

`scripts/test-concept-register-drift-local.py --self-test` runs the **same
functions** the CronJob runs — the three GitHub calls are swapped for `git`, the
logic is never re-implemented — against `8f998e86b^`, the last tree carrying the
defect, and requires exactly 34 entry-without-row findings with `CLM-001` among
them; then requires clean one commit later; then **mutates** the fetched index text
by deleting one row and requires the check to name that exact id. Without the
mutation, an `analyse()` that returned empty sets would pass every other assertion.

### State, and what is owed

Committed, self-test green, manifests render, both `secretKeyRef` keys confirmed
present in `personae-platform-secrets`. **NOT deployed — the owner is running the
deploy.** DOC-074's status says so plainly and lists what is owed before it may
claim live: the apply, one manual job, and a `doc_notes` row.

⚠ **`make release` does not cover this.** Its `deploy-core`/`deploy-agents` targets
apply a hardcoded list of service overlays; a new CronJob directory is not in it.
`make deploy-concept-register-drift-check` is its own target — which is also the
honest reason the fleet release run this afternoon left the cluster without it.

⚠ **The second hand-pinned branch ref.** `REGISTER_REF` joins `SWEEP_REF`. A stale
one fails in the worst way: it reports on a register nobody is editing, so every
finding is unfalsifiable **and every "clean" run is meaningless**. Both manifests
need bumping when the working branch moves; the check refuses on a ref that does
not resolve, but cannot know that a resolving ref is no longer the one in use.

---

## 2026-08-09 — retiring the stored counts, and what the watcher's first four days measured

Owner ruling: take option (1) of the two recorded on DOC-074 — **delete the
hand-maintained headline** rather than keep watching it drift. Extended in the
doing to every stored count in the register, because the same disease was in all
109 category files and nothing had ever looked at them.

### The evidence for the ruling, gathered before acting

**Four commands count the index; all four answers are correct.** Measured the same
minute on 08-08: documented row regex **1,792**, loose row regex **1,799** (and the
loose one is quoted *inside the index's own header chain*, which is how it keeps
being picked up), unique entry ids **1,792**, raw `###` headings **1,800**. Twice
in four days a careful session published the wrong one.

**The per-file counts were worse, and unwatched.** All 109 category files carried a
stated count; **32 were already wrong**, 90 concepts of drift in total.
`batch-processing.md` claimed 5 concepts and held **0**. `documentation-system.md`
said 66 against 75. Nobody had ever compared them, because there was no reason to —
they look like documentation, not like state.

That measurement is the argument. A convention cannot fix a number that has three
plausible rival answers sitting beside it; only removing the number can.

### What was done

Index headline deleted; the entire "previously N, re-grepped after X" chain moved
**verbatim** to a frozen log at the foot of the file (the record of how it failed
is worth more than the numbers were). 108 category counts replaced by a pointer.
`README.md`'s live count likewise. The rule that actually mattered now stands alone
and unqualified: **entry and index row, same commit.**

### The inversion, which is the part worth copying

With no headline in the file, check 4 would have found nothing and reported
nothing — **indistinguishable from passing**. A retired rule and a silently dead
one look identical from outside. So the arm was inverted rather than deleted: it
now reports a stored count that has **come back**, in the index or any category
file. Proven by a second mutation in the self-test (re-add `**9,999 concepts**` to
`adapters.md`, require exactly that file to be named), plus the inverse assertion
that the live tree reports none — so a dead arm cannot pass as a satisfied one.

### Three things the tests caught before shipping

1. **The frozen log quotes the old headlines verbatim**, so a whole-file search
   would have reported a finding every run for ever — a watcher crying wolf about
   its own archive. Both count searches are now head-bounded, with the reason in
   the code rather than in a commit message.
2. **The retained command block still carried stale numbers in its comments.**
   Stripped: the commands stay, the figures do not. Leaving them would have
   recreated the exact artefact being retired, three lines below the paragraph
   retiring it.
3. **`LNK-031` was claimed by two lanes hours apart on 08-08** — `af2667453`
   (fragment resolution, holds the index row) and `85390ee33` (form-action repair,
   no row). Found by the duplicate-id arm on live data, not by review. Renumbered
   `LNK-032` with a visible note; the originating commit and `bugs_open/228` still
   say LNK-031, which is why the note exists.

### What the watcher's first four days actually measured — the real result

Five runs (four unattended). Two findings, and the second is the important one:

- a **headline mismatch reported on three consecutive days and corrected by
  nobody** — the report was right and unread;
- **`SCH-024` filed with an entry and no index row on 08-08**, four days after 34
  such rows were backfilled.

So the missing-row class recurs at roughly **one every few days**. 08-04 was not a
backlog that got cleared. That is the single most useful thing these runs produced,
and it is only knowable because something was counting.

### Honest loose end, and why it is not tidied away

`rebuild-cascade.md` still carries its stored count. Another session has had it
dirty in the shared tree since 08-04, and a pathspec commit takes a same-file
passenger — retiring one line would have swept five days of their uncommitted work
into a commit about something else. **Owed, not done.** The live check does *not*
special-case it and will name it daily, which is correct; only the local
self-test's assertion tolerates it, via a named set with a comment saying not to
grow it as a way of silencing findings.

### The uncomfortable observation to carry forward

The headline mismatch sat uncorrected for three days *while the watcher reported it
every morning*. A mechanism that writes into a table nobody opens fails the same
way the convention did. That is the argument for preferring **removal over
observation** wherever it is available: today's fix deletes the drifting artefact,
which needs no reader. The next intervention should try to do the same.

---

## 2026-08-10 — the prediction held, the leak has a rate, and a roll that proves nothing

**The 06:50 run matched the prediction exactly.** Before it fired I recorded what it
would say (1,803 entries / 1,803 rows, one finding, `rebuild-cascade.md`). It said
that. Predicted first, verified after — not reconstructed from the answer, which is
the only version of this that is worth anything.

**The leak now has a measured rate, and it is the reason to build the gate.**
`BIZ-031` and `WFA-012` both landed with an entry and no index row overnight; with
`SCH-024` on 08-08 that is three in three days, from three different lanes, none of
them careless. The 08-09 summary said we would learn the rate by measuring instead
of guessing. We have: **roughly one per day and a half.** That is enough to justify
the pre-commit rule, and it is the first item in the handoff.

**A fresh chassis rolled today, and it is a good example of a roll proving nothing.**
`WFA-012`'s entry says "inert until the next chassis image", so the tempting move
was to mark it live. Its change is **control flow with no new string literal** —
`ExtractNestedField` gained array indexing, and the symbol itself predates the
change and greps 8 times in the binary either way. That is a positive control that
cannot fail, exactly `DOC-073`'s case. **No pod-grep can confirm this one.** Its
row records the built/approved state and says the roll is not grep-verifiable,
leaving the live proof to the lane that owns it and knows how to induce it.

**Handoff written** (`HANDOFF_2026-08-10_continue_here.md`) because the originating
chat grew long. It carries the rate table, the three landmines specific to this
lane, and the ordered next steps — gate first, then the owed stored count, then
staleness.

---

## 2026-08-10 (later) — the gate is built, and the rate we had been quoting was too low

Handoff item 1, done: `check_register_entry_without_row` in `scripts/pattern-check.py`,
run by `.githooks/pre-commit`, registered as **OPP-006** (entry + index row in the
same commit — its own rule, applied to itself). Commit `7db343ee7`.

### What it does, in one line each

- **Arm 1** — a commit ADDS `### ABC-001` to a register category file and the id has
  no `| ABC-001 |` row in the index *as this commit will contain it*.
- **Arm 2** — the id is already taken somewhere in the register.

Both advisory, like everything in that file. 0.12–0.16s when the register is
touched, 0.05s otherwise (early return before any git call).

### The measurement, and the correction it forces

Swept the 14 days to 2026-08-10, in audit mode, one commit at a time:

| | |
|---|---|
| commits touching the register | 398 |
| of those, commits ADDING ≥1 concept entry | 159 |
| did it correctly → silent | **133 (84%)** |
| shipped an entry with no row → fires | **26 (16%)**, 34 findings |
| false positives on inspection | **0** |
| against the whole commit stream | 26 / 3,324 = **0.8%** |
| arm 2 | **1 finding in 398 commits** — the known `LNK-031` collision |

**The fire rate is not the argument, and this is the part worth carrying forward.**
A count of firings cannot distinguish a real leak from a two-commit workflow the
gate would merely nag. The GAP can. Of the 34 findings, 32 eventually got a row and
**the median entry waited 93.1 hours** for it; 23 of 32 took ≥24h; only 3 closed
inside an hour. **21 of them were closed in a single sweep by another session's
08-04 backfill** (`8f998e86b`) rather than by their authors. That is the shape of a
leak somebody else pays for.

> **CORRECTED 2026-08-10 — the leak rate in the handoff and in yesterday's summary
> is too low, and the reason is structural rather than arithmetic.** Both say
> ~1 per 1.5 days, taken from what the watcher reported. The commit-level sweep says
> **~1.2/day** since the 08-04 backfill. **The watcher undercounts BY CONSTRUCTION:
> it can only name entries still missing at 06:50**, so a row backfilled the same
> afternoon never appears in any daily row — `VIZ-015` (gap 0.0h) and `WII-010`
> (0.8h) are both invisible to it and both real instances of the class. A report's
> count is bounded by its sampling interval; a commit sweep is not. Nothing was
> measured wrongly; the instrument could not see the whole population, and neither
> the handoff nor the summary said so.

### The misstep, which is the most transferable thing here

**Arm 2 shipped inert and passed every test I had.** `git grep <pattern> --cached`
is not a synonym for `git grep --cached <pattern>`: git reads the trailing word as
a REVISION and dies `fatal: unable to resolve revision: --cached`. `pattern-check.py`'s
`sh()` captures stdout only, so that fatal arrived as an **empty string** — empty
corpus, no collisions, no findings, **no error**, which is exactly what a healthy
tree also produces.

It passed the full 398-commit audit sweep, because in `--commit` mode the argument
after the pattern IS a sha, which is legal. **Only staged mode was dead — and staged
mode is the only mode the hook runs.** A green sweep over real history told me
nothing about the code path that actually executes.

What caught it: a positive control that staged a known-duplicate id into a temporary
index and REQUIRED a finding. Nothing else could have — every other test I had was
consistent with the arm being dead.

Two fixes, both kept: the invocation, and an assertion that the corpus it read
contains the very headings the commit adds (true by construction), printing an
audible skip when it does not. So a broken read can never again pass as "no
collisions". Filed fleet-wide in `LANDMINES.md` (`728d7d891`) — measured while
filing: **7 of the 22 helpers under `scripts/` that use `capture_output=True` never
test `returncode` and never pass `check=True`**; `pattern-check.py` was the 7th.

### The five controls, all in an isolated `git worktree` so the shared tree and index were never touched

1. entry staged, no row → **fires** (arm 1)
2. row added in the WORKTREE but not staged → **still fires** — the pathspec hole; a
   worktree read would call this clean
3. row staged too → **silent**
4. duplicate id staged → **fires** (arm 2) — the control that caught the inert arm
5. mutate the grep back to the broken form → **audible skip**, not silence

Plus the honest one: staged without its own index row, the gate named **OPP-006**.

### Two live facts checked while here, both correcting the record

- **The CronJob redeploy owed since 08-09 HAS happened.** The 08-09 summary and the
  owner log both end saying the ConfigMap is stale until `make
  deploy-concept-register-drift-check` is re-run. It is not: the live ConfigMap
  (`concept-register-drift-check-script-bg959m9c7f`, the one the CronJob actually
  mounts) is **byte-identical to the repo's `check.py`** — 16,496 bytes, `diff`
  clean, and it carries the inverted stored-count arm. So the daily run is doing
  all four current checks, not the old four.
- **`REGISTER_REF` is pinned to `087_towards_multiple_domains`**, which is the live
  working branch — the stale-ref landmine is not biting today. But **65 commits sit
  unpushed** on that branch (`git ls-remote origin refs/heads/087_towards_multiple_domains`
  → `5a68d6caf6d9`, local HEAD well ahead), so tomorrow's 06:50 run will still name
  `BLD-018` and `DIAG-042` even though both rows are committed. That is the watcher
  reading the PUSHED branch, working as designed — not a regression, and not
  something this lane should push on other sessions' behalf.

### State at the end of the session

- **1,817 entries / 1,817 index rows.** The missing-row class is closed at HEAD:
  `BLD-018` and `DIAG-042` rows written from their entries (`a332522df`), neither
  claiming more than its entry does (BLD-018 inert, no caller; DIAG-042's TRUE
  branch has never fired).
- The drift check's **only** remaining finding is `rebuild-cascade.md`'s stored count.
- **Handoff item 2 is still correctly blocked, re-checked today.** The file is still
  dirty in the shared tree, and `ls -l` puts its last write at **2026-08-08 20:41** —
  two days ago, so that edit is active work, not an abandonment. Committing my one
  line would take their REB-003 rewrite as a same-file passenger. Left owed, again,
  and the reason is now dated rather than inherited.

---

## 2026-08-10 (later still) — the staleness survey: the register's evidence is ageing, and the obvious mechanism is ruled out

Handoff item 3, surveyed rather than built — full write-up in
`FINDINGS_2026-08-10_staleness_survey.md`. The short version and the two things
that cost me a wrong number.

**Four signals, measured against 1,818 entries:**

| signal | result |
|---|---|
| entries citing a chassis version | 129 — **80 of them 50+ versions behind** (fleet on v1.0.1280; `SYS-077`/`HITL-020` cite **v1.0.407**, 873 back) |
| status still claiming "not live" | 44 matched, **6 already corrected in place**, 38 remaining → **~20 genuine** after reading |
| `sources:` paths that no longer resolve | **96 of 2,611** judgeable citations (3.7%), mostly the numbered-docs tree deleted 08-04 |
| entries citing `bugs_open/NNN` now in `bugs_closed/` | 156 — **one-directional**: the owner's 08-06 ruling means a *non*-moved bug proves nothing |
| `verify-later:` present | 1,754/1,818 — **a template field, so not a signal**; recorded so nobody quotes it as one |

**Proven at the artefact, both replicas of `v1.0.1280`, negative control 0 in the
same exec** — not inferred from the roll count:

- `FIX-055` said **"NOT yet live"**; `hasGatingObjection` (1) and
  `gatesOnlyBecauseTruncated` (2) are in the running binary. **False for some part
  of 13 days and 22 rolls.** Corrected in place.
- `SCR-002` said **"inert until the chassis image rolls"**; `fetch_provenance.go`
  is in the binary, ~23 rolls later. But it is **live and unexercised** — vet
  collection has been off since 2026-03-18. Corrected, and the two claims
  **separated**, because the old wording conflated a claim that expires by itself
  with one that does not.

### Misstep 1 — a publishable-looking number that was wrong

The broken-citation count came out at **187**. Sampling 22 of them showed a large
minority were **artefacts of my own regex**, not defects in the register: tails of
brace notation (`{PLAN,NOTES}_x.md` → `_x.md`; `check_site_unreachable{,_test}.go`
→ `_test.go`) and abbreviated citations containing a literal ellipsis
(`docs021.../025_….md`). Excluding 92 such tokens gives **96**. I had the 187 in a
draft paragraph before the sample. **The check that caught it was asking git
whether each path had EVER existed** — "never existed under that name" is the
signature of a parsing artefact, and "existed once, now deleted" of a real one.

### Misstep 2 — and this one is the design finding

My status regex said **38 entries still claim not-live**. Reading all 38 said
**~20**. The overcount was not a weak pattern; it was the field doing four things
no pattern can classify:

- `WFA-006` — **"runtime-inert BY DESIGN"**: a permanent property that reads
  exactly like an expiring one. A checker flagging it is wrong for ever.
- `VONC-011` — **"deployed — UPDATED 2026-08-02, was `built, not live`"**: the
  stale claim quoted *inside* its own correction. The same shape as the frozen-log
  trap that forced both count searches in `check.py` to be head-bounded — **a
  watcher crying wolf about its own archive**, one level down.
- `CLC-013`, `STY-056`, `WFA-009`, `CGV-031` — **half live, half not.** One entry,
  two statuses, two clocks; there is no single answer to grade.
- `PBP-037` — **"INERT end to end until three things happen IN ORDER"**: a
  precondition chain, not a state.

**So: a staleness checker must NOT parse `status:`.** It should key on things with
no prose ambiguity — a version number, a file path, a bug id, a date — and report
**"this entry's evidence has expired"**, never "this entry is wrong". That is the
bar the drift check already holds ("nothing here is a claim that an entry is
WRONG") and is why it is trusted enough to be read at all.

The one §1-shaped check worth building does not need the prose either: pair the
**entry's own commit date** against the **roll clock** (`IMAGE_TAG` bump commits —
107 in 14 days) and emit a **candidate** with the pod-grep suggested, not a verdict.

**Cheapest first move, and it is not a checker:** make **version lag** visible. 129
entries already carry the number and the fleet's current version is one `kubectl`
call away.

**The 20-entry worklist is in the FINDINGS doc, not corrected here on purpose** —
each needs a pod-grep against a symbol its own lane chose, and `WFA-012` cannot be
settled that way at all (control flow, no new string literal, `ExtractNestedField`
greps 8 times either way — `DOC-073`'s positive-control-that-cannot-fail). Writing
"live" on an entry I had not proved would manufacture exactly the false confidence
this survey exists to measure.

**The finding underneath all of it:** every mechanism this lane has built —
coverage, drift, and today's authoring gate — asks whether the register agrees
**with itself**. Nothing has ever asked whether it agrees with **the platform**.

---

## 2026-08-10 (evening) — the worklist died four hours after it was written, and the mechanism that killed it is the interesting part

A fresh chassis rolled: **`v1.0.1283`**, carrying `BLD-019`'s build provenance.

**The stamp reads back.** `strings /app/agent-chassis | grep -oE '^[0-9a-f]{40}(-tree)?$'`
returns `d3c09cc746e563b6339831cfb69576eb52135c43` on **both** replicas, no `-tree`
suffix (a clean committed build, as designed), `buildinfo.GitCommit` present 4×.
`BLD-019`'s own entry corrected from "INERT until the fleet rolls" → **LIVE**.

**What it replaces.** The survey's §1 worklist said each of ~20 entries needed a
pod-grep against a symbol its own lane chose. That is now one command:

```bash
git merge-base --is-ancestor <the entry's own commit> d3c09cc746e563b6339831cfb69576eb52135c43
```

**Controlled before use** — an ancestry test that always answers yes answers
nothing. Positive: `FIX-055`'s `3a59b5012` → IN, **agreeing with the independent
pod-grep that proved the same entry hours earlier**, which is the best kind of
corroboration because the two methods share no machinery. Negative: `3ac87646a`,
an off-branch merge → NOT IN. The test can return false.

**Result.** The build was made from *exactly current HEAD* (`rev-list --count
stamp..HEAD` = 0), so every commit on this branch is in the image and the
roll-conditional class settled wholesale. **19 entries annotated in place.** The
annotation states only what is proven — the Go code is in the running binary — and
**explicitly declines** to say the feature is exercised, because for several it is
not: `CQ-019` awaits migration 303, `PLAN-047` seed 306, `PBP-025` a `run_checks`
array naming it, `TL-038`/`TL-040` a live fence. Conflating those two is the exact
error `SCR-002` was corrected for this afternoon.

**The new finding, and it is an authoring rule, not a checker.** **13 of the 29
entries examined cite NO commit sha.** Provenance can then only infer inclusion
from the entry's date — sound, unverifiable, and a slide straight back into the
guesswork the stamp exists to remove. Those 19 annotations come in two variants
for that reason, and the no-sha variant says so on its face.

> **An entry whose status is conditional on a roll must NAME ITS COMMIT.** Nine
> characters at authoring time; a one-command check for ever after. Candidate for
> the authoring gate (OPP-006), not for a watcher — same argument as the missing
> row: put the check where the error is made.

**`WFA-012` is the clean demonstration of why `BLD-019` earns its place.** It is
unsettleable by pod-grep — control flow, no new string literal, `ExtractNestedField`
greps 8 times either way, `DOC-073`'s positive-control-that-cannot-fail. It cites
two commits. Both IN. **Provenance settled in one command what marker-hunting
structurally could not settle at all.**

Untouched by this roll, and still the open work: version lag (80 entries 50+
versions behind), unresolvable citations (96), moved bug references (156).

---

## 2026-08-11 — the gate was never ignored, it was inaudible: 45% of commits never see the pre-commit advisory

Picking up `HANDOFF_2026-08-10b`'s READ FIRST item, which asked the right question
and said nothing recorded could answer it: *did that session never see OPP-006's
output, or see it and judge the row could wait?* Something can answer it — the
harness's own session transcripts, which are on this disk and which neither the
watcher nor the register can see.

**Read back the offending commit.** `5c7b115c5` (DES-082/083, no index rows) was
made by session `c8172188`, and the recorded tool result contains **no pre-commit
output at all** — no scope block, no pattern-check block, stderr empty — while the
**commit-msg** hook's council-trailer nudge is present in full. The command ended
`2>&1 | tail -8`. The recorded stdout is **exactly 8 lines**.

**The mechanism.** The pre-commit hook prints FIRST; git prints `[branch sha]` and
`N files changed` LAST. `tail -N` is a head-cutter, so it removes precisely the
part that warns you and keeps precisely the part you asked for. The commit-msg
nudge survives because it prints later — which is why the estate reads as "hooks
are quiet, so all is well".

**It is fleet-wide, not one session.** `scripts/advisory-delivery-sweep.py`, over
every commit made through the tool since the scope report shipped (2026-07-18):
**2,669 multi-file commits; 1,470 (55%) delivered the block; 1,199 (45%) did not;
1,137 of those misses (95%) were cut by the session's own `| tail`; 258 distinct
sessions.**

Two controls, because a 45% with no control is a correlation:
- **the pipe width predicts delivery** — misses are `tail -N` ≤ 8 in 1,031 of
  1,137 (median 5); commits that delivered *despite* a pipe are N > 8 in 465 of
  595. If the cause were something else, N would not separate them, and this could
  have come out flat.
- **exactly-N arithmetic** — for 666 misses the recorded stdout is exactly N lines
  for a `tail -N`. Output existed and was cut. No content can fake that.

**What was being suppressed is much bigger than this lane.** Not just OPP-006: the
commit-scope report (CLAUDE.md's main defence against committing another session's
work), all 17 `pattern-check.py` checks, and the architecture RFC-trigger signal.

### The wrong turn, and it was the interesting-theory trap

My first hypothesis was that `git commit <pathspec>` — the form CLAUDE.md
*mandates* — leaves those files out of the index, so every check reading
`git diff --cached` is blind to exactly the commits the house style produces.
Tidy, structural, explains the symptom, **false**. Git builds a temporary index
for a partial commit and points the hook at it. Refuted in a scratch repo in about
a minute across four shapes: bare commit after `add`; pathspec commit of an
unstaged modification; pathspec commit mixing `add`ed new files with unstaged
modifications (`5c7b115c5`'s exact shape); and a pathspec commit naming one file
while a different file sits staged. All four saw a faithful index.

The lesson is not "test your hypotheses". It is that **two mechanisms explained
the symptom and I reached for the interesting one.** The dull one — the session
piped the output away — was correct, and one grep of the transcript for the
command that made the commit would have found it first. I also had to be told by
my own data: the first sweep left 495 "misses" where nothing had been cut, and
chasing those is what surfaced that 71 shas were from other repos entirely (the
auto-memory git dir, scratch repos) where no hook is installed.

### The fix — OPP-007, and deliberately not teeth

`scripts/commit-advisory-postuse.py`, a `PostToolUse` hook on `Bash`: on a command
containing `git commit` whose output carries git's summary line, re-run
`commit-scope-report.sh --commit <sha>` and `pattern-check.py --commit <sha>` and
deliver via `hookSpecificOutput.additionalContext` — out of band, where no pipe in
the session's own command can reach it. `commit-scope-report.sh` gained the
`--commit <sha>` mode `pattern-check.py` already had (by then the index is clean);
staged mode re-controlled and unchanged.

- **No git hook can fix this**: `post-commit` output also lands before git's
  summary (verified in the scratch repo), so the same `tail -3` eats it.
- **`additionalContext` on stdout, exit 0.** stderr + exit 0 reaches nobody —
  `scripts/memory-index.py` was wired that way and was mute for six days.
- **Controls on the hook itself**: positive (the real `5c7b115c5` payload
  reproduces both blocks); negatives (non-commit Bash call → silent; output that
  already carried the scope block → scope dropped, pattern-check still emitted;
  a sha from another repo → silent; a failed commit with no summary line →
  silent; malformed payload → silent).

**And it is NOT an argument for making OPP-006 blocking — it is the opposite.**
The evidence that read as "the gate is being ignored" was an artefact of a pipe.
`pattern-check.py`'s standing argument against blocking on a shared tree is
untouched. The honest next test is OPP-007's verify-later: if delivery rises and
the watcher's missing-row count does *not* fall, then delivery was never the
binding constraint and enforcement reopens on real evidence.

Filed: `FINDINGS_2026-08-11_advisory_delivery.md`, `RUNBOOK` §B11, register OPP-007
(+ the dated answer written back onto OPP-006's verify-later), and a fleet-wide
`LANDMINES.md` entry synced to `doc_notes` (6 footprint rows, verified in the DB).

**The live control on the fix itself, 2026-08-11.** OPP-007 shipped in `05d8b379e`.
That commit was made deliberately *without* a pipe, and the scope block printed —
10 files across 3 areas, all mine, and pattern-check silent because the OPP-007 row
was already at HEAD. **This second commit is the real test: it is piped through
`| tail -3`, so the pre-commit blocks cannot survive in its output, and the only way
the advisory can reach me is the new hook.** Recorded here before running it, so the
prediction is on the record either way. ⚠ If nothing arrives, the likeliest cause is
not the hook but the harness: `.claude/settings.json` was edited mid-session and a
running session may hold the hook set it started with — in which case the live proof
falls to the next session and this claim stays [UNVERIFIED] until then. Rerunning
`scripts/advisory-delivery-sweep.py --since 2026-08-12` is the fleet-level version of
the same question.

> **RESULT — PROVEN, same minute.** Commit `aa78871d8`, piped through `| tail -3`;
> its recorded stdout is 2 lines and carries neither block. The advisory arrived
> anyway, as a `PostToolUse` system-reminder next to the tool result: the scope
> block for both files, plus the re-run line. **So delivery is verified at the
> artefact — the reader received it — on the exact commit shape that has been
> losing it 45% of the time.** Two secondary facts fell out: the hook took effect
> **mid-session with no restart** (`.claude/settings.json` was edited minutes
> earlier, so the [UNVERIFIED] caveat above is retired rather than deferred), and
> the de-duplication arm behaved — `05d8b379e`, committed without a pipe, printed
> the block in its own output and drew no duplicate from the hook.

---

## 2026-08-12 — the verify-later I wrote yesterday could not verify, and it printed a regression

**First act of this session was to run the verify-later `FINDINGS_2026-08-11` names:**
`scripts/advisory-delivery-sweep.py --since 2026-08-12`. It printed **38.2% delivered**,
against a documented pre-fix baseline of 55%. Taken at face value: OPP-007 made delivery
worse.

**It is the instrument.** The sweep decided "delivered" with
`"commit scope:" in toolUseResult.stdout` (line 91 as written). OPP-007 delivers through
`hookSpecificOutput.additionalContext`, which is **not in `toolUseResult` at all** — the
harness writes a separate record:

```
type: "attachment", attachment.type: "hook_success",
attachment.{hookEvent,hookName,command,stdout,stderr,exitCode,toolUseID}
```

with the hook's JSON in `attachment.stdout`. So the sweep scored **every** out-of-band
delivery as a miss. It was blind to the only path the fix uses.

**How the field was found, rather than guessed:** a recursive JSON-path dump over every
transcript line containing the hook's own wording. That printed the answer directly —
`.attachment.stdout` (51) and `.attachment.content[0]` (51), against 3 in
`.message.content[tool_result].content` (which turned out to be yesterday's *manual control
runs*, not deliveries). Guessing at `system-reminder` shapes would have cost an hour;
enumerating paths cost one command.

**Two slips inside the repair, both recorded because both looked like real failures:**

1. **The join matched nothing at first.** The hook quotes git's **9-char** abbreviated sha;
   the population carries the full 40. `sha in set` → 0 hits, which reads identically to
   "the hook never fired". Fixed by bucketing the hook's shas by length and prefix-testing.
2. **My pre-ship control failed: "24 deliveries before the hook shipped".** That reads as a
   broken parse. It was **my threshold** — transcripts are UTC, this estate writes BST. The
   hook shipped 19:18 BST = 18:18Z, and the earliest genuine delivery is **18:11:38Z**,
   *seven minutes before its own commit*, because it was live in the working tree first
   (`.claude/settings.json` and the script were wired, tested, then committed in
   `05d8b379e`). Live and committed are independent facts, and here the gap was visible in
   the data. Control restated in UTC: **0 before 18:11Z**, all time.

**The patched sweep, and the true reading.** Channels reported separately (so the original
`tail`-width mechanism claim stays falsifiable), plus a control that channel 2 cannot
predate its own hook:

```
2026-08-12: 36 multi-file commits — REACHED 36 (100%): own output 13, out of band 23, missed 0
controls: oob=0 on every day 08-05→08-10 · pre-fix 55–56% unchanged
          channel-1 tail width still separates (1031 vs 131 at N<=8; 106 vs 474 at N>8)
```

**The second half of the verify-later is NOT answered, and I nearly reported that it was.**
The register's leak signal looks perfect — **0 OPP-006 findings across all 17
register-touching commits** since the hook went live, no entry-without-row at HEAD. But the
demand is **4 entry-adding commits** (`7d2377149`, `3c962927d`, `ef1374426`, `05d8b379e`),
and at OPP-006's measured 16% historical leak rate **P(zero | nothing improved) = 0.50**.
Uninformative. **~14 entry-adding commits for 90% power, ~18 for 95%.** Also used the
per-commit sweep rather than a HEAD count, because this lane's own landmine says a snapshot
cannot see a leak repaired the same afternoon.

**The transferable shape:** *when you move where a signal is delivered, the thing that
measures delivery is the first thing to go stale — and it stays green while doing it.* This
sweep had four controls (tail-width separation, exact-line arithmetic, foreign-sha filter,
stated denominator); all four passed and **all four were on the old channel**. A verify-later
needs a positive control on the **new** path. The one-line version:
`grep -c additionalContext scripts/advisory-delivery-sweep.py` → **0**.

Filed: patched sweep; corrected banner + new §"What the 100% does and does not buy" in
`FINDINGS_2026-08-11`; `RUNBOOK` §B11 (five new gotchas); OPP-007's entry gains a fleet-wide
status-evidence line, a landmine, and a verify-later split into PROVEN/OPEN halves;
`WRONG_CALLS.md`; `LANDMINES.md` + `landmines-sync.py --apply` (2143 owned rows).

**Not done, deliberately:** `rebuild-cascade.md`'s stored count is still owed, fourth session
running — the file is still dirty in the shared tree with another session's REB-003 rewrite
(mtime unchanged at 2026-08-08 20:41, 3 added / 3 deleted), so retiring the line would take a
same-file passenger. The drift check's single HEAD finding is that stored count and nothing
else. The three staleness signals (version lag 80, unresolvable citations 96, moved bug refs
156) remain untouched.

---

## 2026-08-12 (second piece) — version lag: surveyed the premise, and it needed narrowing

Handoff item: *"Cheapest next move, and it is not a checker: make version lag visible. 129
entries already carry the number; the fleet's version is one kubectl call."* Did that, but
measured the premise first, and the measurement changed the design.

**The premise, tested:** *"the cleanest mechanical signal in the register; needs no prose
parsing."* Extraction is clean. **Interpretation is not.** Classifying 315 citations by the
words before them: **244 (77%) unclassified**. Cause — `"deployed in chassis v1.0.1029"`
(permanent fact) and `"both replicas of v1.0.1218 return X"` (expiring verification) are
indistinguishable by pattern. Raw lag flags 111 items, mostly permanent. That is the
"report nobody reads" failure the 08-10 design conclusion names, reached from another
direction.

**What worked: the register's own field vocabulary, which is structure, not prose.**
`status:`/`status-evidence:` = current-state claims by convention. 273 citations → **206
across 139 entries**, 67 excluded (24% — the exclusion is itself the control that the key
does work). `status-evidence` median lag **103** vs `status` **28**: status lines get
updated, evidence does not get re-verified.

**The sharp class, and it rests on a fleet fact rather than a linguistic one:**

```sql
-- control first: 187 live rows HAVE a tag, so a zero elsewhere is absence not blindness
SELECT image_tag, count(*) FROM agent_definitions
 WHERE is_active AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL GROUP BY 1;
--  v1.0.1290 | 187      ← uniform. all of them.
```

So a tag read off a live row dates the observation. **`SYS-077`** claimed a row "still
references `v1.0.407`" (883 behind) — the row's tag is `v1.0.1290`; corrected in place.
**`HITL-020`** cites the same `v1.0.407` and is **NOT** stale: it describes the *seed file*,
which still says it. Same version, same day, opposite verdicts.

**Three wrong turns, all recorded because each looked like a result:**

1. **`default_config::text LIKE '%v1.0.407%'` → 0, and my control also → 0.** The predicate
   was blind: the image lives in dedicated columns `image_repository`/`image_tag`, not in
   `default_config`. `\d agent_definitions` first, as the rule says. **The control is what
   saved this** — a bare zero would have been written up as "the claim is false".
2. **`type ILIKE '%hitl%' OR display_name ILIKE '%content-approval%'` → 0 rows, and I said so
   in chat: "not in live config at all". WRONG.** The agent's type is
   `simple-content-writer-with-approval` — no "hitl" in it — and the group is stored as
   **"Content Approval with HITL"**, so the slug `content-approval-hitl` returns 0 as well.
   Both are loaded. A grep proves absence only for the spelling it searches, and I searched
   the spelling the *entry* used rather than the one the *rows* use.
3. **My own detector: a proximity window, then a misleading display.** "`image` within 12
   chars" gave **2 of 7** precision (it read "inert until an image roll" and
   `IMAGE_TAG=v1.0.1190` as live evidence). Adjacency — strip the punctuation, require an
   image token immediately before — is **9 of 9**. Then the worse one: the report printed each
   hit's *line head*, truncated to 200 chars, so for multi-citation lines it showed a
   different citation from the one that matched, and **three correct hits read as false
   positives.** I was one step from loosening a 9-of-9 detector because its own output lied
   about what it had tested. **Show the evidence you tested, windowed on the match.**

**Shipped:** `scripts/report-register-version-lag.py`, registered **DOC-077** (entry + index
row in the same commit; pairing verified in the working tree — the drift harness reads a
**ref**, so it cannot see an uncommitted entry, and HEAD moved twice during this session).
`SYS-077` and `HITL-020` corrected in place with the measurements inline. `HITL-020`'s
verify-later answered.

**Not done:** the two remaining signals (96 unresolvable `sources:` citations, 156 moved bug
refs). The transferable question for both: **is there a key that does not require reading
prose?** Version lag only became trustworthy when it stopped parsing sentences and keyed on
structure the register already maintains. Also still not done: `rebuild-cascade.md`'s stored
count, fourth session running, still dirty with another session's REB-003 rewrite.

**Postscript, same session — my `SYS-077` correction shipped under another session's commit.**
I named 8 paths on the pathspec commit and 7 landed. `system-architecture.md` was gone from my
working tree by then: commit `4a6e39c28` (15:24 BST, "owner rulings 08-12 recorded + fresh
cold-start handoff…", an unrelated lane) had taken it as a same-file passenger between my edit
and my commit. **Nothing is lost** — the correction is at HEAD, verified
(`git grep -c "CORRECTED 2026-08-12: that was true when written and is false now" HEAD -- …` →
1) — and forward-only forbids an amend, so it stays where it is. Recording it because
`git log` on that file now attributes the SYS-077 correction to a commit about owner rulings,
and the only place that trail can be repaired is here. **The tell was arithmetic, not intuition:
8 paths named, 7 files in the commit.** Count them.

---

## 2026-08-12b — signals 2 and 3 closed (`DOC-078`), and the field key does NOT transfer

Picked up from `HANDOFF_2026-08-12_continue_here.md`. Took the carried question — *is there a
key that does not require reading prose?* — into the two open staleness signals. Full account
in `FINDINGS_2026-08-10_staleness_survey.md`'s **UPDATE 2026-08-12b**; this is the technical
log.

**Answer: no, not the same key.** Unresolved-citation rates by field run 10–37% with no break
anywhere (`sources:` 19%, `what:` 25%, `verify-later:` 10%, `status-evidence:` 23%). The reason
is structural: a citation's field predicts nothing about whether its target was renamed. What
IS total and mechanical is **what git can say about the cited target** — at HEAD / moved /
deleted / never — and the middle verdicts name their own repair. The field key returns as
**severity**, not as a filter.

**Built:** `scripts/report-register-citation-rot.py` (`DOC-078`, committed `b9b32ba92`).
Read-only, ~1.5s, no cluster, no DB, not scheduled, not a checker. `--self-test` (10 cases,
each naming the wrong answer it guards) · `--worklist` (every citation git can locate, with
its target) · `--list <VERDICT>`.

**The measurement** — 7,793 path citations / 1,767 entries: **75% resolve as written**, 286
`MOVED-AT-HEAD` (target printed), 316 `BUG-MOVED`, 194 `DELETED`, 769 `MOVED-AMBIGUOUS` (a
bare filename matching several files — under-specified, never wrong), 345 declared unjudgeable,
**4 `NEVER-REPO-PATH`**. The four are listed in the FINDINGS update; `ADP-018`'s is the sharp
one (right bug number, wrong directory, date and slug).

### Misstep 1 — the instrument manufactured the headline, and I said it out loud first

Stripping the `(N)` suffix unconditionally produced **27 of 34** "never existed" findings.
`(N)` is an extraction-unit id in some citations and **part of the actual filename** in others
(`002e_concept_spark(6).md`, `016b_debugging_guide_merged(3).md`). Caught by looking up what
the file is called before writing the repair. Logged in `WRONG_CALLS.md` with the general
check: **never report an absence without printing the near-miss git does have.**

### Misstep 2 — the fix repeated it one level down, and the number never moved

Resolving as-cited *and* stripped, I kept the **last** verdict rather than the **best**, so a
citation that resolved exactly as written was overwritten by the stripped form's failure —
same 15 entries, same wrong count, different bug. **A figure that reproduces across two runs of
the same instrument is not corroborated**; both runs shared the instrument, not the world.

### Misstep 3 — the first control failed, and that is the landmine

`git rev-list --objects --all` **dedups by object**: 791 of 9,301 HEAD paths absent from its
output, **791/791 content-identical duplicates**. A path-existence check built on it reports
live files as never having existed. Enumerator that is actually total:
`git log --all --no-renames --pretty=format: --name-only`. In `LANDMINES.md` (synced), and the
script refuses to print if `HEAD ⊆ ever` fails.

### Three more, cheap but worth the line

- **The self-test caught a bug the full run could not show me.** A bare `NOTES.md` was
  "resolving" to whichever file the index happened to list first — a basename with no directory
  carries no evidence about which file it means. Fixed to require uniqueness; `MOVED-AMBIGUOUS`
  went 81 → 769, i.e. **the honest number was nine times the flattering one**. A synthetic case
  can fail in a way 7,793 real ones cannot, because in the real corpus the wrong answer still
  looked like a path.
- **`Council-Submitted: n/a` is REFUSED by the commit-msg gate**, correctly — the trailer is a
  join key for the 098 report and a non-UUID value resolves to nothing. If your change is out of
  the gate's scope (`platform/`, `internal/`, `pkg/`), **omit the trailer entirely** and say so
  in prose. Cost: one blocked commit.
- **`pattern-check` fires `new-capability-surface` on a doc that QUOTES a dead path.** The
  FINDINGS table names `internal/gateway/hitl_handler.go` in order to say it does not exist, and
  the checker read that as proposing a new `internal/gateway/`. Advisory, never blocks, and it
  is the documented false-positive shape ("naming a path you have deliberately decided against
  fires this too"). Recorded so the next reader of that commit does not chase it.

### ⚠ A HARD RESET DESTROYED EVERY SESSION'S UNCOMMITTED WORK AT 18:38:52

Between writing this lane's docs and committing them, another session ran a reset — reflog
`1ee940968 HEAD@{2026-08-12 18:38:52 +0100}: reset: moving to HEAD`. It discarded **all
uncommitted tracked-file changes across the whole repository**, not just this lane's.

**What this lane lost and re-did:** the FINDINGS 08-12b update, this NOTES entry, and the
`README_where_we_are` entry — all three re-appended from the session's own transcript, content
unchanged (`724359d72` and this commit).

**What the tree lost, measured rather than asserted** — files that were modified at session
start, are clean now, and were NOT committed today, so their changes went with the reset:
`CLAUDE.md` (last commit 08-11 13:30), `cmd/config-key-audit/main.go` (08-11 17:14),
`internal/tools-api/clientip/clientip.go` (**07-30**), `platform/orchestration/actions/store_generated_component_action.go`
(**08-03**), `check_endpoint_health_action.go`, and `register/rebuild-cascade.md` — the REB-003
rewrite that four consecutive handoffs recorded as "stalled, do not touch", dirty since
**2026-08-08 20:41**. Untracked files survived (a hard reset does not touch them), which is why
`clearideas.bash`, `live.html` and `ctacalibrate` are still there and can mislead you into
thinking nothing happened.

**Two things follow.** First, the standing rule "commit each task the moment it is coherent,
narrowly" is not merely about tidy history — the window between writing and committing is a
window in which another session can delete your work outright, and here it was about ninety
seconds. Second, **`rebuild-cascade.md`'s stored count is no longer owed**: the WIP that blocked
it is gone, so the drift check's last HEAD finding is now removable by whoever next touches that
file. That is a bad way to become unblocked and it should be said plainly to its owner rather
than quietly banked.

**Addendum, same session — my own "did it land?" check was a false positive, in the day's own
shape.** Verifying the restored files at HEAD, I grepped `LANDMINES.md` for
`rev-list --objects` and got a match, so I recorded the landmine as safely committed. **It was
not there at all** — the match was a pre-existing entry from another lane about the same
command's `%(rest)` trap. A second, discriminating grep (for the entry's own heading) found
nothing, and the entry had to be written a third time. **A verification pattern that a file
could satisfy without your change is not a verification** — and this one was the same failure
the whole session is about, committed while writing it up: I tested a tidied-up proxy for the
claim rather than the claim. The `doc_notes` rows were fine throughout, because the sync ran
before the reset — so `landmines-sync.py --check` reporting clean would ALSO have reassured me
wrongly. Only the file is the system of record.

> **CORRECTED 2026-08-12, same session — it was a `git stash`, not a hard reset, and NOTHING
> WAS DESTROYED.** The entry above says another session's reset "discarded" every uncommitted
> change and that the losses are unrecoverable. **Both halves are false.** `git stash` performs
> a `git reset --hard` internally and writes exactly the reflog line I read as evidence —
> `reset: moving to HEAD`. The work is in **`stash@{0}`** ("WIP on 087_towards_multiple_domains:
> 1ee940968"), 38 files, including all five of the documents I re-wrote **and**
> `register/rebuild-cascade.md`'s REB-003 changes.
>
> **What caught it:** another session had already diagnosed it correctly and committed a
> landmine about it (`4e0de34ec`, "a bare `git stash` by another session deletes YOUR
> uncommitted work, and git status reads clean") — I saw the subject line in `git log` while
> checking my own commits had landed. **The check that would have cost one command: `git stash
> list`.** I inferred a destructive mechanism from a reflog line, escalated it to
> "unrecoverable", and put it in three documents and a commit message, without running the one
> command that distinguishes the two mechanisms. A reflog line names an OPERATION, not an
> INTENT, and several porcelain commands share this one.
>
> **Two consequences.** (1) `rebuild-cascade.md`'s stored count is **still owed** — the WIP that
> blocks it is stashed, not gone, and its owning session can restore it, so the "do not touch"
> instruction in every handoff still stands. My note above that it is "no longer owed" is
> withdrawn. (2) **Do not `git stash pop`** to check any of this: that would drop 38 files into
> a shared tree, five of them files I have since committed, and the conflict would be mine to
> explain. `git stash show --name-only stash@{0}` answers the question read-only.
>
> The re-writing was not wasted (the content is committed and verified at HEAD), but it was not
> necessary either, and the wrong diagnosis is the more expensive half: it is now in three
> commit messages that cannot be amended.

---

## 2026-08-12c — "if clean, do it" fired on an ACCIDENT, and the owed line survives with a sharper check

Picked up the lane from `HANDOFF_2026-08-12b_continue_here.md` at 18:37 BST, four minutes
after it was written. Its owed bullet says `rebuild-cascade.md` is *"still owed, FIFTH
session running … another session's REB-003 rewrite still dirty in the shared tree … mtime
2026-08-08 20:41, 3 added / 3 deleted"*, and instructs: **"Re-check `git status` before
assuming; if clean, retire this line."**

**I re-checked, and it WAS clean** — `git status --short` empty, `git diff` empty, mtime
changed to 18:38. By the instruction as written, the blocker had cleared and the owed work
was mine to do.

**It had not cleared. The file was clean because a session ran a bare `git stash` at
18:38:51**, which reverted every dirty tracked file on this shared tree — 38 files across
about ten lanes, that rewrite among them. Restored from `stash@{0}` and verified per file;
the diff is back to the same 3 added / 3 deleted, and it is real work (a `what:` expansion
for `REB-003`, a `bugs_open/182 → bugs_closed/182` citation, a `verify-later:` update). Had
I followed the instruction I would have deleted `rebuild-cascade.md` from
`KNOWN_STORED_COUNTS` on the strength of an accident, and reported the blocker gone.

**The hole is that "clean" has two causes and the check cannot tell them apart:** the other
session committed (the blocker really is gone), or the other session's work was swept out of
the tree (the blocker is very much still there, and now invisible). On a tree this many
sessions share, the second is not exotic — it happened within four minutes of the handoff
being written.

**The corrected check asserts the POSITIVE fact, not the absence of a diff:**

```bash
git status --short <path>                    # necessary, not sufficient
git log --oneline -1 --date=iso --pretty='%h %ad %s' -- <path>   # did the work LAND?
```

Clean **and** a commit newer than the stall date means it landed. Clean **and** a last
commit still older than the stall means the work has gone somewhere — look for a stash
(`git stash list`) before concluding anything. The same asymmetry the lane already knows from
elsewhere: an absence is only evidence when you have established what would have made it
present.

**So the owed line stands, unchanged in substance:** `rebuild-cascade.md`'s stored count is
still owed, still same-file-blocked, and it remains the drift check's only HEAD finding
(`./scripts/test-concept-register-drift-local.py`, run at HEAD `287cdffe2`: 1842 entries,
1842 index rows, one finding — `rebuild-cascade.md` states 7, actual 7).

Nothing else in this session touched the register. The two landmines filed
(`b3aa8c45c`, `19eb8fdf8`) are fleet-wide, not lane work — though the second is the lane's
own subject one layer up: an advisory that is delivered but keyed so that nothing can match
it has not reached anyone.

---

## 2026-08-14 / 08-16 — the stash ban lands, the landmine keys are repaired, and the verifier is finally fired

Two owner-directed pieces and one carry-through, across two sittings.

### 1. `git stash` forbidden and mechanically blocked (owner ruling 2026-08-14, `371317eb6`)

Owner asked for the ban in CLAUDE.md and "can we ban it in .git?". **The honest answer to the
second is no**: git has no pre-stash hook, `.githooks/` cannot see a stash coming, and any
filesystem trick against `.git/refs/stash` would corrupt the very stash we are still recovering
from. The enforceable layer is the session harness — every actor on this tree is a session — so
the ban is a `PreToolUse` hook, `scripts/block-git-stash.py`, wired in the versioned
`.claude/settings.json`. Denies every MUTATING form (bare/push/save/pop/apply/drop/clear/branch)
in any compound shape; `git stash list`/`show` and `git show 'stash@{N}:<path>'` stay allowed
because they are the documented recovery. 14/14 self-test, pipe-tested on the real stdin shape,
then **proven live**: a real stash against a scratch repo was denied in-session, the scratch
repo showed 0 stashes afterwards. Fails open on malformed input but a crash of its own surfaces
— a silently dead safety gate is what this fortnight already demonstrated once.

Evidence the ban holds: `git log -g refs/stash` — the newest stash is still 2026-08-12 18:38:51,
across 680 commits since (checked 08-16 11:01).

**Same commit carried the owner's 08-12 "explaining decisions" CLAUDE.md note**, which had sat
uncommitted — and therefore stash-sweepable — for two days. Named in the message.

**The overlays question closed itself**: by 08-14 the release flow had already rewritten the
working-tree overlays to the live tag (fleet, makefile, tree all `v1.0.1299`; `v1.0.1303` by
08-16), so there was nothing to protect by committing them. Owner ruled the same.

### 2. `split_footprints` fixed and `doc_notes` re-keyed (`f92e0b3ca`) — and the check changed the work twice

Owner asked me to verify the fix was right before doing it. It was, and the verification
reshaped it:

- **The defect had grown**: 59 → 63 `·`-collapsed entries in two days. The file's own convention
  was outrunning the parser.
- **A second defect of the same class, twice the size**: commas INSIDE parentheticals did split,
  so `` `path.go (FuncA, FuncB)` `` shipped as junk fragments — 143 entries. Total re-keyed:
  **185 of 482**, rows 2,452 → 2,398, and **zero entries without either defect were touched**
  (asserted on the full corpus, before/after).
- **My first paren fix was wrong** and the new 8-case self-test (`python3 scripts/landmines_lib.py`)
  caught it before it touched anything: it stripped `snapshot_agent(text, text)` to
  `snapshot_agent`, which `is_prose`'s own docstring records as a correct footprint. Fixed by
  requiring a SPACE before `(` for the qualifier strip — the discriminator between a
  trailing qualifier and a signature.
- **The sync's own detector could not have seen the repair.** `refootprinted` compared row
  COUNTS; 6 of the 185 entries changed key IDENTITY at an unchanged count and would have kept
  stale subject_keys for ever, every sync reporting clean. `existing_sources()` now returns the
  sorted subject_keys and the comparison is identity. Simulated against the live DB before
  applying: 185 rewrites, both same-count probes present in the set.
- **Zero verifier cost by construction**: refootprinted entries are excluded from
  `NEEDS_VERIFICATION` (body unchanged), and the dispatcher is manual.

Applied and asserted at the artefact: DB keys == parser output on all 482 entries, 0 mismatches,
0 strays, 0 middots; immediate re-run "nothing to apply". Delivery on identical inputs, old
parser vs new: nothing lost; discriminating probe gained the WireGuard entry firing for a file
under `deployments/kustomize/services/wireguard/` — a match the collapsed string could not
structurally make. **Honest caveat**: full-path dirty files often matched collapsed strings by
substring anyway; the fix's real wins are directory-prefix matching and exact-greppable
`doc_notes` keys.

### 3. Verifier fired (08-16) — three entries, one of them owed since 08-12

`--apply` on 08-12 armed two entries (`NEEDS_VERIFICATION`) that I never dispatched — the same
sync-before-dispatch trap the CLAUDE.md correction of 08-15 now names. Confirmed at the DB that
none of the three had a `landmine-verification` row, confirmed the chassis was outside its
300s spawn-drop window (pods up since 08-15 18:45Z), then fired
`trigger-landmine-verifier.sh` per entry:
- `a-shared-tree-git-stash-reverts…` → `4dd05e8a-f1ae-4306-8f8c-6d782ed125b6`
- `appending-a-landmine-with-this-file-s-own-separator…` → `ef045a9a-8da6-40d1-8a3c-84087882f8be`
- `git-commit-paths-silently-commits-fewer-files…` → `52b70a74-7b69-4f8b-aefe-b92d69fcaea6`
Arrival to be proven at `orchestration_states`, not from the print (`kcat -P` can drop at exit
0). See the handoff for the read-back query.

### 4. The fresh chassis roll (`v1.0.1303`) does not touch this lane

Checked rather than assumed: none of `block-git-stash.py`, `landmines_lib.py`,
`landmines-sync.py` appears in any `build/docker/backend/*.dockerfile`. All three are
harness/tooling; nothing here ships in an image.

### 5. Register state at 08-16 11:01 — a NEW finding, and it is the predicted one

`test-concept-register-drift-local.py` at HEAD `88897190e`: **1867 entries, 1868 index rows,
2 findings.** The old one (`rebuild-cascade.md`'s stored count — sixth session running, its
REB-003 rewrite still dirty, still same-file-blocked, last commit still 07-27) and a new one:
**`PUB-005` — an index row with NO register entry.** The mechanism is exactly the 08-12b
handoff's landmine: the row rode into HEAD as a passenger in another lane's commit
(`88897190e`, the `286`/TL-044 commit, 11:00:35 — one minute before my check) while the entry
sits complete and DIRTY in `register/public-api.md` (+13/−2). Owner is live (transcript
`f15c870e` writing at 11:02). **Not mine to commit** — flagged in the handoff; the owning lane
closes it by committing `public-api.md`. It is a transient window, but it is the exact state
OPP-006 exists to prevent, and it is now the register's second HEAD finding.

**Postscript, same sitting — verdicts read back.** All three spawned within 20 s of dispatch
(`orchestration_states` rows 10:03:19–10:03:50Z, COMPLETED) and the verdicts landed at
10:03:44 / 10:04:02 / 10:04:05Z: **UNVERIFIABLE ×3**, each stating the entry is "internally
consistent" and that every footprint lies outside the Go-only `code_symbols` index. That is the
`verify_unverifiable` branch behaving as documented (WRONG_CALLS 08-16, first entry), not a
refutation; nothing objects. **Two corrections to my own first draft of the handoff's read-back
query**, both caught before commit: it used `fix_correlation_id` (the COUNCIL-GATE key; this
trigger puts `correlation_id` in the envelope headers and `input_data` is `{source, ref}`), and
the jsonb-path scan hung >120 s on `orchestration_states` — keyed on `source` + a time bound
instead. Verified against `scripts/trigger-landmine-verifier.sh`, not memory.

**Postscript 2 — `PUB-005` closed itself, and a second transient came and went.** Its owner
committed `public-api.md` at 11:06:12 (`f967d9307`), six minutes after the row rode out.
Between two consecutive drift runs minutes apart the count went 1867/1868 → 1868/1869 → clean:
a *different* row-without-entry appeared and disappeared as another lane's pair of commits
landed in the wrong order and then the right one. Three live lanes were touching the register
this morning; the harness reads a fresh HEAD each run, so a single reading is a snapshot of a
window that may already have shut. **Rule for reading it: a row-without-entry finding younger
than ~10 minutes is a lane mid-commit, not a leak — re-run before filing anything.** The 08-12b
handoff's landmine ("expect the index to ride out under another session's commit") is now
measured at 2 occurrences in one morning; both self-healed. That is data for OPP-006's
enforcement question, and it points both ways: leaks are real and frequent, AND the repair
arrives in minutes without a gate. HEAD `0d1cac6cc`+: 1 finding, `rebuild-cascade.md`, still owed.

---

## 2026-08-17 — fresh cold-start handoff, and two corrections found while grounding it

Wrote `HANDOFF_2026-08-17_continue_here.md` (supersedes 08-16, which supersedes 08-12b) so a new
session can start cold. Re-measured every figure rather than carrying any forward — HEAD
`7d832ebc8`, re-confirmed at `b4db98f0b` twenty minutes later, 244 commits on from my last sitting.

**Everything from the last three sittings is holding, measured not assumed:**
- **The stash ban**: no new stash in **650 commits since the ban** (`371317eb6`), 1,043 since the
  stash itself. `stash@{0}` is still the 2026-08-12 18:38:51 one.
- **The landmine re-key**: 546 entries file-side, 546 DB-side, **0 key mismatches, 0 strays, 0
  `·` left in any `subject_key`**, 2,758 rows. **64 entries have been added by other lanes since
  the 08-14 splitter fix and all are keyed correctly** — so the fix holds for new arrivals, which
  the backfill alone could not have shown.
- **Register**: 1880 entries / 1880 index rows, agreeing exactly; **1 finding**, the perennial
  `rebuild-cascade.md` stored count (still dirty +3/−3, last commit still 07-27 — seventh session).
- **Verdicts**: UNVERIFIABLE ×3, stable, no objections.
- **Citation health is stable against a growing corpus**: 8,279 citations / 1,806 entries, still
  **75%** resolving as written and still exactly **4** never-existed, against 08-12's 7,793 /
  1,767. The corpus grew ~490 citations and the ratio did not move.
- **Version lag has a real worklist now** and its controls pass: newest citation `v1.0.1305` at
  lag 0 (so the live version resolves correctly), field-keying excluding 101 of 345 citations
  (29%, so the key is doing work), and **77 entries ≥50 releases behind on `status-evidence`**.

### Correction 1 — "the branch is unpushed" was wrong, and both prior handoffs carried it

`git status -sb` shows no upstream and `git rev-parse @{u}` errors, which I had been reading as
"nothing is pushed". **That is a LOCAL tracking-config fact, not a statement about the remote.**
Asked the remote directly: `git ls-remote --heads origin 087_towards_multiple_domains` returns
`896c5aeeb`, and `git rev-list --count 896c5aeeb..HEAD` is **66**. So the branch IS pushed and is
66 commits behind — which matters, because the drift CronJob reads the *pushed* branch, so its
morning row is real but lags HEAD by hours rather than being absent. Corrected in the new handoff
and flagged on the 08-16 banner. Same shape as this lane's other absence errors: I interrogated a
proxy (local config) for a claim about a remote, and the proxy answered a narrower question.

### Correction 2 — my own self-test passed VACUOUSLY on its first run, and said so

The handoff first carried the key-identity assertion as a pasted heredoc. Two problems: the
heredoc terminator was indented, so copy-pasting it verbatim would hang rather than run; and a
check this reusable should not live as prose. Promoted it to
**`scripts/landmines-keys-check.py`** — read-only, exits 1 on mismatch, and it exists because
`landmines-sync.py --check` **cannot** settle key identity (its drift test is `new or gone`, which
is exactly how six of the 185 re-keyed entries hid at an unchanged row count).

Its `--self-test` mutates a copy of the corpus and requires the mutation to register. **On the
first run it reported FAIL — and it was right.** My mutation inserted a footprint at the first
`- **footprint:**` line in the file, which sits in the PREAMBLE, above the `# Entries` marker that
`parse()` starts from. The mutation was a no-op, so an unmutated and a mutated corpus scored
identically. Had the self-test simply asserted "no mismatches" it would have passed and certified
nothing. Fixed the mutation (anchor below `ENTRIES_MARKER`), not the assertion: now 0 problems
unmutated, 1 mutated, PASS. This is the `mutate-the-code-to-prove-the-guard` rule catching its own
harness, and it is worth recording that the failing run was the valuable one.

**Deliberate deviation, stated:** I edited the `SILENTLY INERT` landmine entry to point at the new
script, then ran `landmines-sync.py --apply` **without** `landmines-verify-dispatch.sh`. That
consumes the entry's "changed" status so no verifier run fires — chosen on purpose: this entry's
footprints are wholly non-Go and unchanged, it was verified on 08-16, and the verdict would be a
byte-identical UNVERIFIABLE for the cost of a fleet run. Any genuinely new entry still gets its own
`NEEDS_VERIFICATION`.

---

## 2026-08-24 — INCOMING from another lane: an unregistered shared mechanism, and DBI-014 has drifted

**Not this lane's work and not written by its owner.** Left here because it is register-shaped and
because the lane is in a quiet state with a ranked list — this is an item for that list, to take or
decline, not a claim on the lane. Everything below carries the command that produced it and the date
it was run; re-run before repeating any of it outward.

**Where this came from.** Migration `566`
(`docs/agent_docs/sql_for_agents/566_database_cleanup_reaps_every_terminal_status.sql`, commit
`ccc851a42`, applied 2026-08-23 17:46Z) fixed a leak in the `database-cleanup` scheduled task. While
verifying it I went looking for the register's account of the mechanism and could not find one. The
full account of the fix lives in `bugs_open/354_HANDOFF_2026-08-22_a_workflow_that_ends_at_its_error_terminal_is_recorded_COMPLETED_with_error_NULL.md`
and in `docs/agent_docs/docs024_key_docs_latest/orchestration_status_lifecycle/RUNBOOK_orchestration_status_lifecycle.md`
— **do not re-derive it from here**; this note is only the register-facing part.

### 1. `orchestration_status_vocabulary` is not in the register at all

**What the thing is, plainly.** It is a small table, created by migrations `465`/`466`, that lists
every legal value of `orchestration_states.status` — one row per status, with two boolean columns,
`is_terminal` and `is_pausable`. Since `466` the status column has a foreign key to it, so adding a
new orchestration status means inserting a row here first. Seven rows as of 2026-08-24.

**The bar, as CLAUDE.md states it:** *another workstream could call this and would not know it
exists.* It is met — any lane adding an orchestration status must write to this table, and the FK
means it cannot avoid it.

**The absence, measured 2026-08-24** (from `docs/agent_docs/docs026_concept_register/`):

```bash
grep -ril "orchestration_status_vocabulary" register/   # -> no files
grep -ril "is_terminal" register/                       # -> no files
grep -ril "is_pausable" register/                       # -> no files
grep -in "vocabulary\|orchestration_status\|is_terminal" 102_coverage_ratchet.txt   # -> no hits
```

So it is neither registered nor ratcheted. `"status vocabulary"` does match several register files,
but every hit is a different thing — `sites.status` (DBI-018), work-item statuses, adapter response
statuses — which is exactly the kind of near-miss that makes a `grep` read as covered when it is not.

**Why it is worth an entry now rather than whenever.** The table did not merely go unregistered; on
2026-08-23 it acquired a **new guarantee**. Before `566`, `is_terminal` was a lifecycle/dispatch
marker. After it, `database-cleanup`'s arm 3 deletes `status IN (SELECT status FROM
orchestration_status_vocabulary WHERE is_terminal)` — so **`is_terminal` is now a deletion
predicate**, and marking a status terminal also decides its retention (24h after `updated_at`). A
reader of the table today has no way to learn that from the register.

There is a live hazard attached, raised as a medium advisory objection by the council's `guardian`
seat on `566`'s own round (correlation `9d23ccd9-c16c-422d-8bf9-7b60e8b52795`, verdict APPROVED):
nothing prevents a row being **both** `is_terminal` and `is_pausable`, in which case arm 3 deletes it
while arm 4 deliberately spares it. Zero rows are both today, and no pause/human-named status is
terminal — I ran the seat's check — but there is no CHECK constraint, so it is prospective. It is
written up in `docs/agent_docs/docs024_key_docs_latest/LANDMINES.md` ("Setting `is_terminal` … now
ARMS a 24-hour DELETE"), which is the landmine home, not the register one.

**I did not file the entry myself**, for two reasons worth stating rather than leaving implied: the
mechanism is the `465`/`466` lane's, not mine, and CLAUDE.md's register bar explicitly excludes bug
fixes — `566` is a bug fix. Adopting another lane's mechanism into the register under my name is the
same move commit-per-task exists to prevent. Hence this note.

### 2. `DBI-014` (`register/database-and-infrastructure.md`) has drifted — partly

This is the entry that already covers `database-cleanup`, so it is where a reader would land. Its
claims split cleanly into still-true and no-longer-true, and the split matters more than a blanket
"stale" would. **All checked 2026-08-24 against the live `scheduled_tasks.pre_query` row**, not
against a migration file:

| DBI-014 says | measured 2026-08-24 |
|---|---|
| `awaited_requests 7 days` | **holds** — oldest row 2026-08-17, only 9 of 33,082 older than 7 days, and `cleanup_expired_awaited_requests` exists in `pg_proc` |
| `orchestration_requests FK made CASCADE` | **holds** — `pg_constraint` shows `fk_orch` with `confdeltype='c'` |
| `agent_error_log 14/30 days` | **drifted** — the live arm uses `INTERVAL '30 days'` and `INTERVAL '365 days'` |
| `orchestrations 7 days/24h stuck` | **drifted** — the live sweep contains **no 7-day interval at all** (`grep -c "7 days"` on the dumped `pre_query` returns 0); both orchestration arms are 24h |
| *"A uniform retention discipline"* | **was false until 2026-08-23** for one class: a status that was terminal but not named in arm 3's literal pair was reaped by *nothing*. `CANCELLED` sat in exactly that position for 35 days (24 rows, oldest 2026-07-19). `566` is what finally made the word "uniform" true |
| *"always mark itself executed (the 'always-return-a-row HAVING fix')"* | **contradicted by this register's own `SCH-007`**, which carries a 2026-08-17 correction stating the always-return-a-row rule is no longer true and that believing it causes the opposite defect. Two entries in two files now disagree |

Reproduce the whole table with one dump and then read it, rather than trusting the above:

```bash
kubectl -n ai-persona-system exec -i postgres-clients-0 -- psql -U clients_user -d clients_db -At \
  -c "SELECT pre_query FROM scheduled_tasks WHERE name='database-cleanup';" > /tmp/sweep.sql
grep -n "INTERVAL\|deleted_" /tmp/sweep.sql
```

⚠ **One trap if you verify by hashing:** `length()` on that column counts CHARACTERS while `md5()`
hashes BYTES, and the row holds a multi-byte character, so a locally-computed md5 of a `psql` dump
does **not** match `md5(pre_query)`. Compute hashes in the database. (Known family — `LANDMINES.md`
carries it as "`length()` on stored HTML is CHARACTERS".)

**The cross-file disagreement in the last row is probably the most interesting part for this lane**,
because it is not staleness in the sense the two shipped reports detect. `SCH-007` was corrected and
`DBI-014` was not, so both entries are individually well-formed and internally consistent, and each
cites its own evidence. Nothing that keys on version lag or missing citations would surface it —
what makes it findable is that the two entries make opposite claims about the same mechanism. Offered
as a possible worked case for the lane's open "are the reports actionable?" question (handoff item 2),
since it is a real discrepancy someone could try to repair and time.

**No action is claimed or owed here.** If the lane takes it, item 1 is a new entry and item 2 is a
correction to an existing one; if it declines, this note is still the record that the gap was seen on
2026-08-24 and by whom.

**ADDENDUM, same day, 2026-08-24 — the `agent_error_log` drift now has a named cause, found by
re-reading the live row an hour later.** The row changed again at 12:31Z (another lane; my own
`566` predicate is untouched and both arms still read the vocabulary — checked, not assumed). The
new text carries a comment block explaining the retention that DBI-014 records as `14/30 days`:

> *"Retention is BY FINDING CODE since migration 567: codes in the list below … expire at 30 days;
> EVERY OTHER CODE LIVES 365 DAYS, because a deliberate finding outlives the plumbing it shares this
> table with. `resolved` does NOT shorten a row any more (it used to halve it to 14 days, which was
> backwards …)"*

So DBI-014's `14/30 days` was **correct when written** and was deliberately superseded by migration
`567` (`docs/agent_docs/sql_for_agents/567_finding_codes_outlive_the_plumbing.sql`, the
`bugs_open/358` lane). That is the more useful framing than "the entry is wrong": the entry did not
rot, a mechanism changed underneath it, and the lane that changed it documented the reasoning **in
the `pre_query` itself** rather than anywhere a register reader would look. Worth a glance when
repairing DBI-014 — the `resolved`-halves-retention premise is the part that inverted.

It also sharpens the note above: two of DBI-014's four checkable figures have now been superseded by
**two different lanes in two different weeks** (`567` for the log arm, `566` for the orchestration
arms), neither of which knew the entry existed. That is the register's own case for existing, and it
is a cleaner example than a single stale number.

**CORRECTION to the note above, 2026-08-24 — I said I would not file the entry, and then I filed
it. Stating that plainly rather than leaving the two in contradiction.** Item 1 above says the
mechanism belongs to the `465`/`466` lane and that the register bar excludes bug fixes, so the
entry was theirs to write. That reasoning held while I was only *fixing* the mechanism. It stopped
holding when the owner asked me to go ahead with the terminal/pausable guards (`589`), because I am
now *altering* what the mechanism guarantees — and the 2026-07-28 owner ruling's condition (2)
requires such a seam to be **registered in the same commit that ships it**, with its landmine and
open review question. That obligation is mine, not this lane's.

So `DBI-026` exists as of commit `9e0b0daa9`, in `register/database-and-infrastructure.md` with its
index row, no DBI drift. **What it does NOT do is close item 2** — `DBI-014`'s drifted figures are
untouched and still that entry's own repair, and `DBI-026` deliberately points at it in `relations`
rather than quietly restating it. **`DBI-026`'s status line says `589` is written but NOT YET
APPLIED**, which was true when written; if you are reading this later, check
`pg_constraint` for `chk_status_not_terminal_and_pausable` rather than believing the entry.

The change of circumstance is the only thing that changed. If the lane would rather own `DBI-026`
outright, or fold it into `DBI-014`, treat mine as a first draft written by the session that had
the mechanism open — not as a fait accompli.

---

## 2026-08-25 — `DOC-078`'s sharpest category had a false-positive mode (fixed, `a9665268f`)

Picked the lane up from `HANDOFF_2026-08-17_continue_here.md` and re-ran its whole verification
table first, which was the right call: HEAD had moved **1,481 commits** since that doc was written
(and a further ~1,100 by the time the fix landed, at `5c45e1fac`).

**What the re-run said**, each figure with its command in the handoff's table format:

| what | 08-17 | 08-20 re-run |
|---|---|---|
| register self-consistency | 1880 / 1880 | 1918 entries / **1919 rows** |
| landmine keys | 546 entries, 0 mismatch | 640 entries, 3314 rows, **0 mismatch**, exit 0 |
| stash ban | clean 650 commits | **clean 2,131 commits**; `stash@{0}` still the 08-12 one |
| fleet | `v1.0.1305`, 18 deploys, 3 tags | `v1.0.1317`, 20 deploys, **uniform** (makefile `v1.0.1319`, 2 ahead — a build bumped, not rolled) |
| citations (`DOC-078`) | 8279 / 75% / **4** dead | 8741 / 75% / **7** dead |
| version lag (`DOC-077`) | 345 cites, 182 entries | 400 cites, 208 entries; `status-evidence` median 111 → **123**, 82 entries ≥50 behind |

The 75% resolution ratio has now held across **three** measurements while the corpus grew ~950
citations. That is the useful reading and it is unchanged: the citations are abbreviated, not rotting.

**The finding.** `NEVER-REPO-PATH` went 4 → 7, and not one of the three was rot.

- **Two were `SEO-005`**, citing `platform/orchestration/actions/head_assembly.go`. The file existed
  on disk and was git-**staged** (`A `) but uncommitted. The report reads the WORKING TREE for
  entries but resolves paths against git HISTORY — so an added-but-uncommitted file reads as
  "no file, ever". A harness asymmetry, not a citation defect. **Recorded on 08-20 as a prediction
  that it would self-heal; it did** — the SEO lane committed it in `4abcd55a4`, and by 08-25 both
  citations resolve. Same root as the drift check's `SEO-005` row-without-entry finding the same
  morning (index row committed as a passenger in a `PBP-040` commit, entry half still dirty).
- **One was a genuine resolver bug.** `FIX-061` cites
  `scripts/migration/run-migrations.sh:65,:283`; the file is tracked and clean.
  `clean()` stripped line refs with `:L?\d+([-,]L?\d+)*$` — which handles `:151,227` (there is a
  self-test case for exactly that form) but **not the repeat-colon form**. On `:65,:283` it stripped
  `:283`, left `:65,`, and the trailing `rstrip` made it `…run-migrations.sh:65`. Fix is one
  character class: `([-,]:?L?\d+)*`.

**Why it was worth doing at a blast radius of one.** `grep -rhoE '[A-Za-z0-9_./-]+:[0-9]+(,:[0-9]+)+'`
over `register/` returns **1** occurrence as of 2026-08-25. The volume is not the argument. The
argument is that the last three handoffs used *"NEVER-REPO-PATH is still exactly 4"* as the
corpus's stability evidence, and that number had a way of being wrong that no reader would have
read as wrong — a false positive in the one category the report states as decisive.

**Guarded, not just fixed.** A self-test case sits beside the one it generalises. It is
**non-vacuous, and proven so by mutation**: reverting the regex on a *copy* while keeping the case
makes that case fail with exactly `NEVER-REPO-PATH` (10/11, exit 1), reproducing the false positive;
the other ten are unaffected either way. This lane has been bitten by a vacuous self-test before
(08-17: `parse()` starts at the `# Entries` marker, so a mutation above it registered nothing) —
mutating a copy and watching the *specific* case fail is the cheap version of not repeating it.

**Missteps this sitting**, since that is what this file is for:

- I printed `MUTANT EXIT=0` and nearly read it as the script's exit code. It was `tail`'s — the
  pipe masks it. Re-ran unpiped: mutant 1, fixed 0. **A `$?` after a pipeline is the LAST command's**,
  and every self-test invocation here ends in `| tail`, so this trap is permanent in this lane's
  own runbook commands.
- First instinct on seeing 4 → 7 was "the citations are rotting". They were not; two of three were
  my own harness looking at an uncommitted file, and I only avoided filing that by checking
  `git status` on the cited path. **A dead citation naming a file that exists on disk is an
  uncommitted-file transient until proven otherwise** — check `git status <cited path>` before
  believing `NEVER-REPO-PATH`.
- On landing the fix I had to resist reporting "7 → 4, fixed". **Only one of the three was mine.**

**Not in council-gate scope**: `COUNCIL_SCOPE_CODE_RE` anchors `scripts/pattern-check\.py$` with a
trailing `$`, so no submission was owed for this file. Checked, not assumed.

**Still owed, EIGHTH session running:** `rebuild-cascade.md`'s stored count. Last commit still
`7272d59d4` (2026-07-27), still dirty +3/−3, still same-file-blocked. Asserted the positive fact
per the handoff's own warning rather than reading a status as resolution. It is once again the
drift check's ONLY finding.

### 2026-08-25 (later) — logging the landmine, and five things that went wrong doing it

Entry appended to `LANDMINES.md`: *"A `NEVER-REPO-PATH` citation naming a file you can `ls` is an
UNCOMMITTED file, not a dead citation"*, slug
`a-never-repo-path-citation-naming-a-file-you-can-ls-is-an-uncommitted-file-not-a`. **Delivered
end to end** — 7 `doc_notes` rows (one per footprint), keys check **843 entries both sides, 0
mismatches**, verifier dispatched (`78ca5fa5`), spawned `call_verifier` → `verify_unverifiable`,
verdict **UNVERIFIABLE 19:45:35Z**. UNVERIFIABLE is the branch WORKING: the verifier's
`code_symbols` index holds Go only and these footprints are a Python script plus doc paths.

The entry itself was cheap. Getting it *delivered* took five corrections, and they are the reusable part.

1. **I got the heading level wrong, and my sampling method is why.** The format is `###`; I wrote
   `##`. I had sampled the convention with `grep -n '^## ' | tail -5` and read the results back as
   house style — but **that grep selects FOR the non-conforming minority and cannot show the
   majority it does not match**. Actual split: **717 `###` against 148 `##`**. The entry still
   parsed (7 footprints, correct slug), so every check I had run passed; the heading is a separate
   contract from the parse, and `landmines-sync.py` reports it in a block titled *"163 warning(s)
   that cost DELIVERY"*. Fixed in `358a4ae4a`. **The one-command check:
   `grep -c '^## ' <f>; grep -c '^### ' <f>` — compare the two, never sample one.**
2. **The footprint check earned its keep.** My first draft's last footprint was a **103-character
   prose string**; parsing the entry and asserting a LIST of short targets caught it, exactly as
   this lane's own landmine says to (test for MATCHING, not for syncing). All 7 now ≤41 chars,
   both path-shaped ones substring-match real files (1 and 339 hits), no globs.
3. **`to insert/refresh: 842` is NOT the delta — it is `len(want)`, the whole corpus** (line 224).
   I read it as evidence that the sync was resending everything and had nearly written that up as
   the cause of the transport failures. Disproved it in one query: a stable entry's body was
   **byte-identical in DB and file (same md5, 1245 bytes)**, so it was not in the delta at all.
   **A count printed next to the word "delta" is not necessarily the delta — read the print.**
4. **A `psql failed:` from this sync may have ALREADY COMMITTED its write.** Five dispatch attempts:
   four printed transport errors (i/o timeout, connection reset, and twice `unexpected EOF` — the
   *verbatim* error the script's own comment records as its scale limit, hit at ~2,155 rows and now
   at **4,645**), the fifth printed `nothing to apply — already in sync` **and
   `Nothing needs verification`**. The rows were already there. So an earlier "failure" wrote
   successfully and then died on a later full-corpus READ (`existing_bodies()`), **consuming the
   new-entry status without dispatching** — CLAUDE.md's documented trap, fired for real by a
   partial failure rather than by running `--apply` alone. Remedy is the one CLAUDE.md names:
   `./scripts/trigger-landmine-verifier.sh '<source key>'`. **Never read a transport error from
   this script as "nothing happened" — query for your rows before retrying.**
5. **My entry was swept into another lane's commit, and I took two passengers of theirs.** The
   entry landed in `c2f65f287` (the `bugs_open/364` lane's close-out), whose message does not
   mention it — the documented direction that per-task pathspec commits cannot protect against.
   Nothing was lost and the *tightened* version is what landed (verified by re-parsing HEAD's copy).
   In the other direction my two commits carried another lane's in-place RETIREMENT of the
   claims-number-scan entry and a `bugs_open/392` addendum. The first I declared in the message; the
   second arrived between my `--numstat` gate and the commit **seconds later** and I found it only
   afterwards. **On this file the gate and the commit are not atomic — re-read the diff inside the
   same command as the commit, and expect to declare a passenger anyway.**

Also worth keeping: the pattern check fired `shared-ledger-not-appended` on `358a4ae4a` (1 line
removed from an append-only ledger). **True on shape, benign in substance** — the removed line was
my own `##` heading being replaced by its `###` form, confirmed by reading the diff rather than the
count. And the pod-vs-git timezone trap nearly bit again: the spawn read `19:45` against a dispatch
at `20:45` local, which is UTC vs BST and not a missing spawn.

**Recommendation, NOT done (would be scope creep on a "log the landmine" task):** point 4 is itself
a landmine-shaped trap — *a `psql failed` that has already committed, leaving the entry synced,
the status consumed, and the verifier silent.* Its wrong result looks exactly like its right one.
Worth its own entry, with the check being "query for your rows before you retry".

### 2026-08-31 — item 1 built, measured and shipped (`1efc84362`, `7006eb2d8`); the measurement was wrong twice first

The sha-citation authoring rule is live as `check_register_roll_claim_without_commit`,
the 24th check in `scripts/pattern-check.py`, registered as **OPP-011**. Council
`Council-Submitted: 37b0bec4-f503-4b9a-8fc4-688ba29aa2bc` (that file entered gate scope
2026-08-24, so a submission was owed; admission tested free with `DRY_RUN=1` first).

**The rule:** a register status conditional on a roll must name the commit carrying it.
`[MEASURED 2026-08-31]` at HEAD `028c3e112`: 2,027 entries, 125 roll-conditional statuses,
24 withdrawn via strikethrough, 101 live, **28 (27%) naming neither commit nor version tag**.
The handoff's long-standing figure was "13 of 29 entries examined" — a sample, from 08-17
and ~3,700 commits stale. The census replaces it.

**Three design decisions carry the false-positive rate, and only one of them was obvious.**

1. *Strip `~~strikethrough~~` before testing.* The register withdraws a claim by striking
   it and appending the correction, so `~~inert until the next roll~~ → LIVE` still
   contains the trigger phrase while asserting its opposite. 24 of 125 are this shape.
2. *A version tag counts as an answer.* "LIVE on chassis v1.0.1322" dates a claim as well
   as a sha does — 4 more wrong fires gone (SEO-003, SYS-092, PLAN-027, LNK-034).
3. **A hex token counts only once `git cat-file -e <tok>^{commit}` RESOLVES it.** This is
   the load-bearing one and I nearly shipped without it. A council correlation id is 8 hex
   characters with digits and is *pattern-identical* to a short sha; `LNK-040` carries
   `corr e9bda035` today. **Regex-only, the fire rate is 2% and seven of the eight real
   cases are silently exempted** — a check that reads as quiet and useful while missing
   most of its population. With the git lookup: **8 of 45 register-touching commits (17%)**,
   0 false positives on inspection.

**Two ways the measurement was wrong before it was right. Both are the point.**

- **The first fire-rate run reported `0/45`, and the zero was MY HARNESS, not the check.**
  I matched findings with `grep -o "^   .roll-claim-without-commit"`, assuming one
  character between the indent and the name; the ANSI prefix is `ESC[1m` — four. A clean
  zero, a plausible story ("advisory checks should be quiet"), and completely false. It was
  caught by a **demand control in the same loop body** — running a known positive
  (`bf1fbc5b7`) through the identical pipeline, which returned 0 and could not have. The
  estate's rule (*a post-fix ZERO needs a DEMAND control*) earned its keep; note the control
  has to run through the **same loop body**, because it was the loop, not the check, that
  was blind.
- **The correlation-id hole was found by verifying the ONE fire rather than banking it.**
  At 2% I had a single hit (`LNK-040`) and every incentive to call it a true positive and
  stop. Reading its status is what exposed that `e9bda035` was exempting entries on a token
  that answers nothing. **A low fire rate is a hypothesis about the world OR about your
  filter, and only inspection tells you which.**

Also worth keeping: the fire I did verify was a true positive *at its commit* but would not
fire today, because a later edit added the corr id. **A check's verdict is a fact about a
commit, not about the file** — re-deriving it at HEAD would have made a correct fire look wrong.

**Not done:** the `_RELOCK` migration-suffix warning `scripts/council-scope.sh` prints is
another lane's (`bugs_open/314`), left alone.

### 2026-09-03 — council APPROVED; both objections verified before acting; and the roll makes item 1's point for it

**Verdict read** (owed since 08-31): **APPROVED round 1**, 2026-08-31 13:08, 2 advisory
objections, none high-severity. `Council-Reviewed: 37b0bec4-f503-4b9a-8fc4-688ba29aa2bc`.
The `Council-Submitted:` trailers on the earlier commits are credited automatically by `098`,
so no amend was needed — which forward-only forbids anyway.

⚠ **The orchestration row was GONE** (terminal rows reap at 24h; it had been three days). The
verdict lives in `doc_notes` / `diagnosis_artifacts`, which is where to look after a day.
And my first two attempts to read the objections returned *nothing at all* because I piped a
`column "content" does not exist` error through `grep` — **a hard SQL failure and "no
objections found" are the same output through a pipe.** `\d diagnosis_artifacts` first, as the
rules say; the column is `body`.

**Both objections were verified before acting, and they landed differently.**

- **`guardian` (medium): harden the `git cat-file` subprocess. Right conclusion, WRONG REASON.**
  Its ground was that a throw could abort the shared pre-commit run *"fleet-wide, every
  session"*. It cannot — `.githooks/pre-commit:49` runs `pattern-check.py || true`, and the
  seat's own `missing` block admits it never confirmed this and was inferring from landmine
  text. What a throw *does* do is abort the script mid-run so **the other 23 checks report
  silence**. That is a real degradation, so the hardening went in (`7dbe5b8fd`) on the
  corrected rationale: `try/except` + `timeout=5` + an 8-token cap that **fails OPEN**.
  **Proven by inducing the failure** — `PATH` stripped so git is unfindable → returns False
  instead of throwing. An untested `except` is not a guard.
- **`prior_art_librarian` (medium): BLD-019 says provenance is "INOPERATIVE on agent-chassis",
  so the payoff is overstated. A MISREADING, and I checked rather than believed it.** BLD-019
  says the opposite — LIVE and READ BACK on agent-chassis **both replicas**, all 14 backend
  services stamped. The *"unsafe as written"* text it quoted is about the `strings` recipe on
  debian-slim images, not the stamp. BLD-019 states this check's **exact premise**: *"The stamp
  only settles an entry that NAMES a commit — 13 of 29 surveyed entries cite none."* Recorded
  in the `OPP-011` entry so nobody re-opens it.
- **Three seats independently objected that the submission asserted sibling helpers
  (`REGISTER_ROOT`, `raw_diff`, `committed_content`) without citing them.** A fair hit on the
  SUBMISSION, not the code — the helpers exist and the check was built and fire-tested before
  I submitted. **The evidence existed and I simply did not put it in `grounded_in`.** Next
  submission: name the symbols you are reusing, the same way you name the concepts.

**The roll (owner, 2026-09-03, `v1.0.1356`) makes item 1's case better than any argument.**
Verified at the artefact, not from the statement: the tag CHANGED (`v1.0.1317` → `v1.0.1356`),
so the same-tag-cached-image trap does not apply, and chassis pods were 17–18m old. Post-roll
census: **136 roll-conditional statuses, 26 withdrawn, 110 live, 29 undated (26%).** So 110
entries assert they are waiting for a roll that has now happened; **81 can be settled with one
command each and 29 cannot be settled at all.** That is the population `OPP-011` stops growing
and does not shrink — it is now item 1 of the new handoff,
`docs026_concept_register/HANDOFF_2026-09-03_continue_here.md`.
