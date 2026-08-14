# 212 — 47 components redefine the `--section-*` tokens the renderer says it owns, and win

**Filed** 2026-08-06 by the `bugfix_122_contrast_ink_slots` lane, which found it
while answering a council objection and **deliberately did not fold it into that
fix** — it is an unenforced contract, not a missing variable.

> **CORRECTED 2026-08-07 — the title and that last clause are both wrong, and the
> fix ranking in §5 is refuted. Read §8 before acting on anything above it.**
> The contract is **not** unenforced: on gamesdesign.co.uk the platform *detected*
> this exact defect, described it correctly, routed it to a live fixer, and stamped
> the work item `complete` — with nothing written. And "and win" understates it:
> on the motivating case the renderer's own contrast-checked value measures **1.71:1,
> slightly worse than the 1.72:1 literal that overrides it**, so the two candidates
> §5 ranks highest are no-ops or regressions. What caught it: reading
> `buildSectionDefaults`' first three lines, then asking the work-item queue what it
> already knew about gamesdesign. Both were available at filing time and neither was
> done. Logged in `WRONG_CALLS.md`.

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

Added 2026-08-07: `bugs_closed/077` (the same detector's remit split),
`architecture_review/RFC_017` (the completion-verifier registry, which names
`hardcoded_section_colors` explicitly), owner ruling 2026-08-02 / RFC_010
narrowing 1 (N producers on one `item_type`).

---

## 8. What the 2026-08-07 session measured, and what it changes

Session picked this up from the handoff after the v1.0.1262 chassis roll. Two
`090` runs were filed before anything durable was asserted (§8.5). Everything
below is first-hand and dated **2026-08-07** unless marked.

### 8.1 The renderer emits nothing at all unless something is dark [MEASURED]

`buildSectionDefaults` (`color_util.go:185-187`) opens with:

```go
if !bgIsDark && !surfaceIsDark {
	return ""
}
```

So on a site whose background *and* surface are light, **step 10 emits no block
at all**. That is the mechanism behind trap 4's guess about idea.uk, now
confirmed at the source rather than inferred from an absence. Served
stylesheets, fetched today:

| site | `Renderer-enforced section defaults` block | `--section-text:` declarations |
|---|---|---|
| gamesdesign.co.uk | 1 | 2 |
| vonc.com | 1 | 2 |
| idea.uk | **0** | **0** |

And when it *does* emit, it covers exactly two grounds — `body`, plus a
**hardcoded list of five classes** (`color_util.go:239-243`):

```
.differentiators-section, .features-section, .faq-section,
.services-section, .about-section
```

`.system-stats-section` is not among them, and never could be: the list is a Go
literal. So the comment's claim — *"the renderer owns this"* — is far wider than
what the renderer can actually answer for. **A component that paints its own
ground is outside the model, not in breach of it.**

### 8.2 On the motivating case the renderer's own value is WORSE [MEASURED]

gamesdesign's served palette is `--color-primary: #00bcd4`, and its emitted block
is `--section-text: #e2e2e2` / `--section-text-muted: rgba(226,226,226,0.75)`.
The section's ground is `--section-surface` (5% white) over primary =
`rgb(13,191,214)` — reproducing §3's backdrop exactly.

WCAG 2.1 with alpha compositing, all four inks against that same ground:

| ink | source | ratio |
|---|---|---|
| `rgba(255,255,255,0.9)` | component literal | 2.04:1 FAIL |
| `rgba(255,255,255,0.7)` | component literal | **1.72:1** FAIL |
| `#e2e2e2` | **the renderer's own** | **1.71:1** FAIL |
| `rgba(226,226,226,0.75)` | **the renderer's own** | **1.46:1** FAIL |
| `#0f0f0f` | `--color-primary-text`, already in the served CSS | **8.65:1 PASS** |

The 1.72:1 row reproduces §3's independently browser-measured 1.72:1 — that
agreement is what licenses the other four rows, which are counterfactual and
could not be browser-measured. Script:
`RUNBOOK_contrast_ink_slots.md` § "the section-token counterfactual".

**This refutes §5 candidates 2 and 3 on the case that motivated the file:**

- **Candidate 2** (emit at a specificity components cannot beat) replaces 1.72:1
  with **1.71:1**, and the muted slot **regresses 1.72 → 1.46**. It is not a class
  fix; on this ground it is a very slightly worse repaint.
- **Candidate 3** (`var(--section-text, <today's literal>)`) resolves to the
  renderer's value wherever the block is emitted — i.e. it *is* candidate 2, with
  the same numbers — and falls back to today's failing literal where it is not
  (idea.uk). Zero improvement in both branches.

Neither was wrong to write down; both were rankable as "the class fix" only
because nobody had put numbers on the counterfactual. §5's own framing —
"candidate 2 is the only class fix and also the only one that can break something
that works" — had it backwards: it breaks the motivating case itself.

### 8.3 The 47 are not rogue — 44 of them are declared dark sections [MEASURED]

Census re-run today, unchanged at **47 declaring / 173 active**, and split:

```sql
SELECT count(*) FILTER (WHERE decl)                        AS declaring,           -- 47
       count(*) FILTER (WHERE decl AND is_dark_section)    AS declaring_flagged,   -- 44
       count(*) FILTER (WHERE decl AND NOT is_dark_section) AS declaring_unflagged,--  3
       count(*) FILTER (WHERE is_dark_section)             AS flagged_dark,        -- 46
       count(*)                                            AS active               -- 173
FROM (SELECT is_dark_section,
             html_template ~ '--section-(text|text-muted|heading|surface|border)\s*:' AS decl
      FROM content_components WHERE is_active AND forked_from IS NULL) t;
```

**44 of the 46 components flagged `is_dark_section` declare `--section-*`** — and
`ValidateDarkSectionContract` (`validate_dark_sections.go:42`) *warns when a dark
section is missing them*. So a second live mechanism requires the very
declarations `buildSectionDefaults`' comment forbids. Candidate 1 (a write guard
refusing `--section-*`) would put those two in direct conflict.

