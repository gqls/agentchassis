# HANDOFF 2026-08-13 — bugfix 213, continue here (supersedes the 08-12 handoff)

**Read this, then `NOTES…md` (last two sections) for the mutation matrix and the missteps,
then `RUNBOOK…md` (last two sections) for the two commands you actually need.** The 08-12
handoff stays accurate on the D1 *measurements*; it is superseded on "what happens next"
and on one claim it made about the sibling lane's option transferring cheaply here.

---

## THE ONE-PARAGRAPH STATE

D3 is done and live. **D1 half one is BUILT and COMMITTED (`96c53bc18`) and is INERT until
the next chassis roll** — completion gate 1b, which refuses to stamp `complete` on a
`dark_section_audit` item when the handler's own payload says it changed nothing. It is
opt-in per `item_type` with the unsafe default OFF, registers no verifier, and touches no
registry. **D1 half two (retraction at the next audit) is specified and deliberately NOT
started**, because a precondition nobody had noticed turned up: this detector is an LLM, so
its silence is not a measurement's silence, and retracting on it could close real defects
on model variance. **Two things are blocked on the owner:** the kubeconfig token has
expired, which stopped the council submission from dispatching at all, and the bug's own
closure criterion is still unsatisfiable (three costed options, unchanged from 08-12).

---

> ## ✅ UPDATE 2026-08-13 (evening) — TOKEN RESTORED; THE TWO BLOCKED ITEMS ARE DONE
>
> 1. **Council submission DISPATCHED**, correlation **`0c8e7f5b-e510-4d24-893d-e3abb0bbb7b6`**
>    — and confirmed live in `orchestration_states`, not merely printed. The section below
>    on the failed first attempt stays as the record of why `96c53bc18` carries no trailer;
>    forward-only forbids an amend, so that commit will list as un-reviewed in the `098`
>    report and this correlation is the join a human needs.
> 2. **The stability question is ANSWERED, and item 2 is unblocked:** the audit re-reported
>    the colour defect on **7 of 7** post-closure re-visits across 4 sites, on findings
>    independently known not to have been repaired. **But 0 misses in 7 bounds the miss rate
>    at ~35%, not at zero** — so item 2 must trigger on **N consecutive silences**, never on
>    one. Full working, and the two ways I got the query wrong first, in `NOTES`
>    (§"the stability measurement"). ⚠ **The `item_key` join reads 0-of-6 for
>    `hardcoded_section_colors` and that is an artefact of our own Half A rename** — match
>    on site+page, not on item_key.

## ⚠ BLOCKED ON THE OWNER — read before doing anything at the cluster

**The kubeconfig token has expired.** `kubectl -n ai-persona-system get pods` returns
`You must be logged in to the server (Unauthorized)`. Everything live is refused: DB
queries, pod logs, the council trigger, migrations. The owner refreshes it.

The one thing this cost that is not obvious: **the council submission for gate 1b did not
dispatch.** The trigger printed a full, convincing correlation block
(`SUBMISSION_CORR=4c2028f6-c3c6-4dbf-9113-5ebc8705c7b2`) and *then* its `kcat` publish
returned `Unauthorized`. The printout happens **before** the publish, so it names nothing.
The commit therefore carries **no trailer at all** — `Council-Submitted:` asserts a
submission was made, and this one was not.

**First action once the token is back**, in this order:

1. Re-fire the submission — payload is complete and validated at
   `scratchpad/213_d1_gate1b_submission.json`. Record the NEW correlation in `NOTES`.
2. Run the D1-half-two stability measurement below. It is one query and it decides a design.
3. After the next roll, run the RUNBOOK's gate-1b behavioural check (both directions).

---

## WHAT IS BUILT, AND THE ONE THING THAT NEARLY MADE IT WRONG

`platform/orchestration/actions/complete_work_item_no_change.go` — **completion gate 1b**,
sibling of `handlerReportedFailure`. If the handler's result payload reports zero changes at
every declared counter, completion is blocked and the item routes into the existing attempt
machinery with reason code `handler_reported_no_change`.

**Why it is NOT a verifier, which is the load-bearing design fact.** `VerifyTarget` carries
the **spec**, not the result. `load_work_item_actions.go:871` reads the handler's report as
an ACTION INPUT and marshals it into `site_work_items.result` at `:918` — *after* the gates
run. A verifier querying that column would have graded **the row's previous value** and
would have looked like it worked. Do not "improve" this by moving it into the verifier
registry.

Properties to preserve if you touch it:

- **Opt-in, unsafe default OFF.** An `item_type` absent from `noChangeGates` takes a map
  miss; nothing changes for it. This is not decoration — "the handler changed nothing" is a
  legitimate SUCCESS for other handlers (an idempotent repair finding its work already
  done), so a fleet-wide version would block real completions. The first test case asserts
  exactly this.
- **No verifier registered**, so `RegisteredVerifierItemTypes`, `verifier_coverage_test.go`
  and the `sql_for_agents/220` claim-timeout exclusion are all untouched — no lockstep to
  keep. That is a real advantage of gate 1b over a verifier and it should survive review.
- **The third arm ABSTAINS and records** (`agent_error_log`, `NO_CHANGE_GATE_UNREADABLE_RESULT`)
  when the declared counters cannot be resolved. Live today: 10 of the 14 completed rows
  carry a payload that is not this handler's (`bugs_open/213` §D, mechanism NOT ESTABLISHED).
