# HANDOFF — bugs_open/384 page-list invalidation · continue here

**Written 2026-09-04 ~12:1xZ. SUPERSEDES `HANDOFF_2026-09-03b_continue_here.md`**, which is still
correct in almost every part — read its §6 work list, §7 "belongs to other lanes" and §8 traps in
full. **What this file changes is one thing: the cause recorded for the last generic blank.**

Cold-start: **this file** → `bugs_open/384_…md`, the **2026-09-04 12:0xZ** update (tail-first) →
`HANDOFF_2026-09-03b_continue_here.md` §6–§8 → `RUNBOOK_page_list_invalidation.md` (three new
sections at the end) → `NOTES_…` tail → `WRONG_CALLS.md` (7 entries from this lane).

---

## 1. STATE — the seam is SOUND and is now **22 of 22** since the fix

`[MEASURED 2026-09-04 11:50:57Z]`, `scripts/census_repair_rate.sql`, four eras:

| era | writes over a real deficit | repaired | left blank |
|---|---|---|---|
| 1. before `94f81cc60` | **130** | **130** | 0 |
| 2. DURING 454's regression | 12 | 5 | **7** |
| 3. post-fix, build `d0252fd4` | **15** | **15** | 0 |
| 4. post-fix, build `239ab3626` (current) | **7** | **7** | 0 |

Build unchanged: `239ab3626fc7fb9cd4b121c82480bedafe2f555c` (v1.0.1360), **one commit fleet-wide**.

⚠ **Quote era 4 with its demand, not just its zero.** `last_write` is **08:12:16** — nothing
qualifying was written in the 3.6 h to the measurement. Era 4 grew by ONE. Re-run after a busy
period before treating 22/22 as settled; the instrument cannot tell "nothing failed" from
"nothing was asked".

**Residual `[MEASURED 11:5xZ]`, unchanged from 08:1xZ:** generic **701 carded / 1 blank (99.9%)**,
owned **14 / 14 blank (0.0%)**.

## 2. ⚠ THE 08:1xZ CAUSE FOR THE LAST GENERIC BLANK IS RETRACTED — read this before acting on it

That entry said of leopardessconsulting.co.uk `/blog`: *"**Cause:** `pages.sections` … is an empty
array … the re-render skips the page as sectionless and the listing can never be re-resolved."*
**Both halves are wrong.** Full account: `bugs_open/384` 12:0xZ §3–§4; `WRONG_CALLS.md` 2026-09-04.

- **The run never reached the re-resolve.** `rerender_page_sections_action.go:428-459` pre-checks
  each stored section and returns without rendering if one is missing a required `source:"llm"`
  field. `blog-listing_pre_037` declares **`section_heading`/`section_intro`** required-llm; the
  stored `content_data` carries the older dialect **`section_title`/`section_subtitle`**. The
  template renders the former pair and reads neither of the latter — **so the gate is RIGHT.**
- **`pages.sections` suppresses only the ESCALATION** that follows that bail. **Refilling it
  repairs nothing directly** — it converts a silent skip into a `needs_page` item, and repair then
  depends on that item draining. That is the action my sentence invited and it would have failed.
