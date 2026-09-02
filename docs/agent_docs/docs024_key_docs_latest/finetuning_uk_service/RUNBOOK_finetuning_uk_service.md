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

## 5. Provision claims — the duplicate guard, and how to clear a stuck one

Added 2026-08-12 with the `bugs_open/259` fix (`10659b419`, migration **396**
applied). The council's guardian seat objected — correctly — that a claim held
after a failure has **no operator surface**, leaving a manual DB edit under
production pressure. This is that surface.

**What the guard does.** `thunder_provision_claims` holds one row per
`correlation_id`, taken *before* the vendor call. A second provision under the
same correlation is refused. That is deliberate: the chassis retry driver
re-dispatches an expired await with a fresh `request_id`, and each re-dispatch
used to build another billable GPU.

**A failed attempt KEEPS its claim.** So after any failure, that correlation can
never provision again without a human clearing it. That is the safe side of the
trade, but it means a stuck claim looks exactly like a broken provisioner.

```sql
-- 1. Did the guard fire? attempts > 1 means the retry driver leaned on the door
--    and the guard held. This is the bug staying fixed, not a new problem.
SELECT correlation_id, attempts, status, thunder_instance_id,
       left(last_error, 120) AS err, created_at
FROM thunder_provision_claims
WHERE attempts > 1 OR status = 'failed'
ORDER BY created_at DESC LIMIT 20;

-- 2. Claims stuck in 'claimed' — the adapter died between claim and create.
--    Any box it built has NO thunder_instances row: check the vendor (§1b) and
--    the FTW-042 orphan sweep before clearing, or you may clear a claim whose
--    instance is still billing.
SELECT correlation_id, created_at, requested_by
FROM thunder_provision_claims
WHERE status = 'claimed' AND created_at < now() - interval '30 minutes';

-- 3. CLEAR one, deliberately, by correlation. Never bulk-delete: each row is the
--    only durable record that a provision was attempted (bugs_open/258 defect 3).
--    Confirm at the VENDOR first that nothing is billing for it.
DELETE FROM thunder_provision_claims WHERE correlation_id = '<the correlation>';
```

⚠ **Do not clear a claim to "retry" a provision.** If the workflow needs another
attempt, re-trigger it so it gets a **new** correlation — clearing the row
removes the audit trail and re-opens the exact hole 259 closed. Clearing is for a
claim that is genuinely orphaned (its attempt died and nothing is billing).

**Before unpausing** (`thunder_config.is_paused`), confirm the fix is in the
*running* binary — a committed fix is not a shipped fix:

```bash
kubectl -n ai-persona-system logs -l app=thunder-adapter --tail=300 | grep -m1 'build provenance'
git merge-base --is-ancestor 10659b419 <the sha that prints>   # exit 0 = the fix is in
```

## 6. Provision sizing and the wait deadline (258 fixes — LIVE from the roll after `236810e4e`)

### The vCPU count is no longer ours to choose

You do **not** need to pass `vcpus` any more. The adapter reads
`GET /v1/specs` and picks the lowest count Thunder publishes for the spec. The old
workaround (`"vcpus": 6` for a6000) still works and is still honoured verbatim —
and an explicit value **skips the catalogue lookup entirely**, so it is the thing
to reach for if `/specs` is ever down.

Check what it resolved, at the adapter rather than by inference:

```bash
kubectl -n ai-persona-system logs -l app=thunder-adapter --tail=500 \
  | grep 'Resolved vCPU count from Thunder specs'
# spec_key=a6000_x1_prototyping vcpus=6 vcpu_options=[6,8]
```

The catalogue itself, when you want to see the menu (read-only, free):

```bash
POD=$(kubectl -n ai-persona-system get pods -l app=thunder-adapter -o jsonpath='{.items[0].metadata.name}')
kubectl -n ai-persona-system exec "$POD" -- sh -c \
  'wget -qO- --header "Authorization: Bearer $THUNDER_COMPUTE_API_KEY" \
   https://api.thundercompute.com:8443/v1/specs' \
  | python3 -c "import sys,json;d=json.load(sys.stdin)['specs'];[print(f\"{k:30} {d[k].get('vcpuOptions')}\") for k in sorted(d) if d[k].get('gpuCount')==1]"
```

Measured 2026-08-13: a6000 `[6,8]` · a100xl `[8,12,16]` · l40 `[6,8,12]` · h100
`[4,8,12,16]` · every `*_production` spec wants exactly one value (15, or 10 for
l40). Note `l40s` is a GPU constant in our source with **no** live single-GPU
spec — asking for it now refuses rather than 400s.

### Tuning the wait deadline — read the coupling first

