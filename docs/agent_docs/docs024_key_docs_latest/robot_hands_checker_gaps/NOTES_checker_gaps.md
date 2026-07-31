# NOTES — robot-hands checker gaps (append-only, newest at the bottom)

## 2026-07-30 — the owner's two defects, and what measuring them exposed

Owner reported, from the live site: (1) the **Tools** nav link goes to a page with
no content; (2) the **Run MatchMatrix** card in the "Calculation and Comparison
Tools for Gripper Specification" component has no image while its three siblings
do. Brief was "check the site through the framework … we want the framework to be
able to pick them up."

Site `00ff3af5-dad8-4770-9f70-3edc267a3c92`.

### Defect 1 — /tools.html serves chrome only

`curl https://robot-hands.com/tools.html` → HTTP 200, 17,254 bytes, of which
**578 chars of visible text, all of it nav + footer**. Zero body content.

DB state (queried by `url`, not by id — see the trap below):

```
url         | status | build_status  | in_header | nav_label | planned_sections | components | deployed_at
/tools.html | active | needs_rebuild | t         | Tools     | 3                | 0          | 2026-05-10 17:37:02+00
```

`sections = ["hero","tool-list","call-to-action"]`, `rendered_header`/`_footer`
both NULL. So: the page is **active and in the header nav**, has three sections
planned, has **zero `page_components` rows**, and what visitors are served is the
artefact deployed **2026-05-10** — 81 days stale. It has been sitting at
`needs_rebuild` with nothing retrying it.

### Defect 2 — the MatchMatrix card image

`tool-list` on `/index.html` (pc `382ed98e-6529-42f7-bee3-917748fe516c`),
`content_data.items[0]`:

```json
{"name":"tool-matchmatrix","url":"/tools/matchmatrix/index.html","image":""}
```

The other three items carry real `/assets/images/card-tool-*.jpg` paths. Cause:
`tool-matchmatrix` is the **only** tool page with no `purpose='card'`
entity-linked asset, and `tool-list.input_schema` sources `items` from
`query.pages_where_type:tool` — so the resolver emits `image: ""` for a tool page
with no card. The empty string is **derived, not authored**: editing
`content_data` cannot hold it (same family as the derived-field trap recorded in
`robot-hands-site-fixes-workstream`).

The hero DID land: `content_hero_tool_matchmatrix`, active, created 2026-07-25
09:03. What never happened is the **derive** of the card from it.

## The three-link chain, all three measured

The useful finding is not either defect. It is that the framework's
detect → schedule → dispatch chain is **working in one link of three**.

### Link 1 — DETECTION: works, and is not the problem

`content_image_missing` is enabled on `design-discovery-agent` and is correct
about defect 2. Replaying its predicate verbatim, read-only, before touching
anything:

```
tool-grip-force-friction-calculator | content_hero ✓ | card ✓ | (fulfilled)
tool-gripper-cycle-time-estimator   | content_hero ✓ | card ✓ | (fulfilled)
tool-gripper-payload-calculator     | content_hero ✓ | card ✓ | (fulfilled)
tool-matchmatrix                    | content_hero ✓ | card — | derive
```

One finding, zero false positives, exactly the page the owner pointed at. Then I
fired the real sweep and it emitted precisely that:

> `needs_content_image | detected | asset-deployer | Article "tool-matchmatrix" has no card image — derive from its hero`

Pass 1 (generate the hero) ran 07-25 and completed. Pass 2 (derive the card) had
never run, because the design lane had not swept this site since **2026-07-24**.

### Link 2 — SCHEDULING: nothing drives the checker layer

Of **24 enabled `scheduled_tasks`, none targets `completeness-`, `design-` or
`quality-discovery-agent`.** The only discovery task ever created is
`oneshot-discovery-aao-20260726`: **disabled**, aimed at a different site, and
pointed at `system.agent.generic.requests` — the default topic **nothing
consumes** (18 of 18 working enabled tasks use `system.agent.scheduled.requests`).
So it never ran anything either.

The ~60 checks in `discovery_checks/` (57 enabled entries across the three
agents) therefore run **only when a thread fires them by hand**. That is why
robot-hands' design findings all dated 07-24 and its completeness/quality
findings 07-28, with nothing since.

> **Careful, and corrected against a prior thread's recorded mistake:** a
> disabled scheduler is **not** a dead subsystem (`WRONG_CALLS.md` ~L10600 — a
> thread called this package dead while two of three agents had produced 252
> items three days earlier by other routes). The claim here is narrower and is
> what I measured: **no enabled recurring task targets them**, so cadence is
> whatever a human supplies. Not "the checks never run".

### Link 3 — DISPATCH: findings land where no handler can reach them

