# 214 — imagery scope_refs are LLM-minted free text, never validated against the plan they point into; orphans ship silently and their assets are already paid for

**Filed 2026-08-07** by the brochure_component_library lane, promoting the latent
defect RFC_016 §1 flagged ("worth its own bug file") — and upgrading it: **it is
not latent. 5 of 131 section-scope imagery rows fleet-wide are orphaned today**,
one of them minted THIS MORNING by the newest planner prompt, and four of them
have active, generated assets that no page build can ever pick up.

**Verification route, declared per the 2026-07-31 owner ruling:** this file was
not put through the 090 diagnosis loop. Substituted first-hand verification:
every function in the chain read at HEAD this session with citations below
(mint site, drop sites, ordering assignment, both consumer joins), plus a live
fleet census whose result could have come out otherwise (it found 5 orphans,
not 0 — and would have found 0 on a healthy fleet). The mechanism is confined
to one writer and two consumers; nothing here rests on inference from grep hits.
A fixing thread that wants the loop's independent read should still run it.

## The mechanism

The planner LLM emits an `imagery` block whose section-scope entries are keyed
by **free-text `"page_name:ordering"` strings** (documented shape:
`write_site_plan_action.go:827-846`). `flattenImageryBlock` copies each key
verbatim into `site_plan_imagery.scope_ref`
(`write_site_plan_action.go:883-897`) — **nothing anywhere checks that the page
name or the ordinal resolves to a section that actually survived planning.**

Three ways the reference goes wrong, all observed or mechanically reachable:

1. **Minted out of range.** The LLM writes an ordinal past the page's section
   count. Live instance: fundamentallyai.com current plan (written 2026-08-07
   08:24 UTC, replan corr `801b0732`), `scope_ref='about:4'` for
   `illustration_people_approach` — the about page has sections 0–3; the
   illustration was almost certainly meant for `people-feature-block` at
   ordinal 2. A `needs_imagery` item for it was queued the same second, so an
   asset will be generated against a reference that resolves to nothing.
2. **Minted against a page-name variant.** The LLM keys imagery `about:...`
   while emitting the page's sections under `about-index`. Live instance:
   gamesdesign.co.uk current plan (2026-06-05), four icon rows at `about:2`
   (`icon_no_ads`, `icon_math_first`, `icon_browser_based`,
   `icon_practitioner`); the plan has **no** `about` page. All four assets were
   generated 2026-06-06 and are `status='active'` — paid for, deployed, and
   unreachable via the section-imagery join below.
3. **Ordinal-shift after drops** (the flavour RFC_016 §1 documents). The keys
   are minted against the section array AS EMITTED, but `ValidateSitePlanAction`
   drops entries after that: unresolvable section names
   (`v3_site_actions.go:3168-3173`), nameless objects (`:3227`), unrecognised
   shapes (`:3236`). Every drop compacts the array, and `write_site_plan`
   persists `ordering` = position in the POST-drop array
   (`write_site_plan_action.go:383`). One dropped entry mis-keys every later
   imagery ref on that page — silently, since both sides remain "valid".
   No live instance today (2026-08-07's fundamentallyai run persisted all 71
   emitted entries, zero drops), but the door is open on every replan.

## Who consumes scope_ref, and what each failure does to them

- **Page-level LIKE join** (`plan_sections_action.go:362-386`):
  `scope_ref LIKE <page> || ':%'`. Tolerates a wrong ordinal (flavours 1 and 3
  degrade to "imagery attributed to the right page, wrong section" — currently
  harmless because mapping is by key/kind, not ordinal), but a page-name
  mismatch (flavour 2) makes the row **invisible to every build** — the
  gamesdesign icons are exactly `bugs_open/114`'s symptom class (generated,
  deployed, never referenced), reached by a different door.
- **Lock carry-forward across plan rewrites**
  (`write_site_plan_action.go:651-720`): matches old-plan locked imagery to
  new-plan rows on exact `(scope, scope_ref, category, subject, ordering)`.
  Ordinal drift between plans breaks the match → a lock silently fails to
  carry (churn of pinned imagery, the webdesign colour-churn class) — or, on a
  page whose ordinals shifted, carries a lock onto the WRONG section's imagery.
- **`flag_page_image_rebuild_action.go:116`**: parses scope_ref only to find
  the page — same tolerance profile as the LIKE join.

## Evidence (all 2026-08-07, live DB)

Census — could have returned 0, returned 5:

```sql
SELECT count(*) AS total_section_scope,
       count(*) FILTER (WHERE scope_ref !~ '^[^:]+:[0-9]+$') AS malformed_ref,
       count(*) FILTER (WHERE scope_ref ~ '^[^:]+:[0-9]+$' AND
         (split_part(scope_ref,':',2))::int >= COALESCE((
           SELECT count(*) FROM site_plan_sections sps
           WHERE sps.plan_id=spi.plan_id
             AND sps.page_name=split_part(spi.scope_ref,':',1)),0)
       ) AS orphaned_ordinal
FROM site_plan_imagery spi WHERE spi.scope='section';
-- 2026-08-07: 131 | 0 | 5
```

The 5: fundamentallyai `about:4` (current plan, same-day), gamesdesign
`about:2` ×4 (current plan, minted 2026-06-05 in the same instant as the plan —
wrong at birth, not drifted). Asset check: all four gamesdesign keys have
`assets.status='active'` rows dated 2026-06-06.

## Fix candidates, ordered by what closes the door

1. **Validate at the only door rows enter.** `flattenImageryBlock` runs inside
   `write_site_plan`, AFTER validate_plan — the surviving sections array is in
   the same `CollectedData` it walks. Resolve each `page:ordinal` against it:
   unknown page or out-of-range ordinal → degrade the row to page scope (or
   skip) with a durable log, exactly the pattern `FACT_SCOPING_EMPTY_COMPOSITION`
   set for fact assignments. Makes an orphaned scope_ref unrepresentable;
   ~30 lines in one function; no prompt change; no consumer changes.
2. **Move imagery inside the section entry** — RFC_016 §1's stated rule for the
   next structured per-section field ("object-form entries live ONLY between
   the planner LLM and validate_plan's normalise pass"; alignment travels
   inside the entry, positional keying is the named counter-example). The
   door-closing fix at the contract level, but it is a planner-prompt +
   flatten + normalise change — fleet-wide prompt breadth is precisely what
   drew 151's guardian veto, so this goes through its own council round and
   should cite RFC_016 §5.1's ratification when it lands.
3. **Repair the 5 rows** (retarget fundamentallyai's `about:4`→`about:2`;
   re-key gamesdesign's four to `about-index:<n>` or re-scope to page). Needed
   under either fix; repairs nothing structurally on its own.

Candidate 1 is compatible with candidate 2 (validation stays as the guard even
after the carrier moves) — 1 need not wait for 2's council round.

## How to verify a fix

Re-run the census above: `orphaned_ordinal` must read 0, and STAY 0 across the
next replans (the fundamentallyai row proves the current prompt re-mints
orphans, so a one-time cleanup that passes the census once is not a fix —
re-run after the next planner run on any fact-rich or icon-heavy site).

## Relations

`RFC_016` §1 (contract + the counter-example this file promotes) ·
`bugs_open/151` (the lane that surfaced it; its fact assignments deliberately
rejected this keying scheme) · `bugs_open/114` (generated-imagery-never-
referenced — the gamesdesign four are that symptom via this cause) ·
`bugs_open/204` (positional slot NAMES unresolvable in pages.sections — sibling
positional-keying defect, different table, different consumer).
