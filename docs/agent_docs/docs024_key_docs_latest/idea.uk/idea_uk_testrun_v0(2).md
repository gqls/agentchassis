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

## Run 3 — gaswholesalers.com (data-asset candidate test)

**Audience (as stated):** high-paid executives buying oil/gas in bulk, and oil/gas
traders. **Willingness to pay (individually):** high — but they are **already
over-served** by enterprise tooling (Bloomberg, Refinitiv, direct Platts/Argus
terminals at four-to-five-figure monthly costs). For this audience, "the free
substitute" is *their existing institutional terminal*, not a free model.

**Generated, then cut (named the specific free substitute, per the v1 method):**

- *Buy proprietary feed (Argus/Platts/ICIS), apply AI on top* — **held for
  verification** (the strong-defensibility version).
- *Cheap public APIs (EIA, OilPriceAPI) + AI commentary* — **cut**: free models
  with tool-use can call the same APIs directly. No defensibility.
- *AI news synthesis from oil/gas RSS* — **cut**: Perplexity-style tools and free
  chat with search already do this. No defensibility.
- *Specialised tool (basis differential calculator, scenario analyser, supplier
  risk dashboard) with AI commentary* — **held**: tools resist replication better
  than prompts. Verification on willingness needed.
- *Audience pivot: SME/small wholesale buyers without enterprise tools, using
  cheap data + AI* — **held**: different audience, possibly underserved.

**Verified:**

- *Buy proprietary feed* — The big-three PRAs are billion-dollar enterprises
  selling via enterprise channels (LSEG, Bloomberg-style); public pricing isn't
  listed, but the channel signals five-figure-plus annual fees. Licensing on
  these feeds typically restricts redistribution and LLM-grounding. **Premise
  fails:** economically infeasible at our scale and likely contractually
  forbidden as a resale base. Drop.
- *Specialised tool for high-paid traders* — the audience has institutional
  tools (Bloomberg+) that already do scenarios, basis, risk. A £2-pass retail
  tool isn't competing with that. Willingness for our price point essentially
  zero. Drop.
- *SME audience pivot* — cheap data feeds exist and are affordable ($0-$129/mo:
  EIA free, OilPriceAPI from $9/mo, others up to $129/mo with real-time WTI/Brent/
  Henry Hub/TTF/NBP). So building IS feasible. But the underlying data is widely
  available, and a free model with tool-use can hit the same APIs. Defensibility
  collapses to "is our tool/UX/curation meaningfully better than what an SME can
  get from a $9/mo API + ChatGPT?" — a thin claim.

**Scoring the only candidate that even reaches the scoring step:**

```
SME-focused cheap-data + AI commentary + simple tool
   Scores: Defensibility 2 · Willingness 2-3 · Buildability 3 · Reuse 2-3 · Durability 2
   Result: FAILS the gate (Defensibility < 3, Willingness probably < 3).
           Even if it scraped through, Durability 2 flags short-lived.
```

**Read — no candidate advances.** For gaswholesalers.com **as currently scoped**,
the method finds no payable differentiator without acquiring a new asset (an
exclusive partnership, a paid-feed redistribution deal at a price that won't work
at our scale, a domain expert producing curated insight, or a genuinely
differentiated tool with a moat of its own).

This is a useful negative result, not a method failure:

- The "obvious win" (buy a proprietary feed) **fails on verification**, not on
  taste — the licensing/cost reality is concrete.
- The cheap-data version **fails the same v1 cut** that caught websitedesign's
  bare-prompt version — naming the specific free substitute (here: a free model
  with tool-use against EIA/OilPriceAPI) kills it.
- The stated audience is **over-served**; the audience that fits the product is
  different and lower-willingness.

**Implication for the build list:** gaswholesalers.com is **not** in the first
batch of paid-chat domains. Either (a) pivot it to a different monetisation
(lead capture, free-with-caps for an underserved SME niche), or (b) shelve it
until we acquire a new asset for it (partnership/expert/data deal).

---

## Method behaviour across three runs (summary)

| Run | First-pass result | After critique/verify | Where the method earned its keep |
|---|---|---|---|
| websitedesign (Run 1) | "test-now", sum 15 | failed on willingness/durability (1a); 1b survives but expensive | the critique-against-the-real-substitute step |
| idea.uk dogfood (Run 2) | advances but expensive | unchanged; fake-door first | gate flagged "consider", not "test-now" |
| gaswholesalers (Run 3) | three plausible candidates | all fail; no candidate advances | verification killed the proprietary-feed option on cost/licensing; cut killed the public-data versions |

The method gives **three different verdicts on three runs** — test-now (rare),
consider-with-demand-test, and don't-build-yet. That range is healthy: a method
that always advances something is rubber-stamping.

