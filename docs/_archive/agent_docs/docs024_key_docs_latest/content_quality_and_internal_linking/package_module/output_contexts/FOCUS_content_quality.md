# FOCUS — Content Quality (gamesdesign.co.uk)

**Status:** current as of 2026-06-09. Source of record for defects: `CATALOGUE_gamesdesign_post_sync_fix_defects(9).md`. Companion to `FOCUS_internal_linking.md` — the hero-CTA defect spans both.

---

## Scope
Whether the words and per-component data on built pages are correct, on-brand, and complete: CTA text, hero/section copy, list-hub card titles, footer brand/contact, tool/game descriptions. Distinct from **design fidelity** (palette/fonts/layout reproduction — see the April design docs, stale-risk) and from **internal linking** (link *destinations* — see `FOCUS_internal_linking.md`), though the hero CTA sits on the content/linking seam.

## The defects (CATALOGUE)
1. **Hero CTA text↔destination mismatch, site-wide** — e.g. "Browse Tools" pointing at `/contact.html`. Text is content; destination is linking. **Lead item.** (Field-vs-template question: `FOCUS_internal_linking.md` Q1.)
2. **Guide copy is tool-flavoured** — guide heroes/CTAs describe simulators ("Launch RNG Simulator"). Option to give guides real embedded interactive demos (`019_tool_library`/`020_tool_lifecycle`) if simple.
3. **Brand-suffix in card titles** — "… - GameDesign.uk" / "| GameDesign.uk" leaking into list-hub card titles.
4. **Empty footer brand-tagline + contact** — footer brand/tagline blank; contact mailto/phone empty.
5. **Empty tool description** — at least one tool card (TTK) with no description.

## The machinery (reuse — STEP ZERO)
1. **`validate_page_content.go`** — the content validator and the gate. Returns an error on blockers (placeholder text, unrendered templates, empty required sections, cross-site contamination); the page-build-handler routes `validate_content` `error_step → mark_needs_review` → `needs_human_review`. A content fix must pass it. (This is Mode 2 of the silent-completion work, already confirmed fixed.)
2. **Auditors + finding routing** — `PLAN_design-note-recommendation-specialists.md` (PROPOSED; verify what's actually deployed). Auditors (`content-quality-auditor`, `visual-design-auditor`) emit findings; the plan classifies each:
   - **bug** (factually broken: placeholder, broken/phantom link, empty section, contamination) → `content_rewrite` → page-build-handler; the validator catches regressions.
   - **gap** (missing content/page) → `needs_content_page` (P9 already routes this).
   - **recommendation** (tone/branding/CTA-strategy opinion) → a specialist that returns apply/dismiss/escalate; `identity-advisor` for contact/email, `component-template-fixer` for CTA, tone/differentiation specialists for the rest; `sites.approval_mode` gates whether recommendations auto-apply.
   - **The hero-CTA phantom destination is a `bug`** (broken link), not a recommendation — so it routes to `content_rewrite`/`component-template-fixer`, not HITL. The plan notes `component-template-fixer` "already handles CTA fixes" — verify and extend rather than build new.
3. **Routing reality check.** P9 fixed `gap → needs_content_page`. The three-way `finding_type` classification and the specialist agents are PROPOSED, not confirmed built. Before relying on `identity-advisor` / `component-template-fixer` / `sites.approval_mode`, confirm each exists in `agent_definitions` / schema.

## Work order (next chat)
1. **Hero CTA (lead — both content and link).** Read June-02 handoff + `FOCUS_internal_linking.md`; settle field-vs-template (Q1); fix destinations to real pages and text to match; reuse `component-template-fixer`'s existing CTA handling rather than writing new resolution.
2. **Footer/contact + brand-suffix titles + empty tool description** — likely component-template/spec data (footer brand-tagline spec; card-title composition that strips the brand suffix; tool-description spec). The polish batch.
3. **Guide copy** — re-flavour guide content away from simulator language; decide whether guides get real interactive demos.

## Relationship to other docs
- Internal linking: `FOCUS_internal_linking.md` (CTA destinations, phantom-link validation, nav, hubs).
- Design fidelity (separate axis, stale-risk): `FOCUS_design_and_styling_adoption_problems.md` + `..._WORK_PLAN_v2.md` (header/footer/layout reproduction); `PHASE_4_4_cleanup_summary.md` (theme composition).
- Defect source of record: `CATALOGUE_gamesdesign_post_sync_fix_defects(9).md`.

## Open questions
- Which auditor/finding-routing pieces are actually deployed (`finding_type`, `identity-advisor`, `component-template-fixer`, `sites.approval_mode`)?
- Does `validate_page_content` already flag the brand-suffix / empty-contact / empty-description cases, or do they slip through silently?
- Hero CTA field-vs-template (shared with `FOCUS_internal_linking.md` Q1).
