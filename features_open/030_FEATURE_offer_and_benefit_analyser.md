# 030 FEATURE — the offer and benefit analyser

**Raised:** 2026-08-02 by the owner, opening the `vigilant_designer_offer_analysis`
programme; **written up here 2026-08-08** after the owner asked how far the discussion had
got and then widened the scope (see §3).
**Status:** specified, not built. **Two pieces promoted ahead of the rest — see §7.**
**Owning lane:** `docs/agent_docs/docs024_key_docs_latest/vigilant_designer_offer_analysis/`
(Programme B of `PLAN_2026-08-02`; state review in `REVIEW_2026-08-08_…`).

---

## 1. The originating intent, in the owner's framing

> "a constant vigilant designer… and the offer analyser and benefit analyser… all of the
> above can also be made into checkers and handlers too for continual improvement. It needs
> a strong focus or we'll keep tripping up on the detail." — 2026-08-02

> "a dedicated offer and benefit analysis step which is probably a full blown agent and
> checker, handler and probably part of the council, and probably needs to correspond with
> several other agents e.g. copywriting, design, planning, imagery, tool designer,
> experience loops, spec, etc etc at several points." — 2026-08-08

The question it exists to keep asking, per site, for ever: **does this site actually answer
its target market's need, in a way that pays us?** Not "is it well built" (the designer's
question) and not "is it true" (the claims auditor's question) — *is the offer right, and is
the benefit to the visitor visible?*

## 2. The gap

Every site gets a written monetisation analysis and audience definition at build time, and
**nothing ever reads them again.** The nearest live thing —
`site-review-agent.run_strategic_review` — runs inside every improvement sweep (16 LLM calls
in the 14 days to 2026-08-08) and asks genuinely offer-shaped questions, including *"what
single change would most improve conversion?"*. It asks them with **domain, company name,
`dream_spec`, the site plan and two counts** in context. No strategy. No audience. No
identity. No content direction.

So the platform already asks the offer question, blind, sixteen times a fortnight.

Meanwhile doc 028 states the doctrine — *"the revenue model shapes the site, not the other
way round… defaulting to the consultancy shape when the signal is absent is a failure mode,
not a safe fallback"* — and **nothing checks any live site against it.** The council's
always-on `review_mission` seat enforces exactly that rule, on **platform code submissions
only**; it never looks at a site.

## 3. What is new here, versus PLAN §B

PLAN_2026-08-02 §B4 scopes the analyser deliberately small so it could ship: a **config-only**
agent — `load_premise → load_offer_surface → run_offer_analysis → write_findings` — emitting
**existing item types only**, routed to **existing handlers**, corresponding with nobody.

The owner's 2026-08-08 framing is larger on four axes. This file exists because that
difference is a decision, not a misunderstanding, and it had nowhere to live.

| axis | PLAN §B | owner's framing 2026-08-08 |
|---|---|---|
| agent | config-only reader | "full blown agent" |
| checker | 2 checks (§B3) | checker **and** handler of its own |
| council | not mentioned | "probably part of the council" |
| correspondence | none by design | 7 named counterparts, "at several points" |

### 3.1 The correspondence surface, measured

Of the seven counterparts the owner named, **two are wired today, three are wired but
fragile, and two have no route at all.** The last pair is where the actual design work is.

| owner's term | live counterpart | route today |
|---|---|---|
| spec | `domain-strategist`, `site_specs` | **wired** — `needs_strategy` is a live type (§5.2) |
| copywriting | `content-quality-auditor`, `content-reviewer`, `page-content-writer` | **wired** — `content_rewrite` / `tone_shift` |
| design | `design-audit-agent` → `visual-design-auditor`, `webdesign-agent` | fragile — `needs_design_review` exists; the taste critic (`018`) is unbuilt |
| planning | `build-site-planner`, `site-design-planner` | fragile — `needs_page_recompose` has **no handler built** (lane's A3) |
| imagery | `image-build-handler` | fragile — `needs_imagery` needs an imageryplan spec row, not a bare item |
| tool designer | `tool-auditor`, tool-lifecycle checks | **no route** |
| experience loops | `experience-planner`, `experience-approval-council`, `experience-register-writer` | **no route** |

### 3.2 "Part of the council" resolves into two different questions

- *"No platform change should quietly break the revenue-shape doctrine."* — **already true.**
  `review_mission` is one of 17 council-gate seats, always-on, and its prompt carries doc
  028's rule verbatim down to the "tools site with a Start a Project CTA" example. **Nothing
  to build.**
- *"No site should ship with an offer that does not answer its market."* — **a new seat, on a
  different council.** The precedent is `experience-approval-council`: site-facing, 5 seats
  (`checkability`, `deferral_honesty`, `honesty`, `observable_outcome`, `prior_art`), 36 LLM
  calls all-history, last run 2026-07-29.

These are not substitutes, and a third answer is legitimate: **an auditor is not a reviewer.**
An offer seat gates *a proposal*; the analyser judges *a live artefact*. Most of what the
owner described is auditor work. Open question, §8.

## 4. The hard constraint: we have no outcome data

The analyser can grade **the artefact against the stated premise**. It cannot grade either
against **what visitors actually did**, because the platform collects no engagement data —
the mission's own word "measured" is currently unbacked. PLAN §B4's prompt therefore forbids
"users want…" phrasing outright, and that restriction must survive into any larger version.

This is the single most important honesty constraint on the feature. An offer analyser that
sounds like it knows what converts, while reading only our own specs, would be the most
confidently wrong instrument on the platform. The cheapest future fix is named but not
scoped: the disabled intent-collection task.

## 5. Facts that bind the design (measured live 2026-08-08)

### 5.1 The premise substrate exists and is rich
`site_specs` current aspects: `audience` 29, `identity` 20, `content_direction` 20,
`strategy` 17, `evidence_base` 10, `vertical_landscape` 9, `mission_brief` 8, `cta` 4,
`portfolio` 4, `commercial` 2. The `portfolio_positioning` lane measured (2026-07-31, their
figure) that `audience` — the most populated aspect of all — **is read by nothing.**

### 5.2 `needs_strategy` is already live, with a producer and a handler
Three rows, all `complete`: lendzy.co.uk and mortgagecalculator.co.uk (08-02), webdesign.uk
(08-04). Producer `vertical-exemplar-researcher`; handler `domain-strategist` — the same
handler PLAN §B3 names.

**Governance consequence.** A `check_premise_incomplete` makes this lane a **second producer
converging on an existing `item_type`**, which is squarely the **owner ruling of 2026-08-02
(RFC_010 §1)**: no architecture round needed, *provided* the concept-register entry names the
full producer set and states the shared `item_key` shape. Cheaper than the PLAN assumed, but
the register entry is mandatory, not optional.

### 5.3 The recorded revenue shapes — a real population for `check_revenue_shape`
`revenue_models.primary_model` across the 17 current strategy specs:

| primary_model | sites |
|---|---|
| `direct_business` | **10** |
| `saas_tools` | 3 |
| `display_advertising` | 2 |
| `lead_generation` | 1 |
| `sponsored_listings` | 1 |
| *(absent — old shape)* | 1 (gaswholesalers.com) |

**10 of 17 sites are recorded as the consultancy shape** doc 028 names as the failure mode
when the signal is absent. **This is not a claim that 10 sites are misclassified** — several
genuinely are businesses (finetuning.uk, leopardessconsulting.co.uk, webdesign.uk, oufe.com).
It establishes that the check has a real population, a day-one question (*does each site's
CTA lexicon match its own recorded shape?*), and that the disconfirming answer — "all 17
agree, nothing to check" — was available and is **not** what came back. Start with the
`direct_business` rows whose domain reads as a topic or a tool rather than a brand.

> **CORRECTION carried in from the review, so nobody re-derives it.** An earlier draft of
> this analysis stated that `primary_model` did not exist on any site and called two lines of
> PLAN §B a defect. **False** — it exists on 16 of 17, nested one level down under
> `revenue_models`; the PLAN is correct and must not be "fixed". The error came from reading
> `data->>'primary_model'` at the top level and enumerating only top-level keys.
> `WRONG_CALLS.md` 2026-08-08 has the class and the check.

### 5.4 The strategy shape has drifted, and the better shape is the older one
16 sites carry the current 12-key shape (`value_proposition`, `revenue_models`,
`competitive_position`, `growth_path`, `search_intent`, `site_type`, `domain_type`, `tone`,
`content_strategy`, `recommended_page_types` + two `_reasoning` fields).

**One site — gaswholesalers.com, written 2026-04-17 — carries a completely different 6-key
shape**: `satisfaction_condition`, `trust_threshold`, `recurring_value`, `monetisation`,
`primary_intent`, `visitor_type`.

Those are, almost exactly, the four "Q-fields" PLAN §B2 proposes to add. **So B2 is a
restoration of an abandoned shape, not an invention** — an easier argument, and there is one
live worked example to copy from rather than a blank page. This is the genuine plan
correction (the `primary_model` one was not).

> **UPDATED 2026-08-12 — B2 SHIPPED and the restoration is working; the counts above are
> stale, and the new split is a decision for §7 rather than a finding for §5.**
> `satisfaction_condition` + `trust_threshold` + `recurring_value` are now on **7 of 22**
> current strategy rows, and the boundary is a **vintage, not a property of the sites**:
> **every `domain-strategist` row written 08-08 or later carries them (6 of 6); every one
> written 08-02 or earlier does not (13 of 13).** The 7th is loanandmortgagecalculator.co.uk,
> written by another lane's oneshot. So "one live worked example to copy from" is superseded —
> there are seven, all current, including gaswholesalers.com itself, which was re-written by
> the strategist on 08-11 16:19 and now carries BOTH shapes.
> ⚠ **The 15 without are not uniformly refreshable.** Thirteen carry strategist-written rows
> and a refresh is the safe operation B2 exists to allow. **Two carry human-authored current
> specs** — `mortgagecalculator.co.uk` (`source='owner_direction'`, 08-11, the owner's own
> voice direction) and `leopardessconsulting.co.uk` (`source='hitl'`) — and a refresh writes a
> new `is_current` row over the top. Any "refresh the estate" task must exclude those two by
> `source`, not by hand. Measurement, options and the recommendation are in
> `PLAN_2026-08-02_vigilant_designer_offer_analysis.md`, decision log 2026-08-12.

### 5.5 A live hazard blocks any premise-refresh automation
`domain-strategist`'s workflow is `read_specs → analyze_strategy → write_strategy_spec →
create_next_item → complete` with **no conditional anywhere**. `create_next_item`
unconditionally files `needs_briefing` → `build-briefing-agent`, which files `needs_site_plan`
→ `build-site-planner`. **Verified chain, three links: strategy → briefing → re-plan.**

All 3 `needs_briefing` rows in history were produced by `domain-strategist` on greenfield
builds, where the chain is correct. **It has never been run against a deployed site**, and on
one it would re-plan a live site as a side effect of refreshing its premise.
[UNVERIFIED beyond the planner: whether the re-plan then causes pages to rebuild —
`build-site-planner` files no further work items, so any onward effect is via its own writes.]

### 5.6 The checker layer is ready
`quality-discovery-agent` carries 6 checks (`broken_nav_links`, `placeholder_contact`,
`generic_theme`, `unverified_claims`, `voice_tells`, `literal_markdown`) — the array §B3's
two checks belong in. Recipe is settled: registry `init()` + sqlmock test +
handler/verifier/registration coverage tests; name into the array **only after the image
rolls**; IMP-016 order (observe-only → handler live → one clean cycle → enable). Since
bugfix 149 B4 an unregistered check name is **fatal**, not skipped.

## 6. What makes this non-obvious

- **Detection without a drain is the platform's proven failure mode**, and this feature is a
  pure detector by nature. `bugs_open/115` is the worked case: a correct brief-fidelity audit
  predicted the owner's own complaints three days early and died at `status='detected'`.
  **Every detector ships with its handler and promotion, or it does not ship** — the lane's
  organising rule, and it applies hardest here.
- **The analyser will be graded on an artefact it cannot see the effect of** (§4).
- **It judges the premise, which may itself be wrong.** A site whose recorded strategy is
  generic will pass a fit test against a generic premise. Grading the *premise's* quality is a
  second, harder instrument — deliberately out of scope, named so it is not forgotten.
- **Its findings land on other lanes' agents.** Two of the seven correspondents have no route
  at all; three route into machinery with known contracts that a bare work item does not
  satisfy (§3.1).

## 7. Promoted ahead of the rest — owner decision, 2026-08-08

The owner has ruled that **B1 and B2 jump the queue**, reversing the 2026-08-02 build-order
decision *for these two items only* (the designer track otherwise keeps its priority).

The dependency argument: **neither depends on anything in Programme A.** B1 improves a review
that already runs; B2 removes a live hazard. Both are config-only.

- **B1 — give the strategic review its own premise.** Widen
  `site-review-agent.load_strategic_context` (a `query_database` step, so this is one SQL
  rewrite) to load `strategy`, `identity`, `content_direction`, `mission_brief` and
  `audience`; extend the prompt to judge the site against its own recorded
  `revenue_models.primary_model`. Verify by a **planted marker** reaching the assembled
  prompt — not by the query returning rows.
- **B2 — make a premise refresh safe.** Gate `create_next_item` so a refresh on a
  **deployed** site completes *without* filing `needs_briefing`; add the four §5.4 Q-fields
  to the strategy shape with a refresh-preserves instruction. Verify on
  loancalculator.co.uk: a new strategy row, and **zero** `needs_briefing` items.

Both are `agent_definitions` config, so they are **live the moment the migration applies** —
no image roll, and no gap in which they are committed-but-inert.

## 8. Open questions for the owner

1. **Which council, or none** (§3.2)? The mission seat already exists for code; a site-facing
   offer seat is new; "auditor, not reviewer" is a legitimate third answer.
2. **Which of the two missing routes matters** — tool design, or the experience loops (§3.1)?
   Each is a design job in its own right.
3. **Does the analyser get its own handler**, or keep routing to the existing three
   (`domain-strategist`, `component-template-fixer`, `content-gap-planner`)? Reuse is the
   platform's stated preference; a dedicated handler is justified only once real offer
   findings are seen not to fit.
4. **Does grading the premise's own quality come into scope** (§6), or stay out?
5. **Enrolment order** once built. PLAN §B5 proposes webdesign.co.uk as the end-to-end proof
   and defers the rest to an owner call at the time.

## 9. Relates to

`features_open/018` (the taste critic — the designer's counterpart to this; same lane, same
"specified not built" status) · `bugs_open/115` (detection wired to nothing — the failure this
must not repeat) · `bugs_open/198` (the lane's live blocker) · doc 028
`platform_mission_and_pipeline_direction` (the revenue-shape doctrine both this feature and
the `review_mission` seat enforce) · `portfolio_positioning` (owns premise→writer wiring for
new builds; this feature owns the read-back side of the same spec fields — **do not collide**)
· owner ruling 2026-08-02 / RFC_010 §1 (the second-producer rule that governs §5.2).
