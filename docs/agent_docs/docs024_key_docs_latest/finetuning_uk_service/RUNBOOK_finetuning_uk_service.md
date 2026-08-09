# RUNBOOK — finetuning.uk service

Commands that were hard to get right, each with its gotcha attached. Update HERE,
not in scrollback. Companion: `PLAN_2026-07-31_finetuning_uk_service.md`.

---

## 1. Phase −1 — rotate the Thunder Compute token (OWNER ACTION, blocks all GPU work)

> **STATUS 2026-08-08: RESOLVED — token live and verified from the pod**
> (`instances/list` → `{}`, authenticated, zero instances). **The mechanism of
> record changed (owner, 2026-08-05): the token is set via TERRAFORM, not
> `kubectl patch`** — see §1c below, which supersedes steps 2–3 for this key.
> The 401 episode's cause is now explained: the old key came FROM the tfvars, so
> any hand-patched key would have been silently reverted by the next apply.
>
> ~~STATUS 2026-08-03: STILL BLOCKED — the key in the cluster returns 401.~~
> (Old key `a73…96` rejected; controls — bogus token, no header — also 401,
> proving 401 = auth-reject. An invalid key cannot provision, so nothing billed.)

The key is `THUNDER_COMPUTE_API_KEY` in secret **`personae-default-secrets`**
(ns `ai-persona-system`), consumed via `envFrom` by the `thunder-adapter`
deployment (`deployments/kustomize/services/thunder-adapter/base/deployment.yaml:54-57`).
The adapter reads it once at startup (`internal/adapters/thunder/adapter.go:161`)
and refuses to start if empty.

1. Mint the new token in the Thunder Compute console.
2. Patch **only that one key** (the secret carries 11 other providers' keys —
   do not recreate the whole secret):
   ```bash
   kubectl -n ai-persona-system patch secret personae-default-secrets \
     --type merge -p "{\"stringData\":{\"THUNDER_COMPUTE_API_KEY\":\"<NEW_TOKEN>\"}}"
   ```
   ⚠ The token lands in shell history if typed inline — `read -rs TOK` first, then
   use `$TOK`.
3. Restart the adapter (envFrom is read at pod start; a secret edit alone changes
   nothing in the running pod):
   ```bash
   kubectl -n ai-persona-system rollout restart deployment thunder-adapter
   kubectl -n ai-persona-system rollout status deployment thunder-adapter
   ```
4. **Verify from the adapter pod, not a laptop** (proves the pod's own env, its
   egress path, and the token in one check). Read-only call:
   ```bash
   POD=$(kubectl -n ai-persona-system get pods -l app=thunder-adapter -o jsonpath='{.items[0].metadata.name}')
   kubectl -n ai-persona-system exec "$POD" -- sh -c \
     'wget -qO- --header "Authorization: Bearer $THUNDER_COMPUTE_API_KEY" \
      https://api.thundercompute.com:8443/v1/instances/list'
   ```
   Expect a JSON object (keyed by instance id; `{}` when none). A 401 means the
   token; a hang means egress. (Auth shape verified from
   `internal/adapters/thunder/api/client.go:235` — `Authorization: Bearer`.
   NOTE the container may lack `wget`/`curl` — if both are missing, run the same
   request from any pod that has curl, exporting the key via
   `kubectl get secret personae-default-secrets -o jsonpath='{.data.THUNDER_COMPUTE_API_KEY}' | base64 -d`.)
5. While the token is fresh, **capture current per-GPU rates** for pricing.
   ⚠ **`GET /v1/specs` carries NO prices** (verified 2026-08-08: 32 spec entries,
   hardware only — no price/rate/cost field anywhere in the payload). Rates come
   from https://www.thundercompute.com/pricing. Captured 2026-08-08:
   **a6000 $0.35/hr · l40 $0.79/hr · a100xl $1.09/hr · h100 $2.19/hr**, billed
   per minute; +$0.04/vCPU/hr beyond 4; storage $0.03/100GB/hr beyond 100GB;
   snapshots $0.05/GB/mo. ⚠ **There is NO `l40s` gpu_type** — the earlier plan
   text guessed one; the real menu is a100xl, a6000, h100, l40 (in x1/x2/x4/x8
   and prototyping/production variants). ⚠ `thunder_config` still says $1.80/hr
   for a100xl — the live rate is $1.09, so the gate over-estimates. Correct it
   with the UPDATE below **only when Phase 0 is actually about to run** — the
   gate also protects every other lane that might provision:
   ```sql
   -- rates informing the cost gate; $20 estimated_new_run_cost_usd is sized for
   -- 70B runs and will over-refuse small ones if left
   UPDATE thunder_config SET default_hourly_rate_usd = <rate>,
          estimated_new_run_cost_usd = <small-run estimate>;
   ```

