# idea.uk — Method v0 test run

First run of `idea_uk_method_v0.md`, by hand, against one real domain
(websitedesign.com) plus the dogfood run on idea.uk itself. Purpose: judge whether
the method produces honest, useful candidates — not to decide anything yet.

---

## Run 1 — websitedesign.com

**Audience:** people who want a website built — from non-technical small-business
owners to semi-technical founders already using AI app builders.
**Willingness to pay:** moderate — they already spend on builders/tools; a few
pounds to improve their result is plausible, but the value is hard to perceive
*before* they've seen it.

**Generated (step 2), then cut (step 3):**
- *Brand/style starter pack (palette, type, components)* — **cut**: a free model
  does this well. Low defensibility.
- *"Which builder should I use" recommender* — **cut**: a free model with search
  recommends this. Low defensibility.
- *Generic mind-map / site-structure tool* — **cut**: free tools exist; no AI edge
  on its own.
- *Package our site-spec-and-plan as a starter prompt/PRD for Bolt/Lovable/v0* —
  **survives**.
- *Pre-flight: generate the spec, then critique the idea's scope/feasibility before
  they burn builder credits* — **survives** (a variant of the same asset).

**Verified (step 4) — via web research:**
- Premise holds: builder output quality varies with prompt precision; detailed
  specs/PRDs improve results, and PRD/spec input is already a supported input
  pattern across these tools.
- **Competitor risk is real:** Lovable already runs an internal planning stage
  before building; spec-driven generation is already a product approach elsewhere
  (e.g. Remy derives code from a spec). So the idea is *not novel* and is partly
  commoditised.
- **Defensibility therefore hinges on one thing:** our pipeline's spec/plan being
  *demonstrably better* than (a) what the user can prompt a free model to write and
  (b) what the builder generates internally. Contestable, not assured.

**Advancing candidate:**

```
1. Spec/plan starter pack for AI builders
   Idea: use our build pipeline's site-spec-and-plan output as a ready-to-paste
         starter prompt/PRD the customer gives to Bolt/Lovable/v0 to get a better
         first result.
   Beats free AI because: it's our pipeline's structured output, not a generic
         prompt — IF that output is genuinely better than what the user/builder
         produces alone (the contestable part).
   Verification: prompt precision improves builder output (confirmed); but builders
         already plan internally and accept PRDs, and spec-driven competitors
         exist — so the edge is narrow and rests on spec quality.
   Scores: Defensibility 3/5 · Willingness 3/5 · Buildability 5/5 · Reuse 4/5  (sum 15)
   Cheapest test: a "generate a starter prompt for your idea" page with a pay/intent
         button; measure click-through and conversion before building anything.
   [test now]
```

**Dropped:** brand pack, builder recommender, generic mind-map tool (all "free AI
does it").

**Read:** the strongest idea on the board scores *middling on defensibility and
willingness, high on buildability and reuse*. That is an honest result: it's worth
testing **because it's cheap to build and reusable**, not because it's a strong
moat. The method let it through the gate (≥3/≥3) and correctly flagged it
test-now rather than overselling it.

---

## Run 2 — idea.uk (dogfood)

**Audience:** business owners/operators wanting AI opportunities for their
business. **Willingness to pay:** moderate-to-high (businesses pay for ideas and
prototypes) but sceptical of AI-generated suggestions.

**Generated, then cut:**
- *"AI suggests AI ideas for your business"* — **cut**: a free model does a version
  of this. No defensibility on its own.
- *Verified ideas + we prototype/build them, kept current with capabilities the
  base models don't yet know about* — **survives** (the §5 combination).

**Verified:** defensibility is process + freshness + integration (per `PLAN_idea_uk.md`
§5), not a static asset; the build-bridge (we can actually prototype the idea) is
the hardest part to copy and the part unique to us.

**Advancing candidate:**

```
1. Verified AI-opportunity report + prototype
   Idea: run this method (multi-model generate → critique → web-verify → score),
         kept current via the capability watchlist, and turn the top pick into a
         cheap test or working prototype using our build pipeline.
   Beats free AI because: candidates come verified, kept current with newly-shipped
         capabilities the base models don't reliably know about, and we can build
         the prototype — not just describe it.
   Verification: §5 analysis — durable parts are currency, verification, and the
         build bridge; all effort-sustained, not a static moat.
   Scores: Defensibility 3/5 · Willingness 4/5 · Buildability 2/5 · Reuse 3/5  (sum 12)
   Cheapest test: a landing page offering "a verified AI-opportunity report + a
         prototype for £X", measure intent BEFORE building the watchlist/agent.
   [consider — not test-now: buildability 2]
```

**Read:** idea.uk advances the gate but is expensive to build, so the method's own
output says: **fake-door demand test first** (landing page + intent) before
building the agent, watchlist, and ensemble. That matches `PLAN_idea_uk.md` §7's
sequence — the method validates the caution we'd already planned.

---

## Verdict on the method (v0)

- The **cut step worked** — most candidates died there, as intended.
- The **verification step earned its keep** — it downgraded the "strongest" idea
  from an assumed win to a contestable one, with a concrete reason. This is the
  thing a single prompt wouldn't have surfaced honestly.
- The **gating rule behaved** — it passed genuinely-payable candidates and flagged
  cheap-to-test vs expensive correctly.
- **Weaknesses to fix in v1:** scores are still one-person judgement (the
  multi-model critique wasn't actually run here — both runs were single-model);
  "willingness to pay" needs real evidence, not estimate; and the rubric doesn't
  yet capture *time-to-obsolescence* (how fast an improving base model erodes the
  idea), which matters given the currency theme.

**Next:** run the multi-model critique for real on the websitedesign candidate
(generate/critique with different models) to see if it changes the scores, and
build the websitedesign fake-door test since it's the cheapest, highest-reuse
advancing candidate.
