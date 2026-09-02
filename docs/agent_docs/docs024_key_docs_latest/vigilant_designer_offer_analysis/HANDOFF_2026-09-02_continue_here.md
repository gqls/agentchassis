# HANDOFF — vigilant designer + offer analyser (2026-09-02)

> ## ⚠ LANE OWNERSHIP — CONSOLIDATED 2026-09-02 (owner's instruction)
> **TWO sessions worked this lane today, both named `offer analyser benefit analyser visual
> designer`.** `[4628f9]` had been running 7 days; `[6c226c]` was started on 09-02 on the belief that
> `[4628f9]` had crashed — it had not. **The owner has consolidated the lane into `[4628f9]`, which
> owns it; `[6c226c]` handed over in full and stood down.**
>
> **⚠ WE SHARE A NAME, so a `SendMessage` to the bare name will not resolve.** Address the ref.
> `copy_quality_two_stage [e41906]` has been told to route here.
>
> **What the duplication cost, recorded because it is the argument for the rule:** both sessions
> independently verified that the gate was in the binary and called by nothing, within about half an
> hour of each other — and `[4628f9]`'s §H1d in the superseded handoff ("WIRED TO NOTHING") was
> **true at 16:47:21Z and false by 17:14:53Z**, when `[6c226c]` applied `681`. Neither was wrong.
> ⚠ **The superseded file therefore contains TWO sections numbered §H1d, saying different things**
> (lines ~463 and ~689). Nothing was clobbered — verified with `--numstat` — but they need
> reconciling by whoever now owns the file.
>
> **A contributing cause worth fixing elsewhere:** the memory cold-start pointer for this lane had
> been stale at `HANDOFF_2026-08-15` for **18 days**, so a session picking the lane up landed before
> the imagery fix, the gate, the audit and both rulings. Repointed 09-02.

**COLD-START = this file, then:**
- `docs/agent_docs/docs024_key_docs_latest/vigilant_designer_offer_analysis/HANDOFF_2026-08-26b_continue_here.md`
  — **still the authority on everything this file does not mention**: §C (imagery supply), §E
  (residuals), §G (predicate vocabulary / gate 1c), §H2b (carousels), §H3 (logo chain), §H4 (the
  axis finding in full). Its §H1c/§H1d/§H1e are **discharged** — this file replaces them.
- `docs/agent_docs/docs024_key_docs_latest/vigilant_designer_offer_analysis/SUMMARY_2026-09-02_the_gate_holds_the_mint.md`
  — the plain-prose read-out; start here if you want the story rather than the state.
- `docs/agent_docs/docs024_key_docs_latest/vigilant_designer_offer_analysis/NOTES_vigilant_designer_offer_analysis.md`
  — the technical log; every number below is derived there with its query.
- Register **CQ-034** in `docs/agent_docs/docs026_concept_register/register/content-quality.md`.

**What changed since 08-26b:** the producer register gate was built, reviewed twice, applied, and is
now live and repairing. That work is **DONE**. The next work is **Decision D**, and it is a fresh
build rather than a continuation.

---

## §1 — WHAT IS LIVE (all verified at the artefact, 2026-09-02)

