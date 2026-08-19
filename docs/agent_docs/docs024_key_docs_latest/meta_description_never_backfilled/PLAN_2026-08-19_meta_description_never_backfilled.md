# PLAN 2026-08-19 — `pages.meta_description` has writers but no BACKFILL producer

## How this lane started (it is not the lane I was given)

I was handed `bugs_open/309`'s **case repair**: the fundamentallyai.com Platform Log
index renders six cards with zero anchors. The config half is done, applied and live
(migration `478`). §10 of that bug says the one remaining step is:

> **Backfill `meta_description` on five pages … By the framework, not by hand (owner
> ruling 2026-08-06). This is the whole blocker.**

and it routes the reason those five are empty to `head_essentials_missing` /
`bugs_open/083` ("detected findings never reach a handler").

**That routing is wrong, and the size is wrong.** Both corrections are measured below.
The five pages are ordinary members of a fleet-wide class, and the framework has no
mechanism that can fill any of them.

## Two corrections to `bugs_open/309` §10

1. **`head_essentials_missing` does not check meta descriptions.** It asserts exactly
   three things — a non-empty `<title>`, a skip-link, and a `<footer>`
   (`check_site_structural_validity.go:991-1044`, and `headEssentials()` is nine lines
   long). It would never have fired for a missing description, so its 606 handler-less
   rows are not why these five are empty, and fixing `083` would not fill them.
2. **It is not five pages.** `[MEASURED 2026-08-19]` **407 of 731 active pages (55.7%)
   across 26 of 27 sites** carry an empty `meta_description`. Three sites are at 100%
   (`loancalculator.co.uk` 43/43, `adversecreditmortgage.co.uk` 19/19, `loanzy.uk`
   13/13).

## The mechanism, as measured

`pages.meta_description` has **seven writers** and every one of them is a
**create-or-upsert** path — the column is populated at page birth or not at all:

| writer | how it supplies the value |
|---|---|
| `site_db_actions.go:1131,1163` | from the LLM-written site plan (`GetStringField`) — empty if the plan omitted it |
| `create_tool_component_action.go:302,352,504` | `composedToolMetaDescription` / `composedGuideMetaDescription` (Go templates) |
| `deploy_tool_action.go:395,428,589` | same two composers |
| `create_report_page_action.go:172` | a hardcoded literal |
| `apply_adoption_plan_action.go:517,549` | from the adoption plan |
| `adopt_verbatim.go:452,474` | `extractHTMLMetaDescription` off the source page's own HTML |
| `cmd/webdesignport/import.go:178` | the import tool |

**No UPDATE path anywhere sets it on a page that already exists.** Nor is it a field a
writer agent can reach: `llm_fields` / `llm_field_specs` are **per-section**
(`plan_sections_action.go:864-897`), scoped to a component's `input_schema`, and
`meta_description` is a **`pages` column**. This is precisely the shape of the
bugfix-203 lesson — *a precedent transfers on mechanism, not on wording; ask who owns
the field before asking the framework for it.*

And **nothing detects it**: of the 58 checks in
`platform/orchestration/actions/discovery_checks/`, not one names the column.

## The live proof that the obvious route does not work

`content_rewrite` items **are** filed about missing meta descriptions and routed to
`page-build-handler`. Two of them are `complete`. Neither wrote the column:

| item | site / page | status | `meta_description` today |
|---|---|---|---|
| `ec701bb3-85e7-40e7-bff1-2ce1ae104861` — *"neither page carries a meta description"* | gaswholesalers.com `about` | **complete** 2026-08-15 16:30Z | **0 chars** |
| `13522562-2392-4db9-96b5-204ab67cb999` — *"the home page meta description … leads with a self-description of inventory"* | webdesign.co.uk `index` | **complete** 2026-08-15 19:15Z | **0 chars** |

The gaswholesalers page was itself updated at 19:46Z that day, so the handler ran and
touched the page — and the column stayed empty. **A `complete` work item is not a
repaired artefact.** This is the measured version of the 203 trap: filing a
`content_rewrite` for a meta description is a request the framework cannot satisfy, and
it reports success.

## Where the boundary sits, and why I stopped at it

The owner's 2026-08-06 ruling is explicit — *"if the generator does not exist yet, that
is the finding to report, not a gap to fill personally."* So:

- I will **not** hand-write the five descriptions. That is the prohibited move, and it
  would also hide a 407-page defect behind one unblocked page.
- I will **not** file `content_rewrite` items for them. Measured above: it completes and
  does nothing.
- I will **not** lower `section_shrink_floor` (309 §9 — step-config only, therefore
  fleet-wide, on the highest-volume pipeline there is).

## Phasing

1. **Verify the mechanism through the diagnosis loop before asserting it.** Cross-cutting
   structural claim ⇒ owner ruling 2026-07-31 applies. Run correlation
   `d7a9ab39-4917-4479-b1d5-68aae982f79c` (intake `e674f39b-…`, which is NOT the
   artefact key). — *in flight*
