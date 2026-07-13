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

### Capability list (starter — the watchlist maintains this)

Text reasoning · web-search grounding · structured extraction from messy sources ·
document/long-context processing · vision (understanding images/PDFs) · image
generation & editing · code generation · tool/function calling · voice (TTS/STT).
*Add new entries as models ship them (e.g. video generation, agentic browsing);
each new entry triggers a re-run across domains — this is the early-adopter step.*

---

## Procedure

Run the steps in order. Steps marked **[diff-model]** are better done with a
*different* model than the generator, to widen the idea set and avoid a model
rubber-stamping itself. Steps marked **[web]** use web research.

1. **Frame the audience.** State the audience and their willingness to pay in two
   sentences. If willingness to pay is clearly low, note it now — it caps
   everything later.
2. **Generate candidates.** For each plausible (asset × capability) pair, write one
   candidate: *"For {audience}, use {capability} on {asset} to {do X}."* Plus a
   one-line reason it beats a free model with a good prompt. Aim for 6–12 raw
   candidates. **[diff-model]** optionally regenerate with a second model and merge
   — different models surface different candidates.
3. **Cut the generic ones.** **[diff-model]** Critique each: could a free model with
   a good prompt do this about as well? If yes, drop it (it has no defensibility).
   Be ruthless here — this is where most candidates should die.
4. **Verify the survivors.** **[web]** For each, check the claims the idea rests on:
   does the data feed/partnership actually exist and roughly what does it cost; do
   competitors already offer it; is the willingness-to-pay real (evidence, not
   assertion). Attach what you find. Drop candidates whose premise fails.
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

**Gating rule:** a candidate only advances if **Defensibility ≥ 3 AND Willingness
≥ 3**. Below that, it fails regardless of the other two (a cheap, reusable idea
nobody will pay for, or that free AI replicates, is not a product).

**Rank** advancing candidates by the sum of all four.

**Test-now flag:** mark "test now" if Buildability ≥ 4 **or** a demand test (e.g. a
"buy" button that records intent before building) is cheap. Only the expensive
ones need deliberation; cheap ones just get tested.

---

## Output template

```
Domain: <domain>
Audience: <one line> | Willingness to pay: <one line>

ADVANCING CANDIDATES (Defensibility ≥3 and Willingness ≥3), ranked:

1. <title>
   Idea: use <capability> on <asset> to <do X>.
   Beats free AI because: <one line>
   Verification: <what was checked — feed exists/cost, competitors, WTP evidence>
   Scores: Defensibility _/5 · Willingness _/5 · Buildability _/5 · Reuse _/5  (sum _)
   Cheapest test: <the demand test>
   [test now | consider]

DROPPED (and why): <title — reason, e.g. "free AI does it" / "premise failed verification">
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
