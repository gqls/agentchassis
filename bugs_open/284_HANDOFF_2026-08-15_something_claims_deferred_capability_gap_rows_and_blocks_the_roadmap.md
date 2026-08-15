# 284 — something claims `deferred` capability_gap rows, turning parked roadmap entries into `blocked` — and it is still happening

**Filed 2026-08-15** by the `bugs_open/279` owner-decision session, spun out of the
`bugfix_213` lane's contribution into 279 (its "candidate 3: stop capability_gap
being claimed at all — that is arguably the root"). **Status: OPEN, evidence
measured, ROOT CAUSE NOT DIAGNOSED — the claimer is unidentified.** Per the
2026-07-31 owner ruling: this file does NOT assert a structural root cause; it
records the measured symptom and the one mechanism read first-hand, and names the
`090` run as the next step. Grep found no existing bug on this mechanism
(`deferred.*claim` hits only 255/279, both different defects).

## The design being violated (first-hand read, both files)

`capability_gap` rows are the platform's "found work I have no handler for" shape
(`bugs_closed/077`). Two producers file them deliberately **non-dispatchable**:

- `discovery_checks/remit.go` `CapabilityGapItem`: status `'deferred'`, **empty**
  `handler_agent`, with the reason IN the spec: *"naming a real agent on an
  undispatchable row is an invitation for someone to promote it — which re-creates
  077 exactly"*.
- `write_audit_findings` (since `d6d56e540`): same pair, for unrouted audit
  categories.

The dispatch loader (`load_work_item_actions.go:651`) selects
`status IN ('triaged','approved')` — `deferred` is not loadable there, and the
`detected`→`triaged` promoter (`triage_detect_items_action.go`) touches
`status='detected'` only. **By every mechanism read so far, a deferred row should
be unreachable by the claim path. It is not:**

## `[MEASURED]` 2026-08-15 — 18 blocked capability_gap rows, and the bleed is ongoing

```sql
SELECT count(*), count(DISTINCT site_id) FROM site_work_items
WHERE item_type='capability_gap' AND status='blocked';           -- 18 rows, 14 sites
SELECT DISTINCT left(error,60) FROM site_work_items
WHERE item_type='capability_gap' AND status='blocked';
-- "No handler_agent set — item cannot be routed to any agent"
```

That error string is `claim_work_item_action.go` (~:165) — the claim path's
empty-handler branch. So each of these rows was **claimed**, found handler-less
(as designed), and stamped `blocked`. The filed→blocked timeline shows it is a
standing mechanism, not an incident: rows filed 07-28 through 08-10 were blocked
on 08-02, 08-03, 08-04, 08-05, 08-08, 08-09, 08-10 and 08-11 — new blocks every
few days, `created_by` spanning `completeness-discovery-agent`,
`design-discovery-agent` and `generic`.

## Why it matters (three costs, each measured or read)

1. **The roadmap loses its entries' meaning.** `deferred` + empty handler is the
   honest "parked, awaiting a builder" state that `diagnose_triage`'s roadmap view
   groups; `blocked` means "work stopped by an error". 18 of the estate's
   capability gaps now read as failures.
2. **Until commit `d6d56e540`'s producer-scope fix, these 18 rows armed a
   site-wide mute** on 14 sites: `write_audit_findings`' broader blocked check
   collapsed to "ANY blocked capability_gap on this site" for its own new
   capability_gap filings (279's contribution section has the full mechanism).
   The mute is now fixed at that one reader; the blocked rows that armed it are
   this bug.
3. **claim burns an attempt on a row designed never to be claimed** — pure waste,
   repeated every few days.

## What is NOT established — the diagnosis gap, stated plainly

**What invokes ClaimWorkItemAction against a `deferred` row is unknown.** The two
live claimers by config census (`agent_definitions LIKE '%claim_work_item%'`) are
`build-dispatch-loop` and `diagnose-dispatch-loop`; the build loader's status list
excludes `deferred` at :651, but the file has other UPDATE arms (:1056, :1072,
:1089) not yet read, and the diagnose loop's loader has not been read at all.
Candidate mechanisms (all `[UNVERIFIED]`): a second loader with a wider status
list; a promoter arm that flips `deferred` under some condition; a workflow
passing item ids to claim directly.

**Next step is a `090` diagnosis run** (symptom: the mechanism above; point at
`site_work_items` rows `item_type='capability_gap' AND status='blocked'`,
`claim_work_item_action.go`'s empty-handler branch, both dispatch-loop agent
definitions and `load_work_item_actions.go`). Filed here first because the queue
check comes first — check `needs_diagnosis` open items before firing.

## Fix candidates, ordered by what closes the door (pending the diagnosis)

1. **Make `capability_gap` unclaimable at the claim path** — a guard in
   ClaimWorkItemAction (or the loaders) that refuses the type, or refuses any
   `deferred` row, rather than blocking it. Closes the door whatever the claimer
   turns out to be; the 213-lane contribution in 279 proposed exactly this.
2. **Repair the 18 rows** back to `deferred` once (1) ships — until then they
   re-block on the next claim. Do not repair before the mechanism is closed.
3. NOT here: the write-time vocabulary residual (`create_work_item` accepts any
   item_type from workflow config) — recorded in 279's status updates; different
   door.
