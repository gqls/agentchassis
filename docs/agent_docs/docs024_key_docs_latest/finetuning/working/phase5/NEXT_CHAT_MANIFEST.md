# NEXT_CHAT_MANIFEST — Flywheel-C checkpoint upload / loop-await race

What to have loaded for the next session. Tags:
- **[in project]** — already in project knowledge; carries over if the next chat is in the same project.
- **[add]** — produced/needed this session and NOT in the project manifest; save into the project or upload at the top of the next chat.

The gap to close before starting: the **[add]** items below (this-session Go files, the three `109*`
migrations, the four start-here docs, and the code extract). Everything **[in project]** should already be there.

---

## 1. Start here (entry point)
- **`HANDOFF_2026-06-06_checkpoint_upload_loop_await_race.md`** [add] — the plan, verified root cause,
  the fix steps (incl. the now-confirmed `preRegisterAwaitedRequest` body + its 120s/step_name caveats),
  the batch fallback, cleanup, IDs, verify queries. **Read first.**
- `NOTES_phase5_training_launcher_running.md` [add] — institutional log; race entry + §8 signature.
- `RUNBOOK_phase_b_c_d_deploy.md` [add] — deploy/test steps, b2 CLI, the `orchestrate`→model-trainer
  trigger, export verification.
- `PLAN_checkpoint_and_artefact_upload_b2.md` [add] — the upload-path plan.

## 2. The code dump — use the targeted extract, not the 72k-line file
- **`CHASSIS_await_loop_extract.txt`** [add] — **~2,000 lines (2% of the full dump)**. Whole functions
  for the async-await + loop machinery, each headed with its ORIGINAL line range for cross-ref. Contains:
  `ProcessResponse` (L167), `ClaimAwaitedRequest` (L396), `processAwaitResponse` (L1678),
  `persistAwaitingStateWithRetry` (L1831), `extractRequestID`/`determineResponsesTopic`/
  `determineRequestsTopic`/`createAwaitedRequest`/`extractTargetAgent*` (L1899–2010),
  `handleCompleteResponse` (L2246), `continueExecution` (L753), `handleRequestTimeout` (L2969),
  `routeToErrorStepOrFail`/`failWorkflow` (L3159/3233), `getTimeout` (L3594),
  `GetAwaitedRequestStatus`/`InsertAwaitedRequest`/`GetAwaitedRequest`/`GetAwaitedRequestWithRetry`
  (L3721/5300/5350/5399), loop fns `setLoopVariable`/`propagateIterationOutputs`/
  `shouldContinueLoopOnError`/`skipToNextLoopIterationForAsync` (L6700/6849/7090/7263),
  `LoopCompleteAction` (L39489), `preRegisterAwaitedRequest` (L57855, the fix helper), and the spawn
  pre-register call-site pattern (L24985–25035). **Use this for the race work.**
- `production_agent-chassis-actions-current_context.txt` [in project] — the FULL dump. Only re-add if
  you go OUTSIDE the await/loop area (e.g. other actions). For this task the extract is enough.

## 3. Go — the fix and the path it touches
- **`thunder_prepare_object_url_dispatch.go`** [add] — **the file the fix edits** (add the
  `preRegisterAwaitedRequest` call before `ProduceWithValidation`).
- `spawn_actions.go` [in project] — the proven register-before-send call site to mirror (L25020).
  (Also in the extract as the "pattern to mirror" snippet.)
- `loop_expansion_handler.go` [add] — `setLoopVariable` / `propagateIterationOutputs` loop item mechanics.
- `compute_checkpoint_keys_action.go`, `assemble_upload_manifest_action.go`,
  `flatten_presign_results_action.go` [add] — Phase B actions feeding the loop + manifest.

## 4. Go — the batch fallback (if pre-register is fragile)
- `adapter.go` [add], `data_url_actions.go` [add] — thunder adapter: `prepare_object_url` handler +
  reply path; where a plural `prepare_object_urls` would go.
- storage layer: `s3.go`, `url_helpers.go`, `interface.go` [add] — presign helpers
  (`GetPresignedPutURL`/`PresignPutObject`) to reuse.
