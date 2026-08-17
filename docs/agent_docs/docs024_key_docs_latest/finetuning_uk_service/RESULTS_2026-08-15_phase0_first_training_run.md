# RESULTS — Phase 0 first complete training run, 2026-08-15

The reference statistics from the first end-to-end proof of the finetuning.uk
training pipeline: provision → size → boot → train → upload → **verify at the
artefact** → destroy. Companion to `SUMMARY_2026-08-15_…` (the prose read-out);
full evidence trail in `NOTES_…` §"2026-08-15 evening". These are the numbers
pricing will be set from — **do not re-quote them without their dates**: boot
time in particular is proven day-variable (§4).

## 1. The run

| | |
|---|---|
| Date / operator | 2026-08-15, session-driven (concierge path), owner-approved |
| Model | `HuggingFaceTB/SmolLM2-1.7B-Instruct` (Apache 2.0), ChatML via `CHAT_TEMPLATE=auto` |
| Dataset | `finetuning/datasets/phase0-2026-08-12/training.jsonl` — 300 rows, 3,145,061 B |
| Rows trained | **295** — 5/300 dropped by Unsloth's response-marker filter (truncation), the correct behaviour the 08-12 template fix exists to enable |
| GPU | Thunder a6000 (`a6000_x1_prototyping`), NVIDIA RTX A6000, 49,140 MiB VRAM, driver 610.43.02, `base` template, 100 GB disk |
| vCPUs | **6 — derived live** from `GET /v1/specs` (`vcpu_options=[6,8]`, lowest picked; no `vcpus` passed) |
| Trainable parameters | 18,087,936 of 1,729,464,320 (**1.05%**, LoRA) |
| Batch shape | 1 per device × 8 gradient-accumulation = 8 effective |

## 2. Stage timings (all 2026-08-15 UTC)

| stage | measured |
|---|---|
| Provision dispatched | 16:22:32 |
| Launch (working) | 16:39:00 (`RUN_SH_START` 16:38:58) |
| Setup: venv + torch cu124 + Unsloth | **≈ 5.5 min** (16:38:58 → torch installing at 16:40:30 poll → smoke by ~16:44) |
| Smoke: 20 rows, 1 epoch | `train_runtime` **40.52 s**, loss **1.408** |
| Full train: 295 rows, 3 epochs, 111 steps | `train_runtime` **1363.43 s (22.72 min)**, ≈ 12.3 s/step |
| Final adapter upload (68 MB → B2) | ≤ 1 min (`HTTP 200` box-side, 17:09) |
| `RUN_SH_DONE` | **17:10:01** |
| Decommission → vendor `{}` | 17:13:15 |
| **Provision-to-decommission total** | **≈ 50 min** (includes one failed launch + 11 min of live debugging; a clean re-run ≈ 35–40 min) |

GPU utilisation mid-train (single sample): 32%, 2,498 MiB VRAM.

## 3. Training result

| | |
|---|---|
| Final `train_loss` | **0.730** (smoke baseline 1.408 — halved over 3 epochs) |
| Adapter (`adapter_model.safetensors`) | 72,396,376 B |
| Artefact (`adapter.tar.gz`) | **67,989,958 B** at `finetuning/artefacts/phase0-20260815-1621/adapter.tar.gz` |
| Durability proof | Independent presigned GET **from outside the box**: HTTP 206, `Content-Range: bytes 0-0/67989958`, gzip magic `1f8b`. Box-side `uploaded final_adapter.tar.gz (0.07GB) -> HTTP 200` corroborates. |

**This is the load-bearing result: `RUN_SH_DONE ⟹ adapter durable in B2`, proven
at the artefact.** It closes FTW-032 (June → "deployed, proven-in-prod") and
meets FTW-035's enablement condition (`thunder-training-monitor` may now be
safely enabled — owner switch, still off).

## 4. Cost & boot statistics (feed the pricing decision)

| | booked (flat $1.80/hr) | real (advertised rates) |
|---|---|---|
| Training run (50 min, a6000) | $1.500 | **≈ $0.29** ($0.35/hr) |
| Whole day, 4 boxes | $1.630 | **≈ $0.32** |

⚠ `cost_usd`/`total_24h_spend` are OUR estimate at a flat $1.80/hr for every GPU
type (`provision_action.go:429` → `decommission_action.go:152`) — an upper bound
**4–5× over** for a6000. Safe direction (the $30 daily cap trips early). The true
a6000 rate ($0.35 vs $0.43/hr) is `[UNVERIFIED]` until an invoice.

**Boot times — day-variable by ~20×, never quote without a date:**

