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

## 5a. Live verification log — 2026-06-02

Confirmed against the running production system (not source-only):

- **Adapter `prepare_object_url` is live on `thunder-adapter:v1.0.1049`.** A manual
  round-trip (produce to `system.adapter.thunder.requests`, consume on
  `system.agent.generic.responses`) returned `success:true` with `presigned_url`,
  `key:finetuning/scripts/bundle.tar.gz`, `method:GET`,
  `sender_agent_type:thunder-adapter`, signing against `personae-model-training`.
  So 1049 carries our adapter change; the 5d22h pod age is fine. This resolves the
  §10 handoff item (adapter was Phase-4 on 1048; now correct on 1049).
- **Envelope shape confirmed:** the adapter reads `action` + `reply_to_topic` from
  the value's `body` (handleMessage L309–314); Kafka `-H` headers only supply
  `correlation_id`/`request_id`. Round-trip body must put `action`,
  `reply_to_topic`, and the B2 `key` inside `body`.
- **Scripts bundle uploaded** to `personae-model-training / finetuning/scripts/
  bundle.tar.gz` (7422 bytes, 2026-06-02 16:17). Contents verified flat and
  internally consistent (run.sh runs 00_vm_setup.sh; venv `${HOME}/unsloth_env`;
  `DATA=/workspace/training_iter0.jsonl`; train args match the manifest). Closes
  the "bundle not in B2" blocker.
- **Presign verification gotcha:** the presigned URL is signed for **GET**
  (`x-id=GetObject`). `curl -I` (HEAD) returns `403 SignatureDoesNotMatch` because
  SigV4 signs the method — this is NOT a presign/bundle failure. Verify with a GET
  (`curl -s -o /dev/null -w '%{http_code}'`, expect 200). Also: `kcat -f '%s'`
  escapes `&` as `\u0026`; decode the URL through `jq -r '.body.presigned_url'`
  before curling, or the literal `\u0026` yields a 400.
- **`b2 ls` syntax:** this CLI (v3+) needs a URI —
  `b2 ls --long "b2://personae-model-training/finetuning/scripts/"` — not the
  two-bare-arg form in `UPLOAD_bundle.sh`. (`b2 file upload <bucket> <local> <key>`
  works as-is.)

Still NOT confirmed: the chassis topic fix (D4) on 1049 — nothing has exercised
it. The 2026-05-27/17:10 manual run left `awaited_requests` stuck in `waiting`
(orchestration `c09c94a1…`) and an orphaned `running` instance (`60d89697…`),
which is the reply-not-matched signature. Diagnose before re-firing: what was
`c09c94a1` and what `await_responses_topic` did its provision dispatch log (§6).
The DB cleanup `UPDATE thunder_instances … decommissioned` does NOT terminate the
Thunder VM — confirm it's actually gone or it keeps billing.

### Update — 2026-06-02 16:12: D4 CONFIRMED live; provision smoke test passed

Fired `gpu-provisioner` standalone (a100/prototyping). Result: the provision await
**resolved** (`awaited` row `processed`, 42.66s latency; instance `running`, ip
216.81.200.234, db row `52996164…`). Decisive evidence for D4: the adapter's reply
went to `system.agent.generic.responses` — the agent's **own**
`ExecutionContext.ResponsesTopic` — while `__parent_responses_topic__` for that run
was `system.generic.responses` (a *different* topic). Pre-D4 code would have
replied to the parent topic → no consumer → hang (the 17:10 signature). It didn't,
so 1049 carries the own-topic derivation. `__my_responses_topic__` was seeded
(`system.agent.generic.responses`), so the §6 fallback never fired.

Caveats this does NOT yet close:
- It ran gpu-provisioner **standalone on the generic worker**, not as a spawned
  child of model-trainer. The launcher in iter_0 is a spawned child; whether the
  child's `__my_responses_topic__` is seeded the same way is verified only when
  iter_0 steps through (watch the first child await; launcher advancing past
  `presign_dataset` is the proof).
- The 17:10 hang was almost certainly the pre-D4 chassis (old image); now resolved
  on 1049.
- Cost gate is open only because `estimated_new_run_cost_usd=2`; a ~$18 real run
  exceeds `daily_cap_usd=15`. Raise both before iter_0.
- Test instance `ikbj4ogi`/`52996164…` left RUNNING (2h reaper) — decommission for
  real (adapter `decommission_instance`, not a DB UPDATE).

### Update — 2026-06-02 16:32: iter_0 fired; spawned-child seeding CONFIRMED; blocked at call_data_preparer

Fired iter_0 → model-trainer (orch `23863e2e`, corr `0f9ee0f6`). All three spawns
resolved: data_preparer `c29bed97` (16:32:42), provisioner `33989ad4` (16:32:58),
launcher `21b7653e` (16:33:15), each `initialized:true`. Spawned-child seeding
[verified-db/source] confirmed: `spawn_provisioner` created child topics with
`parent_responses_topic: system.agent.generic.responses`, pre-registered the await
on that topic, and the spawned gpu-provisioner pod replied to its `initialize` on
`system.agent.generic.responses` (offset 405); the parent matched `5c6e21be`. So a
spawned child's reply lands on the parent's own ResponsesTopic and the parent
consumes it.

Still NOT exercised: the launcher's **adapter** dispatch reply path (presign /
ssh_exec, the §6 `await_responses_topic` check) — the chain didn't reach
`call_launcher`. The §6 caveat is therefore still open.

Blocked: chain reached `call_data_preparer` (first `call_agent`) and hit an
`error_`-prefixed `debug_collected_data` dump at 16:33:15.520, ~0.5s after the
launcher spawn resolved. No `call_provisioner`/`call_launcher` followed. Cause
PENDING the actual error line (not yet captured).
- Leading hypothesis [assumed, unconfirmed]: `call_data_preparer.input_mapping`
  requires `input_data.orchestration_id` and `input_data.triggered_by`, but the
  trigger supplied only `export_id` + `hyperparameters`; neither field carries the
  `?` optional marker (`call_launcher` uses `?` for its optionals). If
  `ResolveInputMapping` is strict on missing sources, it fails here. The sub-second
  timing fits parent-side resolution, not a child round-trip.
