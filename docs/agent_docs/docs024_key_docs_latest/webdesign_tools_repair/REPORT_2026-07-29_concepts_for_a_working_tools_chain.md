# The concepts that compose into a working-tools build chain

**2026-07-29, for the owner.** Written after three deep sweeps — the concept
register, the travelling-docs machinery, and the experience/fix loops — and then
a second pass that **adversarially re-checked every load-bearing claim against
the live system**. That second pass mattered: it corrected four findings, found
one new gap, and got that gap fixed and rolled before this report was finished.
Where a claim below is verified, the evidence is named; where it is not, it is
marked.

The instruction this answers: *"search hard for the concepts in the concept
register and elsewhere and start a report on what concepts would touch the
tools and in which way. Together they will almost certainly be enough to help
us create a working workflow to build working tools."*

**The answer, up front: you are right, and my earlier "can't be automated" was
wrong.** The platform already holds nearly the entire chain, live. The distance
between "responds to clicks" and "asserts its actual claim" is **five small
pieces of wiring** — one of which was fixed during the writing of this report —
plus a per-tool authoring campaign that the existing composer rules already
describe. Nothing fundamental is missing.

---

## 1. What "working" means — three bars

| bar | question | who tests it today |
|---|---|---|
| **loads** | does the page render without throwing? | Tier 1 `tool_health`, `orphan_element_refs` (TL-032), the audit harness |
| **responds** | does driving a control change something? | the session harness (`toolaudit.py`); Tier 4's generic checks |
| **delivers its claim** | does the contrast tool compute the *right* ratio? does the fluid-type tool *show* fluidity where visitors are? | **the ```criteria fence, and nothing else.** Tier 2 checks it statically, Tier 4 behaviourally, Tier 3 could reason about it (gap G2) |

Both of today's owner-reported failures were bar-3 failures that passed bar 2:
`micro-cms` "responded" while unusable; `fluid-typography` was *correct CSS*
that demonstrated nothing at desktop widths. **No liveness check can see bar 3.
A claim check can, and the vocabulary for it already exists** — the pilot in §7
asserts smart-contrast's actual arithmetic with known-answer pairs
(`#767676` on `#ffffff` → `4.54 : 1`) using only check types that are live today.

---

## 2. The chain that already exists (all LIVE unless marked)

```
BIRTH   tool-generator: generate → compose_plan → write_plan → index_plan
        → enqueue_rerender                                   [migs 131/158/162;
        the PLAN's ```criteria fence is written AT BIRTH      enqueue_rerender
        by an LLM under composer rules corrected 4 times]     VERIFIED in the
                                                              live agent row]
  |
TIER 0  check_tool_completeness + HasToolDocHeader (TL-007)   flags-but-passes
TIER 0.5 check_tool_fabrication (mig 189)                     synthetic-data → HITL
        DeployToolToSiteAction orphan-ref refusal (TL-032)    cannot fire on any
                                                              current template
  |
TIER 1  check_tool_health — structure, script/style, deps     → improve_tool
TIER 2  check_tool_acceptance — the fence, statically,        → improve_tool
        under the ANCHOR RULE (confirm, never refute;
        the two attribute checks are the ruled exception)     [TL-013, TL-031]
TIER 3  tool-auditor — Sonnet, six categories, confidence-    → improve_tool /
        routed                                                  needs_human_review
TIER 4  tool_acceptance_due → tool-acceptance-agent →         → improve_tool,
        browser-runner-adapter (real Chromium, desktop +        convergence guard
        mobile, interaction checks, overflow attribution)       (2 cycles → human)
  |
FIX     tool-improver → section_edit → section-editor         [mig 195 — the
        → redeploy → re-verified by the due-sweep (≤7 days)    needs_rerender
                                                               path is a dead
                                                               end on owned pages]
