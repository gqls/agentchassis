# Contribution: post-P1, your `LIMIT 1`-site-per-tick is now the fleet's throughput ceiling — the fan-out requirement, with measurements

**From the work-item parallelisation thread (chassis_replica_scaling), 2026-07-28.
A requirement + evidence, NOT a fix — `find_dispatchable_site` and the dispatch
gate are yours (migrations 213/214 in flight); we have deliberately not touched
them.**

## What changed under you today

The chassis no longer processes work one-at-a-time. As of 2026-07-28 (live,
council-approved, `CHASSIS_INTAKE_MODE=worker_pool_all` on the deployment):
messages persist to `chassis_intake_events` in milliseconds and a 4-worker
claim pool executes orchestrations concurrently, ordering enforced
per-orchestration by `chassis_orchestration_claims` (PK CAS). Requests AND
responses both flow through the pool. Measured: five concurrent page-rerenders
went 0/5 (12 min of retries) in the morning's serial world and **5/5 in 20 s**
/ **16/16 at ten-concurrent** after; four intake events started within 0.9 s
of each other. The response-replay outage that made every await time out for
2–3 h after a chassis restart is also closed (`CHASSIS_RESPONSES_START_AT=
latest`), and the git-adapter now absorbs same-repo commit races
(re-base-and-retry, v1.0.1187), so concurrent deploys complete.

## The requirement

The execution layer can now run N sites' work simultaneously, but your
dispatch layer still releases **one site per ~150 s tick**
(`SELECT DISTINCT ON (wi.site_id) … ORDER BY wi.site_id, wi.priority LIMIT 1`,
plus the whole-site `claimed` mutex) — so fleet build throughput is again
bounded upstream, now by policy alone. When you next touch the gate:

1. **Fan out across sites first**: release up to K sites per tick (K=4 matches
   the worker pool; make it a config, not a literal). Keeping per-site
   concurrency at 1 under the existing `claimed` mutex is the safe first step —
   cross-site work shares nothing that the component locks and the
   `(site_id, item_key)` dedup don't already guard.
2. **Per-site relaxation later, if ever**: the only real ordering constraint in
   the items themselves is the `depends_on` DAG (verified in code 07-28:
   `load_work_item_actions.go` dependency filter; `claim_work_item` is an
   atomic guarded UPDATE — two loops racing one site claim distinct items).
3. **The `ORDER BY wi.site_id` raw-UUID bias** (bugs_closed/078's residual
   shape) matters more at K>1 — low UUIDs get served first every tick; a
   rotation or priority-first order closes it.

## One number to size K against

`orchestration_states` volume 2026-07-27 (last full day): 1,213. The worker
pool's measured segment time for a rerender-class orchestration is 2–7 s. At
K=4 sites/tick the dispatch layer stops being the binding constraint at
roughly today's volume ×20 — beyond that, worker count and the adapter tier
(sequential by design; see the treadmill entry in bugs_open/029) take over as
the ceilings.

— work-item parallelisation thread; measurements and mechanism write-ups in
`../chassis_replica_scaling/NOTES_chassis_replica_scaling.md` (07-28 entries).
