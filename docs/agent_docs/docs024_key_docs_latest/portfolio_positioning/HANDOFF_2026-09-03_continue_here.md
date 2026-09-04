# HANDOFF — portfolio_positioning — 2026-09-03 (evening). **START HERE.**

Supersedes `HANDOFF_2026-09-02_continue_here.md` (banner added there). Owner read-out:
`SUMMARY_2026-09-03_the_brief_that_two_agents_could_not_read.md`. Every count carries its date.

## ⚠ 0-NOW. STATE AS OF 2026-09-04 ~11:45Z — START HERE (supersedes §0, §1w–§1z and §2 below; they stay for the record)

**copyonline.co.uk is PLANNED and BUILDING, from a sighted reading of the owner's brief.** Nothing is
public (`sites.published_at`/`publish_target` NULL, the domain still serves the old Drupal install).
Chain this morning: `needs_briefing` (filed by hand 07:47Z, item `479614c9`) → briefing spec 09:02:42Z →
`needs_site_plan` complete 11:09:35Z → **plan: 115 rows, 34 new pages (44 total, 38 active)**. The
plan follows the brief: `index` (AI-first hero), `ai-commercial-copy`, `writing-with-ai`,
`editing-ai-copy`, the four brief tools (`tool-headline-scorer`, `tool-readability-checker`,
`tool-cta-tester`, `tool-length-counter`), glossary, ASA/CAP guide, copy-length tables, checklists,
house-view, Copy Clinic (index + 3), every how-to guide, `choose-and-brief-a-copywriter`, `about`,
`contact`, and the two surviving library tools kept. In flight: `needs_page:triaged ×30`,
`needs_imagery:triaged ×20` — the pipeline's normal work; do not prod.

**Migration 764 is FULLY PROVEN**: classifier half 21:25Z 09-03 (NOTES (ccc)); planner half at the
11:09:25Z `plan_site` render — its `## Mission` block carries the brief object (`"audience": {"primary"…`).
Both consumers read a `text`-less brief. `Council-Reviewed: 888e7319…`.

**TWO PAGES THE OWNER CARES MOST ABOUT ARE NOT IN THE PLAN, deliberately and traceably:**
- **the lead route (`/get-copy-written/`) and the copywriter DIRECTORY.** The planner's `strategy_notes`:
  *"converting … via a single lead-route page … The entity-page type from strategy is deferred: the
  directory compilation has not yet populated individual entity pages, and planning them without
  entries would create empty pages that rule 3 holds back."* The home page's closing CTA already says
  *"Ready to have a human write it? One page connects you with UK copywriting companies"* (rev-5
  wording) — pointing at a page that does not yet exist; the CTA resolver will retarget it until it does.
