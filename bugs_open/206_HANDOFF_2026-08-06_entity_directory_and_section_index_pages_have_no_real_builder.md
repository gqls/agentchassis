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

## COUNCIL VERDICT — **APPROVED**, round 6 (corr `52dbd067-10ed-4a6e-84eb-3fbf47d099dd`)

Six rounds on one correlation: REVISE ×5, then **approved with 4 advisory objections, none
high-severity**. The five earlier rounds are not overhead — each changed something real, and two
changed the code:

| round | gating seat | what it actually found |
|---|---|---|
| 1 | `editquality` | a `site_plan_pages.page_type` column that does not exist (in my text, not my code) — and its checks returned **87 items / 16 sites** against my "five", which is what led to `bugs_closed/187` |
| 2 | `bug_historian` | three "you asserted, you did not query" HIGHs. **Changed the code**: a `deferred` row with a non-empty `handler_agent` naming a non-existent agent is the `bugs_closed/078` shape |
| 3 | `guardian` | I had put the HELD swap in the submission's **executable** `edits[]` array to document that it must not ship — an applier would have patched a file I had measured as red |
| 4 | `bug_historian` | **Changed the code, and shrank the fix**: the two-producer divergence is not untidiness — shared `item_key` + `idx_swi_dedup` means the first writer silently wins |
| 5 | `editquality` | my submission had gone **stale against my own code** — sketches still showed the pre-round-2 and pre-round-4 versions |
| 6 | — | **APPROVED** |

**Re-measured at approval time**, because two advisories rightly flagged load-bearing facts as
having been measured several rounds earlier:

- `tool-builder` / `entity-page-builder` — **still ZERO rows**. The absence the gap arm rests on is real *now*, not just at round 2.
- `directory-build-handler` — **still `active=true, snapshot=false, deleted=false`**.
- the "47 of 47 `capability_gap` rows carry an empty `handler_agent`" claim — now **49 of 49**. Two arrived during the session and both are empty, so the shape this change matched is confirmed rather than drifting.

**Advisories NOT actioned, named so the silence is not mistaken for oversight:**

- **`guardian` / `reuse_agent` / `bug_historian` (medium, ×3): the second authority.**
  `builderForPageType` and `WriteBuildItemsAction`'s inline map are byte-identical duplicates, and
  the swap that removes one is blocked on another lane's ownerless, measured-red file. Recorded in
  the code header, **BLD-027**, `LANDMINES.md` and here. Whoever finds that file clean owes it.
- **`guardian` (medium): the `capability_gap` arm reverses `bugs_closed/187`'s explicit decision
  and "should not ship silently".** It has not shipped silently — notice is filed in 187's own
  file, the question was put to the council in every round, and the revert is still offered. **If
  the 187 lane objects, that arm comes out and the routing half stands alone.**
- **`editquality` (low): the ownership-guard test uses `role='tool'`, which the guard catches by
  ROLE, so it does not exercise the guardian's actual edge case** (`role='landing'` +
  `page_type='tool'`). Correct — and that case is untestable at the emit level because the one
  such page fleet-wide (`idea.uk/report`) is `deployed`, so `decideEmit` never reaches emit.
  Measured: **78 of 79** tool-type pages are taken by the role guard first.
- **`debug_historian` (low): the pod-enumeration recipe.** ACTIONED — the RUNBOOK now enumerates
  every pod with its image and probes each, rather than `head -1`, with the "a selector can return
  2 of 41" trap stated.

The code commits carry `Council-Submitted:`; `098` credits them automatically now the correlation
is approved, with no amend (forward-only).

## CORRECTION + a fourth and fifth casualty type, from the `loanzy_uk_example_site` lane's greenfield build (2026-08-24, post-approval)

That lane ran an unaided greenfield build of `garden-tools.uk` overnight and measured the
finished site. Two things it found that this file had wrong or missing.

### 1. CORRECTION — "no worse off than today" understated the cost of the narrowing

This file's amendment says of the two `section-index` pages left unrouted:

> ~~"They are no worse off than today; they simply are not fixed yet."~~

