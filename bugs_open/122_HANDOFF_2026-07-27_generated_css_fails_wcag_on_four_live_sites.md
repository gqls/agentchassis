# 122 — generated stylesheets fail WCAG AA on four live sites, and nothing checks

**Filed** 2026-07-27 from the oufe.com workstream, after the owner reported a link
on oufe.com as "dark blue on the black background and not easily readable".
**Severity** high — user-visible, on live public sites, and an accessibility
failure rather than a cosmetic one. On three of the four, links are effectively
invisible.
**Status** ~~OPEN. oufe.com fixed and browser-verified; the generator unchanged.~~
**FIXED IN SUBSTANCE 2026-08-10 — engine live, config applied, propagation verified,
closures measured. Stays in `bugs_open/` per the owner's 2026-08-06 ruling. Read the
status block immediately below before anything else in this file.**

> ## STATUS 2026-08-10 — what shipped, what it closed, and what it provably cannot
>
> **The generator is no longer unchanged: it now answers the question this bug is
> about.** `buildLegibleInkDefaults` (VIZ-014) emits `--color-<x>-ink` — the palette
> colour *made legible as an ink on the page* — checked against **every** ground the
> text may sit on. Live since chassis v1.0.1266, pod-proven on both replicas.
> Components and layouts opted in via migrations **`338`** (4 components, 5 layouts) and
> **`368`** (`info-card-grid`). 11 sites' stylesheets and 13 pages re-rendered and
> verified at the served artefact.
>
> **Measured result, graded per selector against `BASELINE_2026-08-06_render_audit.txt`
> (banked as `AFTER_2026-08-10_render_audit.txt`): all 10 predicted closures delivered.**
> gaswholesalers 6, robot-hands 2, finetuning 2, dartsonline 1 — and **dartsonline's
> homepage now measures `contrast=0`**, the first fully clean page in this workstream.
>
> **⚠ Do NOT grade this bug by the fleet total.** It rose 109 → 112 while every targeted
> failure closed: other lanes ship to these same sites continuously and a page
> re-rendered for any reason carries every change since it last rendered. Grade per
> selector at the named ratio, against a banked before-state, or you will conclude the
> opposite of the truth.
>
> **The defect class RECURS, and did so inside the fix window.** Two days after 338
> closed dartsonline's `image-hover-card-grid` failure, that lane swapped in
> `info-card-grid` and reintroduced the identical fault six times over. Found by a hand
> re-audit; fixed by 368 the same afternoon. **This is why the standing control matters
> more than the repair:** migration **`369`** now runs the render audit **weekly per
> site** automatically (VIZ-015), filing firm failures as `contrast_failure` work items.
> It fired within 70s of apply and found **171 firm failures on robot-hands' INTERIOR
> pages** — which the 15-page homepage baseline had never looked at — filing 34 of them
> (111 dropped by a `max_items` cap that reports itself; 6 of the site's 31 pages never
> swept at all by a 25-page cap that does **not** — `bugs_open/242`).
>
> **UPDATE 2026-08-11: the rotation has now swept the whole fleet — 19 sites, one per
> hour, exactly as designed — and filed 220 `contrast_failure` items. All 220 sit in
> `detected`.** They are not blocked by type: `triage_detected_items` promotes every
> detected item for a site with no type filter
> (`triage_detect_items_action.go:162-173`, `WHERE site_id = $1 AND status = 'detected'`).
> They are waiting on **`improvement-loop` running for their site at all** — and its
> scheduled driver `improvement-sweep` has been `enabled = false` since **2026-05-02**.
> 776 items across 20+ types are queued behind the same gate, oldest 2026-07-24.
> **That is an owner decision, not a bug fix** — see the handoff's §7.
>
> **What this fix provably CANNOT reach, and the proof:** the ink companions are computed
> against `pageGrounds = {background, surface}` only. An element sitting on a
> **component-painted** ground is out of reach — `legibleInkFor` correctly returns the
> source unchanged while the element stays illegible. Not inferred: gamesdesign's
> `.stats-eyebrow` measured **1.44:1 after the fix shipped, byte-identical to its
> pre-fix failure**; vonc's the same at 1.63:1 after two re-renders. That is ~24 further
> failures and it is **`bugs_open/212` §8's open architecture question** — an owner
> decision, not a bug patch. Repointing components one at a time (338, 368) works and is
> the sanctioned route meanwhile.
>
> **One live dependency:** the 34 findings above route to `css-patch-agent`, where
> **`bugs_open/213`**'s false-complete defect is unfixed. Detection is now strong; the
> repair half is known-defective. **Grade repairs at the next audit, never at the item
> status.**
>
> Full account: `docs/agent_docs/docs024_key_docs_latest/bugfix_122_contrast_ink_slots/`
> — `HANDOFF_2026-08-10_continue_here.md` (state), `SUMMARY_2026-08-10_contrast_ink_slots.md`
> (plain prose), `NOTES_contrast_ink_slots.md` (evidence + 15 recorded missteps).

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

### 2026-08-10 — 353 APPLIED and verified in the template; plus two corrections to the section above

**Applied 2026-08-10 11:41:22 UTC** by the owner. The guard passed and all three
post-conditions asserted: chip rule written once, **4** remaining `--color-text-muted`
uses, **4456** bytes. Read back from `content_components`, the stored rule is now
`color: var(--color-text, #475569); background: var(--color-border, #e2e8f0);`.

