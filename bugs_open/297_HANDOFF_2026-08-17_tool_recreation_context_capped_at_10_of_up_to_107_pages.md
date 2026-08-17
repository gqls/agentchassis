# 297 — `tool-recreation-handler` is shown 10 of up to 107 related pages, and the cut is by `nav_order`

**Filed 2026-08-17** by the `bugfix_275_silent_row_caps` lane, from the census `bugs_open/275`'s fix
made possible. **Same defect class as 275, different step, and worse in ratio.**

## The defect

`tool-recreation-handler`'s `load_related_context` step (`agent_definitions`,
`{workflow,steps,load_related_context,config,query}`) ends:

```sql
SELECT p.name, p.title, p.page_type, rr.summary
FROM pages p
LEFT JOIN research_results rr ON rr.page_id = p.id AND rr.result_type = 'adoption_page'
WHERE p.site_id = $1 AND p.name != $2
ORDER BY p.nav_order LIMIT 10
```

The rows become the *related context* an LLM is given when recreating a tool. A row-count `LIMIT`
feeding a prompt is **silent by construction**: the model produces a plausible tool whether it saw
10 related pages or 107, so there is no wrong output to notice and no error to catch. This is
`bugs_open/275`'s mechanism exactly — see it, and register **LCO-009**, for the class.

The cut is by `nav_order`, so what survives is "whatever sits highest in the navigation", which is a
judgement about menus, not about what a tool needs to know.

## Measured 2026-08-17 (live `clients_db`)

Population per site, using the query's own predicate (all pages of the site bar the subject):

| | value |
|---|---|
| cap | **10** |
| sites | 24 |
| **sites where the population EXCEEDS the cap** | **19 of 24** |
| median site | **26** |
| worst site | **107** |

At the median site the model sees **10 of 26 (38%)**; at the worst, **10 of 107 (9%)**. For
comparison, `bugs_open/275` — the bug that prompted this census — was 30 of 74 (41%). **This one is
worse, at more sites.**

## It is LIVE, not latent — the agent demonstrably runs

Checked, because a dormant agent would be a different severity:

- `llm_call_log` for `agent_type='tool-recreation-handler'`: **290 calls, most recent 2026-08-11.**
- `site_work_items` with `handler_agent='tool-recreation-handler'`: 4, most recent 2026-08-11.

⚠ `SELECT ... FROM orchestration_states WHERE owner_agent_type='tool-recreation-handler'` returns
**0** and is the WRONG instrument — it does not carry the handler's name. Use `llm_call_log`.

## Fix candidates, ordered by what closes the door

1. **Bound the PAYLOAD, not the row count** — the shape that worked for 275. Measure which column
   dominates (there, `description` was 80% of the bytes), bound that, and drop the row cap. Closes the
   door: the visible set becomes the population. **Do the measurement first** — 275's `LIMIT 30` was
   not arbitrary, and "just remove it" would have tripled that prompt.
2. **Rank before capping.** If a cap must stay, order by something that means "related" — shared
   entities, page_type affinity, link graph distance — rather than `nav_order`. Weaker: still a silent
   cap, but the cut becomes a judgement instead of a menu artefact.
3. **Ask whether the step needs a corpus at all.** `rr.summary` is a research summary per page; if the
   recreation only needs the site's *shape*, a compact aggregate may beat 10 full rows.
4. **Do NOT just raise the number.** 10 → 30 leaves 5 sites still cut and re-creates this file in a
   month, because the population grows and the constant does not. That is precisely how 275 arrived.

## How to verify a fix

- Re-run the census above: **0 sites over the cap**, or no cap at all.
- The disconfirming pair: pick a page sorting past position 10 by `nav_order` on a site with >10
  pages, fire a tool recreation, and confirm it appears in the `plan`/recreate step's rendered prompt
  (`llm_call_log.prompt_rendered`). Before the fix it **cannot** appear; after, it can.
- ⚠ **If you bound a column instead of rows, MARK the truncation.** 275's first fix cut descriptions
  silently and the council caught it — an unmarked cut is the same defect one level down
  (`016b` §9; migration 446 is the worked remedy, `[…truncated]`).

## Filing basis (owner ruling 2026-07-31)

**No `090` run; substitution stated plainly.** This file asserts no new mechanism — the mechanism is
`bugs_open/275`'s, already diagnosed and **council-approved** (corr `b684a399`), and registered as
LCO-009. What is new here is arithmetic on live data: the query text read from `agent_definitions`,
the population counted with that query's own predicate, and the agent's activity confirmed in
`llm_call_log`. Every figure is reproducible by one query. **Grepped `bugs_open/` and `bugs_closed/`
before filing** — nothing covers this step under any spelling.

## Related

`bugs_open/275` (the sibling, fixed and approved) · register **LCO-009** (the detector — once the
chassis rolls, `QueryDatabaseAction` will WARN when this query fills, which is how this becomes
visible without a census) · `bugs_open/298` (the third instance) · `bugs_open/242` (same class, render
audit) · `016b` §9 (silent caps).
