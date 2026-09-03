# HANDOFF — portfolio_positioning — 2026-09-03 (evening). **START HERE.**

Supersedes `HANDOFF_2026-09-02_continue_here.md` (banner added there). Owner read-out:
`SUMMARY_2026-09-03_the_brief_that_two_agents_could_not_read.md`. Every count carries its date.

## 0. STATE IN ONE PARAGRAPH

Five remakes exist. Four are live from 2026-09-02 (advertise, websitepromotion, seotools,
designblog). The fifth, **copyonline.co.uk, is NOT live and is STALLED** one step before its page
plan, with no composition and nothing queued to make one — and **nothing of copyonline has ever been
published** (`sites.publish_target/published_at/last_deployed_at` all NULL, `build_status` `pending`,
the domain still serving the owner's old Drupal 7 install). The eight planned tool pages are built,
repaired and serving. The afternoon's real finding is `bugs_open/453`: a brief-writer `mission_brief`
is invisible to the classifier and the planner, which both render
`{{.site_specs.specs.mission_brief.text}}` — a child brief-writer output does not carry — while
`domain-strategist` renders `{{.site_specs}}` and reads it fine. That third agent is why copyonline's
strategy is CORRECT despite the fault. **Nothing was applied to the running build; the owner's
instruction not to disturb it stands and was the right call.**

## 1. FIRST TASKS

1a. **Is copyonline still stalled?** (2 min)
```sql
SELECT (SELECT count(*) FROM site_specs WHERE site_id='3d965325-519a-4515-b79f-50c886954a80'
          AND aspect='resolved_composition' AND is_current) AS has_composition,
       (SELECT count(*) FROM site_work_items WHERE site_id='3d965325-519a-4515-b79f-50c886954a80'
          AND status NOT IN ('complete','cancelled','rejected','failed')) AS open_items;
```
⚠ **`failed` is NOT terminal in the codebase's list** — omit it from that filter and an "open work"
query silently returns only failures. Bit me today; RUNBOOK §9.
If `has_composition` is still 0 and no `needs_composition` is open, the site needs its composition
step re-filed. **Read the diagnosis first (1b) before deciding how.**

1b. **Read the diagnosis I filed at 18:22Z.** Correlation `ef0ec49c-ff6f-4566-a344-db2bf590c619`.
```sql
SELECT status, spec->>'dispatch_correlation_id', left(result::text,2000) FROM site_work_items
 WHERE item_type='needs_diagnosis' AND spec->>'dispatch_correlation_id'='ef0ec49c-ff6f-4566-a344-db2bf590c619';
```
⚠ **The orchestration row is NOT joinable on `collected_data->'input_data'->>'fix_correlation_id'`**
for this run type — that query returns rows briefly and then zero, which reads as a dropped dispatch
and is not one. Find the run by its symptom text instead:
```sql
SELECT current_step, status, updated_at FROM orchestration_states
 WHERE collected_data->'input_data'->>'symptom' ILIKE '%needs_composition%' ORDER BY updated_at DESC;
```
Confirmed alive and at step `route` at 18:31:15Z, ~9 minutes after filing.

Symptom filed: a `needs_composition` item whose required specs are absent records
`validated_inputs.ready=false`, spawns a classifier, reaches terminal `complete`, and nothing files a
replacement when those specs later arrive. **I deliberately asserted no cause** — if the verdict
REFUTES it, that is a success and it goes in `WRONG_CALLS.md`, not a retry.

1c. **Judge the plan if one appeared.** The planner renders `{{toJSON .site_specs.specs.strategy}}`,
and copyonline's strategy is faithful to the owner's brief (`site_type: authority-portal`, the brief's
four tools by name, the lead route as the single converting page, the directory, randomised listings,
the webdesign.uk referral). So a plan produced now should be broadly right **even though the planner
cannot read the brief directly.** It also renders the WRONG classification, so check the plan against
`BRIEF_2026-09-03c_copyonline_co_uk_REV3_for_review.md` rather than assuming.

## 2. THE TWO DECISIONS THE OWNER STILL OWES

- **Retire or keep three tool pages that duplicate seotools** — `serp-snippet-previewer`,
  `title-tag-scorer`, `keyword-intent-classifier`. ⚠ **These are now BUILT and deployed** (16:22:43Z,
  16:34:42Z, 16:39:58Z on 2026-09-03). I told him they were unbuilt and cheap to cancel; they shipped
  while the question sat. Not published, so retiring them is an internal tidy.
- **Where the lead route finally points.** He ruled "organisations in preference to myself, but don't
  change things if it's already running". `apply_r4.sql` was prepared and **correctly refused by its
  own guard**. Apply at the first cheap window, or fold into a rebuild.

## 3. WHAT IS OWED TO OTHER LANES (all delivered — do not redo)

- **`bugs_open/453`** — three CONTRIBs plus corrections. The last one is the useful one: the fix is
  **four template expressions** (`mission_brief.text` and `roadmap_brief.text`, in `build-site-planner`
  and `domain-research-classifier`) and that is the COMPLETE fleet-wide blast radius `[MEASURED
  2026-09-03 18:22Z]`. `roadmap_brief` is latent, not live (4 of 4 current specs carry the key;
  `mission_brief` is 7 of 23 without, which is the non-zero control proving the predicate discriminates).
  **Verify any fix at `mission_brief`** — a `roadmap_brief` test passes before and after and proves nothing.
- **`bugfix_445_layout_fit`** — CONTRIB plus addendum: copyonline has no `resolved_composition`, and its
  `industry_tags` came from the blind classifier so they must not enter their tag simulation. Includes
  the one-column admission check and the seven sites that fail it.
- **`static_site_form_endpoint`** — CONTRIB naming copyonline's lead route as a first customer.
- **17 unbuilt remakes** contributed to the 445 lane as a forward-looking tag-simulation population.

## 4. TRAPS THIS SESSION PAID FOR (all in RUNBOOK §7-9, LANDMINES and WRONG_CALLS)

1. **`pages.deployed_at` is page-level and says nothing about publication.** Ten pages carried it while
   the site had never been published. Probe served bytes with an **invented-URL control** (a parked
   domain 200s every path) and read the `sites` row, never the `pages` rows.
2. **Enumerating the readers of a FIELD is not enumerating the readers of the INFORMATION.** I measured
   two agents that cannot read `mission_brief.text`, then told the owner none of them could, and
   recommended interrupting his build. A third reads it via a whole-object render. RUNBOOK §7 has the
   query, including the capture-group and active-filter gotchas that made it silently wrong twice.
3. **A question put to a human does not pause the pipeline.** See §2's first bullet.
4. **An LLM step's own `reasoning` field is testimony about its inputs** — the classifier wrote "no
   mission brief was supplied" unprompted. Cheaper than any prompt-render harness, and it is the
   behavioural pass/fail after a fix. But recall is ~1 in 7: confirm with it, never survey.
5. **Writing a `mission_brief` copies a shape the consumers cannot read.** Three independent producers
   reproduced it, including me, twice, while diagnosing this very bug. LANDMINES has it, plus a
   correction entry — read BOTH; the first asserts "nothing reads the structured object" and that is false.

## 5. HELD AND NOT TO BE APPLIED

`SQL_2026-09-03c_make_briefs_visible_to_the_classifier_and_planner_HOLD.sql` — a per-site data
workaround (adds a `text` key). **Do not apply it** if 453's template fix ships; it repairs one site,
leaves the trap armed, and the real fix is smaller and fleet-wide. Kept because it is evidenced and
because its footer carries the corrected two-step reasoning.

## 6. STILL TRUE FROM THE PREVIOUS HANDOFF

`bugs_open/444` (listing pages ship empty with brief-echo prose) and the theme-kit differentiation
levers still gate the remaining briefs. Sitemap machinery (642) remains self-maintaining. The
chrome-pin experiment stays RE-SCOPED (37 of 39 sites supply zero header keys) and the
"prefer the nine unused layouts" fire-direction was RETRACTED — do not reinstate either.