Discovery writes items at `status='detected'`. `build-pipeline-trigger`'s
`pre_query` dispatches only `status='triaged' AND pipeline='build'`, and
`claim_work_item_action` filters `status IN ('triaged','approved')`. The bridge is
`triage_detected_items`, run by exactly three agents — `improvement-loop`,
`site-review-agent`, `design-audit-agent`. **None has an enabled scheduled task.**
`improvement-sweep` is disabled, last triggered **2026-05-02**, and on the dead
generic topic.

Fleet-wide consequence, measured:

```
detected  | 263 items | 9 sites | oldest 2026-07-14
triaged   |   0 items
```

**263 findings stuck since mid-July and not one item anywhere at `triaged`.** So
tonight's sweep will not self-repair the MatchMatrix card either — the item is
correct, queued, and unreachable.

## The detector gap behind defect 1 — three near-misses, three different reasons

No check can see an active in-nav page with planned sections and zero
components. The three that look like they should each miss for an unrelated
reason, which is why this survived:

| check | why it is blind here |
|---|---|
| `empty_sections` | iterates `FROM page_components pc` — there are **no rows to scan**. Replayed live for this site: **0 rows returned.** |
| `sectionless_pages` | requires `sections` NULL/`[]`; this page has **3** |
| `unresolved_sections` | requires `p.build_status='deployed'`; this page is **`needs_rebuild`** |

`empty_sections` is the instructive one: it is a *component*-driven detector and
the defect is component **absence**. An empty set is not an empty section.

**Empirical confirmation, not just SQL reading:** all three lanes fired tonight,
all 57 configured checks ran, and work items filed against
`ee3cfcfb-b27a-4dbb-8d91-d3f5131f2304` (`/tools.html`) from those sweeps =
**zero**.

Fleet census of the class:

```
gaswholesalers.com          | 4 shell pages | 0 in nav
ai-agent-orchestration.com  | 3             | 0
dartsonline.com             | 3             | 0
finetuning.uk               | 2             | 0
fundamentallyai.com         | 1             | 0
robot-hands.com             | 1             | 1  ← the only one a visitor can reach from a nav
```

14 pages / 6 sites. Only robot-hands' is user-visible, which is why the owner saw
this one and no previous thread saw the other thirteen.

## Missteps and traps this cost

- **I guessed the page UUID and it was wrong.** I typed
  `ee3cfcfb-d876-4b8b-8d91-d3f5131f2304`; the real id is
  `ee3cfcfb-b27a-4dbb-8d91-d3f5131f2304`. Same 8-char prefix, different body —
  a truncated-prefix eyeball match is not an id. Every load-bearing finding above
  was established `WHERE url='/tools.html'` or via `NOT EXISTS`, so nothing is
  contaminated, but one query silently returned rows for the wrong page.
  **Cheap check: resolve the id in the query, never carry it by hand.**
- **`\d` first, three times over.** `pages` has no `content_data` (it has
  `sections`); `page_components` has no `status` (it has `build_status`);
  `content_components` has no `site_id`; `scheduled_tasks` has no `agent_type`
  (it is `target_agent_type`). Four schema guesses, four errors.
- **`run_checks` is the STEP name; `run_discovery_checks` is the ACTION.** My
  first enablement query filtered on `action='run_checks'` and returned 0 rows —
  which would have read as "no agent runs any check", a much bigger and wholly
  false claim than the true one. The near-miss is the point: **a zero from a
  filter you invented is not evidence.**
- **`triage_detect_items` vs `triage_detected_items`** — the file is named
  `triage_detect_items_action.go` but the registered action is
  `triage_detected_items`. Grepping the filename's spelling finds nothing.
- **Equal `last_triggered_at`/`last_completed_at` proves nothing** — that is the
  normal fire-and-forget stamp. I confirmed the design sweep downstream instead:
  `orchestration_states` row `923f6cab`, COMPLETED, `site_record.domain =
  robot-hands.com`, plus the 12 items it wrote.

## Incidental findings, not acted on

- `claims_unverified` fired **6 times** on this site tonight (2 banned claims, 4
  unregistered-number sets) — notable given this site's 043 fabrication history.
  **2 of the 6 are on `/reports/` pages**, which per the gripper-dossier lane must
  run `check_claims:false` (every report figure is per-request, not in the site
  register). Those two look like false positives of the report-page kind already
  documented in that workstream. Left for its owner; not touched here.
- `capability_gap`: palette emits 1 unreadable pairing, `#1A1F2E` on `#0F1218` at
  **1.14:1** against a 3.0:1 floor.
- 3 `image_url_404` for `content-hero-tool-*-guide.jpg`, and 3 stale site
  components (head/footer/header).
- `triage_detected_items`' callers all set `target_domain`, which **nothing
  reads**; the key actually read is `target_pipeline`, set by zero definitions
  fleet-wide. Invisible only because every caller wants `"build"` and that is the
  default — the half-landed rename recorded as `bugs_open/136`.

