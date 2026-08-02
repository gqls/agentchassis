# RUNBOOK — bugs_open/091, work-item conflict refresh

Every command here had a gotcha attached. Change it HERE when it changes.

---

## Is the durable record still wrong? (the measurement that sized this bug)

Compares what each open `stale_evidence` item SAYS drifted against what the last
`evidence-freshness` run actually FOUND. Four of five disagreed on 2026-08-02.

```sql
WITH live AS (
  SELECT r->>'domain' AS domain,
         (SELECT string_agg(f->>'fact_id', ', ' ORDER BY f->>'fact_id')
            FROM jsonb_array_elements(CASE WHEN jsonb_typeof(r->'facts')='array'
                                           THEN r->'facts' ELSE '[]'::jsonb END) f
           WHERE f->>'outcome'='drifted') AS live_drifted
  FROM orchestration_states os,
       LATERAL jsonb_array_elements(os.collected_data#>'{complete,result,refresh_result,results}') r
  WHERE os.owner_agent_type='evidence-freshness'
), item AS (
  SELECT s.domain, swi.created_at,
         (SELECT string_agg(d->>'fact_id', ', ' ORDER BY d->>'fact_id')
            FROM jsonb_array_elements(CASE WHEN jsonb_typeof(swi.spec->'drifted')='array'
                                           THEN swi.spec->'drifted' ELSE '[]'::jsonb END) d) AS item_drifted
  FROM site_work_items swi JOIN sites s ON s.id=swi.site_id
  WHERE swi.item_type='stale_evidence'
)
SELECT COALESCE(l.domain,i.domain) AS domain, i.created_at::date AS item_filed,
       i.item_drifted, l.live_drifted,
       (COALESCE(i.item_drifted,'') IS DISTINCT FROM COALESCE(l.live_drifted,'')) AS record_is_wrong
FROM live l FULL JOIN item i ON i.domain=l.domain
WHERE i.domain IS NOT NULL OR l.live_drifted IS NOT NULL
ORDER BY 1;
```

**Three gotchas, all of which cost a query today:**

1. **`orchestration_states` is retention-clocked** — terminal rows last ~24h. Exactly
   ONE `evidence-freshness` run is retained at any moment. Run this the same day or
   `live` comes back empty and the whole comparison silently reads "no disagreement".
2. **`jsonb_array_elements` on a scalar aborts the whole statement** —
   `ERROR: cannot extract elements from a scalar`. Some rows carry `facts` as
   something other than an array. The `jsonb_typeof(...)='array'` guard is not
   defensive noise; without it you get no output at all, which reads like "no rows".
3. **There is no `agent_type` column** on `orchestration_states` — it is
   `owner_agent_type`. `\d orchestration_states` first, as the platform conventions say.

## Per-site outcome of the last run (drifted vs work_item_created)

```sql
SELECT r->>'domain' AS domain, r->>'drifted' AS drifted, r->>'work_item_created' AS work_item_created
FROM orchestration_states os,
     LATERAL jsonb_array_elements(os.collected_data#>'{complete,result,refresh_result,results}') r
WHERE os.owner_agent_type='evidence-freshness'
ORDER BY 1;
```

On 2026-08-02: four sites drifted (2, 5, 12, 1 facts) and **every one** reported
`work_item_created=false`. That is candidate 2 working — the report is honest — and it
is also the direct measurement of how much is being dropped.

## The dedup index the ON CONFLICT clause must keep matching

```sql
SELECT indexdef FROM pg_indexes WHERE indexname='idx_swi_dedup';
```

```
CREATE UNIQUE INDEX idx_swi_dedup ON public.site_work_items USING btree (site_id, item_key)
  WHERE ((item_key IS NOT NULL) AND (status <> ALL (ARRAY[
    'complete','verified','rejected','wont_fix','failed','unresolved','cancelled'])))
```

