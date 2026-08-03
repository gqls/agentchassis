# PLAN — bugs_open/116, link-integrity check coverage

**Opened** 2026-08-03 by session "bugfix 100", working the standing instruction to
take the next unowned `bugs_open/` bug, verify it, and fix it at the framework
level rather than the individual case.

**Outcome: no code. The plan is to NOT build the fix, and the reason is the
deliverable.** This document records the decision and its reasons, per the
working-docs rule that corrections to the originating brief live here and are
marked as corrections.

---

## What the brief asked for

`bugs_open/116` says the three link-integrity discovery checks
(`phantom_internal_links`, `dead_controls`, `misdirected_cta`) have never run on
any site, and ranks four fix candidates, of which candidate 1 — *run the checks on
every build or change* — is called "the durable answer" and is backed by the
owner's own steer of 2026-07-27: *"whilst the improvement loop will return the
checkers should run after every build or change I think."*

Framework-level, per-build, unrepresentable-bad-state: on its face exactly the kind
of change the standing instruction prefers.

## Correction 1 — the premise expired before I got here

> **CORRECTED 2026-08-03:** the title claim ("has never run, on any site") is
> **false**. The checks ran on 2026-08-03 21:03–21:04Z against three sites and filed
> real findings. Evidence and the plural/singular naming trap that produced the
> original reading are in the bug file's STATUS block and in NOTES.

The surviving, defensible claim is narrower and was already on record from another
lane (`robot_hands_checker_gaps/NOTES_checker_gaps.md:95-96`, 2026-07-30): **no
enabled scheduled task targets any discovery agent, so cadence is whatever a human
supplies.**

## Correction 2 — the fix is gated on a decision nobody in a session can take

This is the load-bearing finding, and it inverts the bug file's own ordering.

The three checks file work items at status `detected`. The **only** thing that
promotes `detected` → `triaged` is `TriageDetectedItemsAction`, and it exists only
inside `improvement-loop`, which the owner **stopped deliberately** on 2026-07-29
(`bugs_open/136:32-35` — *"a decision, not a defect … do not re-enable them"*).

Fleet census 2026-08-03: **204 `detected` across 10 sites against 2 `triaged`.**

So a per-build detector would file findings into a queue with no consumer. That is
not a judgement call — it is refused in writing in three places:

1. **IMP-016** (`register/improvement-loop.md:130-136`): *"a discovery check should
   only be enabled once its handler agent actually exists — otherwise findings
   accumulate unconsumed."*
2. **`validate_page_content.go:644-650`**, which already faced this exact choice at
   the exact seam candidate 1 names, and declined: *"This writes a work RECORD, not
   a work ITEM, and that is a deliberate choice. A `site_work_items` row would
   promise a repair that nothing performs."*
3. **`bugs_open/077`**, cited there, against filing items whose handler has no remit.

**Therefore the correct sequence is the reverse of this bug file's:** the promotion
gap (`bugs_open/083`) is answered first; widening detection comes second. Building
candidate 1 now would produce a mechanism that reads as coverage and drains
nowhere — and it would be a *shared seam*, so by the 2026-07-28/29 platform-seams
rulings it would also owe an architecture round.

## The candidates, and why each is closed today

| # | candidate | verdict |
|---|---|---|
| 1 | per-build / per-write detection | **Forbidden by IMP-016 + the gate's own precedent** until `083` is answered |
| 2 | seat the checks on `design-discovery-agent` | **Warned against** by `bugs_open/149:395-398` — unattributable until 149-B2 is settled |
| 3 | recurring fleet-wide scheduled task | **G1** — an explicit separate owner go; migration 290 exists so it is one flag flip when it comes |
| 4 | re-enable the improvement loop | **Owner ruling 2026-07-29 forbids it** |

## What I did instead

1. Re-measured the bug against the live system, including the one measurement that
   can distinguish "clean" from "unexamined"
   (`sites.settings->'maintenance_profile'->'last_audit'`), and marked the residual
   uncertainty `[UNVERIFIED]` rather than publishing a count I cannot defend.
2. Filed the mechanism to the diagnosis loop rather than asserting it from here
   (run corr `54bf4506-5192-4528-8395-eb2c636a7fad`).
3. Corrected the bug file in place, visibly, with the evidence and the four
   citations — so the next thread does not spend a validity check discovering the
   title is stale. **One already did, earlier the same day.**
4. Recorded the misleading-scheduler-row trap in `LANDMINES.md`, and my own
   flattering-filter mistake in `WRONG_CALLS.md`.
5. Left the bug **OPEN**, with the owner question stated in the terms the owner is
   already being asked it in (the 204 parked findings).

## What would change this plan

An owner answer on the parked findings / G1. If the loop returns, candidate 1
becomes buildable and `SavePageSectionsAction` is the seam to use — it is the one
chokepoint every body-section writer passes through, it already carries two
pre-persist guards of exactly this shape, and `bugs_open/149` C1's rationale
applies verbatim: *"a gate can be forgotten by whoever writes the seventh agent; a
floor cannot."* The structural gap to plan around is that
`DiscoveryCheckContext` has no `PageID` (`discovery_checks/registry.go:41-50`), so
every check is site-scoped by construction, and all three load the site's `pages`
table separately.

That map is recorded in RUNBOOK so the next thread does not re-derive it.
