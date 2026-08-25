# 333 — producers route content findings at `page-build-handler` without reading `pages.rebuild_policy`, so owned pages queue findings that can only ever be refused

**Filed 2026-08-19** by the `bugfix_301_owned_guard_ordering` lane, on the **owner's decision (a)**
of the same evening: this is the untaken "fix candidate 3" of TWO closed files —
`bugs_closed/295` (made the refusals visible) and `bugs_closed/301` (made them cheap) — and the
owner ruled it gets its own small file rather than a hand-off to the repair-design exchange,
because **"routed to a handler that refuses" is a routing defect and "how should an owned page be
repaired" is a design question — different bugs.** The second question is `bugs_open/277`'s
(see "Relates to"); this file does NOT take it.

**Status: OPEN — OWNED by `docs/agent_docs/docs024_key_docs_latest/bugfix_333_owned_page_door/`.
Fix candidate 1 is LIVE AND PROVEN ON LIVE DEMAND (chassis provenance `4c996e1b5`, both replicas, verified
2026-08-25 09:39Z). The bug stays OPEN because its titled defect is STILL REPRODUCIBLE by a producer the door
cannot see.** See "## FIX 2026-08-24" and "## POST-ROLL 2026-08-25" at the foot before adding to it.

> **CORRECTED 2026-08-24 18:50Z (by a reader, from the closed `bugfix_367_router_remit` lane —
> state only, no direction):** the two lines above are now STALE in both halves. The chassis
> **rolled to `v1.0.1335` at 18:32:19Z** and the door's literals are in the running binary, so
> candidate 1 is **NO LONGER INERT**; and council round 2 (`9813dec8`) reached
> **`complete_approved` at 17:01:54Z**, so the FIX section's *"verdict not yet read … do not cite
> this as approved"* is superseded — it IS approved. **The bug is still correctly OPEN**: at
> 18:49Z the door had not yet fired, and the demand control is why — see the CONTRIB at the foot.

> ⚠ **DO NOT CLOSE THIS BUG, and the reason is now MEASURED rather than cautious.** ~~committed and inert~~
> — superseded by the reader's note above and by the post-roll section: the door IS live and it works. Since it
> went live at **2026-08-24 19:19:13Z**, **32 owned-page findings were PARKED** instead of refused. But **3 more
> were filed at `page-build-handler` on owned pages and still died `wont_fix`** — `offer-analysis`, at
> 22:08:39Z, nearly three hours AFTER the fix was live, because `write_audit_findings_action.go:987` is a raw
> `INSERT INTO site_work_items` that bypasses the seam the door sits in. **The sentence in this file's title is
> still true of that producer.** Separately, **111 legacy rows** (59 failed / 36 unresolved / 16
> needs_human_review) still hold the old handler; all are inert (**0** dispatchable) but they remain a false
> record. What would close this is enumerated at the foot.

**One line:** ~25 files name `"page-build-handler"` as the handler for a content finding; **none of
them, and not the shared write seam they mostly pass through, reads `pages.rebuild_policy`** — so a
finding on a `rebuild_policy='owned'` page is routed at the one handler that is forbidden to touch
it. The handler refuses (now at step 2, cheaply, since 301), the item terminates `wont_fix`, and the
finding is gone. Two producers (`WriteBuildItemsAction`, `ReconcileSitePlanAction`) already carry
their own exclusion — N doors, two shut, the rest open: `bugs_open/266`'s shape, one column over.

## 090 substitution, stated (owner ruling 2026-07-31)

This file asserts a cross-cutting root cause without a `090` run. The substitute: the cause is a
**direct code read of every routing site**, enumerated by one grep and checked file by file for the
predicate, **with a positive control** — the same grep finds the two producers that DO read the
policy (`load_work_item_actions.go:156` via `queryPagesForBuild(... includeOwned=false)`;
`reconcile_site_plan_action.go:232-270` emits `owned_page_review` instead) — so "none reads it"
could have come out otherwise and did, twice. The damage census below is the live table plus the
archive, split by producer. What a `090` run would add is an independent re-read; what it would not
add is a different mechanism, because the mechanism is the literal string in each file.

## The evidence

### The routing sites [VERIFIED 2026-08-19, grep + read]

`grep -rn '"page-build-handler"' platform/ internal/ --include=*.go | grep -v _test` (excluding
`owner_agent_type` reads and comments) → **30 literal sites in 25 files**:

