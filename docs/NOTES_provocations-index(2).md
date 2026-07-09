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


---

## STATE as of 2026-07-07 — PAGE LIVE (gate passed); archive list awaiting Step 4
**Working:** /provocations/index.html is DEPLOYED and live (200): built by item eb3cdf42 (~5 min —
the first real build after ten silent no-ops), deployed_at 2026-07-06 16:38:00; one section instance
(provocations-archive-list, rendered_len 7455) with marker + clone template + empty state all
present in the rendered HTML; snippets.js referenced in this page's head; content-writer header copy
in voice (eyebrow "THE RECORD", title "Every <em>provocation</em> on record.", subtitle "...The
positions are permanent. The splits don't lie."); empty state "Nothing filed yet. The first
provocation drops soon."; CTA /index.html "Enter today's Arena" (static fallback working); zero
<no value>. Every primary CTA on the site now resolves here.
**Not yet:** the archive LIST is empty by design until Step 4 (loader js_snippet + asset-renderer
bundle + `archive` object in /data/provocations.json) — explicit 4.1–4.4 procedure in
RUNBOOK_phase2_provocation_js at YOU ARE HERE.
**Unblock history (one line):** plan row (wrong store for the build path) → pages.sections UPDATE
(the actual gate: load_page_record returns pages.sections; empty ⇒ complete_error silent success) →
component 70d6662a with generation-time guards → build eb3cdf42.
**Bookkeeping observation:** the build set deployed_at but left last_built_at NULL — audit-list item.
`Categories:` silent-noop-success (resolved for this page), planning-gap (resolved for this page),
empty-shell/mode-b-template (resolving via Step 4 runtime-fill)


## 2026-07-07 — Step 4 VERIFIED in browser; template ghost-row defect + CTA circularity found
Archive FILLS: 8 rows (5 Jul → 28 Jun) with dates/titles/teasers/dot-stats; empty state hidden;
index unaffected; bundle = 3 active snippets; js_len 3281 (source 3287, ±6 paste drift — bundle +
browser authoritative). DEFECT: the hidden clone template renders as an empty dot-row — author
`display: grid` beats the `hidden` attribute (guide entry added). Fix pack:
fix_archive_template_display.sql + 085 rerender (make_085_rerender_provocations.sh copies the index
trigger). DESIGN QUESTION (user): "Enter today's Arena" → /index.html; no arena page exists, but the
Gauntlet tool DOES (/tools/gauntlet/index.html, deployed, live js). CTA graph circular (hero →
archive; archive → home; gauntlet-cta → archive). Option A (recommended): retarget archive CTA to
the Gauntlet (SQL ready). Broader CTA-map pass queued (baked URLs need section rebuilds).
`Categories:` css-specificity (new), cta-graph (new)


## 2026-07-07 ~16:30 — CTA decision: OPTION B; 085 rerender proven; display fix still pending
User chose B: leave the CTA graph (archive → /index.html; hero → archive) until the real take-filing
arena exists; retarget SQL parked. 085_rerender-provocations-vonc.sh created via sed-copy of the
index trigger and RUN (corr 0ea43016) — page redeployed (deployed_at 2026-07-07 16:20:17), proving
the page-level rerender path for this page. HOWEVER rendered_len unchanged at 7455 ⇒ the
ghost-row display fix was NOT applied before the rerender (fix adds ~110 chars) ⇒ dot-row presumably
still live. Remaining: fix_archive_template_display.sql → 085 again → verify (rendered_len ~7565;
live grep for the display:none rule = 1; dot-row gone). last_built_at confirmed dead through both
build and rerender paths (audit list).
`Categories:` css-specificity (fix pending), cta-graph (decision: parked, Option B)
