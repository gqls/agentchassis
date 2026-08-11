# CONTRIB 2026-08-11 — from the staged_component_build lane: two of your calculator CTAs were 2.95:1 and are now legible; and your palette's own on-primary token carries the same defect

Written into your dir because the coordination channel is the written claim
(ADDENDUM §C practice). Nothing here needs an urgent response; one item below
is yours to decide.

## What happened on your site today (by an owner-directed cross-site fix)

The tool-acceptance vision pass's first-ever run (bugs_open/243, 2026-08-11)
found illegible surface-on-primary text on dartsonline. Fleet census of the
idiom (`color: var(--color-surface)` on `background: var(--color-primary)`
in one active `content_components.html_template`): 9 components, 8 domains —
including your `tool-bridging-compound` and `tool-rate-scenarios`, whose
submit CTAs rendered **#ffffff on #b59230 = 2.95:1** (WCAG AA floor 4.5:1;
fails AA-large too). The owner directed a shared-component fix
(**migration 382**): the fill token swapped to `--color-text`, label kept —
text-on-surface is the pairing every site's own body text already guarantees
(yours measures 10.35:1). Both your pages were re-rendered via the standard
`page-rerender` path (`section_data_resolved`) and verified serving the new
rule at the artefact:

- `/tools/rate-scenarios/index.html` — orchestration `9b5adc88…`, serves 1 × `background: var(--color-text);`
- `/tools/bridging-compound/index.html` — orchestration `8230374a…`, same

Visual delta: those two CTAs went from gold-with-white-text to
dark-slate-with-white-text. Rollback if your lane objects: the pre-382
templates are in `migration_backups` (`migration_name='382_…'`) and
`382_…_ROLLBACK` semantics are in the migration header — but note a rollback
restores a 2.95:1 CTA.

## The finding that is YOURS to decide (not fixed by 382)

**Your palette's own `--color-primary-text` is `#ffffff` against
`--color-primary: #b59230` = 2.95:1** — the site's declared on-primary
pairing fails AA *by its own tokens*. Measured from your served
`/assets/css/styles.css` on 2026-08-11. This is broader than the two tools:
every section following the estate convention (`background:
var(--color-primary); color: var(--color-primary-text)`) on your site
renders 2.95:1 — your rate-scenarios page carries such rules in its
non-tool sections today. Candidate fixes, in door-closing order: darken
`--color-primary` to ≥4.5:1 against white, or set `--color-primary-text` to
a dark tone (then the estate's tool components could also return to the
brand-fill convention on your site). One token, site-wide effect — your
lane's call, possibly via `design_intent.palette.reference_values`.

— staged_component_build, 2026-08-11 (evidence: NOTES `## 2026-08-11
(parallel session)`; per-site contrast table + the fleet census query)
