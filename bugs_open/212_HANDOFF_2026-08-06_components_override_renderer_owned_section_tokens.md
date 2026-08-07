# 212 — 47 components redefine the `--section-*` tokens the renderer says it owns, and win

**Filed** 2026-08-06 by the `bugfix_122_contrast_ink_slots` lane, which found it
while answering a council objection and **deliberately did not fold it into that
fix** — it is an unenforced contract, not a missing variable.

---

## 1. The contract, in the renderer's own words

Every generated stylesheet carries this, emitted by `buildSectionDefaults`
(`color_util.go:179`, appended as step 10 of `RenderCSSFromSpecAction`):

```css
/* ── Renderer-enforced section defaults ── */
/* Text/heading colours are picked from the site palette by contrast ratio. */
/* Themes MUST NOT declare --section-* defaults; the renderer owns this. */
body { --section-text: …; --section-text-muted: …; --section-heading: …; … }
```

The values are **chosen by contrast ratio** against the site's own background —
that is the whole point of the block.

## 2. What actually happens [MEASURED 2026-08-06]

```sql
SELECT count(*) FILTER (WHERE html_template ~ '--section-(text|text-muted|heading|surface|border)\s*:') AS declaring,
       count(*) AS active
FROM content_components WHERE is_active AND forked_from IS NULL;
--  declaring = 47   active = 173
```

**47 of 173** active unforked components declare a renderer-owned `--section-*`
token. **32 of them do it with a raw `rgb()`/`rgba()` literal** — a fixed colour
that cannot know what it will land on.

And they **win**: the component's `<style>` is scoped to its own section class,
which beats the renderer's `body`-level block on specificity. The contrast-checked
value loses to the literal every time.

## 3. The measured damage, so far

| site | failures | element | measured |
|---|---|---|---|
| gamesdesign.co.uk | 8 | `.stat-label`, `.stat-description` | `rgba(255,255,255,0.7)` on `rgb(13,191,214)` = **1.72:1** |
| idea.uk | 14 | `.brief-explanation__*` | `rgb(131,124,114)` on `rgb(232,223,204)` = **3.11–3.35:1** |
| vonc.com | 2 | `.gauntlet-panel-body`, `.gauntlet-micro` | `rgba(255,255,255,0.7)` on `rgb(131,70,255)` = **3.19:1** |

**~24 of the 109 firm contrast failures** on the fleet, from this one shape.

The `system-stats` case, worked through, because it shows why the literal is not
obviously wrong to whoever wrote it:

```css
.system-stats-section { --section-text: rgba(255,255,255,0.9);
                        --section-text-muted: rgba(255,255,255,0.7); }
```

White-on-dark is a perfectly sensible assumption — **for a dark section**. On
gamesdesign the section fills with `--color-primary` (`#00bcd4`, a light cyan),
composited to `rgb(13,191,214)`. The component assumed a scheme it does not
control. That is exactly what the renderer's block exists to decide, and exactly
what the override discards.

> Arithmetic that identifies the ground rather than assuming it:
> `--section-surface` is `rgba(255,255,255,0.05)`, and
> `0.05×255 + 0.95×[0,188,212] = [12.75, 191.35, 214.15]` — the measured backdrop.
> So the fill is `primary`, not a guess.

## 4. Why this is its own bug and not part of `bugs_open/122`

122's fix (register **VIZ-014**, committed `1d2c93a87`) makes a palette colour
reachable as a legible ink. **It cannot touch this class**: these components are
not naming a palette slot wrongly, they are overriding the renderer's own
contrast-checked answer with a constant. Folding a 47-component contract change
into a bug fix is the opportunistic bundling the council's guardian seat vetoes,
and the `editquality` seat was told this was scoped out, by count, at submission.

## 5. Fix candidates, ordered by what makes the bad state unrepresentable

1. **Enforce the contract where it is written.** A component-write guard that
   refuses a `--section-*` declaration in `content_components.html_template`
   (there is already `component_write_guard.go` to hang it on). Makes the state
   unrepresentable going forward; does nothing about the 47.
2. **Emit the renderer block at a specificity the components cannot beat**, or
   scope it to the same selectors. Fixes all 47 at once with no template edits —
   but is a real cascade change and needs its own council round, and it will
   silently repaint any component that was *right* to override.
3. **Repoint the 32 literal-carrying components** to `var(--section-x, <today's
   literal>)`. Mechanical, safe by fallback, and the same shape VIZ-014's
   consumers use. 32 edits, so it wants the migration-ledger discipline.
4. Fix the three failing sites by hand. Rejected — "operators must remember to
   check contrast" is the defect.

**Not obvious which.** Candidate 2 is the only one that is a class fix and also
the only one that can break something that currently works. That trade is the
decision this file is asking for, and it is worth a `090` before anyone commits
to it — this file has NOT had one.

## 6. Traps for whoever picks it up

- **Do not count `--section-*` declarations as damage.** Some components override
  legitimately (a genuinely dark band on a light site). The failing subset is the
  one whose literal contradicts the *computed* ground, and only a browser
  measurement can tell you which — `scripts/render_audit.py`, not a regex.
- **`css_snippets` is the wrong table.** Component CSS lives inside
  `content_components.html_template`. A census against `css_snippets` returns 0
  and looks like a clean bill of health (21 rows, 0 mentioning `--color-primary`
  at all).
- **The stored artefact and the live template can disagree.** Check
  `page_components.rendered_html` as well as the component row; on gamesdesign
  they happened to match, which is not guaranteed.
- **A site with no renderer block at all is a different case.** idea.uk's
  stylesheet has **no** section-defaults block (step 10 emits only under some
  conditions), so there the component literal is not overriding anything — it is
  all there is. Same failure, different repair.

## 7. Relations

`bugs_open/122` (parent), `bugs_open/211` (`--section-heading` going invalid on
ai-agent-orchestration.com — adjacent, possibly the same family), register
VIZ-014, `component_write_guard.go`, `color_util.go:179` (`buildSectionDefaults`).