2. **File the class as a bug** in `bugs_open/` with the census and the code map.
3. **Correct `bugs_open/309` §10** in place, visibly, so the next reader does not spend
   the day I spent re-deriving that `head_essentials_missing` is the wrong detector.
4. **Present the fix options to the owner.** Building a producer for a public,
   outward-facing content field on 407 pages across 26 sites is not a session's
   unilateral call — it is new authority on a shared seam (2026-08-02 §2) and it writes
   copy that appears under the owner's name.

## Decisions and their reasons

- **A new lane rather than 309's or 284's.** 284 is closed; 309's own lane owns the
  phantom-source-component class. This is a third, distinct class that happens to block
  309's case repair. Keeping it separate is what stops 309 growing a second account of
  a defect that is not its subject.
- **Detection alone is not a candidate fix.** A `meta_description_missing` check with no
  handler reproduces exactly the state this estate is already in — 606 detected
  `head_essentials_missing` rows nobody can action. Ranked accordingly when options go
  to the owner.


---

# PHASING — UPDATED 2026-08-19 after the owner's ruling

**Owner ruled: option (c), everything including a backfill producer.** Phases 1-4 of the
original plan are done; the plan below is what that turned into.

| # | phase | state |
|---|---|---|
| 1 | Verify through the diagnosis loop before asserting | **DONE** — run `7375631a-…` returned **UNVERIFIABLE** on its iteration cap. Not a confirmation, and not written up as one. Its value was refusing the "frozen at creation" framing, supplying M2, and naming the `site_snapshots` test that settled M2's occurrence. |
| 2 | File the class | **DONE** — `bugs_open/320`, committed `4cd81e211` |
| 3 | Correct `309` §10 in place | **DONE** — same commit, purely additive (37 added / 0 deleted, checked with `--numstat` per the deleted-bullet landmine) |
| 4 | Put the options to the owner | **DONE** 2026-08-19 |
| 5 | M2 guard — `upsertPage` ON CONFLICT | **DONE**, committed `aeccfc595`. **Rides the next fleet roll** |
| 6 | M1 fix — migration `485`, the planner is asked | **DONE + APPLIED + ledger-recorded.** Config, live on apply |
| 7 | `save_page_meta_description` (SEO-004) | **DONE**, committed `aeccfc595`. Rides the roll |
| 8 | The driver — migration `486` | **WRITTEN, HELD.** `_HOLD` suffix; runner lists it under Sidecars, verified with `--no-probe` |
| 9 | Council gate | **SUBMITTED** `46734ae9-…`; committed with `Council-Submitted:` per the pre-verdict rule. **Still owed: read the verdict and act on a REVISE/REJECTED** — the code is already on the shared branch |

## The decisions taken while building, and why

- **Only the PERSIST half is new.** `query_database` already finds pages;
  `execute_llm_prompt` already writes sentences. A monolithic Go action would have
  re-implemented both and taken authorship from the framework, which is what the
  2026-08-06 ruling exists to prevent.
- **`overwrite_existing` defaults FALSE** (owner ruling 2026-08-02 §2). The unsafe
  authority is replacing published copy, so that is the side that must be asked for. The
  workflow omits the key entirely, and `486`'s verify block RAISEs if a later edit adds it.
- **The policy lives in the UPDATE's `WHERE`**, not in a read-then-write, so it cannot
  race a concurrent writer.
- **Reuse `MetaDescriptionLooksInternal`** rather than re-derive the brief-vs-public rule.
  Its own doc comment says it is exported so "a discovery check or a backfill" can ask the
  same question; this is that backfill.
- **`486` is NOT scheduled.** Nothing dispatches it. A generator of public copy on a timer,
  before anyone has read its first page, is how `bugs_closed/103` published 1,206 characters
  of build brief to Google. Hand-driven, one site at a time, until the output has been read.
- **A discovery check was ranked BELOW the fixes and not built.** Detection with no handler
  reproduces the state the estate is already in — 606 `head_essentials_missing` rows
  detected and unactionable. If one is added later it is a wider seam and gets its own
  register entry, which is written into the ratchet line's RE-CHECK condition.

## What is NOT done, stated so it cannot read as finished

1. **No page has been filled yet.** 407 of 731 remain empty. Phases 5-8 stop the cause and
   supply the tool; they do not do the work.
2. **`486` cannot be applied until the roll.** `[MEASURED 2026-08-19]` live chassis
   `v1.0.1314` / `d3590ca4…`, `merge-base --is-ancestor aeccfc595 d3590ca4` FALSE, 145
   commits unshipped.
3. **`save_page_meta_description` is UNEXERCISED.** Registered SEO-004 as *"built, not yet
   exercised in production"* deliberately — its tests pass, and a capability with no live
   call has an untested dependency on its environment.
4. **The council verdict is unread.**
5. **`bugs_open/309` is still blocked.** Its five pages are plan-managed, so they are
   reachable once the roll lands and fundamentallyai.com is replanned — or directly by
   `486`. Neither has happened.
6. **M2's scale is unknown and should stay marked so.** Four pages proven blanked, out of a
   139-page / 7-site / mostly-April sample. Existence, not a rate.
