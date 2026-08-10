# PLAN — 2026-08-04 — finish the imagery, then put the site to the visual designer

Owner, 2026-08-04: *"after we've sorted out the current design problems with the
site — missing images and so on — please can we put the site to the visual designer
to improve all the components and imagery."*

Sequenced deliberately, and this plan's whole job is to make the first half real
before the second half runs. A visual designer pass over a site with five 404
images will spend its judgement on the holes rather than on the design.

## OWNER RULING 2026-08-04 — per-site loop firing is the standing method

Asked on 2026-08-03 (see `PLAN_2026-08-03_…` §"the open question") and now
answered: **keep firing the improvement loop deliberately, per site.** Not
re-enabling `improvement-sweep`, not bulk-promoting `detected → triaged`.

So `294_TRIGGER_improvement_loop_v1.sh <site_id> [domain]` is the sanctioned
entry point rather than a one-off, and IMP-050 is a standing mechanism. The
fleet's 204 parked findings stay parked until someone fires their site, which is
the intended trade: control over coverage.

**What this makes load-bearing.** The trigger is now the only route anyone has, so
its two pre-flight refusals (the 300s post-roll window, the claimed-item check) are
the fleet's protection rather than one session's caution — do not weaken them, and
do not add a `FORCE=1` to a runbook as if it were the normal path.

## Where the imagery actually stands

Yesterday's loop run fixed more than the broken icons. **All ten of the site's
imagery work items completed** — 5 `needs_imagery` (tool page heroes) and 5
`unfulfilled_hero_variant` — routed to `image-build-handler`, which works
[MEASURED: `handler_agent='image-build-handler'` on this site, 10 rows, all
`complete`]. Imagery generation is not broken.

**What remains is five case-study images, and nothing in the framework will fix
them.**

```
404  /assets/images/case-study-facilities.jpg
404  /assets/images/case-study-legal-rag.jpg
404  /assets/images/case-study-private-ai.jpg
404  /assets/images/case-study-financial-data.jpg
404  /assets/images/case-study-logistics-strategy.jpg
```

They are detected — eleven `image_url_404` items — and every one is stuck:

```
status = blocked
error  = "No handler_agent set — item cannot be routed to any agent"
```

**This is by design and the design is what needs revisiting.** `check_image_url_404`
is deliberately flag-only: its header states that repairing a stale reference means
removing or repointing it, "which no image generator can decide". That reasoning is
right for a *stale* reference. It is wrong for this case, where five real
case-study cards want five real images that were simply never generated — a
condition `image-build-handler` handles competently ten times over on this same
site.

This is exactly the objection the council's `bug_historian` seat raised against
yesterday's checker fix, in its own words: *"The fix makes the silent loss VISIBLE
but does not make it ACTIONABLE."* It was right, and here is the case that proves
it.

