# HANDOFF 2026-08-15b — the provision half of Phase 0 is PROVEN ON REAL HARDWARE; the training half is next

**COLD-START document for the lane.** Supersedes
`HANDOFF_2026-08-15_continue_here.md`, whose title and §2 ("ONE step remains:
unpause and run Phase 0") are now **out of date** — that step was authorised by the
owner and run at 14:26–14:30Z on 2026-08-15. Everything else in that file is still
accurate and it remains the better reference for the traps (§4) and the lane's
fleet-wide records (§8).

PLAN = approved design · **RUNBOOK = commands (§6 wait deadline, §7 how to fire a
provision by hand — NEW, §8 the logs trap — NEW)** · NOTES = evidence trail ·
README = owner's plain-prose log.

---

## 1. State, verified 2026-08-15 ~14:30Z

| thing | state |
|---|---|
| thunder-adapter | `v1.0.1302`, pod started 11:29:23Z, stamp **`194907d5b`** |
| **259** claim guard (`10659b419`) | **LIVE** — ancestor of that stamp (re-checked with a control) |
| **258** defects 1+2 (`236810e4e`) | **LIVE** — ancestor of that stamp; migrations 396 + 400 applied |
| **258 council round** | **APPROVED** 2026-08-15 10:55Z, `d24f9829-0a3f-47a8-bdcb-4b63ced63f1b` — **verdict read** |
| `is_paused` | **true** (re-armed 15:00Z after the credits hold; reason string records why) |
| claims / live instances / vendor | 3 `succeeded` audit rows / **0** / `{}` |
| `total_24h_spend` | **$0.113** (3 test boxes) — our estimate at flat $1.80/hr; real ≈ $0.03 (see §3) |

⚠ **Re-check the stamp before you trust it.** The previous handoff's stamp expired
within three hours when another session rolled the fleet for unrelated work. Ask the
binary, per service: `kubectl -n ai-persona-system logs <pod> | grep -m1 'build
provenance'`, then `git merge-base --is-ancestor <commit> <stamp>`.

## 2. What the test proved, and what it did not

**PROVEN on a real billed machine** (full evidence + timeline in NOTES 2026-08-15
"PHASE 0 PROVISION TEST FIRED AND PASSED"):

1. **258 defect 1.** `Resolved vCPU count from Thunder specs spec_key=a6000_x1_prototyping
   vcpus=6 vcpu_options=[6,8]`, corroborated independently at the vendor
   (`"cpuCores":"6"` on the live box). No `vcpus` was passed in. The old hardcoded
   `4` would have 400'd.
2. **258 defect 2.** `Provision wait deadline from live config` with `wait_timeout:540`
   — live config read, **no** compiled-in-default warning. Box reached `RUNNING` and
   was **not** deleted mid-provision.

**NOT proven — and now much harder to prove: 259.** The claim row shows
`attempts=1, status=succeeded`. The await never expired, so the retry driver never
fired and the guard was never asked to refuse anything. **See §4 — this is now an
owner decision, not a task.**

## 3. Two measurements that change the lane's premises

- **a6000 cold boot was ~16 seconds TODAY — but boot time is DAY-VARIABLE by ~20×.**
  > **CORRECTED later the same day:** this bullet originally said ">5 min does not
  > hold for a6000" and called 540s "~33× the measured need". **The lane's own 08-12
  > notes refute that framing: the same a6000 spec measured 4m39s/4m49s still
  > `STARTING`, twice, three days earlier.** Measured today: a6000 11–16s, a100xl
  > 12–17s (second spec, `vcpus=8` derived — another live proof of 258 defect 1).
  > The honest statement: boot ranges **~16s to >4m49s on different days**, so the
  > 540s deadline is protecting the slow days, not wasting time. Any boot-time
  > claim must carry its date. RUNBOOK §8b.
- **`cost_usd` is OUR estimate, not the vendor's charge.** It is
  `default_hourly_rate_usd` (a flat **$1.80/hr for every GPU type**) × uptime —
  `provision_action.go:429` stamps the rate, `decommission_action.go:152` computes
  from it. The 129s test booked $0.0645; at the advertised $0.35–0.43/hr it was
  really ~$0.0125–$0.0154. **`total_24h_spend` is an upper bound ~4–5× over for
  a6000** (safe direction — the $30 cap trips early). **The real price question from
  the previous handoff §5 is UNCHANGED and still needs an invoice.**

## 4. THE OPEN OWNER DECISION — how, or whether, to prove 259

The claim guard is live, unit-tested, mutation-proven, council-approved, and **has
never fired in production.**

> **CORRECTED later the same day:** this section originally argued a natural firing
> "will never happen on a6000" at a 16s boot. **Withdrawn** — boot time is
> day-variable (§3): on a day like 08-12 (4m49s+ still STARTING) a provision could
> genuinely outrun the 540s wait and the 600s await. So "wait for a naturally slow
> provision" is back on the table as the safe route — it is just unschedulable.
> After ANY provision failure, check the claim row (`attempts > 1` = the guard
> held) before doing anything else.

The deliberate routes remain:

1. **Induce it deliberately** — lower `provision_wait_timeout_seconds` so a provision
   outruns the await. ⚠ **This deliberately creates the quiet-success condition**
   (previous handoff §4): the workflow reports FAILED while a real billed instance
   runs on with nobody watching. Requires someone actively watching the vendor, and
   a plan to clean up by hand.
2. **Try it on a slower spec** (`a100xl`) where >5 min may still be real — costs more
   per attempt and is not guaranteed to be slow either.
3. **Accept it as unproven** and rely on the unit + mutation evidence, recording that
   the live-proof gap is permanent under current boot times.

**Do not pick one of these on a session's own authority.** Route 1 in particular is
the exact failure mode the guard exists to prevent.

## 5. The next actual work — the training half of Phase 0: **STAGED TO THE LAUNCH LINE, held on owner credits**

The owner approved it on 2026-08-15 and a training a6000 was provisioned — then a
credits hold arrived and the box was decommissioned **before any command ran on it**
($0.019, ~37s). Nothing about the staging is lost:

- **RUNBOOK §9 has the complete launch recipe** — presign (via `presign.py`, now
  committed in this directory), the exact ssh_exec command with the env vars and
  real-newline markers, the polling markers, and the B2 durability proof. Verified
  this session: bundle md5 matches (it IS the env-var bundle), both GETs resolve,
  `02_train`'s manifest shape confirmed (`{"final":{…}}` suffices at
  `SAVE_STEPS=0`).
