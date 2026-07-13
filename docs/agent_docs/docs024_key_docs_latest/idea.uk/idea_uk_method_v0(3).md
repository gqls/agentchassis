# idea.uk — Ideation Method v0 (runnable)

The §10 method made concrete enough to run the same way every time — by hand now,
by the §11 agent later. Companion to `PLAN_idea_uk.md`. This is **v0**; expect to
revise the rubric after the first real runs.

The principle it encodes: a payable idea is **one hard-to-reproduce asset × one
current AI capability, aimed at an audience that will pay**, doing something a free
model with a good prompt cannot.

---

## Inputs

1. **Audience** — who the chat serves, and how much they can and will pay (a
   sentence each). For the hosted version this comes from the user; for internal
   runs we supply it.
2. **Asset list** — the hard-to-reproduce things this business has or could
   acquire. Categories: proprietary/paid data; an owned process or output; a tool
   we could build well; a commercial partnership; early access/timing on a new
   capability. *Passed in as data — never baked into the method (so the same
   method serves a stranger's business on idea.uk).*
3. **Capability list** — the maintained watchlist of AI capabilities worth using
   now (below). The watchlist's own upkeep is a separate recurring workflow.

### Capability list (v1 — grouped by what specialism can do uniquely)

The v0 list was generic ("text reasoning, vision, voice..."); too coarse to provoke
novel combinations. v1 groups capabilities by **what specialism does that
generalist LLMs don't**, with concrete examples of what each enables.

- **Knowledge & retrieval.** Curated/grounded RAG over the right sources;
  long-context (loading entire user corpora or regulatory bodies at once);
  persistent cross-session memory. Beats generalists on stale or shallow knowledge.
- **Reasoning & computation.** Domain-tuned reasoning models; *actual* computation
  (not LLM approximation — call a solver/calculator/simulator); multi-step
  workflows with checkpoints and verification. Beats generalists on confident-wrong
  outputs in technical/regulated areas.
- **Multi-modal input.** Technical image understanding (engineering drawings,
  medical images, agricultural conditions, forms); voice/audio; sensor/data feeds.
  Beats generalists on domain-specific inputs they weren't tuned for.
- **Multi-modal output.** Precise image editing/in-painting; video generation;
  voice; structured data with schema guarantees. Beats generalists on output
  fidelity for specific use cases.
- **Action-taking.** Agentic browsing / computer use; integration with the user's
  stack (calendar, CRM, accounting, devices); multi-step workflow execution.
  Beats *description-only* generalists — does the thing rather than telling you how.
- **Coordination.** Multi-model ensembles (cheap-fast for some steps, deep-slow
  for others); specialised sub-models per task. Beats single-model chats on
  quality-at-price.
- **Personalisation & continuity.** Cross-session memory; user-profile tuning;
  ongoing-project context. Beats stateless chats for coaching, learning,
  case-management, long projects.
- **Quality & safety.** Domain-aware validation; source-grounded outputs with
  citations; explicit uncertainty signalling; refusal/escalation rules. Beats
  generalists in regulated/sensitive domains where confident-wrong is harmful.

*Add new entries as models ship them; each new entry triggers a re-run across
domains — this is the early-adopter step.* Currently watching: agentic
browsing/computer-use reliability, million-token contexts, reasoning-model
pricing, real-time voice agents, video generation usefulness, precise image
editing.

---

## Procedure

Run the steps in order. Steps marked **[diff-model]** are better done with a
*different* model than the generator, to widen the idea set and avoid a model
rubber-stamping itself. Steps marked **[web]** use web research.

1. **Frame the audience.** State the audience and their willingness to pay in two
   sentences. If willingness to pay is clearly low, note it now — it caps
   everything later. *(v2 lesson from the robot-hands run: also ask "is this the
   right audience, or would a different audience for the same domain pay more or
   be on a softer free substitute?" — challenge the audience now, not after.)*

2. **Generate candidates — multi-lens (v2).** The v0 single asset×capability
   pass is supply-side only and surfaces only obvious moves. Run all four lenses
   below; aim for 3–6 candidates *per lens* (12–24 total before dedup). Merge
   and dedupe before the cut. **[diff-model]** optionally regenerate one or more
   lenses with a different model.

   - **a. Demand lens.** What does this audience deeply want, struggle with, or
     pay for today? What workflow takes them longest? What error-prone task do
     they re-do? What expertise is unevenly distributed (insiders have, outsiders
     lack)? What do they currently pay specialists for that's mostly pattern-
     following? What can they not get done because they lack a piece of expertise?
   - **b. Generalist-failure lens.** Where does a generalist LLM fail this
     audience? Stale or wrong on technical/regulated specifics? Confident-wrong
     on details that matter? Generic when domain-specific would be better? Can't
     take action, only describe? Forgets between sessions when continuity is
     needed? Can't compute precisely when precision matters? Can't access live
     or proprietary data? Each "yes" is a candidate seed.
   - **c. Frontier lens.** What capability just became possible or cheap in the
     last 6–12 months that could enable a new product for this audience?
     (Pull from the watchlist.) For each, ask "is there a thing for *this*
     audience that wasn't possible 18 months ago?"
   - **d. Outcome lens.** What's the *dream outcome* for this user — not "help
     me with X" but "X is done, correctly, ready to use"? Reverse-engineer to a
     product that delivers that outcome.
   - **e. Asset × capability sweep (v0 step, kept).** Cross the asset list with
     the v1 capability menu as a final pass to catch obvious combinations the
     lenses missed.

3. **Cut against the specific free substitute.** **[diff-model]** For each
   candidate, name the *specific* free alternative the audience would actually
   use — the concrete thing they'd otherwise do (describe their idea to Bolt;
   call the supplier's free application engineer; use the manufacturer's free
   training app; ask Perplexity). If that gets them most of the way, drop it. Be
   ruthless — most candidates should die here.

   *Also (v2 from the robot-hands runs):* check whether the seller of the
   underlying product *already gives this support away free* as part of their
   sales process — for high-margin products, they usually do. If so, drop unless
   the candidate clearly improves on the seller's offer.

4. **Verify the survivors.** **[web]** Check the claims the idea rests on: does
   the data feed/partnership actually exist and what does it cost; do competitors
   already offer it; is the willingness-to-pay real (evidence, not assertion).
   Attach what you find. Drop candidates whose premise fails.
5. **Score** each survivor on the rubric below.
6. **Rank and split.** Order by composite; split into **test now (cheap)** and
   **score/consider (expensive)**. For each, name the cheapest demand test.

---

## Scoring rubric

Four factors, 1–5, where **5 is always more attractive**. Define each honestly;
guessing high defeats the point.

**Defensibility — how hard to reproduce**
- 1: a free model with a good prompt does this now.
- 3: needs our process/output or assembled curation; a determined expert could
  copy it with effort.
- 5: depends on an asset others can't get (exclusive/paid data, a held
  partnership, or a genuinely hard-to-build tool).

