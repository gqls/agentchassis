# RUNBOOK — bugs_open/179 finding A

Every command that had to be got right, with its gotcha attached.

---

## The census — match the JSON shape, never the bare word

**Gotcha, and the bug file records it because it cost a wrong conclusion:**
`collected_data::text LIKE '%deploy_path%'` returns 93 orchestrations and 2 agent
definitions. **None is a value.** They are the declaration, the step descriptions,
and this lane's own council submissions (a council run stores the submission JSON,
whose rationale argues about `deploy_path` at length).

```sql
-- VALUES, the only population that matters. Expect 0 / 0 / 0.
SELECT count(*) FROM site_work_items
 WHERE spec::text LIKE '%"deploy_path":"%' AND status NOT IN ('complete','cancelled','rejected');
SELECT count(*) FROM agent_definitions
 WHERE is_active AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL
   AND default_config::text LIKE '%"deploy_path":"%';
SELECT count(*) FROM orchestration_states WHERE collected_data::text LIKE '%"deploy_path":"%';
```

**Second gotcha: a value census cannot see this defect's real exposure.**
`ExtractActionInputs` hunts every *declared* field through the whole of
`collected_data` with a depth-20 recursive search, so while the field is declared,
a stray key nobody set deliberately is still bound. The population that mattered
was therefore the **declaration**, not the values:

```sql
SELECT type,
       default_config->'workflow'->'steps'->'deploy_asset'->'config'->'input_fields' AS input_fields,
       input_contract->'optional' AS contract_optional
  FROM agent_definitions
 WHERE is_active AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL
   AND (default_config::text LIKE '%deploy_path%' OR input_contract::text LIKE '%deploy_path%');
```

**Third gotcha:** a bare-word hit on a *second* agent is not a second caller. Read
the actual step — `image-build-handler`'s hit was its step **description**; its
`input_mapping` passes no `deploy_path`:

```sql
SELECT jsonb_pretty(default_config #> '{workflow,steps,call_asset_deployer}')
  FROM agent_definitions WHERE type='image-build-handler' AND is_active
    AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;
```

## The source census that made the class ban cheap

```bash
grep -rn "AssetPaths{" --include="*.go" platform/ internal/ pkg/ | grep -v _test | grep -v "platform/storage/"
# exactly ONE hit before the change (the override); empty after.
```

## Running the tests, and the mutations

```bash
export TMPDIR=/home/ant/.cache/buildtmp      # /tmp is a small tmpfs on this box
go build ./...
go test ./platform/storage/ ./platform/orchestration/actions/ -run 'DeployImageAsset|AssetPaths|BrandHead' -v
```

**Gotcha — a source-scanning test makes your COMMENTS load-bearing.** The ordering
assertions use `strings.Index`, i.e. the **first** occurrence of each token. Naming
`storage.DeployedAssetPath(` in a doc comment *above* the guard fails the test on
ordering alone, and the failure looks like a code-ordering bug. Two of my own
comments tripped this. Both the action's doc comment and the test now say so.

**Mutation recipe** (the tree is shared — back up, mutate, test, restore in one
command so the window is seconds, then assert residue-free):

```bash
SC=<scratchpad>; A=platform/orchestration/actions/deploy_image_asset_action.go
cp $A $SC/A.bak
# … mutate …
go test ./platform/orchestration/actions/ ./platform/storage/ -run 'DeployImageAsset|AssetPaths'
cp $SC/A.bak $A
grep -c "zzz-never-matches\|declined: deploy_path" $A   # expect 0
```

The mutation worth keeping: insert `_ = storage.DeployedAssetPath("x","y")` **above**
the guard. It fails **only** the ordering assertion, isolating ordering from
existence — which "delete the guard" cannot, because that fails everything at once.

## Applying seed 307

```bash
kubectl -n ai-persona-system exec -i postgres-clients-0 -- psql -U clients_user -d clients_db \
  -v ON_ERROR_STOP=1 < docs/agent_docs/sql_for_agents/307_asset_deployer_stops_declaring_deploy_path.sql
```

**Gotcha — a verify block made of `SELECT`s cannot stop the `COMMIT`.**
`ON_ERROR_STOP` ignores a non-empty result, so a "verification" that prints the
wrong rows still commits them. Use `DO … RAISE EXCEPTION`, which 307 does.

**And induce it.** A guard that can never fire passes identically to one that
works. After applying, run an inverted copy against the live row and watch it
raise:

```sql
DO $$
DECLARE bad int;
BEGIN
  SELECT count(*) INTO bad FROM agent_definitions
   WHERE type='asset-deployer' AND is_active AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL
     AND (input_contract->>'optional') LIKE '%purpose%';          -- deliberately TRUE
  IF bad <> 0 THEN RAISE EXCEPTION 'CONTROL FIRED as designed (%)', bad; END IF;
END $$;
-- ERROR:  CONTROL FIRED as designed (1)      <- what you want to see
```

`snapshot_agent` has two overloads — `(text)` and `(text, text)`; 307 uses the
second so the reason is recorded with the snapshot.

## Verifying the roll

**Take the baseline BEFORE the roll**, and use a positive AND a negative control in
the same exec — this change *removes* a string as well as adding one, which is the
strongest form of the `bugs_open/153` recipe:

```bash
for POD in $(kubectl -n ai-persona-system get pods -l app=agent-chassis -o name); do
  echo "== $POD"
  kubectl -n ai-persona-system exec $POD -- sh -c '
    echo -n "NEW refusal (expect >=1 post-roll): "; strings /app/agent-chassis | grep -c "refused: deploy_path"
    echo -n "REMOVED override (expect 0 post-roll): "; strings /app/agent-chassis | grep -c "Using custom deploy_path"
  '
done
```

Pre-roll that reads `0 / 1`; post-roll it must read `>=1 / 0` on **every** replica.
Without the removed-string control, a stale image and a fresh one look alike.

**Behavioural induction** (wait ≥300s after any pod restart — a spawn inside that
window is silently dropped): dispatch `asset-deployer` with a `deploy_path` and
assert `deploy_result` is `deployed:false, skipped:true, reason LIKE 'refused: deploy_path%'`,
that no git commit fired, and that the probe path 404s. Then the **healthy
control**: the same dispatch without `deploy_path` must still deploy. Without that
second half, a guard that refuses *everything* looks like success.