- `registry.go` [in project] + `registry_patch_phase_b.txt` [add] — action registration.
- `call_agent.go` [add, optional] — register-before-send contrast / the call_agent-helper route.

## 5. Migrations (this session)
- `109_launcher_upload_manifest_wiring.sql` [add]
- `109a_fix_write_manifest_workspace_perm.sql` [add]
- `109b_fix_presign_one_loop_item_keypath.sql` [add]

## 6. Schemas (check before any SQL)
- **`awaited_requests`** — the central table for the race (status / request_id / step_name / timeout_at).
  Inspect via `schemas_all` [in project] or `\d awaited_requests` live. (Its columns are also visible in
  the `InsertAwaitedRequest` / `preRegisterAwaitedRequest` SQL in the extract.)
- `019_model_lifecycle_schema.sql` [in project] — `training_runs`, `thunder_instances`, `training_exports`.
- `agent_definitions` + `training_exports` schemas — in `schemas_all` / `schemas_some` [in project].

## 7. Table content (run live — cannot be pre-baked)
- Launcher def config — verify `presign_one` shows `key_path:"ckpt_key"` (query in handoff §Verify).
- Export rows — pick a non-zero `actual_rows` via the `LEFT JOIN training_exports.rows` count.
- `awaited_requests` rows during the test — to see `status_before` at a stuck iteration.
- Reference IDs (in handoff): model-trainer `94f5a069`, training-launcher `1223bdc1`, data-preparer
  `71ab9361`, monitor-orchestrator `c3b4c052`, monitor-worker `470c6b3f`; exports `146a9a12` / `fef7be6b`
  good, `a8484922` bad; chassis image `v1.0.1057`.

## 8. Docs / standards
- `001_development_guide_3_.md` [in project] — loop reference + agent-creation guidelines.
- `003_contracts_and_standards.md` [in project] — message headers/body contracts.
- `002_system_architecture.md` [in project] — system overview.
- `FOCUS_finetuning_flywheel_and_service_23_.md` [in project] — training_exports schema + flywheel context.

## 9. Debug docs
- `016_debugging_guide.md` / `016_debugging_guide_v2_27.md` /
  `016_debugging_guide_addendum_adopted_tools_no_widget_3_.md` [in project].
- `016_debugging_guide_v2_30.md` [add] — the version updated this session (latest), if you want it.

## 10. Scripts / training assets (for the re-run + bundle)
- `run.sh`, `02_train_llama_3_3_70b.py` [add] — training scripts (SAVE_STEPS / bundle; B2 bundle is still
  the `=50` build).
- `isolation_test_phase_a.py` [add] — Phase A isolation test.
- kcat trigger scripts [add] — the `orchestrate`→model-trainer trigger is in RUNBOOK §4b.

## 11. Deploy / ops (all [in project])
- `thunder-adapter.yaml`, `deployment.yaml`, `service.yaml`, `kustomization.yaml`, `rbac.yaml`,
  `makefile.txt`.
- ns `ai-persona-system`; Kafka `-n kafka` (cluster `personae-kafka-cluster-...`). B2 bucket
  `personae-model-training`, endpoint `https://s3.us-east-005.backblazeb2.com`, **b2 CLI** (not aws).

---

### First actions next session
1. Read the handoff; confirm `ActionParams.CurrentStep` holds the expanded loop-substep name at dispatch time.
2. Add `preRegisterAwaitedRequest(...)` to `thunder_prepare_object_url_dispatch.go` before the send
   (guard `params.DB != nil`); keep returning `await_response:true`.
3. Rebuild + redeploy chassis (bump from v1.0.1057). Re-run with export `146a9a12` or `fef7be6b`.
4. Confirm each `presign_checkpoints_iter_N_presign_one` logs `ClaimAwaitedRequest: status_before=waiting … claimed:true`; loop → flatten → `MANIFEST_WRITTEN` → `ssh_exec_launch`.
5. Cleanup first if not done: delete stuck launcher jobs `24eb2c59` / `48c12fe3`; verify no live `thunder_instances`.
