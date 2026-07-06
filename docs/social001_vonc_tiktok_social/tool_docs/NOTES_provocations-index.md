# NOTES — provocations-index (page)

Append-only provenance record for the `provocations-index` page (the Provocations
Archive — the arena's destination). Dated states of what works and what doesn't.
`Categories:` uses the shared taxonomy (TOOL_DOCS_convention.md).

**page id:** e4b3b195-919f-45ad-854e-201d3e846ea8
**url:** /provocations/index.html  •  **title:** "Provocations Archive | Spark"
**page_type:** section-index  •  **site:** vonc.com

---

## STATE as of 2026-07-04
**Working:** nothing user-visible — the page does not exist on B2 (404 NoSuchKey).
**Linked from:** header nav ("Provocations"), hero "Enter the Gauntlet",
site_specs.cta.primary_url (gauntlet-cta + system-stats CTAs), lobby-grid "Enter the
Arena" + all six arena card urls, provocation-card primary CTA. Every primary action on
the site dead-ends here.

**NOT working / the finding (2026-07-04):**
- pages row: build_status='planned', updated_at = creation instant (2026-06-22 17:13:08)
  — planned in the original build, never built.
- The current site plan has ZERO site_plan_sections for page_name='provocations-index'
  (0 rows; query shape proven on the index page; spelling-confirm via DISTINCT page_name
  pending). pages.sections fallback = '[]'.
- SEVEN work items for the page, ALL 'complete', no errors: needs_page:provocations-index
  (06-22 17:13:22, +14s after page creation), 2× manual needs_page rebuilds (06-26),
  4× page_rerender (06-23 → 07-01 12:53). The page row was untouched by every one —
  the handler exits on the zero-sections path before any step that writes the page;
  the rerender skips deploy when there are no page_components. A success status masked
  a no-op for two weeks. Full route + prevention: debugging guide §9 "Page build
  completes having built nothing"; runbook App I.
- ROOT CAUSE two layers: (a) planning gap — the planner emitted the page but no sections
  (likely systemic for section-index/archive pages: no component vocabulary for
  dynamic-list pages); (b) build + rerender treat "nothing to do" as success.

**Fix path (pending):** give the page sections — realistically a header/hero + an archive
LIST component (kind=dynamic, feed from provocations.json — Phase-3 family; ties to the
section-descriptor + loader-builder design and the complex-tool loop in the parallel
chat) — then needs_page build → deploy → 404 gone. Guards/invariants from the guide entry
are the framework-side prevention.
`Categories:` planning-gap (new), silent-noop-success (new)
