# PLAN — the tool fabrication fence (bugs_open/482)

**Status: DRAFT, pre-review.** Written before the adversarial plan review returns, deliberately,
so that what the review changes is visible as a correction rather than absorbed silently. Every
correction below this line gets a dated `> **CORRECTED**` block; nothing gets edited away.

---

## 0. What the bug actually is, restated

Not *"a boxing countdown has stale dates"*. The class is:

> **A `component_level='tool'` component asserts specific real-world facts that nothing on the
> estate verified, and a fact-backed fixture is byte-indistinguishable from a fabricated one.**

Dates are one facet of one instance. The worst known instance
(`tool-vet-comparison-vetcomparison-uk`, 30 invented veterinary practices) **contains no date at
all**. Writing that sentence first is the whole of what my first census got wrong — see
`WRONG_CALLS.md` 2026-09-04 and NOTES.

Two symptoms, one cause: a tool that needs real-world data and has no seam to any either
**invents** it (countdown, vetcomp, sfi26, budget-kit) or **ships empty** (fighter-comparator,
0 `<option>` / 0 `<select>`).

## 1. The design constraint that rules out the obvious fix

**Enumerating fabrication shapes is a losing game and we have now lost it four times in two days.**
ISO date string → numeric `year/month/day` triplet → object keyed `fight1..fight6` rather than an
array → 6 records rather than 15 → *no date whatsoever*. Every widening is specified against the
examples in hand; the next generator writes the next shape.

Four detectors demonstrated blind or self-discarding on this class:

| detector | status | why it fails here |
|---|---|---|
| 427 §23.2 Layer 1 (`check_event_fixture_completeness`) | built, armed on 0 agents | wrong side of the seam — checks *facts declared without evidence*, not *tool content with no fact behind it* |
| 427 §23.2 Layer 2 (birth refusal on ISO-date-keyed array) | proposed | shape mismatch; and **0** subjects fleet-wide |
| 427 §23.2 Layer 3 (`data-fact-id` resolves to a fact) | proposed | **structurally unreachable** — `data-fact-id` is 0 in served markup even for the component doing it right |
| `check_tool_fabrication` | **LIVE** | **detects and discards** — see §2 |

## 2. The finding the plan is built on

`check_tool_fabrication_action.go` does not fail to see these tools. On three of five census hits
it **computes the corpus signature** (20, 24, 30 entity-record objects, threshold 15) and returns
`Fabricated=false`, because Tier B gates on `dataBacked && !preserved` and `dataBacked` derives
from an `original` that a **born** tool has never had. `Signals` is populated; nothing gates on it.

**Consequence for the plan: the birth arm is not new detection work.** The evidence is already
computed, already correct, and thrown away. That makes the change small enough to be defensible,
which no pattern-widening version was.

## 3. Draft plan (PRE-REVIEW — expect this to change)

Organising principle, from `order-fix-candidates-by-what-closes-the-door`: **rank by what makes the
bad state unrepresentable, not by what detects it.** "Operators must remember X" is a defect.

- **A — Wire the existing gate onto the uncovered authorship paths.** `create_tool_component_action.go`
  (covers birth *and* regeneration, since `regenerateToolComponentInPlace` is called from `:307`,
  after the `:127`/`:160` gates) and `update_component_html_action.go`.
- **B — Give Tier B a birth arm**, so a computed signature is not discarded for want of an
  `original` that cannot exist. ⚠ Calibrate against the **134-tool** dataset-bearing set
  (`bugfix_427_event_render/CENSUS_2026-09-04_…txt`), per `component_write_guard.go`'s doctrine:
  every threshold there was simulated against the full live history first, and **two candidate
  checks were dropped because the simulation caught them misfiring**.
- **C — Coverage ratchet**, mirroring `component_template_writer_coverage_test.go`. Working name
  `tool_content_writer_coverage_test.go`, decision map `provenanceExemptWriters`. ⚠ **Contract
  wording, at the 427 lane's request:** *"declares provenance OR is listed with a written reason"*,
  **not** *"is fenced OR listed"* — 427's rail writers will satisfy the obligation by emitting
  provenance rather than by calling this gate, and a fence-membership contract would make compliant
  writers read as exemptions. Restate the inherited weakness in the header: it reads SOURCE, so it
  proves the call exists, not that it executes.
