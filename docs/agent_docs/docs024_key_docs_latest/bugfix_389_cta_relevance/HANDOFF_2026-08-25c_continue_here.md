# HANDOFF — 2026-08-25c. **START HERE.** `bugs_open/391` — step 2 is 11 of 12 done and verified live; retirement is the next real work

**Supersedes `HANDOFF_2026-08-25b_continue_here.md`** (keep it for the reasoning trail). Read this,
then `bugs_open/391` from the bottom — its last section is the step 2 result and it amends the
verification recipe that file previously gave.

> **Deploy facts have a shelf life of hours — re-probe, do not quote.** Chassis at handoff:
> `a7459a44b68b8c67b7d7bb0ca7c064e0729d59f5`, pods up `2026-08-25 19:07Z`. Re-probe with an
> absent-control before trusting it. Nothing in this handoff depends on a new build.

---

## 0. THE BUG IN ONE PARAGRAPH (unchanged)

`chooseCTATargets` (`resolve_internal_links_action.go:651`) picks a site's primary CTA by sorting
every `tool`/`game` page on `COALESCE(nav_order,100)` then `name` and taking `[0]` — **no topic, tag
or semantic input at all.** A password-strength toy carried the fossil value `nav_order = 1` on three
sites and so won the primary button on every page. The framework then writes button copy **naming**
whatever it picked (`stampCTADestinationGuidance:362`), so a wrong pick **locks itself in**.

## 1. STATE — what is done, what is in flight, what is next

| # | work | state |
|---|---|---|
| 1 | `nav_order` demoted 1 → 900 on three sites | ✅ done + verified |
| 2 | the 12 label-locked pages (canary + 11) | ✅ **11 of 12 verified at the served bytes**; the 12th is in flight, below |
| 3 | **retire the three tool pages** + footer + 3 `/tools.html` listings | ⏳ **NOT STARTED — this is the next real work** |
| 4 | re-resolve the 60 label-less fields (44 pages) | blocked on 3, by design |
| 5 | the platform lever (owner decision 3) | not started; design notes in §5 |

## 2. ⏳ IN FLIGHT — one page only

`SQL_2026-08-25_retry_model_directory_pair.sql`, dispatched 2026-08-25 ~20:44Z: one
`content_rewrite` + one `page_rerender` for `ai-agent-orchestration.com/model-directory.html`, the
relink carrying `depends_on` (proven armed by a `DO`/`RAISE` in the same transaction).

```sql
SELECT item_type, status, error FROM site_work_items
WHERE item_key LIKE '%retry:2c7c836c%' ORDER BY item_type;
```

Budget **~25–35 min per item**. **DO NOT bypass the queue** — see §6.

**Why the first attempt failed, because it will look like our bug and is not.** It was refused at
`validate_content` for `unregistered_number "150"` — the CTA `<h2>` says *"More than 150 agents are
listed here"*, that sentence **was already live before this lane touched the page**, and it is false:
the page's own `/data/model-directory-full.json` reports `"count": 30`. The gate is correct. The
retry asks for a heading with **no count in it at all**. Routed to the owning lane as
`model_directory_pipeline/CONTRIB_2026-08-25_from_bugfix_391_the_directory_page_claims_150_and_lists_30.md`.

> **Where the real error message lives.** The work item's `error` and the orchestration's
> `collected_data->'__step_errors'` both say only *"0 blockers, 1 errors"*, and the orchestration row
> reads `COMPLETED` with `error` NULL. The detail is in a third place:
> ```sql
> SELECT jsonb_pretty(context) FROM agent_error_log
> WHERE work_item_id='<item>' AND error_code='CONTENT_VALIDATION_BLOCKER_DETAIL';
> ```

## 3. ⚠ THE VERIFICATION RECIPE HAS CHANGED — the old one nearly passed a destroyed page

Verify each page **as a matched pair at the served bytes**, never by work-item status. Three checks,
and **the distinctness one is now the load-bearing one**:

```bash
curl -s "https://<domain>/<page>.html?cb=$(date +%s)" > after.html
grep -c 'password-entropy' after.html               # 0 on finetuning.uk / leopardess;
                                                    # on ai-agent-orchestration.com expect 2 (FOOTER)
                                                    # and 3 on /tools.html (footer + listing card)
grep -oE '<a [^>]*class="[^"]*(btn|cta)[^"]*"[^>]*>[^<]*</a>' after.html   # label and href must name the SAME tool
```

```sql
-- THE CONTROL THAT ACTUALLY WORKS: components vs distinct openings, per page.
-- components > distinct  ⇒  the rewrite replaced a section with a copy of a neighbour.
SELECT page_id, count(*) AS components,
       count(DISTINCT left(regexp_replace(regexp_replace(rendered_html,'<[^>]*>',' ','g'),
                                          '\s+',' ','g'), 80)) AS distinct_openings
FROM page_components WHERE page_id = ANY(<pages>) GROUP BY 1;
```

