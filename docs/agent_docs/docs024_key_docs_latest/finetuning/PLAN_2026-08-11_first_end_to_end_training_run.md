# PLAN — the first end-to-end training run (2026-08-11)

**Goal, stated so it can be failed:** one training run that reaches
`RUN_SH_DONE` with a final adapter **durable in B2** and a `training_runs` row
that carries real metrics — not a hand-flipped status.

**Why this and nothing else:** the pilot service *is* the run-to-adapter path.
Until one run has finished once, there is no product, the eval report (the actual
deliverable) has nothing to report on, and §3D/§3E/§3F of the handoff are all
downstream. Measured today: **14 runs, 0 finished** — see HANDOFF §9.5.

Owner decision 2026-08-11: plan it, come back before spending GPU money.

---

## 1. What is already true — verified today, not carried forward

| thing | state | evidence |
|---|---|---|
| all 10 training agents | `is_active`, live rows | `agent_definitions`, non-snapshot |
| launcher workflow | complete, 11 steps | `presign_dataset → presign_scripts → compute_keys → presign_checkpoints → presign_final → check_resume → assemble_manifest → write_manifest → ssh_exec_launch → mark_running → complete` |
| migrations 110 + 111 | **still applied** | live config shows `dispatch_thunder_prepare_object_url**s**` (batch) and `check_resume → dispatch_thunder_prepare_resume_url` |
| **the June checkpoint race** | **FIXED and live** | `preRegisterAwaitedRequest` + `prepare_object_url` probe PRESENT in chassis v1.0.1288; negative control absent |
| **the June blocker (broken adapter image)** | **GONE** | `thunder-adapter:v1.0.1288`, ready, **0 restarts**; provenance `bb5348642`, and all three June commits are ancestors of it |
| adapter reachability | working now | `thunder-orphan-scan` 17:46Z → `list_instances success:true` |
| training data | **one verified-good export** | `146a9a12` = 1958 recorded / **1958 actual** |
| spawn→call handshake | much healthier than recorded | 21 timeouts fleet-wide/48h (my memory said 79 in 29h for one agent); 2 since the 17:13Z roll |

**So the two things that stopped this lane in June are both gone.** Start from the
run, not from those bugs.

## 2. What is NOT true — the real gaps

1. **`thunder-training-monitor` is DISABLED and has never once been triggered**
   (`enabled=f`, `last_triggered_at` NULL). Nothing watches a run. The launcher's
   job ends at `mark_running`; the ~24h train then proceeds unobserved, and
   nothing flips `running → complete/failed` or decommissions the box. **This is
   why the one "complete" row has NULL metrics — it was flipped by hand.**
2. **`model_lifecycle.artefacts` has NO WRITER.** It exists only in its own DDL.
   A perfect run today still registers nothing. The registry a pilot would hand a
   customer is unwired — see the correction in HANDOFF §9.5.
3. **The 18h uptime cap is shorter than the 3-epoch run** (~25h measured). June
   worked around it by hand-bumping one instance to 48h. At $1.80/h a 25h run is
   **~$45 against a $30/day cap** — it cannot legally complete as configured.
4. **Every launch silently gets the 70B model** (§3).

## 3. The recommendation: prove it on a SMALL model first

`run.sh` **already supports `BASE_MODEL`**, and its own comment says what it is
for — *"passed through as `--base-model` (e.g. an unsloth 1B/3B for the paid-demo
runs)"*. `02_train_llama_3_3_70b.py` takes `--base-model`; the 70B id is only its
**default**. But the live launcher never sets it, so today every run is
`unsloth/Llama-3.3-70B-Instruct-bnb-4bit`.

Setting it is a one-line change to the launcher's `command_template`:

```
… && chmod +x /workspace/run.sh && BASE_MODEL="<small-model-id>" /workspace/run.sh > /workspace/train.log 2>&1'
```

**Why the first proof should be small, not 70B:**

- **It fits the caps.** A 1–3B LoRA on 1958 rows is hours, not ~25h — inside both
  the 18h uptime cap and the $30/day cap, with no hand-bumping.
- **It costs single-digit dollars**, so a failed attempt is cheap. We are proving
  *the pipeline*, and the pipeline does not care how big the model is.
- **It is the product.** The offer is "a choice of **small** models" (owner ruling
  2026-08-05). A 70B proof would validate a path we do not sell.
- **It de-risks the expensive run.** Every trap below fires identically on a small
  model, at ~1/10th the cost and ~1/10th the wait.
- **It starts §3D's model menu** rather than deferring it: the menu is this field.

