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

| check | result |
|---|---|
| orchestrations for `multipage-website-builder` | **0** in the retained window |
| orchestrations ever mentioning `links_extracted` | **0** |
| `link_registry` rows, all-history (no retention job on that table) | **0**, `max(created_at)` NULL |
| the live build pipeline, same window | `build-dispatch-loop` 588 · `build-pipeline-trigger` 587 · `page-rerender` 22 · `page-build-handler` 1 — this agent absent |

**Scoped honestly:** `orchestration_states` is retention-clocked (oldest row
2026-07-13), so "0 orchestrations" means *~20 days*, not *never*. The all-history
half is `link_registry` itself. It is the **combination** that makes the
conclusion safe.

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
