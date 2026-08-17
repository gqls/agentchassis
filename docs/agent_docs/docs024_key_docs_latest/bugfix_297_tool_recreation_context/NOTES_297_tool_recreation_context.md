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

## 2026-08-17 — council submitted (FORCED, with the reason stated), then applied

`SUBMISSION_CORR = 4b9265c3-f6f4-4ed6-a038-f6aaf10b52d8`.

**The gate's client-side scope filter is `^(platform|internal|pkg)/` and this change is
config-only**, so it needed `FORCE=1`. I did NOT do that silently: the rationale opens with the
scope note. The justification is 275's own round — the sharpest objection there
(`bug_historian`, unmarked column truncation) was against the MIGRATION half, not the Go half, so
a config migration to a live shared agent is exactly a shape this gate has caught things in. The
ruling's purpose is that *docs and site content* never spend credits; a live agent's SQL is
neither.

## 2026-08-17 — a risk I raised, then closed rather than leaving for a reviewer

Writing the submission's risks block I noted the LATERAL's `ORDER BY r.created_at DESC` and asked
whether it wanted `NULLS LAST` — `created_at` is nullable, and a NULL sorts FIRST under plain
DESC, which would let an untimestamped row win the "newest" tie for ever.

**I measured instead of asking: 0 of 21 adoption_page rows are NULL today.** So the guard costs
nothing right now, and it makes the bad state unreachable rather than merely unlikely — which is
the "rank by what closes the door" test. Added `NULLS LAST` to the query, the verify block's
literal, the file header and the submission before firing it. **A risk you can close for free is
not a risk to hand a reviewer.**

## 2026-08-17 — applied, verified, recorded

Applied by hand (own file only — `--apply` takes every pending file, and 452 sits in the dir as a
`_HOLD`). Snapshot captured, both gates passed, `UPDATE 1`, post-state verify passed, COMMIT.

| check | result |
|---|---|
| live query | LATERAL present, **no multi-row LIMIT** |
| worst site | **107 rows = full population** (was 10) |
| fan-out site | population rows, duplicate `index` gone |
| snapshot | `agent_definitions_backup` 16:21:26Z (NOT an `is_snapshot` row — the landmine) |
| ledger | `--record-only` with a note, never a hand INSERT |

**A live reminder of the bug's own premise, caught in passing:** the fan-out site's population read
41 while I was validating and **42** ten minutes later, at verify time — another session was adding
pages underneath me. A fixed constant against a growing population is exactly what candidate 4
warned about; this fix leaves no constant.

**Owed, not claimed:** `llm_call_log.prompt_rendered` confirmation needs the next real recreation
run (most recent call 2026-08-11). The query-level disconfirming pair is what is asserted today.
