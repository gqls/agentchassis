# HANDOFF — 2026-08-25b. **START HERE.** `bugs_open/391` — CTA destinations: work IN FLIGHT, 22 items dispatched, retirement still pending

> **⚠ SUPERSEDED 2026-08-25 by `HANDOFF_2026-08-25c_continue_here.md`.** Kept for the reasoning trail.
> Two things in this file are now wrong: §2 presents `grep -c '<p'` as *the* prose control (it read
> 17 → 20 on a page whose copy had been destroyed — use the distinctness query in 25c §3 instead),
> and §2/§4 describe 22 items as in flight (they resolved: 11 of 12 pages verified, the 12th retried).

**Supersedes `HANDOFF_2026-08-25_continue_here.md`** (keep it for the reasoning trail; this file is
the current state). Read this, then `bugs_open/391` from the bottom — its last three sections are
corrections that reverse earlier claims in the same file.

> **Deploy facts have a shelf life of hours — re-probe, do not quote.** Chassis at handoff:
> **`a7459a44b68b8c67b7d7bb0ca7c064e0729d59f5`**, pods up `2026-08-25 19:07Z`. Capability re-probed
> on the running binary, with its absent-control: `rendered_html_transform` 8,
> `code_span_to_code_tag` 5, `cta_links_stale` 3, control **0**. Both today's lane commits are
> ancestors of that stamp.

---

## 0. THE BUG IN ONE PARAGRAPH

