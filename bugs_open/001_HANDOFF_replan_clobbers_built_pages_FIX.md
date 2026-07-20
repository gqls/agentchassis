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

---

## FRESH EVIDENCE — 2026-07-18, leopardessconsulting.co.uk (severity raise)

> **CORRECTED 2026-07-19 — this whole section is MISATTRIBUTED. The leopardess damage was
> real, but this bug did not cause it.** Read the correction at the end of the section
> before acting on anything in it. The section is left standing, unedited, because the
> damage it records is real and still needs owners — just not this one.

This bug is not dormant and it is not confined to idea.uk. It hit leopardess **twice in 24
hours**, each time undoing hand-verified human work, and it produced an owner-visible defect.

**Recurrence 1 — homepage, 2026-07-17 14:14.** A hand-written, owner-approved plain-voice
homepage (4 curated sections) was rebuilt into 6 sections. The rebuild re-added
`system-stats` and `case-studies-grid` and populated them with **fabricated content**:
- `system-stats`: *"Functional Areas: 150+"* — resurrecting the exact "functional areas"
  fabrication that had been audited out of this site days earlier, plus `150+%` / `150++`
  garbage from the shared component's forced `%/ms/+/x` suffixes;
- `case-studies-grid`: **invented case-study titles** ("Validation Layer Stops Bad Data
  Reaching the Warehouse", "Content Operation Running Without Manual Handoffs") — this site
  has no clients and no case studies.

**Recurrence 2 — services page, 2026-07-18 07:50.** A restored, reviewed services page was
rebuilt again. Among the generated content it created a link to
`/tools/tool-monitoring-coverage-gap-finder.html` — **a tool page that does not exist**. The
site owner clicked it and landed on a blank 404. That is the user-visible bug report that led
here; the root cause is this one.

### What this adds to the original diagnosis

1. **It is not only *loss*, it is *injection*.** The original write-up frames the bug as built
   composition being discarded. The worse half is what replaces it: audited-out fabrications
   return, and phantom links to non-existent pages are created. On a site whose entire
   governing rule is "no claim ships without an evidence row", a re-plan is a fabrication
   *source*.
2. **It defeats human review.** Copy that was read, corrected, and verified by a person is
   silently replaced by unreviewed LLM output. Any human content pass on a non-adoption-locked
   site has an undefined shelf life, so the review effort cannot be trusted to hold.
3. **It is fast enough to outrun a fix.** Two clobbers inside one working day on one site,
   while a session was actively repairing that site.

### Mitigation that demonstrably survives a clobber (useful for the fix design)

Per-page hero images wired through `site_plans` / `site_plan_imagery` rows **survived both
recurrences**, while `page_components` content did not. Whatever the fix is, the imagery-plan
join is a working example of state the re-plan does not trample — worth reading before
designing the guard.

### Suggested fix direction (unchanged in spirit, sharpened)

A re-plan must not be able to overwrite a page whose content a human has touched. The existing
`locked_at` / lock_type machinery on `page_components` is the obvious lever: treat
human-reviewed content as locked and make the planner **refuse** (loudly, as a work item)
rather than silently regenerate. A `needs_human_review` item saying "the plan wants to change
this page; here is the diff" is strictly better than a silent rewrite.

**Verification case:** leopardess index + services. Both restored by hand; if the guard works,
they stay restored across a subsequent `build-site-planner` run.

---

## CORRECTION — 2026-07-19: the leopardess evidence above belongs to two OTHER bugs

The section above attributes both leopardess recurrences to this bug and says "the root cause
is this one". **It is not.** Checked against the live DB before writing the fix:

**`reconcile_site_plan` has never emitted a work item on leopardess — not once in the site's
entire history.**

```sql
SELECT source, item_type, count(*), min(created_at), max(created_at)
FROM site_work_items WHERE site_id='4851f6fc-71cf-4160-a270-e03d6d3e0732'
GROUP BY 1,2 ORDER BY 4 DESC;
-- 48 rows across 20 sources. reconcile_site_plan is absent from all of them.
```

Since `reconcile_site_plan` is the only route from a plan to a page rebuild, no re-plan ever
touched that site. What actually did:

- **Recurrence 2 (services, 2026-07-18 07:50) — `tool-suggester`, not the planner.** At
  02:32 the suggester wrote nine `content_rewrite` items, one of them *"Add Monitoring
  Coverage Gap Finder tool reference to services page"* with spec `suggestion` = *"Weave a
  natural reference to 'Monitoring Coverage Gap Finder' …"*. Status `complete`. The
  suggester writes these at **suggestion** time, alongside (not after) its `add_tool` item,
  so the rewrite instruction names a tool that has no page:
  `SELECT name FROM pages WHERE site_id='4851…' AND name ILIKE '%monitoring%'` → **0 rows**.
  The writer duly emitted a link to `/tools/tool-monitoring-coverage-gap-finder.html`. That
  is the owner's 404. Confirmed at the orchestration too — run `615bee1d` at 07:44:54 carries
  `{"page":"services","source":"tool-suggester","suggestion":"Weave a natural reference to
  'Monitoring Coverage Gap Finder'…"}`, and the services `page_components` rows all update at
  07:50:41.
- **Recurrence 1 (homepage, 2026-07-17 14:15) — a `page-rerender` run.** Orchestration
  `f8a5bbbf`, workflow steps `check_rerender_mode → rerender_sections → save_sections →
  render_page → deploy_page`. Again not the planner.

**What this changes.** The fix in this file is still correct and still worth shipping — the
idea.uk evidence is sound and the code plainly does what the diagnosis says. But:

1. **Shipping it will not stop the leopardess clobbers.** Anyone who believed the section
   above would have shipped this fix, watched leopardess get rewritten again, and concluded
   the fix failed.
2. **The severity raise does not transfer.** This bug needs someone to fire a re-plan, which
   is rare and deliberate. The tool-suggester path runs autonomously — which is the more
   urgent property, and it belongs to that bug, not this one.
3. The phantom-link mechanism is filed separately as **`/bugs_open/029`**.

**What caught it:** checking `site_work_items` by `source` for the affected site before
trusting the attribution. The lesson is the cheap one — *"a page was rewritten"* does not
identify **what** rewrote it, and on this platform at least four independent paths can rewrite
a page. Attribute by the emitting source row, not by the damage.

A `needs_diagnosis` item was also filed for the tool-suggester mechanism (item_key
`needs_diagnosis:tool-suggester-writes-content-rewrite-wo`), but **the run never dispatched** —
no orchestration exists for correlation `a8b483ff-55af-463d-9622-837c73780e48`. The finding
above rests on the primary DB evidence quoted, not on a loop verdict.

---

## FIX APPLIED — 2026-07-19

Both halves are written; the Go half is **inert until the chassis image rolls**.

- **Go** — `v3_site_actions.go`: preservation set widened from adoption-locked to
  **adoption-locked OR `build_status='deployed'`** (fix steps 1 + 3); new **Pass B2** snaps a
  built page's composition back by name, **gated on the realised sections being non-empty**,
  which closes the main defect and the second (empty-sections) defect together (fix step 2);
  truncation must-keep widened to match. `realisedPageIsBuilt` mirrors `decideEmit`'s
  `skip_built` test so "built" has one definition across planner and reconciler.
- **DB** — `sql_for_agents/173_load_existing_pages_build_status.sql`, **applied and verified
  live**: the step query now surfaces `build_status` (it previously did not, so the Go side
  could not have seen it). Also deduplicated a repeated `p.sections, p.meta_description,
  p.nav_order` tail left when the carry-fields migration landed on top of 054.
- **Landing order is free.** `realisedPageIsBuilt` returns false on a missing column, so the
  Go half degrades to the old adoption-locked-only behaviour; the query change is inert on the
  current chassis. Neither half can break the other while in flight. There is a test for this.
- **Tests** — `v3_site_reconcile_test.go`, 7 cases from the idea.uk fixtures. Verified
  **discriminating**: neutralising `realisedPageIsBuilt` fails exactly the two built-page
  tests and reproduces the recorded symptoms (`about` re-composed to a generic `hero`,
  `report` dropped from the plan).

**Deliberately NOT done, and why:**

- **Pass C2 (item-topic stem dedup) was left at the adoption-locked scope.** It is a name-stem
  heuristic; a false positive suppresses a legitimately new page (a new `tool-pricing` beside
  a built `guide-pricing` shares the stem `pricing`). Bounded to the 90-day window that is
  acceptable; permanent for every built page it is not. It is also not needed here — invented
  pages carry new topics and so collide with nothing.
- **"Pages invented" (original symptom 3) is NOT fixed.** Nothing here stops the LLM proposing
  net-new pages; truncation now merely can't evict a built one to make room. Suppressing
  invention outright would block legitimate site growth, so it needs the explicit-intent
  plumbing of fix step 4 (a `rebuild:true` / scope field on the `needs_site_plan` spec) — a
  policy call, not a bug fix.
- **Fix step 4 (explicit-intent rebuild gating) is not built.** A deliberate redesign of a
  built page is now impossible through a re-plan rather than merely un-silent. If that blocks
  a real workflow, step 4 is the designed way back in.

**Known residual — Pass B still carries emptiness forward (found 2026-07-19, after committing).**
Pass B (the URL-match rename snap-back) replaces the LLM page with
`normaliseRealisedToPlanPage(rp)` **wholesale**, which carries the realised `sections` *including
an empty `[]`*. Pass B2's non-empty gate does not protect this path, because Pass B `continue`s
before B2 runs. So the second defect survives in one narrow case: **a catalogued (uncomposed) page
whose URL the LLM reuses under a different name** is snapped back to the realised identity *and* to
its empty composition, and can never be filled. The common case (same name) is fixed; this one is
not. The fix would be to keep the realised identity but take the LLM's sections when the realised
ones are empty — i.e. give Pass B the same gate — but that changes which fields Pass B is allowed
to carry, so it wants its own review rather than a quiet widening. Not attempted here.

**Council review: NOT obtained.** Two rounds were voided by `/bugs_open/019` (a truncated
`review_editquality` at the 8000-token cap discards the whole round). Round 1 was a fresh
9,655-byte plan; round 2 a 6,026-byte lean resubmission — **both** voided on the first seat, so no
reviewer read this change. A third round was not attempted (see the fourth reproduction in 019 for
the reasoning and for what these two rounds add to that bug). Commits `c41e9ddbc` and `fcd8812f3`
therefore carry **no `Council-Reviewed:` trailer**, and will show as unreviewed in the 098 report.
That is this bug's doing, not a skipped review.

## VERIFIED LIVE — 2026-07-20, dartsonline.com

Chassis rolled (pod `agent-chassis-55d7774dc4-pzt9j` on `v1.0.1138`; the planner runs in its own
ephemeral pod, `v1.0.1139`). Both binaries grep positive for `realisedPageIsBuilt`,
`reconciled with realised pages` and `snapped built page composition`. Migration 173 re-checked and
still applied (no re-seed clobber).

Test site **dartsonline.com** (`5fe8785b-…`): 3 `deployed` pages with real compositions, 10
catalogued-but-empty, one current plan from 2026-07-06 — i.e. a genuinely **non-adopted** site, the
exact case where the function used to no-op. Snapshotted to `_darts_bak_20260719_{pages,plan_pages,
plan_sections}` first, then emitted a `needs_site_plan` (item `2845b23d`).

**Trap hit on the way:** the work item read `status='complete'` with `updated_at` **identical to
`created_at`** — which reads exactly like "nothing ran". It had run. The real evidence is the new
`site_plans` row (`5d438145`, `is_current`, 09:43:29). This is CLAUDE.md's "trust the artefact, not
the status" in its most literal form; `updated_at` on `site_work_items` is not maintained.

**The decisive comparison** — `orchestration_states.collected_data` for run `c342cffa` stores BOTH
the LLM's raw proposal (`llm_plan.result.pages`) and the post-convergence plan (`validate_plan.pages`),
so the two can be diffed directly rather than inferred from logs (the validate step's pod is
ephemeral and its logs were already gone):

| page | build_status | LLM proposed | plan got | realised |
|---|---|---|---|---|
| `about` | deployed | `…, differentiators-section, …` | `…, differentiators, …` | `…, differentiators, …` |
| `new-arrivals` | deployed | same as realised | unchanged | unchanged |
| `shipping-returns` | deployed | same as realised | unchanged | unchanged |
| `guides-index` | planned | `[hero, content-listing]` | kept | was `[]` |
| `index` | **needs_rebuild** | dropped `differentiators`+`content-listing`, added `features` | **taken as proposed** | 7 sections |

**PROVEN: all three `deployed` pages kept their exact composition**, and Pass B2 provably fired on
`about` — the plan records `differentiators` where the LLM wrote `differentiators-section`, which can
only happen via the snap-back. Pre-fix the convergence early-returned and `pages.sections` would now
read the LLM's string.

> **SUPERSEDED 2026-07-20 by the second re-plan below — the clean rescue case this paragraph said
> was missing has now been observed twice.** The caveat below remains accurate about *run 1*; keep
> it, because it is the reason run 2 was worth doing.

> **Honest limit on that one, corrected before claiming more than it shows.** `differentiators` is
> the `function` and `differentiators-section` is the `name` of the SAME component row
> (`content_components`: `name='differentiators-section', function='differentiators'`; likewise
> `about-hero`/`hero-about`). So this snap **proves the code path executes** but did not
> demonstrably rescue the page from a visible regression — both strings point at one component.
> A clean rescue case is still unobserved in production. Do not quote this row as "prevented a
> regression"; quote it as "convergence now runs and rewrites the plan, where before it returned
> early".

**Rendered artefact checked, not just the DB row.** `about` was rebuilt (09:48:51) and its
`page_components` are `about-hero, about-content, differentiators-section, call-to-action` — which
*looks* like a mismatch against `pages.sections` `["hero-about","about-content","differentiators",
"call-to-action"]` but is not: sections store the **function**, `page_components` reference the
**component**. Verified against `content_components`. The rebuild rendered the preserved composition.

### Three things this run measured that the fix does NOT address

**All now filed as their own cases** (2026-07-20): `/bugs_open/037` (needs_rebuild unprotected),
`/bugs_open/038` (every deployed page still rebuilt, content regenerated), and — from the near-miss
described above — `/bugs_open/039` (function-vs-name, and a section naming a missing component
rendering a hollow stub). `/bugs_open/035` records the `updated_at` trap. Summaries follow; the case
files hold the evidence and fix candidates.

1. **`needs_rebuild` pages are not protected.** `index` had a live 7-section composition and lost
   `differentiators` + `content-listing` (both distinct components, not aliases — checked). It is
   `needs_rebuild`, so `realisedPageIsBuilt` excludes it. This is arguably fix step 4's intended
   escape hatch ("a page whose `build_status` was set to `needs_rebuild`" as explicit rebuild
   intent) — but it means **"built page" == "deployed" only**, and a flagged page still loses its
   composition silently. Whether that is the wanted boundary is an open decision, not a settled one.
2. **Every deployed page is still REBUILT.** `reconcile_result` came back `pages_skipped_built: 0` —
   all three deployed pages emitted as `stale` because `built_from_plan_version != <new plan id>`.
   So composition survives, but the page is re-rendered and its **content regenerated**. This fix
   secures structure, not copy. Anything relying on "a re-plan won't touch my reviewed text" is
   still wrong. (`decideEmit`'s stale test is the lever; re-stamping `built_from_plan_version` when
   the composition is unchanged would close it.)
3. **Pages are still invented.** Two net-new pages (`grip-styles`, `shaft-length`) — as documented
   above, unfixed by design.

## SECOND RE-PLAN — 2026-07-20, the genericity proof (plan `5d438145` → `dcc7834e`)

Run 1 left a hole: every deployed page it "preserved" had been proposed identically by the LLM
except `about`, whose snap turned out to be a naming variant. So preservation was demonstrated but
never actually *exercised*. The owner's instruction closed that hole — rather than hand-restoring
`index`, refresh it through the framework and re-plan again, so the guard has to prove itself on a
page whose status changed by the platform's own route rather than by a hand edit.

`index` was rebuilt via its own `needs_page` item and reached `build_status='deployed'` (that
rebuild also exposed `/bugs_open/040`). `guides-index` had likewise gone `planned`→`deployed`. Both
were then in the preserved set for the first time. Second `needs_site_plan` emitted; run
`b6c30dba`, LLM proposed 18 pages.

**Two genuine snap-backs, on distinct components — this is the rescue case:**

| page | LLM proposed | plan got (realised) | what was prevented |
|---|---|---|---|
| `index` | `hero, product-grid, features, call-to-action, testimonials, content-listing` | `hero, product-grid, category-listing, features, call-to-action, testimonials` | **`category-listing` dropped and `content-listing` swapped in** |
| `shipping-returns` | `generic-text-block, faq` | `generic-text-block` | an unrequested `faq` section added to a built page |

Distinctness checked, not assumed (the run-1 trap): `category-listing`, `content-listing` and `faq`
are three separate rows in `content_components`, each its own `function`. So a real section would
have been lost from `index` and a foreign one added, and the guard stopped both.

`about`, `new-arrivals` and `guides-index` were re-proposed identically this run, so preservation was
not exercised on them — recorded so the table above is not read as five rescues.

**Why this proves the fix is generic and not a dartsonline artefact.** `index` is the *same page*
that **lost** `differentiators` + `content-listing` on run 1. Nothing about the page, the site or
the plan changed in kind between the two runs — only its `build_status`, `needs_rebuild` →
`deployed`, and that transition was performed by the framework's own build path, not by hand. Same
site, same page, opposite outcome, explained solely by the status the guard keys on. That is the
mechanism working as designed rather than a site-specific coincidence, and it simultaneously
demonstrates `/bugs_open/037` from the other side: a page is unprotected as `needs_rebuild` and
protected as `deployed`.

**`/bugs_open/038` reproduced identically**: `reconcile_result` again returned
`pages_skipped_built: 0`, so all five deployed pages were re-emitted `stale` despite four of them
being byte-identical to the plan. Composition preserved, rebuild still requested.

Also this run: `brands-index` and `shop-index` (both `planned`, empty) were composed from the LLM's
proposal — the second defect's fix working again — and one page was invented (`brand-comparison`),
consistent with that remaining unfixed. The run's 6 emitted rebuilds were cancelled once the result
was read.

**Cleanup state:** run 1's re-plan queued 16 `needs_page` items. `about` completed; 13 were paused
(`triaged`→`detected`) once the verification result was in, to stop unnecessary spend. Snapshot
tables retained. `index` is paused and still carries the re-proposed composition; restoring it from
`_darts_bak_20260719_pages` would desync `pages.sections` from the new `site_plan_sections`, so it
was left alone pending a decision.
