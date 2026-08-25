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

> **⚠ CORRECTED 2026-08-24 — "`tool-deployer` has never run" is FALSE, and the query cited could
> never have shown otherwise.** `orchestration_states` retention is **24 hours, sliding** — measured
> today, oldest COMPLETED row **24.5 h**, oldest FAILED **23.5 h**, and no non-reaped status to lift
> the floor. So `owner_agent_type='tool-deployer' → 0` meant *"none in the last day"*, never
> *"none ever"*, and the "retention reaches 2026-07-19" figure quoted in the round-3 submission was
> the documented FALSE FLOOR (non-reaped `CANCELLED`/`RUNNING` rows), not a real window.
> **Re-sourced from a table with no retention job:** `agent_error_log` holds **10** `tool-deployer`
> rows spanning **2026-08-03 17:55:49Z → 2026-08-15 00:08:28Z** — so it ran **at least 10 times**
> during the damage window, every one logging *"workflow completed but its result could not be
> delivered to the parent (failed_transient)"*. **The workflow COMPLETED**; only the reply to its
> parent failed.
> **What this does and does not change.** The DAMAGE is untouched — it was measured from durable
> rows (32 withheld tools, 30 with zero cross-links ever), not from this table. What changes is the
> MECHANISM sentence: the correct statement is not "the re-emission path has never run" but
> **"it ran, and the withheld tools still have no cross-links"** — which is a sharper and so far
> UNANSWERED question (did it run for other tools, or run for these and withhold again?).
> **[UNMEASURED]** — and it is now open item (c'), restated in §13.2.
> Caught by the round-3 council's `prior_art_librarian` seat citing this estate's own retention
> landmine. Do not restate the old claim.


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

> **⚠ CORRECTED 2026-08-24 — the attribution in the sentence above is WRONG, and the true source
> refutes this file's central mechanism claim.** `tool-automation-savings-estimator`'s 3 items did
> NOT come from "a LATER rebirth". They were emitted by **`tool-deployer`** on **2026-08-11**
> (`site_work_items` where `item_key LIKE 'tool_crosslink:%'`, `source='tool-deployer'`).
> **So the re-emission path this file calls dead did run, did reach the emitter, and did repair one
> of the withheld tools.** The exception §3 records as a curiosity was the proof all along.
> Measured in §14. (The other exception, `tool-affordability-complaint-checker`, has no
> `tool_crosslink` rows in `site_work_items` today — **not evidence of absence**: that table is a
> rolling window and completed items archive out of it, so its original source is now unrecoverable
> from this table. Do not read the missing row as a contradiction of §3.)


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

## 11. ⚠ **§10 IS WRONG AND IS RETRACTED — the links WERE being served all along.** The backfill is COMPLETE at the artefact (2026-08-24 ~11:0xZ)

> **CORRECTED 2026-08-24:** §10 claimed "51 of 51 pages hold the link in stored HTML and have not
> redeployed", and recommended a 51-page fleet redeploy. **That finding was an artefact of MY OWN
> MEASUREMENT and the recommendation was withdrawn before it ran.**

**What I did wrong: I CONSTRUCTED the page URLs instead of reading `pages.url`.** I curled
`https://dartsonline.com/barrel-shapes.html`; the page's real URL is **`/blog/barrel-shapes.html`**.
Every "0 hits" in §10 — including its *control* — was a 404-shaped miss on a URL that does not
exist. **This is precisely the defect `bugs_closed/029` exists for** ("a tool page's URL CANNOT be
constructed from its name — it has to be LOOKED UP from `pages.url`"), committed while working
inside that very file, on the bug that file's guard caused.

**The re-measurement, using `pages.url` read from the database, with controls:**

| page (real URL) | tool | hits |
|---|---|---|
| dartsonline.com `/blog/barrel-shapes.html` (rerendered) | tungsten-diameter-visualiser | **1** |
| dartsonline.com `/blog/barrel-weight.html` (**never rerendered by me**) | tungsten-diameter-visualiser | **2** |
| negative control (same page, a fake tool slug) | — | **0** |

The second row is the decisive one: **a page I never touched was already serving its link**, so
nothing was waiting on a redeploy. A random sample of **12 backfilled pages across 8 domains, all
at their DB-read URLs: 12 of 12 serving, 1–4 hits each.**

**The `deployed_at < completed_at` comparison that produced §10 was a red herring**, and the very
consistency I cited as proof ("51 of 51 — an ordering property, not a race") was the tell I
misread: `deployed_at` is simply not stamped by the rewrite's own deploy path. **A 100% result
should have prompted "what would make this true trivially?" rather than a mechanism story.**

**So §9.3's bar is MET and the damage half of 353 is CLOSED:** 74 items created, 61 complete, and
the links are live on the pages. What remains open is only **(a)** the forward fix, inert until a
roll (council round 2, corr `642ecc3c`), and **(b)** `tool-deployer`'s unexercised path.

**One redeploy was fired before the correction** (dartsonline `barrel-shapes`, corr `c0fd334d`,
COMPLETED) — harmless, and it is what exposed the error: it "fixed" nothing because nothing was
broken, and re-checking the control is what showed the control had been serving all along.
**The 50-page redeploy was NOT run.**

## 12. 2026-08-24 — **the forward fix is LIVE** (open item (a) discharged), round 2 came back REVISE on a STALE SKETCH OF MINE, and the call-site gap it named is now a test

### 12.1 The fix SHIPPED — probed at the artefact, with both controls

Chassis pods `agent-chassis-8bbb57765-{6q6vp,j5gdd}`, both `v1.0.1332`, started 09:39:19Z /
09:39:50Z, one replicaset. The startup `build provenance` line had already scrolled out of
`--tail=3000`, so this is the **capability probe**, not a commit inference:

| literal probed | expected | result |
|---|---|---|
| `emitted_ungated_build_enqueued_by_caller` (the new arm) | present if shipped | **PRESENT** |
| `tool_page_will_not_go_live` (**control +**, old literal) | must be present | **PRESENT** |
| `zzz_synthetic_literal_that_cannot_exist` (**control −**) | must be absent | **ABSENT** |

**So §11's open item (a) — "the forward fix, inert until a roll" — is DISCHARGED.**

### 12.2 ⚠ But LIVE is not PROVEN, and the zero here proves nothing — the demand control says why

`agent_error_log` since the roll: **0** `tool_page_will_not_go_live`, **0**
`emitted_ungated_build_enqueued_by_caller`, and **3** rows all `no_related_pages`
(10:19:50Z–11:12:00Z). The withhold count was **5 on 08-23** and is **0 today**.

**Do not read that as the fix working.** Demand control: `orchestration_states` with
`owner_agent_type='tool-generator'` since the roll = **3 runs, 3 COMPLETED** — so there WAS birth
demand, but all three recorded `no_related_pages`, which short-circuits **before** Guard 2 is ever
reached. Nothing since the roll could have exercised the new arm whatever it does. **The correct
statement is live-and-unexercised.** The INFO row exists precisely so this stays measurable — the
first non-zero is the proof, and it has not happened yet.

*(Adjacent, unfiled: 3 of 3 births carrying no `related_pages` at all may be the same
generation-side degradation as the `tool_birth_instance_scope_refused` rate in the lane handoff §3.4.
Two data points is not a trend — noted, not claimed.)*

### 12.3 Council round 2: **REVISE**, `decided_by` = gating objection from editquality

Run `9a6e4350`, 2026-08-24 10:28:53Z → 10:35:17Z; 11 reviewers, 6 abstained, not truncated.

**The high-severity objection was RIGHT ABOUT MY TEXT AND WRONG ABOUT THE CODE — and that is my
defect, not a reviewer error.** It read edit 3's sketch, which still said
`crossLinkEmitDecision(false, err == nil, …)`, and objected that if that ships the `pageLive`
branch is dead in production. I had fixed the code in round 2 and **left the sketch describing the
old code**. The reviewer set exactly the right clearing condition ("…or a code_check confirms the
committed code differs from the sketch") and named the check. Run:

```
create_tool_cross_link_items.go:245  pageLive := toolPageLive(buildStatus)
create_tool_cross_link_items.go:264  switch crossLinkEmitDecision(pageLive, gateItemFound, req.pageBuildIsEnqueuedByThisWorkflow) {
```

The code was right; the submission was not. **From outside, a sketch that contradicts the diff is
indistinguishable from an unfixed defect.**

The low-severity objection (edit 6 is a comment-only change on the `replace_existing` return) is
**agreed and not contested**: it is disclosure of a residual, not a fix, and is credited as nothing
more. The residual it names — **a regeneration whose spec ADDS a related page never emits a
cross-link for it** — is recorded here rather than left in a comment alone, and is open and unowned.

### 12.4 The `missing` item was the substantive one, and it is answered with a test

> "No test or wiring change proves the production caller of `crossLinkEmitDecision` passes the real,
> computed `pageLive` rather than a literal — `TestCrossLinkEmitDecision` only pins the pure
> function's table, not the call site's argument."

Correct, and it is **this bug's own failure mode recurring**: 353 survived 19 days in a
DB-dependent branch no unit test could reach while its inputs' tests stayed green. Answering that in
prose would have repeated the mistake.

**`TestCrossLinkCallSitePassesTheRealPageLive`** (commit `027461e3d`) drives
`emitToolCrossLinkItems` through sqlmock with the one setup that discriminates — **tool page
SERVED, opt-in OFF**. Correct wiring reads `deployed` → `pageLive` TRUE → emit. The literal-`false`
wiring falls to withhold and creates nothing. The assertion is the **EFFECT** (`created == 1`),
never the absence of a query — that shape passes vacuously the moment the call fails for any other
reason.

**MUTATION-PROVED** in a `git archive HEAD` copy (the shared tree is never left mutated): restoring
`crossLinkEmitDecision(false, …)` fails **exactly** this test while **both pre-existing tests keep
PASSING** — the reviewer's point demonstrated rather than argued. Reverted; full
`./platform/orchestration/actions` green (3.815s).

### 12.5 Round 3 submitted — corr `642ecc3c` (same correlation, `RESUBMIT_CORR`, so the trail accumulates)

Seven edits: substance unchanged from round 2, edit 3's sketch replaced with the committed shape
verbatim, and the call-site test added as edit 7. Submission file:
`docs/agent_docs/docs024_key_docs_latest/staged_component_build/COUNCIL_SUBMISSION_2026-08-24_bug353_crosslink_withhold_round3.json`

### 12.6 What 353 is still open on

- **(b) The new arm is live but UNEXERCISED** (§12.2). First non-zero
  `emitted_ungated_build_enqueued_by_caller` row closes it; nothing else does.
- **(c) `tool-deployer` still has 0 runs in retained history** — the emitter's second caller remains
  an unexercised path, unchanged by any of this work.
- **(d) The regeneration residual** (§12.3): `replace_existing` returns before the emitter, so a
  regeneration that ADDS a related page never emits for it. Unowned.
- Round 3's verdict.

**Item (a) — the forward fix pending a roll — is CLOSED (§12.1). The damage half stays CLOSED
(§11).**

## 13. 2026-08-24 — council round 3 **APPROVED**, and its three advisory objections found two real defects in my own evidence

Run `53e3812f`, `complete_approved`, *"approved with 3 advisory objection(s) — none high-severity"*.
Corr `642ecc3c` throughout (rounds 1–3 on one correlation).

**Advisory objections are not decoration.** All three were about EVIDENCE rather than code, two were
right, and one of those refutes a claim that had stood in this file since filing. Recording them in
full because "APPROVED" is exactly the moment they stop being read.

### 13.1 `debug_historian` (medium) — my deploy probe was too narrow. **Right about the method; the conclusion survives on wider evidence.**

> "Deploy verification checked `/proc/1/exe` on only the two pods matching `-l app=agent-chassis` …
> a per-run agent runs in its OWN EPHEMERAL POD … If `tool-generator` spawns as a per-run pod, the
> binary probe proves nothing about the pod that matters."

**Correct, and §12.1 as written did not establish what it claimed.** I inferred completeness from
"one replicaset" on a **two**-pod label selector. The namespace holds **159** pods, of which **68**
run this binary as per-run `agent-*` pods — and `tool-generator` is one of those (no
`agent-tool-generator-*` pod exists at rest; it spawns per run).

**Re-checked properly, and it holds:** all **68** per-run agent pods are on `v1.0.1332`, and a
per-run pod **on a different node** from the two chassis pods probes clean —
`emitted_ungated_build_enqueued_by_caller` **present**, control **+** present, control **−** absent.
Node-level check matters here because a same-tag rebuild can serve a stale cached image per node.

**The lesson is the reviewer's, not mine:** `-l app=<service>` answers "which pods carry this label",
never "which pods run this binary", and on this estate those differ by a factor of 34.

### 13.2 `prior_art_librarian` (medium) — **THE REAL ONE. "`tool-deployer` has never run" is refuted.** See the correction banner in §2.

> "This is exactly the documented landmine pattern for this column … A 0-row result on this predicate
> does not establish the claimed absence."

Retention is **24 h sliding** (measured today: oldest COMPLETED **24.5 h**). Re-sourced from
`agent_error_log`, which has no reaper: **10 `tool-deployer` rows, 08-03 → 08-15**, every one
*"workflow completed but its result could not be delivered to the parent"*. **It ran. The workflow
completed.**

**Open item (c) is therefore RESTATED, and it is now a better question than the one it replaces:**

> ~~(c) `tool-deployer` still has 0 runs in retained history — an unexercised path.~~
> **(c′) `tool-deployer` RAN at least 10 times inside the damage window and the 30 withheld tools
> still have zero cross-links. Did it run for other tools, or for these and withhold again?**
> **[UNMEASURED]** — the discriminator is whether any of those 10 runs names a withheld tool.

The damage figures are untouched: they came from durable rows, never from `orchestration_states`.

⚠ **This is the landmine's own predicted failure mode, third recorded instance.** The entry says a
TABLE-footprinted landmine cannot reach anyone via the `SessionStart` hook (which matches dirty
FILE paths), so its delivery moment is the first query — and it names two other lanes it failed to
reach on 08-23. I read this file, then queried the table, and still restated the false floor.
Logged to `WRONG_CALLS.md` and appended to the landmine's own tally.

### 13.3 `bug_historian` (medium) — the regeneration residual must not ride this bug's closure. **Agreed; now filed separately.**

> "A comment pointing at `bugs_open/353` is not durable tracking for a DIFFERENT residual than the
> one 353 is about — recommend a distinct work item so it doesn't silently ride the parent bug's
> closure."

**Right, and it matches this estate's documented recurring shape** ("one call site of a shared
judgement gets the rigorous fix; the sibling stays heuristic"). §12.6(d) is exactly a different
defect parked inside this bug's open list, where it would close when 353 closes.

**Filed as `bugs_open/379`.** Removed from 353's open list — see §13.4.

### 13.4 353's open list, restated

- **(b)** the new arm is **live but UNEXERCISED** — first non-zero
  `emitted_ungated_build_enqueued_by_caller` row closes it; the zero today is uninformative (§12.2).
- **(c′)** `tool-deployer` ran ≥10× in the damage window and the withheld tools still have nothing —
  **[UNMEASURED]**, §13.2.
- ~~(d) the regeneration residual~~ → **moved to `bugs_open/379`**, no longer gated on this file.

**CLOSED:** the damage half (§11), and item (a), the forward fix (§12.1, method corrected in §13.1).
**The forward fix is APPROVED (round 3) and LIVE** — nothing further to ship.

## 14. 2026-08-24 (later) — **item (c′) is ANSWERED and CLOSES: `tool-deployer` is not broken and not unexercised. It ran, it emitted, and it repaired one of the withheld tools itself.**

Measured after the `v1.0.1334` roll. This retires the last mechanism question on this bug.

### 14.1 The question

§13.2 restated open item (c) as: *`tool-deployer` ran at least 10 times inside the damage window and
the 30 withheld tools still have zero cross-links — did it run for OTHER tools, or for THESE and
withhold again?* The discriminator named there was "whether any of those 10 runs names a withheld
tool".

### 14.2 The answer: **for other tools, and it worked — plus one of the withheld ones**

The decisive instrument was not the runs at all; it was the emitter's own output, which records its
caller. Two queries, both against durable tables:

```sql
-- who called the emitter, and did it ever withhold for them?
SELECT context->>'emitted_by', context->>'skip_reason', count(*) FROM agent_error_log
 WHERE error_code LIKE 'tool_crosslink%' GROUP BY 1,2;
-- tool-generator | tool_page_will_not_go_live | 37
-- tool-generator | no_related_pages           | 12
--   *** NO ROW FOR tool-deployer — it never withheld, not once ***

-- what did it actually produce?
SELECT source, count(*), min(created_at)::date, max(created_at)::date
  FROM site_work_items WHERE item_key LIKE 'tool_crosslink:%' GROUP BY 1;
-- backfill-353   | 74 | 08-23 | 08-23
-- tool-generator | 56 | 08-17 | 08-21
-- tool-deployer  | 12 | 08-03 | 08-19   <-- the "dead" path
-- tool-suggester |  5 | 04-08 | 07-24
```

**`tool-deployer` emitted 12 cross-link items for 5 tools between 08-03 and 08-19, and produced ZERO
withhold rows.** Every one of its calls into the emitter succeeded.

| tool the deployer emitted for | first item | in the withheld set? |
|---|---|---|
| `tool-model-approach-selector` | 2026-08-03 | no |
| `tool-ab-test-calculator` | 2026-08-05 | no |
| `tool-llm-cost-calculator` | 2026-08-05 | no |
| **`tool-automation-savings-estimator`** | **2026-08-11** | **YES** |
| `tool-archetype-clash-calculator` | 2026-08-19 | no |

**The fourth row is the finding.** It is one of the two "exceptions" §3 has recorded since filing as
having got its items *"from a LATER rebirth that found the page already live"*. That attribution was
wrong: the items came from `tool-deployer`. **The path this file declared dead had already repaired
one of this bug's own casualties, eight days before the bug was filed.**

### 14.3 What this means for the bug — and it is a simplification, not a new defect

`deploy_tool_action.go` calls the emitter on **both** its arms, and the early-return arm says so in
its own comment: *"Cross-links are emitted even on this early return: re-running the deployer is the
supported way to backfill them for a tool that was deployed before `bugs_open/029` was fixed."*

**So the re-emission path was never the problem.** It exists, it is exercised, it succeeds, and it is
documented as the supported repair route. What was missing was a **TRIGGER**: nothing re-runs the
deployer for a tool whose birth withheld, and a withheld tool emits no signal that would prompt one.

**Both ends are now closed by work already done** — the forward fix stops new withholds at birth
(§12.1/§13.1, live in `v1.0.1334`), and the backfill repaired the historical 30 (§9.2/§11). **There
is nothing to build here.**

### 14.4 What the 10 `agent_error_log` rows actually were — a known, CLOSED, fleet-wide class

Every one reads *"workflow completed but its result could not be delivered to the parent
(failed_transient)"*. That is **not** a deployer defect and has nothing to do with cross-links: it is
`bugs_closed/274_…_completed_workflows_cannot_deliver_their_result_to_the_parent_fleetwide.md`, and
fleet-wide it dwarfs this lane (`page-rerender` **5,869**, `build-dispatch-loop` **2,954**,
`feed-ingester` **1,218** …). `tool-deployer`'s 10 are an unremarkable tail.

⚠ **Consequence for the count, and it cuts the right way:** those rows only capture runs whose
DELIVERY failed. Successful runs leave no error row at all, so **10 is a floor, not a total** — the
12 emitted items across 5 tools are an independent lower bound on successful runs. Do not quote
"10 runs" as the number of times `tool-deployer` has executed.

### 14.5 353's open list after this

- **(b)** the new arm is **live but STILL UNEXERCISED** — re-checked after the `v1.0.1334` roll:
  **4** tool births since the first roll, **all 4** `no_related_pages`, **0** cross-link items
  created today. Guard 2 has still not been reached once. The zero remains uninformative for the
  same reason as §12.2; the first non-zero
  `emitted_ungated_build_enqueued_by_caller` row is what closes it.
- ~~(c′) `tool-deployer`~~ → **CLOSED, §14. Not a defect.**
- ~~(d) regeneration residual~~ → moved to `bugs_open/379`.

**Everything on this bug is now closed except (b), which is a wait, not a task.**

⚠ **Sharpening on the adjacent generation-side concern** (handoff §3.4, still unfiled): `no_related_pages`
has **12** rows all-time since 07-31 and **4 of them are today**. A third of every occurrence ever
recorded landed in one day, and it is the same day tool births stopped reaching Guard 2 entirely.
Still **[UNMEASURED]** as a rate — births/day is not established — but it has moved from "two data
points is not a trend" to worth someone's morning.


## 15. 2026-08-25 — **ROUTED HOME as the `staged_component_build` lane closes. Item (b) is the ONLY thing left, and the producer everyone would try it with CANNOT satisfy it.**

The lane that has carried this bug since 08-22 is closing. §12.6's list reduces to one item, and it
comes with a correction that will otherwise cost the next session a wasted filing.

### 15.1 The only open item

**(b)** the ungated-emit arm is live and has **never fired**. `emitted_ungated_build_enqueued_by_caller`
count: **0** [MEASURED 2026-08-25 ~10:2xZ]. Nothing else is open here.

### 15.2 ⚠ The correction: a `webdesign_tool_rebuilds`-style filing can NEVER exercise it

I wrote in the 08-25 handoff that migration 602 (the related-pages picker) would unblock (b) "as a
side effect", because a birth carrying pages now reaches Guard 2. **Half right, and the half that
matters is wrong** — proven by the first real filing on 2026-08-25:

- The picker ran (09:50:53Z) and supplied two pages; two `tool_crosslink` items were created.
- **No skip row of any kind was written**, and (b)'s INFO row did not appear.
- Because the tool page `/tools/smart-contrast/index.html` was **created 2026-07-25** and was already
  `deployed` when the emitter ran, so Guard 2 took the ordinary `pageLive` arm — which emits ungated
  and writes **no** INFO row. That arm is not (b)'s arm.

**That lane's whole programme rebuilds ported tools AT THE SAME URL**, so its page is always already
live. No volume of their filings will ever reach the new arm.

### 15.3 What WOULD exercise it, stated so nobody re-derives it wrong

A tool birth where **all three** hold at the moment the emitter runs: the spec (or the picker) supplies
`related_pages`; the tool page is **NOT yet live** (`build_status` neither `deployed` nor
`needs_rebuild`); and no open `needs_content_page` / `page_rerender` / `needs_page` item exists for it
yet. That is a **brand-new** tool page on a site, not an adoption of an existing one — the shape
`dartsonline`'s five births had on 08-23, one day before the fix rolled.

```sql
SELECT count(*), min(occurred_at) FROM agent_error_log
 WHERE error_code = 'tool_crosslink_not_emitted:emitted_ungated_build_enqueued_by_caller';
```
Non-zero ⇒ (b) closes and so does this bug. **Still zero ⇒ check whether any NEW-page tool birth has
happened at all before concluding anything** — the same demand-control discipline this bug's §12 was
already burnt by twice.

### 15.4 Everything else is closed and evidenced

(a) forward fix LIVE + council-APPROVED (`642ecc3c`); (c′) `tool-deployer` runs, emits, and had already
repaired one of this bug's own casualties (§14); (d) → `bugs_open/379`, unowned; the damage half §11.
The `330` producer split that kept (b) starved is fixed and live (`bugs_open/330` §12–§13).
