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

> **CORRECTED 2026-08-09 — the query that used to be here UNDERCOUNTED BY 32%, and
> read as complete while doing it.** It walked `->'workflow'->'steps'` and nothing else,
> so it could not see a step nested in a loop's `sub_workflow.steps`. It reported 13
> carriers of the three old key names; there were **19**. The six it missed
> (component-quality-auditor, internal-linker, tool-auditor ×2, tool-suggester ×2) are
> exactly the shape `validation.WalkSteps` was extracted to abolish — `bugs_open/144`:
> two hand-written traversals blind in the same direction, agreeing with each other.
> **The lane's own handoff table was built from the blind version and is wrong.** The
> AUDIT was never blind (`cmd/config-key-audit` walks `WalkSteps` at every call site),
> which is why the acceptance number stayed trustworthy while the census did not.

**Ask at ALL depths.** A text-level scan cannot be fooled by nesting, and it is the
cheapest honest census — carry a positive control so a zero is readable:

```sql
SELECT
  count(*) FILTER (WHERE default_config::text ~ '(item|check|target)_domain')   AS old_spelling,
  count(*) FILTER (WHERE default_config::text ~ '(item|check|target)_pipeline') AS new_spelling,  -- positive control
  count(*) FILTER (WHERE default_config::text ~ 'zzz_invented_control')         AS neg_control
FROM agent_definitions
WHERE deleted_at IS NULL AND COALESCE(is_snapshot,false)=false AND is_active;
```

For the exact paths (which is what you need to WRITE a fix), recurse — one recursive
term only, Postgres refuses two references to the CTE:

```sql
WITH RECURSIVE walk(agent_type, path, node) AS (
  SELECT ad.type, ARRAY[]::text[], ad.default_config
    FROM agent_definitions ad
   WHERE ad.deleted_at IS NULL AND COALESCE(ad.is_snapshot,false)=false AND ad.is_active
     AND ad.default_config::text ~ '(item|check|target)_domain'
  UNION ALL
  SELECT w.agent_type, w.path || e.k, e.v
    FROM walk w CROSS JOIN LATERAL jsonb_each(
      CASE jsonb_typeof(w.node)
        WHEN 'object' THEN w.node
        WHEN 'array'  THEN COALESCE((SELECT jsonb_object_agg((i-1)::text, v)
                                       FROM jsonb_array_elements(w.node) WITH ORDINALITY a(v,i)), '{}'::jsonb)
        ELSE '{}'::jsonb END) AS e(k,v)
)
SELECT agent_type, array_to_string(path,' > ') AS json_path, node #>> '{}' AS value
FROM walk
WHERE path[array_length(path,1)] IN ('item_domain','check_domain','target_domain')
ORDER BY 1,2;
```

**Gotcha, and it is the tell for this whole bug class:** a key with **zero** live steps
that the code reads, beside an old name with several, is a rename that landed on one side
only. Query for **both** names in one statement — asking only about the new one returns an
empty result that reads like "nothing to see".

**Gotcha 2:** all three of `deleted_at IS NULL`, `COALESCE(is_snapshot,false)=false` and
`is_active` are load-bearing. Snapshots carry old configs and will give you counts that
describe history.

**Gotcha 3 (the 08-09 one):** `jsonb_each(default_config->'workflow'->'steps')` is a
**top-level-only** descent, and there is no error, no warning and no empty result to tell
you so — it returns a confident, plausible, incomplete number. Any census of live step
config must either recurse or scan the text. If you are writing Go rather than SQL, call
`validation.WalkSteps` and inherit the fix instead of re-deriving the bug.

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

> **CORRECTED 2026-08-08 after the council's `debug_historian` seat objected — the first
> version of this recipe used `-l app=agent-chassis` and a positive-only grep. Measured:
> that selector returns 2 pods, and 25 RUNNING pods carry an agent-chassis image (34
> including non-Running). It was verifying 8% of the surface.** Enumerate by IMAGE.

