# Analysis — Two defects found during Phase 2F runtime testing

**Date observed:** 2026-05-11
**Status:** Discovery, parked for deep discussion. Not blocking the migration's
SQL-level correctness; both surface during runtime exercise of the new path.
**Context:** Triggered as part of Phase 2F end-to-end test. Migration is
correctly applied to `agent_definitions` (verified via SQL). The runtime
behaviours below are independent of the migration.

This document complements `ANALYSIS_chassis_response_consumer_group_race.md`
(2026-05-10). Together they make up three discovered defects parked for the
deep review session.

---

## Defect 1: Chassis ships pre-migration workflow definition despite correct DB state

### What was observed

The Phase 2F migration applied successfully at 16:09 UTC on 2026-05-10. The
active row of `agent_definitions` for `asset-deployer` was updated:

```json
"input_fields": ["s3_uri", "deploy_path", "purpose", "domain", "asset_key"]
"input_contract.optional": ["deploy_path", "purpose", "asset_key"]
```

(Confirmed by direct SELECT against the active row — `is_snapshot=false, is_active=true`.)

Two test runs followed: one against a chassis pod that pre-dated the migration,
one against a chassis pod that started **after** the migration was applied (post
`rollout restart`). Both runs produced spawn messages with the **pre-migration**
workflow embedded in `agent_config`:

```json
"workflow": {
  "steps": {
    "deploy_asset": {
      "config": { "input_fields": ["s3_uri", "deploy_path", "purpose", "domain"] }
    }
  }
}
```

The extractor in each spawned pod requested only those four fields, ignoring
the `asset_key` present in `input_data`, and the deploy went to the canonical
`assets/images/hero.jpg` instead of the variant-keyed path.

The chassis restart (which would have cleared any in-memory cache) **did not
change the behaviour**. So the original "stale in-memory cache" hypothesis is
refuted. The chassis is reading something stale from a source that survives
a process restart — i.e. from the database itself, just not the row we'd
expect it to read.

### Revised hypothesis: snapshot row is shadowing the active row

The `snapshot_agent()` function in `021_model_swap_and_rollback.sql` inserts
snapshots with `version + 1000` to avoid unique constraint conflicts:

```sql
INSERT INTO agent_definitions ( ..., version, is_snapshot, is_active, ... )
SELECT  ..., version + 1000, true, false, ...
```

So after the migration we have:

| id                                      | version | is_snapshot | is_active | input_fields                               |
|-----------------------------------------|---------|-------------|-----------|--------------------------------------------|
| `e9a9bac9-…` (active, migrated)         | 1       | false       | true      | `[..., asset_key]`                          |
| `10a8978f-…` (snapshot, pre-migration)  | 1001    | true        | false     | pre-migration                              |

If the chassis loads agent definitions with a query like:

```sql
SELECT default_config FROM agent_definitions
WHERE type = 'asset-deployer' AND deleted_at IS NULL
ORDER BY version DESC
LIMIT 1;
```

— without filtering `is_snapshot = false` — then **version 1001 (snapshot)
sorts above version 1 (active)** and the chassis reads the snapshot's
workflow. This is exactly the pattern observed: the spawn message ships
pre-migration `input_fields`, but the active DB row has post-migration values.

### Why this is structural, not transient

If this hypothesis is correct, the issue isn't with caching. It's with the
snapshot mechanism in `021_model_swap_and_rollback.sql` creating rows that
sort ahead of active rows in version-descending queries. Any code path that
picks "the most recent" definition by version without excluding snapshots
will pick the snapshot.

This makes the snapshot/rollback feature a structural trap: every snapshot
shadows the active definition until either the loader is fixed or the
snapshot is deleted.

### Diagnostic SQL to confirm

To definitively confirm the hypothesis, run this query — it mimics a
naive "most recent" lookup:

```sql
SELECT id, version, is_snapshot, is_active,
       default_config->'workflow'->'steps'->'deploy_asset'
         ->'config'->'input_fields' AS input_fields
FROM agent_definitions
WHERE type = 'asset-deployer' AND deleted_at IS NULL
ORDER BY version DESC
LIMIT 1;
```

**If this returns the snapshot row's input_fields (the pre-migration list),
the snapshot-shadowing hypothesis is confirmed.** The active row is correct;
the chassis just isn't reading it.

