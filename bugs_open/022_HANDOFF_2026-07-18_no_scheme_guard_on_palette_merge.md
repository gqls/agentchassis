# 022 — Nothing stops a spec from flipping a site's colour scheme: light palette rendered onto a dark layout

**Filed:** 2026-07-18 from the robot-hands R1 thread. This is the **damage
mechanism** behind that incident; `bugs_open/017` and the `generic_theme` check
fix (commits 3437f2212 + 3b52da8ec) only reduce how often it is *triggered*.
Raised by the council gate's `bug_historian` seat on correlation
`e0ebf6ee-dcc0-4a7b-9a3d-438ce9af5fff` (round 2), then **withdrawn from that
submission on the council's own recommendation** — guardian's stability point:
a shared render-merge change should not ride along with a single-file check fix.
It needs its own submission, which this file is the groundwork for.

## Symptom (real, shipped)

robot-hands.com runs layout `tool-portal-dark` (`layouts.scheme='dark'`, a user
decision taken twice: B7 on 2026-07-10 and again at the D13 gate). On
2026-07-17 20:31 a routine `webdesign-agent` run committed
`robot-hands.com/assets/css/styles.css` (gqls/sites `302702fc96`) carrying:

```
--color-background:     #F4F5F7;    /* light — on a scheme=dark site */
```

No error, no warning, no gate. It went live. Four CSS rewrites landed that day
(`5dfe168347` 08:54, `4b6685a422` 13:34, `bd814de701` 16:24, `302702fc96`
20:31), each with different core colours; only the last was catastrophic.

## Mechanism

`analyze_design` (webdesign-agent, `execute_llm_prompt`) emits a fresh
`color_scheme` on **every** run. `RenderCSSFromSpecAction` merges it at
`render_css_from_spec_action.go:119`:

```go
mergedPalette := buildPaletteMap(comp.Palette, specPalette)
```

`buildPaletteMap` (`render_css_composition_helpers.go:72`) applies the
documented core-vs-specialised rule: for **core** slots — `primary`,
`secondary`, `accent`, `background`, `surface`, `text`, `text_muted`, `border` —
**the spec always wins**. That rule is right for site identity and wrong for
*scheme*: nothing anywhere compares the proposed `background` against the
layout's declared `scheme`. The layout is the user's choice; a per-run LLM
guess silently overrides it.

Per-site mitigation applied (robot-hands only): pinned
`design_intent.palette.reference_values` so the prompt stops inventing —
`docs024_key_docs_latest/robot_hands/SQL_2026-07-17_r1b_design_intent_palette_pin.sql`.
That is a data patch on one site; every other site with a `scheme` layout and no
pin has the same exposure.

## Verifications already done (the council asked; these all resolve in favour)

- `parseHexColor` — **exists**, `platform/orchestration/actions/color_util.go:26`.
  `relativeLuminance` at `:64`. No new helper needed (reuse_agent's concern).
- `layouts.scheme` — **exists**: `text`, `CHECK (scheme IN ('light','dark',
  'neutral'))`; robot-hands' `tool-portal-dark` row is `scheme='dark'`.
  It is NOT currently carried on `themeComposition`
  (`render_css_composition_loader.go:37`) — the loader's JOIN must select it.
- Call sites — `buildPaletteMap` and `loadThemeComposition` have **exactly one
  non-test caller each** (`render_css_from_spec_action.go:119` and `:107`).
  So a guard at that boundary is NOT the "one call site patched, mechanism left
  generic" shape bug_historian feared; today it *is* the mechanism. (It must be
  re-checked at implementation time — that count is the load-bearing fact.)

## Fix candidate

Add `LayoutScheme` to `themeComposition` (populate from `l.scheme` in
`loadThemeComposition`'s JOIN), then guard immediately after the merge at
`:119`: when the layout declares `dark` and the merged `background` has
luminance > 0.5 (or the mirror case for `light`), keep the **theme's**
`background` *and* `text` together — never half-swap, that breaks contrast —
and `logger.Warn` naming both the rejected and kept values. Sites with
`scheme` NULL/`neutral` are untouched.

Open question for the submission (guardian): confirm no legitimate workflow
renders a light palette against a dark layout — a deliberate rebrand should
swap the layout first, which is the sanctioned 025 FK-swap order anyway.

## How to verify

- Unit: merged palette with a light background + `LayoutScheme="dark"` returns
  the theme background, and emits the Warn.
- Live: re-run webdesign-agent on a `scheme=dark` site with the design_intent
  pin REMOVED; the committed styles.css must stay dark and the pod log must
  carry the rejection line.
- Deploy (debug_historian's round-2 point): Go is inert until an image
  build+roll — verify with a pod-binary grep for the new symbol, never the
  commit hash or tag.