- **Why no entries:** the `copywriter-directory-discovery` task (weekly, `directory-researcher`,
  organisations-only query) fired ONCE, 09-03 15:50:03Z, and **FAILED at `search_web`** (*"search query
  not found — check 'query', 'topic', or 'query_field'"*) — the researcher's config was then changed at
  22:05:57Z (no migration in git, no `agent_definitions_backup` snapshot — an unsnapshotted live edit
  by someone; not chased). `directory_entities` kind `copywriter` = **0**. **Re-fire NUDGED 11:4xZ**
  (`last_triggered_at`/`last_completed_at` set NULL so the scheduler fires it on its next tick via the
  real path). Watcher `bk751pk1k`. Check:
  `SELECT last_triggered_at, last_completed_at FROM scheduled_tasks WHERE name='copywriter-directory-discovery'; SELECT count(*) FROM directory_entities WHERE kind='copywriter';`
- **Once entries exist, the plan must be re-run to add the two pages**: file `needs_site_plan`
  (`source='manual-replan'`, key `site_plan_copyonline.co.uk`, handler `build-site-planner`, the fleet
  has done this 3×) — an OWNER-visible step; say so before firing.
- **OWNER DECISION:** `classification.content_features.copywriter_directory` was never set as a
  structured key (the brief's own recipe step 2; live shape on loanandmortgagecalculator:
  `{"kind":"mortgage-lender","reason":…,"recommended":true,"separate_page":…}`). The classifier rewrites
  that object on every run, so a hand-set flag is the `design_intent pinned=true does not hold`
  landmine. The strategy already carries the directory ("entity-page type from strategy"), so entries
  may suffice; if the planner still omits it after a replan, set the flag AND expect to re-set it.

**Also pending, owner-facing:** `owned_page_review:needs_human_review ×4` — the four brief tools
(`tool-headline-scorer`, `tool-cta-tester`, …) are *"not_built — needs owner-aware build, not the
generic builder"*. Whatever the owner-aware build path is, it is waiting on a human.

**Housekeeping to expect:** the two old `tool-*-guide` survivors are NOT in the plan (the plan made
`guide-insight-injector-guide` / `guide-website-brief-starter-guide` instead) → the orphan check will
flag them; archiving them then is a tidy, not a bug. 15× `image_source_unsatisfiable` include ARCHIVED
pages (`bugs_open/266` class). 6 archived rows still carry `deployed_at` (315/359 shape).

**Bugs from this lane, open:** `bugs_open/478` (the strategist's deployed-gate; oxenunity + cookly
stalled the same way, unowned) · `bugs_open/453` (fixed by 764 for the 4 expressions; lint is theirs).
**Owner rulings applied:** the negation phrase stays (`wont_fix` + doc_notes ruling).
**Watchers do not survive a session restart** — every watcher named in this file is dead if you are
reading it in a new session; re-arm from the checks above.

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

## 1y. ⚠ ADDED ~20:35Z — RETIREMENT IS DONE; §1z's re-fire task is CLOSED. What is live now:

- **Migration 764 is LIVE (20:55:27Z) and PROVEN (21:25Z, NOTES (ccc)): copyonline's classification is now sighted (category editorial, the brief's tools named, rev-5 wording quoted), the rendered mission block carries the brief object with zero `<no value>`. The planner's own run-once is the site's next `plan_site` — read its rendered `## Mission` block then. The HOLD file's guard refuses "already applied"; do not re-run it.**
- ~~**Migration 764 was APPLIED at 20:55:27Z (DB clock) on the owner's "It is quiet now" — the HOLD file's guard now REFUSES with "already applied"; do not re-run it.**~~ Run-once proof item `17bac4d6…` filed 20:56:06Z; its pass/fail is the NEW `classification.reasoning` for copyonline and the rendered `## Pre-Defined Mission` block. If this handoff is read before NOTES (ccc) exists, the proof had not landed — check the item and the artefact before believing 764 works; the ROLLBACK file is beside it.
- ~~**Migration 764 is council-APPROVED (round 2, `888e7319…`, 4 advisory/none high; `Council-Reviewed:` on `a5dbbffab`) and HELD — NOT applied. THE OWNER SAYS WHEN.**~~
  Apply: `kubectl -n ai-persona-system exec -i postgres-clients-0 -- psql -U clients_user -d clients_db -v ON_ERROR_STOP=1 -f - < docs/agent_docs/sql_for_agents/764_classifier_and_planner_render_the_brief_object_when_it_has_no_text_HOLD.sql`
  (refuses on md5 drift — re-pin if either template moved). Then MAKE IT RUN ONCE before claiming
  anything (header steps); the classification.reasoning sentence is the pass/fail. Rollback file beside it.
  Decision records: `SELECT subject_key, body FROM doc_notes WHERE categories ? 'migration-764';`
- ~~**Migration 764 is HOLD, in council ROUND 2 (same correlation `888e7319…`, run orch `b07fdcdd…`), NOT applied.**~~
  Round 1 was REVISE (bug_historian gating); every objection is answered in the file header and the
  proof now runs through the fleet's own renderer (`tplproof/`, `go test -tags tplproof`, README).
  ⚠ PRC-003 (`681b0ee65`, rides the next roll) STRIPS `<no value>` inside `RenderPromptTemplate` —
  after the roll, a `<no value>` census at `llm_call_log.prompt_rendered` reads clean for the wrong
  reason; measure with `ScanMissingValues` on the raw execution, as the harness does.
- ~~**Migration 764 is HOLD, council-submitted (`888e7319-01ae-4371-846d-76fe227a1ebc`), NOT applied.**~~
  Read the verdict: `SELECT body FROM doc_notes WHERE categories ? 'council-gate' AND body LIKE '%888e7319%' ORDER BY created_at DESC LIMIT 1;`
  APPROVED → the OWNER picks the moment (two shared agents, fleet-wide). Apply by hand, then **make it
  run once** exactly as the file header says — read the new `classification.reasoning` for a
  `.text`-less site; do not call it done before that. REVISE → resubmit with `RESUBMIT_CORR=888e7319…`.
- **Retirement, final state:** six archived + six retracted (repo commits `1b0c148…`, `44da691f…`,
  `f14a638e…`), three nav rows `inactive`, survivors untouched, verified at the repo with a present
  control. One `phantom_internal_link` item on `tool-insight-injector-guide` may still read `triaged` —
  its page's `content_data` is already clean; it is a no-op, not a stall. Six archived rows keep
  `pages.deployed_at` (315/359 shape) — expected members of any archived-and-deployed census.
- Everything in §1z below about the re-fire is DONE; keep it for the record.

