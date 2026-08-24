# 206 — `entity-directory`/`entity-page`/`section-index` pages have no builder that produces real sections; "the machinery is proven live" (vetcomparison PLAN, 07-26) was an unverified inference, now falsified

**Filed 2026-08-06**, from `features_open/021`'s workstream while following up on an owner
request to build `vetcomparison.uk`'s practice-directory page **through the framework**, not by
hand. Diagnosed first-hand (code read at HEAD + live DB queried directly) rather than via the
090 loop — declaring the substitute per the 2026-07-31 ruling: the claim below is falsified by a
direct query (zero pages using the relevant component, quoted below), not inferred, and the
closest prior claim on record (quoted below) is shown wrong by the same query.

## The ask that surfaced this

`vetcomparison.uk`'s homepage links to "Search the directory" (`/directory/index.html`), which
has never built. The underlying `site_work_items` row (`715ec305-...`, `item_type='needs_page'`,
`page_name='directory-index'`) has sat `needs_human_review` since 2026-07-17. A session in this
site's own workstream reviewed it on 2026-07-26 and recommended:

> "**Reuse, do not build**: the entity-page machinery is proven live — relojistas.com has 8
> `entity-page` + 1 `entity-directory` deployed, vonc.com 8." — `PLAN_2026-07-26_site_strength.md:269`

Taking that at face value would mean this page is a trivial re-trigger away from building. It
isn't — the claim conflates "this `page_type` label has been used before" with "the specific
data → component pipeline vetcomparison needs is proven." Checked directly, it is not.

## What IS live (the data half — genuinely proven, no correction needed)

