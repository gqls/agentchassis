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

**G2 — nothing judges the claim. [REVISED 2026-07-30 after the owner's
correction, which was right and changed the design.]**

My first draft proposed teaching `tool-auditor` to load the PLAN and judge
delivery against promise. The owner's objection: *"I think that probably has
value — maybe for where and when it runs especially — and perhaps could be
improved but not change its function? I think a judgement agent is a different
thing and introduces more flexibility than a hard coded rule filter. Maybe we
could do both but we shouldn't confuse validation with judgement."*

That is a better decomposition and the report now follows it. Two separate
things were collapsed in my draft:

- **Validation** answers a closed question with a fixed rule: does this markup
  label its inputs, are targets ≥44px, are there hardcoded hex values, does the
  script reference an id that exists. The answer is the same every time it is
  asked. `tool-auditor`'s six-category checklist IS this, and it is valuable
  precisely *because* it is fixed — it is comparable across 63 tools and across
  months. **Its function should not change.** What is worth improving is
  **where and when it runs**: today it fires only after Tier 1 passes with no
  blockers, on a 30-day cooldown, off a discovery pass nothing schedules. A
  checklist that runs at birth, at repair, and after any markup edit is worth
  far more than the same checklist made cleverer.
- **Judgement** answers an open question where the answer depends on the case:
  does *this* tool deliver what *this* PLAN promises? Is a fluid-typography
  preview that is technically correct and inert at desktop widths acceptable?
  That cannot be a rule, because the interesting part is the specific promise.

So **G2 splits**:

- **G2a — leave tool-auditor's function alone; fix its cadence.** Add it to
  the birth path and the post-repair path, shorten or scope the cooldown. No
  prompt surgery. Cheap, low-risk, and it makes an existing asset earn more.
- **G2b — add a claim-judgement seat, separately.** Its input is the PLAN's
  promise (and the promise ledger, if the experience loop has written one) plus
  the delivered artefact; its output is a judgement with reasons, not a
  pass/fail against a fixed list. The platform already has the shape for
  this — the council seats are exactly judgement-not-validation, and
  `diagnose_council_decide` already turns seat opinions into a decision
  deterministically. A `review_claim_delivery` seat is a nearer neighbour to
  what is wanted than a modified auditor.

The distinction generalises past tools, which is why it is worth stating
plainly: **G3 is validation** (a fence linter — fixed rules, same answer every
time) and **G2b is judgement** (is this promise honoured). Conflating them
would have produced an auditor that is worse at both: a checklist that drifts
per tool, and a judgement constrained to six categories.

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

**G2c — the criteria fence is authored AFTER the tool is built, from the
finished HTML.** That ordering is why a tool can be born already broken: the
criteria describe what was produced rather than what was required. §7c sets out
the staged alternative — claim first, then build to it, verifying each stage.

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

**Outcome: PASSED, first complete run.** (Correlation `c258967d`; an earlier
attempt, `fcb58019`, failed inside the G6 fix itself — a Postgres
type-deduction error `go build` cannot see — which is exactly what a pilot is
for. Corrected, proven by PREPARE/EXECUTE against the live schema, rolled as
v1.0.1206, re-fired.)

The verdict note, verbatim from `doc_notes` (`source='tool-acceptance'`,
category `acceptance-run`):

> Tier-4 acceptance PASSED — smart-contrast. Observed: all 11 of the tool's
> own checks passed in headless Chromium across profiles: desktop, mobile
> (1 skipped: mobile-fit@desktop). Verified: browser-runner-adapter run;
> checks: boots@desktop, status@desktop, **claim-aa-boundary@desktop**,
> **claim-maximum@desktop**, console@desktop, boots@mobile, status@mobile,
> mobile-fit@mobile, **claim-aa-boundary@mobile**, **claim-maximum@mobile**,
> console@mobile

Zero work items raised — nothing to fix, and the correctly-skipped
`mobile-fit@desktop` shows the profile scoping working rather than a vacuous
pass (the note counts 11 passes, not "no failures").

**What this proves, precisely:** the platform's own acceptance machinery —
dispatched through its own agent, not a session harness — drove a PORTED
tool's real controls in a real browser on two profiles and asserted its
**arithmetic** against known answers. The chain from `doc_plans` fence to
browser to verdict note runs end to end for the estate the ladder was widened
to see. The bar-3 layer is not merely automatable; as of this run, it is
automated. What remains is authoring: 50 tools still need a fence that states
their claim, and the G1–G4 wiring so the loop reads, lints and reasons about
those claims without a human in the middle.

---

## 7b. Check DEEPER — added 2026-07-30, after the owner broke this report twice more

Between drafts the owner opened two tools this workstream had declared repaired
and found both unusable. Neither was a near miss:

- **micro-cms** — the starter copy (mine) said *"Click anywhere and start
  typing. Everything on this page is editable."* Measured: the editable body
  was 248px inside a 743px frame, so `elementFromPoint` returned HTML — not
  BODY — for the bottom two thirds of the visible editor, and a query for
  formatting buttons returned an **empty list**. designMode was on and
  `execCommand('bold')` worked when called directly; there was simply no way
  for a visitor to reach it.
