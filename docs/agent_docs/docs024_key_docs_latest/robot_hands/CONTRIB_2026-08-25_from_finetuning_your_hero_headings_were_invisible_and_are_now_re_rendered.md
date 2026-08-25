# CONTRIB 2026-08-25 — from `finetuning_uk_service`: your site's page HEADINGS were invisible, and I have re-rendered three of your pages to fix it

**You are being told because a shared component change touched your site** (owner ruling
2026-07-29 §3: a shared mechanism's other consumers must be TOLD, not merely measured).
Nothing here needs a decision from you. It is already done and verified; this is the record.

## What was wrong on your site

Your theme puts a **gradient** in `--color-cta-bg`. Three shared hero components
(`about-hero`, `contact-hero`, `services-hero`) substituted that token into a CSS
**`<color>` position** — a gradient colour-stop and a `color-mix()`. A gradient there is
**invalid at computed-value time**, so the whole `background` declaration was discarded and the
hero band painted *nothing*: the page's own background showed through, under white heading text.

Measured with `scripts/render_audit.py` `[MEASURED 2026-08-25]`:

- `robot-hands.com/about.html` — `A.cta-btn` **1.00:1** (white on white), plus the hero band fault
- the hero `H1` measures **1.11:1** where **3.0** is the floor

⚠ **The trap, because it will bite you elsewhere:** the rule carried the standard
progressive-enhancement pair — a plain `background: var(--color-cta-bg, …)` and then an
"enhanced" gradient line. That idiom does **not** work for `var()` substitution: the cascade takes
the LAST declaration, and invalid-at-computed-value-time falls back to *inherit/initial*, **never**
to the earlier declaration in the same rule. So the guard line reads like insurance and can never
fire.

## What was done

- **Migration 619** removed the invalid declaration from the three hero components (the valid
  single declaration above it already covered both a gradient and a colour).
- **Migration 630** converged `tool-cta`'s button face onto the palette slot its sibling already
  used, instead of a hard-coded `#fff`.
- **Migration 631** filed one `page_rerender` / `reason='template_changed'` per broken page —
  because a template edited by SQL ships **nothing** on its own, and a reason-less rerender
  re-assembles stored HTML. Three of your pages were in that set.
- A Go change adds `--color-cta-bg-ink`, the legible-ink companion the CTA buttons now read.
  **That half is INERT until the next chassis roll**, so your `A.cta-btn` stays 1.00:1 until then.

## What you may want to check

After the next chassis roll, the CTA-button half needs one more fan-out (deliberately held so your
pages re-render once, not twice). Until then:

```bash
scripts/render_audit.py https://<your-domain>/about.html
```

Full workup, the controls, and the two repairs that look right and are wrong:
`bugs_open/398_HANDOFF_2026-08-25_cta_bg_may_be_a_gradient_and_five_components_use_it_where_only_a_colour_is_valid.md`
(⚠ the number 398 is ambiguous — another lane filed a different one the same day; resolve by slug).

— `finetuning_uk_service`, 2026-08-25
