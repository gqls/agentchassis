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
> 2026-07-28 in headless Chromium (see the note on tooling at the foot of this file).
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

Measured in a browser — computed style, real painted
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
2. **Wire `scripts/render_audit.py` over the fleet as a post-deploy check** and
   raise `contrast_failure` at high severity. **The tool already exists** (built
   2026-07-27 by the brochure workstream) and is **not wired to anything** — that
   is the whole of the remaining work for this candidate. It is the only thing
   that catches 026 family 3, which `check_palette_contrast` states in its own
   header it cannot see. This is the candidate that would have caught finding 2,
   which nobody noticed on a live site.

   > **CORRECTED 2026-07-28:** this bug previously said the candidate was "built"
   > as `scripts/render_audit.py`. That was a duplicate of `render_audit.py`, written
   > without finding it because the prior-art grep was `--include=*.go` and the
   > prior art is Python. The Go tool has been deleted.
3. **Gate the webdesign agent's CSS output** at generation. Still worth doing, and
   note it would NOT have caught findings 1–3, because none of them are in the
   generated stylesheet.
4. Fix sites by hand. Done for oufe only.

# How to verify, and the trap in verifying

```bash
python3 scripts/render_audit.py https://<site>/      # every element, not a selector list
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
`scripts/render_audit.py`, which is a prior-art miss on my part — the grep that would
have caught it is `AuditPalette|contrast` over `platform/`, and I searched
`cmd/` and `scripts/` only.

**They are complementary, and this is worth stating explicitly so neither is
deleted as a duplicate of the other.** They audit different layers:

- `AuditPalette` reads the **composed palette** from
  `site_specs.resolved_composition.palette_id`. DB-only, microseconds, and it can
  run *before* a deploy. Its own load-bearing insight is that intent != artefact,
  which is why it reads the composed row rather than `design_intent`.
- `scripts/render_audit.py` reads the **painted page**. Seconds per page, post-deploy
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

## Contribution 2026-07-29 — dartsonline.com: the mechanism is the LAYOUT'S OWN LIGHT LITERALS, and the fix already shipped

From the `dartsonline_traffic` workstream. Contributing here rather than opening a
parallel account, and **not** starting a competing generator fix — candidate 1
still belongs to whoever takes it.

**Re-measured on the live page** with `scripts/render_audit.py`, 2026-07-29 16:45Z:

```
FAIL https://dartsonline.com/                      contrast=13
FAIL https://dartsonline.com/blog/barrel-weight.html  contrast=5
```

The homepage's whole card grid is unreadable, and the numbers say why:

| pair | ratio | element |
|---|---|---|
| `rgb(240,242,247)` on `rgb(255,255,255)` | **1.12** | `.info-card-grid__card-title` — "Find Your Barrel Weight" |
| `rgb(240,242,247)` on `rgb(255,255,255)` | **1.12** | `.info-card-grid__card` link — "Read the tungsten guide" |
| `rgb(17,21,32)` on `rgb(13,16,25)` | **1.04** | `.info-card-grid__eyebrow` — "SPEC-FIRST BUYING GUIDES" |

Near-white text on white cards, on a site whose background is near-black. Four
cards, every one of them.

**The component is not at fault, and this is the part worth carrying.** Every rule
is variable-driven and correct:

```css
.info-card-grid__card       { background: var(--color-card-bg, var(--color-surface)); }
.info-card-grid__card-title { color: var(--color-heading); }
.info-card-grid__eyebrow    { color: var(--color-primary); }
```

The defect is in the served `styles.css`, which is the **`ecommerce-storefront`**
layout with its light-scheme literals intact:

```css
--color-card-bg:   #ffffff;
--color-header-bg: #ffffff;
--color-product-bg:#ffffff;   /* "Product cards stay neutral regardless of
                                  palette — product images demand a clean
                                  backdrop." */
--color-background:#0D1019;   /* the palette DID land here */
--color-text:      #F0F2F7;
```

So the site's dark palette filled the eight core slots and the layout's white
literals survived in the nine it does not define. That is **exactly** the defect
`platform/orchestration/actions/palette_specialised_slots.go` was written for on
2026-07-27 — its own header names fundamentallyai.com at 1.21:1 and measures "16
of 31 palettes define no `card_bg`, 12 of those dark". dartsonline is one of them
and nobody had connected it to this bug file.

**So there is nothing to fix in the generator for this site — the fix is live and
the CSS is simply stale.** Pod-verified on both replicas before acting:

```
$ kubectl exec -n ai-persona-system agent-chassis-6fd7d88c4d-f6pgj -- \
    sh -c 'strings /app/agent-chassis | grep -c "a card is a raised surface"'
