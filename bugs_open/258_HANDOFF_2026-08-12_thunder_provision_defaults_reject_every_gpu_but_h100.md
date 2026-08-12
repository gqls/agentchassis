# 258 — the Thunder provision path rejects every GPU but h100, and a slow boot destroys the box it just paid for

**Filed** 2026-08-12 by the `finetuning_uk_service` lane, from a live Phase 0
attempt. **Status: OPEN.** Three defects in one code path, found together;
a fixer should take them together. None is site-specific.

**How this was found:** trying to provision the cheapest GPU on the menu
(a6000, the playground's chosen box) through `gpu-provisioner`, for the paid
demo service. Nothing about this is specific to that lane — the same call from
any lane hits the same wall.

---

## Defect 1 — the default `vcpus: 4` is invalid for 9 of the 11 single-GPU specs

`ProvisionInstanceRequest.VCPUs` documents `// default 4`
(`internal/adapters/thunder/provision_action.go:77`), and the adapter sends
`cpu_cores: 4` when the caller does not override it. Thunder rejects that for
almost everything:

```
POST /instances/create -> 400
{"error":"validation_error","message":"validation: invalid vCPU count 4; valid options: [6 8]","code":400}
```

Thunder publishes the valid set per spec, so this is not a guess — it is
`vcpuOptions` on the (already-used, read-only, free) `GET /v1/specs`.
Measured live 2026-08-12, single-GPU entries only:

| spec | vcpuOptions | is the default 4 valid? |
|---|---|---|
| `a100xl_x1`, `a100xl_x1_prototyping` | `[8, 12, 16]` | **no** |
| `a100xl_x1_production` | `[15]` | **no** |
| `a6000_x1`, `a6000_x1_prototyping` | `[6, 8]` | **no** |
| `h100_x1`, `h100_x1_prototyping` | `[4, 8, 12, 16]` | yes |
| `h100_x1_production` | `[15]` | **no** |
| `l40_x1`, `l40_x1_prototyping` | `[6, 8, 12]` | **no** |
| `l40_x1_production` | `[10]` | **no** |

**9 of 11 invalid.** The two that work are h100 — the most expensive GPU on the
menu ($2.19/hr). So with adapter defaults, the only provisionable box is the
dearest one, and every cheaper option 400s. A caller who does not know to pass
`vcpus` cannot provision an a6000, an l40, or an a100xl at all.

**Fix candidates, best first (ranked by what makes the bad state unrepresentable):**
1. **Derive from `/v1/specs`** — the adapter already talks to this endpoint's
   host; read `vcpuOptions` for the requested spec and pick the lowest valid
   value when the caller says nothing. Then no caller can choose an invalid
   count, and the table above can never drift from Thunder again.
2. Validate against `vcpuOptions` and fail with the valid set in the message
   (better error, still needs the caller to know a number).
3. Change the constant 4 to something valid more often — rejected: there is no
   single value valid for all (a6000 needs 6+, a100xl needs 8+, `*_production`
   wants exactly 15 or 10). A constant cannot be right here, which is the point.

## Defect 2 — `waitTimeout` is 5 min, hardcoded; an a6000 does not boot that fast, and the compensation DELETES the box

`NewProvisionAction` sets `waitTimeout: 5 * time.Minute`
(`provision_action.go:141`) with no config path — it is not in `thunder_config`
(schema read live 2026-08-12: the table has no timeout column) and not an env
var. Changing it needs a code change, an image build and a roll.

Observed live, first a6000 attempt:

```
13:52:13  Thunder create accepted        uuid=fi3966m0 (billing starts)
13:52:13  ... "starting cleanup-tracked phase"
13:56:52  vendor still reports STARTING  (4m39s in)
13:57:13  WARN Compensating cleanup starting
          reason="WaitForRunning failed"
          error="wait for instance running: WaitForRunning: context deadline exceeded (instance 0)"
13:57:14  POST /instances/0/delete -> 200
13:57:14  Deleted SSH keypair secret
13:57:15  error_recoverable / infrastructure_error
```

So the box is created, billed for ~5 minutes, and then **destroyed at the
moment it was probably about to become useful** — the caller gets an error and
pays for nothing. The compensation itself is correct and worked exactly as
designed (this is the first time that saga path has fired for real, and it
fired cleanly — worth knowing); the defect is the deadline it is compensating
for.

⚠ **The row is INSERTed only after `WaitForRunning` succeeds** (`insertOnce`,
status hardcoded `'running'`). So for the whole boot there is a live, billing
Thunder instance with **no `thunder_instances` row** — invisible to the reaper
and to every check that reads that table. This is the window `FTW-042`'s orphan
scan carries a 30-minute grace for, and it is fine *provided* the compensation
runs. If the adapter pod dies between create and insert, the box bills until a
human looks.

**Fix candidates:**
1. Make the wait deadline configurable (`thunder_config` column, live, no roll)
   and raise the default. Boot time is a vendor property, not ours to guess.
2. Do not delete on wait-timeout: write the row (status `provisioning`) and let
   the reaper own it — it already covers stuck `provisioning` since
   `sql_for_agents/280`. Turns "destroy the thing we paid for" into "hand it to
   the component whose job is exactly this".
3. Insert the row BEFORE `WaitForRunning`, so the orphan window closes and the
   reaper is the backstop from the first second. Ordering change, needs care.

## Defect 3 — a failed provision leaves NO durable record anywhere

After two failed provisions in one hour, checked live:

- `thunder_instances` — no row (insert is post-wait; see above)
- `agent_error_log` — **no row**. Not a quiet table: 8 other agents logged
  errors in the same window, so it is live and working; the provision path
  simply does not write to it. (Counted the honest way, per LANDMINES:
  `count(*) FILTER (WHERE domain IS NULL)` / `= ''` / `<> ''`.)
- `orchestration_states` — a row, stuck (see below)
- adapter pod logs — the only real evidence, and they rotate

So **"how often does provisioning fail, and why" is currently unanswerable**,
and unanswerable retrospectively — which for a service customers pay for is the
gap that matters most of the three. It also means this bug's own history cannot
be reconstructed: I cannot tell you whether defect 1 has been failing silently
since Thunder tightened validation, or since forever.

## Related observation, NOT diagnosed — an error response may not clear the await

Both failed provisions left their orchestration in `AWAITING_RESPONSES` at
`dispatch_provision`, with `awaited_requests.status='waiting'` and
`processed_at` NULL, ~9 minutes after the adapter had logged
`Sent error response ... error_unrecoverable`. A *successful* provision resolves
its await in ~42s (`RUNBOOK_iter0_pretrigger(8)`, confirmed 2026-06-02), so the
topic wiring works for the success path.

**This is an observation, not a root cause.** [UNVERIFIED] — I did not read the
consuming code, and my chassis-log check was structurally blind: `-l
app=agent-chassis` samples 2 pods of ~34 running that binary
(`LANDMINES.md`), so finding nothing there is not evidence of anything. If it
holds, a failed provision hangs the caller instead of failing it, which for a
paid run is worse than the failure. Worth a `090` before anyone asserts a
mechanism.

## How to verify a fix

```bash
# 1. the specs API is the source of truth for valid vCPU counts (read-only, free)
POD=$(kubectl -n ai-persona-system get pods -l app=thunder-adapter -o jsonpath='{.items[0].metadata.name}')
kubectl -n ai-persona-system exec "$POD" -- sh -c \
  'wget -qO- --header "Authorization: Bearer $THUNDER_COMPUTE_API_KEY" \
   https://api.thundercompute.com:8443/v1/specs' | python3 -m json.tool | grep -A1 vcpuOptions

# 2. provision the CHEAPEST box with NO vcpus override — today this 400s
#    (that is the regression test for defect 1)

# 3. and it must reach 'running' without being deleted (defect 2), so watch BOTH:
kubectl -n ai-persona-system exec -i postgres-clients-0 -- psql -U clients_user -d clients_db \
  -c "SELECT status, instance_ip, hourly_rate_usd FROM thunder_instances ORDER BY created_at DESC LIMIT 1;"
kubectl -n ai-persona-system exec "$POD" -- sh -c \
  'wget -qO- --header "Authorization: Bearer $THUNDER_COMPUTE_API_KEY" \
   https://api.thundercompute.com:8443/v1/instances/list'
# vendor EMPTY + no row = the compensation ate it again.
```

**Cost of reproducing:** defect 1 is free (rejected before anything is created).
Defect 2 costs whatever ~5 minutes of the chosen GPU costs — about $0.03–0.04
on an a6000.

## Workaround until fixed

Pass an explicit valid `vcpus` in `input_data` (it is forwarded when `> 0` —
`thunder_provision_dispatch.go:119`). `{"gpu":"a6000",...,"vcpus":6}` gets past
defect 1. Nothing works around defect 2 except a GPU that boots inside 5
minutes.
