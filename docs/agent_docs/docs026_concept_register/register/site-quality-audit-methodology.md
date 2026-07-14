# Register — site-quality-audit-methodology

2 concepts, consolidated from 4 raw extractions (2 unique blocks, each doubled by
an input-file duplication artifact) across unit U13.

### SQAM-001 — Three-way split quality-gap diagnostic method (stuck / poor / out-of-scope)
- **status:** aspirational
- **status-evidence:** Framed as the mission-defining method for the whole runbook: "measure which legs never dispatched, which dispatched but produced poor output, and which were never in scope at all, then fix in that order."
- **what:** A triage discipline for closing the gap between "a site deploys" and "a site is best-in-class for its vertical": split every quality shortfall into (A) dispatched-but-stuck-or-never-delivered (chrome/design/imagery — fix first), (B) delivered-but-poor (content depth, link integrity — fix second), and (C) never-in-scope (feeds/news/graphics/games never planned for — planning-criteria fixes, not build bugs — fixed third). Explicitly evidence-first: "0-rows not decisive until the query is checked." This is the audit method that the Site Quality Programme (SQ-001) applies operationally.
- **sources:** dartsonline.com_site_quality/RUNBOOK_site_quality.md#THE TASK,#THE THREE-WAY SPLIT
- **relations:** Site-chrome rendering gap (SQ-002); Baseline mechanical quality measurement methodology (SQAM-002); Site Quality Programme (SQ-001)
- **verify-later:** RUNBOOK_builder_route.md §B0-B5 (sibling doc, out of this scope)

### SQAM-002 — Baseline mechanical quality measurement methodology
- **status:** deployed
- **status-evidence:** A concrete, dated (2026-07-06) measured table exists: per-page bytes/nav/img/svg/script/css-var-refs/stylesheet-links counts for four rendered pages.
- **what:** A cheap, deterministic, pre-LLM measurement pass over a site's rendered HTML output (byte count, `<nav>` count, `<img>` count, `<svg>` count, `<script>` count, CSS custom-property reference count, stylesheet link count) used to form falsifiable hypotheses about missing pipeline stages before spending any LLM audit budget.
- **sources:** dartsonline.com_site_quality/RUNBOOK_site_quality.md#MEASURED BASELINE
- **relations:** Site-chrome rendering gap (SQ-002); Three-way split quality-gap diagnostic method (SQAM-001)
- **verify-later:** repeatability of this measurement as a scheduled/automated check vs one-off manual pass