1                       # and 1 on ...-ktnzr
```

Corroborating evidence that the stylesheet predates the palette: the served CSS
does not match the `palettes` row at all. Served `--color-background: #0D1019`,
`--color-primary: #111520`, `--color-surface: #1A1F2E`; the palette row
(`b6f34e0e`) says `background #111520`, `primary #1A1F2E`, `surface #1E2436` —
the served values are the palette's, shifted by one role, plus a `#0D1019` that
appears nowhere in it. A re-render was triggered at 16:47Z via `webdesign-agent`
(correlation `4555d081-65c2-47fc-a89b-e7c6a96badc5`).

**Two things this adds to the file's own conclusions.**

1. The 07-28 correction says `--color-primary` "is not the link colour on the
   other sites". Correct — but on dartsonline it *is* an ink: the card-grid
   eyebrow uses it, at 1.04:1. So the original 1.11 figure was arrived at by an
   unsound method **and** there is a real near-invisible element using that
   variable. Both statements are true, and the second one does not rescue the
   first.
2. The file's closing argument — palette audit is the cheap pre-deploy gate,
   render audit the post-deploy witness — has a **third** gap between them that
   dartsonline demonstrates: a palette can be perfectly legible, the component
   CSS can hardcode nothing, and the page can still fail, because the LAYOUT
   supplies literals for slots the palette never names. `AuditPalette` cannot see
   it (the value is not in the palette), and `render_audit.py` sees it only after
   it is already live and public. The pre-deploy check that would catch it is
   over the *composed* layout+palette pair, which is what
   `palette_specialised_slots.go` now derives — so the door is closed for new
   renders and **open for every site whose CSS has not been re-rendered since
   2026-07-27**. That set is not enumerated anywhere; it is probably the rest of
   this bug's population.

### Result, measured on the live page 2026-07-29 17:00Z

The re-render deployed (`gqls/sites`, "Update stylesheet via webdesign-agent") and
the derivation fired — `--color-card-bg` went `#ffffff` → `#1A1F2E`,
`--color-header-bg` `#ffffff` → `#0E1019`, `--color-cta-bg`/`--color-cta-text`
`#111827`/`#ffffff` → `#1A1F2E`/`#F0F2F7`:

```
before   FAIL  https://dartsonline.com/                      contrast=13  h-overflow
         FAIL  https://dartsonline.com/blog/barrel-weight.html  contrast=5   h-overflow
after    FAIL  https://dartsonline.com/                      contrast=1
         FAIL  https://dartsonline.com/blog/barrel-weight.html  contrast=0
```

18 failures → 1, and the horizontal overflow went with it. **No site-specific
palette edit was needed and none was made** — the site was simply carrying a
stylesheet older than the fix.

**The one that remains cannot be fixed at site level, and the arithmetic says
why.** `.info-card-grid__eyebrow` uses `color: var(--color-primary)`, which is
`#111520` on a `#0E1019` ground: **1.04:1**. But the same layout also uses
`--color-primary` as a FILL, with `--color-primary-text` on top of it. Both
consumers are on the served homepage:

```
background: var(--color-primary);  color: var(--color-primary-text);   /* button */
color: var(--color-primary);                                            /* eyebrow, card links */
border: 3px solid var(--color-primary);
```

So one variable is asked to be legible *on* the dark page and to *be* a dark page
for light text. Measured candidates, against background `#0E1019`, card `#1A1F2E`
and text `#F0F2F7`:

| value | as ink on bg | as ink on card | `#F0F2F7` on it as fill |
|---|---|---|---|
| `#111520` (current) | 1.04 ✗ | 1.11 ✗ | **16.28 ✓** |
| `#E8311A` (the brand accent) | 4.41 ✗ | 3.82 ✗ | 3.84 ✗ |
| `#FF5A3C` (lighter tint, same hue) | **6.13 ✓** | **5.30 ✓** | 2.77 ✗ |