**No artefact carries it yet, and that is expected.** Every one of the 9 placements had
last rendered *before* 11:41, so at 11:43 `robot-hands.com/news/` still served the old
declaration. Propagation is per page re-render (§ the cascade note above — no stylesheet
rebuild is needed).

**Measured re-render cadence, which decides whether this self-heals:** 7 of the 9
placements re-rendered on 2026-08-10 between **02:24 and 08:45**. So those pages pick the
fix up on their own within about a day. **Two will not:** `webdesign.co.uk` (last
2026-08-08) and `idea.uk` (last **2026-07-14**, four weeks).

**832 chips are rendered fleet-wide** across the 8 pages that have any — robot-hands 105
(×2 pages), gaswholesalers 120, relojistas 109, ai-agent-orchestration 107, dartsonline
100, fundamentallyai 95, webdesign.co.uk 91, **idea.uk 0**.

> **CORRECTION 1 — "7 of 8 sites fail today" mixed an observation with an arithmetic
> prediction, and one of them is unobservable.** Only `robot-hands` and `dartsonline` were
> render-measured. The other six numbers were computed from their served palettes: they say
> *"a chip on this site would fail"*, which is the right basis for a fleet fix but is not a
> count of anything seen. **`idea.uk`, which I called the worst at 2.25:1, renders no chips
> at all** — its news page serves the `{{else}}` placeholder. Its true contribution to this
> bug today is **zero**. The fix is still right for it (the moment a chip renders there it
> would have been unreadable) but it repairs nothing visible on that site now.
> *The shape, again:* this file's own 2026-07-28 correction says a stylesheet cannot answer
> a rendered-contrast question. I used palette values to extend a render measurement to six
> unrendered sites and wrote the two kinds of number in one column.

> **CORRECTION 2 — "181 instances" is 181 DISTINCT failing labels, not 181 chips.**
> `render_audit.py:141` dedupes on `class|colour|first-40-chars-of-text`, so a tag text
> repeated across articles is counted once per page. The rendered chip counts on the two
> measured sites are **210** (robot-hands, two pages) and **100** (dartsonline) against
> **128** and **53** distinct failures. The **41% share is unaffected** — all 442 failures
> are deduped the same way — but "instances" overstated what a visitor sees per page and
> understated the total.

**Still owed:** re-run `render_audit.py --sitemap` on robot-hands and dartsonline once
their pages have re-rendered, and record the paired before/after. Before = robot-hands
**193**, dartsonline **125** (2026-08-09, full sitemap). Expect −128 and −53.

### 2026-08-10 (evening) — 353 PROPAGATED and graded per selector: robot-hands 193 → 43 solid, chips 128 → 0

**From the `bugs_open/113` contrast front, contributed into this file** — the ink-slots lane
above owns 122 now (`bugfix_122_contrast_ink_slots/`); this is the paired before/after for
the `news-list-tag` change (`sql_for_agents/353`), which that lane's account does not cover.

**Propagation verified with a discriminating control before measuring anything.** At 15:12
`robot-hands.com/news/` re-rendered and its served chip rule became
`color: var(--color-text, #475569)` with **0** `--color-text-muted` inside the rule;
`dartsonline.com/news/`, which had not re-rendered, still served the old ink. Same page
type, opposite results — so the change reaches artefacts by page re-render, as predicted.

**Full sitemap, same tool, same 19 pages, graded per selector** (this file's own rule —
never grade by the fleet total):

| | pages | raw | over-image | **solid** | `.news-list-tag` |
|---|---|---|---|---|---|
| before (08-09) | 19 | 217 | 24 | **193** | **128** |
| after (08-10) | 19 | 67 | 24 | **43** | **0** |
| delta | 0 | −150 | **0** | **−150** | **−128** |

**`over-image` is identical at 24 — that is the internal control.** Those rows are the
probe's own mid-grey placeholder and no CSS change can move them; had they shifted, the two
runs would not have been comparable.

**−150, not −128, and the extra 22 are NOT mine.** Attribution by selector:

| selector | before | after | whose |
|---|---|---|---|
| `news-list-tag` | 128 | 0 | **353, this front** |
| `faq-item__answer` | 20 | 0 | another lane |
| `tl-eyebrow` / `tl-card-link` | 1 / 1 | 0 / 0 | the ink-slots lane (338 — both are named in its own 08-06 sub-shape A list) |

That is the *"a page re-rendered for any reason carries every change since it last
rendered"* warning in the status block at the top of this file, observed from the other
side: **a re-render triggered for my change delivered three lanes' work at once.** Anyone
sizing a fix by a whole-site total will over-credit themselves; the per-selector grade is
the only honest one.

**Corrections to my own 08-10 entry above:**
- **`webdesign.co.uk` self-healed** at 15:33 — I listed it as a straggler that would not.
  **`idea.uk` (2026-07-14) is now the only placement not carrying the fix**, and it renders
  no chips anyway (see the entry above, and `bugs_closed/027`).
- Remaining on robot-hands: **43 solid**, dominated by `LEGEND` 8, `cta-btn` 8, `H3` 7,
  `H2` 5 — i.e. sub-shape A/C, untouched by 353 and unchanged by it.

**Still owed:** the same paired grade for `dartsonline.com` once its news page re-renders
(before = **125** solid, **53** chips).