- Disambiguate before any fix: (1) chassis error line after the dump; (2) preparer
  pod logs (`job/agent-training-data-preparer-c29bed97`) — did it get the request;
  (3) `model_lifecycle.training_runs` — a `pending` row = preparer ran (preparer-side
  error); no row = errored before dispatch (parent-side resolution).
- This run provisioned NO Thunder VM (`call_provisioner` never reached). Three
  spawned chassis pods (`c29bed97`/`33989ad4`/`21b7653e`) idle until 3600s timeout.

### Update — 2026-06-02 16:55: root cause CONFIRMED; fix written (103)

Root cause [verified-log/source/db], no longer a hypothesis. Full chassis trace:
`call_agent.go:962 extractDataForAgent` → `:969 Using explicit input_mapping
mapping_count:4` → `content_search.go:96 Path part not found part=orchestration_id
available_keys=[hyperparameters,export_id]` → `coordinator.go:1534 input_mapping
failed: source path 'input_data.orchestration_id' not found for field
'orchestration_id'`. call_agent found the target preparer by role FIRST
(`call_agent.go:298`, topic `…training-data-preparer-spawn_data_preparer.requests`)
and only then failed building the child input_data, so NO work request was sent.
Preparer logs mirror it: got `initialize` (input_data `{}`), replied initialized,
logged "now starting agent's own workflow", then "No activity for 5 minutes" until
its topic was torn down (16:50). `training_runs` newest row 2026-05-12 — no INSERT.
Error couldn't propagate (`Cannot propagate error to parent - missing
replyToRequestID`, parent_topic system.generic.responses) so the orchestration
dead-ended.

[DECISION D5] Fix at the workflow-contract layer (align with the action's spec),
NOT by jamming values into the trigger. The preparer's `PrepareTrainingDataInputSpec`
already lists `triggered_by`/`orchestration_id` as Optional; `training_runs`
columns are nullable UUIDs; `ExtractActionInputs` is spec-driven (Optional →
skip-if-absent) and sources `export_id` from `input_data.export_id`. So mark the
two mappings optional with `?` (call_launcher convention). Verified `?` semantics
in `input_mapping.go` L101-128 (suffix on the dest KEY; TrimSuffix'd into the
child field; absent+optional → skip, absent+required → hard error).

Fix written: **`/mnt/user-data/outputs/103_call_data_preparer_optional_inputs.sql`**.
In-place `jsonb_set` on `default_config.workflow.steps.call_data_preparer.config.
input_mapping` scoped to id `94f5a069-…` (verified: call_data_preparer lives only
in `default_config`, col 6). New mapping: `{export_id, hyperparameters,
orchestration_id?, triggered_by?}`. No version bump (stable id, picked up next
orchestrate). CAVEAT: if the chassis caches agent_definitions, a rollout may be
needed — confirm on the next run.

Next-step preflight [verified-backup]: the preparer's own `prepare_training_data`
step already has `bucket: personae-model-training` and `s3_key_template:
finetuning/datasets/{export_id}/training.jsonl` (a stale `s3_bucket: finetuning`
is overridden by `bucket`), so after the fix the preparer has what it needs — it
won't trip on a missing s3_key_template. Note: the preparer's step ALSO carries an
`input_mapping` (no `?`), but it is INERT — `ExtractActionInputs` doesn't consult
it (it uses input_data nested sources + the spec), so it was left untouched.

After applying 103, before re-firing: (1) settle the cost gate — the chain now
reaches `call_provisioner` and provisions a real a100 for the ~9h run (estimate→
~$20, daily_cap→~$30, re-check can_provision); (2) decommission the leftover
standalone-test instance `ikbj4ogi` (real decommission_instance, not a DB UPDATE).
Then re-fire; the still-unverified piece is the LAUNCHER adapter-dispatch reply
path (the §6 `await_responses_topic` check), reached only at `call_launcher`.

### Update — 2026-06-03 15:31: 103 applied to WRONG DB (clients_db); target is templates_db

Re-fired iter_0 (orch `f5f4a79f`, corr `7cf3f83b`) after redeploy. Failed again at
`call_data_preparer`, now on `triggered_by` (was `orchestration_id`) — purely Go
map-iteration randomness; both keys are still required. The coordinator's logged
`step_config.input_mapping` still has NO `?` on either key, and the pod was fresh
(`agent-chassis-5c488fcc58-qdvts`), so it's not a cache — the loaded row lacks the fix.

Root cause [verified-doc]: `agent_definitions` source of truth is **`templates_db`**
(`002_system_architecture.md` "The source of truth is agent_definitions in
templates_db"; `011_database_and_infrastructure.md` → `postgres-templates-0`,
`templates_user`/`templates_db`, 5432 direct / 6432 pgbouncer). 103 was applied to
`clients_db` (the session used for `model_lifecycle.training_runs`), a different DB.
The definition load is `... ORDER BY version DESC LIMIT 1`; run loaded id
`94f5a069` v1 (the only model-trainer), so it's not a version issue — just the
wrong database.

ACTION: apply 103 to templates_db —
`kubectl -n ai-persona-system exec -it postgres-templates-0 -- psql -U templates_user -d templates_db`
then run 103 (verify-only SELECT first; after-SELECT must show `orchestration_id?`/
`triggered_by?`). If it already shows `?` there, the long-lived chassis is caching
the definition → restart the chassis deployment. No redeploy needed for the DB edit;
a fresh orchestrate reads templates_db live. NOTE for all future agent_definition /
default_config migrations: target templates_db, NOT clients_db.

### Update — 2026-06-03 15:45: 103 applied to templates_db, CONFIRMED

Applied 103 to templates_db; before/after SELECT confirmed the flip (UPDATE 1):
before `{export_id, triggered_by, hyperparameters, orchestration_id}` →
after `{export_id, triggered_by?, hyperparameters, orchestration_id?}`. The earlier
clients_db apply was the no-op. New run fired (orch `ff041320`, corr `a5bede0e`);
spawns completed again (launcher `a5f9d73e` initialize reply consumed on
system.agent.generic.responses 15:45:45). call_data_preparer outcome not yet
captured — the tell is its logged step_config showing `orchestration_id?`/
`triggered_by?` (vs the bare names on the pre-fix runs). If this run predated the
COMMIT it fails identically; re-fire. Downstream gates: `ikbj4ogi` was a 2h-reaper
instance from 06-02 16:12 → almost certainly already reaped (verify tnr status);
cost gate at $2 est/$15 cap will likely allow provisioning (pre-provision gate, not
a mid-run kill; 18h reaper covers the ~9h train).

### Update — 2026-06-03 15:47: 103 WORKS; chain reaches call_launcher; blocked on provisioner output shape

103 confirmed end-to-end through the two steps that mattered (run orch `ff041320`,
corr `a5bede0e`):
- `call_data_preparer` SUCCEEDED: preparer exported 1958 rows (~21.5MB) to
  `s3://personae-model-training/finetuning/dataset…`, inserted training_runs row
  `27ef7bea-862c-4af0-9d4f-e9039185787e`. (15:46:52)