Note the middle row: **the site's own accent does not clear AA as an ink either**,
at 4.41 against a 4.5 floor. There is no value that satisfies both roles, so
repointing the palette would trade one failing eyebrow for failing text on every
primary button — a strictly worse position, and exactly the "findings get 'fixed'
into real regressions" failure this file already warns about. **So nothing was
changed.**

The fix that closes this door is at the generator: a layout must not spend one
palette slot on both a fill and an ink. That is adjacent to candidate 1 (stop the
chrome hardcoding white) and belongs with whoever takes the generator — flagging,
not claiming.

**The transferable item, and the reason this is worth the space here.** The
population of this bug is probably not "four sites with bad palettes" but **every
site whose `styles.css` has not been re-rendered since 2026-07-27**, for whom the
fix exists and has never run. That set is enumerable in one query per site
(compare the served `--color-card-bg` against `#ffffff` while the palette is
dark) and is not enumerated anywhere. A re-render is cheap — this one took under
three minutes end to end.

### Two more instances of the SAME pair, found 2026-07-29 17:50Z on new pages

Both are `--color-primary`-as-an-ink and one more slot doing double duty. Recorded
here rather than fixed, for the reason already established above: no value
satisfies both roles, so repointing trades one failure for another.

**Homepage went from 1 failure to 2** after a rebuild, not because anything
regressed but because the new card content exposed a second consumer:

```
1.04:1  rgb(17,21,32) on rgb(14,16,25)   .info-card-grid__eyebrow      'SPEC-FIRST GUIDES'
1.11:1  rgb(17,21,32) on rgb(26,31,46)   .info-card-grid__card-link    'Catch up on news'
```

