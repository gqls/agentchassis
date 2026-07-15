# PLAN — Concept Register (docs026)

**What this project is about (read this first).** agentchassis is an autonomous
agent platform that plans, builds, and operates a fleet of content websites. Its
`docs/` tree holds ~4,111 files of accumulated design docs, handoffs, runbooks,
and running notes spanning many workstreams and eras — rich but scattered, with
no single place that says "here is every mechanism/responsibility/behaviour in
this system, and whether it's actually real." The **concept register** is a
three-stage programme to build that place:

1. **Extract** every nameable concept from every doc file, tag it with a
   documentary status signal (deployed/partial/aspirational/superseded/
   abandoned/unknown), and classify it into an open-ended taxonomy of categories.
2. **Verify** each concept's documentary status against the live codebase and DB
   — the extraction's status is what the docs *claim*, not ground truth.
3. **Council** — build an expert agent per concept-area, versed in its
   responsibilities and provenance, to join the
   [[fixloop-workstream]]'s diagnosis/fix council (currently a 2-agent skeleton;
   see `docs024_key_docs_latest/fixloop_eg_dartsonline/SUMMARY_write_step_position_2026-07-12.md`).
   The register's category structure is designed to map onto council seats
   directly — that's the whole point of building it.

Companion documents:
- `RUNNING_NOTES_concept_register.md` — turn-by-turn discussion log, updated every turn.
- `RUNBOOK_concept_register.md` — the tasks the human operator performs (rotations, judgment calls, resumes).
- `README.md` — the extraction/consolidation method spec and status vocabulary (source of truth for format).
- `register/000_concept_index.md` — the master index, 1,627 concepts.
- `006_VERIFICATION_stage2.md` — full stage-2 method and findings.

Nothing outside `docs/agent_docs/docs026_concept_register/` is modified by this
work.

---

## Stage 1 — Extract (COMPLETE, 2026-07-13)

Swept all ~4,111 files under `docs/` via 34 extraction units (26 planned + 8
recovery/gap-fill), producing 2,185 raw concept blocks. Consolidated into 107
category files (`register/<category>.md`), 1,627 final concepts, one master
index (`register/000_concept_index.md`). Taxonomy started from a 30-category
seed but was kept open — extraction agents proposed 65 `NEW:` categories,
settling at 107. Full detail: `002_PLAN_extraction.md` (work-unit ledger),
`005_TAXONOMY_final.md` (taxonomy comparison + known limitations).

## Stage 2 — Verify (COMPLETE, 2026-07-14)

Ran as three Workflow batches: one agent per category work-unit follows each
concept's `verify-later` pointers into the live repo; every proposed status
correction gets an independent adversarial re-check before acceptance.

**Full scope covered, all 1,627 concepts:**
- All 314 `partial`/`unknown` concepts (the originally-planned scope; batch 2).
- All 871 `deployed` concepts, swept for false positives (added mid-stream —
  batch 1's hand-verification found the one true status error, MCL-002, in the
  `deployed` bucket, not partial/unknown; batch 2).
- All 102 `superseded` + 72 `abandoned` concepts (batch 3, completing coverage).

**Result:** 124 corrections confirmed total across all three batches (1 hand-verified
in batch 1 + 105 in batch 2 + 18 in batch 3), a ~7.6% error rate — with 106
proposed corrections overturned by the adversarial pass across batches 2+3,
showing the gate doing real work in both directions (see
`006_VERIFICATION_stage2.md` for examples). All applied directly to the
register files and master index. A 7th status, `convention`, was added to the
vocabulary in `README.md` for design doctrines/working practices that stage 1
had defaulted to `deployed` for lack of a better slot. One duplicate (PUB-001)
retired to a pointer entry.

**Two distinct failure-mode classes found**, worth remembering for any future
extraction/verification pass:
1. **Present-tense-plan misreading** (batches 1-2): a plan document narrating
   its own design in the present tense reads as evidence of completion.
2. **Search-scope gaps** (batch 3, e.g. DOC-064): a concept tagged `abandoned`
   because extraction's search never reached a sibling doc location holding a
   live, byte-identical copy — evidence that was never found, not misread.

Final status distribution: 853 deployed / 257 partial / 290 aspirational / 90
superseded / 67 abandoned / 21 unknown / 49 convention (was 871/274/271/99/72/
40/0 before stage 2).

## Stage 3 — Council agents (DESIGN GROUNDED, NOT YET BUILT)

Design a council-member agent per concept-area to join the fix-loop's
diagnosis/fix council. Each council agent should be "fully versed in its
responsibilities and provenance" per the original charter
(`001_PROMPT_charter.md`) — seeded from its category's register file (now
stage-2-verified) plus enough of the underlying source docs to answer
questions about that area with grounded, current confidence.

