# CONTRIB 2026-08-04 — Phase 4 groundwork: the composition mechanism, mapped

**From:** the bug-sweep session that picked up
`bugs_open/107_…every_site_gets_the_same_homepage_skeleton.md` on 2026-08-04,
re-validated it, mapped the mechanism — and then found your Phase 4.1/4.2 IS
that bug's fix, owner-approved, in an active lane. Contributing rather than
competing. Nothing in your lane was touched.

**What you get, and where it lives:**

- **A dated mechanism map appended to the bug file itself** (the 2026-08-04
  section): where the composition is born (`build-site-planner.plan_site`,
  mirror at `053_build_site_planner.sql:1978-2228`), the unfiltered
  `load_components` menu, the brochure one-shot example inside the prompt,
  the discarded `site_type` plan field, the starved
  `component_selector.suitable_site_types` hook (+0.35, never fed — 86/178
  active components carry values), the TWO post-planner fallbacks that
  re-impose the skeleton (`defaultSectionsForPage`, sibling modal synthesis),
  and the `validate_components` silent-drop that would kill any
  roadmap-specified novel section.
- **Fresh validity evidence:** lendzy.co.uk (built 08-02, classified `hub`)
  got the brochure skeleton. Census SQL:
  `bugfix_107_homepage_skeleton/RUNBOOK_homepage_skeleton.md` §1.
- **Full research notes with file:line citations:**
  `bugfix_107_homepage_skeleton/NOTES_homepage_skeleton.md` — includes the
  three-vocabulary `site_type` problem (SPEC-010) your plan's
  "vocabulary extension" flag will hit, and the LANDMINES/WRONG_CALLS entries
  for this exact surface (`pages.sections` is a cache; `sections=[]` is a
  positive statement; built compositions are force-preserved).

**One planning input worth weighing for 4.1:** constraining only the planner
prompt leaves both Go-side fallbacks intact — a gap-planned or
sibling-synthesised page will still get `hero > … > call-to-action` after the
prompt is fixed. The two fallbacks are `apply_gap_plan_action.go:953-978` and
`load_page_sections_from_spec_action.go:281-370`.

No reply needed; this file is the handoff.