### Fleet state at chassis v1.0.1280, pod-grepped 2026-08-10

The audit engine and its cadence are both live — recording the controls because the status
block's claims are now independently confirmed from outside that lane:

| binary | symbol | count |
|---|---|---|
| `browser-runner-adapter` (`render-audit-adapter` pod) | `render_audit` / `RenderAudit` / `overImage` | 8 / 4 / 6 |
| `agent-chassis` | `write_render_audit_findings` / `contrast_failure` / `fillDarkSchemeSpecialisedSlots` | 11 / 2 / 4 |
| both | `ZZZ_invented_control_symbol` | **0** |

`site-render-audit-rotation` is **enabled, hourly**, last fired 15:54:33. **This retires
the "nothing dispatches it" finding in the 08-06 section** — that gap was closed by
migration 369.

**The queue is filling and not draining.** `contrast_failure` items went **34 → 68 within
this session**, every one `detected`, every one routed to `css-patch-agent`, against **4**
`complete` (all the 08-04 hand-run). Consistent with `bugs_open/213`'s false-complete
defect being unfixed. **Detection is now strong and the repair half is the constraint** —
which is what the status block says, measured here from the item table rather than asserted.

### 2026-08-11 — dartsonline graded: chips 53 → 0, no regression on any existing page, and 3 NEW pages arrived already failing

Completes the pair owed above. `dartsonline.com/news/` re-rendered **2026-08-11 02:25:30**
and serves `color: var(--color-text, #475569)` with **0** `--color-text-muted` inside the
chip rule (100 chips still on the page, so this is a real repair, not an empty page).

**All 9 `news-listing` placements now carry 353 except `idea.uk`** (still on its 2026-07-14
render, and it renders no chips at all — `bugs_closed/027`).

**The raw totals are NOT comparable and quoting them would be wrong: the sitemap grew 18 →
21 pages between the two runs.** Restricted to the **18 pages present in both**:

| | solid |
|---|---|
| before (08-09) | **125** |
| after (08-11) | **60** |
| delta | **−65** |

**Every selector that moved on those 18 pages moved DOWN — there is no regression anywhere:**

| selector | delta | whose |
|---|---|---|
| `news-list-tag` | **−53** | 353, this front |
| `info-card-grid__card-link` | −6 | this lane's migration 368 |
| `LABEL` | −3 | other |
| `info-card-grid__eyebrow` / `H3` / `db-submit` | −1 each | 368 / other |

**The apparent `A` +8, plus new `LEGEND` and `btn-compare` failures, are entirely on three
pages that did not exist on 08-09** — `/tools/dart-weight-comparator/`,
`/blog/grip-styles.html`, `/guides/tool-dart-weight-comparator-guide.html`, carrying
**3 + 4 + 4 = 11 solid failures between them.**

> **That is this file's "the defect class RECURS" claim, measured on a second site and a
> second week.** Three pages built in the last two days shipped with 11 AA failures on a
> site that was being actively repaired the whole time. It is the argument for the weekly
> cadence (369) being the real deliverable rather than any individual repair — **and note
> the recurrence is on NEW pages, which no amount of re-rendering old ones would surface.**

*Method note, because it nearly produced a false result:* the first pass of this comparison
reported `A` **50 → 58** and read as a regression on a site I had just helped fix. It is an
artefact of comparing two different page sets. **A sitemap is not a fixed population —
intersect the URL sets before differencing, or a growing site will manufacture both
regressions and improvements.** Same family as this file's own "grade per selector, never
by the fleet total", one level down: grade per selector *on the same pages*.

---

## Contribution 2026-08-12 — a fourth consumer of `--color-primary`-as-ink, named: the shared `article-body` component's in-prose links, 91 instances / 17 sites

Not from the `bugfix_122_contrast_ink_slots` lane — reported by the owner from a live
screenshot of `dartsonline.com/guides/tool-dart-weight-comparator-guide.html` (in-prose
links reading as invisible dark-on-dark). Contributing the diagnosis here rather than
opening a parallel account, and **not touching the shared component** — that lane owns
the renderer-level fix and is mid-deploy of the retraction mechanism (`5639a1103`,
council-approved, not yet live per its own 2026-08-12c handoff).

**Confirmed live, `scripts/render_audit.py`, 2026-08-12:**

```
FAIL dartsonline.com/guides/tool-dart-weight-comparator-guide.html contrast=5
  1.11:1 need 4.5  rgb(26,31,46) on rgb(17,21,32)  .A  'Dart Weight Comparison Tool'
  1.11:1 need 4.5  rgb(26,31,46) on rgb(17,21,32)  .A  'tungsten darts guide'
  1.11:1 need 4.5  rgb(26,31,46) on rgb(17,21,32)  .A  'barrel weight guide'
  1.11:1 need 4.5  rgb(26,31,46) on rgb(17,21,32)  .A  'Dart Setup Builder'
```

`rgb(26,31,46)` is dartsonline's `--color-primary` (`#1A1F2E`) — identical to its own
`--color-surface`, so as an ink on `--color-background` (`#111520`) it is nearly invisible.
This is a live-and-already-parked instance, not a new one: `site_work_items` already carries
`d7044a72-b209-4d8b-9410-1032e4b7d52b`,
`contrast_failure:/guides/tool-dart-weight-comparator-guide.html#A.A`, status `deferred` —
i.e. this page is already inside the 226-row park the lane's 12c handoff describes, no new
item needed.

