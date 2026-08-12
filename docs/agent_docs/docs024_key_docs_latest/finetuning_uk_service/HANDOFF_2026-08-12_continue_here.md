# HANDOFF 2026-08-12 — finetuning.uk service: training side READY, provisioning BROKEN and PAUSED

**This is the COLD-START document for the lane** (supersedes
`HANDOFF_2026-08-09_continue_here.md`, which remains accurate about FTW-042 and
the council trail; the 08-08 one behind it carries the Phase −1 background).
RUNBOOK has the commands, PLAN_2026-07-31 the approved design, NOTES the
evidence and every misstep, README the owner's plain-prose log.

## ⛔ READ FIRST — GPU provisioning is PAUSED fleet-wide

> **A FLEET ROLL TO `v1.0.1291` LANDED 14:55Z 2026-08-12 AND CHANGES NOTHING HERE.
> Do not re-check this; it is measured.** The roll is real (pods restarted), but
> `thunder-adapter` v1.0.1291 is built from **`da5a7eb8f`** — its own build-provenance
> line — which was HEAD *before* this session began, and in any case
> `git diff da5a7eb8f HEAD -- internal/adapters/thunder/ platform/kafka/consumer.go`
> is **empty**. Both defects are still in the running binary because **the fixes have
> never been written** — this session filed them, it did not fix them. A newer tag is
> not a fix. **Stay paused.**
>
> ⚠ And note what a build could never have carried: **this lane's training-script
> changes are NOT in any binary.** `run.sh` / `02_train_llama_3_3_70b.py` ship in the
> **B2 bundle**, so they deploy by re-upload and are unaffected by rolls in either
> direction. Verified live at the artefact after the roll: bundle md5
> `a19557ccf61ac951c28e81254a8d76f7`, unchanged. Do not go looking for them in a pod.

```sql
-- current state, set 2026-08-12 ~14:05Z, OWNER DECISION: leave until fixed
SELECT is_paused, pause_reason FROM thunder_config;   -- t | 'phase0 2026-08-12: kafka redelivery ...'
-- undo, ONLY once bugs_open/259 is fixed:
UPDATE thunder_config SET is_paused = false, pause_reason = NULL;
```

**This is not this lane's private setting — it blocks every lane** from
provisioning a GPU. It is deliberate: `bugs_open/259_…_billable_gpus` — **one provision request
can build several billable GPUs**, with no bound in code. Containment verified
live (2 redeliveries denied, 0 creates, vendor `{}`), not assumed.

## What this lane is

finetuning.uk (live Class A site) is to offer a real, paid, demo fine-tuning
service: a few pounds via Stripe Payment Link, one small-model QLoRA fine-tune
on a Thunder Compute GPU, before/after eval + adapter/GGUF download + a booked
GPU playground hour. Concierge first, automate later. Front door = a `finetune`
route group in tools-api on the island; cluster only ever pulls; Thunder
strictly stop/start with artefacts in B2.

## ⚠ Lane boundary (unchanged)

The site's FRONT END belongs to the `finetuning_uk_repair` thread
(`7b4e88a8-…`). This lane is service backend only. Phase 1 (offer page) is
BLOCKED on coordinating with that thread — do not author page content or fire
rerenders at finetuning.uk without checking `MEMORY_workstreams.md`.

## State of the world

### The training side is READY — this is the good news
Fixed and deployed 2026-08-12 (`270dbfd98`), bundle live in B2 with md5
round-trip verified:
- **`BASE_MODEL` alone was never enough.** The chat template and BOTH
  response-masking literals were hardcoded Llama. `--chat-template` /
  `--instruction-part` / `--response-part` now move with it, forwarded by run.sh
  as `CHAT_TEMPLATE` / `INSTRUCTION_PART` / `RESPONSE_PART`; every default is
  the previous literal, so a 70B run is byte-identical (verified both ways).
- A **guard** refuses to start if the markers do not occur in the first 25
  rendered rows — proven offline to fire on the old markers and pass on the
  corrected ones.