**The live mechanism this plugs into** (read directly from `register/fix-loop.md`,
FIX-014/015/020/036/043 — not invented): the council today is two sequential LLM
reviewer steps (`review_editquality`, `review_guardian`) inside the
`fix-proposer` agent's workflow definition (`fixloop_eg_dartsonline/0NN_fix_proposer.sql`),
aggregated by a deterministic Go action (`diagnose_council_decide`: hard veto →
rejected; any veto → rejected; any objection → revise; all approve → approved).
`hard_veto_from` is a plain config array, currently `["guardian"]`. **Adding a
reviewer is mechanically just: a new named step in that workflow JSON + a role
prompt + (optionally) an entry in `hard_veto_from`.** Both existing reviewers
currently get identical context — role prompt + fix plan + diagnosis + a live
DB schema hint — with "no per-reviewer curated corpus" yet (FIX-043, Q-G,
explicitly still open). FIX-036 itself says reviewer areas "are expected to
correlate with the docs024 documentation categories, the direct bridge to this
concept register's own council-agent goal" — i.e. the register was already
anticipated as the answer to Q-G.

**Design recommendation for the three open questions, resolved against that mechanism:**

- **Granularity — not 107 fixed reviewers.** 107 always-on synchronous LLM
  steps per fix-loop run would be a cost/latency non-starter, and most
  categories are irrelevant to any given bug (a CSS regression has no business
  consulting `canine-biology`). Recommend a **relevance-filtered subset per
  run**, not a fixed roster: given the fix plan's touched files/symbols/tables
  (the same signal `review_guardian` already inspects for blast radius), match
  against each category's `sources`/`verify-later` footprint to select maybe
  2-5 relevant category-reviewers for that specific fix, added as extra
  parallel steps alongside the existing two.
- **Activation — data-driven selection, not a symptom-keyword match.** Derive
  relevance from the fix plan's concrete artifacts (which files/tables/actions
  it touches), not the bug's English description — the register's own
  `verify-later` fields are exactly the right join key (they already name
  files/tables/actions per concept). This reuses the register as a lookup
  index rather than requiring new classification machinery.
- **Freshness — curated corpus for framing, live grep for the load-bearing
  claim.** Stage 2 found errors even in `deployed` (MCL-002) — a council
  reviewer must not cite a register status as ground truth for anything that
  will gate a real decision. Recommend: seed the reviewer's role prompt +
  context from its category's register file (satisfies Q-G's "curated
  corpus"), but require it to re-check the *specific* `verify-later` pointers
  relevant to the current fix plan live (grep/read), mirroring the exact
  verify-then-adversarially-recheck pattern already proven at stage-2 scale in
  `stage2_workflow.js`. Cheap per-run since one review only touches a handful
  of concepts, unlike stage 2's full-corpus sweep.

**Scope boundary — deliberately not yet done:** this design has not been
implemented against the live `0NN_fix_proposer.sql` workflow. That file belongs
to the actively-developed [[fixloop-workstream]] (separately live as of
2026-07-14 — TRIAGE + Phase 2 silent-check), and FIX-035's own standing rule is
"awareness before autonomy." Changing a production council's decision-gating
workflow is a cross-workstream architecture change and a judgment call on
scope/risk that belongs to the user, not something to just proceed on. Next
concrete step, when ready: pick one pilot category (candidate: whichever
category the next real fix-loop incident's blast radius touches) and add it as
a single extra reviewer step, rather than building all category-reviewers at
once.

---

## Backlog / open items

- **Superseded + abandoned sweep** (102 + 72 concepts) — DONE 2026-07-14. 18
  corrections confirmed (9 overturned), applied. This closes out stage 2 —
  every concept in the register has now been checked at least once. See
  `006_VERIFICATION_stage2.md` batch 3.
- **PUB-001 duplicate retirement** — DONE 2026-07-14. Retired to a pointer
  entry in `register/public-api.md` (kept the ID and full entry for its
  distinct source citation, added a duplicate-of note); `register/000_concept_index.md`
  row updated to reflect it. Did not fold `public-api.md` into
  `admin-dashboard-and-api.md` — a single-concept category file is already
  normal in this register (e.g. `entity-data.md`, `canine-biology.md`), so
  there was no forcing reason to eliminate the category itself.
- **Credential rotation** — two live-looking credentials found during stage 1
  extraction, not yet rotated (human task; see `RUNBOOK_concept_register.md`).
- **Stage 3 design** — RESOLVED 2026-07-14, see Stage 3 section above for the
  concrete recommendation (relevance-filtered subset, data-driven activation
  via `verify-later` join, curated-corpus-plus-live-recheck freshness model).
  Not yet implemented against the live fix-loop workflow — that's a separate
  user decision (see Stage 3's "scope boundary" note).
