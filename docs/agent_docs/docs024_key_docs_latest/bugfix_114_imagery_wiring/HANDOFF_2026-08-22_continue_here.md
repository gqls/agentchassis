# HANDOFF 2026-08-22 — bugs_open/114 imagery wiring: continue here

Cold-start for whoever picks this lane up. Read this, then `PLAN_2026-08-22_imagery_wiring.md`
for the mechanism and `NOTES_imagery_wiring.md` for the evidence and the missteps.
`README_where_we_are.md` is the owner's plain-prose log — append to it, never rewrite it.

## State in one paragraph

`bugs_open/114` is "imagery is planned, generated, deployed, and then nothing points a page
at it". The framework half is **built, council-approved, LIVE on chassis `v1.0.1326`, and
proven end to end on the designated fixture**. The data repair is applied. What remains is
one detection check, some queue hygiene, and two owner decisions. **The bug is not closed**,
and the bar for closing it is stated at the bottom.

## What is LIVE and PROVEN (2026-08-22)

| thing | register | proven how |
|---|---|---|
| `store_asset` no longer poisons the site-wide default | IMG-072 | binary probe on BOTH replicas, 3 new literals PRESENT + a present-control + an absent-control |
| the hero resolver says when it fell through to the site default | — | same probe |
| card derivation fires at the landing event, not at a sweep | IMG-073 | consumer path proven live (below); emitter itself awaits a natural landing |
| the 19 poisoned `content_data.hero_url` rows repaired | migration `562` | applied; guard induced to refuse; idempotent re-run 0 rows; wire-probed 200 |

**The acceptance test, run live 2026-08-22 18:03Z on mortgagecalculator.co.uk:** a
`needs_content_image` item filed at `triaged` in the exact shape the emitter produces was
claimed by `asset-deployer`, ran `derive_card_asset`, and produced
`card_tool_repayment` — **active, `entity_type='page'`, `entity_id` set, origin lineage
set** — which the listing-card reader join (`queryresolve.pageImageJoins`) now resolves,
and whose file serves **200**. That site had **0** card assets before; it has 1 now, and
it is the first entity link any of its ten 08-15 heroes has ever had.

> ⚠ **The file 404'd on my first probe and served 200 on the second, ~4 minutes later.**
> Do not conclude a deploy failed from one probe — this file's own bug records the same
> trap at fleet scale (41 reported broken, 35 served 200 on an unhurried retry). Re-probe
> before filing anything.

## What is NOT done

1. **The emitter has never fired naturally.** `emitContentCardDerive` is proven by unit
   tests (5 mutation proofs) and its *consumer* is proven live, but no real imagery
   landing has run through `flag_page_image_rebuild` on the new binary yet. **First
   natural landing is the thing to watch.** Grep the chassis logs for
   `emit_content_card_derive` — every disposition is logged, including each skip.
2. **The other nine mcalc heroes** still have no cards. The emitter will not retro-fill
   them (it fires at a landing). Either file nine more derive items in the shape above, or
   let the design-discovery sweep do it **if** that lane is ever revived (see below).
3. **Part 3 — the detection check.** Nothing detects "this page has a page-scoped asset,
   deployed and serving, and the page points at the generic site fallback instead".
   Specified in PLAN as a flag-only check; not written.
4. **Queue hygiene.** 8 `image_landed` items parked in `needs_human_review`: 5 whose pages
   resolve no sections (reuse `sql_for_agents/300_sweep_187…` pattern), 3 marked
   "satisfiable now" by the 08-21 revalidation (re-triage).
5. **The residual poisoned keys.** `icon_url`, `content_hero_url`, `illustration_url`,
   `sprite_sheet_url` — one distinct value fleet-wide each, all naming files the deployer
   cannot produce. Migration 562 deliberately did **not** touch them: no Go reader, 0
   deployed `page_components` reference their literals, **but `illustration_url` has a live
   template consumer (`brief-explanation`.html_template)**, so it needs its own analysis.

## Two owner decisions, not mine to take

- **Widen `check_content_image_missing`'s surface to `page_type='content'`.** It sweeps
  only `blog-post` and `tool` (`:129-146`), so case-study pages are outside every producer.
  This is fleet-wide image-generation spend. Measure the population first:
  `SELECT count(*) FROM pages WHERE page_type='content'`.
- **The 5 unsatisfiable parked rows** — cancel, or build the pages they name?

## Things found that belong to OTHER lanes — report, do not fix

- **⚠ `design-discovery-agent` has not run since 2026-08-11.** `site_discovery_rotation`'s
  newest `last_selected_at` for that lane is 08-11; all four siblings are current. The
  `site-discovery-staleness` CronJob (`bugs_open/230`) reports the gap **daily** and
  nothing consumes the report. This is why sweep-driven convergence never happened and why
  IMG-073 moved it to the event. **Owner: `bugs_open/230`.**
- **`gamesdesign.co.uk`'s `card_tool_gacha_pity_designer` 404s**, asset created
  2026-08-17 — re-probed, not a lag. A derived card whose file never deployed. Different
  failure from 114's class; unfiled.
- **`noted.co.uk` serves `/assets/images/hero.jpg` 200 with no asset row keyed `hero`** —
  an orphan file with no row behind it. Recorded in 562's header, not acted on.
- **`webdesign.uk` returns 302 on every asset path**, so no probe there is decisive.

## Council trail

- `3c0560f3-2873-439f-8311-61fde3903fc7` — the code (IMG-072 + IMG-073). **APPROVED** at
  round 3. Rounds 1 and 2 each found something real; round 2's residual (a per-step config
  gate cannot see a per-invocation `asset_key`) is now a WARN naming the key.
- `4145fcdc-9ffe-42e0-a547-49e07bda04db` — migration 562. **Submitted, verdict owed a
  read.** Read it and act on a REVISE; the migration is already applied, and the file's
  header carries the full before/after so the baseline is not lost.

## The bar for closing 114

Owner's rule is **fixed AND live**. Two of the three mechanisms are. Before moving the file
to `bugs_closed/`:

1. a natural imagery landing files the derive item **without** a sweep (watch the log line);
2. the resulting card is entity-linked **and** its file serves on an unhurried re-probe;
3. no new site acquires a poisoned `content_data.<purpose>_url` — re-run the
   `count(DISTINCT v)` census in the RUNBOOK and confirm it has not grown;
4. the detection check (item 3 above) exists, or the decision not to build it is recorded.

Until then it stays open, and the reason is written in the bug file rather than implied by
the commits.

## Traps this lane hit — read before repeating the work

All in `NOTES_imagery_wiring.md` and `WRONG_CALLS.md`, but the four that cost most:

1. **A test that asserts a helper's output does not pin that the code under test calls the
   helper.** Made twice in one session, hours apart, the second time after writing the
   first up. The control is structural: make the code under test return the value.
2. **A JSON path probe cannot distinguish "not declared" from "not there".** Reading
   `input_schema->'background_image'` returned NULL for four rows because the fields live
   under `input_schema.fields`. Four rows agreeing perfectly is evidence about your PATH.
3. **`LIKE`'s `_` is a wildcard.** I filed a defect against `check_undeployed_assets` for
   matching `content_hero-%`; it matches `content-hero-…` fine. Refuted the same day.
4. **Backticks in `git commit -m` execute.** Cost two words out of migration 562's message.
   Use `git commit -F -` with a heredoc. This trap is in `MEMORY.md` and I hit it anyway.
