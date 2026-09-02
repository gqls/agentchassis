# RUNBOOK — calendar_component

Commands that were hard to get right, or worth not re-deriving.

## Current adoption (period-calendar placements, live)

```sql
SELECT s.domain, p.url, p.page_type, pc.created_at
FROM page_components pc
JOIN content_components cc ON cc.id=pc.component_id
JOIN pages p ON p.id=pc.page_id
JOIN sites s ON s.id=p.site_id
WHERE cc.function='period-calendar'
ORDER BY pc.created_at;
```
Run via:
```bash
kubectl -n ai-persona-system exec -i postgres-clients-0 -- psql -U clients_user -d clients_db -t -c "<query>"
```
`pages` has no `page_name` column — it's `name`/`url`/`page_type`. (Tripped once, 2026-09-02.)

## Explicit zero-check for the sibling components (don't trust a GROUP BY silence)

A plain `GROUP BY` on `cc.function` silently omits a component with 0 placements — it
looks identical to "I forgot to check it". Use a LEFT JOIN from the component side:
```sql
SELECT cc.function, count(pc.id)
FROM content_components cc
LEFT JOIN page_components pc ON pc.component_id = cc.id
WHERE cc.function IN ('checklist','period-calendar','comparison-table')
GROUP BY cc.function;
```

## Where the component is defined

- Seed: `docs/agent_docs/sql_for_agents/605_period_calendar_component.sql`
- Rollback: `605_period_calendar_component_ROLLBACK.sql`
- Concept register entry:
  `docs/agent_docs/docs026_concept_register/register/visualisation-and-charts.md`,
  entry VIZ-017 (shared with `checklist`/`comparison-table` — read the whole entry, not
  just the calendar lines)
- Index row: `docs/agent_docs/docs026_concept_register/register/000_concept_index.md`,
  row `VIZ-017`

## Where the design reasoning lives (don't re-derive the boundary decisions)

`bugs_closed/381_HANDOFF_2026-08-24_the_planner_composes_pages_from_components_that_cannot_express_the_page_it_planned.md`
— the motivating bug, closed 2026-08-25, verified at served bytes (after a false
"shipped" claim on a parked domain was caught and corrected — see `WRONG_CALLS.md`,
2026-08-25 entry, before citing "served" from this lane's own history).

Lane docs (now closed, but the reasoning is the record):
`docs/agent_docs/docs024_key_docs_latest/bugfix_381_inexpressive_composition/` — read
`HANDOFF_2026-08-25_continue_here.md` first; it supersedes its own earlier working notes.

## Checking ownership before touching anything calendar-adjacent

```bash
python3 scripts/who-owns.py 381      # closed bug, resolves to bugs_closed/381
python3 scripts/who-owns.py 384      # the listing-invalidation bug some calendar-themed pages tripped — NOT calendar-scope, don't take it on
```
Neither `who-owns.py period-calendar` nor `who-owns.py calendar` matches anything —
there's no bug filed under that slug, only the component seed. Search by function name
in the DB, not by bug slug.

## The "calendar-shaped site" observation (PLAN §2)

Its only written record is in
`docs/agent_docs/docs024_key_docs_latest/loanzy_uk_example_site/NOTES_loanzy_uk_example_site.md`,
around line 1727 ("And the interaction neither lane predicted"), and the shorter version
contributed to the 206 lane:
`docs/agent_docs/docs024_key_docs_latest/bugfix_206_directory_build_handler/CONTRIB_2026-08-25b_from_loanzy_lane_the_canary_cannot_test_your_fix_and_17_of_its_21_pages_are_the_role_you_excluded.md`,
§3.