## 1w. ⚠ ADDED 2026-09-04 ~07:50Z — the plan's real blocker, and what is in flight now
- **copyonline never had `needs_briefing` or `needs_site_plan`**: `domain-strategist.gate_next_item`
  skips the briefing when `check_site_deployed` says `is_deployed`, and that query counts ANY page
  with `deployed_at` — the five blind tool pages satisfied it. 090 `467f0283` → UNVERIFIABLE (loop
  cannot read `default_config` steps); **`bugs_open/478` FILED** on the first-hand chain; §9.100 in 016b.
  oxenunity.com and cookly.uk are stalled the same way (no plan); loancalculator was redriven by hand 08-15. `needs_briefing` filed by hand 07:47:25Z (item `479614c9`) → expect
  `needs_site_plan` → the plan. Watcher `bwdqsjvzs`. If the plan lands, read the planner's rendered
  `## Mission` block (764's second half) — but note PRC-003 is LIVE (v1.0.1360, image label
  `239ab362…`), so count holes with `ScanMissingValues` on the raw execution, never `<no value>`.
- Composition resolved 02:07Z (`content-hub-tools`, `library_match`); 445 lane has the fields.
- OWNER decision open: `brief_supplies_negation` — "practical craft notes, not legal or compliance
  advice" in `content_direction`; the check says only the owner edits it.
- 15× `image_source_unsatisfiable` (`tool-guide-intro` → `site_assets.image`, nothing generates it)
  include ARCHIVED pages — `bugs_open/266`'s class; noted, not chased.

## 1x. ⚠ ADDED ~21:40Z — a chassis roll was announced for ~22:00Z (baseline v1.0.1359)
Nothing of this lane rides it. After it: (a) read the new pods' `build provenance` line and test
`git merge-base --is-ancestor 681b0ee65 <stamp>` — PRC-003 live means `<no value>` censuses at
`llm_call_log.prompt_rendered` read clean for the WRONG reason from then on; (b) confirm no `claimed`
item is stranded on copyonline (`status='claimed'` older than ~30 min after the roll = stranded);
(c) expect ~300s of silently dropped dispatch after the restart — do not re-file on that evidence.

## 1z. ⚠ ADDED ~19:30Z — the owner ANSWERED (evening), and these are now the first tasks — SUPERSEDED BY §1y for the retraction items

Both §2 decisions below are MADE and executed; §2 is kept for the record. What is live and owed:

- **Ten repair items on copyonline** (`created_by LIKE 'portfolio_positioning lane 2026-09-03%'`,
  types `page_rerender`×3 `cta_links_stale`, `needs_rerender`×1 chrome refresh, `phantom_internal_link`×6
  → page-build-handler). Watcher `b14z2xpln` was armed; if this session is gone, check them yourself:
  ```sql
  SELECT item_type, status, COALESCE(spec->>'page_name','(site)') FROM site_work_items
   WHERE site_id='3d965325-519a-4515-b79f-50c886954a80' AND created_by LIKE 'portfolio_positioning lane 2026-09-03%' ORDER BY 1,3;
  ```
  **When all are terminal, re-run the three-source inbound audit (RUNBOOK_unpublish_primitive §"Audit")
  for the three tool URLs; if it reads zero, RE-FIRE the tools' retraction:**
  ```sh
  SITE_ID=3d965325-519a-4515-b79f-50c886954a80 PAGE_IDS='["3ae2096f-98c6-4589-92cd-b2f343140fbb","9fae1345-84a3-42dd-aa7b-b22ca314d335","09fdbca9-4d88-4011-907b-b5adf1206a82"]' ./docs/agent_docs/sql_for_agents/216_TRIGGER_page_retraction.sh
  ```
  Expected on success: 3 files removed from `gqls/sites`, 3 nav rows → `inactive`. The first run
  (`45a7eba9`, 18:58Z) was REFUSED on 13 named links — correct behaviour, see NOTES (vv).
- **The three retracted guides still carry `pages.deployed_at`** (`bugs_open/315`/`359` shape). Not
  fixed; if a census of "archived-and-deployed" pages is run, these three are expected members.
- **Brief revision 5 is the owner's record, not yet a live input** — the classifier/planner cannot read
  it (`bugs_open/453`) and the strategist has run. When 453 ships, or if the strategist re-runs, verify
  the lead route renders as "list of companies primary, enquiry secondary".
- **Analytics:** copyonline now carries `GTM-PQ3WCTBD` in `site_config.analytics`; expect one
  `stale_chrome` → `needs_rerender` at its next discovery run, then `check_gtm_state.sh --db` bucket D → 0.
- **Composition:** re-files itself when `design-discovery-agent` next selects the site
  (`MissingStyleCollectionCheck`, `sites.style_collection_id IS NULL`). Cadence unmeasured. Do not prod.

## 2. THE TWO DECISIONS THE OWNER STILL OWES — ~~still owes~~ ANSWERED 2026-09-03 evening (kept for the record)

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
