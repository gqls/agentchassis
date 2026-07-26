# NOTES — bugs_open/074, inline workflow silently ignored

Append-only, newest at the bottom. Evidence, commands, what the system actually said, and every
misstep.

---

## 2026-07-26 — reading the code before touching anything

**Ownership check first** (`scripts/who-owns.py 074`): last touched 07-25 by the filing session,
no owning workstream identified, no competing fix. `site_work_items` has nothing in flight for
scheduled tasks; the `needs_diagnosis` queue is empty. `git log` since 07-25 shows nothing on
`cmd/scheduler/` or the bug file.

**The bug file's mechanism is right but incomplete, and the missing half changes the fix.**
`selectWorkflow` (`platform/messaging/processor.go:882-985`) has three priorities, in order:

1. `msgBody["config"]["workflow"]` — inline override, **taken before group discovery**;
2. group/agent discovery from `config.agent_type` when the action is an orchestration action;
3. the agent definition's own `default_config.workflow`.

So an inline workflow IS honoured — at `body.config.workflow`. Verified it is not a dead path:

```sql
SELECT count(*) FROM orchestration_states WHERE workflow_plan::text LIKE '%call_ingester%';
--  58
```

Those come from `DispatchFeedSourcesAction` (`dispatch_feed_sources_action.go:224`), which builds
`"config": {"workflow": inlineWorkflow, "processing_mode": …, "timeout_seconds": …}` in the
message. Production, not testing.

The scheduler is the one dispatcher that cannot express it. `fireTrigger`
(`cmd/scheduler/main.go:427-433`):

```go
body := map[string]interface{}{
    "action": "orchestrate",
    "config": map[string]interface{}{"agent_type": task.TargetAgentType},
    "input_data": inputMap,
}
```

`config` is built here, from the row's *columns*. Anything the author writes in the `input_data`
column lands under `input_data` — so `input_data.config.workflow` becomes
`body.input_data.config.workflow`, one level below the only place anything reads it.

**Live state, measured (not inferred):**

```sql
SELECT name, enabled, target_agent_type, jsonb_typeof(input_data->'config'->'workflow') AS inline_wf
FROM scheduled_tasks WHERE input_data ?| array['action','config','input_data'];
```

| name | enabled | target | inline_wf |
|---|---|---|---|
| ai-endpoint-health-check | t | endpoint-health-checker | (none — `action` only, 054 family) |
| diagnose-pipeline-trigger | t | diagnose-dispatch-loop | (none — full 054 envelope, empty payload) |
| **evidence-freshness** | **t** | **generic** | **object** |
| adoption-tracker-freshness | f | generic | object |
| ch-fetch-accounts | f | ch-accounts-fetcher | (none — `config.agent_type` only) |

**The model_directory repair is proven, and its own success criterion is a red herring.** Their
seed says the fingerprint of a working sweep is `directory_claims` re-verified > 0. Measured today
that is still **0 of 108** — but the sweep is not broken:

```sql
SELECT count(*) FROM orchestration_states WHERE workflow_plan::text LIKE '%refresh_directory_claims%';
--  1   (COMPLETED, step 'complete', 2026-07-26 09:30)
SELECT collected_data->'refresh_result' FROM orchestration_states WHERE … ;
--  {"checked": 0, "flipped": 0, "dry_run": false, "results": []}
SELECT count(*) FILTER (WHERE is_current AND verified_at + (staleness_days||' days')::interval <= NOW()) FROM directory_claims;
--  0 due, of 82 current
```

It ran, found nothing due, and said so. Before the repair it could not have run at all. That is
the evidence the repair pattern works, and it is the pattern this fix copies.

**Blast radius of activating `evidence-freshness`**, checked before deciding anything:

```sql
SELECT s.domain, count(*) FILTER (WHERE f->'source' ? 'sql') AS sql_facts, count(*) AS all_facts,
       ss.data->>'writer_block_managed' AS managed
FROM site_specs ss LEFT JOIN sites s ON s.id=ss.site_id,
     jsonb_array_elements(COALESCE(ss.data->'facts','[]'::jsonb)) f
WHERE ss.aspect='evidence_base' AND ss.is_current GROUP BY 1,4;
```

| domain | sql facts | facts | writer_block_managed |
|---|---|---|---|
| leopardessconsulting.co.uk | 9 | 18 | true |
| fundamentallyai.com | 7 | 15 | true |
| relojistas.com | 0 | 13 | (unset) |

Eight sites hold an `evidence_base` spec; three hold facts. `fundamentallyai`'s was written
**today at 17:35** by the brochure_component_library thread and feeds their new chart component —
which is why the first pass is a `dry_run`, not a write.

`refresh_evidence_base` is in the live image (`strings /app/agent-chassis | grep -c
refresh_evidence_base` → **11**), so the sweep has been runnable all along and simply had no
caller.