- `call_provisioner` SUCCEEDED (15:47:41): a REAL a100 was provisioned
  (provisioner `1fb11543`). Provision row has `attached_to_training=true`
  (training_run_id passed) → ~18h reaper, NOT 2h. Since call_launcher then failed the
  box is an IDLE PAID ORPHAN → decommission immediately (thunder_instances newest row
  → its `id` = provisioning_id → adapter `decommission_instance`, not a DB UPDATE).
- `call_launcher` FAILED [verified-log]: `input_mapping failed: source path
  'provisioning_result.provisioning_id' not found`. First time call_launcher reached.

Root cause: the launcher mapping reads `provisioning_result.provisioning_id` (the
intended contract — `provision_action.go ProvisionInstanceResult` field names are
documented to match `provisioning_result.*`). resolveSourcePath auto-unwraps ONE
`.response` (why `preparation_result.dataset_uri` works — preparer returns a FLAT map).
But the provisioner result is NOT flat:
`provisioning_result.response.dispatch_provision.response.<ProvisionInstanceResult>`
plus `provisioning_result.response.input_data`. provisioning_id sits behind the
`dispatch_provision` step-name key, which auto-unwrap won't cross (it unwraps
`.response`/`result`, not arbitrary step names).

[DECISION] Fix on the PROVISIONER (flatten its output so `provisioning_result.response`
IS the provision response), NOT the launcher mapping (pointing the launcher at
`…dispatch_provision.response.provisioning_id` would couple it to the provisioner's
internal step name). The backup's gpu-provisioner `complete` uses singular
`output_field: provision_response` with a comment that this makes
`provisioning_result.instance_ip` resolve via auto-unwrap — the intended flat shape —
but the LIVE result shows `dispatch_provision`/`input_data`, so the deployed def has
DIVERGED from the backup. PENDING: dump live gpu-provisioner workflow from templates_db
before writing the fix (do NOT trust the backup — same lesson as clients_db/templates_db).
Expected fix: small jsonb_set making `complete` emit the single dispatch output
(singular output_field) so the wrapper drops and `provisioning_result.provisioning_id`
resolves; launcher mapping unchanged. Apply to templates_db.

### Update — 2026-06-03 17:07: orphan decommissioned; pre-next-run checklist

- ORPHAN GONE: fired `decommission_instance` (provisioning_id `40811b3e-fc82-4aa4-a96c-d344737f7bd4`,
  thunder_instance_id `0`, ip 216.81.200.234, up since 15:47:39) via the adapter; `tnr status`
  now reports no instances. Cost stopped. (`attached_to_training` is a LOG field, not a
  thunder_instances column — the reaper backstop is `max_uptime_hours`.)
