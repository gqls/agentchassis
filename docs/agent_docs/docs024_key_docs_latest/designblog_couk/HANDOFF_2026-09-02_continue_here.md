# HANDOFF 2026-09-02 — designblog.co.uk lane — continue here

**Session name: "designblog.co.uk".** Owner-created lane for the designblog.co.uk site
(remake №4 of portfolio positioning's 22 hosted-site remakes, live 2026-09-02). The lane's
job: hold the owner's critique of the site, verify everything at the served artefact, route
class defects to their owning mechanism threads, track receipt and answers, and execute
what the owner rules. One day old and dense — everything below happened 2026-09-02.

**Reading order for a fresh session:**
1. This file.
2. `docs/agent_docs/docs024_key_docs_latest/designblog_couk/CRITIQUE_2026-09-02_owner_site_review.md`
   — the owner's critique verbatim + per-point verification + corrections.
3. `docs/agent_docs/docs024_key_docs_latest/designblog_couk/NOTES_designblog_couk.md`
   — append-only; every thread answer digested with attribution, newest at bottom.
4. `docs/agent_docs/docs024_key_docs_latest/designblog_couk/README_where_we_are.md`
   — the owner's plain-prose log (append-only, his document — never rewrite).
5. `docs/agent_docs/docs024_key_docs_latest/SITE_DEFECT_CATEGORIES.md`
   — the fleet acceptance checklist this session extended (1.5, 2.6, 4.6, 4.7, §9).

---

## 1. FIRST ACTIONS for the next session (in order)

1. **Kubeconfig token was EXPIRED at handoff time** (the standing 3-day expiry; the
   owner refreshes it — `kubeconfig-token-expires-every-3-days` memory). Nothing
   cluster-side can be checked until then.
2. **Read the 718 council verdict — the run may have been ROLL-KILLED.** A fresh chassis
   build deployed right at handoff time, and a roll kills in-flight councils (the lendzy
   lane's 695 was roll-killed twice). Last seen: `gate_tooling_provenance | EXECUTING_STEP`
   at ~20:25Z. Check by payload, never by printed id:
   ```sql
   SELECT current_step, status, updated_at FROM orchestration_states
   WHERE collected_data->'input_data'->>'fix_correlation_id' =
         '2dae4f20-3baf-46f9-96c5-fb35609ed7bd';
   ```
   - APPROVED → nothing owed (the commit carries `Council-Submitted:`, 098 credits it).
   - REVISE/REJECTED → read the objections (`SELECT body FROM doc_notes WHERE categories ?
     'council-gate' ORDER BY created_at DESC LIMIT 1;`) and ACT — the change is live.
   - Run dead/missing after ~an hour → resubmit:
     `097_TRIGGER… docs/agent_docs/docs024_key_docs_latest/designblog_couk/COUNCIL_SUBMISSION_718.json`
     with `RESUBMIT_CORR=2dae4f20-3baf-46f9-96c5-fb35609ed7bd` in the environment so the
     trail accumulates. (Submission JSON preserved in this directory for exactly this case.)
3. **Confirm 718 still in effect** (it's DB config — a roll cannot revert it, but verify
   rather than assume; ~30 s):
   ```sql
   SELECT (length(t)-length(replace(t,'Content-carrying imagery is EXPECTED','')))>0 AS new_bullet,
          position('Use sparingly in v1' in t)=0 AS suppressor_gone
   FROM (SELECT default_config #>> '{workflow,steps,plan_site,config,prompt_template}' AS t
         FROM agent_definitions WHERE type='build-site-planner' AND is_active
           AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL) s;
   ```
4. **Remember the fresh roll's dispatch rule**: no orchestration dispatch within ~300s of a
   chassis pod restart — spawns silently dropped (CLAUDE.md).

## 2. What is DONE and PROVEN (all 2026-09-02)

- **Owner critique verified point-by-point at the served site** — all 8 points confirmed
  (4 listing pages with ZERO content items; no tools nav link; chrome-only imagery;
  AI-sounding copy verbatim; design sameness). See CRITIQUE doc §2 + addenda.
- **All 7 owner-named threads received the critique and ANSWERED with measurements**
  (routing table + digests in NOTES): Portfolio positioning, components, experience loop,
  theme kits, site design planner, vigilant designer (session "offer analyser benefit
  analyser visual designer" — needs the `[4628f9]`-style ref, name collides with an offline
  RC session), copy quality two stage. Supplementary routes: editorial design uplift,
  feed lane (never ACKed — chase), inline guide imager, bugs_open/444 session,
  gamedesign.uk, analytics ("google" session).
- **MIGRATION 718 EXECUTED — the owner's ruled change** ("go ahead with the planner prompt
  and exemplar changes"): build-site-planner's plan_site prompt now EXPECTS content
  imagery. Five anchored count-guarded replaces (the row is co-edited — never wholesale
  jsonb_set, LANDMINES); rule 16 kept + verify-asserted; rule-3/article pages exempt
  (bugs_open/114). Applied 19:59:56Z, snapshot taken, `schema_migrations` row written,
  independently re-verified at the live row (suppressor ×0, new bullet ×1, infographic
  exemplar ×1, rule 16 ×1, length 28,117→30,164). Files:
  `docs/agent_docs/sql_for_agents/718_planner_imagery_content_expected_prompt_and_exemplar.sql`
  (+`_ROLLBACK`, roundtrip proven byte-exact), commit `40dbeaea4`.
- **SITE_DEFECT_CATEGORIES.md extended** (owner: "so we can look over new sites with these
  in mind"): new 1.5 (AI-tell copy through the gate), 2.6 (live surface with no nav path),
  4.6 (chrome-only imagery plan — REMEDY-LANDED note: post-718 a zero is a defect signal),
  4.7 (one-slab articles), §9 (cross-site sameness — the class no per-site audit sees);
  extended 2.2 (four-empty-pages + 444 upstream checks) and 4.2 (visible CORRECTION: an
  `<img>` census is half-blind — always pair it with the `background-image:.*url()` grep).
- **Two WRONG_CALLS of mine logged**: "content imagery is absent" (CSS-background heroes;
  commit `5d76472d8` — two agreeing measurements sharing an encoding are one measurement).
  The same trap then caught gamedesign and nearly caught inline guide imager — three
  sessions in one night.

## 3. The convergent findings (what to tell anyone who asks "what's actually wrong")

- **Sameness = SELECTION, not library poverty** (three independent measurements):
  chrome hardcoded (`ChromeSlotFunction()`, 36/37 sites identical header/footer, 10 unused
  headers, only 6/40 `style_collections` pin `header_component_id`); layout library gap
  (all three remakes → `magazine-grid`; no "content hub with tools" archetype —
  `bugs_open/445`); top-10 components carry 78–87% of slots, two thirds of the library
  unused/single-site. ⚠ components have THREE names (`name` ≠ `function` on 76%;
  `slot_name` a third) — LANDMINE `0b3151337`; resolve via `component_id`.
- **Emptiness = NO PRODUCER** (`bugs_open/444`, class fix in council, their gate migration
  renumbered 718→**720**): feeds have 0 `content_sources` rows (mechanism live elsewhere);
  "uk-design-studio" isn't a DIR-001 kind (⚠ and the served page uses the BARE
  `directory-listing` component via `directory-json-exporter` — decide WHICH producer
  before populating); glossary/inspiration have no item producer anywhere in the estate.
  Detection was blind by construction (experience loop's rule C drops empty indexes before
  the rule runs — empty-index rule IN BUILD with them; content-quality-auditor post-694
  now asks the owner's exact question in record mode).
- **Imagery**: the planner was OBEYING its own prompt (fixed by 718). Display side: the
  orphaned-hero class — **7 components / 157 instances / 61 pages render the wrong hero
  while passing every check** (hero-tool 23 sites/76 instances the biggest); 4 pages DO
  get their own asset via an unidentified route (leopardess tool page) — **identify that
  route before editing seven components** (with components thread + `bugs_open/114`
  second CONTRIB). In-article imagery blocked on article structure (114: max 1 prose
  section on all 462 article pages) — kept as a SEPARATE ask, deliberately not forced
  by 718.
- **Copy**: the build ran the gated path; four known classes (repair `no_answer_for_target`
  — candidate fix: one re-ask; banned words detection-only, no page repairer; ruled
  AI-tell/mistype classes; the empty-room class whose listing half 444(1) closes upstream
  — copy lane deliberately narrowed their title-promise gate to the non-listing half,
  their `1b5209a1a`).

## 4. WATCH LIST (things that will move without this session)

| Item | Who | What arrives |
|---|---|---|
| 718 council verdict | this lane | read + act (first action above) |
| First post-718 plan = imagery canary | Portfolio positioning | plan's `imagery.sections` should carry illustration/infographic entries; check they land on display-capable sections (instruction-only limit — residual §4.1 risk) |
| 444 gate (mig 720) verdict + apply | bugs_open/444 session | listing pages refused/degraded at plan time; section-index arm proven (feed shape = their test fixture) |
| Empty-index detector rule | experience loop | they re-run over designblog's four pages when built |
| Orphaned-hero route + fix shape | components (+ inline guide imager census in 114) | don't accept a 3-component fix; leopardess route first |
| GTM key + chrome rerender wave (17 pages) | analytics ("google") | expected, not damage; next serve-verify must expect the GTM head; empty pages re-render empty |
| gamedesign.uk re-plan post-718 | gamedesign.uk session | second canary; their 446 + 444/114 CONTRIBs |
| planner-vs-fallback share measurement | theme kits (asked) | sizes page_archetypes vs planner-prompt as the structure lever |
| feed lane ACK | this lane | never ACKed the design-vertical source ask — chase |

## 5. OWNER DECISIONS GATHERED (put these to him when he asks "what's waiting on me")

1. theme kits' `page_archetypes` apply/roll (committed, inert).
2. Pre-flight gate assignment (vigilant designer offered: cohort component-overlap +
   imagery counts per new site; queries in `RUNBOOK_designblog_couk.md`).
3. Cheap per-site fixes now vs class fixes only: tools nav (a tools section-index page
   with `in_header` + nav rebuild) and a pinned non-default header.
4. Who designs the "content hub with embedded tools" layout archetype (445).
5. Priority on the two unblocked-and-unstarted flips — Illustrated Text Block selection,
   info-card-grid carousel default (vigilant designer owns; effect UNVERIFIED — editorial
   design uplift's caution stands).
6. /the-design-feed/ fill route: replan as news-index vs child pages (444 session: both
   resolve cleanly; a news source ALONE cannot fill the current section-index shape).

## 6. Traps this lane learned (do not re-derive)

- **An image census must run BOTH greps** (`<img` AND `background-image:.*url(`) — three
  sessions caught in one night; SITE_DEFECT_CATEGORIES 4.2 carries the correction.
- **Migration numbers collide constantly**: 715 ×2, 716 ×2 exist; 718 was nearly taken
  twice more (444's gate → 720; 719 taken the same night). Check max AND expect a race.
- **Editing `build-site-planner`**: anchored `replace()` with exact-count guards, snapshot
  first, dry-run with COMMIT→ROLLBACK, verify with DO/RAISE; the row is co-edited by
  other lanes (591/595/598 precedent; 718 pattern is the worked example in this lane).
- **Council runs die with rolls**; find runs by payload
  (`fix_correlation_id`), budget ~30 min dispatch latency, never retry on a missing row.
- **"The vigilant designer thread"** = session "offer analyser benefit analyser visual
  designer" (needs ref on send — name collision with an offline RC session).
- A population derived from the reported case is a GUESS — derive from the predicate,
  then sample the consequence (the 7-component census vs the 3-component hunch).

## 7. Commits this session (all pathspec, this lane's own files only)

`e7d84fdcd` lane opened · `38a5a740a` ratchet line · `d9c4e97fb` thread digests ·
`5d76472d8` WRONG_CALLS · `22aeb6223` imagery correction · `4dec46dd2` copy closure ·
`a379d977d` SITE_DEFECT_CATEGORIES additions · `40dbeaea4` **migration 718** ·
`7f859c833` 718 trail · `b3e3d2a21` README · `472055960`/`745ce6460`/`0d3786c6b`/
`4ffa59f74` follow-ons · this handoff.
