# HANDOFF — `kind:"hero"` routes to a model that cannot render text; the lane that can was never used

**Filed:** 2026-07-18, from the leopardessconsulting.co.uk rebuild (owner review).
**Severity:** Medium. No code change is strictly required to get good infographics today — the
capability is already wired. The bug is a routing default plus an unused lane.
**Status:** OPEN (routing fix + guard). The infographic capability itself is **PROVEN WORKING**.

> ## ⚠️ CORRECTION — read this before anything else
> The first version of this handoff (same number, 2026-07-18 morning) claimed *"generated
> images cannot render readable text"* and recommended building an SVG renderer because
> diffusion models "synthesise glyph-shaped texture, not text". **That claim was wrong**, and
> a thread acting on it would have built the wrong thing.
>
> The owner produced two Gemini infographics with perfectly legible, correctly-spelled text,
> then asked whether we could wire that up. We already had: the deployed
> `BANANA_DEFAULT_MODEL` is **`gemini-3-pro-image-preview`**, and `kind:"infographic"`
> **already routes to it**. Generating through that lane produced a production-quality
> infographic on the first attempt — legible throughout, correct figures, on-brand.
> Evidence: `https://leopardessconsulting.co.uk/assets/images/infographic-what-we-build.jpg`
> (asset `infographic_what_we_build`, `origin_model=banana/gemini-3-pro-image-preview`).
>
> The generalisation "image models can't do text" was true of SDXL and is no longer true of
> the current Gemini image model. Corrected in place rather than deleted, because the wrong
> inference is an easy one to repeat.

---

## 1. What is actually broken

**The garbled homepage hero was a routing accident, not a capability limit.**

`internal/adapters/imagegenerator/dynamic_adapter.go` switches provider on `kind`:

```
icon | logo | illustration | infographic | sprite_sheet | content_hero  → Banana (gemini-3-pro-image-preview)
everything else (including "hero")                                     → Stability (SDXL v1.0)
```

So `kind:"hero"` gets SDXL. SDXL genuinely cannot render text. When a hero prompt implies any
structure ("a diagram of a pipeline…"), SDXL returns a convincing-looking flowchart full of
gibberish words — which is exactly what shipped as this site's homepage hero
(`/assets/images/hero.jpg`, still live as the site-wide fallback and still on how-it-works).

Two consequences worth separating:

1. **Routing default.** For a site whose house style is flat illustration, `hero` is the one
   kind that lands on the photographic model least able to serve it. Heroes on this site only
   became good when explicitly requested as `kind:"illustration"`.
2. **Unused lane.** `infographic` has routed to the capable model all along. Nothing on this
   site had ever used it. The "we can't make infographics" belief was self-inflicted.

## 2. What works today (verified, no code change needed)

- `kind:"infographic"` → Banana → `gemini-3-pro-image-preview`, with a **richly specified
  prompt**, produces publishable infographics with legible, accurate text.
- Prompt specificity is the dominant variable. The successful prompt names the layout, every
  column header, every card's heading and body text verbatim, the exact figures permitted, the
  palette by hex, the icon for each card, and ends with an explicit instruction that all text
  must be correctly spelled and real, and that **no number outside the supplied list may
  appear**. Thin prompts are what produced the earlier rubbish.
- `kind:"illustration"` → Banana with hard no-text constraints produces good text-free heroes
  (three now live on this site).

## 3. Remaining real work

**R1 — fix the hero routing default.** Choose the provider from the site's
`design_intent.imagery_direction` (or an explicit per-site provider preference), not from the
kind string alone. A site declaring a flat-illustration house style should never have its
heroes sent to the photographic model. Low risk, fleet-wide benefit.

**R2 — a text-legibility guard before publish.** Even the good model is not perfect: the
owner's own Gemini map rendered "REPRETITIVE" for "REPETITIVE". A typo in a generated
infographic is a real defect on a professional site, and no pipeline signal catches it —
generation reports success. Add an OCR/vision pass after generation that (a) extracts the
rendered text and (b) flags misspellings and any number not present in the request. Route
findings to human review; never auto-publish an image whose text failed the check. This is the
same check→work-item→HITL shape as the claims and voice gates.

**R3 — numbers must come from the evidence base.** The generated infographic is accurate
because the prompt carried audited figures and forbade any others. That should be structural,
not a matter of prompt discipline: build infographic prompts from
`site_specs.evidence_base` facts so an infographic cannot state an unverified number. Ties
directly into the claims-verification layer.

**R4 — keep code-rendered SVG for exact data.** Generated infographics are now good enough for
*explanatory* graphics. They are still the wrong tool for a chart whose values must be exactly
right, selectable, translatable and screen-reader accessible. The L7 chart component (Go emits
SVG from real values) remains worth building for data; it is no longer needed for explanation.

## 4. Blast radius

Fleet-wide. Every site's heroes route by kind, so every site with an illustration house style
has the same mismatch, and no site has used the infographic lane. Any already-deployed image
generated as `hero` with a structural prompt is a candidate for the R2 sweep once the detector
exists.

## 5. Key files

- `internal/adapters/imagegenerator/dynamic_adapter.go` — the provider switch (the routing fix)
- `internal/adapters/imagegenerator/banana/provider.go` — Banana/Gemini provider; model from `BANANA_DEFAULT_MODEL`
- `platform/orchestration/actions/generate_image_actions.go` — prompt assembly, `constraints` → negative prompt, per-kind defaults
- `docs/leopardessconsulting/PLAN_imagery_and_design_2026-07-18.md` — the site-side plan this came from
- Working example prompt + result: scratchpad `infographic.json`; live asset `infographic-what-we-build.jpg`
