# PLAN 2026-08-06 — bug 122, the surviving contrast class: a palette colour asked to be an INK

**Lane opened** 2026-08-06 against `bugs_open/122_HANDOFF_2026-07-27_generated_css_fails_wcag_on_four_live_sites.md`.
**Picked** because it was the coldest bug in the backlog by reference-heat across the
42 live session transcripts (28 hits, next coldest 38), and because `who-owns.py`
named only `dartsonline_traffic` — which contributed *into* 122 on 2026-07-29 and
explicitly declined the generator fix ("candidate 1 still belongs to whoever takes
it"). No `*122*` workstream directory existed before this one.

---

## 1. What I re-measured before planning anything, and what had changed under the file

The bug file's findings are from 2026-07-27/28/29. **Two of its three named findings
are now fixed, and the fix it recommends first has already shipped.** Measured today
against the live fleet, not read from the file:

### Finding 1 (its candidate 1, "stop the header chrome hardcoding `color: white`") — SHIPPED

The live shared header template is var-driven:

```
SELECT substring(html_template from '\.header-cta\s*\{[^}]*\}')
  FROM content_components WHERE name='header-theme-chrome';
-->  background: var(--color-cta-bg, var(--color-accent));
     color:      var(--color-cta-text, var(--color-primary-text));
```

And the stored chrome rows have caught up — 19 header rows fleet-wide,
**0 carrying a hardcoded white CTA ink**: 14 `var-driven`, 4 with no `header-cta` at
all, 1 (`leopardessconsulting.co.uk`) on the bespoke `header-leopardess` component,
which is the only active component still holding the literal.

> So the file's headline candidate is done and its "five of six measured sites fail
> on the same hardcoded declaration" is **spent**. A reader taking that table as
> current would spend a roll re-fixing it.

### Finding 2 (robot-hands.com's invisible white-on-white primary CTA) — FIXED

`.cta-btn.cta-btn-primary` at 1.00:1 is gone from the served page. robot-hands still
fails, but on entirely different elements (below).

### Finding 3 (vonc.com's Gauntlet buttons) — STILL LIVE, and owned elsewhere

23 failures, `.gauntlet-btn-primary` at 1.76:1 as filed. **The Gauntlet workstream
owns that surface** and the file already says to coordinate. Out of scope for this
lane; not touched.

### The fleet today — 15 homepages, browser-measured

`python3 scripts/render_audit.py`, 2026-08-06, one homepage per site:

| site | firm failures | dominant shape |
|---|---|---|
| ai-agent-orchestration.com | 30 | `--color-heading` == its own background |
| vonc.com | 23 | Gauntlet buttons (owned elsewhere) |
| gamesdesign.co.uk | 17 | white/70% on a cyan accent fill |
| idea.uk | 14 | `text_muted` on a light surface |
| finetuning.uk | 10 | white on a mid-tone accent fill |
| gaswholesalers.com | 6 | accent as an ink on white |
| robot-hands.com | 3 | `--color-primary` as an ink |
| oufe.com / webdesign.co.uk / webdesign.uk | 1–2 | over-image approximations only |
| dartsonline.com | 1 | `--color-primary` as an ink |
| fundamentallyai.com, leopardessconsulting.co.uk | 0 | — |
| relojistas.com, vetcomparison.uk | 0 | **were failing on 07-28; now clean** |

**109 firm failures across 12 sites.** Two sites the file lists as failing are now
clean, which is why every figure here is dated and re-derived rather than carried.

---

## 2. The mechanism, stated as narrowly as the evidence allows

Three distinct sub-shapes survive. They are **not** one bug, and the file's framing
("generated stylesheets fail WCAG") covers none of them well, because in every case
the generated stylesheet is doing exactly what it was told.

### A. A palette slot spent on both a FILL and an INK — `--color-primary` [MEASURED]

dartsonline `.image-hover-card-grid__eyebrow` **1.04:1**; robot-hands `.tl-eyebrow`
**1.14:1** and `.tl-card-link` **1.07:1**. The component rules are variable-driven
and individually correct:

```
.image-hover-card-grid__eyebrow { color: var(--color-primary); }   -- content_components
.tl-eyebrow                     { color: var(--color-primary); }
.tl-card-link                   { color: var(--color-primary); }
```

The palette's `primary` is a dark near-background colour — a legitimate choice for a
*fill*, and `16.28:1` for light text sitting on it. As an *ink on the page* it is
invisible. **17 of 18 layouts use `color: var(--color-primary)` as an ink**
(`social-lobby` 11 times, `affiliate-hub` 9, `magazine-grid` 8, …), so this is the
fleet's most-used ink and nothing checks it.

`warnUnusablePrimary` (`palette_specialised_slots.go:328`) **already detects exactly
this condition** at < 3.0 and only logs. The dartsonline contribution to 122 worked
out the arithmetic and correctly concluded no single value satisfies both roles — so
repointing the palette trades an invisible eyebrow for invisible button text. That is
why nothing was changed then, and it is the right call.

**What is missing is a second slot.** `darkSchemeDerivations` derives `primary_text`
— the ink that goes *on* a primary fill. There is **no** derived slot for the inverse:
primary *made legible as an ink on the page*. The platform computes this answer twice
already (`pickInkOn`, `pickReadableOnBackground`) and never offers it in that
direction.

### B. `--color-heading` collapsing onto its own background — cause [UNMEASURED]

ai-agent-orchestration.com serves **six `.H3` headings at 1.00:1** —
`rgb(13,17,23)` on `rgb(13,17,23)` — plus an `.H2` at 1.04 and a `.section-heading`
at 1.10. This is the single worst instance on the fleet and **appears in no bug
file**, including 122.

It should be impossible: `darkSchemeDerivations` has `{name: "heading", from: "text"}`
and the renderer's alias block has `{"--color-heading", "var(--color-text)"}`. Yet the
served value equals the *background*, not the text. **I have not established why, and
I am not going to guess** — the alias is appended only when `--color-heading:` is not
already defined anywhere in the CSS, and this site's stylesheet does define
`--color-heading: var(--color-text)`, so the resolved chain needs reading rather than
reasoning about. Filed for diagnosis rather than asserted (§5).

### C. A component hard-coding an ink over a themed fill — 026 family 3 [MEASURED]

finetuning `.csg-cta-btn` white on `#C8873A` = **3.01**; `.cta-btn cta-btn-primary`
white on white = **1.00**; gaswholesalers `.A` `#E8A020` on white = **2.22**;
gamesdesign `.stats-eyebrow` `#00E5FF` on `#0DBFD6` = **1.44** and eight
`rgba(255,255,255,0.7)` labels on the same cyan = **1.72**.

`accent_text` was derived on 2026-07-27 for precisely this — its own code comment says
so: *"It is emitted so a component can stop hard-coding white over an accent fill."*
**It has zero consumers.** Measured across every surface that could name it:

```
content_components 0 · layouts 0 · css_snippets 0 · site_components 0 · page_components 0
```

That is the LANDMINE `A palette slot no LAYOUT declares is never emitted — deriving it
is dead config`, already recorded from this bug's own dartsonline round, sitting
unfired in the very list it was written about.

**Note the scheme boundary breaks here.** `fillDarkSchemeSpecialisedSlots` is
dark-only by deliberate design, and gaswholesalers (`#F4F1EB`) and finetuning
(`#F5F3EF`) are LIGHT sites. "Is this colour legible as text on that ground" is a
scheme-independent question, so a fix scoped dark-only cannot reach sub-shape C.

---

## 3. Fix candidates, ordered by what makes the bad state unrepresentable

### Candidate 1 — emit legible-ink slots from the RENDERER, not from 18 layouts *(preferred)*

Add a renderer-owned `:root` block that defines, for each palette colour used as an
ink, a companion variable holding *that colour made legible on the ground it lands
on*: `--color-primary-ink`, `--color-accent-ink`. Value = the colour itself when it
clears AA against the background, else the palette colour that does (the existing
`pickInkOn` walk, which prefers a palette colour so the site keeps its character).

**Why the renderer and not the layout templates.** The LANDMINE above is decisive: a
palette slot reaches the stylesheet *only* through `{{palette "X" "literal"}}` in a
layout, so adding to `darkSchemeDerivations` alone ships nothing. The alternative is
editing all 18 layout `css_template`s, which is exactly the two-hand-maintained-lists
drift class this estate keeps getting bitten by. The renderer already owns two such
blocks — `buildSectionDefaults` (`--section-*`) and `buildTokenAliases` (the
compatibility aliases) — with a stated contract that *themes must not declare these
themselves*. This is a third instance of an established pattern, in one place, that
cannot drift across layouts.

**It is additive and inert.** Nothing changes until a component writes
`var(--color-primary-ink, var(--color-primary))`. The fallback preserves today's
rendering exactly where the slot is undefined — the same shape 113's repair used. By
the owner ruling of 2026-07-29 §1, additive-and-inert is normal-council-gate scope,
not RFC scope: no shared *guarantee* changes, and nothing can reach the new value
until a template opts in. **The opt-in is structural**, which is what the 2026-08-02
§2 ruling asks for — a component author's own edit, visible to a reviewer of the
component, not a rule in a doc comment.

**Scoped to both schemes, deliberately**, unlike `fillDarkSchemeSpecialisedSlots`.
Sub-shape C is on light sites. This widening is the one thing in this candidate a
reviewer should push on, so it is stated here rather than buried: the dark-only
boundary exists because a light literal is only *wrong* on a dark site, whereas "this
ink is illegible on this ground" is false-or-true independent of scheme.

Then repoint the shared components that use a palette colour as an ink for small text
— `image-hover-card-grid` (1 page, 1 site) and `tool-list` (6 pages, 4 sites) close
sub-shape A's measured instances; both are shared, unforked, active.

### Candidate 2 — give the render audit a cadence *(complementary, and now one row)*

**This is no longer the build task the file describes.** As of v1.0.1257 the whole
chain exists and is live:

- `write_render_audit_findings` is **in the running binary** — pod-grepped on
  `agent-chassis-5b9fd84984-hqc5d`: **11** occurrences, invented control **0**,
  positive controls `scanStoredStatClaims` 2 / `fillDarkSchemeSpecialisedSlots` 4.
- `render-audit-agent`'s live row carries the full workflow:
  `site → audit → write_findings → complete`.
- It files firm contrast failures as `contrast_failure` routed to `css-patch-agent`,
  deduped on `contrast_failure:<page-path>#<selector>`.

**And nothing dispatches it.** 28 enabled `scheduled_tasks`, none targeting
`render-audit-agent`. Total `contrast_failure` items ever raised: **4**, all
relojistas.com, all 2026-08-04, all `complete` — i.e. it has been run by hand once.
The remaining work is one `scheduled_tasks` insert. This is the `bugs_open/083` /
`093` / `115` shape one more time: *a mechanism made correct and then guarded behind
something that never runs.*

Ordering: run the audit **first**, to bank a baseline, because findings dedup by
page+selector and the next audit is the de-facto verifier — so a pre-fix baseline
makes candidate 1's effect measurable rather than merely asserted.

### Candidate 3 — repoint individual palettes. NOT RECOMMENDED

The dartsonline round already proved this trades one failure for another on sub-shape
A. Recorded so nobody re-derives it.

### Candidate 4 — fix pages by hand. Rejected as a class fix

What produced this state. "Operators must remember to check contrast" is the defect.

---

## 4. How a fix gets verified, and the trap

**Never grade this on a stylesheet or a palette row.** Both were the unsound method
that produced this file's superseded table (its own 07-28 correction) — a stylesheet
cannot resolve the cascade, and a palette cannot see a literal that is in no palette.

1. Bank a pre-fix baseline with `python3 scripts/render_audit.py <urls>` over the 12
   failing sites. Keep the output file; a falling count is **not** evidence of repair
   on its own (this file's dartsonline round showed the count is content-dependent —
   the same defect reports 1 or 2 depending which cards a page renders).
2. Re-render a stylesheet on a dark site whose `primary` fails as an ink, and assert
   the new variable is PRESENT in the served CSS with a value that differs from
   `--color-primary`. A slot that renders identical to its source is the dead-config
   failure wearing a success.
3. **Induce the no-op case**: a site whose `primary` already clears AA must get the
   slot with `primary`'s own value, and its rendered pages must be byte-unchanged.
   A gate that fires on everything is as useless as one that fires on nothing.
4. Only then re-run the render audit and compare like for like.

## 5. Sequencing, and what is filed rather than asserted

- **Sub-shape B is going to the diagnosis loop before I write a cause anywhere.** It
  is cross-cutting (a shared alias + a shared derivation, on a live commercial site),
  the cause is not where the symptom is, and CLAUDE.md's own corrected section plus
  the 2026-07-31 owner ruling both point at `090` for exactly this. A confident-feeling
  guess about the alias chain is the failure mode that section was rewritten to stop.
- **Sub-shape A and C are the code change**, and go through the council gate before or
  alongside the commit, registered in the concept register in the same commit
  (2026-07-29 condition 2, which stands).
- **Candidate 2's cadence row is live config**, so it needs no build — but it should
  not land before the baseline in §4.1 is banked.

## Corrections to the originating brief

- 122's candidate 1 is recorded there as outstanding. **It has shipped**; its
  five-of-six table is spent evidence. Corrected in the bug file, dated.
- 122's candidate 2 is recorded there as "the tool already exists and is not wired to
  anything — that is the whole of the remaining work". **The Go port, the orchestration
  and the work-item drain all now exist and are live**; the remaining work is a
  cadence row, not wiring.
