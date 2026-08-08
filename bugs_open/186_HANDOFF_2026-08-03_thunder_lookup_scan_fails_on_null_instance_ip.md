# 186 — thunder_instances lookup scans a NULLABLE column into a non-nullable Go string, so decommission dies before it reaches Thunder

**Filed:** 2026-08-03 · **Status:** FIXED AND LIVE — verified by re-drill 2026-08-08 17:52Z on adapter `v1.0.1267` (file stays in `bugs_open/` per owner ruling 2026-08-06) · **Severity:** was latent; now closed in substance
**Found by:** the reap drill in `docs024_key_docs_latest/finetuning_uk_service/` (see NOTES 2026-08-03)

> **FIX 2026-08-08 (`f83927375`, `Council-Submitted: 862583b1`):** candidate 1
> below, exactly — `InstanceIP`/`RequestedBy` → `sql.NullString`, the two
> ssh_exec readers updated, and `114_thunder_reaper.sql`'s smoke template now
> **omits `instance_ip`** so the drill can produce the failing input.
> `go test -vet=off ./internal/adapters/thunder/` ok (the vet failures at
> `provision_action.go:161` and `api/client_test.go` reproduce on clean HEAD —
> pre-existing). ~~**Stays OPEN until rolled and re-drilled per "How to
> verify"** — the running adapter still carries the defect.~~ **Rolled and
> re-drilled, see below.**

