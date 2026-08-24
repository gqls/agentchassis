# HANDOFF 2026-08-24 — continue here (`bugs_open/206`)

**Supersedes `HANDOFF_2026-08-08b_continue_here.md`** (keep it; accurate for its own day).
Read `bugs_open/206` bottom-up first — its last four sections are today, in order.

## State in one paragraph

The 08-08 fix holds (both vetcomparison pages re-verified serving, and they survived the 08-23
fleet re-render). Today's work found the class **still live through a second producer**:
`reconcile_site_plan` hardcoded `handler_agent='page-build-handler'` and never consulted the
builder map, parking five typed pages on three other sites — one an `entity-directory` page
unbuilt for fifteen days while its builder ran. Fixed by a single shared routing authority
(`builder_routing.go:builderForPageType`, register **BLD-027**), council **APPROVED** at round 6
(corr `52dbd067-10ed-4a6e-84eb-3fbf47d099dd`), **live and binary-verified on v1.0.1334**.
**The bug stays OPEN**: nothing has yet been proven at the artefact, and five rows are still
parked.

## What is DONE (do not redo)

- **Code, 6 commits**, all credited REVIEWED by `098`: `d1aa231aa` (the fix), `0baa8a107`
  (operator recipe comment), `03e2bbdb7` (round-2: gap row's `handler_agent` empty),
  `90448d175` (routed-handler-must-be-registered test), `200d54bdf` (round-4: `section-index`
  held out), plus doc commits.
- **Live verification at the binary**, both replicas, negative controls 0:
  `builderForPageType` ×2 and the round-4 log literal `page_type not in the builder map` ×1 —
  the literal is what dates the binary to the *approved* revision, not merely to "some 206 code".
- **Docs**: all five standing docs now exist (README and RUNBOOK were created today), plus
  `LANDMINES` (two entries: the two-doors trap, and a further-correction to the 090 budget
  entry), **BLD-027** + index row, and **eleven** `WRONG_CALLS` entries.
- **Cross-lane**: notice filed in `bugs_closed/187` (whose deliberate "reconcile NOT guarded"
  decision one arm of this touches, revert offered); `bugs_open/345` told the ownerless trio is
  measured-red; `loanzy_uk_example_site` supplied the greenfield measurements.

## What is LEFT — three items, in priority order

### 1. Prove the fix at the artefact. BLOCKED on another lane's authorisation.

**The closure test both lanes originally wrote CANNOT FIRE** — see the POST-ROLL section of
`bugs_open/206` and `WRONG_CALLS` (eleventh entry). Two reasons: `reconcile_site_plan` has no
timer (it runs inside a build/publish pipeline, so a quiet site never re-reaches the fixed code),
and the parked row **holds its own `item_key`**, so reconcile skips the page as "queued" and the
new routing never applies to it.

**The honest proof is the MINT, not the page:**

1. Free the key — close the parked row to a **terminal** status (`cancelled`/`wont_fix`; **never**
   `complete`) with the reason in `error`.
2. Trigger a build/publish for the site so reconcile actually runs.
3. **Assert on the new row**: a fresh `needs_page:brand-directory-index` carrying
   `handler_agent='directory-build-handler'`, `created_by='reconcile_site_plan'`, with nobody
   having set the handler by hand. The page building and the home-page link going live follow.

Best case is `garden-tools.uk/brand-directory-index` (`entity-directory`, linked from that site's
own home page, on an unaided greenfield build). **Step 1 is an operator action on the
`loanzy_uk_example_site` lane's deliberately-unrepaired site — asked, not assumed.** They have
been messaged with three options (they run it / they authorise us / we record the gap honestly
and leave the site pristine). **Check for their reply before doing anything on that domain.**

⚠ **A hand re-triage fixes the PAGE and proves NOTHING about this fix** — setting `handler_agent`
yourself only re-demonstrates that `directory-build-handler` works, known since 08-08.

### 2. The five parked rows (an operator action, post-roll, now safe)

`garden-tools.uk` brand-directory-index / brand-profile / buying-guides-index,
`dartsonline.com` brand-detail, `loanzy.uk` guides-index. RUNBOOK step 4 has the recipe and the
`priority ASC` trap. Expected outcomes differ by type and **that is by design**:
`entity-directory` → routes and builds; `entity-page` → deferred `capability_gap` (no builder,
deliberately); `section-index` → **still parked**, because the entry is held out (below).

### 3. Two follow-ups that are NOT this lane's to sneak in

- **The `WriteBuildItemsAction` swap.** `section-index` is deliberately absent from
  `builderForPageType` so the map stays byte-identical to `WriteBuildItemsAction`'s inline copy —
  with it present the two producers disagreed on one page_type and, since both mint the same
  `item_key` under `idx_swi_dedup`, the first writer silently wins (round-4 guardian). **Add the
  entry in the SAME commit that makes `WriteBuildItemsAction` call this function**; two tests fail
  if you add it early and the producer test says so in its message. Blocked on that file's
  ownerless dirty hunk (breaks HEAD's build; fails three `TestUpdateWorkItemStatus` tests) —
  **re-check whether it is clean before assuming it still blocks.**
- **Residual (b), the larger class.** Two failures wear the `no sections ready to build` string:
  **(a)** a type with no builder or the wrong one — what this fix addresses; **(b)** a type mapped
  to a handler that *cannot fill a missing layout*, i.e. everything on bare `page-build-handler`,
  because `ensure_page_section_layout` exists only in `directory-build-handler`'s workflow.
  **(b) is bigger and better** — `blog-post`/`blog-index` casualties measured on four sites — and
  the right shape is making the layout-ensuring step reachable from the generic path, not routing
  more types to `directory-build-handler`. Its own submission, not an amendment to an approved one.

## The single most useful thing to know before touching anything

**A census of this class must join `pages.page_type`; NEVER filter `swi.spec->>'page_type'`** —
`reconcile_site_plan` writes no such key, so that filter returns a confident **zero** for the
population it exists to count. It did, here, and the zero reached a council submission. The
correct query is RUNBOOK §1. And **grep by the SYMPTOM, not the bug number**: this population is
named under four different numbers (206, 220, 328, closed 187) and `who-owns.py` cannot find it.

## Closure test — when may this move to `bugs_closed/`?

All three, at the artefact, not at a status:

1. A `reconcile_site_plan`-minted row for a typed page carrying `directory-build-handler`, with
   no hand routing (item 1 above).
2. That page built and serving, verified by `curl`, not by `build_status`.
3. The parked `entity-directory` and `entity-page` rows resolved to their *designed* outcomes
   (built / deferred gap respectively).

`section-index` pages staying parked is **not** a failure of closure — it is the council's
deliberate narrowing, and its cost is real and recorded: on a greenfield build the parked page is
the target of three dead links, one of them from the home page.