**The named cause is new: it is not `.info-card-grid__eyebrow`, `.tl-eyebrow`/`.tl-card-link`
or `.news-list-tag`** (the three consumers this file has measured so far). It is the shared
`article-body` component (`content_components.id = 5835b2e1-50d7-4f20-8a9c-8da4d270ae3d`),
whose inline `<style>` block sets in-prose links unforked:

```css
.article-body-section .article-body__content a { color: var(--color-primary,#1e40af); text-decoration: underline; }
```

— the base site-wide rule (`a { color: var(--color-text); }`, high contrast on every site
checked) is overridden inside every article body, on purpose, and this is the same
double-duty mechanism sub-shape A already names: one palette slot asked to be both a fill
(buttons, borders) and an ink (this rule), and no slot exists for "primary made legible as
an ink" — this file's own `PLAN`'s `--color-primary-ink` proposal is exactly the fix that
would close this consumer too, once it ships and `article-body` is repointed alongside
`tool-list` and `image-hover-card-grid`.

**Blast radius, and the honest limit of it: template reach is measured, failure reach is
not.** The component renders **91 times across 17 sites** (`page_components` join, counted
2026-08-12) — every one of them a blog or guide article, which is the single most common
page type this platform generates, so this is plausibly the largest unmeasured consumer of
the class. But **spot-checking two more sites that use the same component did NOT reproduce
this failure**:

```
ok    ai-agent-orchestration.com/blog/why-most-ai-agent-frameworks-fail-at-the-orchestration-layer.html   contrast=0
FAIL  finetuning.uk/blog/what-can-ai-actually-do-for-a-small-business.html   contrast=3   (all 3 are CTA buttons, none `.article-body__content a`)
```

So this is **not** "every site with this component is broken" — it reproduces only where a
site's own `--color-primary` already sits close to its background/surface, i.e. sites that
would already trip `warnUnusablePrimary` (the existing log-only check this file already
names as detecting-but-not-acting). **Which of the 17 sites that is remains
`[UNMEASURED]`** — a real render pass per site is needed before quoting a count, per this
file's own repeated correction about mixing palette inference with rendered measurement.

**Nothing applied.** No template edit, no DB write — flagging for whoever next repoints
consumers onto `--color-primary-ink`, the same way `tool-list` and `image-hover-card-grid`
were repointed by migrations 338/368.

---

## Contribution 2026-08-13 — `--color-primary-ink` IS `--color-text`. The ink companion cannot preserve brand character, and 16 of 16 sites prove it

Not from the `bugfix_122_contrast_ink_slots` lane. I picked up the 08-12 contribution above — the
`article-body` consumer it named and explicitly did not fix — intending to repoint it and then the
other 167 components carrying the same shape. **I did not apply the repoint, because measuring the
thing it repoints onto refuted the premise.** Recording that here rather than opening a parallel
account. No template edit, no DB write, no migration.

### 1. The measurement, and it could have come out otherwise

`curl` of all 22 live domains' served stylesheets, 2026-08-13. 18 are palette-driven (three
calculator sites carry no `--color-*` vocabulary at all — `loancalculator.co.uk`,
`loanandmortgagecalculator.co.uk`, `noted.co.uk`; `webdesign.uk` 302s). For every site where an ink
companion **differs** from its source slot, I compared it against that site's own `--color-text`:

| site | `--color-text` | primary → primary-ink | accent → accent-ink |
|---|---|---|---|
| dartsonline.com | `#F0F2F7` | `#1A1F2E` → `#F0F2F7` ✓text | `#E8311A` → `#F0F2F7` ✓text |
| robot-hands.com | `#E2E8F0` | `#1A1F2E` → `#E2E8F0` ✓text | `#E8500A` → `#E2E8F0` ✓text |
| loancash.co.uk | `#1a1a1a` | `#e8f5ee` → `#1a1a1a` ✓text | `#b7791f` → `#1a1a1a` ✓text |
| mortgagecalculator.co.uk | `#334155` | `#b59230` → `#334155` ✓text | `#b59230` → `#334155` ✓text |
| ai-agent-orchestration.com | `#E6EDF3` | `#0D1117` → `#E6EDF3` ✓text | *(same)* |
| oufe.com | `#E8E2D9` | `#1B2A3B` → `#E8E2D9` ✓text | *(same)* |
| vonc.com | `#f0eeff` | `#7c3cff` → `#f0eeff` ✓text | *(same)* |
| cookly.uk | `#2C2C27` | *(same)* | `#C8502A` → `#2C2C27` ✓text |
| finetuning.uk | `#1A1A2E` | *(same)* | `#C8873A` → `#1A1A2E` ✓text |
| gaswholesalers.com | `#1C1C1C` | *(same)* | `#E8A020` → `#1C1C1C` ✓text |
| lendzy.co.uk | `#1A1A1A` | *(same)* | `#E8700A` → `#1A1A1A` ✓text |
| relojistas.com | `#1a1a1a` | *(same)* | `#b8952a` → `#1a1a1a` ✓text |
| vetcomparison.uk | `#0f172a` | *(same)* | `#10b981` → `#0f172a` ✓text |
| webdesign.co.uk | `#2b2b2b` | *(same)* | `#d4a373` → `#2b2b2b` ✓text |

