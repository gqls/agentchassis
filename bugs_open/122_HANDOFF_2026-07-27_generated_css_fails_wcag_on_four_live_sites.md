# 122 — generated stylesheets fail WCAG AA on four live sites, and nothing checks

**Filed** 2026-07-27 from the oufe.com workstream, after the owner reported a link
on oufe.com as "dark blue on the black background and not easily readable".
**Severity** high — user-visible, on live public sites, and an accessibility
failure rather than a cosmetic one. On three of the four, links are effectively
invisible.
**Status** OPEN. oufe.com fixed and browser-verified; the generator unchanged.

> **CORRECTED 2026-07-28 — the measurement below was UNSOUND and the mechanism
> named in it was WRONG. Read this before using any figure in this file.**
>
> The table was produced by a regex over each site's `styles.css`, taking
> `--color-primary` to be "the link colour" because that is what oufe's generated
> CSS used. **That variable is not the link colour on the other sites.**
> dartsonline's `--color-primary` is `#111520`, near-identical to its background,
> which is why it "scored" 1.11 — the audit was comparing the background against
> itself, not measuring any rendered text. A stylesheet cannot answer this
> question at all: it cannot resolve the cascade to say which `a` rule wins, and
> the answer also depends on ancestors, alpha and gradients.
>
> **Re-measured in a real browser** (computed style, alpha-composited backdrop) on
> 2026-07-28 with the tool this bug asked for, now built as `cmd/contrastscan`.
> The corrected picture is in "What is actually wrong" below. In short: **body
> link colour was largely a false alarm; call-to-action buttons are the real
> defect, they fail on nearly every site, and on robot-hands the primary CTA is
> literally invisible.** The bug is more serious than filed, and almost none of
> the specifics were right.
>
> What was NOT wrong: the owner's original report. The Thames Water link really
> was dark blue on black, that fix was real, and oufe now passes every measurable
> pair at worst 6.86.

## Measurement (SUPERSEDED — kept because the error is the lesson)

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

---

# What is actually wrong (browser-measured, 2026-07-28)

Measured with `go run ./cmd/contrastscan <url>` — computed style, real painted
backdrop, alpha composited. "Not measurable" means a gradient or image backdrop,
which computed style cannot resolve; those are excluded rather than guessed.

## Finding 1 — the shared header CTA hardcodes white text on the site's accent colour

This is the fleet-wide defect, and it is in **stored chrome**, not in `styles.css`
(`.header-cta` does not appear in the stylesheet at all — see `bugs_open/117`,
chrome is a stored artefact in `site_components.rendered_html`):

```css
.site-header--gradient .header-cta {
    background: #C49A3C;     /* the site's accent, per site */
    color: white;            /* hardcoded, every site */
}
```

Whether that passes is **luck**: it depends entirely on how dark the site's accent
happens to be. Measured on the "Get Started" / equivalent header button:

| site | ratio | AA 4.5 |
|---|---|---|
| dartsonline.com | 4.30 | FAIL (marginal) |
| vonc.com | 3.01 | FAIL |
| relojistas.com | 2.85 | FAIL |
| oufe.com | 2.61 | **FIXED 2026-07-28** → 6.86 |
| vetcomparison.uk | 2.54 | FAIL |
| robot-hands.com | — | not measurable (gradient) |

**Five of six measured sites fail on the same hardcoded declaration.** No site's
palette can fix it, because the text colour is not derived from the palette.

## Finding 2 — robot-hands.com's primary call-to-action is invisible

Not "low contrast". **Invisible**, and confirmed by screenshot: a white button
with white text, so the homepage's primary CTA renders as a blank rectangle. The
secondary button beside it renders correctly.

| element | ratio | note |
|---|---|---|
| `.cta-btn.cta-btn-primary` "Run a MatchMatrix Query" | **1.00** | white on white |
| `.tl-card-link` "Open tool" | **1.07** | `#1A1F2E` on `#1E2535` |

