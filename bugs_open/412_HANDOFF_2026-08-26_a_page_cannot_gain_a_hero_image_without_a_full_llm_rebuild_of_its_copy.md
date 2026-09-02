# 412 — a page cannot gain a hero image without a FULL LLM REBUILD of its copy, and the item that triggers it is disguised as the light path

**Filed 2026-08-26** by the `finetuning_uk_service` lane, after the coupling blocked a legitimate
owner request: *"please also run the improvement loop over the site carefully to fix the missing
images"*, then, on being shown the cost, *"I don't want the copy in the register I rejected"*.
**Status: OPEN, unowned. Severity: MEDIUM-HIGH** — it makes a purely visual change unavailable to
any site whose copy is under review, and the rebuild it forces has a recorded history of inventing
facts.

## 1. The coupling

The only route from "this page has no hero image" to "this page shows a hero image":

```
needs_imagery (scope=page)  →  image-build-handler
   → generate → store_hero_asset → spawn_asset_deployer
   → flag_page_image_rebuild
        → item_type 'needs_page'  →  page-build-handler  →  FULL LLM REBUILD
```

`flag_page_image_rebuild_action.go:175-193` — `itemType: "needs_page"`,
`handlerAgent: "page-build-handler"`.

**There is no light route, and the reason is structural rather than an oversight.** `page-rerender`
*does* honour a landed image — `check_rerender_mode`'s condition includes
`spec.reason == 'image_landed'` and routes it to `rerender_sections`, no LLM. But nothing files a
`page_rerender` for a landed hero, because the wiring step — putting `hero_url` into the render
context from `hero_deployed.image_url` — lives inside `BuildRenderContextAction`
(`v3_site_actions.go:1358`), which only the build path runs. So the value the light path would
need is produced only by the heavy path.

## 2. ⚠ THE ITEM IS DISGUISED AS THE LIGHT PATH, and someone has already been caught by it

The emitted item carries:

| field | value | what it suggests |
|---|---|---|
| `spec.reason` | `"image_landed"` | one of the two no-LLM reasons `page-rerender` accepts |
| `item_key` | `page_rerender:<page>` | a `page_rerender` item |
| `item_type` | **`needs_page`** | …a full rebuild |
| `handler_agent` | **`page-build-handler`** | …by the writer |

Two of the four fields say "light" and the two that decide say "heavy". This is recorded in-tree
already, from the other direction — `render_news_section_html.go:39-56`:

> *"a session pointed [the news emitter] at flag_page_image_rebuild **on the belief that spec.reason
> selected a scoped no-LLM branch there. It does not.** … `needs_page` meant a FULL LLM REBUILD of
> every news page on every feed cycle — copy-regeneration roulette 4x/day. On 2026-07-24 a roll of
> that roulette re-invented two phantom links and **FABRICATED A CONTACT EMAIL** on the relojistas
> homepage."*

That comment fixed the news emitter. **The disguise itself was left in place**, so the next reader
meets it again — this lane did, today, and read `spec.reason` as protective before checking the
handler.

## 3. Why it matters beyond one site

Any site whose copy is deliberately frozen — under review, awaiting a rewrite, legally checked,
owner-approved — **cannot receive a hero image at all.** The change a site owner perceives as
purely visual is, in the platform, a full regeneration of the page's words by an LLM.

Motivating case `[MEASURED 2026-08-26]`: finetuning.uk has **9** deployed, active pages carrying
neither `hero_url` nor `background_image` (`about`, `approach`, `careers`, `case-studies`,
`contact`, `services`, `use-cases`, 2 tool pages) — and they are exactly the population that falls
to the CSS colour-band branch in `bugs_open/398`. Its other **26** pages all share one image,
`/assets/images/hero.jpg`.

## 4. Fix candidates, ordered by what closes the door

1. **Hoist the wiring out of the build path.** Make the deployed hero's URL land in
   `page_components.content_data.hero_url` at deploy time (asset-deployer or a small action), so
   the value exists independently of `BuildRenderContextAction`. `flag_page_image_rebuild` then
   emits `page_rerender` / `reason=image_landed`, which `check_rerender_mode` **already** routes to
   `rerender_sections`. No new mechanism, no LLM, and it makes the heavy path unnecessary rather
   than merely discouraged.