```

Supporting machinery around the spine, each with its register ID:

- **Travelling docs** (TL-017): the criteria live in the tool's `doc_plans` PLAN
  as a fenced block; NOTES accumulate every repair. *Writing the PLAN is the
  act that makes a tool testable* — for a ported tool it is the only act
  (TL-033).
- **Experience register** (PLAN-043..046): parameterised claims
  (`criteria_template` + `binding_schema`), shared invariants held once
  (`no-inert-control`), a ten-rule criteria validator, and the only
  anti-vacuous verdict function on the platform. Schema + write path + council
  LIVE; `bind`/`verify` deployed but **called by no workflow**;
  `apply_experience_verdict` **does not exist** [verified: zero grep hits], so
  the register lifecycle is blocked at its first gate.
- **Experience loop**: the promise ledger (CTA copy → what the destination
  must deliver), journeys with observable outcomes, the honesty council
  ("a not-yet feature must be absent or labelled coming-soon, never
  simulated"). One plan approved to date.
- **Diagnosis→fix loop** (FIX-003, FIX-002, FIX-014..025): for when a tool
  defect's *cause* is in the chassis — cite-or-abstain diagnosis, constrained
  fix plans, 17-seat council, build-gated PRs. This is what repaired today's
  chassis-side gaps.
- **Immune system** (FIX-051..053): triage keeps operational noise out of the
  diagnosis queue; the silent-check finds the class no work item records.
- **Interactive fingerprint** (DYN-002): the only concept that captures what a
  SOURCE tool claimed to do (canvas, handlers, library signals, a type_hint) —
  action exists; the C2–C6 workflow integration reads as never shipped
  [UNVERIFIED whether any orchestration calls it].
- **The session harness** (`toolaudit.py`, this workstream): nine checks
  including the two classes nothing platform-side covers — external-script
  reference resolution and viewport-band scaling. A candidate donor for
  TL-019's behavioural QA loop, which is **aspirational** (nothing built).

**Marked NOT BUILT, so nobody plans against them:** TL-019 (behavioural QA
loop), TLIB-013 (known-good solution library), GML-001 (games parity), DYN-002
C2–C6, DYN-011 (loader-builder), TL-009/TL-010, the appeal-dimension check
types (`computed_style`, `contrast_ratio`, …), `apply_experience_verdict`.
**Superseded, never revive:** TL-021, TLIB-009 (tag matching deployed a
password checker to a gas wholesaler), TLIB-010, TLIB-019, and the
`asset_loads /tools/assets/*.js` criterion class (migs 143/144).

---

## 3. The criteria vocabulary — what a claim can say today

From the two evaluators' own switch statements (not from docs):

| check type | Tier 2 (static) | Tier 4 (browser) |
|---|---|---|
| `selector_exists` / `selector_count` | anchor rule | live DOM |
| `interaction` (fill/click/select + expect text_matches) | anchors only | **runs it** |
| `page_status_ok` | fetch 200 | nav 200 |
| `asset_loads` | string presence | — (skipped) |
| `attribute_absent` / `attribute_matches` | **refuting**, zero-matches-SKIPS | — (skipped) |
| `no_console_errors` | — (skipped) | whole-session, evaluated last |
| `no_horizontal_overflow` | — (skipped) | with culprit attribution |

`interaction` + `text_matches` is already enough to assert most of these tools'
claims — a calculator's known-answer pair, a generator's output naming its
inputs, a formatter's before/after. That is what the pilot uses. What the
vocabulary cannot yet say: computed-style/appeal assertions, event-listener
presence, fault injection, multi-page journeys (all named, none built).

---

## 4. The wiring gaps — the whole distance, measured

Each verified against the live system on 2026-07-29, not read from docs.

**G1 — tool-improver never reads the PLAN or NOTES.** The docs say it does in
four places, including the judge's own note text. The live workflow has no
`load_doc_context` step [verified: step list read from `agent_definitions`;
only `tool-acceptance-agent` carries that action]. Sharper: the improve_tool
item's spec *carries* the failing criteria as `acceptance_test`, and dispatch
passes spec whole — **the prompt just never references it** [verified: prompt
consumes `{{.input_data.issue}}` only]. So the fixer repairs against a prose
issue line while the machine-readable claim, the PLAN's *deliberate decisions —
do not re-fix*, and every prior repair note sit unread. Fix: one prompt line
(`acceptance_test`) + one migration (a `load_doc_context` step). This is the
highest-value single change: without it, the fixer can undo a deliberate
decision every time it fires — and the PLANs written this week carry exactly
the traps it would walk into (pasteboard genuinely defines `saveHistory`;
vibe-equalizer's output must stay a TEXTAREA).

**G2 — Tier 3, the only tier that reasons, never sees the claim.** tool-auditor
reviews `html_template` against a generic six-category checklist [verified: no
doc/criteria/claim term in the live prompt; model claude-sonnet-4-6]. Teaching
it to load the PLAN and judge "does this code deliver what this tool
promises?" converts the LLM audit from code review into claim assertion — the
only tier that could have caught fluid-typography's "correct but demonstrates
nothing" *before* a human did. Fix: one migration (a load step + a prompt
paragraph). Consider Sonnet 5 while touching it.

**G3 — criteria fences are never linted.** 23 current fences fleet-wide
[measured now — an earlier figure of 78 was a point-in-time RFC measure; never
quote a count you didn't re-run]. They are LLM-written, and TL-016 records the
composer inventing selectors twice. The experience register's ten-rule
validator (P1–P10: placeholder closure both ways, no invented `-EDIT` ids, no
fields no checker reads, an attribute check must assert something) exists,
exported, and is applied only to `experience_patterns`. Routing `write_doc_plan`
(or a Tier-2 pre-pass) through it closes the invented-criteria class
mechanically — the remedy TL-016 already concluded was "validation, not
sterner prompts."

**G4 — the Tier-4 judge can pass on nothing.** It guards the no-criteria and
no-results cases [verified in code], but a result set where every check was
individually *skipped* — all `-EDIT` placeholders, all Tier-2-only types, all
wrong-profile — yields `len(Failed)==0` → **PASS note + 7-day cooldown**. The
experience register's `experienceVerdict` ("≥1 PASS and 0 FAIL, else
*inconclusive*") is the anti-vacuous rule, one exported function away. Note:
changing the judge's verdict semantics changes what a shared mechanism
GUARANTEES — under RFC_002's ratified rule that goes to the council gate, and
plausibly an RFC. G1/G3 are additive and do not.

**G5 — enablement is where composed workflows die, and one loop is off ON
PURPOSE.** The improvement-sweep schedule is disabled [verified:
`enabled=f`, last scheduled completion 05-08; fired manually today] — and the
owner has ruled the improvement loop stays stopped, so **this report does not
recommend re-enabling it.** What matters instead: the acceptance checks
(`tool_health`, `tool_acceptance`, `tool_acceptance_due`) are wired to
**design-discovery-agent** [verified]; `orphan_element_refs` is on
completeness-discovery-agent, which no enabled schedule targets [verified] —
so it fires only on manual sweeps. The register's "continuous sweep not in the
binary" (TL-014) is **stale**: `tool_acceptance_due` is in the running
v1.0.1205 binary [pod-verified]. The honest statement: the machinery runs when
a discovery pass runs, and discovery passes are currently manual-fire. Any
composed workflow must say who fires them and when.

**G6 — found by this report's own challenge pass, fixed, and live.** The
morning's eligibility widening (TL-033) taught the *discovery* tiers to key a
ported tool by its page name minus `tool-`; but `request_browser_run` still
resolved pages by `name = function` — so every acceptance run for a ported
tool would have hard-errored at the ladder's first attempt to look at the
population it was widened to see. Fixed (two-candidate lookup, exact name wins),
rolled as **v1.0.1205**, pod-verified both replicas. The pilot below runs
through this fix.

---

## 5. The composed workflow — what to wire, in order

1. **G1** (one prompt line + one migration) — the fixer reads the claim and the
   deliberate decisions before touching a tool.
2. **G3** (route fences through the exported P1–P10 validator) — no invented
   criteria can be written again.
3. **G2** (one migration) — the LLM audit judges against the claim.
4. **G4** (export `experienceVerdict`; adopt in the judge) — via the council
   gate, likely an RFC, because it changes a guarantee.
5. **The fence campaign**: 50 of 63 webdesign tools still carry no fence
   (13 written this week, incl. the pilot's). Author them under the composer's
   rules — *never invent a selector; describe delivered reality; watch every
   criterion pass before writing it* — and the widened ladder holds them from
   then on. This is authoring work, not construction: the chain that consumes
   them exists end to end.
6. **Then** the aspirational layer in value order: `apply_experience_verdict`
   (unblocks the whole experience register lifecycle), the appeal check types,
   TL-019's temporal/cross-browser loop (toolaudit.py is a working prototype
   of half of it).

**What the architecture council is for here** (the owner's framing is right):
G4 and any check-type additions that let a tier refute where it confirmed are
guarantee changes — RFC_002's ratified trigger. G1/G2/G3 are additive opt-ins:
normal council gate, register in the same commit, no RFC needed.

---

## 6. Constraints and hazards the workflow must respect

- **bugs_open/126**: Tier 4 navigates once per profile and state accumulates
  across checks — a consent-gated tool cannot pass, and the failure aims the
  fixer at a legally load-bearing disclaimer. Keep gated tools out of the
  campaign until fixed.
- **bugs_open/084 / TL-032's blind spot**: 13 of the 63 keep logic in external
  script files invisible to every DB-side static check. Their claims are
  Tier-4-only. (micro-cms's fence already reflects this.)
- **TL-011**: `tool_health`'s INNER JOIN reports "no tools" as a pass for a
  page with no component row — Tier 1 stays the weakest tier for this estate,
  and TL-033 deliberately did not widen it (noise).
- **TL-001**: a rebuild can still clobber a widget; the guard is live but
  carry-forward is unconfirmed. Today's evidence is mildly good: the 1203/1204
  rolls did not revert any of the nine DB-side repairs [checked].
- **TLIB-014/TL-027**: the quality auditor regenerates BY FUNCTION — all 63
  ported tools share `function='ported-page'`, so a regeneration aimed there is
  aimed at the shared blob. Treat as hazardous until ported tools get real
  component rows.
- **Tier numbering has drifted** across docs (the LLM audit is Tier 2 in older
  files, Tier 3 in the register) — cite mechanisms by name, not tier number.
- **TL-012, the standing argument**: "completeness + validation passed" ≠
  working — twice demonstrated. Quote it whenever a static pass is offered as
  proof.

---

## 7. The pilot — one tool through the whole chain

**smart-contrast**, chosen because its claim is pure arithmetic and therefore
honestly assertable with today's vocabulary, and it has no consent gate.

What was done, in order:
1. Watched the claim pass by hand on the live page (never author a criterion
   you have not watched pass): `#767676`/`#ffffff` → `4.54 : 1`,
   `#000000`/`#ffffff` → `21.00 : 1`.
2. Wrote its PLAN + fence (`SQL_t06`): the composer's four standard checks plus
   two known-answer interaction checks, selectors read off the live DOM.
3. Fired the Tier-4 run (087 trigger, correlation `fcb58019`), through the
   v1.0.1205 fix.

**Outcome:** *recorded below when the run completes.*

<!-- PILOT RESULT APPENDED AFTER THE RUN -->

---

## 8. Corrections this report made to its own inputs

The three exploration sweeps were themselves challenged before anything above
was asserted. What changed: the fence count (78 → 23, measured); TL-014's
"sweep not in the binary" (stale — it is, at v1.0.1205); G1 sharpened (the
criteria reach the fixer's inputs and are simply unread by the prompt); G4
narrowed (two vacuous cases already guarded; the all-skipped case is not);
G5 reshaped (the improvement loop is off deliberately — the report's earlier
draft would have recommended re-enabling something the owner ruled stays
stopped); and G6 found, fixed and rolled. Four register cross-reference
defects found during the sweeps are fixed in the same commit family as this
report.