- **pasteboard, logic-architect, mind-map** — work areas measuring **1146x0**.
  Present in the DOM, invisible. One cause for all three: each was ported from
  a standalone page whose `body` was the flex container, so `flex: 1` had no
  flex parent and the height collapsed — while that same `body` rule restyled
  the host page (`display:flex`, `overflow:hidden` on the site's own body).

**How they got past me, stated exactly, because the method is the finding:**
I verified pasteboard by calling `addItem(src)` — an internal function. I
verified logic-architect by calling `loadTemplate('code')`. Both returned the
right answer, so both "passed". **A visitor cannot call a function.** Their
entry points are a paste event and a click on a visible control, and the areas
those act on had no height.

### The two rules this produced

1. **Verify through the visitor's gesture, never through the tool's internal
   functions.** If the entry point is a paste, dispatch a paste. If the
   vocabulary cannot express the gesture, that is a MISSING CHECK TYPE to
   record as a deferral — not a licence to substitute a function call. (Both
   repaired tools were then re-verified this way: a synthetic paste event
   carrying a PNG produces a sticker and hides the empty state.)
2. **A fence must assert the tool's TERMINAL value, not the first observable
   state change.** "Status reads LIVE EDITING" is a waypoint; "text can be
   edited and emphasised" is the point. My micro-cms fence asserted the
   waypoint and passed while the tool was unusable.

### What was built, so the rule is enforced and not merely written

**TL-034 `has_visible_area`** — a Tier-4 check type measuring
`getBoundingClientRect()` against a floor (default 24x24; per-check
`min_width`/`min_height`). It fails on a collapsed box and on a missing
element, and it is **Tier-4-only by necessity**: it measures rendered layout,
which no static read of HTML can compute. A Tier-2 equivalent is not unbuilt,
it is impossible — which is the sharpest available answer to "why does the
ladder need a browser tier at all".

**Measured before shipping:** exactly 3 pages fleet-wide carry page-scope CSS
that escapes onto the host page — these three. That is a second, static defect
class worth a check of its own (a `<style>` inside a section that targets
`body`/`html`/`*` with layout properties is always wrong); it is named here and
not yet built.

**Where this lands in the chain:** it is a **validation** primitive (fixed
rule, same answer every time), so it belongs beside G3 and not beside G2b. It
does not judge whether a tool is good; it refuses to call an invisible element
present.

---

## 7c. Building a tool in testable stages — the owner's recollection, and what I could find

The owner: *"somewhere in the docs I have previously discussed not necessarily
building a tool all at once but in testable stages, it may or not be in the
concept register or maybe somewhere else, I can't find the discussion."*

I could not find that specific discussion either, and I want to be straight
about that rather than substitute the nearest thing and call it found. **A
targeted search agent for it was cut off by a session limit before reporting**,
so this section rests on what my own greps turned up. Two clear statements of
the principle exist, both owner-authored in substance:

- **`022_dynamic_applications.md` §5 "Incremental complexity"**, verbatim:
  *"Start with the simplest version that works. A contact form starts as a
  mailto: link. Then it becomes a Formspree integration. Then a Cloudflare
  Worker that stores to D1. Then a full backend with CRM integration. **Each
  step is a separate work item, not a big-bang rewrite.**"* — the principle
  exactly, stated for capability tiers rather than for one tool's construction.
- **`FOCUS_interactive_content_generation(4).md` §"Path C — work breakdown"**,
  verbatim: *"The full Path C work is split into six incremental steps, each
  independently verifiable. **After each step we pause and check the output
  before moving on.**"* — the *method* the owner is describing, applied to
  building the interactive-extraction pipeline. C1..C6 are even numbered and
  status-tracked individually.

Also adjacent: **DYN-006** (tool builder tiers: static / dynamic /
application — a triage vocabulary for what an LLM may be asked to generate at
all) and **AGOV-009** (*"thin vertical slice before the six-contract
infrastructure"* — the same instinct at platform scale, and marked deployed).

**Why it matters to this chain, concretely.** Today a tool is born in ONE LLM
pass (`generate_tool_html`), and its PLAN and criteria fence are written
*afterwards*, from the finished HTML. That ordering is what lets a tool be born
already broken: the criteria describe what was produced rather than what was
required. The staged alternative inverts it — **write the claim first, then
build to it, verifying each stage before the next**:

1. **Skeleton** — the markup contract only: every id and control the tool will
   need, with a fence of `selector_exists` + `has_visible_area`. Verifiable
   before a single line of logic exists, and it would have caught every one of
   the fifteen defects repaired in this workstream, because every one of them
   was missing or invisible markup.
2. **One real behaviour** — the tool's single most important claim, as an
   `interaction` check with a known answer (smart-contrast's `#767676 → 4.54`).
   Verify. This is the stage that proves the thing works at all.
3. **The rest of the behaviours**, one criterion at a time.
4. **Polish** — mobile fit, console cleanliness, accessibility.

Each stage is a work item with its own verdict, which is exactly the shape the
existing machinery already has (`improve_tool` items, per-check verdicts,
`max_fix_attempts`). **Nothing new is needed to run this except authoring the
fence in stages rather than at the end.**

**Open question for the owner**, and the reason this section says "what I could
find": if the discussion you remember is a third document — particularly one
that stages *a single tool's build* rather than a pipeline's — it will have
detail these two lack, and it should supersede this section. I have not found
it; the search is worth one more pass with fresh session budget.

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
