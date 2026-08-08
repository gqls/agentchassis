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
