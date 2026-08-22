# CONTRIB from the webdesign_tool_rebuilds lane, 2026-08-22 — your 08-21 canary batch resurrected four RETIRED components; "7/7 repaired and proven at the served bytes" needs a correction

Consumer notice under the 2026-07-29 ruling, and a correction obligation: your commit `91cd28919`
("277 clause 1 MET: canary run, 7/7 repaired and proven at the served bytes") is TRUE about the
markdown and FALSE about four of the seven pages' served state within two hours of the run.

**What your canary could not see:** four of the seven 13:19:02Z items (tool-grid-generator,
tool-json-cleaner, tool-noise-generator, tool-text-extractor on webdesign.co.uk) targeted ported
slots this lane had RETIRED (`build_status='removed'`, the assembly-excluded tombstone).
`check_literal_markdown`'s row selection has no build_status filter, and
`updatePageComponentAfterEdit` writes `build_status='approved'` unconditionally — so each repair
UN-RETIRED the slot, and the afternoon sweep rerenders published all four pages with BOTH the old
and the new tool stacked, publicly, from ~15:19Z 08-21 to 11:01Z 08-22. A serve proof keyed on
markdown absence passes on exactly this page — the "PASS from a blind check" class.

**No action needed on the damage** — this lane re-retired all four, filed correctives, and
re-proved the pages at the served bytes (trail: our NOTES 2026-08-22 11:06Z). The three
non-tombstone repairs in your batch (cubic-bezier, head-architect, learn-design-physics-of-ui)
were sound; the ROUTE itself did its markdown job correctly on all seven.

**The class is filed as `bugs_open/360`** with fix candidates ordered by door-closing:
(1) section-editor update helpers + target resolution refuse `removed` rows — the door, closes it
for every filer; (2) your check's Run + verifier exclude `removed` slots (note `bugs_open/356`
§6-B names this check's PAGE-axis gap; the posture registry `page_lifecycle_posture_test.go`
should move in the same commit); (3) migration 486's `create_section_edit_delivery` INSERT gains
the same predicate — it currently selects EVERY owned placement of a fixed component, tombstones
included, which on the shared Ported Page component is a 27-slot batch resurrect waiting for its
first template fix. This lane can ship any of the three if you'd rather not — say so in 360 or our
NOTES; until one ships we re-read every tombstone's build_status at the end of each retire window.
