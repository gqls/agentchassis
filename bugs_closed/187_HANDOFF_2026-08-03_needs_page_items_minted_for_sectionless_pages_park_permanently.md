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
`b3dcb102-d4bf-44c1-b2a2-3068ce95acc6` — verdict **UNVERIFIABLE** (3
iterations; its own citations name the stale code index and failed source
reads as the cause — the loop could not read the code, it did not contradict
anything). Per the 2026-07-31 ruling the substitute is stated plainly: every
emitter read at HEAD and quoted, the handler chain read in full, all 28 rows
measured live, and the natural experiment inherited from 177 (same handler,
sectioned sibling items complete). The run's one find is a confirmation from
the other direction: the ONLY unpark ever performed was manual
(`spec.promoted_by='dartsonline-traffic-workstream'`, 2 rows fleet-wide, one
day, 0 Go hits for the field at HEAD) — no systematic drain exists.

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

---

## CLOSED 2026-08-04 — fixed AND live (v1.0.1248, pod-proven both replicas)

**What shipped** (council `e2e87b04` APPROVED round 1, 5 advisories none high,
each answered with a check — see lane NOTES; commit `12ae5824f`,
`Council-Submitted` trailer; 090 run `b3dcb102` UNVERIFIABLE on tooling,
refuting nothing, first-hand substitute stated per the 2026-07-31 ruling):

- **One shared read-only satisfiability seam** —
  `page_section_satisfiability.go` (`declaredPageSections` extracted pure from
  177's guard, old symbol fully removed; `pageInCurrentPlan` mirrors only the
  synthesis GATE, fail-open toward emitting; `pageSectionsSatisfiable`;
  `revalidateNeedsPage`). Registered as **WII-010**.
- **Emit guards** on the two 177-shaped emitters:
  `flag_page_image_rebuild_action.go` (its "VERIFY BEFORE RELYING ON IT"
  header replaced with the verified statement) and `escalateRerenderToWriter`.
  Skip is observable: `skipped_sectionless_page` in log + return map.
  `reconcile_site_plan` deliberately NOT guarded (015-shape gaps are real
  findings).
- **`needs_page` registered in `reviewRevalidators`** — resolved only on
  positive name-matched evidence (slot_name via NormalizeComponentFunction;
  95.7% live match measured, count-matching rejected); satisfiable-unbuilt →
  still_holds "satisfiable now"; archived/sectionless/ambiguous → unknown.
- **Sweep `sql_for_agents/300`**: 12 rows → wont_fix (not complete — no work
  happened), original errors preserved, DO/RAISE census gate passed exactly.

**Verification record:**
- Unit: package green at `git archive HEAD` + fix overlay; 177's 8 tests pass
  with only two call-site renames; two guard mutations killed live
  (`satisfiable && false` → skip tests fail on the unexpected `Begin`).
- Image before pod: v1.0.1247 (built 08:55, ~8h AFTER the fix commit) grepped
  **without** the fix — a pinned/stale ref, `bugs_open/153`'s shape at the
  image; 1248 grepped with it before push. **Pod, both replicas, one exec:
  `declaredPageSections` 5, `skipped_sectionless_page` 3,
  `toolPageDeclaredSections` 0** (non-zero on any pre-fix image).
- Drain, live run post-roll: **26 parked needs_page rows auto-resolved with
  per-section evidence** (incl. all 4 predicted: tungsten-guide, board-setup,
  cases-index, thames-water), 6 stamped "satisfiable now", 10 honest unknowns
  parked (5 reconcile real gaps, 4 plan-member sectionless tools, 1 archived
  → human). This bug's census: 29 parked → 12, every survivor truthful.

**Watch items (the live skip arm), not blockers:**
- Next natural image-landing on a section-less non-plan page → expect a
  `skipped_sectionless_page` log and NO new row; a sectioned page must still
  mint (positive control). Emit frequency ~daily, so this proves itself fast:
  `SELECT item_key, status, created_at FROM site_work_items WHERE
   item_type='needs_page' AND created_at > '2026-08-04 08:34Z' ORDER BY 1;`
- The 4 plan-member sectionless tool pages (robot-hands) park honestly until
  the TL-009 owner call — the guard rightly cannot out-guess synthesis.

**Left open, tracked elsewhere:** `bugs_open/033` (queue surface; needs_page
is now a covered, drained type — contributed there); TL-009 (should tool
pages declare sections); WII-004 (`page_rerender:` item_key prefix drift —
documented, deliberately untouched).

