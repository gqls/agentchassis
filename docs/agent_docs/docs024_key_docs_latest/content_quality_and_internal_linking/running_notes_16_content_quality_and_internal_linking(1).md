# Running notes 16 — content quality & internal linking

Continues from `running_notes_15_skinner_box_and_adoption_sections(9).md` (skinner-box + sections-durability) and `HANDOFF_2026-06-09_sections_durability_and_content_quality.md`. Site: gamesdesign.co.uk. Standing rules unchanged (snapshot before DB writes; fresh `\d` before SQL; STEP ZERO reuse; complexity in Go; user runs all SQL/kubectl/builds; validate Go by brace balance only; no "perfect/critical").

This file is the live running notes going forward.

---

## Carried-forward state

**Deploy-pending (from chat 15 — do before/with the content work):** 2b (`load_page_sections_from_spec` sibling fallback), S1 (`check_sectionless_pages`), S2 (page-build-handler `mark_no_sections`), Fix A (`complete_work_item` guard — prerequisite for S2). Deploy per `RUNBOOK_section_sectionless_durability.md`. FOCUS silent-completion modes 1–3 already fixed in prod; Fix B deferred.

**Pending from June-02 work:** registry entries are now **in** `registry.go` for both `flag_page_image_rebuild` (was already present) and `reconcile_section_data` (added). Both are callable. **Still open:** `reconcile_section_data` has no host — registering it makes it callable, but nothing calls it yet. Wire it into either a periodic loop discovery check or a post-build finalize step in the rerender path (plan-time is too early — listed pages don't exist yet). Until hosted, query-resolvable `needs_section_data` items still sit at `needs_human_review`. Per `HANDOFF_2026-06-02_hero_resolver_and_section_data_reconciler.md`.

---

## Part 1 — internal-linking subsystem mapped (from code, 2026-06-09)

Read `multipage_actions.go`, `site_db_actions.go`, `queryresolve.go`, plus the June-02 handoff. Findings:

- **The "hero resolver" (June-02) is about hero IMAGES**, not CTA URLs — per-page hero/logo images rendered the fallback `/assets/images/hero.jpg`; fix was per-page resolution in `plan_sections` `ensureAssets` + `flag_page_image_rebuild`. The hero-CTA-URL bug is a SEPARATE problem.
- **Phantom `/services.html` has a concrete source:** `multipage_actions.go` lines 310–318 — when nav resolution returns empty, `AssembleMultipageSiteAction` injects a hardcoded brochure nav (`Home/About/Services/Contact` → `/index/about/services/contact.html`). Generic default leaking the phantom.
- **Existing linking machinery to reuse:** `link_registry` via `ExtractAndSyncLinksAction` (per-page link inventory — substrate for a phantom-link check); `buildNavigationFromPages`/`GetNavigationStructure` (real nav from pages); `fixAnchorLinks` (`#anchor`→`/page.html`); `queryresolve` `pages_where_type`/`pages_under_section` (list-hub targets); `upsertPage` writes `slug`/`url`/`nav_label`/`nav_order` (the page URLs of record); `*_index_url` site_specs feed "Browse All" (unpopulated → empty href; inconsistent sources).
- **`med_url_discovery_action.go` is NOT relevant** — business_intel pet-med price scraper (retailer product URLs via Firecrawl). Keyword false-positive; ignore.

Full write-up: `FOCUS_internal_linking.md`.

## Part 2 — content-quality subsystem summarised

`FOCUS_content_quality.md` written. Key points: `validate_page_content` is the gate (blockers → needs_human_review). Finding routing per `PLAN_design-note-recommendation-specialists.md` (bug/recommendation/gap; `component-template-fixer` "already handles CTA fixes"; `identity-advisor` for contact/email) is PROPOSED — verify what's deployed. Hero-CTA phantom destination is a *bug*, routes to content_rewrite/component-template-fixer, not HITL.

## Open questions to settle FIRST next chat
1. Hero CTA: resolvable field or hardcoded in the hero component `html_template`? `SELECT name, input_schema->'fields' FROM content_components WHERE component_level='hero' AND is_active=true LIMIT 5;` (June-02 follow-up #2 analog).
2. Does `link_registry`/`syncLinksToDB` validate link targets, or only record? Read `syncLinksToDB` + schema.
3. Is gamesdesign's nav coming from DB (`GetNavigationStructure`) or falling through to the hardcoded default? Check `nav_data` + rendered header.
4. Which auditor/finding-routing pieces actually exist (`finding_type`, `identity-advisor`, `component-template-fixer`, `sites.approval_mode`)?

## Lead work order
Hero CTA (content + link) → footer/contact + brand-suffix titles + empty tool desc (polish) → guide copy re-flavour. Source of record: `CATALOGUE_gamesdesign_post_sync_fix_defects(9).md`.
