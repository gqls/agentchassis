# Proposal — vetcomparison accent home + industry-hub section-imagery treatment

Design work only, per `vetcomparison`'s "no writes, no wait" (imagery) / "hold
behind 701" (everything palette-related). Nothing in this file is applied.
Blast radius fact requested before shipping: **3 deployed sites use
`industry-hub`** — `vetcomparison.uk`, `farmerinsurance.uk`, `garden-tools.uk`.

## 1. Accent values — HOLD behind 701 (touches a palette + a shared component)

**Palette (`palette-vetcomparison-uk`), two currently-unset slots:**

| slot | proposed value | why |
|---|---|---|
| `independence_bg` | `#ecfdf5` | a light emerald tint (pairs with `#10b981`, not a re-used blue) — text inside stays on `--color-accent-text`/`-ink` (`#0f172a`), so background lightness is not itself contrast-critical |
| `independence_border` | `#10b981` (i.e. `--color-accent` verbatim) | non-text, 3:1 threshold — emerald-500-class green on a near-white tint clears this comfortably; re-using the accent variable directly (not a new literal) keeps it tied to the palette rather than a second hardcoded green |

**Constraint honoured:** nothing here sets accent as TEXT colour. Any copy
inside the callout keeps `color: var(--color-accent-text, var(--color-primary))`
— falls back to primary, never to raw accent, if the derived slot is ever
absent.

**Amber-fallback normalisation, same change, also held:** the `latest-news`
component (`content_components` id `77dafa26…`, live on **8** deployed sites)
hardcodes `#d97706` (amber) as the CSS fallback on every `var(--color-accent, …)`
usage — inconsistent with the accent's own real value and with `industry-hub`'s
own `:root` fallback (`#1e40af`, blue). Proposed: change all five occurrences'
fallback literal from `#d97706` to `#10b981` (the palette's actual accent) —
a fallback should degrade to something coherent with the brand, not a third,
unrelated hue. **This is a shared-template edit reachable from vetcomparison's
page, so it's held behind 701 with everything else, not shipped separately.**

## 2. Section-illustration CSS treatment for `industry-hub` — cleared to design now, NOT yet applied

**Additive and inert by construction**: nothing below fires unless a section
wrapper carries the new modifier class, so it changes nothing for any page
(on any of the 3 industry-hub sites, or none) until something opts in.

**Shape:** a section-level image sits beside its content on desktop, stacks
above it on mobile, alternating sides is author-controlled (a second modifier),
matching the layout's existing `768px` mobile breakpoint and its established
spacing tokens (`--section-pad-y`, `--card-pad`) rather than inventing new ones.

```css
/* ── Section illustration — additive, opt-in per section ── */
.section-with-illustration {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: clamp(2rem, 5vw, 4rem);
  align-items: center;
}
.section-with-illustration.illustration-right { grid-template-columns: 1fr 1fr; }
.section-with-illustration.illustration-right .section-illustration { order: 2; }
.section-with-illustration.illustration-left  .section-illustration { order: -1; }

.section-illustration {
  border-radius: var(--radius);
  overflow: hidden;
  box-shadow: var(--shadow-md);
}
.section-illustration img {
  width: 100%;
  height: 100%;
  object-fit: cover;
  display: block;
}

@media (max-width: 768px) {
  .section-with-illustration {
    grid-template-columns: 1fr;
    gap: 1.5rem;
  }
  .section-with-illustration.illustration-right .section-illustration,
  .section-with-illustration.illustration-left  .section-illustration {
    order: -1; /* image above content on mobile, both variants */
  }
}
```

**Deliberately NOT specified here:** which sections on vetcomparison's index
get this treatment, or the actual image content/generation (that is
`site_assets`/`site_plan_imagery` + IMG-075's own machinery, and the site's
existing `design_intent` imagery direction — photographic white/teal, no
close-up generated faces — already governs asset selection; nothing in this
CSS needs to know that). This file specifies the layout capability, not the
per-page decision of where to use it.

**Sizing note for whoever reviews this**: 3 sites, additive-only — low risk to
ship to the shared `industry-hub` template once reviewed. If any site later
needs non-additive behaviour (e.g. a different grid ratio), a per-site override
via `sites.style_overrides` is the right lever, not editing this shared block.

## 3. What's still open, not this file's to answer

- Which of vetcomparison's four index sections (hero / info-card-grid /
  latest-news / call-to-action) actually gets an illustration, and how many —
  a content/imagery-planning decision, not a CSS one.
- Exact review path this goes through before the CSS is committed to
  `layouts.css_template` — `vetcomparison` asked that the site count be stated
  "in whatever review it goes through"; done here (§ heading), routing is
  theirs to run.