> **VERIFIED LIVE 2026-08-08 17:52Z.** Fleet roll 16:27Z put adapter
> `v1.0.1267` up (image built 14:03 +0100, after the 10:19 fix commit; `make
> build-*` builds committed HEAD). **The usual pod-grep controls are
> structurally unavailable for this diff** — it changes only Go types and
> comments, adding and removing no string literal — so the proof is the
> behavioural re-drill per the 114 template: synthetic overdue row
> `2890abab-9d00-4e95-9bb3-2ce232e0adc7` / `thunder_instance_id` `999999`
> (numeric was safe ONLY because `instances/list` was `{}` at 17:51Z),
> `instance_ip` NULL — **the exact input that produced the Scan error on
> 2026-08-03**. Tick→terminal ~30s: the DB lookup **passed** (adapter log
> 17:52:31 shows the flow continuing to a real authenticated
> `POST /instances/999999/delete` → 404 → "Thunder instance already deleted
> (404)" treated as success), `Decommission complete` with `cost_usd` 3.60
> stamped, row `decommissioned` with `instance_ip` still NULL, reaper
> orchestration run COMPLETED 17:52:34. Drill row deleted (matched `id` AND
> `thunder_instance_id`); table verified back to 23 rows / 23 decommissioned /
> 0 running. Note the row reached `decommissioned`, one state *beyond* the
> `decommissioning` this file's "How to verify" predicts — full terminal
> success, not just past the lookup.

## On the "file it through the loop" ruling (CLAUDE.md, 2026-07-31)

**I did not run `090`, and this states plainly why**, as the ruling's named escape
hatch requires. The claim here is not a theory about where a cause might live: I
**induced the failure on the live cluster**, read the failing line, and then
**isolated the cause by changing exactly one field** and watching the error
disappear. The root cause is a two-line type mismatch in a single function, with
the DB error naming the column. This is the "local and self-evidencing" case the
Diagnosis section exempts — I can watch it fail, change it, and watch it pass.
It is filed here because it is *durable*, not because it is uncertain.

## Symptom

The `thunder-reaper` selects an overdue instance, dispatches
`decommission_instance` correctly, and the adapter refuses it:

```
"error":"decommission_failed","success":false,
"detail":"thunder_instances lookup: sql: Scan error on column index 3,
          name \"instance_ip\": converting NULL to string is unsupported"
```

Status `error_unrecoverable`. **No Thunder API call is ever made** — the failure is
at the DB read, step 1 of 6, so a billing instance would keep billing.

## Root cause

`internal/adapters/thunder/store/instances.go:29`

```go
InstanceIP        string          // ← schema says: instance_ip text NULL
```

`lookupOne` (`:81-96`) scans `instance_ip` (column index 3 of its SELECT list)
straight into that `string`. `database/sql` cannot put NULL into a `string`, so
**any row with a NULL `instance_ip` fails the scan**, and both public lookups go
through it:

- `LookupByID` (`:47`) — used by `decommission_action.go:197` and `ssh_exec_actions.go:127`
- `LookupByThunderIdentifier` (`:70`) — used by `decommission_action.go:207`

The neighbouring fields get this right — `SSHPort` is `sql.NullInt64`,
`TrainingRunID`/`CostUSD`/`ProvisionedAt`/`RunningSince`/`DecommissionedAt` are all
`sql.Null*`. `InstanceIP` was simply missed.

**Second instance of the same class, unexercised:** `RequestedBy string` (`:36`)
against a nullable `requested_by`. Currently 0 of 23 rows are NULL there, so it has
never fired — but it is the identical defect and should be fixed in the same edit.

## Why it has never bitten, and why that is about to change

[verified-db 2026-08-03] All 23 historical rows have `instance_ip` set, because
the only writer — `provision_action.go:408-413` — INSERTs the row **after** the
instance is up, with the IP in hand and `status` hardcoded to `'running'`.

Two things kept it hidden:
1. **No writer ever produces a NULL-IP row**, so production never generated the input.
2. **The documented smoke test cannot reach it.** `sql_for_agents/114_thunder_reaper.sql:186-199`
   builds its synthetic row with `instance_ip = '10.0.0.42'`. The one procedure
   written to prove the reaper works is blind to this by construction.

It matters now because the **plausible recovery action for an orphaned instance is
to hand-insert a row** for a box seen at Thunder — and the operator will often not
know its IP. That is the first realistic NULL-IP row, and it would fail on the one
path that exists to clean it up. See `bugs_open/`-adjacent orphan gap below.

## Fix candidates, ordered by what closes the door

1. **Make the Go types match the schema** (preferred). `InstanceIP sql.NullString`
   and `RequestedBy sql.NullString`, fixing call sites. Makes the bad state
   unrepresentable rather than relying on a writer's discipline.
2. **`COALESCE(instance_ip,'')` in both SELECTs.** One line, no call-site churn —
   but it silently converts "unknown" to "empty", and the next nullable column
   added repeats the bug. A patch, not a structural fix.
3. **`NOT NULL DEFAULT ''` on the column.** Rejected: it lies about the data, and
   "no IP yet" is a real state the adapter should be able to represent.

⚠ Whichever is chosen, **update `114_thunder_reaper.sql`'s smoke test to omit
`instance_ip`**, or the drill will keep passing over the defect it should catch.

## How to verify a fix

Induce it exactly as the drill did (safe: costs nothing, touches no real instance):

```sql
INSERT INTO thunder_instances
  (thunder_instance_id, instance_type, status, max_uptime_hours, hourly_rate_usd,
   ssh_key_secret_name, requested_by, running_since, provisioned_at, created_at)
VALUES ('T-DRILL-<date>', 'drill-not-a-real-instance', 'running', 18, 0, 'none',
        'reaper-drill', NOW()-interval '30 hours', NOW()-interval '30 hours', NOW()-interval '30 hours');
-- instance_ip deliberately omitted → NULL. That is the whole point of the test.
```

Wait one 900s tick. **Before the fix:** adapter logs the Scan error, row stays
`running`. **After the fix:** the row reaches `decommissioning`.
Then `DELETE` the row (match on `id` AND `thunder_instance_id`).

⚠ **Use a NON-NUMERIC `thunder_instance_id`.** Real ids here are bare integers
(`0`, `1`); `decommission_action.go:123-129` parses the id with `strconv.Atoi` and
refuses a non-parseable one *before* calling Thunder, so a non-numeric drill id
cannot delete a real box. A numeric guess could.

## Related, NOT fixed by this

- **The orphan gap.** Because the row is INSERTed only *after* the Thunder instance
  exists (`provision_action.go:408`), a crash between the API create and the INSERT
  leaves a **billing instance with no row at all**. Nothing we run can see it —
  every check reads `thunder_instances`. `api.Client.ListInstances`
  (`internal/adapters/thunder/api/client.go:91`) is built and unit-tested but no
  orchestration action exposes it. Manual check: `finetuning_uk_service/RUNBOOK` §1b.
- **`sql_for_agents/280`** (applied 2026-08-03) widened the reaper's selector to
  three stuck states. Honest scope: `decommissioning` is genuinely reachable
  (`store.MarkDecommissioning`); **`provisioning` is written by nothing** — no Go or
  SQL writer sets it, `provision_action.go:413` hardcodes `'running'` — so that
  branch is defensive only. Do not cite 280 as fixing a live provisioning leak.