> **⚠ `grep -c '<p'` is NOT a sufficient prose control and the previous handoff was wrong to present
> it as one.** On the page this repair destroyed it read **17 → 20** — a `+3` that looks like a
> writer adding a sentence. The damage was **duplication**, and duplication does not move a count of
> anything; it moves distinctness. The count had been trusted because it held at 15/15 on the canary
> — but the canary was undamaged, so it had never been shown to discriminate. Keep it as a secondary
> signal; do not sign a page off on it.

**And probe every CTA target with a per-domain absent control** (`/tools/this-page-does-not-exist-391.html`
→ must 404). A parked-domain catch-all 200s every path and would score every href as live.

## 4. ⚠ WHAT THIS REPAIR IS CAPABLE OF BREAKING — read before dispatching another `content_rewrite`

A `content_rewrite` commissioned for **labels only** rewrote the page **body** on one of eleven pages
and destroyed two authored sections by overwriting them with copies of a third. The spec text
*"Reword ONLY the … LABELS. Leave all other prose exactly as it is"* did not bound the write —
**prompt text is not a control.**

- Restored from the offending write's own archive (`page_component_history.source_item_id`), by
  subquery, nothing retyped: `SQL_2026-08-25_restore_your_own_model_blocks.sql` +
  `SQL_2026-08-25_rerender_after_restore_your_own_model.sql`. Verified back at the served bytes.
- **Restore `content_data` only** and let a `page_rerender` regenerate `rendered_html`.
- **Guard the restore with `DO`/`RAISE`, and run the guard against the damaged state FIRST** — a bare
  `SELECT` cannot abort a `COMMIT`, and a guard that has not failed is not evidence when it passes.
- **Refuse if any section has `content_data IS NULL`** — that escalates the next rerender to the
  content writer, which regenerates the copy and silently undoes the restore.
- Same family as `bugs_open/403` (owned by the leopardess lane); CONTRIB filed there with the new
  shape and the detector query. `LANDMINES.md` has the prospective check; `WRONG_CALLS.md` the incident.

## 5. NEXT: retirement (step 3), and the ordering is still load-bearing

1. **Retire the three tool pages** (owner decision 1 — the shared library component
   `tool-password-entropy` **STAYS** `is_active=true`), **with the footer entry and the three
   `/tools.html` listings in the same operation**, through the framework (2026-08-04 ruling), never
   hand-edits. Blast radius measured 2026-08-25 **before** step 2 ran: **91** `page_components` refs,
   1 footer (`ai-agent-orchestration.com`), 3 live listings, 0 visible nav. **RE-MEASURE — step 2 has
   just reduced it**, and a count carries the date it was counted.
2. **Then** re-resolve the 60 label-less fields (44 pages) — `cta_links_stale`, no LLM.
   **This cannot be brought forward.** `applyCTARecompute`'s KEEP #2
   (`rerender_page_sections_action.go:1114`) returns early for any stored destination that is a valid
   page, so while `/tools/password-entropy.html` exists it is **kept** and the positional pick is
   never reached. Only retirement makes it invalid; KEEP #1 (utility-area) and KEEP #3
   (`IsAuthoredNonPageCTADestination`, `links_tel.go:36`) do not catch a relative `/tools/…` path, so
   control then reaches the positional write — which the `nav_order` demotion makes correct.
3. **Decision 3 — the platform lever.** Candidate 1 (explicit `eligible_as_cta_target` opt-out)
   **paired with** candidate 4 (a detector for the anomalous-`nav_order` shape). Three constraints,
   all from review: read at the **RANKING**, not the loaders (`render_site_components_action.go:182-190`
   calls the loaders directly, takes `ordered[0]`, and its output is never persisted); it must also
   bind `LoadCTALabelUniverse` or the opt-out has a hole exactly the shape of this bug; and **engage
   RFC_022 and enumerate the consumers** — asserting the opt-in shape without the query is itself the
   objection.
4. **Final sweep:** zero `password-entropy` references across the three sites at the served bytes —
   which after retirement includes the footer and the listing cards, currently the only residue.

## 6. ⚠ Traps this lane already fell into — do not repeat them

- **I called the queue stalled and bypassed it 5 seconds after its own run had started.** Before
  calling this queue dead, measure its **service interval**, not your row's age. Full account: NOTES
  MISSTEP 9 + `WRONG_CALLS.md`.
- **`spec.suggestion`, not `spec.content_guidance`** — `suggestion` is the key the handler reads.
- **`spec.page_name` is load-bearing on a rerender** — without it the rerender discards its own
  result (`sections_saved: 0, success: true`) and deploys the stale assembly.
- **A churn/word-count figure is a comparison against a baseline you chose.** Mine ranked
  `leopardessconsulting.co.uk/services.html` as the worst-damaged page; it had been damaged the day
  before by someone else, and a concurrent session restored it **while I was measuring**. On a shared
  estate, re-read before writing a number down, and say when you read it.
- **Item keys dedup in ANY status** (`bugs_open/326`) — a retry needs a fresh `item_key`.

## 7. Session-start checklist
`git log --oneline -10` · re-read this from disk · `scripts/who-owns.py 391` · chassis stamp +
capability probe **with an absent-control** · the in-flight query in §2 · then §5.
