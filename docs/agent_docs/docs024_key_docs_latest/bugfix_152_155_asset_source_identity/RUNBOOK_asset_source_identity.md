# RUNBOOK — asset source identity (bugs 152 + 155)

Every command here was needed, with its gotcha attached. Change it HERE, not in
your scrollback.

## R1 — census `assets` by url form vs storage_path (the population query)

```sql
SELECT CASE WHEN url LIKE '%X-Amz%'     THEN 'presigned url'
            WHEN url LIKE '/assets/%'   THEN 'local web path'
            WHEN url LIKE 's3://%'      THEN 's3 uri'
            WHEN url LIKE 'http%'       THEN 'other http' ELSE 'other' END AS url_form,
       (storage_path IS NOT NULL) AS has_storage_path, count(*),
       count(*) FILTER (WHERE status='active') AS active
FROM assets GROUP BY 1,2 ORDER BY 3 DESC;
```

⚠ **Split by BOTH columns or the number lies.** "156 rows carry a local web path"
is the same figure whether their source is recoverable or gone — and that is the
entire difference between cosmetic staleness and unrecoverable data. The 107/49
split is the finding; the 156 is not.

## R2 — which sites can actually hit 155 (2+ active same-purpose assets)

```sql
SELECT s.domain, a.purpose, count(*) FROM assets a JOIN sites s ON s.id=a.site_id
WHERE a.status='active' GROUP BY 1,2 HAVING count(*)>1 ORDER BY 3 DESC;
```

## R3 — read the purpose cache (and the jsonb trap)

```sql
SELECT s.domain, k.key, left(s.content_data->>k.key, 70)
FROM sites s, LATERAL jsonb_object_keys(s.content_data) k(key)
WHERE jsonb_typeof(s.content_data)='object' AND k.key LIKE '%\_uri' ORDER BY 1,2;
```

⚠ **`jsonb_object_keys` ERRORS on a jsonb ARRAY**, and at least one site's
`content_data` is one — without the `jsonb_typeof(...)='object'` guard the whole
query aborts and you get no rows for any site, which reads exactly like "no site
has this key". The `\_` escape matters too: bare `_` is a single-char wildcard.

## R4 — prove a `_uri`-style key has no live agent reader (with its positive control)

```sql
-- the claim:
SELECT type FROM agent_definitions
WHERE is_active AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL
  AND default_config::text ~ '(hero|logo|icon|content_hero)_uri';        -- expect 0
-- the control that makes the 0 mean something:
SELECT type, (SELECT string_agg(DISTINCT m[1], ', ')
              FROM regexp_matches(default_config::text,'([a-z_0-9]+_uri)','g') m)
FROM agent_definitions WHERE is_active AND COALESCE(is_snapshot,false)=false
  AND deleted_at IS NULL AND default_config::text LIKE '%\_uri%' GROUP BY 1,default_config;
```

Run both in the same session. A 0 from the first with an empty second means your
regex is broken, not that the key is unread.

## R5 — pod-grep for THIS change, with the negative control that does the work

```bash
kubectl get pods -n ai-persona-system -l app=agent-chassis \
  -o custom-columns=NAME:.metadata.name,IMAGE:.spec.containers[0].image,START:.status.startTime --no-headers
for p in <each pod>; do kubectl exec -n ai-persona-system $p -- sh -c \
  'strings /app/agent-chassis | grep -c "AssetSourceRef";
   strings /app/agent-chassis | grep -c "Resolved s3_uri from site content_data via asset_id"'; done
```

**Pre-roll the pair reads `0` / `1`; post-roll it must read `≥1` / `0`.** The second
string is the deleted branch's own log line, so one command answers both "did my
code ship" and "is the defect still present" — and the second arm is the one that
cannot be faked by a stale image (`bugs_open/153`). Compare the pod START time
against your commit time first: a build that predates your commit cannot contain it,
however fresh the roll.

## R6 — apply the backfill (roll-independent)

```bash
kubectl -n ai-persona-system exec -i postgres-clients-0 -- \
  psql -U clients_user -d clients_db -v ON_ERROR_STOP=1 < docs/agent_docs/sql_for_agents/323_backfill_assets_storage_path.sql
```

⚠ The verify block is `DO`/`RAISE`, not a bare `SELECT`: `ON_ERROR_STOP` does not
fire on a non-empty result set, so a verification made of `SELECT`s cannot stop the
`COMMIT`. Expect `BEGIN / UPDATE 205 / DO / COMMIT` — a silent `DO` is the pass.

## R7 — closure test for 155 (the only one that proves it)

On a site with 2+ active same-purpose assets, deploy each by `asset_id` alone (NO
`s3_uri` in spec) and `sha256sum` the downloaded files: they must DIFFER, and at
least one must be opened and looked at against its own `origin_prompt`. dartsonline's
6 founding icons are the natural re-run. **`success:true` and distinct destination
paths were both already true while the bug was shipping identical bytes** — neither
is evidence.

## R8 — before claiming a code path is fixed, prove it is REACHABLE

```sql
SELECT a.type, s.key AS step,
       (s.value->'config'->'input_fields')::text AS input_fields,
       (s.value->'config'->>'<the_field_you_rely_on>') AS explicit_path_cfg
FROM agent_definitions a, LATERAL jsonb_each(a.default_config->'workflow'->'steps') s
WHERE a.is_active AND COALESCE(a.is_snapshot,false)=false AND a.deleted_at IS NULL
  AND s.value->>'action' = '<the_action>'
ORDER BY 1,2;
```

⚠ **A field being Optional in the Go `ActionInputSpec` does NOT mean callers pass
it.** `ExtractActionInputs` Strategy 1 wins whenever the step config has
`input_fields`, and it extracts **only** the names listed there — the recursive
all-fields hunt (Strategy 2) is reached only when `input_fields` is ABSENT
(`action_inputs.go:441-467`). So a step with `input_fields` silently drops every
spec field it does not name, however present the value is in `collected_data`.

This is how I proved a branch I had just "fixed" was unreachable through the very
agent its bug was filed against: `asset-deployer` has never listed `asset_id`. Run
this BEFORE writing that a defect is closed — an unreachable branch is neither
the cause of your symptom nor evidence of your fix.
