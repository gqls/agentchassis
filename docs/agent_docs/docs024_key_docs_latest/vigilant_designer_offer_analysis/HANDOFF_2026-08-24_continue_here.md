# HANDOFF — vigilant designer + offer analyser (2026-08-24)

**COLD-START = this file + `features_open/030` §10 (the v2 batch, now FOUR items) + `features_open/034`.**
**This supersedes `HANDOFF_2026-08-21_continue_here.md`.**
`bugs_closed/335` is closed — read it only for its three residuals, which are restated below.

> **Re-run every liveness claim here before acting.** This branch takes hundreds of commits a day.
> Verify against `git archive <resolved-sha>` — never the working tree (another lane is often
> mid-edit and it may not compile) and never the moving name `HEAD` (two archives minutes apart came
> off different commits and produced a confident false conclusion; `WRONG_CALLS.md` 2026-08-21).

## The one-line state

> **`bugs_open/335` is CLOSED — fixed, live and proven estate-wide. The lane's next work is the v2
> batch, and its strongest item, v2(d), was missing from the feature file until today.**
> The attribution gate (`verify_cited_cardinals`, CLM-023) runs on chassis `v1.0.1332`; **6 post-537
> runs across all 5 enrolled sites as of 2026-08-24**, zero unsourced cardinals estate-wide.

## What is DONE, so nobody re-takes it

| | state |
|---|---|
| `bugs_open/335` | **CLOSED** → `bugs_closed/335`, 2026-08-24 |
| the gate (`verify_cited_cardinals`, CLM-023) | LIVE on `v1.0.1332`, capability-probed with two controls |
| migration `537` (gate step + prompt rule) | **APPLIED** 2026-08-22 11:03Z |
| the `B2B` false-positive fix | LIVE, rolled 2026-08-24 |
| council `9a8f1283` | **APPROVED** round 3 (2 REVISE rounds, each found a real defect) |
| `scripts/fire-offer-analyser.sh` | NEW — the surgical single-agent dispatcher (see below) |

**⚠ FIRE B4 WITH `scripts/fire-offer-analyser.sh <domain>`, NOT `run_improvement_sweep_once.sh`.**
The latter fires the *whole* improvement loop, and its `triage_findings` promotes every `detected`
work item on the site into live handler dispatches that change live pages — **[MEASURED 2026-08-22]
111 items on webdesign.co.uk and 37 on leopardessconsulting.co.uk**, including other lanes'. To
exercise one agent that is the wrong instrument by two orders of magnitude. The new script publishes
exactly one dispatch, and refuses if the live workflow does not carry the gate step.

## The three residuals `335` left — conditions on the next change, not open defects

1. **The gate's ENFORCEMENT arm has NEVER FIRED.** Total drops estate-wide, ever: **0 as of
   2026-08-24**. The *prompt* half prevents the bad output, so the gate has had nothing to catch. It
   is proven by mutation-tested unit tests built from verbatim live strings — **not** by a live
   firing. ⚠ **A run containing no cardinals passes this gate trivially, so never quote a clean run
   as evidence the gate works.** That mistake cost this lane two days of false confidence.
2. **`dropped_unsourced` has no automated consumer.** Tolerable only while `offer_ordering` has none
   — **measured 2026-08-24: `strpos` finds the literal `lead_with` in `offer-analyser` alone, and no
   Go file reads it.** ⚠ Before `offer_ordering` gains its first automated consumer, drop-mode must
   surface as a **work item**, and per the council's `bug_historian` seat that requirement belongs in
   the **ACTION**, not in one call site — any future caller opting into `drop` inherits the gap.
3. **`GPT-4` still yields the cardinal `4`** (the character before the digit is `-`, not a letter).
   `B2B`, `S3`, `IPv6`, `Web3` are handled. Pinned by a test that **fails if anyone fixes it**, so
   the doc comment cannot rot.

## What the next session should do

### 1. THE v2 BATCH — `features_open/030` §10. One migration, one re-proof.

> **⚠ READ THIS FIRST: §10 said "three changes" for a week and v2(d) — the STRONGEST item — was not
> in it.** It lived only in `NOTES` (*"v2(d) CENSUS"*) and the 08-17/08-18 handoffs, while those
> handoffs pointed at *"`features_open/030` §10"* for it. **Folded into §10 on 2026-08-24**, heading
> corrected to FOUR. A session doing the right thing — reading the canonical backlog — would have
> dropped the best item. Watch for that shape: *a pointer naming a canonical home the content never
> reached.*

