# HANDOFF — nav membership (`bugs_open/149` Group A). COLD-START: read this first.

**Written 2026-07-31 ~18:10 UTC.** Everything below was measured; re-measure anything
you are about to act on, because this tree moves in minutes.

---

## 1. State in one paragraph

`bugs_open/149` A2 and A6 are **FIXED, council-APPROVED, LIVE and PROVEN**; A4's
recordable half is done; C2 is **answered** (not a defect). The bug file **stays open** —
it has **13 items** (not 12, as its own headers said until I corrected it): **6 done, A4
half, 6 open.** Nothing is in flight. No work is half-finished. A fresh session can pick
any open item cold.

**Verified live on `v1.0.1218`, both replicas, 18:07 UTC:**

```
A2  "declares no nav membership"              1
A6  "nav_membership_declared"                 1
    "navLabelSegmentFromURL"  (label fix)     2
    "classifyPagesForNav: skipping child page" 0   <- the string the fix REMOVED
    "CONTENT_CLAIMS_FLOOR_DETAIL"             1   <- positive control (149 C1)
```

## 2. What was wrong and what the fix is (the one paragraph worth keeping)

`check_orphan_pages` raises `nav_drift` for a page that declared nav membership and is
missing from nav, routing it to `nav-updater` — whose workflow is
`populate_nav_tables → render_site_components → create_rerender_items →
get_pages_for_rerender`, i.e. it derives nav, re-renders chrome AND propagates. The
handler had everything it needed. `classifyPagesForNav` then discarded the page with a
bare `continue` **before reading either nav flag**, because its URL sat under `/tools/`,
`/blog/`, `/guides/` … So the item completed, changed nothing, and the next sweep raised
it again.

**The rule now installed, and pinned by `nav_membership_test.go`:**
> `pages.in_header`/`in_footer` **DECLARES** nav membership. A page's URL shape may decide
> **WHERE** it appears. It may never decide **WHETHER** it appears.

Plus: `site_nav_items` had **two writers** and now has **one**. `addToolToNav` is deleted;
creators declare on the page row and file a rebuild **request** (`RequestNavRebuild`).

## 3. THE ONE THING STOPPING THIS BEING VISIBLE TO VISITORS

**The build dispatch loop is DEAD and it is not this lane's.** Measured 18:07 UTC:
newest claim anywhere `15:52:35` (**136 minutes ago**); gamesdesign alone has **34
`page_rerender` items stuck at `triaged`**; the served footer is unchanged.

So the chain is: nav tables ✅ → stored chrome ✅ → **served pages ⛔**.

- Handed to the `robot_hands_checker_gaps` lane (their Link 3), appended to their
  `NOTES_checker_gaps.md` with the measurements and the trap.
- **The trap, because it will fool you too:**
  `scheduled_tasks.last_triggered_at`/`last_completed_at` keep advancing while nothing
  runs — a fire-and-forget stamp. The real tells are `max(claimed_at)` fleet-wide and
  the newest `complete_idle` orchestration. See RUNBOOK **R12**.
- **Bypass for one site while it is down:** `./TRIGGER_nav_rebuild.sh <domain>` in this
  directory. It completed in ~30s with the loop dead, and refuses by default if a
  rebuild would fail to reproduce an existing nav row.

## 4. What a fresh session could do next, in order of value

