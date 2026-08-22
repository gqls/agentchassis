# 353 — a new tool's cross-links are withheld at birth and nothing ever emits them again: 30 live tools across ~24 domains have silently lost theirs since 2026-08-03

Filed 2026-08-22 ~09:3xZ by the staged-component-build lane, exposed by `bugs_open/330`'s
owner-approved positive control (NOTES `## 2026-08-22`; run corr `8ea2140b`, robot-hands.com).

**One sentence:** two individually-sound guards compose into a permanent withhold — the 029
crosslink emitter gates new-page crosslinks on a `needs_content_page` item that the 177 fix no
longer raises for pure-tool pages, and the only re-emission path (`tool-deployer`) has never run
— so every genuinely NEW tool born since 08-03 whose page was not already live lost its
related-page cross-links, forever, at `warning` severity.

## 1. Symptom

A tool build completes cleanly (`complete`, component + page + guide all created), its spec
carries a valid `related_pages` list — and **zero** `tool_crosslink:*` work items exist, with no
error anywhere. The only trace is one `agent_error_log` row,
`tool_crosslink_not_emitted:tool_page_will_not_go_live`, severity `warning`, and
`"cross_links_added": 0` inside the step output nobody reads.

## 2. Mechanism — every link read from the system's own records, none inferred

1. **Birth path** (`create_tool_component_action.go`, new-tool arm): the page row is INSERTed
   with `build_status='planned'`; `raiseToolContentItem` is asked for a content item and — since
   fix **177** (`74655b709`, 2026-08-03 11:06Z) — **declines when the page declares no prose
   sections**, which a pure-tool page never does. Worked case's step output:
   `"content_item": "skipped_no_prose_sections"`.
2. **The emitter's Guard 2** (`create_tool_cross_link_items.go`, built for
   `bugs_closed/029_HANDOFF_2026-07-19_tool_suggester_writes_phantom_tool_links.md` — **029 is a
   documented AMBIGUOUS number shared with the unrelated hung-spawns case; resolve by slug**):
   page not live (`toolPageLive` = `deployed|needs_rebuild`; `planned` is neither) → look for an
   open `needs_content_page` item on the tool page to `depends_on` → **none exists (see 1)** →
   withhold ALL crosslinks, write one skip row, return 0. Its own message: "cross-links withheld
   rather than pointed at a page that may never deploy". Correct in isolation; starved by 1.
3. **Nothing re-emits.** `emitToolCrossLinkItems` has exactly two production callers: the birth
   path (once, above) and `deploy_tool_action.go` — and **`tool-deployer` has 0 orchestrations
   in all retained history** `[MEASURED 2026-08-22: SELECT count(*) FROM orchestration_states
   WHERE owner_agent_type='tool-deployer' → 0]`. ⚠ **That is a RETENTION-BOUNDED statement, not
   an all-time one** — `orchestration_states` retention is per-status (a council seat has
   objected to "all-time" claims from this table before, correctly). It does not need to be
   all-time: for the 32 withheld tools, what matters is that no deployer run followed any of
   their births, and the skip census (08-03 → today) sits inside the same window as the zero.
   When the page later deploys via the ordinary page-build path, no crosslink emission happens
   — emission is a birth-time one-shot that already fired into the withhold.

## 3. Damage, measured (2026-08-22 ~09:2xZ)

```sql
SELECT context->>'skip_reason', count(*), min(occurred_at), max(occurred_at)
  FROM agent_error_log WHERE error_code LIKE 'tool_crosslink_not_emitted%'
 GROUP BY 1;
-- tool_page_will_not_go_live | 32 | 2026-08-03 17:53:48 | 2026-08-22 09:05:20
```

**32 withholding events, 32 distinct tools, ~24 domains — the first 6h47m after 177's guard
shipped.** Every one had a real, non-empty `related_pages` list (this skip fires only after the
`no_related_pages` check has passed). Joined to `pages` today: **30 of the 32 pages are NOW
`deployed` and carry zero `tool_crosslink` items ever** — the census query is in the lane NOTES
(`## 2026-08-22`) and regenerates the full table. The two exceptions
(`tool-affordability-complaint-checker`/lendzy, `tool-automation-savings-estimator`/finetuning,
3 items each) got their items from a LATER rebirth that found the page already live — which is
also why the class stayed invisible: **repeat births of established tools pass Guard 2 via an
already-deployed page**, and repeat births are most of the traffic (e.g. the rebuild lanes —
whose `replace_existing` arm, separately, never reaches the emitter at all).