> **CORRECTED, same session, 18:25.** The table above says three sites hold facts. That was
> already false when I wrote it. The dry run found **four** sites with sql-sourced facts — 24 of
> them: leopardess 9, fundamentallyai 7, ai-agent-orchestration 5, robot-hands 3. My query had
> been run minutes earlier, and in between, the **043 lane wrote facts[] into three of those very
> rows** (`site_specs.updated_at` on robot-hands = 18:19:06 UTC; their commit `0c994f2ee`,
> 18:19:29 UTC — "seed real facts[] — the claims checkers were a no-op on the sites that
> fabricated"). Two threads editing the same rows minutes apart, exactly as CLAUDE.md says to
> expect. **What would have caught it:** re-running the count immediately before writing the
> figure down, instead of carrying a number forward from earlier in the same session. The measured
> figure is the sweep's own report, which I now quote instead of my query.

## 2026-07-26 18:22 — the repair, staged through a dry run

Applied `SEED_evidence_freshness_agent.sql`: a new `evidence-freshness` agent_definitions row
carrying the workflow (`dry_run: true` for the first pass), task repointed, `input_data` → `{}`,
`last_triggered_at = NULL`. It fired on the next tick.

**The structural proof, first:**

```sql
SELECT count(*) FROM orchestration_states WHERE workflow_plan::text LIKE '%refresh_evidence_base%';
--  before: 0     after: 1   (correlation 534c70b4-…, COMPLETED at step 'complete')
```

That count could not have been non-zero before the repair, for any reason — the plan the chassis
executed never contained the action. It is the discriminating check; the timestamps are not.

**The dry-run report** (`collected_data->'refresh_result'`, and note it is `collected_data`, not
`final_result`, which is empty on these runs):

| domain | facts checked | would update | drifted | writer_block |
|---|---|---|---|---|
| robot-hands.com | 3 | 0 | 0 | unmanaged |
| fundamentallyai.com | 7 | 4 | 3 | regenerated |
| ai-agent-orchestration.com | 5 | 3 | 1 | unmanaged |
| leopardessconsulting.co.uk | 9 | 6 | 1 | regenerated |
| vonc, oufe, gamesdesign, relojistas | 0 | 0 | 0 | unchanged |

8 sites checked, 5 drifted, 0 errors — and it wrote **nothing**: 0 new `site_specs` rows, 0
`stale_evidence` items, confirmed by query rather than assumed from the flag.

## 2026-07-26 18:24 — writes on, and what actually landed

Flipped `dry_run` to `false`, asserted the flag read back `false`, forced again.

- **3 new `site_specs` revisions** by `created_by='evidence-refresher'`, each carrying a
  `V4 freshness pass: N live-verifiable fact(s) checked, …` note. robot-hands got none, correctly —
  nothing had changed there.
- **`pinned` survived exactly**: leopardess `t`→`t`, the other three `f`→`f` (checked against the
  backup table `_bug074_evidence_base_backup_20260726`, taken before any of this).
- **3 `stale_evidence` work items** raised, `needs_human_review`: fundamentallyai (3 facts),
  ai-agent-orchestration (3), leopardess (1).

**A drifted fact IS still updated.** I first read the per-fact report as "drifted facts keep their
stored value" — wrong, and caught by reading the artefact instead of the report:
`F11-council-rounds-revise` now stores `109` (was 108) with `verified_at` today. The policy is
both/and: the number is mechanically re-synced from the fact's own query, *and* a human is told,
because the published copy quoting it may now be wrong. The counters in the summary line
(`checked / updated / drifted`) overlap; the per-fact `outcome` field is the authoritative read.

**The world moved between the two passes, two minutes apart.** `aao-agent-definitions` read live
176 at 18:22 and 174 at 18:24 — another session was adding and removing agent definitions while I
worked (one of the 176 was my own `evidence-freshness` row). That is not noise to explain away: it
is why `exact`-tolerance facts drift, and why the register re-verifies rather than trusts.

## 2026-07-26 18:28 — the constraint (migration 217)

Hand-applied, then `--record-only` (never `--apply`: 13 files were pending, most of them other
threads').

- The trap is **refused**: `ERROR: new row for relation "scheduled_tasks" violates check
  constraint "scheduled_tasks_no_inline_workflow"`, SQLSTATE 23514.
- **Positive control**: the same probe row with an ordinary payload inserts fine, so the refusal
  is the constraint doing its job and not an unrelated failure.
- 0 surviving violators.

**A false positive worth knowing about:** the runner's idempotency lint flagged 217 as "INSERT
into scheduled_tasks with no guard". 217 contains no INSERT — the lint reads **comment text**, and
my header had a pasteable probe in it. Moved the pasteable SQL to the RUNBOOK, where the
working-docs rules say commands belong; the warning went away. Recorded rather than fixed: the
lint has a twin in `scripts/pattern-check.py` that must change with it, which is the 007 lane's
call, not this one's.

## 2026-07-26 18:35 — the scheduler's own detector

`discardedInlineWorkflow` + refusal in `runTick`, committed `16315d5ab`. **Inert until the next
kafka-scheduler build** — said plainly because a committed fix that hasn't rolled is not a live
fix.

The test was falsified before being trusted: blinding the detector (`if !ok` → `if true`) makes
three subtests fail, and they pass again on restore. A detector test that cannot fail is the same
class of defect as the bug it is testing for.
