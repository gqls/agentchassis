# NOTES — native rebuild of ported tools (append-only, newest at the bottom)

## 2026-08-15 ~17:00Z — lane opened; mechanism established; safety restore done; pilot filed

- Corrected my own earlier claim to the owner: the tool-generator was NOT "replacing ported
  tools" — `check_missing_tools`/tool-suggester does GAP analysis and ADDS tools. css-unit-
  converter (00:07 today) was an addition; the 63 ported tools had no replacement path at all.
- The one page where native met ported (`tool-ab-test-calculator`) is the cautionary case:
  native slot attached ALONGSIDE (position 2), ported slot left `pending`, and a raw
  `{{.section_heading}}` served to visitors. Junk census across all 4 native slots: ONLY
  ab-test carries a raw tag — the generator's newest output is clean.
- SAFETY RESTORE executed (owner-sanctioned, closes bugs_open/281 Finding A hazard): shared
  `ported-page` wrapper template restored from the poisoned 8,864-char asset-formatter markup
  to v3's pre-poison snapshot (4,664, `{{.body}}` intact); poisoned state banked as v4;
  114 `pending` placements un-flipped to `deployed`. Verified: template 4,664/passthrough=t,
  0 pending. NOTE: the 281 addendum said "restore v1 (77-char seed)" — v1 was NOT the
  pre-poison state (legitimate 08-05→08-14 edits sit between); v3's pre-edit snapshot is.
- Retired ab-test's ported slot (`removed` — the assembly-excluded tombstone, verified in
  code before use) and filed `section_edit:owner-gate:tool-ab-test-calculator-rawtag` to fix
  the raw tag; its reassembly also drops the removed slot from the served page.
- PILOT filed: `add_tool_novel_webdesign.co.uk` → tool-generator, `tool-aspect-ratio`
  (simple; fleet-wide function claim checked empty; same URL, page_id pinned). Next session:
  RUNBOOK — retire, re-render, grade at the artefact, then batch simple tools serially.
- [ASSUMED] the deploy will attach to the existing page by matching name/function, as it did
  for ab-test. If it instead creates a duplicate page/URL, stop and re-plan before batching.