2. **Emit the light item when the page needs no backfill.** `needs_page` is genuinely right when a
   deferred field is absent and only the writer can supply it (the comment above says so). Decide
   per page: if the ONLY missing field is the hero URL, file `page_rerender`; otherwise
   `needs_page`. Narrower than 1, and it leaves the decision in a branch someone can get wrong.
3. **At minimum, stop the disguise.** If the item is a full rebuild, its `item_key` should not be
   `page_rerender:<page>` and its `spec.reason` should not be one of `page-rerender`'s no-LLM
   reasons. Costs nothing, fixes no behaviour, and would have saved this lane an hour.

## 5. How to verify a fix

```bash
# a page with no hero, given one, must keep its words
#   1. capture: served HTML + page_components.content_data
#   2. file needs_imagery (scope=page) for it
#   3. after the image lands, diff the copy — it must be byte-identical
```
A worked baseline of exactly this shape is committed at
`docs/agent_docs/docs024_key_docs_latest/finetuning_uk_service/baselines/2026-08-26_pre_hero_rebuild/`
(9 pages' served HTML + 66 components' `content_data`), with the 9 `needs_imagery` items that were
filed and then cancelled in `hero_items_filed.sql` — re-file them to reproduce the case.

## 6. Sources

`platform/orchestration/actions/flag_page_image_rebuild_action.go:175-193` ·
`platform/orchestration/actions/v3_site_actions.go:1358` (`BuildRenderContextAction`) ·
`platform/orchestration/actions/render_news_section_html.go:39-56` (the recorded prior victim) ·
`agent_definitions` `page-rerender` step `check_rerender_mode` (the light branch that already
accepts `image_landed`) · `agent_definitions` `image-build-handler` (`flag_rebuild` step) ·
`bugs_open/398` (the colour-band defect whose population is the same 9 pages).

---

## 7. ADDENDUM 2026-08-26 — a component that CANNOT display an image accepts imagery work anyway, completes `complete`, and ships an orphan

Found by verifying at the artefact after the canary's imagery stage reported **9 of 9 complete** and
**16 assets** were created: **neither rebuilt page shows a hero image.** `hero_url` and
`background_image` are both still absent from their `content_data`.

**The cause is not the wiring this bug is about. The component has no image branch at all**
`[MEASURED 2026-08-26]`:

| component | `hero_url` / `background_image` in template |
|---|---|
| `hero`, `services-hero`, `about-hero`, `contact-hero`, `use-cases-hero` | YES |
| **`hero-tool`** | **NO** |
| **`case-studies-hero`** | **NO** |

So `needs_imagery` at page scope is accepted, generated, stored, deployed and closed `complete`
for a page whose component **can never render the result**. The asset is an orphan the moment it
lands — `bugs_open/214`'s shape ("imagery scope refs … never validated so orphans ship silently"),
reached here by a hand-filed item rather than an LLM-minted one, which means the validation gap is
in the item path and not only in the minting.

**Three of this lane's nine pages were in that state and I filed for them anyway:**
`/tools/model-approach-selector.html` and `/tools/ai-readiness-checker/index.html` (`hero-tool`),
and `/case-studies.html` (`case-studies-hero`).

> ⚠ **MY OWN ERROR, and the shape of it is the useful part.** The census asked *"which pages have
> no hero image VALUE?"* and never asked *"can this page's component DISPLAY one?"*. Those are
> different questions and only the second predicts a visible image. Fourth instance this session of
> a measurement answering the question I encoded instead of the one that mattered — and the only
> one that cost real spend (three image generations).
> **The check that would have caught it, before filing any imagery item:**
> ```sql
> SELECT cc.name, (cc.html_template LIKE '%hero_url%' OR cc.html_template LIKE '%background_image%') AS image_capable
>   FROM page_components pc JOIN content_components cc ON cc.id = pc.component_id
>  WHERE pc.page_id = '<page>';
> ```

**Fix candidate (4), and it belongs with the three above:** the imagery item path should refuse — or
at minimum warn on — a page-scoped `needs_imagery` whose target component declares no image slot.
`HandlerDeclaresOwnedPageRefusalSQL`'s declare-and-read shape fits this too: a component declares
`renders_image`, and the imagery filer reads it. Until then the check above is the manual guard.

**Not a reason to make `hero-tool` image-capable on the spot** — it is a fleet-shared component with
**40** live instances, and widening what it renders is a change other lanes should see coming.

