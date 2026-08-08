# RUNBOOK — `bugfix_136_config_key_aliases`

Every command here had a gotcha attached when I first got it right. The gotcha is the point.

## Is this bug still live? (no roll needed, reads the production DB)

```bash
./scripts/audit-config-keys.sh        # ~40s: go run + a kubectl psql export
```
Read **UNKNOWN KEYS** and the new **DEPRECATED KEYS** section together. As of 2026-08-08
after the fix: UNKNOWN = `plan_sections: domain` only; DEPRECATED names the three renames.

**Gotcha:** exit code is deliberately **unaffected** by the DEPRECATED section. Those keys
are wired; it is a migration list, not a defect list. Do not write CI that treats it as one.

## Which live steps carry a key, and which carry the one the code reads

```sql
SELECT ck.key, count(*) AS live_steps, string_agg(DISTINCT ad.type, ', ') AS agents
FROM agent_definitions ad, jsonb_each(ad.default_config->'workflow'->'steps') AS e(k,v),
     jsonb_object_keys(e.v->'config') AS ck(key)
WHERE ad.deleted_at IS NULL AND COALESCE(ad.is_snapshot,false)=false AND ad.is_active
  AND jsonb_typeof(e.v->'config')='object'
  AND ck.key IN ('check_pipeline','check_domain','target_pipeline','target_domain',
                 'item_pipeline','item_domain')
GROUP BY 1 ORDER BY 1;
```

**Gotcha, and it is the tell for this whole bug class:** a key with **zero** live steps
that the code reads, beside an old name with several, is a rename that landed on one side
only. Query for **both** names in one statement — asking only about the new one returns an
empty result that reads like "nothing to see".

**Gotcha 2:** all three of `deleted_at IS NULL`, `COALESCE(is_snapshot,false)=false` and
`is_active` are load-bearing. Snapshots carry old configs and will give you counts that
describe history.

## Does an action actually read a key? (the check I got wrong)

```bash
grep -n '"priority"' platform/orchestration/actions/create_work_item_action.go
```

**Gotcha:** grep the **key name**, never the access pattern. `grep 'config\["'` misses every
key that arrives through a helper — `GetIntField`, `GetBoolField`, `resolveAIServiceConfig` —
and those are precisely where the non-obvious reads are, because a key simple enough to read
inline is a key nobody wrapped. This mistake is logged in `WRONG_CALLS.md` 2026-08-08.

## Which alias field does my key belong in?

- Value is a **dot-path into `collected_data`**, action reads it via `inputs.Get(...)`
  → `Deprecated`.
- Value **IS the value**, action reads it via `config["k"].(string)`
  → `DeprecatedConfigKeys`, honoured by `datahelpers.ResolveConfigSetting`.

**Gotcha:** putting a setting in `Deprecated` is worse than declaring nothing — Strategy 3
resolves the value as a path, finds nothing, takes the default, and `UnknownConfigKeys`
recognises the key, so the detector goes quiet too. See the LANDMINE.

## Which discovery checks propagate `dctx.Pipeline` (the blast-radius query)

```bash
grep -rln "dctx.Pipeline" platform/orchestration/actions/discovery_checks/*.go | grep -v _test
```
then intersect with each agent's live `checks` array:
```sql
SELECT ad.type, e.v->'config'->'checks'
FROM agent_definitions ad, jsonb_each(ad.default_config->'workflow'->'steps') AS e(k,v)
WHERE ad.is_active AND COALESCE(ad.is_snapshot,false)=false AND ad.deleted_at IS NULL
  AND e.v->>'action'='run_discovery_checks';
```

**Gotcha:** most checks hardcode their own `Pipeline:` in the `WorkItemSpec` they return, so
the population that cares about `check_pipeline` is **only** the ones that propagate
`dctx.Pipeline`. Counting all checks overstates the blast radius by about 5×. This
intersection is not stable — it changed between 2026-07-28 and 2026-08-04 and that change is
what turned this bug from latent to live. **Re-run it; do not cite the last answer.**

## Proving the fix after the roll

```bash
POD=$(kubectl -n ai-persona-system get pods -l app=agent-chassis -o jsonpath='{.items[0].metadata.name}')
# discriminating literal — verified to exist nowhere else in the tree before relying on it
kubectl -n ai-persona-system exec "$POD" -- sh -c "strings /app/agent-chassis | grep -c 'config setting arrived via a deprecated alias'"
kubectl -n ai-persona-system exec "$POD" -- sh -c "strings /app/agent-chassis | grep -c 'ResolveConfigSetting'"   # positive control
kubectl -n ai-persona-system logs -l app=agent-chassis --tail=-1 --since=24h | grep "deprecated alias"
```

**Gotcha:** `-l app=agent-chassis`, never `logs deploy/…`, which reads one pod of N.
**Gotcha 2:** before trusting any pod-grep marker, confirm the literal is new —
`grep -rc "<literal>" --include=*.go .` must return exactly your own file. `093`'s marker was
vacuous because the new code deliberately mirrored existing wording and greped `1` before
anything shipped.

## Testing on a shared tree

```bash
rm -rf /tmp/x && mkdir -p /tmp/x && git archive HEAD | tar -x -C /tmp/x
# copy ONLY your files over it, then:
cd /tmp/x && go build ./... && go test ./platform/orchestration/... ./cmd/...
```

**Gotcha:** this is how you test *the commit you are about to make* rather than *your tree*,
which contains every other session's WIP. **And match the instrument**: `go test` runs vet,
`go build` does not, so a package that builds at HEAD can still fail `go test ./...` for
reasons that have nothing to do with you. Compare vet with vet.

## Council + landmine bookkeeping

```bash
./docs/agent_docs/docs024_key_docs_latest/fixloop_eg_dartsonline/097_TRIGGER_council_review_v1.sh <submission.json>
./scripts/landmines-sync.py --apply && ./scripts/landmines-sync.py --check
```
Verdict, keyed on the **submission correlation** (not the printed run id):
```sql
SELECT created_at, metadata->>'decision' FROM diagnosis_artifacts
WHERE correlation_id='433de2c0-682f-4d8d-8c48-28637309f1ba' AND kind='council_report'
ORDER BY created_at;
```
**Gotcha:** budget ~30 minutes, not ~2 — the council itself takes 2–5 but the dispatch queues
behind the fleet. A missing row is latency, not a dropped dispatch; retrying costs a
duplicate round.