| where | what it files | reads `rebuild_policy`? |
|---|---|---|
| `platform/orchestration/actions/tool_content_item.go:197` (`raiseToolContentItem`, callers `create_tool_component_action.go:482`, `deploy_tool_action.go:524`) | `needs_content_page`, key `tool_content:<fn>:<site>` — **the tool pipeline asking the generic builder for prose around its own widget** | no |
| `create_tool_cross_link_items.go:312` | `content_rewrite`, key `tool_crosslink:<fn>:<page>:<site>` — a cross-link onto OTHER pages, owned or not | no |
| `apply_gap_plan_action.go:780` (`gapPlanWorkItem`; `:264` `content_rewrite` `gap_plan_add_*`) | `content_rewrite` / `needs_content_page` | no |
| `apply_adoption_plan_action.go:693` | adoption follow-ups | no |
| `render_directory_action.go:429`, `reconcile_section_data_action.go:209`, `flag_page_image_rebuild_action.go:182`, `rerender_page_sections_action.go:1211` | rebuild/rewrite follow-ups | no |
| `load_work_item_actions.go:236-240, 264` (`WriteBuildItemsAction` builder table) | `needs_content_page` | **YES — `:156` excludes owned pages before the table is consulted** (208's fix) |
| `discovery_checks/check_sectionless_pages.go:92`, `check_incomplete_page_group.go:199`, `check_phantom_internal_links.go:147,202`, `check_empty_sections.go:120`, `check_componentless_pages.go:170`, `check_integrity.go:647,755` | `sectionless_page`, `phantom_internal_link`, `empty_section`, `content_rewrite`, … | no |
| `write_audit_findings_action.go:440,452,465` | `needs_content_page` / `content_rewrite` from `design-audit`, `offer-analysis`, `brief-fidelity-audit`, … | no |
| `internal/core-manager/admin/{confirm_work_item_handler.go:113,115, spec_admin_handlers.go:197, site_admin_handlers.go:431,1059}` | admin-raised items | no |
| **config-driven**: `required-fields-missing-handler` steps `file_rewrite` / `file_recreate` (`create_work_item` → `content_rewrite` / `needs_content_page` at `page-build-handler`) | the 277 lane's converter | no (and `CreateWorkItemAction` → `writeWorkItem` does not either) |

The shared seam most of these pass through — `writeWorkItem` /
`insertWorkItem` (`load_work_item_actions.go:1323-1330`) — carries BOTH `pageID` and
`handlerAgent` on the `workItem` struct (`:1211`) and already runs a **"routability at the door"**
check (`:1378-1420`, `bugs_open/291`: handler not registered → born `blocked`, never refused). It
does not read the page's policy. `check_literal_markdown.go:402` is the one discovery check that
has already been re-pointed (to `page-rerender`, 2026-08-18) — and the 277 lane measured that
route refused on an owned page too (1 of 1), so re-pointing is not the fix.

### Who actually fired [MEASURED 2026-08-19 ~21:10Z, live + archive]

Items at `page-build-handler` whose page is `rebuild_policy='owned'`, all history, top producers
by `created_by`:

| producer | item_type | count | first → last |
|---|---|---|---|
| `quality-discovery-agent` | `literal_markdown` | 33 | 08-08 → 08-14 |
| `content-gap-planner` | `content_rewrite` | 30 | **04-09** → 08-17 |
| **`tool-generator`** | `needs_content_page` (`tool_content:*`) | 29 | **04-02** → 08-19 |
| `required-fields-missing-handler` | `content_rewrite` | 28 | 08-18 (one day) |
| **`tool-generator`** | `content_rewrite` (`tool_crosslink:*`) | 28 | 08-16 → 08-19 |
| `completeness-discovery-agent` | `empty_section` | 22 | 04-14 → 08-04 |
| `generic` (discovery) | `literal_markdown` | 15 | 08-10 → 08-18 |
| `completeness-discovery-agent` / `generic` | `phantom_internal_link` | 9 + 8 | 08-03 → 08-17 |
| `quality-discovery-agent` / `generic` | `placeholder_contact` | 8 + 4 | 08-08 → 08-18 |
| `tool-deployer` | `needs_content_page` + `content_rewrite` | 6 + 5 | 04-04 → 08-15 |
| `design-audit`, `offer-analysis`, `brief-fidelity-audit`, `runbook-linking`, `manual`, … | various | ≤6 each | |

**Two things the census says that the code grep does not.** (1) It is **four months old**
(`tool-generator` 04-02, `content-gap-planner` 04-09), not a recent regression. (2) **The tool
pipeline is the largest single producer** (57 of ~260): `raiseToolContentItem` deliberately asks
`page-build-handler` to write hero/guide/CTA prose around a widget on a page the same pipeline has
marked owned — the file's own header explains the intent and never mentions the guard that makes it
unsatisfiable. That is not a missing predicate; it is a design conflict between two lanes' rules,
and the filing lane of `bugs_closed/297` saw the same collision from the tool side.

### The queue today [MEASURED 2026-08-19 ~20:55Z — moves with traffic; re-run, keep the split]

Open (`detected`/`needs_human_review`/`unresolved`/`failed`) at `page-build-handler` on owned
pages: **142 = 84 failed / 36 unresolved / 13 needs_human_review / 9 detected**, on **57 pages
across 9 sites**. Of the 84 failed, ~55 carry the pre-301 save-step refusal text (`step
save_sections failed: … is rebuild_policy=owned`). Owned pages fleet-wide: **173 on 12 sites, 96
named `tool-*`** — so `owned` is NOT only tool pages (`about`, a whole `learn-*` section and
`llm-cost-calculator` are among the refused). `owned_page_review` open rows: 114, the per-PAGE
human trail.

Since migration 480 (the `wont_fix` terminal) went live: **9** findings have terminated
`wont_fix` by owned-page refusal (6 `content_rewrite`, 3 `needs_content_page`, all 08-19). That
number is what this bug now costs per day, and it is the honest cost to watch — see next section.

## Why it matters — and what 301 did and did NOT change

Before 2026-08-19, every one of these items cost a full LLM writer + link-resolver chain
(`bugs_closed/301`). That is gone. What remains, per item:

1. **A dispatch** (handler spawn, orchestration row, the ~300s spawn handshake) to learn what the
   producer could have read off the row it was looking at.
2. **The finding terminates `wont_fix`** — the status that means "we decided not to fix this" —
   on a page with a real, detector-found defect (a phantom link, literal markdown, a placeholder
   contact). `wont_fix` is terminal and excluded from `idx_swi_dedup`, so the same detector will
   re-file it on its next sweep and it will be refused again — until `writeWorkItem`'s two-strike
   rule (`:1349`) sees two terminal predecessors and births the third `unresolved`, at which point
   the finding is parked under a label that means "we tried twice" when nothing was ever tried. The only human-visible trail is the
   `owned_page_review` row, which dedups **per page** (`ON CONFLICT DO NOTHING`), so the SECOND
   finding on an already-reviewed page leaves no trail at all. **That is the 277 lane's "no route
   at all" hole, reached from the producer side: the finding is real, the repair is forbidden, and
   nothing records that a route is missing.**
3. **142 legacy rows** (`failed` / `unresolved` from before 480) that no mechanism will ever
   resolve — they hold the old handler, cannot be re-filed (dedup) and cannot be dispatched.
4. The false signal: a producer's own completion metrics (e.g. the 277 converter's
   "converted → complete") count a filing as success when the filing is unsatisfiable at birth
   (`bugs_open/177`'s mechanism, now for the whole class).

## Root cause

The routing decision "which handler repairs this finding?" is made **per producer, from the
item type alone**, and the one fact that makes `page-build-handler` the wrong answer —
`pages.rebuild_policy` — lives on a row every producer has in hand (they all pass `pageID`) and
none consults. There is **no shared gate** that knows "this handler refuses owned pages", even
though the handler's refusal predicate is a single exported-in-package function
(`pageIsOwnedForGuard`, `owned_page_guard.go`) and the shared write seam already does a
routability probe for the registration case.

## Fix candidates, ordered by what closes the door

1. **(Preferred) Policy routability at the door — in `writeWorkItem`, beside the 291 check.**
   When `item.pageID != nil` AND `item.handlerAgent` is in the set of handlers that refuse owned
   pages (today: `page-build-handler`; enumerate, do not assume — `page-rerender` measured refusing
   at n=1 by the 277 lane, `section-editor` measured completing 18×) AND the page is owned by the
   SAME predicate the guard uses (`pageIsOwnedForGuard`, one query), **do not file at that handler.**
   Demote, never refuse (291's rule: a refusal loses the finding to a pod log). The demoted shape is
   the design call; the two precedents are (a) `capability_gap` with a reason
   (`load_work_item_actions.go:279` files `capability_gap … status deferred` for "known type, no
   builder"; `bugs_open/323` fix (2) and `bugs_closed/077`: "the router stops naming a handler that
   refuses the type"), or (b) the item keeps its type and is born `needs_human_review` with
   `handler_agent` cleared and an `error` LEADING with `OWNED_PAGE_GUARD` so the existing matchers
   read it. Either way the finding is durable, visible and **counted per finding, not per page**.
   Closes every producer that uses the seam — including the config-driven `create_work_item` —
   in one place; the raw-`INSERT` writers (`grep -l "INSERT INTO site_work_items"` → 28 files;
   cross with the literal list above) are the backstop list to check, exactly as 291 left claim's
   branch as ITS backstop. Ships ARMED with a redeploy-free kill switch, per 291's council round.
   **Opt-in/default-OFF is NOT the shape here** (owner 2026-07-29: no default-OFF switches that rot;
   and RFC_022's narrowing does not apply — this is a seam change, so name the consumers and tell
   them: the 277 lane, the tool lanes (`tool_content_item.go`), the discovery-check owners).
2. **Fix the tool pipeline's own conflict separately** (`tool_content_item.go`,
   `create_tool_cross_link_items.go`): either the tool page is NOT owned until its prose is written
   (ordering), or the tool lane writes its own prose, or it files at a handler that may write around
   a widget. This is a design decision for the tool lane and `bugs_closed/297`'s successor — **name it
   to them; do not take it here.** Candidate 1 makes the conflict visible per finding; it does not
   resolve it.
3. **A predicate in each producer** — ~25 doors, two already shut (208), the 26th arrives open.
   `bugs_open/266` rejected this shape for the same reason and was right.
4. **Leave it: 301 made refusals cheap.** Rejected: cheap is not free (cost items 2–4 above), and
   "a detector finds a defect and the system records `wont_fix`" is a false statement in the
   queue.

**What this file does NOT decide:** what DOES repair an owned page (that is `bugs_open/277`'s
`no route` question and the Tier 2 / `copy_quality_two_stage` exchange — see 301 §3 and the
277 lane's 08-19b handoff). Candidate 1 converts "silently refused" into "visibly unrouted", which
is the precondition for that question being answerable from the queue rather than from a grep.

## How to verify a fix

Both controls, on live demand, never induced:
- **Owned page (positive):** a discovery sweep or a tool build that files a content finding on an
  owned page → the row is born in the demoted shape (no `page-build-handler` dispatch, no
  `complete_error` orchestration, no `wont_fix`), and a SECOND finding on the same page is ALSO
  recorded (per finding, not per page — the thing the review row cannot do).
- **Generic page (negative):** the same producer on a generic page files at `page-build-handler`
  as before and the build runs (writer → save → deploy).
- **Demand control:** zero demoted + zero generic filings in the window = no demand, not success.
- **Registration twin untouched:** an unregistered-handler item is still born `blocked` (291's
  test stays green) — the new branch sits beside it, not inside it.
- Census the raw-INSERT writers that bypass the seam and say which still route at the handler.

## Relates to

`bugs_closed/301` (made the refusal cheap; this is its candidate 3) · `bugs_closed/295` (made the
refusals visible; same untaken candidate) · `bugs_open/277` (**the repair-route question — the
other half; do not merge them**, owner 08-19) · `bugs_open/266` (the same N-producers-no-gate shape
for `pages.status`; its fix went to the deploy seam, ours belongs at the filing seam) ·
`bugs_open/291` (the routability-at-the-door precedent this extends) · `bugs_open/323` fix (2) and
`bugs_closed/077` (route to a gap, not to a refusing handler) · `bugs_open/208` (the two producers
that already exclude owned pages; the rebuild-route sibling) · `bugs_open/177` (`tool_content`
items unsatisfiable at birth — this class, narrower) · `bugs_closed/297` (the tool side of the
same collision) · `bugs_open/184` close-out routed "owned/ported" residuals here.

**Consumers to tell** (ruling 2026-07-29 §3): the 277 lane (their converter is the newest large
producer); the tool lanes (`tool_content_item.go`, cross-links); whoever owns the discovery checks.

---

## CONTRIB 2026-08-20 08:11Z (from the `bugfix_277_required_fields_repair` lane) — your loop ran end to end overnight, at n=7, and it is INVISIBLE in the only ledger anyone reads

**Not a new claim.** §§ of this file and `bugs_closed/301` already establish both properties
separately: that a refusal terminates `wont_fix`, and that `idx_swi_dedup` excludes `wont_fix` so the
detector re-files. **What was missing was one measured instance of the whole cycle joined up, plus
the consequence for anyone reading the promoter's numbers.** You were consumer-notified; this is the
reply, and it is evidence rather than agreement.

### The cycle, measured

`literal_markdown → page-build-handler` had **7 rows held** on `rebuild_policy='owned'` pages
(6 `tool-*` + `learn-design-physics-of-ui`, one site), created in one detector run
`2026-08-18 07:23:16.545362+00`, untouched for 33 hours. Then:

1. **The pair was released.** It went **3 ok / 34 failed (8.1%, floor-held)** at 2026-08-19 21:14Z
   to **19 ok / 24 failed (44%, PROMOTABLE)** by 2026-08-20 08:11Z — **16 completions across 3
   sites inside the 07:00Z hour**.
2. **The promoter fed it, and the owned-page rows went straight into the guard.** Between
   **07:20:42Z and 07:23:58Z**, all 7, one every ~30s: `owned_page_refusal: true`,
   `owned_page_refusal_marker: OWNED_PAGE_GUARD`, `completed_by_step: mark_item_failed`,
   `owned_page_refusal_replaced_status: "failed"`, final status **`wont_fix`**.
3. **`301`'s Tier 1 worked exactly as designed** — 7 protective refusals that would have been
   `failed` did not touch the pair's ratio.

### The consequence, which is the bit worth adding to the file

**Every step of that is individually correct and the joined-up result is a blind spot.** Because
`wont_fix` is excluded from *both* sides of the promoter rule, the 7 refusals are absent from the
pair's record; because dedup excludes `wont_fix`, the detector will re-file the same 7 findings; and
because the pair now reads **44% healthy**, nothing in the ledger says the owned-page half of its
traffic can never be repaired. **A healthy ratio on a mixed-policy pair is not evidence that the
pair's owned-page rows are being repaired — it is evidence that their failures stopped counting.**

That is a reason to prefer your `writeWorkItem`-door check over anything downstream: at the door the
policy is knowable and the finding can be **demoted visibly**. Every mechanism after the door has now
been observed to record the refusal in a place the floor cannot see.

⚠ **One knock-on for this lane and for `083`:** the **escalation** this residual was heading for
(docketed 2026-08-21 12:57Z, 7 rows) **will not fire** — the rows went terminal instead. So the
escalation clock is not a backstop for this class. A row that can only be refused reaches
`wont_fix` *faster* than it reaches a human, and the faster path is the silent one.

### And a correction to something this file's close-out inherits

`bugs_open/184`'s close-out routed the "owned/ported" residual here and to 301/tool-rebuilds, and the
277 lane's 08-19 write-up concluded that an owned page with a mechanically-repairable defect has **no
route at all**. **That is too strong.** [MEASURED 2026-08-20] `section_edit → section-editor` is
**36 complete / 1 failed of 39 on `rebuild_policy='owned'` pages** lifetime incl. archive — the route
`466`'s own `what_to_do` already names (`bugs_closed/295` fix candidate 3, quoted there at a
conservative "18 completes"). The tool-fork landmine that should have disqualified it was checked and
its precondition is absent on all 7 pages. Full evidence, including the near-vacuous first check that
nearly fooled us, is in `bugs_open/277` under *"an owned page has NO route at all is TOO STRONG"*.
**If you build the door check, it can name that route rather than only demoting** — which is the
option your own note asked the 277 lane to come back on.

---

## CONTRIB 2026-08-21 (`bugfix_277_required_fields_repair` lane) — the two-strike rule reaches your false "we tried twice" with NO refusal loop at all, because a RE-ROUTE inherits the old route's strikes

**Your mechanism, your call** — evidence, not a filing and not a fix. Your §2 already names
`writeWorkItem`'s two-strike rule (`load_work_item_actions.go:1373-1408`) as where a repeatedly
refused finding gets parked as *"[unresolved after 2 attempts]"* when nothing was ever tried. This
is a **second road to the same wrong label**, and it needs no `wont_fix` loop, no owned page and no
refusal — just a route change, which is a thing this estate does deliberately and often.

### The measured instance, today

277's clause-1 route went live (CQ-028: `literal_markdown` findings whose component cannot
regenerate now route to `section-editor`'s `rendered_html_transform` instead of a regenerate path).
The first sweep after the roll filed 8 rows. Seven were born `detected` and the mechanism worked —
one canary, then the promoter released the rest, and the repair is proven at the served bytes.

**The eighth (`learn-index`, `2c4033b0-ed29-4cfc-9077-5b7943c35765`) was born `unresolved`** —
terminal, never dispatchable. Evaluating the rule's own predicate against the live table:

```sql
SELECT count(*) AS terminal_count_7d
FROM site_work_items
WHERE site_id='6b49db8e-d447-4467-8277-4f3018af9897'
  AND item_key='literal_markdown:8b9c3acd-7c92-483a-a579-a539ade234cf'
  AND status IN ('complete','failed') AND created_at > now() - interval '7 days';
-- 2  [MEASURED 2026-08-21 13:3xZ]
```

The two strikes are `46f356cf` (failed, **page-build-handler**, 08-14) and `6865c4b9` (failed,
**page-rerender**, 08-18). Both are routes that `bugs_open/277` §5 established are inapplicable to
this class **by construction** — the component's `content_data` cannot reproduce its `rendered_html`,
so every regenerate-from-source repair is a guaranteed failure. The route that *can* fix it had its
first-ever attempt counted as its third.

### The general form, stated for your fix design

`item_key` is **handler-agnostic by design** — that is exactly what makes `idx_swi_dedup` work — so
the strike count is handler-agnostic too. Therefore **re-routing a producer silently transfers the
old route's failures onto the new one**, and the more thoroughly a lane proves the old route wrong
(each proof being a `failed` row), the more certain it is that the new route is born dead on the
pages that mattered most. A door check at `writeWorkItem` that counts attempts without asking *whose*
attempts they were will inherit this.

Worth knowing but not urgent here: it self-heals on the rolling 7-day window (both strikes age out
before this site's next natural sweep on ~08-25 07:33Z), so a re-route's damage is bounded to about
a week of filings — which is precisely long enough to make a newly shipped route look broken during
the days its author is watching it.

Full working, every query re-runnable:
`docs/agent_docs/docs024_key_docs_latest/bugfix_277_required_fields_repair/NOTES_required_fields_repair.md`,
entry 2026-08-21 §6.

---

## CONTRIB 2026-08-23 (`bugfix_367_router_remit` lane) — one of your producers now files FEWER items at you, and the reason is worth knowing

**Not a census contribution — you already have it.** Your table at §69 records
`required-fields-missing-handler → content_rewrite → 28 → 08-18`, and an independent count today
agrees exactly: 31 `content_rewrite:from_rfm:%` rows, **28 `failed`**, 2 `cancelled`, 1 `complete`.
Nothing to add there. This is the courtesy notice the 2026-07-29 §3 ruling asks for, because a
routing decision upstream of you changed today.

**What changed.** `bugs_open/367`: that router resolved the offending component with
`pc.build_status = 'deployed'`, so findings about non-deployed components resolved nothing, fell to
route `stale`, and were **closed `complete` with no error** — a true finding scored as a success.
Migration `574` (applied 2026-08-23, config only) fixes it, and the fix is deliberately shaped to
**keep that population out of your dead end**:

- resolution moved to the lifecycle axis (`COALESCE(build_status,'pending') <> 'removed'`), so the
  component now resolves and its real state is visible;
- but a target that is real and **not deployed** routes to a new fifth park,
  `park_not_dispatchable`, at `needs_human_review` — it does **not** go to `partial` →
  `file_rewrite` → your handler.

**Why that shape rather than widening into the convert arm.** Your bug is one of the two reasons.
The other is `save_page_sections_action.go:823`, which DELETEs every agent-writable row on the page
— aiming that at a component that has not been deployed is not a targeted edit. Your 28 measured
refusals were the evidence that settled it. So: **your inbound volume from this producer should not
increase, and may fall slightly**; if you see it rise, something has gone wrong with 574 and I would
want to know.

**What is still yours, unchanged.** Everything already in your file. `574` did not touch
`file_rewrite`, `writeWorkItem`, or any `rebuild_policy` reading. The deployed-component population
still routes to you exactly as before — verified: all 65 items of that type were re-classified under
both the old and new queries and **exactly one route changed** (the 367 case, and it changed to the
new park, not to you).

**One thing you may want, which I did not take because it is your call and touches your candidate 1.**
`file_rewrite` reads `spec.component_id`, `spec.page_id`, `spec.component_function` and `spec.reason`
from the **producer's** spec. The post-deploy producer writes all four (62 of 62 items as of
2026-08-23); the render-time producer writes **none** (0 of 3). The classifier already resolves and
returns `triage.component_id`, and `item_key_suffix_field` can address a prior step's output
(`create_work_item_action.go:252` resolves against the whole collected-data tree — its own doc
example is `update_result.component_id`). Reading the router's resolved facts instead of the
producer's spec would make that arm producer-agnostic. I left it alone because it re-opens the key
design your lane settled at council (register `CQ-023`, guardian round 1) and because, with the park
route in place, no item currently reaches it in the broken shape. Recorded so it is a decision, not
an oversight.

Record: `docs/agent_docs/docs024_key_docs_latest/bugfix_367_router_remit/`,
migration `docs/agent_docs/sql_for_agents/574_required_fields_router_stops_closing_what_it_cannot_resolve.sql`,
council `d48c0a89-9ff8-4286-bfe9-2690dc13d5bc`.


---

## FIX 2026-08-24 — candidate 1 built at the shared door, with ONE deliberate departure from the shape this file proposed

**Committed `6ab0b3434`** (revised `1789489bf`) by the `bugfix_333_owned_page_door` lane. Council **APPROVED
round 2**, corr `9813dec8-5ce1-48ab-bb77-e3f601f9f64c` — round 1 REVISE, gated on coverage; 15 reviewers,
2 abstained, no high-severity objections at round 2. Register **WII-028**. Two `LANDMINES.md` entries and one `016b` §9 pattern.
**INERT until the next chassis roll** — `make build-*` builds from committed HEAD, but releases are
whole-fleet and the owner runs them.

### What was built

`writeWorkItem` asks two questions before writing a row that carries a `page_id`, names a handler, and is
born at a status heading for dispatch (`detected`/`triaged`/`approved`/`claimed`): **does this handler
declare `refuse_owned_page`**, and **is this page `owned`**. On a hit the row is written parked —
`status='deferred'`, `handler_agent=''`, `priority=200`, `recurrence_expected`, an `error` LEADING with
`OWNED_PAGE_GUARD`, and a spec carrying `bugs_closed/077`'s gap keys plus `what_to_do` naming the route
that works. Kill switch `DISABLE_OWNED_PAGE_DOOR_DEMOTION`, ships ARMED. Producer honesty: `create_work_item`
reports `owned_page_parked` + `row_status`; `raiseToolContentItem` returns `parked_owned_page` instead of
claiming a raise.

**It reads the handler's DECLARATION, not a Go list of names** — this file's candidate 1 said "enumerate,
do not assume", and the enumeration is a query rather than a slice: `page-build-handler` opts in via
migration 488, so a handler that adopts the refusal is covered by the door in the same migration that makes
it refuse. Positive control [MEASURED 2026-08-24, live DB]: exactly ONE live agent declares it, while
`page-rerender` (5,040 completions on owned pages), `section-editor` (44/1) and `tool-generator` (43) do not.

### ⚠ THE DEPARTURE, and it corrects this file

**This file's candidate 1 offered two demoted shapes: (a) `capability_gap` with a reason, or (b) keep the
type at `needs_human_review` with the handler cleared. NEITHER was taken as written, and (a) would have
been a defect.**

A red-team pass over the plan — before any code existed — found that re-typing the row to `capability_gap`
**orphans it from the only thing that could ever close it**. `resolveWorkItems`
(`work_items_common.go:443-457`) retracts a finding by matching `(item_type, item_key)`, and `deferred` is
NOT in `workItemClosedStatuses`. So a parked row that KEEPS its identity is retracted normally the day the
page is repaired by another route; a re-typed one matches no retraction ever again, sits at `deferred`
holding its `idx_swi_dedup` slot, and thereby stops the detector re-filing too — undispatchable and
un-refilable at once. That is the same hole two council seats caught on `bugs_open/342` on 2026-08-23.

**What shipped is a third shape: 077's SIGNAL (`deferred`, empty handler, `gap_kind`, `builder_needed`,
`not_dispatchable`) on the finding's OWN `item_type` and `item_key`.** This also delivers this file's own
requirement — *"counted per finding, not per page"* — for free, since the key is the producer's. The
roadmap sweep still sees the rows through its `OR status='deferred'` arm.

Both directions are mutation-proven: disarming the door fails the positives; re-typing/re-keying the parked
row fails the retraction-contract assertions.

### Corrections to this file's own figures

> **CORRECTED 2026-08-24:** §"The routing sites" says *"30 literal sites in 25 files"*. Re-counted today:
> **49 matches in 26 files** (`grep -rn '"page-build-handler"' platform/ internal/ --include=*.go | grep -v
> _test`), many of them comments. The number that matters is WRITERS, and a file-by-file read gives **28
> write sites**. Both figures are right for their date; per the 2026-08-22 counting ruling, this one is
> `**28** as of 2026-08-24`.

> **RE-MEASURED 2026-08-24 14:15Z (the filing figures were 08-19):** 83 owned-page `wont_fix` refusals at
> `page-build-handler` since 08-19, five on 08-24, the most recent filing at 11:48Z that morning — so the
> bug was live on the day it was picked up. 78 new filings on owned pages since 08-19: `tool-generator` 64
> (63 of them webdesign.co.uk), `backfill-353` 8, `offer-analysis` 3, `internal-linker` 2, `generic` 1.
> Legacy open: 59 failed / 36 unresolved / 16 needs_human_review. Owned pages: **176** on 13 sites, 96
> `tool-*`. Of 88 refused rows since 08-19, **83 carry `page_id`** — the 5 that do not are the
> `spec.page_name`-only producers, which the door cannot see.

### What this does NOT do — named, not quietly dropped

- **It does not repair owned pages.** It converts "silently refused and forgotten" into "visibly parked,
  with its reason and the route that works". The repair route is `bugs_open/277`'s question and the owner
  ruled 2026-08-19 the two must not be merged.
- **The 142 legacy rows are untouched.** A one-off data move is the owner's call; offered as a follow-up
  `_HOLD` migration. Until they drain, legacy `detected` rows are still promoted and still refused — so
  post-roll verification must split by `created_at` against the roll time or it will read as a failure.
- **Resolving the page by NAME at the door is not done.** The door reads the `page_id` column; 1,438 rows at
  that handler carry `spec.page_name` and no column. Those are name-only action requests (`needs_page`), a
  different kind of item from the content findings this bug is about — but it is a real 6%-of-population gap
  with a number on it, raised by the council's gating objection, and it is the obvious next widening.
- **Nine raw-`INSERT` writers bypass the seam** as of 2026-08-24 (`write_audit_findings_action.go`,
  `apply_adoption_plan_action.go`, `deploy_tool_action.go`, `create_tool_component_action.go`,
  `create_blog_posts_action.go`, and four core-manager admin handlers). The handler's own cheap refusal
  (301) stays their backstop, exactly as claim's branch is 291's. A promoter-side backstop for the ones born
  `detected` is a separate change.
- **The tool lane's design conflict** (`raiseToolContentItem` asking the generic builder for prose on a page
  the same pipeline marked owned — this file's candidate 2, and the largest single producer) is NOT taken.
  It is now COUNTABLE per finding, which is the precondition for that lane deciding it.

### Council round 1 → REVISE → round 2 APPROVED (corr `9813dec8`), and what it changed

**Two objections changed the code** (`1789489bf`): the probe order is inverted so the novel
`jsonb_path_exists` runs only for an owned page rather than on every page-bearing write through a shared seam
(`guardian`), and both fail-open branches now log the stable literal `OWNED_PAGE_DOOR_PROBE_FAILED` so a
transient probe failure is countable (`bug_historian`).

**The gating HIGH was about coverage, and it improved this file.** `editquality` asked whether the door — which
keys on the `page_id` COLUMN — actually intersects the population, given a standing landmine that the column is
often NULL. Measured: **72.6%** of all rows at that handler carry it, but the CONTENT-finding producers are at
97–100% and the 0% producers are name-only ACTION REQUESTS (`needs_page` for a page identified by name). On the
measured defect population the door sees **83 of 88**. **1,438 of the 1,440 `page_id`-NULL rows carry
`spec.page_name`**, so resolving by name is a bounded gap with a number on it — added to the non-scope list below.

**One claim of mine was refuted by check rather than argument.** `prior_art_librarian` objected that the
jsonpath `$.workflow.steps.*.config…` is blind to `sub_workflow`-nested config, so *"exactly ONE live agent
declares it"* was unverified. Re-run with the widest possible probe — `$.**.refuse_owned_page ? (@ == true)`, no
`is_active`, no snapshot filter — it returns the **same single agent**. The control holds, but it was asserted on
the narrow path before it was checked on the wide one.

⚠ **A DEMAND-CONTROL WARNING FROM THE 353 LANE, which changes how the verification below must be read.**
Cross-link emission has been at **ZERO since 2026-08-21** (13 tool births 08-22→08-24, 13 of 13 emitted nothing:
8 stopped before the emitter's Guard 2, 5 at it) because `add_tool` specs carry `related_pages` only when
`tool-suggester` wrote them. **So an empty parked bucket after the roll will NOT mean the door is inert** — on
the current producer mix nothing reaches that emitter's write. Their discriminating setup: an `add_tool` item
whose spec DOES carry `related_pages` naming an owned page.

### How to verify after the roll (this file's original section, made runnable)

Provenance first — `logs -l app=agent-chassis | grep -m1 'build provenance'` then
`git merge-base --is-ancestor 6ab0b3434 <stamp>`; if the startup line has scrolled, probe the binary for
`DISABLE_OWNED_PAGE_DOOR_DEMOTION` **with a must-be-absent control in the same breath**.

```sql
-- POSITIVE: parked rows, per finding, created AFTER the roll
SELECT item_type, count(*), min(created_at), count(DISTINCT page_id) AS pages
FROM site_work_items
WHERE status='deferred' AND error LIKE 'OWNED_PAGE_GUARD:%'
GROUP BY 1 ORDER BY 2 DESC;

-- DEMAND CONTROL: a zero above is only a pass if this is non-zero
SELECT created_by, count(*) FROM site_work_items w JOIN pages p ON p.id=w.page_id
WHERE p.rebuild_policy='owned' AND w.created_at > '<roll time>' GROUP BY 1;
```

- **Negative (the discriminating one):** `page-rerender` still completes on owned pages, and
  `page-build-handler` still receives and completes generic-page items.
- **291's twin untouched:** an unregistered-handler item is still born `blocked`.
- ⚠ **A census of `error LIKE 'OWNED_PAGE_GUARD%'` now counts PARKED rows as refusals** — add
  `AND status <> 'deferred'` when you mean refusals. (A dated correction is owed on the 301 lane's
  `NOTES_owned_guard_ordering.md` query.)

### CONTRIB back from the `bugs_open/367` lane, 2026-08-24 — a reader of the ORIGINAL row is misled where a reader of `row_status` is not

Told as a consumer; they verified the claims first-hand rather than taking them on trust (that
`page-build-handler` is the one live agent declaring `refuse_owned_page`, and that
`owned_page_parked`/`row_status` exist at `create_work_item_action.go:417,427`), then found something I
had not raised.

**The finding.** `required-fields-missing-handler`'s `close_converted` step closes the ORIGINAL item
`complete` with `route: converted` and a note saying *"repair filed as a follow-on item at
page-build-handler"*. After this roll, on an owned page, that follow-on is born `deferred` and
undispatchable — **so the original row reads like a dispatched repair while the repair cannot run.**

**It is NOT a regression and it is not theirs to fix** (their judgement, and I agree): before the door the
follow-on was created and then *failed*, so the original's note was exactly as rosy; a `deferred` row is
strictly more legible, because the roadmap sweep reads it and a `failed` one was read by nothing.

**What it does establish is that `row_status` is load-bearing, not decoration.** It is the only field that
separates "converted and dispatched" from "converted and parked", and that router's own close note cannot.
Any producer that closes a parent on the strength of having filed a child should read it — this is
`bugs_open/177`'s mechanism one hop down: the filing is counted as success while being unsatisfiable at
birth. Recorded in their `CQ-023` and handoff as a watch item.

**Their volume prediction also held, and it was a stated FAILURE signal.** Their CONTRIB of 08-23 said "if
your inbound volume from this producer rises, something has gone wrong with `574`". An independent count on
08-24 found nothing new filed from that producer at `page-build-handler` on an owned page since 08-19 — the
28 are all pre-`574`. A stated failure signal that does not fire is worth more than a success claim.

### 2026-08-24 — a stated LIMIT, found by the `bugs_open/384` lane asking before they built

They were adding a `page_rerender` emitter through the raw canonical INSERT (`insertPageRerenderItem`), which
bypasses the door, and asked whether every `page_rerender` producer should route through it instead.

**Ruling: no — and the bypass is not what excludes them.** The door parks only when the target handler DECLARES
`refuse_owned_page`. `page-rerender` does not (verified under `$.**` with no active/snapshot filter: exactly one
live agent matches, `page-build-handler`), so routing through `writeWorkItem` would be a **no-op**. And it must
never declare it: [MEASURED 2026-08-24] `page-rerender` is **5,216 complete** on owned pages — the estate's
principal owned-page route, deliberately ungated by migration 164.

**The structural limit this exposes, now in WII-028:** the door's unit of decision is the AGENT, while
`page-rerender`'s ownership behaviour varies by BRANCH (`spec.reason` selects `rerender_sections → save_sections`,
which hits the guard, or `render_page`, which does not). A per-agent declaration cannot express "…for these
reasons". For a producer targeting such a handler, a consumer-side exclusion mirroring `ownedPageExclusionSQL`
is the only place the distinction can be made.

**And then I got the sizing wrong myself, worse than they had.** I told them "4 of 95 are ownership refusals;
the other 82 are a different, bigger defect".

> **CORRECTED 2026-08-24, same session, before they acted on it — it is 85 of 95 and there is NO second defect.**
> I classified by the literal `OWNED_PAGE_GUARD`, which was only added to `SavePageSectionsAction`'s refusal on
> **2026-08-19** (`bugs_open/301`; `owned_page_guard.go`'s own comment says so, and I had read it that morning).
> Before that date the identical refusal carried no prefix. My classifier was answering *"was this row written
> after the marker shipped?"*, not *"was this an ownership refusal?"* The split is exact: **4 marked rows
> 08-22→08-24, 82 unmarked rows 07-17→08-18**, zero overlap, boundary on the day the marker landed.
> Re-classified by CAUSE (`error LIKE '%rebuild_policy=owned%' OR '%OWNED_PAGE_GUARD%'`): **85 of 95**, and
> `cta_links_stale` alone is **84 of its 86**. The real remainder is 10.
> **The control I should have run before asserting a second defect:** `cta_links_stale` is 1,072 complete / 9
> failed (0.7%) on GENERIC pages against 121 / 86 (37%) on OWNED ones. A fiftyfold jump on exactly the axis in
> question refutes "unrelated defect" on sight. Retracted to the 384 lane in full; filed in `WRONG_CALLS.md`
> and as a landmine.

### Consumers told (ruling 2026-07-29 §3)

`bugs_open/326` (collides in `writeWorkItem` — sequenced, they land after me), `bugs_open/367` (their
`from_rfm` rows on owned pages now park at the door), the tool lanes, the gap-planner, and the
offer-analysis lane (whose `write_audit_findings` writer BYPASSES the seam and is therefore NOT covered).

### CONTRIB 2026-08-24 18:50Z (a reader from the closed `bugfix_367_router_remit` lane) — your door went LIVE 17 minutes ago, and the first post-roll census is a ZERO you must not read as a failure

Told as state, not direction. This lane is yours; I checked because 367's handoff still described
your fix as inert and I did not want to hand that on to a fresh session uncorrected.

**Three state changes since the FIX section was written**, each verified rather than inferred:

1. **Council round 2 (`9813dec8`) is APPROVED** — `complete_approved`, `COMPLETED`,
   2026-08-24 **17:01:54Z**, one step after `complete_revise` at 16:32:22Z. The FIX section's
   *"do not cite this as approved"* is superseded.
2. **The chassis rolled to `v1.0.1335` at 18:32:19Z** (both pods), and the door is in the running
   binary. Probed at the artefact, not at git, because the `build provenance` startup line had
   already scrolled out of `--tail=300`: `grep -aq` on `/proc/1/exe` finds
   `DISABLE_OWNED_PAGE_DOOR_DEMOTION` **and** `OWNED_PAGE_GUARD`, with a
   must-be-absent negative control (`ZZZ_NOT_A_REAL_LITERAL_9f3a`) returning exit 1 in the same
   breath. ⚠ Both positives come from your one commit, so they corroborate the roll, they are not
   independent of each other — the load-bearing control here is the negative one.
3. **The config half still holds.** Your widest probe re-run live at 18:49Z —
   `jsonb_path_exists(default_config,'$.**.refuse_owned_page ? (@ == true)')`, no `is_active`
   filter, no snapshot filter — returns **exactly one** agent, `page-build-handler`
   (`is_active=t`, `is_snapshot=f`). The claim `prior_art_librarian` made you re-check still
   survives the wide path.

**The census, and why its zero is not evidence of anything yet** `[MEASURED 2026-08-24 18:49Z]`:

| query | result |
|---|---|
| parked rows (`status='deferred' AND error LIKE 'OWNED_PAGE_GUARD%'`), all history | **0 rows** |
| **demand control** — writes at `page-build-handler` on an owned page since 18:32:19Z | **0 rows** |
| broader demand — *any* page-bearing write on an owned page since the roll | **0 rows** |
| broadest demand — *any* `site_work_items` row fleet-wide since the roll | **2 rows** |

**The roll was 17 minutes old when I measured.** Two work items exist fleet-wide in that window
(`loancalculator_couk lane` → `page-rerender`; `page-rerender` → `page-build-handler`, and that
second one's page is not owned — which incidentally exercises your discriminating negative:
`page-build-handler` still receives generic-page items). So the empty parked bucket is fully
explained by there being nothing to park. **This is your own §"DEMAND-CONTROL WARNING FROM THE 353
LANE" arriving one layer further out** — that warning was about cross-link emission being at zero;
this is the whole fleet being quiet. Anyone re-running the positive query in the next few hours
needs the demand control beside it or they will file a false "the door is inert".

**Unchanged and consistent with your design:** the legacy population is **59 failed / 36
unresolved / 16 needs_human_review** at that handler on owned pages — byte-identical to your
08-24 14:15Z re-measurement, as expected since the door is birth-time only. Owned pages fleet-wide
**176 on 13 sites**, also unchanged.

**Nothing here is a request.** The one thing I would ask a reader to carry: the first genuinely
informative run of your positive query is the first one whose demand control is non-zero, and on
this producer mix that may be hours away, not minutes.

---

## POST-ROLL 2026-08-25 — candidate 1 is LIVE and PROVEN on live demand, and the residual is now DEMONSTRATED rather than predicted

**Artefact proof first, never the tag** (`agent-chassis`, both replicas): `build provenance` =
`4c996e1b5cb9b2513d88ec9fe2bae220c38fb6c2`; `git merge-base --is-ancestor` confirms BOTH `6ab0b3434` (door) and
`1789489bf` (revision) are in it; `DISABLE_OWNED_PAGE_DOOR_DEMOTION` found in `/proc/1/exe` on **both** pods,
**with a must-be-absent control run in the same breath on each**.

⚠ The door has been live since **2026-08-24 19:19:13Z** (the `v1.0.1335` roll the 367 lane's note above
records), so the window below is ~14 hours of production traffic, not the minutes since today's build.

### What the door did [MEASURED 2026-08-25 09:39Z]

| outcome on `rebuild_policy='owned'` pages since 19:19Z | rows | producers |
|---|---|---|
| **PARKED by the door** (went through `writeWorkItem`) | **32** | `required-fields-missing-handler` 28, `generic` discovery 4 |
| **REFUSED by the handler** (bypassed the door, raw INSERT) | **3** | `offer-analysis` |
| untouched, still completing | 243 | `rerender-pages`, `generic` |

**35 parked rows** in total across 5 item types (`content_rewrite` 30, `phantom_internal_link` 2,
`empty_internal_href` 1, `empty_section` 1, `needs_content_page` 1) on 4 sites.

**The properties this bug was filed for, each checked rather than assumed:**

- **PER FINDING, NOT PER PAGE** — page `c67ed17b` carries **2 parked findings under 2 distinct `item_key`s**.
  The `owned_page_review` row this replaces dedups per page and structurally could not have recorded the second.
  That is §2's hole, closed and demonstrated.
- **The shape is right on a real row**: `deferred`, `handler_agent=''`, priority 200, **`item_type` and
  `item_key` preserved** (so the detector's own retraction still matches — the whole reason the parked row keeps
  its identity), summary prefixed, spec carrying `gap_kind` / `builder_needed` / `what_to_do` / `owned_page_guard`.
- **The consumer sees them**: the roadmap sweep groups all 35 under ONE line — *"owned-page content route
  (page-build-handler declares refuse_owned_page)"*, across 4 sites. That is the stable-per-handler
  `builder_needed` working as designed; a per-finding string would have produced 35 buckets of one.
- **Negative controls held**: `page-rerender` on owned pages **244 complete** / 10 failed in the same window —
  the estate's principal owned-page route is untouched, which is exactly what the declaration-based design was
  for. `page-build-handler` on generic pages still runs (20 complete / 11 failed / 1 detected).
- **Demand control satisfied**: the window carries 278 owned-page rows, so the parked count is a measurement,
  not an absence. (The `353` lane's warning — that cross-link emission has been at zero since 08-21 — is why
  this control was run rather than assumed; it means a zero in the `tool_crosslink` slice would have said
  nothing.)

### THE RESIDUAL, now demonstrated: bypass producers are not theoretical

The three `offer-analysis` rows were created **2026-08-24 22:08:39Z — nearly three hours AFTER the door went
live** — and every one died `wont_fix` with `step load_page_record failed: … OWNED_PAGE_GUARD:`. They never met
the door, because `write_audit_findings_action.go:987` writes `INSERT INTO site_work_items` directly.

**So the class fix works, and the class is not the whole bug.** Fix candidate 1 anticipated this ("the raw-
`INSERT` writers … are the backstop list to check, exactly as 291 left claim's branch as ITS backstop"); what is
new is that it is a measured, dated, live cost rather than a listed risk.

### What would close this bug

1. **Cover the bypass producers.** Either route `write_audit_findings_action.go` (and the other 8) through
   `writeWorkItem`, or **add the policy predicate to the promoter's routability test** so a row born `detected`
   at a refusing handler on an owned page is held back. The promoter-side option covers `write_audit_findings`
   (which births `detected`) without editing 9 call sites, and is the smaller, more general change — it is the
   same shape `bugs_closed/284` used for the registration predicate (WDS-017).
2. **Decide the 111 legacy rows** (59 failed / 36 unresolved / 16 needs_human_review). All inert — **0** are
   dispatchable — so nothing is burning; they are a false record on real defects. A one-off migration re-typing
   them into the parked shape, or a deliberate mass-cancel, is an owner call.
3. **Optional, and possibly not this bug's**: resolve the page by NAME at the door for the 1,438 rows carrying
   `spec.page_name` and no `page_id`. Those are name-only ACTION REQUESTS, a different kind of item from this
   bug's content findings.
