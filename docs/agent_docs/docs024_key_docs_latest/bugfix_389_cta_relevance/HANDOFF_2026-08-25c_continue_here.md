# HANDOFF — 2026-08-25c. **START HERE.** `bugs_open/391` — step 2 is 11 of 12 done and verified live; retirement is the next real work

> **⚠ SUPERSEDED 2026-08-26 by `HANDOFF_2026-08-26_continue_here.md`.** Kept for the reasoning trail.

**Supersedes `HANDOFF_2026-08-25b_continue_here.md`** (keep it for the reasoning trail). Read this,
then `bugs_open/391` from the bottom — its last section is the step 2 result and it amends the
verification recipe that file previously gave.

> **Deploy facts have a shelf life of hours — re-probe, do not quote.** Chassis at handoff:
> `a7459a44b68b8c67b7d7bb0ca7c064e0729d59f5`, pods up `2026-08-25 19:07Z`. Re-probe with an
> absent-control before trusting it. Nothing in this handoff depends on a new build.

---

## 0a. FLEET LLM OUTAGE — **RESOLVED 2026-08-26 ~09:00Z**, kept for the residue it left

> **RESOLVED.** Credit restored; verified at the artefact — the 09:00 hour ran **124 calls, 124
> ok, 0 failed**. The residue below still stands and still needs re-firing.

`[MEASURED 2026-08-26 08:5xZ]` The Anthropic account ran out of credit at ~2026-08-25 23:46Z.
**Zero successful LLM calls across 00:00–08:00**, 691 attempts, 690 of them verbatim *"Your credit
balance is too low to access the Anthropic API"* (controls: 157 and 124 successes in the 21:00 and
22:00 hours). Owner notified by the `loanzy_uk_example_site` lane.

- **Not blocked:** `cta_links_stale` rerenders (explicitly no-LLM), and all measurement/SQL work.
- **Residue that survives the top-up:** **21** work items fleet-wide have exhausted `max_attempts`
  against the dead API and will **not** self-heal — 7 `unbuilt_internal_link`, 4 `content_rewrite`,
  4 `improve_tool`, 2 `dead_fragment_link`, 2 `needs_page`, 1 `needs_diagnosis`, 1 `page_rerender`
  (20 of 21 cite the credit error; the `page_rerender` is a different failure that landed in the
  window). None are this lane's. Re-fire recipe lives in the `loanzy_uk_example_site` RUNBOOK.

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
| 2 | the 12 label-locked pages (canary + 11) | ✅ **COMPLETE and verified** — the lock query returns **0 rows fleet-wide** (§2a). ⚠ **2 of the 12 lost authored copy to the repair itself** — both restored; one publish may still be in flight (§2) |
| 3 | **retire the three tool pages** + footer + 3 `/tools.html` listings | ⏳ **NOT STARTED — this is the next real work** |
| 4 | re-resolve the 60 label-less fields (44 pages) | blocked on 3, by design |
| 5 | the platform lever (owner decision 3) | not started; design notes in §5 |

## 2. ⏳ IN FLIGHT — one re-render, and it is a CONTENT restore, not CTA work

**The CTA work is done.** `/model-directory.html` completed 2026-08-25 21:00:09Z and verified at the
served bytes (`password-entropy` 4 → 2, both survivors in the footer; the refused claim gone; four
anchors whose labels and hrefs agree; targets 200 against a 404 control; 3 components / 3 distinct).

**Still in flight:** `content_restore_rerender:a32b8822-…` — the re-render that publishes the two
restored text blocks on `finetuning.uk/technical-details.html`. The DB is already correct; until this
lands the page still SERVES the same licence section three times. Verify with §3's distinctness
control **and** at the served bytes:

```bash
curl -s "https://finetuning.uk/technical-details.html?cb=$(date +%s)" > td.html
grep -c 'The base model itself is a small open-weight model'      # must be 1
grep -c 'meaning the underlying weights are published'            # must be 1
grep -c 'meaning the company that built it has published'         # must be 1
```

```sql
SELECT item_key, item_type, status, error FROM site_work_items
WHERE item_key LIKE '%retry:2c7c836c%'
   OR item_key LIKE 'content_restore_rerender:%'
ORDER BY item_key;
```

**Three items, two unrelated jobs.** The `retry:` pair is the CTA repair for `/model-directory.html`.
`content_restore_rerender:a32b8822-…` is the canary re-render that publishes the restored text blocks
on `finetuning.uk/technical-details.html` — until it lands, that page still SERVES the same licence
section three times, though the DB is already correct.