## Run 4 — robot-hands.com (audience-fit stress test)

Ambiguous domain name. The method needs an audience as input, so this run
deliberately tests **three plausible audience framings** rather than guessing
one — also a direct stress test of the audience-fit gap flagged in Run 3.

### Framing 4A — Prosthetic / bionic hand users (medical / assistive)

**Audience:** people with upper-limb difference and their families.
**Willingness to pay individually:** complex. Devices cost £8,000–60,000+; the
established players (Open Bionics in Bristol, Ottobock, Touch Bionics, Aether
Biomedical) are the channel, and **over 70% of US orders are insurance-funded**.

**Generated, then cut against the real free substitute:**

- *Funding/insurance navigator* — **cut**: Open Bionics provides a Customer
  Success Officer giving *free* advice on funding pathways, insurance appeals,
  payment plans, grants. The seller of the £8k device gives the help away.
- *Device comparison/selection* — **cut**: free models with search compare
  Hero Arm vs TrueLimb vs Zeus; the audience is small and well-researched.
- *Training/personalisation help* — **cut**: Open Bionics ships the Sidekick
  App for exactly this.
- *Activity attachment design help* — **cut**: Open Bionics open-sources
  community designs for activity attachments on Printables. Free.
- *Peer/lived-experience chat* — **cut**: free Reddit/Facebook communities and
  the Lucky Fin Project / Open Bionics Foundation already cover this.

**Verification finding:** the high-margin product seller (Open Bionics, etc.)
**bundles support free** because the device sale is large enough to fund it.
There is no support gap to monetise. Also worth flagging: this is a vulnerable
audience and a healthcare context — charging £2 per pass for advice that's
already free is questionable on ethics, not just economics. **All candidates
fail.**

### Framing 4B — Industrial robotic grippers / end-effectors (B2B automation)

**Audience:** automation engineers, integrators, procurement at manufacturers
buying grippers ($500–10,000+) for KUKA/Fanuc/UR arms.

**Generated, then cut:**

- *Gripper selection AI* — **cut**: Robotiq, OnRobot, Schunk run free
  application-engineer pre-sales; their sales process *is* the consultation.
- *Compatibility/payload calculator* — **cut**: vendor sizing tools exist; free
  models with tool-use compute this; CAD/PLM workflows increasingly include AI.
- *Vision-based "what gripper for this part?"* — **considered**: more
  defensible because it uses image capability. But it competes with vendor
  application-sizing tools and B2B procurement does not go through a £2-pass
  retail chat.

**Verification finding:** same pattern as gaswholesalers and 4A — *the audience
is over-served by suppliers' free pre-sales engineering*, because the
underlying product (the gripper) carries the margin. B2B procurement also
doesn't fit a £2-pass model regardless. **All candidates fail.**

### Framing 4C — Hobbyist makers / DIY robotics

**Audience:** makers, students, hobbyists building robotic arms/hands with
Arduino/ESP32/ROS and 3D printers.

**Generated, then cut:**

- *AI tutor for build/code* — **cut**: free models walk through Arduino servo
  code, inverse kinematics, BOM generation; YouTube libraries are vast.
- *Personalised parts list / code debugger* — **cut**: same; ChatGPT and
  Claude do this well with no asset of ours involved.
- *3D-printable design generator from grasp requirements* — **cut**: free
  open-source designs (Open Bionics' community Printables) cover this.

**Verification finding:** maker culture is free/open-source by norm; the
willingness-to-pay floor is essentially zero for retail learning help. **All
candidates fail.**

### Cross-framing result

```
| Framing                        | Advances? | Why                                                    |
|--------------------------------|-----------|--------------------------------------------------------|
| 4A Prosthetic/bionic users     | No        | Manufacturer (Open Bionics) bundles support free       |
| 4B Industrial gripper buyers   | No        | Supplier pre-sales is the free substitute; B2B not £2  |
| 4C Hobbyist makers             | No        | Free AI + open-source; ~zero willingness to pay        |
```

**Read.** Second consecutive "no candidate advances" run, and the audience-fit
stress test paid off — running three framings showed that **no plausible
audience for robot-hands.com works as paid chat at our price point**, for three
different reasons. The method correctly refused to advance any of them.

A new pattern across this run AND gaswholesalers AND the bionics result:
**where the underlying product has high margin (a £30k prosthetic, a £10k
gripper, a £50k oil-data terminal), the seller already gives away expert
support for free** because that support is part of how they close the sale.
So a paid chat that "helps you with X" is structurally undercut by X's seller
in any high-value market. That's a general finding worth carrying forward.

**Implication:** robot-hands.com, like gaswholesalers.com, should come out of
the first paid-chat batch. Better fits: lead-capture / free-with-caps if there
is an operator monetisation route, or shelve the domain for paid chat.
