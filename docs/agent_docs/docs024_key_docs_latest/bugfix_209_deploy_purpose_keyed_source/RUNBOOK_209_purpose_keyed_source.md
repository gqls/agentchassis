# RUNBOOK — 209 purpose-keyed source

Every command that was hard to get right, with its gotcha attached. Fix them
here, not in scrollback.

---

## 1. Which live definitions carry the action (a DB fact, never a repo fact)

A Go call-site count is not the caller count — 221's lane was corrected by the
council for exactly that. Ask the database:

```sql
SELECT type FROM agent_definitions
WHERE is_active AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL
  AND default_config::text LIKE '%deploy_image_asset%'
ORDER BY type;
```

⚠ A bare token (`deploy_image_asset`) is safe in `::text LIKE`. A **key/value**
pair is not — jsonb renders a space after the colon, so `'%"k":"v"%'` matches
nothing. Induce a non-zero result before trusting a zero.

## 2. Enumerate a definition's steps (steps is an OBJECT, not an array)

```sql
SELECT ad.type, s.key AS step, s.value->>'action' AS action,
       COALESCE(s.value->'config'->>'purpose','-')            AS cfg_purpose,
       COALESCE((s.value->'config'->'input_fields')::text,'-') AS input_fields
FROM agent_definitions ad,
     LATERAL jsonb_each(ad.default_config->'workflow'->'steps') s
WHERE ad.is_active AND COALESCE(ad.is_snapshot,false)=false AND ad.deleted_at IS NULL
  AND ad.type = '<type>'
ORDER BY s.key;
```

Use `jsonb_pretty(s.value)` for the full step when you need `next_step` /
`error_step` / conditional routing — branch exclusivity is invisible in a flat
listing, and that exclusivity is what decided this bug.

## 3. Read a run's `collected_data` shape — enumerate keys, never probe a path

A path read cannot see a shape change underneath it. Ask which keys exist:

```sql
SELECT k, count(*) AS n_runs
FROM orchestration_states o, LATERAL jsonb_object_keys(o.collected_data) k
WHERE o.owner_agent_type = 'asset-deployer'
GROUP BY k ORDER BY n_runs DESC, k;
```

⚠ The agent-type column is **`owner_agent_type`**, not `persona_id`.

## 4. Before reading anything into an ABSENCE in `orchestration_states`

The table looks like it holds a month (oldest row 2026-07-13) — it does not.
Bucket by status first:

```sql
SELECT CASE WHEN created_at > now() - interval '24 hours' THEN 'last_24h'
            WHEN created_at > now() - interval '7 days'  THEN '1-7d'
            ELSE 'older_than_7d' END AS age,
       status, count(*)
FROM orchestration_states GROUP BY 1,2 ORDER BY 1,3 DESC;
```

Measured 2026-08-08: **13** `COMPLETED` rows older than 24h, **0** older than 7
days; the old tail is `CANCELLED`/`RUNNING`/`INITIALIZED`. So for completed runs
this is a **~24-hour window** — "no rows" means "not today", nothing more.

## 5. `llm_call_log` is BLIND to these agent types — check the control first

It retains from 2026-03-25 (50,861 rows), which makes it look like the right
place to ask "has this agent ever run". It returned **0 rows for
`asset-deployer`** on a day it ran 16 times. Always run the positive control:

```sql
SELECT agent_type, count(*), max(created_at)::date
FROM llm_call_log
WHERE agent_type IN ('<the type you care about>', '<a type you KNOW ran today>')
GROUP BY 1;
```

If the known-good type returns 0, the measurement is blind — discard it, do not
report it as an absence.

## 6. Run the 209 evidence harness

```bash
go test ./platform/orchestration/actions/ \
  -run 'TestFindStorageURI_|TestExtractActionInputs_LegacyShape|TestDeployImageAsset_LegacyShape' -v
```

The `AssetIDIsNotStable` test runs the real helper 400× and prints the split. It
is **not** a flake if the numbers move — Go map order is randomised per run, so
the ratio varies (344/400 hero-wins when first measured). It *would* be a real
finding if the split ever collapsed to a single value: that means resolution
became deterministic and the hazard changed.

## 7. Ownership re-check (the `who-owns.py` false positive)

```bash
./scripts/who-owns.py 209
```

⚠ It reads **commits**, so a lane that merely *cites* a bug in a handoff reads as
owning it, and a session mid-fix is invisible. Sweep live transcripts too:

```bash
ls -t ~/.claude/projects/-home-ant-projects-agentchassis/*.jsonl | head -16 | while read f; do
  c=$(grep -c "findStorageURI\|209_HANDOFF" "$f" 2>/dev/null); [ "$c" != "0" ] && echo "$c $(basename $f)"; done
```

Then check each hit's *last* entries for what it is actually doing — most hits
are incidental (a `git status` listing the filename, or a neighbouring lane
reading the same package).