**The producer register gate.** Every `lead_with` benefit point the `offer-analyser` mints is scanned
against `BANNED_REGISTER_v1` (the owner's banned words AND the shape family) *before* it is
persisted; a violating point is restated by one judged model call.

| what | where |
|---|---|
| Go | `platform/orchestration/datahelpers/registerwords.go`, `platform/orchestration/actions/repair_ordering_register_action.go` (+ tests), `registry.go`. Commit `f7156fb54` |
| wiring | `docs/agent_docs/sql_for_agents/681_offer_analyser_producer_register_gate_HOLD.sql` — **APPLIED** 17:15Z, in `schema_migrations` |
| the model | `docs/agent_docs/sql_for_agents/682_offer_analyser_register_gate_ai_service.sql` — **APPLIED** 17:18:22Z ⚠ see the number clash below |
| council | `4054f4d9-cd75-4b9c-8b8c-b7b86f11de1e` — **APPROVED round 2**, 16 reviewers. Round 1 was REVISE |

Live chain: `verify_ordering_cardinals` → `repair_ordering_register` → `write_offer_ordering`.

> ### ⚠ MIGRATION NUMBER 682 IS USED TWICE, AND IT IS MY DOING
> `682_content_listing_collapse_empty_card_slots.sql` (another lane, committed `f57f5ad1f` at
> 11:51) and `682_offer_analyser_register_gate_ai_service.sql` (mine, 18:17) share a number.
> **Cause:** I took 681 on **08-31**, when 680 genuinely was the maximum, then took 682 **on 09-02
> without re-running the check** — and the runbook says to check *at write time* precisely because
> "concurrent sessions take numbers hourly". They took twenty while I worked: the max is now **710**.
> **What it does NOT break:** the runner keys on FILENAME and both rows are in `schema_migrations`,
> so neither is pending and neither will re-apply. Both are already applied, so ordering is moot.
> **Deliberately NOT renamed.** Renaming needs a live `UPDATE` on the ledger to keep the row matching
> the file — a write against production to fix a cosmetic clash, with the `git mv` + pathspec
> copy-trap on top. The estate's own precedent for colliding numbers is *resolve by slug*, and these
> two slugs could not be less alike. **Recorded so the next reader is not surprised; the receiving
> session may overrule this and rename, in which case update the ledger row in the same breath.**
> ⚠ **And take the number at WRITE TIME, from `ls`, every time — not from the last one you used.**

### What it has actually done — and what that does NOT yet prove

| site | at | checked | violations | repaired | still violating |
|---|---|---|---|---|---|
| farmerinsurance.uk | 17:16 | 6 | 1 | 0 | 1 |
| garden-tools.uk | 17:31 | 6 | 2 | **2** | 0 |
| webdesign.co.uk | 17:53 | 6 | 2 | **1** | 1 |

The repairs are good — read them in the artefact, not the status. On `garden-tools.uk` both were
surgical: the banned word deleted, the grammar corrected around it (`an`→`a`, capitalisation),
everything else byte-identical.

> ### ⚠⚠ DO NOT REPORT THE GATE AS WORKING YET. THE NUMBERS LOOK BETTER THAN THEY ARE.
> `[MEASURED 2026-09-02]` points minted **since** the gate went live: **1 of 12 dirty = 8.3%**,
> against an all-history baseline of **23.3%**. That looks decisive and **is not evidence**:
> **P(≤1 dirty in 12 | the rate is unchanged) = 0.193** — a result this good happens about **one
> time in five by chance alone**.
> **To DETECT a halving (23% → 11.5%) at p≤0.05 you need ~25 minted points.** At the historical
> ~133/day that is a few hours; at today's burst rate, sooner. **So the honest statement today is
> "the mechanism works, the rate is unmeasured", and the measurement is cheap and close.** Re-run
> the query in NOTES (the 8-arm predicate — use the SAME one, see §3).

### ⚠ The corpus got DIRTIER while the gate was inert, and that is expected

`[MEASURED 2026-09-02, identical 8-arm predicate]` live corpus **36 of 190 = 18.9%**, up from
**25 of 185 = 13.5%** on 08-31. Nothing regressed: the mint ran unchecked for two days while the Go
was inert awaiting a roll. **The gate repairs at the MINT, so a site only gets clean when it is next
re-analysed** — 32 sites at roughly hourly means days to cycle through. **A rising corpus number is
not the gate failing.** Judge the gate on the post-17:17 mint, never on the corpus.

---

## §2 — WHAT TO DO NEXT: DECISION D (the question hierarchy)

**This is the whole of the next build.** ⚠ **RELAYED, second-hand** via `copy_quality_two_stage`
(owner's words as relayed: *"yes to both, we can see if it works"*, their rulings ledger 2026-09-02).
Treat as a strong lead; it has not been heard first-hand by this lane.

**Two rulings arrived together and THE ORDER MATTERS:**

1. **Build `question_hierarchy` + `answered_by`.**
2. **The ranking axis inverts** — buyer-relevance + readability govern heroes; **differentiation
   demotes to an INPUT**.

> ### ⚠ DO 1 BEFORE 2, AND THE REASON IS MEASURED, NOT AESTHETIC
> `offer_ordering.lead_with[]` currently ranks on differentiation, and `[MEASURED 2026-09-02, n=190]`
> that axis is completely intact — % `differentiated` by rank: **100 / 100 / 85 / 65 / 32 / 53**.
> So the ranking is on the axis the owner has just demoted.
> **But re-ranking first achieves nothing.** `HANDOFF_2026-08-26b` §H4 measured that the gap is
> **ABSENCE, not order** — only 19 of 186 points (10%) addressed effort or practicality at all.
> **Re-ranking cannot surface material that was never derived.** A prompt migration done first would
> spend a round producing the same seller-axis list in a different order.
> **Derive the hierarchy, then re-rank against it.**

**⚠ AND DO NOT QUOTE ANY IMPROVEMENT IN THAT 10% GAP.** A re-measure on 09-02 read 30%, which looks
like it closing on its own. **It is not comparable — I widened the regex** (added cost/price/£/free/
quick/simple/step to the original effort/time terms). A different instrument is not a different
result. **Re-run §H4's exact proxy before quoting movement.**

### The seam split — proposed by this lane, ACCEPTED by `copy_quality_two_stage`

**THIS LANE (production side)**, all inside `offer-analyser`, all passing through the register gate
like everything else it writes:
- a **`question_hierarchy`** array in `offer_ordering`'s shape: ranked buyer doubts, each with a
  `why` citing the source field it was derived from;
- **`answered_by`** on each `lead_with` point, pointing at the question it addresses, or explicit
  **`unanswered: true`**.

**THEIRS (consumption side):** writers read the hierarchy + `answered_by`, and copy ORDERING follows
it.

- ⚠ **THE JOIN IS THE DELIVERABLE, NOT THE LIST.** A hierarchy with no link to the copy would be the
  third provenance-stamped artefact nobody reads — and this lane's own `offer_ordering` (32 sites,
  **zero** writer consumers until they wire it) is the argument.
- ⚠ **BOUNDARY, agreed identically both sides:** the hierarchy is **unserved rationale, structured
  input only**. **Never rendered into a prompt as prose** — *"most visitors first ask X"* in a
  writer's or critic's window **IS** the presumption shape the owner banned, and will be copied
  verbatim.
- ✅ **PRE-REGISTERED ACCEPTANCE CRITERION** (both lanes, recorded before any run): **the first pass
  comes back MOSTLY `unanswered` at the top. That is the finding, not the failure.**

---

## §3 — TRAPS. Do not re-derive these; each cost real time.

- **⚠ `\m`/`\M` are the word boundaries in Postgres. `\b` is a BACKSPACE there** and has already
  produced two confident zeroes in this lane. The 8-arm dirty predicate lives in NOTES — **use it
  verbatim**, and never compare a number against one produced by a different arm set (that mistake
  is logged in `WRONG_CALLS.md`, 2026-09-02).
- **⚠ IF THE GATE GOES QUIET, CHECK `ai_service` FIRST.** `offer-analyser` has **no root
  `ai_service` block**; the model config sits on individual steps. Without `682`'s block the gate is
  live, firing, and repairing **nothing** — it returns `"no ai_service configuration resolvable"`,
  keeps every point and records the reason. That happened for real in a two-minute window and is the
  `params.StorageClient` shape: **a capability with no live caller has an untested dependency on its
  ENVIRONMENT, and the first real call is what finds it.**
- **⚠ THE SEED IS STALE ON THE FIELD THAT THREADS THE CHAIN.** `408_offer_analyser_agent.sql` says
  `write_offer_ordering.spec_data = 'offer_analysis.result.ordering'`; the **live row** says
  `ordering_register_checked.object`. A migration written from the seed would silently un-wire two
  live gates while reporting success. **Derive insertion points from the live row.** Now a LANDMINE
  (`docs/agent_docs/docs024_key_docs_latest/LANDMINES.md`).
- **⚠ A REPAIRED POINT IS CLEAN AGAINST THE REGISTER AS WRITTEN, NOT AGAINST THE OWNER'S
  OBJECTION.** `garden-tools.uk` rank 4 is marked `repaired` and still ends *"— not a default to
  premium"*: the em-dash form of `x_not_y`, which **v1 could not see** (both the register pattern and
  `negXNotYRe` are comma-anchored). `copy_quality_two_stage` is fixing it in **v2**. **Do not widen
  `negXNotYRe` from this lane** — it is shared with the page gate, the voice gate and the nightly CLI.