```sql
SELECT provision_wait_timeout_seconds FROM thunder_config;   -- 540 by default
UPDATE thunder_config SET provision_wait_timeout_seconds = 570;  -- live, no build
```

⚠ **`adapter wait` MUST stay BELOW `dispatch_provision`'s `timeout_seconds`.**

```sql
SELECT default_config->'workflow'->'steps'->'dispatch_provision'->'config'->>'timeout_seconds'
FROM agent_definitions WHERE type='gpu-provisioner' AND is_active
  AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;   -- 600
```

Going above that does **not** give you a longer wait — it gives you a *quiet
success*. The await expires first, the chassis retry driver re-dispatches, the
259 claim guard refuses the duplicate (correctly), and **the workflow reports
FAILED while a real, billed instance keeps running with nobody watching it.**
The bad outcome of raising this carelessly is a provision that worked and that
nobody knows about.

**So, to raise it: raise the STEP's `timeout_seconds` FIRST, then the column.**
In that order, so the invariant never inverts even briefly. The column's CHECK
allows up to 1800 — that is a bound on absurdity, not permission.

If a run reports a timeout, check which deadline actually applied before changing
anything — a silently-defaulted one logs a warning:

```bash
POD=$(kubectl -n ai-persona-system get pods -l app=thunder-adapter -o jsonpath='{.items[0].metadata.name}')
kubectl -n ai-persona-system logs "$POD" \
  | grep -E 'Provision wait deadline from live config|using compiled-in default'
```

`using compiled-in default (is migration 400 applied?)` means the binary has the
fix but the database does not.

⚠ **The deadline logs as an integer, not a duration.** The line reads
`"wait_timeout":540`, **not** `wait_timeout=9m0s`. Earlier handoffs predicted the
duration form; grepping for `9m0s` finds nothing and reads as "the fix is missing".
Corrected 2026-08-15 against the real line.

⚠ **Use the POD NAME, not `-l app=…`.** See §8.

## 7. Firing a provision by hand — the whole recipe (added 2026-08-15)

This lane's most-needed command was **not written down anywhere** until now;
reconstructing it took a repo-wide search. The two `scripts/initial_messages/`
files that look like triggers (`270_model_trainer/081_gpu_provisioner.sh`,
`300_thunder_flywheel/081b_thunder_adapter_gpu_provisioner.sh`) are **stale
scratchpads from May/June**, non-executable, with destructive SQL interleaved
below the part that looks like the end. **Do not run them.**

**This spends money. `is_paused` is fleet-wide. Get the owner's word first.**

```bash
# 0. PRE-FLIGHT — must be can_provision=f for reason "paused", spend 0, active 0,
#    and the vendor must agree there is nothing already running.
#    (Vendor call is in §1b; `thunder_instances` count(*) includes HISTORY —
#     group by status or the 23 decommissioned rows will scare you.)

# 1. UNPAUSE (the money step)
UPDATE thunder_config SET is_paused = false, pause_reason = NULL;

# 2. DISPATCH exactly one. Topic system.agent.generic.requests; the chassis reads
#    the agent type from body.config.agent_type (processor.go:1240-1265).
C=$(cat /proc/sys/kernel/random/uuid); R=$(cat /proc/sys/kernel/random/uuid)
M=$(cat /proc/sys/kernel/random/uuid); O=$(cat /proc/sys/kernel/random/uuid)
echo "SAVE THIS: correlation=$C"
kubectl -n kafka run -i --rm kcat-prov-$(date +%s) --image=edenhill/kcat:1.7.1 \
  --restart=Never --quiet -- \
  kcat -P -b personae-kafka-cluster-kafka-bootstrap.kafka.svc.cluster.local:9092 \
  -t system.agent.generic.requests \
  -H correlation_id=$C -H request_id=$R -H message_id=$M \
  -H orchestration_id=$O -H orchestration_name=manual-gpu-provision-$(date -u +%Y%m%d-%H%M%S) \
  -H step_name=start -H client_id=demo_client -H message_type=request \
  -H action=orchestrate -H from_agent_type=cli -H from_agent_id=cli-manual-phase0 \
  <<<'{"action":"orchestrate","config":{"agent_type":"gpu-provisioner"},"input_data":{"gpu":"a6000","mode":"prototyping"}}'
```

- **Pass NO `vcpus`.** The adapter derives it from `GET /v1/specs` (bug 258 defect
  1). Supplying one tests nothing. `input_data` is flat: `gpu`, `mode`,
  `training_run_id`, `template`, `num_gpus`/`vcpus`/`disk_size_gb` (ints, forwarded
  only when > 0 — `thunder_provision_dispatch.go:113-127`).