## 1c. Token rotation THE MECHANISM OF RECORD — terraform (owner direction, 2026-08-05)

`personae-default-secrets` is **owned by Terraform root
`deployments/terraform/environments/production/uk001/047-base-configs`**
(`kubernetes_secret.personae_default_api_keys`, `main.tf:114-134`). A
`kubectl patch` on this secret is DRIFT and the next `terraform apply` reverts
it — which is exactly what made the 07-31 rotation vanish.

1. Owner puts the fresh token in `~/.config/thundercompute/token` (one line).
2. Update the tfvars **without printing the value** (from the root dir):
   ```bash
   python3 - <<'EOF'
   import re
   new = open('/home/ant/.config/thundercompute/token').read().strip()
   assert len(new) == 64 and '"' not in new
   p = 'terraform.tfvars.secret'
   s, n = re.subn(r'^(default_thunder_compute_api_key\s*=\s*)".*"$',
                  lambda m: m.group(1)+'"'+new+'"', open(p).read(), flags=re.M)
   assert n == 1; open(p,'w').write(s)
   EOF
   ```
   (`terraform.tfvars.secret` is gitignored — verified; never commit it.)
3. **Fingerprint-compare every key in the tfvars against BOTH live secrets
   before applying** (len + first-3/last-2 only). This root manages the whole of
   `personae-default-secrets` AND `personae-platform-secrets` AND the prod
   configmap — an apply against a drifted tfvars silently reverts someone
   else's hand-rotated key. 2026-08-05: all 19 values matched, so the apply was
   provably single-key.
4. `terraform plan -var-file=terraform.tfvars.secret` — expect exactly
   `0 to add, 1 to change, 0 to destroy`. Then apply (same `-var-file`).
   ⚠ The tool classifier blocks `terraform apply` from an agent session — the
   owner runs it (`!` prefix works).
5. The adapter reads the key at pod start: `rollout restart deployment
   thunder-adapter`, then verify from inside the pod (§1 step 4). `{}` =
   authenticated, zero instances.

## 1b. "AM I BEING BILLED RIGHT NOW?" — the check to run before bed

**Our database is not the source of truth for your bill; Thunder is.** Every
automated check we have reads `thunder_instances`. An instance Thunder is
charging for that has no row in that table is invisible to all of them. So the
authoritative check is the API, and it takes ten seconds:

```bash
POD=$(kubectl -n ai-persona-system get pods -l app=thunder-adapter \
        -o jsonpath='{.items[0].metadata.name}')
kubectl -n ai-persona-system exec "$POD" -- sh -c \
  'wget -qO- --header "Authorization: Bearer $THUNDER_COMPUTE_API_KEY" \
   https://api.thundercompute.com:8443/v1/instances/list'
```
`{}` (or an empty object) = nothing running = nothing billing. **Anything else,
read the `status` of each entry.** (If the container has neither `wget` nor
`curl`, run the same request from any pod that does — see §1 step 4.)

Then compare against what we *think* is running:
```sql
SELECT thunder_instance_id, status, instance_type,
       ROUND(EXTRACT(EPOCH FROM (NOW()-COALESCE(running_since,provisioned_at,created_at)))/3600.0,1) AS age_h
FROM thunder_instances WHERE status <> 'decommissioned';
```
**A row in the API output with no matching row here is an orphan** — it will
never be reaped automatically. Kill it by hand at the Thunder console, and file
it, because it means a provision wrote to Thunder and failed to write to us.

