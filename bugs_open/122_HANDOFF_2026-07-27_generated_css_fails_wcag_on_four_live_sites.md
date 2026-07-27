# 122 — generated stylesheets fail WCAG AA on four live sites, and nothing checks

**Filed** 2026-07-27 from the oufe.com workstream, after the owner reported a link
on oufe.com as "dark blue on the black background and not easily readable".
**Severity** high — user-visible, on live public sites, and an accessibility
failure rather than a cosmetic one. On three of the four, links are effectively
invisible.
**Status** OPEN. oufe.com fixed by hand; the other three untouched and the
generator unchanged.

## Measurement

WCAG relative-luminance formula, `--color-primary` (which the generated CSS uses
as the link colour) against each site's own background and surface:

| site | link on background | link on a card | verdict |
|---|---|---|---|
| **oufe.com** | **1.23** | **1.00** | fixed by hand 2026-07-27 |
| **dartsonline.com** | **1.11** | **1.06** | untouched |
| **robot-hands.com** | **1.14** | **1.07** | untouched |
| **vonc.com** | **3.71** | **3.48** | untouched (fails AA body, passes large text) |
| idea.uk | 14.39 | 13.37 | fine |
| leopardessconsulting.co.uk | 18.32 | 19.44 | fine |
| relojistas.com | 10.83 | 11.50 | fine |
| fundamentallyai.com | 8.30 | 7.19 | fine |
| gamesdesign.co.uk | 8.16 | 7.26 | fine |
| webdesign.co.uk | 5.32 | 5.65 | fine |
| vetcomparison.uk | 4.94 | 5.17 | fine |

AA requires **4.5** for body text. **A ratio of 1.00 is not low contrast, it is
the same colour** — on those palettes `--color-primary` and `--color-surface` hold
an identical hex, so a link inside any panel is invisible.

## The cause, and why it is not a palette problem

The generator's own prompt is **already correct**. `031_webdesign_agent.sql`
instructs, in three separate revisions of the prompt:

```
- a { color: var(--color-accent); }
```

oufe's deployed stylesheet contains:

```css
a {
  color: var(--color-primary);
```

So the instruction is right and the model did not follow it. Nothing downstream
noticed, because **nothing checks generated CSS for contrast** — a gap already
recorded in `docs/leopardessconsulting/RUNBOOK.md:156-157`: *"Validate every
text/bg pair with the WCAG formula in `color_util.go` first — the platform does
NOT gate specialised-slot contrast."*

The maths already exists: `wcagContrastRatio` at
`platform/orchestration/actions/color_util.go:43`, with a documented AA floor for
section backgrounds at `:110`. It is simply never run over the stylesheet the
webdesign agent emits.

## A second finding: specialised slots leak a template's palette

oufe is navy and gold. Its generated stylesheet also declares:

```css
--color-primary-hover:  #6b0000;   /* dark red */
--color-cta-bg:         #8b0000;   /* dark red */
--color-badge-bg:       #8b0000;   /* dark red */
--color-card-bg:        #ffffff;   /* white cards, on a dark site */
--color-header-bg:      #ffffff;
```

Those reds have no relationship to the site's palette and are template defaults
that were never forked. `--color-card-bg: #ffffff` is latent rather than live —
`.article-card` is not currently rendered on oufe — but light body text on a white
card is the same invisibility bug waiting for a news page to be added.

## A trap this bug set for me, worth repeating

My first audit compared **every** foreground colour against the page background
and flagged the header, logo and nav as failing at ~1.03. They are fine: they sit
on `--color-header-bg: #ffffff`, where the near-black text scores **17.40**.
Switching them to the accent, as the naive fix would, drops them to **2.6** and
breaks what works.

**A contrast check must resolve the element's ACTUAL background, not the page's.**
Any automated gate built from this bug has to do the same, or it will "fix" sites
into a worse state than it found them.

## Fix candidates, ordered by what closes the door

1. **Gate the webdesign agent's output.** After CSS generation, parse the
   `:root` block, resolve every `color:`/`background:` pair a rule creates, and
   fail the step when a text pair is under 4.5 (3.0 for large text). The maths is
   already in `color_util.go`. Makes the bad stylesheet unrepresentable rather
   than merely discouraged, and catches the leaked-slot class too.
2. **A post-deploy discovery check**, in the shape of `check_voice_tells` — scan
   the live stylesheet, raise `contrast_failure` at high severity to human review.
   Catches drift and the eight sites already deployed, but only after the fact.
3. **Fix the four sites by hand.** What was done for oufe. Immediate, and it does
   nothing for site number twelve.

1 and 2 are complements rather than alternatives: one stops new failures, the
other finds the ones already live. Neither is a substitute for the other, and 3
alone is the state that produced this.

## How to verify a fix

Build a site with a dark palette where `primary` and `surface` share a hex — the
exact shape that produced 1.00 here. Candidate 1 must fail the CSS step. Then
build one where they differ and confirm it passes: **a gate that fires on
everything is as useless as one that fires on nothing**, and this one has real
scope to over-fire because of the header trap above.

## Related

- `docs/leopardessconsulting/RUNBOOK.md:156` — the gap, recorded months ago.
- `platform/orchestration/actions/fix_forced_text_colours_action.go` — an existing
  action in the same family, worth reading before writing a new one.
- `bugs_open/107` — every site gets the same skeleton; same root, a generator
  whose output nothing measures against the site it is for.
