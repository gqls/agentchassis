# Retired: `multipage-website-builder` — 2026-08-02

**Retired by owner decision**, raised out of `bugs_closed/092` / `bugs_closed/165`.
Backup: `BACKUP_2026-08-02_multipage-website-builder.json` (both rows, full
`default_config`, 32 KB). Restore: `RESTORE_multipage-website-builder.sql`.

> **This directory is deliberately NOT `docs/agent_docs/sql_for_agents/`.** The
> migration runner applies **every** pending file in that directory, so a restore
> script parked there would be executed by the next `--apply` and silently
> un-retire the agent. Keep retirement artefacts here.

## Why it was retired

It was the sole carrier of `extract_and_sync_links`, and the question "is
`link_registry` empty because the site-id exposure fires, or because the agent
never runs?" had been deferred between two bug files until neither owned it. The
answer was **the agent never runs**:

**THE DURABLE EVIDENCE — `site_specs`, which has no retention job:** across
**1,874 rows, 36 sites, back to 2026-02-25**, the only value ever recorded for
`recommended_builder` is **`pageflow-builder`** (1,216 rows, 14 sites, latest
2026-07-29). `multipage-website-builder` was **never chosen, not once, in five
months**. That is the finding the retirement rests on.

```sql
SELECT v, count(*) AS rows, count(DISTINCT site_id) AS sites, max(created_at)::date
FROM site_specs s, LATERAL jsonb_path_query(s.data, '$.**.recommended_builder') AS v
GROUP BY v ORDER BY rows DESC;
--  "pageflow-builder" | 1216 | 14 | 2026-07-29     <- the only row
```

Supporting, and all-history because that table has no retention job either:
`link_registry` holds **0 rows**, `max(created_at)` NULL — so the action this
agent uniquely carries has never written anything.

> **CORRECTION, made the same day this file was written.** The first version of
> this section rested on orchestration counts and said *"0 orchestrations in the
> retained window … `orchestration_states` is retention-clocked (oldest row
> 2026-07-13), so that means ~20 days, not never."* **The 20 days was wrong: it is
> ~24 HOURS.** `COMPLETED` rows are reaped after about a day (measured: 2,504
> COMPLETED, oldest 24.7h; FAILED oldest 25.4h), and the whole-table
> `min(created_at)` reads 2026-07-13 only because `CANCELLED`, `RUNNING` and
> `INITIALIZED` stragglers are *not* reaped — a floor set by the statuses the
> census was not about. I watched a row I had quoted at 09:40 vanish by 10:40.
>
> So the orchestration evidence never supported "never runs"; it supported "not in
> the last day". The conclusion survived because `site_specs` answers it flatly
> over five months — **right answer, wrong reason, and the reason is what was
> published.** Filed fleet-wide in `LANDMINES.md`
> ("`orchestration_states` keeps terminal rows ~24 HOURS").

## It was NOT dead code — it was a live menu option

This is the part worth knowing before reviving or removing anything else.
`intake-orchestrator` discovers builders dynamically:

```json
"action": "query_agent_definitions",
"config": {"filter": {"type_pattern": "%-builder"}},
"output_field": "available_builders"
```

…then spawns via `agent_type_field: "confirmed_type.recommended_builder"` — a
string chosen at runtime by an LLM classifier or by a human in the HITL
`dynamic_select`. **A config grep can never see that reference**, because the
value does not exist until the run. `build-briefing-agent` has the same shape
(`site_specs.specs.classification.recommended_builder`).

So the agent was reachable, in two ways:

1. `site-classifier`'s prompt asks the model to "recommend the appropriate
   builder" from the live `available_builders` list — unconstrained.
2. A human could pick it from the intake confirmation form.

What made it *safe* to retire is that the primary classifier
(`domain-research-classifier`) is pinned: its prompt says
`recommended_builder should always be "pageflow-builder" for now`.

## Pre-flight checks that were run (all clear)

| check | result |
|---|---|
| other agent configs referencing it | 0 |
| `site_specs.data` / `.source_agent` | 0 |
| `site_work_items.handler_agent` / `.spec` | 0 |
| `orchestration_states` (collected + initial request) | 0 |
| FK referrers (`agent_definitions.previous_version_id`, `client_system.agent_instances.template_id`) | 0 |
| snapshots / already-deleted rows for this type | 0 / 0 |

## How it was retired, and why not a physical `DELETE`

```sql
UPDATE agent_definitions
SET is_active = false, deleted_at = now(), updated_at = now()
WHERE type = 'multipage-website-builder';
```

`is_active = false` **is** the delete that matters here: `query_agent_definitions`
defaults to `active_only`, so this is precisely what removes it from the builder
menu — verified against the action's own WHERE-clause construction
(`query_agent_definitions_actions.go:126`), not assumed. `deleted_at` additionally
matches the platform-wide read convention (`deleted_at IS NULL` filters
everywhere).

A physical `DELETE` would change no behaviour that a reader can observe, and would
trade the one-line reversal below for a JSON-reimport. If the rows are ever
physically removed, restore from the JSON backup rather than by hand — the
`default_config` is 4.4 KB of workflow per row.

## Consequence for `bugs_closed/165` site C

`extract_and_sync_links` is now unreachable, so its completeness floor
(`link_registry_prune_floor.go`) guards an action nothing can invoke. It was
already inert by construction (`Stored=0` cannot refuse); it is now inert because
its only carrier is retired. **Keep the file** — it costs nothing, it is
mutation-proven, and it arms itself if the agent is ever revived. But its live
induction is not "pending", it is **moot** unless this retirement is reversed.

## What the same query says about the OTHER builders — likely a bigger decision

The builder menu went 7 → 5 with this retirement (it held
`multipage-website-builder` **twice**, once per live row, so the classifier and any
human were being offered the same builder as two separate options). The five that
remain:

`content-site-builder` · `landing-page-builder` · `pageflow-builder` ·
`report-builder` · `website-builder`

Run the same durable `site_specs` census against them and only **`pageflow-builder`**
has ever been chosen. The other four have **never** been selected in five months —
the same standing this agent was retired for. They were left alone because
retiring them is the owner's call, not a rider on this one; but the decision is
plainly about the whole menu rather than about one member of it.

Related: new sites *are* being created (25 in the last 30 days, including today),
yet `intake-orchestrator`, `site-classifier` and `build-briefing-agent` show no
recent runs — so whatever creates sites now, it is not the intake→classify→spawn-a-builder
path this menu belongs to. **That is a question about the build path's shape, and
it is worth answering before retiring anything else in the menu** — the honest
reading is "this whole subsystem may be superseded", not "these four agents are
individually dead".