- **All four ids must be real UUIDs** — `orchestration_states.correlation_id` is a
  `uuid` column, so a friendly string fails the cast at verify time.
- **Single-line here-string `<<<`, never a heredoc.** `kcat -P` sends **one message
  per line**; a pretty-printed body becomes N invalid fragments that still carry the
  headers, and the run reads `owner_agent_type='generic'`, `execution_path=[]`,
  `status=COMPLETED` — a fast fake success.
- **`kcat` exits 0 having sent nothing.** The exit code is not evidence. Verify:

```sql
SELECT owner_agent_type, current_step, status FROM orchestration_states
WHERE correlation_id = '<C>'::uuid;          -- expect gpu-provisioner | complete | COMPLETED
SELECT attempts, status, thunder_instance_id FROM thunder_provision_claims
WHERE correlation_id = '<C>';                 -- expect 1 | succeeded | <id>
```

### Cleaning up — do NOT leave it to the reaper

The reaper's deadline for a non-training box is **2 hours** (`max_uptime_hours: 2`
in the provision log), and `default_hard_uptime_hours` is 18. Either is far too long
for a test. Decommission explicitly, straight at the adapter topic:

```bash
C=$(cat /proc/sys/kernel/random/uuid); R=$(cat /proc/sys/kernel/random/uuid)
O=$(cat /proc/sys/kernel/random/uuid); PROV=<provisioning_id from the provision response>
BODY=$(printf '{"headers":{"correlation_id":"%s","orchestration_id":"%s","client_id":"demo_client","step_name":"decommission","request_id":"%s","message_type":"request","action":"decommission_instance","sender_agent_type":"cli","sender_agent_id":"%s","responses_topic":"system.agent.generic.responses","reply_to_topic":"system.agent.generic.responses"},"body":{"action":"decommission_instance","reply_to_topic":"system.agent.generic.responses","provisioning_id":"%s","reason":"manual cleanup"}}' "$C" "$O" "$R" "$O" "$PROV")
echo "$BODY" | kubectl -n kafka run -i --rm kcat-decom-$(date +%s) --image=edenhill/kcat:1.7.1 \
  --restart=Never --quiet -- \
  kcat -P -b personae-kafka-cluster-kafka-bootstrap.kafka.svc.cluster.local:9092 \
  -t system.adapter.thunder.requests \
  -H correlation_id=$C -H orchestration_id=$O -H request_id=$R \
  -H client_id=demo_client -H message_type=request -H action=decommission_instance \
  -H step_name=decommission -H sender_agent_type=cli -H sender_agent_id=$O \
  -H reply_to_topic=system.agent.generic.responses -H responses_topic=system.agent.generic.responses
```

Note the **adapter topic takes a wrapped envelope** `{"headers":{…},"body":{…}}` as
the message value (unlike the chassis intake in step 2, which takes a bare body).
`provisioning_id` is the `db_row_id` / `provisioning_id` from the provision
response. Then **re-pause**, and verify at the vendor, not at our tables:

```sql
UPDATE thunder_config SET is_paused = true, pause_reason = '<what you did, and the result>';
```

### Measured 2026-08-15, first successful end-to-end run

- **a6000 cold boot: ~16s** (`createdAt` → first `RUNNING` poll; 5s poll interval,
  so true value is 11–16s). The lane's long-standing `> 5 min` figure **does not
  hold for a6000** — though both historical `>5 min` rows were `a100xl`, so it may
  still hold there. 540s is ~33× the measured a6000 need.
- Provision → `Provision complete`: **16.5s**. Decommission → gone: **~7s**.
- ⚠ **`cost_usd` is OUR estimate, not Thunder's charge.** It is
  `default_hourly_rate_usd` (a flat **$1.80/hr**, all GPU types) × uptime —
  `provision_action.go:429` stamps the rate, `decommission_action.go:152` computes
  from it. **SETTLED by the invoice (2026-08-18): a6000 bills $0.35/hr FLAT
  (6-vCPU minimum included — no surcharge), a100xl $1.09/hr, per-minute; our
  books run exactly 5.1× over for a6000 ($5.72 booked vs $1.12 invoiced for the
  whole of Phase 0).** The flat $1.80 stays DELIBERATELY — safe direction for
  the $30 daily cap, and the single flat column cannot carry per-type rates;
  lowering it weakens the fleet's spend guard for every card at once.

## 9. The Phase 0 TRAINING run — staged 2026-08-15, ready to fire (postponed on an owner credits hold)

Everything below was verified staged on 2026-08-15; the run itself was postponed
minutes before launch when the owner called a credits hold. **Nothing here has run
— firing it is the next session's first move once the owner clears spend.**

