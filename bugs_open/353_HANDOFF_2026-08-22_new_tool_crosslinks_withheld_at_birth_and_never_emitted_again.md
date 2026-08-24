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

## 9. CLAIMED, FIXED AND BACKFILLED — 2026-08-23 (staged-component-build lane, owner-approved)

**Claimed** after `who-owns.py 353` showed only this lane's own filing commits.

### 9.1 The forward fix — commit `323b63a00`, council `642ecc3c` (submitted)

Widening the gate's item types alone does **not** fix the birth path, and finding out why is the
substance of this fix: **the ordering**. tool-generator runs `save_tool` (create_tool_component,
which calls the emitter) and only THEN `enqueue_rerender` (create_rerender_items), which files the
`page_rerender` item. On the worked case the withhold is stamped **09:05:20Z** and the gate item
**09:06:11Z** — Guard 2 ran **51 seconds before** the item it was hunting for existed.

So the fix is two parts:
1. **Widen the gate query** to the channels that actually build a page today —
   `item_type IN ('needs_content_page','page_rerender','needs_page')`.
2. **An OPT-IN field, unsafe default OFF** (the owner's 2026-08-02 shared-seam ruling):
   `pageBuildIsEnqueuedByThisWorkflow`, set by **only** `create_tool_component`, which creates the
   page and whose own workflow enqueues its build. `deploy_tool_action` is deliberately left on the
   default — it promises no build. A zero-valued request keeps today's behaviour, pinned by test.
3. **The decision is EXTRACTED** into `crossLinkEmitDecision(pageLive, gateItemFound,
   buildEnqueuedByCaller)` — because the branch that caused this bug sat inside a DB-dependent
   function where no unit test could reach it, which is how it survived 19 days while tests of its
   *inputs* passed. **Mutation-proved:** making the opt-in inert fails exactly the named
   `THE_353_FIX` case; `./platform/...` green from a git-archive-HEAD copy.
4. The permissive arm writes its own countable INFO row
   (`emitted_ungated_build_enqueued_by_caller`), so the new branch is as measurable as the withhold
   was. **Go change ⇒ inert until a roll.**

### 9.2 The backfill — DONE, 2026-08-23 ~17:1xZ

**`cmd/backfill-tool-crosslinks`** (dry-run default, `--only <fn>` canary, `--apply`). It calls the
**real emitter** rather than inserting rows, so the item shape, the `item_key` namespace, the dedup
clause and the two-strike anti-churn cannot drift into a second implementation.

**Its input is the guard's own telemetry** — `related_pages` is recorded verbatim in every skip row,
so nothing is reconstructed or guessed. *The design decision that made this bug findable is what
made it repairable*, which is the strongest possible argument for countable skips.

| figure (as of 2026-08-23 17:1xZ) | value |
|---|---|
| withheld tools found | **37** across **24** domains |
| eligible (page live, no items yet) | **34** (3 already had items from a later rebirth) |
| **cross-link items created** | **74** (1 canary + 73) |
| fleet total now | **151** items, **65** tools, **19** sites |
| withheld tools still at zero | **5 — ALL CORRECT, see below** |

**Canary first** (`tool-gripper-torque-moment-calculator`, robot-hands): verified at the artefact
before the rest — right `item_key` namespace, `content_rewrite`/`triaged`/`page-build-handler`,
priority 110, `source='backfill-353'`, the REAL tool URL, `mode=edit_live`, acceptance_test present.

**⚠ THE FIVE ZEROS ARE CORRECT, AND WERE CHECKED RATHER THAN ACCEPTED.** `tool-bridging-compound`,
`tool-rate-scenarios`, `tool-combat-balance-comparator`, `tool-stat-budget-allocator`,
`tool-wave-difficulty-ramp` created nothing because **every page they name is itself a tool page**
(12 of 12 named pages `LIKE 'tool-%'`, all of which exist) — and the emitter deliberately skips
tool-to-tool cross-linking. Several other tools created *fewer* items than pages named for the same
reason. A zero here is the filter working, not a failure.

### 9.3 What is NOT yet true — the bar for closing this bug

**A created item is a REQUEST, not a link.** These are `content_rewrite` items for
page-build-handler; the pages carry the references only once they dispatch and rerender. **Do not
close 353 on the 74.** The close condition is the artefact: named pages actually serving an inline
link to their tool. Also note two of the robot-hands targets are `rebuild_policy='owned'`, so their
completions may legitimately sit in human review — "74 created, N completed" is the owned-page
control working, not a partial failure.

**Still open:** (a) the forward fix is inert until a chassis roll; (b) the 74 items must dispatch
and be verified at the artefact; (c) `tool-deployer` still has 0 runs in retained history — the
second emitter caller remains an unexercised path, unchanged by this work.

## 10. 2026-08-24 — the rewrites LANDED but are NOT SERVED: 51 pages hold the link in stored HTML and have not redeployed

**61 of the 74 backfilled items are `complete`** (8 `wont_fix`, 3 `needs_human_review`, 2 `failed`
— those 13 are the ordinary handler outcomes, not a backfill defect). So the writers did the work.

**And the pages do not serve it.** Checked at the artefact per this file's own §9.3 bar, with a
control:

```
curl https://dartsonline.com/barrel-shapes.html | grep -c /tools/tungsten-diameter-visualiser/  -> 0
curl https://dartsonline.com/barrel-weight.html | grep -c /tools/tungsten-diameter-visualiser/  -> 0
control: /about.html (no backfill item)                                                         -> 0
```

**The writer is NOT at fault — the link IS in `page_components.rendered_html`** (1 component on
barrel-shapes, 4 on barrel-weight). The page simply has not redeployed since:

| page | deployed_at | rewrite completed | order |
|---|---|---|---|
| barrel-shapes | 2026-08-23 18:32:57Z | 18:33:04Z | deploy **7 s BEFORE** the rewrite |
| barrel-weight | 2026-08-23 18:29:13Z | 18:29:20Z | deploy **7 s BEFORE** the rewrite |

**Fleet-wide it is systematic, not a race on two pages: of 51 pages carrying a completed backfill
rewrite, 51 deployed BEFORE their rewrite landed and 0 after** (2 are already `needs_rebuild`).
A ~7-second gap repeated 51 times is an ORDERING property of the rewrite path, not a coincidence.

**So the backfill is DONE AT THE DATA LAYER AND UNFINISHED AT THE SERVING LAYER.** The remaining
step is a redeploy of those 51 pages. That is a real fleet action across ~19 live sites and is
**NOT being fired unilaterally** — it is the same class of decision as the backfill itself and
wants the owner's word, exactly as the backfill did. Note the standing caveat when it is run: a
re-render carries **every** improvement made since each page last rendered, so it must not be
sized by this change alone.

**Whether they redeploy on their own is UNPROVEN in both directions** — none has in ~17 h, and the
window is too short to call it. Do not record "they will pick it up naturally" without measuring it.

**353 therefore stays OPEN** on: (a) the forward fix, inert until a roll (council round 2 in
flight, corr `642ecc3c` — round 1's two objections were both right and are answered in the code);
(b) these 51 pages serving their links; (c) `tool-deployer` still unexercised.
