# HANDOFF — bugs_open/384 page-list invalidation · continue here

**Written 2026-09-03 ~09:30Z. SUPERSEDES `HANDOFF_2026-09-02_continue_here.md`.**
Read that one only for the traps in its §7 (still valid) — its §2 experiment has now RESOLVED.

Cold-start: **this file** → `bugs_open/384_…md`, updates dated **2026-09-03** first (tail-first;
they supersede everything above) → `RUNBOOK_…` → `NOTES_…` tail → `WRONG_CALLS.md` (5 entries from
this lane, 09-02/09-03).

---

## 1. STATE: 384 STAYS OPEN (owner ruling). The starvation is resolved — but the defect is
## REPRODUCING LIVE on the seam's own path

> **OWNER RULING 2026-09-03: *"keep it open until those are checked and fixed."*** The decision
> that stood in §2 is settled — **384 does not close on "recovers by rotation"**. The remaining
> items are to be checked AND fixed. My Option-A recommendation is superseded; do not re-open the
> question. **And the first check found the defect live, which falsifies the framing that decision
> rested on.**

`[MEASURED 2026-09-03 09:1x–09:2xZ]`

- **The page is repaired.** leopardessconsulting.co.uk/blog serves **13 of 13 card images**
  (42,483 bytes; was 11 of 13 / 38,319 bytes byte-identical across three reads on 08-31 and 09-02).
- **The repairing write:** `page_component_history` 2026-09-02 **23:20:15**,
  `action:rebuild_blog_listing`, pre-image 13 articles / **2 blank**.
- **The §4 chain of the previous handoff is CONFIRMED**, by the brand (the state discriminator),
  not by timing: three items filed 23:12–23:22 were **unbranded**, dispatched, and completed —
  after six days in which every such item was born `unresolved` with `[unresolved after N attempts]`.
  The growth-posture door parked nothing (`growth_release_recipe` absent on every row).
- **Fleet census now:** generic **5 blanks / 2 pages, all IN-FLIGHT** (cards landed 4.3h and 7.1h
  ago — nothing stuck); owned **14 / 3**, unchanged.

## 2. ⚠ THE LIVE REPRODUCTION — read this before touching anything

`[ALL MEASURED 2026-09-03 09:3x–10:0xZ]` **designblog.co.uk/index is broken right now**, on the
current build, and it **is not a blog listing**:

- Component `content-listing`, field `articles`, source **`query.blog_posts`**. The page is
  **`save_page_sections`-maintained** — the path I previously said the seam "repairs correctly".
- **The seam fired correctly.** 5 `page_rerender` items from `derive_card_asset`,
  `reason=section_data_resolved`, all complete by 05:25:51; two carry
  `consumes: ["query.blog_posts"]`, matching the component's own source.
- **Every projection input is present and correct.** Four target pages `active` / `page_type
  blog-post`; four card assets `active` with non-empty `asset_key` and matching `site_id`; cards
  landed 04:56:39–05:05:10.
- **The array was rewritten twice after all four cards existed** (05:06:21, 05:25:28) and holds
  **4 entries, 4 blank**, pre-image blank each time.

**THE CARRY HYPOTHESIS IS REFUTED.** The runs are still inside `orchestration_states` retention
(hours old, not days — the problem that defeated the leopardess diagnosis):

| run | section_count | rerendered | carried | escalated | skipped |
|---|---|---|---|---|---|
| 05:06:19 | 4 | **4** | **0** | false | false |
| 05:08:24 | 4 | 4 | 0 | false | false |

Nothing was carried. So **`plan.Status != "ready"` (`rerender_page_sections_action.go:509`) is NOT
the cause** — I carried that candidate through two handoffs and it is now dead for this case.

**Remaining hypothesis, UNTESTED — do not build a fix on it:** the `query.blog_posts` resolve
returns without populating `articles`, so `plan.ResolvedData` lacks the key, `mergedContent` keeps
the stored blank array, and the section still counts as `rerendered` because it did render HTML.
That would also explain `content_data` unchanged while `rendered_html` changed.

**A `090` was fired on this live case 2026-09-03 ~10:1xZ** with the seeding corrected (whole files,
not symbols; the live workflow routing quoted in the symptom). **Find its verdict before
re-diagnosing** — intake slug `query_blog_posts_resolves_empty_image_despite_active_cards`,
`RUNTIME_SITE=designblog.co.uk`. If it returned UNVERIFIABLE again, read its "still needed" list:
last time that list was what actually cracked the case.

### Three corrections this forces to my own earlier claims

1. **RETRACTED: "the defect is BLOG-LISTING-specific."** It is not.
2. **WEAKENED: "four of five demonstrations are genuine."** I verified a `save_page_sections`
   **write happened** on finetuning at 19:13/19:14/19:15 — **not that those writes produced
   non-blank images.** The 0-blank figure comes from the 08-26 census. **Re-verify before quoting.**
3. The leopardess repair (09-02 23:20) is still real and the starvation chain is still confirmed —
   but it repaired via `rebuild_blog_listing`, and that tells us nothing about whether the seam's
   own path works. On this evidence it does not.

## 3. What else must be settled before close, whichever option

1. **The owned-page residual — 14 blanks / 3 pages, unchanged.** Structurally out of this seam's
   reach (`save_sections` refuses an owned page). Remedy shape exists: migration `486`'s
   `section_edit` → `section-editor` route. **It must NOT close inside 384** — file it or carry it
   into the owned-page seam's own round.
