# HANDOFF — robot-hands.com — START HERE (2026-07-20)

**Supersedes `HANDOFF_2026-07-19_robot_hands_start_here.md`.** That file's R1–R6
status table still holds; its "Next actions" list is done or superseded — see
Corrections below, several of its premises were wrong.

Read order for a fresh chat: **this file → `NOTES_…` Turn 9 (technical log,
missteps) → `RUNBOOK_…` (the commands) → `README_where_we_are.md` (owner's log)**.

Site: **robot-hands.com**, site_id `00ff3af5-dad8-4770-9f70-3edc267a3c92`.
Deploy repo: **gqls/sites**, files under `robot-hands.com/` (GitHub-API commits,
no local checkout). DB:
`kubectl -n ai-persona-system exec -i postgres-clients-0 -- psql -U clients_user -d clients_db`

---

## UPDATE 2026-07-22 — R7 landed: the six-technology claim is now BACKED (read NOTES Turn 10)

The "42 prose fields" owner decision below (next action 4) is **resolved**. Owner chose
**expand the catalogue**. Five real, datasheet-sourced grippers added — one per missing
actuation technology (Festo DHPS-10-A pneumatic, OnRobot VG10 vacuum, OnRobot Soft Gripper
SG soft-robotic, OnRobot Gecko SP5 adhesive, Schmalz SGM-HP 50 magnetic). The `products`
index now holds **10 grippers across 6 technologies**, sourced, and the stat blocks were
re-pointed to computed subqueries over the catalogue (10 models / 6 manufacturers / 39
figures; `index`'s "6 technologies" now backed). Live in the DB **now**; the prose claim is
true immediately (it was always rendered, just unbacked). Only the two stat *numbers* on
about/gripper-detail await a re-render (`robot-hands-r7-catalogue-expand`), which is sitting
behind the bug-029 queue stall — until it drains those pages show 5, an undercount, not a
lie. Details: `SQL_2026-07-22_r7_…`, NOTES Turn 10, `bugs_open/043` (Update 2026-07-22),
PLAN D10. Commit `efa9dea1b`. **Residual (flagged, not part of the decision):** MatchMatrix
still scores only the parallel-jaw subset, so *tool-scope* prose remains ahead of the tool;
new grippers have no browsable catalog rows.

**VERIFIED LIVE after the v1.0.1149 roll (2026-07-22), and two follow-on fixes landed:**
- `about` renders 10/6; `/entities/gripper-detail.html` renders **10 / 6 / 4 / 39** bare.
  Note gripper-detail's real url is `/entities/…` (`pages.url`) — NOT root; a stale root
  file misled me briefly (WRONG_CALLS).
- **R7b→R7c:** gripper-detail still carried bug 043 point (b) — the placeholder unit
  suffixes ("10%", "6ms", "4+", "39x"). Root-caused to the SHARED `system-stats` component's
  `input_schema` `fallback`s (`%/ms/+/x`, `source=static`), which is **fleet-wide** (all 5
  consumers; vonc "14,203%", ai-agent-orch "1,000sms"). R7c set the fallbacks to `""`
  (zero live-page impact on the other 4 — their suffixes are persisted non-empty) and
  cleared+re-rendered robot-hands. VERIFIED bare live. Commits `569b160c3`, `642bb57cb`.
  See `bugs_open/043` Update 2026-07-22 (b), NOTES Turns 11–12, `SQL_…_r7b/r7c_…`.
- **Residual → 043 thread:** the other 4 `system-stats` sites still carry persisted junk
  suffixes; the schema fix does not retro-clear them and the intended unit is each owner's
  call. Not touched from this lane.
A SUMMARY is now warranted (R7 closed end-to-end).

## STATE AT 2026-07-21 (v1.0.1144 roll) — read this first

**Everything R4 is still live and correct after the roll.** Re-verified against the
rendered pages, not statuses, 2026-07-21 ~10:00 BST: `/`,
`/tools/matchmatrix/index.html`, `/about.html`, `/entities/gripper-detail.html` all
**200**; **0** fabricated figures (`1,200+` / `2,400+` / `140+`); **0** links to the
dead payload-budget tool on the homepage. robot-hands has **0** pending build work —
both R4 batches (13 items) are complete. **There is nothing outstanding on this
workstream that needs the cluster.**

**The roll halted builds fleet-wide again (`/bugs_open/029`, third roll running).**
Not this site's problem — robot-hands has nothing queued — but if you dispatch
anything, know that ~37 build items were backed up with ~1 completing per 20 min,
newest trigger stuck at `call_dispatch`. I applied 029's cancel recovery; under
v1.0.1144 it buys **one** dispatch then re-stalls, so it is relief not repair. Do
**not** loop it. Owned by the 029 thread; the reaper fix is still pending.

**`/bugs_open/052` gained a correction from another thread** (the bugfix-049
session): the never-built predicate is **`deployed_at IS NULL`**, not
`build_status='planned'` — four `needs_rebuild` pages have never deployed and 404,
including `gaswholesalers`'s framework page linked from 28 live footers. Fix
candidate 1 in 052 inherits the gap; use `deployed_at IS NULL`. See next action 6.

**`/bugs_open/023` is progressing (still owned, still observe-only in the pod).**
Migration `181` gated the live CTA anchors (class C+E) and `/bugs_open/049` was filed
for 312 live 404s from stale chrome — same derived-field class. The resolver flip is
**not** live in v1.0.1144 (only the observe stage from 1140), so robot-hands' primary
CTAs still revert to nav-order defaults on render; the ones that read correctly do so
because the resolver would pick that URL anyway. Do not start a competing fix
(`scripts/who-owns.py 023` shows the owner).

**Tooling built alongside this workstream** (committed, fleet-wide, documented in
CLAUDE.md): `scripts/who-owns.py` (does a bug already have an owning thread — run it
before routing work at any bug) and `scripts/memory-git-snapshot.py` (auto-versions
the shared auto-memory dir). Both came out of missteps logged in
`docs024_key_docs_latest/WRONG_CALLS.md`.

---

## Status of the original six defects

| # | Defect | State (2026-07-20) |
|---|---|---|
| R1 | Dark theme lost, blue brochure chrome | **DONE, verified live** |
| R2 | learning-center-hub yellow on white | **DONE, verified live** |
| R3 | learning-center URL sprawl | **DONE, verified live** |
| R4 | Tools broken / MatchMatrix "blank" | **DONE** — MatchMatrix built & live; CTA layer repaired; payload-budget card removed |
| R5 | Dead "Load More" | **DONE, verified live** |
| R6 | 6 of 9 listed articles 404 | **DONE** |

**R4 is closed.** `/tools/matchmatrix/index.html` serves 200 with a real
interactive tool (`gqls/sites` @ `0a6dc426`, 38,144 B, 1 form / 4 inputs /
4 selects / 4 `addEventListener`, 19/19 logic tests). The other dead tool,
`tool-robot-payload-budget-calculator`, was **not** built — its card was removed
from the homepage tool list per owner ruling, and its single inbound CTA repointed.

---

## Corrections to the 2026-07-19 handoff (do not re-walk these)

- **"Five homepage CTAs point at those two 404s" undercounted by 4×.** It was
  **20 components across 11 pages**, plus **20 more mispaired secondary CTAs that
  never 404'd at all**. The dead URL had become a default dumping ground.
- **"Repoint those CTAs at `/matchmatrix.html`" would have been wrong twice over.**
  (a) Only ~6 of the 20 labels named MatchMatrix; a URL-keyed repoint cements every
  mismatch. (b) `/matchmatrix.html` is a **brochure** — 0 forms, 0 inputs, 1 button
  (the mobile nav toggle) — whose own two primary CTAs pointed at the same 404. It
  would have moved the dead end one hop.
- **`bugs_open/039` does not gate a tool page**, despite looking like it should.
  Tool pages here carry `sections = []` and have **zero** `site_plan_sections`
  rows *even when they work*. 039's Part 1 naming rule is still worth knowing; it
  just is not the trap here.
- **`017…unregistered_action…` is now CLOSED and lives in `/bugs_closed/`.**
  This handoff first reported it as "reads fixed-and-live in v1.0.1139, own file
  says otherwise, left for its owning thread" — they have since closed it
  (`93edb02f7`, pod-verified). **Number 017 is still ambiguous**: the *other* 017
  (`static_cutover_orphans_backend_entry_forms`) remains open. Resolve by slug.
- **The image rolled to v1.0.1139** (pod started 2026-07-20 07:35Z), so the
  previous handoff's optional item 4 (`5151d4a79`) is moot.

## The reference implementation for a tool page (verified, use this)

A working tool page on this site is **one** `page_components` row at position 2
pointing at a bespoke `content_components` row:

```
name          tool-<slug>-robot-hands-com
function      tool-<slug>
render_mode   template
html_template the ENTIRE interactive tool, 17–24 KB (form + inputs + JS)
```

`pages.sections` stays `[]`. There are no `site_plan_sections`. The deployed
artefact lives at `gqls/sites:robot-hands.com/tools/<slug>/index.html`.

**Do not route a new tool through `tool-generator`.** Live prompt comparison:

```
tool-recreation-handler | has_no_fake_data_rule=t | forbids_fetch=f
tool-generator          | has_no_fake_data_rule=f | forbids_fetch=t
```

`tool-generator` — the path for a *new* tool, not the recreation handler the
owner's `020` hold names — has **no fake-data rule at all** and forbids fetch. A
data-backed tool routed through it is structurally pushed to invent its dataset.
MatchMatrix was hand-authored for exactly this reason.

---

> **CORRECTED 2026-07-20, same session — read this before acting on R4 below.**
> The CTA repair is **partly not durable**, and I found out by watching a page
> re-render rather than by trusting my own verify query.
> `resolve_internal_links_action.go` **owns** the CTA url fields
> (`ctaFieldNames`, `:99-105`) and recomputes them on every render via
> `chooseCTATargets` (`:319`), which **never reads the label** — it sorts
> interactive pages then hubs by `NavOrder`/`Name` and takes `[0]` and `[1]`.
> So the "dumping ground" was deterministic assignment, not drift, and
> `content_data` is not the source of truth for a CTA URL.
> **Durable:** the statistics (`stat_*` is not resolver-owned; `about.html`
> renders 5/5/4 live) and the tool page (a committed artefact).
> **Not durable:** the primary CTA URL edits. The ones that still look right do
> so because the resolver would choose that URL anyway — *building the tool* is
> what fixed them, not the SQL. `/contact.html` is permanently excluded as a CTA
> destination (`areasExcludedFromCTA`, `:72-74`), so "Request Integration
> Support" cannot be paired correctly from `content_data` at all.
> Full write-up in `NOTES` and in the addendum on `/bugs_open/023`.

## Next actions, in order

1. **DONE — the re-render batch drained and is verified live.** 12/12 complete,
   plus a 13th (`robot-hands-r4e-archive`) for the tool list. Verified against the
   rendered pages, not the statuses, at 2026-07-20 19:30 BST:
   `/`, `/about.html`, `/entities/gripper-detail.html`,
   `/tools/matchmatrix/index.html`, `/matchmatrix.html`, `/services.html` all
   **200**; **zero** occurrences of `1,200+` / `2,400+` / `140+` /
   "Actuation Technologies" anywhere; **zero** links to the dead
   `robot-payload-budget-calculator` across five pages. `gripper-detail` renders
   5 / 5 / 4 / 24 with the `2,400+%` and `140+ms` placeholder suffixes gone.
   Nothing outstanding on this item.

2. **`bugs_open/023` — DO NOT START IT HERE. It has an active owning thread.**
   The `cta_link_integrity` workstream
   (`../cta_link_integrity/`) is six council rounds in and its **observe-only
   stage shipped LIVE in v1.0.1140** (`f6b4aea5a`, trail `2525f980`). Its PLAN
   already carries the defect classes, including class H — `ctaFieldNames` is a
   hardcoded 6-component map and anything outside it is "detectable but not
   repairable". **Coordinate, do not duplicate**; that thread owns the fix.
   What robot-hands contributes is an instance and one mechanism I did not find
   in their notes, now written into `/bugs_open/023`'s addendum: `chooseCTATargets`
   (`resolve_internal_links_action.go:319`) picks `primary = ordered[0]`,
   `secondary = ordered[1]` after sorting by `NavOrder`/`Name` and **never reads
   the label**, so every CTA of a kind on a site converges on the *same* two
   destinations. That is why 20 components here collapsed onto one URL. Their
   framing is about which components are *repairable*; this is about the *choice*
   being label-blind. Hand it to them rather than acting on it.
   Consequence for this site meanwhile: robot-hands' primary CTAs will keep
   reverting to nav-order defaults until that thread lands its fix. The ones that
   currently read correctly do so because the resolver would choose that URL
   anyway.
   **Moved on since:** that thread has rescoped 023 and split the library gap out
   as `/bugs_open/045` — the only tool-hero component in the library is hardwired
   to a Bayesian ranker — so 023 is no longer gated on it (`0bd6aa09d`).
3. **`bugs_open/043` — generated copy invents quantitative claims.** Filed this
   turn, robot-hands contained. **The fleet-wide sweep has NOT been run** — that
   is the first thing the fixing thread should do (sweep query is in the file).
   Same family as `020`, different path, so a site with no tools is still exposed.
4. **Owner decision: the 42 remaining prose fields.** The unsupported
   "six actuation technologies" claim survives in 42 further `content_data` fields
   on this site (body prose, `features`, `subheadline`, FAQ `questions`, `cards`).
   Correcting a statistic was containment; rewriting 42 paragraphs is a decision
   about what the site claims to be. Query to list them is in `043`.
5. **DONE by another thread — `bugs_open/022` is CLOSED.** The scheme guard
   went live in v1.0.1140, verified against the pod and a real run
   (`7812d4664`); the file has moved to `/bugs_closed/`. The previous handoff
   listed it as open and so did this one, until `scripts/who-owns.py` surfaced
   the closure. Nothing to do.

6. **`bugs_open/052` — filed this turn, and the one build-workflow change worth
   arguing for.** A listing regenerates from the page set with no build-state
   filter, so it re-advertises never-built pages; it undid the dead card removal
   on this site's homepage. **Another thread (bugfix-049) has since corrected the
   predicate** — see 052's addendum: use **`deployed_at IS NULL`**, NOT
   `build_status='planned'`. Four `needs_rebuild` pages have never deployed and
   404 (one is `gaswholesalers`'s framework page, linked from 28 live footers), so
   the `planned`-only predicate I originally proposed misses the worst real
   instance. The `needs_rebuild`-usually-serves trap still holds; the fix is just
   to key on `deployed_at`, not the flag. Related: `/bugs_open/049` (chrome 404s)
   and `/bugs_open/053` (legal nav) are the same derivation of "is this page
   fetchable" — coordinate, do not solve three times.

7. **Optional: `tool-robot-payload-budget-calculator`.** Still `planned`, no page,
   card removed, CTA repointed — so nothing is user-visibly broken. If it is ever
   built, it is the *safer* of the two (formula-based, not data-backed) and the
   reference implementation above is the route.

## Landmines specific to this site

- **Repoint CTAs by LABEL, never by old URL.** The single most important line in
  `SQL_2026-07-20_r4_matchmatrix_and_cta_pairing.sql`.
- **CTA fields are not consistently named.** Most use `cta_text`/`cta_url`; some
  use `primary_cta`/`primary_cta_url`; `content-block-about` uses **`cta_label`**
  and was missed by the first pass because of it. Stat blocks are worse —
  `stat_1_value` on two components, `stat1_value` (no underscore) on a third.
  A sweep keyed on one spelling will silently under-report.
- **`build_status` on tool components is meaningless** — all three working tool
  components sit at `pending` with `deploy_commit` NULL since April while serving
  200. `bugs_open/024` leg (a), confirmed live.
- **Check the rendered page, not just `content_data`.** The `2,400+%` / `140+ms`
  placeholder-suffix bug is invisible in the DB row.
- **Backlog is inherited, not new damage** — ~115 unresolved, 53
  needs_human_review. Almost all predates this workstream. `bugs_open/033`: that
  queue has no consumer, so anything failing into it is invisible.

## Artefacts from this turn

| file | what |
|---|---|
| `SQL_2026-07-20_r4_matchmatrix_and_cta_pairing.sql` | CTA label↔URL pairing, tool-card removal, overstated copy on touched components |
| `SQL_2026-07-20_r4b_fabricated_about_stats.sql` | the `about` stat block (1,200+ → 5) |
| `SQL_2026-07-20_r4c_fabricated_stats_sitewide.sql` | `gripper-detail` + `index` stat blocks |
| `SQL_2026-07-20_r4d_queue_rerenders.sql` | the 12 re-render items |
| `/bugs_open/043_…invents_quantitative_claims.md` | the platform bug |
| `WRONG_CALLS.md` (2026-07-20 robot-hands row) | the 11 kg / 140 N misread |

All four SQL files are **applied**. Re-render queued, **not yet drained**.

## Process notes

- The standing five are current: PLAN, RUNBOOK, NOTES (Turn 9), README (owner's
  log, appended), and the SUMMARY series. **No new SUMMARY was written this turn**
  — per CLAUDE.md's 2026-07-20 cadence cut, a summary is for a real milestone, and
  `SUMMARY_2026-07-19` plus this handoff already say where we are. Write one when
  the re-render lands and R4 can be reported closed end-to-end.
- `bugs_open/022` and `bugs_open/017…unregistered_action…` were authored before
  the 090 diagnosis-loop default came in and have not been through it. If either
  becomes load-bearing for someone else — especially `022`, whose fix changes
  behaviour fleet-wide — file it to 090 first.
