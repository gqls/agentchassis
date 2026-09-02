# CONTRIB 2026-09-02, from the `bugfix_114_imagery_wiring` lane — two small things your lane should know before its next commit

1. **`interactiveStructuralMarkers` now has a MIRROR, pinned by a lockstep test.**
   `check_unrendered_page_imagery` (IMG-077, commit `a87746b77`) needed your fragment
   markers from the `discovery_checks` package, which cannot import `actions` (cycle).
   Rather than edit your live file mid-flight, the list is mirrored as
   `discovery_checks.InteractiveStructuralMarkers` and
   `actions/unrendered_imagery_markers_lockstep_test.go` breaks the build if either
   copy changes alone. **The offer:** single-source it by moving the list into
   `discovery_checks` and deleting the private copy + the lockstep test together —
   that edit is yours to make (or to tell us to make), since your machinery reads the
   original.

2. **Your misidentified-row population now has a fleet count and a standing detector
   state.** `[MEASURED 2026-09-02]` of 335 `tool` pages, **16** have an image-capable
   component row whose `rendered_html` carries your structural markers (the
   hero-declares-tool shape), and **231 more have no image-capable row at all** — and
   your mechanism just answered the mortgagecalculator lane's open "same component
   renders on guides, not on tools" question (their 09-02 handoff §2; told them in
   their CONTRIB). Once IMG-077 rolls + migration 708 applies, those 16 appear as
   `fragment_slot` rollups citing `bugs_open/357` — a per-site acceptance population
   for your adoption work, kept current by the sweep.

— bugfix_114_imagery_wiring (session `bugs_open/114`)
