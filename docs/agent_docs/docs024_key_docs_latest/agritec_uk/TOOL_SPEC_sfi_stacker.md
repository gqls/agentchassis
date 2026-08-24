# Tool spec — SFI26 revenue stacker (replaces the retired SFI Revenue Stacker)

Written 2026-08-22, at the owner's instruction: *"make the SFI calculator use the correct up to
date facts and figures. Deconstruct it as necessary."* This is the deconstruction and the
specification. The figures it depends on are **already in the evidence register** —
`SEED_2026-08-22g_sfi26_action_rates.sql`, 72 attested facts — so this document names them
rather than restating them.

Ledger row: **T6**. Register facts: `ATT-sfi26-actions-table` plus `ATT-sfi26-<CODE>` ×71.

---

## 1. Why the old tool could not be patched

Audited against the SFI26 scheme rules, 2026-08-22. Of its **nine** revenue lines:

| Old line | Old rate | SFI26 position | Verdict |
|---|---|---|---|
| SFI Management Payment | £20/ha, first 50 ha | **abolished** | remove |
| SAM1 Soil Assessment Plan | £6/ha | no equivalent action | remove |
| IPM1 Integrated Pest Mgmt | £989/yr fixed | no equivalent action | remove |
| NUM1 Nutrient Mgmt Review | £652/yr fixed | no equivalent action | remove |
| HRW1 Assess Hedgerows | £3/100m | no equivalent action | remove |
| SAM2 Multi-species cover crop | £129/ha | **CSAM2** £129/ha | recode only |
| IPM4 No insecticide | £45/ha | **CIPM4** £45/ha | recode only |
| SAM3 Herbal leys | £382/ha | **CSAM3 £224/ha** | rate 41% too high |
| HRW2 Manage hedgerows | £10/100m | **CHRW2 £13 per 100m for one side** | rate low **and** scope changed |

**Two of nine are correct, and only after recoding.** Four of the actions no longer exist —
they are the paid assessment and planning ones, which is coherent with DEFRA's own stated reason
for dropping the management payment: releasing funding to offer more agreements.

That is why this is a rebuild. Every code, every rate and the entire payment model changed; what
survives is the *idea* of stacking compatible actions across a holding.

## 2. What the tool is, and what it is not

**Is:** a scenario model. The reader describes a holding — areas, lengths, counts — picks actions,
and sees what that combination would pay under the published SFI26 rates, with the scheme's own
caps applied.

**Is not:** an eligibility determination, an application, or advice. It cannot see the holding's
land classification, existing agreements, or obligations. The banned-claims already in the
register enforce this at publication: `you (will|would) (receive|get|claim)…`,
`(eligible|qualifies) for …(SFI|ELMS)`, `you will receive …[0-9]`.

## 3. Inputs

| Input | Unit | Notes |
|---|---|---|
| Total agricultural area | hectares | drives the 3 ha eligibility check and the 25% limited-area cap |
| Per-action quantity | ha, 100m, m², ponds, plots, tonnes | **unit follows the action**, see §5 |

No "farm size auto-fills every action" behaviour. The old tool had a `farmSize` field with an
empty auto-fill handler; either make it drive something real or leave it out.

## 4. The action set — all 71, grouped

Sourced from the register; **do not hard-code a rate anywhere in the tool that is not a registered
fact** (`bugs_open/288`: a figure a calculator encodes is checked by nothing). Families as
published, 71 actions as of 2026-08-22:

`AHW`×10 · `UPL`×7 · `WBD`×6 · `OFM`×6 · `OFC`×5 · `GRH`×4 · `CAHL`×4 · `PRF`×3 · `CIPM`×3 ·
`CIGL`×3 · `SPM`×2 · `SOH`×2 · `SCR`×2 · `HEF`×2 · `CSAM`×2 · `CNUM`×2 · `BND`×2 · `BFS`×2 ·
`AGF`×2 · `CLIG`×1 · `CHRW`×1

Offering the full set is the upgrade: the old tool's eight actions could not show a stack worth
discovering. Group them, let the reader filter, and show each rate **with its unit** on screen —
which also satisfies the `bugs_open/288` control, since the tool's visible table is the copy the
claims scanner can actually read.

## 5. Payment units — seven shapes, not one

The old tool assumed per-hectare or fixed. SFI26 has **seven** distinct shapes (as of 2026-08-22):

