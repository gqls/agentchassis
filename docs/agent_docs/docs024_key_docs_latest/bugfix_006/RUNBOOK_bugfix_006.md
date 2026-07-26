# RUNBOOK — bugfix_006

Every query and command this workstream had to get right, with its gotcha attached. When one
changes, change it **here**.

DB access throughout:
```bash
kubectl -n ai-persona-system exec -i postgres-clients-0 -- psql -U clients_user -d clients_db
```

---

## §B — contact forms

### The live form audit (which sites, which action)

```sql
SELECT s.domain, p.name,
       COALESCE(substring(pc.rendered_html from 'action="([^"]*)"'), '(none)') AS action
FROM page_components pc
JOIN pages p ON pc.page_id = p.id
JOIN sites s ON s.id = p.site_id
WHERE pc.rendered_html ~* 'data-component="contact-form"'
ORDER BY 1;
```

**Gotcha:** scope on `data-component="contact-form"`, **not** on `rendered_html ILIKE '%<form%'`.
The looser predicate returns 16 rows, six of which are working tool calculators — the first draft
of the discovery check made exactly that mistake and had to be narrowed against live data.

### Which address a form will heal to

```sql
SELECT domain, COALESCE(NULLIF(email,''), '(none)') FROM sites ORDER BY 1;
```

**Gotcha:** `RenderContext.Email` comes from the **top-level `sites.email` column**
(`loadSiteDataFull`, `COALESCE(si.email,'')`) — **not** `content_data.email` and **not** the
identity `site_spec`. Setting the wrong one changes nothing. Verified 2026-07-21 against a
council objection.

### Interpreting the result

A site still showing `#contact` is only a signal **if a discovery cycle has run there since the
fix went live**. Discovery is pipeline-triggered; the improvement-sweep that would drive it on a
cadence has been disabled fleet-wide since ~2026-05. Owner ruling 2026-07-25: leave the tail to
organic cycles, no batch runs.

---

## §C — the claim-timeout sweep

### Read the sweep as it actually is (never trust a doc's copy of it)

```bash
kubectl -n ai-persona-system exec -i postgres-clients-0 -- \
  psql -U clients_user -d clients_db -At \
  -c "SELECT pre_query FROM scheduled_tasks WHERE name='claimed-item-timeout';"
```

**Gotcha:** the column is `name`, **not** `task_name` — `\d scheduled_tasks` first, as always.

### Churn measurement (the number that says whether §C is biting)

```sql
SELECT item_type,
       count(*) FILTER (WHERE error LIKE 'Claim timed out%')  AS timed_out,
       count(*) FILTER (WHERE error LIKE 'Auto-completed%')   AS auto_completed
FROM site_work_items
WHERE updated_at > now() - interval '14 days'
GROUP BY 1 ORDER BY 2 DESC;
```

**Gotcha:** `updated_at` on this table was **not maintained** until migration `216` added a
trigger (2026-07-26, `bugs_open/035`). Figures from before that date under-count. `completed_at`
carries the historical truth for completions.

### Did the new generic branch fire? (the discriminating marker)

```sql
SELECT item_type, count(*) FROM site_work_items
WHERE error = 'Auto-completed: handler orchestration completed after claim'
GROUP BY 1 ORDER BY 2 DESC;
```

**Gotcha:** this string is deliberately **different** from the artifact branch's
`'Auto-completed: work verified done despite lost response'`. Counting the old string proves
nothing about the new branch — that is the whole point of the separate marker.

### Link a claimed item to its handler orchestration

```sql
SELECT wi.id, wi.item_type, now()-wi.claimed_at AS age, wi.handler_agent,
       o.owner_agent_type, o.status, o.current_step
FROM site_work_items wi
LEFT JOIN orchestration_states o
  ON o.initial_request_data->'input_data'->>'work_item_id' = wi.id::text
WHERE wi.status = 'claimed'
ORDER BY wi.claimed_at;
```

**Gotcha — the big one:** `orchestration_states` is purged at **~2 days**. Any coverage or
backtest measurement over a window longer than that will report a purge as a missing link. Check
retention before trusting a coverage figure:

```sql
SELECT date_trunc('day', created_at) AS day, count(*) FROM orchestration_states
GROUP BY 1 ORDER BY 1 DESC LIMIT 10;
```

### Who can even hold a claim

```sql
SELECT type FROM agent_definitions
WHERE is_active AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL
  AND default_config::text LIKE '%claim_work_item%';
```
Returns exactly two: `build-dispatch-loop` and `diagnose-dispatch-loop`. **Gotcha:** the second
does **not** use `status='claimed'` — it uses `diagnosing` and runs its own 75-minute reaper, so
it is outside this sweep entirely. Do not include it when reasoning about claim-timeout coverage.

---

## Applying a `pre_query` migration safely

