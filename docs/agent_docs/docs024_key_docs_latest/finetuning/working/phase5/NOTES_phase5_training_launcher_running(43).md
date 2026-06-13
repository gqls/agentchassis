# Running notes — Phase 5 training-launcher (Flywheel C)

Decision + reasoning log for automating the `iter_0` training launch. Append as
we go. Epistemic tags used below: **[verified-source]** read from the actual
code/schema this session; **[verified-db]** confirmed by querying production;
**[deployed?]** changed in code and reported deployed but image tag not yet
re-confirmed; **[assumed]** inferred from surrounding code, not directly proven;
**[gap]** known missing piece.

Last updated: 2026-06-09.

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

NEXT (reasonable steps, in order): (1) ✅ DONE — classifier action `classify_training_probe` written (`classify_training_probe_action.go`): pure logic, no DB; reads the prior `ssh_get_status` step's `.response.{stdout,reachable}` (NOTE: `ExtractNestedFieldString` returns "" for a bool, so `reachable` is read via `ExtractNestedField` + type-assert), parses the `STATUS=` token, and routes via a `next_step` override (branching is supported — `getNextStepFromResult`, coordinator L1091; an action returning `{"next_step": ...}` overrides the configured one). Verdicts alive/done_ok/done_fail/gone_unknown/unreachable/no_status → config-named target steps (complete_step/failed_step/alive_step/unreachable_step), keeping it decoupled from specific step names. Testable NOW against the running box via a one-off `ssh_get_status` with the probe `status_command` (expect `STATUS=ALIVE`; `STATUS=DONE_OK` in ~9h). **Counter step — DONE (this step):** "unreachable for N consecutive ticks → lost → decommission" is now implemented as its OWN step, keeping the classifier pure. Migration `106_thunder_unreachable_counter.sql` (clients_db, public.thunder_instances) adds `consecutive_unreachable_probes int NOT NULL DEFAULT 0` + `last_probe_at timestamptz` (idempotent ADD COLUMN IF NOT EXISTS; the existing updated_at trigger handles updated_at). Action `record_probe_streak` (`record_probe_streak_action.go`): `mode=reset` (reachable/ALIVE → zero counter, route ok_step) / `mode=bump` (unreachable/no_status → counter+1 RETURNING; if ≥ `unreachable_threshold` [default 3] → route `lost_step` = the shared mark_failed→decommission path; else `ok_step` = leave for next tick). The classifier routes alive_step→`reset_streak` and unreachable_step→`bump_streak`. (106 numbering: a project DOC `106_claude_anthropic_skill.md` exists — confirm the next free SQL migration number in the runner.) (2) ✅ DONE — `mark_training_run_terminal` (`mark_training_run_terminal_action.go`): running→complete|failed via config `status`, stamping `completed_at` on both and `error_message` on failed (mirrors the column shape of the existing unexported `markTrainingRunFailed`, which is NOT a registered action). Guarded `WHERE id=$1 AND status='running'` → idempotent, never clobbers a terminal row. `status` read as a config literal via `configOrInput` (ExtractActionInputs resolves paths/keys, not bare literals); `training_run_id` via `ExtractActionInputs` like the running-sibling. Schema-verified: CHECK is `('pending','running','complete','failed')` — literal is `'complete'`. (3) ✅ DONE — `find_active_training_instances` (`find_active_training_instances_action.go`): queries clients_db `public.thunder_instances WHERE status='running' AND training_run_id IS NOT NULL AND decommission_requested_at IS NULL`, returns `{instances:[{provisioning_id,training_run_id,thunder_instance_id,instance_ip}], count}` (id/training_run_id cast ::text for safe scanning; key always present even as `[]`). Fan-out contract = `LoopAction` (loop_actions.go): loop step config `iterate_over: "<find step>.instances"`, `loop_var` (default `loop_item`), `substeps`/`sub_workflow.steps`; resolves the collection by dot-path, accepts []interface{}/[]map/[]string, `max_iterations` default 20 (bump if >20 concurrent training boxes), EMPTY array = graceful skip but MISSING path = error unless `allow_missing` (hence always return the key). Each substep reads `loop_item.provisioning_id` / `loop_item.training_run_id` via input_mapping. (4a) ✅ DONE — worker sub-agent definition `107_thunder_training_monitor_worker.sql` (clients_db): inserts agent type `thunder-training-monitor-worker` (INSERT…ON CONFLICT (type,version) DO UPDATE, image fields copied from model-trainer, workflow as a JSON literal). Workflow `probe→classify→{reset_streak|bump_streak|mark_complete|mark_failed}→[decommission]→done`, exercising all five actions. Verified shape against the launcher def (102) and the agent_definitions schema (type/display_name/category/default_config NOT NULL; unique (type,version)). No per-step input_mapping needed — provisioning_id/training_run_id arrive in input_data from the orchestrator's call_agent at spawn, and classify reads the probe result by step name (results stored under BOTH step name and output_field, coordinator L1636/L2408). probe step named "probe" so classify's default probe_step matches.

(4b) ✅ DONE — orchestrator + schedule: `108_thunder_training_monitor_orchestrator.sql` (clients_db). Orchestrator `thunder-training-monitor` workflow `find_active_training_instances → loop(items_field "find_instances.instances", item_variable "current_instance", sub_workflow: spawn_worker[role monitor_worker] → call_worker[target_role monitor_worker, input_mapping provisioning_id/training_run_id from current_instance]) → done`. **Design choice (grounded, not reaper-style):** the scheduler merges only the FIRST pre_query row + fires once per tick (010 §Pre-Queries), so the reaper's scheduler-finds shape would starve newer instances (an ALIVE box sits atop the ordering forever, since ALIVE boxes are not decommissioned). Hence the orchestrator does the find+fan-out; `find_active_training_instances` is the right call. **Deployment-wiring question RESOLVED — no manifest change (001 §Topics):** scheduler fires the orchestrator via the generic entry point (`system.agent.generic.requests`, config.agent_type), it runs in the existing generic chassis pods (pure coordination), each worker is spawned as its own Job pod with a per-spawn `job.<id>.requests` topic, and workers reply to the orchestrator (their parent). Loop SEQUENTIAL (call_agent awaits) — same-step spawn-topic reuse is safe only because the prior worker Job finishes first; not fire-and-forget. Schedule: `scheduled_tasks` row, 300s, gated pre_query (skip when nothing training), `concurrency_group` max_concurrent 1, **inserted DISABLED** (won't fire workers before the actions exist); ON CONFLICT does NOT overwrite `enabled` so a manual enable survives re-apply.

**MONITOR BUILD COMPLETE — all artifacts written (pending go build + deploy + apply + test + enable).** Deploy order: (1) `go build ./...`/`go vet`, deploy chassis with the 5 new actions + their registry lines; (2) apply 106 → 107 → 108 (clients_db); (3) manual one-off worker test (post to generic entry, config.agent_type=`thunder-training-monitor-worker`, input_data {provisioning_id, training_run_id} of a known box → expect STATUS=ALIVE→reset_streak→done while training, or complete/fail+decommission when finished); (4) `UPDATE scheduled_tasks SET enabled=true WHERE name='thunder-training-monitor';`. Reuses unchanged: adapter `ssh_get_status`, `dispatch_thunder_decommission`, `spawn_agent`/`call_agent`/`loop`/`complete_workflow`, `datahelpers` helpers.

