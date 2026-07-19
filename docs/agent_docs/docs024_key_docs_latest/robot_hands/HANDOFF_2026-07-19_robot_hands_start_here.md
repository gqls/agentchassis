# HANDOFF — robot-hands.com — START HERE (2026-07-19)

**Supersedes `HANDOFF_2026-07-17_robot_hands_site_fixes.md`** (the original
defect list; still worth reading for the R1–R6 framing, but several of its
stated causes were WRONG — corrections below).

Read order for a fresh chat: **this file → `SUMMARY_2026-07-19_…` (read-aloud
state) → `NOTES_robot_hands_site_fixes.md` (technical log, missteps) →
`RUNBOOK_robot_hands_site_fixes.md` (the commands)**.

Site: **robot-hands.com**, site_id `00ff3af5-dad8-4770-9f70-3edc267a3c92`.
Deploy repo: **gqls/sites**, files under `robot-hands.com/` (GitHub-API
commits, no local checkout). DB:
`kubectl -n ai-persona-system exec -i postgres-clients-0 -- psql -U clients_user -d clients_db`

---

## Status of the original six defects

| # | Defect | State (2026-07-19) |
|---|---|---|
| R1 | Dark theme lost, blue brochure chrome | **DONE, verified live** — 0 hits of `#3b82f6` / `site-header--gradient` on index, hub, matchmatrix, about, tools |
| R2 | learning-center-hub yellow on white | **DONE, verified live** — 0 `#fbbf24`, 0 white backgrounds; cards resolve `var(--color-card-bg)` `#1E2535` on `#0F1218` |
| R3 | learning-center URL sprawl | **DONE, verified live** — header links `/learning-center-hub.html`; residue index archived |
| R4 | Tools broken / MatchMatrix "blank" | **PARTLY** — "blank" diagnosed and fixed (was white cards + `--color-primary` used as text colour). **2 of 5 tool pages still 404** — see Next actions |
| R5 | Dead "Load More" | **DONE, verified live** — button gone from hub; fleet fallback flipped to opt-in |
| R6 | 6 of 9 listed articles 404 | **DONE** — 6 rows archived, 12 assets superseded; hub lists exactly the 3 real guides |

## Corrections to the 2026-07-17 handoff (do not re-walk these)

- **The blue header did NOT come from the 2026-07-16 regeneration.** It entered
  between gqls/sites `af0ead8da1` (07-08 22:44) and `78532b8c63` (07-09 09:13) —
  a week earlier. The regen only *spread* it to 37 pages.
- **`component_versions` holds NO snapshots** for header/footer/head. The
  "restore from snapshots" route the handoff proposed does not exist.
- **`hardcoded_section_colors` never rewrote anything.** Both "completed" items
  failed inside their workflow (`WORKFLOW_INVALID … requires a topic`) and were
  stamped complete anyway. It did not cause the white sections; the palette did.
  Filed as a bug — see below.
- **MatchMatrix was never blank** — 6.3KB of visible text throughout.

## Root cause of R1 (the real one)

B7 (2026-07-10) swapped **only** `css_themes.layout_id`. **Three palette copies
stayed light/blue** — `palettes.colours`, `style_collections.color_palette`
(the one component rendering actually reads), `css_themes.color_palette` — and
`site_components` still pointed at **deactivated** components that bake colours
into inline `<style>`. `renderAndStoreSiteComponent` ignores `is_active`. On top
of that, `webdesign-agent` re-rolled the palette on every run because the
`generic_theme` check misfired fleet-wide.

Fix applied: new var()-only `header-theme-chrome` / `footer-theme-chrome`
components, slots repointed, all three palette copies rewritten dark,
`design_intent.palette` pinned. SQL artifacts are in this directory.

---

## Next actions, in order

