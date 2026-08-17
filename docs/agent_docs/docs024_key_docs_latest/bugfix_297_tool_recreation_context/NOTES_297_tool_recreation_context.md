# NOTES — bugs_open/297, tool-recreation-handler's related-context cap

Append-only, newest at the bottom. Missteps are the point, not an appendix.

## 2026-08-17 — picked up, ownership checked two ways

`who-owns.py 297` → only the filing commit (275 lane spinning the census out); no owning
workstream. Live-transcript grep for the slug + `tool-recreation-handler` across ~25 recent
`.jsonl` sessions: the only real hit counts are the 275 lane itself (last message: "both filed…
nothing of mine uncommitted") and the 291 lane (tool-auditor). No competing fixer.

## 2026-08-17 — validity re-check, from the live DB not the bug file

- One active `tool-recreation-handler` row (id `8701375f…`, v1) — not a duplicate-active-row type.
- Query unchanged: ends `ORDER BY p.nav_order LIMIT 10`.
- Census: 25 sites (was 24 at filing), 19 over cap, median 26, worst 107 (webdesign.co.uk).
  Verbatim live query against the worst site returns exactly 10 rows today.

## 2026-08-17 — the measurement flipped the remedy away from 275's shape

Expected to repeat 275 (bound the dominant column, mark the truncation). Measured instead:

- Rendered line is `- name (page_type): title`; `rr.summary` is selected and NEVER rendered —
  and is nearly absent anyway (21 of 727 pages, max 48 chars).
- Column extremes: name ≤ 66, title ≤ 144, type ≤ 16. Nothing to bound.
- Whole-population rendered block at the worst site: **8,810 chars (~2.2k tokens)** vs 735 capped —
  in a prompt already carrying the page's full raw HTML.

So the honest fix is simply: **no cap, no bounding, no marker** — there is nothing to truncate.
Wrote that down BEFORE the council could ask why I didn't cargo-cult `left(…,200)` from 445.

## 2026-08-17 — found a second live defect in the query: join fan-out

The plain `LEFT JOIN research_results … result_type='adoption_page'` has no one-row guarantee.
Measured: page `0747e2fc…` (`index` of site `00ff3af5…`) has 2 adoption rows and sits at
nav_order 1 — **today's prompt on that site lists `index` twice inside the visible 10.** With the
cap gone this door opens wider (N research rows → N lines), so the same edit closes it:
`LEFT JOIN LATERAL (… ORDER BY r.created_at DESC LIMIT 1)` — newest per page, shape-preserving,
indexed (`idx_research_page`, `idx_research_created`).

Read-only validation of the proposed text: worst site → 107 rows = population; fan-out site →
40 rows = population, duplicate gone.

## 2026-08-17 — scope decisions, stated so the reviewers don't have to ask

- No Go change: LCO-009's detector is committed (`eb137faed`) and covers the class at the shared
  point; it rides the next roll.
- `rr.summary` kept in the SELECT although unrendered — 275's own `category` reasoning (dropping
  a column a future consumer might read = scope creep for negligible saving). Adjacent finding.
- Prompt template untouched.
- Inner `LIMIT 1` is the fetch-one idiom — outside the silent-cap class by LCO-009's stated
  design (n=1 excluded; end-anchored regex ignores mid-query subquery LIMITs — both arms
  vindicated on live cases in 275's round 2).