`[MEASURED 2026-08-13, at the served artefact]` — **16 divergences, 16 of them equal that site's
own `--color-text`. Zero exceptions.** No site in the fleet receives `accent`, `text_muted`,
`secondary`, or an achromatic fallback. The remaining 4 palette-driven sites
(`fundamentallyai.com`, `gamesdesign.co.uk`, `idea.uk`, `leopardessconsulting.co.uk`) have
ink == fill on both slots, so they contribute no divergence either way.

**This could have come out otherwise, which is what makes it evidence.** Any single site returning a
third colour — a tinted primary, `text_muted`, `#ffffff` — would have falsified it. Fourteen sites
with fourteen different palettes, two slots each, and the answer is the same value every time.

### 2. The cause is one line, and on every site in the fleet it makes `text` win

`platform/orchestration/actions/palette_specialised_slots.go:350`:

```go
	for _, key := range []string{"text", "accent", "text_muted", "secondary", "primary"} {
```

`legibleInkFor` returns the **first** palette colour that clears 4.5:1 against every ground, and
`text` is first. `text` is by construction the palette slot chosen to be read on `background` — so
it clears the grounds whenever anything does. **It wins on every site in the fleet**, and §1's
measurement is what that looks like from outside.

> **CORRECTED 2026-08-13, within hours, by a reviewer from the `bugfix_122_contrast_ink_slots` lane —
> and the correction is worth more than the sentence it fixes.** I first wrote that `accent`,
> `text_muted`, `secondary` and `primary` are "unreachable in production". **That is a measured fleet
> fact, not a logical necessity.** The walk *would* return `accent` on a site where `text` failed one
> ground and `accent` cleared them all — `text` is chosen to be legible on `background`, `grounds` is
> `{background, surface}`, and the two can come apart. **No such site exists in the fleet.**
>
> Why it matters that I got this wrong: stating it as a necessity converts a **disconfirmable
> measurement into an unfalsifiable claim** — and §1 exists precisely because the measurement could
> have come out otherwise. The mechanism explains the result; it does not license a stronger version
> of it. The part that IS necessary is narrower and sufficient for everything below: the walk can
> never return a colour related to `<x>`, because `<x>` is last in the list and is the one colour
> already known to have failed.

### 3. So the mechanism's stated purpose is not what it does

`PLAN_2026-08-06_contrast_ink_slots.md:189-195` describes candidate 1 as:

> "Value = the colour itself when it clears AA against the background, else **the palette colour
> that does (the existing `pickInkOn` walk, which prefers a palette colour so the site keeps its
> character)**."

and `palette_specialised_slots.go:401` names the slot:

> `--color-<x>-ink    <x> ITSELF, made legible as an ink on the page.`

> **CORRECTED 2026-08-13: `--color-<x>-ink` is not `<x>` made legible. It is `--color-text` under
> another name, on every site in the fleet.** The "keeps its character" clause is false as built —
> not aspirational, not partly true: the walk cannot return a colour related to `<x>` at all. The
> half of the description that IS true is the first clause: where the source already clears AA it is
> returned unchanged, which is why the opt-in is a genuine no-op on the 4 healthy sites.

Nothing about the four shipped repoints (migrations 338, 368) is wrong — those elements were
**invisible** at 1.06–1.14:1 and now read at 13–16:1, which is strictly better and was measured at
the artefact. What is wrong is the belief that repointing a consumer *preserves the brand colour*.
It replaces it with body text.

### 4. Why this stopped a fleet-wide repoint that was otherwise ready

The corpus, `[MEASURED 2026-08-13]` by regex over `content_components.html_template` with a
property-boundary anchor (`(^|[;{\s])color:`, so `background-color:`/`border-color:` cannot match):

> **Two notes on the anchor, both `[MEASURED 2026-08-13]` and both corrections to what I first
> wrote here.** (a) **The anchor loses nothing.** `[;{ ]` and a whitespace-inclusive
> `(^|[;{[:space:]])` return the **identical** 168 components / 453 declarations, so the missing
> `\n` in the character class — which I expected to cause an undercount — does not. (b) **The
> 11-component gap between anchored (168) and unanchored (179) is compound colour properties
> GENERALLY, not `background-color`.** I first wrote "the gap is `background-color`", inheriting it
> from a subagent; only **5** components carry `background-color: var(--color-primary|accent…)`. All
> 11 gap rows score 0 on the anchored pattern and ≥1 on `[a-z-]color:`, i.e. `border-color`,
> `text-decoration-color`, `-webkit-text-fill-color` and friends. The anchor is still doing exactly
> its job; the *reason* I gave for it was wrong.

- **168 components (135 active)** carry `color: var(--color-primary|accent…)` — **423 in-block
  occurrences** (259 primary / 164 accent).
- Of those, **~347 sit in blocks that paint no background of their own**, which was to be the
  eligibility rule.
- Only **4** components have ever been repointed, by hand, in the 6 days since the mechanism
  shipped. The 5th was found by the owner's eye on a screenshot. That ratio is the argument for
  automating it.

Had it been automated on the rule above, it would have written the equivalent of
`color: var(--color-text)` onto the rendered placements of those components on the 14 sites whose
served ink diverges. On light sites that turns every brand-accent link near-black — on the lane
whose product is web design among them.

