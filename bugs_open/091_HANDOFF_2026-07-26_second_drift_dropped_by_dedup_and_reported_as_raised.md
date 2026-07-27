# 091 — a second, different drift is dropped by the work-item dedup while an earlier item is open, and the run reports it as raised

**Filed:** 2026-07-26, by the bugs_open/074 session, **found by inducing a fault** — the same
technique that found 074, in the sweep 074 had just brought to life.
**Severity:** Medium. Nothing is permanently lost (the next pass after the open item is resolved
re-raises), but for as long as the item is open: the durable record names the *wrong* fact, and
the run's own report says a record was created when none was.
**Class:** detected-then-discarded (`bugs_open/071`, `083`) crossed with green-report-for-an-
absent-record — which is the class `074` itself belongs to.
**Status:** **OPEN — candidate 2 done, candidates 1 and 3 still owed.** Committed
`027a8f9e3`, council APPROVED (`a5b70424-b2b5-4d58-aa61-978e8bcf1234`, 11 reviewers,
0 unreadable, 3 advisory objections), inert until the next chassis roll.

> ## STATUS 2026-07-27 (bugs thread) — the report is honest; the finding is still dropped
>
> **Candidate 2 applied, and it turned out to be three sites rather than one.**
> `createStaleEvidenceItem` now returns `(bool, error)` and the caller sets
> `WorkItemCreated` from what was actually inserted, plus a Warn naming this bug when
> drift is found and the open item still describes the earlier drift.
>
> **Two more sites of the identical shape, which this file does not name:**
> `apply_gap_plan_action.go` hardcodes `"item_created": true` in **both** its
> `new_page` and `retype_existing` arms, over an `INSERT … ON CONFLICT DO NOTHING`,
> having discarded the `sql.Result` that knows. Found by grepping every *reporter of a
> creation* rather than only the filed call site — the same discipline that turned up
> the second call site in `bugs_open/103`.
>
> **Nothing branches on these fields**, checked before changing them: no `conditional`
> step in any active `agent_definition` references `item_created`. The six definitions
> that mention it declare it as an `output_field` on the **different**
> `create_work_item` action.
>
> ### Still owed — deliberately not done here
>
> - **Candidate 1** (`ON CONFLICT … DO UPDATE` behind a `refreshOnConflict` opt-in).
>   The second drift is still dropped; the run now says so instead of claiming
>   otherwise. This changes a helper every detector in the fleet calls and is
>   `work_item_completion_integrity`'s remit, exactly as this file says.
> - **Two findings raised by council reviewers**, both for that same workstream:
>   the `guidelines` seat notes `apply_gap_plan`'s `ON CONFLICT DO NOTHING` against
>   `site_work_items` is itself a **WORK-ITEM DEDUP rule violation** (the rule mandates
>   DELETE+INSERT); and `reuse_agent` asks whether those two raw INSERTs should route
>   through `insertWorkItem` and inherit its dedup semantics and its boolean rather
>   than a new helper. Both are probably right and both are refactors with real blast
>   radius, not reporting fixes.
>
> ### After the roll
>
> ```
> kubectl exec -n ai-persona-system <chassis pod> -- \
>   sh -c 'strings /app/agent-chassis | grep -c "an open stale_evidence item already"'
> ```
> Then induce it as this file prescribes — corrupt a second sql-sourced fact on a site
> that already has an OPEN `stale_evidence` item — and require `work_item_created` to
> read **false**. A clean sweep proves nothing: the happy path reported true both
> before and after.

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

## Post-roll triage 2026-07-27 (~15:55 UTC) — unchanged by the roll; framing the decision

The fleet rolled to **v1.0.1174** (`2026-07-27T15:11:15Z`). This bug is **not** one of
the "fixed but inert" class — no fix has been written, so the roll changes nothing.
Re-read at the two cited sites today, both still exactly as filed:

```
platform/orchestration/actions/refresh_evidence_base_action.go:367-370
  if err := createStaleEvidenceItem(...); err != nil { ... } else { res.WorkItemCreated = true }
platform/orchestration/actions/load_work_item_actions.go:1146   ON CONFLICT (site_id, item_key)
```

`grep -n "refreshOnConflict" platform/orchestration/actions/` → **no hits**. Still OPEN,
still unfixed.

**The decision the owner actually needs to make, stated plainly.** The three
candidates are not alternatives at the same altitude — candidate 2 is free and the
other two are not, and they answer different questions:

- **Candidate 2 (propagate `inserted` to `res.WorkItemCreated`) is not a design
  question at all.** It is one line, it changes no behaviour, and it makes the run's
  own report stop asserting a write that did not happen. It is correct under *every*
  answer to the real question below, and it should not wait on that answer. Owned by
  either `work_item_completion_integrity` or `claims_verification`; safe for both.
- **The real question is: when a second, different finding arrives while an item is
  open, should the durable record be updated, or should it stay as first written?**
  Candidate 1 (`DO UPDATE`) says update; candidate 3 (key per fact) says one row per
  finding. Both have a cost the other does not: candidate 1 changes the semantics of
  a helper every detector in the fleet calls, so it needs the `refreshOnConflict`
  opt-in the file already proposes; candidate 3 multiplies rows into a queue that
  `bugs_open/033`'s own owner ruling says **should not fill** — and `033` is measured
  today at **380** parked `needs_human_review` items, still growing (newest 15:14
  today). On that evidence candidate 3 is the weakest, and it is the one that needs
  the `033` owner's consent that it cannot presently get.
- **Sizing:** candidate 2 is minutes of code plus an image window. Candidate 1 is a
  council round plus an induced-fault verification on a shared helper — a session,
  not an afternoon. They should not be bundled: candidate 2's value is that it makes
  the report honest *while* candidate 1 is still being decided.

## Related

- `bugs_open/074` — the case that brought this sweep to life; the induced fault that found this
  was its verification step.
- `bugs_open/071`, `083` — detected-then-discarded siblings.
- `bugs_closed/021` — durable write guard; same question asked of a different write path.