A `pre_query` is executable SQL in a **text column**. Nothing validates it on write, the
scheduler runs it every 120 s, and a typo silently kills the fleet's only claim self-heal while
`enabled`, `last_triggered_at` and the logs all keep reading healthy. So:

### 1. Capture the current value first, byte-for-byte

```bash
kubectl -n ai-persona-system exec -i postgres-clients-0 -- psql -U clients_user -d clients_db -At \
  -c "SELECT pre_query FROM scheduled_tasks WHERE name='claimed-item-timeout';" > prequery_before.sql
md5sum prequery_before.sql
```
This becomes the `_ROLLBACK.sql` twin. **Never hand-edit an 84-line SQL string back by eye.**

### 2. Dry-run the whole migration with COMMIT swapped for ROLLBACK

```bash
perl -pe 's/^COMMIT;$/ROLLBACK;/' 220_claimed_item_timeout_generic_evidence.sql > /tmp/dryrun.sql
kubectl -n ai-persona-system exec -i postgres-clients-0 -- \
  psql -U clients_user -d clients_db -v ON_ERROR_STOP=1 < /tmp/dryrun.sql
```
Expect `BEGIN / UPDATE 1 / DO / DO / DO / ROLLBACK`. **Gotcha:** `-v ON_ERROR_STOP=1` is not
optional — without it psql reports the error and carries on, and a `DO` block that raised looks
like a pass in the tail of the output.

### 3. Fault-inject each guard and watch it fail

A guard that has never failed is not a guard. Induce each fault, confirm the diagnostic, restore.
The four for `220` are tabulated in `NOTES_bugfix_006.md`.

### 4. Re-check the row is unchanged since capture, then apply

```bash
diff prequery_before.sql <(kubectl ... -At -c "SELECT pre_query FROM ...")
```
**Gotcha:** several sessions work this cluster. Between capture and apply, another thread may have
edited the row; a rollback file generated from a stale capture would then restore the wrong thing.

### 5. Prove the SCHEDULER runs it, not just psql

```bash
kubectl -n ai-persona-system logs <kafka-scheduler-pod> --since=10m | grep claimed-item-timeout
```
Expect `"Pre-query task completed (no message fired)"`. **Gotcha:** `last_triggered_at` advancing
is weaker evidence than it looks — read the log line. And `"Skipping task — concurrency group at
max"` is normal, not a failure.

### 6. Prove the BEHAVIOUR live, with a negative control

Plant two probe work items at `status='claimed'`, `claimed_at = now()-20 minutes`, with an
`item_type` no artifact branch can match (`migration_probe`), plus two `orchestration_states`
rows carrying each item's id at `initial_request_data->'input_data'->>'work_item_id'` — one
`COMPLETED`, one `FAILED`. Wait one tick. Assert the first completed **with the new marker** and
the second did **not** move. Then delete all four rows and verify zero remain.

**Gotcha:** the probes are inert while they exist — a `claimed` item is not dispatchable (claiming
requires `triaged`/`approved`) and a `complete` one is terminal — but they are *live rows in the
production database*, so delete them in the same session that created them.

---

## Council gate — the traps this lane has already paid for

- `fixPlan.risks` must be a **string**, not an array.
- `operation` allowlist is `modify | add | remove | config_change`. **`create` is not valid** — a
  new file is `add`.
- An invalid run writes **no artifacts**, so polling `diagnosis_artifacts` waits forever. Poll
  `orchestration_states.current_step` too:
  ```sql
  SELECT status, current_step FROM orchestration_states
  WHERE collected_data->'input_data'->>'fix_correlation_id' = '<SUBMISSION_CORR>';
  ```
- **No row at all means QUEUED, not dropped.** Budget ~30 minutes. Do not resubmit — it costs a
  duplicate round.
- Quote fidelity in `grounded_in` is load-bearing: reviewers cannot open the file, so an
  abbreviated quote is a **different claim**. A trimmed SQL `WHERE` once manufactured a MEDIUM
  objection against byte-identical queries.
- The trailer `Council-Reviewed: <corr>` is earned by **APPROVED only** — read `decided_by`
  first. A verdict that post-dates its commit can never carry one; say so in the file.

---

## Testing Go in this repo

```bash
git archive HEAD | tar -x -C /tmp/cleantree
cp <your changed files> /tmp/cleantree/<same paths>
cd /tmp/cleantree && go test ./platform/orchestration/actions/discovery_checks/
```
**Gotcha:** the shared working tree frequently does not compile — it is not yours. This session hit
`undefined: hardcodedColourCandidate` from another session's in-flight edit. Never "fix" their
file to make your test run.

**Second gotcha:** `TestRegisteredVerifiersMatchClaimTimeoutExclusion` globs for the migration at
`../../../../docs/agent_docs/sql_for_agents/*_claimed_item_timeout_generic_evidence.sql`, so the
clean tree needs the migration copied in too, or the test `t.Fatal`s (by design — it refuses to
pass silently when it cannot find its other half).