Emergency stop (halts all new provisioning immediately; DB config is live, no
deploy needed):
```sql
UPDATE thunder_config SET is_paused = true, pause_reason = '<why>';
-- undo: UPDATE thunder_config SET is_paused = false, pause_reason = NULL;
```

### What the automated reaper does and does not cover

- **It is ENABLED and firing** every 900s (verified 2026-07-31, `scheduled_tasks`
  name `thunder-reaper`, last tick 21:01Z). The training *monitor* is a separate
  task and is deliberately **disabled** until Phase 0 proves durability.
- **It has never actually reaped anything** — 0 of 23 all-time rows have
  `reaped_at` set. Firing ≠ working; the ticks have simply never had a target.
- Its original query missed three billing states (stuck `provisioning`, stuck
  `decommissioning`, `running` with a NULL `running_since`) — **measured 0 of 3
  on synthetic rows.** `sql_for_agents/280_thunder_reaper_widen_stuck_states.sql`
  fixes that (3 of 3), with the rollback in its header.
  **APPLIED 2026-08-03** — verify anytime with:
  ```sql
  SELECT enabled,
         pre_query LIKE '%stuck_provisioning%'    AS covers_provisioning,
         pre_query LIKE '%stuck_decommissioning%' AS covers_decommissioning
  FROM scheduled_tasks WHERE name='thunder-reaper';   -- expect t | t | t
  ```
- ⚠ **The dispatch field names disagree with the query's, and that is FINE** —
  do not "fix" it. The workflow description says
  `provisioning_id`/`thunder_identifier`, the pre_query returns
  `thunder_instance_id`. `thunder_decommission_dispatch.go:62-67` requires
  **either** `provisioning_id` **or** `thunder_identifier`; the query supplies
  `provisioning_id`, the form the source marks preferred. Checked 08-03.