The next confirmation step is to find the actual loader query in the chassis
code:

```bash
grep -rn "FROM agent_definitions" --include="*.go" .
grep -rn "default_config" --include="*.go" .
```

Look for the WHERE clause. If it lacks `is_snapshot = false` (or equivalent),
the hypothesis stands.

### Possible fixes

1. **Add `is_snapshot = false` to the chassis loader's WHERE clause.**
   Structural fix. Doesn't change the snapshot mechanism. Doesn't affect
   anyone else's queries (each loader manages its own filter).

2. **Change the snapshot mechanism to not use `version + 1000`.** For
   example: store snapshots in a separate `agent_definitions_snapshots`
   table; or use the same `version` but rely on `is_snapshot` flag and
   never let snapshots sort ahead in a version-ordered query.
   Higher-impact — touches `021_model_swap_and_rollback.sql` and possibly
   restore/rollback logic.

3. **Delete the offending snapshot (workaround).** Removes the shadowing
   row but loses the rollback path. Fast, reversible only by re-running
   `snapshot_agent()`. Not a real fix.

The structural fix is #1 if the hypothesis is right. The work is a
single-line SQL change in one Go function.

### What this is not

This is NOT a "workflow override is broken" issue. The chassis intentionally
supports override: a caller can include `config.workflow` (or `agent_config`)
in the trigger to override DB defaults. That's correct, expected behaviour —
used for testing, one-off variants, gradual rollout. None of that is
affected by this defect. The defect is only about which DB row the chassis
reads when nothing in the trigger specifies an override.

### Workaround that DOES NOT work

Restarting the chassis Deployment was tried and made no difference. The
pod that ran the second test was 2 minutes old at trigger time, started
20+ hours after the migration committed. It still shipped the
pre-migration workflow. So in-memory cache was not the problem and
"rollout restart" is not the workaround.

### Workaround that WOULD work (untested)

If the snapshot-shadowing hypothesis is correct, deleting the snapshot
row should restore correct behaviour:

```sql
-- TEST WORKAROUND — do not apply structurally
DELETE FROM agent_definitions
WHERE id = '10a8978f-07e5-4a7e-8a32-a53aaa8c55c7';  -- the snapshot
```

Then re-trigger the test. If `assets/images/hero-about.jpg` lands, the
hypothesis is confirmed. If `hero.jpg` is overwritten again, the
hypothesis is wrong and we need to look elsewhere.

This costs the rollback path for asset-deployer (no snapshot to revert
to), so it's a one-time test workaround, not an operational fix.

### Severity assessment

High. Every snapshot taken via `snapshot_agent()` since the system
launched may be shadowing its corresponding active row in any
agent-definition lookup. We've only noticed because Phase 2F is the
first work to actually depend on a value that differs between the
active and snapshot rows. There may be other latent cases where a
recent change is silently using stale config because a snapshot is
shadowing it.

The fix is small but consequential. Worth a careful review before
applying because it changes the semantics of the loader query.

### Questions for the deep discussion

1. **Confirm the hypothesis with the diagnostic SQL above.** Five
   minutes; highest information yield of any next step.

2. **Find every agent-definition loader.** `grep -rn "agent_definitions"`
   across the Go source. Each one needs its WHERE clause audited for
   `is_snapshot` handling.

3. **Audit other tables with the same snapshot pattern.** Does
   `021_model_swap_and_rollback.sql` define snapshot functions for other
   tables? If so, the same shadowing risk applies to them.

4. **Decide on the canonical filter.** Should it be `is_snapshot = false`,
   `is_active = true`, or both? Subtle: a definition could exist that's
   neither active nor snapshot (e.g. mid-deploy). Need a clear rule.

5. **Add a regression test.** A SELECT against agent_definitions that's
   expected to return the active row, run as part of CI or migration
   verification, would catch this immediately. The migration's
   verification queries currently filter by `is_active=true` explicitly,
   so the migration "looks correct" even though the chassis behaves
   incorrectly. The gap is between the verification query and the
   chassis's actual query.

---

## Defect 2: Git adapter responds on partition 1 of a single-partition topic

### What was observed

