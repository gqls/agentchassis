# 478 — the strategist's "is this site deployed" gate reads ANY shipped page as "this site was planned", so a site whose tool pages ship before its plan never receives a briefing or a plan

**Filed 2026-09-04** by the `portfolio_positioning` lane. **Status: OPEN, UNOWNED.** Cross-cutting.

**Diagnosis-loop note (owner ruling 2026-07-31):** 090 run `467f0283` returned **UNVERIFIABLE** —
*"agent_definitions for domain-strategist has orchestrator_workflow, task_workflow AND
orchestration_workflow all NULL, so the actual gate_next_item/check_site_deployed/create_next_item
step JSON … is not readable from this row"*. That is the loop's documented static-tier limitation
(concept register `diagnosis-loop.md` line 268: workflow definitions live in
`agent_definitions.default_config` as JSON and are invisible to it). **This file therefore substitutes
first-hand verification**, all of it quoted below from the live rows and reproducible by the queries given.

## 1. Symptom

A greenfield site is briefed, classified, given a strategy and a composition, and **never gets a page
plan**. No `needs_briefing` and no `needs_site_plan` item ever exists for it (live table + archive).
Audits then pile up record-mode verdicts naming the missing pages (`needs_content_planning:deferred`),
because the audits can see the brief's plan and the builder never ran it.

## 2. Mechanism — three config steps on `domain-strategist` (`agent_definitions.default_config->workflow->steps`, live 2026-09-04)

```
check_site_deployed  (query_database):
  SELECT (COUNT(*) > 0) AS is_deployed FROM pages
   WHERE site_id = $1 AND NOT (deployed_at IS NULL AND COALESCE(build_status, '') <> 'deployed')
gate_next_item       (conditional_branch):
  condition "site_state.is_deployed == true"  then_step "complete"  else_step "create_next_item"
create_next_item     (create_work_item):
  item_type needs_briefing → handler build-briefing-agent → (which files needs_site_plan → build-site-planner)
```

The gate was added by migration **341** (2026-08-08, "refresh-safe") and its predicate corrected by
**359** (2026-08-09) to `PageHasShippedPredicateFor("")`. Its purpose, in 359's own words: *"a site whose
shipped pages are all flagged needs_rebuild would read is_deployed = false and the gate would CHAIN A
RE-PLAN of a serving site — precisely what the gate exists to prevent."* **The intent is right. The
predicate answers a different question.** "At least one page has shipped" is read as "this site has
been planned and is live". That holds only if pages can exist solely downstream of a plan — and they
cannot: the **tool path** (`evaluate_tools` → tool-suggester → tool-deployer, which creates and deploys
its own pages, `bugs_open/450` §7) and at least one **seeding** path run in parallel with the build
path from the moment a site goes active. Whichever finishes first decides the gate.

## 3. Evidence, first-hand

**copyonline.co.uk** (site `3d965325…`), timeline `[MEASURED]`: `evaluate_tools` 15:55Z 09-03 → tools
spec 16:01 → **ten tool pages created 16:15–17:20** → classification 16:57 → **strategy 17:44**. Both
strategist orchestrations (17:44Z and 22:31Z) carry `collected_data.site_state = {"is_deployed": true}`
and **no `next_item_created` key**. Briefing/plan items ever: **0**. Pages archived later; the survivors
still carry `deployed_at`, so the gate would answer the same today.

**designblog.co.uk** (control, 09-02): `evaluate_tools` 15:44 → strategy **15:58** → `needs_briefing`
15:59 → `needs_site_plan` 16:02 → first page **16:10**. Strategy won the race by 12 minutes; briefed.

**oxenunity.com** (09-02): six `tool-*` pages created 19:20, strategy 19:57, never briefed, **no plan**.
**cookly.uk** (08-09): `index`, `about`, `contact` created 13:10 (a seeded skeleton, not tools),
strategy 19:20, never briefed, **no plan** — a second producer of pre-plan pages.
**loancalculator.co.uk**: pages 08-03, strategy 08-08, gate-skipped, then **briefed by hand 08-15**
(`created_by = manual-redrive-2026-08-15-post-b2-gate`) — this defect was hit and worked around
manually twenty days ago, and the 359 header names that very site as its "witness" for `is_deployed = true`.