This is a live commercial site with an unreadable primary CTA and an unreadable
"Open tool" link on every tool card. It has been shipping like that unnoticed,
which is the strongest argument for candidate 2 below.

## Finding 3 — vonc.com's Gauntlet buttons

| element | ratio | note |
|---|---|---|
| `.gauntlet-btn-primary` "Enter the Gauntlet" | **1.76** | purple text on a pink button, screenshot-confirmed |
| `.gauntlet-btn-secondary` "Find Your Archetype" | 4.28 | marginal fail (needs 4.5) |

**Coordinate before touching these** — the Gauntlet workstream owns this surface.

## Finding 4 — assorted, lower severity

`relojistas.com` `.news-more-link` 2.63; several `rgba(255,255,255,0.7)` nav and
secondary items in the 3.3–4.4 band across sites.

# Fix candidates, re-ordered by what closes the door

1. **Stop the header chrome hardcoding `color: white`.** Derive the CTA text
   colour from the accent's own luminance (black or the site background, whichever
   passes) at generation time. This is the single change that fixes five sites,
   and it makes the failure unrepresentable rather than per-site-lucky. Note the
   chrome is a **stored artefact**: existing sites need the `nav_drift` →
   `nav-updater` path, since no page re-render rebuilds chrome (`117`).
2. **Run `cmd/contrastscan` over the fleet as a post-deploy check** and raise
   `contrast_failure` at high severity. Built 2026-07-28. This is the candidate
   that would have caught finding 2, which nobody noticed on a live site.
3. **Gate the webdesign agent's CSS output** at generation. Still worth doing, and
   note it would NOT have caught findings 1–3, because none of them are in the
   generated stylesheet.
4. Fix sites by hand. Done for oufe only.

# How to verify, and the trap in verifying

```bash
go run ./cmd/contrastscan https://<site>/            # exits non-zero on any failure
```

**Screenshot anything you are about to report on a live site.** Three separate
false positives were produced while building this measurement, each of which
looked like a serious live defect and each of which was disproved by looking at
the page: comparing against the page background instead of the element's;
treating a gradient header as transparent and falling through to the body; and
treating `rgba(255,255,255,0.1)` as opaque white. All three are now guarded in
the tool, and the guards are commented with the false positive that motivated
them.

The general form: **a contrast audit that over-reports is worse than none**, because
its findings get "fixed" into real regressions, and because people stop reading it.

## Relationship to the 026 palette contrast check (found 2026-07-28, after building the above)

Another thread shipped `platform/colour.AuditPalette` on 2026-07-27 (`6dd8667ea`,
"finds defects on 7 of 10 live sites"). I found it only after building
`cmd/contrastscan`, which is a prior-art miss on my part — the grep that would
have caught it is `AuditPalette|contrast` over `platform/`, and I searched
`cmd/` and `scripts/` only.

**They are complementary, and this is worth stating explicitly so neither is
deleted as a duplicate of the other.** They audit different layers:

- `AuditPalette` reads the **composed palette** from
  `site_specs.resolved_composition.palette_id`. DB-only, microseconds, and it can
  run *before* a deploy. Its own load-bearing insight is that intent != artefact,
  which is why it reads the composed row rather than `design_intent`.
- `cmd/contrastscan` reads the **painted page**. Seconds per page, post-deploy
  only.

The gap that needs both: **a colour can be legible in the palette and illegible on
the page**, because chrome and component CSS carry hardcoded literals that are in
no palette. Finding 1 above is precisely that — `color: white` hardcoded in
`site_components.rendered_html`. I checked oufe's composed palette and the CTA
text colour **is not in it**, so no palette audit at any level of quality can see
this defect.

So: `AuditPalette` is the cheap pre-deploy gate that should run every build;
`contrastscan` is the post-deploy witness for everything the palette does not
govern. Candidate 1 (stop the chrome hardcoding white) remains the fix that
closes the door, because it removes the literal rather than detecting it.
