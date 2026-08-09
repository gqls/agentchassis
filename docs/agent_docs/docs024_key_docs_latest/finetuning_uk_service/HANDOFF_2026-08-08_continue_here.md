# HANDOFF 2026-08-08 — finetuning.uk service: Phase −1 complete, start Phase 0 from here

> **SUPERSEDED 2026-08-09 → read `HANDOFF_2026-08-09_continue_here.md` FIRST.**
> This file stays for the Phase −1 background it carries (token rotation
> mechanics, pricing, model menu) — the 08-09 handoff points back here for it.

**This is the COLD-START document for the lane.** Read this, then the RUNBOOK for
commands. The PLAN (`PLAN_2026-07-31_finetuning_uk_service.md`) holds the full
approved design; NOTES holds the evidence trail and every misstep; README is the
owner's plain-prose log. Everything below was verified on 2026-08-08 unless
dated otherwise — re-verify anything load-bearing, this tree moves fast.

## What this lane is

finetuning.uk (already a live Class A site) is to offer a real, paid, demo
fine-tuning service: a few pounds via Stripe Payment Link, customer dataset or
sample, one small-model QLoRA fine-tune on a Thunder Compute GPU, deliverables =
before/after eval + adapter/GGUF download + a booked GPU playground hour or two.
Concierge first, automate later. Full architecture in the PLAN — the front door
is a `finetune` route group in tools-api on the island (`tune.finetuning.uk` →
existing Cloudflare tunnel), the cluster only ever pulls (owner ruling
2026-07-25), Thunder is strictly stop/start with artefacts persisted in B2.

## ⚠ Lane boundary — read before touching ANYTHING site-facing

The site's FRONT END belongs to another thread (session `7b4e88a8-…`,
"finetuning design"; its docs live in `../finetuning_uk_repair/`). It was fixing
broken design/images and re-rendering pages. **This lane is the service backend
only**: Thunder, reaper, training path, island routes. PLAN Phase 1 (authoring
the offer page as chassis sections) is **BLOCKED on coordinating with that
thread** — do not author page content or fire rerenders at finetuning.uk
without checking its state first (`MEMORY_workstreams.md` names the lanes).

Two of that thread's findings that change OUR plan (from its PLAN, read 08-03):
- **Fleet-wide dispatch is dead**: `detected → triaged` promotion lives only
  inside the improvement loop, whose only scheduled route has been `enabled=f`
  since 2026-05-02 (204 detected / 2 triaged across 10 sites). PLAN Phase 2
  routes customer orders through `site_work_items` — **verify the promoter is
  live before depending on it, or drive the `finetune_request` item type with
  its own dispatcher.**
- Their `check_image_url_404` finding is the same "green check that never
  looked" class as our reaper findings. Treat every inherited green check as
  unproven until drilled.

## State of the world (all [verified 2026-08-08] unless noted)

### Thunder token — LIVE, and terraform-owned
- The key is **owned by Terraform root
  `deployments/terraform/environments/production/uk001/047-base-configs`**
  (`kubernetes_secret.personae_default_api_keys`). Its value comes from
  `terraform.tfvars.secret` (gitignored), which is updated from
  `~/.config/thundercompute/token`. **A `kubectl patch` on that secret is
  DRIFT and the next apply reverts it** — that is the entire explanation of the
  07-31→08-05 401 episode (three tokens in play: `5a2…c2` pasted 07-31 and
  never effective; `a73…96` = the old tfvars value terraform kept restoring,
  rejected 401; `f39…ff` = current, live, working).
