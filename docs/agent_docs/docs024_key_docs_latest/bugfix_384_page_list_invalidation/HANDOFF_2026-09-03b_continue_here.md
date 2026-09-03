# HANDOFF — bugs_open/384 page-list invalidation · continue here

**Written 2026-09-03 ~13:00Z. SUPERSEDES `HANDOFF_2026-09-03_continue_here.md`**, whose §2 and its
"~37%" defect statement are **RETRACTED** — read that file only for its §5 checks and §7 traps.

Cold-start: **this file** → `bugs_open/384_…md`, the **12:4xZ** update first (tail-first) →
`RUNBOOK_page_list_invalidation.md` (two new sections at the end) → `NOTES_…` tail →
`WRONG_CALLS.md` (6 entries from this lane, 09-02/09-03).

---

## 1. STATE — the seam is SOUND. The defect that has been breaking it since 09-02 is `bugs_open/454`, and it is fixed and live.

**384 STAYS OPEN** (owner ruling 2026-09-03: *"keep it open until those are checked and fixed"*).
But what is left is smaller and different from what the last handoff said.

`[MEASURED 2026-09-03 12:1x–12:4xZ]` Every archive-trigger write over a `query.*`-sourced listing
array in 10 days, counting only entries whose target page had an **active card created before the
write** (so the resolver genuinely had something to project), `LEAD` for the value each write
produced, split at `94f81cc60`:

| era | writer | writes over a real deficit | repaired | left blank |
|---|---|---|---|---|
| **before** `94f81cc60` | `save_page_sections` | **131** | **131** | **0** |
| **before** `94f81cc60` | `action:rebuild_blog_listing` | 1 | 1 | 0 |
| **after** `94f81cc60` | `save_page_sections` | 11 | 4 | **7** |
| **after** `94f81cc60` | `action:rebuild_blog_listing` | 1 | 1 | 0 |

And in the post-regression window the attribution is total — each write joined to the last
orchestration on its page within 20 minutes:

| attributed to | writes | repaired | left blank |
|---|---|---|---|
| `page-rerender`, `reason=section_data_resolved` (**the light re-render**) | 7 | **0** | **7** |
| `page-build-handler` (**full build**) | 1 | 1 | 0 |
| no run in window (full-build chains, keyed differently) | 4 | 4 | 0 |

**Every light re-render failed; everything else repaired.** That is `bugs_open/454` exactly.

## 2. `bugs_open/454` — what it is, and why it is NOT this lane's to fix

`classifyStoredSection` (`rerender_page_sections_action.go`) computes `plan := planSection(...)`,
branches on `plan.Status`, and returns **without assigning `c.plan`**. `renderPlannedSection` reads
`cls.plan` and gets the zero value, so `plan.ResolvedData` is **nil for every section of every
light re-render**. The render context becomes `base ⊕ stored content_data` with no fresh data at
all, and the persisted `content_data` is the stored map unchanged. Nothing errors; `rerendered`
equals the row count; no carry bucket names it. **The only observable is a negative.**

- **Introduced** `94f81cc60`, 2026-09-02 11:27:53Z ("035 P1: extract classifyStoredSection").
- **Fixed** `9831e9ab4`, 2026-09-03 10:00:40Z. Council **APPROVED** round 1 (`075cfedd`).
- **LIVE** in chassis `d0252fd4dab2a3a583d1cc8eb8e1b26e9c422d85` (v1.0.1358) since ~12:05Z
  `[MEASURED 12:33Z]`. Proven live at the artefact by the owning lane (`6f3116af0`).
- **Owned by** `bugfix_427_event_render` (session `bugs_open/427 [2e9752]`, live). I sent them this
  lane's census as corroboration 12:5xZ. **Do not fork an account of it.**

**`WebPath()` is also exonerated** — the last surviving candidate from the previous handoff, and a
two-line read as it predicted: `DeployedWebPath` → `DeployedAssetPath` (`platform/storage/url_helpers.go:317-347`)
builds a `RelativeURL` from the asset key and **cannot return "" for a non-empty `CardKey`**.