DEPLOY/VERIFY (2026-06-04, root cause CONFIRMED from the full chassis state dump): migrations 106/107/108 applied (both defs `active`). The empty-grep earlier was indeed the label selector — the worker runs in-process on the generic chassis pod (`agent-chassis-99cdd59bc-f4v67`), and the full log shows it executing the FULL `WorkflowPlan` (probe→classify→…), reaching `probe`, dispatching `dispatch_thunder_ssh_get_status` to thunder-adapter (its own request_id `a9e722e8`), and pausing AWAITING_RESPONSES. **Root cause of the stuck row:** a reply-topic mismatch. The coordinator registered the awaited request on `system.agent.generic.responses` (`determineResponsesTopic` priority #1 = env `RESPONSES_TOPIC`, which pins it to the pod's OWN responses topic — it overrode the action's result topic), but the OLD `dispatch_thunder_ssh_get_status` (cloned from `dispatch_thunder_ssh_exec`) put `__parent_responses_topic__` = `system.generic.responses` (the CLI-derived `reply_to_topic`, no `.agent`, nothing consumes it) into the adapter request envelope. So the adapter was told to reply where the coordinator wasn't listening → orphaned await → never reached `reset_streak` (the only writer of `last_probe_at`) → row stayed 0/NULL/running. **FIX APPLIED** (`thunder_ssh_get_status_dispatch.go`): envelope reply-to now prefers `ExecutionContext.ResponsesTopic` (= what the coordinator awaits, in both spawned and generic-entry paths), `__parent` only as fallback — DIVERGES intentionally from the ssh_exec dispatch (which works only because it's called from spawned children where the two topics coincide; latent same bug if ever fired top-level). The probe await deadline is per-step `GetStepTimeout` → `DefaultRequestTimeout` (the step sets no explicit timeout), adequate for the SSH round-trip. **VERIFIED (2026-06-04 18:21, orch `66e33188`, pod `agent-chassis-6499dbd8fc-xs6tr`):** redeployed with the fix and re-fired → the full ALIVE path ran end-to-end: `probe` → thunder-adapter replied `STATUS=ALIVE` / `reachable:true` / exit 0 (await RESOLVED, response from `agent_type:thunder-adapter` in_response_to the probe step) → `classify` `verdict:alive` → `reset_streak` `streak:0` → `done` COMPLETED. DB confirms `last_probe_at=2026-06-04 18:21:12`, `consecutive_unreachable_probes=0`, `status=running`; no decommission (correct for ALIVE — box left running). The reply-topic fix is confirmed; this run also validated `classify_training_probe` parsing/routing, `record_probe_streak` (reset mode), `complete_workflow`, and the probe `status_command` running against the live box (`pgrep` matched `02_train`). **CORRECTION to the prior note:** the stale stub `agent_config` (def `470c6b3f`, `start_step:complete`, "scheduled task pre_query already did the work", `timeout_seconds:10`) PERSISTED across the redeploy (new pod hash) — so it is NOT in-pod cache as I previously said; it loads from a persistent source on every request while `WorkflowPlan` is built from the full def. It is cosmetic (the full plan executes; the stub's timeout does not govern the per-step await) but a real inconsistency — investigate at the DB level: `SELECT id, version, is_snapshot, is_active, status, default_config->'workflow'->>'start_step' FROM agent_definitions WHERE type='thunder-training-monitor-worker' ORDER BY version, is_snapshot;` (look for a snapshot or second row carrying `start_step=complete`; then ask why the envelope fields load from it). **Still NOT exercised:** the terminal/unreachable verdict paths (`mark_training_run_terminal` → `decommission`; `bump_streak`→lost) and the ORCHESTRATOR (`find_active_training_instances` + `spawn_worker` + `call_worker` loop). NEXT: orchestrator test (`config.agent_type='thunder-training-monitor'`, no input_data → find + loop + spawn + call); then exercise a terminal verdict; then enable the schedule — but NOT until the B2 adapter-upload exists (a real `DONE_OK` would decommission and lose the adapter). **Schedule still disabled.**

STUB SOURCE NARROWED (2026-06-04): `SELECT … WHERE type='thunder-training-monitor-worker' ORDER BY version, is_snapshot` returned exactly ONE row (`470c6b3f`, version 1, is_snapshot=f, is_active=t, status=active, `start_step=probe`). So the definition is clean and the stub `agent_config` is NOT a snapshot or duplicate row — it is a stale representation from a NON-DB source (a shared/out-of-process definition cache the SQL UPDATE did not invalidate, surviving pod restarts). Cosmetic while the full `WorkflowPlan` executes. ORCHESTRATOR TEST (in progress): fire `config.agent_type='thunder-training-monitor'`, `input_data={}` → expect `find_active_training_instances` → 1 instance (`fabfd7fa`) → `monitor_loop` spawns a worker Job + `call_worker` (awaits) → worker runs the ALIVE path → `last_probe_at` advances to a NEW timestamp (> 18:21:12) → orchestrator COMPLETED. This is the FIRST run of the worker as a SPAWNED CHILD (confirms the reply-topic fix holds when `__parent` is a real consumed topic) AND the functional check of the stale stub: if the spawned Job runs the no-op `complete` workflow instead of the full one, `last_probe_at` will NOT advance — that is the stale source biting the spawn path, at which point the shared-cache hunt becomes worthwhile. Safe: `fabfd7fa` is ALIVE → reset_streak, no decommission.

PENDING: watch `train.log` to `RUN_SH_FULL_OK`/`RUN_SH_DONE`; **scp `/workspace/adapter_out` off `fabfd7fa` BEFORE anything decommissions it** (run.sh does NOT upload — the adapter is local-only); then decommission + reconcile `1cd65dd7` (+ `e6ab9fad`). Structural fixes: **checkpointing + final-adapter upload to B2 now has an agreed design — see `PLAN_checkpoint_and_artefact_upload_b2.md`** (presigned single-object write-only PUTs minted by the adapter, pre-minted at launch into a `/workspace/upload_manifest.json`; keyed by save-index; `RUN_SH_DONE` moves to after the final upload, which is what makes the monitor's `DONE_OK → decommission` safe; resume via cluster-side checkpoint selection + presigned GET). Remaining queued: cap-sizing from smoke s/step at provision, FA2 + fewer epochs, and correcting the runbook "30–90 min" line.

### Guideline audit (001 dev guide / 002 architecture / 003 contracts) — 2026-06-04
Checked all 7 monitor artifacts (5 Go actions + migrations 106/107). **Compliant:** agent types kebab-case (`thunder-training-monitor`, `-worker`); action names snake_case; status/`complete`/`failed` single-word lowercase (003 §String-Value Naming); SQL parameterised `$1/$2`, no `{{.}}` interpolation (003 §Parameterisation); `ProduceWithValidation` not `Produce`; no `logger.Debug` in my code; no wrapper+core split — each file exports only `XxxAction` (+`XxxInputSpec`) (001 §Actions are the unit of work); dispatch reply routes to the PARENT responses topic (001 §Agents respond to caller's topic); `complete_workflow` uses `output_fields` plural; migrations idempotent; adapter `ssh_get_status` + `dispatch_thunder_decommission` reused unchanged. **Fixed:** (1) `record_probe_streak` dropped custom `configIntDefault` → `datahelpers.GetIntField` (+sanity floor) per 001 §"grep datahelpers before adding a utility" (GetIntField handles the float64 a JSON-number config yields); (2) `107` now copies `category` from model-trainer (was a hardcoded `'training'` guess). **Flagged → both resolved:** (C) ✅ `find_active_training_instances` KEPT as a dedicated query action (your call; backed by the `LoadWorkItemsAction` precedent and 001 §"DB queries belong in Go"). (D) ✅ aligned `classify`/`record_probe_streak`/`mark_training_run_terminal` to the strict 001 §Config-value split: literals (`status`, `error_message`, `mode`, `probe_step`, `*_step`) now read directly from `params.StepConfig.Config` via `datahelpers.GetStringField`; input_data values (`provisioning_id`, `training_run_id`) via `ExtractActionInputs` (+ new `RecordProbeStreakInputSpec` with `init()` registration). `configOrInput` now appears only in the thunder dispatch family (its native pattern). `go build ./...`/`go vet` still pending in-repo.

### Update — 2026-06-04 ~1x: training-monitor VERIFIED live (both paths); reply-topic orphan fixed
[verified-db] Orchestrator test passed: `find_active_training_instances` → spawn worker Job → `call_worker` (awaits) → worker ALIVE path → `last_probe_at` advanced → orchestrator COMPLETED. Worker-direct ALIVE path also confirmed. The earlier "no chassis log" was a reply-topic orphan in `thunder_ssh_get_status_dispatch.go` (built own-topic wrongly) → fixed to `execCtx.ResponsesTopic`. The stale-stub `agent_config` did NOT bite the spawn path (cosmetic, non-DB cache). **Terminal/decommission branch still never run live** — fires on the next finishing box. Not enabled.

### Update — 2026-06-05: iter_0 CLOSED OUT
[verified-db] `fabfd7fa` reached its final save; `/workspace/adapter_out` scp'd to `~/projects/agentchassis/iter0_adapter_out` (`adapter_model.safetensors` 828MB + tokenizer + run-metadata `manifest.json`); `training_run 1cd65dd7` reconciled to `complete`; box decommissioned + confirmed gone (`tnr` "No instances found"). Schedule never enabled on this run. **Still open:** stale `e6ab9fad` is `running` (box gone) — reconcile to `failed`; sweep other `running` runs/instances (folded into RUNBOOK pre-flight).

### Update — 2026-06-05: upload path Phases A / B / C BUILT (A Tier-1 PASSED)
[verified-source unless noted] Closing the "run.sh writes adapter local-only" gap — see `PLAN_checkpoint_and_artefact_upload_b2.md`.
- **A (`02_train`):** gated flags `--save-steps` / `--save-total-limit` / `--upload-manifest`; `CheckpointUploader` callback PUTs each save by **save-index** (best-effort); **final upload is a hard gate** (raises → non-zero exit); resume staging if `manifest.resume`. With no flags it is byte-for-byte the old script. **Tier-1 box-free B2 round-trip PASSED** [verified-db, personae-model-training @ us-east-005]: presigned PUT accepts `application/octet-stream`, the `checkpoints/` exclusion holds, GET+extract byte-identical.
- **B (launcher):** 3 pure actions `compute_checkpoint_keys` / `flatten_presign_results` / `assemble_upload_manifest`; a `key_path` source added to `dispatch_thunder_prepare_object_url` (for the plain-local `presign_final`); migration `109` wiring `compute_keys → presign_checkpoints(loop) → flatten → presign_final → assemble → write_manifest → ssh_exec_launch`.
- **C (`run.sh`):** `--save-steps 50 --upload-manifest` on full-train only (smoke untouched; absent-manifest = old behaviour). **No marker moved** — `set -e` + A's final-upload raise means `RUN_SH_DONE` only prints on exit 0 = trained AND uploaded. This is what makes the monitor's `DONE_OK → decommission` safe.

### Update — 2026-06-05: Phase D adapter side BUILT — reused `ListObjects` (the "list-keys gap" was wrong)
[verified-source] `prepare_resume_url`: `DataURLAction.ResumeURL` + `handlePrepareResumeURL` + dispatch switch case in `adapter.go`. **Reuses the EXISTING `storage.Client.ListObjects`** (interface.go; `*S3Client` aws-sdk-go-v2 `ListObjectsV2Paginator`, same `c.bucket` the dataset/artefact presigns use) — NOT a new method. The plan's "ONE genuine adapter gap / list-keys needed" was wrong: item-18 blast-radius is about **adding** to that broad interface; listing was already on it. Removed the narrow `objectLister` interface I had first written. `ResumeURL` → `ListObjects(prefix)` → `latestCheckpointKey` (max `ckpt-<N>`) → `GetPresignedURL`; `found=false` on an empty prefix ⇒ launcher trains fresh. **List (B2 network) failure → `error_recoverable`** (chassis retries `check_resume`); presign (local signing) + bad-input → `error_unrecoverable`. **LEFT:** `dispatch_thunder_prepare_resume_url` + migration `110` (`check_resume` step); `assemble_upload_manifest` already emits `resume` only when `resume_url` is non-empty.

### Update — 2026-06-05: guideline audit — Phase D + Phase B
[verified-source] Ran the new code against 001/002/003.
- **Adapter Response Envelope (003 §832): PASS by reuse.** The adapter's `sendSuccessResponse`/`sendErrorResponse` already use a typed `responseHeaders` (real `bool` `is_complete`/`is_error`), reuse the incoming `request_id` for both id fields, set `message_id`, and send via `ProduceWithValidation` (the 2026-05-22 matcher fix). `handlePrepareResumeURL` inherits all four checklist items — hand-rolling the envelope would have risked re-introducing that exact fault.
- **Phase B loop fix.** Checked all 11 production loops in `bk_agent_definitions_backup.sql`: every one ends its sub_workflow on an explicit `loop_complete` substep with the async `call_agent`/work substep chaining into it. `presign_checkpoints` had `presign_one` (an awaited async dispatch) as the terminal with **no** `loop_complete` → **FIXED** (added `presign_done` `loop_complete`, `presign_one.next_step → presign_done`). Matters because the iteration boundary fires on `loop_complete`'s empty `next_step`, and the `ErrLoopExpansionHandled` race is precisely about fast async responses there.
- **Actions conform.** All 3 + `key_path` + registry: correct signature, `initialize` early-return, `RegisterActionInputSpec` in `init()`, `ExtractActionInputs`, canonical `datahelpers` reuse (`ExtractStringListHelper`, `ExtractNestedFieldString`, `GetIntField`/`GetStringField` — no new field-path fn), `IsLocal:true`, and no const/func name collisions in the actions dump. `compute_checkpoint_keys` emits `checkpoint_keys` (the loop dep — confirmed). `continue_on_error:false` on the loop is **deliberate** (manifest `checkpoints[]` is index-aligned; a skipped presign would shift it, and presign is local/unrecoverable anyway). `109` is a static `jsonb` migration → 003 §611 parameterisation and §946 agent-def conventions don't bind.

### Update — 2026-06-05: docs + deploy runbook; deploying to test the environment
`PLAN_checkpoint_and_artefact_upload_b2.md` corrected (resume/`ListObjects`, final key = `finetuning/artefacts/`, `run.sh` `set -e` mechanism, A/B/C/D statuses). **`RUNBOOK_phase_b_c_d_deploy.md`** written for steps 2–6: verify-live + apply `109` → re-pack/re-upload bundle (both edited scripts at root) → Tier-2 short launch (B+C integration; `SAVE_STEPS` low for the test, restore 50 after) → resume (⛔ blocked on D3 + `110`, and on step 4 being green) → enable monitor (last; gated on step 4 proving `DONE ⟹ uploaded`). Deploying the current build now.

### Update — 2026-06-05: deploy step 2 done; `write_manifest` /workspace perm bug found + fixed (`109a`); B2 CLI not aws
[verified-db] `109` applied + verified live: `presign_scripts.next_step=compute_keys`; the six new steps chain through to `ssh_exec_launch`; loop substeps `presign_one → presign_done`. Bundle key confirmed `finetuning/scripts/bundle.tar.gz`; `ssh_exec_launch` carries `scripts_url`/`dataset_url` at **top-level** config (token interpolation OK). State clean: `e6ab9fad` reconciled to `failed`, zero non-terminal `thunder_instances`. NOTE: four `pending` training_runs (`efd9d9f7`/`ca582ce0`/`f02b586a`/`27ef7bea`) sit unstarted — watch which `run_id` the test trigger actually claims.
- **BUG + FIX (`109a`).** `write_manifest` is the FIRST launcher step to touch the VM filesystem (all prior steps are presign/pure — no SSH) and used a **non-sudo** `mkdir -p /workspace`. `ssh_exec_launch` (next step) creates /workspace with `sudo mkdir -p` + `sudo chown $(id -u):$(id -g)` — which proves `/` is **not** writable by the ssh user — so `write_manifest` failed (mkdir, or the `> /workspace/upload_manifest.json` redirect) before the manifest was written, dying the launch there. iter_0 never hit this (old launcher had no `write_manifest`). Fixed to mirror `ssh_exec_launch`: `sudo mkdir -p /workspace && sudo chown $(id -u):$(id -g) /workspace && …`. `109a` patches the live row (109 was already applied); canonical `109` corrected too. **Applied + verified live.** [verified-db]
- **B2 CLI, not aws.** The runbook's `aws s3` examples fail here ("Unable to locate credentials") — we use the native **`b2`** CLI. Upload: `b2 file upload personae-model-training bundle.tar.gz finetuning/scripts/bundle.tar.gz` (b2 v4) or `b2 upload-file …` (v3). List: `b2 ls --long "b2://personae-model-training/finetuning/checkpoints/<run_id>/"` (v4) or `b2 ls --long personae-model-training finetuning/…/` (v3). aws works only if `AWS_ACCESS_KEY_ID`/`AWS_SECRET_ACCESS_KEY` are exported from the B2 key id/secret. `RUNBOOK_phase_b_c_d_deploy.md` updated to b2.

### Update — 2026-06-06: Tier-2 / B+C test LAUNCHED (export a8484922, epochs 1)
[verified-db/source] Fired the 4b trigger with export `a8484922-4e3d-4c11-a2ed-4056b13a67a6`
(page-content-writer, 1957 rows, ~21MB), `epochs:1`. correlation `999f532f`, orchestration
`8c50a756`. Watched the chassis log through `spawn_launcher`.
- **Chassis live at `v1.0.1057`** — the `training-launcher` Job spawned with this image, so
  the new build (the four Phase B actions + registry) is deployed. (`spawn_actions` "Agent
  definition loaded for spawn", image `docker.io/aqls/agent-chassis:v1.0.1057`.) The
  thunder-adapter `prepare_resume_url` is a SEPARATE image and is only needed for resume
  (step 5); the upload test uses the adapter's existing `prepare_object_url`.
- **model-trainer flow confirmed live:** `spawn_data_preparer → spawn_provisioner →
  spawn_launcher → call_data_preparer → call_provisioner → call_launcher → complete`. It
  spawns all three workers first, then calls them in order. `call_data_preparer` INSERTs the
  new `training_runs` row (pending) and streams the export to S3 JSONL; `call_provisioner`
  provisions the A100; `call_launcher` runs our 109 chain.
- **Box-resolution path RESOLVED (earlier open question).** The LIVE model-trainer's
  `call_launcher` input_mapping passes **`provisioning_id`**
  (`provisioning_result.dispatch_provision.provisioning_id`) to the launcher, alongside
  `instance_ip?`, `ssh_user?`, `ssh_key_secret_name?`, `dataset_uri`, `training_run_id`,
  `hyperparameters`. So the launcher's ssh_exec / `write_manifest` / presign dispatches
  resolve the box via `input_data.provisioning_id` (configOrInput). The launcher receives
  this when CALLED (`call_launcher`, after `call_provisioner`), not at spawn — the init
  message has empty `input_data`, which is fine. (`bk_agent_definitions_backup` passed only
  `instance_ip`; the live def is newer.) Confirms `write_manifest` will reach the box.
- **Stale-stub `agent_config` confirmed cosmetic.** The message carries a no-op
  `agent_config` (`start_step: complete`, "scheduled task pre_query already did the work"),
  but the **`WorkflowPlan`** (full 7-step flow) is what executes — the known non-DB-cache
  artefact; not cleared by redeploy, does not affect the run.
- Transient `Topic not yet on broker` for the launcher `.responses` topic self-healed on
  attempt 2 (topic-creation race) — normal.
- IN PROGRESS at time of writing: awaiting `call_launcher` (the 109 chain:
  `compute_keys → presign loop → MANIFEST_WRITTEN → ssh_exec_launch`), the box, and the
  first `ckpt-*.tar.gz`. The new `training_run_id` is created at `call_data_preparer` — get
  it from `training_runs` (newest) and watch `finetuning/checkpoints/<run_id>/` in B2.

### Update — 2026-06-06 (2): empty-export bug; model-trainer fall-through; re-fired on 146a9a12 → 109 chain executing; loop-output OPEN
[verified-db] First Tier-2 attempt (export `a8484922`) FAILED at `prepare_training_data`:
"export a8484922… has no training rows" → run `693656ce` failed, no box provisioned.
Confirmed: `a8484922` has **0** rows in `training_exports.rows` despite
`training_exports.runs.rows_exported=1957`; `146a9a12` and `fef7be6b` both have 1958.
**Lesson: verify an export with `SELECT count(*) FROM training_exports.rows WHERE
export_id=…`; `runs.rows_exported` can disagree with the actual `rows`.**
- **BUG (model-trainer, separate) — fall-through after a failed `call_agent`.** When
  `call_data_preparer` failed, the model-trainer did NOT abort — it ran `call_provisioner`,
  which failed differently: `source path 'preparation_result.training_run_id' not found`
  (because `preparation_result` then held the error, not the prep output). A failed
  `call_agent` step should terminate the orchestration, not continue with an empty
  output_field. Harmless here (run already failing) but produces a confusing second error.
  **Flagged for a model-trainer fix — not ours, not blocking the upload test.**
- Re-fired with `146a9a12` (correlation `cdba2808`). The 109 chain is now EXECUTING in the
  launcher pod: `presign_dataset` succeeded (adapter returned a presigned GET for
  `finetuning/datasets/146a9a12…/training.jsonl`), `compute_checkpoint_keys` produced 40
  keys for the new run `70156952-180b-41bf-ac54-e0288238a288`, dataset JSONL is in B2.
- **RESOLVED (`109b`) — presign_checkpoints loop presigned the DATASET key 40×.** Root cause
  confirmed from the launcher pod: every `presign_checkpoints_iter_N_presign_one` dispatch
  logged `object_key="finetuning/datasets/146a9a12…/training.jsonl"`. `presign_one` used
  `input_mapping {key: ckpt_key}`, but input_mapping does NOT populate `input_data.key` for a
  (local-action) loop substep, so `dispatch_thunder_prepare_object_url` ran its chain
  (explicit key → key_path → **dataset_uri fallback**) and derived the dataset key. The
  `page_html`/`rendered_html` fields in the earlier loop results were downstream noise, not
  the cause — the expanded plan's `substep_output_fields:["ckpt_presign"]` was already
  correct. Fix: `presign_one` reads the item via `key_path:"ckpt_key"` (CollectedData, where
  setLoopVariable puts each iteration's item; resolves BEFORE the dataset_uri fallback) —
  matching the proven production-loop pattern (config dot-path, e.g.
  `"work_item_id":"current_item.id"`) and `presign_final` in this same def. No Go change.
  `109b` patches the live row; canonical `109` corrected. **CORRECTS a load-bearing
  assumption: input_mapping is NOT live for (local-action) loop substeps — reference the loop
  var via a config dot-path the action resolves (key_path), not input_mapping.**
  VERIFIED against source + docs before applying: `setLoopVariable`
  (loop_expansion_handler.go) writes the item to `CollectedData[loop_var_name]`=`ckpt_key`
  (string → stored as-is) right before the substep runs; `key_path` reads exactly that;
  `LoopCompleteAction` Strategy 1 aggregates `ckpt_presign_<N>` (the `page_html`/`rendered_html`
  were Strategy-3 generic-fallback padding, fired only because `ckpt_presign_<N>` was absent in
  the broken run); `prefixDataReference` leaves `ckpt_key` alone (not a substep output_field).
  Dev guide loop reference confirms substeps read the item via a config dot-path
  (`"input_from":"current_item.field_name"`) and lists `input_mapping` as the `call_agent`
  caller mechanism — so input_mapping was the wrong tool for this local-action substep.

### Update — 2026-06-06 (3): 109b CONFIRMED working in prod; NEW blocker — loop-await race (send-before-register), OPEN
[verified-source] After applying 109b the adapter presigns the CORRECT per-iteration key —
logged `key="finetuning/checkpoints/a2a41ae2-…/ckpt-9.tar.gz"`, method PUT, for
`presign_checkpoints_iter_9_presign_one`. The key_path binding fix is confirmed live; the
dataset-key bug is closed.
- **NEW BLOCKER (OPEN) — presign_checkpoints loop stalls intermittently at a LATER iteration.**
  The loop now progresses (iters 0–8 complete) then wedges on one: iter_6 in run `cdba2808`,
  iter_9 in run `a2a41ae2`. A MOVING stall point ⇒ a RACE, not a deterministic wiring bug.
  The adapter replies (~1s, and even twice — original + retry — for the stuck iteration), but
  the launcher never clears the await; the timeout handler re-dispatches every ~3 min with a
  fresh request_id, so RetryVersion stays 0 and it never reaches the max-retries fail path →
  effectively infinite retry. Boxes get reaped at the 18h cap meanwhile.
- **Root cause (verified against the actions/coordinator dump).** Our loop substep is a LOCAL
  dispatch (`dispatch_thunder_prepare_object_url`): it PRODUCES the adapter request, returns
  `await_response:true`, THEN the coordinator registers the awaited request
  (`processAwaitResponse` → persist state → `InsertAwaitedRequest`). Send-BEFORE-register. For
  a ~1s reply the response can arrive before the `awaited_requests` row is inserted →
  `ClaimAwaitedRequest` finds no `waiting` row → the reply is dropped → timeout → retry → same
  race. `presign_dataset` (first dispatch) wins the race; later loop iterations lose it more
  often. CONTRAST: `spawn_agent`/`call_agent` call `preRegisterAwaitedRequest` (L57855) to
  register BEFORE sending — which is why the working production loops (call_agent substeps)
  don't stall and ours does.
- **Fix to try (non-framework, NOT yet implemented).** Make OUR dispatch
  (`thunder_prepare_object_url_dispatch.go`) call the EXISTING `preRegisterAwaitedRequest`
  just before producing the adapter request — mirroring `spawn_agent`. Closes the
  send-before-register window without touching the coordinator/loop machinery. The
  coordinator's later `processAwaitResponse` re-register hits "already exists" and no-ops
  (L1780). Affects all presign dispatches (dataset/scripts/loop/final/resume) — makes them all
  more robust, no harm. See the handoff for exact steps.
- **Fallback (structural) if the pre-register fix is fragile.** Replace the 40-iteration loop
  with ONE batch adapter call: a `prepare_object_urls` (plural) handler taking the key array
  (compute_checkpoint_keys already emits all 40) returning `[{key,url}]` (all local SigV4
  presigns, no per-key network). One async round-trip like `presign_dataset` → no loop, no
  flatten, no race class.

### Update — 2026-06-08: `preRegisterAwaitedRequest` body confirmed (the fix's exact mechanics + the one pre-check); race now also in the debug guide (§9, v2_35). Still OPEN.
[verified-source] Read `preRegisterAwaitedRequest` (L57855) end-to-end. It does:
`INSERT INTO awaited_requests (… status) VALUES (…, 'waiting') ON CONFLICT (request_id) DO NOTHING`,
setting `step_id = params.ExecutionContext.StepID`, `step_name = params.CurrentStep`,
`retry_version = 0`, `reply_to_request_id = requestID`, `timeout_at = now + 120s` (hardcoded).
- **Double-registration is safe.** The `ON CONFLICT DO NOTHING` means the coordinator's later
  `processAwaitResponse → InsertAwaitedRequest` no-ops — so pre-registering in the dispatch and
  still returning `await_response:true` is correct; no schema or flow change needed.
- **Two things to mind when wiring it into the dispatch:**
  1. **VERIFY FIRST (the one pre-check):** `step_name = params.CurrentStep`. The dispatch must
     have `params.CurrentStep` holding the EXPANDED loop-substep name
     (`presign_checkpoints_iter_N_presign_one`) at dispatch time, or `handleCompleteResponse`
     resumes the wrong step. Confirm `ActionParams.CurrentStep` before relying on the fix.
  2. The **120s `timeout_at` is hardcoded** and, via `ON CONFLICT DO NOTHING`, wins over the
     step's configured timeout once pre-registered. Fine for ~1s presigns, but it pins every
     presign dispatch's await to 120s — note it, don't be surprised by it.
- **Status: the fix is fully specified but NOT yet implemented or tested — still the open
  blocker.** Next action, in order: (a) confirm `params.CurrentStep` is the expanded substep
  name; (b) call `preRegisterAwaitedRequest(ctx, params, requestID, targetAgentID="",
  targetAgentType="thunder-adapter", requestsTopic="system.adapter.thunder.requests",
  responsesTopic=<the await-responses topic the dispatch computes>)` just before
  `ProduceWithValidation`, guarded `if params.DB != nil`; (c) rebuild + redeploy the chassis
  (bump from `v1.0.1057`); (d) re-run with export `146a9a12` or `fef7be6b`; (e) confirm each
  `presign_checkpoints_iter_N_presign_one` logs `ClaimAwaitedRequest: status_before=waiting …
  claimed:true`, then loop → `flatten_checkpoint_urls` → `MANIFEST_WRITTEN` → `ssh_exec_launch`.
- **Debug guide updated.** The race is now a standalone §9 entry in
  `016_debugging_guide_v2_36.md` — "Adapter reply dropped at a fast loop iteration — the
  local-dispatch await is send-before-register (race)" — recorded as the FOURTH cause of the
  `awaited_requests`-stuck-`waiting` symptom (distinct from the `is_complete` bool-unmarshal and
  the envelope-recognition causes), with a cross-ref bullet added to the existing "Adapter
  response silently dropped" entry. (v2_35 = v33 + the restored v30 "Traps A–C" section + this
  branch.)

### Update — 2026-06-08 (2): pre-register fix APPLIED in `thunder_prepare_object_url_dispatch.go`. Pending chassis rebuild + verify.
[verified-source] The one pre-check from the previous entry is confirmed and the edit is in.
- **Pre-check resolved.** `params.CurrentStep` IS the expanded loop-substep name at dispatch
  time. `buildActionParams` (coordinator) sets `ActionParams.CurrentStep = state.CurrentStep`,
  and loop expansion sets `state.CurrentStep` to the injected iteration step, which each
  iteration runs through the normal `executeStep` path. So the dispatch sees
  `presign_checkpoints_iter_N_presign_one`. Equivalence: `createAwaitedRequest` (the coordinator's
  current path) writes `StepName = state.CurrentStep`, so the pre-register writes the SAME
  `step_name` — no step regression; `handleCompleteResponse` (`state.CurrentStep = awaited.StepName`)
  resumes the correct substep.
- **Request-id consistency confirmed.** The dispatch generates `newRequestID` once and uses it in
  THREE places: the envelope header `request_id`, the returned `result["request_id"]`, and now the
  pre-register. So `extractRequestID(result, …)` returns that same id and the coordinator's later
  `InsertAwaitedRequest` reuses it → `ON CONFLICT (request_id) DO NOTHING` → `RowsAffected==0` →
  "already exists" → `processAwaitResponse` treats it as success and does NOT start a second
  timeout goroutine (the `if !alreadyExisted` guard). One row, one timeout owner. `target_agent_id`
  passed as `""` matches `extractTargetAgentID(result)` (our result carries no agent_id key).
- **The edit.** Just before `ProduceWithValidation`, guarded `if params.DB != nil`:
  `preRegisterAwaitedRequest(ctx, params, newRequestID, "", "thunder-adapter", thunderAdapterTopic, myResponsesTopic)`.
  Note the arg order is `(requestsTopic, responsesTopic)` — requests first. On error it Warns and
  continues (race mitigation, not a hard gate), mirroring spawn. No variable renamed; no import
  added (`preRegisterAwaitedRequest` is in `spawn_actions.go`, same `actions` package). This makes
  ALL presign dispatches (dataset/scripts/loop/final/resume) register-before-send; the working
  ones only get more robust.
- **Carry-forward caveats.** (1) The helper's 120s `timeout_at` is hardcoded and, via the
  ON CONFLICT path, pins every presign dispatch's await to 120s regardless of the step's configured
  timeout — fine for ~1s presigns. (2) Because the coordinator's insert now no-ops, the per-request
  `handleRequestTimeout` goroutine is skipped (same as spawn); the background expiry sweep is the
  safety net. This also removes the current ~3-min infinite re-dispatch, which was that goroutine
  firing on a row it could never claim.
- **Status: code applied, NOT yet built/deployed/verified.** Remaining, in order: (a) `gofmt` +
  `go build ./...` + `go vet` in-repo; (b) rebuild the chassis bumping from `v1.0.1057`, roll the
  launcher; (c) cleanup (kill stuck launcher jobs, confirm no live `thunder_instances`); (d) re-run
  with export `146a9a12` or `fef7be6b`; (e) confirm each `presign_checkpoints_iter_N_presign_one`
  logs `Pre-registered awaited request in database` then `ClaimAwaitedRequest: status_before=waiting
  … claimed:true`, through the loop → `flatten_checkpoint_urls` → `MANIFEST_WRITTEN` →
  `ssh_exec_launch`. If a stuck iteration persists, grep the launcher pod for
  `ProcessResponse|ClaimAwaitedRequest|status_before|no matching|InsertAwaitedRequest` to tell
  race-closed-other-issue from reply-not-consumed; the structural fallback (one batch
  `prepare_object_urls`) is the next move only if this proves flaky.

### Update — 2026-06-09: pre-register fix CONFIRMED in prod; loop hit an O(n²) state-cost wall; DECISION — go batch.
Re-ran the Tier-2 launch (export `146a9a12`, chassis rebuilt past v1.0.1057, box `0zcmvvhb`/`d96d6530`).
- **Race fix works [verified-log].** Every `presign_checkpoints_iter_N_presign_one` logged
  `ClaimAwaitedRequest: status_before=waiting … claimed:true` — the awaited row existed when the
  reply landed, claimed not dropped. No `no matching`, no ~3-min re-dispatch, no double replies.
  The send-before-register race is closed. presign_dataset/scripts (GET) and the per-iteration PUT
  keys (`ckpt-0…`) were all correct.
- **BUT a SECOND problem surfaced — loop state-bloat, O(n²) [verified-log].** Per-iteration time
  grew badly: iter_0–4 ~2–3s each, iter_5 ~11s, iter_6 ~25s, iter_7 ~60s, iter_8 ~100s (one
  message's processing `duration:276s`), iter_9 a single `EXECUTING_STEP` persist took 53s. State
  `Version` climbed 62→86 over those iterations. Cause: every step does a full read-modify-write of
  the orchestration state, which embeds (a) the entire EXPANDED workflow — a 40-iteration loop is
  ~80 substeps, each with its verbose description; (b) `collected_data`, +1 presign result/iter;
  (c) `ProcessingHistory`, +~4 entries/iter. Blob grows O(K), with O(K) persists → O(K²), large
  constant from the verbose expansion. By iter_9 (~9 min in) the pod logged Kafka `i/o timeout` to
  broker prod-2 and "No activity for 5 minutes". The launcher never reached `write_manifest` (no
  `/workspace/upload_manifest.json` on the box). At this rate 40 iterations won't finish, and the
  GPU bills throughout. This is NOT the await race (claims all succeeded) — it is the loop's own cost.
- **Cleanup done.** Killed the launcher job; decommissioned the box via the decommission trigger
  (`tnr`: "No instances found" — confirmed gone). Mark the run failed to keep the table clean.
- **DECISION: replace the per-iteration loop with ONE batch adapter call (the documented fallback).**
  Trying to shave per-step cost inside the loop means coordinator surgery on what gets persisted
  (fragile, framework-touching, against the structural-fix preference). The batch route deletes the
  whole class. Plan (NOT yet built — next session):
  1. **Adapter handler `prepare_object_urls` (plural)** in `data_url_actions.go` + a `case` in
     `adapter.go`'s action switch (alongside `prepare_object_url`, switch ~L323–338). Body:
     `{keys []string, method, expiry_minutes}`. Loop the EXISTING `DataURLAction.ObjectURL` primitive
     per key (GET→`GetPresignedURL`, PUT→`GetPresignedPutURL`; reuse — no new signing path), reply
     `{urls:[{presigned_url,key,expires_at,method}], count}` via `sendSuccessResponse`. (`ObjectURL`
     returns `PreparedURLResult{PresignedURL,Key,ExpiresAt,Method}` — confirmed in the uploaded
     `data_url_actions.go`.)
  2. **Launcher dispatch** that reads the key array (`compute_checkpoint_keys` already emits all K),
     sends ONE request, awaits ONE reply — pre-registering the await exactly like the singular
     dispatch (keep the race fix). Likely a sibling action `dispatch_thunder_prepare_object_urls`
     reusing the singular file's envelope/topic/pre-register code; the difference is keys-array in +
     urls-array out.
  3. **Migration** replacing `presign_checkpoints` (40-iter loop) + `flatten_checkpoint_urls` with
     the single batch step, and adjusting `assemble_upload_manifest`'s inputs (it currently reads the
     flattened `ckpt_presign_0..N`; now it reads the batch `urls[]`).
  Net: one round-trip, one state persist, no 80-step expansion, no `flatten`. First action next
  session: read the adapter's `prepare_object_url` handler + `ObjectURL` + storage presign helpers
  and `assemble_upload_manifest`'s current input shape before writing anything; check the migration
  number in the runner; do NOT touch code until the adapter/handler shapes are confirmed.

### Update — 2026-06-09 (2): batch route — Step A (adapter) BUILT; Step B (chassis dispatch + migration 110) in progress.
Confirmed the contract by reading `assemble_upload_manifest_action.go`: it already consumes two
PARALLEL lists — `checkpoint_keys` and `checkpoint_urls` — paired by index with a hard
length-mismatch error. `checkpoint_urls` was the output of `flatten_checkpoint_urls`. So the batch
only needs to return an ORDERED url list; `flatten` is removed and `assemble_manifest`'s Go is
UNCHANGED — only the def's `checkpoint_urls` config dot-path moves. Migration number = **110**
(confirmed by user; Phase D resume wiring moves to a later number).
- **Step A — adapter, BUILT (additive, ship-alone-safe).** `data_url_actions.go`: new
  `ObjectURLsRequest{Keys,Method,ExpiryMinutes}` + `handlePrepareObjectURLs` — mirrors the singular
  handler and loops the EXISTING `DataURLAction.ObjectURL` per key (no second signing path). Reply
  `{presigned_urls[], keys[], count}`, `presigned_urls[i]` aligned to `keys[i]`; a single presign
  failure fails the whole batch (matches the retired loop's `continue_on_error:false`).
  `adapter.go`: one `case "prepare_object_urls"` next to `prepare_object_url`. No new imports.
- **Step B — chassis dispatch + migration 110 (in progress).** New action
  `dispatch_thunder_prepare_object_urls` (`thunder_prepare_object_urls_dispatch.go`): reads the key
  array via the PROVEN `ExtractActionInputs` + `ExtractStringListHelper` path (same mechanism
  `assemble_manifest` uses for `checkpoint_keys` — config value is a dot-path resolved against
  CollectedData; NOT a guessed extractor), `method`/`expiry_minutes` as literals via `configOrInput`,
  and reuses the singular dispatch's envelope/topic build + `preRegisterAwaitedRequest` race fix +
  the await-return shape. Factored the own-responses-topic derivation into a shared
  `ownResponsesTopic(params)` helper used by BOTH dispatches (reuse-first; the singular file is
  edited only to call it — no behaviour change, noted). The new action must be registered in the
  actions registry alongside the singular.
- **Migration 110 — NOT yet written; needs the live def first.** Must confirm from the running
  `training-launcher` def: the compute step's output_field + the checkpoint-keys field name (trace
  shows `presign_final.key_path = "ckpt_keys.final_key"`, so compute's output_field is `ckpt_keys`,
  NOT `compute_checkpoint_keys` as the action doc-comment loosely says — the real path must be read,
  not assumed), `presign_checkpoints.next_step`, `flatten_checkpoint_urls`'s output field +
  next_step, and `assemble_manifest.config`'s `checkpoint_urls`/`checkpoint_keys` paths. Then 110:
  replace the `presign_checkpoints` loop step with the batch dispatch step (keys path → the real
  compute path, method PUT, expiry 3000, output_field `ckpt_presign_batch`, next_step → whatever
  followed `flatten`), remove `flatten_checkpoint_urls`, repoint `assemble_manifest.checkpoint_urls`
  → `ckpt_presign_batch.presigned_urls`. Per the rules, the def is read before the SQL is written.

### Update — 2026-06-09 (3): batch route CONFIRMED end-to-end in prod. Launcher green.
Applied migration 110 (snapshot_agent captured source_version=1 / id 1223bdc1, reason
"pre-migration-110 batch presign"; 3 UPDATEs; COMMIT), built+deployed the adapter and chassis
images with the new handler/dispatch/registry, and re-ran (export `146a9a12`, epochs:3, correlation
`07dc14fa`, run `0ac806ab`, box `3cf69d7b` / 216.81.200.234).
- **Full launcher path completed in ~26s [verified-trace], state Version 30.** Order: presign_dataset
  → presign_scripts → compute_keys → **presign_checkpoints (ONE batch await, 15:05:23→25, ~2s,
  returned all 40 ckpt-0..39 PUT URLs)** → presign_final → assemble_manifest → write_manifest
  (`MANIFEST_WRITTEN`) → ssh_exec_launch (`exit_code:0`, `LAUNCH_PID=216`, reachable) → mark_running
  → complete → COMPLETED → notified parent success. Contrast the retired loop: Version 86 / still at
  iter_9 nine minutes in, never reached write_manifest. The O(K²) class is gone.
- `assemble_manifest` consumed `checkpoint_keys: ckpt_keys.checkpoint_keys` +
  `checkpoint_urls: ckpt_presign_batch.presigned_urls` (the 110 repoint), no `flatten` step — its Go
  was untouched. `total_steps` dropped 11→10 (loop+flatten replaced by one batch step).
- **Backup correction (this milestone).** Migration 110's backup step uses the sanctioned
  `snapshot_agent('training-launcher', '<reason>')` (rollback: `revert_agent`). The earlier
  hand-rolled `CREATE TABLE IF NOT EXISTS agent_definitions_backup (...)` was wrong — that table
  already exists (mirrors agent_definitions); IF-NOT-EXISTS no-ops then the INSERT fails on missing
  columns. Lesson recorded in debug guide §6.1 + Schema reminders (v2_40): discover DB helpers with
  psql `\df` before hand-rolling SQL. snapshot_agent writes to agent_definitions_backup (not a new
  agent_definitions row), so it does not shadow the live patched def.
- **Still pending — the durability proof plays out on the box over hours, NOT in the launcher.** The
  launcher's job (provision→presign→manifest→launch) is done. Whether checkpoints actually land in
  B2 via the presigned PUTs is run.sh's job on the VM: watch `/workspace/train.log` RUN_SH markers
  and `b2 ls personae-model-training finetuning/checkpoints/0ac806ab-.../` (runbook 4d/4e). At
  SAVE_STEPS=50 the first checkpoint lands ~1.5h in. Reaper deadline 18h; the
  thunder-training-monitor (terminal/decommission branch) is still the last untested piece.

### Update — 2026-06-09 (4): on-box manifest + run.sh verified; one non-blocking follow-up (expiry_minutes).
- **Manifest correct on the VM [verified].** `/workspace/upload_manifest.json` has all 40 checkpoints,
  indices 0→39 contiguous, each `index:i` paired with `ckpt-i.tar.gz`, plus `final` adapter and
  `run_id`, no `resume` key. The contiguous index/key pairing is the real proof the batch reply held
  order (assemble pairs checkpoint_keys[i] with checkpoint_urls[i]).
- **run.sh past the risky part [verified].** `train.log`: `RUN_SH_START → STEP setup → STEP smoke →
  SMOKE_OK → STEP full_train → RUN_SH_UPLOAD manifest=present save_steps=50`. Smoke passed (env +
  model load + train code OK), manifest parsed, armed to upload. First checkpoint PUT lands ~1.5h in
  at save_steps=50; `b2 ls b2://personae-model-training/finetuning/checkpoints/0ac806ab-.../` to confirm
  (note: current b2 CLI needs the `b2://` URI). checkpoints/ + artefacts/ prefixes only appear on first PUT.
- **FOLLOW-UP (non-blocking): `expiry_minutes:3000` override is being ignored.** PUT URLs came back at
  the 24h default (`X-Amz-Expires=86400`), GET at 1h (`3600`) — the ~50h override never applied.
  Isolated: the adapter `ObjectURL` applies ExpiryMinutes>0 and defaults at 0, and the URLs carry the
  defaults, so it arrived as 0 — lost on the DISPATCH side. Suspect `configOrInput` drops the JSON
  NUMBER `3000` via a `.(string)` assertion (string configs like method/key apply; only the numeric one
  fails; happens on BOTH singular and batch dispatch → shared path, predates the batch work). Logged in
  debug guide v2_42. NOT blocking this run (24h > 18h reaper). Fix pending: read `configOrInput` +
  `parsePositiveInt`, then make configOrInput coerce numeric config values (structural) — stopgap is
  quoting `"expiry_minutes":"3000"` in the def.

### Update — 2026-06-09 (5): upload path PROVEN end-to-end (ckpt-0 in B2); expiry_minutes fix written.
- **ckpt-0 confirmed in B2 [verified].** `b2 ls -r b2://personae-model-training/finetuning/` shows
  `finetuning/checkpoints/0ac806ab-71d9-4bd6-af62-ffe0df3b7514/ckpt-0.tar.gz`. That closes the last
  unproven link: a real checkpoint tarball written from the training box via a presigned PUT URL that
  came out of the batch `prepare_object_urls` reply, at the manifest's index-0 key. Full chain
  (launcher presign → manifest → run.sh upload → B2) is proven, not inferred. (The
  `_isolation_test/1780655839/*` objects are the 06/05 standalone test — unrelated.) Expect
  ckpt-1.. every +50 steps (~14 total: eff. batch 8 × ~735 steps / 50), then the final adapter under
  `artefacts/`. Run completion is NOT required to consider the upload path done.
- **expiry_minutes override — FIXED (was the §9 follow-up).** Confirmed from the source: `configOrInput`
  read config via `Config[name].(string)`, so the JSON-number `3000` (decodes to float64) failed the
  assertion → fell through → `""` → omitted → adapter default (24h PUT / 1h GET). Fix: `configOrInput`
  now coerces config scalars via `coerceConfigScalar` (string/float64[integral→no decimal]/json.Number/
  int/int64/bool; non-scalar → "" → falls through unchanged). Rescues `expiry_minutes` + `timeout_seconds`
  + any future numeric config read through the shared helper. `parsePositiveInt` unchanged. Edited file:
  `thunder_ssh_exec_dispatch.go` (added `strconv` import, replaced configOrInput body, added
  coerceConfigScalar — no variable renames). No def change needed; existing `"expiry_minutes":3000`
  works once deployed. Debug guide v2_43.
- **Deploy:** chassis `actions`-package change → rebuild+push chassis, bump launcher (and monitor, when
  wired) def `image_tag`. Future launches only; the in-flight box is unaffected, so this can ship now
  without waiting for the current run. Verify next launch: PUT URLs carry `X-Amz-Expires=180000` (3000 min).

### Update — 2026-06-09 (6): Phase D resume WIRED; monitor decommission branch understood (built, never fired live).
- **Phase D resume — code + def done.** New chassis dispatch `dispatch_thunder_prepare_resume_url`
  (`thunder_prepare_resume_url_dispatch.go`, clone of the singular prepare-object-url dispatch: reuses
  ownResponsesTopic + preRegisterAwaitedRequest), registered in registry.go. Migration 111 APPLIED
  (verified): `presign_final → check_resume → assemble_manifest`; `check_resume` = the resume probe
  (config `{expiry_minutes:3000}` only — training_run_id resolves from input_data via configOrInput
  fallback, NOT a config dot-path); assemble_manifest now reads resume_url/resume_key/resume_index from
  `resume_probe.*`. found=false collapses to a fresh start (assemble emits a resume block only when
  resume_url is non-empty), so ONE launcher workflow serves both fresh and resume launches. Deploy:
  chassis rebuild (resume dispatch + registry) + def image_tag bump, then 111. Adapter side
  (prepare_resume_url) already live.
- **How DONE is signalled (confirmed from PLAN + run.sh build note).** `run.sh` emits `RUN_SH_DONE` to
  `/workspace/train.log` on exit 0; `set -euo pipefail` + 02_train's final-adapter upload HARD GATE
  (final PUT raises → non-zero exit → no DONE marker) means RUN_SH_DONE ⇒ trained AND the final adapter
  is durably in B2 (Phase C, built 2026-06-05). Failure marker: `RUN_SH_FATAL`. The monitor-worker probe
  status_command matches the plan exactly: ALIVE (pgrep 02_train_llama_3_3_70b|/workspace/run.sh) →
  reset_streak; DONE_OK (RUN_SH_DONE + /workspace/adapter_out/adapter_config.json) → mark_complete →
  decommission; DONE_FAIL (RUN_SH_FATAL) → mark_failed → decommission; GONE_UNKNOWN (process gone, no
  marker) → bump_streak, mark_failed at consecutive_unreachable_probes ≥ 3.
- **Monitor decommission branch is BUILT, not missing.** thunder-training-monitor = periodic fan-out
  (find_active_training_instances → per-instance spawn+call of thunder-training-monitor-worker). The
  worker holds probe→classify→mark_complete/mark_failed→decommission, all wired. Only the ALIVE path was
  verified live (2026-06-05); the terminal/decommission branch has NEVER fired live — it triggers on the
  next box that finishes. Gates the PLAN set for enabling it (Phase C: DONE means durable; Phase D resume
  check) are now both satisfied.
- **Correction: training_runs DOES exist — in schema `model_lifecycle`, not public.** The
  thunder_instances FK is `... REFERENCES model_lifecycle.training_runs(id)`. An unqualified
  `\d training_runs` in clients_db/public says "not found" because `\d` only sees the search_path. So the
  run lifecycle is two tables: `model_lifecycle.training_runs` (logical run; mark_training_run_terminal
  writes it) + `thunder_instances` (the box; decommission writes it). The monitor reconciles both. (Logged
  in debug guide v2_44 with the resolution: read the FK schema, `\d model_lifecycle.training_runs`, `\dt *.*`,
  information_schema.tables, or SET search_path.)
- **Unproven in prod:** no run has reached RUN_SH_DONE yet (test box decommissioned early). Checkpoints
  proven (ckpt-0). The final adapter upload (finetuning/artefacts/<run_id>/adapter.tar.gz) shares the same
  presigned-PUT mechanism (ckpt-0 + Phase A isolation test), so low risk — but the first completing run is
  the first live test of BOTH the final upload and the monitor terminal branch. Hard-gate means a failed
  final upload degrades to GONE_UNKNOWN→mark_failed (never a false DONE_OK).

### Update — 2026-06-09 (7): regression — re-running 109 reverted 110+111; restore + guard added.
- **What happened.** During deploy prep, migration `109` was re-applied (the runbook said "single
  idempotent transaction; re-running it is safe"). `109` rebuilds `presign_checkpoints` as a `loop` +
  recreates `flatten_checkpoint_urls`; `110` had replaced that loop with the batch step + deleted flatten,
  and `111` had chained `check_resume` off `presign_final`. So re-running `109` silently REVERTED both —
  the live def went back to the O(K²) loop chain with no resume. Caught by the verify query before any
  launch (presign_checkpoints=loop, flatten present, presign_final→assemble_manifest).
- **Root lesson.** A migration is idempotent only against its OWN prior application, never against LATER
  migrations that mutate the same object. The runbook's blanket "safe to re-run" was wrong; corrected in
  RUNBOOK 2b + logged in debug guide v2_45. Don't re-run an earlier migration to "make sure" — check state.
- **No ledger.** 109/110/111 are hand-applied jsonb mutations to the launcher def, not tracked in a
  schema_migrations table, so the def shape is the only "did it run" source of truth. Added RUNBOOK **2d**
  — a per-migration state-check query (110: presign_checkpoints.action=dispatch_thunder_prepare_object_urls
  + no flatten; 111: check_resume + presign_final.next_step=check_resume) to run after EVERY deploy.
- **Restore (after the chassis with the resume dispatch + configOrInput fix is up + image_tag bumped):**
  re-apply `110` then `111` (in that order; both re-snapshot + guard, safe on the reverted chain), then run
  the 2d check (expect m110_batch_ok=t, m111_resume_ok=t, action=dispatch_thunder_prepare_object_urls,
  flatten_present_BAD=f, presign_final_next=check_resume). Do NOT launch until 2d passes.
- **Structural fix (recommended, prioritise).** File 110 and 111 in the canonical migrations directory in
  number order so a directory-based migrate applies the full ordered set instead of stopping at the last
  filed migration (109). Hand-applied one-offs outside that directory revert on the next deploy — that is
  the mechanism behind this incident; it recurs until 110/111 are filed.
- **Also:** confirm box `0ac806ab` is actually GONE on Thunder, not just a DB row flipped. If a manual
  UPDATE set it terminal without the real decommission (Thunder API DeleteInstance), the VM is still
  billing AND the reaper won't catch a row already in a terminal status. Check `tnr status` / console, or
  fire the real decommission_instance.
- **training_runs schema confirmed** (model_lifecycle.training_runs): status is TEXT with CHECK
  `status IN ('pending','running','complete','failed')` — so the monitor's mark_complete/mark_failed
  values are valid. Not an enum (the enum_range probe errored because there is no enum type). Lifecycle:
  pending → running (launcher mark_running, started_at) → complete (completed_at) | failed.

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
- `prepare_training_data: export <id> has no training rows` → that export's
  `training_exports.rows` is empty even though `runs.rows_exported` is non-zero (recorded
  count ≠ actual rows). Verify with `count(*)` on `training_exports.rows` before launching.
- presign_checkpoints loop substep presigns the DATASET key (every iter's dispatch logs
  `object_key=…/datasets/…/training.jsonl`) / loop_complete results show raw keys with no
  `ckpt_presign` → `presign_one` used `input_mapping` (dead for local loop substeps), so the
  dispatch fell through to its `dataset_uri` fallback. Fix: read the item via `key_path`
  `"ckpt_key"` (109b). Rule: loop substeps read the item via a config dot-path the action
  resolves, NOT input_mapping. [resolved 2026-06-06]
- presign_checkpoints loop progresses then stalls intermittently at a LATER iteration; adapter
  replies (even twice) but the await never clears, retry every ~3 min forever → send-before-register
  race in the local dispatch async path: the dispatch produces the adapter request before the
  coordinator inserts the `awaited_requests` row, so a fast reply beats the insert and
  `ClaimAwaitedRequest` drops it. Fix: pre-register the await in the dispatch before sending
  (`preRegisterAwaitedRequest`, as spawn/call do), or batch the presigns into one adapter call.
  [FIX APPLIED 2026-06-08 in `thunder_prepare_object_url_dispatch.go` (register-before-send);
  pending chassis rebuild + verify]
- Launch never reaches `ssh_exec_launch`; `write_manifest` errored, no `MANIFEST_WRITTEN`,
  no `/workspace/upload_manifest.json` on the box → `write_manifest` used a non-sudo
  `mkdir -p /workspace` but `/` isn't ssh-user-writable. Fix: `sudo mkdir` + `sudo chown`
  (`109a`).
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

- **Training-monitor: BUILT + VERIFIED live 2026-06-04 (both paths).** `running →
  complete/failed` reconcile + release-on-completion is the `thunder-training-monitor`
  orchestrator + `-worker` (106/107/108 + 5 actions). **Terminal/decommission branch
  still never run live** — fires on the next finishing box. **Not enabled**; enabling is
  RUNBOOK step 6, gated on the upload path proving `DONE ⟹ durable`.
- **Upload path (`PLAN_checkpoint_and_artefact_upload_b2.md`): Phases A/B/C BUILT (A
  Tier-1 PASSED); Phase D adapter side BUILT** (reuses `ListObjects`; `prepare_resume_url`).
  The PUT/presign path is no longer "untested" at Tier-1. **LEFT:** D3
  (`dispatch_thunder_prepare_resume_url`) + migration `110`, plus the on-a-real-box
  Tier-2 / B+C integration test and the resume test (`RUNBOOK_phase_b_c_d_deploy.md`).
- The full train runs ~24h (not ~9h — corrected; 1958 examples × 3 epochs ≈ 726
  steps × ~127 s/it) **outside any await window**; nothing watches it but us until
  the monitor is enabled (which is gated on the upload plan landing so `DONE` means
  durable).

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
- Plan/own-handoff: resume needs a NEW "list-keys-under-prefix" storage method ("the ONE
  genuine adapter gap"). **Reality (2026-06-05, [verified-source]):** `storage.Client`
  already declares `ListObjects(ctx, prefix) ([]ObjectInfo, error)` and `*S3Client`
  already implements it (aws-sdk-go-v2 paginator). Reused directly; the narrow
  `objectLister` interface first written was removed. Item-18 blast-radius is about
  **adding** a method to that broad interface — not reusing what is already on it.