Budget **~25–35 min per item**. **DO NOT bypass the queue** — see §6.

**Why the first attempt failed — kept because it will look like our bug and is not.** It was refused at
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

## 2a. Step 2 is provably complete — by the lock query, not by the dispatch list

Do not take "we did the pages on the list" as coverage. Re-run the RUNBOOK's label-lock query
fleet-wide; it returns every `%_url` field pointing at the retiring tool **whose own label names
it**, which is the population a `nav_order` fix can never reach:

```sql
SELECT s.domain, p.url, kv.key,
       COALESCE(pc.content_data->>(replace(kv.key,'_url','_text')),
                pc.content_data->>(replace(kv.key,'_url','_label')),
                pc.content_data->>(replace(kv.key,'_url',''))) AS label
FROM page_components pc
JOIN pages p ON p.id=pc.page_id JOIN sites s ON s.id=p.site_id
CROSS JOIN LATERAL jsonb_each(pc.content_data) kv
WHERE kv.key LIKE '%\_url' AND kv.value #>> '{}' LIKE '%password-entropy%'
  AND COALESCE(pc.content_data->>(replace(kv.key,'_url','_text')),
               pc.content_data->>(replace(kv.key,'_url','_label')),
               pc.content_data->>(replace(kv.key,'_url',''))) ILIKE '%password%';
```

`[MEASURED 2026-08-25 ~21:00Z]` **0 rows fleet-wide.** It was **20 of 80** when this bug was filed,
and 2 an hour before this line was written. So the label-locked population is cleared, and
what is left is the label-**less** one: **31 `page_component` rows across ~22 pages** on
`ai-agent-orchestration.com` (`hero` / `call-to-action`), **2 in the `footer` `site_component`**, and
1 in the `/tools.html` `tool-list`. That is step 3/4 work and is blocked on retirement by KEEP #2 —
not a gap in step 2. **Re-run the query, do not quote the numbers.**

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
       count(DISTINCT md5(regexp_replace(regexp_replace(rendered_html,'<[^>]*>',' ','g'),
                                     '\s+',' ','g'))) AS distinct_sections
