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

## Defect 1: Chassis appears to cache stale agent definitions across migrations

### What was observed

The Phase 2F migration applied successfully at 16:09 UTC on 2026-05-10. The
active row of `agent_definitions` for `asset-deployer` was updated to:

```json
"input_fields": ["s3_uri", "deploy_path", "purpose", "domain", "asset_key"]
"input_contract.optional": ["deploy_path", "purpose", "asset_key"]
```

(Confirmed by direct SELECT.)

A test trigger 20 hours later (2026-05-11 12:20 UTC) spawned an asset-deployer
Job pod. The spawn message embedded an `agent_config` block carrying the
asset-deployer's workflow definition. That embedded workflow contained the
**pre-migration** `input_fields`:

```json
"input_fields": ["s3_uri", "deploy_path", "purpose", "domain"]
```

The extractor in the spawned pod then logged:

```
"requested_fields": ["s3_uri", "deploy_path", "purpose", "domain"]
"available_keys":   ["domain", "s3_uri", "purpose", "asset_key"]
```

`asset_key` was present in `input_data` (delivered by the parent's input_mapping
in the inline wrapper) but the spawned action didn't extract it, because the
embedded workflow didn't ask for it. Result: the deploy ran with no `asset_key`,
the action fell through to its purpose-derived default path
(`assets/images/hero.jpg`), and the variant-keyed path
(`assets/images/hero-about.jpg`) was not produced.

### Where the stale workflow came from

The spawning chassis pod assembled the spawn message. The workflow embedded in
that message was the pre-migration version. The pod was older than the
migration's commit time (chassis Deployment had not been restarted since
before 2026-05-10 16:09). Most likely explanation: the chassis caches agent
definitions in memory at startup or first lookup, and that cache is not
invalidated by `agent_definitions` UPDATEs.

This is an inference, not yet traced to the cache itself. A direct read of
the agent-definition load path (likely something like `LoadAgentDefinition`
in the chassis startup or per-spawn lookup) would confirm whether it queries
the DB on every spawn or caches.

### Important nuance: this is not "override is broken"

The chassis intentionally supports override: a caller can include
`config.workflow` (or `agent_config`) in the trigger to override the DB
definition. That's correct, expected behaviour — used for testing,
one-off variants, and gradual rollout. The defect is **not** that override
exists. The defect is that **when nothing in the trigger specifies a workflow
for the spawned child, the chassis appears to fall back to a cached copy
rather than re-reading the DB**.

The test trigger that surfaced this had an inline wrapper workflow at the
top level (`config.workflow`), but did NOT specify an override workflow for
the asset-deployer child. The child's workflow should therefore have come
from the DB's current row. It didn't.

### Workaround applied today

Rolling restart of the chassis Deployment will pick up the current
agent_definitions row on next startup:

```bash
kubectl -n ai-persona-system rollout restart deployment agent-chassis
kubectl -n ai-persona-system rollout status deployment agent-chassis
```

This is a workaround, not a fix. It implies every agent-definition change
needs a chassis restart to take effect, which isn't documented anywhere and
isn't enforced by tooling.

### Questions for the deep discussion

1. **Where is the cache?** Trace the code path from spawn_agent → DB lookup
   → embedded workflow assembly. Is it `LoadAgentDefinition`? Is there an
   in-memory map keyed on agent_type?

2. **What's the intended cache invalidation model?** Options:
   - Reload on every spawn (cheap, but does N DB queries per orchestration
     instead of 1; arguably the right default since spawns are sparse).
   - TTL-based cache (e.g. 60s). Cheap most of the time; bounds staleness.
   - Pub/sub invalidation via a kafka topic when agent_definitions changes.
   - Manual flush via API or signal. Worst option for routine ops.

3. **Does this affect other config sources?** If agent definitions are
   cached, are `style_collections`, `content_components`, `site_specs`
   also cached and similarly stale? Worth checking before deciding on a
   cache model — a unified approach would be better than per-table fixes.

4. **Is there a "reload-on-version-bump" pattern available?**
   `agent_definitions` has a `version` column. The chassis could be told to
   reload by an external trigger that bumps version. Less convenient than
   automatic invalidation but explicit.

### Severity assessment

Medium. Not a data-loss bug, but it silently produces wrong behaviour:
migrations applied to `agent_definitions` don't take effect until a chassis
restart, with no signal that anything is wrong. Anyone applying an agent-
definition change without knowing this will wonder why their change isn't
working. We just spent a debug cycle on exactly that.

The fix needs the cache located and invalidation designed. The workaround
(restart after agent-definition changes) is sufficient operationally as
long as it's documented and routine.

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

### Why this matters

The asset-deployer entered AWAITING_RESPONSES at 12:20:28 expecting a reply
from the git adapter on this topic. The reply was never delivered. After
~3 minutes the asset-deployer's coordinator fired a timeout
(`coordinator.go:3000 — Request timed out`).

Meanwhile, the deploy itself had completed successfully. The orchestration
reports failure because the response didn't arrive, even though the
underlying work succeeded.

This shifts every asset-deployer call from "completes cleanly" to
"deploys but reports timeout". From the parent's perspective (the inline
wrapper, and by extension `image-build-handler` post-migration), this looks
like the deploy failed. Downstream logic that branches on `deploy_result`
will go down the failure path. Visible symptom: the file IS in git but the
orchestration is FAILED.

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

All three were latent and hadn't surfaced together because:

- The **consumer-group race** affects responses on shared topics
  (`system.agent.generic.responses`). It manifests when multiple chassis
  pods are running.

- The **stale agent-definition cache** affects DB-driven changes that
  aren't surfaced until a pod restart. It manifested today because we
  applied a workflow change and tried to use it.

- The **git adapter partition** affects responses on per-spawn topics from
  adapters. It manifested today because asset-deployer is the first agent
  to round-trip through the git adapter for image deploys (previously the
  deploys ran in-chassis and there was no async response).

Phase 2F is the first work to exercise all three paths in one orchestration:
inline workflow override → spawn child → child uses DB-defined workflow →
child awaits adapter response. Each defect blocks a different segment.

The right sequence to address them is probably:

1. **Cache invalidation (defect 1).** Fastest to work around (restart).
   Real fix is bounded — likely a single function in the chassis.
2. **Adapter partition routing (defect 2).** Blocks all async adapter
   responses. Likely affects multiple adapters; one fix may cover all.
3. **Consumer-group race (separate doc).** Largest scope: wiring change
   plus rollout planning. Held last.

None of these are caused by the Phase 2F migration. All three were already
present and surfaced because Phase 2F is the first work to exercise these
code paths together at runtime.

---

*Parked for deep discussion. Nothing in this document changes anything;
material for review.*
