# 335 — the offer analyser imports an unverified PAGE claim into the strategy layer and stamps it `from_field`, so the field built to prove sourcing vouches for a number the premise never contained

**Filed 2026-08-20** by the `vigilant_designer_offer_analysis` lane — **this is a defect in this
lane's own agent (`offer-analyser`, BIZ-032).** Found because the **leopardess lane caught it and
held the findings** rather than letting them reach a writer. **OPEN, owned by this lane.**

**One line:** `load_offer_surface` passes page **meta descriptions** into the analysis; the model
lifted a stale factual claim out of one (*"eight live sites"*, true count **23**) and wrote it into
`offer_ordering.lead_with` **rank 1** — with `from_field: "trust_threshold"`, a premise field that
does not contain the number.

---

## The evidence, all read at the artefact 2026-08-20

**The output.** `site_specs` aspect `offer_ordering`, leopardessconsulting.co.uk, `is_current`,
written by run `2026-08-19 15:14:56` (COMPLETED):
> `lead_with[0].point` — *"Your agent system will run on Kubernetes, Kafka, and Postgres — the same
> stack that runs **eight live sites** built by this team…"*
> `lead_with[0].from_field` = **`trust_threshold`** · `differentiated` = true

**The number is false.** `SELECT count(*) FROM sites WHERE status='deployed'` = **23**.

**Where it actually came from — a PAGE, not the premise.** Pages on that site whose
`meta_description` carries the phrase: `about` (*"…a platform that runs eight live sites"*) and
`index`. **`load_offer_surface` passes `title` and `meta_description` for every page** (NOTES
2026-08-14, honest limit 1). Searching every `is_current` spec for the claim returns **only
`offer_ordering` itself** — the analyser's own output. The `strategy` aspect does not contain it;
neither does any other premise aspect. **So the analyser is the first spec-layer carrier of a page's
claim.**

**The attribution is the defect, not just the staleness.** `from_field` exists so a reader can see
which premise field a ranked point came from — this lane's own honesty machinery. Here it names
`trust_threshold`, and the `why` clause reasons correctly *about* that field — but the **specific,
checkable number** in the point is not in it. A reader auditing the artefact sees a sourced claim.

**It was caught by a human lane, not by us.** The leopardess lane held all five findings at
`needs_human_review` on 2026-08-19 with `held_reason` recording an owner request for a design
report, and `grading`: *"the run is still degraded:true and repeats the stale 'eight live sites'
figure … so its rank-1 suggestion would put a false number in the hero."* **Nothing in this lane's
own machinery would have stopped it** — the finding was well-formed, the ordering artefact passed
every structural check B4 makes, and rank 1 is exactly what a writer would consume first.

## Why this is NOT `bugs_open/161`, and not `features_open/034`

- **`161`** is *the evidence register ratifies the claim it was built to catch* — a register
  vouching for a claim already on a page. **This is the opposite direction:** a page claim being
  promoted UP into the strategy layer, where 034 and B4 both treat spec prose as the authority.
- **`034`** (claims-audit over `site_specs` prose, owner-approved 2026-08-14) would **catch this
  after the fact** — it checks premise prose for invented specifics. **It does not stop the import**,
  and on this estate the ordering artefact is read by writers, so the window between writing and
  auditing is a window in which a false number reaches a hero. 034 remains the right track; this is
  a separate producer-side defect.
- **`bugs_closed/262`** (claims revalidator certifies DB state while the served page drifts) is the
  same family one layer down.

## Fix candidates, ordered by what closes the door

1. **(Preferred) Forbid numerals and named quantities in `lead_with[].point` unless they appear in
   the cited `from_field`.** A prompt line plus a verify assertion at write time: if the point
   contains a cardinal, the cited premise field must contain it too, else drop the clause. This
   makes the bad state hard to represent rather than asking the model to be careful, and it is
   checkable in `write_offer_ordering` without a new mechanism.
2. **Stop passing `meta_description` into the offer surface**, or pass it labelled as *unverified
   page copy, not evidence*. ⚠ **Costly:** the meta descriptions are load-bearing for the surface's
   real job (two of the first five gaswholesalers findings were grounded in missing/generic metas).
   Removing them would blunt a working check to fix an attribution bug — **not recommended alone**.
3. **Batch with v2(b)** (`features_open/030` §10) — the attribution line for `why` clauses. Same
   surface, same prompt, one migration. **This defect makes v2(b) load-bearing rather than cosmetic:**
   v2(b) was filed as "intermittent, does not justify a migration alone", and this is the case that
   justifies it.

## How to verify a fix

Both controls, on a re-run against **this same site** (its premise is unchanged and the pages still
carry the phrase, so the input that produced the defect is still live):
- **Positive:** the new `offer_ordering` for leopardess contains no cardinal in any `lead_with`
  point that is absent from its cited `from_field`.
- **Negative control:** gaswholesalers.com, whose rank-1 point legitimately carries premise-sourced
  specifics, must **keep** them. Without this, "no numbers anywhere" passes trivially and destroys
  the artefact's usefulness.

## Verification basis (owner ruling 2026-07-31)

