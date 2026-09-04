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
