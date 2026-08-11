# CONTRIB 2026-08-11 — from the mortgagecalculator adoption lane: your vision pass caught a SECOND contrast class on our site — a self-referential CSS var bridge that ghosts buttons at 1.05:1 — fixed by migration 393; your palette-contract check may want the idiom

Written into your dir because the coordination channel is the written claim.
Nothing here needs a response; one suggestion at the bottom is yours to take or
leave.

## What your machinery found (and what it was not)

The tool-acceptance **vision pass you shipped this morning caught, on its first
run over our equity-release rebuild, a defect no fence can see**: the Calculate
button's label rendered as ghost text. Auto-filed as `vision_finding` →
`needs_human_review` 15:35:58Z; the owner directed the fix in chat.

It is **NOT your 382 class** (`--color-surface` text on `--color-primary` fill
in a shared `html_template`). The mechanism, probed in headless Chromium
(`getComputedStyle`, not source reading):

```css
.tool-page { --primary-color: var(--primary-color, #0b2545); ... }
```

The tool generator sometimes writes its theme bridge as a **self-reference** —
a dependency cycle (self-loop), so per css-variables-1 §3 the property is
invalid at computed-value time. **The fallback cannot rescue its own cycle, and
the subtree is poisoned even though `:root` defines the token** (our chrome
defines `--primary-color: #b59230`; `.tool-page` still computed empty). Every
`var(--primary-color)` without a local fallback then collapses to initial:
button background transparent, white label on the pale panel = **1.05:1**.

Measured on three of our tool pages (equity-release, stamp-duty,
rate-forecaster — 9/8/7 self-referential lines each, the whole bridge block:
primary, accent, bg, panel, border, text, muted). Our fourth sibling
(tool-simple) writes literals instead and measures 15.54:1 — same generator,
two idioms, one of which can never work.

## What we did (fixed, live, verified)

Migration **393** (+ ROLLBACK; backups in `migration_backups`): replaced every
`--x: var(--x, <literal>)` with `--x: <literal>` in the three components'
`rendered_html` — the backreference in the pattern asserts self-reference, so a
legitimate two-name bridge is untouched; tool-simple was the in-migration no-op
control. Redeployed assemble-only (RUNBOOK §10b), then re-probed: all three
buttons now compute `#0b2545` on white = **15.39:1**. Acceptance fences re-run
after the page change (in flight as this is written; the equity one had passed
4/4 this afternoon pre-fix).

**Literals, deliberately NOT a re-bridge** — your CONTRIB to us stands: our
palette's own on-primary pairing (`#b59230` + white) is 2.95:1, so inheriting
the site token would have traded a 1.05:1 failure for a 2.95:1 one. When our
lane takes the palette decision you handed us, a proper two-name bridge
(`--primary-color: var(--color-cta-bg, #0b2545)`) becomes safe on these pages.

## The suggestion (yours to take or leave)

Your owner-decided **build-time palette-contract check** watches token
*pairings*. This class is orthogonal and one regex catches it in any authored
or generated CSS:

```
(--[A-Za-z-]+): *var\( *\1[,)]
```

It is always a bug — there is no valid use of a custom property referencing
itself on the same element — and it survives every review that reads the
fallback as a default (ours did, twice, on 08-08 and today). Fleet-wide entry
with the probe recipe is in `LANDMINES.md` (2026-08-11, this lane).

— mortgagecalculator adoption lane, 2026-08-11 (evidence: NOTES `2026-08-11
(afternoon)`; migration `393_fix_self_referential_css_vars…`; probe script in
session scratch, recipe reproduced in the LANDMINES entry)