**70B stays available** — the default is unchanged for anyone who wants it.

## 4. The sequence

**Phase 0 — pre-flight (free, do all of it before spending anything).** §5.

**Phase 1 — make the small-model run possible.**
1. Migrate the launcher's `ssh_exec_launch.command_template` to set `BASE_MODEL`.
   Agent-def migrations here are **hand-applied, no runner/ledger** — take a
   snapshot, and never re-run an earlier phase5 migration.
2. Rebuild + re-upload `finetuning/scripts/bundle.tar.gz` from
   `working/scripts/` (§5, trap 1).

**Phase 2 — fire one run.** RUNBOOK `iter0_pretrigger(8)` §6 has the exact
envelope; it already names export `146a9a12`. Use `epochs: 1` (the June attempt
had already moved to 1). Watch `train.log` for
`RUN_SH_START → RUN_SH_MODEL → RUN_SH_STEP setup → RUN_SH_SMOKE_OK →
RUN_SH_STEP full_train → RUN_SH_FULL_OK → RUN_SH_DONE`.
**`RUN_SH_MODEL` prints the base model — that is the line that proves Phase 1
took.** The smoke gate fails fast and cheap; most failures die there.

**Phase 3 — prove DONE means durable.** `RUN_SH_DONE` only prints on exit 0, and
`set -e` plus the final-upload raise means it implies *trained AND uploaded*.
Verify the object exists in B2 at `artefacts/<run_id>/` — **at the bucket, not
from the marker.**

**Phase 4 — close the two gaps the run exposes.** Enable
`thunder-training-monitor` (its terminal/decommission branch has still never run
live), and wire an artefact row so the result is recorded. Both are justified by
a real run; neither should be built speculatively first.

## 5. Pre-flight checks — each one is a silent failure

1. **⚠ The scripts bundle is the biggest trap.** `presign_scripts` serves a fixed
   key, `finetuning/scripts/bundle.tar.gz`. If the uploaded bundle predates
   `run.sh`'s `BASE_MODEL` support, **the env var is ignored and the box trains
   70B anyway** — no error, no warning, and you find out ~20h and ~$40 later.
   Re-upload from `working/scripts/` and confirm `RUN_SH_MODEL` in `train.log`.
2. **Never trust `rows_exported`.** `a8484922` records **1957** and actually holds
   **0** — it is what killed run `693656ce`. Count `training_exports.rows`:
   ```sql
   SELECT r.id, r.rows_exported, count(x.id) FROM training_exports.runs r
   LEFT JOIN training_exports.rows x ON x.export_id=r.id GROUP BY 1,2;
   ```
   Use `146a9a12` (1958/1958) or `fef7be6b` (1958/1958). **Not `a8484922`.**
3. **Settle the cost gate before firing**, not after: `daily_cap_usd`=30,
   `estimated_new_run_cost_usd`=20, `max_concurrent_instances`=2,
   `default_hard_uptime_hours`=18, `is_paused`=f. A small-model run fits; a 70B
   3-epoch run does not.
4. **No orchestration dispatch within ~300s of a chassis restart** — the spawn is
   silently dropped. The fleet rolled at 17:13Z today; check before firing.
5. **`kcat -P` can exit 0 having sent nothing**, and sends one message per LINE.
   Count messages at the topic before theorising about a run that never started.
6. **Decommission is manual until Phase 4.** The reaper is enabled (900s) and the
   hard cap is 18h, so a forgotten box is bounded — but bounded at ~$32.

## 6. Cost

| | small model (recommended) | 70B, 3 epochs |
|---|---|---|
| wall clock | a few hours | ~25h (measured) |
| cost @ $1.80/h | **single-digit $** | **~$45** |
| fits 18h uptime cap | yes | **no** — needs a hand bump |
| fits $30/day cap | yes | **no** |

## 7. What I need from the owner before Phase 1

1. **Approve the small-model-first shape** (§3) — or say you want 70B proven
   first, in which case the uptime and daily caps both need raising.
2. **Name the model**, or let me propose one and confirm. It wants to be an
   unsloth 4-bit instruct model in the 1–3B class so the existing
   `load_in_4bit=True` path is unchanged.
3. **Approve the spend** for one run.

Phase 0 (§5) is free and I can do all of it now.

## 8. Not in scope here

The eval report, the model menu as a customer-facing choice, hosting, and the
front end. All are downstream of one run finishing. §3E's ruling stands: follow
the webdesign.uk lane for hosting, do not fork one.

---

*Standing-five doc for the finetuning SERVICE lane. Cadence: update when a
decision or correction lands, not at handoff time.*