**Pre-verified this session:** bundle at `finetuning/scripts/bundle.tar.gz` (GET
presign resolves; md5 `a19557ccf61ac951c28e81254a8d76f7` — matches the handoff, so
it IS the env-var-parameterised bundle) · dataset at
`finetuning/datasets/phase0-2026-08-12/training.jsonl` (GET resolves, HTTP 206) ·
`02_train` takes `--instruction-part`/`--response-part` **literally, no unescape**,
so the env values must carry REAL newlines (`$'...\n'`), and its marker guard
(`02_train:280-288`) fails fast before any GPU time if you get that wrong.

```bash
# 1. Mint the three URLs (presign.py now lives IN THIS DIRECTORY — creds come
#    from the live k8s secret, never hardcoded):
export B2_KEY_ID=$(kubectl -n ai-persona-system get secret personae-storage-secrets -o jsonpath='{.data.B2_APPLICATION_KEY_ID}' | base64 -d)
export B2_KEY=$(kubectl -n ai-persona-system get secret personae-storage-secrets -o jsonpath='{.data.B2_APPLICATION_KEY}' | base64 -d)
export S3_ENDPOINT=https://s3.us-east-005.backblazeb2.com
RUN_TAG=phase0-$(date -u +%Y%m%d-%H%M)
python3 presign.py GET finetuning/scripts/bundle.tar.gz 240                       # → BUNDLE_URL
python3 presign.py GET finetuning/datasets/phase0-2026-08-12/training.jsonl 240   # → DATASET_URL
python3 presign.py PUT finetuning/artefacts/$RUN_TAG/adapter.tar.gz 240           # → FINAL_PUT_URL
# Verify both GETs with a range request BEFORE spending GPU time:
#   curl -s -o /dev/null -w '%{http_code}' -r 0-0 "$URL"   → expect 206
#   (a first 503 from B2 is transient — retry before concluding anything)

# 2. Provision an a6000 (§7 — unpause, dispatch, save the provisioning_id).

# 3. ONE ssh_exec (§7's decommission shows the envelope shape; body action is
#    "ssh_exec" with provisioning_id/command/timeout_seconds). The command
#    (CORRECTED against the 2026-08-15 live launch — see the traps below):
sudo mkdir -p /workspace && sudo chown ubuntu:ubuntu /workspace \
 && cd /workspace \
 && curl -sf -o bundle.tar.gz '<BUNDLE_URL>' && tar -xzf bundle.tar.gz \
 && curl -sf -o training_iter0.jsonl '<DATASET_URL>' \
 && printf '%s' '{"final":{"key":"finetuning/artefacts/<RUN_TAG>/adapter.tar.gz","url":"<FINAL_PUT_URL>"}}' > upload_manifest.json \
 && chmod +x run.sh 00_vm_setup.sh \
 && { BASE_MODEL=HuggingFaceTB/SmolLM2-1.7B-Instruct CHAT_TEMPLATE=auto \
    INSTRUCTION_PART=$'<|im_start|>user\n' RESPONSE_PART=$'<|im_start|>assistant\n' \
    SAVE_STEPS=0 MIN_VRAM_MIB=8000 nohup ./run.sh > train.log 2>&1 & } \
 && echo LAUNCHED_FOR_REAL

# 4. Poll: ssh_exec 'tail -20 /workspace/train.log' — markers, in order:
#    RUN_SH_START → step=setup (5-10 min venv+CUDA) → step=smoke →
#    RUN_SH_SMOKE_OK → step=full_train → RUN_SH_UPLOAD manifest=present →
#    RUN_SH_FULL_OK → RUN_SH_DONE. Measure each stage — that is FTW-032/035's ask.
# 5. Prove durability at B2, not at the marker: presign a GET for the adapter
#    key and range-request it → 206 = the artefact is really there.
# 6. Decommission (§7), re-pause, record timings in NOTES + §6.
```

Traps already known: the dataset must land as `training_iter0.jsonl` (run.sh's
hardcoded `DATA` path) · `SAVE_STEPS=0` = no checkpoints (right for a short run;
the FINAL upload still happens via the manifest and is the hard gate) · the
manifest JSON travels inside a kafka JSON envelope inside a shell command —
**escaping is the whole reason the old session had a `build_launch.py`**; compose
carefully and echo the command back before firing · `thunder-training-monitor`
stays **disabled** (its DONE_OK path decommissions the box; enabling early
destroys the artefact — FTW-035).

Traps found by the first live launch (2026-08-15):