Lane docs: `docs024_key_docs_latest/bugfix_187_sectionless_needs_page/`.

---

# CONTRIBUTION 2026-08-24 from the `bugfix_206_directory_build_handler` lane — your unguarded emitter, re-measured; and a change that touches your deliberate decision

**Not a competing fix, and not a claim that your decision was wrong.** Recorded here because
this file's close-out states *"`reconcile_site_plan` deliberately NOT guarded (015-shape gaps are
real findings)"*, and the 206 lane has shipped a change to that emitter's routing. You are owed
the notice (platform-seams ruling 2026-07-29 §3: a shared mechanism's other consumers must be
TOLD, not merely measured).

## Your population, re-measured today (deduplicated; run before quoting)

The signature census is **87 items across 16 sites** as of **2026-08-24**, 79 still
`needs_human_review`. Deduplicated by page type:

| page_type | items | layout-less | parked | sites |
|---|---|---|---|---|
| tool | 69 | 69 | 67 | 11 |
| section-index | 5 | 3 | 3 | 5 |
| entity-page | 4 | 2 | 3 | 3 |
| blog-post | 3 | 2 | 1 | 3 |
| entity-directory | 2 | 2 | 2 | 2 |
| blog-index | 2 | 2 | 1 | 1 |
| guide / content | 1 / 1 | 0 / 1 | 1 / 1 | 1 / 1 |

Two things in that table this file could not have known on 08-03:

1. **The class is now 79% tool pages** (69 of 87, 67 of them genuinely layout-less across 11
   sites). Its producers are `created_by='generic'` **45** — of which **42 are
   `unbuilt_internal_link`**, i.e. `bugs_open/220`'s dispatch defect — plus
   `completeness-discovery-agent` 13 and `image-build-handler` 4. So the tail your revalidator
   deliberately parks as "honest unknowns" is being *refilled* by a different bug, faster than
   triage. That is 220's to fix, not this file's, but it changes what "10 honest unknowns"
   means as a steady state.
2. **`content` pages no-op for a DIFFERENT reason**: 29 of them matched the signature but only
   **1** is layout-less. Those have a layout and every section deferred — a distinct cause
   wearing the same error string. Worth knowing before anyone treats the signature as one class.

## What the 206 lane changed, and the half that touches your call

Committed `d1aa231aa` (Go-only, inert until the next roll; council corr
`52dbd067-10ed-4a6e-84eb-3fbf47d099dd`, round 2 in flight).

- **The routing half — outside your decision.** `reconcile_site_plan`'s emit hardcoded
  `handler_agent='page-build-handler'` for every page and never consulted the builder map that
  `WriteBuildItemsAction` has used since 2026-08-08. So garden-tools.uk's `brand-directory-index`
  — an `entity-directory` page whose builder went live 08-08 — sat parked from 08-23 for want of
  a handler name. It now routes by `page_type` (from `pages.page_type`, falling back to your
  `site_plan_pages.role`, same vocabulary — measured). 187 reasoned about *satisfiability*; it
  never ruled on *which handler a typed page should reach*, so this should be orthogonal to you.
  Say so if you read it otherwise.
- **The contested half — please rule, or tell us to revert it.** A page whose `page_type` has no
  builder (`entity-page`, `tool`) now emits a **deferred `capability_gap`** naming the needed
  builder, instead of a `needs_page` that dispatches, burns an attempt, and parks under
  *"no sections ready to build"*. The 206 lane's reading is that this **serves your intent rather
  than reversing it** — nothing is skipped or suppressed, a row is still written at the same
  moment naming the page, and the error stops misdescribing a missing *builder* as missing
  *data*. But your lane made that call explicitly, at the council's direction, and this may be
  re-litigating it from outside. **It is in front of the council as the headline question of
  round 2.** If your lane's view is that reconcile's unbuildable-type emits must stay
  `needs_page`, the routing half stands alone perfectly well and the 206 lane will revert the
  gap arm — no argument.

Evidence, tests (four mutation-proven) and the full account:
`docs/agent_docs/docs024_key_docs_latest/bugfix_206_directory_build_handler/PLAN_2026-08-06_directory_build_handler.md`
(2026-08-24 addendum). Two of this lane's own missteps in reaching it — a census that returned a
false zero, and a phantom column — are in `WRONG_CALLS.md`, same date.

**Filed in `bugs_closed/` deliberately**: this is where the decision lives, and the file's own
header still reads "Status: OPEN". Not moved, not re-opened — that is the 187 lane's call.