- **D — Prompt half.** `generate_tool_html`'s `prompt_template` (5,118 chars, 22 rules, **zero**
  about provenance) gains the rules migration 183 already gives the *recreation* prompt.
  ⚠ **DB config is live immediately; Go is inert until a roll.** Ordering matters — see §5.
- **E — Census + remediation.** Birth-time refusal is a control on the FUTURE only; every known
  violation predates any gate that could have refused it.
- **F — Serve-time backstop** (the owner's ask, 482 §5). Placement genuinely open: `experience_loop`
  audit machinery vs the tool pipeline itself.

## 4. The objection I already accept, unprompted

**A provenance requirement the fact supply cannot satisfy converts "fabricated tool" into
"unbuildable tool".** `evidence_base` for boxingonline holds 8 facts, 7 dated, **2** forward. The
fighter-comparator needs fighter attribute data for which **no source exists on the estate at
all**. So a refusal with only two outcomes (build / refuse) makes a whole category unbuildable —
and *"a guard that refuses good work gets switched off, and then it protects nothing"* is this
estate's own written doctrine (`component_write_guard.go`).

**Therefore the refusal needs a third outcome: build it honestly EMPTY, with a stated
"data to follow".** That is 427 §22.3's own precedent (*"an empty calendar that says so is not a
false claim; twelve invented ones are"*). Whether that is expressible in the current tool contract
is a question for the review, not an assumption.

## 5. Sequencing, and the trap in it

**The refusal must not land before the generator can express the good shape**, or whole categories
of tool become unbuildable — 427 §23.2 flagged exactly this and it is the same trap that made their
own draft's `input_schema` step a latent full-rebuild trigger for a paid page.

Order therefore: **D (prompt teaches the shape) → A/B (gate refuses the bad shape) → C (ratchet) →
E (remediate the past) → F (serve-time backstop)**. ⚠ Against this: **D is DB config and goes live
immediately, A/B are Go and are inert until a roll**, so shipping D first is automatic and safe;
shipping A/B first would refuse generations the prompt has not yet been taught to satisfy. This is
the same deployment-order note already written into `create_tool_component_action.go:124-126` for
the tool-doc header gate — *"apply the prompt update before or with the binary carrying this gate,
or every generation fails here"*. **That precedent is directly on point and must be followed.**

## 6. Explicitly NOT in scope

- **Widening `ExtractAssertionText` to script bodies.** Architecture-scope under the 2026-07-29
  ruling (it changes what the shared claims gate guarantees). Both lanes agree not to file it: if
  the rail lands and the fence consumes it, the widening becomes *duplicative* rather than
  deferred. 427 is recording that in its §23.2.
- **`bugs_open/449`** (no acceptance fence asserts a number) — the numeric sibling, **actively
  owned** by the `mortgagecalculator_couk_adoption` lane. Share vocabulary, cross-file, **do not
  extend**.
- **Remediating live sites.** boxingonline is with the owner via its lane; vetcomparison is with
  its lane. This lane dispatches nothing at either.

## 7. Open questions the review must answer

1. Where does the draft cause a **second incident**? (427's equivalent review found their draft's
   `input_schema` step would have triggered a full LLM rebuild of a paid page.)
2. Is C (provenance declaration) genuinely bounded, or self-deception?
3. Is there a cheaper framing missed entirely? *Reuse before build* is a hard estate rule.
4. What is the **smallest change closing the largest fraction** of risk, if the owner wants one
   commit rather than six?
5. Does the third outcome (§4) exist in the current tool contract, or must it be built?

---

# PLAN v2 — 2026-09-04, after simulation and peer review

The draft above stands unedited. Three of its load-bearing assumptions were refuted within the
day; each is marked below with what refuted it. **Nothing here was refuted by argument — all three
fell to a measurement**, which is the reason the plan is written this way round.

## v2.0 What was refuted, and by what

| draft claim | refuted by | outcome |
|---|---|---|
| §3B "give Tier B a birth arm; calibrate the threshold against the 134-set" | **Simulation** over `427`'s calibration set: dropping corroboration at threshold 15 → **28 convictions, 3 true, 25 false (89%)**. And the motivating bug sits at **6 records**, below the threshold, while the corpus's largest legitimate dataset is **73**. | **Withdrawn.** Record count is not under-tuned — it is the wrong axis, in both directions. |
| "port `ea` because it is the narrower, conservative column" | `427`'s diagnosis + first-hand check: `ds` and `ea` are **independent predicates, not subset/superset**. `ds` requires a human-readable string (≥8 chars with a space); `ea` requires identity+attribute keys. Three single-word-name records (`Secateurs`, `Loppers`, `Wheelbarrow`) pass `ea` and fail `ds`. | **Right choice, false reason.** `ds` would have discarded single-word entity names — and an invented product, practice or fighter is very often one word. |
| "`FactBearingFields(schema)` lets the birth arm ask a structural question instead of a content one" | `[MEASURED 2026-09-04, this lane]` **287 of 335** active tools have `input_schema` NULL — **85.7%**. | **Near-inert as an inculpatory signal.** Leaning on it reintroduces the 89% shape through a different door. |

## v2.1 The governing constraint

**Declaring nothing is what 86% of tools do, and most of them are innocent.** Therefore:

- the provenance declaration is a sound **exculpatory** test — declares a fact-bearing field, ids
  resolve ⇒ clean, cheaply and certainly, and this strengthens as `427`'s rail lands;
- it is a near-useless **inculpatory** one today;
- so **the `ea` content predicate carries essentially all the discrimination**, and this plan says
  so rather than presenting a two-part conjunction as though both halves bore weight.

⚠ **Do not restore the symmetry later without re-running the census.** The moment schema adoption
rises materially above 48/335, the inculpatory half becomes worth something — and until then a
reviewer reading "declares nothing" as suspicious is reading the corpus's default state as a
finding.

## v2.2 The plan

Ordered by **what it makes unrepresentable**, not by what it detects.

### Phase 1 — the prompt half (DB config; live immediately, no roll)
`generate_tool_html`'s `prompt_template` gains the provenance rules migration 183 already gives the
*recreation* prompt: do not invent real-world records; if you have no real data, ship the
interface with an explicit empty state rather than illustrative-looking data.

**Why first:** DB config is live immediately and Go is inert until a roll, so this cannot be
overtaken by the gate. The inverse order refuses generations the prompt was never taught to
satisfy — the trap `create_tool_component_action.go:124-126` already records for the tool-doc
header gate (*"apply the prompt update before or with the binary carrying this gate, or every
generation fails here"*). **That precedent is directly on point.**

*Makes unrepresentable:* nothing. It is a prompt. It is first because of ordering, not strength.

### Phase 2 — the `ea` content predicate, as a shared, testable pure function
`platform/orchestration/datahelpers/` — a pure `DetectEmbeddedEntityClaims(html) ClaimsResult`
returning the records found and the identity/attribute keys that matched. Pure core, no DB, no
`ActionParams`, mirroring `DetectToolFabrication`'s own testability design.

*Verified by:* replaying it over `427`'s **134-tool calibration set** and requiring the published
confusion matrix — this estate's doctrine (`component_write_guard.go`) is that a threshold is
simulated against the full live corpus **before** commit, and that two candidate checks were
dropped when the simulation caught them misfiring.

*Makes unrepresentable:* nothing yet — it is the instrument. Shipped alone it is inert and safe.

### Phase 3 — wire it into the uncovered authorship paths, REPORT-ONLY, default OFF
`create_tool_component_action.go` (covers birth **and** regeneration —
`regenerateToolComponentInPlace` is called from `:307`, after the `:127`/`:160` gates) and
`update_component_html_action.go`.

**Opt-in field, unsafe side default OFF**, per the owner ruling of 2026-08-02 §2. Report to
`recordComponentWriteRejection`'s existing observability path; refuse nothing.

⚠ **Optional-key budget:** adding a key to these actions counts against **N=10**
(`cmd/config-key-audit --optional-key-budget`, register WFA-013). **Run it before committing**, and
if either action crosses N, record the accumulated-surface review in
`architecture_review/optional_key_budget_acks.json`.

*Makes unrepresentable:* still nothing — deliberately. This phase exists to produce the live
false-positive figure on real traffic before anything refuses.

### Phase 4 — the coverage ratchet
`tool_content_writer_coverage_test.go`, map `provenanceExemptWriters`, mirroring
`component_template_writer_coverage_test.go`.

**Contract wording, agreed with `427`:** *"declares provenance OR is listed with a written
reason"* — **not** *"is fenced OR listed"*. Their rail's writers satisfy the obligation by emitting
provenance rather than by calling this gate; a fence-membership contract would make compliant
writers read as exemptions, which is worse than a missing check because it reads as a decision
someone made.

Inherited weakness restated in the header, unsoftened: **it reads SOURCE, so it proves the call
exists, not that it executes.**

*Makes unrepresentable:* **a future authorship path appearing unfenced without a human decision.**
This is the phase that actually closes the door, and it is the reason the whole plan is worth more
than a checker.

### Phase 5 — turn refusal on, to `needs_human_review`, never to breakage
Only once Phase 3's live figure justifies it. Refusal routes to human review — the existing gate's
own design — never to a silent failure.

⚠ **Needs a THIRD outcome besides build and refuse.** `evidence_base` on boxingonline holds 8
facts, 7 dated, **2** forward; `tool-fighter-comparator` needs fighter attribute data for which
**no source exists on the estate**. A refusal with two outcomes converts *fabricated tool* into
*unbuildable tool* for a whole category, and *"a guard that refuses good work gets switched off,
and then it protects nothing"* is this estate's own doctrine. The third outcome is **build it
honestly empty with a stated empty state** — `427` §22.3's precedent: *"an empty calendar that
says so is not a false claim; twelve invented ones are."*

*Makes unrepresentable:* a new tool being born with undeclared invented entity records.

### Phase 6 — remediation of what already shipped
Birth-time refusal is a control on the **future**; every known violation predates any gate that
could have refused it. Census is `427`'s calibration file plus this lane's verification.

**Not this lane's to execute on live sites.** boxingonline is with the owner via its lane;
vetcomparison is with its lane. This lane supplies evidence and the verification contract.

## v2.3 Smallest change closing the largest fraction

**Phase 4, the ratchet — but only if exactly one thing ships.**

Phases 2+3 produce a measurement; Phase 5 produces a control on new tools. **Phase 4 is the only
one that prevents the bug's actual recurrence mechanism**, which was not "a detector was too
narrow" but *"a live, correct, council-reviewed gate covered one authorship path of several and
nobody noticed for seven weeks"* — `bugs_closed/021`'s named class, recurring.

Second choice, if the owner wants something that moves today rather than at the next roll:
**Phase 1**, because it is DB config, is live immediately, needs no build, and addresses the cause
the evidence actually supports — `[MEASURED]` the generator was never told not to fabricate and
has no step that could consult a real fact.

## v2.4 Architecture scope

**None of the above is architecture-scope** under the 2026-07-29 ruling, on the current reading:
each phase is additive, opt-in with the unsafe side OFF, and changes no shared mechanism's
*guarantee*. Under the 2026-08-11 RFC_022 narrowing, an opt-in field with an unsafe-OFF default
that no live consumer names is explicitly **not** architecture-scope — but the consumers must be
**enumerated, not asserted**, and that enumeration is owed at submission.

**Explicitly out of scope, and both lanes agree not to file it:** widening `ExtractAssertionText`
to script bodies. If the rail lands and the fence consumes it, that widening is *duplicative*
rather than deferred.

## v2.5 Still open

1. Does the "build honestly empty" third outcome exist in the current tool contract, or must it be
   built? **Unanswered — and Phase 5 cannot ship without it.**
2. Placement of the serve-time backstop (owner's §5 ask): `experience_loop` audit machinery vs the
   tool pipeline. Both named in 482 §5; neither chosen.
3. `deploy_tool_action.go` forks a component to a second site **without re-inspecting it**. Not an
   authorship path, so outside the ratchet as scoped — but it propagates a fabrication. Recorded,
   not solved.
4. The `ea` extractor's exact record-splitting: `ds=18 / ea=20` on one row does not close
   arithmetically even after the subset confusion is resolved (18+3=21). Does not move any
   conclusion; must be settled before the predicate is ported verbatim.
