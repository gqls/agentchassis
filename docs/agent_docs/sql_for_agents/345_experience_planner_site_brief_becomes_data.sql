-- 345_experience_planner_site_brief_becomes_data.sql
--
-- bugs_open/227 — `experience-planner`'s prompts hardcode ONE site's diagnosis,
-- so an experience plan for any other site describes that site's pages.
-- Filed 2026-08-08 by the loancalculator_couk lane; this is fix candidate 1
-- ("move the worked case out of the prompt and into the input").
--
-- THE DEFECT, as measured 2026-08-08 ~22:45Z against the live row
-- (e0194bee-3b8e-4a38-a402-a031d4fe7a15, the ONLY experience-planner row).
-- CASE-INSENSITIVE census of provocation|gauntlet|arena|vonc|spark, because a
-- case-sensitive one is what hid two of these steps from my own first pass:
--   compose             41
--   review_feasibility   2   (holds a veto)
--   review_honesty       2   (holds a HARD veto)
--   review_mvp           2
--   reframe              1   (the step that rewrote the vetoed plan)
--                       48 total
-- The bug file names only `compose`. It is FIVE prompts. A case-SENSITIVE grep
-- returns 37 and shows only three steps: "Gauntlet" is capitalised everywhere it
-- appears in reframe and review_honesty, so those two steps read as clean. If you
-- re-measure this, measure lower().
--
-- THE CHANGE, config only, live on apply (no image, no roll):
--   1. NEW step `load_brief` (query_database, output_format=object, aliased
--      `AS text` exactly like load_context) reading the site's OWN brief from
--      doc_notes WHERE subject_type='experience' AND subject_key=$1 AND
--      categories ? 'experience-brief'. load_schema_hint.next_step repoints to
--      it; load_brief.next_step = compose. The query is a scalar COALESCE so it
--      ALWAYS returns exactly one row: a miss yields the visible sentinel
--      "(no brief on file …)", never zero rows and never NULL. That sentinel is
--      what makes a mis-keyed brief disconfirmable — a silent empty string would
--      read exactly like a site that has no brief.
--   2. compose.config.input_fields gains "experience_brief". WITHOUT THIS THE
--      TEMPLATE RENDERS EMPTY AND NOTHING ERRORS — the step only receives the
--      fields it lists (live: ["experience_context","input_data"]). This is the
--      single most likely way for this migration to look applied and do nothing.
--   3. compose.prompt_template rewritten: the four site-specific sections (the
--      2026-07-17 diagnosis of vonc's three surfaces, decisions D1/D2/D2-HARD,
--      the /data/provocations.json + tools.apis.uk contract, and the scattered
--      Gauntlet/provocation examples in HARD RULE 2, §2, §3, §5 and the LENGTH
--      section) are REPLACED by {{.experience_brief.text}} plus an explicit
--      instruction that a site with no brief may not import surfaces from
--      anywhere else.
--   4. review_mvp / review_feasibility / review_honesty / reframe: surgical
--      replace() of the sentences that hold one site's surfaces as the general
--      rule. Safe against a silent no-op because the drift guard below pins all
--      five prompts by md5 — if the md5 matches, the substring is present by
--      construction — and the verify block re-censuses the whole row.
--   5. vonc-spark-game's brief is INSERTED into the new channel IN THE SAME
--      TRANSACTION, verbatim, before the prompt loses it.
--
-- WHY 5 IS NOT OPTIONAL: D1 and D2 are owner rulings carrying "do NOT
-- relitigate", and 59 of the 61 experience plans all-history are vonc's. A fix
-- that only deletes the vonc text trades a contaminated plan for a de-briefed
-- one on the only site that has ever used this agent in anger.
--
-- THE REVIEWERS TOO — the finding the bug file does not have, and the reason
-- fixing `compose` alone would have been a trap. THREE OF THE FOUR COUNCIL SEATS
-- HOLD ONE SITE'S CRITERIA AS THE GENERAL RULE:
--   review_feasibility (veto)  judges whether data is "in /data/provocations.json
--                              or client-computable", watches for "the daily
--                              emitter", assesses "the client-side scoring/timer
--                              for the minimal-real Gauntlet"
--   review_mvp                 is told the core loop is "land on a provocation ->
--                              file a position -> see the day's record; enter a
--                              real timed Gauntlet round"
--   review_honesty (HARD veto) is told "vonc's evidence_base has ZERO facts, so
--                              ANY hard number that is not read from real client
--                              state is fabricated"
-- and `reframe` — the step that rewrites a vetoed plan, i.e. the one that produced
-- the SECOND contaminated debt-difficulty-help plan — carries "If the Gauntlet is
-- what was vetoed …".
--
-- So the seat that correctly vetoed the loancalculator plan holds vonc's criteria
-- itself. Two consequences, both material:
--   (a) a CORRECT post-fix plan can still be objected to by a seat looking for a
--       feed, a timer and a game loop this site was never going to have — and that
--       would look exactly like "the fix did not work";
--   (b) the hardcoded premise is not merely site-specific, it is STALE, inside a
--       hard veto. vonc's evidence_base had zero facts when this was written on
--       2026-07-18. It has FOUR today, written 2026-08-08 08:58Z (site_specs
--       aspect='evidence_base', is_current, data->'facts'; loancalculator has no
--       such spec row at all). A vonc plan run today is told by its own honesty
--       auditor that four verified facts do not exist. Nothing updates a premise
--       pinned inside a shared prompt when the site moves underneath it — which is
--       the general argument for the brief being data, independent of 227.
--
-- CONSUMERS TOLD (owner ruling 2026-07-29 §3): the only producer of
-- needs_experience_plan is ./092_TRIGGER_experience_plan.sh (no dispatch loop
-- claims that status — the trigger's own header note verifies this), and the
-- only consumer of the output is doc_plans subject_type='experience', today
-- vonc-spark-game (59 plans) and debt-difficulty-help (2, both contaminated,
-- both already demoted by hand). The vonc lane's guarantee is UNCHANGED by
-- design: its brief text moves channel but not content. What changes for it is
-- that the brief is now editable per experience without touching a fleet-shared
-- agent row.
--
-- NOT FIXED HERE — the second, separable defect in 227: `persist_plan` runs
-- immediately after compose (compose -> persist_plan -> review_journeys -> …),
-- so the plan is written is_current=true BEFORE the council sees it and nothing
-- demotes it when the verdict is rejected. write_doc_plan has no config to write
-- a non-current row (platform/orchestration/actions/write_doc_plan_action.go:104
-- INSERTs and relies on doc_plans.is_current DEFAULT true), so the two routes are
-- (a) rewire the graph so persist happens only on the approved path — config
-- only, but complete_escalated.output_fields lists plan_persisted and the
-- escalation path is meant to surface a plan, so that coupling must be answered
-- first; or (b) add an is_current/`set_current_when` config field to
-- write_doc_plan — a platform seam, so council gate + concept register.
-- Deliberately left out of this file: it is independent, and mixing a Go seam
-- into a config fix is the scope-veto shape CLAUDE.md warns about.
--
-- ROLLBACK: 345_..._ROLLBACK.sql restores from agent_definitions_backup
-- (two-arg snapshot_agent below → agent_definitions_backup, NOT an is_snapshot
-- row — LANDMINES 2026-07-30), picking by snapshot_taken_at DESC with
-- snapshot_reason LIKE 'pre-update: 227%'. Every backup row for one agent shares
-- the source row's id and created_at, so ordering by created_at returns an
-- arbitrary snapshot — order by snapshot_taken_at.
--
-- VERIFY (after apply) — the census is necessary but NOT sufficient:
--   SELECT count(*) FROM agent_definitions WHERE type='experience-planner'
--     AND is_active AND default_config::text ILIKE '%provocation%';   -- expect 0
-- That only proves the text is gone. The behavioural proof is a run:
--   ./092_TRIGGER_experience_plan.sh loancalculator.co.uk debt-difficulty-help \
--     "getting help when you cannot keep up with a loan repayment"
-- then, on that correlation, assert BOTH directions at the rendered prompt —
-- llm_call_log.prompt_rendered is the only place that shows what the model was
-- actually handed, which is the field that would have caught this bug on day one:
--   SELECT step_name,
--          prompt_rendered ILIKE '%provocation%'            AS leaked,      -- expect false
--          prompt_rendered ILIKE '%no brief on file%'       AS got_sentinel -- expect TRUE for this site
--     FROM llm_call_log WHERE correlation_id='<CID>' AND step_name='compose';
--   SELECT body ILIKE '%provocation%' AS still_wrong        -- expect false
--     FROM doc_plans WHERE subject_type='experience' AND subject_key='debt-difficulty-help'
--    ORDER BY created_at DESC LIMIT 1;
-- POSITIVE CONTROL, so a passing run cannot be a run that never read the channel:
-- the same two assertions for vonc-spark-game must come out the OTHER way —
-- leaked=true (its brief legitimately contains the word) and got_sentinel=false.
-- A fix that silently fails to load ANY brief passes the loancalculator half and
-- fails this one.

-- ============================================================================
-- Probe guard: refuse a second application.
-- ============================================================================
DO $$
BEGIN
    IF EXISTS (
        SELECT 1 FROM agent_definitions
        WHERE type = 'experience-planner'
          AND is_active AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL
          AND default_config #> '{workflow,steps,load_brief}' IS NOT NULL
    ) THEN
        RAISE EXCEPTION '227/345: already applied — load_brief step exists';
    END IF;
END $$;

-- ============================================================================
-- Drift guard: composed against the exact live texts fetched 2026-08-08 22:30Z.
-- The agent_definitions row was bulk-touched at 22:01:02.606329Z along with 186
-- others (one statement, no snapshot taken, mechanism not identified) — so this
-- guard is not ceremony on this table.
-- ============================================================================
DO $$
DECLARE
    c_md5 text; m_md5 text; f_md5 text; h_md5 text; r_md5 text; nx text;
BEGIN
    SELECT md5(default_config #>> '{workflow,steps,compose,config,prompt_template}'),
           md5(default_config #>> '{workflow,steps,review_mvp,config,prompt_template}'),
           md5(default_config #>> '{workflow,steps,review_feasibility,config,prompt_template}'),
           md5(default_config #>> '{workflow,steps,review_honesty,config,prompt_template}'),
           md5(default_config #>> '{workflow,steps,reframe,config,prompt_template}'),
           default_config #>> '{workflow,steps,load_schema_hint,next_step}'
      INTO c_md5, m_md5, f_md5, h_md5, r_md5, nx
      FROM agent_definitions
     WHERE type = 'experience-planner'
       AND is_active AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;
    IF c_md5 IS DISTINCT FROM '8b05c372ddd8ca1696900516db880c04'
       OR m_md5 IS DISTINCT FROM 'dfb8111e0e06ff9ff7ed0c90d158498e'
       OR f_md5 IS DISTINCT FROM '4a86a799fe89bf65f1c17b99f5da810a'
       OR h_md5 IS DISTINCT FROM 'fffda6802ad25fd26e90bfac5a814d21'
       OR r_md5 IS DISTINCT FROM 'ebe84e4fedf431ae165c19470085cae3'
       OR nx    IS DISTINCT FROM 'compose' THEN
        RAISE EXCEPTION '227/345: DRIFT — live config differs from what this file was composed against (compose %, mvp %, feas %, honesty %, reframe %, next_step %). Re-read the five prompts and recompose.',
            c_md5, m_md5, f_md5, h_md5, r_md5, nx;
    END IF;
END $$;

BEGIN;

SELECT snapshot_agent('experience-planner',
    'pre-update: 227 — site-specific brief becomes data (loancalculator_couk lane)');

-- ============================================================================
-- 1. vonc-spark-game keeps its brief — same content, new channel. Verbatim from
--    the compose prompt this file is about to rewrite. Idempotent.
-- ============================================================================
INSERT INTO doc_notes (subject_type, subject_key, site_id, body, categories, source, source_agent, created_by)
SELECT 'experience', 'vonc-spark-game',
       (SELECT id FROM sites WHERE domain = 'vonc.com' LIMIT 1),
$brief$## The diagnosis you are fixing (three broken surfaces, artifact-verified 2026-07-17)
1. /provocations/index.html — archive entries are runtime-filled into a template whose href="#" was never given a destination; per-provocation detail pages were never planned (needs_page:provocation, now owned_page_review).
2. /tools/arena/index.html — the tool-arena widget does NOT fetch /data/provocations.json (the feed is live, 200/5.6KB) so it shows "Loading… DAY 0" forever.
3. /tools/gauntlet/index.html — HISTORY: shipped as a mock (both CTAs href="#", fabricated stats, an invented "Live" leaderboard). PARTIALLY REMEDIED 2026-07-22: the dead hrefs and fabrications were removed — but the tool remains a hollow shell whose controls produce no outcome a visitor can perceive (owner-verified: the primary CTA starts an already-visible timer; checkboxes tick a progress bar bound to nothing). build_status=needs_rebuild. The REAL fix is D1's debate build.

## Decisions already made (owner-accepted; do NOT relitigate)
D1 (REVISED 2026-07-23 by owner ruling — supersedes the earlier minimal-real client-side-only cut) — the Gauntlet ships as a REAL DEBATE against an AI opponent, backed by the fleet's first live backend API (built in the platform concurrently with this plan). The flow: the visitor reads today's provocation, files a written Position, the AI opponent returns a REAL opposing Position plus a challenge to the visitor's, the visitor defends within the clock, and the AI returns an honest verdict with reasons on whether the take held up. The three objective checkboxes become REAL self-checking steps bound to these events (position filed / defence sent / verdict received before the clock runs out). The opponent is honestly labelled an AI competitor while the site has no human traffic. NO leaderboard in this round; zero fabricated numbers. DEGRADED MODE: if the API is unreachable the tool says so honestly and disables its controls (e.g. "the AI opponent is offline — try again later") — never simulate, never fall back to a mock.
D2 (REVISED 2026-07-18 by owner ruling — this supersedes any earlier "static detail pages" instruction) — per-provocation detail is rendered CLIENT-SIDE on the EXISTING archive page (/provocations/index.html), which is already a runtime-fill shell hydrating from /data/provocations.json. There are NO per-provocation static pages and NO new page_type in this MVP. Rationale, so you do not re-derive it: a prior council round established at HIGH severity that page_type='provocation' has zero prior rows and no proven build/render pipeline, so authoring one is not MVP work. Static per-provocation pages and the daily emitter both move to LATER.
D2 HARD CONSTRAINT — client-side detail must be a REAL, OBSERVABLE outcome, never a dead control. The original defect on this site was archive entries whose href was "#" with nothing behind it; you must not reproduce it in a new form. Opening an archive entry must produce an observable state change that an acceptance check can assert (for example: a deep-linkable URL fragment or query parameter, plus a detail region that becomes populated and visible with that entry's real content from the feed). A control that only toggles a class, or that leads to an empty/placeholder region, is the same defect wearing a different coat. If a provocation in the feed has no detail content to show, the entry must not be presented as openable at all.

## The data contract shape (fixed by the existing client loader)
/data/provocations.json has keys: today {eyebrow, headline (may contain <em>), body, primary_cta{url,label}, secondary_cta{url,label}, stats[3]{value,label}}, lobby[≤4]{icon,title,desc,url}, arena {…reserved for the Arena widget}. Static site with ONE exception (D1): the Gauntlet debate calls the fleet's tools API. Base URL is API_BASE = https://tools.apis.uk — DECIDED and LIVE, not a placeholder (the built JS carries it as a literal constant). VERIFIED LIVE 2026-07-25: a full real round-trip was run against this exact URL through the public internet (Cloudflare -> tunnel -> island VM -> tools-api), not a mock or a local test. POST https://tools.apis.uk/api/v1/tools/gauntlet/round with Origin: https://vonc.com -> HTTP 200, real body incl. round_id and today's actual provocation JSON. POST .../position with {round_id, position_text} -> HTTP 200, a genuine Anthropic-generated counter_position and challenge (not templated). POST .../defend with {round_id, defence_text} -> HTTP 200, a genuine Anthropic-generated verdict and reasons; two full rounds completed this way, both verdict="opponent wins" (the AI judges honestly, not a pushover). Denied-origin POSTs return 403; a missing round_id returns 404; a malformed/oversized body is rejected; the AI-unavailable path returns 503 with a clean JSON error body (Cloudflare replaces raw 502 bodies, so 503 is the status that survives to the browser — write the front-end's degraded-mode handling against 503, not 502). Every one of these was exercised for real, not asserted. ROUND-5 GAPS (from the council's own 2026-07-25 escalation, corr fcdf8e72 — address these directly rather than rediscovering them): (a) the EXACT verified JSON response shapes, taken from the tools-api source and a live round-trip, so the data contract can be exact rather than approximate: POST .../round -> {"round_id":"<uuid>","provocation":{"eyebrow":"...","headline":"...","body":"...","primary_cta":{"label":"...","url":"..."},"secondary_cta":{"label":"...","url":"..."},"stats":[{"value":"...","label":"..."}]}} (provocation is the verbatim 'today' object from /data/provocations.json — do not invent extra fields); POST .../position -> {"counter_position":"...","challenge":"..."} (two string fields, nothing else); POST .../defend -> {"verdict":"...","reasons":"..."} (two string fields, nothing else). (b) gauntlet-interface's CURRENT live js_content actively simulates entering a round on its enter-button click (starts the real timer, scrolls to the challenge panel, focuses the first objective) — this is the 2026-07-22 partial-remedy build and it must be SEQUENCED for change (not left as-is), so that in disabled/offline-scaffolding mode the click instead sets the honest offline status text and does nothing else; never let a click both look disabled and still run the old simulate-a-round behaviour. (c) two existing loaders/components need their own gaps named as sequenced steps if the plan touches them: provocations-archive-loader's current source only does cloneNode+setText for date/title/teaser/stat plus a conditional href — it has no detail_body/slug read and sets no --linked/--static class, so any archive-detail journey needs an explicit loader-modification step, not an assumption that it already splits entries; and any Journey referencing tool-arena-interface must have that component's actual html_template/js_content pulled into context and quoted before the plan asserts its selectors exist. The EXACT access paths the new gauntlet JS must implement — greenfield pairs: pin them verbatim in §3 and back EACH with a §5 criterion that fails if it is never built: POST {API_BASE}/api/v1/tools/gauntlet/round -> {round_id, provocation:{headline, body}}; POST {API_BASE}/api/v1/tools/gauntlet/position with {round_id, position_text} -> {counter_position, challenge}; POST {API_BASE}/api/v1/tools/gauntlet/defend with {round_id, defence_text} -> {verdict, reasons}. Everything else dynamic remains client-side JS reading this JSON.

## Site-specific notes for the plan's sections
- Evidence: vonc's evidence_base has ZERO facts, so an MVP number must come from the feed / client-computed real state, or not appear.
- §2 Promise ledger, worked example for this site: "Enter the Gauntlet" -> a playable timed round actually starts.
- §3 Data contracts: name the emitter decision from D2.
- §5: the feed this site serves is /data/provocations.json.
$brief$,
       '["experience-brief"]'::jsonb, 'bugs_open/227', 'cli', 'sql:345'
WHERE NOT EXISTS (
    SELECT 1 FROM doc_notes
     WHERE subject_type = 'experience' AND subject_key = 'vonc-spark-game'
       AND categories ? 'experience-brief'
);

-- ============================================================================
-- 2. The agent: new load step, the input field, and three de-contaminated prompts.
-- ============================================================================
UPDATE agent_definitions
SET default_config =
    jsonb_set(
    jsonb_set(
    jsonb_set(
    jsonb_set(
    jsonb_set(
    jsonb_set(
    jsonb_set(
    jsonb_set(
        default_config,

        -- 2a. the new per-site brief channel
        '{workflow,steps,load_brief}',
        $step$
        {
            "action": "query_database",
            "config": {
                "query": "SELECT COALESCE((SELECT string_agg(body, E'\\n\\n---\\n\\n' ORDER BY created_at) FROM doc_notes WHERE subject_type = 'experience' AND subject_key = $1 AND categories @> '[\"experience-brief\"]'::jsonb), '(no brief on file for this experience — there is no prior diagnosis, no prior decision and no fixed data contract. Plan from the live site context alone.)') AS text",
                "params": ["input_data.experience_key"],
                "error_step": "complete_refused",
                "output_format": "object"
            },
            "next_step": "compose",
            "output_field": "experience_brief",
            "description": "The site's OWN brief for this experience (doc_notes, category experience-brief). The per-site channel that replaced the hardcoded single-site diagnosis — bugs_open/227. Always returns one row; a miss returns the visible no-brief sentinel, never NULL."
        }
        $step$::jsonb),

        -- 2b. chain it in
        '{workflow,steps,load_schema_hint,next_step}', '"load_brief"'::jsonb),

        -- 2c. the step must be GIVEN the field, or the template renders empty
        '{workflow,steps,compose,config,input_fields}',
        '["experience_context","experience_brief","input_data"]'::jsonb),

        -- 2d. the de-contaminated compose prompt
        '{workflow,steps,compose,config,prompt_template}',
        to_jsonb($prompt$# PROMPT — compose the EXPERIENCE_PLAN

You own the WHOLE experience, not a page or a tool in isolation: the promise every button makes, the journey a visitor takes, the data a widget needs, the honesty of every number on the page. Write an EXPERIENCE_PLAN for the "{{.experience_name}}" experience on {{.experience_domain}}. This document travels; a build round and an acceptance ladder will execute it verbatim.

## The brief for THIS experience — the ONLY site-specific instruction you have
Everything between the next paragraph and "## HARD RULES" is this site's own brief: the diagnosis being fixed, decisions already taken (owner-accepted — do NOT relitigate them), and any fixed data contract. It is written per experience and travels with it.

If it reads "(no brief on file …)" then there is no prior brief, and that is normal, not an error: plan from the live site context below alone. In that case you must NOT import a diagnosis, a decision, a file path, an endpoint, a component or a tool name from anywhere else — you have not been told about any other site, and a surface you cannot find in the live context does not exist.

{{.experience_brief.text}}

## HARD RULES
1. A not-yet feature is ABSENT or labelled coming-soon — NEVER simulated. No dead controls (href="#"/no-op), no fabricated numbers, no invented users, ever.
2. Every quantitative claim traces to a runtime data source THIS site actually serves, or to an evidence_base fact. If this site has neither for a number, the number does not appear.
3. Reference REAL pages and selectors from the live context below; do not invent page names. Every page, path, endpoint, component and selector you name must appear in the live context or in the brief above — if it is in neither, it does not exist and you may not plan against it.
4. Tool-owned pages (rebuild_policy=owned) are rebuilt via the tool pipeline, never the generic page builder.

## Live site context
The "Attached components" list is the COMPLETE ground truth for this site: every component attached to every page, with its level and active flag. If a component is not in that list it is NOT attached. Do not claim a component exists, is active, or is "confirmed by query" unless the list shows it — and equally, do not object that something is unverifiable when the list settles it. Anything missing must be sequenced as a create/attach step.

{{.experience_context.text}}

## Write the plan as markdown with EXACTLY these sections
### 1. Journeys
Each journey is an ordered list of steps; every step names: page, control (a real CSS selector), action, and the OBSERVABLE outcome. No step may end at "#".
### 2. Promise ledger
A table: CTA copy → the page/state the destination must deliver. Every CTA in the table is either copy this site already uses or copy this plan specifies; the destination is what a visitor actually gets, not what the label implies.
### 3. Data contracts
What data this experience needs at runtime, WHICH file or endpoint serves it — name the exact path, taken from the brief or from the runtime loader source in the live context, never a path from another site — who writes it and when, and what is client-side-only. Name any emitter decision the brief records. If the experience computes ANY number a visitor will read as a score/metric, define the EXACT computation and the honest meaning of its label here — an undefined "score" is a soft fabrication, and an acceptance check that only asserts "some digits appeared" cannot tell a real computation from an arbitrary one.
### 4. MVP cut + LATER
The round-1 scope (the smallest honest playable loop) and an explicit LATER list. Restate any constraint the brief marks as a decision, as it applies to the cut. Write the MVP cut as an ORDERED, GATED step list: any prerequisite DATA step (e.g. committing the feed file this experience reads, with real content) is step 0 with an explicit gate ("do not proceed until it returns 200 with real content"), because a rebuild has nothing to fetch until it exists. Every later step names what must be true before it starts, and any claim that an existing work item is already resolved must be re-verified at build time, not merely cited. A dependency mentioned only in prose is NOT sequenced.
### 5. Acceptance criteria
A fenced ```criteria block of JSON the runner executes, using ONLY these check types (multi-page journeys are described narratively in §1; the runner journey type is a later phase):
   {"profiles":["desktop","mobile"],"container":".tool-container","checks":[
     {"id":"...","type":"selector_exists","selector":".real-selector"},
     {"id":"...","type":"asset_loads","path":"<the exact path §3 says this site serves>"},
     {"id":"...","type":"interaction","steps":[{"action":"fill|click|select","selector":"#real","value":"x"}],"expect":{"selector":"#real","text_matches":"..."}},
     {"id":"...","type":"page_status_ok"},{"id":"...","type":"no_horizontal_overflow","profiles":["mobile"]}]}
   Every selector must be one you expect the built page to actually have, and every path must be one §3 says this site serves.

## LENGTH AND QUOTING DISCIPLINE (a run has already died here)
The site context above includes REAL JAVASCRIPT SOURCE for the runtime loaders and component-owned scripts. That source is REFERENCE MATERIAL FOR YOU TO READ. It must NEVER be reproduced in the plan. Do not paste loader bodies, function definitions, or long code blocks into any section. When a data field or selector matters, name it and give the one-line access path (e.g. `data.entries[]`, `.tool-container__status`) — never the surrounding code.

The whole plan must come in WELL UNDER the output cap: aim for about 14,000 characters, and never exceed 20,000. A plan that hits the cap is DESTROYED, not shortened — the run fails outright with stop_reason=max_tokens and produces nothing. Approved plans have run to about 14,000 characters, so this is comfortable, not tight.

Priority if you are running long: the ```criteria fence in §5 has ABSOLUTE priority and must always be complete, valid, closed JSON followed by the END trailer. Cut narrative, prose rationale and LATER-list detail first. A terse plan is fine; a cut-off plan is worthless.

## Output format (IMPORTANT)
Output the whole plan as markdown: start with "# EXPERIENCE_PLAN — {{.experience_name}}", the five sections, the ```criteria fence, and then — AFTER the closing ``` of that fence — one final line exactly: <!-- END EXPERIENCE_PLAN -->. No preamble before the "#", no commentary. (The trailer line is required so the criteria fence is preserved verbatim in storage.)
$prompt$::text)),

        -- 2e. the MVP seat stops holding one site's core loop as THE core loop.
        --     `default_config` here is the OLD row value (a single UPDATE reads the
        --     pre-update snapshot of its own row), which is exactly what replace()
        --     needs — and neither 2a–2d nor 2f touches review_mvp.
        '{workflow,steps,review_mvp,config,prompt_template}',
        to_jsonb(replace(
            default_config #>> '{workflow,steps,review_mvp,config,prompt_template}',
            'without breaking the core loop (land on a provocation → file a position → see the day''s record; enter a real timed Gauntlet round)',
            'without breaking the core loop — where "the core loop" is the one §4 of THIS plan defines for THIS experience, not a loop you expect from another site'
        ))),

        -- 2f. same for the feasibility seat, which holds a veto
        '{workflow,steps,review_feasibility,config,prompt_template}',
        to_jsonb(replace(
            default_config #>> '{workflow,steps,review_feasibility,config,prompt_template}',
            'i.e. in /data/provocations.json or client-computable — with NO server; (c) does the plan depend on anything unbuilt (e.g. the daily emitter) without saying so and sequencing it; (d) is the client-side scoring/timer for the minimal-real Gauntlet genuinely doable without a backend.',
            'i.e. served by a data file or endpoint THIS site actually has, as named in §3, or client-computable — with NO server; (c) does the plan depend on anything unbuilt (an emitter, a feed, an endpoint that does not exist yet) without saying so and sequencing it; (d) is any client-side computation, timing or scoring the MVP relies on genuinely doable without a backend. Judge the plan against THIS site: do not object that it lacks a feed, a tool or a loop that another site has.'
        ))),

        -- 2g. the HARD-veto honesty seat stops carrying another site's (stale)
        --     evidence_base premise as its rule. See "(b)" in the header: the
        --     zero-facts assertion went false on 2026-08-08 08:58Z.
        '{workflow,steps,review_honesty,config,prompt_template}',
        to_jsonb(replace(replace(
            default_config #>> '{workflow,steps,review_honesty,config,prompt_template}',
            '— vonc''s evidence_base has ZERO facts, so ANY hard number that is not read from real client state is fabricated;',
            '— a number is fabricated unless it is read from real client state at runtime, or traces to a fact in THIS site''s own evidence_base spec. Do not assume this site has facts, and do not assume it has none: if the plan cites one, it must say which;'),
            'This is the anti-fabrication rule that the current Gauntlet violates (12,847 competitors, a fake Live leaderboard). The plan must not reproduce it in any form.',
            'This rule exists because a tool on this estate once shipped fabricated competitor counts and a fake "Live" leaderboard. The plan must not reproduce that in any form.'
        ))),

        -- 2h. reframe — the step that rewrites a VETOED plan, and therefore the
        --     one that produced the second contaminated plan in the 227 run.
        '{workflow,steps,reframe,config,prompt_template}',
        to_jsonb(replace(
            default_config #>> '{workflow,steps,reframe,config,prompt_template}',
            'If the Gauntlet is what was vetoed and no honest minimal-real round is achievable, demote it to a labelled coming-soon panel and move the real round to the LATER list — that is an acceptable honest MVP.',
            'If the vetoed feature admits no honest minimal-real version, demote it to a labelled coming-soon panel and move the real version to the LATER list — that is an acceptable honest MVP.'
        )))
WHERE type = 'experience-planner'
  AND is_active AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;

-- ============================================================================
-- VERIFY — a DO block that RAISEs, because a verify block of bare SELECTs cannot
-- stop the COMMIT (ON_ERROR_STOP ignores a non-empty result).
-- ============================================================================
DO $$
DECLARE
    n int;
BEGIN
    SELECT count(*) INTO n
      FROM agent_definitions
     WHERE type = 'experience-planner'
       AND is_active AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL
       AND default_config #>> '{workflow,steps,load_schema_hint,next_step}' = 'load_brief'
       AND default_config #>> '{workflow,steps,load_brief,next_step}' = 'compose'
       AND default_config #>> '{workflow,steps,load_brief,output_field}' = 'experience_brief'
       AND default_config #> '{workflow,steps,compose,config,input_fields}' @> '["experience_brief"]'::jsonb
       AND default_config #>> '{workflow,steps,compose,config,prompt_template}' LIKE '%{{.experience_brief.text}}%'
       -- The whole point: no site-specific vocabulary left ANYWHERE in the row.
       -- Case-insensitive and the same five terms as the header census — a
       -- case-sensitive check here would pass while "Gauntlet" survived in two
       -- steps, which is exactly how those two steps were missed in the first place.
       AND default_config::text !~* 'provocation|gauntlet|arena|vonc|spark'
       AND default_config::text NOT ILIKE '%tools.apis.uk%';
    IF n <> 1 THEN
        RAISE EXCEPTION '227/345 VERIFY FAILED: expected exactly 1 fully-updated active row, found %. (A surgical replace() that matched nothing leaves the old text in place and is the likeliest cause — check which of the five terms still matches.)', n;
    END IF;

    -- vonc must not be left briefless: the channel has to hold what the prompt lost.
    IF NOT EXISTS (
        SELECT 1 FROM doc_notes
         WHERE subject_type = 'experience' AND subject_key = 'vonc-spark-game'
           AND categories ? 'experience-brief'
           AND body LIKE '%D2 HARD CONSTRAINT%'
           AND body LIKE '%tools.apis.uk%'
    ) THEN
        RAISE EXCEPTION '227/345 VERIFY FAILED: vonc-spark-game has no experience-brief note carrying D1/D2 and the API contract — the prompt would lose its brief with nothing to replace it';
    END IF;

    -- the snapshot must hold the PRE-change config or it restores nothing
    IF NOT EXISTS (
        SELECT 1 FROM agent_definitions_backup
         WHERE type = 'experience-planner'
           AND snapshot_reason LIKE 'pre-update: 227%'
           AND default_config::text ILIKE '%provocation%'
    ) THEN
        RAISE EXCEPTION '227/345 VERIFY FAILED: no backup row carrying the PRE-change (contaminated) config';
    END IF;
END $$;

COMMIT;
