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

<!-- Append new turns below this line. Format: ## Turn N — date — one-line summary -->