The git adapter successfully committed `assets/images/hero.jpg` to the
robot-hands.com repo at 12:20:31. Commit SHA `105a59e124c8ea00c915f059a812a920ff98e17f`
is in the repo.

Immediately after, it tried to publish the success response back to the
asset-deployer's per-spawn responses topic:

```
topic="job.9bff1ece-2a6b9417-asset-deployer-spawn_deployer.responses"
partition=1
```

Kafka rejected with:

```
fetch request error: topic partition not found
  (topic="job.9bff1ece-2a6b9417-asset-deployer-spawn_deployer.responses"
   partition=1)
```

The per-spawn topic was created with **one** partition (partition 0). The
adapter tried to write to partition 1. The response was lost.

Confirmation that the topic has only one partition comes from the kafka
broker's cleanup log later that day:

```
Deleting log /var/lib/kafka/data/kafka-log2/
  job.9bff1ece-2a6b9417-asset-deployer-spawn_deployer.responses-0
```

The suffix `-0` is the partition number. Only `-0` was logged — no `-1`,
`-2`, etc. existed.

### Update from 2026-05-11 13:27 run (second test)

A second test under correlation `ac09c48d-…` ran the same trigger after a
chassis restart. This time:

- The git adapter wrote successfully to the per-spawn responses topic.
- No partition error in the git-adapter logs.
- Asset-deployer received the response cleanly.
- Both orchestrations (parent generic and child asset-deployer) reached
  `COMPLETED` status.

So defect 2 did not reproduce on the second run. This downgrades severity
significantly. Possibilities for why the first run hit it and the second
didn't:

1. **Transient kafka metadata propagation.** The per-spawn topic
   `job.9bff1ece-…` was created moments before the git adapter tried to
   write to it. If the adapter's kafka client hadn't yet refreshed
   partition metadata for the new topic when it constructed the write,
   the balancer could have picked an out-of-range partition.

2. **Git-adapter pod restart picked up a fix.** The git-adapter pod that
   handled the second run (`git-adapter-68cc4dd455-nbmb9`) was restarted
   between the two test runs. If a fix existed in the running image
   (`v1.0.44`) but was masked by the first pod's state (e.g. cached bad
   metadata), the restart would have cleared it.

3. **The kafka topic cleanup cron** that ran between the two test runs
   may have cleared some metadata state that contributed to the first
   failure.

None of these is confirmed. The defect happened once, did not reproduce
on a repeat under similar (but not identical) conditions. It may be a
genuine transient kafka-go metadata-cache issue that's hard to reproduce
on demand.

### Revised severity assessment

Lower than initially stated. The defect happened once and did not
reproduce. Not blocking Phase 2F. Worth keeping on the watch list because:

- Same code path (kafka-go writer with `LeastBytes` balancer on per-spawn
  topics) is used by multiple adapters. If one hit it, others can.
- If it happens occasionally, that's a worst-case bug: orchestrations
  fail intermittently with the underlying work having succeeded.

Action: monitor adapter logs for partition errors over the next week of
runtime. If it recurs, escalate to a real investigation. If it doesn't,
the root cause can be left as "probably transient kafka-go metadata
caching, not actionable without a repro".

### Where the wrong partition comes from

The git adapter uses `kafka.NewProducerWithValidator(...)` which constructs
a `kafka.Writer` (kafka-go library). That writer has a `Balancer` field
that determines partition selection. In the standard producer construction
(see `platform/kafka/producer.go`):

```go
writer := &kafka.Writer{
    Addr:         kafka.TCP(brokers...),
    Balancer:     &kafka.LeastBytes{},
    RequiredAcks: kafka.RequireAll,
    ...
}
```

`kafka.LeastBytes` picks the partition that has received the least data.
For per-spawn topics with only one partition, it should always pick 0 —
unless the partition count is not yet known to the writer's metadata cache,
in which case it may default to a fixed assumption (which might be 1).

This is a plausible explanation but not confirmed. The actual cause needs
a trace through kafka-go's `LeastBytes.Balance` logic and how partition
metadata is loaded for a topic the producer hasn't written to before.

### Possible fixes (for discussion)

1. **Use `&kafka.Hash{}` or `&kafka.RoundRobin{}` balancer instead of
   LeastBytes.** Both interrogate partition metadata; both should pick a
   valid partition. May still hit the same metadata-cache issue if it's a
   library-level problem.

