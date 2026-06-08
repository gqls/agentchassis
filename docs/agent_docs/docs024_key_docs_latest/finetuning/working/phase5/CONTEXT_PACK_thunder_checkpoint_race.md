# Context pack — Flywheel-C checkpoint upload: loop-await race (fresh thread)

Starting context for the open blocker. Flywheel C fine-tunes Llama 3.3 70B on Thunder A100s; checkpoints + the final adapter upload to Backblaze B2 via **presigned PUT URLs** in an upload manifest. The `training-launcher` (def type `training-launcher`, id `1223bdc1`) builds the manifest, writes it to the VM, SSH-launches training; it's spawned + called by the `model-trainer` orchestrator (id `94f5a069`). Chassis image `docker.io/aqls/agent-chassis:v1.0.1057`. DB `clients_db`; ns `ai-persona-system`, Kafka `-n kafka`. B2 bucket `personae-model-training`, endpoint `https://s3.us-east-005.backblazeb2.com`, **b2 CLI (not aws)**.

---

## State + next action

**Blocker:** the `presign_checkpoints` loop presigns correct keys but **stalls intermittently at a later iteration** (moving point ⇒ a race). The adapter replies (~1s, sometimes twice), but the launcher never clears the await; the timeout handler re-dispatches every ~3 min with a fresh request_id, so RetryVersion stays 0 and it never hits the max-retries fail path → effectively infinite retry.

**Root cause (verified against the code extract):** the loop substep is a **local** dispatch (`dispatch_thunder_prepare_object_url`) — it produces the adapter request and returns `await_response:true`, and the coordinator registers the awaited request **after** that (`processAwaitResponse` → persist → `InsertAwaitedRequest`). **Send-before-register.** For a ~1s reply the response can land before the `awaited_requests` row exists → `ClaimAwaitedRequest` (`WHERE status='waiting'`) finds nothing → reply dropped → timeout → retry → same race. `spawn_agent`/`call_agent` call `preRegisterAwaitedRequest` (register-**before**-send), which is why the working production loops (vet-batch `process_batch`, content-feed `process_sites`, both `call_agent`) don't stall and this local-dispatch loop does.