- **"Can never be re-resolved" is false.** `action:rebuild_blog_listing` resolves the card image
  correctly (this lane's **decision 3**, 2026-08-25, splicing `PageImageProjectionSQL`), on the
  same `purpose='card' AND status='active'` join the residual query uses. It runs inside
  **`rerender-pages`, per site**; leopardess's last was **17:48:31** and the card landed
  **17:50:13** — **93 seconds later**. Ordinary rotation (65 runs / 22 sites in 25.5 h), **not**
  a recurrence of 389's six-day leopardess starvation.

**So the blank is a race plus a wait, on a page where the seam is separately blocked.**

**⚠ The general form, and this lane has now hit it four times:** *a disposition names the step that
RECORDED an outcome, not the step that DECIDED it.* `skipped_sectionless_page` has **two** call
sites; the stored output tells them apart (the pre-check sets `skipped`+`section_count`, the other
does not). Grep the literal and count its callers before quoting it as a cause.

## 3. The PREDICTION on the record — check this first if you pick the lane up

**The blank clears on leopardess's next `rerender-pages` run, with no intervention.** If that run
lands and the entry is still blank, §2 is wrong and this is the sharpest 384 case there has been.

```sql
SELECT o.created_at::timestamp(0) FROM orchestration_states o
  JOIN sites s ON s.id::text = o.collected_data->'input_data'->>'site_id'
 WHERE o.owner_agent_type='rerender-pages' AND s.domain='leopardessconsulting.co.uk'
 ORDER BY 1 DESC LIMIT 1;
```
then re-run `scripts/residual_by_policy.sql` — the `generic` row should be **gone**.

⚠ `orchestration_states` holds **~25 h** (`min(created_at)` 2026-09-03 10:21Z when measured), so
past that the run may have aged out; the residual query is the durable half. ⚠ Grade against the
**card date**, not the clock.

## 4. The hole, measured — 4 slots / 3 pages / 3 sites; ONE is ours

*Refused by the render gate **and** escalation suppressed* — cannot render, cannot ask for help.
Query (all three `declaredPageSections` sources reproduced) is in the RUNBOOK.

| domain | url | slot | 384 consumer? |
|---|---|---|---|
| ai-agent-orchestration.com | `/blog.html` | `hero`, `call-to-action` | no |
| gaswholesalers.com | `/tools/tool-gas-unit-converter.html` | `tool-gas-unit-converter` | no |
| **leopardessconsulting.co.uk** | **`/blog.html`** | **`blog-listing`** | **YES** |

So 08:1xZ's "1 page, 1 site" bound **survives on a better predicate** — good news, and the reason
to keep the wider query rather than the narrow one. **Do not raise it as a class.**

⚠ **A separate population this lane did NOT measure: 73 slots / 66 pages are refused but WOULD
escalate.** Whether those `needs_page` items drain is `bugs_open/187`/`389` territory. Do not quote
my 4 as the cost of the render gate.

> **⚠ CORRECTED 12:3xZ, after the `ai-agent-orchestration` lane re-ran my query.** Three changes,
> all in the bug file's 12:3xZ update.
> 1. **My predicate expressed only ONE of the gate's two branches.** It refuses on (a)
>    `len(s.contentData) == 0`, schema-independent, AND (b) a required `source:"llm"` field empty;
>    I wrote only (b). **The hole's membership survives at 4** — empty content_data also leaves
>    every required field absent, so those slots matched (b) by accident — but the split is
>    **3 branch (a) / 1 branch (b)**, and only leopardess is (b). Branch (a) means the writer
>    authors the WHOLE slot; different repair. **The escalatable figure was 64/60 and is 73/66**
>    (struck above).
> 2. **The "latent" population is ZERO, not the large number both of us expected.** Keying on
>    *unsatisfiable alone* (fallback already gone, held out only by intact content) returns
>    **121 pages / 29 sites** — and **120 of them carry a self-contained tool component that the
>    loop SKIPS before either branch**, so they can never be refused; the 121st has no components.
>    **Ask whether a population is ELIGIBLE for the mechanism before publishing it as at-risk** —
>    measuring both conjuncts is what produced the 121.
> 3. **`input_schema.fields` is an OBJECT, not an array** — `jsonb_path_query_array($.fields[*] ...)`
>    returns a clean empty result that reads as "declares no required fields". Use `jsonb_each`.
>
> **The RUNBOOK query is corrected**; the version first published there undercounts branch (a).

## 5. What 384 still owes — 09-03b's §6 list, restated with today's status

| # | item | status | whose |
|---|---|---|---|
| 1 | post-fix proof | ✅ DONE 09-03, and 22/22 as of today | this lane |
| 2 | owned-page residual 14/14 | ✅ CONTRIB in `bugs_open/389` §2 + remedy-safety addendum | 389's call |
| 3 | the `page_list_stale` sweep | ⚠ 13 of 14 lifetime items still born `unresolved` — blocked on 389's two-strike arm. **Re-do the escalation watch from ZERO after 389 lands** | this lane, later |
| 4 | `bugs_open/404` | ⚠ taken 2026-08-26, then dormant. **Not unclaimed** — message the holder | another lane |
| 5 | re-run the census | ✅ DONE today (§1). Re-run after a BUSY period (§1 caveat) | this lane |
| 6 | 1 frozen NULL-`component_id` row (gamesdesign `/game-jelly-invaders`) | noted-not-owned; not a listing this seam feeds | nobody yet |
| 7 | **the last generic blank** | **NEW STATUS: transient, prediction filed (§3)** | this lane |

## 6. Traps — 09-03b §8 stands IN FULL, plus two from today

1. **A bulk `kubectl exec … psql` census TRUNCATES MID-ROW and the short file looks complete.**
   Already a landmine since 2026-07-30 and I did not grep for it: exit **1**, 81 perfect rows, the
   last one cut off with the error text spliced on. Bucketing it would have read "everything
   pre-regression, everything repaired". **Aggregate server-side** — recipe in the RUNBOOK.
2. **⚠ `grep -c 'src=""'` CANNOT FAIL on half our templates** — a blank image behind `{{if .image}}`
   renders **no element at all**. `[MEASURED 2026-09-04]` **8 of 15** components that render
   `.image` guard it. This lane used *"N `<img src>` and zero `src=\"\"`"* as its 09-03 proof; the
   second half was vacuous. **Count `<article>` blocks against `<img>` tags** and cross-check
   against `jsonb_array_length(content_data->'articles')`. Now a landmine; recipe in the RUNBOOK.

## 7. Where the knowledge lives

`bugs_open/384_HANDOFF_2026-08-24_a_landed_card_image_never_invalidates_the_listing_that_renders_it.md`
(**2026-09-04 12:0xZ** update first) · `docs/agent_docs/docs024_key_docs_latest/bugfix_384_page_list_invalidation/`
— `HANDOFF_2026-09-03b_continue_here.md` (§6–§8 still current), `RUNBOOK_page_list_invalidation.md`,
`NOTES_page_list_invalidation.md`, `README_where_we_are.md`,
`scripts/census_repair_rate.sql`, `scripts/residual_by_policy.sql` ·
`docs/agent_docs/docs024_key_docs_latest/WRONG_CALLS.md` ·
`docs/agent_docs/docs024_key_docs_latest/LANDMINES.md` ·
peers: `bugs_open/389` (owned by `bugfix_308`) · `bugs_open/427` lane owns `bugs_closed/454` ·
`bugs_open/187` owns the escalation guard · `bugs_open/204`/`443` own the `pages.sections` cache.
