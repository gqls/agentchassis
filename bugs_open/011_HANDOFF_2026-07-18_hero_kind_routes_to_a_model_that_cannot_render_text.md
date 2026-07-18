# HANDOFF — `kind:"hero"` routes to a model that cannot render text; the lane that can was never used

**Filed:** 2026-07-18, from the leopardessconsulting.co.uk rebuild (owner review).
**Severity:** Medium. No code change is strictly required to get good infographics today — the
capability is already wired. The bug is a routing default plus an unused lane.
**Status:** **R1 FIXED in code 2026-07-18** (inert until a chassis + image-generator-adapter
image ships — see §6). R2/R3/R4 remain OPEN. The infographic capability itself is
**PROVEN WORKING**.

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

**R1 — fix the hero routing default. ✅ FIXED IN CODE 2026-07-18 — see §6 for what shipped
and what is still owed.** Choose the provider from the site's
`design_intent.imagery_direction` (or an explicit per-site provider preference), not from the
kind string alone. A site declaring a flat-illustration house style should never have its
heroes sent to the photographic model. Low risk, fleet-wide benefit.

> **Route not taken, and why.** The obvious reading of R1 — infer the provider by keyword-
> matching the free-text `design_intent.imagery_direction` — was tried against all 11 live
> values first and **rejected**: it misfires on at least three. Site `9ec3b9ee` reads
> *"Minimal photography. Prefer abstract geometric constructions…"* and `1244516d` reads
> *"Photography and illustration should be minimal…"* — both contain "photography" while
> intending the opposite, so a substring match would misroute them **silently**. An explicit
> per-site field plus a better default beats a fuzzy guess over prose the planner never wrote
> for this purpose.

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

## 6. R1 — what shipped, 2026-07-18 (and the one thing still owed)

Fixed as one commit on `085_debug_and_feature_loops`. Council gate correlation
`e996bf0a-4cdd-40fa-8ff0-1f1a76c3d181` — **three rounds, final verdict REVISE, not APPROVED**;
what remains open is recorded below rather than quietly dropped.

**The bug was bigger than "hero is on the wrong model".** The real defect is the *mechanism*:
provider selection was a hand-maintained `switch` whose `default:` branch routed to Stability
**silently**. `content_hero` fell through it and shipped mis-routed; `hero` fell through it and
shipped a gibberish diagram as a client homepage. Both were found months later by a human
looking at an image, because generation reports success and dropped brand anchors say nothing.
Adding `hero` to the switch's list would have fixed instance three and left instance four to be
found the same way — this is the council's `bug_historian` objection, and it was right.

What landed:

1. **`internal/adapters/imagegenerator/routing.go` (new).** The switch is now an enumerable
   table `kindProviderRouting` plus a pure `routeProvider` function. Because the routed set is
   data the code can interrogate, a non-empty kind *absent* from it is **detected**: the adapter
   logs `UNROUTED KIND` naming the kind and listing the valid set, instead of quietly serving
   from the weaker provider. Adding a kind is still a code change; **forgetting to is no longer
   silent.** An empty kind deliberately does *not* warn — legacy callers predating the field are
   a documented Stability path, and a warning that fires constantly is one nobody reads.
2. **`hero` joined the routed set** — the last kind left behind, by omission, and the largest:
   **84 of 155** `site_plan_imagery` rows.
3. **Per-site escape hatch.** `imagery_style_guide` gained an optional `provider` field
   (`"banana"` | `"stability"`), guide-level and per-kind, resolved by `providerForKind` —
   mirroring `avoidForKind`'s override-wins-**even-when-empty** contract — and passed to the
   adapter as `provider_hint`. A site wanting SDXL heroes back sets **data, not code**.
   The adapter has no DB handle, which is *why* routing was hardcoded; resolving in the action
   layer and shipping the answer as data is what makes the decision site-owned.
4. **Tests** for the two widest-blast-radius behaviours plus the guard itself
   (`routing_test.go`, `TestProviderForKind`).

**Verified this round** (both were council escalate-conditions, so they are checked, not
asserted): `ImageRequestData{}` is constructed **nowhere** in the repo — the adapter only
unmarshals it from Kafka JSON — and only three files touch the adapter topic
(`topic_manager.go` declares it, `generate_image_actions.go` is the sole producer, the adapter
consumes). **`GenerateImageAction` is the exclusive path, so no caller bypasses the plumbing.**
The `imagery_style_guide` JSON likewise has exactly one reader (`getImageryStyleGuideForSite`)
plus one seed file — no UI, frontend or other service — so the new field is safe to add.

### Still owed — the residual objection (`bug_historian`, medium; `guardian`, low)

**`UnmigratedKind` is a log line, not a record.** It closes the silent-failure gap only for
someone reading logs on the right pod — which this repo's own history says is unreliable. The
platform already has the right shape for this: `agent_error_log(severity, resolved,
work_item_id, context)`, and `site_work_items` for anything that should demand action. A
fleet-wide dashboard should be able to catch the next unmigrated kind; today only a human
tailing logs can.

**Why it was not done in this commit, and the trap for whoever does it.** The image-generator
adapter has **no database handle at all** (`grep sql.DB|pgxpool` in `dynamic_adapter.go`
returns nothing) — persisting from there means giving an adapter service a DB dependency, which
is an architectural change, not a small one. The tempting shortcut is to detect it in the
**action** layer instead, which does have a DB — **do not do this**: the action and the adapter
are *separate services on separate images*, so the action would be predicting a routing table
that may not match the one actually deployed in the adapter. That is the
dedup-index↔Go-list drift class this platform has already been bitten by. The structurally
correct shape is **adapter reports the condition in its response → the action, which has the DB
and the orchestration context, persists it**. That is a coherent task of its own.

**Also unresolved and owner-facing: cost and latency.** This moves the fleet's largest image
kind onto Gemini. **No cost or latency parity has been established** — no billing data was
available for either provider and none is asserted here. The adapter's HTTP timeout (120s) was
tuned around SDXL's 30-60s generation. Reversible per-site as data (`provider:"stability"`) and
fleet-wide by one line in `kindProviderRouting`.

**Content risk:** photographic-house-style sites (`00ff3af5` robot-hands, `5fe8785b` darts,
`ecf15e75` relojistas) will get *new* heroes from a different model than their existing ones,
so a page can mix two visual languages until regenerated.

**Deploy state: INERT.** Go changes do nothing until an image is rebuilt and rolled. Verify
against the **running pod**, never the tag:
```
kubectl exec -n ai-persona-system <image-generator-adapter-pod> -- \
  sh -c 'strings /app/image-generator-adapter | grep -c "UNROUTED KIND"'
```

## 5. Key files

- `internal/adapters/imagegenerator/dynamic_adapter.go` — the provider switch (the routing fix)
- `internal/adapters/imagegenerator/banana/provider.go` — Banana/Gemini provider; model from `BANANA_DEFAULT_MODEL`
- `platform/orchestration/actions/generate_image_actions.go` — prompt assembly, `constraints` → negative prompt, per-kind defaults
- `docs/leopardessconsulting/PLAN_imagery_and_design_2026-07-18.md` — the site-side plan this came from
- Working example prompt + result: scratchpad `infographic.json`; live asset `infographic-what-we-build.jpg`
