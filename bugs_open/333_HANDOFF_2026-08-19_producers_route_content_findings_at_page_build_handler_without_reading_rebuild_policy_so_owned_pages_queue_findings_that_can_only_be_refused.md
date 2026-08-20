# 333 — producers route content findings at `page-build-handler` without reading `pages.rebuild_policy`, so owned pages queue findings that can only ever be refused

**Filed 2026-08-19** by the `bugfix_301_owned_guard_ordering` lane, on the **owner's decision (a)**
of the same evening: this is the untaken "fix candidate 3" of TWO closed files —
`bugs_closed/295` (made the refusals visible) and `bugs_closed/301` (made them cheap) — and the
owner ruled it gets its own small file rather than a hand-off to the repair-design exchange,
because **"routed to a handler that refuses" is a routing defect and "how should an owned page be
repaired" is a design question — different bugs.** The second question is `bugs_open/277`'s
(see "Relates to"); this file does NOT take it.

**Status: OPEN, unowned.** Nothing in the tree is changed by this filing.

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