- **Orphans: the automated sweep is LIVE AND VERIFIED** (2026-08-09, FTW-042).
  Fleet `v1.0.1273` carries the code (pod-grepped both chassis replicas + the
  adapter); `sql_for_agents/342` applied AND recorded in the
  `schema_migrations` ledger (`--record-only` — that ledger DOES cover this
  directory; the earlier "no ledger" note was wrong, WRONG_CALLS 08-09).
  `thunder-orphan-scan` runs every 6h; first verified run COMPLETED
  2026-08-09 11:42:06Z with counts truthful against a same-session manual
  `instances/list` and table read. Orphans are filed as `thunder_orphan`
  work items on `system.internal`:
  ```sql
  SELECT summary, status, created_at FROM site_work_items
  WHERE item_type = 'thunder_orphan' AND status NOT IN ('complete','cancelled','rejected');
  -- run result of any scan:
  SELECT status, collected_data->'reconcile_result' FROM orchestration_states
  WHERE owner_agent_type='thunder-orphan-scan' ORDER BY created_at DESC LIMIT 1;
  ```
  §1b stays valuable as the independent cross-check: a 0-findings scan proves
  it looked via `db_rows`/`vendor_billing`, but a same-day manual call is the
  cross-check that doesn't share the scan's code. ⚠ The first live run FAILED
  on a config-nested `output_field` (INERT — must be step-level; LANDMINES
  08-09, the reaper's own seed models the wrong form). ⚠ The filing path has
  never fired against a real orphan — first real orphan is its live proof.
- **Prove it can reap before trusting it** (drill, after the token is live):
  insert one synthetic `running` row with an obviously bogus
  `thunder_instance_id` (e.g. `T-DRILL-1`, never a bare small integer — real
  Thunder ids here are `0` and `1`, so a bad guess could kill a real box),
  `running_since` 30h ago; wait one tick; confirm the dispatch fired and the row
  moved. Decommission is idempotent on a 404, so a bogus id is safe. Delete the
  row afterwards.
- ✅ **PROVEN END TO END 2026-08-08** (drill row id `999999`, numeric-but-safe:
  `instances/list` was `{}` so no id could match a real box). Tick → terminal in
  32s: `running` → `decommissioned`, cost stamped, adapter log shows the real
  authenticated Thunder delete with `Thunder instance already deleted (404)`
  treated as success. Both drills together cost **$0.00**. Two caveats: the
  lookup passed because the row carried an IP — the NULL-IP case is
  `bugs_open/186` (fix committed `f83927375`, council-APPROVED `862583b1`);
  and orphans (below) remain uncovered.
- ✅ **186 VERIFIED LIVE 2026-08-08 17:52Z** — after the 16:27Z fleet roll
  (adapter `v1.0.1267`), the NULL-IP re-drill per the 114 template passed:
  `instance_ip` NULL row `999999` went tick→`decommissioned` in ~30s, lookup
  passed, real Thunder delete 404≡ok, cost stamped, drill row deleted, table
  back to baseline. The NULL-IP caveat above is closed; **orphans are now the
  only uncovered gap**. (Pod-grep controls were structurally unavailable — the
  diff adds/removes no string literal — so the drill is the proof of the roll,
  not a supplement to it.)

## 2. Scripts bundle — the training deploy unit

The on-VM scripts ship as `finetuning/scripts/bundle.tar.gz` in B2 bucket
`personae-model-training`. **Re-uploading the object IS the deploy** (FTW-031);
editing a script without re-tarring deploys nothing (byte-identical md5 trap).

Bundle must be FLAT (files at archive root):
```bash
cd docs/agent_docs/docs024_key_docs_latest/finetuning/working/scripts
tar -czf /tmp/bundle.tar.gz run.sh 00_vm_setup.sh 02_train_llama_3_3_70b.py 03_inference_test.py
# upload — see working/phase5/UPLOAD_bundle.sh for the exact b2/aws invocation + creds source
```
**HOLD the upload until Phase 0 actually runs.** The git copy is the reviewed
truth; the B2 object is the live one; between edit and upload they intentionally
differ. Verify after upload: download it back and `md5sum` against the local tar.

`run.sh` small-model support (added 2026-07-31, backward compatible): env vars
`BASE_MODEL` (default unchanged 70B) and `SAVE_STEPS` (default 50; set `0` for
short runs — checkpoints pointless, but the **manifest is still required**: it
carries the final adapter's presigned PUT URL).

## 3. Live-state queries (schema verified this session)

```sql
-- cost gate: can we provision right now, and why not
SELECT can_provision, denial_reason, total_24h_spend, active_count FROM thunder_provision_check;

-- training runs, newest first (NO base_model column — it lives in hyperparameters jsonb)
SELECT id, status, started_at, completed_at, final_loss, cost_usd,
       hyperparameters->>'base_model' AS base_model, left(error_message,60) AS err
FROM model_lifecycle.training_runs ORDER BY created_at DESC LIMIT 5;

-- instances (column is thunder_instance_id, NOT instance_id; type is instance_type)
SELECT thunder_instance_id, instance_type, status, created_at::date, cost_usd, reaped_reason
FROM thunder_instances ORDER BY created_at DESC LIMIT 5;

-- an export's REAL row count (runs.rows_exported can lie: a8484922 says 1957, holds 0)
SELECT count(*) FROM training_exports.rows WHERE export_id = '<uuid>';
```

## 4. Gotchas inherited from the flywheel (do not rediscover)

- Kafka trigger JSON must be flat single-line via here-string `<<<'{...}'` —
  heredocs mangle silently (016 §9).
- `thunder-training-monitor` stays **disabled** until Phase 0 proves
  `RUN_SH_DONE ⟹ adapter durable in B2` — its DONE_OK path decommissions the box
  (FTW-035). Enabling early destroys the artefact.
- Thunder gpu_type has **no plain "a100"** — it is `a100xl`; the composite spec
  key (`a100xl_x1_prototyping`) is NOT a valid gpu_type.
- Templates: `base` is the plain GPU template; named templates (`ollama`,
  `unsloth`, …) are pre-built stacks — retest for small runs, the ~25 min
  environment build dominates short-run cost.
- Healthy run markers in `train.log`: `RUN_SH_START → step=setup → RUN_SH_SMOKE_OK
  → step=full_train → RUN_SH_FULL_OK → RUN_SH_DONE`. Failure signatures table:
  `finetuning/working/phase5/NOTES_phase5_training_launcher_running(45).md` §8.