**Willingness to pay**
- 1: users expect this free, or won't pay.
- 3: some would pay a small amount; mild pain.
- 5: audience has budget and clear, repeated pain; pays readily.

**Buildability — cheap/fast to deliver**
- 1: major bespoke build, new infrastructure.
- 3: moderate build, mostly assembling things we have.
- 5: trivial, or we already produce it.

**Reuse across domains**
- 1: bespoke to this one domain.
- 3: reusable across a few similar domains.
- 5: reusable across many/most domains.

**Durability — resistance to base-model improvement** *(added in v1 after the test
run; the whole moat rests on currency, so this must be scored)*
- 1: the next base-model release likely erases the advantage; the substitute is
  improving fast (e.g. builders adding their own planning).
- 3: holds for a while; needs periodic refresh to stay ahead.
- 5: rests on something base-model progress doesn't erode (exclusive data, a held
  partnership, the build bridge).

**Risk to the operator** *(added in v3 after the agritec/SFI run — the rubric had
scored SFI single-farm assessment as test-now, with no flag that wrong output
could cost a farmer £5k–50k of lost grant money. The dimension wasn't there to
catch.)* Score the CONSEQUENCE of being wrong, not the probability of being
wrong. Higher = safer.
- 1: regulated profession territory (medical advice, legal advice, FCA-regulated
  financial advice). Should NOT be built without proper qualifications regardless
  of how attractive it looks on the other factors.