- `SAVE_STEPS` default restored to **10** (the live bundle's real prior value;
  it had been written as 50 under a claim of "identical to before").
- SmolLM2-1.7B-Instruct verified from its own `tokenizer_config.json`: ChatML,
  `<|im_start|>user\n` / `<|im_start|>assistant\n`. Licence in that exact repo:
  **Apache 2.0** — the standing Phase 0 licence obligation is discharged.

**Ready and waiting in B2:** `finetuning/datasets/phase0-2026-08-12/training.jsonl`
(300 rows, every row carries a user+assistant pair). Presigned PUT proven with a
real round-trip. Helper scripts kept in the session scratchpad are trivial to
rebuild — `presign.py` (mint URLs + prove the PUT) and `build_launch.py` (compose
the ssh_exec command with the env vars the launcher cannot yet supply); both are
described in NOTES if they are gone.

### The provisioning side is BROKEN — four defects, two bug files
- **`bugs_open/259_…_billable_gpus` — one request, several billable GPUs.** THE BLOCKER.
  Synchronous handler blocks 5 min in `WaitForRunning`; consumer deadlines are
  60s; offset never commits; message redelivered; another box built. Measured:
  2 requests → 3 GPUs. **Fix idempotently on `correlation_id`** — raising the
  consumer timeouts only moves the boundary.
- **`bugs_open/258` defect 1 — default `vcpus: 4` is invalid for 9 of 11
  single-GPU specs.** Only h100 (the dearest) provisions on defaults. Workaround:
  pass `"vcpus": 6` (a6000) / `8` (a100xl, l40) in `input_data`.
- **258 defect 2 — `waitTimeout` is a hardcoded 5 min**; an a6000 does not boot
  that fast (measured twice: 4m39s, 4m49s still STARTING), and the compensating
  cleanup then deletes the box. Not in `thunder_config`, not an env var — needs
  a build + roll.
- **258 defect 3 — a failed provision leaves NO durable record**: no
  `thunder_instances` row, no `agent_error_log` row. Unauditable.

⚠ **The two interact:** fixing 258's timeout by RAISING it, before fixing 259,
makes 259 worse (longer block = more redelivery).

### FTW-042 orphan sweep — DONE, council-APPROVED (08-09), unchanged
Live, verified, nothing owed. See the 08-09 handoff for the full trail. Its
30-minute grace is what makes the boot-window orphan (below) tolerable.

⚠ **Newly understood this session:** the `thunder_instances` row is INSERTed
only **after** `WaitForRunning` succeeds, so for the whole boot there is a live,
billing instance with **no row** — invisible to the reaper and to every check
reading that table. Fine while the compensating cleanup runs; if the adapter pod
dies between create and insert, the box bills until a human looks.

## Next steps, in order

1. **Fix `bugs_open/259_…_billable_gpus`** (idempotent create keyed on correlation), then 258.
   Both need a thunder-adapter build + roll — **owner runs the fleet release.**
