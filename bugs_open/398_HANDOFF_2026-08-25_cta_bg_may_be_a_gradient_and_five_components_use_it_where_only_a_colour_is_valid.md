# 398 — `cta_bg` may legitimately hold a GRADIENT, and five components use it where only a `<color>` is valid, so the declaration is silently discarded and text lands on its own background

> ## ⚠ THE NUMBER 398 IS AMBIGUOUS — refer to this case BY SLUG
> Another lane filed `bugs_open/398_HANDOFF_2026-08-25_scheduled_tasks_row_is_not_single_flight.md`
> the same day. Both are real, neither is renumberable (numbering is never reassigned). This case is
> **`cta_bg_may_be_a_gradient_and_five_components_use_it_where_only_a_colour_is_valid`**.
> `git log` the FILE PATH, not the number.

**Filed 2026-08-25** by the `finetuning_uk_service` lane, from an owner report on his own site.
**Status: OPEN — fix committed `7fe5bd7b6`, council `Council-Submitted: f0591cb2-d65d-4517-a676-0334a7ff29a8`.**
Hero half goes live on migration `619` + a re-render; button half is inert until the next chassis roll.
**Severity: HIGH** — it makes headings and CTA button labels invisible on live pages, including a
paid front door, and it has been doing so for at least fifteen days with the platform's own
detector having already found it.

> **No `090` run — first-hand verification substituted, per the 2026-07-31 ruling, and stated
> rather than omitted.** The cause is read directly from: the component template
> (`content_components.html_template`), the site's palette (`css_themes.css_content`), the served
> pages measured with `scripts/render_audit.py`, and **three cross-site controls** including one
> that isolates the single causal variable. The claim is structural but not inferential.

## 1. The owner's report, which was more precise than it looked

> *"a couple of the pages have no hero images which has meant that the copy is also unreadable.
> e.g. services.html"* — owner, 2026-08-25

Both halves are true and the causal "which has meant" is **correct**, which is why this file leads
with it. The hero template takes an imaged branch whenever the page has one:

```
class="hero hero-services{{if or .hero_url .background_image}} hero-services--imaged{{end}}"
  {{if or .hero_url .background_image}} style="background-image: linear-gradient(rgba(6,11,20,0.62),
  rgba(6,11,20,0.72)), url('...'); ... --hero-ink: #fff;"{{end}}
```

A page WITH an image gets a dark scrim and an inline white ink — readable. A page with **no** image
falls through to the CSS band below, which is where the defect lives. So "no hero image" is the
precondition, and every fleet site checked (oufe, homegarden, garden-tools, cookly, idea, agritec)
takes the imaged branch — which is exactly why this has stayed hidden.

## 2. The mechanism

`cta_bg` is the **one** palette slot whose contract admits a gradient as well as a colour.
That is not a defect in the palette and it is documented in-tree:

> `derive_brand_head_assets_action.go:275` — *"Gradient values (e.g. cta_bg) are rejected by
> parseHexColour downstream."*

**10 of the fleet's `css_themes` hold a gradient in it** `[MEASURED 2026-08-25]`: `calm-minimal`,
`dark-modern`, `default`, `professional-dark`, `tech`, `warm-friendly`, `theme-finetuning-uk`,
`theme-gaswholesalers-com`, `theme-loanandmortgagecalculator-co-uk`, `theme-robot-hands-com`.
(A looser `LIKE` also matches `soft-editorial`, `theme-dartsonline-com`, `theme-vonc-com`,
`theme-webdesign-uk` — those four are **solid**; discriminate on the regex, not the `LIKE`.)

Five component rows put that token where CSS requires a `<color>`:

| component | the declaration | position |
|---|---|---|
| `about-hero`, `contact-hero`, `services-hero` | `background: linear-gradient(135deg, var(--color-cta-bg) 0%, color-mix(in srgb, var(--color-cta-bg) 82%, ...) 100%)` | colour STOP + `color-mix()` |
| `call-to-action` | `.cta-btn-primary { background: var(--color-cta-text); color: var(--color-cta-bg, var(--color-primary)); }` | `color:` |
| `tool-cta` | `.tool-cta-btn-primary { background: var(--color-white,#fff); color: var(--color-cta-bg); }` | `color:` |

