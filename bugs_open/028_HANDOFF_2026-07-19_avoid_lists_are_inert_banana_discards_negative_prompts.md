# HANDOFF — every `avoid` list in the fleet is inert: Banana discards the negative prompt, and all declared kinds route to Banana

**Filed:** 2026-07-19, from the imagery workstream, after 9 generated tool heroes on
gamesdesign.co.uk violated their own `avoid` list in exactly the ways it forbids.
**Severity:** Medium-high. Nothing crashes; the imagery style guide's entire negative
half has simply had no effect since the Banana migration, and at least one documented
"hard-won fact" was attributed to it.
**Status:** OPEN. Diagnosed from code + 9 live samples. **No fix applied** — the fix
changes generation behaviour fleet-wide and belongs at the council gate.

---

## 1. The mechanism, verified in code

`internal/adapters/imagegenerator/banana/provider.go:18`:

> `// NegativePrompt on provider.Request is ignored here (Gemini has no`
> `// negative-prompt concept). Provider logs at debug level if one is`
> `// provided; callers shouldn't rely on it being honoured.`

and at line 105 it does exactly that — logs at **Debug** and drops it:

```go
if req.NegativePrompt != "" {
    p.logger.Debug("NegativePrompt provided but Banana ignores it", ...)
}
```

Meanwhile `avoid` has exactly one destination. `generate_image_actions.go:333`:

```go
if avoid := styleGuide.avoidForKind(kind); avoid != "" {
    negativePrompt += ", " + avoid          // ← the ONLY use of avoid
}
```

`directionForKind` composes the POSITIVE prompt from medium/mood/palette only. So for
any Banana-routed kind, `avoid` is **fully inert** — it is assembled, logged, shipped
over Kafka, and thrown away.

## 2. The blast radius is "all imagery"

`internal/adapters/imagegenerator/routing.go:59` — **every declared kind is on Banana**:

```go
var kindProviderRouting = map[string]string{
    "icon": banana, "logo": banana, "illustration": banana, "infographic": banana,
    "sprite_sheet": banana, "content_hero": banana, "hero": banana,
}
```

SDXL — the provider that *does* honour negative prompts — is now reached only by an
empty/legacy kind or an explicit per-site `provider:"stability"`. So:

- every `avoid` in every site's `imagery_style_guide`, and
- every `NegativePrompt` in `kindDefaults` (`generate_image_actions.go:51`, incl. the
  logo entry described in-code as "biggest expected win is logo getting any negative
  prompt at all"),

have no effect on essentially all imagery the platform now generates.

> **Caveat on `hero`:** its Banana routing is `bugs_open/011` R1, **fixed in code but
> not yet live** at time of filing. So `hero` may still be reaching SDXL in prod, where
> its negative prompt does work. Everything else has been Banana-routed and live since
> v1.0.1135 or earlier.

## 3. The live evidence that prompted this

Nine `content_hero` images generated on gamesdesign.co.uk, 2026-07-19, whose
`kinds.content_hero.avoid` explicitly lists *"text, lettering, words, numerals,
labels, captions … white background, pale background, bright full-bleed colour field"*:

| # | asset | violation |
|---|---|---|
| 1 | `drop_rate_simulator` | renders the numerals **"100 100.100"** |
| 3 | `ehp_calculator` | pale near-white diagonal bands |
| 7 | `progression_architect` | large pale region |
| 9 | `xp_curve_designer` | sits on a **near-white ground** |

**4 of 9 violate the list, in its own terms.** The other 5 comply — which is what an
ignored constraint looks like: compliance by luck, not by instruction.

## 4. A documented "hard-won fact" is very likely a misattribution

This is the expensive part, because it is written down in three places and has been
repeated as a lesson.

`HANDOFF_imagery_best_in_class.md` (D14 findings), the imagery memory, and RUNBOOK
**A6.5** all record:

> "**Style drift in the ground colour is fixed via the style guide's `avoid`, not its
> `medium`** — 'deep charcoal ground' in `medium` did not stop a white background;
> adding 'white background, pale background, light background' to `avoid` did."

Those were `content_hero` generations — **Banana-routed**, therefore with `avoid`
discarded. The `avoid` edit cannot have caused the improvement. What was observed was a
**re-roll that happened to come out darker**, and the change made alongside it took the
credit. The 4-of-9 white/pale grounds above are that supposed fix failing on its first
real test at n=9.

I am stating the code fact as **verified** and the misattribution as **strongly
implied** — I have not re-run the original D14 generations with and without the `avoid`
edit, which is the experiment that would settle it beyond doubt.

## 5. Fix candidates

1. **Fold `avoid` into the POSITIVE prompt for providers with no negative-prompt
   concept.** Gemini responds to plain instruction ("no text or lettering, no white or
   pale background"). The positive prompt is already how `content_hero` gets its "no
   text or lettering in the image" clause — which is *also* frequently ignored, so
   phrasing matters and wants testing, not assuming. Mind
   `maxImageryDirectionInPrompt = 200` (`/bugs_open/027` §4b) — appending avoid terms
   to a capped string will silently truncate something else.
2. **Make the drop loud.** `provider.go`'s Debug line is invisible in practice. A
   provider that discards a caller's constraint should say so at Warn, or the
   capability should be declared on the provider interface so the action layer can
   choose a strategy instead of shipping a field into a void.
3. **Do not "fix" this by routing back to SDXL** — SDXL was abandoned for good reasons
   (no `ReferenceImageURIs`, weak style adherence, illegible text). Losing brand
   anchoring to regain a negative prompt is a bad trade.

## 6. How to verify a fix

- Generate one `content_hero` on gamesdesign.co.uk (its guide's `avoid` names white
  grounds and numerals) and check the produced image for both.
- Confirm the constraint reached the model: read `assets.origin_prompt` — under fix (1)
  the avoid terms should be visible **in the stored prompt**, which is the same
  evidence trail that proved `/bugs_open/027` §4b.
- n=1 proves nothing here. Both defects in this file were only visible across a set;
  use 5+ and count violations rather than eyeballing one.

## 7. Related

- `/bugs_open/027` §4b — the palette-truncation defect, found in the same session. Both
  are the same shape: **a structured, brand-approved instruction is silently discarded
  between the style guide and the model**, and the output looks deliberate.
- `/bugs_open/011` — the routing work that put every kind on Banana. This bug is its
  unnoticed consequence, not a defect in it.