- **⚠ A COUNCIL "MISSING" ITEM ON AN *APPROVED* VERDICT IS NOT A FORMALITY.** `llm_reliability`'s
  missing item named the `ai_service` failure two days before it happened and asked for it to be
  confirmed at application time. **Read the report, not the decision field** — this approval carried
  nine objections, two of them real.
- **⚠ A GUARD FOR A STATE THAT HAS NEVER EXISTED IS THE ONE MOST LIKELY TO BE VACUOUS.** My register
  lockstep was anchored to the `v1` filename and would have gone **green** on a v2 cut — the exact
  thing I had promised a peer it would catch. Fixed (`91e2ad706`) by globbing for the highest version
  present. **Induce the state before believing the guard.**

---

## §4 — CROSS-LANE STATE

**`copy_quality_two_stage`** is the close partner; everything below is agreed with them.
- **v2 of the register is IN FLIGHT IN THIS SHARED TREE** (uncommitted as of 17:5xZ): a new
  `BANNED_REGISTER_v2.json` plus edits to `registerwords.go`, `registerwords_test.go`,
  `negationtells.go`. It adds the **plain-words class** and the **em-dash widening**. My suite passes
  against their working state. ⚠ **If you find `platform/orchestration/datahelpers/` dirty and red,
  it is probably theirs mid-cut — check `git status` before assuming you broke something.**