A gradient in a `<color>` position is **invalid at computed-value time**. The declaration is
discarded and the property falls back to inherit/initial:

- the hero band paints **nothing**, so the page's own background shows through;
- the button's ink **inherits**, and inside a CTA band what it inherits is white.

> ### ⚠ THE TRAP, and the transferable half (also filed as `016b` §9)
> The hero's own safety net is what defeats it. The rule contains TWO `background` declarations:
> ```css
> background: var(--color-cta-bg, var(--color-primary));                    /* valid for a gradient */
> background: linear-gradient(135deg, var(--color-cta-bg) 0%, ...);         /* invalid  */
> ```
> This is the standard progressive-enhancement idiom — *"simple version first, enhanced version
> second, old browsers keep the first"*. **It does not work for `var()` substitution.** The cascade
> picks the LAST declaration, and invalid-at-computed-value-time then falls back to
> **inherit/initial — never to the earlier declaration in the same rule.** So the fallback line is
> dead weight that reads like insurance. Anyone auditing this file sees a guarded declaration and
> moves on.

## 3. Measured, with controls

`scripts/render_audit.py` (VIZ-010), all `[MEASURED 2026-08-25]`:

| page | element | ratio | needs |
|---|---|---|---|
| finetuning.uk/services.html | `.hero-content H1` | **1.11:1** white on `rgb(245,243,239)` | 3.0 |
| finetuning.uk/about.html | `.hero-content H1` | **1.11:1** | 3.0 |
| finetuning.uk/services.html, /about.html, **/your-own-model.html** | `A.cta-btn` | **1.00:1** white on white | 4.5 |
| robot-hands.com/about.html | `A.cta-btn` | **1.00:1** | 4.5 |

**The measured background is the page cream `rgb(245,243,239)`** — neither the blue gradient nor
the `var(--color-primary)` fallback `#1A1A2E`. That is what identifies the declaration as
*discarded* rather than merely ugly, and it is the observation the whole diagnosis rests on.

**Control 1 — same component, solid token, passes.** `noted.co.uk/contact.html` runs the identical
non-imaged `hero-contact` band with `--color-cta-bg: #e9e2d3`. **Zero** hero contrast failures.
The only variable is the token's TYPE.

**Control 2 — in-repo.** `hero`, `use-cases-hero` and `case-studies-hero` carry the SAME
two-declaration idiom against `--color-primary`, which is always a colour. They are correct, and
they are deliberately untouched by the fix.

**Control 3 — second site.** robot-hands.com reproduces the 1.00:1 button independently.

## 4. The platform found this on 2026-08-10 and lost it — twice, by two different mechanisms

`site_work_items`, finetuning.uk, `item_type='contrast_failure'`: **11 rows, all created
2026-08-10** `[MEASURED 2026-08-25]`.

- **4 `cancelled`** — including `contrast_failure:/about.html#H1.H1`, `/approach.html#H1.H1`,
  `/contact.html#H1.H1`, each summarised *"Contrast 1.11:1 (needs 3.0:1)"*. That is **this exact
  defect**, filed fifteen days ago. They were filed with the invented `H1.H1` selector
  (`bugs_closed/352` / VIZ-016 — a class-less element filed under its own tag name, matching
  nothing) and retired wholesale by **migration 587 on 2026-08-24**. The selector was withdrawn;
  the defect was never repaired. A reader of those rows today sees `cancelled` and concludes it was
  handled.
- **7 `deferred`** — the `A.cta-btn` 1.00:1 rows, plus two near-misses. Untouched since 2026-08-11.
  Per **`bugs_open/396`** (*work items parked at `deferred` with a named `handler_agent`*) these are
  undispatchable, un-promotable **and un-re-filable**: `deferred` is NOT in `idx_swi_dedup`'s
  terminal list, so a fresh audit's insert on the same `item_key` fails `23505`, *which reads as
  "already queued" and means "queued and abandoned"*.