## 2026-07-31 — the framework repaired defect 2, and then stopped one step short

Owner authorised promoting robot-hands' whole detected backlog. Promoted **90 of
94** rows detected → triaged (mirroring `TriageDetectedItemsAction`), deliberately
**excluding the 4 with an empty `handler_agent`** (3 `image_url_404`, 1
`capability_gap`): an item no handler can clear gets claimed, fails twice and is
relabelled "[unresolved after 2 attempts]", which asserts a handler FAILED when
none could ever have succeeded — the `bugs_open/077` class `remit.go` guards.
`handler_agent=''` is the canonical "no handler" since migration 217.

**It worked, up to the last step.** `asset-deployer` derived the card:
`card_tool_matchmatrix`, purpose `card`, entity-linked, `origin_asset_id` set (so
lineage is right), and the file is **live: HTTP 200, 46,149 bytes** at
`/assets/images/card-tool-matchmatrix.jpg`. Work item `complete`,
`attempt_count=0`.

**And the card still does not appear on the page.** `/index.html` was re-rendered
and deployed at **12:48:31**, an hour and a half AFTER the asset landed at
**10:54:42**, and `tool-list`'s `items[0].image` is still `""`, with zero
occurrences of `card-tool-matchmatrix.jpg` in the served HTML.

The reason is a distinction PBP-022 already draws and I had to rediscover: the
re-render that ran was a **light/assemble** one, which re-renders from EXISTING
`content_data`. The `items[]` array is **query-backed**
(`source: query.pages_where_type:tool`) and query sources are resolved by
`plan_sections` — i.e. **only a FULL rebuild re-derives it.** The resolver logic
itself is correct and would produce the right path: `pageImageJoins` joins
`assets` on `entity_type='page' AND entity_id=p.id AND purpose='card' AND
status='active'` (which the new asset satisfies) and `webPath()` prefers the card
key, returning `storage.DeployedWebPath(...)` — never `assets.url`, because that
holds an expiring presigned URL.

So this is **"a complete work item is not a repaired artefact"** with an extra
turn: the item was honestly complete, the asset genuinely landed and genuinely
deployed, and the *visible defect the owner reported is still there*. Verified at
four layers, which is what separated them: item status → asset row → live file →
served HTML. Only the fourth disagreed.

> **[UNRESOLVED, flagged not smoothed]** my promotion `UPDATE` reported **90**
> rows; **93** now carry that identical `triaged_at`. Not reconciled.

**The rebuild carries a known risk on THIS page.** A full rebuild regenerates
copy, and `robot-hands/index` is the recorded site of a live 043 recurrence — a
routine re-render re-invented "2,400+" four days after it had been corrected.
Migration 201 and the `evidence_base` writer_blocks exist to mitigate exactly
that, and they are live, but the honest position is that rebuilding this page is
not free of fabrication risk and its stat blocks need re-checking afterwards.
Not fired unilaterally for that reason.

## Corrections to my own earlier claims in this file

- **"ZERO at `triaged`" became my own footprint.** Another session read 119
  `triaged` rows as evidence that something now drives `triage_detected_items`.
  It was me. Re-measured: **every** `triaged` row fleet-wide is robot-hands, and
  `min(triaged_at) = max(triaged_at)` to the microsecond — one statement, one
  transaction. A running loop would leave a spread across sites and times.
  **`min=max` on a timestamp is the cheap discriminator between a batch write and
  a live process**, and I have corrected the memory note that had it wrong.
- **I planned to fork the `features` component and should not have.**
  `hero-card-carousel` (`82274d36`, `hero-carousel`, render_mode `agent`) already
  exists with exactly the needed shape — `cards[]` of image/image_alt/title/
  teaser/link_url/link_label, an `autoplay` default of false, a pause button and
  hover/focus pause, and llm_guidance grounded in the ~89% first-slide figure. It
  has **0 instances**: built, never exercised. "Reuse existing machinery before
  building new" would have caught this before I read the `features` template.
- **I nearly generated three carousel images straight into `bugs_open/155`.**
  LANDMINES.md already carries it: `resolveStorageURIFromAsset` Priority 1
  resolves the source by `sites.content_data->>'{purpose}_uri'` — keyed on
  **purpose only, never asset_id** — so N assets sharing one purpose all deploy
  as the SAME file, every one reporting `success:true` (proven live: 6
  byte-identical deploys, sha256 `e647f9fb…`). Three carousel images share a
  purpose by construction. **The landmine's instruction: pass `spec.s3_uri`
  explicitly as `s3://bucket/key` derived from that asset's own `storage_path`
  — NOT its `url` (`bugs_open/152`: already overwritten post-deploy) — then
  `sha256sum` the files and confirm they differ, and actually LOOK at one.**
  I found this via another session's SQL rather than by grepping LANDMINES first,
  which is the check I should have run before planning any image work.
