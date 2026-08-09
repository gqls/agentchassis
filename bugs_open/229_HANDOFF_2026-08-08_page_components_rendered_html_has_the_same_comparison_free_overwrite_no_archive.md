# 229 — `page_components.rendered_html` has the same comparison-free overwrite as bug 226, and no archive: a section rebuild can still silently destroy artefact-only content

**Filed 2026-08-08 by the `bugfix_226_chrome_divergence` lane**, at the council's
direction in substance: the round-1 `bug_historian` seat (corr `cffbfec4`)
objected that fixing `site_components` alone is "a mechanism-scoped fix, not a
pattern-scoped one" — the identical silent-drop mechanism lives on
`page_components.rendered_html`, and its most iconic instances are page-side.

**090 substitution, stated per the owner ruling of 2026-07-31:** not through the
diagnosis loop. The substitute: the mechanism is the one 226 just measured and
closed for chrome — this file asserts only that the SAME shape exists page-side,
verified by reading the writers (below) and by the platform's own prior filings
(STY-025 register entry: "Full rebuilds silently discard interactive tools
stored only as rendered_html"; the A* pathfinding game destroyed by a
DELETE+INSERT rebuild, recurring on a second site — both cited by the
`bug_historian` seat as this pattern's iconic cases).

## The mechanism (226's, one table over)

Writers that replace `page_components.rendered_html` with no comparison against
the outgoing bytes and no archive (inventory 2026-08-08, `grep -rn "UPDATE
page_components"` + `INSERT INTO page_components`):

- `section_editor_actions.go:1123,1132` — section edit replace
- `create_report_page_action.go:216` — report dossier replace
- `adopt_verbatim.go:510` — ported-page replace
- `fix_harcoded_colours_action.go:242`, `fix_forced_text_colours_action.go:359` — colour-fix rewrites
- `rebuild_blog_listing_action.go:322,357` — listing rebuild (UPDATE + INSERT arms)
- `save_page_sections_action.go:868` — full save path (and the DELETE+INSERT
  family is worse than overwrite: the trigger shape needs a DELETE arm here)
- core-manager admin handlers; raw psql (the 226 origin class, page-side too)

`page_component_history` exists (14,396 rows) **but archives `content_data`
only** — for a component whose real value lives in `rendered_html` (interactive
tools, hand-tuned sections; `rebuild_policy='owned'` pages), the history table
records the wrapper and loses the artefact. Measured for 226: **zero** rows of
`page_components.content_hash` (1294 rows) and `pages.content_hash` (619) are
populated — there is no live artefact-provenance mechanism page-side at all.

## OWNER CALL NEEDED — two council seats disagree on the record about this file's existence (2026-08-09)

In 226's round 2 (trail `cffbfec4`), the `bug_historian` seat **gated** on this
bug being *filed* rather than *fixed*: "the exact shape already indexed under
016b §9: one call site of a shared judgement gets the rigorous fix; the sibling
stays heuristic" — and it is right that page content is arguably the
higher-impact surface. In the SAME round, the `architecture` seat wrote the
opposite condition: "Fine at two instances; a third table adopting the same
shape without a shared abstraction would be the point this needs an RFC" — i.e.
the page-side guard is a legitimately separate decision, and folding it into a
bug patch is the very move the 2026-07-28 seams ruling exists to stop. The
fixing lane held 226 to chrome (per the scope-veto guidance: record the
disagreement, let a human break it). **Whoever takes this bug: the first
decision is whether the archive shape generalises (shared abstraction, maybe an
RFC) or page_components gets its own contained trigger — do not inherit 226's
answer by copy-paste; the DELETE+INSERT family below is why it may not fit.**

## Why this was NOT folded into 226's fix (scope decision, stated)

- `page_components` is orders of magnitude hotter than `site_components` (1294
  vs 57 rows; every section build writes it). A fail-closed archive trigger
  there needs its own volume/pruning design, not a copy-paste of mig 344.
- The DELETE+INSERT rebuild family means an UPDATE trigger is not sufficient
  page-side; the archive shape must handle row replacement, which chrome does
  not have.
- 226's round-1 `debug_historian` seat separately flagged the platform-wide
  blast radius of the chrome trigger; doubling the surface in the same round
  would have been the opposite of contained.

## Fix candidates, ranked by what closes the door

1. **Extend the 344 shape** (`site_component_history` / `trg_site_component_archive`,
   register STY-054): a `page_component_history.rendered_html` column (nullable;
   the existing content_data rows stay valid) + an archive trigger on
   `page_components` covering UPDATE-of-rendered_html AND DELETE, with the
   digest stamp (`rendered_html_digest`) written by the render/save paths
   same-statement. Decide fail-closed vs fail-open for the hot path on its own
   evidence, not by copying chrome's answer.
2. **Ledger only** (no digest, no work items): archive on differing overwrite
   and on delete; recoverable, not yet loud. Cheap first slice.
3. **Convention only**: already the written rule; already failed twice
   (STY-025's cases). Not a candidate, listed for completeness.

## How to verify a fix

226's protocol, one table over: hand-patch a throwaway string into a test
page's section `rendered_html`, run the section rebuild, require the outgoing
bytes recoverable and (candidate 1) a divergence signal; negative control: an
untouched section rebuilds with no archive row when byte-identical, and a
DELETE+INSERT rebuild leaves the pre-delete artefact recoverable.

## Relations

`bugs_open/226` (the mechanism, closed for chrome; this is its page-side
sibling) · STY-054 (the shape to extend) · STY-025 (interactive-section
clobber — the page-side incident register entry) · `page_component_history`
(the content_data-only archive) · council corr `cffbfec4` round 1
(`bug_historian` objection this file answers).