**True of the PAGES, false of the SITES.** Measured on the finished build
`[MEASURED 2026-08-24 09:05Z, served pages, cache-busted, by the loanzy_uk_example_site lane]`:
`garden-tools.uk/buying-guides/index.html` — a parked `section-index` — is the target of **three
dead links from three different live pages**, one of them `/index.html`. **A visitor meets a 404
from the front page.** Nothing suppresses a link when its target parks (that is `bugs_open/328`,
a separate defect — but it is the mechanism through which this fix's cost is *paid*). Full set on
that site: 9 dead-link instances, 4 distinct targets, settled composition —
`/tools/finder/` ×4, `/buying-guides/` ×3, `/brand-directory/` ×1,
`/blog/buying-guide-post.html` ×1, with 3 of the 9 instances sitting on the home page.

> **Why the composition is recorded and not just the total** (the lane corrected its own figures
> to me, 2026-08-24): these numbers MOVED DURING THE BUILD, and **the total did not move with
> them.** Mid-build the count was 9; on the finished site it was still 9 — but `seasonal-planner`
> came live in between, which removed one dead link (it was itself a dead target) and contributed
> two of its own. **A stable total is not evidence of a stable population**, and a dead-link
> census taken while a build is still running measures a different site from the one that ships.
> Quote the composition and the timestamp, or quote nothing.

So the honest cost sentence is: **leaving `section-index` unrouted costs a 404 from the home page
per greenfield site.** That does not make the narrowing wrong — a dead link is *visible* where a
silent mis-route is not, which is exactly the guardian's argument — but the next reader should
price it correctly rather than read "no worse off".

### 2. `blog-post` and `blog-index` are the same defect, and this file had not counted them

The lane asked whether `blog-post` is in the map. **It is — and so is `blog-index`, and both are
mapped to `page-build-handler`, which has no layout-filling step.** So a layout-less page of
either type no-ops identically. Measured 2026-08-24:

| site | page | type | layout-less | status | producer |
|---|---|---|---|---|---|
| garden-tools.uk | `buying-guide-post` | blog-post | yes | needs_human_review | reconcile_site_plan |
| lendzy.co.uk | `blog-post` | blog-post | yes | rejected | reconcile_site_plan |
| leopardessconsulting.co.uk | `blog` | blog-index | yes | needs_human_review | offer-analysis |
| leopardessconsulting.co.uk | `blog` | blog-index | yes | cancelled | required-fields-missing-handler |

**This changes the shape of the residual, and it is worth stating precisely because it is the
framework-level version of this bug.** There are two different failures wearing one error string:

- **(a) a type with no builder, or the wrong one** — what `bugs_open/206` is about, and what
  `builderForPageType` fixes at the reconcile door.
- **(b) a type mapped to a handler that CANNOT FILL A MISSING LAYOUT** — which is every type
  routed to bare `page-build-handler`, because `ensure_page_section_layout` exists **only** in
  `directory-build-handler`'s workflow. The map is "correct" here: it names a real, live handler.
  The handler simply cannot do the layout-less case.

**(b) is the larger class and the better fix, and it is NOT what shipped.** Routing more types to
`directory-build-handler` is the wrong shape for it — the right one is for the layout-ensuring
step to be reachable from the generic build path itself (either a step in `page-build-handler`'s
workflow, or a fallback in `load_page_sections_from_spec` to `defaultSectionsForPage` when every
source is empty). That is a bigger blast radius than an approved point fix should absorb, so it
is recorded here as the named next step rather than smuggled in. **Whoever takes it: the census
in this file's producer section is the population, and `defaultSectionsForPage` is already the
single shared chooser both paths would use.**

### 3. A real greenfield verification case, offered and accepted

`garden-tools.uk` is a live, unaided, deliberately unrepaired build. Its
`/brand-directory/index.html` is an `entity-directory` page linked from its home page — i.e. this
lane's headline case, on a site nobody contrived for it. **Post-roll closure check gains a step:
that link should go live on `garden-tools.uk` without anyone touching the site.** That is a
better proof than re-triaging a page by hand, because nothing about it was set up to succeed.

## POST-ROLL 2026-08-24 (v1.0.1334) — the code is LIVE and verified; the closure test as written CANNOT fire, for two independent reasons

### The code is live, at the artefact

`[MEASURED 2026-08-24, fleet on v1.0.1334, both `agent-chassis` replicas]` — binary probe, not
build provenance, with controls in the same breath:

| probe | fr8dn | xl2zk | expected |
|---|---|---|---|
| `builderForPageType` | 2 | 2 | > 0 |
| `page_type not in the builder map` (the round-4 log literal) | 1 | 1 | > 0 |
| `builderForPageTypeXYZZY` (**negative control**) | **0** | **0** | 0 |

The round-4 log literal is the useful one: it is unique to this change and postdates the
round-2 revision, so its presence dates the binary to the final version, not merely to "some
206 code".

### …and it will not do anything yet. Both halves of the closure test are blocked.

**Reason 1 — reconcile has not run on any affected site since the roll.** `sites.last_reconciled_at`:
`garden-tools.uk` **2026-08-23 20:15** (the roll was 15:39 today), `loanzy.uk` 08-18,
`dartsonline.com` 07-22, `vetcomparison.uk` 07-17. There is no timer that runs
`reconcile_site_plan` on a cadence — it runs inside a build/publish pipeline. **So a site with no
pipeline activity never re-reaches the fixed code**, and four of the five parked pages are on
sites that have been quiet for days to weeks.

**Reason 2 — and this is the one that matters — the parked row BLOCKS its own re-mint.** The fix
routes what is minted; it does not touch rows already filed. `loadOpenPageItems` treats
`needs_human_review` as OPEN (its filter excludes only
complete/verified/rejected/wont_fix/failed/cancelled), so when reconcile does next run for
`garden-tools.uk` it will find `needs_page:brand-directory-index` already held and skip the page
as "queued" — **taking the correct new routing with it.**

### Consequence: a shared expectation was wrong, and it was wrong in BOTH lanes' docs

The `loanzy_uk_example_site` lane's handoff §3a and this lane's RUNBOOK both said, in substance,
*"post-roll `garden-tools.uk/brand-directory/index.html` should go 404 → 200 with nobody touching
the site"*. **It will not.** Confirmed at the artefact just now: still `404` (and
`buying-guides/index.html` also `404`, which IS expected — that is the deliberate narrowing).
`vetcomparison.uk`'s two pages remain `200`, so nothing regressed.

Neither of us checked what actually re-triggers a build for a quiet site, or whether an existing
parked row would suppress the new mint. **The check was designed to prove the fix and was
incapable of firing** — which is this estate's own recurring shape (a control that cannot
produce the disconfirming result), reached this time by two lanes agreeing with each other.

### What an honest live proof requires

Freeing the key, then giving the site a reconcile — in that order:

1. Close the parked row to a **terminal** status so `idx_swi_dedup` and `loadOpenPageItems` both
   release the key (`cancelled` or `wont_fix`; **never** `complete`, which would assert work that
   did not happen), with the reason in `error`.
2. Trigger a build/publish pipeline for the site so `reconcile_site_plan` actually runs.
3. Then the proof is the MINT, not the page: a fresh `needs_page:brand-directory-index` carrying
   `handler_agent='directory-build-handler'`, filed by `reconcile_site_plan`, with no hand
   routing — followed by the page building and the home-page link going live.

**Step 1 is an operator action on another lane's deliberately-unrepaired greenfield site**, so it
is theirs to authorise, not this lane's to take. Asked, not assumed.

