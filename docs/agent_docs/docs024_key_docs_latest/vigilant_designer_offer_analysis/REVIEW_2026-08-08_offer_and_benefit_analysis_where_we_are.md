# REVIEW 2026-08-08 — the offer and benefit analyser: what we've got, where we are

Written at the owner's request ("research the docs and threads to determine how far the
discussion has got with preparing a dedicated offer and benefit analysis step… prepare a
review of what we've got and where we are").

Every figure below is measured against the live system on 2026-08-08 unless marked
otherwise. Where a figure is inherited from another lane's measurement it says so.

---

## 1. The short answer

**The discussion has got exactly as far as one approved plan and no code.** The offer and
benefit analyser is **Programme B** of this lane, approved by the owner on 2026-08-02, and
it is **deliberately queued behind Programme A** (the vigilant designer) by the owner's own
build-order decision that day. Programme A is roughly 40% delivered and has been **idle for
three days**. Nothing of Programme B exists.

There is also a **scope gap** worth a decision now: the owner's current framing (a full
agent + checker + handler, a council seat, corresponding with copywriting, design,
planning, imagery, tool design, experience loops and spec at several points) is materially
larger than what PLAN §B4 describes, which is a **single config-only agent that reads and
files findings against existing item types and existing handlers, corresponding with
nobody**. That is not a defect in the plan — B4 was scoped small on purpose so it could
ship — but the two descriptions should be reconciled before anyone builds to the smaller
one. §6 sets out the choice.

---

## 2. Where the work actually lives

`docs/agent_docs/docs024_key_docs_latest/vigilant_designer_offer_analysis/` — the lane owns
both tracks. `PLAN_2026-08-02_…md` §Programme B is the only place the offer analyser is
specified. Cold-start for the lane is `HANDOFF_2026-08-03_continue_here.md` + the NOTES tail.

The originating owner words are recorded at the head of the PLAN:

> "a constant vigilant designer… and the offer analyser and benefit analyser… all of the
> above can also be made into checkers and handlers too for continual improvement. It needs
> a strong focus or we'll keep tripping up on the detail."

The four owner decisions of 2026-08-02 that still govern:

| decision | answer | consequence for the offer track |
|---|---|---|
| Cadence | manual per-site triggers; `improvement-sweep` stays `enabled=false` | the offer analyser will be hand-fired, not scheduled |
| Build order | **designer first, offer analyser second** | **this is what is holding B up, and it is the owner's own call — it can be reversed** |
| Critic model | trial Gemini and Claude | applies to A2; the offer analyser is text-only and unaffected |
| Authority | broad autonomy, locks respected | offer findings would auto-apply through existing handlers |

