
<!-- SOURCE: U13_docs024_small_dirs.md -->
### Three-way split quality-gap diagnostic method (stuck / poor / out-of-scope)
- **category:** NEW:site-quality-audit-methodology
- **status-signal:** aspirational
- **status-evidence:** Framed as the mission-defining method for the whole runbook: "measure which legs never dispatched, which dispatched but produced poor output, and which were never in scope at all, then fix in that order"
- **what:** A triage discipline for closing the gap between "a site deploys" and "a site is best-in-class for its vertical": split every quality shortfall into (A) dispatched-but-stuck-or-never-delivered (chrome/design/imagery — fix first), (B) delivered-but-poor (content depth, link integrity — fix second), and (C) never-in-scope (feeds/news/graphics/games never planned for — planning-criteria fixes, not build bugs — fixed third). Explicitly evidence-first: "0-rows not decisive until the query is checked."
- **sources:** dartsonline.com_site_quality/RUNBOOK_site_quality.md#THE TASK,#THE THREE-WAY SPLIT
- **relations:** Site-chrome rendering gap; Baseline mechanical quality measurement methodology; Design/composition work-item emission gap
- **verify-later:** RUNBOOK_builder_route.md §B0-B5 (sibling doc, out of this scope)

<!-- SOURCE: U13_docs024_small_dirs.md -->
### Baseline mechanical quality measurement methodology
- **category:** NEW:site-quality-audit-methodology
- **status-signal:** deployed
- **status-evidence:** A concrete, dated (2026-07-06) measured table exists: per-page bytes/nav/img/svg/script/css-var-refs/stylesheet-links counts for four rendered pages
- **what:** A cheap, deterministic, pre-LLM measurement pass over a site's rendered HTML output (byte count, `<nav>` count, `<img>` count, `<svg>` count, `<script>` count, CSS custom-property reference count, stylesheet link count) used to form falsifiable hypotheses about missing pipeline stages before spending any LLM audit budget.
- **sources:** dartsonline.com_site_quality/RUNBOOK_site_quality.md#MEASURED BASELINE
- **relations:** Site-chrome rendering gap; Three-way split quality-gap diagnostic method
- **verify-later:** repeatability of this measurement as a scheduled/automated check vs one-off manual pass

<!-- SOURCE: U13_docs024_small_dirs.md -->
### Three-way split quality-gap diagnostic method (stuck / poor / out-of-scope)
- **category:** NEW:site-quality-audit-methodology
- **status-signal:** aspirational
- **status-evidence:** Framed as the mission-defining method for the whole runbook: "measure which legs never dispatched, which dispatched but produced poor output, and which were never in scope at all, then fix in that order"
- **what:** A triage discipline for closing the gap between "a site deploys" and "a site is best-in-class for its vertical": split every quality shortfall into (A) dispatched-but-stuck-or-never-delivered (chrome/design/imagery — fix first), (B) delivered-but-poor (content depth, link integrity — fix second), and (C) never-in-scope (feeds/news/graphics/games never planned for — planning-criteria fixes, not build bugs — fixed third). Explicitly evidence-first: "0-rows not decisive until the query is checked."
- **sources:** dartsonline.com_site_quality/RUNBOOK_site_quality.md#THE TASK,#THE THREE-WAY SPLIT
- **relations:** Site-chrome rendering gap; Baseline mechanical quality measurement methodology; Design/composition work-item emission gap
- **verify-later:** RUNBOOK_builder_route.md §B0-B5 (sibling doc, out of this scope)

<!-- SOURCE: U13_docs024_small_dirs.md -->
### Baseline mechanical quality measurement methodology
- **category:** NEW:site-quality-audit-methodology
- **status-signal:** deployed
- **status-evidence:** A concrete, dated (2026-07-06) measured table exists: per-page bytes/nav/img/svg/script/css-var-refs/stylesheet-links counts for four rendered pages
- **what:** A cheap, deterministic, pre-LLM measurement pass over a site's rendered HTML output (byte count, `<nav>` count, `<img>` count, `<svg>` count, `<script>` count, CSS custom-property reference count, stylesheet link count) used to form falsifiable hypotheses about missing pipeline stages before spending any LLM audit budget.
- **sources:** dartsonline.com_site_quality/RUNBOOK_site_quality.md#MEASURED BASELINE
- **relations:** Site-chrome rendering gap; Three-way split quality-gap diagnostic method
- **verify-later:** repeatability of this measurement as a scheduled/automated check vs one-off manual pass