- The presigned URLs minted 2026-08-15 ~14:57Z expire after ~4h — **re-mint, don't
  reuse** (30 seconds with presign.py).
- First move when the owner clears spend: RUNBOOK §9 top to bottom. Measure each
  run.sh stage, prove `adapter.tar.gz` in B2 by presigned GET, then GGUF and the
  playground timing. Closes FTW-032/035.
- `thunder-training-monitor` stays **disabled** until the run proves
  `RUN_SH_DONE ⟹ adapter durable in B2` — its DONE_OK path decommissions the box
  and would destroy the artefact.

## 6. Council follow-up on 258 — approved, with one well-corroborated gap

APPROVED, 8 approve / 1 object, none high-severity, `gated_by_truncation: false`.
The commit carries `Council-Submitted:`, so `098` credits it automatically at report
time — **no amend needed and none permitted** (forward-only).

**Four seats independently objected to the same thing** (`editquality`, `guardian`
at medium, `guidelines`, `architecture`): the wait/await coupling — `thunder_config
.provision_wait_timeout_seconds` must stay below `gpu-provisioner`'s
`dispatch_provision.config.timeout_seconds` — is enforced by prose and a static test
assertion, **not mechanically**. Nothing stops a future session moving one without
the other. The two named remedies are a startup check that reads the
`gpu-provisioner` row, or moving the await timeout into `thunder_config` too.
Measured live 2026-08-15: **540 < 600, 60s headroom, invariant holds.** The §3
boot-time finding lowers the urgency (the margin is enormous for a6000) but does not
close the gap, which protects the *slow* case.

Three cheaper open items were **closed by direct check** this session — no other
production call site hardcodes vCPUs; `GetSpecs` is genuinely new, not a rediscovery;
`LoadConfig` names its columns (no `SELECT *`) so the added column breaks no other
consumer. Evidence in NOTES.

⚠ One residue: **`api.DefaultCPUCores = 4` is an exported constant with no production
consumer**, kept alive only by the mutation test that uses it as the old-wrong-value
sentinel. Its value is invalid for 9 of 11 specs. Legitimate today, and exactly the
shape a future session reaches for. Left untouched deliberately — changing it is a
code edit for the gate, not a drive-by.

## 7. Lane boundary (unchanged)

The **front end** of finetuning.uk belongs to the `finetuning_uk_repair` thread
(`7b4e88a8-…`). This lane is service backend only. **Phase 1** (offer page + payment
link) is blocked on coordinating with them. Owner calls still outstanding: final
price, playground booking shape, sample datasets, Stripe posture.

## 8. Pre-existing problems, neither this lane's (unchanged)

- `internal/adapters/thunder/api/client_test.go` does not compile at HEAD
  (`unknown field Identifier in struct literal of type Instance`) — so
  `go test ./internal/adapters/thunder/api/` cannot run at all.
- `adapter.go:393` swallows a reply-produce error (`silent-reply-drop`). Adoption of
  `DeliverReply` beyond webscrape is RFC-gated (`bugs_open/158` item 1) — **do not fix
  casually.**

## 9. This lane's fleet-wide records

`bugs_open/258` · `bugs_open/259_…_billable_gpus` (resolve **by slug**) ·
`architecture_review/RFC_026` · register **FTW-043**, **FTW-044** · **six**
`LANDMINES.md` entries (the `kubectl logs -l` trap added 2026-08-15) · **four**
`WRONG_CALLS.md` entries (the `9m0s` log-rendering prediction added 2026-08-15) ·
migrations **396** and **400**, both applied.