- **Mutation matrix, all one-at-a-time:** roster entry removed → RED; any-non-zero early
  return removed → RED; `json.Number` arm broken → RED on that case; `int` arm broken → RED
  on that case; restored → GREEN.
- **NOT PROVEN by any test:** that `verifyBeforeComplete` actually calls the gate. Needs a
  `*sql.DB`, so it is behavioural-only. That check is the RUNBOOK's, and it needs its
  two-sided control — a gate that never fires and a gate that is not wired look identical.

---

## D1 HALF TWO — specified, not started, and here is the precondition

The plan of record (and the `bugfix_122` lane's own conclusion for `contrast_failure`) is
**retraction on the discovery path**: let the next audit's silence close the item.

The 08-12 handoff said that option "transfers here and is cheaper here than there". **The
first half stands; the second was too quick.** `WriteAuditFindingsAction` (`:509-545`) takes
its findings from an **LLM response**, parsed out of JSON and ```` ```json ```` fences. So:

> `contrast_failure` retracts on the silence of a **measurement** — a browser computed a
> ratio and the bad pairing is gone. `dark_section_audit` would retract on the silence of an
> **LLM**. A model that does not mention a defect on run N+1 has not established that the
> defect is gone; it may simply not have said so this time.

Known: the *wording* varies — the two `finetuning.uk` filings of the same defect on the same
component, one day apart, carry different `description` and `acceptance_test` text.
**[UNVERIFIED]** whether the *finding set* varies. On 08-11 the audit filed 14 items across
14 sites; on 08-12 exactly one was re-filed, although nothing repaired the other 13
(0 of 61 bodies changed). That is equally explained by the rotation visiting one site that
day. **Separating those two is task one, and it is cheap:**

```sql
-- Did a site the audit demonstrably RE-VISITED re-file its finding?
SELECT s.domain, w.batch_id, w.created_at::date, count(*)
FROM site_work_items w JOIN sites s ON s.id = w.site_id
WHERE w.item_type = 'dark_section_audit'
GROUP BY 1,2,3 ORDER BY 1,3;
-- then: for any site appearing on two dates, was the SAME item_key filed both times?
-- (finetuning.uk is the one known instance: design-audit_dark_section_audit_index_1368e337-…)
```

Widen it to the whole producer if the dark-section population is too thin — any
`spec->>'audit_source'='design-audit'` type will do, since the question is about the
auditor's stability, not this item type's.

**If the finding set is stable on an unchanged site**, build retraction: `resolveWorkItems`
(`work_items_common.go:249`) already exists in the producer's own package, and the dedup key
is page-level (`{audit_source}_{item_type}_{page_name}_{site_id}`,
`write_audit_findings_action.go:291`) so no per-section identity is needed. The remaining
blocker is the one 122 also has — the audit must report WHICH pages it examined;
`loadSitePages` is the site's inventory, not the audit's coverage.

**If it is unstable**, retraction is the wrong design here and a `pages_audited` list will
not save it. Say so and re-open the fork; do not build it anyway because the neighbouring
lane did.

---

## DECISIONS OWED BY THE OWNER

1. **Refresh the kubeconfig token.** Everything live is blocked behind it.
2. **The closure criterion, unchanged from 08-12 and still unanswered.** `bugs_open/213`
   says it stays OPEN until a `hardcoded_section_colors` item without `spec.check` reaches
   completion and lands `triaged`/`failed` — but Half A permanently moved that producer to
   `dark_section_audit`, so **the fix removed the traffic that would have demonstrated the
   fix**. (a) accept the unit + mutation proof and close, recording the one unexercised
   branch; (b) exercise it with one synthetic no-`spec.check` row on a throwaway site driven
   to completion (real dispatch, needs a yes); (c) leave OPEN, accepting the file no longer
   describes a reproducible defect.
3. **Whether gate 1b's blocked items should stay `failed`.** Once it rolls, up to 15
   `dark_section_audit` rows will land `failed` rather than `complete` — the honest state
   for a route that provably cannot repair them, but it is a population that did not exist
   before and someone reading that column will ask. The alternative is re-routing them to a
   handler that does not yet exist, which is D1 half three and nobody's task yet.

---

## STILL OPEN, NOT THIS LANE'S TO ADOPT (carried forward from 08-12, unchanged)

- **`RFC_024`** — nine CronJob meta-checks, no shared harness; three council seats have
  asked for a consolidation pass. Nobody has picked it up.
- **12 live item_types in neither half of `verifier_coverage_test.go`** (89 rows). Belongs
  to `bugs_open/021` §INSTANCE 2; contributed as a census comment, deliberately not adopted.
- **Two `design-audit` rows vanished from `site_work_items` between 08-11 and 08-12.** No
  standing pruner exists in code. Unexplained. It matters because `verifier-remit-check`'s
  retraction cannot tell a fix from a deletion.
- **The 10-of-14 payload split** (`bugs_open/213` §D) — recorded as an OBSERVATION with the
  mechanism NOT ESTABLISHED. Gate 1b now instruments it rather than guessing; the
  `agent_error_log` rows after the roll are the evidence that will settle it.

---

## COMMITS TODAY

`96c53bc18` — gate 1b (4 files, no council trailer, and the message says why).
Docs commit follows this file. Yesterday's evidence: `5c27a85a2`, `13d0bc588`.
