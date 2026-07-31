# RUNBOOK — finetuning.uk service

Commands that were hard to get right, each with its gotcha attached. Update HERE,
not in scrollback. Companion: `PLAN_2026-07-31_finetuning_uk_service.md`.

---

## 1. Phase −1 — rotate the Thunder Compute token (OWNER ACTION, blocks all GPU work)

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
5. While the token is fresh, **capture current per-GPU rates** for pricing:
   `GET /v1/specs` (same header). Record the a100xl, l40, l40s, a6000 hourly
   rates in NOTES, and correct `thunder_config` if drifted:
   ```sql
   -- rates informing the cost gate; $20 estimated_new_run_cost_usd is sized for
   -- 70B runs and will over-refuse small ones if left
   UPDATE thunder_config SET default_hourly_rate_usd = <rate>,
          estimated_new_run_cost_usd = <small-run estimate>;
   ```
   ⚠ Do this UPDATE only when Phase 0 is actually about to run — the gate also
   protects every other lane that might provision.

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
