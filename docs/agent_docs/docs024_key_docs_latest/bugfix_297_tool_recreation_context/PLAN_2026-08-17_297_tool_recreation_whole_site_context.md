# PLAN 2026-08-17 — bugs_open/297: tool-recreation-handler sees 10 of up to 107 pages

Bug: `bugs_open/297_HANDOFF_2026-08-17_tool_recreation_context_capped_at_10_of_up_to_107_pages.md`.
Sibling and worked precedent: `bugs_open/275` (`bugfix_275_silent_row_caps` lane — migrations 445/446,
council trail `b684a399` APPROVED, framework detector LCO-009 committed at `eb137faed`).

## 0. Ownership checked two ways (the 275 lane's own method)

- `who-owns.py 297`: only the FILING commit (`8912607df`, the 275 lane spinning the census out as
  tickets); "likely owning workstream: none identified".
- Grep of ~25 live session transcripts for the slug and `tool-recreation-handler` (the only
  instrument that sees uncommitted work): the two sessions with real hit counts are the 275 lane
  (`612ff4f1`, last message: both tickets filed, tree clean) and the 291 lane (`318d9c48`, working
  tool-auditor). **No session is fixing 297.** Taken.

## 1. Still valid, and marginally grown

Re-measured 2026-08-17 against live `clients_db`, not copied from the bug file:

| | at filing (2026-08-17 am) | re-measured (2026-08-17 pm) |
|---|---|---|
| sites | 24 | **25** |
| sites where population exceeds the cap | 19 | **19** |
| median population | 26 | **26** |
| worst site | 107 | **107** (webdesign.co.uk, 108 pages) |

The live query (one active row, id `8701375f-81f7-4d92-ba39-c85f8489dada`, version 1 — the
duplicate-active-row trap from 446's round does NOT apply, and the migration gates on type-count
anyway) still ends `ORDER BY p.nav_order LIMIT 10`.

## 2. The measurement the bug asked for first — and it inverts 275's remedy

The bug's candidate 1 says "measure which column dominates, bound that, drop the row cap" — 275's
shape, where `description` was 80% of the payload and needed `left(…,200)` + a truncation marker.
**Here the measurement says no column needs bounding at all:**

- The prompt renders one line per row: `- {{.name}} ({{.page_type}}): {{.title}}` (read from the
  live `analyze_tool` template, not assumed).
- Column extremes across all 727 pages: `name` max 66, `title` max 144 (p99 ≈ 114), `page_type`
  max 16. Nothing pathological; no `left()`/marker machinery is warranted.
- **`rr.summary` is SELECTED and never RENDERED** — and is nearly empty estate-wide: 21 of 727
  pages have an `adoption_page` research row, max 48 chars. Same dead-payload shape as 275's
  `category` finding.
- Rendered-block arithmetic per site (`- ` + name + ` (` + type + `): ` + title + newline):
  worst site **8,810 chars (~2.2k tokens) uncapped vs 735 today**; median site ~2.2k chars. The
  `analyze_tool` prompt already embeds the original page's full raw HTML (typically tens of KB),
  so the whole population is a small fraction of the prompt it joins.

**So: drop `LIMIT 10`, bound nothing, keep `ORDER BY p.nav_order`** (with no cap, order no longer
decides visibility — nav order is a sensible presentation order for "how does this tool fit the
site"). No truncation marker is needed because nothing is truncated. Candidate 4 ("do not just
raise the number") is honoured: there is no number left to outgrow.

## 3. A second defect in the same query, found by measuring: the join can FAN OUT

`LEFT JOIN research_results rr ON rr.page_id = p.id AND rr.result_type = 'adoption_page'` has no
one-row guarantee, and one page already has TWO adoption rows (`0747e2fc…`, the `index` page of
site `00ff3af5…`, nav_order 1) — **so today's prompt on that site renders `index` twice inside the
visible 10.** Measured, not hypothetical. With the cap dropped and the join left plain, a page
with N research rows would spam N lines.

Fix in the same edit, because the coherent task is "the step returns the site's page population,
exactly": replace the plain join with a **`LEFT JOIN LATERAL (… ORDER BY r.created_at DESC LIMIT 1)`**
— newest summary per page, shape-preserving (every row keeps `name, title, page_type, summary`
keys, so the sole consumer sees an identical structure). The inner `LIMIT 1` is the fetch-one
idiom, explicitly outside the silent-cap class (LCO-009 excludes n=1, and its end-anchored regex
correctly ignores a mid-query subquery LIMIT — both decisions vindicated on live cases in 275's
round). `idx_research_page` + `idx_research_created` make the lateral cheap.

**Validated read-only before writing the migration:** proposed query returns 107 rows = population
on the worst site, and 40 rows = population 40 on the fan-out site (duplicate gone).

## 4. What is deliberately NOT in scope

- **No Go change.** The framework half of this class (LCO-009's WARN in `QueryDatabaseAction`) is
  built, mutation-tested and committed at `eb137faed` by the 275 lane; it rides the next chassis
  roll. This fix removes 297's instance (and with it the future WARN noise for this step).
- **`rr.summary` stays selected** though nothing renders it — same call as 275's `category`:
  dropping a column a future consumer might read is scope creep for a negligible saving. Recorded
  as an adjacent finding, and the LATERAL keeps its shape byte-compatible.
- **The `analyze_tool` prompt template is untouched** — the LANDMINE about prompt/token-cap depths
  is not in play.
- No new config keys, no new authority on a shared seam — RFC_022 / optional-key budget not in play.

## 5. The change: migration 453 (config-only, LIVE ON APPLY)

`docs/agent_docs/sql_for_agents/453_tool_recreation_whole_site_context.sql` + `_ROLLBACK.sql`,
following 445/446's hardened shape exactly:

1. `snapshot_agent('tool-recreation-handler', …)` FIRST (sketches must show the safety lines —
   275's missteps 3 and 5).
2. Pre-state gate by **type-count** (the 446 lesson): exactly 1 live row for the type, and the
   query still carries `LIMIT 10` + the plain join — refuse rather than clobber a concurrent change.
3. `jsonb_set` of the one query string, scoped by id AND type AND live predicates.
4. Post-state verify in `DO`/`RAISE` (never bare SELECTs): no multi-row LIMIT survives
   (`LIMIT [2-9]|[0-9]{2,}` — the inner `LIMIT 1` is allowed by design), `LATERAL` present,
   `nav_order` ordering intact, `params` untouched.
5. Rollback sidecar gated the same way in reverse (refuses unless the row carries 453's text).
6. Applied by hand (own file only — `--apply` takes every pending file), then
   `run-migrations.sh --record-only` with a note. Never a hand-written ledger INSERT.

Migration number 453 confirmed free at write time: ledger max 451; dir holds 452 only as `_HOLD`.

## 6. Council

One run for the coherent task, submitted alongside the commit (`Council-Submitted:` trailer;
review here is after the fact by design — owner ruling 2026-07-29 §2). Sketch = the WHOLE
migration file, per 275's misstep-5 rule.

## 7. How to verify (from the bug file, adapted)

- Census re-run: **no cap exists**, so "sites over cap" is vacuously zero; assert instead that the
  live query returns population rows on the worst site (107) and the fan-out site (40, no dup).
- Disconfirming pair: a page sorting past position 10 by nav_order structurally CANNOT appear
  before (LIMIT 10); after, show it in the query result. Full end-to-end confirmation in
  `llm_call_log.prompt_rendered` arrives with the next real recreation run — recorded as owed,
  not claimed.
- No truncation marker check needed — nothing is truncated (§2).
