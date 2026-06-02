# Running notes — Phase 5 training-launcher (Flywheel C)

Decision + reasoning log for automating the `iter_0` training launch. Append as
we go. Epistemic tags used below: **[verified-source]** read from the actual
code/schema this session; **[verified-db]** confirmed by querying production;
**[deployed?]** changed in code and reported deployed but image tag not yet
re-confirmed; **[assumed]** inferred from surrounding code, not directly proven;
**[gap]** known missing piece.

Last updated: 2026-06-02.

---

## 1. Objective and system shape

Build the real `training-launcher` agent (replacing a stub) so a fine-tune runs
end to end without manual VM steps: Llama 3.3 70B QLoRA on a Thunder Compute GPU,
dataset + artefacts in Backblaze B2. Target first run: `export_id 146a9a12-…`.

Orchestration chain (all Kafka/saga, namespace `ai-persona-system`, DB
`clients_db`): `model-trainer` (orchestrator) spawns and then calls, in order,
`training-data-preparer` → `gpu-provisioner` → `training-launcher`. Each
`call_*` is a `call_agent` step that awaits the child's final result. The
children that touch Thunder do so through `thunder-adapter`, which is the
credential boundary (holds the B2 keys); the child dispatches a request to
`system.adapter.thunder.requests` and awaits the adapter's reply.

Two-level await, and the distinction matters for the topic bug below:
- A child's **intermediate** adapter calls (e.g. the launcher's presign / ssh_exec)
  are awaited by **that child's own** coordinator, on the child's **own** response
  topic.
- A child's **final** result goes up to its **parent** via the coordinator's
  `notifyParentOfSuccess`/`Failure`, on the **parent's** topic.

---

## 2. Load-bearing architecture facts [verified-source unless noted]

