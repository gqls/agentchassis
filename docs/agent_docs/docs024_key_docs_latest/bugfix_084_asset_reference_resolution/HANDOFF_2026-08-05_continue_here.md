# HANDOFF — bugs_open/084, continue here

**Written 2026-08-05 ~22:00 BST.** Everything below is committed. Nothing is
in flight. Pick this up cold.

## One paragraph

`bugs_open/084` says nothing in the platform ever asserts that a `<script src>`
on a deployed page actually resolves. Still true. I built the fix for its
candidate 2 — a discovery check called **`asset_reference_404`** — got it
**council-APPROVED at round 1**, and committed it. It is **deliberately not
enabled**, so the bug **stays OPEN**. What remains is a roll, an ordered
enablement, and one induced-fault proof in production.

## State

| thing | where | status |
|---|---|---|
| the check | `platform/orchestration/actions/discovery_checks/check_asset_reference_404.go` | committed `e526a5196` |
| tests (23) + mutation pins | `..._test.go` same dir | committed `e526a5196`, `9496f2cbd` |
| item-type classification | `verifier_coverage_test.go` | committed `e526a5196` |
| concept register | **IMP-051**, `register/improvement-loop.md` + index row | committed `e526a5196` |
| council verdict | corr `3675ec9a-4a8b-4c3f-b3b4-cc95498584c8` | **APPROVED r1**, 5 advisory, none high |
| the [medium] objection | liveness predicate was hand-rolled | **fixed** `9496f2cbd` |
| lane docs | this directory — PLAN / RUNBOOK / NOTES / README_where_we_are | committed |
| bug file | `bugs_open/084…md` | updated, **still OPEN** |
| `WRONG_CALLS.md` | 2 entries added | committed |

Both commits carry the correlation (`Council-Submitted:` then
`Council-Reviewed:`), so `098` credits them.

## The three things most worth knowing before you touch it

1. **It finds nothing today, and that is measured, not assumed.** All 541
   deployed pages fetched and DOM-parsed: 854 `<script src>` elements, 96
   distinct assets, **96 of 96 returning 200**. It is a regression guard. **Do
   not read a future "0 findings" as proof it works** — that is what a silently
   broken check reports too. The only evidence it bites is the mutation table in
   `RUNBOOK_asset_reference_resolution.md` and the induce-and-revert procedure.
2. **Do not "fix" `asset_loads`.** 084's candidate 1 is RFC-scope and was ruled
   so in a council round on 2026-07-29 (`experience_register/harvest/entries/
   CC-001_feed-driven-teaser-list.json:255`). It would also flip the type from
   `experienceStaticConfirming` to `experienceStaticRefuting`, the verbatim
   RFC_002 trigger. An RFC is owed; I did not write it.
3. **084's candidate 3 is already DONE** — `tool_eligibility.go` / TL-033,
   2026-07-29. The bug's §1 is stale and I have marked it so. `bugs_open/146
   (ported_tool_pages…)` is stale in the same way; re-read it before working it.

## STATUS UPDATE 2026-08-06 — LIVE AND ENABLED. Two steps of four are done.

Chassis **v1.0.1257** carries it; pod-grepped before the config change with a
negative control (`asset_reference_404` 13, UA 1, kind 1, `asset_reference_405`
**0**, `image_url_404` 7, both replicas). `design-discovery-agent` 22 → 23 checks,
`DO`/`RAISE` verify block, fixture updated in the same commit `42e117c5e`.

Also fixed while there, and named as not-mine: `literal_markdown`
(`bugs_open/184`'s lane) was live on `quality-discovery-agent` and missing from
`liveConfiguredChecks` — the fixture was under-asserting by one. It resolves, so
no production risk, but a silently drifting roster is what that file exists to
prevent.

## Next actions, in order

1. ~~Roll~~ **DONE** — v1.0.1257.
2. ~~Pod-grep with controls~~ **DONE** — see above.
3. ~~Enable + fixture in one commit~~ **DONE** — `42e117c5e`.
4. **MAKE IT RUN — this is the next thing, and it needs a decision.** The check
   has never executed. `improvement-sweep` is `enabled=f` (IMP-016), so the only
   route is `docs024_key_docs_latest/finetuning_uk_repair/294_TRIGGER_improvement_loop_v1.sh <site_id> [domain]`,
   which fires the FULL loop — `discovery → triage_findings → call_dispatch` — and
   **dispatches real content fixers at a real customer site**. That is an
   outward-facing action taken for the sake of a verification. Get the owner's
   go-ahead, or find a discovery-only path, before firing it. Its own pre-flight
   refusals (300s post-restart window; in-flight queue check) are built in;
   `FORCE=1` overrides them and should not be used casually.
5. **PROVE IT BITES.** A clean run proves nothing — the population is zero and a
   silently broken check reports exactly that. Induce one unresolvable reference
   on a page we own, assert exactly one item keyed on the resolved URL, revert,
   assert the retraction. Stated gap: retraction needs a still-referenced URL that
   now returns 200, so *deleting* the reference leaves the item open by design.
   Inducing it means writing to a live `page_components.rendered_html`, so plan
   the revert before the induction.
6. **Then narrow, do not close.** Candidates 4 (T5.1 post-hydration dead-control
   assertion) and 5 (generalise `checkScriptParity`, overlaps `bugs_open/178`/`198`)
   are untouched. Rename the bug to what actually remains rather than closing a
   file whose §3 and §4 are still live.

## Open question left for a human, not for the next session to argue away

The council's `bug_historian` seat, [medium], advisory:

> *"the platform keeps adding new detected-but-unhandled work item types (071,
> 079, 083, and now this one) faster than it drains them… a human should confirm
> [the owner ruling] still holds given the growing size of the undrained pile."*

Flag-only routing is defensible here (the repair is a judgement no generator can
make, and the 2026-08-02 §1 ruling covers it) — but the seat is right about the
trend. It is recorded in `bugs_open/084` and in IMP-051.

## Two missteps of mine that are cheap to repeat

Both in `WRONG_CALLS.md`, 2026-08-05, and both about **trusting a blank result**:

- I regexed `rendered_html` for `<script src>` and my only "live 404" was a
  **comment inside a tool's own JavaScript**. Parse the DOM; a regex cannot tell
  an element from a mention of one, and tool pages are the worst case.
- My mutation test reported **0 failures** and I nearly recorded the guard as
  inert. The mutation had never applied — a `str.replace` that matched nothing.
  **Assert the mutation applied before you interpret its result.**
