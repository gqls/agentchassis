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

> **ANSWERED 2026-08-02 evening, and the paragraph that used to sit here was
> wrong.** It said `intake-orchestrator`, `site-classifier` and
> `build-briefing-agent` "show no recent runs", concluding the whole subsystem
> might be superseded. That came from `orchestration_states`, which reaps terminal
> rows after ~24h — the same trap corrected further up this very file.
>
> **The intake path is NOT superseded; it was re-plumbed** to a work-item model.
> Live and running *today*, evidenced from `site_work_items.handler_agent` and
> `site_specs.created_by` (both durable, no reaper):
> `domain-submitter` → work item → `domain-research-classifier` →
> `domain-strategist` → `vertical-exemplar-researcher` → `site-design-planner` →
> `build-briefing-agent` → **`build-site-planner`**, which is the builder now
> (`plan_site`, `sync_pages`, `populate_nav`, `emit_design`).
>
> What *is* superseded is narrower: `intake-orchestrator` (0 work items, 0 specs,
> no scheduled task, nothing spawns it), `site-classifier` (same), and **the whole
> `%-builder` menu — no work item has EVER named a builder, all history.**
>
> So the remaining five can be retired on *better* evidence than this agent was.
> **But fix `domain-research-classifier`'s prompt first if you retire
> `pageflow-builder`**: it mandates `recommended_builder` and pins it to that name,
> has written it into 1,216 spec rows, and nothing consumes it — an LLM instructed
> on every run to fill a field for a consumer that no longer exists.

---

# Second retirement — 2026-08-02 evening: three more builders

**Retired:** `content-site-builder`, `landing-page-builder`, `website-builder`.
Backup: `BACKUP_2026-08-02_three_unused_builders.json` (3 rows, 43 KB, 13 workflow
steps each). Same restore semantics as above — the rows are soft-retired, so
recovery is one `UPDATE`.

**Kept, by owner instruction:** `pageflow-builder` (so
`domain-research-classifier`'s pinned `recommended_builder` still resolves — no
prompt change needed).

**KEPT AGAINST THE INSTRUCTION, and this is the finding:** `report-builder`. The
owner said "retire the others", on my report that no work item had ever named a
builder. That report was **scoped too narrowly** — it filtered
`site_work_items.handler_agent LIKE '%-builder'`, which is empty, and I read that
as "no builder is dispatched". `report-builder` is dispatched by a different
route:

- **`report-dispatch` is an ENABLED scheduled task on a 90-second tick**, last
  fired 2026-08-02 21:13. Its `report-dispatch-loop` claims
  `pipeline='reports' AND item_type='report_request' AND status='awaiting_report'`
  from `site_work_items` and spawns **the handler named on the item** — its own
  step description says "(report-builder)".
- **8 rows in `client_system.agent_instances` reference it via `template_id`**,
  newest 2026-07-31. That is a live FK, and a hard `DELETE` would fail on it.

The queue is empty *today*, which is why every "has it run" check reads zero. **An
empty queue plus an enabled dispatcher is not a dead agent — it is a live one with
nothing to do**, and retiring it would break the loop the moment a report is
requested. Left active; the owner can decide separately with this in front of them.

## Pre-flight for the three that were retired (all clear)

| check | content-site | landing-page | website |
|---|---|---|---|
| `site_work_items.handler_agent` / `.spec` | 0 | 0 | 0 |
| `site_specs.created_by` / `.data` | 0 | 0 | 0 |
| `agent_instances.template_id` (FK) | 0 | 0 | 0 |
| other agent configs naming it | 0 | 0 | 0 |
| live rows / snapshots / already-deleted | 1 / 0 / 0 | 1 / 0 / 0 | 1 / 0 / 0 |

**One false positive worth recording:** the blast-radius query also matched **15
`orchestration_states` rows** — all owned by `council-gate`, and all of them my own
council submissions from earlier today, whose rationale text *names these builders
while arguing about them*. Text under review scoring as usage of the thing under
review is the `llm_call_log.prompt_rendered` landmine in a second table. Check the
`owner_agent_type` before treating a `collected_data` match as evidence of use.

---

# Third retirement — 2026-08-02 late: the last two of the old shape