2. **For per-spawn topics (always single-partition), explicitly write to
   partition 0.** Set `kafka.Message.Partition: 0` when producing. Bypasses
   the balancer entirely. Most targeted fix but couples the producer to
   the knowledge that per-spawn topics are single-partition.

3. **Force a metadata refresh before the first write.** Probably what's
   actually missing.

4. **Make per-spawn topics multi-partition by default.** Mostly works
   around the bug. Per-spawn topics don't need multiple partitions
   semantically — they have exactly one producer and one consumer — so
   this is wasteful.

The right fix needs trace data first.

### Severity assessment

High. Every asset-deployer call (and by extension the new Phase 2F path
in image-build-handler, and the existing pageflow-builder /
site-work-orchestrator deploy paths once those move to the asset-deployer
pattern) hits this. Deploys succeed but orchestrations report failure.
Downstream branches on failure. Looks like a regression of work that
isn't actually regressed.

Workaround: none clean. The downstream parent could be made tolerant of
"deploy succeeded but no response" by checking the git repo state — but
that's a substantial change for a workaround.

Most likely the same defect affects other adapters that respond to
per-spawn topics (webscrape, image-generator). Worth a sweep.

### Questions for the deep discussion

1. **Is this a recent regression?** Did adapter→asset-deployer response
   routing ever work? The STATUS doc says `hero_about` reached
   `store_variant_asset` on 2026-05-08 — but that path doesn't use the
   git adapter response; store_asset is a local action. The first time
   asset-deployer would await a git adapter response is exactly the
   migration we just tested. So this may always have been broken.

2. **Are other adapters affected?** The git adapter and image-generator
   adapter both write responses to per-spawn topics. The image-generator
   adapter's recent timeout error (separate bug —
   `json: unsupported type: func() (io.ReadCloser, error)`) prevented us
   from observing whether it would have hit this same partition issue.

3. **What was the original intent for partition count on per-spawn
   topics?** Is it explicit somewhere (topic_manager.go) or accidental?

---

## How the three parked defects interact

The three defects are independent in mechanism but were all latent because
nothing prior to Phase 2F exercised this exact path: trigger via generic
chassis → spawn child → child uses DB-defined workflow → child awaits
adapter response.

- The **consumer-group race** (separate doc) affects responses on shared
  topics (`system.agent.generic.responses`) and manifests only when
  multiple chassis pods are running.

- The **snapshot-shadowing defect** (defect 1 here) affects any chassis
  read of an `agent_definitions` row that has a snapshot taken from it.
  It manifests whenever the active row differs from the snapshot. Latent
  prior to today because the migration we just applied is the first time
  the active row diverged from a snapshot in a way that mattered to runtime
  behaviour.

- The **adapter partition routing** (defect 2 here) affects adapter
  responses on per-spawn topics. Manifested once; did not reproduce on
  second run. May be transient. Worth monitoring.

Phase 2F is the first work to depend on all three paths cleanly. Each
defect blocks (or threatens) a different segment:

- Defect 1 prevents the chassis from shipping the right child workflow.
  Hard block on Phase 2F's intended behaviour.
- Defect 2 (when it occurs) prevents the child from receiving its adapter
  response. Causes orchestration timeout even when underlying work
  succeeded.
- Consumer-group race prevents multiple chassis replicas from running
  cleanly. Currently worked around with replicas=1.

The right sequence to address them, given updated severity:

1. **Confirm and fix defect 1 (snapshot shadowing).** Highest priority.
   Hypothesised cause is concrete; one-line fix likely. Blocking Phase 2F's
   `asset_key` path from working.
2. **Consumer-group race.** Real architectural defect. Required before
   scaling chassis back up.
3. **Defect 2 (adapter partition).** Monitor; may not need intervention.
   If recurrence is observed, reopen.

None of these defects is caused by the Phase 2F migration. The migration's
SQL is correct (verified). All three were already present in the system
and surfaced because Phase 2F is the first work to exercise these paths
together at runtime.
present and surfaced because Phase 2F is the first work to exercise these
code paths together at runtime.

---

*Parked for deep discussion. Nothing in this document changes anything;
material for review.*