### 8.4 The real defect: detected, routed, handled, closed — and untouched

This is the finding that reframes the file. `color-variable-fixer` is active
(since 2026-02-26) and **is** driven: 15 `site_work_items` all-history, 9
`complete`, most recent 2026-08-05. One of them is gamesdesign's, and its summary
is this bug:

> **`8200cee6-2529-4e82-915f-6df953a5809c`**, created 2026-08-03 21:05:06,
> `complete` 21:08:23 (3m17s):
> *"system-stats-section uses var(--color-primary, #1a1a2e) as its background, but
> the palette defines --color-primary as #00bcd4 (cyan). This means the section
> renders with a bright cyan background instead of a dark surface, making white
> text nearly illegible."*
> `spec.acceptance_test`: *".system-stats-section computed background-color is a
> dark colour (luminance < 0.1), not the #00bcd4 cyan primary"*

Nothing was written, and the timestamps prove it rather than suggesting it:

- `content_components` (`function='system-stats'`) `updated_at` =
  **2026-08-03 10:31:15** — *10.5 hours before the item was created*. The template
  still matches both `--section-text:\s*rgba\(255,255,255` and
  `background:\s*var\(--color-primary`.
- `page_components.rendered_html` `updated_at` = **2026-08-03 21:24:24** — 16
  minutes *after* the close, and still carrying the literal. So the page did
  re-render; it re-rendered the unchanged source.
- §3's browser measurement (2026-08-06) and §8.2's served CSS (2026-08-07) both
  still show the defect, four days on.

**Why it closed.** The item_type is the join, and it is shared by two producers
with different predicates:

1. `write_audit_findings_action.go:117` maps design-audit category `dark_section`
   → item_type **`hardcoded_section_colors`**, handler `color-variable-fixer`.
2. That item_type has a registered completion verifier,
   `VerifyHardcodedSectionColorsResolved`
   (`discovery_checks/check_hardcoded_section_colors.go:290`), whose verdict
   helper applies `PartitionByRemit(candidates, ReplaceHardcodedColors)` — the
   *discovery check's* population, i.e. **hardcoded dark hex literals the fixer's
   regexes would rewrite to `var(--color-primary)`**.

gamesdesign's defect is `background: var(--color-primary, #1a1a2e)` — **already a
`var()`**. `ReplaceHardcodedColors` would not change it, so the population is
empty and the verifier returns, correctly for its own predicate:

> `Resolved: true` — *"no unlocked component carries a colour within the fixer's
> remit"*

This is **not** RFC_017's fail-open path: the verifier did not error, and it was
not wrong. It answered a different question from the one the item asked. The
item's own `acceptance_test` — which states the right question — is written by
`write_audit_findings_action.go:236` and **read by nothing on this path**; every
`acceptance_test` consumer in the tree belongs to the `improve_tool` / tool-
acceptance family. [grep over `--include=*.go` for `acceptance_test|AcceptanceTest`;
absence holds only for those two spellings.]

Filed as its own bug — it is a work-item contract defect, not a CSS one, and it
will mis-close items that have nothing to do with colour. See §8.6.

### 8.5 Fix candidate 5, which §5 did not have

**Repoint the ink to the token that names the ground the component actually
paints.** A section whose CSS says `background: var(--color-primary, …)` takes
`var(--color-primary-text, …)` for its ink — `#0f0f0f` on gamesdesign, **8.65:1**,
and already present in the served stylesheet.

This is not new machinery, and **it is not VIZ-014**: `--color-primary-ink` is
primary *made legible as an ink*; `--color-primary-text` is *the ink that goes on
a primary fill*. Opposite directions — the distinction 122's handoff draws, and
the wrong one here would be a silent no-op.

Nor is it new code. `fix_forced_text_colours_action.go` **already implements
exactly this**: `classifySectionPainting` derives what the template paints from
the CSS itself (`paintPaletteBand` matches
`background[^;{}]*var\(--color-(primary|secondary|accent)\b` — which
`system-stats` matches), and `rewriteSectionDeclarationsInHTML` rewrites literal
`--section-*` declarations to the on-colour family, deliberately ignoring
`is_dark_section` (`:526` — *"metadata only and must never key styling"*).

