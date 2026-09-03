# HANDOFF — bugs_open/384 page-list invalidation · continue here

**Written 2026-09-03 ~13:00Z. SUPERSEDES `HANDOFF_2026-09-03_continue_here.md`**, whose §2 and its
"~37%" defect statement are **RETRACTED** — read that file only for its §5 checks and §7 traps.

Cold-start: **this file** → `bugs_open/384_…md`, the **12:4xZ** update first (tail-first) →
`RUNBOOK_page_list_invalidation.md` (two new sections at the end) → `NOTES_…` tail →
`WRONG_CALLS.md` (6 entries from this lane, 09-02/09-03).

---

## 1. STATE — the seam is SOUND: 132/132 before the regression, 0/7 during, 18/18 after as of 15:27Z (§5). The defect that broke it 09-02→09-03 is `bugs_open/454`, fixed and live.

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

**AND THE THIRD ERA IS IN** (added 15:3xZ). The `components` lane (`bugs_open/425`) drained its
whole class on the fixed binary: **17 new shape / 0 old at 15:27Z**, six of them a batch with
baselines recorded BEFORE dispatch. With designblog that is **18 repaired / 0 blank after the fix**.
So the seam reads **132/132 before the regression · 0/7 during it · 18/18 after (as of 15:27Z —
that column is still moving, so quote it with the time)**.

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
- **CLOSED 2026-09-03** and moved to `bugs_closed/454_…md`, owned by `bugfix_427_event_render`.
  That lane took this lane's census as its §14 evidence and re-ran the SQL rather than quoting it
  (reproduced exactly: 144 rows, pre 132/132/0, post 12/5/7), and verified my designblog canary
  independently by WebFetch. **Do not fork an account of it.**

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

## 5. ✅ PROVEN 12:54Z — the seam's own path repairs on the fixed binary

**RESULT (added 13:0xZ, after this file was first written).** The verification below RAN and
REPAIRED. designblog.co.uk/index `content-listing.articles`: **4 blank → 0 blank**, `updated_at`
05:25:29Z → **12:54:43Z**, `rendered_html` 2,494 → **3,327 B**, and the served page returns HTTP 200
with **four card `<img src>` and zero `src=""`** (all four .jpg files 200, 35–62 KB). The run was
`page-rerender` / `section_data_resolved`, COMPLETED 12:54:41Z. Item `complete`, `attempt_count 0`.

**⚠ Do NOT quote the run's counts as the proof.** `rerendered=4, carried=0, escalated=false` is
exactly what the BROKEN runs reported — that is 454's signature. The proof is the artefact plus an
independent baseline: the `bugs_open/427` lane read this row at 12:54Z as "4 articles, 0 with
images, updated_at still 05:25:28Z, html 2,494 bytes", seconds before the write landed.

**Scope of the claim:** this is a positive for the **`image` field only**. The `components` lane
(`bugs_open/425`) checked and the deck already carried `articles[0].excerpt` from an earlier BUILD,
so the item-SHAPE axis could not have failed here and this run says nothing about it. See the bug
file's 13:0xZ update for their discriminator — *"a re-render wrote a row carrying key K" is not
"the re-render produced K"*.

**So item 1 of §6 is CLOSED.** What follows is the original in-flight section, kept for its recipe
and its controls.

### The original in-flight section (recipe + controls, still the right ones to reuse)

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

## 6. What 384 still owes — EVERY ITEM RE-CHECKED 2026-09-03 16:1xZ, and two of them had gone stale

**Do not work this list from the previous handoff.** Three of its five items were wrong or out of
date by the time it was written, and I only found that by opening each bug rather than restating it.
Status below is checked, with the query or commit that checked it.

| # | item | status as of 16:1xZ | whose |
|---|---|---|---|
| 1 | post-fix proof | ✅ **DONE** (§5) | this lane |
| 2 | owned-page residual, 14/14 | ✅ **CONTRIB filed into `bugs_open/389` 16:4xZ** | 389's call now |
| 3 | this lane's sweep | ⚠ **NO LONGER "never run" — it worked today** | this lane |
| 4 | `bugs_open/404` | ⚠ **NOT unclaimed — taken 2026-08-26** | another lane |
| 5 | re-run the census | outstanding | this lane |
| 6 | 2 stranded NULL-id rows | unowned, not a listing | nobody yet |

### 2. ✅ DONE 16:4xZ — the owned-page residual is contributed into `bugs_open/389` §2, not re-filed

The previous handoff said "file it or carry it". Neither: **contribute the measurement into 389 §2**,
which is the class, and which already carries a correction from the `bugfix_333_owned_page_door`
lane that matters more than anything this lane has measured:

> the door **structurally cannot cover** owned pages. It parks only when the TARGET HANDLER declares
> `refuse_owned_page` (migration `488`), and `page-rerender` **must never declare it** — a
> per-agent/per-branch ruling, register **WII-028**, made **with the 384 lane**. `bugfix_333` is
> CLOSED (2026-08-26) and this is the part it could not fix.

And the half that makes it a live cost rather than a static blank: **the `save_sections` refusal has
no `wont_fix` terminal** (migration `480` covers `load_page_record` only), so these items loop
`failed` → `triaged` for ever. `[MEASURED 2026-09-03 16:0xZ]` owned-page `page_rerender` items in
14 days: **1,436 `complete` · 402 `unresolved` · 76 `failed`, last failure 15:56:07Z** — minutes
before this was written. The 1,436 "complete" are 389's own complete-and-unchanged class.

**The CONTRIB is filed** (`bugs_open/389`, section dated 2026-09-03), and it carries three things:

1. **A retraction of this lane's OWN 09-02 contribution in that file**, which ended *"so it has
   never run once"* about the sweep. It ran today (item 3 below). True when written, false now, and
   it was load-bearing in how we put it to them.
2. **Artefact-level proof of their class 2** — the three pages carry **26 `page_rerender` items that
   reached `complete`** while not one of the three arrays has been written since July/mid-August:
   leopardess `llm-cost-calculator` (frozen 2026-07-17, 7 completes), leopardess
   `tool-ai-vendor-trust-checklist` (2026-07-30, 9), finetuning `llm-cost-calculator`
   (2026-08-12, 10). That is "complete provably meant unchanged" measured on the stored artefact,
   which is what their Verification bar asks for and what this lane could supply cheaply.
3. **A refinement of their "these loop `failed`→`triaged`" line**, on a WIDER cut (all owned-page
   `page_rerender`, 14 days, not just `cta_links_stale`): `[MEASURED 16:33:10Z]` 76 `failed` **all at
   `attempt_count 3`** — ladder exhausted, not cycling — and 405 `unresolved` all at 0. Offered as a
   different population, explicitly NOT as a refutation; they should re-cut on their own predicate.

**No fix proposed and no new bug opened** — it is their class on their handler, and a second account
would drift. Migration `486`'s `section_edit` → `section-editor` route is the remedy shape; the call
is 389's. Flagged to them that register **WII-028** was a ruling made with this lane, so a candidate
that would change it should re-open it jointly rather than underneath us.

⚠ **No live session holds 389 or 308** (checked 16:4xZ) — the bug file is the channel, as it was for
this lane's 09-02 contribution. Do not wait for a reply.

### 3. The sweep — ⚠ **IT RAN TODAY. The "never run" line is retired.**

`[MEASURED 2026-09-03 16:0xZ]` `check_page_list_stale` (migration `603`) lifetime items:
**13 `unresolved` + 1 `complete`**. The `complete` one is the first ever, and it worked:

- filed **14:42:47Z** by `completeness-discovery-agent` on oxenunity.com
  `/tool-take-strength-scorer`, naming a real deficit — `tool-cta` `items`, entry
  `/tools/community-growth/index.html`, `stored_image: ""` against
  `current_image: /assets/images/card-tool-community-growth.jpg`;
- dispatched its own `page-rerender` with `reason=section_data_resolved`, **`cause=page_list_stale`**
  at 15:02:10Z; item `complete` 15:02:36Z, `attempt_count 0`. The slot is **0 blank** now.

⚠ **AND A NEAR-MISS I ALMOST FILED AS A 384 RESIDUAL — read this before grading any sweep run.**
The history says the sweep's own write took the slot **2 blanks → 2 blanks**, and the repair to 0
came 30 minutes later from an unrelated `card_landed` re-resolve. That reads as "the sweep detected
correctly and its remedy did nothing", which is what I was about to write. **It is wrong.** At
15:02:10 only ONE of the three cards existed (`card_tool_community_growth`, 02:15:47); the other two
landed at 15:21:56 and 15:22:29. And the one repairable entry had already been fixed at 14:58 by a
`cta_links_stale` re-render three minutes earlier. **So there was nothing left for the sweep's run
to repair, and 2 → 2 was the correct outcome.** Same lesson as the ~37% one level along: *a blank
with no card is correct*, and here it made a working mechanism look broken. **Always date the cards
against the run before grading a repair.**

**What is still owed:** 13 of 14 lifetime items are still born `unresolved`, so `bugs_open/389`'s
two-strike arm still binds and one working run is not a working sweep. **Re-do the escalation watch
from zero after 389 lands** — the old "zero escalations against 1-in-36" is zero over an empty
denominator. Keep the oxenunity run as the worked example of what a good one looks like.

### 4. `bugs_open/404` — ⚠ **NOT unclaimed. It was taken on 2026-08-26 and this lane has been saying otherwise for a week.**

Filed by *this* lane 2026-08-25 (`efc0db7bc`, out of a council advisory). **Taken 2026-08-26** —
`98c48e3a1` "404 CONTRIB (taking the bug)" — followed by four corrections that day
(`a975530ce`, `87e942570`, `4a31c6b8f`, `b640c696a`). Quiet since; the file has not moved in 8 days.
So the honest status is **taken, then dormant**, not "unclaimed".

⚠ **And a trap in the tool: `scripts/who-owns.py 404` names THIS lane as the most likely owner**
(10 mentions, ACTIVE) — purely because our own handoffs keep citing it as unclaimed. **A citation
count is not ownership**, and a lane that repeatedly lists a bug it is not working will rank itself
top. Resolve by `git log` on the FILE PATH, which is what the CLAUDE.md landmine says and what
settled it here. Before touching 404, message whoever holds it rather than assuming it is free.

### 5. Re-run §1's census once the fixed binary has had a day of traffic

`scripts/census_repair_rate.sql` (commit `f8110df1e`). Expect the post-`94f81cc60` failures to stop
accumulating. If they do not, the residual is real. Quote the "after" column **with its time** — it
is still moving.

### 6. Two stranded NULL-`component_id` rows — unowned, and NOT this seam's

finetuning.uk `/blog` (`article-grid`) and gamesdesign.co.uk `/game-jelly-invaders` (`section`) —
the only two live rows whose `slot_name` matches neither an active component's `name` nor its
`function`, so no re-render can ever resolve them (§8 trap 5). Neither is a listing this seam feeds,
so they are noted, not owned. If a stuck page ever appears in the census, check that column first.

## 7. What belongs to other lanes — do not fix here

- **`bugs_closed/454`** → `bugfix_427_event_render`. Fixed, approved, live, **CLOSED**. §2 above.
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