`chooseCTATargets` (`resolve_internal_links_action.go:651`) picks a site's primary CTA by sorting
every `tool`/`game` page on `COALESCE(nav_order,100)` then `name` and taking `[0]` — **no topic, tag
or semantic input at all.** A password-strength toy carried the fossil value `nav_order = 1` on three
sites (set at page creation, 2026-03-13) and so won the primary button on every page. Worse, the
framework then writes button copy **naming** whatever it picked (`stampCTADestinationGuidance:362`
feeds the destination title into the writer's spec for the label field), so a wrong pick
**locks itself in**: the next resolve label-matches the copy back to the same page.

## 1. OWNER DECISIONS — all five answered 2026-08-25

| # | answer | state |
|---|---|---|
| 1 | tool "can disappear everywhere" — **but the shared library component STAYS** (`tool-password-entropy` remains `is_active=true`, available to new sites) | retirement of the **three site pages** still TO DO, sequenced 3rd |
| 2 | yes, correct the menu-order numbers | ✅ **DONE** — 1 → 900 on three rows |
| 3 | yes, build the platform lever | **NOT STARTED** — the only design work left |
| 4 | repair: "whatever you suggest" | sequenced last, verify at served bytes |
| 5 | re-scope the commission | ✅ done — ~20 fields by query, not 16 sites |

## 2. ⏳ IN FLIGHT RIGHT NOW — 22 items, dependency-ordered, no babysitting needed

`SQL_2026-08-25_step2_remaining_11_pairs.sql`, applied 2026-08-25 ~19:2xZ. Eleven
`content_rewrite` + eleven `page_rerender`, **each relink carrying
`depends_on = ARRAY[<its rewrite id>]`.**

**Verified with the dispatcher's own predicate at dispatch time: 11 rewrites eligible, 0 relinks
eligible.** `load_work_item_actions.go:713` refuses a row whose `depends_on` is not yet
`complete`/`verified`, so **a page's relink cannot be served before its own rewrite lands** and the
eleven pages progress independently. That is what makes a single batch safe — see §5 for why
batching them any other way is not.

**Check progress:**
```sql
SELECT item_type, status, count(*) FROM site_work_items
WHERE created_by='bugfix_391_cta_relevance' GROUP BY 1,2 ORDER BY 1,2;
```
Budget **~25–35 min per item** (measured: `content_rewrite` claimed after 24 min; `page_rerender`
run ~30 min after creation). **Both are normal. DO NOT bypass the queue** — see §6.

**Then verify each page as a MATCHED PAIR at the served bytes**, never by work-item status
(`bugs_open/389`: a `cta_links_stale` rerender reports `complete` whether or not any CTA moved):
```bash
curl -s "https://<domain>/<page>.html?cb=$(date +%s)" > after.html
grep -c 'password-entropy' after.html              # must be 0
grep -o 'href="[^"]*"[^>]*>[^<]*</a>' after.html   # label and href must name the SAME tool
grep -c '<p' after.html                            # PROSE CONTROL — must equal the before count
```
The prose control is the one that catches an over-reaching rewrite. On the canary it held at 15/15.

## 3. WHAT IS DONE AND PROVEN

- **`nav_order` demoted** 1 → 900 on the three sites (`SQL_2026-08-25_demote_password_entropy_nav_order.sql`).
  ⚠ **900, not 200:** at 200 it ties with the sites' other tools and the tiebreak is alphabetical on
  `name` — `password-entropy` precedes every `tool-*`, so it would still have won on two of three.
- **Canary page COMPLETE and verified at the served bytes**: `finetuning.uk/technical-details`,
  `password-entropy` refs **2 → 0**, both buttons now name *and* link the Fine-Tuning vs RAG vs
  Prompting Decision Guide (`/tools/model-approach-selector.html`, 200), prose control **15 → 15**.

## 4. WHAT REMAINS, IN ORDER — the ordering is load-bearing

1. **(in flight)** the 11 label-locked pages — 18 fields.
2. **Retire the three tool pages** (owner decision 1), **with the footer entry and the three
   `/tools.html` listings in the same operation**, through the framework (2026-08-04 ruling), never
   hand-edits. Blast radius measured 2026-08-25: **91** `page_components` refs
   (`content_data` **and** `rendered_html`; 45/25/21), 1 footer (`ai-agent-orchestration.com`), 3
   live listings, 0 visible nav. **Re-measure before acting — step 1 is reducing it now.**
3. **Re-resolve the 60 label-less fields** (44 pages) — `cta_links_stale` rerender, **no LLM**.
   ⚠ **This CANNOT be brought forward.** `applyCTARecompute`'s **KEEP #2**
   (`rerender_page_sections_action.go:1114`) returns early for any stored destination that is a valid
   page and not the page itself — so while `/tools/password-entropy.html` exists it is **kept**, and
   the positional pick that `nav_order` governs is never reached. Only retirement makes it invalid.
   Verified that retirement then unblocks rather than freezes: KEEP #1 needs utility-area, KEEP #3
   needs `IsAuthoredNonPageCTADestination` (`links_tel.go:36` — true only for
   `tel:`/`mailto:`/`http(s)://`/`//`/`#frag`), and a relative `/tools/…` path matches neither, so
   control reaches the positional write. **The demotion in §3 is what makes that write correct.**
4. **Decision 3 — the platform lever.** Candidate 1 (an explicit `eligible_as_cta_target` opt-out)
   **paired with** candidate 4 (a detector for the anomalous-`nav_order` shape). Three constraints,
   all from review:
   - **read at the RANKING, not the loaders** — `render_site_components_action.go:182-190` (the site
     **header** CTA fallback) calls the loaders directly, takes `ordered[0]`, and its output is
     **never persisted** (`site_components` holds 0 `cta_url` keys), so a loader change moves every
     site's header button with no `content_data` diff to show it;
   - **it must also bind `LoadCTALabelUniverse`**, or the opt-out has a hole exactly the shape of
     this bug (the label match runs first);
   - **engage RFC_022 and enumerate the consumers** before booking a council round — asserting the
     opt-in shape without the query is itself the objection.
5. **Final sweep:** zero `password-entropy` references across the three sites, at the served bytes.

## 5. Why the pairs are dependency-linked rather than run in two waves

Between a page's rewrite and its relink, the page serves a button whose **text names one tool and
whose href points at another** — `bugs_closed/299`'s defect, manufactured by our own repair. On the
canary that window was ~32 minutes. Dispatching eleven rewrites and then eleven relinks would put
**all eleven pages in that state at once**; `depends_on` keeps the window per-page and unattended.

## 6. ⚠ Two traps this lane already fell into — do not repeat them

- **I called the queue stalled and bypassed it 5 seconds after its own run had started.** The queue
  claimed the item at 13:37:22; my direct fire began 13:37:27. The queue did the work; mine was a
  redundant duplicate write against a live page, harmless only because a CTA recompute is idempotent.
  **Before calling this queue dead, measure its service interval, not your row's age** — and note I
  had already measured 593 `page_rerender` completions in six hours and argued past it. Full account:
  NOTES MISSTEP 9 + `WRONG_CALLS.md` 2026-08-25.
- **`spec.suggestion`, not `spec.content_guidance`.** `suggestion` is the key the handler reads;
  `content_guidance` is only *aliased* into it (`bugs_open/271`) and an author-supplied `suggestion`
  wins. The inherited lane RUNBOOK says the wrong one; a live completed item said the right one.
- And the precondition that could sink a rerender: **if ANY section has `content_data IS NULL` the
  whole page escalates to the content writer and the copy IS regenerated.** Check with `IS NULL`,
  not "is empty" — they are different tests.

## 7. Session-start checklist
`git log --oneline -10` · re-read this from disk · `scripts/who-owns.py 391` · chassis stamp +
capability probe **with an absent-control** · the progress query in §2 · then §4.