> **`[MEASURED 2026-08-14 by me, after the token was restored — and the inherited figure was an
> UNDERCOUNT]`** The real blast radius is **330 rendered placements across the 14 diverging sites**,
> not "~224 across 13". Per site: `webdesign.co.uk` **50** (inherited said 49), `finetuning.uk` 48,
> `ai-agent-orchestration.com` 44, `robot-hands.com` 29, `gaswholesalers.com` 28, `relojistas.com`
> 23, `loancash.co.uk` 20, `dartsonline.com` 16, `vetcomparison.uk` 11, `oufe.com` 10,
> `lendzy.co.uk` 8, `mortgagecalculator.co.uk` 7, `cookly.uk` 5, `vonc.com` 31. The fleet-wide
> figure is **461 of 1,485 `page_components` rows across 20 sites** — the 20 reconciles exactly with
> the per-site sum, which is the arithmetic that makes both trustworthy. The divergence count is
> **14** by §1's table, and the inherited "13" was simply wrong. Query in §9(b).

**And `scripts/render_audit.py` would have scored the result a clean pass**, because near-black text
on a light ground has excellent contrast. This is the shape this file keeps recording: the wrong
result and the right result are indistinguishable at the instrument.

### 5. A second, independent defect — in the eligibility rule, not the derivation

The rule "skip any declaration block that sets its own background" is wrong, and the ground truth
refutes it in one query. There are exactly **6 hand-repointed rules** in the corpus. Five use
`--color-primary-ink`. The only `--color-accent-ink` consumer is:

```css
/* content_components: system-stats */
.system-stats-section .stats-eyebrow {
  ... color: var(--color-accent-ink, var(--color-accent, #7dd3fc));
  border: 1px solid var(--section-border);
  background: var(--section-surface);          /* <- self-painted */
}
```

**That block sets a background, so the rule skips it** — i.e. the rule refuses the one repoint a
human made from the owner's own evidence. 1 of 6 ground-truth rules, and 1 of 1 on accent.

The reason is that `--section-surface` is `rgba(255,255,255,0.05)` (`color_util.go:220`) — a
translucent overlay. The element *is* on the page ground. **"Sets a background" is not "the ink
lands on a different ground"** — and the `system-stats` row above is a first-hand, sufficient
counter-example on its own: one ground-truth rule that the eligibility rule provably refuses.

> **`[MEASURED 2026-08-14 by me, after the token was restored — supersedes the inherited figure]`**
> **41 of the 76 self-painted blocks (54%) paint a translucent, `transparent` or `--section-*`
> background.** The inherited figure was "38 of 75 (51%)" — directionally right, wrong in detail,
> and now replaced. Query in §9(a). The conclusion never depended on it: it rests on the
> `system-stats` counter-example, which I read from the live corpus myself. **More than half of what
> the eligibility rule excludes, it excludes wrongly.**

Prior art that already solved this properly and must be reused rather than re-derived:
`fix_forced_text_colours_action.go:164-188` carries a **calibrated four-way** `paintClass`
classifier (`paintAmbient` / `paintPair` / `paintInk` / `paintPaletteBand`) over the same corpus,
already council-reviewed, and it deliberately classifies from the CSS over the **whole template**
rather than per block. `bugs_open/122:205-206` already names that file as "worth reading before
writing a new one". It was right.

### 6. Population, stated honestly: this defect lives on FIVE surfaces, not one

A `content_components`-only transform would be canonical for about one fifth of the corpus, and
saying otherwise is the claim that would not survive review:

| surface | rows carrying the pattern |
|---|---|
| `content_components.html_template` | 168 components / 423 in-block **+ 30 in inline `style="…"` attributes** (16 tool components) — the 30 have no enclosing block, so any block-walking transform is blind to them |
| `layouts.css_template` | **17 of 18** — and these ship into every site's `styles.css` |
| `css_snippets.css_content` | 2 of 21 |
| `site_components.rendered_html` (stored chrome) | **33 of 66 rows, across 19 sites** — no page re-render rebuilds chrome (`bugs_open/117`). Both figures now mine `[MEASURED 2026-08-14]`; the inherited "19 sites" checked out exactly |
| `page_components.rendered_html` (stored artefacts) | **461 of 1,485 rows, across 20 sites** `[MEASURED]` (I first wrote 460, inherited). Of those, **330 are on the 14 sites where the ink diverges** — the ones a repoint would actually change |

All four non-`content_components` rows `[MEASURED 2026-08-13]` by me, same anchored pattern, one
query per surface. The `content_components` row reconciles exactly and that is worth stating,
because it is the arithmetic that makes the block scan trustworthy: **423 in-block + 30 inline =
453 total declarations**, and 453 is what the anchored pattern returns fleet-wide. Every occurrence
is accounted for in exactly one bucket, which is the property a transform over this corpus needs.

### 7. What is now the actual fix, and what must not be done first

**The repair belongs in the derivation, not in 347 consumers.** `legibleInkFor` should try to make
the source colour *itself* legible — step its lightness toward whichever achromatic extreme gains
contrast until it clears `inkMinContrast` against **every** ground — and only then fall through to
the existing palette walk and achromatic fallback. That makes `--color-<x>-ink` mean what its own
doc comment says, retroactively improves all four shipped repoints, and is what makes a fleet-wide
repoint a legibility fix instead of a de-branding.

Constraints that must survive any such edit, already recorded on this bug: `grounds` stays a
**slice** (R8), no `isDarkHex` narrowing (R7), and the already-clears-AA branch must still return
the source unchanged.