- **`/workspace` does not exist on the `base` template until `00_vm_setup.sh`
  creates it** — but run.sh requires the files already placed there BEFORE setup
  runs. The launch command must `sudo mkdir -p /workspace && sudo chown
  ubuntu:ubuntu /workspace` first (now in the §9 command above; the originally
  staged command lacked it and failed on it live).
- **`… & echo LAUNCHED` lies.** The `&` backgrounds the whole `&&` chain and the
  `echo` runs unconditionally — the first launch returned `exit_code 0, stdout
  LAUNCHED` while stderr held `cd: /workspace: No such file or directory` and
  nothing had run. Group the backgrounding so the marker is conditional:
  `… && { ENV nohup ./run.sh > train.log 2>&1 & } && echo LAUNCHED_FOR_REAL` —
  **and always read stderr, not just exit_code/stdout: ssh_exec reports the
  SESSION's exit, not your chain's.**
- **Small-model runs must ALSO set `MIN_VRAM_MIB`** (e.g. `8000` for a 1.7B) —
  `00_vm_setup.sh`'s VRAM gate assumed the 70B default and refuses an a6000
  otherwise (fixed `2094a02e2`, default 79000 = old behaviour). It joins the
  move-together env set: `BASE_MODEL`/`CHAT_TEMPLATE`/`INSTRUCTION_PART`/
  `RESPONSE_PART`/`SAVE_STEPS`/`MIN_VRAM_MIB`.
- ~~⚠ B2 bundle redeploy pending~~ **DEPLOYED 2026-08-15 ~17:45Z** (owner-directed):
  bundle md5 **`6f27b21a6a4236c3c23679892337d0c3`** at
  `finetuning/scripts/bundle.tar.gz`, round-trip verified by boto3 read-back AND
  by the launcher's own presigned-GET path; the fetched `00_vm_setup.sh` carries
  `MIN_VRAM_MIB` and the old `-lt 79000` literal is gone from the gate line.
  `deploy_bundle.py` (this directory) is now the deploy — it refuses to report
  success unless the read-back md5 matches. The earlier `curl -T` was
  permission-blocked; boto3 is the same PUT and is not.

## 8b. ⚠ Boot time is DAY-VARIABLE — do not plan around one day's measurement

2026-08-15 measured a6000 ~16s and a100xl 12–17s. But 2026-08-12 measured the
**same a6000 spec** twice at **4m39s / 4m49s still STARTING**. Same vendor, same
spec, three days apart, ~20× difference. So: the 540s wait deadline is NOT
over-generous (it protects the slow days, which really happen), and a natural
259-guard firing (an await expiring on a slow boot) **remains possible** — rare on
a fast day, plausible on a slow one. Any claim shaped "boot takes X" must carry
its date.

## 8. ⚠ `kubectl logs -l app=<x>` can return NOTHING for a live pod

Bit this lane on 2026-08-15. `logs -l app=thunder-adapter --since=15m | grep …`
returned **empty** for lines that were definitely there; `logs <pod-name>` on the
same single, zero-restart pod returned the full history. On the empty result the
natural conclusion is "the code never ran", which would have sent someone hunting a
bug in working code.

**Always resolve the pod name first**, and when a grep comes back empty, **run a
control — a line you have already seen with your own eyes:**

```bash
POD=$(kubectl -n ai-persona-system get pods -l app=thunder-adapter -o jsonpath='{.items[0].metadata.name}')
kubectl -n ai-persona-system logs "$POD" | grep -c "Provision complete"   # control: must be > 0
kubectl -n ai-persona-system logs "$POD" | grep "Resolved vCPU count"     # the real question
```

If the control returns 0, your query is broken and the absence means nothing.

## Test-render a writer-prompt block against real loop data (added 2026-09-02)

`render_test_641/` holds a standalone harness that builds the template the way
`datahelpers.RenderPromptTemplate` does. `cd render_test_641 && go run .` renders `fixtures.json`
and prints each result; `!! ERROR` / `!! CONTAINS <no value>` are the failure markers. Refresh
fixtures from live rows with the query in NOTES 2026-09-02 (late). **Gotcha the harness exists
for:** the prompt is rendered against `ExtractFields(CollectedData, input_fields)`, NOT
CollectedData. A key at the CollectedData root is invisible unless the step names it:

```sql
SELECT default_config->'workflow'->'steps'->'process_sections_loop'->'config'->'sub_workflow'
       ->'steps'->'generate_content'->'config'->'input_fields'
FROM agent_definitions WHERE type='page-content-writer' AND is_active
  AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;
```

Run that BEFORE writing `{{.anything}}` into a prompt; if the key is absent the render is silently
empty (a nil range) rather than an error.
