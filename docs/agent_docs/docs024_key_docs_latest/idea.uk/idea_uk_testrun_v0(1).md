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

**Candidate after the critique pass (rescored with the v1 rubric):**

```
1a. Spec/plan as a starter prompt for AI builders (the "bare prompt" version)
   Idea: hand the customer our site-spec/plan as a ready-to-paste prompt for
         Bolt/Lovable/v0.
   Critique (the real free substitute): the customer can describe their idea to
         Bolt themselves for free, and Bolt/Lovable already run their own planning
         step. Paying £2–5 for "a better prompt" is a hard sell against that.
   Verification: prompt precision helps (confirmed), but builders plan internally,
         accept PRDs, and spec-driven competitors exist.
   Scores: Defensibility 3 · Willingness 2 · Buildability 5 · Reuse 4 · Durability 2
   Result: FAILS the gate (Willingness < 3) and Durability ≤ 2 (builders' own
         planning is improving). Not a paid product on its own.

1b. Build-orchestration / improvement layer (the version that survives)
   Idea: produce the spec, run it through the builder, critique the output,
         iterate — sell the improvement loop, not the prompt.
   Beats the free substitute because: the value is the iteration/quality lift the
         user can't easily do themselves, not a one-off prompt.
   Scores: Defensibility 3 · Willingness 4 · Buildability 2 · Reuse 3 · Durability 3
   Result: advances, but expensive (Buildability 2) → [consider], not test-now.
   Cheapest test: a fake-door page offering "we'll spec it, build it, and improve
         the result for £X" and measure intent before building the loop.
```

**Dropped:** brand pack, builder recommender, generic mind-map tool (the free
substitute covers them).

**Read (revised after the critique):** the first pass scored the bare-prompt idea
15 and flagged it test-now. The critique pass — testing against the *real* free
substitute (describe it to Bolt yourself) and adding durability — **failed it**:
willingness-to-pay for a bare prompt is too low and the advantage erodes as
builders improve. The version that survives is a bigger build (an orchestration
layer), which is expensive and therefore "consider", not "test now". So the
critique materially changed the conclusion — the cheap thing we'd have rushed to
test probably wasn't worth testing, and the thing worth pursuing needs a fake-door
demand test first, not a build.

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

## Verdict on the method (after the critique pass)

- The **cut step worked** — most candidates died there, as intended.
- The **verification step earned its keep** — it downgraded the "strongest" idea
  from an assumed win to a contestable one, with a concrete reason.
- **The critique pass changed the conclusion** — this is the important result.
  Re-examining the websitedesign candidate against the *real* free substitute
  (describe it to Bolt yourself) and adding durability **failed the bare-prompt
  version that the first pass had flagged "test now"**. A single generative pass
  would have sent us to build/test a weak product. So the multi-pass / multi-model
  step is not decoration — it caught a strategic flaw. (Caveat: the critique here
  was a rigorous opposing pass by the *same* model; the real tool should use a
  genuinely different model, which would likely catch more.)
- **Rubric improved to v1:** added a **Durability** factor and sharpened the cut
  step to name the actual free substitute (both now in `idea_uk_method_v0.md`).

**Still to fix:** willingness-to-pay is still estimated, not evidenced (the
fake-door tests are how we fix that); and the critique should be run with a truly
different model once the tool exists, not a same-model opposing pass.

**Next options:**
- Build the **fake-door demand test** — but note it should test the *orchestration
  layer* (1b: "we'll spec, build, and improve it for £X"), not the bare prompt
  (1a) the critique just failed.
- Run the method against a third domain with a different asset type (e.g.
  gaswholesalers' "buy a proprietary feed" — verify whether such feeds exist and
  their cost), to test the method on a data-asset candidate rather than a
  process-asset one.