1. **R4 — build the two missing tool pages.** `tool-matchmatrix` and
   `tool-robot-payload-budget-calculator` are planned but never built:
   `/tools/matchmatrix/index.html` and
   `/tools/robot-payload-budget-calculator/index.html` both **404**, while the
   other three tool pages serve 200. This is user-visible: **five homepage CTAs
   point at those two 404s** ("Search the Gripper Catalog", "Run a MatchMatrix
   Query", and the tool cards). Two open `incomplete_page_group` items already
   track them.
   **Coordinate with the experience_loop workstream** —
   `../experience_loop/HANDOFF_2026-07-17_experience_loop_start.md` covers
   exactly this class fleet-wide (broken tools, tool_acceptance sweeps,
   page-ownership markers). Do not duplicate its machinery. An interim
   alternative, if tool-building is deferred: repoint those CTAs at
   `/matchmatrix.html` (a real 200 page) so the homepage stops 404-ing.
2. **`bugs_open/022` — the scheme guard.** The damage mechanism behind the
   light-on-dark incident is still unguarded fleet-wide. The case file has all
   three verifications the council demanded already done. It needs its own
   council submission and a fix.
3. **`bugs_open/017…unregistered_action…`** — `fix_forced_text_colors` is
   registered nowhere (so every `color-variable-fixer` run dies), and
   `CompleteWorkItemAction` stamps `complete` on failed sagas. Both open.
   **Refer to it by SLUG: there are two `017` files** (the other is
   `017_HANDOFF_2026-07-18_static_cutover_orphans_backend_entry_forms.md`).
   Numbering is never reassigned — this is a documented trap in
   `/bugs_closed/README.md`.
4. **Optional**: `5151d4a79` (generic_theme consolidation) is committed but not
   in the running image. It is **structural only** — classification verified
   identical before and after — so there is nothing to chase; it lands with the
   next roll.

## What is already live and needs no action

- **The `generic_theme` fix is LIVE and PROVEN.** Another session rolled
  **v1.0.1137** (pod started 2026-07-18 22:21); the running binary carries the
  predicate (`strings /app/agent-chassis | grep -c 'jsonb_typeof(content_data'`
  → 1) and there have been **zero `generic_theme` detections fleet-wide since
  the roll**. The perpetual misfire is over.
- Council correlation **`e0ebf6ee-dcc0-4a7b-9a3d-438ce9af5fff`**, three rounds,
  final: **7 of 8 seats approve**, one dissent (bug_historian) whose two points
  were both answered — one empirically (no other discovery check has the
  existence-only shape), one in code (`5151d4a79`). The loop was stopped there
  deliberately: CLAUDE.md says one run per coherent task, not per iteration.

## Queue reality (checked 2026-07-19)

robot-hands has a **large inherited backlog**: 115 unresolved, 54 failed, 53
needs_human_review, 17 detected — almost all predating this thread. Nothing of
mine is pending. Do NOT read that backlog as new damage.

**Dispatch is slow by design**: `build-pipeline-trigger` (30s) processes **one
item per site at a time** (a site is skipped while anything is `claimed`), so a
50-item backlog takes hours. To jump a user-visible batch ahead of churn, lower
its `priority` and stamp `triaged_at` — see the RUNBOOK.

## Two process notes for whoever picks this up

- **CLAUDE.md changed on 2026-07-19**: the diagnosis loop (090) is now the
  **DEFAULT before asserting any durable claim**, reversing the previous
  "debug directly" advice. `bugs_open/017…unregistered_action…` and
  `bugs_open/022` were both authored *before* that correction and have **not**
  been through the loop. Their evidence is cited and re-checkable, but if
  either is about to become load-bearing for someone else's work — especially
  022, whose fix changes behaviour fleet-wide — file it to 090 first.
- The workstream now keeps the **standing five** (PLAN / RUNBOOK / NOTES /
  README_where_we_are / SUMMARY) in this directory. `README_where_we_are.md` is
  the **owner's** document: append only, never rewrite or reorder.
