# Handoff — FIX: re-planning a site silently discards its built pages' composition

You are fixing a **structural, fleet-wide** bug in the chassis site planner. It causes **silent data
loss**: re-running `build-site-planner` on a site that already has built pages throws away those
pages' composition and replaces it with whatever the LLM proposes (or fails to propose) that run.
This was found the hard way on idea.uk on 2026-07-14 (see "Evidence" below). Nothing is on fire —
idea.uk was recovered — but the bug is live for every non-adopted site, so any future re-plan is a
loaded gun.

## Working rules (hold these)
Go, not Python. British English. **Schema first**: read `\d <table>` before any SQL, and read the
function before you change it. Prefer **structural fixes over quick patches**. **Reuse and adapt
existing functions before writing new ones.** Keep workflows simple; put complexity in Go action
code. `logger.Info`, never `logger.Debug`. Do not treat a 0-row result as decisive until you've
cleared the query. Go changes are **inert until the chassis image is rebuilt** (`make
quick-agent-update`) — unlike DB config, which is live immediately.
DB: `PSQL="kubectl -n ai-persona-system exec -i postgres-clients-0 -- psql -U clients_user -d clients_db"`.

## The system in one paragraph
Multi-agent Go/Kafka/Postgres platform that plans and builds multipage sites from a domain.
`build-site-planner` proposes a plan (LLM), validates it, writes `site_plans` /`site_plan_pages`
/`site_plan_sections`, then a cascade builds and deploys pages. A site's *realised* pages live in the
`pages` table; `pages.sections` is an ordered JSON array of component names, and
`build_status='deployed'` marks a built page. Reference site: idea.uk, site_id
`1244516d-014d-421c-88c6-090bb1e9552a`.

---

## The bug, precisely

**File:** `platform/orchestration/actions/v3_site_actions.go`.
**Function:** `ValidateSitePlanAction` (the `validate_plan` step), around **line 2636–2695**, which
calls **`reconcilePlanWithRealised`** (defined **:4516**), which uses **`normaliseRealisedToPlanPage`**
(**:4458**).

`reconcilePlanWithRealised` is the ONLY thing that carries a realised page's composition into a
re-plan. **It force-preserves only pages under a live adoption lock** (`:4521-4535` filters
`existingPages` down to `adoption_locked == true`, and returns the LLM plan untouched if none are
locked). Its own header comment says so (`:4499-4505`), and the caller comment is blunt
(`:2656`): *"reconcilePlanWithRealised no-ops for every site (adopted pages never preserved…)."*

**Consequence for any site NOT under a live adoption lock** (i.e. most sites: adoption locks expire
after 90 days, and from-scratch builds are never locked):
- `reconcilePlanWithRealised` is a **no-op** → the plan is the **raw LLM proposal**.
- A built page the LLM **re-proposes by the same name** falls through every pass and is kept as the
  LLM's version (`:4627 kept = append(kept, lm)`). Its realised `sections` are **not** carried. →
  **the built page is silently re-composed** (usually worse: specific components like `hero-about`
  replaced by a generic `hero`, sections dropped).
- A page the LLM **omits** is dropped from the plan entirely (only the adoption-locked union at
  `:4640-4652` would re-add it, and that subset is empty).
