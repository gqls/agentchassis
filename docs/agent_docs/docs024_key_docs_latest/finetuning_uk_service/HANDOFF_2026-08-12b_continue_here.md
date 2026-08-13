# HANDOFF 2026-08-12b — 259 FIXED (not live), and its filed root cause was REFUTED

> ## ⛔ SUPERSEDED by `HANDOFF_2026-08-13_continue_here.md`
> Still accurate — and the refutation in §1 is the important part of this lane's
> record, so read it. What changed: **259's fix is now LIVE** (thunder-adapter
> `v1.0.1295`, stamp `69612d692`), and **258 defects 1 and 2 are fixed but not
> live** (`236810e4e`). Its "next steps" are done or moved on.

**This is the COLD-START document for the lane.** It supersedes
`HANDOFF_2026-08-12_continue_here.md`, which is accurate about the training side
and the lane's shape but **states the wrong root cause for 259** — do not act on
its "Kafka redelivery" framing or its step-1 implementation notes without reading
§1 below. PLAN has the approved design, RUNBOOK the commands, NOTES the full
evidence trail, README the owner's plain-prose log.

## ⛔ READ FIRST — GPU provisioning is STILL PAUSED, and must stay that way

```sql
SELECT is_paused, pause_reason FROM thunder_config;   -- t | 'phase0 2026-08-12: kafka redelivery ...'
```

The pause blocks **every** lane, not just this one. It stays until the 259 fix is
**in a running binary** — which it is not. A committed fix is not a shipped fix.

> ⚠ The `pause_reason` string still says "kafka redelivery". That wording is
> **wrong** (see §1) and is left as-is deliberately: rewriting it would change
> live state for cosmetic reasons while the containment is doing its job. Correct
> it when you unpause, not before.

## §1 — THE ROOT CAUSE CHANGED. This is the most important thing here.

`bugs_open/259_…_one_provision_request_builds_several_billable_gpus` was filed as
*"Kafka redelivers while the handler blocks"*. **It does not.** Corrected in place
in that file, with evidence.

**What actually happens.** The `dispatch_provision` step carries
`timeout_seconds: 600`. When that await expires, the chassis retry driver
(`coordinator.go` `retryExpiredAwaitedRequest`, budget `RetryVersion < 3`)
**re-executes the step**. Each execution re-runs `DispatchThunderProvisionAction`,
which mints a **new `request_id`** (`thunder_provision_dispatch.go:99`) and
publishes a **new message**. Four executions → four billable GPUs.

**The evidence, and why it is conclusive** (orchestration `8c5bf926`,
correlation `23c9bc6a`): 4 `awaited_requests` rows, **4 DISTINCT `request_id`s**,
1 `orchestration_id`, each row `processed_at` *exactly* its own `timeout_at`, each
next dispatch ~1s later. A redelivered Kafka message replays identical bytes, so
`request_id` would be **constant**. It is not — redelivery is **disproved**, not
merely unsupported.

```sql
-- the whole refutation, in one query
SELECT correlation_id, count(*) AS rows, count(DISTINCT request_id) AS req_ids
FROM awaited_requests WHERE target_agent_type='thunder-adapter'
  AND sent_at > now() - interval '2 days' GROUP BY 1;
-- rows == req_ids  => RE-DISPATCH (correlation_id is the key)
-- rows >  req_ids  => genuine redelivery as well
```

**Consequences you must not miss:**
- **The previous handoff's "raise the Kafka deadlines as an interim" is RETIRED.**
  It addresses a mechanism that is not firing. It would cost a build cycle,
  change nothing, and slow dead-member detection for every consumer on that config.
- **`correlation_id` is the right key — but not for the reason originally given.**
  It is the only identifier stable across the attempts. `request_id` looks
  canonical (per-message, a UUID, the `awaited_requests` PK) and would produce a
  guard that can **never fire**, with a fully green test suite. That trap is in
  `LANDMINES.md`.
- **A co-cause is still unfixed.** The adapter answered ~5 min in
  (`Sent error response`) and the await did **not** clear — it expired on its own
  clock every time. Had the error response cleared it, the step would have failed
  after ONE attempt and built ONE box. This is the "does an error response clear
  an await?" question from the morning handoff: not a curiosity, but *half the
  reason this happened*. **Undiagnosed — worth its own `090`.**

## §2 — What shipped this session

Commit `10659b419`, `Council-Submitted: 20d8b725-f4fc-4b8b-ba58-37606ffddacd`.
Register entry **FTW-043**. Landmine + `WRONG_CALLS.md` entries filed and synced.

`thunder_provision_claims` (PK `correlation_id`), claimed **before** the vendor
call, in one statement (`ON CONFLICT … DO UPDATE … RETURNING (xmax = 0)`) so
claim and count cannot race. A held claim → `ErrProvisionDuplicate` →
`provision_duplicate` / **`error_unrecoverable`** (recoverable would ask the
chassis to retry the thing that builds the second GPU).

Three deliberate calls, all challengeable:
1. **A failed attempt KEEPS its claim.** Every attempt on 08-12 failed and was
   cleaned up, so release-on-failure would leave the loop exactly as it was. Cost:
   a transient vendor failure now ends the workflow instead of self-healing.
   Bounded retry (`attempts <= N`) is the obvious next iteration.