**Do not repoint more consumers before that lands.** Every repoint applied today is a consumer
permanently pointed at `--color-text`; fixing the derivation afterwards repairs them all at once,
whereas repointing first spends the brand on 13 sites to buy contrast that the derivation fix would
have delivered without it.

**Disconfirming test for whoever does it** (Control D): after the change, dartsonline's
`--color-primary-ink` must be a **navy** — a lightened relative of `#1A1F2E`. If it is still
`#F0F2F7`, the derivation fix did not land and the repoint must not run.

### 8. Method note — the 090 substitution, declared

The claim in §2 is structural, concerns shared infrastructure, and is the kind this repo's
2026-07-31 ruling says should go through `090` before being asserted. It has **not** been through
`090`. Substituted first-hand verification, declared rather than omitted: 16 served stylesheets
fetched and compared at the artefact (§1, disconfirmable and nearly disconfirmed by any single row),
plus the deciding line read in the source (§2), plus the ground-truth counter-example queried from
the live corpus (§5). A `090` run on "the preference walk cannot return anything but `text`" remains
cheap and welcome; it is not owed for the *damage*, which is measured.

### 9. Owed measurements, with their queries — and why they are listed rather than quoted

> **RESOLVED 2026-08-14 — all three were re-run by me once the token came back, and the corrected
> figures now stand inline in §§4–6. Two of the three inherited numbers were wrong: the blast radius
> was an undercount (330 across 14 sites, not ~224 across 13) and the translucent share was 41 of 76,
> not 38 of 75. The third (19 sites of chrome) checked out exactly. Queries kept below so the next
> reader can re-run them rather than trust me.**

The kubeconfig token expired part-way through my verification pass (routine 3-day expiry; the owner
refreshes it). Three figures that a subagent produced were therefore **not re-run by me at the time**,
and were marked `[UNVERIFIED — inherited]` where they appeared above rather than left wearing a
`[MEASURED]` badge. None of the section conclusions rested on them — each section's load-bearing
evidence was first-hand — and the marking is what let me come back and settle them the same night,
which is the whole argument for marking rather than omitting.

**The queries stayed here after being run, deliberately.** Two of the three inherited numbers turned
out wrong when I re-ran them, so the next reader should be able to check mine the same way rather than
inherit them from me — which is exactly the mistake this section records.

```sql
-- (a) §5's translucent-ground share of the self-painted blocks.
--     Inherited claim: 38 of 75 (51%).
WITH blocks AS (
  SELECT cc.id, m[1] AS blk
  FROM content_components cc,
       LATERAL regexp_matches(coalesce(cc.html_template,''), '\{[^{}]*\}', 'g') AS m)
SELECT count(*) FILTER (WHERE blk ~ '[;{ ]color[ ]*:[ ]*var\(--color-(primary|accent)[,)]'
                          AND blk ~ 'background')                              AS self_painted,
       count(*) FILTER (WHERE blk ~ '[;{ ]color[ ]*:[ ]*var\(--color-(primary|accent)[,)]'
                          AND blk ~ 'background'
                          AND blk ~ 'background[^;]*(rgba|transparent|--section-)') AS translucent
FROM blocks;

-- (b) §4's blast radius, per site. Inherited claim: ~224 placements across 13 sites,
--     webdesign.co.uk 49. My own containing figure is 461 of 1,485 rows fleet-wide.
SELECT s.domain, count(*) AS placements
FROM page_components pc
JOIN pages pg ON pg.id = pc.page_id
JOIN sites s  ON s.id  = pg.site_id
WHERE pc.rendered_html ~ '[;{ ]color[ ]*:[ ]*var\(--color-(primary|accent)[,)]'
GROUP BY 1 ORDER BY 2 DESC;

-- (c) §6's site/page spreads for the two rendered surfaces.
SELECT count(DISTINCT site_id) AS sites
FROM site_components
WHERE rendered_html ~ '[;{ ]color[ ]*:[ ]*var\(--color-(primary|accent)[,)]';
```

**Read (b) against §1's table, not against the inherited count.** §1 lists **14** diverging sites;
the inherited figure says 13. That discrepancy is unreconciled and is exactly the kind that turns
into a quoted fleet number if nobody notices — the join in (b) settles it.

### 10. The trap that produced §9, recorded because it is cheap to repeat

I ran three subagents to map this mechanism, and one of them was tasked with attacking my own plan.
It did, correctly and decisively — §1's finding is downstream of its challenge. **Then I wrote its
supporting figures into this file marked `[MEASURED]` without re-running them.** They were measured;
they were not measured *by me*, and one of the two I did re-check turned out to be wrong (the
`background-color` claim in §4, actual 5 components not 102).

The general shape, which this file has recorded in other forms and which now has a subagent-specific
instance: **a delegated measurement arrives in the same voice as a first-hand one, and the marker
system cannot tell them apart.** `[MEASURED]` records that a number came from a query, not that the
writer ran it. A borrowed number needs its own marker, and re-running it is usually one command.
Logged in `WRONG_CALLS.md`.

### 11. A sibling defect the derivation fix will NOT close: a `var(--x, fallback)` whose `--x` is DEFINED BUT OF THE WRONG TYPE

