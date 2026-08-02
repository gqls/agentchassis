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

> ## STATUS 2026-08-03 (bugs-sweep lane) — **CANDIDATE 1 BUILT.** And the severity in the header above is too low: the record is wrong on 4 of 5 sites TODAY
>
> **Workstream:** `docs024_key_docs_latest/bugfix_091_workitem_conflict_refresh/`
> **Council:** `Council-Submitted: 8e7357ae-9f8d-49bf-81c0-669d9a97a205`
> **Still OPEN** — Go code, so it is inert until the next chassis image. Do not close
> it on the commit; the defect stays reproducible until the roll (`bugs_closed/` bar).
>
> ### The header says "Medium … a delay, not a loss". Measured today, that undersells it
>
> `evidence-freshness` is **enabled** and ran at **2026-08-02 18:36:07Z**. Comparing what
> each open `stale_evidence` item SAYS drifted against what that run FOUND (query in the
> workstream RUNBOOK — note `orchestration_states` keeps ~24h, so this is only
> reproducible on the day):
>
> | site | item filed | the item says | the live run found | correct? |
> |---|---|---|---|---|
> | leopardessconsulting.co.uk | 07-26 | `C4-orchestration-state-records` | `C4-agent-definitions-catalogue` | **NO — a different fact** |
> | fundamentallyai.com | 07-26 | F11, F12, F13 | F9, F10, F11, F12, F13 | **NO — 2 drifts invisible** |
> | ai-agent-orchestration.com | 07-26 | 3 facts | 2 (one re-synced) | **NO — over-reports** |
> | vonc.com | 08-01 | `vonc-tools` | *(nothing)* | **NO — describes drift that is gone** |
> | oufe.com | 07-27 | 12 × CIT-* | the same 12 | yes |
>
> **Four of five open items are factually wrong.** The handler is `human-review`, so the
> only consumer is a person — and on leopardess that person is sent to a fact that never
> moved. "A delay" is true of the FINDING; it is not true of the RECORD, and the record is
> the artefact. Candidate 2 is visible working in the same run: all four drifting sites
> reported `work_item_created: false`, honestly, while dropping 20 facts between them.
>
> ### What was built — candidate 1, with three deliberate deviations from this file
>
> `conflictPolicy` on the shared writer: `dropOnConflict` (default, byte-identical to
> today) and `refreshOnConflict`, which updates the open row's `summary`/`spec`. One
> caller opts in. Registered as **BATCH-005** in the concept register, with its two
> landmines, in the same commit that ships it.
>
> **1. NOT `ON CONFLICT … DO UPDATE`, which this file proposes — it would have re-created
> candidate 2's defect inside candidate 1's fix.** `DO UPDATE` affects a row, so
> `RowsAffected()` returns 1, and `insertWorkItem` returns `rows > 0`, which is what
> `work_item_created` is set from. The literal fix makes the run start reporting a
> creation that never happened again. Instead: a separate `UPDATE` in the conflict branch
> (the shared INSERT stays byte-identical for ~20 callers) and a three-state outcome,
> `workItemWrite{Inserted, Refreshed}`. `work_item_created` still means *created*;
> `work_item_refreshed` is a new, separate field.
>
> **2. NOT a `workItem` field, which this file also proposes.** As a field, a caller can
> set it and still call `insertWorkItem`, whose single bool cannot express a refresh — a
> silent wrong answer at the call site most likely to be copied. It is a **parameter of
> `writeWorkItem`**, so the old function cannot receive it and the mistake does not
> compile.
>
> **3. The refresh has TWO guards this file does not mention**, both in the UPDATE's own
> predicate, which is also how the unlocked gap between the two statements is lost safely:
> a row that went terminal in between does not match (never resurrected), and a row a
> handler HOLDS (`claimed`, `diagnosing`) is skipped (its spec is not changed underneath a
> running handler). Only `summary` and `spec` are written — `status`, `priority`,
> `handler_agent`, `severity` may all have been moved by a human.
>
> ### The two council objections this file records as owed are both answered
>
> `apply_gap_plan_action.go`'s three hand-rolled `INSERT … ON CONFLICT DO NOTHING`
> statements now route through `insertWorkItem` (the `guidelines` and `reuse_agent` seats
> on `a5b70424`). **The reason they forked is the finding worth keeping:** they set
> `parent_item_id` and the shared `workItem` struct had no field for it, so the shared
> door was unusable to anyone needing a parent. The field exists now. They adopt with
> `recurrenceExpected: true`, which is what makes the adoption behaviour-PRESERVING: a gap
> plan asking for a page to be built is an action request, and without the flag adoption
> would newly suppress items within 3h of a terminal predecessor and brand them
> `unresolved` after two — `bugs_open/024`'s regression, re-created by an unrelated fix.
>
> ### Candidate 3 stays refused, and the reason is now sharper
>
> One row per fact would multiply into `needs_human_review`, which `bugs_open/033`'s owner
> ruling says must not fill. Re-measured 2026-08-02: **368** parked rows (this file
> recorded 380 on 07-27). The refresh keeps the row count at exactly one per site.
>
> ### Guards are MUTATION-PROVEN, not merely present
>
> Four mutations run, each confirmed to fail the suite: policy check removed (2 tests),
> held-status clause dropped (1), a refresh reporting `Inserted: true` (1),
> `recurrenceExpected` cleared (1). **The fourth exposed a harness blind spot worth more
> than the mutation:** the anti-churn probe discards its own error
> (`if err == nil && terminalCount > 0`), so a sqlmock test that does not expect the probe
> still passes — **no behavioural test in this package can see a change to
> `recurrenceExpected`**. Filed as a landmine; it is why that adoption is covered by a
> direct assertion on the built item.
>
> ### What is still owed on this file
>
> - **The roll**, then this file's own §"How to verify a fix" — induce a second, different
>   drift on a site with an open item and require the open item's `spec->'drifted'` to name
>   the NEW fact while `work_item_created` reads false and `work_item_refreshed` reads
>   true. **A clean sweep proves nothing.** Pod-grep markers are in the workstream RUNBOOK.
> - **The council verdict** on `8e7357ae`, and the one judgement I want challenged:
>   `needs_human_review` is deliberately NOT a held status, so a refresh CAN change an item
>   under a human who is reading it. It is a queue, not a claim, and the alternative is
>   leaving the record false — but that is a judgement, not a measurement.
> - **THREE SIBLING CALL SITES IN EXACTLY THIS SHAPE, DELIBERATELY NOT SWITCHED.** All
>   HITL-terminal, all carrying a list in `spec` that will differ next run:
>   `evidence_citations.go:426` (`citation_unverified:<site>`, spec `rejected[]`);
>   `directory_claims.go:333` (`directory_citation_unverified`, constant key);
>   `directory_claims.go:575` (`stale_directory_claim`, constant key — its summary even
>   carries a count of the findings it is about to drop). Each needs the same judgement
>   made on its own evidence. Building the capability and switching one caller is not the
>   same as fixing the class, and this list is the honest statement of what is left.

> ## STATUS 2026-07-27 (bugs thread) — the report is honest **and LIVE on v1.0.1177**; the finding is still dropped
>
> **Rolled 19:22:02Z.** Verified in the running pod:
> `grep -c "an open stale_evidence item already"` → **1**.
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
