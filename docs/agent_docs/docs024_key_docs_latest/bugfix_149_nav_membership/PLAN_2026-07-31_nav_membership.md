# PLAN 2026-07-31 — nav membership: one declaration, one writer

Workstream for `bugs_open/149` **Group A** (routing). Items in scope: **A2, A6,
A5's child-page half, A4's recordable half**. Deliberately out of scope: A1/B2/B3
(dispatch + cadence — owned live by the `robot_hands_checker_gaps` lane as of
2026-07-31 09:17), A3 (a new route + a listing rebuild — a different mechanism),
C2 (`report-builder`'s `check_claims:false` — a question, not a defect).

## The defect in one sentence

`nav_drift` is raised for a page that declared nav membership and is missing from
nav; its handler `nav-updater` rebuilds the nav tables from `pages`; and the
rebuild **discards the declaration** for any page whose URL sits under `/tools/`,
`/blog/`, `/guides/`, `/articles/`, `/case-studies/`, `/news/`, `/resources/` or
`/insights/`. So the item completes, nothing changes, and the next discovery run
raises it again.

## The evidence, with its positive control

Both from the live DB, 2026-07-31 (`RUNBOOK` R1/R2 re-run them):

**The defect.** `site_work_items` id-shape `nav_drift`, source `discovery`,
gamesdesign.co.uk, raised 2026-07-29, `spec.affected_pages` naming exactly four
pages, **status `complete` at 17:27:50 on 07-29**:

| page | url | in_header | in_footer | in `site_nav_items` on 07-31 |
|---|---|---|---|---|
| bayesian-ranking | /tools/bayesian-ranking.html | t | f | **no** |
| tool-drop-rate-tuner | /tools/tool-drop-rate-tuner.html | t | t | **no** |
| tool-loot-table-balancer | /tools/tool-loot-table-balancer.html | t | t | **no** |
| tool-xp-curve-designer | /tools/tool-xp-curve-designer.html | t | t | **no** |

Same shape on ai-agent-orchestration.com: `tool-ai-agent-roi-estimator`, item
complete 2026-07-25, still absent six days later.

**The control, which is what makes this a mechanism and not a cadence story.**
robot-hands.com, `nav_drift` raised 2026-07-28, `affected_pages` =
`["learning-center", "news"]` — `/learning-center.html` and `/news.html`, neither
under a child prefix. Both are **in `site_nav_items` today**. Same check, same
handler, same action. The only difference between repaired and not-repaired is the
URL prefix.

So: the handler works. One predicate stops it having anything to do.

## Why it happens — `classifyPagesForNav`, and the order of two rules

`populate_nav_tables_action.go` already holds the right rule and never reaches it:

```go
if isChildPageURL(page.URL) && !isSectionIndexType(page.PageType) {
        continue                      // ← drops the page ENTIRELY, flags unread
}
if neverPrimaryTypes[page.PageType] { // blog-post, tool, entity-page
        if page.InFooter || page.InHeader {
                utility = append(utility, page)   // ← the rule that should apply
        }
        continue
}
```

The platform has **already decided** what to do with a page that must not appear
in the main menu but has declared nav membership: it goes to `utility`, which
renders in the footer. Two overlapping notions of "never primary" exist — one
keyed on URL shape, one on `page_type` — and the URL-shaped one runs first and
answers a different question (*whether* the page appears, not *where*).

`LANDMINES.md` records the same mechanism from the other end, found independently
by the leopardess lane on 2026-07-30: `populate_nav_tables` does
`DELETE FROM site_nav_items WHERE site_id = $1` and rebuilds, so running the agent
whose name says navigation **deletes every child-path nav link and puts none back**.

## The design: one declaration, one writer

> `pages.in_header` / `in_footer` is the declaration of nav membership. A page's
> URL shape may decide **where** it appears. It may never decide **whether** it
> appears.

Corollary, and the reason this is not just a predicate patch: `site_nav_items` is
a **derived** table with **two** writers today — `populate_nav_tables` (derives
from `pages`, authoritative, full DELETE+rebuild) and `addToolToNav`
(`create_tool_component_action.go`, hand-written, incremental, into a bespoke
`tools` group typed `primary`). The authoritative writer cannot express what the
incremental one writes, so it destroys it. Making two writers agree is a
coordination invariant, and this file is the record of that invariant drifting.
**So: one writer.** Creators declare intent on the page row and *request* the
rebuild; the derivation writes every row.

### Edit 1 — `classifyPagesForNav` (A2, A5's child half, the landmine)

Collapse the two never-primary notions into one, so a child-path page falls
through to the rule that already exists:

```go
neverPrimary := neverPrimaryTypes[page.PageType] ||
        (isChildPageURL(page.URL) && !isSectionIndexType(page.PageType))
if neverPrimary {
        if page.InHeader || page.InFooter { utility = append(utility, page) } else { /* logged skip */ }
        continue
}
```

### Edit 2 — `create_tool_component_action.go` (A6 half 1, A4's recordable half)

- Its page INSERT omits `in_header`/`in_footer`, so both inherit the **schema
  default `true`** while the action has computed `inHeader`/`inFooter` from its own
  config and uses them only to decide whether to call `addToolToNav`. A step
  configured `in_footer: false` therefore produces a row saying `in_footer = true`.
  **The writer discards its own decision.** Record both columns explicitly.
  This is a prerequisite, not a tidy-up: after Edit 1 that inherited `true` becomes
  load-bearing — it is what puts the page in the footer.
- Delete `addToolToNav`; request the rebuild instead (below).

### Edit 3 — `deploy_tool_action.go` (A6 half 2)

It records the flags and writes no nav row and no request, so a deployed tool waits
for a discovery sweep to notice it. Request the rebuild.

### The rebuild request, and why it is not a nav row

Writing the nav row at creation looks like the stronger fix ("unrepresentable beats
detectable", 149's own ordering) and **is worse on this platform**: chrome is a
stored artefact (`bugs_open/117`/`118`), so a nav row alone changes no served page,
while `check_orphan_pages` treats the presence of a nav row as reachability. A
creation-time nav row would therefore leave the page exactly as unreachable and
**silence the only check that would have noticed** — my own write blinding my own
detector.

What actually makes the link real is `nav-updater`, whose live workflow is
`populate_nav_tables → render_site_components → create_rerender_items →
get_pages_for_rerender` — derive, re-render chrome, propagate to deployed pages.
So the creators emit one work item asking for exactly that, through the existing
`withWorkItemTx` → `insertWorkItem` path, `handler_agent = nav-updater`,
`status = triaged` (matching the `needs_content_page` item the same actions already
emit, which is dispatchable without waiting on triage).

Two details that are load-bearing:

- **`recurrenceExpected: true`.** `insertWorkItem`'s two-strike rule brands a third
  item with a repeated `item_key` as `unresolved` — terminal, never dispatched.
  That is right for a detected defect and wrong for an action request, where a
  completed predecessor means success. The flag exists for precisely this
  (`bugs_open/024`). Without it, the third tool added to a site would silently stop
  reaching the nav. This is also 149 A1's "born `unresolved`" mechanism seen from
  the writing side.
- **A distinct `item_key`** (`nav_rebuild:<site_id>`, not the detector's
  `nav_drift:<site_id>`). Sharing the key would make my request inherit the
  detector's strike history and would blur a real signal — a `nav_drift` that keeps
  recurring means the repair is not working, and that is worth keeping legible.

### Edit 4 — tests

`nav_membership_test.go` pins the invariant directly: a nav-flagged child page
reaches `utility` and never `primary`; an unflagged child page stays out; a
`section-index` child page still reaches `primary`; and a `tool`-typed page with a
flag still reaches `utility`. The first test fails if the URL-shaped rule ever
regains the power to drop a flagged page.

## Blast radius, measured before submission (not left for the reviewer)

Fleet-wide, 2026-07-31, over all 14 sites (`RUNBOOK` R3/R4/R5):

- **Rows a rebuild destroys today and would keep after the fix: 6 of 7.** Seven
  active nav rows are not reproducible by today's derivation (2 in `tools`/primary
  from `addToolToNav`, 5 hand-written into leopardess's `utility` group). After the
  fix, six are reproduced. The seventh — leopardess `/tools/password-entropy.html`,
  `in_header=f in_footer=f` — is not reproducible either way: the row asserts a
  membership the page does not declare. **That is a pre-existing exposure, not a
  regression introduced here**, and the honest repair is to set the flag. Told to
  that lane rather than changed here.
- **Rows added on the next rebuild: 26, across 9 sites, at most 5 per site.**
  21 are deployed pages whose link ships on the next chrome render; 5 are
  `needs_rebuild`/`planned` and are held back by the **existing**
  `NavFetchableOnly` filter until they deploy — no new mechanism needed for that.
- **Footer sizes stay in family.** Utility groups already run to 14 items
  (leopardess, live). Post-fix maxima: leopardess 15, finetuning 15,
  ai-agent-orchestration 14, gaswholesalers 13, gamesdesign 6. Every live caller
  passes `maxItems = 0`, so nothing is truncated.
- **No live site changes as a direct result of this commit.** Nav rows are read by
  chrome renders, and chrome is a stored artefact. The two `tools`/primary rows
  that exist today are **not in any served header** — checked on gamesdesign and
  fundamentallyai — because chrome has not been re-rendered since they were
  written. The change alters what the *next* re-render produces.
- **Placement change to declare:** a tool page created through
  `create_tool_component` used to land in a `tools` group typed `primary`, i.e. the
  **header**. After this it lands in `utility`, i.e. the footer. That is the
  platform's own stated contract — tier 4, "never primary: individual tool pages" —
  and the two affected rows have never reached a served header, so nothing visible
  moves.

## Consumers to tell (owner ruling 2026-07-29)

`site_nav_items` is shared. What changes about their guarantee:

- **`nav-updater` / `populate_nav_tables` callers** — a rebuild now *keeps*
  child-path items that declare a flag, where before it deleted them. This makes
  the landmine's "what to use instead" workaround (`nav-link-fixer` +
  assemble-mode rerender) unnecessary for the flagged case; that landmine must be
  amended, not deleted, because the unflagged case still loses rows.
- **`check_orphan_pages` (`nav_drift` branch)** — its items become fixable. Expect
  the `nav_drift` recurrence on gamesdesign and ai-agent-orchestration to stop.
- **Chrome renderers** (`render_site_components`, `section_editor_actions`,
  `v3_site_actions`) — footer nav gains up to 5 items per site on the next render.
- **`create_tool_component` / `deploy_tool_to_site` callers** — a tool creation now
  emits one extra work item (`nav_rebuild:<site_id>`, handler `nav-updater`).

## What this does NOT fix, said plainly

- **A3 stands.** A tool page with *no* nav flags is still invisible: skipped by the
  derivation (correctly — it declared nothing), excluded from `check_orphan_pages`
  (`page_type='tool'` without flags), and absent from the parent listing, which
  nothing keeps in sync. That is a different mechanism and wants its own item.
- **A4's schema half stands.** `in_header`/`in_footer` still DEFAULT TRUE at the
  column. Edit 2 stops one writer inheriting it silently; it does not change the
  default, which is a shared-schema change and architecture scope.
- **A5's non-child half stands.** `buildServicesHTML` still queries `pages`
  directly with its own predicate for non-child pages. Routing it through
  `GetNavItems` would change the footer's "Our Services" column on every site and
  wants its own measurement.
- **Cadence.** None of this runs until a discovery sweep or a nav rebuild happens.
  That is the other lane's work (263 items `detected`, 0 `triaged`).

## Decision log

- **2026-07-31 — no diagnosis-loop run filed, and why.** The cause is not in doubt:
  the code path, a completed work item naming four pages, those four pages still
  absent two days later, and a positive control where the same handler succeeded on
  a non-child page. The structural claim is also not new — it is 149 A2, filed
  2026-07-29, and `LANDMINES.md`'s nav-updater entry, found independently by another
  lane on 2026-07-30. A run would be spent confirming what two sources already
  measured. The council gate is the right instrument here, and it is being used.
- **2026-07-31 — took Group A, not B.** `robot_hands_checker_gaps` was live in the
  tree at 09:17 the same morning on schedule/dispatch (B1/B2/B3). Overlapping would
  have competed, which `scripts/who-owns.py` exists to prevent.