2. **This lane's sweep has NEVER RUN.** `check_page_list_stale` (migration `603`): 12 items in its
   lifetime, all born terminal. Cause is 389's arm. **After 389 lands, re-validate the sweep and
   re-do the escalation watch from zero** — the old "zero escalations against 1-in-36" is zero over
   an empty denominator and must not be quoted.
3. **`bugs_open/404`** — still unclaimed, still latent, unchanged by any of this.

## 4. What belongs to other lanes — do not fix here

- **`bugs_open/389`** (owned by `bugfix_308`): the two-strike arm counting SUCCESSES as strikes.
  Evidence contributed 09-02, including the cross-producer shared-key property. **If you want one
  thing added: stamp the 2026-09-02 23:12/23:20 resume into 389's evidence** — the
  `dispatch_throughput` lane asked for it and it is now measurable. A fix candidate is noted there
  (`insertWorkItem` already exempts `recurrenceExpected`); deliberately not taken.
- **`dispatch_throughput`** (live session `throughput`): the `detected` backlog is **designed
  behaviour, not a bug** — handler-less rows are flags, excluded upstream of the doors by
  `scored`'s `COALESCE(wi.handler_agent,'') <> ''`. **My claim that the handler door parked them
  was WRONG and is retracted. Do not re-open it.** They also supplied the recovery-differential
  control that separated leopardess from its comparators, and adopted the `unresolved`-is-terminal
  overcount caveat into their runbook.

## 5. Checks worth running (NOT a closing checklist any more — see the owner ruling in §1)

1. **Measure the rotation latency** so Option A closes on a number:
   ```sql
   -- per blog-listing page: gap between a card landing and the next rebuild_blog_listing write
   SELECT s.domain, p.name, max(h.created_at) AS last_rebuild,
          round(extract(epoch from (now()-max(h.created_at)))/3600.0,1) AS hours_since
     FROM page_component_history h JOIN pages p ON p.id=h.page_id JOIN sites s ON s.id=h.site_id
    WHERE h.application_name='action:rebuild_blog_listing'
    GROUP BY 1,2 ORDER BY 4 DESC;
   ```
2. **Re-read the in-flight five** (they were 4–7h old at writing). Expect 0; a still-blank entry
   after ~24h is a live instance of §2's gap and changes the decision.
   **NOT CHECKED at writing, and it bears on §2:** whether `designblog.co.uk/index` and
   `oxenunity.com/tool-take-strength-scorer` are `rebuild_blog_listing`- or
   `save_page_sections`-maintained. Same history query as above, filtered to those two pages.
3. **Re-run the fleet census** (in `bugs_open/384` §7 of the 08-26 handoff, or the 09-03 update).
   Expect generic ≈ 0 and owned 14.
4. **Move BOTH paths in one commit** — `git mv` plus a one-sided pathspec ships a COPY:
   `git commit bugs_open/384_….md bugs_closed/384_….md -m "..."`, then verify at HEAD:
   `git ls-tree -r --name-only HEAD -- bugs_open/ bugs_closed/ | grep 384` → exactly one line.

## 6. Build state — re-read it, do not trust this paragraph

`[MEASURED 2026-09-03 09:14Z]` Mid-roll: `0d2feee2ff61` (228 pods) → **`7bf1ff674021`** (88 pods,
strict descendant). **All four lane commits are ancestors of both**, and of every build this lane
has seen. The chain's files are **unchanged** between `0d2feee2ff61` and `7bf1ff674021`.
This file has already been overtaken by two rolls in 12 hours — **re-read the stamp and re-run the
confound check** (`git log <old>..<new> -- <the chain's files>`) rather than trusting the above.

## 7. Traps — the previous handoff's §7 still applies IN FULL, plus one more

Five `WRONG_CALLS` entries from this lane on 09-02/09-03. The new one:
**I dated a rolling-window prediction from the key carrying the SYMPTOM, not the key gating the
SERVICE** — 21 hours out, and my own "still broken on 09-04 ⇒ REFUTED" instruction would have told
you to discard a correct mechanism. What saved it was the STATE discriminator (is the row branded?).
**Prefer a state test to a timing test; when you offer both, say which one decides.**
Also still live from 09-02: `updated_at` is not trigger-maintained (use `page_component_history`,
and key on `page_id` — `save_page_sections` deletes and re-inserts, orphaning 98.3% of history from
a live-row join); a NOW-census of `triaged`/`approved` reads zero on every site; `unresolved` is
terminal and reads as open work to the 090 coverage clause.

## 8. Where the knowledge lives

`bugs_open/384_…md` (09-03 updates first) · `bugs_open/389_…` (CONTRIB 09-02) ·
`RUNBOOK_page_list_invalidation.md` (incl. the `page_component_history` recipe and 090 seeding notes) ·
`NOTES_page_list_invalidation.md` · `README_where_we_are.md` (owner prose, four entries) ·
`WRONG_CALLS.md` · 090 run correlation `149ec925-ffb7-41eb-806a-1595b8ff2226` (verdict
UNVERIFIABLE; its missing-evidence list is what cracked the case) · peers: `throughput` session,
`bugfix_308` lane owns 389, `leopardess [5c2e15]` session **has NOT been told** its site was out of
rerender service for six days.