| spec | 2026-08-15 | 2026-08-12 |
|---|---|---|
| a6000 | **11–16 s** (twice) | **4m39s / 4m49s still `STARTING`** (twice) |
| a100xl | **12–17 s** | — |

vCPU derivation proven on two specs: a6000 → 6 (`[6,8]`), a100xl → 8
(`[8,12,16]`). Wait deadline in force: 540 s live config, under the 600 s await.

## 5. Defects found by the run (all fixed same hour)

1. `/workspace` absent on the `base` template until setup creates it — launch now
   `mkdir`s first (RUNBOOK §9).
2. `00_vm_setup.sh` VRAM gate hardcoded 79,000 MiB (the 70B assumption) →
   `MIN_VRAM_MIB`, default unchanged (`2094a02e2`). Bundle **deployed to B2
   17:45Z**, md5 `6f27b21a…`, verified via the launcher's own presigned-GET path.
3. `… & echo LAUNCHED` printed success unconditionally while the chain had died —
   ssh_exec reports the *session's* exit, not the chain's. Marker now grouped and
   conditional; read stderr always.

Plus one instrument note: never size a presigned object by HEAD (returns a
163-byte error body against a GET-signed URL) — use `Content-Range` on a range GET.

## 6. What a customer-shaped job costs us (the business read)

A 300-example fine-tune on a 1.7B model: **~50 minutes wall clock, ~$0.30 of GPU.**
Even at 10× headroom for bigger models, longer epochs and retries, unit cost is
pennies-to-low-dollars against a service price in the tens-to-hundreds — the
margin question is operator time, not GPU time. ~~Remaining before pricing: GGUF
conversion + playground-hour timing~~ **BOTH MEASURED 2026-08-17 — §7.** Only the
invoice question remains.

## 7. GGUF conversion & playground rehearsal (measured 2026-08-17)

**GGUF conversion** (a6000, `base` template, attempt 2 — attempt 1 failed, see below):

| stage | measured |
|---|---|
| toolchain + setup (apt update, cmake asserted, venv+torch+unsloth) | ≈ 291 s |
| merge + convert + quantise q4_k_m (incl. llama.cpp build) | **170 s** |
| upload 1.06 GB → B2 | **16 s** |
| **end-to-end on-box** | **489 s (8.2 min)** |

Artefact: `finetuning/artefacts/phase0-20260815-1621/smollm2-1.7b-phase0-q4_k_m.gguf`,
**1,055,609,504 B**, verified at B2 by range-GET (`Content-Range …/1055609504` +
literal `GGUF` magic bytes).

⚠ **Attempt 1's root cause, confirmed by attempt 2: unsloth's
`save_pretrained_gguf(out, …)` writes to `<out>_gguf/`, not `<out>/`, while
printing success naming `<out>`.** Attempt 1's guard looked only in `--out`,
found nothing, and correctly refused to claim success — but the artefact then
died with the box ($3.66 booked / ≈$0.71 real, and the box idled to the reaper's
2 h deadline: **the reaper's first real reap, unattended, correct**). Attempt 2
searches everywhere ≥50 MB and uploads what it finds.

**Playground rehearsal** (a6000, **`ollama` template**, dispatch 18:24:28 Z):

| stage | measured |
|---|---|
| provision dispatch → box ready | **27 s** |
| ollama ready (binary **preinstalled**; service NOT running — started by hand) | 42 s |
| fetch 1.06 GB GGUF from B2 | **12 s** (~88 MB/s) |
| `ollama create` | 18 s |
| cold first token (incl. 38.5 s model load) | **78 s** |
| **DISPATCH → FIRST TOKEN, total** | **≈ 3 m 23 s** |
| warm first token / throughput | **0.36 s · 139.3 tok/s** (cold-request rate 7.5 tok/s) |

**The booking read** (PLAN line 154's `[TO MEASURE]`, now measured): on a
fast-boot day a playground box is conversational **~3½ minutes after dispatch**;
on a slow-boot day (§4: boot varies ~20×) budget **~9 minutes**. So: start the
box ~10 minutes before a booked hour and the customer never sees a cold model.
Warm behaviour during the hour: sub-second first token, ~140 tok/s on a 1.7B.
Template verdict: the `ollama` template carries the binary (saves an install)
but not a running service — the boot script must `ollama serve` itself.

**Phase 0 total spend, all sixteen days of it settled in two sessions:**
$5.72 booked / **≈ $1.12 real** across 7 boxes, every one decommissioned, vendor
`{}` at close. Phase 0 is COMPLETE; pricing is unblocked (invoice still owed).
