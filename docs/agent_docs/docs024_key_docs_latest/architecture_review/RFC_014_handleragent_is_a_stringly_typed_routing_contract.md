# RFC 014 — `HandlerAgent` is a stringly-typed routing contract with no spec-shape check, and the same defect has now shipped five times

**Filed 2026-08-05** by the `bugfix_201_page_content_writer_dispatch` lane.
**Raised by the council's `architecture` seat, on record, in an APPROVED round**
(correlation `71523705-07d1-4067-9c5d-af371ba84b89`, 15 reviewers, 5 advisory objections,
none high). The seat said **"Approve the edits as written … ship it"** and then said this
belongs on the record as the RFC-shaped question. This file is that record; it is **not** a
request to revisit `bugs_open/201`'s fix.

## The seat's words

> "this is now the FIFTH site … where a discovery check named `HandlerAgent:
> "page-content-writer"` for an item shape the writer's self-plan cannot consume. The only
> guard in the estate — `TestEveryCheckHandlerAgentExistsOrIsADeclaredGap` — checks that the
> string names a KNOWN agent, not that the agent can consume the spec shape being filed. That
> is a stringly-typed routing contract with no structural check tying HandlerAgent choice to
> call-shape compatibility, and it has now produced the identical defect class five times, each
> fixed by a one-off string edit rather than by narrowing what's sayable."

## What is actually true, measured

A discovery check declares a repair route as a bare string on `WorkItemSpec.HandlerAgent`. Two
things are checked, and one important thing is not:

| checked | by what |
|---|---|
| the string names an agent that exists | `handler_coverage_test.go` → `knownHandlerAgents` + `TestEveryCheckHandlerAgentExistsOrIsADeclaredGap` (source-scanning; **proven to fail on a bogus value** by mutation, 2026-08-05) |
| the dispatcher can carry what the handler declares | `scripts/audit-relay-gaps.sh` (`cmd/config-key-audit --relay-gaps`) — 175 agents decoded, 0 findings, 2 uncovered dispatcher-shaped relays |
| **the handler can CONSUME the spec shape the check files** | **nothing** |

The gap is the third row. `page-content-writer`'s self-plan reads
`input_data.current_page.sections`; a discovery spec has no `sections` key — measured, all 14
such work items in history carry only `{check, findings, fix, original_pipeline, page_id,
page_name, page_url}`. `plan_sections_action.go:867-875` then early-returns
`ready_count: 0, reason: "no sections to plan"` and the run hard-fails. 11 of 11 attempts,
2026-08-04.

**The five sites.** `check_empty_sections.go` and `save_sections_claims_guard.go` (both already
migrated, their headers still record "(was `page-content-writer`)"), plus
`check_literal_markdown.go`, `check_placeholder_contact.go`,
`check_component_standards.go:477` (migrated 2026-08-05 by `37afbb847`).

**Why CI could not catch any of them:** `page-content-writer` is a real, active agent, so it is
legitimately in `knownHandlerAgents`. Every one of the five passed the guard. The guard is not
weak — it is answering a different question.

## Why this is architecture-scope and not a bug

Under the owner ruling of 2026-07-29 §1, the test is whether a change alters what a **shared
mechanism guarantees**. It is not the routing string that is shared — it is the *implicit
contract* that naming an agent means that agent can act on your item. That contract has no
representation anywhere, so it cannot be violated loudly; it is violated silently and
discovered by a hard-fail in production months later. Each of the five fixes narrowed nothing:
the next check author can still write `page-content-writer` and pass CI.

Also relevant, and the reason "just patch the string" keeps looking sufficient: the failure is
**loud** in this instance. It need not be. Which brings in the second finding of the same round.

## The complication that makes the cheap fix less attractive than it looks

`bug_historian` [medium] and `guardian` [medium ×2] objected that the 201 fix trades a loud
hard-fail for a pipeline with filed history of silent partial success (016b §9: "Regenerated
content section is deferred by `plan_sections` … and dropped from the page"; "Page build
completes having built nothing — zero planned sections treated as success"). Gate 2 was
checked and is clear — `build-dispatch-loop`'s `call_handler` forwards `domain` and `site_id`,
both of `page-build-handler`'s required contract fields, item-type-agnostically, and
`spawn_handler` reads `current_item.handler_agent` from the row. But the deeper point stands:

**a spec-shape compatibility check would want to assert "this handler will ACT on this item",
and the estate cannot currently express that** — see `LANDMINES.md:4433`,
`page-build-handler`'s writer never sees a page's own stored prose unless `spec.mode="recreate"`,
"and there is today no workflow channel that passes a page's LIVE stored section content to its
own writer for editing." So "compatible" is not a single boolean; it has at least two axes
(can the handler *plan* the item; will the repair *preserve* what it should).

## Options, costed

1. **Remove `page-content-writer` from the legal direct-dispatch set for discovery-filed items.**
   Cheapest, and it closes the exact recurrence. A second, narrower map beside
   `knownHandlerAgents` — say `directlyDispatchableByDiscoveryChecks` — with the writer absent
   and a comment saying why. Catches site six at CI. Does **not** address the "will it act"
   axis. **Recommended as the floor**, because it is the only option whose cost is hours.
2. **Assert spec-shape compatibility per handler.** Each handler declares the spec keys it
   consumes; the coverage test cross-references a check's emitted spec against its named
   handler's declaration. Real, and real work: the emitted shape is built inline in Go
   (`specJSON`, per sub-check), so this needs the shape to become inspectable — a typed
   struct, or a declared key list per check. This is the version that would have caught all
   five.
3. **Do nothing and keep patching.** Cost, stated so it is a decision and not a drift: one
   council round, one commit and one deploy per undiscovered site, each surfacing as a live
   hard-fail on a customer site first. Five so far.

## What this RFC does NOT ask for

`bugs_open/201`'s fix is approved and committed. Nothing here is a reason to revert or hold
it. The question is whether option 1 or 2 gets funded, and that is an owner call — the
architecture seat and the `guardian` seat agreed on the diagnosis and neither asked for a
veto.

## Evidence

- Council report body, correlation `71523705-07d1-4067-9c5d-af371ba84b89`, `diagnosis_artifacts`
  `kind='council_report'` — `decided_by: "approved with 5 advisory objection(s) — none
  high-severity"`, `reviewers: 15`, `abstained: 2`, `gated_by_truncation: false`.
- `plan_sections_action.go:867-875` (the empty-input early return).
- `load_work_item_actions.go:242-243` and `:220-225` (every page type routes to
  `page-build-handler`; the item-type collision behind the third of the five).
- `handler_coverage_test.go:151` failure text, induced by mutation 2026-08-05.
- `LANDMINES.md:4433` (`spec.mode="recreate"`, and no channel for live section content).
- `bugs_open/201`, `bugs_closed/087` (the self-plan branch this is the untested half of),
  `bugs_open/178` (the same missing-prose mechanism, root cause confirmed 2026-08-03).