> **Consequence for anyone verifying this fix: a re-run of the render audit will NOT re-file the
> button rows.** The 4 `cancelled` H1 rows can re-file (`cancelled` IS terminal in the index); the
> 7 `deferred` ones cannot. Their honest disposition once the fix is live and measured is to close
> them naming `619`, which also unblocks those keys.

## 5. The fix, and the two things it deliberately does not do

Committed `7fe5bd7b6`. Reuses **VIZ-014**'s legible-ink companion mechanism rather than inventing a
parallel one, because this IS that mechanism's stated problem — *"a palette colour used as a FILL
and the same colour used as an INK are different questions"*.

- **`--color-cta-bg-ink`** (`palette_specialised_slots.go`): source is `cta_bg` reduced to a solid
  (`solidCTAFill` → `firstGradientStop`), ground is the inverted button's own `cta_text` face.
- **Migration `619`**: heroes lose the second `background:` declaration (the valid one above it is
  already correct for both types); the two buttons repoint to `var(--color-cta-bg-ink, var(--color-cta-bg))`.

> ### ⚠ THE OBVIOUS FIX IS WRONG, and it looks right on the site that reported the bug
> Using the gradient's first STOP directly as the ink fails AA against the button's white face on
> **6 of the 10** gradient themes `[MEASURED 2026-08-25]`: `#3b82f6` → 3.68, `#059669` → 3.82,
> `#8b5cf6` → 4.28. Only finetuning's `#1e40af` (8.8) clears it — and that is the theme in front of
> whoever is debugging. The stop is the SOURCE for the derivation; `legibleInkFor` re-tints it.
> `TestBuildLegibleInkDefaults_CTAInkRetintsAnIllegibleFirstStop` goes red if anyone simplifies it.
>
> The other tempting fallback is worse: `--color-primary` on a white button is illegible on **4 of
> 12** themes, worst `theme-loanandmortgagecalculator-co-uk` at **1.24:1** — trading one invisible
> button for another.

**Not done, deliberately:**

1. **`tool-password-entropy_pre_037`** keeps three `background: color-mix(in srgb, var(--color-cta-bg)
   N%, transparent)` declarations — the same invalid-value fault. Its harm is a **missing panel
   tint, not unreadable text**, and it has **ZERO `page_components` uses**. A correct repair needs a
   solid *FILL* companion, a different token from this legible *INK*, with no live consumer — the
   accumulating-opt-in-surface shape `RFC_022` exists to discourage. **Pinned** by 619's verify
   block at exactly 1, so a NEW component taking the same wrong turn refuses the migration.
   ⚠ It was found only because the verify block was INDUCED — the by-name census that preceded it
   had encoded its own answer and returned 3.
2. **No `css-patch-agent` append.** It would repair finetuning.uk and leave the other gradient-theme
   sites broken — and per **`bugs_open/396`** (*a design run erases every appended CSS repair*)
   `persist_css_to_theme` rewrites `css_themes.css_content` byte-for-byte, so the repair would
   expire at the site's next design run while its work item stayed `complete`.

## 6. How to verify

```bash
scripts/render_audit.py https://finetuning.uk/services.html https://finetuning.uk/about.html \
                        https://finetuning.uk/your-own-model.html \
                        https://noted.co.uk/contact.html https://robot-hands.com/about.html
```
Every `1.11:1` hero row and every `1.00:1 A.cta-btn` must be gone; noted.co.uk must stay clean (it
already is — it is the control, not a target). **Discount the `over an image` lines**: the tool says
so itself, they are pre-existing, and they are not what this fixes.