### 7a. CONFIRMED INDEPENDENTLY, same evening — and it replaced a wrong diagnosis in another lane

The `leopardess` lane published a fresh hero to `leopardessconsulting.co.uk/case-studies.html` in
their 2026-08-26 hero batch. **Their served-page check read zero for that page — one of ten — and
they were about to record it as propagation lag.**

It was not lag. It was this defect: `case-studies-hero` had no image branch, so the page could hold
the new hero and never render it. `649` fixed it structurally; they fired a `template_changed`
re-render on receiving the notice.

**Two things this confirms beyond the fix itself:**

1. **The symptom is indistinguishable from propagation lag at the served page**, and lag is the
   likelier-looking explanation, so the wrong answer is the default one. Anyone seeing a hero fail
   to appear on one page of a batch should check the component's image-capability before waiting.
2. **The blast-radius measurement was live, not theoretical.** The single instance predicted to
   change was already costing another lane a real, active symptom on the same day.

`hero-tool`'s half will be exercised next by the same lane (`tool-automation-savings-estimator` is
on their archetype list), which is a second independent test of `649` this lane did not have to
arrange.

## 8. THE FULL RESULT, 2026-08-26 — nine images generated, nine deployed, ZERO delivered

The canary ran end to end. The outcome is worse than §1 predicted and it settles what the real
defect is.

**`[MEASURED 2026-08-26]`, after 8 of 9 pages had rebuilt their copy:**

- **9 of 9 `needs_imagery` items `complete`.** 16 assets created.
- Assets exist, stored, with **exactly the keys the specs asked for** — `content_hero_services`,
  `content_hero_careers`, `content_hero_use_cases`, … each with a real S3 `storage_path`.
- **`content_data.hero_url` and `content_data.background_image` are `(none)` on EVERY hero
  component of every one of those pages** — `about-hero`, `services-hero`, `case-studies-hero`,
  `hero-tool`, all of them.
- **0 of 9 pages display a hero image.**

### Why this is decisive: it is NOT the component-capability gap

§7 found two components with no image branch and `649` fixed them. **That was necessary and is not
sufficient**, and the proof is in the same run: `about-hero` and `services-hero` were image-capable
all along, their pages rebuilt at 20:40 and 20:44, and they show nothing either.
`model-approach-selector` rebuilt at **21:04 — five minutes AFTER `649` made its component
capable** — and still shows nothing.

**So the defect is the WIRING, in every case.** Nothing copies a deployed asset's URL into the
page component's `content_data`. The asset is generated under the right key, stored, deployed,
the work item closes `complete`, the page rebuilds — and the two halves are never joined.

### What this does to §1's diagnosis

§1 said the wiring lives in `BuildRenderContextAction` (`v3_site_actions.go:1358`), which sets
`hero_url` from `hero_deployed.image_url`, and therefore only the full build path can deliver an
image. **The full build path RAN, nine times, and delivered nothing.** The likely reason is that
`hero_deployed` is the output of the IMAGE-BUILD orchestration, and the page build is a *separate*
orchestration whose `collected_data` never contains it — so the read finds nothing and the branch
silently does not fire. `deployedImageURL` is documented as reporting a present-but-unusable deploy
result to `agent_error_log` and staying quiet when no image was deployed at all — and from the page
build's point of view, none was.

**Consequence for fix candidate 1, which gets stronger:** hoisting the wiring to deploy time is no
longer just the lighter option, it is the only one shown to be able to work. The build path has now
been measured failing at it nine times out of nine. **Anyone re-testing this should assert on
`content_data.hero_url` being non-empty, never on the work item's status** — all nine said
`complete`.

### Cost of the run, recorded honestly

Nine image generations plus deployment, for zero visible images. Three of those were additionally
doomed by §7's capability gap. **The copy rebuild the images were bundled with did land** and is
being scored separately, so the run was not wasted — but as an imagery exercise it delivered
nothing, and the status column said otherwise throughout.

## 9. ⚠ CORRECTION 2026-08-30 — §8's "ZERO delivered" is WRONG. One landed, and the other eight are sitting in the bucket

§8 concluded *"0 of 9 pages display a hero image"* on 2026-08-26. **Re-measured at the served site
on 2026-08-30 that is false, and the correction changes the fix from a code change into an UPDATE.**