- The LLM may **invent pages** not asked for; `max_pages` (80 in idea.uk's config) does not stop this
  until the cap.

**Second, independent defect (bites even adoption-locked sites):** `normaliseRealisedToPlanPage`
(`:4469-4480`) carries `sections` faithfully **including an empty `[]`**. So a
catalogued-but-uncomposed page is preserved **as empty** — a re-plan can never fill it. The union
carrying emptiness forward means "re-plan to compose the missing pages" is structurally impossible
for a page the LLM doesn't spontaneously compose.

**The doctrine that led here is wrong.** Prior handoffs
(`idea_uk_section_data_missing/HANDOFF_claude_code_continue.md`, OPEN ITEM 1) state: *"re-running
build-site-planner … safely unions already-built pages with the new ones via
normaliseRealisedToPlanPage."* That is true **only** for the adoption-locked subset. On a
non-adopted site it is false, and doing it regresses the built pages. Correct that doc as part of
this fix.

---

## Evidence (idea.uk, 2026-07-14, plan `32be2797` → `ff03bdef`)

A `needs_site_plan` re-plan produced, from a clean 9-page site:
- **4 built pages regressed** — `index` dropped `info-card-grid`; `about` dropped `hero-about` +
  `info-card-grid` and gained a generic `hero`; `contact` dropped `hero-contact` for a generic
  `hero`; `report` dropped `generic-text-block` + `info-card-grid`. `about` and `contact` actually
  **re-rendered and re-deployed** the regressed artefact (proven via `page_components`: both rendered
  slot 1 = generic `hero`, not their specific component).
- **10 pages invented** — 5 extra tools + 5 blog posts, all uncomposed.
- **`tool-audience-check` stayed empty** (the second defect — no sibling, union preserved `[]`).

Recovery is documented in `idea_uk_vm_site/RUNNING_NOTES_idea_uk_vm_site.md` §I–K and
`idea_uk_vm_site/sql/p1_02_replan_rollback.sql` (restore built pages' sections from the prior plan,
delete invented pages+items, re-emit rebuilds for the two that redeployed regressed). idea.uk is
**already fixed**; do not re-run its rollback. This handoff is about the **code**, so it never
happens again.

---

## The fix

The principle: **a re-plan must never silently redesign or drop an already-built page.** Adoption
locks are the wrong (too narrow) gate — a built page deserves preservation whether or not it is
adoption-locked. Preservation should key on **"is this page built?"** (`build_status='deployed'`),
not on the adoption lock.

**Schema-first before coding:** read the `load_existing_pages` query
(`load_site_pages_action.go`, and the workflow step config in the `build-site-planner`
`agent_definitions.default_config`) to confirm which columns it surfaces per realised page. It
currently carries `adoption_locked`; **the fix needs `build_status` (and `sections`) surfaced too**
if they are not already. `\d pages` first.

**Preferred structural shape** (adapt, don't rewrite — extend the existing passes):
1. In `reconcilePlanWithRealised` (or a sibling it calls), widen the force-preserve set from
   "adoption-locked" to "**adoption-locked OR built (`build_status='deployed'`)**". Do NOT make it a
   no-op when only built (non-locked) pages exist.
2. Add a pass, mirroring Pass B's snap-back, keyed on **name match** (not just URL): for an LLM page
   whose name equals a realised page, **snap its sections back to the realised composition** —
   **but only when the realised sections are non-empty.** This single rule fixes both defects:
   - realised **non-empty** (a built page) → preserved, LLM cannot clobber it;
   - realised **empty** (a catalogued page) → keep the LLM's proposed sections, so composition is
     finally *allowed* to happen.
3. Keep the truncation guard (`:2683-2695`) preserving the widened must-keep set, not just the
   locked subset, so a re-plan can never truncate away a built page.
4. Consider gating a deliberate rebuild behind explicit intent (a per-page `rebuild:true` in the
   `needs_site_plan` spec, or a page whose `build_status` was set to `needs_rebuild`), so a genuine
   redesign is still possible — just never the silent default.

**Do not** "fix" this by forbidding re-plans, or by hand-writing `site_plan_sections` — the writer is
`write_site_plan_action.go` (the only `INSERT INTO site_plan_sections`, `:391`) and it must stay the
single writer.

## Related gap (decide whether in scope)
The intended single-page retrigger, `discovery_checks/check_sectionless_pages.go`, only fires for a
sectionless page that has a **same-role sibling with sections** (its query's `EXISTS` clause). A page
whose role has no sibling (e.g. the only `tool` page) cannot be auto-composed by any route today. If
the platform needs to compose such a page, this check needs a no-sibling fallback (a role-default
layout, or routing that one page to the LLM composer). idea.uk sidestepped it by making
`tool-audience-check` a pointer page (url → the live tool, `build_status='deployed'`, no sections);
that is a workaround, not a fix for the general case.

## How to verify the fix (before trusting it on a real site)
1. Pick or make a disposable multi-page site with several `build_status='deployed'` pages carrying
   distinct, specific components (not generic `hero`), and at least one catalogued page with
   `sections=[]`.
2. Snapshot `pages.sections` + `site_plan_sections` for the current plan.
3. Emit a `needs_site_plan` (shape: see `idea_uk_vm_site/sql/p1_01_replan_emit.sql`). Let it run.
4. Assert: every built page's `sections` is **unchanged**; the empty catalogued page is now
   **composed** (or intentionally left, if the LLM declined and no sibling exists); **no** invented
   pages; `page_components` for the built pages still shows their **specific** components.
5. Only then is it safe to point the improvement loop or a re-plan at any populated site.

## Key references
- Bug site: `v3_site_actions.go` — `ValidateSitePlanAction` (~:2636), `reconcilePlanWithRealised`
  (:4516), `normaliseRealisedToPlanPage` (:4458).
- Reconcile emit logic: `reconcile_site_plan_action.go` `decideEmit` (:293) — `build_status='deployed'`
  + matching `built_from_plan_version` ⇒ `skip_built`.
- Plan writer: `write_site_plan_action.go:391`.
- Single-page retrigger: `discovery_checks/check_sectionless_pages.go`.
- Doc to correct: `idea_uk_section_data_missing/HANDOFF_claude_code_continue.md` OPEN ITEM 1.
- Full incident + recovery: `idea_uk_vm_site/RUNNING_NOTES_idea_uk_vm_site.md` §I–K; memory
  `replan-clobbers-built-pages`.