```bash
PODS=$(kubectl -n ai-persona-system get pods -o json | python3 -c "
import json,sys
d=json.load(sys.stdin)
print(' '.join(p['metadata']['name'] for p in d['items']
      if p.get('status',{}).get('phase')=='Running'
      and any('agent-chassis' in c.get('image','') for c in p['spec'].get('containers',[]))))")
echo "pods carrying the image: $(echo $PODS | wc -w)"
for P in $PODS; do
  kubectl -n ai-persona-system exec "$P" -- sh -c "
    echo -n '$P new=';         strings /app/agent-chassis | grep -c 'config setting arrived via a deprecated alias'
    echo -n '$P pos_control='; strings /app/agent-chassis | grep -c 'Using deprecated config pattern'
    echo -n '$P neg_control='; strings /app/agent-chassis | grep -c 'zzz_invented_control_string_136'"
done
kubectl -n ai-persona-system logs -l app=agent-chassis --tail=-1 --since=24h | grep "deprecated alias"
```

**Baseline measured pre-roll 2026-08-08:** `new=0 pos_control=1 neg_control=0` on every pod
sampled. After the roll `new` must be 1 everywhere.

**Gotcha:** the **positive control is the load-bearing one** and a new-string-only check
cannot supply it. `Using deprecated config pattern` is Strategy 3's long-live warn: it proves
the grep works *and* that you are reading the binary you think you are. A `0` from a
new-string grep is ambiguous between "not shipped" and "I mistyped it / exec'd the wrong
container" until a control disambiguates.
**Gotcha 2:** before trusting any marker, confirm the literal is new —
`grep -rc "<literal>" --include=*.go .` must return exactly your own file. `093`'s marker was
vacuous because the new code deliberately mirrored existing wording and greped `1` before
anything shipped. **But source-uniqueness is not binary-presence**: it tells you the string
would be new, not that it is spelled right in the image. That is what the two controls are for.
**Gotcha 3:** `logs deploy/…` reads one pod of N; `-l` is right for logs and wrong for
establishing coverage.

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

## Witnessing an alias at runtime (the recipe that finally worked, 2026-08-08)

Log-sweeping CANNOT witness it — an active chassis pod retains **<1s** of log (measured;
see the LANDMINES entry "chassis pod's retrievable log holds less than a second"), and
`agent-job-cleanup` deletes Completed carrier pods within minutes. The witness must be a
**DB row whose value could come out otherwise**.

Recipe: a throwaway agent definition whose single `create_work_item` step config carries
ONLY the deprecated key with a NON-default value, filing a born-`cancelled` item on
`system.internal`:

```sql
-- the step config that discriminates (item_pipeline absent, default is "build"):
--   "item_domain": "content", "status": "cancelled", "item_type": "alias_witness_136",
--   "handler_agent": "none", "site_id": "input_data.site_id"
```

then one orchestrate publish to `system.agent.generic.requests` (payload in the container
COMMAND with a `PUBLISH_OK` marker — the kcat stdin trap), then read the row:

```sql
SELECT pipeline, status FROM site_work_items WHERE item_type='alias_witness_136';
-- content → alias honoured; build → alias fell through to the default
```

Working scripts (worked first time, ~3 min dispatch→row on a quiet lane):
`witness_136_fire.sh` / `witness_136_poll.sh`, preserved in the session scratchpad and
reproduced in full in `bugs_open/136` §11's commit. Clean up: `UPDATE agent_definitions
SET is_active=false, deleted_at=now() WHERE type='alias-witness-136'` — the row stays,
it IS the evidence.

**Gotcha 1:** the filed row must be born `cancelled`, NOT `detected` — `triage_detect_items`
REWRITES `detected` rows to `pipeline='build'` with no pipeline filter, which would destroy
the witness value in a way that reads exactly like the alias failing.
**Gotcha 2:** the witness value must differ from the hardcoded default. All nine live
carriers set `"build"` = the default, which is why no live execution can ever witness this
(§9's wrong call).
**Gotcha 3:** find the run by payload, not the printed orchestration id:
`WHERE collected_data->'input_data'->>'witness_marker' = '<uuid you sent>'`.