## 3. ⚠ THE ~37% FIGURE IS RETRACTED, AND SO IS THIS MORNING'S "ATTRIBUTION DEFECT"

Both were mine, from earlier today. Read `WRONG_CALLS.md` 2026-09-03 (this lane, 6th entry) before
you quote any number from the previous handoff.

1. **"the re-resolve repairs a blank listing only ~37% of the time"** — the census counted every
   entry with `image=''` as a failure. **An entry whose target page has no card is CORRECTLY blank.**
   Same rows, same window, same method: bare empties = 19 writes / 5 repaired; card-joined = 8 / 6.
   The gotcha was already in this lane's own RUNBOOK, written in August after the same mistake.
2. **"the lane credits the seam with repairs it did not perform"** — generalised from **two** cases,
   both of which fall inside 454's 19-hour regression window. §4's "proven four times" is not
   falsified by anything; it is merely un-recheckable now (see the retention limit below).

## 4. Two limits — carry them wherever these numbers go

- **The pre-regression 132 cannot be attributed to a code path.** `orchestration_states` holds
  **25.0 hours** (oldest run 2026-09-02 11:44Z) `[MEASURED 12:4xZ]` — re-measure, it moves. So
  132/132 is an OUTCOME over whatever mix of light re-renders and full builds was running. It is
  strong evidence of **no standing failure**; it is **not** proof the light re-render did the
  repairing. Only the 12-write post-regression table discriminates by path.
- **Every failure count here is a LOWER BOUND.** The archive triggers fire only on a *change*
  (`trg_page_component_artefact_archive_upd` on `rendered_html`;
  `trg_page_component_content_archive_upd` on `content_data` when `rendered_html` did not move). A
  write that changed neither leaves **no row** — and a byte-identical no-op is exactly what 454
  produces.

## 5. IN FLIGHT — the one thing left to prove, and how to read it

A `section_data_resolved` re-resolve at **designblog.co.uk/index**, filed 12:35:51Z:

- item `80a1c536-b75f-416d-ac72-952177229b5c`, `created_by='bugs_open/384_postfix_verify'`,
  item_key `page_rerender_index_aa51d9b8-511a-4bda-8207-a7e65c3abc4a_section_data_resolved_384verify`
- page `902083a9-15b9-4dae-9aa9-71fb9c6f2815`, site `aa51d9b8-511a-4bda-8207-a7e65c3abc4a`
- **the control:** the `content-listing` `articles` array has been **4 entries, 4 blank since
  05:25:29Z**, with all four cards active, correct and present since 05:05. **Success = 0 blank.**

```sql
SELECT w.status, pc.updated_at::timestamp(0),
       (SELECT count(*) FROM jsonb_array_elements(pc.content_data->'articles') e
         WHERE coalesce(e->>'image','')='') AS blank
  FROM site_work_items w, page_components pc
 WHERE w.created_by='bugs_open/384_postfix_verify'
   AND pc.page_id='902083a9-15b9-4dae-9aa9-71fb9c6f2815' AND pc.slot_name='content-listing';
```

**⚠ Do NOT re-fire if it is still `triaged`.** `build-dispatch-loop` runs **per site**; designblog's
turn came at 11:35:45Z and the rotation is ~15 sites / 90 min `[MEASURED 12:5xZ]`, so the next turn
is due ~13:05Z. A missing orchestration row is latency, not a dropped dispatch.

**⚠ Before reading a null result as a 384 failure, rule out `bugs_open/450`.** Its
`pageRefusesGenericBuild` (`owned_page_guard.go:175`) went live in the SAME image and makes
`save_page_sections` refuse an `owned` page or a pending tool shell — a refusal and a failed
re-resolve are indistinguishable at the array. **I checked this target before filing:**
`page_type='landing'`, `rebuild_policy='generic'`, so neither arm fires. If you pick a different
page, check it again — 450's guard refuses **58 tool pages across 12 sites** on the 427 lane's
measurement.

If it repairs: the seam is demonstrably sound on its own path on the fixed binary, and item 1 of §6
is closed. If it does not: that is a **genuine 384 residual on a clean binary** and the sharpest
case this lane has ever had — the runs will be inside retention, unlike everything before.