**Re-triaging by hand instead (the RUNBOOK's step 4) fixes the PAGE but proves nothing about this
fix** — setting `handler_agent` yourself demonstrates only that `directory-build-handler` works,
which has been known since 08-08. Distinguish the two before reporting either as closure.

## 2026-08-24 — the authorisation was NOT spent, and why. Where this fix actually fires.

The owner authorised clearing the stuck job on `garden-tools.uk`. **It was not cleared**, because
checking what would have to happen next showed the action would either do nothing or do harm.

### First, the state was not what anyone thought

The owner recalled having already cleared it in another lane. Measured: **all three
`garden-tools.uk` rows are untouched** — `needs_human_review`, `handler_agent='page-build-handler'`,
born 2026-08-23 20:15, and `sites.last_reconciled_at` unchanged at the same timestamp. (The likely
source of the recollection is the `vetcomparison` `practice` page, which a different lane re-typed
this morning — a real action on a different site.)

### Second, and decisively: `reconcile_site_plan` runs in ONE place, and it is a full re-plan

`[MEASURED 2026-08-24]` — every agent whose steps carry the action:

```sql
SELECT ad.type, st.key FROM agent_definitions ad, jsonb_each(ad.default_config->'workflow'->'steps') st
WHERE st.value->>'action' = 'reconcile_site_plan'
  AND COALESCE(ad.is_snapshot,false)=false AND ad.deleted_at IS NULL;
-- build-site-planner | reconcile_site_plan     ← the ONLY row
```

and **no scheduled task targets it** (`SELECT count(*) FROM scheduled_tasks WHERE
target_agent_type='build-site-planner'` → **0**). `build-site-planner`'s step order is:

> `plan_site` (LLM) → `read_specs` → `sync_pages` → `emit_design` → `ensure_site` → `load_styles`
> → `emit_imagery` → `populate_nav` → `validate_plan` → `load_components` → `write_site_plan` →
> `load_existing_pages` → **`reconcile_site_plan`**

So reaching the fixed code on an existing site means re-running the LLM planner, rewriting the
site plan, overwriting `pages.sections` via `sync_pages`, and re-emitting design, imagery and nav
— i.e. **a full re-plan**, which is `bugs_closed/001`'s named hazard, on a site another lane is
deliberately holding as a clean measurement of the unaided route. **Clearing the row without that
achieves nothing** (reconcile never runs, so nothing re-mints); **clearing it with that** destroys
the measurement and re-plans someone's site. The authorisation covered clearing a stuck job, not
re-planning a site — *"an approval covers the action as DESCRIBED, not its side effects"*.

### What this means for the fix — stated plainly, because it narrows the claim

**This fix corrects the routing at the moment a site is PLANNED — which is exactly when the bug
is committed, and is why it is worth having.** Proof, from the same site: **13 work items were
born at `2026-08-23 20:15:50.199268`, byte-identical to `last_reconciled_at`** — the greenfield
build's own reconcile run minted them, and minted them at the hardcoded generic handler. **That
is this bug caught in the act, on a real greenfield build, by a lane that was not looking for
it.**

So:

- **New and re-planned sites get correct routing from now on.** That is the fix's real delivery
  and it needs nothing from anybody.
- **Existing sites with already-parked pages do NOT self-heal**, and are not worth a re-plan to
  rescue. Those five rows are legacy damage; the operator action for them is a re-triage, which
  fixes the pages and — say it again — **proves nothing about this fix**.

### The live proof is now FREE and needs no site touched

The `loanzy_uk_example_site` lane runs greenfield builds; that is its route. **The next greenfield
build of any site carrying an `entity-directory` or `entity-page` page is the proof**, with no
re-plan, no cleared rows, and nothing pristine disturbed: the assertion is that
`reconcile_site_plan` mints `needs_page:<name>` with `handler_agent='directory-build-handler'`
(entity-directory) or files a deferred `capability_gap` with an EMPTY handler (entity-page),
where `garden-tools.uk` got `page-build-handler` yesterday. Same producer, same moment in the
pipeline, opposite outcome.

**That is what this bug now waits on, and it costs nobody anything.**

## The PRE-FIX baseline, captured in the wild (the `loanzy_uk_example_site` lane, 2026-08-23)

The FAIL condition of this bug's closure test, observed on a real unaided greenfield build by a
lane that was not looking for it — **the best single piece of evidence either lane produced**.
`garden-tools.uk`, all 13 items filed by `reconcile_site_plan` at `20:15:50.199268`:

```
brand-directory-index | needs_page        | page-build-handler   ← entity-directory, WRONG handler
brand-profile         | needs_page        | page-build-handler   ← entity-page,      WRONG handler
buying-guides-index   | needs_page        | page-build-handler
buying-guide-post     | needs_page        | page-build-handler
tool-finder           | owned_page_review | (empty)              ← correctly gated by the role guard
about / care / index / contact / how-we-assess / seasonal-planner / affiliate-disclosure
                      | needs_page        | page-build-handler   ← correct for content/landing
```

Two things this pins that no test of mine could:

1. **The bug, in the act, at plan time** — an `entity-directory` and an `entity-page` both routed
   to the generic builder by the producer, with no hand involvement, on a site built from nothing.
2. **The owned-page role guard working correctly in the same run** (`tool-finder` →
   `owned_page_review`, no handler), which is the control: the producer is not simply broken, it
   is broken *specifically* on the type axis. That is the discriminating detail.

It is recorded as a dated observation in that lane's `NOTES_loanzy_uk_example_site.md` as well, so
it does not depend on either lane remembering it. **Re-run the corrected closure query in the
handoff over these exact rows before trusting it on new ones** — it must return `FAIL` for the
first two and `n/a` for the rest. It does, as of 2026-08-24.

## 2026-08-25 — the swap LANDED (both doors, one authority), and the closure test had a hole I had to find first

### 1. The blocker cleared, so the deferred half shipped

`HANDOFF_2026-08-24_continue_here.md` item 3 said the `WriteBuildItemsAction` swap was blocked on
an ownerless dirty hunk in `load_work_item_actions.go`, and told the next session to **re-check
rather than assume**. Re-checked: `git status --porcelain` is empty for that file and
`scripts/verify-head-builds.sh` reports HEAD builds. So the swap is done — commit `efec862f4`,
council corr `b92e624d-15c7-4ef7-a2e5-4a7f41187b38` (submitted, verdict pending at time of
writing; the commit carries `Council-Submitted:` per the forward-only rule).

What landed: the inline `builderInfo` / `availableBuilders` / `unavailableBuilders` maps are gone
from `WriteBuildItemsAction`, which now calls `builderForPageType`; `section-index` was added to
the shared map **in the same commit**, which is the condition council round 4 imposed; and the
`capability_gap` row's `handler_agent` went EMPTY at this door, matching the round-2 ruling at the
sibling door. Five tests were added — this door had **zero** direct coverage before.

### 2. The closure test as written could be passed by a HAND RE-ROUTE. It nearly was.

**This is the important part of today, and it is a defect in our instrument, not in the fix.**

The handoff's corrected closure query asserts `PASS` when a `reconcile_site_plan`-minted row for an
`entity-directory` page carries `handler_agent='directory-build-handler'`. It guards against hand
routing only with a *procedural* caveat — *"with nobody having set the handler by hand"*. That is
not a discriminator. `handler_agent` is a **mutable column**, and re-pointing a parked row at a
working handler is the estate's documented operator escape hatch, used at least twice by us.

So the column has two causes and the query cannot tell them apart. Measured, rather than argued —
every row in existence, live `UNION` archive, all history:

```sql
WITH allrows AS (SELECT site_id,created_by,handler_agent,item_key,spec,created_at,updated_at FROM site_work_items
  UNION ALL SELECT site_id,created_by,handler_agent,item_key,spec,created_at,updated_at FROM site_work_items_archive)
SELECT s.domain, a.item_key, a.handler_agent,
       (a.spec ? 'page_type') AS spec_has_page_type,
       (a.updated_at > a.created_at + interval '1 second') AS touched_after_mint
FROM allrows a JOIN sites s ON s.id=a.site_id
WHERE a.created_by='reconcile_site_plan' AND a.handler_agent='directory-build-handler';
```

`[MEASURED 2026-08-25]` — **three rows, and all three are hand re-routes:**

| domain | item_key | spec has page_type | touched after mint |
|---|---|---|---|
| vetcomparison.uk | `needs_page:practice` | **f** | t |
| vetcomparison.uk | `needs_page:directory-index` | **f** | t |
| vetcomparison.uk | `needs_page:guides-index` | **f** | t |

Their `created_at` is 2026-07-17 — months before the fix — and their `updated_at` is 08-08 and
08-24, the two dates this estate re-routed pages by hand. **Had anyone run the closure query
without a domain filter, all three would have returned `PASS`, and the fix would have been
declared proven by rows minted by the very hardcode it replaced.** That is the whole failure mode
this lane keeps logging: a check that reports the right answer for the wrong reason.

### 3. The discriminator that works, and why it is airtight *right now*

`spec ? 'page_type'`. The fixed emit stamps `"page_type": routeType` into the spec
(`reconcile_site_plan_action.go`, the `// emit` block); the old hardcoded emit did not, and a hand
re-route only touches `handler_agent` — **it cannot add a spec key**.

`[MEASURED 2026-08-25]` fleet-wide, live `UNION` archive, all history (2026-05-12 → 2026-08-24):
**508 `reconcile_site_plan`-minted rows, and `spec ? 'page_type'` is FALSE on every one of them.**
Zero, with no exceptions, because reconcile has not run anywhere since the roll.

That zero is what makes it airtight: the stamp's population is currently **empty**, so the first
row that ever carries it was necessarily minted by the fixed binary. Use it as the **gate**, not as
a corroborator:

```sql
-- PASS requires BOTH: the fixed code minted this row, AND it routed correctly.
WHERE swi.created_by='reconcile_site_plan'
  AND swi.spec ? 'page_type'            -- ← the mint fingerprint; a hand re-route cannot forge it
  AND swi.created_at > '<the build start>'
```

> The 08-24 handoff said *"do not use `spec_stamped` as the primary discriminator — it is absent
> exactly when the fix has not shipped, which is when you most need the test to work."* That is
> right about detecting **FAIL** and wrong about confirming **PASS**, and the two need different
> instruments. For FAIL, keep joining `pages.page_type` and reading `handler_agent` (a wrongly
> routed row is exactly the one with no stamp). For PASS, the stamp is the *only* thing that
> separates the fix working from a human having fixed the page.

### 4. State of the wait: nothing has changed, measured not assumed

`[MEASURED 2026-08-25]` no greenfield build and no reconcile run since the 08-24 15:39 roll:
`site_work_items` rows with `created_by='reconcile_site_plan'` created after the roll: **0**.
`sites` with `last_reconciled_at` after 2026-08-24 12:00: **0**. The newest reconcile anywhere is
`agritec.uk` at 2026-08-24 11:26, i.e. **before** the roll. The five parked rows are untouched.

So item 1 of the handoff — the free proof on the next greenfield build — has not arrived. It is
still the right thing to wait for, and the query above is now the right thing to run when it does.

### 5. Two more measured facts about this class, recorded rather than fixed

**(a) `needs_directory` is a write-only item_type.** `[MEASURED 2026-08-25, live UNION archive,
all history]` **0 rows, ever, minted by any producer**; **0** Go readers outside
`builder_routing.go` itself; **0** rows from `SELECT type FROM agent_definitions WHERE
default_config::text LIKE '%needs_directory%' AND NOT is_snapshot AND deleted_at IS NULL`. It has
been the map's value for `entity-directory` since the 08-08 fix and has never reached a row.
`section-index` inherits it for byte-consistency rather than inventing a third answer. Retiring it
is a separate question: `create_tool_cross_link_items.go:263` gates cross-links on `item_type IN
('needs_content_page','page_rerender','needs_page')`, so a `needs_directory` row is invisible
there — in the **safe** direction (it withholds links rather than pointing at a page that may
never deploy), but invisible all the same.

**(b) `loadOpenPageItems` and `idx_swi_dedup` disagree by exactly one status, and it is the
damaging one.** The index's partial predicate excludes **seven** statuses; `loadOpenPageItems`
(`reconcile_site_plan_action.go:713`) excludes **six** — the same list minus `unresolved`. So a
row at `unresolved` is treated as **OPEN** by reconcile (the page is skipped as "already queued",
and new routing therefore **never reaches it**) while the index does **not** cover it. It is also
undispatchable, since both claim gates filter `status IN ('triaged','approved')`. The page is
parked with nothing able to free it. `[MEASURED 2026-08-25]` one live instance:
**`adversecreditmortgage.co.uk` `blog-index`**, a `section-index` `needs_page` at
`page-build-handler`, `unresolved` since 2026-08-18. **Today's routing change does not reach it**,
and not for any reason to do with routing. Both facts are now corrections in the code itself.

### 6. Closure test — unchanged in substance, sharper in instrument

The three conditions in the 08-24 handoff still stand. Condition 1 gains a clause:

1. A `reconcile_site_plan`-minted row for a typed page carrying `directory-build-handler`,
   **and carrying `spec.page_type`** — the mint fingerprint, without which the row may simply be a
   page a human fixed.
2. That page built and serving, verified by `curl`, not by `build_status`.
3. The parked `entity-directory` and `entity-page` rows resolved to their designed outcomes.

`section-index` pages staying parked is **no longer** part of the deliberate narrowing — the
hold-out ended today. But note (b) above: the `unresolved` ones still will not move.

## 2026-08-25 — COUNCIL VERDICT **APPROVED** (round 1, corr `b92e624d-15c7-4ef7-a2e5-4a7f41187b38`), and the one question three seats asked, answered with a census

13 reviewers, 4 abstained, **all approve, first round**, no vetoes. Every objection was `low` and
filed "for the record". `decided_by: all reviewers approve`. The commit (`efec862f4`) carries
`Council-Submitted:`, so `098` credits it automatically now the correlation is approved — no amend,
which forward-only forbids anyway.

### The one substantive gap, and it was the same in three seats

`bug_historian` listed it under **missing**: *"Whether any other producer besides
WriteBuildItemsAction and reconcile_site_plan mints `needs_page:` items with a third routing copy —
not checked."* `reuse_agent` asked to *"confirm no third copy of the routing maps survives
elsewhere in the codebase"*. `guardian` wanted *"a fleet confirmation"* on the section-index arm.

They were right that I had not checked. **Answered now, and the answer is NO — but the first half
of it is not what I expected.**

**There are SIX Go sites minting a `needs_page:<name>` item_key, not two** (`grep -rn '"needs_page:'
--include=*.go`, 2026-08-25):

| site | branches on page_type? | handler |
|---|---|---|
| `reconcile_site_plan_action.go:305` | yes — `builderForPageType` | routed |
| `load_work_item_actions.go:324` (WriteBuildItems) | yes — `builderForPageType` | routed |
| `rerender_page_sections_action.go:1424` | **no** | hardcoded `page-build-handler` |
| `apply_adoption_plan_action.go:731` | tool vs static only | hardcoded `page-build-handler` |
| `discovery_checks/check_incomplete_page_group.go:202` | **no** | hardcoded `page-build-handler` |
| `page_build_failure_guard.go:194` | n/a — mints `page_build_failed`, no handler | — |

So **no third copy of the MAP exists** — nothing else duplicates the page_type→handler decision.
But three further producers choose a handler *without consulting it*, and one of them is not
marginal: `page-rerender` is **the fleet's most active `needs_page` producer**, `[MEASURED
2026-08-25, live UNION archive, all history]` **414 rows, 21 of them on typed pages, still minting
today (2026-08-25)**. `site-adoption-agent`: 23 rows, 4 typed. For comparison the door I fixed
today, `site-planner`, has 25 rows and 1 typed, last minted 2026-04-18.

### …and then the evidence refused my inference, which is the useful part

Having found that, the obvious next line to write was *"so 206's class is still live through the
rerender door"*. **It is not, and I nearly filed it as a residual on an inference.** Same census,
restricted to typed pages from those two producers:

`[MEASURED 2026-08-25]` **26 rows. `error ILIKE '%no sections ready to build%'` → 0. Not one.**

Their failures are a different and *deliberate* mechanism — 10 of them are the owned-page guard
refusing (`save_page_sections: page X is rebuild_policy=owned`, plus one `OWNED_PAGE_GUARD` at
`load_page_record`), one is content validation, and 7 completed normally.

**The mechanism explains the zero, which is why I believe it rather than just recording it.** 206
is specifically the *layout-less* case: a page with no sections in any source, which
`page-build-handler` cannot fill because `ensure_page_section_layout` lives only in
`directory-build-handler`'s workflow. A rerender or adoption target **already has a layout** — it
is an existing, built page being rebuilt. The doors that skip the map are exactly the doors whose
pages do not need it. The two doors that *do* consult it are the two that mint at **plan** time,
which is when a page has nothing yet.

**So: the routing consolidation is complete for the population that has this defect**, and the
hardcoded handlers at the other three sites are not latent 206s. They are still a copy of an
answer, and if a layout-less page ever reaches one of them this reasoning expires — but there is
no evidence today that one does, and I am not filing a residual on a mechanism I cannot show
firing.

⚠ **One caveat on the census, stated because it is the kind that bites**: the join reads
`pages.page_type` as it is **now**, not as it was when the row was minted, and a page can be
re-typed (this estate does it by hand as a repair). So "21 on typed pages" is a current-state
read. It cannot manufacture the zero above — a row whose page was re-typed *into* a directory
type would still have to carry the error string, and none does — but it would matter to anyone
recounting the totals later.

### Other seat notes worth carrying

- **`debug_historian`**: the zero-row claims were quoted as `[MEASURED ... live UNION archive]`
  without the query. Fair. The queries are now inline in §2 and §5 of the section above and in this
  one. It also flags that rollout verification must **grep the running pod**, not green tests —
  noted for when this ships; today's change is committed, not yet rolled.
- **`guidelines`**: file the `unresolved` status-set divergence as a follow-up work item so it does
  not rot. It is recorded in §5(b) above with its one live casualty
  (`adversecreditmortgage.co.uk` `blog-index`); a lane that wants it should take it from there.
- **`editquality`**: the comment-correction edit is ancillary and does not count as coverage.
  Agreed and never claimed otherwise.

## 2026-08-25 (later) — an adversarial review of THIS DAY'S OWN WORK found four real defects, including in the correction I had just published

I put the day's work through a fresh reviewer on a different model, told to refute rather than
confirm. It found four things that stand. Three are mine; the fourth is a pre-existing asymmetry
nobody had written down. **The pattern across them is one thing: I wrote a correction, and then
did not apply my own correction's discipline to the correction.**

### 1. My "airtight" mint fingerprint was forgeable, TWO ways — and I dropped the discriminator I had myself measured

§3 above, the RUNBOOK §7 query and the new `LANDMINES` entry all said: gate on
`spec ? 'page_type'`, because *"a hand re-route cannot add a spec key"*. True, and **insufficient**.

- **The stamp dates the ROW, not the HANDLER value.** `handler_agent` stays mutable after a stamped
  mint. If the fixed binary ever *mis*-routes — the exact failure this test exists to catch — and an
  operator then repairs the row, it carries the stamp AND the correct handler and reads `PASS`. The
  fix takes credit for a human repair. **The same false-PASS shape as §2, one generation later, in
  the check written to prevent §2.**
- **This lane's own operator recipe forges the gate without touching a spec key.** Reconcile's
  `capability_gap` spec **already carries `page_type`** (`reconcile_site_plan_action.go`, the
  `gapSpec` block), and the recipe eleven lines below it — added by this lane in `0baa8a107` — says
  *"promote this row in place: set `item_type='needs_page'`, `status='triaged'` and `handler_agent`
  to a handler that can actually build it."* A promoted gap row is `created_by='reconcile_site_plan'`,
  stamped, `needs_page`, at `directory-build-handler`: **every PASS condition, entirely by hand.**
  `[MEASURED 2026-08-25]` prospective, not live — **0** stamped `capability_gap` rows exist
  anywhere — so it arms itself the moment the fix files its first gap, i.e. exactly when this test
  starts being used.

**What makes this galling rather than merely wrong: §2 of this same section — the measurement that
exposed the original defect — computes `updated_at > created_at + interval '1 second'`, and the
query I shipped four paragraphs later has no `updated_at` clause at all.** I had the discriminator
in hand, used it to catch the first version, and left it out of the fix.

**Corrected**: `AND swi.updated_at < swi.created_at + interval '1 second'` is now gate 2 in RUNBOOK
§7 and in the landmine. It closes both holes, because `trg_site_work_items_updated_at` is
`BEFORE UPDATE … FOR EACH ROW` (verified in `pg_trigger` 2026-08-25) so *any* write bumps it. **Its
cost is stated: it expires** — a legitimate claim or completion bumps `updated_at` too, so the mint
must be read while the row is still `triaged`. **The durable fix is a code change, named and not
smuggled in:** stamp the routed handler into the spec at mint (`"handler": route.handler`) and
assert `spec->>'handler' = handler_agent`, so the column carries its own provenance.

### 2. Two mutations survived my new test file, and one is a wall my own change removed

`[VERIFIED 2026-08-25 by running them]`, against my tests alone (the package baseline is currently
red from another lane's concurrent work, which masked this on the first attempt):

- **Gap row `status: "deferred"` → `"triaged"`: nothing failed.** This is the sharp one, and it is a
  gap *I introduced today*. Council round 2's argument for the change was that `deferred` parks the
  row **twice over**; but making `handler_agent` empty means `writeWorkItem`'s registration probe —
  gated on `handlerAgent != ""` at `:1925` — **no longer runs at all**. So `deferred` is now the
  **only** thing between an empty-handler row and both claim gates, and I left it unpinned.
- **`itemType := route.itemType` → hardcoded: nothing failed** in that file (only the separate unit
  table caught it). The constants `wbiArgItemType` and `wbiArgHandlerAgent` were declared and then
  **never used in any pin** — the assertion was intended and dropped when I rewrote the test bodies.

Both are pinned now (`0777eb297`) and both re-verified failing their own test. The file's header
claim *"Every expectation below is mutation-proven"* was **overbroad** and is corrected in place: a
claim that a file is mutation-proven must name **which** mutations were run, or it reads as
coverage nobody has.

### 3. The "26 rows, ZERO with the 206 signature" census has a blind window I did not state

The signature text `no sections ready to build` is written by flag steps that **migration 149
introduced on 2026-07-14** — its own header records that pre-149 no-ops were *"stamped 'complete'"*
with no error at all (`sql_for_agents/149_page_build_handler_noop_flags.sql`). `[MEASURED
2026-08-25]` **3 of the 26 rows are `site-adoption-agent` / `complete` / 2026-06-05** — before the
instrument existed, and *exactly* the false-complete shape 149 was written to fix. For those three,
a no-op and a success are indistinguishable, so my zero says nothing about them.

**The zero is genuine for the 23 post-149 rows** (all 21 of `page-rerender`'s included), and the
mechanism argument — 206 is the layout-less case, these doors act on pages that already have a
layout — still carries the conclusion. But I presented the *census* as the thing that refuted my
inference, and its stated caveat (`pages.page_type` read as-now) was not this one. **A zero from a
detector is bounded by the date the detector started existing**, which is a caveat I have written
elsewhere and did not apply here.

Minor arithmetic, corrected: the split above says "site-adoption-agent: 23 rows, **4** typed" — that
4 is the `page-build-handler` subset; its typed total is **5** (3 on 06-05, 2 on 07-31), and
21 + 5 = 26. Two tables, two denominators, stated as one.

### 4. Pre-existing and unwritten: reconcile counts an emit it did not make

`reconcile_site_plan_action.go:484` does `emitted++` **unconditionally** after an
`ON CONFLICT DO NOTHING` insert — while the `capability_gap` arm four lines above (`:426`) correctly
reads `RowsAffected` and cites `bugs_open/091` for doing so. So whenever a `needs_page:<name>` row
exists but is **invisible to `loadOpenPageItems`** (its `item_type` is outside that filter's three
values — which is true of every row `WriteBuildItemsAction` mints, `needs_directory` and
`needs_content_page` alike), every reconcile over that site reports a phantom `pages_emitted` and,
because `emitted > 0`, re-mints `needs_rerender`. Born with the file (2026-05-05); reachable only
when the per-page arm of `write_build_items` fires, which is wired live in `site-work-orchestrator`
but dormant since ~April. **Not fixed here** — it is one line, and it is not this lane's approved
scope.

**And the consequence for the closure test, which IS this lane's business:** RUNBOOK §7 filters
`created_by='reconcile_site_plan'`. If a future greenfield build mints through the
`WriteBuildItemsAction` door instead, that query returns **empty on a successful build** — a false
FAIL. Yesterday's evidence says reconcile is the greenfield door in practice (garden-tools.uk's 13
items were minted by it at plan time), but "the proof arrives free on the next greenfield build"
silently assumes it. **Check `created_by` on whatever rows the build actually produces before
reading a zero as failure.**

## 2026-08-25 (evening) — a real greenfield build arrived, and it CANNOT close this bug. That is the finding.

The `bugs_open/381` lane built `homegarden.uk` and, on request, snapshotted the mint before the
dispatcher touched it — 43 rows, plan minted `11:31:05Z`, captured `11:34:22Z`, 37 still untouched
since emit. Their file:
`docs/agent_docs/docs024_key_docs_latest/bugfix_381_inexpressive_composition/evidence/MINT_SNAPSHOT_homegarden_20260825T113422Z.txt`

**This is the artefact four lanes have been waiting on, and for this bug it is negative.** Worth
stating precisely, because "we got the build" would otherwise read as progress.

### 1. The build passes every gate and still cannot discriminate the fix

`[MEASURED 2026-08-25]` the site's 21 planned pages: **17 `section-index`, 2 `content`, 1 `landing`,
1 `blog-post`. Zero `entity-directory`. Zero `entity-page`.**

Every one of those types routes to `page-build-handler` under **both** the old hardcoded literal and
the new map. So the mint is stamped, untouched-since-emit, filed by `reconcile_site_plan`, and
**identical to what the bug would have produced.** It confirms the new code *ran*; it cannot show
that it *routes differently*.

**So "the proof arrives free on the next greenfield build" was wrong as written, and is corrected
everywhere it appeared** (handoff item 1, RUNBOOK §7b). The closure artefact requires a site
carrying an **`entity-directory`** page — the one type where the map and the literal disagree today
— or an **`entity-page`** (whose expected outcome is a deferred `capability_gap` with an empty
handler). Check the plan for one *before* treating a build as closure:

```sql
SELECT page_type, count(*) FROM pages
WHERE site_id=(SELECT id FROM sites WHERE domain='<domain>') GROUP BY 1 ORDER BY 2 DESC;
```

### 2. Two of my caveats settled — one wrong, one retired

- **Gate 1's "empty population" is spent, and the gate never meant what I wrote.** That build minted
  **21 stamped rows**, so the stamp's population is no longer empty (it emptied by *time*: no
  reconcile had run since the roll). And `git log -S` settles the ambiguity in "the first stamped row
  is necessarily the fix": the `"page_type": routeType` stamp was added by **`d1aa231aa` (08-24
  11:50)**, live since `v1.0.1334`. **Today's swap never touched that emit.** A stamped row proves
  reconcile consulted `builderForPageType`; it says nothing about `section-index` routing. *The 381
  lane caught this by contradicting a caveat I had put into circulation — they were right.*
- **The "which door fired?" caveat is RETIRED, free.** `reconcile_site_plan` minted all 22;
  **`WriteBuildItemsAction` did not appear as a `created_by` at all.** Reconcile is the greenfield
  door. One build is not a guarantee (that door is wired live in `site-work-orchestrator`), but §7's
  `created_by` filter is not the hazard it looked like.

### 3. And I sent them a false warning, which they had already adopted

I warned that lane their `section-index` pages would **no-op** on the current fleet, since today's
swap is unrolled. **False for their build**, and logged in `WRONG_CALLS`. 206's no-op needs a page
with **no layout from any source**; `[MEASURED]` all 17 of their `section-index` pages carry
`["hero","generic-text-block","content-listing"]`, `april-index` had **already built and deployed**
through `page-build-handler`, and `sig_206` across all 21 `needs_page` rows is **0**.

They had put it in their acceptance guide, which would have told an investigator to check
`handler_agent` first on pages where `page-build-handler` is the **correct** handler — handing their
own bug an alibi. Retracted within the same build; the error string stays in their guide as a
discriminator (still valid — if you *do* see it, it is mis-routing), the prediction is gone.

### 4. What this build DOES say about the class — and it belongs to 381, not here

`[OBSERVED, theirs]` the planner expressed a "month by month" promise as **seventeen
`section-index` pages** rather than one page carrying a period-calendar component. All 17 will
build. Nothing will fail or error; the site will serve 17 near-identical three-section index pages
where the plan promised one structure. **That is `bugs_open/381`'s symptom with no 206 in it** — and
my warning would have muddied exactly that. Recorded here only so this file does not later claim it.

### 5. Closure test — unchanged in substance, one precondition added

Conditions 1–3 stand (see the 2026-08-25 morning section, plus the two-gate correction). **New
precondition, ahead of all of them:** the build must carry an `entity-directory` or `entity-page`
page. Without one, no amount of stamped, untouched, correctly-routed rows can close this bug.

## 2026-08-25 (night) — the swap is LIVE, verified with a working control; and the closure gate no longer expires

### 1. `v1.0.1339` carries the swap — ancestry, not a symbol probe

`[MEASURED 2026-08-25]` both `agent-chassis` replicas (started `19:07:18Z` / `19:07:49Z`) report the
same build provenance, read from the pods' own startup line (in range, because the pods are fresh):

```
git_commit: a7459a44b68b8c67b7d7bb0ca7c064e0729d59f5
```

| commit | ancestor of the stamp? | |
|---|---|---|
| `efec862f4` — the swap | **IN** | today's `section-index` routing + gap alignment |
| `0777eb297` — the test pins | **IN** | |
| `d1aa231aa` / `200d54bdf` — 08-24 | **IN** | |
| `c591d8d61`, `08bfce067` (committed 20:19–20:20, after the build) | **NOT in** | ⬅ **the control** |

**The control is the load-bearing part**: two commits made after the build are correctly excluded,
so the test discriminates. This is the method §7c prescribes after yesterday's three-wrong-probes
episode — and it worked first time.

**So `section-index` now routes to `directory-build-handler` at both doors, live.**

### 2. …and nothing has happened, exactly as predicted

`[MEASURED 2026-08-25, post-roll]` reconcile rows created since `19:07`: **0**. Sites reconciled
since: **0**. The six parked rows are untouched. Nothing schedules `reconcile_site_plan`, and a
parked row holds its own `item_key` and so blocks its own re-mint. **The roll changes what happens
on the NEXT plan; it cannot reach back.** Do not read the unchanged parked rows as the fix failing.

### 3. The closure gate no longer expires (commit `1887a116b`, council `9ff151d6`)

Both producers now stamp the handler they **chose** into the item's spec, so the closure check
becomes `spec->>'handler' = swi.handler_agent`.

Why this was worth a round rather than living as a query: the interim gate `updated_at < created_at
+ interval '1 second'` is sound but **expires on the first legitimate dispatcher claim** — within
minutes, and long before anyone reads the result. A proof that dies when the system does its job is
not one you can rely on being able to run. The stamp is permanent, and an `UPDATE` of
`handler_agent` (the documented operator repair) cannot forge agreement because it does not touch
the spec. **Divergence becomes the signal**: `spec.handler <> handler_agent` says *a human
re-routed this*, which is precisely what the closure test must exclude.

Both doors stamp it, deliberately — a stamp at one door only means a reader cannot tell a mint from
a repair at the other, and the absence reads as "repaired" rather than as "not stamped". That is the
same two-doors-disagree trap this bug is about.

Mutation-proven before submitting: dropping either door's stamp fails **only** that door's test;
making the stamp disagree with the column fails **both** — the case that matters, since a stamp
agreeing with nothing would read as a forged repair. Baseline and restored green, tree verified free
of mutations afterwards.

> ⚠ **`1887a116b` is NOT live.** `v1.0.1339` predates it, so `spec ? 'handler'` is **FALSE on every
> row in the database right now** and a query gated on it returns **zero** — which reads exactly
> like the fix failing. RUNBOOK §7d carries the warning and the gate table. Until the next roll use
> §7's gates and read the mint promptly.

### 4. What actually remains

Unchanged and still the only thing standing between this bug and closure: **a greenfield build of a
site carrying an `entity-directory` or `entity-page` page.** Not schedulable by this lane, arrives
free when a lane builds one, and §7b (iii) gives the query to check a plan for one before spending a
build on the question.

## 2026-08-25 (night, later) — provenance stamp **APPROVED** (`9ff151d6`), and the three objections answered by MEASUREMENT

**APPROVED** — 9 reviewers, 8 abstained, `gated_by_truncation: false`,
*"approved with 3 advisory objection(s) — none high-severity"*. Commit `1887a116b` carries
`Council-Submitted:`, so `098` credits it automatically.

**Three seats objected on the same ground, and they were right: I asserted an absence instead of
measuring it.** `prior_art_librarian` named it exactly — *"no lookup is cited for this absence — only
the allow-list argument, which protects against unintended input_mapping propagation, not against
some other reader examining `spec->>'handler'` directly."* That is this lane's own recurring failure,
caught by someone else for the fourth time in two days. Answered now, with the queries:

| objection | seat | answer `[MEASURED 2026-08-25]` |
|---|---|---|
| "nothing reads `spec.handler`" is asserted, not verified | guardian (low), prior_art (medium) | **0** Go readers outside the writers; **0** live `agent_definitions` naming `spec.handler` |
| WriteBuildItemsAction's callers never enumerated — an added key could collide | guardian (**medium**) | one registration (`registry.go:712`), **one** live caller (`site-work-orchestrator :: write_build_items`); and **0** rows in `site_work_items` ∪ archive, all history, all producers, carry a `handler` spec key — nothing to collide with |
| nothing enforces the two doors draw the handler from one vocabulary; spelling drift would make the stamp itself a false-divergence source | guardian (low) | both take **`route.handler`** — `load_work_item_actions.go:243` is `handlerAgent := route.handler` — so drift is **not representable**, not merely unlikely |
| `builder_needed` precedent claimed without a lookup | prior_art (low) | exists at **9** sites incl. both gap arms (`reconcile_site_plan_action.go:381`, `load_work_item_actions.go:251`) |

### The one thing I actually owed and had missed

`guidelines` filed it as **MISSING**, and it is a standing owner ruling, not a preference:
*"adding a key inside an already-passed object does NOT require input_contract re-declaration, but
MUST be named in the seam's concept-register entry in the same commit"* (owner ruling 2026-08-11).
I did not. **Done now as a forward-only follow-up** — `spec.handler` is declared in **BLD-027** with
the measurements above. Late by one commit, and recorded as late rather than backdated.

### A consumer the objection led me to, and a comment my change had falsified

Chasing guardian's medium objection surfaced `resolveGuardedPage` (`owned_page_guard.go`), which
consumes `current_item.spec` for build items. It reads `spec.id` / `spec.name` **by explicit path**,
so an additive key is inert — but its comment said the spec is *"exactly what queryPagesForBuild
returned"*, and **my change made that false.** Corrected in the same commit rather than left
standing. *The objection did not find a defect; it found a document I had quietly invalidated* —
which is the same class as the two false comments this lane corrected yesterday.