2. **No `correlation_id` → refuse, not provision.** No key, no dedup.
3. **Its own table**, because a pre-create row in `thunder_instances` would be
   filed as a `ghost_row` by our own FTW-042 sweep on every in-flight provision.

Test proven able to fail: mutating the refusal out gives
`CreateInstance called 2 times for one logical request`.

## Next steps, in order

1. ~~**Read the council verdict**~~ **DONE — APPROVED** (`20d8b725-…`), 11
   reviewers, 6 abstained, `unreadable: 0`, `gated_by_truncation: false`,
   "approved with 7 advisory objection(s) — none high-severity". **All seven were
   read and six acted on** — see the council section at the foot of
   `bugs_open/259_…_billable_gpus`. What changed as a result:
   - `classifyProvisionError` extracted + 2 tests pinning that a duplicate
     refusal stays `error_unrecoverable` (mutation-checked).
   - **`architecture_review/RFC_026`** filed — the retry driver re-executes
     side-effecting actions fleet-wide; survey found **54 live `call_agent` steps
     across 33 agents**, not the 3 a narrower query showed.
   - **`thunder_config.pause_reason` rewritten** to name the corrected cause
     (`is_paused` still `true`).
   - **RUNBOOK §5** added: how to clear a stuck provision claim.
   - Migration **396 APPLIED and recorded**, duplicate refusal induced against
     the live table.
   One objection is deliberately unactioned (the success-path race) — recorded in
   RFC_026 §6.

   The original polling queries, if you need them again:
   ```sql
   SELECT created_at, metadata->>'decision' FROM diagnosis_artifacts
   WHERE correlation_id='20d8b725-f4fc-4b8b-ba58-37606ffddacd' AND kind='council_report';
   SELECT body FROM doc_notes WHERE categories ? 'council-gate' ORDER BY created_at DESC LIMIT 1;
   ```
   The code is already on the shared branch, so a REVISE/REJECTED needs a
   follow-up commit, not a hold. **Do not write `Council-Reviewed:` on a verdict
   you have not read.**
2. **Apply migration 396 BEFORE the new binary rolls.** The adapter treats an
   absent `thunder_provision_claims` as a hard error and refuses to provision —
   fails closed, but it *will* refuse.
   `./scripts/migration/run-migrations.sh` (dry run first, per session).
3. **`make build-thunder-adapter` from committed HEAD**, then the whole-fleet
   `make release` — **the OWNER runs the release.** Verify per service, not per
   fleet: `kubectl -n ai-persona-system logs -l app=thunder-adapter --tail=300 | grep -m1 'build provenance'`
   then `git merge-base --is-ancestor 10659b419 <that sha>`.
4. **Then unpause** and re-run Phase 0. Expect the **first a6000 attempt to fail
   once, cleanly** — 258 defect 2 (5-min `waitTimeout` vs a box that needs longer)
   is untouched. It should leave exactly ONE claim row marked `failed` and ONE
   instance at most at the vendor. Verify by counting per **correlation**, and
   independently at the vendor.
5. **Fix 258**, now safer: raising the wait timeout no longer multiplies boxes.
6. **Phase 1** (offer page + payment link) — still BLOCKED on the front-end
   thread (`finetuning_uk_repair`, `7b4e88a8-…`). Owner calls still pending: final
   price, playground booking shape, sample datasets, Stripe posture.

## Unchanged from the previous handoff (still true)

- **The training side is READY.** Bundle live in B2, md5 `a19557ccf61ac951c28e81254a8d76f7`,
  dataset `finetuning/datasets/phase0-2026-08-12/training.jsonl` (300 rows)
  uploaded, presign proven. `BASE_MODEL`/chat-template/masking-marker
  parameterisation and its start-up guard all shipped `270dbfd98`.
  **Training-script changes ship in the B2 bundle, not in any binary** — rolls do
  not affect them in either direction.
- **FTW-042 orphan sweep** — done, council-APPROVED 08-09, nothing owed.
- **Lane boundary** — the site's FRONT END belongs to `finetuning_uk_repair`.
  This lane is service backend only.
- **Open pricing question** — a6000 floor is **$0.35–$0.43/hr** until an invoice
  settles whether $0.35 already assumes the 6-vCPU minimum. `[UNVERIFIED]`.
- **How long an a6000 takes to boot is still unknown** — both attempts were killed
  at 5 min while still STARTING, so `> 5 min` and nothing more.

## Two things found in passing, neither mine to fix

- **`internal/adapters/thunder/api/client_test.go` does not compile at HEAD**
  (`unknown field Identifier in struct literal of type Instance`). Untouched.
- **`adapter.go:393` swallows a reply-produce error** (`silent-reply-drop`, flagged
  by `pattern-check`). Its own note says adoption of `DeliverReply` beyond
  webscrape is RFC-gated (`bugs_open/158` item 1) — **do not fix casually.**
- Fixed in passing because it blocked this work: `fmt.Errorf(reason)` was a
  non-constant format string that **vet rejects, so `go test ./internal/adapters/thunder/`
  could not build at HEAD** — no test in that package had been running.