## 6. What 384 still owes before it can close

1. **The post-fix proof above.** In flight.
2. **The owned-page residual — 14 blanks / 3 pages.** Unchanged, structurally out of this seam's
   reach (`save_sections` refuses an owned page). Remedy shape exists (migration `486`'s
   `section_edit` → `section-editor` route). **Must NOT close inside 384** — file it or carry it
   into the owned-page seam's own round.
3. **This lane's sweep has NEVER RUN.** `check_page_list_stale` (migration `603`): 12 items in its
   lifetime, all born terminal, cause is `bugs_open/389`'s two-strike arm. **After 389 lands,
   re-validate the sweep and re-do the escalation watch from zero** — the old "zero escalations
   against 1-in-36" is zero over an empty denominator.
4. **`bugs_open/404`** — still unclaimed, still latent, unchanged.
5. **Re-run §1's census in a day or two, once the fixed binary has had traffic.** Expect the
   post-`94f81cc60` failures to stop accumulating. If they do not, the residual is real.
   SQL: `scripts/census_repair_rate.sql` in this directory (commit `f8110df1e`).

## 7. What belongs to other lanes — do not fix here

- **`bugs_open/454`** → `bugfix_427_event_render`. Fixed, approved, live. §2 above.
- **`bugs_open/450`** → its own lane. Its guard refuses tool pages; the 427 lane raised the reach
  measurement with them and the scope call is theirs.
- **`bugs_open/389`** (owned by `bugfix_308`): the two-strike arm counting SUCCESSES as strikes —
  the leopardess six-day starvation, which is **pre-regression and unaffected by any of today**.
  Still worth stamping the 09-02 23:12/23:20 resume into 389's evidence.
- **`dispatch_throughput`** (session `throughput`): the `detected` backlog is designed behaviour.
  This lane's earlier claim that the handler door parked them was WRONG and is retracted.

## 8. Traps — the previous handoff's §7 stands IN FULL, plus three

1. **A listing entry with `image: ""` is usually CORRECT.** Now in `LANDMINES.md` (entry
   *"A listing entry with `image: ""` is USUALLY CORRECT…"*, committed inside `09e7aa75b` — see
   trap 3). Join the card, require it to pre-date the write, and restrict to `query.*`-sourced
   fields.
2. **`WITH … AS MATERIALIZED` on every stage of a jsonb census.** PostgreSQL does not guarantee
   `AND` evaluation order, so a `jsonb_typeof(x)='array'` guard gets reordered behind
   `jsonb_array_length(x)` and the query dies with `cannot get array length of a scalar` — which
   reads like a data problem and is a planner one. Killed two attempts here. Runbook, not a
   landmine: it fails loudly.
3. **My `LANDMINES.md` append was swept into another session's commit** (`09e7aa75b`, a fence-file
   landmine from a different lane) ~30 seconds before my own `git commit` on the same file, which
   then found nothing to commit. Nothing is lost and forward-only holds — but if you go looking for
   the 384 landmine by commit message you will not find it. This is the documented same-file
   passenger that no hook can prevent.
4. **Grep `bugs_open/` when you FORM a hypothesis, not only when you file one.** The previous
   session's surviving hypothesis was exactly right and had been filed, fixed and approved by
   another lane 90 minutes before it was written down.

## 9. Where the knowledge lives

`bugs_open/384_…md` (**12:4xZ** update first) · `bugs_open/454_…` (the cause; owned elsewhere) ·
`bugs_open/389_…` (CONTRIB 09-02) · `RUNBOOK_page_list_invalidation.md` (+ the corrected census,
the provenance-table route, the 450 confound) · `scripts/census_repair_rate.sql` ·
`NOTES_page_list_invalidation.md` · `README_where_we_are.md` (owner prose, five entries) ·
`WRONG_CALLS.md` · `LANDMINES.md` · peers: `bugs_open/427 [2e9752]` owns 454 and has been told;
`bugfix_308` owns 389; `throughput` session; `leopardess [337f17]` is offline and **still has not
been told** its site was out of rerender service for six days.