`directory_export_action.go` + the `directory-export-json` scheduled task (enabled,
`interval_seconds=172800`, `last_triggered_at = last_completed_at = 2026-08-04 10:57:05Z`) export
vetcomparison.uk's veterinary practices to `data/vet-full-index.json` in the site's git repo,
generic across verticals by design (its own header: "Serves any comparison vertical/domain —
nothing site-specific may be hardcoded here"). Live data behind it:

```sql
SELECT count(*), count(*) FILTER (WHERE verification_status='verified') AS verified
FROM business_intel.businesses b JOIN business_intel.business_verticals v ON v.id=b.vertical_id
WHERE v.slug='veterinary';
-- 3419 total, 2337 verified (checked 2026-08-06)
```

## What is NOT live — the falsified claim, with the query that falsifies it

A `directory-listing` component exists in the catalogue (`content_components`, function
`directory-listing`) — the plausible renderer for the exported JSON. It has **never been used on
any live page, anywhere in the fleet**:

```sql
SELECT s.domain, p.name FROM pages p JOIN sites s ON s.id=p.site_id
WHERE p.sections @> '"directory-listing"'::jsonb;
-- 0 rows (checked 2026-08-06)
```

The pages the 07-26 claim cited as proof do not exercise this pipeline at all:

```sql
SELECT s.domain, p.name, p.page_type, p.sections FROM pages p JOIN sites s ON s.id=p.site_id
WHERE s.domain IN ('relojistas.com','vonc.com') AND p.page_type IN ('entity-directory','entity-page');
```
- `relojistas.com/glosario-index` (`entity-directory`): sections `["hero","archetype-grid"]` — a
  glossary grid, LLM-authored, no external data feed.
- `vonc.com`'s 8 `entity-page`s: sections `["hero","content-block-about","call-to-action"]` —
  plain generic content (persona bios), same shape as any other content page.

Neither reads from a per-site exported dataset or uses `directory-listing`. The label
`entity-directory`/`entity-page` has been assigned to pages before; the mechanism vetcomparison
needs (real external entities → a real listing component) has not.

## Root cause: two gaps in the same map, read at HEAD

`platform/orchestration/actions/load_work_item_actions.go` (the code that decides, when a
`needs_page` item is minted, which handler builds it):

```go
availableBuilders := map[string]builderInfo{
    "content": {...}, "index": {...}, "landing": {...}, "blog-index": {...}, "blog-post": {...},
    // Add here as builders become available:
    // "entity-directory": {handler: "directory-build-handler", itemType: "needs_directory"},
    // "entity-page":      {handler: "entity-page-build-handler", itemType: "needs_entity_page"},
}
unavailableBuilders := map[string]string{
    "tool": "tool-builder", "entity-directory": "directory-builder", "entity-page": "entity-page-builder",
}
```
`entity-directory`/`entity-page` are explicitly, in-code, marked as builders that don't exist yet
(`directory-build-handler`/`entity-page-build-handler` — named, commented out, never implemented).
The live `715ec305` row predates this map (created 2026-07-17) with `handler_agent` defaulted to
generic `page-build-handler`, which is why it wasn't silently deferred — it was dispatched, and
correctly no-op'd instead (`page-build-handler no-op: no sections ready to build`, quoted in full
in the row's `error` column) because nothing had ever populated its plan.

The second gap is upstream of the first: `apply_gap_plan_action.go`'s `defaultSectionsForPage`
(the generic gap-planner's section-choosing fallback) has no case for `entity-directory`,
`entity-page`, or `section-index` — it falls through to `["hero", "generic-text-block",
"call-to-action"]` for all three, which would not render a real directory, a real entity, or a
real index of sibling pages; it would render filler text with the right page_type label. This
means **`guides-index` (`page_type='section-index'`) has the same underlying gap**, even though
it isn't in `unavailableBuilders` and was separately recommended as the "cheapest" of the three —
cheap to trigger, but nothing today would make its sections actually list the three live guide
pages rather than generate generic prose.

## Impact

Three `vetcomparison.uk` pages parked `needs_human_review` since 2026-07-17 cannot be built
through the framework as it stands: `directory-index` (`715ec305`), `practice` (`3cce980c` —
separately on HOLD pending a different blocker, P1's company-number crawl), `guides-index`
(`2f50bfda`). Likely affects any other site on the fleet carrying these `page_type`s with a
similar expectation (not surveyed here — scope this before generalising further).

## Fix candidates, not yet attempted

1. **Implement the commented-out `directory-build-handler`** — given a page with
   `page_type='entity-directory'` and a sibling `directory_export_json` artefact for the site,
   declare `["hero", "directory-listing"]` sections and confirm the component's expected
   `content_data` shape matches what `directory_export_action.go` actually writes (unverified
   here — the binding between the two has never been exercised, so don't assume the shapes agree
   without checking). Same shape of work for `entity-page-build-handler`, and a `section-index`
   case in `defaultSectionsForPage` that lists sibling deployed pages instead of the generic
   fallback.
2. **Narrower**: a one-off, reviewed action that plan-scaffolds sections for specific named pages
   — still needs the same `content_data` shape-matching, and per this repo's platform-seams
   ruling this is new capability (a builder that did not exist gets invoked), not a one-page
   config tweak — route through the council gate rather than a hand SQL edit, which is exactly
   the "every site goes through the framework" ruling's target.

**Deliberately not done**: did not hand-build the page (would bypass evidence_base/claims gating
on a site with documented legal remediation history — see
`docs024_key_docs_latest/vetcomparison/PLAN_2026-07-26_site_strength.md` "Constraints"), did not
re-trigger `page-build-handler` (would no-op identically and spend `attempt_count` 2 of 3 for
nothing), did not touch `practice` or `tool-compliance-deadline-calculator` (separate, already
correctly triaged by the 07-26 review).

## Related

- `docs024_key_docs_latest/vetcomparison/PLAN_2026-07-26_site_strength.md` — the stale claim,
  corrected in place (see file) rather than silently.
- `features_open/021` — the operator rebuild path used this session to prove `page-rebuild`
  works; unrelated pipeline, REBUILD-only, cannot create a new page regardless of this bug.
- `bugs_closed/001` — general re-plan risk for this site; not the mechanism here (this is about
  a single unplanned page's sections, not a full re-plan).

---

# CLOSURE EVIDENCE 2026-08-08 — fixed AND live, both pages serving; file stays in bugs_open/ by owner direction (2026-08-06)

**The fix**: `directory-build-handler` (fix candidate 1, as designed in the lane PLAN) —
`ensure_page_section_layout` + `queryresolve.resolveBusinessDirectory` + the builder-map flip,
council **APPROVED round 3** (corr `5b8e4cf7-31c3-4793-a550-d6b9be1f00e8`, 09:18Z), code live
on v1.0.1264→1266 (pod-grepped both replicas each roll, negative control 0), config via
migrations **325, 326, and two live-fire corrections 336/337** (326's delegation input_mapping
was defective twice over — prefixed keys, then missing spec/current_page; each found by a real
dispatch failing, each fixed by migration same-day; the seed 326 alone does NOT match the live
row).

**Proof at the artefacts, not the statuses** (both built by ordinary `build-pipeline-trigger`
dispatch of the ORIGINAL parked work items — no manual dispatch, which was the point):

- `715ec305` (`needs_page:directory-index`) → **complete**, page `deployed_at 2026-08-08
  17:02:22Z`, repo commit `65ade0ee`; `https://vetcomparison.uk/directory/index.html` HTTP 200
  serving **61 real practices, 49 postcodes, alphabetical** (24 Hour Vetcare … 608 Equine &
  Farm Vets …), sourced from `business_intel` via the site's own directory-export config.
- `2f50bfda` (`needs_page:guides-index`) → page `deployed_at 2026-08-08 17:07:31Z`, repo commit
  `836fd73b`; the page lists **exactly the three real guide pages** by their real URLs
  (cma-compliance, cma-market-investigation, independent-strategy) + the real
  obligation-checker CTA — no fabricated entries. (URL 200 confirmed after CDN lag; directory
  URL took ~60s to flip.)
- `site_plan_sections` carries both pages' layouts (`hero, directory-listing` /
  `hero, guide-list`), written by `ensure_page_section_layout` — its first production runs.

**Corrections to this file's own account, discovered in closing it** (per the lane NOTES
2026-08-08b/c/d, where the full trail lives):

1. The re-triage plan's "guides-index needs NO new handler" was **wrong** — bare
   `page-build-handler` has no layout-filling step and no-op'd again when the improvement loop
   re-dispatched it (live refutation). Both pages route to `directory-build-handler`; its
   `ensure_layout` step is page-name-generic.
2. The impact section's "not surveyed here" fleet question got a partial answer for free: the
   improvement loop detects this page-class symptom (`unbuilt_internal_link`, since 08-02) but
   its remediation dispatch rebuilds the WRONG page and self-reports success —
   **`bugs_open/220`**, filed from this lane with live reproduction. The loop cannot fix this
   bug's class; it CAN now route it correctly once a `needs_page` item exists (proven: it
   revived and re-routed `715ec305` itself via `incomplete_page_group` + `refreshOpenWorkItem`).
3. `practice`/`entity-page` remains deliberately unbuilt (P1 crawl 10/~2,109 at 08-06) — that
   was scope-out, not omission, and stands.

Operator note that cost 45 minutes: the dispatcher orders `priority ASC` — LOWER dispatches
first (`load_work_item_actions.go:683`; WRONG_CALLS 2026-08-08 second entry).

---

# RE-VERIFICATION 2026-08-24 — the 08-08 fix HOLDS; the CLASS does not. Bug stays OPEN, and why

Lane resumed 2026-08-24 at the owner's request. Three findings, in descending order of what
they change.

## 1. The 08-08 fix is still live and still correct (verified at the artefacts)

- `https://vetcomparison.uk/directory/index.html` — HTTP 200, 52,699 B, `last-modified`
  **2026-08-23**; 49 postcode-shaped strings, "24 Hour Vetcare" … "608 Equine & Farm Vets"
  present. `/guides/index.html` — HTTP 200, `guide-list` component serving.
- **Both pages were re-rendered by a fleet wave on 08-23 and the real listings SURVIVED it** —
  so this is not a stale artefact from August 8th; the pipeline reproduces it.
- A **third page shape** was proven on `directory-build-handler` today by the `vetcomparison`
  lane (not this session): `needs_page:practice` (`3cce980c`) re-routed 10:11Z, deployed
  **10:17:38Z first attempt** — a plain content page via `ensure_page_section_layout`'s default
  branch, with human-authored `content_direction` rails. Their commits `a0c8fa18b`,
  `98beb8b92`, `aa26df458`.

## 2. The defect CLASS is live today, through a producer this file never examined

`reconcile_site_plan_action.go`'s emit **hardcodes `handler_agent='page-build-handler'`** for
every page and never consults the builder map that `WriteBuildItemsAction` has used since 08-08.
So the routing decision had two copies that disagreed. Parked right now with this file's exact
signature (`page-build-handler no-op: no sections ready to build`), each having burned an attempt:

| site | page | page_type | parked since |
|---|---|---|---|
| garden-tools.uk | `brand-directory-index` | **entity-directory** | 2026-08-23 |
| garden-tools.uk | `brand-profile` | entity-page | 2026-08-23 |
| garden-tools.uk | `buying-guides-index` | section-index | 2026-08-23 |
| dartsonline.com | `brand-detail` | entity-page | **2026-07-20** |
| loanzy.uk | `guides-index` | section-index | 2026-08-18 |

The first row is the sharp one: an `entity-directory` page **sat unbuilt while its builder had
been live for fifteen days**, purely because a second copy of the routing decision didn't know
the builder existed. This file's own "Impact" section said the fleet question was "not surveyed
here — scope this before generalising"; that survey is now done and the answer is that the fix
was never fleet-wide, because the fleet has more than one door.

**Fix committed `d1aa231aa`** (council corr `52dbd067-10ed-4a6e-84eb-3fbf47d099dd`): the routing
decision now exists ONCE (`builder_routing.go:builderForPageType`), `reconcile_site_plan` asks
it, `section-index` gets the map entry that two hand re-routes had been standing in for, and a
page whose type has no builder files a **deferred `capability_gap`** instead of a dispatch item
that cannot succeed. Six tests over a decision that had **zero** coverage; four mutations each
kill their own test. Go-only — **inert until the next image roll**.

## 3. Prior art this lane should have found first, and one open question

**`bugs_closed/187`** measured this same population on **2026-08-03**, names these very rows
(`reconcile_site_plan | needs_page | needs_human_review | 9`), and its shipped fix
**deliberately left this emitter unguarded** — *"015-shape gaps are real findings"*. So the
`capability_gap` arm above touches another lane's considered, council-directed decision. It is
in front of the council as round 2's headline question, the notice is recorded in 187's own file,
and **if that lane's view is that reconcile's unbuildable-type emits must stay `needs_page`, the
routing half stands alone and the gap arm gets reverted.** (I found this only because the
council's round-1 check returned "87 across 16 sites" against my "five" — see WRONG_CALLS
2026-08-24, four entries, including grepping the bug NUMBER instead of the ERROR STRING.)

**Corrected class census, 2026-08-24, deduplicated** — 87 items / 16 sites carry the signature,
79 still parked. **69 are tool pages** (67 layout-less, 11 sites), of which 42 are
`unbuilt_internal_link` items from `bugs_open/220` — this lane's own filing — so the tail is
being *refilled* by a different bug faster than it drains. This change reaches the ~11 non-tool
typed items and **must not be read as fixing the tool class**.

## Status: STAYS OPEN (deliberate)

The original defect is fixed and live, which would normally move this file to `bugs_closed/`
under the owner's restored fixed-AND-live bar (2026-08-12). It stays open because **the class is
reproducible on the fleet today**: garden-tools.uk's `brand-directory-index` is parked, right
now, for exactly the reason this file names, and the fix for it is committed but **inert until
the next roll** — which is precisely the case CLAUDE.md says keeps a bug open ("a fix committed
but inert until the next roll stays OPEN, because the defect is still reproducible until it
ships"). **Move it when:** the roll lands (check build provenance, per service), a
`reconcile_site_plan` run mints a typed page's item carrying the right `handler_agent`, and the
five parked rows above are re-triaged and build. Not before.

## Diagnosis method for the 2026-08-24 structural claim (declared, per the owner ruling of 2026-07-31)

The claim above — *the routing decision has two copies and one of them never consults the map* —
is cross-cutting and structural, so it owes either a `090` diagnosis run or a plain statement of
the equivalent first-hand verification substituted for it. **`090` was not run. The substitute,
stated so a reader can judge it rather than take it on trust:**

1. **The code path was read, not grepped**: `reconcile_site_plan_action.go`'s emit
   (`'page-build-handler'` as a SQL literal at ~:297), both dispatch predicates
   (`claim_work_item_action.go:102`, `load_work_item_actions.go:711`), and
   `WriteBuildItemsAction`'s map.
2. **The population was measured live and the first measurement was WRONG and corrected** — the
   `spec->>'page_type'` filter returned a false zero; the corrected census (join `pages.page_type`)
   returned 87/16 with the per-type and per-producer breakdown quoted above. A census that could
   only ever have agreed with me is what the marker rules exist to catch, and this one did not
   survive its own demand control.
3. **Four independent mutations kill their own tests** (hardcoded handler restored, gap arm
   deleted, ownership guard reordered, role fallback dropped) — so the mechanism is pinned by
   something that fails when it is absent, not merely asserted.
4. **The council gate ran THREE rounds against it** (corr `52dbd067`) and each round found
   something real: a phantom column, then three "you asserted, you did not query" HIGHs, one of
   which changed the code. That is adversarial review by readers who could not see my working —
   the property `090` is valued for.
5. **The `landmine-verifier` independently confirmed the structure**, unprompted by my prose:
   *"Both `needs_page` producers (`WriteBuildItemsAction` and `ReconcileSitePlanAction`) confirmed
   present with the dual-producer structure described"* — verdict **STILL_VALID**
   (`doc_notes`, `categories ? 'landmine-verification'`, 2026-08-24).

**Where that substitute is weaker than a `090` run, said plainly:** every one of those checks was
aimed at the mechanism I had already named. None of them could have told me the cause was
somewhere else entirely, which is the specific thing the diagnosis loop is good at — and this
lane has already been burned once today by a check that could not return the disconfirming
answer. A reader who wants that assurance should run `090` on the symptom string rather than
treat this list as equivalent.

## Producer census 2026-08-24 — "is the cause anywhere ELSE?", asked properly and answered NO

The one thing this lane's first-hand verification could not do was rule out a *different* producer
with the same shape. So it was asked as a census rather than left as a caveat. **Five code sites
mint page-build work items. Only one had the defect.**

| producer | handler choice | verdict |
|---|---|---|
| `WriteBuildItemsAction` (planner) | consults the builder map | **correct by design** — but still lacks the `section-index` entry until the held swap lands |
| `ReconcileSitePlanAction` | `'page-build-handler'` as a SQL literal | **the defect** — fixed, `d1aa231aa` |
| `check_sectionless_pages` | `'page-build-handler'`, **with a stated reason** | not the defect — scoped to pages a same-role sibling can repair, and that fallback lives in *that* workflow (its own header says so) |
| `check_componentless_pages` | `'page-build-handler'`, **with a stated reason** | not the defect — *"the page's own sections array is intact, so page-build-handler can build it"* (`:55`) |
| `check_incomplete_page_group` | `'page-build-handler'`, **no stated reason**, consults only `role` (the TP-004 tool guard) | **suspected, then REFUTED by measurement** — see below |

The fifth looked like a second instance: it mints `needs_page` for a page that may well have no
layout, and unlike the two above it offers no reasoning for the handler choice. The disconfirming
query — `spec->>'reason'` across every row carrying the no-op signature:

```
(none)                  70    ← overwhelmingly tool pages, created_by generic/discovery (bugs_open/220's class)
not_built                9    ← reconcile_site_plan's own reason string
image_landed             4
content_data_backfill    3
sectionless_pages idiom  1
incomplete_page_group    0    ← the suspicion, refuted
```

**Zero.** Every typed-page instance of this defect carries reconcile's `not_built`, and the large
remainder is the tool class with a different cause. So the fix's scope is not an assumption — the
alternative was enumerated, one candidate was suspected on structure, and the evidence said no.

**What this does NOT rule out**, stated because the census can only see failures the platform has
already recorded: `check_incomplete_page_group` retains the *shape* (a hardcoded generic handler,
blind to `page_type`) and would produce this defect the first time it mints a typed, layout-less
page. It has not yet. That is a latent instance, not a live one — worth a comment pointing at
`builderForPageType` when someone next touches that file, not a fix ahead of demonstrated need.

> **Correction to this lane's own 2026-08-08 account, found by the same read:** the
> `HANDOFF_2026-08-08b` says the improvement loop *"re-routed directory-index to
> `directory-build-handler` **via the builder map**"*. `check_incomplete_page_group` hardcodes
> `page-build-handler` and consults no map, so the map cannot be what re-routed it — the handler
> on that row came from this lane's own earlier hand-update, which `refreshOpenWorkItem` then
> preserved. The loop genuinely did revive the row (that part stands); it did not choose the
> handler. `[The mechanism of the preservation is INFERRED from the hardcode — I have not read
> refreshOpenWorkItem's field-merge behaviour.]`

## Status of the 2026-08-24 fix at end of session — read THIS, the sections above are as-written

The sections above were written across one day and their council-status lines went stale as they
were written; this is the current state. Forward-only — nothing above has been edited.

**Committed and inert.** Five commits, Go-only, all carrying `Council-Submitted: 52dbd067-…`:

| commit | what |
|---|---|
| `d1aa231aa` | `builderForPageType` + reconcile routes through it + the routing tests |
| `0baa8a107` | the operator recipe the `capability_gap` arm replaces (comment only) |
| `03e2bbdb7` | round-2 answer: the gap row's `handler_agent` is now **empty** |
| `90448d175` | every ROUTED handler must be a known-registered agent (the 078 class) |
| — | plus this file, NOTES, PLAN, RUNBOOK, README, SUMMARY, LANDMINES, BLD-027, WRONG_CALLS ×7 |

**Council: FOUR rounds on one correlation, three REVISE and the fourth in flight at session end.**
Each round found something real, which is the argument for the gate rather than against it:

1. **REVISE** (`editquality`) — a phantom `site_plan_pages.page_type` column in my submission text.
   The code never had it. Its checks also returned **87 items / 16 sites** against my "five", and
   chasing that discrepancy is what found `bugs_closed/187`.
2. **REVISE** (`bug_historian`) — three HIGHs, all "you asserted, you did not query". Two were
   answered by one query each (`directory-build-handler` is live; `tool-builder` /
   `entity-page-builder` are genuinely absent). **The third changed the code**: a `deferred` row
   with a non-empty `handler_agent` naming a non-existent agent is the `bugs_closed/078` shape.
   Now empty, matching 47 of 47 existing `capability_gap` rows.
3. **REVISE** (`guardian`) — I had put the HELD `WriteBuildItemsAction` swap into the submission's
   **executable** `edits[]` array in order to document that it must not ship. An applier would
   have patched a file I had myself measured as red. Removed; it is prose-only now.
4. In flight at session end. **Pre-committed stopping rule** (NOTES, written before the verdict):
   a real defect → fix it; a false premise → one evidenced answer; **a third objection to either
   question I asked the council to RULE on — the `capability_gap` arm versus 187, or the transient
   split-brain — means taking the contained option instead of a fifth round: revert the gap arm,
   keep the routing half.**

**Two things are the council's to decide, not this lane's**, and both have a revert offered:
the `capability_gap` arm against `bugs_closed/187`'s deliberate "not guarded" ruling (notice filed
in 187's own file), and the transient split-brain while `WriteBuildItemsAction` keeps its inline
maps (blocked by another lane's ownerless change, measured red — breaks HEAD's build and fails
three existing tests).

**A `090` diagnosis run was attempted twice** (the owner-ruling escape hatch is declared above).
The first was doomed by the cumulative bundle-budget trap and cancelled with its reason recorded;
the second, scoped to one symbol, read the target and was still iterating at session end.
**Neither produced a verdict, so nothing here rests on one.**

**Nothing above changes the closure test**, which is unchanged and is at the artefact: after the
next roll, probe the running pod for `builderForPageType` with a negative control, watch a
`reconcile_site_plan` run mint a typed page's item at the right handler, then re-triage the five
parked rows using the RUNBOOK's recipe — **not before the roll**, because until then the rows
still name `page-build-handler` and a re-triage burns an attempt of three for nothing.

## AMENDMENT — the fix was NARROWED on the council's evidence (round 4), and the diagnosis runs produced nothing

Two changes to what the sections above describe. Both are the record correcting itself; nothing
above has been edited.

### 1. `section-index` is no longer routed. The divergence was the defect, not the untidiness.

Rounds 2–4 saw four seats (`guardian`, `reuse_agent`, `editquality`, `bug_historian`) object to
the same thing: `builderForPageType` was called by *one* producer while `WriteBuildItemsAction`
kept its own inline copy. I answered twice that the divergence was bounded and harmless — one
page_type, one path, no regression, and a shared dedup key so a page cannot hold two items.

**Round 4's `guardian` turned that last point against me, correctly.** Both producers mint the
same `item_key` (`needs_page:<name>`) and `idx_swi_dedup` is UNIQUE over non-terminal statuses, so
whichever fires **first** wins and the other is dropped by `ON CONFLICT`. A page
`WriteBuildItemsAction` reaches first therefore keeps the *wrong* handler and this fix never fires
for it — **with no signal anywhere**. That is not tidiness; it is a silent no-op, which is this
estate's most-recurring failure shape.

`section-index` was the **only** line on which the two maps differed. It is now removed, and the
maps are byte-identical (verified: both sort to the same 8 entries). So:

- **The divergence does not exist** — no page_type can route differently depending on which
  producer files first. The race has nothing to race over.
- **The headline case still lands.** `garden-tools.uk/brand-directory-index` is
  `entity-directory`, a type `WriteBuildItemsAction` *already* mapped correctly — routing
  reconcile through the shared authority fixes it with no divergence at all.
- **The cost, stated where it is decided:** two of the five parked pages are `section-index` and
  **stay parked** — `loanzy.uk/guides-index` and `garden-tools.uk/buying-guides-index`. They are
  no worse off than today; they simply are not fixed yet.
- **The entry goes in with the swap**, when one line moves both doors at once — which is what the
  consolidation was for. Re-adding it early fails two tests, and the producer test's failure
  message says so in terms.

### 2. The `090` runs produced no verdict — twice — so nothing here rests on one

Declared plainly because the owner ruling of 2026-07-31 makes the loop the route by which a
structural claim becomes "filed", and it was **not available** for this mechanism:

- **Run 1** (`edcbc57b`), multi-symbol: doomed at bundle 1 — `WriteBuildItemsAction` (25,397
  chars) never fitted, 35,233 of the 60,000 budget already spent. Cancelled with the reason in
  its `error` column rather than left `diagnosing`.
- **Run 2** (`0f5a40da`), **single symbol**, i.e. exactly what the landmine prescribes: bundle 1
  clean, bundle 2 clean, **bundle 3 truncated** — the loop widens its own scope between
  iterations. Ended `status='complete'` with **zero** non-bundle artifacts.

So the declared first-hand substitute above is the whole of the verification, plus what the
council found — and the council found more than either run would have. **The landmine has been
further-corrected with this measurement** (checking after the first bundle is necessary and not
sufficient), and re-verified.

### What still stands unchanged

The bug stays **OPEN** — `garden-tools.uk/brand-directory-index` is parked right now for exactly
the reason this file names, and the fix is inert until the roll. The closure test is unchanged:
probe the running pod, watch a reconcile-minted `entity-directory` item carry
`directory-build-handler`, then re-triage the parked rows per the RUNBOOK — **after** the roll,
never before.
