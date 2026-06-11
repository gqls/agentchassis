# RUNBOOK — deploy the linking phantom fixes (Step 1 + Layer 1b + Layer 2)

Applies the shipped changes that stop hero/CTA and header/footer phantom links at source, plus the shared link-validation layer. Order matters: the Go `resolve` fix must be live before the schema SQL, or the schema change alone leaves the phantom (the resolver still fabricates a truthy `/contact.html`).

Site: gamesdesign.co.uk → `site_id := (SELECT id FROM sites WHERE domain='gamesdesign.co.uk')` (changes on teardown; resolve fresh).

## 0. Pre-flight
- Fresh `\d content_components`, `\d pages` (confirm columns before SQL).
- Confirm the Go files compile in your tree (assistant validated brace/paren balance only).

## 1. Code (one chassis image roll, `ai-persona-system`)
Place all of:
- `plan_sections_action.go` — `resolve` `pages` case (no fabrication).
- `datahelpers/links.go` — NEW (validator + audit import it; must be present first).
- `validate_page_content.go` — gate on datahelpers.
- `check_phantom_internal_links.go` — audit on datahelpers (self-registers; stays inert until enabled in step 5).
- `render_site_components_action.go` — header/footer `ContentData` from real pages.

Roll the chassis image. The Go changes are interdependent (datahelpers is imported by the other two); ship together.

## 2. Schema/template SQL (snapshot-first)
Run, in order, checking each reports the expected row counts and the verification SELECTs:
1. `step1_hero_cta_phantom_fix.sql` — snapshots `hero`,`call-to-action`; schema + template edits. Each template `replace` must report `UPDATE 1`; verification row must show `tmpl_has_contact_literal=f`, `tmpl_has_services_literal=f`, `tmpl_has_features_literal=f`, `tmpl_has_tightened_gate=t`. If a literal is still `t`, the stored whitespace differed and that replace no-op'd — fix the match string, do not assume.
2. `layer1b_header_footer_phantom_fix.sql` — snapshots `header-bold-gradient`,`footer-4-column`; CTA gate + data-driven footer legal. Verification must show `header_cta_gated=t`, `footer_has_privacy_literal=f`, `footer_has_terms_literal=f`, `footer_legal_data_driven=t`.