- VERIFIED [schema]: training provisions (TrainingRunID set) get `cfg.DefaultHardUptimeHours`
  = **18** (schemas_all; Thunder enforces server-side). ~9h train sits under it → reaper won't
  kill mid-train. No change needed. (Confirms the orphan would've billed ~18h.)

PRE-NEXT-RUN CHECKLIST:
1. BLOCKING — gpu-provisioner flatten fix. Awaiting live `default_config #> '{workflow,steps}'`
   dump from templates_db (do NOT use backup — diverged). Without it the next run re-fails at
   call_launcher AND orphans another a100. After fix, all 7 launcher fields resolve.
2. Free the recycled Thunder id: `tnr status` empty → next provision likely gets id `0` again
   (the orphan's id). Partial unique index `thunder_instances_live_identifier_uniq` on live rows
   → if the `40811b3e` row didn't flip to `decommissioned`, next provision INSERT fails dup-key
   (23505). Verify no row in (provisioning/running/decommissioning); reconcile `40811b3e` if stuck.
3. Cost gate: every prior run died pre-training, so the gate is untested against a full ~9h run
   (~$18, not ~$2). Re-check `thunder_provision_check.can_provision` + `thunder_config.daily_cap_usd`;
   bump cap to cover ~$18 + today's spend if needed. (cost-gate tables are in clients_db.)
Optional hygiene: mark leftover training_runs `27ef7bea` cancelled (launcher never ran; next run
makes a fresh row); idle spawned chassis pods self-clean at 3600s.

### Update — 2026-06-03 17:2x: gate + id checks pass; gpu-provisioner dump was wrong DB again

- COST GATE — funded, NO change. `thunder_provision_check`: can_provision=t, no denial,
  is_paused=f; daily_cap_usd=30, estimated_new_run_cost_usd=20, default_hourly_rate_usd=1.80,
  default_hard_uptime_hours=18, max_concurrent=2, total_24h_spend=0.57, active_count=0.
  0.57+20=20.57 < 30 → clears with ~$9 headroom; estimate already realistic for ~9h.
- RECYCLED ID — free, NO change. 0 rows in (provisioning/running/decommissioning) → the
  `40811b3e` row reconciled to decommissioned; Thunder id `0` released → next provision
  won't dup-key.
- POST-TRAIN OPS (no monitor): after train.log shows completion (~9h), MANUALLY decommission —
  nothing auto-tears-down, so the box idles to the 18h reaper (~$16 wasted, and 0.57+~32 would
  blow the $30 cap, though only matters if another provision followed).

**gpu-provisioner workflow dump was run on clients_db (WRONG) — `templates_db` still needed.**
agent_definitions is a SEPARATE physical table in BOTH clients_db and templates_db (103 proved
the chassis reads templates_db; the clients_db copy is inert). The clients_db copy shows the
clean `output_field: provision_response` shape, but the LIVE runtime result was
`provisioning_result.response.{dispatch_provision, input_data}` (keyed by step name + input_data
echo) — which a provision_response-only `complete` would NOT emit (cf. endpoint-health-check
`complete` keying by field name `health_result`). So the live templates_db def has DIVERGED from
the clients_db/backup copy; the flatten fix must be written against the templates_db dump:
`kubectl -n ai-persona-system exec -it postgres-templates-0 -- psql -U templates_user -d templates_db -c "SELECT jsonb_pretty(default_config #> '{workflow,steps}') FROM agent_definitions WHERE type='gpu-provisioner' AND is_active=true ORDER BY version DESC LIMIT 1;"`
PIN: for agent_definitions, always read AND patch templates_db, never clients_db.

### Update — 2026-06-03 ~17:3x: CORRECTION — chassis reads clients_db, NOT templates_db; call_launcher root cause found

**DB DIRECTION CORRECTED (supersedes the earlier templates_db pin above).** The chassis
reads the flywheel-C / rich-schema agent_definitions from **clients_db**, not templates_db.
Proof [verified-db]: templates_db.agent_definitions has the OLD schema (NO `version` column —
`ORDER BY version` errors there) and holds only the 8 original website-builder agents
(2025-08-21). The chassis loader query is `... AND is_active ... ORDER BY version DESC LIMIT 1`
filtering `is_snapshot` — columns that exist ONLY in clients_db's rich schema, so it cannot
run against templates_db. gpu-provisioner (`0bf9fa8a`, v1.0.1051) + model-trainer + the whole
chain live in clients_db. 103 worked because it landed on clients_db (where model-trainer is),
NOT because of a templates_db apply — the earlier "apply to templates_db" guidance was wrong.
PIN (corrected): for the flywheel-C agent_definitions, always read AND patch **clients_db**.
(The 002_system_architecture.md "templates_db" line refers to the old website-builder catalog.)

**call_launcher root cause [verified-source].** `extractWorkflowResult` (the fn that builds an
agent's final result) reads `completeStep.Config["output_fields"]` — PLURAL only. The
gpu-provisioner `complete` uses `output_field` (SINGULAR), which is never read → falls to the
fallback branch that dumps every non-internal collected key → provision result comes out as
`{dispatch_provision, input_data}` (step-name-keyed), with provisioning_id buried at
`…dispatch_provision.response.provisioning_id`. The backup comment claiming singular output_field
yields a flat `provisioning_result.instance_ip` describes behavior the chassis does NOT implement.
`ExtractNestedField` confirms: literal-path descent + one-level `.response` unwrap per part, no
global search — can't pluck a buried field to a clean top-level key.

FIX (two layers):
- Proper/structural (chassis, deferred): make extractWorkflowResult honor singular `output_field`
  as "return that field's value as the body." Fixes gpu-provisioner AND thunder-reaper (same
  singular key); launcher mapping stays the documented `provisioning_result.provisioning_id`.
  Needs Go change + chassis rebuild/redeploy + guard for reaper's absent `reaper_summary`. NOT now.
- Targeted/immediate (DB migration, clients_db): point the launcher's 4 provisioning fields at
  `provisioning_result.dispatch_provision.<field>` — resolves via the resolver's per-part
  `.response` auto-unwrap (same mechanism that makes `preparation_result.dataset_uri` work).
  Couples launcher to the provisioner's `dispatch_provision` step name (the tradeoff). Revert
  once the chassis fix lands.
PENDING: dump live call_launcher input_mapping from clients_db (id 94f5a069,
`default_config #> '{workflow,steps,call_launcher,config,input_mapping}'`) to write the exact
jsonb_set without guessing `?`/key names.

### Update — 2026-06-03 ~17:4x: chassis change vetoed; fix at launcher mapping only

User decision: do NOT change `extractWorkflowResult`. `output_fields` (plural) is the
well-used standard the chassis honors; the gpu-provisioner is the non-compliant one
(singular `output_field` → falls into the fallback dump). So fix at def/mapping level.

CHOSEN FIX (minimal, no redeploy): repoint the launcher's 4 provisioning fields in the
model-trainer `call_launcher` input_mapping from `provisioning_result.<field>` to
`provisioning_result.dispatch_provision.<field>` (provisioning_id required; instance_ip,
ssh_user, ssh_key_secret_name — needed for ssh_exec even though `?`-optional in the map).
`dispatch_provision` is the only key the provision result lands under in the fallback dump.
Single jsonb_set on clients_db (id 94f5a069), launcher mapping path
`default_config #> '{workflow,steps,call_launcher,config,input_mapping}'`.

DEFERRED cleanup (not now): switch gpu-provisioner `complete` from singular `output_field`
to standard `output_fields: ["dispatch_provision"]` — brings it onto the standard, but
changes the result shape, so re-point the launcher at the same time.

PENDING before writing the jsonb_set (to avoid a wrong-path run that orphans a real a100):
(1) live call_launcher input_mapping (exact keys + `?` placement);
(2) exact literal depth of provisioning_id under dispatch_provision — from
`platform/datahelpers/content_search.go` (FindByPath/path resolver) OR a grep of the saved
trace `trace_a5bede0e*.json`. Prefer writing a literal path GetValueAtExactPath resolves
outright over relying on the resolver's `.response` auto-unwrap through an intermediate key.

### Update — 2026-06-03 ~17:5x: 104 written — gpu-provisioner output_fields + launcher re-point

User chose the structural fix (fix the gpu-provisioner, not just patch the launcher).
104 (clients_db, in-place, no version bump → no restart) does two edits:
1. gpu-provisioner (0bf9fa8a) complete.config: `{"output_field":"provision_response"}` →
   `{"output_fields":["dispatch_provision"]}`. extractWorkflowResult honours only plural
   `output_fields`; the provision result lands under the STEP NAME `dispatch_provision`
   (await storage keys by step name — `provision_response` is never a collected key, per
   runtime). Drops the input_data echo too.
2. model-trainer (94f5a069) call_launcher.input_mapping: 4 provisioning fields repointed
   `provisioning_result.<field>` → `provisioning_result.dispatch_provision.<field>`. Resolve
   via the same `.response` auto-unwrap that makes `preparation_result.dataset_uri` work
   (dispatch_provision = immediate child after unwrapping provisioning_result.response);
   robust whether or not ExtractStepData strips an inner `.response`, and order-independent
   with edit #1. dataset_uri/training_run_id/hyperparameters unchanged.
Live call_launcher mapping keys confirmed from clients_db dump (7 keys, `?` on the 3
optional provisioning fields). Optional pre-fire check: grep trace_a5bede0e* for the
provisioning_id path to confirm depth before firing (saves an orphan-a100 adjust-run);
if provisioning_id sits at `…dispatch_provision.response.provisioning_id`, add `.response`
to the 4 launcher paths. gpu-provisioner dispatch step keeps output_field provision_response
(cosmetic for await results) — left minimal.

### Update — 2026-06-03 ~18:0x: docs updated; 104 validated against guidelines

Per request, updated (in /mnt/user-data/outputs):
- RUNBOOK_iter0_pretrigger.md → new "## 2b. Data-path verification (input_mapping ↔
  producer result shape)": before firing, confirm each call_* step's mapped source
  paths resolve against the producing step's REAL collected_data shape (dump
  orchestration_states.collected_data or grep the trace), since a wrong path fails only
  after an a100 is booked. References 104 + gotcha #23.
- 016_debugging_guide_v2_30.md (copied from v2_29) → (a) new discipline gotcha #23:
  child result shape = `complete.output_fields` (plural); singular `output_field` is
  ignored → fallback dump keyed by step name (producer-side twin of #2/#12); (b) new
  subsection under the `input_mapping failed: source path … not found` entry covering
  the "field exists but producer shaped it wrong" root cause + the two-layer fix.

104 vs guidelines (003 contracts / 001 dev guide / 016): CONFORMS. `?` on the 3 optional
launcher keys, provisioning_id required (016 #2/#12); uses the documented jsonb_set-on-
input_mapping fix pattern; reuses output_fields (no new Go); no subworkflow; targets
clients_db. The gpu-provisioner half REMOVES a latent violation (it was non-conformant
using singular output_field). One accepted tradeoff vs the dev guide's decoupling spirit
(001 L362, find-by-role): the launcher now reads through the provisioner's
`dispatch_provision` step name — unavoidable for data paths (no find-result-by-role
exists); the fully-decoupled version is the deferred chassis change.

### Update — 2026-06-03 ~18:04: 104 CONFIRMED live — FIRST full chain end-to-end; then VM-side /workspace failure

Re-fired with 104 applied (orch `cd906623`, launcher orch `b002b359`, corr `40ff2d91`). **104 worked.** call_launcher resolved — the launcher (agent `a0eb4a2f`) received all 7 fields populated (instance_ip 216.81.200.234, provisioning_id `1634d7f8`, ssh_user ubuntu, ssh_key_secret_name thunder-ssh-1634d7f8, dataset_uri, training_run_id `e6ab9fad`, hyperparameters) and ran its WHOLE workflow — the previously-unverified stretch, all confirmed live: presign_dataset → 200; presign_scripts → 200; ssh_exec_launch → SSH connected attempt 1, exit_code 0; mark_running → `e6ab9fad`→running; complete → returned launch_result via `output_fields:["launch_result"]` (extractWorkflowResult "using configured output_fields", field_count 4 — clean, no fallback dump); parent notified on system.agent.generic.responses. D4 + adapter prepare_object_url + ssh_exec all validated live for the first time.

**BUT training did NOT start (VM-side).** ssh_exec_launch stderr = `mkdir: cannot create directory '/workspace': Permission denied` + `bash: line 1: /workspace/launch.log: No such file or directory`; stdout `LAUNCH_PID=193`. The command_template did a *plain* `mkdir -p /workspace` (no sudo); `ubuntu` can't create at `/`. The `&`-backgrounded setsid's `> /workspace/launch.log` redirect then failed → the job died immediately → no training. exit_code 0 only because the command's last token is `echo` (the known detached-ssh_exec false-success — gap #1 in §9). a100 `1634d7f8` (Thunder `tvlqb4u1`) provisioned but idle; `e6ab9fad` flipped to running over nothing.

Confirmed against the bundle scripts (the fix direction follows from them): run.sh L33 `WORKSPACE="/workspace"` (DATA/OUT/`cd`); 00_vm_setup.sh L9 `WORKSPACE="/workspace"` AND **L51-52 `sudo mkdir -p "${WORKSPACE}"` + `sudo chown "$(id -u):$(id -g)" "${WORKSPACE}"`** — so /workspace IS the bundle's convention, created with sudo+chown. The launcher command_template diverged (plain mkdir) AND runs its curls into /workspace before 00_vm_setup runs. 02_train is arg-driven (no change). Decommissioned `1634d7f8`; `thunder_instances` live set confirmed empty (0 rows) before re-fire.

### Update — 2026-06-03 ~18:29: 105 CONFIRMED — ssh_exec clean; setup reached CUDA/torch/Unsloth; died on root-owned ~/.bashrc (set -e)

**105 written/applied to clients_db/CONFIRMED** (BEFORE/AFTER verified, UPDATE 1): training-launcher `ssh_exec_launch.command_template` → `sudo mkdir -p /workspace && sudo chown $(id -u):$(id -g) /workspace; setsid ...` — mirrors 00_vm_setup L51-52. No re-bundle needed for this (run.sh/00_vm_setup keep /workspace; 00_vm_setup's own sudo-mkdir is now idempotent). sudo is known-good (whole setup script uses it); /workspace-on-root has the 100GB (the manual run used it). No version bump; **chassis loads the def per-orchestrate — CONFIRMED (no restart between applying 105 and re-fire, and the new run used the new template).**

Re-fire corr `31328232`, parent orch `bd702ffe`: preparer `6bb1ad73` → 1958 rows, NEW training_run `1cd65dd7-ad74-4f0d-b509-75f821c29d46`; provisioner → NEW a100 `fabfd7fa-ac84-4476-86f3-f7ac57862214` (Thunder `5azm9fpe`, 216.81.200.234:**30340**, 18h reaper); launcher `ced41059` → both presigns 200, **ssh_exec_launch CLEAN** (stderr `""`, exit_code 0, `LAUNCH_PID=197` — empty stderr where the prior run had the two /workspace lines), mark_running → `1cd65dd7`→running, complete. So 105 fixed the launch.

**Watched train.log directly** (SSH: key in secret `thunder-ssh-fabfd7fa-...`, field `private_key`, base64 — `... -o jsonpath='{.data.private_key}' | base64 -d`; needs `chmod 600`; and `-o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null` because Thunder recycled IP 216.81.200.234 → stale known_hosts). Setup got THROUGH the heavy part: torch 2.10.0+cu128, device NVIDIA A100-SXM4-80GB, vram 79.2 GiB, bf16 True; Unsloth 2026.6.1 import OK. **→ the `base` template DOES carry CUDA** (00_vm_setup's `nvidia-smi` guard passed) — earlier worry resolved.

**Died at 00_vm_setup.sh line 153** — `echo 'export HF_HUB_ENABLE_HF_TRANSFER=1' >> "${HOME}/.bashrc"` → `/home/ubuntu/.bashrc: Permission denied`. On this image `~/.bashrc` is root-owned 644 (home DIR is writable — the venv landed in it fine — but the pre-seeded file is not). `set -e` aborted 00_vm_setup at the LAST, cosmetic step (step 9, *after* torch/Unsloth installed+verified) → run.sh (also `set -e`) aborted → training never reached smoke/full (no `RUN_SH_STEP step=smoke` in train.log). This is the same class as the /workspace one: a best-effort step made fatal by `set -e`.

Fix (structural): 00_vm_setup.sh step 9 made best-effort — `if [ -w "${HOME}/.bashrc" ] && ! grep -q HF_HUB_ENABLE_HF_TRANSFER "${HOME}/.bashrc" 2>/dev/null; then echo ...; fi`. Patched file in /mnt/user-data/outputs/00_vm_setup.sh (`bash -n` OK); needs re-bundle + re-upload to `finetuning/scripts/bundle.tar.gz` for future automated runs. Salvage for the CURRENT box (paid-for, setup idempotent + 90% done): SSH one-liner — sudo-seed the HF var to .bashrc so setup's `grep` short-circuits past line 153, then re-launch run.sh detached (setsid; files already in /workspace from the launch, so run.sh's existence check passes). Re-tail to `RUN_SH_SMOKE_OK` → full.

PENDING after this: watch train.log to `RUN_SH_FULL_OK`/`RUN_SH_DONE` (smoke gates the ~9h full; status flip stays untrustworthy — `1cd65dd7` already reads running); decommission `fabfd7fa` after DONE/FATAL (no auto-teardown, 18h reaper); reconcile `e6ab9fad` (still 'running' from the 18:03 attempt → mark failed). Once a run trains clean: fold gotchas into runbook/debug guide — (a) the command_template must stand up its own workspace the way 00_vm_setup does (sudo mkdir+chown), and (b) detached ssh_exec exit-0 masks VM-side failure, plus the corollary that any best-effort VM step under `set -e` becomes fatal (guard it).

### Update — 2026-06-03 ~19:1x: bundle re-uploaded (fixed) + box salvaged — iter_0 SMOKE training RUNNING

Bundle fix shipped to B2 and verified: `finetuning/scripts/bundle.tar.gz` re-uploaded (size 7548, MD5 a44aa32a…), confirmed FLAT (run.sh, 00_vm_setup.sh, 02_train_llama_3_3_70b.py, 03_inference_test.py at archive root) and the step-9 `.bashrc` guard present. So future automated provisions pull the fixed bundle — no DB/chassis change (the agent def only holds the key; re-uploading the object IS the whole deploy).

Current box `fabfd7fa` (216.81.200.234:30340) SALVAGED rather than discarded (setup was 90% done + paid-for). Pushed the fixed `00_vm_setup.sh` onto it and re-launched run.sh detached. **Verified live training is now running** (pgrep + train.log tail): `bash /workspace/run.sh` → `python 02_train … --output /workspace/smoke_out --limit 20 --epochs 1` (the SMOKE pass) + torch inductor compile workers. Log: Unsloth 2026.6.1, Num examples=20, Total steps=3, Trainable params 207,093,760 of 70,760,800,256 (0.29%), "Starting training", step 1/3 (~170s/it on the first step — includes model load/compile). So the `.bashrc` guard worked (setup ran through, venv activated, smoke started); smoke (3 steps) gates the ~9h full run.

SSH commands used (key + state-check + salvage + watch — `scp` uses `-P`, `ssh` uses `-p`; `StrictHostKeyChecking=no` because Thunder recycles the IP → stale known_hosts):

```
# key from the per-instance secret (field private_key, base64); chmod 600
SECRET=thunder-ssh-fabfd7fa-ac84-4476-86f3-f7ac57862214
kubectl -n ai-persona-system get secret "$SECRET" -o jsonpath='{.data.private_key}' \
  | base64 -d > /tmp/k && chmod 600 /tmp/k

# state-check: is run.sh/training running? what does the log say?
ssh -i /tmp/k -p 30340 -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null \
  ubuntu@216.81.200.234 'pgrep -af "run.sh|02_train|[p]ython"; echo ---; tail -n 8 /workspace/train.log'

# salvage an idle box that has the OLD on-disk bundle: push the FIXED setup script,
# then re-launch run.sh fully detached (setsid; files already in /workspace from launch)
scp -i /tmp/k -P 30340 -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null \
  00_vm_setup.sh ubuntu@216.81.200.234:/workspace/00_vm_setup.sh
ssh -i /tmp/k -p 30340 -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null \
  ubuntu@216.81.200.234 \
  'setsid bash -c "/workspace/run.sh > /workspace/train.log 2>&1" < /dev/null > /workspace/launch.log 2>&1 & echo LAUNCH_PID=$!'

# live-watch the run.sh markers
ssh -i /tmp/k -p 30340 -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null \
  ubuntu@216.81.200.234 'tail -f /workspace/train.log'
```

PENDING: watch for `RUN_SH_SMOKE_OK` → `RUN_SH_STEP step=full_train` → (full 3-epoch run on 1958 rows, ~9h) → `RUN_SH_FULL_OK output=/workspace/adapter_out` → `RUN_SH_DONE`; status flip stays untrustworthy (`1cd65dd7` already reads running). Decommission `fabfd7fa` after DONE/FATAL (no auto-teardown, 18h reaper). `e6ab9fad` still 'running' from the 18:03 attempt → mark failed. Once `adapter_out` lands: the iter_0 milestone (first automated training launch end-to-end → trained adapter) is reached; then fold the two confirmed VM-side gotchas into runbook/debug guide — (a) command_template must stand up its own workspace (sudo mkdir+chown) like 00_vm_setup, (b) detached ssh_exec exit-0 masks VM failure + any best-effort VM step under `set -e` becomes fatal (guard it).

### Update — 2026-06-04 ~11:50: full run overran 18h cap → reaper SAVE via per-instance max_uptime_hours; training-monitor design + probe started

**Full run is ~24h, not the 30–90 min `run.sh` claims.** At ~119 s/step × 726 steps (3 epochs, 1958 rows) ≈ 24h. Loss is healthy (1.49 → ~0.2 by epoch 1; 3 epochs likely overkill, and FA2=False/xformers is a big slowdown at seq 4096). The smoke rate (116 s/step) predicted this — nobody extrapolated it, and I wrongly propagated `run.sh`'s "30–90 min" into the runbook. The 18h cap < 24h train → the run was structurally doomed to be reaped mid-train, and `save_strategy="no"` (no checkpoints) means a reap = total loss.

**Reaper SAVE (done, verified).** `tnr` can't extend uptime — `tnr modify` has no uptime flag and restarts the instance (which would kill the run). But the cap is OURS, not Thunder's: the Thunder create call carried NO uptime field, and `thunder_instances` has a per-instance `max_uptime_hours` (default 18) + `running_since`, with NO separate deadline column → the reaper computes cutoff = `running_since + max_uptime_hours`. So bumped THIS row: `UPDATE public.thunder_instances SET max_uptime_hours=48 WHERE id='fabfd7fa-…' AND status='running';` → deadline moved 2026-06-04 12:29 → 2026-06-05 18:29; `NOW()`=11:47 UTC → ~30h headroom. Per-instance, reversible, no restart, no Thunder call. (Also learned: `tnr status` prints LOCAL/BST — "12:24" was 11:24 UTC vs a 12:29 UTC deadline, so ~1h of slack, not 4 min.) MUST still manually decommission once `RUN_SH_DONE` (pushed auto-reap is now 48h → ~$60 if forgotten).

**Decision: build a separate `thunder-training-monitor`, NOT bolted into the time-reaper.** Distinct responsibilities: the time-reaper is the last-line cost backstop and must stay dead-simple/dependency-free (pure DB + Thunder) so it still works if the adapter is down; the monitor depends on the adapter + SSH. They overlap only in calling `decommission_instance` (shared action — reuse, not duplicate). The launcher returns long before training ends (detached run), so the monitor can't be a workflow step that awaits — it's a periodic poller (scheduler-triggered, ~5–10 min), spawning a sub-agent per active training instance (no subworkflow). It closes TWO current gaps: (1) release a detached box on completion, (2) the `running → complete/failed` reconcile that never happens today — exactly why `1cd65dd7` and `e6ab9fad` are stuck `running`.

**Reuse audit before building (per the reuse-first rule):**
- Adapter PROBE: `ssh_get_status` ALREADY EXISTS (handleSSHGetStatus → SSHExecAction.GetStatus) — takes an optional arbitrary `status_command`, returns `{provisioning_id, exit_code, stdout, stderr, reachable}`, and crucially returns `reachable:false` as a VALID answer (not an error) when the box is down. That is the probe primitive; no adapter change. (The user's "existing status/poll action to extend" — it already does what we need.)
- Chassis DISPATCH: there was `dispatch_thunder_ssh_exec`, `dispatch_thunder_decommission`, `dispatch_thunder_provision`, `dispatch_thunder_prepare_object_url` — but NO `dispatch_thunder_ssh_get_status`. So one new dispatch action, a near-clone of the ssh_exec dispatcher (same envelope/publish/AwaitResponse; status_command optional). Written: `thunder_ssh_get_status_dispatch.go` (reuses package-local thunderAdapterTopic/defaultIfEmpty/configOrInput/interpolateCommandTemplate — no redeclaration). Registry line to add next to `dispatch_thunder_ssh_exec`.
- RELEASE: reuse `dispatch_thunder_decommission` (the reaper's path).
- RUN-STATUS write: `MarkTrainingRunRunningAction` is hardcoded `pending→running` (WHERE status='pending') — the launcher depends on it, so DON'T mutate it; the terminal `running→complete|failed` write will be a NEW sibling action, same file shape.

**Probe `status_command` (grounded in the real run.sh markers — RUN_SH_START/STEP/SMOKE_OK/FULL_OK/DONE on success, RUN_SH_FATAL only pre-train; a mid-train crash leaves NO marker):**
```
pgrep -af '02_train_llama_3_3_70b|/workspace/run.sh' >/dev/null 2>&1 && { echo STATUS=ALIVE; exit 0; }; if grep -q RUN_SH_DONE /workspace/train.log 2>/dev/null && [ -f /workspace/adapter_out/adapter_config.json ]; then echo STATUS=DONE_OK; elif grep -q RUN_SH_FATAL /workspace/train.log 2>/dev/null; then echo STATUS=DONE_FAIL; else echo STATUS=GONE_UNKNOWN; fi; tail -n 3 /workspace/train.log 2>/dev/null
```
Verdicts → monitor action: `ALIVE` leave it; `DONE_OK` (RUN_SH_DONE + adapter_config.json present) → training_runs→complete + decommission; `DONE_FAIL`/`GONE_UNKNOWN` (crash/OOM/aborted with no end marker) → training_runs→failed + decommission; `reachable:false` for several consecutive cycles → treat as `lost` and decommission anyway (an unreachable box still bills). `GONE_UNKNOWN`=total-loss case reinforces the checkpointing fix (`save_strategy="no"`).

NEXT (reasonable steps, in order): (1) ✅ DONE — classifier action `classify_training_probe` written (`classify_training_probe_action.go`): pure logic, no DB; reads the prior `ssh_get_status` step's `.response.{stdout,reachable}` (NOTE: `ExtractNestedFieldString` returns "" for a bool, so `reachable` is read via `ExtractNestedField` + type-assert), parses the `STATUS=` token, and routes via a `next_step` override (branching is supported — `getNextStepFromResult`, coordinator L1091; an action returning `{"next_step": ...}` overrides the configured one). Verdicts alive/done_ok/done_fail/gone_unknown/unreachable/no_status → config-named target steps (complete_step/failed_step/alive_step/unreachable_step), keeping it decoupled from specific step names. Testable NOW against the running box via a one-off `ssh_get_status` with the probe `status_command` (expect `STATUS=ALIVE`; `STATUS=DONE_OK` in ~9h). **Deferred (needs cross-tick state):** "unreachable for N consecutive ticks → lost → decommission" — a single probe can't count ticks and `thunder_instances` has no consecutive-unreachable counter; for now unreachable = leave for next tick, time-reaper (per-instance `max_uptime_hours`) backstops a truly-lost box. Would need a small schema add (counter column) to honor the user's "several cycles" ask. (2) terminal `mark_training_run_terminal` action (running→complete|failed via config `status`), mirroring `MarkTrainingRunRunningAction`'s DB idiom — NOTE `markTrainingRunFailed` already exists but is an UNEXPORTED helper in `prepare_training_data_action.go` (the preparer's error path), NOT a registered workflow action, so the monitor needs a new registered one. (3) `find_active_training_instances` query action (thunder_instances WHERE status='running' AND training_run_id IS NOT NULL). (4) monitor agent definition (thin workflow: find → per-instance spawn → probe → classify → mark_complete|mark_failed → decommission → done) + scheduler wiring. Reuses `dispatch_thunder_ssh_get_status` (new, written) + `dispatch_thunder_decommission` (existing). Static-only so far (no Go toolchain here) — `go build ./...` pending in-repo.

PENDING (unchanged): watch `train.log` to `RUN_SH_FULL_OK`/`RUN_SH_DONE`; decommission `fabfd7fa` after; reconcile `1cd65dd7` + `e6ab9fad`. Structural fixes still queued: checkpointing (save_strategy="steps"+B2+resume — top priority), cap-sizing from smoke s/step at provision, FA2 + fewer epochs; and correct the runbook "30–90 min" line.

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

Status as of 2026-06-02 in brackets.

1. **Both images on `v1.0.1049`** [DONE for adapter — round-trip proves 1049 has
   `prepare_object_url`; chassis 1049 deployed but the D4 topic fix is unverified
   until a real run]. Agents run on the chassis workers (`agent-chassis`,
   `business-intel`, `vet-intel`), not their own pods.
2. **`go build ./...` + `go vet`** in-repo [the compile gate; still the user's to
   run].
3. **Bundle in B2 + `prepare_object_url` round-trip** [DONE — uploaded 16:17,
   round-trip returned `presigned_url`; verify the GET fetch returns 200 with the
   jq-decoded URL, not `curl -I`].
4. **Cost gate** [OPEN]: `SELECT can_provision, denial_reason FROM
   thunder_provision_check;` after cleanup → expect `t | NULL`. Make
   `estimated_new_run_cost_usd` realistic (~$20 vs the current $2) and raise
   `daily_cap_usd` above one ~$18 run; confirm the orphaned VM is really
   terminated first.
5. **Re-fire provision (`gpu-provisioner` standalone) and read
   `await_responses_topic`** [DONE 2026-06-02 — PASSED: replied to own topic
   `system.agent.generic.responses`, await resolved 42s; D4 confirmed live on 1049.
   Decommission the test instance `52996164…` for real. Spawned-child seeding for
   the launcher is still only verified when iter_0 runs].
6. **Trigger `iter_0`** [BLOCKED→FIX READY 2026-06-02 — root cause confirmed:
   call_data_preparer input_mapping hard-failed on absent input_data.orchestration_id.
   Fix written: 103_call_data_preparer_optional_inputs.sql (marks orchestration_id?/
   triggered_by? optional). Apply 103, then settle the cost gate + decommission
   ikbj4ogi, then re-fire. See the 16:55 update above.]
   `input_data` must carry `export_id 146a9a12-…` and `hyperparameters`.

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