| Shape | Count | Example |
|---|---|---|
| per hectare | 63 | CSAM2 £129 per hectare |
| per 100m **for one side** | 2 | CHRW2 £13, BND2 £11 |
| per 100m **for both sides** | 2 | BND1 £27, WBD2 £4 |
| per square metre | 1 | HEF1 £5 |
| per pond | 1 | WBD1 £257 |
| per plot | 1 | AHW4 £11 |
| per tonne | 1 | AHW2 £732 |

**"For one side" against "for both sides" is a factor-of-two error waiting to happen** on a
boundary length, in the farmer's favour on screen and against them at audit. It must be visible in
the UI, not buried in a tooltip. This is exactly why those four rates are registered **without** a
numeric `value` — a compound rate reduced to one number loses the qualifier.

## 6. Constraints the tool MUST model — the old one modelled none

**Scheme-level** (all registered, citation-verified):
1. **£100,000 annual agreement value cap.** The application service refuses more. A stack that
   exceeds it must show the cap binding, not a bigger number.
2. **3 hectare minimum** to be eligible to apply at all.
3. **One SFI26 agreement per farm business.**
4. **Limited-area actions ≤ 25% of total agricultural area**, in any combination.

**Action-level, embedded in the published rate string** — three cases:
- **AHW2** supplementary winter bird food: max 1 tonne per 2 hectares **of CAHL2** — a cross-action
  dependency, so AHW2's ceiling depends on another action's area.
- **WBD1** manage ponds: max 3 ponds per hectare.
- **AHW4** skylark plots: **minimum** 2 plots.

(`UPL5`'s "minimum" sits in its action *name*, not its rate; its rate is a clean £18/ha.)

A stacking calculator blind to these totals up figures the scheme would never pay, which is the
single most damaging thing this tool could do.

## 6b. Every rate carries a link to its source (owner instruction, 2026-08-24)

Each action row shows its rate **and an anchor to the document that states it** — for all 71,
that is the SFI26 scheme rules on GOV.UK, with the capture date shown. Every fact in the register
now resolves a structured URL at `source.citation.url`, so the tool builds the link from data
rather than from a hard-coded string.

Three constraints on how, each earned:

- **HTML anchor, never markdown.** No markdown renderer exists in this platform, and
  `check_literal_markdown` strips `[text](url)` from text-typed fields as a defect. A markdown
  citation disappears silently and leaves the figure looking sourced when it is not. A tool's
  `html_template` is raw HTML, so this is free here — but it is the reason the same rule must be
  stated for the explainers.
- **This is also the `bugs_open/288` control.** A figure a calculator encodes is checked by
  nothing; the claims scanner reads a page's visible words, never its code. Putting each rate and
  its source in the *visible* rate table is what brings the tool's numbers inside the gate at all.
- **Show the capture date next to the link.** Rates move — that is the whole reason this rebuild
  exists — and a link with no date invites the reader to assume it is current for ever.

Do not attach a confidence claim to the citations. They let a reader check us; they do not make
us right, and the site says so.

## 6c. Code-level checking: what covers our rates, and what provably does not

Updated 2026-08-24 from measurements run by the `bugs_open/288` lane against this site's live
register. This section exists because I asked them whether their fix made my visible-rate-table
control redundant. **It does not, and the answer is quantitative.**

**Their code probe refuses our figures, by design.** Phase 3a reads a tool's `<script>` out of
stored `rendered_html` and looks for a registered value as a guarded literal. It has a
*measured* distinctiveness floor — false-positive rates over the script text of all 161
script-bearing tool pages, using invented values so every match is a false positive by
construction:

| digits | 1 | 2 | 3 | 4 | 5+ |
|---|---|---|---|---|---|
| FP rate | 32.75% | 3.79% | 0.06% | 0.03% | 0.00% |

The floor is **1000**, and below it the probe refuses rather than guesses. Against our register:
**75 of 105 facts refused, 25 have no numeric value, 5 probed.** Every figure in this tool —
382, 224, 45, 129, 20, 13, 10 — is under the floor. Lowering it would not rescue us: our rates
cluster at two and three digits, and 3.79% wrong on a rate table is noise that teaches people to
ignore the check.

**So §6b's visible-rate-table requirement is not belt-and-braces. It is the only thing covering
this tool's shape**, and it must not be dropped.

**A second, structural gap on their side:** their suggester's population is
`page_type='tool' OR component_level='tool'`, and agritec currently has **zero** tool-level
components. A calculator embedded in an article would never be selected — fleet-wide that class
is 50 of 222 script-bearing pages (23%). Another reason this tool must be built as a real
`page_type='tool'` with a `component_level='tool'` component, not as a script inside a
`blog-post`.

### What we DO gain: a per-fact `artifact_check` on each rate

`artifact_check` became reachable for **citation facts** in their Phase 2 (RFC_025 stage 2b) —
previously the per-fact loop handled `source.citation` and `continue`d before ever testing it,
which is why it had 0 consumers of 294 facts — and it is now addressable by `subject_key` rather
than a `page_components.id` that dies on decomposition.

So each rate this tool encodes gets, **on the fact itself**:

```json
"artifact_check": {
  "subject_key": "<this tool's subject key>",
  "pattern": "HERBAL_LEYS_RATE\\s*=\\s*224",
  "must_be_present": true
}
```

**The pattern must carry context.** A bare `224` is refused by the platform's own guard, and
correctly — that refusal is exactly the floor above. A human-authored pattern is admissible
because the constant NAME does the discriminating. This is what makes a sub-1000 legislated
figure checkable at all, and it is per-fact, so we only pay for the rates that matter.

**Two cautions, both from them, both load-bearing:**
- **A green acceptance run on a fence carrying `facts` means nothing about the figures.** Both
  tiers ignore the key by design; only the nightly `evidence-freshness` sweep reads it. Do not
  read a Tier-4 pass as evidence the rates are right.
- **None of their four phases is live** — all Go, inert until the next chassis roll — and Phase
  3a is **annotation only** by construction. It records what it saw and changes no routing.

### We are the live proof, deliberately

They asked, and it is worth saying yes: this is the one case where a tool and a register are
known to disagree *today*, so it is the only real test of whether the mechanism fires rather than
reads green. **The result we most need is the negative one** — if the first sweep after the roll
does not flag a tool encoding £382 against a register saying £224, the mechanism is inert on the
case it was built for, and that is far better learned here than from a synthetic fixture.

Sequencing, because it matters: they cannot see this tool's code today (they probed — 0 pages
containing `Math.min`, 0 containing `382`), for the simple reason that it does not exist yet.
Neither lane may quote their probe's verdict on this tool until it is built and deployed.