**This index and the Go slice `workItemTerminalStatuses` are ONE contract.** Drift
between them is a fleet-wide `42P10` ("no unique or exclusion constraint matching the
ON CONFLICT specification") — every work-item write fails, everywhere. The refresh
`UPDATE` added by this fix builds its predicate from the same slice for the same
reason; do not inline a status list anywhere.

## Live status vocabulary (needed to pick the "in-flight" guard)

```sql
SELECT status, count(*) FROM site_work_items GROUP BY 1 ORDER BY 2 DESC;
```

2026-08-02: `complete` 1955, `needs_human_review` 368, `unresolved` 235, `detected`
204, `cancelled` 93, `deferred` 30, `failed` 23, `verified` 9, `wont_fix` 7,
`rejected` 7, `triaged` 3, `blocked` 2, `claimed` 1, `diagnosing` 1.

`claimed` and `diagnosing` are the two a handler actually holds — the refresh skips
them. Note `approved` appears in dispatch filters but has zero live rows; it is a
queued state, not a held one, so it is refreshable.

## Inducing the fault (the acceptance test — a clean sweep proves NOTHING)

```sql
-- pick a site that already has an OPEN stale_evidence item
SELECT s.domain, swi.id, swi.spec->'drifted' FROM site_work_items swi
  JOIN sites s ON s.id=swi.site_id WHERE swi.item_type='stale_evidence';

-- corrupt one sql-sourced fact so it MUST drift, on that site
UPDATE site_specs SET data = jsonb_set(data, '{facts,<idx>,value}', '<wrong>'::jsonb)
 WHERE id = '<current spec row>';

-- re-arm the sweep
UPDATE scheduled_tasks SET last_triggered_at = NULL WHERE name = 'evidence-freshness';
```

Then require BOTH, or it is not verified:

* the open item's `spec->'drifted'` now names the NEWLY drifted fact, and
* the run's `work_item_created` reads **false** while `work_item_refreshed` reads true.

Restore afterwards: the sweep re-syncs the register value itself, so only the work
item needs cleaning up.

## Pod-grep markers after a roll

Use a string this change CREATES, plus a positive control that must already be 1, plus
a negative control that must be 0 (`bugs_open/153` — a roll is not evidence your fix
shipped, and the image may predate your commit).

```bash
POD=$(kubectl -n ai-persona-system get pods -l app=agent-chassis -o jsonpath='{.items[0].metadata.name}')
# NEW — 0 before the roll, 1+ after
kubectl -n ai-persona-system exec "$POD" -- sh -c "strings /app/agent-chassis | grep -c 'refreshed the open work item'"
# POSITIVE CONTROL — candidate 2's wording, live since v1.0.1177, must stay 1
kubectl -n ai-persona-system exec "$POD" -- sh -c "strings /app/agent-chassis | grep -c 'an open stale_evidence item already'"
# NEGATIVE CONTROL — a string this change REMOVED, must reach 0
kubectl -n ai-persona-system exec "$POD" -- sh -c \
  "strings /app/agent-chassis | grep -c \"holds this site.s key, and its spec still describes\""
```

> **CORRECTED 2026-08-03, before anyone used it.** This block first named
> `'create work item: '` as the negative control. **It is not one** — that string is
> an `fmt.Errorf` wrapper in `applyNewPage` which this change deliberately KEEPS, so
> it greps 1 before and 1 after and would have read as "the fix did not ship" on a
> perfectly good image. Caught by grepping the post-change source for the control
> itself, which is the check: **a negative control has to be verified ABSENT from
> your own working tree before you trust it against a pod.** The control above is
> verified absent (the sentence was rewritten in `refresh_evidence_base_action.go`)
> and measured present on the pre-roll image.

Run it on **every** replica, not `deploy/agent-chassis` (that reads one pod of N).

## Building and testing the way the tree is actually built

`make build-<service>` builds from committed HEAD. So test what HEAD will compile,
not the dirty tree — other sessions' WIP is in the tree and not in the image:

```bash
rm -rf /tmp/claude-*/head-check && mkdir -p /tmp/claude-*/head-check
git archive HEAD | tar -x -C <that dir>
cd <that dir> && go build ./... && go test ./platform/orchestration/actions/ -run WorkItem
```

## Prove the SQL before the roll — `go build` cannot parse it

Both new statements, PREPAREd against the LIVE schema inside a transaction that
ends in ROLLBACK (real schema, nothing written, nothing to clean up). Run this
after any edit to either statement — a typo'd column or a predicate that no longer
matches the partial index is invisible to the compiler and fatal at runtime.

```sql
BEGIN;
PREPARE swi_refresh (uuid, text, text, jsonb) AS
UPDATE site_work_items
   SET summary = $3, spec = $4::jsonb, updated_at = NOW()
 WHERE site_id = $1 AND item_key = $2
   AND status NOT IN ('complete','failed','verified','rejected','wont_fix','unresolved','cancelled')
   AND status NOT IN ('claimed','diagnosing');

PREPARE swi_insert_parent (uuid,text,text,text,text,text,text,uuid,uuid,int,text,text,text,text,uuid,uuid[],uuid) AS
INSERT INTO site_work_items (
    site_id, source, pipeline, item_type, severity, summary, spec,
    page_id, component_id, priority, handler_agent, status, created_by,
    item_key, batch_id, depends_on, parent_item_id
) VALUES ($1,$2,$3,$4,$5,$6,$7::jsonb,$8,$9,$10,$11,$12,$13,$14,$15,$16::uuid[],$17)
ON CONFLICT (site_id, item_key)
    WHERE item_key IS NOT NULL
      AND status NOT IN ('complete','failed','verified','rejected','wont_fix','unresolved','cancelled')
DO NOTHING;
ROLLBACK;
```

**Both returned `PREPARE` on 2026-08-03.** The second one is the load-bearing
check and it answers the pre-commit hook's `unpaired-change` warning
("touches `workItemTerminalStatuses` but not `idx_swi_dedup`"): adding
`parent_item_id` to the column list does **not** disturb partial-index inference —
the `ON CONFLICT … WHERE` still resolves to `idx_swi_dedup`, so there is no
SQLSTATE 42P10. The warning is correct to fire (that pairing has bitten the fleet
before) and is a false positive here, because the terminal list is only READ: the
new UPDATE interpolates it into a plain `WHERE`, which touches no index predicate
at all.

## Baseline pod-grep, taken BEFORE the roll (this is what makes the after-grep mean something)

Measured 2026-08-03 on both replicas, image `v1.0.1234`:

```
NEW  'refreshed the open work item'                    -> 0   (change is NOT live)
NEG  "holds this site's key, and its spec still describes" -> 1   (the string this change REMOVES)
POS  'an open stale_evidence item already'             -> 1   (control: grep + binary path work)
```

After the roll, require **NEW ≥ 1 and NEG = 0 on every replica.** NEW alone is not
enough — an image can carry a new string and still predate other parts of your
commit; the negative control is what proves the OLD code is gone rather than
merely accompanied. (`bugs_open/153`: a roll is not evidence your fix shipped.)