Order: apply `619` → `page_rerender` the affected pages (re-assembles from `content_data` + the
template, so the CSS updates and **the copy is not regenerated** — which is what keeps this
compatible with the finetuning lane's standing copy hold) → measure. The button half needs the
chassis roll first; probe the binary for `--color-cta-bg-ink` with a present-control and an
absent-control through the same NUL-split pipeline, never a bare `strings`.

## 7. Residual

- The 7 `deferred` rows (§4) need closing once measured, naming `619`.
- `firstGradientStop` reads hex stops only. Today every gradient `cta_bg` is hex-stopped; an
  `rgb()`/`hsl()`/named stop would emit no companion and revert those consumers to present-day
  behaviour **silently**. Bounded and stated, not detected.
- No detector yet refuses a component template using `--color-cta-bg` in a `<color>` position.
  `component_validation.go:175` already carries the token allow-list and is the natural home.

## 8. ROUND 2 — what the council found, and what the artefact found (2026-08-25, same day)

The round-1 submission came back **REVISE** on a gating objection from `editquality`, and the
objection was right. Two further defects came out of chasing it — one from the seat, one from
measuring the served page instead of trusting a green status.

**8a. The button had TWO faces, and grounding on both would have been wrong.**
`[MEASURED 2026-08-25]` `call-to-action` painted the inverted button
`background: var(--color-cta-text, var(--color-primary-text))`; `tool-cta` hard-coded
`background: var(--color-white, #fff)`. Same UI object, two faces, one of them underivable from
the palette — the `bugs_open/113` layout-literal shape. The obvious repair (ground the ink on
both faces, since `grounds` is a slice for exactly that) is **wrong, not merely expensive**:
**7 live themes** pair a light band with a near-black ink — `cta_bg #e9e2d3` / `cta_text #1a1a1a`
(noted, idea, lendzy, loanzy, mortgagecalculator, remortgagecalculator, webdesign.co.uk) — so the
two faces are `#ffffff` and `#1a1a1a`, and **no single colour clears 4.5 against both**;
`legibleInkFor` falls through to its terminal branch and emits black at a worst ratio of ~1.2, a
second invisible button in the opposite direction. **Migration 630 converges the face instead**,
and the ink then grounds on the one real surface.

**8b. Migration 619 shipped NOTHING on its own.** Nothing files `template_changed` re-renders for
a template edited by SQL — that fan-out lives in `component-template-fixer.create_rerender`, keyed
to the component the *fixer* changed, and the fixer is an LLM repair agent, not something you
invoke to publish a hand-authored template (`bugs_open/283` §13). So the pages kept serving their
stored `rendered_html`, **with a green status and no error anywhere**.

> **How it was caught, and it was not by reasoning.** A `page-rerender` dispatch reported
> `COMPLETED` for `/about.html`, `/approach.html` and `/contact.html` — and all three still served
> the invalid declaration. `page_components.updated_at` still read **2026-08-17 / 2026-08-24 /
> 2026-08-23**. A reason-less rerender re-assembles STORED component HTML. Only `/services.html`,
> whose components happened to regenerate at 19:46:04, actually carried the repair.
> **`COMPLETED` on the orchestration meant the deploy ran, never that the template was re-rendered.**
> Migration **631** is the fan-out, shaped from `615`, and `template_changed` is confirmed to work:
> gaswholesalers.com's three pages regenerated at 21:19 and their stored HTML is clean.

**8c. The round-1 census could not have proven its own claim.** The verify block pinned
`color-mix` uses at 1 with a needle (`color-mix(in srgb, var(--color-cta-bg`) too narrow to have
found a differently-spaced one. Re-run **position-classified across all 13 components** that
reference `--color-cta-bg`, post-619 `[MEASURED 2026-08-25]`: **as-a-colour 0 · in-a-gradient-stop
0 · in-color-mix 1** (the unshipped `tool-password-entropy_pre_037`) **· as-a-background 13** (the
legal use). The conclusion held; the evidence for it did not. **That is the second census in this
file that encoded its own answer** — the first was by component name, and it missed a row until a
`DO`/`RAISE` was induced. Grep by POSITION.

**Scope of the fan-out: 9 pages, not 55.** 55 active pages still hold the stale hero, but only the
9 on gradient-`cta_bg` sites are BROKEN (finetuning.uk 3, gaswholesalers.com 3, robot-hands.com 3).
On the other 46 the declaration is *valid* and renders a gradient sheen; re-rendering them today
would be a cosmetic change to sites nobody reported. They converge on their next natural render.

⚠ **Two of the three sites are other lanes'** (gaswholesalers.com, robot-hands.com). Included
because the defect is live there and the repair regenerates no copy — but owner ruling 2026-07-29
§3 requires **telling** them, and a CONTRIB in each lane's directory is still owed.
