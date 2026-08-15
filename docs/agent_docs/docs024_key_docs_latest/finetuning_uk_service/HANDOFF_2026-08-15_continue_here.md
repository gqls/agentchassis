# HANDOFF 2026-08-15 — everything is live and staged; ONE step remains: unpause and run Phase 0

> **⚠ SUPERSEDED 2026-08-15 by `HANDOFF_2026-08-15b_continue_here.md` — DO NOT ACT ON
> §2.** The owner authorised the money step and it was **run at 14:26–14:30Z**: one
> a6000 provisioned, both 258 defects proven live, box decommissioned, `is_paused`
> re-armed. Cost $0.0645 (our estimate; see 15b §3 — it is not the vendor's charge).
> **This file's title and §2 are now false.** Two of its figures are also refuted:
> a6000 boot is **~16 seconds**, not `> 5 min` (§2 "the thing nobody knows"), and the
> deadline line renders `wait_timeout:540`, **not** `wait_timeout=9m0s` — grepping for
> `9m0s` finds nothing and reads as "the fix is missing" (WRONG_CALLS 2026-08-15).
> **Still accurate and still worth reading: §3 (verdict, now READ — APPROVED), §4
> (traps), §5 (open unknowns, minus the boot time), §6, §7, §8.**

**COLD-START document for the lane.** Supersedes
`HANDOFF_2026-08-13_continue_here.md` (its blocked items are all done).
PLAN = approved design · RUNBOOK = commands (**§1b** billing check, **§2** bundle,
**§5** provision claims, **§6** sizing + wait deadline) · NOTES = evidence trail ·
README = owner's plain-prose log.

---

## 1. State — everything that was blocked is done

| thing | state, verified 2026-08-15 |
|---|---|
| thunder-adapter | `v1.0.1301`, pod started 10:14:33Z, stamp **`0115f2b45`** |
| **259** claim guard (`10659b419`) | **LIVE** — ancestor of the stamp. Migration 396 applied. Council **APPROVED** (`20d8b725`). |
| **258** defects 1+2 (`236810e4e`) | **LIVE** — ancestor of the stamp, **and** migration 400 applied. |
| `thunder_config.provision_wait_timeout_seconds` | **540s** |
| `dispatch_provision` await | **600s** → invariant OK, 60s headroom |
| 258 council round | **SUBMITTED 2026-08-15**, `d24f9829-0a3f-47a8-bdcb-4b63ced63f1b` — **verdict not yet read** |
| landmines → `doc_notes` | synced; all three of this lane's entries readable |
| `is_paused` | **true** — nothing has been provisioned |
| claim rows / live instances | 0 / 0 |

⚠ **The 44-hour lesson, worth carrying:** the binary shipped on 08-13 **without**
migration 400, so `provision_wait_timeout_seconds` did not exist and the adapter
silently fell back to the compiled-in **5 minutes** — i.e. defect 2 was *unfixed in
behaviour* while every build check said "shipped". A config-backed fix has two
halves and the stamp only proves one. The tell is a WARN line naming the migration
(see §4).

## 2. THE ONE REMAINING STEP — unpause and run Phase 0

The owner chose this sequence on 2026-08-13 ("fix 258 first, then test once") and
steps 1–4 are complete. This is step 5. **It spends money (~$0.04–0.10 for a short
a6000) — confirm with the owner before firing if any time has passed.**

```sql
-- 1. UNPAUSE (this is the money step)
UPDATE thunder_config SET is_paused = false, pause_reason = NULL;
```

Then dispatch ONE provision through `gpu-provisioner`. **Pass NO `vcpus`** — that
is the whole point of 258 defect 1's fix, and overriding it tests nothing:

```json
{"gpu": "a6000", "mode": "prototyping", "training_run_id": "<optional>"}
```

**RE-PAUSE as soon as the test is done**, until Phase 0 is genuinely routine:
`UPDATE thunder_config SET is_paused = true, pause_reason = 'phase0 testing complete, re-armed';`

### What this run has to show — three separate things

1. **258 defect 1.** The adapter logs the derived count *before* any spend:
   `Resolved vCPU count from Thunder specs spec_key=a6000_x1_prototyping vcpus=6 vcpu_options=[6,8]`.
   A `400` from Thunder means the derivation is wrong — read the spec key it names.
2. **258 defect 2.** The box reaches `running` and is **NOT** deleted. Also confirm
   the deadline in use: `Provision wait deadline from live config wait_timeout=9m0s`.
3. **259's live proof — and it is the one you probably will NOT get.** The claim
   guard has never been observed refusing a real re-dispatch. It only fires if an
   await expires, which now needs a provision slower than 600s. **If the provision
   succeeds quickly, the retry driver never fires and 259 still has no live proof.**
   Do not record success here as proof of 259. See §5.

### And measure the thing nobody knows

**How long an a6000 actually takes to boot.** Every prior attempt was killed at 5
minutes while still `STARTING`, so the honest figure is `> 5 min` and nothing more.
This run is the first chance to get a real number. Watch the vendor directly:

```bash
POD=$(kubectl -n ai-persona-system get pods -l app=thunder-adapter -o jsonpath='{.items[0].metadata.name}')
kubectl -n ai-persona-system exec "$POD" -- sh -c \
  'wget -qO- --header "Authorization: Bearer $THUNDER_COMPUTE_API_KEY" \
   https://api.thundercompute.com:8443/v1/instances/list'
```

Record it in NOTES and RUNBOOK §6 — it is the number that says whether 540s is
right, and it feeds the a6000 price question below.

### Then the rest of Phase 0 (unchanged, all staged)

Bundle live in B2, md5 `a19557ccf61ac951c28e81254a8d76f7`; dataset
`finetuning/datasets/phase0-2026-08-12/training.jsonl` (300 rows); presigned PUT
proven. Drive `run.sh` over `ssh_exec` with the four env vars
(`CHAT_TEMPLATE` / `INSTRUCTION_PART` / `RESPONSE_PART` / `SAVE_STEPS`), measure per
stage, confirm `adapter.tar.gz` really lands in B2, then GGUF and the playground
timing. Closes FTW-032/035. Helper scripts (`presign.py`, `build_launch.py`) are
described in NOTES if the scratchpad copies are gone.

## 3. Read the 258 council verdict

```sql
SELECT metadata->>'decision', body FROM diagnosis_artifacts
WHERE correlation_id='d24f9829-0a3f-47a8-bdcb-4b63ced63f1b' AND kind='council_report';
```

The code is already on the shared branch, so a REVISE needs a **follow-up commit**,
not a hold. **Do not write `Council-Reviewed:` on a verdict you have not read.**
**Expect push-back on the wait/await coupling being a stated invariant rather than a
mechanical one** — that is flagged in the submission's own risks deliberately, and
the honest answers are a startup check that reads the `gpu-provisioner` row, or
moving the await timeout into `thunder_config` too.

For reference, 259's round approved with 7 advisory objections, 6 acted on; the
seventh (a success-path race) is recorded in `RFC_026` §6.

## 4. Traps, in the order you will meet them

- **⚠ Raising `provision_wait_timeout_seconds` above the 600s await does NOT give a
  longer wait — it gives a QUIET SUCCESS.** The await expires, the retry driver
  re-dispatches, 259's guard refuses the duplicate *correctly*, and the workflow
  reports **FAILED while a real billed instance runs on with nobody watching it.**
  **Raise the STEP timeout FIRST, then the column.** The tell that it has already
  happened:
  ```sql
  SELECT c.correlation_id, c.attempts, c.status, c.thunder_instance_id
  FROM thunder_provision_claims c
  JOIN orchestration_states o ON o.correlation_id = c.correlation_id
  WHERE o.status='FAILED' AND c.status IN ('created','succeeded');
  -- any row = a box nobody is watching. Check the VENDOR, not just our tables.
  ```
- **A silently-defaulted deadline logs a warning, and that is the only tell:**
  `provision_wait_timeout_seconds not available — using compiled-in default (is
  migration 400 applied?)`. If you see it, the binary has the fix and the database
  does not.
- **A failed attempt KEEPS its claim, by design.** So a correlation that failed can
  never provision again until a human clears it — RUNBOOK §5 has the queries.
  **Never clear a claim in order to "retry"**: re-trigger so it gets a *new*
  correlation. Clearing is only for a genuinely orphaned claim with nothing billing.
- **Check the vendor, not only our tables.** `thunder_instances` gets its row only
  after the box is up, so during boot a live billing instance has no row. FTW-042's
  orphan sweep carries a 30-minute grace for exactly this.
- **`is_paused` blocks every lane, not just this one.** Leaving it off is a
  fleet-wide decision.

## 5. What is still genuinely unknown

- **259 has no live behavioural proof.** The guard is live, unit-tested and
  mutation-proven, and has never fired in production. A deliberate slow case is
  needed — e.g. temporarily lowering `provision_wait_timeout_seconds` toward the
  bottom of its range so a provision reliably outruns the await… **but note that
  doing so deliberately creates the §4 quiet-success condition.** Think it through
  before trying it; safer is to wait for a naturally slow provision and check the
  claim row afterwards.
- **a6000 boot time.** `> 5 min`, nothing more. §2 measures it.
- **The real a6000 price.** Advertised $0.35/hr, but its minimum is 6 vCPU and the
  pricing page charges +$0.04/vCPU/hr beyond 4. Whether $0.35 already assumes 6 is
  `[UNVERIFIED]` — floor is **$0.35–$0.43/hr** until an invoice settles it. Now that
  we provision 6 vCPU by derivation, the first invoice answers this.
- **Does an error response clear an await?** Every await in 259's evidence expired
  on its own clock, never on the adapter's prompt error response — which is *why*
  the retry loop ran at all. **Undiagnosed. Worth a `090`.** `RFC_026` §6.
- **`RFC_026` needs an owner ruling** — the retry driver re-executes side-effecting
  actions fleet-wide (**54 live `call_agent` steps across 33 agents**); how many are
  side-effecting is unclassified.

## 6. Lane boundary (unchanged)

The **front end** of finetuning.uk belongs to the `finetuning_uk_repair` thread
(`7b4e88a8-…`). This lane is service backend only. **Phase 1** (offer page +
payment link) is blocked on coordinating with them. Owner calls still outstanding:
final price, playground booking shape, sample datasets, Stripe posture.

## 7. Two pre-existing problems, neither this lane's

- **`internal/adapters/thunder/api/client_test.go` does not compile at HEAD**
  (`unknown field Identifier in struct literal of type Instance`). Untouched — it
  means `go test ./internal/adapters/thunder/api/` cannot run at all.
- **`adapter.go:393` swallows a reply-produce error** (`silent-reply-drop`).
  Adoption of `DeliverReply` beyond webscrape is RFC-gated (`bugs_open/158` item 1)
  — **do not fix casually.**

## 8. This lane's fleet-wide records

`bugs_open/258` · `bugs_open/259_…_billable_gpus` (resolve **by slug** — a different
259 exists) · `architecture_review/RFC_026` · register **FTW-043**, **FTW-044** ·
five `LANDMINES.md` entries · three `WRONG_CALLS.md` entries · migrations **396**
and **400**, both applied.