1. **Prove the label fix on a second site** (the `guardian` seat asked for more than a
   one-site spot check before any fleet-wide dispatch). gamesdesign's labels were
   authored into `pages.nav_label` by hand, so it **does not exercise** the code fix. A
   site with flagged tool pages and NULL `nav_label` does — candidates from RUNBOOK R5:
   **gaswholesalers (4), finetuning (4), ai-agent-orchestration (4), vonc (3), oufe (2),
   fundamentallyai (2)**. Run `TRIGGER_nav_rebuild.sh <domain>`, then read the labels
   (RUNBOOK R8's second query). Expect clean names, no `/`, no `Index`.
2. **`bugs_open/149` A3** — nothing keeps a parent listing in sync, so a tool page with
   *no* nav flags is still invisible (correctly out of nav, absent from `/tools/index`,
   which nothing rebuilds). Genuine new build work; the analogue exists for blogs
   (`orphan_blog_posts` → `rerender-pages`).
3. **A5's non-child half** — `buildServicesHTML` still queries `pages` directly with its
   own predicate and `LIMIT 6`. Routing it through `GetNavItems` is the structural answer
   (`nav_tables.go`'s header lists the 8 query-time nav functions it replaced; this is a
   ninth that was missed) but it changes the footer "Our Services" column on **every**
   site — own measurement, own council round.
4. **A4's schema half** — `pages.in_header`/`in_footer` still `DEFAULT TRUE`, so any
   writer that omits them records no decision. Shared-schema change: architecture scope,
   and the fleet needs sweeping for rows correct only by inheriting `true`.
5. **A1** — recurrence branding (repeat detections born `unresolved`). Pair it with B2;
   `recurrenceExpected` (see §6) is the same mechanism from the writing side.
6. **B1/B2/B3 — DO NOT TAKE.** The `robot_hands_checker_gaps` lane owns them and was live
   in the tree today.

## 5. One loose end I created and did not close

**leopardess `/tools/password-entropy.html`** has an active `utility` nav row whose page
declares **neither** flag. No derivation reproduces it, so the next `populate_nav_tables`
run on that site deletes it and does not put it back. **This is pre-existing, not caused
by the fix** (it was equally unreproducible before), but the fix does not save it either.
Repair is one `UPDATE` setting `in_footer = true` — it is another lane's site, so I wrote
it down rather than reaching in. `TRIGGER_nav_rebuild.sh` refuses on this condition by
default, so it cannot be lost silently.

## 6. Landmines this work produced — read before touching the code

Both are in `LANDMINES.md` (synced to `doc_notes`), and the second bit me:

1. **`recurrenceExpected` is load-bearing on any repeated-`item_key` REQUEST.**
   `insertWorkItem` brands the **third** item on a repeated key as `unresolved` —
   terminal, never dispatched — and the emitter's return value is indistinguishable from
   ordinary coalescing. Without it, the third tool added to a site silently stops
   reaching the nav. That is `bugs_open/024`'s mechanism.
2. **A test asserting a query is NOT issued passes VACUOUSLY against `insertWorkItem`.**
   sqlmock errors on an unexpected query and `insertWorkItem` *swallows* that error
   (`if err == nil && terminalCount > 0`), so the branding never happens and every
   expectation is still met. My first guard test was green with the flag turned off.
   Assert the mechanism's **effect** (the INSERT's `status` argument) instead.

And the transferable one, now in the concept register (**NAV-013**) and the memory index:
**a change that widens what REACHES a function can break that function without editing
it.** Routing child pages into nav fed URL *paths* into `navSimplifyLabel`, which had only
ever seen flat page names — six live footer labels read `Tools/Damage Formula
Designer/Index`. **Invisible in the diff**, because nothing in the fix touches that
function; found only by reading the rendered rows. A blast-radius query counting ROWS
cannot see a LABEL defect.

## 7. Wrong calls logged (all in `WRONG_CALLS.md`, 2026-07-31)

1. **A 7-day aggregate answering a "right now" question.** I told the council the dispatch
   lane was healthy citing 1,580/1,664 claims over 7 days — true, and the lane had been
   dead two hours. The `bug_historian` seat was **more right than my rebuttal**.
2. **A guard test that passed while asserting nothing** (§6.2).
3. **A Go classifier replicated in SQL by listing its predicates, not walking its branch
   order** — `legalNames` is tested *before* any flag, so my first regression census
   reported a row as at-risk that was never at risk.
4. **A count inherited without checking** — "12 items" carried from a previous status
   block; the file has 13.
5. **(Not yet logged, minor, same family.)** At ~17:46 I found the pods on `v1.0.1216`,
   nine minutes after another session closed two bugs as "LIVE on v1.0.1217", and was one
   step from filing correction banners on their closed files. By 17:59 the fleet was on
   **`v1.0.1218`** and the whole concern was moot — a 13-minute window I nearly wrote up
   as a regression. **On this tree a cluster snapshot can go stale inside one
   investigation.** Re-read pod state immediately before asserting anything about it.

## 8. Artefacts

| what | where |
|---|---|
| commits | `1884f1ee8` (fix) · `8c41e3eaf` (objections answered) · `c053bb31f` (label regression) · `a8c21d233` (docs) · `5a35a20b2` (count) |
| council | `4486f1a9-6d96-4767-9ddd-6ff5e92ba45c` — **APPROVED**, 12 reviewers, 0 unreadable, 2 medium (both answered), no high |
| diagnosis loop | `1d8085f0-b596-4cce-9417-f48227ac67d3` — **CONFIRMED**, first iteration |
| concept register | **NAV-013**; NAV-008 superseded in part; NAV-006 narrowed |
| standing five | this directory (`PLAN` · `RUNBOOK` R1–R13 · `NOTES` · `README_where_we_are` · `SUMMARY_2026-07-31`) |
| bug file | `bugs_open/149_HANDOFF_2026-07-29_discovery_checker_layer_defect_queue.md` — banners on A2, A4, A5, A6, C2 |

**Why the ticket is still open:** six items remain, three of them another lane's live
work, one needing an owner ruling on a schema default. Closing it would delete the record
of six measured defects. That is a call for the owner, not for this thread.
