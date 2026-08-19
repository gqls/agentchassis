# 317 — the claimed-item-timeout sweep can complete an item with NEITHER completion gate running, because its exclusion list is locked to gate 2 only

**Filed 2026-08-19** by the `bugfix_302_design_repair_verification` lane, spun out of
`bugs_closed/302` so it is not lost when that file closes. **OPEN, UNOWNED. Latent — measured
zero occurrences, filed for the door rather than the damage.**

## The mechanism, verified in the live predicate rather than inferred

`scheduled_tasks.claimed-item-timeout` (enabled, hourly) auto-completes an item that has been
`claimed` past its timeout, on **generic orchestration evidence**: the handler orchestration this
claim dispatched reached `COMPLETED` after the claim was taken. That write does **not** go through
`CompleteWorkItemAction`, so **neither completion gate runs** — not gate 1b (`noChangeGates`,
`complete_work_item_no_change.go`) and not gate 2 (the verifier registry).

Its protection against that is an item_type exclusion list, and its own comment states the contract:

> *"The item_type exclusion is the LOCKSTEP TWIN of the `RegisterVerifier()` calls in
> `platform/orchestration/actions/discovery_checks/`: those item types have a Go verifier that…"*

**So the list keys on gate 2 only.** A type that has a **gate-1b roster entry and no registered
verifier** is not excluded — and that is exactly `dark_section_audit`, the only entry on that
roster. For such a type the sweep is a completion path that no gate can see.

## Why this is filed as latent, and the measurement that says so

[MEASURED 2026-08-19, archive-inclusive — `site_work_items UNION site_work_items_archive`, because
the live table is only a ~7-day window]

| item_type | completions carrying the sweep's `completed_by_step` shape | total completions |
|---|---|---|
| `dark_section_audit` | **0** | 26 |
| `hardcoded_section_colors` | **0** | 498 |

**It has never fired for either type.** The reason is contingent, not structural: both carriers that
dispatch `dark_section_audit` are `enabled=false`, so nothing claims these items, so nothing times
out. **Re-enable either carrier and this becomes reachable** — and the false green it produces is
precisely the class `bugs_open/213` D1 and `bugs_closed/302` were filed about.

## ⚠ Do NOT "just add the type" to the exclusion list

`TestRegisteredVerifiersMatchClaimTimeoutExclusion` enforces the lockstep **in both directions** —
excluded ⇔ has a registered verifier. Adding `dark_section_audit` to
`sql_for_agents/220_claimed_item_timeout_generic_evidence.sql` without registering a verifier for it
**fails that test**, and registering a verifier for it is the thing `bugs_closed/302` established
should *not* be done by that route (the family is classified in `verifier_coverage_test.go` as
needing a browser on the completion path, or as `catJudgement` with nothing to re-run).

## Fix candidates, ordered by what makes the bad state unrepresentable

1. **Widen the lockstep contract from "has a verifier" to "has a verifier OR a gate-1b roster
   entry", and change the parity test to match.** Closes the class: any future type on either
   roster is excluded automatically, with nothing to remember. Touches a **live
   `scheduled_tasks.pre_query`** that no Go test can see, plus the test that guards it — so it needs
   the same lockstep discipline the original had (migration + test in one change), and the live
   column must be read before it is written.
2. **Make the sweep route its completion through `CompleteWorkItemAction` instead of writing the
   row itself.** Strictly better in principle — one completion path, all gates, no list to keep in
   step — and correspondingly larger: the sweep's whole design is that it works "for every item
   type, because every dispatched item carries its id into the handler", and it deliberately writes
   parity with the lost `mark_complete`, not a stricter test (migration 220's header).
3. **Do nothing while both carriers stay disabled.** Defensible today and only today; it makes
   re-enabling a carrier a silent re-arming of this path, which is the shape this estate keeps
   filing bugs about.

## How to verify a fix

The predicate is the artefact: read the LIVE `scheduled_tasks.pre_query` (the repo file is history,
the live row is fact) and assert the exclusion list contains every `noChangeGates` key as well as
every `RegisterVerifier` type. Then induce: a `claimed` `dark_section_audit` item past its timeout
must NOT be auto-completed. ⚠ **A zero here needs a demand control** — with both carriers disabled
nothing will ever be claimed, so "it did not fire" is not evidence in either direction.

## Relations

`bugs_closed/302` (the lane that found this; its gate-1b arm is fixed, live and proven —
`handler_result_unreadable`) · `bugs_closed/213` D1 (gate 1b itself, WII-017) · WII-011 / `RFC_017`
(gate 2's fail-closed policy) · `sql_for_agents/220` + `TestRegisteredVerifiersMatchClaimTimeoutExclusion`
(the lockstep pair) · `bugs_open/230` (the rotation work that would re-enable a carrier and make this
reachable) · LANDMINES `Dedup index ↔ Go list lockstep` (the same "two hand-maintained copies of one
vocabulary" class).
