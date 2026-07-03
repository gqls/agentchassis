# PLAN — Sections dropped between page_components and the deployed page

**Owner strand:** vonc.com index (Spark) — provocation-card + lobby-grid
**Opened:** 2026-07-03
**Status:** DIAGNOSIS (cause not yet confirmed — do not fix until confirmed)
**Related:** RUNBOOK_section_assembly_drop.md, RUNBOOK_phase2_provocation_js.md (App F/G),
NOTES_provocation-card.md, NOTES_lobby-grid.md

---

## What we're doing
Finding out why two section instances that exist in `page_components` (as `deployed`,
with fresh rendered HTML) do not appear in the deployed `index.html`, then fixing it
structurally so all planned sections reach the live page.

## Problem
`page_id = b4d24f8e-fccd-49df-9dad-aa56a0b20a68` (index) has SIX `page_components`
(2026-07-03 rebuild, all `build_status='deployed'`, rendered_html @13:15):
hero(1), provocation-card(2), gauntlet-cta(3), brief-explanation(4), lobby-grid(5),
system-stats(6). The deployed HTML on Backblaze contains only FOUR — positions 1,3,4,6.
**provocation-card(2) and lobby-grid(5) are absent from the deploy.**
These two are the Mode-B (empty input_schema, `<no value>` in template), inline-`<script>`
sections. provocation-card is the daily-provocation card (the site's centerpiece). It WAS
in the 2026-07-02 deploy (as a truncated shell); it is gone from 2026-07-03.

## Flow context (how we got here)
regenerate brief-explanation → deferral root-cause + fix (CTA spec + optional
illustration) → provocation-card truncation repair (close inline <script>) → full index
rebuild: needs_page → build-dispatch-loop → page-build-handler → page-content-writer
(render each ready section + compile_page) → validate_content → save_page_sections
(writes page_components) → page-rerender `deploy_page` (assemble from stored components) →
GitHub → GitHub Actions → Backblaze S3.
The drop is in the assemble+deploy tail (content-writer compile OR page-rerender assemble
OR validate/clean). The three content sections rendered correctly; only the two Mode-B
interactive sections vanished.

## What we know
- page_components: 6 rows, all `deployed`, fresh rendered_html @13:15.
- Deployed HTML: 4 sections (pos 1,3,4,6). Missing pos 2 (provocation-card), 5 (lobby-grid).
- The 2 missing share: Mode-B (empty schema, `<no value>` template), inline `<script>`.
- 2026-07-02 deploy included provocation-card; 2026-07-03 does not.
- page-rerender `deploy_page` is documented as "assemble page from stored components and
  deploy to git" (from the page-build-handler agent definition).

## Hypotheses (to CONFIRM, not assume)
- **H1 — assemble filter.** page-rerender's assemble action selects/filters
  `page_components` on a column or property the two dropped rows share (candidates:
  `<no value>` present, empty/near-empty content, `slot_name`, `parent_instance_id`,
  `schema_mode`, missing `deploy_commit`, an explicit interactive-section skip).
- **H2 — wrong source.** The deploy uses the content-writer's compiled `page_html`
  (which may have had 4) rather than re-assembling from `page_components` (6).
- **H3 — compile/render skip.** compile_page or render_component dropped the two Mode-B
  sections (their render produced empty/invalid output), while save_page_sections still
  wrote/updated their rows — so page_components=6 but compiled HTML=4.
- **H4 — validate/clean strip.** validate_content's `clean_html` removed the two sections
  (e.g., detected `<no value>` / empty and stripped them).

## Diagnostic approach (Phase 0 — see runbook for exact SQL/commands)
- **D1** Compare the 6 page_components rows across filter-candidate columns
  (build_status, slot_name, parent_instance_id, schema_mode, deploy_commit, content_hash,
  locked_at, LENGTH(rendered_html), contains `<no value>`, contains `data-component`).
  Look for what distinguishes the 2 dropped from the 4 kept. CHEAPEST, run first.
- **D2** Read the 2026-07-03 rebuild's `deploy_result` (site_work_items.result) — what
  page-rerender reported assembling/deploying (section count, any skip logging).
- **D3** Get page-rerender's workflow (agent_definitions type='page-rerender') → identify
  the assemble action + its source/filter config.
- **D4** Read the assemble action's Go code → find the query/filter that selects sections.
- **D5** Confirm whether the 2 dropped rows' rendered_html still contains `<no value>`
  (Mode-B) — if the assembly drops `<no value>` sections, that is the filter.

## Fix approach (PROVISIONAL — depends on the confirmed cause)
- If H1/H4 (a filter/clean drops Mode-B / `<no value>` / empty / interactive sections):
  the section drop is a genuine defect — a section legitimately present in the plan and
  page_components must not be silently removed. Fix the filter in the Go action so it
  keeps such sections (a shell with a `data-component` + selectors is valid — the runtime
  loader fills it). ALTER the existing action; do not create a new one.
- If H2 (wrong source): make deploy assemble from `page_components` (the stored,
  authoritative set), matching the documented behaviour.
- If H3 (compile/render skip): fix render_component/compile so Mode-B shells compile
  (they must emit the shell markup even with empty fields).
- Orthogonal but related: properly building provocation-card + lobby-grid OUT of Mode-B
  (real schema, or a runtime-feed contract) would also make them robust — tracked
  separately (see NOTES for each); it is not a substitute for fixing a wrongful drop.

## Success criteria
- All six index sections present in the deployed HTML (provocation-card + lobby-grid back).
- provocation-card's `.pc-*` shell present so the Path-2 loader fills it in-browser.
- No regression to hero / gauntlet-cta / brief-explanation / system-stats.
- The mechanism (why they dropped) is understood and recorded — not patched blind.

## Guardrails
- Confirm the cause (D1–D4) before any fix. No fix into an unconfirmed hypothesis.
- Reuse/alter the existing assemble action; keep the workflow thin, logic in the action.
- Check schema before any SQL. Do not treat a 0-row/empty result as decisive without
  checking the query.