- **Local action steps do not resolve `input_mapping`.** The coordinator resolves
  `output_mapping` for local steps but only honours `input_mapping` for
  `call_agent` (building a child's `input_data`) and loop fan-out. On a plain local
  action, `input_mapping` is **dead config**. The canonical way a local action
  pulls a value (including a prior step's output) is a plain config key whose
  value is a dot-path, resolved from `collected_data` by `ExtractActionInputs`
  Strategy 0 (action_inputs.go L119) or, for ssh command tokens, by
  `resolveTemplateToken`.
- **Reply-topic fields.** Coordinator fills `ExecutionContext.ResponsesTopic` from
  `__my_responses_topic__` **first** (coordinator.go ~L1352).
  `__parent_responses_topic__` is used **only** by `notifyParentOfSuccess`/
  `Failure` (~L3298/L3487) — the child→parent final notification. So an
  intermediate adapter reply must come back on the agent's **own** topic.
- **Launcher's own response topic** is `system.responses.training-launcher`
  (agent_def). Note this is *not* the `system.agent.<type>.responses` shape that
  the dispatch actions' fallback branch builds — see the caveat in §6.
- **Bucket** [verified-source + verified-db reasoning]: `thunder-adapter`
  defaults `TRAINING_BUCKET=personae-model-training`; the preparer's live config
  also writes there (its agent_def `s3_bucket="finetuning"` is stale/logical, not
  the real bucket). `dataset_uri` comes from the preparer's own `S3Client.Upload`,
  so the key the launcher strips out is exactly what was written; the adapter
  presigns it against the same bucket. Aligned.
- **Adapter `ssh_exec` blocks.** It runs `session.Run(command)` (ssh_exec_actions.go
  ~L218), which waits for the remote command to exit, bounded by a 5-min
  `sshCommandTimeout`. This is *why* the launch command must detach (see §4,
  setsid). Reply shape: `{provisioning_id, exit_code, stdout, stderr, reachable}`;
  `LAUNCH_PID` is inside `stdout`.
- **`ProvisionInstanceResult`** carries `provisioning_id` (provision_action.go
  ~L109), so threading it from provisioner → launcher is sound.
- **Proven dispatch actions** (`provision`, `decommission`) derive the reply topic
  from the agent's **own** `ExecutionContext.ResponsesTopic` (with a
  `system.agent.<type>.responses` fallback). These work in production.

---

## 3. Decision log

### D1 — Don't change the orchestrator; fix the thunder side
Initially the launcher's local steps used `input_mapping`, which is dead config
(§2). A first attempt added a `coordinator.go` change to resolve `input_mapping`
on every local step. Reversed after re-reading source and the question "why
change the orchestrator when the other adapters work as-is?": the other adapters
work because their inputs already arrive in `input_data`; the launcher's only
genuine novelty is a **cross-step reference** (a prior presign step's output),
and the existing config-dot-path mechanism already handles that. Teaching the
chassis a new behaviour to serve one agent's misuse was disproportionate.
**Decision: withdraw the coordinator change; fix thunder-side only.** Smaller
blast radius, no chassis-wide behavioural change, makes thunder resolve inputs
the same way every other local action does.

### D2 — How each launcher input is resolved (no `input_mapping` on local steps)
- `presign_dataset` dataset: the action falls back to `input_data.dataset_uri`
  (the preparer's URI, threaded by `call_launcher`) and strips it to a key.
  Config carries only `{method: GET}`.
- `presign_scripts`: literal `key` + `method` in config.
- `ssh_exec_launch` cross-step URLs: config keys `scripts_url` /`dataset_url`
  hold dot-paths (`scripts_url_result.presigned_url`, …); `resolveTemplateToken`
  resolves the `{scripts_url}`/`{dataset_url}` command tokens from
  `collected_data`. `provisioning_id` is read from `input_data` (not config).
- `mark_running`: config `{training_run_id: "input_data.training_run_id"}`,
  resolved by `ExtractActionInputs` Strategy 0 [verified-source: spec Required
  `training_run_id`, then `uuid.Parse`].
- `model-trainer.call_launcher` is a real `call_agent` step, so `input_mapping`
  is correct there; we added `provisioning_id ← provisioning_result.provisioning_id`.

### D3 — Reuse over parallel code in the adapter presigner
The adapter needed a general presign-by-key (`ObjectURL`). Rather than a third
parallel signer, `DatasetURL`/`ArtefactURL` were refactored to **delegate** to
`ObjectURL` (same keys, same GET/PUT, same default expiries, now centralised in
the method switch). One presign path.

### D4 — Launcher reply topic = own topic (NOT parent topic)
The launcher's `ssh_exec`/`prepare_object_url` dispatch actions had preferred
`__parent_responses_topic__`. The proven `provision`/`decommission` use the
**own** topic, and §2 shows `__parent_responses_topic__` is for the final
child→parent notification, not intermediate adapter-reply matching. Routing an
adapter reply to the parent lands it where no `awaited_request` matches → the
launcher publishes fine and **hangs**. **Decision: align both launcher actions to
the provision/decommission own-topic derivation.** This is consistency with the
proven pattern, not a new invention.

---

## 4. The `setsid` launch command

`ssh_exec` blocks to command exit (§2), so the launch must return immediately.
The command (single line, no literal newlines — survives JSON/Kafka transport):

```
mkdir -p /workspace; setsid bash -c 'curl -fsSL "{scripts_url}" -o /workspace/bundle.tar.gz && tar -xzf /workspace/bundle.tar.gz -C /workspace && curl -fsSL "{dataset_url}" -o /workspace/training_iter0.jsonl && chmod +x /workspace/run.sh && /workspace/run.sh > /workspace/train.log 2>&1' < /dev/null > /workspace/launch.log 2>&1 & echo "LAUNCH_PID=$!"
```

The fetch+train chain runs under `setsid` in a new session, stdin `</dev/null`,
stdout/stderr → `/workspace/launch.log`, backgrounded with `&`; the SSH channel
hits EOF right after `echo`, so `session.Run` returns fast with `LAUNCH_PID`.
`run.sh` keeps its own `/workspace/train.log`. Presigned URLs are
percent-encoded so they sit safely double-quoted inside the single-quoted
`bash -c` body. Parses under `bash -n` with `&`-bearing URLs.

---

## 5. Changes that shipped (this work)

Migration (DB):
- `102_training_launcher_real.sql` — **[verified-db] applied and correct**: launcher
  steps `presign_dataset → presign_scripts → ssh_exec_launch → mark_running →
  complete`, no `input_mapping` on local steps, `complete` uses `output_fields`
  (plural array), scoped to the one `is_active=true` row; `model-trainer.call_launcher`
  has `provisioning_id`. **Do not re-run.**

Adapter image (`thunder-adapter`):
- `internal/adapters/thunder/data_url_actions.go` — adds `ObjectURLRequest`,
  `ObjectURL` (presign by key, GET default 60m / PUT 24h), `handlePrepareObjectURL`
  (reply key `presigned_url`); `DatasetURL`/`ArtefactURL` delegate to `ObjectURL`.
- `internal/adapters/thunder/adapter.go` — adds `case "prepare_object_url"` to the
  dispatch switch.

Chassis image (`agent-chassis`):
- `platform/orchestration/actions/thunder_prepare_object_url_dispatch.go` —
  `input_data.dataset_uri` fallback + own-topic reply derivation (D4).
- `platform/orchestration/actions/thunder_ssh_exec_dispatch.go` —
  `resolveTemplateToken` for command tokens + own-topic reply derivation (D4).

Withdrawn: `coordinator.go` change (D1). Coordinator is untouched.

Static review status: no bugs found. Verified each shared package symbol declared
once; all imports used (`prepare_object_url` correctly has no `datahelpers`
import, using package-local `configOrInput`); brace/paren/quote balance; full
launcher data-flow re-traced to resolution; Strategy 0 + `mark_running` confirmed
from source. **Not provable here: a real `go build`/`go vet`** (no Go toolchain
and not the full module) — that remains the compile gate.

---

## 6. Open caveat to watch at runtime (not a code fix)

The own-topic fallback branch builds `system.agent.<type>.responses`, but the
launcher's real response topic is `system.responses.training-launcher`. If
`ExecutionContext.ResponsesTopic` is ever empty and the fallback fires, the reply
routes nowhere → hang. We deliberately did **not** "fix" the fallback, because it
mirrors `provision`/`decommission` (which work), so the primary path
(`ResponsesTopic` seeded from `__my_responses_topic__`) is what carries it; a
divergent fallback would re-introduce the inconsistency D4 removed.

**Runtime check:** the first run's launcher dispatch log should read
`await_responses_topic=system.responses.training-launcher`. If it reads
`system.agent.training-launcher.responses`, then `__my_responses_topic__` isn't
seeded for the launcher — which would also affect `provision`, so it'd be a
pre-existing chassis seeding issue, not this change.

---

## 7. Pre-trigger gates (order)

1. **Confirm both images actually moved** [deployed?]. The earlier "deployed" was
   chassis+migration only — the adapter `prepare_object_url` and the topic fix were
   not in it. Confirm `thunder-adapter` > `v1.0.1048` and the chassis tag is newer
   than the `__parent_responses_topic__` build; pods Ready, no fresh restarts.
   ```
   kubectl -n ai-persona-system get deploy -o jsonpath='{range .items[*]}{.metadata.name}{"\t"}{.spec.template.spec.containers[0].image}{"\n"}{end}' | grep -iE 'thunder|launcher|trainer|chassis'
   kubectl -n ai-persona-system get pods | grep -iE 'thunder|launcher|trainer'
   ```
2. **`go build ./...` + `go vet`** in-repo (the compile gate not runnable here).
3. **`prepare_object_url` round-trip**: reply has `presigned_url`; and the bundle
   exists at `personae-model-training / finetuning/scripts/bundle.tar.gz`. This
   also exercises the shared `ObjectURL` path, covering the `prepare_dataset_url`
   delegation regression.
4. **Cost gate**: a ~9h run is well over `daily_cap_usd=15`. Inspect `thunder_config`
   and raise the cap — verify the gate's comparison direction first so the change
   permits rather than blocks.
5. **Trigger `iter_0`** via the existing mechanism with `input_data` containing
   `export_id 146a9a12-…` and `hyperparameters` (read at
   `input_data.hyperparameters` by the preparer; passed through by the launcher).

---

## 8. Failure signatures → cause (quick reference)

- Adapter reply `not_implemented` on `prepare_object_url` → adapter image didn't
  pick up the new action (gate 1).
- `presign_dataset` dispatched, adapter logged a successful presign, launcher never
  advances, ~600s timeout → reply-topic mismatch / seeding (D4 not deployed, or the
  §6 caveat).
- curl error in `launch.log`, no `train.log` → presigned URL 404: bucket/key
  mismatch, or bundle not uploaded (gate 3).
- `RUN_SH_FATAL missing_required_file=…` → bundle/path issue.
- `ssh_exec` blocks ~5 min then errors → launch didn't detach (setsid not deployed).
- `RUN_SH_SMOKE_OK` absent + error in `train.log` → data format / OOM; the smoke
  pass did its job and stopped before the full run.

Healthy markers in `train.log`: `RUN_SH_START → RUN_SH_STEP step=setup →
RUN_SH_SMOKE_OK → RUN_SH_STEP step=full_train → RUN_SH_FULL_OK → RUN_SH_DONE`.

---

## 9. Known gaps (not introduced here, still open)

- **No training-monitor.** `running → complete/failed` is not automatic. Because the
  launch is detached, `ssh_exec` `exit_code` is always 0, so a later curl 404 leaves
  `training_runs` stuck at `running`. Watch `launch.log`/`train.log` on the first
  run. A monitor (poll `train.log`/artefact, transition the row) is the natural next
  build.
- **`prepare_artefact_url` PUT path untested** until there's an `adapter_out` to
  upload.
- The full train runs ~9h **outside any await window** (1958 examples × 3 epochs per
  the manual `iter_0` manifest); nothing is watching it but us until the monitor
  exists.

---

## 10. Handoff-correction log (institutional memory)

Where the inherited handoff disagreed with the deployed/source code. Pattern:
**verify against code, not the handoff.**

- Handoff: `prepare_object_url` was added to the adapter. **Reality:** deployed
  `v1.0.1048` was Phase-4 only (dataset/artefact by ID); no `prepare_object_url`
  handler or dispatch case. We wrote it (§5).
- Handoff: the launcher must use `__parent_responses_topic__` for replies;
  own-topic was "the bug". **Reality (D4):** backwards — `__parent_responses_topic__`
  is the child→parent final-notification topic; `provision`/`decommission` use
  own-topic and work.
- Handoff: the `provision` dispatch file was "patched to use
  `__parent_responses_topic__`". **Reality:** the uploaded `provision` file uses the
  own-topic derivation.