(Step 2 must run AFTER step 1's image is live, per the ordering note above.)

## 3. Re-render
The schema/template/data changes only affect future renders.
- Force re-render `site_components` (slots header, footer) for the site — the `render_site_components` action with `force_rerender` (the same step `nav-link-fixer`'s workflow uses).
- Re-render the hero/CTA pages (every page carrying a `hero` or `call-to-action` section) so the corrected templates + resolved data regenerate `page_components.rendered_html`. Use the normal page-build/rerender path (e.g. `build_status='needs_rebuild'` reset, or re-issue the page through page-build-handler).

## 4. Verify (re-run the audit dry-run)
Run the dry-run SQL from the phantom-link investigation (or enable the check briefly in a non-loop context). Expected after re-render:
- Page-body `hero … /contact.html` and `… /services.html` rows: **gone** (buttons dropped — no resolver agent yet, so heroes have no CTA button; that is correct-or-absent, not silent — the absence is build-time visible and Step 3 restores them).
- `site_component header /contact.html`: gone if `contact-index` is reachable via footer nav → header CTA now points at the real contact page; otherwise the header CTA is omitted (no phantom).
- `site_component footer /privacy.html`, `/terms.html`: **gone** (no real legal pages → footer renders no legal links).
- Remaining: the three `empty_internal_href` rows on `tool-list`/`guide-list`/`game-list` (B4/B5 — not addressed here; see PLAN).

## 5. Do NOT enable the audit check yet
`phantom_internal_links` self-registers but only runs when added to a discovery agent's `checks` array. Leave it out until:
- the `internal-link-resolver` agent exists (page_component findings route to it), and
- you are ready to re-enable `improvement-sweep`.
Otherwise findings accumulate with no handler for the page_component surface.

## Rollback
- Templates/schema: restore from the snapshot tables (`content_components_bak_cta0610`, `content_components_bak_navfix_0610`) by `UPDATE … SET html_template/input_schema = bak.… FROM <bak> WHERE name=…`.
- Code: roll back the chassis image tag.

---

# Part 2 — B4/B5 hub links + Step 3 resolver (same batch)

## Code (same chassis image as Part 1)
Add to the image:
- `section_index_for.go` (queryresolve) + ONE switch case in `queryresolve.Resolve`:
  `case "section_index_for": return resolveSectionIndexForType(ctx, db, req.SiteID, arg, logger)`
- `resolve_internal_links_action.go` (actions; includes the unresolved_cta emission) + registry entry:
  `"resolve_internal_links": { Handler: ResolveInternalLinksAction, Category: "content", IsLocal: true }`
- `check_phantom_internal_links.go` — UPDATED routing (page_component -> page-build-handler/content; site_component -> nav-link-fixer/build). Still inert until enabled.
- `plan_sections_action.go` — the CURRENT file + the Step 1 one-block fix (this copy supersedes any earlier patch).

## SQL — order matters, all after the image is live
1. `b4_b5_hub_links_schema.sql` — snapshot + repoint the 3 list `cta_url` sources to `query.section_index_for:<type>` (verb must exist first or the source resolves as unknown).
2. `b4_b5_hub_links_template_gate.sql` — gate the 3 Browse-All anchors (each `UPDATE 1`; verify `browse_all_gated=t` ×3).
3. `internal_link_resolver_agent.sql` — the agent row (NOT EXISTS-guarded). Set `image_tag` to the batch's tag before running.
4. `page_content_writer_link_resolver_wiring.sql` — `snapshot_agent()` first (built in), then the chained `jsonb_set` UPDATE; check all 7 verification columns. Rollback: `SELECT revert_agent('page-content-writer');`

## Re-render + verify (extends Part 1 step 4)
- Re-render the pages carrying `tool-list`/`game-list`/`guide-list` (index + hubs) — the 3 `empty_internal_href` rows should clear (Browse-All now resolves to `/tools/index.html`, `/games/index.html`, `/guides/index.html`).
- Rebuild one hero page end-to-end and confirm in the writer's orchestration: `resolve_links` ran, `sections_for_render.sections_ready` populated, hero rendered with `cta_url` = a real hub. Resolver logs: `resolve_internal_links: augmented CTA sections`.
- gamesdesign has all three hubs, so `unresolved_cta` items should be ZERO; any that appear are real findings.

## Enabling the audit check (later, deliberate)
Preconditions now reduce to: the batch is live + `improvement-sweep` re-enable decision. Routing targets (`nav-link-fixer`, `page-build-handler`) both exist. Add `phantom_internal_links` to the discovery agent's checks array; watch the first sweep's work items before re-enabling autonomous processing.

---

# Part 3 — post-deploy verification, then readopt

Baseline recorded pre-roll (2026-06-11): `internal-link-resolver` and `page-content-writer` both `v1.0.1060`; `unresolved_cta` count 0.

## 1. Post-roll checks (existing site — isolates each fix before the big test)
Do these on the EXISTING gamesdesign.co.uk first. Six code changes and four SQLs shipped together; if a fresh readopt misbehaves, these are what make the failure attributable to a layer.

1. Tags moved together: `SELECT type, image_tag FROM agent_definitions WHERE type IN ('internal-link-resolver','page-content-writer','research-agent');` — all on the new batch tag. If `internal-link-resolver` lagged, its spawned pods run the old image (no `resolve_internal_links` action) and every build silently takes the fallback path.
2. Site components re-render (layer1b Go): header CTA → real contact page; footer legal links gone (no real legal pages). The intermediate-state phantoms clear here.
3. List pages re-render (B4/B5): the three Browse-All buttons render with `/tools/index.html`, `/games/index.html`, `/guides/index.html`. Verb log lines: `queryresolve: resolved section_index_for (by area|by url prefix)`.
4. ONE hero page rebuild end-to-end (Step 1 + resolver wiring): orchestration shows `resolve_links → select_sections → process_sections_loop`; `sections_for_render.sections_ready` count = plan count; resolver log `resolve_internal_links: augmented CTA sections` with `hub_count=5`; hero renders both buttons pointing at the top two content hubs.
5. Dry-run audit SQL: zero phantom rows, zero `empty_internal_href` rows. `unresolved_cta` still 0.

## 2. Readopt (gamedesign.uk → gamesdesign.co.uk) — the end-to-end acceptance
After Part 3.1 passes. The readopt is the only test of the FROM-SCRATCH path: fixed `plan_sections` resolver during initial planning, the resolver agent on first writes, hub links on first list renders, clean header/footer from the start. It also produces the fresh, current-pipeline baseline the content-quality package needs (the existing site mixes legacy-pipeline and patched state).

Why it should work (evidence, not hope): planning creates all `pages` rows before content writing (the writer receives `db_sync.pages`), and during the ORIGINAL build the list cards resolved via the same `status IN ('active','deployed')` filter the hub queries use — so hubs are visible to the resolver when the first hero is written.

Watch during the readopt:
- Per page: `resolve_internal_links: augmented CTA sections` (resolver ran; `unresolved=0`).
- `SELECT count(*) FROM site_work_items WHERE item_type='unresolved_cta' AND site_id=<new>;` — stays 0.
- After deploy: the dry-run audit on the NEW site_id — zero phantom/empty rows anywhere, including site_components, on a virgin build.
- EXPECTED recurrences (not linking regressions): the open content-quality defects — brand-suffix card titles, tool-flavoured guide copy, empty tool descriptions, footer brand/contact metadata — live in the adopt/content path and were not touched. They re-manifesting is the input to the next work package.

Note: teardown changes site_id — old-site work items/baselines orphan; re-resolve site_id in every query.