- **My lockstep is the contract**: a register word with no Go rule, a Go rule with no register entry,
  or a new register version Go does not implement, all fail the build. **Shape NAMES are compared,
  never patterns** — their patterns are documentation proxies, `negationtells.go` decides.
- **Owner ruled option (a) on the exemption-as-licence question**: clean the briefs fleet-wide, gate
  semantics untouched. **This lane's 32-of-34 measurement earned that ruling**; the wash campaign is
  theirs, nothing owed from here.
- **The third-path RFC trigger stands**: `architecture` ruled that the *accumulation* of
  analyser→writer paths earns an RFC, not the second instance. **Decision D creates no new path.**

**The audit that bounds this lane's claims** `[MEASURED 2026-09-02]`: `offer_ordering` is **one of
eleven** writer-input specs carrying register violations, and among the smallest — 25 of 1,462 dirty
text fields. `content_direction` 549/1576, `strategy` 282/640, `identity` 174/703, and seven more.
⚠ **Three of them are worse than ungated: they are EXEMPTION-LICENSED** — `content_direction.formatted`,
`identity.key_differentiators`, `identity.target_audience` are on `rewrite_negations`'
`defaultBriefFields`, so a construction there is matched as brief-supplied and **deliberately left in
SERVED copy**. The eight leak; **these three launder.** When the third path triggers the RFC, **those
three rank first by mechanism, not by volume.**

---

## §5 — STILL OPEN (unchanged, from `HANDOFF_2026-08-26b_continue_here.md`)

1. **Imagery SUPPLY (§C3) — the larger half of the owner's imagery ask, still UNOWNED.** The pipeline
   makes heroes and icons and barely any illustrations. Say the supply figure alongside any imagery
   claim.
2. `section/illustration` resolution is **first-wins by kind** (§E2); apis.uk has offered itself as
   the worked test case.
3. Eight components still declare `site_assets.image` and are exposed to the alias trap (§E8). ⚠ The
   architecture seat set the trigger: **a THIRD component hitting it** is when to ask for an explicit
   opt-out rather than another repoint.
4. **H2b's carousel flag**: `info-card-grid` has a declared `carousel` boolean and it is **ON on 1 of
   42 instances**. The cheapest lever for the owner's carousel ask, and *who sets it, on what
   evidence* is a real decision.
5. **Gate 1c's negative control** (§G) — still blocked on this lane's own `features_open/030` §10
   v2(a), unstarted.