2. **Do NOT wait for a 090 verdict on 259 — there isn't one, and two runs were
   already spent.** Run 1 died on a fleet-wide API 529; run 2 completed with
   **5 bundles and no verdict** (the body-budget trap, hit despite every named
   file being well under the 60KB the landmine says to check — the budget is
   cumulative, and `LANDMINES.md` is corrected). 259 records what was
   substituted instead, per the 2026-07-31 escape hatch. **If you want the
   broker-side trigger named, re-file with ONE symbol in scope**
   (`handleProvisionInstance` alone), and read the budget line ~2 min in rather
   than waiting 30:
   ```sql
   SELECT substring(body from '_\(body omitted[^)]*\)_') FROM diagnosis_artifacts
   WHERE correlation_id='<RUN_CORR>' AND kind='bundle' AND body LIKE '%body omitted%';
   -- non-empty => no verdict is coming; re-file narrower.
   ```
   **You do not need it to fix this.** Fix candidate 1 (idempotency on
   `correlation_id`) is correct whichever trigger it turns out to be.
   **Implementation notes for the 259 fix, so a fresh session need not re-derive
   any of it** (full reasoning in the bug file; this is the short form):
   - **Where.** `internal/adapters/thunder/provision_action.go` — `Execute` does
     gate checks (`cfg.IsPaused` :156, `CheckCanProvision` :164) and then calls
     `p.thunderAPI.CreateInstance` (:228). The idempotency check belongs
     **between those**, i.e. after the gate and before the vendor call.
   - **What.** The request already carries a correlation — the adapter reads
     headers in `handleMessage`/`handleProvisionInstance` (`adapter.go` :294–:325)
     and sets `req.RequestedBy` from `sender_agent_type` (:447), so plumbing one
     more header field through is the same shape as an existing one. Persist it
     **before** the create (a `correlation_id` column on `thunder_instances` with
     a UNIQUE index is the smallest change that makes a duplicate
     *unrepresentable* rather than merely detected), then treat a conflict as
     "already provisioned — return the existing row".
   - **Why not just widen the timeouts.** `KAFKA_SESSION_TIMEOUT` /
     `KAFKA_REBALANCE_TIMEOUT` are env-tunable (`platform/kafka/consumer.go` :56)
     and need no roll, so they are tempting. They move the boundary; they do not
     remove it, and they slow genuine dead-member detection for **every** consumer
     on that config. Fine as an interim, not as the fix.
   - **Process.** This is `internal/` — **platform code, so put it through the
     council gate** (`097_TRIGGER_council_review_v1.sh`, ~30 min, budget for it).
     Then `make build-thunder-adapter` **from committed HEAD**, and the roll is a
     whole-fleet `make release` which the **owner** runs.
   - **Regression test that can fail.** The check must drive a provision slower
     than the consumer deadline, which today is any provision at all: count
     **deliveries per correlation** in the adapter log (query in the bug file),
     and independently confirm at the vendor that one request left **at most one**
     instance. Counting creates or rows instead will undercount in both
     directions.

3. **Unpause**, then re-run Phase 0 — the training half needs no further work.
   Everything is staged: bundle deployed, dataset uploaded, presign proven.
   Provision a6000 with `vcpus: 6`, drive `run.sh` over `ssh_exec` with the four
   env vars, measure per stage, confirm `adapter.tar.gz` really lands in B2, then
   GGUF and the playground timing. Closes FTW-032/035.
4. **Phase 1** — page + payment link: BLOCKED on the front-end thread. Owner
   calls still pending: final price, playground booking shape, sample datasets,
   registering the island's widened Stripe posture.

## Open questions this session could not answer

- **The real a6000 price.** Advertised $0.35/hr, but its minimum is 6 vCPU and
  the pricing page charges +$0.04/vCPU/hr beyond 4. Whether $0.35 already
  assumes 6 is **[UNVERIFIED]** — so the floor is **$0.35–$0.43/hr** until an
  invoice settles it. (An earlier claim of "$0.43, not $0.35" was arithmetic on
  an assumption and is corrected in NOTES.)
- **How long an a6000 actually takes to boot.** Both attempts were killed at
  5 min while still STARTING, so we know it is `> 5 min` and nothing more.
- **Whether an error response clears an await.** Three orchestrations from this
  session sit in `AWAITING_RESPONSES` with `awaited_requests.status='waiting'`
  long after the adapter logged `Sent error response`. Observed, **not
  diagnosed** — noted in 258, worth a `090` of its own.

## Where everything is

- **This dir:** PLAN · RUNBOOK (§1 token/§1b billing check/§1c rotation/§2
  bundle/§3 queries) · NOTES (full 08-12 evidence trail) · README (owner log) ·
  council JSONs · SUMMARYs.
- **Code:** `internal/adapters/thunder/` · `platform/orchestration/actions/thunder_*.go`
  · training scripts in `docs/agent_docs/docs024_key_docs_latest/finetuning/working/scripts/`.
- **Fleet-wide records written this session:** `bugs_open/258`, `bugs_open/259_…_billable_gpus`,
  two `LANDMINES.md` entries (synced, verifier fired), one `016b` §9 pattern.