**Note the duplication too**: five paths, eleven items. The pre-2026-07-31 wording
("Pages reference unknown image X") and the post-rewrite wording ("…but no active
asset deploys to that path") are both present for the same five files, plus one
`empty_src`. Worth checking whether the rewrite changed the `item_key` and so
defeated `idx_swi_dedup` — if it did, that is a second finding and it is
fleet-wide.

## Phase 1 — make the five images real, through the framework

**Do not hand-generate them and do not delete the references.** The cards are
legitimate content; the site is meant to show case studies.

The route that already works on this site is a `needs_imagery` item per image,
handler `image-build-handler`, which generated this site's tool heroes yesterday.
Three things to settle before queueing, each a query not an opinion:

1. **What `needs_imagery` requires in `spec`** — read a completed one from this
   site rather than inventing the shape; there are five to copy.
2. **What asset key each card expects.** The rendered path is
   `/assets/images/case-study-<slug>.jpg`, so the asset must deploy to exactly
   that path via `storage.DeployedWebPath`. Getting this wrong produces a
   generated image at a path nothing references — which looks like success and
   fixes nothing.
3. ~~**Whether the component or the content owns the filename.**~~ **RESOLVED
   2026-08-04, before queueing — and the answer is BOTH, which is why it was worth
   checking rather than assuming:**

   | page | component | paths in `html_template` | paths in `content_data` |
   |---|---|---|---|
   | `/index.html` | `case-studies-grid` | no | **yes** |
   | `/case-studies.html` | `case-studies-list` | **yes** | no |

   Two surfaces, two different shapes, **the same five filenames**. That is the
   convenient outcome: because `storage.DeployedWebPath(asset_key, purpose)`
   yields `/assets/images/<key>.<ext>`, five assets keyed
   `case-study-facilities`, `case-study-legal-rag`, `case-study-private-ai`,
   `case-study-financial-data`, `case-study-logistics-strategy` deploy to exactly
   the paths **both** components already reference. **One set of five assets fixes
   both pages**, and no repointing is needed on either.

   Two riders, both stated rather than assumed:
   - **[UNVERIFIED] that the generated extension will be `.jpg`.** Both surfaces
     reference `.jpg`; if the generator emits `.png`, `DeployedWebPath` produces a
     path neither component references and the repair silently fixes nothing while
     reporting success. **Check the extension before queueing, and verify at the
     served URL afterwards** — this is precisely the trap
     `bugs_open/142`/`bugs_closed/168` records for that helper.
   - **`case-studies-list` hardcoding image paths in its template is a latent
     defect of its own**, and the same family the `image_url_404` header names as
     the cause it exists to catch ("a component template hardcoding an image
     reference for which no asset was ever generated"). Generating the assets
     resolves the symptom on that page today; the template stays a landmine for
     the next site that mounts it. Worth filing separately, not fixing here.

Then queue, dispatch via the standing per-site trigger, and **verify at the served
URL** (`curl -o /dev/null -w '%{http_code}'`), not at the work item status.

## Phase 2 — the rest of the current design backlog

Still open on the site, and all of it should clear before the designer runs, so its
findings are about design rather than about breakage:

| item | n | note |
|---|---|---|
| `generic_theme` | 1 | unresolved after 2 attempts — site still on the default theme |
| `needs_design_review` | 1 | shares a style collection with other sites; wants its own |
| `undeployed_asset` | 1 | logo generated but never deployed |
| `deactivated_component` | 2 | needs_human_review — chrome points at deactivated components |

`generic_theme` and `needs_design_review` are the two that most directly shape what
a designer would see, and both have handlers (`webdesign-agent`). They have failed
before, so expect to read *why* rather than just re-firing.

## Phase 3 — the visual designer pass

The agent the owner means is **`visual-design-auditor`**: *"Group auditor for
visual design quality. Loads style collection and rendered HTML samples, runs
algorithmic checks for colour consistency and spacing, then makes one LLM call for
holistic visual assessment. Produces findings as work items."*

It is **not** dispatched directly. It is spawned by **`design-audit-agent`**, the
top-level design orchestrator, which also spawns `content-quality-auditor` and
triages what comes back into work items. And `design-audit-agent` is itself a step
inside the improvement loop (`call_design_audit` / `spawn_design_audit`) — it ran
yesterday, and it is the agent that independently spotted the `src="cpu"` defect in
prose.

**So the visual designer pass is not a new mechanism to build — it is the loop,
fired again, after phases 1 and 2 land.** That is a pleasant answer: the standing
per-site trigger already does it.

Two caveats worth stating so the run is not wasted:

- **Fire it once phases 1–2 are verified at the artefact**, not merely marked
  complete. The designer reads *rendered HTML*; if the images are still 404 its
  holistic assessment will be about missing images again, and we will have spent an
  LLM call to be told what we already know.
- **Expect findings that need a human.** Yesterday's run produced
  `needs_human_review` items about positioning, social proof and audience — real,
  and not automatable. The designer pass will add to that pile. That is the system
  working, not failing, but it means "run the designer" is not the same as "the
  design improves".

## What would make this plan wrong

Stated so it is checkable rather than persuasive:

- ~~If the five case-study paths turn out to be hardcoded in a component template,
  Phase 1 is the wrong shape entirely.~~ **CHECKED 2026-08-04 — see Phase 1 item 3.
  They are hardcoded on one page and content-driven on the other, and both name the
  same five files, so generating five correctly-keyed assets fixes both.** The
  residual risk moved rather than vanished: it is now the **file extension**, and it
  fails silently in the direction of looking successful.
- If `generic_theme` has failed twice for a structural reason (not transient), then
  re-firing the loop will fail a third time and Phase 2 needs a diagnosis run
  (`090`) rather than another attempt.

---

## CORRECTION 2026-08-10 — Phase 1 item 3's resolved table is WRONG, and it changes what Phase 1 achieves

> **CORRECTED.** Phase 1 item 3 states, marked **RESOLVED 2026-08-04, before
> queueing**:
>
> | page | component | paths in `html_template` | paths in `content_data` |
> |---|---|---|---|
> | `/case-studies.html` | `case-studies-list` | **yes** | no |
>
> **The `html_template` column is false.** `case-studies-list`
> (`content_components` `e7fc34f7-ef74-4665-830d-2c130c689002`, unedited since
> 2026-03-09, so it was false when written) contains **no `<img>` tag, no
> `.jpg`, and no `/assets/images/` at all**. It renders `{{.title}}`,
> `{{.client}}`, `{{.summary}}`, `{{.results}}` and nothing more.
>
> The check that produced "yes" was almost certainly
> `html_template LIKE '%case-study-%'`, which matches the **CSS class names**
> `case-study-item` / `case-study-client` / `case-study-results`. I repeated the
> error on 2026-08-09 with the same pattern. Full account: `WRONG_CALLS.md`
> 2026-08-10.

**What this changes.** The plan's convenient conclusion — *"One set of five
assets fixes **both** pages, and no repointing is needed on either"* — does not
hold. `/case-studies.html` was never a consumer. Phase 1 therefore fixes **one**
surface, not two, and only if that surface's content names the files.

**Who actually references `/assets/images/case-study-*.jpg` (measured 2026-08-10,
precise prefix, with a positive control):**

```sql
SELECT 'template', cc.name FROM content_components cc
WHERE cc.html_template LIKE '%/assets/images/case-study-%'
UNION ALL SELECT 'page content_data', p.url
FROM pages p JOIN page_components pc ON pc.page_id=p.id
WHERE pc.content_data::text LIKE '%/assets/images/case-study-%';
-- control: the same shape finds 19 rows for '/assets/images/content-hero-%'
```

**Exactly one row: a `case-studies-grid` on `/enterprise-reference-deployment.html`
— and that page 404s.** `/index.html`'s `case-studies-grid` *did* reference them
and lost the keys to a content regeneration (`bugs_open/238`).

**So, as of 2026-08-10: the five images exist and serve, and no live page shows
them.** That is the honest state of Phase 1 — the assets half is done and
verified, the referencing half is not, and the remaining work is content, not
imagery:

1. `bugs_open/238` — restore image URLs to `/index.html`'s grid, for the
   *rewritten* case studies (the same run changed which case studies the cards
   describe, so the old URL set is not a paste-back).
2. Decide what `/enterprise-reference-deployment.html` is: a page that should be
   built and deployed, or a stale row to retire.

**Phase 3 (the visual-designer pass) is still gated** and now on a firmer
reason than "the images 404": the homepage renders five `<img src="">`.