- 2: high-stakes decisions ride on the output (real money, regulated or
  quasi-regulated matters, legal/medical adjacency). Needs human review of every
  report + PII insurance + carefully reviewed T&Cs before any build.
- 3: meaningful financial or operational decisions, but the customer can verify
  our citations and reverse course before harm. PII insurance recommended.
- 4: minor downstream consequences possible; refunds make the customer whole.
- 5: pure analysis; customer makes their own decisions; no plausible loss beyond
  the fee paid.

**Risk is NOT added to the sum.** Sum is fitness (Def + Will + Build + Reuse +
Dur, out of 25); Risk is hazard. They're shown separately in the report.

**Gating rule:** a candidate only advances if **Defensibility ≥ 3 AND Willingness
≥ 3**. Below that, it fails regardless (a cheap, reusable idea nobody will pay for,
or that the real free substitute already covers, is not a product). **Also flag
any candidate with Durability ≤ 2** as short-lived even if it advances — fine for a
quick cheap test, not for a sustained build.

**Risk rules** *(separate from the gate)*:
- **Risk = 1 is dropped** automatically and listed in a separate "Dropped for
  operator risk" section so the operator sees what got killed for risk vs what
  failed the Def/Will gate.
- **Risk ≤ 2 still advances** if it passes the gate, but flagged with
  "**⚠ needs liability work before building**" — the cheapest_test for these
  candidates must explicitly say "validate demand first; do not build until PII
  insurance is in force and T&Cs are reviewed by a UK solicitor."

**Rank** advancing candidates by the sum of the five fitness factors, with **Risk
as tiebreaker** (prefer safer builds when fitness is equal).

**Test-now flag:** mark "test now" if Buildability ≥ 4 **or** a demand test (e.g. a
"buy" button that records intent before building) is cheap **AND** Risk ≥ 3.
Only the expensive ones need deliberation; cheap ones just get tested — unless
they're risky, in which case the demand test comes before any build regardless.

---

## Output template

```
Domain: <domain>
Audience: <one line> | Willingness to pay: <one line>

ADVANCING CANDIDATES (Defensibility ≥3 and Willingness ≥3), ranked:

1. <title>  [test_now | consider]  [⚠ needs liability work before building, if Risk ≤ 2]
   Idea: use <capability> on <asset> to <do X>.
   Beats free AI because: <one line>
   Verification: <what was checked — feed exists/cost, competitors, WTP evidence>
   Scores: Defensibility _/5 · Willingness _/5 · Buildability _/5 · Reuse _/5 · Durability _/5  (sum _)
   Operator risk: _/5  <short label>
   Cheapest test: <the demand test — includes "PII + T&Cs review required" if Risk ≤ 2>

DROPPED (failed Def/Will gate): <title — reason>

DROPPED FOR OPERATOR RISK (Risk = 1, regulated territory):
  <title — flagged for visibility, not recommended>
```

---

## Notes on running it well

- The cut step (3) is the quality gate. If nothing dies there, the run was too
  soft and the output will be generic.
- Verification (4) is what separates this from a prompt — assert nothing you can
  check. Keep depth proportionate to cost.
- Multi-model: use one model to generate, a different one to critique and to
  verify-read, so the method isn't one model marking its own work.
- Dogfood: run the method on idea.uk itself; if it can't find an advancing
  candidate for its own domain, the method (or the product) isn't good enough yet.