## 4. Blast radius `[MEASURED 2026-09-04 ~08:1xZ]`, with the instrument's blindness stated

`pages.deployed_at` **moves** — every rerender re-stamps it — so "did a stamp exist at strategy time"
is not recoverable from the current column (copyonline's own `min(deployed_at)` is now 18:23Z, AFTER
its 17:44Z strategy, and the census that used it returned zero). The stable proxy is
`pages.created_at < first strategy`:

```sql
WITH st AS (SELECT site_id, min(created_at) fs FROM site_specs WHERE aspect='strategy' GROUP BY site_id),
     fp AS (SELECT site_id, min(created_at) fc FROM pages GROUP BY site_id),
     b  AS (SELECT DISTINCT site_id FROM (SELECT site_id,item_type FROM site_work_items UNION ALL SELECT site_id,item_type FROM site_work_items_archive) u WHERE item_type='needs_briefing')
SELECT s.domain FROM st JOIN sites s ON s.id=st.site_id JOIN fp USING (site_id) LEFT JOIN b USING (site_id)
 WHERE fp.fc < st.fs AND b.site_id IS NULL;
```
→ **5 of 40** sites with a strategy: copyonline.co.uk, oxenunity.com, cookly.uk (greenfield, **no plan**)
and lampenkap.com, loancash.co.uk (**adopted** sites with real pre-existing pages, where "do not re-plan
a live site" is the right answer — excluded from the defect). Control: 33 briefed sites whose first
page postdates their strategy. So **3 greenfield sites stalled, plus 1 manually redriven**, as of today.

## 5. Fix candidates, ordered by what makes the bad state unrepresentable

1. **Ask a site-level question.** Replace the predicate with "has this site ever been PLANNED":
   `EXISTS (SELECT 1 FROM site_plans WHERE site_id = $1)` (or `sites.published_at IS NOT NULL` if
   "serving" is the guard's real meaning). A never-planned site then always gets its briefing, and a
   planned, serving site is still protected — the 359 intent holds by construction. One config edit
   on one step; verify on a planned live site that a strategy refresh still does NOT chain a re-plan.
2. **Exclude pre-plan page producers from the shipped predicate** (`name NOT LIKE 'tool-%'`, seeded
   skeletons…). Weaker: a list to maintain, and the next producer re-opens it.
3. **Sequence the tool path behind the plan.** Correct in principle, widest blast radius, and the
   tool path's independence is used on purpose elsewhere (450).

## 6. How to verify

A greenfield site whose tool pages ship first must acquire a `needs_briefing` at its first strategy
run (copyonline is the natural witness once its manual briefing — item `479614c9`, filed 2026-09-04
07:47Z by this lane — has been cancelled or completed). The regression that must NOT happen: a site
with a current plan and shipped pages receiving a `needs_briefing` from a strategy refresh.

## 7. Related

`bugs_open/450` §7 (the tool-deployer creates its own pages) · `bugs_open/315` (page-level
`deployed_at` written without publishing) · `bugs_closed/037` (why 359 chose the shipped predicate) ·
LANDMINES *"`pages.deployed_at` is set on EVERY page of a site that has never been published"*
(2026-09-03 — the human version of this same read) · `portfolio_positioning` NOTES (eee)/(fff).

**Ownership of the other two stalled sites `[checked 2026-09-04 ~08:35Z]`:** neither `oxenunity.com`
nor `cookly.uk` has a lane directory under `docs024_key_docs_latest/` or a `who-owns.py` match; the
handoffs that mention them are other bugs citing them in passing. So there is nobody to send a CONTRIB
to. Whoever takes this bug: the one-row workaround (§6, copyonline's item `479614c9` is the template)
unblocks each of them without waiting for the gate fix; do not apply it to the two adopted sites.