`--color-primary` (#111520) is the ink for BOTH the eyebrow and the card link, on
the page background and on the derived card background respectively. Worth noting
because it shows the count is content-dependent: the same defect reports 1 or 2
depending on which cards a page happens to render, so **a falling number here is
not evidence of repair**.

**A new component, a new pair.** `/news/index.html` was created today and audits
at 28 failures, every one of them the same:

```
3.94:1  rgb(138,146,168) on rgb(44,52,80)   .news-list-tag   'World Matchplay'
```

That is `text_muted` (#8A92A8) on `border` (#2C3450) — a component using the
BORDER slot as a fill and the muted slot as an ink on top of it. Neither slot was
authored for that job. It is marginal (3.94 against a 4.5 floor) and on small
tags, so it is the least urgent thing in this file, but it is the same class and
it arrived on a brand-new page, which is the point: **the class is still being
reproduced by new components, not just carried by old stylesheets.**

`.news-list-tag` is worth a look by whoever takes the generator — a tag chip
wants `surface` as its fill and `text` as its ink, and both are already derived.

---

## Re-measurement 2026-08-06 — CANDIDATE 1 HAS SHIPPED, two of three findings are FIXED, and the surviving class is a different one

New lane: `docs024_key_docs_latest/bugfix_122_contrast_ink_slots/`
(`PLAN_2026-08-06_contrast_ink_slots.md` carries the full working). Everything below
was measured today against the live fleet, not carried from this file.

> **CORRECTED 2026-08-06 — "Finding 1: the shared header CTA hardcodes white text on
> the site's accent colour" is FIXED, and its five-of-six table is SPENT EVIDENCE.**
> A reader taking that table as current would spend a chassis roll re-fixing a shipped
> fix. The live shared template is var-driven:
>
> ```sql
> SELECT substring(html_template from '\.header-cta\s*\{[^}]*\}')
>   FROM content_components WHERE name='header-theme-chrome';
> -->  background: var(--color-cta-bg, var(--color-accent));
>      color:      var(--color-cta-text, var(--color-primary-text));
> ```
>
> And the stored chrome has caught up — **0 of 19** header rows fleet-wide carry a
> hardcoded white CTA ink (14 var-driven, 4 with no `header-cta`, 1 on the bespoke
> `header-leopardess`, which is the only active component still holding the literal).
> This file's candidate 1 said the chrome is a stored artefact needing the
> `nav_drift` → `nav-updater` path; that has evidently happened.

> **CORRECTED 2026-08-06 — "Finding 2: robot-hands.com's primary call-to-action is
> invisible" is FIXED.** `.cta-btn.cta-btn-primary` at 1.00:1 is gone from the served
> page. robot-hands still fails, on entirely different elements (below).

**Finding 3 (vonc.com's Gauntlet buttons) is STILL LIVE as filed** — 23 failures,
`.gauntlet-btn-primary` 1.76:1. Untouched: that surface is the Gauntlet workstream's,
as this file already says.

### The fleet today — 15 homepages, `scripts/render_audit.py`, 2026-08-06

**109 firm failures across 12 sites.** `relojistas.com` and `vetcomparison.uk`, both
listed as failing here on 07-28, are now **clean**; `fundamentallyai.com` and
`leopardessconsulting.co.uk` are clean too.

| site | firm | dominant shape |
|---|---|---|
| ai-agent-orchestration.com | 30 | `--color-heading` == its own background |
| vonc.com | 23 | Gauntlet buttons (owned elsewhere) |
| gamesdesign.co.uk | 17 | white/70% on a cyan accent fill |
| idea.uk | 14 | `text_muted` on a light surface |
| finetuning.uk | 10 | white on a mid-tone accent fill |
| gaswholesalers.com | 6 | accent as an ink on white |
| robot-hands.com | 3 | `--color-primary` as an ink |
| dartsonline.com | 1 | `--color-primary` as an ink |
| oufe.com, webdesign.co.uk, webdesign.uk | 1–2 | over-image approximations only |

### The surviving class, in three sub-shapes

**A — a palette slot spent on both a FILL and an INK.** dartsonline
`.image-hover-card-grid__eyebrow` 1.04, robot-hands `.tl-eyebrow` 1.14 /
`.tl-card-link` 1.07 — all `color: var(--color-primary)` in shared, unforked, active
component templates. **17 of 18 layouts use `color: var(--color-primary)` as an ink**
(`social-lobby` 11×, `affiliate-hub` 9×, `magazine-grid` 8×). This is the
double-duty finding the 07-29 dartsonline contribution reached, now measured as a
fleet property rather than one site's. `warnUnusablePrimary` already detects the
condition at <3.0 and only logs. **`darkSchemeDerivations` derives `primary_text` —
the ink that goes ON a primary fill — and there is no slot for the inverse: primary
made legible as an ink on the page.** The platform computes that answer twice
already (`pickInkOn`, `pickReadableOnBackground`) and never offers it in that
direction.

**B — `--color-heading` collapsing onto its own background. NEW, in no bug file,
worst on the fleet, cause `[UNMEASURED]`.** ai-agent-orchestration.com serves **six
`.H3` at 1.00:1** — `rgb(13,17,23)` on `rgb(13,17,23)` — plus an `.H2` at 1.04 and a
`.section-heading` at 1.10. It should be impossible: `darkSchemeDerivations` has
`{heading, from: text}` and the renderer alias block has
`{"--color-heading", "var(--color-text)"}`. The served value equals the *background*.
**Deliberately not diagnosed here** — going to the `090` loop rather than into a
guess, per this repo's corrected diagnosis section and the 2026-07-31 owner ruling.

**C — a component hard-coding an ink over a themed fill (026 family 3).** finetuning
`.csg-cta-btn` white on `#C8873A` = 3.01, `.cta-btn cta-btn-primary` white-on-white =
1.00; gaswholesalers `.A` `#E8A020` on white = 2.22; gamesdesign `.stats-eyebrow`
`#00E5FF` on `#0DBFD6` = 1.44 plus eight `rgba(255,255,255,0.7)` labels at 1.72.

> **`accent_text` was derived on 2026-07-27 for precisely this — its own comment says
> "so a component can stop hard-coding white over an accent fill" — and it has ZERO
> CONSUMERS.** Measured across every surface that could name it: `content_components`
> 0, `layouts` 0, `css_snippets` 0, `site_components` 0, `page_components` 0. That is
> the LANDMINE *"a palette slot no LAYOUT declares is never emitted — deriving it is
> dead config"*, recorded from this bug's own dartsonline round, sitting unfired in
> the list it was written about.

**Note the scheme boundary breaks on C.** `fillDarkSchemeSpecialisedSlots` is
dark-only by deliberate design, and gaswholesalers (`#F4F1EB`) and finetuning
(`#F5F3EF`) are LIGHT. "Is this colour legible as text on that ground" is
scheme-independent, so a dark-only derivation cannot reach sub-shape C at all.

### CANDIDATE 2 IS NO LONGER A BUILD TASK — it is one `scheduled_tasks` row

> **CORRECTED 2026-08-06 — candidate 2 above says "the tool already exists and is not
> wired to anything — that is the whole of the remaining work". The Go port, the
> orchestration AND the work-item drain now all exist and are live.** As of chassis
> **v1.0.1257**:
>
> - `write_render_audit_findings` is **in the running binary** — pod-grepped on
>   `agent-chassis-5b9fd84984-hqc5d`: **11** occurrences, invented control **0**,
>   positive controls `scanStoredStatClaims` **2** / `fillDarkSchemeSpecialisedSlots`
>   **4**. (Register VIZ-013, previously "inert until an image roll".)
> - the live `render-audit-agent` row's workflow steps are
>   `site → audit → write_findings → complete` — so the config tail landed too.
> - it files firm contrast failures as `contrast_failure` routed to `css-patch-agent`,
>   deduped on `contrast_failure:<page-path>#<selector>`.
>
> **And nothing dispatches it.** 28 enabled `scheduled_tasks`, none targeting
> `render-audit-agent`. `contrast_failure` items ever raised: **4** — all
> relojistas.com, all 2026-08-04, all `complete`, i.e. one hand-run. This is the
> `bugs_open/083` / `093` / `115` shape again: a mechanism made correct, then guarded
> behind something that never runs.
>
> *The check that stops you getting this wrong:* `orchestration_states` for
> `owner_agent_type='render-audit-agent'` returns **0 rows**, and that does NOT mean
> "never ran" — terminal rows are reaped at ~24h. `scheduled_tasks` has no reaper and
> is what answers the cadence question.

### Fix now proposed (full reasoning + the rejected alternatives in the lane's PLAN)

1. **Emit legible-ink companions from the RENDERER, not from 18 layout templates** —
   `--color-primary-ink` / `--color-accent-ink`, each holding its source colour when
   that clears AA against the ground, else the palette colour that does. Same
   renderer-owned `:root` pattern as `buildSectionDefaults` and `buildTokenAliases`,
   which already carry a stated "themes must not declare these" contract. Additive
   and inert — nothing changes until a component writes
   `var(--color-primary-ink, var(--color-primary))`, so the opt-in is structural and
   visible to a reviewer of the *component*. Scoped to **both** schemes, unlike the
   dark-only derivations, because sub-shape C is on light sites — that widening is
   the thing to push on in review, and is stated rather than buried.
   Then repoint `image-hover-card-grid` (1 page, 1 site) and `tool-list` (6 pages, 4
   sites), which closes sub-shape A's measured instances.
2. **Add the cadence row for `render-audit-agent`** — but bank a pre-fix baseline
   first, since findings dedup by page+selector and the next audit is the de-facto
   verifier.
3. **NOT repointing palettes** (proved to trade one failure for another) and **not**
   hand-fixing pages as a class fix.

**Verification must not be graded on a stylesheet or a palette row** — both are the
unsound methods this file's own 07-28 correction records. Induce the no-op case too:
a site whose `primary` already clears AA must get the slot holding `primary`'s own
value and render byte-unchanged.

---

## Contribution 2026-08-09 — the 08-06 fleet figure is a HOMEPAGE figure, and `.news-list-tag` is not the least urgent thing here, it is the largest

**Contributed from the `bugs_open/113` lane, whose 08-09 census named the three worst
`--color-primary` sites and whose stated next step was to audit them in full.** Adding
the measurement here because this file owns the class; 113 owns only the merge half.

### 1. Homepage sampling understated this bug by ~2 orders of magnitude

The 08-06 section measures **15 homepages** → 109 firm failures, with
`robot-hands.com` at **3** and `dartsonline.com` at **1**. Same tool, same probe,
`--sitemap` instead, 2026-08-09:

| site | 08-06, homepage only | 08-09, full sitemap | pages |
|---|---|---|---|
| robot-hands.com | 3 | **193** | 19 |
| dartsonline.com | 1 | **125** | 18 |
| ai-agent-orchestration.com | 30 | **124** | 24 |
| | | **442 solid, 61 pages** | |

**Nothing regressed and nothing was mismeasured — the 08-06 run simply never opened
these pages.** The failures concentrate on tool, guide and news pages. `dartsonline`
going 1 → 125 is the sharpest case: on its homepage this bug is a rounding error, and
across its site it is 125 failures.

**41 further raw failures were discounted, not counted**, all `overImage` mid-grey
approximations under the two gradient CTAs (`render_audit.py:111-114`). The 08-06
table's "over-image approximations only" rows are the same artefact.

*The check:* **`--sitemap` is not a nicety on this bug, it is the measurement.** A
per-site number taken from `index.html` is a claim about one page.

### 2. `.news-list-tag` — the 07-29 rating needs inverting

That section reads: *"the least urgent thing in this file"*, marginal at 3.94:1, 28
instances on one new page. The mechanism it names is exactly right —
`.news-list-tag { color: var(--color-text-muted); background: var(--color-border) }`
in the active, unforked `news-listing` component, `border` spent as a fill with `muted`
as its ink. Only the **size** was wrong:

| site | instances | ratio |
|---|---|---|
| robot-hands.com | **128** | 3.47:1 (`rgb(122,143,166)` on `rgb(45,58,74)`) |
| dartsonline.com | **53** | 3.94:1 (`rgb(138,146,168)` on `rgb(44,52,80)`) |
| | **181 of 442 = 41%** | |

**It is the single largest contributor to this bug across the three sites measured**,
larger than sub-shapes A and B combined on those sites. It stayed "least urgent"
because it was only ever seen on the one page that happened to be new that day.

It also does not fit A, B or C cleanly: A is `primary` double-duty, C is a component
hard-coding a *literal* ink over a themed fill. This is **two themed slots paired with
each other**, neither authored for the pairing and neither reviewable as a text pair —
`border` is checked as a line colour, `text_muted` as an ink on `surface`. Whether that
is a fourth sub-shape or a widening of A is the owning lane's call.

The 07-29 note already names the repair — *"a tag chip wants `surface` as its fill and
`text` as its ink, and both are already derived"* — and it is **one component template**,
not 18 layouts. Against the proposed `--color-primary-ink` work it is far cheaper and
retires 41% of the measured failures, so it is worth doing **first and separately**,
whatever happens to the renderer proposal. Not done from here: `news-listing` is a
shared fleet component and this is not my lane.

### 3. Sub-shape B, on the site that has it worst

The 08-06 section records `--color-heading` collapsing onto its own background on
`ai-agent-orchestration.com` (six `.H3` at 1.00:1 on the homepage), cause `[UNMEASURED]`,
deliberately routed to `090` rather than guessed. **Unchanged and not diagnosed here.**
Across its full sitemap it is **25 `.H3` at exactly 1:1** plus 17 `.H2`, and **73 of its
124 failures are ≤1.1:1**, i.e. invisible rather than faint. Two facts that may help
whoever files the `090`, both measured, neither a diagnosis:

- its served `--color-primary`, `--color-surface` and `--color-heading` all resolve to
  **`#0D1117`**, while `--color-background` is `#080B10`;
- **there is no `palettes` row for the domain at all** (`source_domain` → 0 rows), and its
  `site_specs.design_intent.color_scheme` is a **light** scheme (`background #ffffff`)
  that plainly did not render, on a site whose `avoid` list forbids white backgrounds.

Where its served palette comes from is the open question, and it is the same question
sub-shape B is stuck on.

### The `.news-list-tag` fix, measured on all 8 consumers and PREPARED but NOT APPLIED

Written as `docs/agent_docs/sql_for_agents/353_news_list_tag_ink_fix.sql` (+ `_ROLLBACK`).
**Not applied: the production write was refused by this session's permission
classifier.** Everything below is measured; only the execution is outstanding.

**The 07-29 note's prescription was half right, and the measurement says which half.**
It proposed *"a tag chip wants `surface` as its fill and `text` as its ink, and both are
already derived."* The **ink** half is correct. The **fill** half is wrong, and the same
render audit that found the bug is what refutes it: `surface` against the section's own
`background` is **1.04–1.22 on all eight sites**, so a surface-filled chip stops reading
as a chip. Keeping `border` as the fill preserves the pill's existing distinctness
(1.11–1.62, unchanged) and makes it a one-declaration change to a shared fleet component
instead of two.

Both candidates clear AA on 8/8; the ink-only change is the smaller one, so that is what
353 does — `color: var(--color-text-muted, #64748b)` → `color: var(--color-text, #475569)`.

| site | today | after | scheme |
|---|---|---|---|
| idea.uk | **2.25** | 9.67 | light |
| robot-hands.com | **3.47** | 9.38 | dark |
| relojistas.com | **3.84** | 12.55 | light |
| dartsonline.com | **3.94** | 10.94 | dark |
| webdesign.co.uk | **4.13** | 11.99 | light |
| fundamentallyai.com | **4.30** | 11.47 | dark |
| gaswholesalers.com | **4.32** | 11.01 | light |
| ai-agent-orchestration.com | 4.95 | 12.88 | dark |

**7 of 8 fail today, not 2** — the two I audited by sitemap are simply where the volume
is. **8 of 8 pass after, minimum 9.38.** The no-theme fallback improves as well:
`#64748b` on `#e2e8f0` = 3.86 (fails) → `#475569` on `#e2e8f0` = 6.15. Light and dark
sites both pass, so this is scheme-independent — unlike `fillDarkSchemeSpecialisedSlots`,
which by design cannot reach sub-shape C at all.

**Scope, checked rather than assumed:** the template uses `--color-text-muted` five
times; 353 changes **one**. The other four are muted text on the section `background` —
the slot's designed pairing — and **none of them appears anywhere in the render audit's
failure list**, which is the evidence for leaving them alone. The two-line anchor
(colour line + the `background:` line under it) occurs exactly once in the template;
that was verified in the DB before the file was written, and 353 re-checks it in a
`DO`/`RAISE` guard that aborts if another session has edited the template since.

**The repair path, which is NOT what `bugs_closed/072` would lead you to expect.** The
rule ships **twice**: inline in every page's `<style>` (from this template) and again in
`styles.css` (frozen). Measured on `robot-hands.com/news/index.html`, the inline copy is
at byte **38425** and the `<link>` to `styles.css` at byte **8412** — the inline copy is
later, so at equal specificity **it wins**. Therefore **a page re-render repairs the page
and no stylesheet rebuild is needed**, which is the cheap path and the opposite of 113's
situation. Six of the eight sites also carry the stale rule in `styles.css`
(`fundamentallyai.com` and `ai-agent-orchestration.com` do not); it is overridden and
harmless, but it will keep reading as the old value to anyone grepping the stylesheet.

**Owed after it is applied:** re-render one news page (robot-hands `/news/` has 105 chips
on it) and re-run `render_audit.py --sitemap` on robot-hands and dartsonline. Expect
−128 and −53. **Run the audit BEFORE as well as after** — 113's own transferable lesson,
and the reason the mirror defect on the cost calculator was caught at all.

---

## 2026-08-10 — pointer: the queued stylesheet re-render will NOT repair `ai-agent-orchestration.com`'s white cards

Short note, not a fork — the finding and its evidence live in
**`bugs_open/113`, section "2026-08-09 (third pass)"**.

Work item **`e97fb5c5`** (`needs_design_review`, triaged 2026-08-09, *"re-render
styles.css so the VIZ-014 legible-ink slots reach the served stylesheet"*) will deliver
the ink slots as intended. **It will not move that site's 44 `#ffffff`-card failures**,
and those are a separate mechanism from this file's `.news-list-tag` chips.

Why, in one line: the site renders from the **shared seed palette `professional-dark`**,
which defines `card_bg: "#ffffff"` explicitly. `card_bg` is a specialised slot (the theme
wins it, a site's spec cannot override it) and `fillDarkSchemeSpecialisedSlots` skips any
slot the palette already defines — so the value is **curated, not fallen-through**, and is
unchanged by re-rendering. Deterministic; it does not depend on what the run's LLM emits.

**What this means for the batch:** nothing needs to change about `e97fb5c5` — just don't
read a persisting white card afterwards as the re-render having failed. If the white cards
are also wanted gone, that is an **input** change (fork the palette for this site, or move
it to a genuinely dark collection); editing the shared seed row would break
`finetuning.uk` and `gaswholesalers.com`, which ride the same palette and are light sites
where `#ffffff` is correct.

*Also relevant to this file's own before/after discipline:* when the audit is re-run here,
`ai-agent-orchestration.com`'s total will drop by the chip count and **not** by 44. That is
the expected result, not a shortfall.