- **Rotation procedure: RUNBOOK §1c.** Key steps: update tfvars from the token
  file WITHOUT printing the value; fingerprint-compare ALL 19 tfvars values
  against BOTH live secrets before applying (this root also owns
  `personae-platform-secrets` + the prod configmap — a drifted apply reverts
  other people's rotations); expect plan `0 add, 1 change, 0 destroy`; **the
  tool classifier blocks `terraform apply` from agent sessions — the owner runs
  it** (`!` prefix); adapter reads the key at pod start → rollout restart;
  verify from inside the pod, never a laptop.
- Verified: pod env `f39..ff`, `GET /v1/instances/list` → `{}` (authenticated,
  zero instances). Fleet: 23 `thunder_instances` rows, all `decommissioned`,
  `total_24h_spend=0`, `can_provision=t`.

### Reaper — widened AND proven end to end ($0 spent proving it)
- `sql_for_agents/280` APPLIED 2026-08-03 (widens the selector to stuck
  `decommissioning` and NULL-clock cases; verify block returns t/t/t).
  Honest scope: only the `decommissioning` branch is reachable — **nothing ever
  writes `status='provisioning'`** (`provision_action.go:413` hardcodes
  `'running'`, INSERTs after the box is up); that branch + NULL-clock are
  defensive.
- **End-to-end drill PASSED 2026-08-08**: synthetic overdue row, id `999999`
  (numeric was safe ONLY because `instances/list` was `{}` — otherwise use a
  non-numeric id; the Atoi guard in `decommission_action.go:123-129` refuses
  those before calling Thunder). Tick→terminal in 32s: real authenticated
  Thunder delete, `Thunder instance already deleted (404)` treated as success,
  cost stamped, row `decommissioned`. Drill row deleted; table verified back to
  23/0/0.
- **`reaped_at`/`reaped_reason`/status `'reaped'` are DEAD — no writer exists.**
  A successful reap ends at `status='decommissioned'`. Never quote `reaped_at`
  as evidence of anything (WRONG_CALLS 2026-08-08 entry).
- **Reaper run rows**: `orchestration_states WHERE owner_agent_type =
  'thunder-reaper'` (**not** `agent_type` — that column doesn't exist, and the
  error reads as "no rows" if stderr is suppressed).
- ⚠ A drill row left in `decommissioning` is re-selected every 900s forever by
  the widened query — ALWAYS delete drill rows (match `id` AND
  `thunder_instance_id`).

### bugs_open/186 — ✅ VERIFIED LIVE 2026-08-08 17:52Z (section kept for the trail)

> **DONE (evening session, 08-08):** the fleet roll landed 16:27Z (adapter
> `v1.0.1267`, built 14:03 — after the 10:19 fix commit). Pod-grep controls
> were structurally unavailable (the diff adds/removes no string literal), so
> the proof is the behavioural re-drill: NULL-IP row `999999` →
> `decommissioned` in ~30s, lookup passed, real Thunder delete 404≡ok, cost
> stamped, drill row deleted, table back to 23/23/0. Recorded in the bug file
> (stays in `bugs_open/`), NOTES, RUNBOOK §1b. **Next step 1 below is done;
> the lane's next work is Phase 0 (owner-ish supervision) and task #6.**
- Defect: `store.Instance` scanned nullable `instance_ip`/`requested_by` into
  plain Go strings → any NULL-IP row killed the decommission at the DB lookup,
  before any Thunder call. Found by the 08-03 drill (reaper's first-ever
  dispatch was refused exactly there); cause isolated by changing one field.
- Fix: both fields → `sql.NullString`; the two readers in
  `ssh_exec_actions.go:134,148` updated; `114_thunder_reaper.sql`'s smoke
  template now **omits `instance_ip`** (it used to supply one, which made the
  documented drill structurally unable to catch this).
- Committed `f83927375`, council **APPROVED round 1** (`Council-Reviewed:
  862583b1-5f58-47cd-962c-1707c7019cbd`, verdict read and quoted in NOTES).
- **NOT LIVE**: inert until the thunder-adapter image is rebuilt and rolled.
  Owner runs whole-fleet releases (`make release`) — do NOT roll one service at
  its own tag. **After the next roll**: pod-grep verify (a string the change
  ADDED and one it REMOVED, expect 0 — see MEMORY "a roll is not evidence"),
  then **re-run the reap drill with `instance_ip` omitted** (the 114 template
  is now exactly that recipe). Expect the row to reach `decommissioning`
  instead of the Scan error. Then mark 186 fixed-and-live **in the file, which
  stays in `bugs_open/`** (owner ruling 08-06: finished bugs stay there).
- Pre-existing failures NOT mine, left alone: `go vet` fails at
  `provision_action.go:161` (non-constant format string) and
  `api/client_test.go` (`Identifier` field renamed under the test) — both
  reproduce on clean HEAD. Use `go test -vet=off ./internal/adapters/thunder/`.

### Pricing — measured 2026-08-08
- **`GET /v1/specs` carries NO prices** (32 entries, hardware only). Rates from
  https://www.thundercompute.com/pricing:
  **a6000 (48GB) $0.35/hr · l40 (48GB) $0.79 · a100xl (80GB) $1.09 · h100
  $2.19** — per-minute billing, +$0.04/vCPU/hr beyond 4, storage $0.03/100GB/hr
  beyond 100GB, snapshots $0.05/GB/mo.
- **There is NO `l40s` gpu_type** (the PLAN guessed one). Playground GPU =
  **a6000**; a 2h window ≈ $0.70.
- ⚠ `thunder_config` still says `default_hourly_rate_usd=1.80` (live rate
  $1.09) and `estimated_new_run_cost_usd=$20` (sized for 70B) — both stale in
  the SAFE direction (over-refusal). **Correct them only when Phase 0 is about
  to run** (the gate protects every other lane too). SQL in RUNBOOK §1 step 5.

### Model menu (owner: no Chinese models, 07-31)
Mistral 7B v0.3 (Apache 2.0) · Phi-3.5-mini (MIT) · SmolLM2-1.7B-Instruct
(Apache 2.0, replaced Qwen2.5-1.5B). Llama excluded (naming + "Built with
Llama" obligations flow to customers). Nothing about models exists in the site
front end (grep-verified 0 matches) — the menu lives only in NOTES. At Phase 0,
read the LICENSE in the exact HF repo actually downloaded.

### Training scripts
`run.sh` in git takes `BASE_MODEL`/`SAVE_STEPS` env vars (added 07-31,
defaults preserve 70B behaviour). **The live deploy unit is the B2 bundle
`finetuning/scripts/bundle.tar.gz` and it has deliberately NOT been re-uploaded**
— git and B2 intentionally differ until Phase 0 re-tars (RUNBOOK §2; re-upload
IS the deploy; verify by md5 round-trip).

## Next steps, in order

1. ~~**(gated on the next fleet roll)** Verify 186 live: pod-grep both controls,
   re-drill with `instance_ip` omitted, record in the bug file + NOTES.~~
   **DONE 2026-08-08 17:52Z** — see the 186 section above. (Pod-grep was
   structurally unavailable for this diff; the re-drill is the proof.)
2. **Phase 0 — the measured rehearsal (~$1–2, needs owner-ish supervision):**
   - Correct `thunder_config` rates (RUNBOOK §1 step 5) — just before, not
     earlier.
   - Re-tar + upload the scripts bundle (RUNBOOK §2), md5-verify round-trip.
   - One small-model run end to end (SmolLM2-1.7B or Mistral 7B; `--limit`/1
     epoch): record wall clock per stage, $ cost, `RUN_SH_DONE` reached,
     `adapter.tar.gz` genuinely in B2 (`b2 ls`, non-zero size). Test `unsloth`
     template vs `base`+`00_vm_setup.sh` — the ~25 min env build dominates
     short runs (FTW-015).
   - Merge → GGUF on the GPU box, upload to B2.
   - **Playground rehearsal on an a6000**: pull GGUF from B2, serve (try
     Thunder's `ollama` template first), measure provision→first-token and
     tok/s, decommission. This decides the booking shape.
   - **Then** enable `thunder-training-monitor` (it stays `enabled=f` until
     `RUN_SH_DONE ⟹ durable` is proven — its DONE_OK path decommissions the
     box, FTW-035).
   - Phase 0 closes flywheel gates FTW-032/035.
3. ~~**Task #6 — orphan sweep** (the top uncovered risk)~~ **BUILT 2026-08-08
   late evening (FTW-042), commits `81484df8a` + `2ef4ab581` + revision
   `ecbb0f362`, council corr `7ffecfa2`.** Round 1: REVISE, read and fully
   dispositioned (gating objection led to the shared-door `insertWorkItem`
   adoption — see NOTES + WRONG_CALLS). **Round 2 DID NOT RUN: fleet-wide
   Anthropic credit exhaustion from 18:25:48Z killed it at the first seat.
   Once the owner tops up, resubmit with the saved JSON — one command,
   recorded in NOTES (do not rebuild it):**
   `RESUBMIT_CORR=7ffecfa2-… 097_TRIGGER… council_r2_submission_ftw042.json`.
   `list_instances` adapter action + `dispatch_thunder_list` +
   `reconcile_thunder_instances` (orphans filed as `thunder_orphan` items on
   system.internal; ghosts reported not filed; 30-min grace) +
   `sql_for_agents/342` (6-hourly scan). **Remaining: after the next fleet
   roll, apply 342 (image FIRST — the seed header has the verify commands),
   then first-run verification: kick the task, check the COMPLETED
   orchestration's counts against §1b's manual API call the same day.** Until
   then the manual net is still **RUNBOOK §1b** ("AM I BEING BILLED RIGHT
   NOW?" — ten seconds, run it before bed).
4. **Phase 1 — page + payment link**: BLOCKED on coordination with the
   `finetuning_uk_repair` thread (see boundary above). Also owner calls: final
   price (after Phase 0 measurement), playground booking shape
   (customer-named vs offered slots — cold-start time decides), which 3–4
   sample datasets, and registering the island's widened Stripe posture in the
   concept register when Phase 2 ships it.

## Open bugs this lane owns / watches

- **186** (above) — **FIXED AND LIVE, verified 08-08 17:52Z.** Stays in `bugs_open/`.
- Orphan gap — not filed as a numbered bug; it is task #6 + RUNBOOK §1b + the
  concept-register follow-up.
- Adjacent, not ours: `bugs_open/190` (raw LLM envelope in a page_components
  row on one finetuning.uk page — do not rerender that page before it's
  decoded); the dead fleet-wide dispatcher (repair thread's finding, "not mine
  to decide" per their PLAN).

## Landmines specific to this lane (fuller set in LANDMINES.md — grep it)

- **Drill ids are a safety control**: real `thunder_instance_id`s are bare
  integers (`0`,`1`). Numeric drill id ONLY when `instances/list` is `{}`;
  otherwise non-numeric (Atoi guard refuses pre-vendor-call).
- **Delete drill rows** or the widened reaper re-selects them every 900s.
- The 114 smoke template omitting `instance_ip` IS the test for 186 — do not
  "helpfully" add an IP back.
- `orchestration_states.owner_agent_type`, not `agent_type`.
- Never `2>/dev/null` a psql check whose empty output equals its success
  signal (cost me a false "CHANGED" on 08-03 — NOTES misstep entry).
- Terraform owns the secret: `kubectl patch` = drift = silent revert.
- `make build-*` builds from committed HEAD; bump `IMAGE_TAG`; verify at the
  pod with positive AND negative grep controls; no orchestration dispatch
  within ~300s of a chassis pod restart.
- Kafka trigger JSON: flat single line via here-string `<<<'{...}'`.
- `training_exports` run `a8484922…` claims 1957 rows, holds 0 — `count(*)`
  the rows table before launching anything against an export.
- Kubeconfig token expires every 3 days — fleet-wide `Unauthorized` = expiry,
  owner refreshes.

## Commit trail (this lane, newest first)

- `250b752f4` docs: Phase −1 complete read-out (Council-Reviewed trailer)
- `bb99b8e64` WRONG_CALLS: reaped_at proof was unfalsifiable
- `e3c1e2c23` bug 186 file: committed-not-live status
- `f83927375` **fix(186)** — the Go change + 114 template (Council-Submitted)
- `fb68631ed` (08-03) reaper drill round 1: bug 186 filed, LANDMINES entry,
  280 applied, NOTES/RUNBOOK
- `cf54e72c2` (07-31) 280 SQL + RUNBOOK §1b + reaper audit + model menu rev
- `a57371d7c` / `5e729b9da` / `475dfbe38` (07-31) standing five · run.sh
  params · coverage ratchet line

Council correlations: `862583b1-…` (186 fix, APPROVED r1). Tasks in the
harness: #6 orphan sweep (pending), #8 = 186 verify-after-roll (in_progress).