## 7. Output

Per-action subtotal, a stack total, and — separately and prominently — **which caps bound the
result**. "£112,400 before the £100,000 cap; £100,000 payable" is honest and useful; silently
showing £112,400 is not, and silently showing £100,000 with no explanation is worse.

Show the capture date of the rates on the page: *SFI26 scheme rules, GOV.UK, captured 22 August
2026*. Every registered fact's `writer_line` already carries it.

## 8. Acceptance criteria

For `tool-acceptance-agent` (Tier 4, headless Chromium). Steps limited to fill/click/select.

```criteria
- fill total agricultural area with 100 and select action CSAM2 with 10 hectares; the CSAM2 subtotal shows 1290
- select action CSAM3 with 10 hectares; the CSAM3 subtotal shows 2240 and NOT 3820
- the page contains no line item named "management payment"
- fill total agricultural area with 2; the tool shows the 3 hectare eligibility floor is not met
- select action CHRW2 and confirm the displayed unit reads "per 100m for one side"
- build a stack exceeding 100000 and confirm the payable total is capped at 100000 and the cap is named
- confirm each visible rate row carries an anchor whose href points at gov.uk, and that the capture date appears on the page
```

The second and third are the regression tests for this whole exercise: £382 must not come back,
and the abolished payment must not reappear.

## 9. How it gets built, and what is blocking

Through `tool-generator` / `create_tool_component` — self-contained HTML, inline `<script>`,
`component_level='tool'`, `function` prefixed `tool-`. Not hand-written: the 2026-08-04 ruling.

**Blocked on the site existing.** agritec.uk has 0 pages as of 2026-08-22. `tool-generator`'s
`load_brand_context` reads the site's specs, and `design_intent` is deliberately unseeded so the
classifier can do its own design research (PLAN Phase 1). Building the tool now yields one that
matches no site. **Order: submit (Phase 3) → design cascade → build this tool (Phase 4).**

**Carry the honesty rules in the tool brief itself.** Measured 2026-08-22: `tool-generator` and
`tool-improver` receive the evidence register but neither prompt names `writer_block`, evidence,
fact, source or invent — unlike `page-content-writer` and `tool-recreation-handler`, which name
`writer_block` explicitly. The register will not instruct this build; the brief must.