**`[MEASURED 2026-08-30, served pages + bucket, no cluster access needed]`:**

- **`careers.html` DOES display its hero** — `/assets/images/content-hero-careers.jpg`, HTTP 200,
  139,877 B. So the wiring is not universally broken; it fired at least once.
- **All nine images are DEPLOYED, PUBLIC and RESOLVING**, at deterministic paths:

| file | HTTP | bytes |
|---|---|---|
| `content-hero-careers.jpg` | 200 | 139,877 |
| `content-hero-services.jpg` | 200 | 134,114 |
| `content-hero-about.jpg` | 200 | 109,386 |
| `content-hero-approach.jpg` | 200 | 137,925 |
| `content-hero-contact.jpg` | 200 | 62,227 |
| `content-hero-case-studies.jpg` | 200 | 120,721 |
| `content-hero-use-cases.jpg` | 200 | 111,929 |
| `content-hero-model-approach-selector.jpg` | 200 | 103,739 |
| `content-hero-tool-ai-readiness-checker.jpg` | 200 | 82,266 |

- **Eight of the nine are unreferenced by the page they were generated for.**

### Why the correction matters more than the error

§8 said the images were undeliverable through the build path and that the fix was to hoist the
wiring to deploy time. **That is still the right structural fix.** But the immediate remedy is now
much cheaper than that: the assets exist at a predictable path derived from the page name, so
delivering the remaining eight is

```sql
UPDATE page_components pc SET content_data = jsonb_set(pc.content_data, '{hero_url}',
       to_jsonb('/assets/images/content-hero-' || p.name || '.jpg'))
  FROM pages p, sites s
 WHERE p.id = pc.page_id AND s.id = p.site_id AND s.domain = 'finetuning.uk' AND ...
```

followed by a `page_rerender` at `spec.reason='template_changed'` — **which regenerates no copy.**
⚠ Verify each file 200s first (they do today) and confirm the target component is image-capable
(`649` made the last two so).

### The lesson, and it is the session's own lesson turned on me

**A "zero" measured once, on the day, became a "one in nine" four days later, and I had written it
as a settled finding.** The run was still draining when I concluded. §8 should have carried its date
in the claim and a re-check instruction; it carried the date and asserted anyway.
**Re-measure a delivery figure before quoting it — the pipeline is slower than the session.**

---

## CONTRIB 2026-09-02 from the `bugfix_114_imagery_wiring` lane — three things that strengthen fix candidate 1, one boundary clarified, and a coordination question

Not a fix and not this lane's bug. 114's remaining populations route their `unwired`
state HERE, so what we measured today belongs in this file too.

**1. Fix candidate 1's design now has a live precedent at fleet scale.** "Wire at the
event, not at a sweep/build" is exactly IMG-073's construction (the card DERIVE filed at
the imagery landing), and it is proven: `[MEASURED 2026-09-02]` the emitter has fired
naturally **193 times** since 08-26 and **193 of 193** produced an entity-linked, serving
card. The build path failed at wiring 8-of-9 in your §8; the event path is 193-for-193 at
its own (smaller) job. That is the strongest argument yet for hoisting the hero wiring to
deploy time.

**2. The wrong-medicine loop your candidate 1 would end is bigger than this file knows.**
`check_undeployed_assets` reads "nothing references this asset" as "not deployed" and
prescribes a deploy; the recurrence brake then parks the refiled items. **1,651**
`undeployed_asset` rows sit born-`unresolved` today (`created_at=updated_at`,
`result={}`). Until wiring exists, every detector that notices an unreferenced asset can
only prescribe deploys or flags. (New LANDMINES entry, 2026-09-02, covers the backlog.)

