# HANDOFF 2026-08-17 — GGUF attempt 1 FAILED (artefact died with the box); playground rehearsal not started; everything else of Phase 0 is DONE

**COLD-START for the lane.** Supersedes `HANDOFF_2026-08-15b_…` (training half:
complete — see its banner). RESULTS_2026-08-15 = the statistics of record.

## State, verified 2026-08-17 ~16:15Z

| thing | state |
|---|---|
| `is_paused` | **true** (re-armed 16:15Z; reason records the failed attempt) |
| vendor `instances/list` | `{}` — nothing billing |
| adapter.tar.gz (LoRA) | **safe in B2**, `finetuning/artefacts/phase0-20260815-1621/adapter.tar.gz`, 67,989,958 B |
| GGUF | **NOT in B2** — attempt 1 failed, artefact lost with the box |
| chassis | **freshly rolled ~14:43Z 2026-08-17** — re-read build provenance before any ancestry claim; **no orchestration dispatch within 300s of a chassis pod start** |
| thunder-adapter pod | `thunder-adapter-85bf4fb6d4-ztbhv`, started 14:43:13Z |

## What happened to GGUF attempt 1 (box `b9358b7e…`, 11:43–13:45Z)

Setup 297s → adapter fetched 3s → `gguf_convert.py` ran, unsloth printed success
(`GGUF_CONVERT_OK out=/workspace/gguf quant=q4_k_m`) — **but wrote NO `.gguf`
into `--out`**. The script's own guard (`ls /workspace/gguf/*.gguf`) caught the
lie and exited `PG_FATAL` before the upload step, so nothing reached B2. Two
leads for attempt 2, in order of likelihood:
1. **unsloth writes the GGUF elsewhere** (model dir / cwd, name like
   `unsloth.Q4_K_M.gguf`) — search `find /workspace -name '*.gguf'` BEFORE
   declaring failure; the fix is likely just widening where the script looks.
2. **`PG_WARN apt install failed`** fired at setup — cmake/build-essential may
   have been missing, so unsloth's llama.cpp build may have failed silently and
   "success" was conversion-skipped. Attempt 2: `sudo apt-get update -qq` first,
   then install, and **assert `command -v cmake`** before the convert; also
   `tail -40 convert.log` on any failure — attempt 1's tail was never read
   (usage-limit gaps), which is why the mechanism is still two hypotheses.

**THE REAPER'S FIRST REAL REAP.** The session went quiet (usage limits); the box
idled ~1h40m after the failure; the reaper killed it at its stamped 2h deadline
(13:45:17Z, cost stamped $3.66 booked ≈ $0.71 real). Every prior reap was a
drill — this was the real thing, unattended, and it worked. Layered defence
held. Worth a line in 016b §durable-invariants / the register when updating.

## To finish Phase 0 (the only remaining items)

1. **GGUF attempt 2** — scripts in the 08-15 session scratchpad are gone with
   it; rebuild from this recipe: RUNBOOK §9-style launch (mkdir /workspace,
   fetch bundle for `00_vm_setup.sh`, base64 the two scripts over, nohup,
   markers `PG_*`). Fix per the two leads above. Upload via the box-side
   presigned PUT (key `finetuning/artefacts/phase0-20260815-1621/smollm2-1.7b-phase0-q4_k_m.gguf`),
   verify at B2 by range-GET `Content-Range` (never HEAD — 163-byte error body).
2. **Playground rehearsal** — provision with
   `{"gpu":"a6000","mode":"prototyping","template":"ollama"}` (template field IS
   supported: `thunder_provision_dispatch.go` forwards it), timestamp the
   dispatch, measure **provision→first-token** (PLAN line 154 `[TO MEASURE]` —
   it sets how named hours must be booked). The rehearsal script from this
   session is preserved below in spirit: fetch GGUF via presigned GET, `ollama
   create phase0 -f Modelfile` (`FROM ./model.gguf`), timed streaming generate
   (first_token, load, tok/s), then a warm second generate. If the template
   lacks ollama, install and RECORD that the template assumption failed.
3. Then RESULTS gets §7 (GGUF + playground numbers), pricing follows.

## Traps for the next session (beyond RUNBOOK §7–§9, all still current)

- **Watcher filters: foreground-test against a line that EXISTS.** Two incidents
  now (WRONG_CALLS 08-15 label-selector, 08-17 impossible-AND) — both mine, both
  read query-failure as system-state.
- The 08-15 presigned URLs and the 08-17 `pg/` URLs are all EXPIRED — re-mint
  everything (`presign.py`, in this directory; creds live from
  `personae-storage-secrets`; `deploy_bundle.py` for bundle deploys).
- ssh_exec responses are read from the ADAPTER POD's producer log — a fleet roll
  orphans pre-roll responses with the old pod. Re-send, don't hunt.
- Costs: `cost_usd` is a flat $1.80/hr estimate, 4–5× over for a6000.
- Phase 0 spend to date: $1.63 (08-15) + $3.66 (08-17 failed convert) booked ≈
  **$0.32 + $0.71 real ≈ $1.03 total**.

## Unchanged

258/259 fixes LIVE + APPROVED; 259's guard still has no live firing (owner
decision, 15b §4); monitor enable = owner switch (gate condition met); bundle
deployed (md5 `6f27b21a…`); lane boundary (front end belongs to `7b4e88a8`);
owner calls outstanding: price, playground booking shape, sample datasets,
Stripe posture.