**So the class fix is written, deployed, and reachable — and the route that
should have run it closed the item instead.** That inverts the work: the open
question is no longer "which of four repairs" but "why did the existing repair
not run", and §8.4 answers it. Candidate 5's remaining cost is a rerender for the
16 placements, not 32 template edits.

### 8.6 State, and what is still open

- **Filed today**: `bugs_open/213` — the predicate mismatch of §8.4.
- **Two `090` runs** were filed before any of the above was asserted, per the
  2026-07-31 owner ruling, since §8.1/§8.4 are structural claims:
  `b6ab22d6-e49c-4b55-a9d9-dd026532a595` (the renderer's grounds) and
  `84c3da66-06c0-41a5-94dc-21fbf71260f0` (the predicate mismatch). Verdicts in
  the lane's `NOTES_contrast_ink_slots.md`.
- **SECOND MEASURED INSTANCE, 2026-08-09 — and it is in the NEW mechanism, not the old
  one.** VIZ-014's ink companions inherit the same two-ground blindness:
  `buildLegibleInkDefaults` computes `--color-accent-ink` against `{background, surface}`
  (`pageGrounds`, `palette_specialised_slots.go`), so on gamesdesign it returned accent
  unchanged (12.46:1 on the page) — and the repointed `.stats-eyebrow`, re-rendered and
  served, still measures **1.44:1 on the primary section fill**, byte-identical to the
  baseline failure. The approved 122 plan counted this closure twice (gamesdesign, vonc);
  neither is reachable. So the question below is not hypothetical: the platform now has
  TWO mechanisms that each answer "is this ink legible?" against grounds a component can
  simply not be standing on.
- **Still genuinely open, and still wanting a human:** §8.1 says the renderer's
  two-ground model cannot answer for a component that paints its own ground.
  Candidate 5 repairs the components one class at a time. Whether the renderer
  should instead learn about component-painted grounds is an architecture
  question, not a bug fix — and it is the one §5 was reaching for.
- **Unchanged and still correct:** every trap in §6, and §4's reason for not
  folding this into 122.

---

## 9. OWNER RULING 2026-08-14 — the renderer learns about self-painted grounds, and one agent owns the repair

Recorded by the `bugfix_122_contrast_ink_slots` lane, which put the §8.6 question to the owner
alongside the ink-colour decision. The ruling, near-verbatim:

> We should make the framework be able to fix it, and it should ideally be one agent's
> responsibility — even if we need a new agent with that responsibility. Making the renderer know
> about self-painted backgrounds would be closer to that goal. We don't want "manual"/CLI fixes.

What this decides, mapped onto §8's options:

1. **The direction is the renderer, not the consumers.** The architecture question §8.6 left open —
   "should the renderer learn about component-painted grounds?" — is answered YES. Candidate 5
   (repoint each self-painting component class at the on-colour token family, one class at a time)
   is **not** the destination: it is at best an interim repair for the ~24 live failures, and any
   use of it must not be read as discharging this ruling.
2. **One agent carries the responsibility end to end.** Detect, decide, repair — not a detection
   that files an item into a queue whose handler half-knows the model (§8.4's predicate mismatch is
   the worked example of why split responsibility fails here). If no existing agent is the right
   home, a new agent with exactly this responsibility is explicitly sanctioned.
3. **No manual/CLI fixes.** A session hand-repointing components — even through migrations — is the
   posture this ruling retires. The framework repairs its own output.

Constraints already measured, which any implementation inherits:

- **The blind spot now lives in TWO mechanisms** (§8.6, measured 08-09): `buildSectionDefaults`'s
  hardcoded five-class ground list, and VIZ-014's `buildLegibleInkDefaults` computing against
  `{background, surface}` — the round-2 composited-ground fix (`8ad05d01a`) widened that to four
  grounds but they are still all *page* grounds; a component-painted primary fill remains outside
  the model. gamesdesign's `.stats-eyebrow` still measures **1.44:1** served, byte-identical to
  baseline, after two shipped "fixes" each computed against grounds it does not stand on.
- **The ground truth for "what does this component paint?" is derivable from the CSS itself** —
  `classifySectionPainting` in `fix_forced_text_colours_action.go` already does it
  (`paintPaletteBand` matches `background[^;{}]*var\(--color-(primary|secondary|accent)\b`), is
  council-reviewed, and deliberately ignores `is_dark_section` ("metadata only, must never key
  styling"). Reuse it; do not re-derive it (§8.5).
- **This is architecture-scope by the estate's own rules** — it changes what a shared mechanism
  guarantees (the renderer's answer would start being conditional on the component's own paint).
  Route it through the architecture track, not a bug patch; RFC-shaped, with this section as its
  brief.

**What this section does NOT do:** assign the work. It records the direction so the next session
picking this up builds toward one owning agent and a ground-aware renderer, instead of costing a
fifth repair option. The ~24 live failures stay open until that lands or the owner separately asks
for the interim repoint.