## 4. Why it was invisible for 19 days

The damage is an **absence**: no failed item, no error above `warning`, the orchestration
completes honestly. The immune system sweeps recorded failures; nothing here records one. And
`bugs_open/330`'s conflict instrument sat one step upstream (wrong VALUES delivered), so its
silencing by migration 516 read as the whole story — the delivery being correct now exposes that
the next stage drops the correct delivery.

## 5. Filed without an 090 run — substitution stated per the 2026-07-31 ruling

No link in §2 is a hypothesis: the guard's own telemetry names itself and its reason
(`skip_reason`, `related_pages_n:3` in the worked case); the composition is read directly in the
two functions; the dead re-emission path and the 30/32 census are single queries reproduced
above. A 090 round would re-read the same three artefacts. Run one if a fix candidate below is
contested — the repro in §7 gives it a live case.

## 6. Fix candidates, ordered by what closes the door

1. **Emit on liveness, not at birth** — run emission (idempotent: the `tool_crosslink:` item_key
   is the dedup unit) at the moment `build_status` flips to `deployed`, or from the deploy-commit
   writer. Makes the withheld state unrepresentable: crosslinks exist iff the page is served.
   Needs the emitter callable with the spec's related_pages at that point (they are in the
   component/spec records — verify which survives to that path).
2. **Teach Guard 2 the channel that actually builds tool pages now** — `needs_content_page` is
   no longer how a pure-tool page goes live (177); let the gate accept whatever item/queue does
   (or accept `planned` + `needs_rerender` as "will go live"), keeping the depends_on semantics.
   Weaker than 1: it re-couples the guard to a second subsystem's policy, which is exactly the
   coupling that broke.
3. **Revert 177 for tool pages** — reopens 177's stall class (unsatisfiable items). Do not.

**Whichever wins, the 30 lost tools need a one-shot backfill** — their births are past and no
mechanism will revisit them. The census query names them all; emission through the central
helper keeps the item shape canonical.

## 7. Ready-made repro / fix verification

`tool-electric-vs-pneumatic-cost-comparator` on robot-hands.com (`00ff3af5-…`): page
`planned`, spec's three related pages (`electric-vs-pneumatic-economics`,
`robot-demand-step-change`, `pneumatic-vs-electric-grippers`) recorded verbatim in the
09:05:20Z skip row. After a fix: exactly 3 items keyed
`tool_crosslink:tool-electric-vs-pneumatic-cost-comparator:<page>:00ff3af5…`. Backfill
verification: the §3 census returns `items_ever > 0` for all 30. **Creation ≠ completion:**
two of the three targets are `rebuild_policy='owned'` pages, so completions may legitimately
gate on human review — do not read "3 created, 1 completed" as a partial failure of this fix.

## 8. Ownership

The mechanism is the phantom-tool-links 029's emitter + 177's guard, both CLOSED bugs — this is
a new defect in their composition, not a reopening of either. **029 is an ambiguous number**
(two unrelated closed cases share it; `who-owns.py 029` warns): the emitter belongs to
`bugs_closed/029_…_tool_suggester_writes_phantom_tool_links.md` (closed 07-26), and **the
countable-skip rows that made this bug findable at all are THAT lane's council round's doing**
(`025f4f34e`, "central insert, countable skips") — credit there, not to the hung-spawns lane.
The OTHER 029 (`…_hung_spawns_saturate_dispatch_group…`) closed **2026-08-20 18:10**
(`75b77f751`) with its live half re-filed as `bugs_open/343`; that session was notified
2026-08-22 (misrouted by the bare number — it confirmed no conflict and supplied these
corrections, including this close date, which an earlier version of this paragraph had merged
with the notification date). Adjacent-but-unaffected:
the tool-rebuild lanes (`replace_existing`, 331/TL-047) whose arm exits before the emitter.
Filing lane: staged-component-build (this find falls out of 330's verification and blocks
nothing in it — 516's resolver half is proven both directions regardless; see 330 §10).
**The FIX is unowned as of filing** — it belongs to whoever claims this file (announce the
claim here, and run `who-owns.py 353` first), not to whichever session reads it next, and not
automatically to the filing lane.