**Not** put through the `090` loop. The substitute, stated plainly: every element was read
first-hand at the artefact today — the ordering row and its `from_field`, the two page
`meta_description`s carrying the phrase, a search of **all** `is_current` specs for it (returning
only the analyser's own output), and the true site count from `sites`. The motivating harm was
independently caught and documented by a different lane before I looked.

## FIX BUILT 2026-08-21 — candidate 1, both halves. STAYS OPEN until the chassis rolls.

**Status: the code is committed and INERT.** The bar for `/bugs_closed/` is fixed AND live; the Go
half does nothing until the next chassis roll, and the config half is deliberately held back until
it has. So this file stays here, and the defect is still reproducible on leopardess today.

- **Go:** `verify_cited_cardinals` (new action) — commit `d79e4243c`, registered **CLM-023**.
  Generic gate for "ranked items each naming their own source field": every cardinal in an item's
  prose must appear in the field that item cites. 12 tests, all fixtures verbatim from live
  `site_specs`.
- **Config:** `docs/agent_docs/sql_for_agents/537_offer_analyser_cardinal_attribution_gate_HOLD.sql`
  — commit `6b1f4cb08`. Splices the gate between `set_audit_source` and `write_offer_ordering`,
  repoints the write at the checked object, and appends the rule to the prompt as well.
  **`_HOLD` is load-bearing:** a step naming an unregistered action does not no-op, it fails the
  **whole** workflow (the validator concludes the action is remote and rejects it), which would take
  `write_offer_findings` down with it. Gate commit `d79e4243c`; the file carries the artefact probe.
- **Council:** submitted `9a8f1283-574e-44d7-8e66-b84789ba0429` (`Council-Submitted:` on the
  migration commit). **Verdict not yet read** — whoever picks this up owes that read, and owes
  acting on a REVISE, because the code is already on the shared branch.

### Two measurements that changed the design, and one CORRECTION to this file

> **⚠ CORRECTED 2026-08-21 — the negative control this file proposes CANNOT DISCRIMINATE.**
> "How to verify a fix" (below) names **gaswholesalers.com**, "whose rank-1 point legitimately
> carries premise-sourced specifics, must keep them". Measured over all six of its `lead_with`
> points: **none contains a cardinal at all.** It therefore passes *any* rule, including one that
> bans every numeral — the exact failure the control was written to prevent, and it would have read
> as a clean pass. The controls that actually bite are **webdesign.co.uk** ("sixty-three tools",
> present verbatim in the cited `value_proposition`), **robot-hands.com** ("six actuation types",
> likewise), and **robot-hands.com rank 5** ("2–3 technical articles per month" against a premise
> that writes "2-3" — an en-dash/hyphen mismatch a naive substring check fails). All three are now
> verbatim fixtures in the test file. The specifics this file had in mind on gaswholesalers are real
> but **non-numeric** ("rack pricing", "gasoline, diesel, and natural gas"), which a cardinal gate
> never touches.

1. **A digits-only gate cannot see this defect, and the nearest precedent is digits-only.**
   `verify_report_prose` is the right precedent and its numeric idea is reused — but its token
   regex is `\d[\d,]*\.?\d*` and the defect was the **word** "eight". Reusing it unmodified
   yields a gate that passes its own motivating case while reading green. Its own doc comment
   already names the hole from the other end. `TestDigitsOnlyScanWouldHaveMissedTheDefect` asserts
   the defect point contains **no digits at all**, so deleting the word vocabulary fails the suite
   rather than quietly widening the gate.
2. **"one" and "zero" are not quantity claims.** `[MEASURED 2026-08-21, all 30 live `lead_with`
   points]` including them in the challenged vocabulary flags **6, of which 5 are false** — "one
   click away", "a restart from zero", "the one you arrived with", "one of those categories", "in
   one workflow". Excluding them leaves **exactly 1**: this defect. They stay admitted on the
   **source** side, so a premise legitimately saying "one" still licenses a point saying "1".

### Why `drop` and not `fail`

The action defaults to `on_violation: "fail"`; the offer-analyser is configured **`drop`**.
`write_site_spec` deep-merges and an array takes the scalar-overwrite arm, so a **successful**
re-run replaces `lead_with` wholesale — but a run that **fails** at the gate writes nothing and
leaves the previous row `is_current`. On the one site that actually carries this defect, fail-mode
would report a working gate while the false rank-1 stayed live, and would also lose the findings
(written by the step after the ordering). Drop writes the survivors, removes the offender, and
records the removal in the artefact under `dropped_unsourced`. It still refuses to write an **empty**
`lead_with`.

### What is still owed

- **Read the council verdict** and act on it.
- **After the roll:** apply `537`, then re-run against leopardess (positive) **and** webdesign +
  robot-hands (the real negatives). ⚠ Applying `537` does **not** repair leopardess — only a
  successful re-run rewrites the row, and the `improvement-sweep` has been disabled since 08-17,
  so nothing will re-run it unprompted. **Coordinate with the leopardess lane first**: it is holding
  this lane's findings pending an owner design report.
- **Unmeasured, stated rather than absorbed:** a legitimate **year** in a point ("updated for 2026")
  is a cardinal and premises rarely carry one, so it would be dropped — no live point contains a
  year today. And the word vocabulary stops at ninety-nine, so "a hundred tools" is not challenged.
- **`features_open/034` is not replaced by this.** 034 asks whether the premise itself is true;
  this stops a page claim being imported into the premise layer and mis-attributed. 335 sharpens
  034's case.

## Relates to

`features_open/030` §10 v2(b)/(d) · `features_open/034` · `bugs_open/161` · `bugs_closed/262` ·
BIZ-032 (register: *"its inputs are unverified prose … until then this ceiling stands"* — this
shows the ceiling is lower than stated, because the inputs include unverified PAGE copy, not only
premise prose) · NOTES 2026-08-14 honest limit 1 (the surface carries metadata; that limit was
framed as *findings may be hypotheses*, and this is the sharper consequence)