**Next action — the non-framework fix (not yet implemented):** in `thunder_prepare_object_url_dispatch.go`, call `preRegisterAwaitedRequest(...)` just before `ProduceWithValidation` (the send), guarded `if params.DB != nil`, reusing the existing helper `spawn_agent` uses; keep returning `await_response:true` (the coordinator's later `InsertAwaitedRequest` no-ops via `ON CONFLICT (request_id) DO NOTHING`).
- **First verify** `ActionParams.CurrentStep` holds the **expanded loop-substep name** (`presign_checkpoints_iter_N_presign_one`) at dispatch time — `preRegisterAwaitedRequest` uses `params.CurrentStep` for `step_name`, so the resume targets the right step only if it does.
- **Caveat:** the helper hardcodes a **120s** timeout and (via `ON CONFLICT DO NOTHING`) it wins over the step's configured timeout once pre-registered. Fine for ~1s presigns; note it pins every presign dispatch's await to 120s.
- It affects **all** presign dispatches (dataset/scripts/loop/final/resume) — they all become register-before-send; the working ones get more robust, the loop stops racing.
- Rebuild + redeploy the chassis (bump from v1.0.1057). Re-run with export `146a9a12` or `fef7be6b`. Confirm each `presign_checkpoints_iter_N_presign_one` logs `ClaimAwaitedRequest: status_before=waiting … claimed:true`, then loop → `flatten_checkpoint_urls` → `MANIFEST_WRITTEN` → `ssh_exec_launch`.

**If it still stalls** at a stuck iteration, capture from the launcher pod:
`kubectl -n ai-persona-system logs <launcher-pod> | grep -Ei "ProcessResponse|ClaimAwaitedRequest|status_before|no matching|InsertAwaitedRequest"` — tells you whether the row now exists when the reply lands (race closed, other issue) or the reply isn't consumed at all (different problem).

**Fallback (structural)** if the pre-register fix is fragile: replace the 40-iteration loop with **one batch** `prepare_object_urls` (plural) adapter call (takes the key array `compute_checkpoint_keys` already emits, returns `[{key,url}]`, all local SigV4 presigns) → no loop, no `flatten`, no race class. Bigger change (adapter handler + dispatch + a migration replacing `presign_checkpoints`+`flatten` with one step + adjust `assemble_manifest` inputs). Reuse the adapter's existing `prepare_object_url` handler + presign helpers (`GetPresignedPutURL`/`PresignPutObject`).

## Cleanup needed now

Kill stuck launcher jobs (`kubectl -n ai-persona-system delete job agent-training-launcher-24eb2c59 agent-training-launcher-48c12fe3`); verify no live `thunder_instances`; optionally mark stalled runs failed.

## Standing rules (the constitution)

Reuse existing funcs/helpers before recreating; **check schema before SQL**; complexity in Go, workflows thin; agents reply on the caller's responses topic; no `logger.Debug`; don't rename vars/fields; structural over patches; plain language. b2 CLI not aws. Namespaces `ai-persona-system`, `kafka`.

## Attach — code

- **`CHASSIS_await_loop_extract.txt`** — the ~2,000-line targeted extract (whole await/loop functions with original line ranges); the diagnosis source. Use this, **not** the 72k-line full dump.
- **`thunder_prepare_object_url_dispatch.go`** — the file the fix edits.
- `spawn_actions.go` — the proven register-before-send call site to mirror.
- `loop_expansion_handler.go` — `setLoopVariable`/`propagateIterationOutputs` loop-item mechanics.
- Phase B actions: `compute_checkpoint_keys_action.go`, `assemble_upload_manifest_action.go`, `flatten_presign_results_action.go`.
- **Fallback route only:** `adapter.go`, `data_url_actions.go`, storage `s3.go`/`url_helpers.go`/`interface.go`, `registry.go`.

## Attach — docs

`HANDOFF_2026-06-06_checkpoint_upload_loop_await_race.md` (start here), `NOTES_phase5_training_launcher_running.md`, `RUNBOOK_phase_b_c_d_deploy.md`, `PLAN_checkpoint_and_artefact_upload_b2.md`, `001_development_guide`, `003_contracts_and_standards`, `016_debugging_guide`.

## Pull — schema (fresh `\d`)

```
dbcontext -psql 'kubectl exec -n ai-persona-system postgres-clients-0 -- psql -U clients_user -d clients_db' \
  -schema awaited_requests,agent_definitions
```
Plus the `model_lifecycle` (`training_runs`,`thunder_instances`) and `training_exports` schemas (in `019_model_lifecycle_schema.sql` / `schemas_all`). `awaited_requests` is the central table (status / request_id / step_name / timeout_at); its columns also show in the `InsertAwaitedRequest`/`preRegisterAwaitedRequest` SQL in the extract.

## Pull — live data (run live; cannot be pre-baked)

- Launcher def config — verify `presign_one` shows `key_path:"ckpt_key"` (no `input_mapping`).
- Export rows — pick a **non-zero** `actual_rows` via the `LEFT JOIN training_exports.rows` count (`146a9a12`/`fef7be6b` good; **`a8484922` is 0 actual rows despite `rows_exported=1957` — do not use**).
- `awaited_requests` rows during the test — to see `status_before` at a stuck iteration.

(Exact queries in the handoff §Verify.)

## Key IDs

model-trainer `94f5a069`, training-launcher `1223bdc1`, data-preparer `71ab9361`, monitor-orchestrator `c3b4c052`, monitor-worker `470c6b3f`; exports `146a9a12`/`fef7be6b` good, `a8484922` bad; image `v1.0.1057`. Extract line refs: `processAwaitResponse` L1677, `ClaimAwaitedRequest` L396, `handleRequestTimeout` L2969, `preRegisterAwaitedRequest` L57855, spawn pre-register call-site ~L25020.

## Also pending (lower priority)

model-trainer fall-through after a failed `call_agent`; `SAVE_STEPS` (B2 bundle is the `=50` build — re-pack with `=10` for a fast checkpoint test); Phase D launcher wiring (D3+migration 110) only after the upload path is green; enable `thunder-training-monitor` last.

## Minimum set to start fast

The HANDOFF + `CHASSIS_await_loop_extract.txt` + `thunder_prepare_object_url_dispatch.go` + `\d awaited_requests` + the launcher-def-config and export-rows queries. Enough to confirm `CurrentStep` and apply the pre-register fix.