FROM page_components WHERE page_id = ANY(<pages>) GROUP BY 1;
```

> **⚠ `grep -c '<p'` is NOT a prose control at all, and the previous handoff was wrong to present it
> as one.** On one destroyed page it read **17 → 20** — a `+3` that looks like a writer adding a
> sentence. On the other, **the canary, it read 15 → 15**: it could not move, because three
> paragraphs had been replaced by three paragraphs. Duplication does not move a count of anything;
> it moves distinctness.
>
> **And the 15/15 is why the control was trusted.** It was validated on the canary and promoted to
> the batch on that basis — while the canary was itself damaged and the control was reporting clean.
> *A control checked only where you believe nothing went wrong has not been shown to discriminate;
> if the thing you checked it against turns out to have been broken, its green result is evidence
> AGAINST the control.* Keep the `<p>` count as a weak secondary signal; never sign a page off on it.

**And probe every CTA target with a per-domain absent control** (`/tools/this-page-does-not-exist-391.html`
→ must 404). A parked-domain catch-all 200s every path and would score every href as live.

## 4. ⚠ WHAT THIS REPAIR IS CAPABLE OF BREAKING — read before dispatching another `content_rewrite`

A `content_rewrite` commissioned for **labels only** rewrote the page **body** and destroyed
authored sections by overwriting them with copies of a sibling — on **2 of the 12 pages this lane
put through it (17%)**: `finetuning.uk/your-own-model.html` and `finetuning.uk/technical-details.html`
(the canary). Both restored. The spec text
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

## 5. NEXT: retirement — and it is THREE steps, not two. The 2-step order DEADLOCKS

> **⚠ CORRECTED 2026-08-26 after reading the retraction machinery.** This section previously read
> *retire the pages → then re-resolve the label-less fields*. **Those two are each other's
> preconditions**, so that order cannot start:
> - `retract_page_deployment` **REFUSES** any page something editorial still links to — by design
>   (`retract_page_graph.go`: *"INBOUND, editorial → REFUSE the retraction and name the referrers"*),
>   so that a retraction cannot create a dead link. Measured with the action's own quote-delimited
>   `href="<url>"` predicate `[MEASURED 2026-08-26]`: **61 rows across 47 active pages**, plus the
>   aiao footer `site_component`.
> - The re-resolution is blocked the other way by `applyCTARecompute`'s **KEEP #2**, which returns
>   early while the stored destination is in `validPages`.
>
> **It breaks on `validPages`.** `loadResolverPageSet` (`resolve_internal_links_action.go:964`)
> selects `WHERE status NOT IN ('deleted','archived')`. **Archiving alone** drops the page out of
> that set, KEEP #2 goes false, KEEP #3 cannot catch a relative `/tools/…` path, and control reaches
> the positional pick — which the `nav_order` 1 → 900 demotion already made correct. **Archive is
> the key that turns both locks**, and it needs no LLM and deletes nothing.
>
> **"Retire" is also two distinct operations.** `pages.status='archived'` is **hand-run SQL** — the
> action's own header: *"nothing in this codebase archives a page … there is no writer of
> `status='archived'` at all"*. Removing the FILE is `retract_page_deployment`, driven by the live
> **`page-retraction`** agent (config `site_id_field`, `page_ids_field`). **Archiving does NOT
> unpublish** — it freezes the last rendered HTML and keeps serving it, verified at the artefact
> (archived canary returns **200**, absent control **404**). So there is no 404 window between the
> steps, and the whole sequence is reversible until step 3.

1. **ARCHIVE** the three pages (SQL; reversible; the page keeps serving). Then **2. RE-RESOLVE**
   the inbound references (`cta_links_stale`, **no LLM**). Then **3. RETRACT** via the
   `page-retraction` agent, which will no longer refuse and which also deactivates the
   `site_nav_items` row itself (structural inbound is *mechanised*, editorial is *refused*,
   newly-stranded outbound is *reported*).

   Owner decision 1 — the shared library component **STAYS** active. ⚠ **There is NO
   `content_components` row named `tool-password-entropy`**, as this file used to say: the active
   library row is **`tool-password-entropy_pre_037`**, and the four per-site `tool-password-entropy-<domain>`
   rows are already `is_active=false`. A guard written to the old wording finds 0 and falsely aborts —
   assert on `count(*) WHERE name ILIKE '%password-entropy%' AND is_active = 1` instead.

   **Canary order, simplest first** (from the per-site table below): `finetuning.uk` →
   `leopardessconsulting.co.uk` → `ai-agent-orchestration.com` **last**, since only that one has a
   `site_nav_items` row and a footer link.

   ⚠ **Before any retraction run, read `LANDMINES.md` on the two siblings' opposite `dry_run`
   defaults** — omitting the key on `retract_page_deployment` **deletes**, where omitting it on
   `retract_asset_files` is a dry run.

   **Blast radius `[RE-MEASURED 2026-08-26 ~09:00Z]`** (the old **91** was taken before step 2 ran):
   **63** `page_components` rows over **49** pages — aiao 30/19, leopardess 20/19, finetuning 13/11.
   `content_data` and `rendered_html` agree row-for-row (0 / 0 / 63), so no stale render is hiding one.

   > **⚠ CORRECTED 2026-08-26: "with the footer entry … in the same operation" DESCRIBES AN EDIT THAT
   > DOES NOT EXIST.** The footer `site_components` row's `content_data` is
   > `{year, email, phone, domain, tagline, company_name}` — **no tool links at all**; the reference
   > lives only in `rendered_html`, and the served footer lists **exactly the six live tool pages in
   > `nav_order` order**, derived from the nav tables. **Retiring the page removes it by
   > construction.** The risk is the inverse: retire and fail to refresh, and every page on that site
   > serves a footer link to a dead page. Recipe: **`nav-link-fixer`** refreshes
   > `site_components.rendered_html` from the existing nav tables, then propagate in **assemble mode**
   > (`page-rerender` with **no** `spec.reason`) — worked script
   > `docs/leopardessconsulting/scripts/reconcile_footer_nav.sh`. Do **not** reach for the agent whose
   > name says navigation; it deletes every child-path link (LANDMINES, NAV-013).

   **The three sites do NOT retire alike** `[MEASURED 2026-08-26]` — only one has any nav presence:

   | domain | `in_header` | `in_footer` | `site_nav_items` | served refs |
   |---|---|---|---|---|
   | ai-agent-orchestration.com | **t** | f | **1** | 1 (in the FOOTER) |
   | finetuning.uk | f | f | 0 | 0 |
   | leopardessconsulting.co.uk | f | f | 0 | 0 |

   `in_header=true` yet no header link is **not** the flag lying: `classifyPagesForNav` demotes a
   child-path page declaring a nav flag into the **`utility`** group, which renders in the footer.
   Flag honoured, placement overridden. finetuning and leopardess need **no** nav work at all.
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