- **v2(d) — machine-checkable acceptance predicates. START HERE; it is the strongest.**
  B4 emits a structured predicate **only when it can**, alongside the prose, and stays silent
  otherwise — **per-finding opt-in, unsafe default OFF** (the 2026-08-02 shared-seam ruling's shape).
  **The census is done and settles it** (all 22 live acceptance tests read 2026-08-17;
  `[CLASSIFIED]`, not measured): **8 of 22 fully expressible**, 6 partly, 8 judgement-only.
  **The load-bearing fact: the test that actually FAILED is in the expressible set** — webdesign's
  *"…before any count of tools or articles"* against a hero opening *"Sixty-three browser tools…"*
  is an ordering assertion over two string positions, ordinary text arithmetic. A predicate would
  have caught the exact failure that shipped.
  ⚠ **THE TRAP IS THE WHOLE RISK: never let the model emit a predicate for a JUDGEMENT test.** Two
  thirds cannot be expressed; a plausible predicate over a judgement clause grades confidently and
  wrongly — **worse than the prose it replaced, because it carries a green tick.** The *silence* arm
  is the load-bearing half.
  ⚠ It also inherits `335`'s lesson: a predicate is a **self-attributing artefact** of exactly the
  shape `016b` §9 now warns about — it asserts its own checkability. See that entry before designing.
- **v2(b) — attribution in the `why` clauses. PARTLY DONE, and check before rebuilding.** `537`
  already added the sourcing rule to the prompt for `lead_with` **points**; v2(b) asked for it on the
  **`why` clauses**, which is still outstanding. Re-read the live prompt before assuming either way.
- **v2(a) — a bounded head-of-hero excerpt per page in the offer surface.** The surface is page
  METADATA, not content, so some findings are hypotheses. ⚠ **This GROWS the surface**, and the
  truncation baseline (`__truncated` absent at 104 pages, webdesign.co.uk, 2026-08-15) is v1's —
  **re-run that check on that same site after (a) before trusting it anywhere.**
- **v2(c) — `primary_model` in the degraded arm's field list.** ⚠ **LATENT, no live instance** (its
  population claim was measured wrong and corrected the same day — zero sites lack one).
  **Must not be the reason to open the batch.** ⚠ Do NOT fix it by letting the model *infer* a
  `primary_model`.

### 2. `features_open/034` — claims audit over `site_specs` prose
Owner-approved 2026-08-14, still not designed. **`335` sharpens its case and does not replace it:**
034 asks whether the premise itself is true; `335` stopped a *page* claim being imported into the
premise layer and mis-attributed. Different layers, both needed.

### 3. Sweep window — 18 of 23 sites still lack a ranked record
Only the 5 enrolled sites have `offer_ordering`. ~1 site / 15 min. **Owner's call** — the
`improvement-sweep` scheduled task is `enabled=false` on his cost control (last fired 2026-08-17).
Enable by direct `UPDATE`, never a migration, and **disable in the same session**.
⚠ Now that the gate is live, new orderings are gated on the way in — so a sweep is *safer* than it
was, but no cheaper.

## Watch-outs that have actually bitten this lane

- **⚠ A clean run is not evidence a content gate works** — see residual 1. The negative control must
  be an item whose specific IS legitimately sourced and must **survive** (`robot-hands`: *"across six
  actuation types"*, verbatim in its cited field). A control carrying no specific passes any rule,
  including one banning every numeral — which is what `335`'s own bug file originally proposed.
- **⚠ Before attributing a difference to your own change, read the rest of the object.** The
  over-suppression scare of 2026-08-22 was refuted by `avoid_leading_with`, **two keys away in the
  JSON already open**. An "n=2, watch it" hedge did not stop it reaching four documents including the
  owner's log. `WRONG_CALLS.md` 2026-08-24.
- **⚠ Reviewers judge the SKETCH.** Three council objections across two rounds were factually wrong
  about the file because the sketch omitted guards that existed (`snapshot_agent`, `BEGIN/COMMIT`,
  the anchor gate, the rollback file). It cost two rounds. Sketch the file's skeleton with line
  numbers.
- **⚠ `orchestration_states` has no `agent_type` column** — it is `owner_agent_type`, and the wrong
  name *errors* rather than returning zero, so behind a `2>/dev/null` it reads as "nothing to see".
  And a zero from the **corrected** query still is not evidence: terminal rows are reaped (~24-48h
  for `COMPLETED`) and the column names the ORCHESTRATION's owner.
- **⚠ `llm_call_log.agent_type` is the DISPATCH context.** A hand-fired B4 run lands under
  `'generic'`. Key on `step_name='run_offer_analysis'` — it read 9 when `agent_type` read 7.
- **⚠ The Anthropic 400 *"you will regain access on <date>"* names the BILLING RESET, not an outage
  window.** Two B4 runs died on it 2026-08-22 18:34/18:40. Read the clock, not the date.
- **⚠ `LIKE '%lead_with%'` LIES** — `_` is a LIKE wildcard, so it matches "lead with" in reviewer
  prose. Use `strpos`. It returned 3 apparent consumers where there is 1.
- **⚠ psql prints UTC, your shell prints BST.** Always toward alarm — make the DATABASE subtract.
- **⚠ A site with `created_at` today, 0 pages, or `status='active'` is UNDER CONSTRUCTION** and is
  not a fact about the estate.

## Who owns what nearby

The **leopardess lane** works leopardessconsulting.co.uk and was holding five of this lane's findings
at `needs_human_review` pending an owner design report — **still open as of 2026-08-24, and B4 was
re-run against that site twice with the owner's authorisation.** Coordinate before firing again.
`bugs_open/333` belongs to the 301 lane — contribute, do not compete.
`copy_quality_two_stage` + the LMC lane still work loanandmortgagecalculator.co.uk.
