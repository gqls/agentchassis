# 091 — a second, different drift is dropped by the work-item dedup while an earlier item is open, and the run reports it as raised

**Filed:** 2026-07-26, by the bugs_open/074 session, **found by inducing a fault** — the same
technique that found 074, in the sweep 074 had just brought to life.
**Severity:** Medium. Nothing is permanently lost (the next pass after the open item is resolved
re-raises), but for as long as the item is open: the durable record names the *wrong* fact, and
the run's own report says a record was created when none was.
**Class:** detected-then-discarded (`bugs_open/071`, `083`) crossed with green-report-for-an-
absent-record — which is the class `074` itself belongs to.
**Status:** OPEN, diagnosed with evidence, **not fixed** — the fix belongs to owners, see below.

---

## Symptom, measured

`refresh_evidence_base` swept leopardessconsulting twice, ten minutes apart:

| pass | drifted fact | new work item? | reported |
|---|---|---|---|
| 18:24:44 | `C4-orchestration-state-records` (live 1,900 vs published 90,790) | **yes**, `needs_human_review` | `work_item_created: true` |
| 18:34:14 | `C1-records-enriched` (live 937 vs published 9,370 — induced) | **no row written** | `work_item_created: true` |

After the second pass there is still exactly one `stale_evidence` row for that site, created at
18:24:45, and its `spec->'drifted'` still describes **C4 only**. The C1 drift exists nowhere
durable — only in that run's `collected_data`.

```sql
SELECT swi.created_at, swi.status, swi.spec->'drifted'->0->>'fact_id'
FROM site_work_items swi JOIN sites s ON s.id = swi.site_id
WHERE swi.item_type='stale_evidence' AND s.domain='leopardessconsulting.co.uk';
--  2026-07-26 18:24:45 | needs_human_review | C4-orchestration-state-records
```

## Mechanism — read in the code, not inferred

1. `createStaleEvidenceItem` keys the item **per site**, not per fact:
   `itemKey: "stale_evidence:" + siteID` (`refresh_evidence_base_action.go:759`).
2. `insertWorkItem` dedups on that key while any non-terminal row exists
   (`load_work_item_actions.go:1148-1152`):
   ```sql
   ON CONFLICT (site_id, item_key) WHERE item_key IS NOT NULL AND status NOT IN (<terminal>)
   DO NOTHING
   ```
   It correctly returns `inserted = rows > 0`, i.e. **false**.
3. **That boolean is then thrown away.** `createStaleEvidenceItem` logs it and returns `nil`; the
   caller sets the report field on the error alone
   (`refresh_evidence_base_action.go:367-371`):
   ```go
   if err := createStaleEvidenceItem(...); err != nil { logger.Warn(...) } else { res.WorkItemCreated = true }
   ```
   So `work_item_created: true` means "no error", not "a record exists".
4. The only trace of the truth is a log line carrying `"inserted": false` — in an ephemeral pod.
   The chassis pod that ran the 18:34 pass was replaced within four minutes of it, taking the line
   with it, which is how a signal that lives only in a log behaves in this fleet.

The dedup itself is right — it exists to stop item spam, and it does. What is wrong is that the
**contents** of the open item are never brought up to date, and that the run reports a write that
did not happen.

## Why it matters even though nothing is lost forever

A human opens the item, reads "1 fact(s) drifted", fixes the C4 copy, marks it complete. The C1
drift was never shown to them. It re-raises on the next sweep — so this is a delay, not a loss.
But the item is also the thing a human uses to decide the site is now *clean*, and while it is
open it is a stale description of a live situation.

## Fix candidates

1. **Refresh the open item instead of dropping the finding.** `ON CONFLICT … DO UPDATE SET
   spec = EXCLUDED.spec, summary = EXCLUDED.summary, updated_at = NOW()` for this item type —
   the row stays single, its contents stay true. Needs care: `DO UPDATE` on a shared helper
   changes behaviour for every caller of `insertWorkItem`, so it belongs behind an explicit
   `workItem` field (e.g. `refreshOnConflict: true`) rather than as a blanket change.
2. **Stop reporting a write that did not happen.** `createStaleEvidenceItem` already receives
   `inserted`; propagate it to `res.WorkItemCreated`. One line, no behaviour change, and it makes
   the report honest whichever way (1) goes.
3. **Key per fact rather than per site** (`stale_evidence:<site>:<fact_id>`). Truest record, but it
   trades one open item per site for one per drifting fact, and the review queue is under an
   explicit "should not FILL" ruling (`bugs_open/033`). Not recommended without that ruling's
   owner.

## Ownership — do not start a competing fix

- The **shared helper** (`insertWorkItem`, and whether a `site_work_items` row means what it says)
  is the `work_item_completion_integrity` workstream's remit.
- The **V4 evidence layer** (`refresh_evidence_base`, `stale_evidence`) is `claims_verification`'s.
- Candidate 2 is contained enough to be safe for either; candidate 1 touches every detector in the
  fleet and should not be done casually.

## How to verify a fix

Not by re-reading the code. Induce it, as this was found:

```sql
-- corrupt one sql-sourced fact so it must drift, on a site that already has an OPEN stale_evidence item
UPDATE site_specs SET data = jsonb_set(data, '{facts,<idx>,value}', '<wrong>'::jsonb) WHERE id = '<current spec>';
UPDATE scheduled_tasks SET last_triggered_at = NULL WHERE name = 'evidence-freshness';
```

Then require BOTH: the open item's `spec->'drifted'` names the newly drifted fact, **and** the
run's `work_item_created` reads false when nothing was inserted. Restore afterwards — the sweep
re-syncs the value itself, so only the work item needs cleaning up.

## Related

- `bugs_open/074` — the case that brought this sweep to life; the induced fault that found this
  was its verification step.
- `bugs_open/071`, `083` — detected-then-discarded siblings.
- `bugs_closed/021` — durable write guard; same question asked of a different write path.
