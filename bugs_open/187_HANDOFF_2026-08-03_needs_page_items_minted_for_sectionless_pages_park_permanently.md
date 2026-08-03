# 187 — `needs_page` items minted for section-less pages park permanently; the 177 shape under five other emitters

**Filed 2026-08-03 by the bugfix_177 lane, at the council's direction** (corr
`982507b0`, `bug_historian` + `architecture` seats: the deferred population must
be "tracked as its own ticket, not just a paragraph in this close-out, given
this platform's history of exactly this gap recurring under different
emitters"). **Status: OPEN, mechanism `[UNVERIFIED]` per emitter** — the
population is measured; nobody has yet read the five emitters to establish
which mint items that were unsatisfiable at birth versus items whose data
legitimately never arrived.

**CLAIMED 2026-08-03 (evening)** by a dedicated thread (workstream dir:
`docs/agent_docs/docs024_key_docs_latest/bugfix_187_sectionless_needs_page/`).
Plan: per-emitter triage per the recipe below, then the shared-resolver
extraction the `architecture` seat flagged, through the council gate.
Contribute findings into this file; do not start a competing fix.

## The measured population (live DB, 2026-08-03; re-run before quoting)

```sql
SELECT source, item_type, status, count(*), max(created_at)::date AS newest
FROM site_work_items
WHERE error LIKE '%no sections ready to build%'
GROUP BY 1,2,3 ORDER BY 4 DESC;
```

| source | item_type | status | n | newest |
|---|---|---|---|---|
| image-build-handler | needs_page | needs_human_review | 11 | 07-29 |
| reconcile_site_plan | needs_page | needs_human_review | 9 | 07-25 |
| page-rerender | needs_page | needs_human_review | 2 | 07-28 |
| gemini-p7-verification | needs_page | needs_human_review | 1 | 07-27 |
| reconcile_site_plan | needs_page | rejected/complete/unresolved | 5 | — |
| json-leak-fix | needs_page | rejected | 1 | 07-15 |

(The `tool_content` rows this query also matched are `bugs_closed/177`, fixed.)

## The pattern, and why this is filed as one ticket not five

016b §9: **"A work item can be UNSATISFIABLE AT BIRTH — a 0% class completion
rate indicts the EMITTER, not the handler."** 177 established the diagnostic
recipe: (1) per-class completion rate over all history; (2) find a sibling
class from the same emitter that succeeds and diff the inputs; (3) read the
handler's input resolution (`load_page_sections_from_spec_action.go`: plan
tables → `site_specs.site_plan` → `pages.sections` → same-role sibling
synthesis) as the contract, and ask whether the emitter ever satisfies it.

**Caution the other way (bugs_closed/015):** a page that SHOULD have sections
arriving with `sections=[]` is a real defect, and `reconcile_site_plan`'s items
are plausibly that case (the reconciler asks for a page the plan wants — the
plan may legitimately gain sections later). So per-emitter triage must sort
**unsatisfiable-at-birth** (177's shape → emit-side guard) from
**data-not-yet-arrived** (legitimate deferral — the item is doing its job) from
**upstream defect left the page section-less** (015's shape → fix the cause,
keep the item). Do NOT blanket-apply the 177 guard to all five.

## Fix template, if an emitter IS the 177 shape

`raiseToolContentItem` (`platform/orchestration/actions/tool_content_item.go`)
is the worked example: resolve the handler's own sources read-only at emit
time, skip with an observable disposition when unsatisfiable, route the write
through `insertWorkItem`. The `architecture` seat flagged (advisory, corr
`982507b0`) that a THIRD copy of the satisfiability-mirror would be the moment
to extract one shared resolver rather than grow a family of mirrors — whoever
takes this ticket should weigh that extraction first.

## Related

- `bugs_closed/177` — the worked case + the §9 pattern entry.
- `bugs_open/033` — the queue these park in; owner ruling 2026-07-25 "the queue
  should not fill"; ~~`needs_page` IS drainable by the revalidator but these rows
  predate/evade it — check why before hand-sweeping.~~
  > **CORRECTED 2026-08-03 (187 lane):** false. `reviewRevalidators`
  > (`revalidate_review_queue_action.go:149`) covers exactly `unresolved_cta`,
  > `required_fields_missing`, `needs_section_data`. **`needs_page` is an
  > UNCOVERED type — nothing drains these rows.** The check was one grep of the
  > map. WRONG_CALLS entry recorded.
- `bugs_open/087` — page-rebuild's writer got no section plan (a consumer-side
  sibling of this emitter-side question).
- `bugs_closed/015` / `bugs_closed/081` — the "page should have sections and
  does not" cause family.

## Per-emitter triage — DONE 2026-08-03 (187 lane; the `[UNVERIFIED]` above is now resolved)

All 28 parked rows measured against live state (join `pages` BY NAME — 27/28
carry NULL `page_id`), every emitter read at HEAD. 090 run filed, correlation
`b3dcb102-d4bf-44c1-b2a2-3068ce95acc6`.

- **image-build-handler (14) — 177's shape, guard the emit.**
  `flag_page_image_rebuild_action.go:132-159` emits from only (site_id,
  page_name); its own header comment says "VERIFY BEFORE RELYING ON IT", and
  the assumption is measured false: every parked row's page declares
  `sections=[]` with no plan membership (except brands-index/shop-index —
  satisfiable, see below).
- **page-rerender (4) — 177's shape, guard the emit.**
  `escalateRerenderToWriter` (`rerender_page_sections_action.go:803`) fires on
  a NULL `content_data` slot and asks the writer to rebuild from a section
  plan that does not exist; a tool page's widget slot rendering from other
  than `content_data` makes the trigger itself a false alarm there.
- **reconcile_site_plan (9) — leave the emitter alone.** 4 rows' pages were
  BUILT since by other routes (tungsten-guide, board-setup, cases-index,
  thames-water — items stale, drainable with evidence); 5 rows point at pages
  with 0 sections + 0 plan rows (directory-index, practice, guides-index,
  brand-detail, platform-log-index) — the `bugs_closed/015` shape, a REAL gap
  the item is correctly surfacing. A guard here would suppress genuine
  findings.
- **gemini-p7-verification (1) / json-leak-fix (1) — manual enqueues, no code
  path exists** (grep: doc mentions only). grip-styles is satisfiable NOW
  (3 declared, 3 plan rows, 0 slots) — genuine pending work, stays parked;
  the json-leak-fix row is already `rejected`.

Fix shipping from the 187 lane (PLAN in
`docs024_key_docs_latest/bugfix_187_sectionless_needs_page/`): shared
read-only satisfiability resolver extracted from 177's guard (the third-copy
moment the architecture seat named), wired into both 177-shaped emitters, plus
a `needs_page` entry in `reviewRevalidators` so satisfied asks close with
evidence instead of parking for ever.

## Verify (per emitter, once triaged)

The 177 recipe: class completion rate before/after; for a guarded emitter, an
induced emit against a section-less page yields a logged+surfaced skip and no
row; positive control: the same emitter against a page with resolvable
sections still mints.