**Contributed 2026-08-13 by the `bugfix_122_contrast_ink_slots` lane (the render-audit retraction
half), carried here at their request to avoid a same-file passenger while I held 181 uncommitted
lines. Their evidence, their finding — recorded under their name, not mine.** Full working in
`NOTES_contrast_ink_slots.md`, 2026-08-13 §5.

On `robot-hands.com`, `.cta-btn-primary` sets:

```css
color: var(--color-cta-bg, var(--color-primary));
```

and that site's `--color-cta-bg` holds **`linear-gradient(135deg,#3b82f6,#2563eb)`**. A gradient is
not a valid `<color>`. Because the variable **is** defined, the sane-looking fallback
(`var(--color-primary)`, itself falling back to `#1A1F2E`) is **never reached** — the declaration is
invalid at computed-value time, `color` inherits `#ffffff` from `.cta-section`, and the button paints
white on white.

Measured across 8 sites by the contributing lane: the *consumer* shape is fleet-wide — **8 of 8 use
`--color-cta-bg` in a `color` slot** — but the gradient is only on **3 of 8** (`robot-hands.com`,
`finetuning.uk`, `gaswholesalers.com`), and all three pair it with `--color-cta-text: #ffffff`, which
is why it has survived unnoticed. The other 5 hold a plain hex there and are fine.

> **`[INFERRED]` LIFTED 2026-08-14 — the check was written down, then run, and it PASSED.** The
> inheritance step was reasoned rather than observed when this was filed (neither of us had a live
> token). The contributing lane's token came back and they read the filed row's own spec:
> **`fg = rgb(255,255,255)` on `bg = rgb(255,255,255)`, ratio 1, `text_sample "Run MatchMatrix"`** —
> which is the primary button's own link text. That is exactly the disconfirming check recorded
> below, and it confirmed rather than refuted.
>
> **Fleet-wide: 16 of 17 filed `%cta-btn%` rows show the same 1.0:1 white-on-white signature.** The
> 17th is a clean control and worth more than the 16: `leopardessconsulting.co.uk`'s
> `A.tool-cta-btn-primary` at **2.27:1** — a *valid* token that is merely too pale. So the two faults
> are distinguishable by ratio alone (≈1.0 = the dead-fallback type error; 2–4 = an ordinary
> contrast failure), which means the signature is diagnostic rather than just suggestive.

**The check that lifted it**, kept because the next reader should be able to re-run it rather than
trust us: the filed `contrast_failure` row's spec carries the audit's measured fg/bg/ratio for the
selector. `fg ≈ bg ≈ #ffffff` at ≈1.0:1 confirms the dead-fallback mechanism; anything else refutes
it.

**Why it belongs on this bug and why my §7 fix does not touch it.** It is the same family — a palette
slot used in a role it was not authored for — but the failure is a *type* error, not a contrast one,
so making `--color-primary-ink` a legible tint does nothing for it. And it is worse than the case
this bug started with:

> **A `var(--x, fallback)` whose `--x` is defined but of the wrong TYPE is strictly worse than one
> whose `--x` is undefined, because the fallback is dead code while the source reads as though it has
> a safety net.**

That generalisation is the contributing lane's, and it is the most transferable thing on this page
today. Note it cuts *toward* the §7 fix rather than against it: `buildLegibleInkDefaults` emits only
hex, and `TestBuildLegibleInkDefaults_NeverEmitsAnEmptyOrIndirectValue` already pins that, so the ink
companions cannot themselves become a dead-fallback trap. **Anything that later teaches the renderer
to emit a gradient into an ink slot would create exactly this defect**, and that test is what stops
it. Filed as its own trap in `LANDMINES.md`; a separate bug number is the contributing lane's call.

### 12. 2026-08-14 — the derivation fix is LIVE (`v1.0.1298`), dormant, and now behind ONE protection instead of two

`[MEASURED 2026-08-14, both controls passing]` Build point **`bc39e7bf5`**, pods up 08:58Z.
`git merge-base --is-ancestor 12cf55015 bc39e7bf5` → true; same for `8ad05d01a`. Controls: HEAD is
correctly NOT an ancestor, yesterday's `69612d692` correctly IS. The stamp came from the
`bugfix_122_contrast_ink_slots` lane (adapter provenance line, plus a binary probe on
`agent-chassis-64cb9c4bb9-6tfxf` where `bc39e7bf5` was present and `69612d692` absent, so the probe
discriminated) and was re-verified here by ancestry rather than taken on trust.

**Both rounds shipped in the same build**, so the round-1-only failure mode — an ink that reads as a
correct navy while measuring 3.93:1 on the composited ground — never reached a site.

**Nothing has changed for any visitor.** Read live the same day:
`dartsonline.com --color-primary-ink: #F0F2F7`, `robot-hands.com: #E2E8F0` — both still the pre-fix
`--color-text`, because a stylesheet only picks this up when it re-renders.

**But the protection count halved, and this is the operative fact for anyone reading §7.** Until this
roll, "no visitor sees a change" rested on two independent facts: the code was not in the binary, and
no stylesheet had re-rendered. **Now it rests on one.** Any re-render of any of the 14 diverging
sites, fired by any lane for any unrelated reason, regenerates that stylesheet with the new
derivation. Nobody has to intend it, and this lane is not driving it but cannot prevent it.

So: the owner's ruling on whether the tinted-brand-colour change is wanted is now a **soft** gate. It
was a hard one yesterday. **Read the served ink on the day; a reading carried forward from an earlier
session is not evidence about now.**