**Notably absent: there is no `features_open/` entry for the offer analyser.** The design
critic has one (`features_open/018_FEATURE_specialist_design_critic.md`, raised 2026-07-24
from the owner's own words, still `Status: specified, not built`). The offer analyser has
none — it exists only inside PLAN §B. If the intent is now larger than PLAN §B, a
`features_open/` file is the right place for it, because that directory is where "designed
but not built" is supposed to live and be found by other lanes.

---

## 3. What is built (Programme B): nothing. Measured four ways.

| what the plan asks for | live state 2026-08-08 | how measured |
|---|---|---|
| B4 `offer-analyser` agent | **does not exist** | `agent_definitions` wildcard on `offer\|benefit\|critique\|recompose` — 0 rows |
| B3 `check_premise_incomplete`, `check_revenue_shape` | **do not exist** | repo-wide grep across `*.go`/`*.sql`: hits only in this lane's own prose |
| any offer LLM traffic, ever | **none** | `llm_call_log` 14d: `site-review-agent` 16 calls, `domain-strategist` 5, nothing else in the family |
| B1 widen `site-review-agent.load_strategic_context` | **not done** | live config still selects only `domain, company, dream_spec, site_plan, deployed_pages, completed_items` — no strategy, identity, content_direction, mission_brief or audience |
| B2 gate `domain-strategist.create_next_item` | **not done** | live config still unconditionally creates `needs_briefing` → `build-briefing-agent`, i.e. **every strategy refresh still enqueues a rebuild chain** |

The `brochure_component_library` lane reached the same conclusion independently on 2026-08-05
after running a real improvement sweep, and wrote the warning into their handoff in these
words: *"Still does NOT exist: the offer and benefit analyser… Do not tell the owner the
offer analyser ran."* Two lanes, two methods, same answer.

---

## 4. What is built that the offer analyser would sit on top of

This is the "what we've got" half, and it is more than nothing.

### 4.1 The nearest live equivalent — and it is judging blind

`site-review-agent.run_strategic_review` is live, `active`, and **runs inside every
improvement sweep** (`improvement-loop → call_site_review`). Its prompt asks six strategic
questions, including *"What single change would most improve conversion?"* and *"Does the
page structure serve the business goal or is it generic?"*, caps itself at five findings,
and emits a closed vocabulary (`content_rewrite | needs_content_page | tone_shift |
cta_improvement | nav_restructure`) with an `acceptance_test` on each.

**But its entire context is domain, company name, `content_data.dream_spec`, the site plan
and two counts.** It does not load the strategy, the audience, the identity, the content
direction or the mission brief. So the platform already asks the offer question sixteen
times a fortnight — and asks it without the premise in front of it. That is precisely what
B1 was written to fix, and B1 is one query rewrite.

### 4.2 The premise substrate exists and is rich

`site_specs` current aspects, live counts: `audience` 29, `identity` 20,
`content_direction` 20, `strategy` 17, `mission_brief` 8, `evidence_base` 10,
`vertical_landscape` 9, `cta` 4, `portfolio` 4, `commercial` 2. The raw material for a
need↔offer judgement is already written at build time for most of the estate.

The `portfolio_positioning` lane measured (2026-07-31, their figure, not re-run here) that
the `audience` aspect — the most populated of the lot at 29 sites — **is read by nothing**.
That is the same disease as §4.1: written once, never consulted.

### 4.3 Two real problems in the strategy shape that the plan does not yet account for

Enumerating the keys of every current `strategy` spec (not path-reading them — a jsonb path
read cannot see a shape change underneath it):

- **16 of 17 sites carry one shape**: `value_proposition, revenue_models, monetisation*,
  competitive_position, growth_path, search_intent, site_type, domain_type, tone,
  content_strategy, recommended_page_types` + the two `_reasoning` fields.
- **1 site carries a completely different shape** — `gaswholesalers.com`, written 2026-04-17:
  `satisfaction_condition, trust_threshold, recurring_value, monetisation, primary_intent,
  visitor_type`. **Those are, almost exactly, the four "Q-fields" PLAN §B2 proposes to add.**
  So B2 is a *restoration of an abandoned shape*, not an invention — which is a much easier
  argument to make, and it means there is one live worked example to copy.
- ~~**`primary_model` does not exist in any strategy row (0/17).** PLAN §B1 says the prompt
  should "judge against `primary_model`" and the §B3 verifier is "strategy row with
  `primary_model`". **Both reference a field that has never been written.** This is a plan
  defect, caught here rather than at build time.~~

  > **CORRECTED 2026-08-08, same session, before anything was built on it — the claim above
  > is FALSE and the PLAN was right.** `primary_model` exists on **16 of 17** sites, nested
  > at **`revenue_models.primary_model`**. I read `data->>'primary_model'` — the *top level*
  > — got 17 empties, and called the field absent.
  >
  > **What caught it:** reading `domain-strategist`'s own prompt while writing up B2. Its
  > output schema puts `primary_model` inside the `revenue_models` object, in plain sight.
  > **The cheap check:** when a key is missing everywhere, read the WRITER's schema before
  > concluding it is absent — a field that no row has is far more likely to be a field you
  > are looking for in the wrong place. I had even enumerated the top-level keys and treated
  > that as "enumerating the shape"; enumerating one level is still a path read.
  > Logged in `WRONG_CALLS.md`.
  >
  > **The distribution is the real finding, and it is better than the error was:**
  > `direct_business` 10, `saas_tools` 3, `display_advertising` 2, `lead_generation` 1,
  > `sponsored_listings` 1, absent 1 (gaswholesalers, the old shape). So **10 of 17 live
  > sites are recorded as the consultancy shape** — the one doc 028 names as *"a failure
  > mode, not a safe fallback"* when the signal is absent.
  >
  > **What that number does and does not establish.** It does NOT show 10 misclassified
  > sites: several genuinely are businesses (finetuning.uk, leopardessconsulting.co.uk,
  > webdesign.uk, oufe.com). It DOES establish that `check_revenue_shape` (§B3) has a real
  > population to run against and a testable question on day one — *does each site's CTA
  > lexicon match its own recorded shape?* — and that the disconfirming answer ("all 17
  > agree") was available and is not what came back. The candidates worth looking at first
  > are the `direct_business` rows on domains that read as topic or tool, not brand.

### 4.4 `needs_strategy` is already a live type with a live producer — B3 would be the second

Three `needs_strategy` rows exist, all `complete`: `lendzy.co.uk` and
`mortgagecalculator.co.uk` (2026-08-02), `webdesign.uk` (2026-08-04). Producer:
`vertical-exemplar-researcher`. Handler: `domain-strategist` — exactly the handler PLAN §B3
names.

This matters for governance, not just for plumbing. Adding `check_premise_incomplete` makes
this lane a **second producer converging on an existing `item_type`**, which is squarely the
**owner ruling of 2026-08-02 (RFC_010 §1)**: that needs no architecture round *provided* the
concept-register entry names the full producer set and states the shared `item_key` shape.
So B3 is cheaper than the plan assumed — but it carries a mandatory register obligation.

### 4.5 The offer conscience already sits on the council — for code, not for sites

The council gate runs **17 seats**, and one of them is `review_mission`, **always-on**. Its
prompt encodes doc 028's revenue-shape doctrine verbatim, down to the example:

> "THE REVENUE MODEL SHAPES THE SITE, not the other way round… Mixing shapes (a tools site
> with a 'Start a Project' CTA) signals the classification is vague or a downstream agent is
> ignoring it."

That is the offer/benefit judgement, already seated, already firing on every submission —
**but it judges platform code changes, not site artefacts.** A "tools site with a
consultancy CTA" will never be caught by this seat, because the seat never looks at a site.
This is the single most important distinction in this review, and §6 turns it into a choice.

There is also a **second, site-facing council already built and exercised**:
`experience-approval-council` — 5 seats (`checkability`, `deferral_honesty`, `honesty`,
`observable_outcome`, `prior_art`), 36 LLM calls all-history, last run 2026-07-29. If offer
judgement is to sit on a council at all, that is the shape and the precedent to copy, not
the code gate.

### 4.6 The checker layer is ready to receive B3's two checks

Three discovery agents carry check arrays: `design-discovery-agent` 23 checks,
`completeness-discovery-agent` 32, `quality-discovery-agent` 6 (`broken_nav_links`,
`placeholder_contact`, `generic_theme`, `unverified_claims`, `voice_tells`,
`literal_markdown`). B3's two checks belong in the last of those. The recipe is settled
(registry `init()` + sqlmock test + handler/verifier/registration coverage tests, name into
the array only **after** the image rolls, IMP-016 order: observe-only → handler live → one
clean cycle → enable). Since bugfix 149 B4, an unregistered check name is **fatal**, not
skipped — so the ordering is not optional.

---

## 5. Why Programme B has not started: the gate is Programme A, and A is stalled

| phase | state | evidence |
|---|---|---|
| A0 — make findings flow | **DONE and proven live** | migrations 290/291/301 applied; witnessed run on relojistas (orchestration `5d36d7ec`, 2026-08-04): gate parsed live, 22 findings promoted, drain proven both ways — one stale finding retracted by observation, one fresh finding worked through to a re-rendered header + a 19-page cascade, all 22 deployed. `bugs_open/171` closed on it |
| A1 — eyes | **DONE and live** | `capture_renders` in browser-runner v1.0.1241; `GenerateWithImages` / `execute_vision_prompt` in chassis ≥ v1.0.1244, both replicas pod-verified with negative controls |
| A2 — the design critic | **NOT STARTED** | no `design-critique-agent` in `agent_definitions`; the Gemini-vs-Claude trial has never run |
| A3 — recompose drain + 016/017 | not started | no `page-recompose-handler`; `needs_page_recompose` has 0 rows |
| A4 — anti-brochure compose-time | not started | — |
| **B1–B5 — offer analyser** | **NOT STARTED** | §3 |

**What stopped it.** On 2026-08-04 the first-ever dispatch to `css-patch-agent` produced
*correct* colour fixes and then **destroyed the stylesheet**: the model returned only the new
rule, and nothing between the model and the two writers checked size or shape, so
`css_themes` went 25,816 → 149 chars and relojistas.com served an unstyled site until the
owner landed the restore. That is `bugs_open/198`.

The defect is now fixed in config — migration 318 makes the save an **append** (shrink is
unrepresentable in SQL concatenation), adds a 1..8192-char guard and a `check_saved` step
that fails loud on zero rows; council **APPROVED r1**. `bugs_open/198` stays open because
its own bar is a witnessed end-to-end run, which has not been fired.

**The lane's last commit is 2026-08-05.** Three days idle. The next action on the plan's own
terms is: fire one css-patch dispatch to discharge 198, then seed A2.

---

## 6. The decision this review exists to surface

PLAN §B4 specifies the offer analyser as **config-only**: `load_premise →
load_offer_surface (all active pages) → run_offer_analysis → write_findings`, emitting
**existing item types only**, routed to **existing handlers**, with a prompt that forbids
"users want…" phrasing because the platform has **zero outcome data** — it can only grade
the artefact against the stated premise, never against real visitor behaviour.

The owner's current framing is bigger on four axes. Each is a genuine choice, not a
misunderstanding:

1. **Checker + handler of its own.** PLAN §B3 gives the offer track two checks
   (`check_premise_incomplete`, `check_revenue_shape`) and reuses `domain-strategist`,
   `component-template-fixer` and `content-gap-planner` as handlers. A dedicated *offer*
   handler does not exist and is not planned. Reuse is the platform's stated preference and
   is cheaper; a dedicated handler is warranted only if offer findings turn out not to fit
   the existing three.

2. **A council seat.** Two different councils, two different answers (§4.5). If the intent is
   "no platform change should quietly break the revenue-shape doctrine", **that seat already
   exists and is always-on** — nothing to build. If the intent is "no *site* should ship with
   an offer that does not answer its market", that is a **new site-facing seat**, and
   `experience-approval-council` is the precedent to copy. These are not substitutes.

3. **Correspondence with the other agents.** PLAN §B4 deliberately has none — it files
   findings and existing handlers pick them up. The live seams the owner named all exist and
   are all reachable:

   | owner's term | live counterpart | how an offer finding would reach it today |
   |---|---|---|
   | copywriting | `content-quality-auditor`, `content-reviewer`, `page-content-writer` | `content_rewrite` / `tone_shift` items |
   | design | `design-audit-agent` → `visual-design-auditor`, `webdesign-agent` | `needs_design_review`; the A2 critic when it exists |
   | planning | `build-site-planner`, `site-design-planner` | `needs_page_recompose` — **handler unbuilt (A3)** |
   | imagery | `image-build-handler`, imagery-plan checks | `needs_imagery` — needs an imageryplan spec row, not a bare item |
   | tool designer | `tool-auditor`, `missing_tools` / `tool_acceptance` checks | no offer→tool route exists |
   | experience loops | `experience-planner`, `experience-approval-council`, `experience-register-writer` | no route exists |
   | spec | `site_specs`, `domain-strategist`, `chief-strategist`, `site-strategist` | `needs_strategy` — **live already (§4.4)** |

   So of the seven, two are wired today (spec, copywriting), three are wired-but-fragile
   (design, imagery, planning), and **two have no route at all** (tool design, experience
   loops). "Corresponds at several points" is a real increment over PLAN §B4, and the two
   missing routes are where the design work is.

4. **Ordering.** B sits behind A2/A3/A4 by the owner's 2026-08-02 build-order decision. That
   decision is reversible and the dependency is only partial: **B1 and B2 depend on nothing
   in Programme A.** B1 is a single query rewrite that would immediately improve a review
   that already runs sixteen times a fortnight. B2 removes a live hazard (every strategy
   refresh currently enqueues a rebuild chain) and is a precondition for B3 whenever it comes.
   **If the owner wants offer movement without disturbing the designer track, B1+B2 are the
   two pieces that can be done now.**

---

## 7. Recommended next actions (for the owner to choose between, not to execute silently)

- **Reconcile the scope.** Write the larger offer/benefit analyser as a `features_open/`
  entry in the owner's own framing, the way `018` was written for the design critic. It is
  the missing artefact and it is where other lanes will look.
- ~~**Fix the plan's two factual defects before anyone builds to it**: `primary_model` does
  not exist (§4.3), and `needs_strategy` already has a producer (§4.4).~~
  **REVISED 2026-08-08 — there is ONE defect, not two.** The `primary_model` half was my own
  error (§4.3 correction); the PLAN's text is correct and must NOT be "fixed". What stands:
  the four Q-fields are a **restoration** of an abandoned shape with one live worked example,
  not an invention (§4.3), and `needs_strategy` already has a producer, which carries a
  register obligation under the 2026-08-02 ruling rather than an RFC (§4.4).
- **Consider unblocking B1 + B2 now** — independent of Programme A, cheap, and each fixes a
  live problem rather than adding a new mechanism.
- **Decide which council** the offer judgement belongs to (§6.2) — the answer may be
  "neither, it is an auditor not a reviewer", and that is a legitimate outcome.
- **Programme A's own next step is unchanged**: discharge `bugs_open/198` with one witnessed
  css-patch run, then seed A2.

---

## 8. What this review did not do

- Did not run the `090` diagnosis loop. No cross-cutting root-cause claim is made here — every
  statement is a direct read of live config, live rows, or a named document, and §3's negative
  claims are each measured two ways where a single grep could have lied.
- Did not fire anything at the cluster, seed anything, or change any plan file.
- Did not re-verify the `portfolio_positioning` lane's "audience is read by nothing" figure
  (§4.2) — it is attributed to them and dated 2026-07-31.
- Did not survey whether the *content* of the 16 existing `strategy` rows is any good, only
  their shape. A premise that exists but is generic is a different finding, and the offer
  analyser is the instrument that would produce it.
