# FOCUS — Content Quality (gamesdesign.co.uk)

**Status:** current as of 2026-06-10. Source of record for defects: `CATALOGUE_gamesdesign_post_sync_fix_defects(9).md`. Companion to `FOCUS_internal_linking.md` — the hero-CTA defect spans both.

---

## Scope
Whether the words and per-component data on built pages are correct, on-brand, and complete: CTA text, hero/section copy, list-hub card titles, footer brand/contact, tool/game descriptions. Distinct from design fidelity (palette/fonts/layout) and from internal linking (link *destinations* — `FOCUS_internal_linking.md`), though the hero CTA sits on the content/linking seam.

## The defects (CATALOGUE) and status
1. **Hero CTA text↔destination mismatch, site-wide** — **addressed (Step 1).** The destination half is fixed at source: the `resolve` `pages`-case fabrication is gone, the `hero`/`call-to-action` schemas drop the phantom fallbacks (`on_missing: skip_field`), and the templates gate each button on a resolved url. So an unresolved CTA now renders no button instead of a phantom. The text half — making the LLM-written label match an intent-appropriate destination — is the `internal-link-resolver` agent's job (Step 3), which will choose the destination per page rather than hardwiring contact/services.
2. **Guide copy is tool-flavoured** — open. Guide heroes/CTAs describe simulators ("Launch RNG Simulator"). Option to give guides real embedded interactive demos (`019_tool_library`/`020_tool_lifecycle`) if simple.
3. **Brand-suffix in card titles** — open. "… - GameDesign.uk" / "| GameDesign.uk" leaking into list-hub card titles.
4. **Empty footer brand-tagline + contact** — partially related to Layer 1b. Footer legal links are now data-driven; footer brand-tagline/contact mailto/phone still to address (site metadata: `company_name`/`tagline`/`email`/`phone` — see `check_component_standards.checkMissingSiteMetadata`, routes to `site-metadata-fixer`).
5. **Empty tool description** — open. At least one tool card (TTK) with no description.

## The machinery (reuse — verified this round)
1. **`validate_page_content.go`** — the content validator and deploy gate. It always had a `validateInternalLinks` check, but emitted missing targets as a single **non-blocking warning** (lumping true phantoms with planned pages), and never inspected `site_components`. Now consolidated onto `datahelpers/links.go` (shared with the post-deploy audit): flags `phantom_link` (no `pages` row) and `empty_internal_href`, both non-blocking per policy; planned pages tolerated. Placeholder/template/contamination/email checks unchanged (those remain blockers/errors).
2. **Auditors + finding routing — reality check (corrected):**
   - `content-quality-auditor`, `visual-design-auditor`, `design-audit-agent` exist (analysts).
   - `component-template-fixer` exists but **explicitly punts on CTAs** (`cta_improvement`/`cta` → `fixed:false, action:"needs_review"`); it does programmatic HTML/CSS fixes (`inject_nav_flex_css`, `responsive_fix`, `remove_element`, `align_slot_name`), not link/CTA resolution. So the PLAN's "already handles CTA fixes" was wrong; there was no CTA resolver to reuse — hence the dedicated `internal-link-resolver` (Step 3).
   - `identity-advisor` does **not** exist. `sites.approval_mode` does **not** exist. The three-way `finding_type` classification and those specialists are PROPOSED, not built.
3. **`check_phantom_internal_links`** (new) — post-deploy audit for broken/empty internal links; routes page-body issues to `page-build-handler` (a rebuild re-runs build-time link resolution), header/footer to `nav-link-fixer`. Inert until enabled.

## Work order (next)
1. ~~B4/B5 empty-href hubs~~ — **done** (2026-06-11): `query.section_index_for:<type>` verb + schema repoint + template gate. In the deploy batch.
2. ~~Step 3 — `internal-link-resolver`~~ — **done** (2026-06-11): agent + writer wiring + `unresolved_cta` HITL signal. In the deploy batch. (Closes the destination half of defect 1; the LLM CTA text now coexists with a real destination — text↔destination *coherence* polish can ride the future intent-matching upgrade.)
3. **Readopt gamedesign.uk → gamesdesign.co.uk** (after RUNBOOK Part 3.1 verification) — from-scratch acceptance for the linking work AND the fresh baseline for this package. The open defects below WILL recur on the readopted site (adopt-path, untouched) — recatalogue against the new site_id.
4. **Footer brand-tagline + contact, brand-suffix card titles, empty tool description** — site metadata via `site-metadata-fixer`; card-title composition to strip the brand suffix; tool-description spec.
5. **Guide copy** — re-flavour away from simulator language; decide whether guides get real interactive demos.

## Open questions
- Guide copy: re-flavour only, or embed real interactive demos in guide pages (feasibility per `019_tool_library`/`020_tool_lifecycle`).
- Does `validate_page_content` flag the brand-suffix / empty-contact / empty-description cases, or do they slip through? (It does not currently — those are content/spec issues, not link/placeholder issues.)