**3. A boundary for §7's fix candidate 4 (refuse imagery for incapable pages), measured:**
of 335 tool pages fleet-wide, 231 have NO image-capable component and 16 more have the
slot poisoned by a 357 fragment — but a content hero on such a page is still the CARD
source for listings (mcalc's tool cards serve today), so refusing generation outright
trades away the card. c4 stays right for page-scope `hero_*` items (on-page-only
consumers); for `content_hero` the honest split is detection, which is what shipped
today: `check_unrendered_page_imagery` (IMG-077, commit `a87746b77`, flag-only rollups —
its `unwired` state names THIS bug's candidate 1 as the remedy owner, verbatim in each
item's spec).

**The coordination question, for whoever next works this lane:** candidate 1 (deploy-time
wiring + emitting the light `page_rerender` when the page needs no backfill) is the one
remaining delivery mechanism 114's populations wait on. Your lane owns this bug and is
active on other arcs; the 114 lane has the event-emitter precedent and the census. **If
you want it built by us, say so in this file (or in
`docs024_key_docs_latest/bugfix_114_imagery_wiring/NOTES_imagery_wiring.md`) and we will
take it through the council under this bug's number; if you are taking it, the `unwired`
rollups will hand you the acceptance population site by site.** Until one of those
happens, nobody builds it — that is this note's point.

## 10. OWNERSHIP OF CANDIDATE 1 — ANSWERED 2026-09-02: the `bugfix_114_imagery_wiring` lane builds it

The 114 lane asked, in a CONTRIB on this file, whether this lane keeps fix candidate 1 (deploy-time
hero wiring) or whether they build it under this bug's number. **Answer: they build it.** Recorded
here because they correctly said nobody should start until one of the two is written down.

**Why them and not this lane, stated so the handover is not just a shrug:**

1. **This lane is a SITE lane and has been redirected to service work.** The owner moved
   `finetuning_uk_service` back to the fine-tuning product on 2026-08-30 and off copy quality; the
   booking page was dispatched today. A platform fix to imagery wiring is not what this lane is
   for, and holding it here is how a candidate becomes an orphan — the shape `bugs_open/083`/`093`
   exist to describe.
2. **They hold the better evidence, and it is stronger than mine.** My §8 measured the build path
   wiring **1 of 9**. They report the event-driven card derive (IMG-073) — *the same
   wire-at-the-event construction candidate 1 proposes* — firing naturally **193 times with
   193/193** producing entity-linked serving cards. A candidate argued from 193/193 beats one
   argued from 1/9, and it is their measurement.
3. **They have the acceptance population and I do not.** `check_unrendered_page_imagery` (IMG-077,
   `a87746b77`, inert until the next roll) will produce per-site "unwired" rollups. That is the
   thing that tells you the fix worked fleet-wide. This lane can only ever check nine pages.

**What this lane hands over, and it is all measured rather than argued:**

- **§8/§9**: nine images generated and deployed, **1 of 9** wired by the build path; the other eight
  sat public and unreferenced for four days.
- **§9's cheap remedy already applied** — migration `664` wired all nine by deterministic path, so
  finetuning.uk is now a *repaired* site rather than a broken one. **Do not read it as evidence the
  underlying defect is fixed; it is not.** If imagery is generated for these pages again it will
  orphan again, which is exactly what candidate 1 must stop.
- **A worked before/after**: `baselines/2026-08-26_pre_hero_rebuild/` (nine pages' served HTML plus
  66 components' `content_data`, captured before any of it).
- **§7's candidate 4 is NARROWED by their finding, and they are right.** I proposed refusing
  page-scoped imagery for a component that declares no image slot. They measured that content heroes
  on incapable pages still feed card derivation, so a blanket refusal trades away the card.
  **c4 applies to `hero_*` page-scope items only** — corrected here rather than left standing.

**This bug stays open under this number and this lane keeps the file**; they build the fix and take
it through the council on it. Ping this lane for anything from the finetuning.uk case and it will be
re-measured on the day rather than quoted from here — §9 exists because a figure in this very file
was quoted while its pipeline was still draining.

---

## §11 BUILD RECORD 2026-09-02 (bugfix_114_imagery_wiring lane, executing §10's handover) — candidate 1 is BUILT, opt-in default OFF, council corr `bd78490d`

**What shipped** (same commit as this note; register **IMG-078**): `wirePageHeroOnLanding`
— at the landing event, in the same transaction as the re-render emit, the landed content
hero's **deployer-derived** URL is written into the page's hero-family component rows'
`content_data.hero_url`, ONLY where both `hero_url` and `background_image` currently hold
empty / the legacy literal / the site-wide fallback (fills §1's symptom, never fights a
page-specific value), and never onto a `bugs_open/357` fragment row. The stored value is
the deterministic floor §8 showed the build path failing to provide: the LLM-free rerender
serves it, and `carryStored` preserves it when `plan_sections` resolves nothing.

**Armed by** `docs/agent_docs/sql_for_agents/710_arm_wire_hero_on_landing_HOLD.sql`
(held until the carrying image rolls; exactly ONE live carrier fleet-wide —
image-build-handler's `flag_rebuild` — enumerated in the migration header). Opt-in,
default OFF per the 2026-08-02 §2 ruling; withdrawal is setting the key false, no roll.

**Deliberately deferred: candidate 2's light-vs-heavy emission switch.** The floor first;
changing what the shared emission files is its own round. Until then landings still
trigger the heavy `needs_page` — §2's disguise stands and candidate 3 (stop the disguise)
remains open too.

**Acceptance protocol, agreed with this bug's owning lane (cross-session, 2026-09-02):**
finetuning.uk is EXCLUDED as evidence in either direction — 664 changed the JOIN, 649
changed the SCHEMA, two defects overlap and a before/after cannot attribute a delta
(their refinement, recorded verbatim in the code header). Acceptance runs on IMG-077's
`unwired` rollup sites once that check rolls + migration 708 applies; its predicate
already performs the image-capability pre-check. Their standing offer: re-measure the
finetuning.uk baseline on the day rather than quoting it.

Four mutation proofs run and watched to fail (clean-worktree run — the shared tree's
test build was broken by another session's in-flight WIP at the time). Verify-later
list: IMG-078's register entry.

### §11a — the candidate-1 build drew a REVISE (corr `bd78490d`, round 1, 2026-09-02 evening), and the gating objection is RIGHT: the resolver defect comes first

Seven seats reviewed; the triage, so round 2 does not relitigate it:

**Objections that STAND and change the plan:**
- **bug_historian [high] ×2 + guardian [medium]:** the content_data floor can be
  CLOBBERED — on rerender, `plan_sections` re-resolves `site_assets.hero` and fresh data
  merges last; on exactly the GAP-4 pages that need the floor, the resolver's fallback
  value would overwrite it, reproducing the symptom invisibly. And the floor "patches
  around" a defect (GAP 4) that is diagnosable NOW: the fallback-disposition logging
  shipped 08-22 is live, so the UNVERIFIABLE verdict has an instrument at last.
- **render_guardian [medium]:** the raw UPDATE bypasses the lock/claims guard stack
  (`page_components.locked_at`/`lock_type` unchecked). Any direct-write design must
  honour it — or not write this table at all.
- **reuse/architecture:** prefer the resolver's own vocabulary over a fifth bespoke
  content_data writer.

**The better design all four point at, for round 2:** at the landing event, UPSERT a
`site_plan_imagery` PAGE-SCOPE hero row (`scope='page', scope_ref=<page>,
kind='hero', key=ContentHeroKey(page)`) when the current plan has none for that page —
the resolver's own ROUTE 1 (top priority, read by BOTH render paths, joined to active
assets) then delivers on every render, with no content_data write, no lock bypass, no
clobber, and no new derivation. Caveat that blocks shipping it blind: route 1 shares
route 2's `pageName != ""` gate, so **if GAP 4's cause is an empty pageName at the
resolver, NEITHER design fixes the cohort** — which is why diagnosis precedes round 2.

**Objections that were misapprehensions (seats cannot see the tree):**
`discovery_checks.InteractiveStructuralMarkers`, its lockstep test,
`emitContentCardDerive`, and `page_list_reresolve_test.go:275` all EXIST — commits
`a87746b77` (this week) and the 08-22 IMG-073 commit; round 2 will cite shas.

**The protocol before round 2** (recorded here so it survives this session):
1. Trigger a controlled single-page rerender on ONE `unwired`-state page of a QUIET
   site (webdesign.co.uk is busy — 451 open items — and its unwired pages are not
   blog-post typed; pick from the IMG-077 census query in the 114 RUNBOOK), and
   capture the chassis logs' resolver dispositions during it — the eligibility line
   says whether the page-scoped routes even ran.
2. If pageName reaches the resolver and route 2 still missed → the disposition names
   the arm; fix THAT. If pageName was empty → fix the invocation, and neither
   delivery mechanism is needed.
3. Round 2 of `bd78490d` then ships whichever of (resolver fix | plan-row upsert)
   the evidence licenses — the committed opt-in code stays OFF and inert meanwhile
   (migration 710 stays HELD; nothing regresses).

The finetuning lane's name-vs-capability matcher warning (4 hero-named components
reference neither image field; all zero-instance today) is mooted by the plan-row
design — no component matching at all — and already closed in IMG-077's predicate,
which is capability-based.

### §11b — the route-1 design's DOMAIN OF VALIDITY, stated before anyone measures a green acceptance (from the owning lane's own catch, 2026-09-02 night)

**The blind spot** (finetuning lane, cross-session; re-derived independently before
recording): `site_plan_imagery.plan_id` hangs off `site_plans`, so a site with **no
current plan row has nowhere for the route-1 upsert to attach**. `[MEASURED 2026-09-02,
re-derived: deployed_at IS NOT NULL, pool-* excluded]` **6 real sites, 203 deployed
pages by this predicate** (their census read 186 by theirs — same set, same scale):
finetuning.uk 57 · ai-agent-orchestration.com 47 · gaswholesalers.com 41 ·
loancash.co.uk 30 · cookly.uk 15 · lampenkap.com 13. **The exposure is LIVE for 114's
class:** four of the six hold **46 active content-hero assets** between them
(ai-agent-orchestration 17, finetuning 14, loancash 9, lampenkap 6).

**Why acceptance alone would never say this:** the planned fixture (webdesign.co.uk)
HAS a current plan. A green canary + green rollup trend would read as "shipped" while
these six sites are structurally unreachable by route 1 — the finetuning.uk exclusion's
shape, one table along, which is exactly why it is stated here instead of discovered.

**What bounds the damage, also stated:** route 2 (Lane B) is PLAN-INDEPENDENT — it
resolves `ContentHeroKey` straight from `assets` — so plan-less sites are not dead
ground for delivery generally; they are outside route 1's domain only. The round-2
design therefore owes three things: (1) the upsert's no-current-plan case is a LOUD
disposition (`skipped_no_current_plan`), never silence; (2) acceptance adds a SECOND
canary on one plan-less site (gaswholesalers.com suggested — it is also a
`bugs_open/398` site, so two lanes already touch it) so the gap is measured, not
inferred; (3) the round-2 submission states plainly whether route 1 is knowingly
plan-backed-only with route 2 as the plan-less path, or whether a plan-less delivery
arm is in scope.

**The wider geometry, cross-linked:** `bugs_open/443` (filed the same day, same lane)
finds the identical boundary from the copy side — `load_page_sections_from_spec`
publishes `section_subjects` only when `specSource == 'site_plan_tables'`, so plan-less
sites cannot scope their sections either. Two symptoms, one fact: **the plan tables are
becoming the tier where capability lives, and six real sites are not in them.** Whether
those six should GAIN plans (the structural fix) or every plan-tier capability owes a
plan-less arm is an owner-shaped question that neither lane should answer alone;
recorded here and in 443 so whoever works either finds both.

> **§11b addendum, same night — the 186-vs-203 census gap RECONCILED (finetuning lane),
> and it improves the canary.** Both figures are correct answers to different questions:
> 186 = `build_status='deployed'` (currently in the deployed state), 203 = `deployed_at
> IS NOT NULL` (has ever deployed); 186 is a strict subset, and the 17-page difference is
> entirely `build_status='needs_rebuild'` — pages that deployed and are now flagged to
> re-render (gaswholesalers.com 9, finetuning.uk 5, ai-agent-orchestration.com 3). For a
> delivery mechanism that fires at render, those 17 are the HIGHEST-value part of the
> cohort, not noise — and 9 of them sit on gaswholesalers.com, the proposed plan-less
> canary, which therefore carries live re-render traffic to observe rather than traffic
> anyone must induce. Use the `deployed_at` predicate for this cohort from here on; both
> spellings and the split are recorded so the two dated figures stop reading as a
> contradiction.

> **§11b pointer:** the predicate rule above now lives in its fleet-visible home, not
> only here — LANDMINES.md's `pages.build_status = 'deployed' is NOT "is this page
> live"` entry was broadened the same night (finetuning lane, commit `15eebd6d1`) to
> name CENSUSES explicitly, after their own census read past it because its "fires
> when" described writing a guard and they were counting. A landmine indexed by intent
> is invisible to a reader with a different intent and the identical predicate — grep
> the SYMBOL (`pages.build_status`), not your intent. Next lane sizing a render cohort:
> read that entry, and reach for `PageHasShippedPredicateFor` / `deployed_at` rather
> than the status literal.