**Retired:** `intake-orchestrator`, `site-classifier`. One live row each.
Backup: `BACKUP_2026-08-02_intake_path_orphans.json` (2 rows, 16 KB).
Restore: `RESTORE_intake_path_orphans.sql` — **they are a pair, restore both or
neither**, because `intake-orchestrator`'s workflow spawns `site-classifier` and
nothing else reaches it.

This finishes the job the section above left open ("the last of the old shape…
not retired, because nobody asked"). Owner asked. Same evidence standard, plus one
axis the previous two retirements did not have to think about.

## The axis that was missing: the caller is a HUMAN, and no query can see one

Every check in this directory so far asks the database who references an agent.
That is sound for an agent spawned by another agent. **`intake-orchestrator` is an
entry point — it is spawned by an operator publishing to Kafka**, so it would read
as an orphan on every DB axis whether it were dead or in daily use:

```bash
# scripts/initial_messages/090_new_build/…/075_new_build_…_intake_orchestrator.sh
kcat -P -b $BOOTSTRAP -t system.agent.generic.requests …
{"action":"orchestrate","config":{"agent_type":"intake-orchestrator"}, "input_data":{…}}
```

So the retirement rests on a **file-date** comparison in `scripts/`, not on a row:

| entry-point script | routes to | last touched |
|---|---|---|
| `090_new_build/…_intake_orchestrator.sh` | **`intake-orchestrator`** | **2026-06-21** |
| `020_build_pipeline/082_submit_domain_unified.sh` | `domain-submitter` / `site-adoption-orchestrator` | **2026-07-30** |

`082` is explicit in its own header that it is "one entry point for FRESH and
ADOPT submissions", and its comment block draws the live graph
(`needs_domain_research → domain-research-classifier → … → build-site-planner`).
The operator path **moved**; it did not lapse. That is what makes these two safe,
and it is evidence of a kind none of the earlier retirements needed.

> **Generalise this before the next retirement.** §3 of the handoff says *an
> absence of WORK is not an absence of WIRING* (`report-builder`, saved by a
> `scheduled_tasks` row). This is the **third** spelling of the same error and the
> one with no row to find at all: **an absence of WIRING is not an absence of a
> CALLER, when the caller is a person with a script.** For any agent that looks
> like an entry point, grep `scripts/` and compare dates — the question is not
> "does something reference it" but "has the habit moved".

## Pre-flight (all clear)

| check | intake-orchestrator | site-classifier |
|---|---|---|
| `site_work_items.handler_agent` / `.spec` text | 0 / 0 | 0 / 0 |
| `site_specs.created_by` / `.data` text | 0 / 0 | 0 / 0 |
| `scheduled_tasks` (target **or** named in `input_data`) | 0 | 0 |
| `client_system.agent_instances.template_id` (FK) | 0 | 0 |
| `agent_definitions.previous_version_id` (FK) | 0 | 0 |
| `orchestration_states` (with `owner_agent_type`) | 0 | 0 |
| other live agent configs naming it | 0 | **1 — `intake-orchestrator` itself** |
| Go / yaml / TS code | 0 | 0 |
| live rows / snapshots / already-deleted | 1 / 0 / 0 | 1 / 0 / 0 |
| **operator scripts** | superseded 2026-07-30 | — |

The single config referrer is the pair bond: they name each other and nothing
else names either. All repo hits outside `scripts/` are seed SQL under
`docs/agent_docs/sql_for_agents/`, which is history, not the system.

## How they were retired

```sql
UPDATE agent_definitions
SET is_active = false, deleted_at = now(), updated_at = now()
WHERE type IN ('intake-orchestrator','site-classifier');
```

Soft, for the reasons given above: reversal stays one `UPDATE`, and `is_active`
is what actually removes an agent from any `query_agent_definitions` lookup.

## What is now true of the builder menu

`intake-orchestrator` was the **only** consumer of the `%-builder` menu
(`fetch_available_builders` → `agent_type_field:
confirmed_type.recommended_builder`). With it retired, **nothing reads that menu
at all**, which retires the open question in the section above about
`domain-research-classifier`'s pinned `recommended_builder`: that field was already
written into 1,216 spec rows for a consumer that no longer existed, and now the
menu it names has no reader either. `pageflow-builder` must still **resolve** —
the classifier's prompt mandates the string — but nothing dispatches on it.
That is a prompt-hygiene item, not a live defect, and it is unchanged by this
retirement.
