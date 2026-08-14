-- 408_offer_analyser_agent.sql
--
-- Programme B, phase B4 (PLAN_2026-08-02, vigilant_designer_offer_analysis):
-- the offer and benefit analyser. `features_open/030` holds the wider scope;
-- this migration ships the config-only v1 the owner approved on 2026-08-14.
--
-- THE QUESTION IT EXISTS TO ASK, per site, for ever: does this site actually
-- answer its market's need, in a way that pays us? Not "is it well built" (the
-- designer's question) and not "is it true" (the claims auditor's question).
--
-- TWO OUTPUTS FROM ONE ANALYSIS (owner decision, 2026-08-14 evening — the
-- alternative shapes considered were auditor-first and ordering-first):
--
--   1. `site_specs` aspect `offer_ordering` — a RANKED "what is this site's
--      reader trying to do, so what should a page lead with?", per site.
--      NEW ARTEFACT, and the reason it exists is a named external consumer:
--      the `copy_quality_two_stage` lane asked for exactly this in
--      CONTRIB_2026-08-12. Three of the four inputs it wanted already existed
--      in `site_specs` aspect `strategy` (satisfaction_condition,
--      value_proposition, trust_threshold, recurring_value — B2 restored them,
--      22 of 22 sites carry them as of 2026-08-14). What did NOT exist, on any
--      site, was any ORDERING of them: four prose paragraphs a human can read
--      and no pass can sort by. That is what this writes down.
--   2. `site_work_items` under audit_source `offer-analysis` — findings where
--      the live surface disagrees with that ordering or with the site's own
--      recorded revenue model. Existing item types, existing handlers, no new
--      routing (the lane's rule: every detector ships with its drain, or it
--      does not ship — bugs_open/115 is the worked case of the alternative).
--
-- ── FOUR THINGS THIS CONFIG IS SHAPED BY, each measured, not assumed ────────
--
-- (a) `bugs_open/272` — THE FINDINGS PATH SILENTLY DROPS AN OBJECT.
--     `write_audit_findings`'s parse switch handles a JSON string and a JSON
--     array and has NO case for a JSON object. `site-review-agent` asks its LLM
--     for `{"overall_score":…,"findings":[…]}` and points `findings_field` at
--     the object one level above the array — so it has filed ZERO work items,
--     ever (checked 2026-08-14: no row anywhere carries audit_source
--     'site-review'). This agent returns an object too, but points
--     `findings_field` at `offer_analysis.result.findings` — the ARRAY —
--     which hits the working `case []interface{}`. That is 272's own fix
--     candidate 1, applied here at birth rather than inherited as a defect.
--     ⚠ If 272 candidate 2 (the missing map case) later lands in Go, this
--     config still works and must not be "simplified" back to the object path.
--
-- (b) ROUTING IS BY `category`, NOT BY `work_item_type`.
--     write_audit_findings classifies deterministically in Go from the
--     finding's `category` (+ whether the named page exists). The
--     `work_item_type` field that site-review-agent's prompt asks for is read
--     by NOTHING. So the closed vocabulary this prompt enforces is a CATEGORY
--     vocabulary, and every one of the seven allowed values has a live route:
--       gap            → needs_content_page / page-build-handler   (page exists)
--                        needs_content_planning / content-gap-planner (it does not)
--       content        → content_rewrite / page-build-handler
--       differentiation→ content_rewrite / page-build-handler
--       structure      → content_rewrite / page-build-handler
--       cta            → cta_improvement / component-template-fixer
--       nav_restructure→ nav_restructure / component-template-fixer
--       tone           → tone_shift / page-build-handler
--     An OFF-vocabulary category does not fail — it mints `audit_finding_<x>`
--     aimed at content-gap-planner, an item type no verifier knows (six such
--     rows exist fleet-wide today, five of them still `detected`). That is why
--     the prompt states the list and says what happens if it is departed from.
--
-- (c) THE DEGRADED VERDICT IS COMPUTED IN SQL, NOT JUDGED BY THE LLM.
--     `load_premise` returns `premise_fields_missing` — the names of the four
--     premise fields that are empty on THIS site — and the prompt copies it
--     verbatim into the artefact. The lane has been bitten twice by a check
--     that examines less on some sites and does not say so, because the
--     resulting silence reads downstream as a clean bill (WII-014 is the whole
--     fix for one instance; bugs_open/255 the other). Today exactly one site
--     is degraded: leopardessconsulting.co.uk has no `recurring_value`, left
--     out by owner decision on 2026-08-14 because the donor prose for it was a
--     fabrication. The field being ABSENT rather than false is the protection;
--     this states it rather than papering over it.
--
-- (d) `write_site_spec` DEEP-MERGES, so an omitted key SURVIVES.
--     Maps recurse, arrays are replaced wholesale (siteSpecDeepMerge,
--     site_spec_actions.go:513). So a re-run replaces `lead_with` cleanly —
--     but a key the model forgets keeps the PREVIOUS run's value while looking
--     current. The prompt therefore requires every key on every run, including
--     empty arrays and `false`. Freshness is the ROW's (`created_at` /
--     `is_current`), never a timestamp inside the payload: an LLM asked for
--     the time invents one.
--
-- ── HONESTY CONSTRAINT (features_open/030 §4) ──────────────────────────────
-- This platform collects NO engagement data. The analyser can grade the
-- artefact against the stated premise and nothing else. The prompt forbids
-- "users want…" phrasing outright, because an offer analyser that sounds like
-- it knows what converts, while reading only our own specs, would be the most
-- confidently wrong instrument on the estate.
--
-- ── AND ITS INPUTS ARE UNVERIFIED PROSE ────────────────────────────────────
-- Nothing on this estate claim-checks a `site_specs` row: `check_unverified_
-- claims` scans deployed HTML and stored `content_data`, never specs, and it
-- never repairs anything by design. A false sentence in a premise is invisible
-- and would be graded against here (bugs_open/161's shape, one layer up). The
-- owner ruled on 2026-08-14 that the claims audit should be extended to cover
-- `site_specs` prose; until it is, this agent's verdicts inherit whatever the
-- strategist wrote. Stated, not solved.
--
-- ── THE OFFER SURFACE, and why this predicate ──────────────────────────────
-- `load_offer_surface` uses the LINKABILITY floor (lifecycle `status='active'`
-- + NOT(deployed_at IS NULL AND build_status='planned')), which is the SQL of
-- datahelpers.PageMayBeLinkedPredicateFor — not PageHasShippedPredicateFor.
-- The question here is "what can a visitor actually reach", and the stricter
-- predicate would drop 11 pages that serve HTTP 200 (measured fleet-wide
-- 2026-08-09, links.go:297-317). ⚠ This is an INLINED COPY of a shared Go
-- predicate — a config-only agent cannot call the helper. If links.go's
-- linkability floor changes, this string does not follow it. Registered as
-- such in the concept register.
--
-- Sizing, measured 2026-08-14: worst-case prompt ≈ 47KB (strategy specs run
-- 12–17KB; the biggest offer surface, webdesign.co.uk at 101 reachable pages,
-- is 14.9KB). B1's live comparator ran 29KB with 1,763 output tokens against a
-- 4,000 cap. max_tokens here is 8,000 for two outputs; `output_tokens ==
-- max_tokens` means the completion was CUT, so read `__truncated` on the step
-- result, never just its status.
--
-- Dispatch envelope (standalone, how B5's proof is run):
--   topic system.agent.generic.requests, action=orchestrate,
--   config={"agent_type":"offer-analyser"},
--   input_data={"site_id":"<uuid>","domain":"<domain>"}
-- Wiring into improvement-loop (due-gated, PLAN §B4) is a SEPARATE migration,
-- deliberately: this one is inert until something calls it.
--
-- ROLLBACK RECIPE (also in 408_offer_analyser_agent_ROLLBACK.sql):
--   UPDATE agent_definitions SET is_active=false, deleted_at=now()
--    WHERE type='offer-analyser' AND deleted_at IS NULL;
-- The `offer_ordering` spec rows it has written are left in place by that
-- rollback — they are data, and nothing else reads them yet.

BEGIN;

INSERT INTO agent_definitions (type, display_name, description, category, agent_category, status, is_active, default_config)
SELECT
  'offer-analyser',
  'Offer and Benefit Analyser',
  'Reads one site''s recorded premise and its reachable page surface, and answers two questions: what should this site lead with for its own reader (written to site_specs aspect offer_ordering, ranked), and where does the live site fail to (work items under audit_source offer-analysis, existing types and handlers only). Judges artefact against premise only — this platform has no visitor behaviour data. Dispatch with input_data {site_id, domain}.',
  'analyst',
  'analyst',
  'experimental',
  true,
  jsonb_build_object('workflow', jsonb_build_object(
    'start_step', 'ensure_site_record',
    'processing_mode', 'orchestrator',
    'timeout_seconds', 900,
    'steps', jsonb_build_object(

      'ensure_site_record', jsonb_build_object(
        'action', 'ensure_site_record',
        'config', jsonb_build_object('store_brief_in_content_data', false),
        'next_step', 'load_premise',
        'description', 'Resolve the site row from domain/site_id without touching content_data',
        'output_field', 'site_record'
      ),

      'load_premise', jsonb_build_object(
        'action', 'query_database',
        'config', jsonb_build_object(
          'query', $q1$WITH st AS (
  SELECT sp.data FROM site_specs sp
   WHERE sp.site_id = $1 AND sp.aspect = 'strategy' AND sp.is_current = true
   ORDER BY sp.created_at DESC LIMIT 1
)
SELECT s.domain,
       COALESCE(s.company_name, s.domain) AS company,
       COALESCE((SELECT data::text FROM st), '{}') AS strategy,
       COALESCE((SELECT data->'revenue_models'->>'primary_model' FROM st), '') AS primary_model,
       array_to_string(ARRAY(
         SELECT f FROM unnest(ARRAY['satisfaction_condition','trust_threshold','recurring_value','value_proposition']) AS f
          WHERE COALESCE((SELECT data->>f FROM st), '') = ''
       ), ', ') AS premise_fields_missing,
       COALESCE((SELECT sp.data::text FROM site_specs sp WHERE sp.site_id = s.id AND sp.aspect = 'audience' AND sp.is_current = true ORDER BY sp.created_at DESC LIMIT 1), '{}') AS audience,
       COALESCE((SELECT left(sp.data::text, 3000) FROM site_specs sp WHERE sp.site_id = s.id AND sp.aspect = 'identity' AND sp.is_current = true ORDER BY sp.created_at DESC LIMIT 1), '{}') AS identity_head_3k,
       COALESCE((SELECT left(sp.data::text, 3000) FROM site_specs sp WHERE sp.site_id = s.id AND sp.aspect = 'content_direction' AND sp.is_current = true ORDER BY sp.created_at DESC LIMIT 1), '{}') AS content_direction_head_3k,
       COALESCE((SELECT sp.data->>'text' FROM site_specs sp WHERE sp.site_id = s.id AND sp.aspect = 'mission_brief' AND sp.is_current = true ORDER BY sp.created_at DESC LIMIT 1), '') AS mission_brief
FROM sites s WHERE s.id = $1$q1$,
          'params', jsonb_build_array('site_record.site_id'),
          'output_format', 'object'
        ),
        'next_step', 'load_offer_surface',
        'description', 'The site''s own premise. premise_fields_missing is computed here, in SQL, so a degraded analysis is a stated field rather than an absence the reader has to notice',
        'output_field', 'premise'
      ),

      'load_offer_surface', jsonb_build_object(
        'action', 'query_database',
        'config', jsonb_build_object(
          'query', $q2$SELECT COALESCE(string_agg(
         format('%s | type=%s | nav=%s | title=%s | meta=%s',
                p.name, COALESCE(NULLIF(p.page_type,''),'-'),
                CASE WHEN p.in_header THEN 'header' WHEN p.in_footer THEN 'footer' ELSE 'not-in-nav' END,
                COALESCE(NULLIF(p.title,''),'(none)'),
                left(COALESCE(NULLIF(p.meta_description,''),'(none)'), 160)),
         E'\n' ORDER BY p.in_header DESC, p.nav_order, p.name), '(no reachable pages)') AS page_list,
       count(*) AS page_count,
       count(*) FILTER (WHERE p.in_header) AS in_header_count
FROM pages p
WHERE p.site_id = $1
  AND p.status = 'active'
  AND NOT (p.deployed_at IS NULL AND COALESCE(p.build_status,'') = 'planned')$q2$,
          'params', jsonb_build_array('site_record.site_id'),
          'output_format', 'object'
        ),
        'next_step', 'run_offer_analysis',
        'description', 'Every page a visitor can actually reach — the linkability floor (PageMayBeLinkedPredicateFor), inlined because a config-only agent cannot call the Go helper',
        'output_field', 'offer_surface'
      ),

      'run_offer_analysis', jsonb_build_object(
        'action', 'execute_llm_prompt',
        'config', jsonb_build_object(
          'prompt', $prompt$You are an offer and benefit analyst. You are looking at ONE live website together with the premise that site recorded for itself. Answer two questions, in this order.

Domain: {{.premise.domain}}
Company: {{.premise.company}}
Recorded revenue model (revenue_models.primary_model): {{.premise.primary_model}}

RECORDED STRATEGY — this site''s own premise. Everything you say must be grounded in it:
{{.premise.strategy}}

Premise fields that are EMPTY on this site (computed from the database, not your judgement): [{{.premise.premise_fields_missing}}]

Target audience (recorded at build time):
{{.premise.audience}}

Identity (first 3000 chars):
{{.premise.identity_head_3k}}

Content direction (first 3000 chars):
{{.premise.content_direction_head_3k}}

Mission brief (the operator''s own words; empty if none was given):
{{.premise.mission_brief}}

THE OFFER SURFACE — every page a visitor can reach. {{.offer_surface.page_count}} pages, of which {{.offer_surface.in_header_count}} are in the header nav:
{{.offer_surface.page_list}}

HONESTY CONSTRAINT. You have NO visitor behaviour data of any kind, and this platform collects none. Never write "users want", "visitors expect", "readers prefer", "this converts better", or any other claim about what people do. Every judgement you make is a comparison between the RECORDED PREMISE and the BUILT ARTEFACT, and must be phrased as one. If a point cannot be grounded in a premise field above or in a page on the list above, do not make it.

DEGRADATION. If the list of empty premise fields above is not empty, you are judging on less than a full premise. Copy that list verbatim into ordering.inputs_missing and set ordering.degraded to true. Do NOT infer a missing field''s content from the others: its absence is a stated limit on this analysis, not a gap for you to fill.

TASK 1 — THE ORDERING. In the recorded strategy, satisfaction_condition says what this reader is trying to achieve; value_proposition says what this site offers that others do not; trust_threshold says what this reader needs before acting; recurring_value says why they come back. Those are four prose paragraphs. Turn them into a RANKED list of what a page on this site should lead with — most beneficial to the reader first, and among equally beneficial points, the most differentiated first.

Rank at most 6 points. Each "point" is one sentence a page could actually open with: a benefit to the reader, never a description of us or of our inventory. Each carries "from_field", naming which premise field it came from, so a later reader can check it.

Also name what a page here should NOT open with ("avoid_leading_with"): typically our own catalogue or page count, our own history, or any self-description carrying no reader benefit.

This ordering is written to the site''s record and other agents will read it as the authority on what matters here. Write it to be used, not to be admired.

TASK 2 — THE FINDINGS. Report at most 5, and only where the live surface disagrees with the ordering you just wrote, or with the recorded revenue model. Fewer is better. Report none at all if none is true — an empty findings array is a valid and useful answer.

"category" MUST be exactly one of these seven strings: gap, content, differentiation, structure, cta, nav_restructure, tone. Any other value routes the finding to a bucket no handler reads, and it is lost. Choose by what must change:
  gap             — something the premise promises has no page at all
  content         — an existing page''s words do not deliver what the premise says
  differentiation — the page leads with what any competitor could equally say
  structure       — the right material exists, in the wrong order or the wrong place
  cta             — what a page asks the visitor to do does not match the recorded revenue model
  nav_restructure — the reachable set is wrong: the header buries what matters most
  tone            — the register is wrong for the reader''s recorded trust_threshold

"page" MUST be one of the exact page names listed in THE OFFER SURFACE above, or the literal string "site-wide". A name not on that list is read downstream as a request to create a NEW page, so do not paraphrase a page name.

"acceptance_test" must be concrete enough that a DIFFERENT agent could check it without reading this analysis. Good: "The index page opens by naming what the visitor''s existing debt does to their mortgage options, before any list of calculators". Bad: "The homepage should be more benefit-led".

OUTPUT. Return ONE JSON object and nothing else — no prose before or after, no markdown fences:
{"ordering": {"reader_goal": "one sentence: what this site''s reader is trying to achieve", "lead_with": [{"rank": 1, "point": "...", "why": "...", "from_field": "value_proposition", "differentiated": true}], "avoid_leading_with": ["..."], "inputs_missing": [], "degraded": false, "primary_model": "{{.premise.primary_model}}", "spec_version": 1}, "findings": [{"category": "differentiation", "severity": "high", "description": "what is wrong", "current_value": "what is there now", "suggestion": "the specific change", "acceptance_test": "a concrete testable criterion", "page": "index", "max_fix_attempts": 2}]}

EVERY key shown in "ordering" must be present on every run, including empty arrays and false. That object is merged over whatever this site''s ordering already holds, so a key you omit silently leaves the previous run''s value standing and looking current.$prompt$,
          'ai_service', jsonb_build_object(
            'model', 'claude-sonnet-4-6',
            'provider', 'anthropic',
            'max_tokens', 8000,
            'api_key_env_var', 'ANTHROPIC_API_KEY'
          ),
          'input_fields', jsonb_build_array('premise', 'offer_surface')
        ),
        'next_step', 'set_audit_source',
        'description', 'One analysis, two outputs: the ranked ordering (Task 1) and the artefact-vs-premise findings (Task 2)',
        'output_field', 'offer_analysis'
      ),

      'set_audit_source', jsonb_build_object(
        'action', 'query_database',
        'config', jsonb_build_object(
          'query', $q3$SELECT 'offer-analysis'::text AS audit_source$q3$,
          'output_format', 'object'
        ),
        'next_step', 'write_offer_ordering',
        'description', 'audit_source is Required with no default since bugs_open/264 — a config literal that fails to resolve now fails the action outright, so it arrives as a resolved query result exactly as the other four auditors do',
        'output_field', 'audit_source_literal'
      ),

      'write_offer_ordering', jsonb_build_object(
        'action', 'write_site_spec',
        'config', jsonb_build_object(
          'site_id', 'site_record.site_id',
          'spec_data', 'offer_analysis.result.ordering',
          'aspect', 'offer_ordering',
          'source', 'offer-analyser',
          'source_agent', 'offer-analyser',
          'notes', 'Ranked reader-priority ordering derived from the strategy spec (satisfaction_condition, value_proposition, trust_threshold, recurring_value). Written by offer-analyser (B4). First named consumer: the copy_quality_two_stage lane. Judged against the premise only — no visitor behaviour data exists on this platform.'
        ),
        'next_step', 'write_offer_findings',
        'description', 'The new artefact. NO error_step by design: if this write fails the run fails visibly — the findings are re-derivable next run, a silently missing ordering is not',
        'output_field', 'offer_ordering_written'
      ),

      'write_offer_findings', jsonb_build_object(
        'action', 'write_audit_findings',
        'config', jsonb_build_object(
          'site_id', 'site_record.site_id',
          'audit_source', 'audit_source_literal.audit_source',
          'findings_field', 'offer_analysis.result.findings'
        ),
        'next_step', 'complete',
        'description', 'findings_field points at the ARRAY, not at the object above it — bugs_open/272: write_audit_findings has no parse case for an object and drops it silently, which is why site-review-agent has never filed a finding',
        'output_field', 'offer_findings_written'
      ),

      'complete', jsonb_build_object(
        'action', 'complete_workflow',
        'config', jsonb_build_object(
          'output_fields', jsonb_build_array('offer_analysis', 'offer_ordering_written', 'offer_findings_written')
        )
      )

    )
  ))
WHERE NOT EXISTS (
  SELECT 1 FROM agent_definitions
  WHERE type = 'offer-analyser' AND deleted_at IS NULL
);

DO $verify$
DECLARE
  cfg      jsonb;
  n_steps  integer;
  prompt   text;
BEGIN
  SELECT default_config INTO cfg
  FROM agent_definitions
  WHERE type = 'offer-analyser' AND is_active
    AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;

  IF cfg IS NULL THEN
    RAISE EXCEPTION 'verify: offer-analyser row absent or inactive after insert';
  END IF;

  SELECT count(*) INTO n_steps FROM jsonb_object_keys(cfg->'workflow'->'steps');
  IF n_steps <> 8 THEN
    RAISE EXCEPTION 'verify: expected 8 workflow steps, found %', n_steps;
  END IF;

  -- (a) bugs_open/272: the findings path must address the ARRAY. Pointing this
  -- at the object silently files nothing, which is indistinguishable from a
  -- clean site.
  IF (cfg->'workflow'->'steps'->'write_offer_findings'->'config'->>'findings_field')
       IS DISTINCT FROM 'offer_analysis.result.findings' THEN
    RAISE EXCEPTION 'verify: findings_field must be offer_analysis.result.findings (the array) — an object is dropped silently by write_audit_findings (bugs_open/272)';
  END IF;

  -- (b) the two writers must address DIFFERENT sub-paths of one analysis.
  IF (cfg->'workflow'->'steps'->'write_offer_ordering'->'config'->>'spec_data')
       IS DISTINCT FROM 'offer_analysis.result.ordering' THEN
    RAISE EXCEPTION 'verify: the ordering write must address offer_analysis.result.ordering';
  END IF;
  IF (cfg->'workflow'->'steps'->'write_offer_ordering'->'config'->>'aspect')
       IS DISTINCT FROM 'offer_ordering' THEN
    RAISE EXCEPTION 'verify: the ordering must be written to its OWN aspect — writing it into strategy would supersede a record two sites hold by owner/hitl decision';
  END IF;

  -- (c) audit_source must resolve from the literal step, not be a bare string.
  IF (cfg->'workflow'->'steps'->'write_offer_findings'->'config'->>'audit_source')
       IS DISTINCT FROM 'audit_source_literal.audit_source' THEN
    RAISE EXCEPTION 'verify: audit_source must resolve from set_audit_source (bugs_open/264)';
  END IF;

  prompt := cfg->'workflow'->'steps'->'run_offer_analysis'->'config'->>'prompt';

  -- (d) the honesty constraint is the single most important line in the prompt
  -- (features_open/030 §4): with zero outcome data, an analyser that sounds
  -- like it knows what converts is the most confidently wrong thing we could ship.
  IF prompt NOT LIKE '%HONESTY CONSTRAINT%' OR prompt NOT LIKE '%NO visitor behaviour data%' THEN
    RAISE EXCEPTION 'verify: the honesty constraint has been edited out of the prompt';
  END IF;

  -- (e) the closed category vocabulary IS the routing contract (write_audit_
  -- findings classifies on category; work_item_type is read by nothing).
  IF prompt NOT LIKE '%gap, content, differentiation, structure, cta, nav_restructure, tone%' THEN
    RAISE EXCEPTION 'verify: the closed category vocabulary is missing — an off-vocabulary category mints an item type no verifier knows';
  END IF;

  -- (f) the degradation instruction, without which a thinner analysis is
  -- indistinguishable from a clean one.
  IF prompt NOT LIKE '%inputs_missing%' OR prompt NOT LIKE '%degraded%' THEN
    RAISE EXCEPTION 'verify: the degradation instruction is missing from the prompt';
  END IF;

  IF (cfg->'workflow'->'steps'->'run_offer_analysis'->'config'->'ai_service'->>'max_tokens')::int < 8000 THEN
    RAISE EXCEPTION 'verify: max_tokens below 8000 — two outputs in one completion, and output_tokens = max_tokens means the answer was CUT';
  END IF;
END $verify$;

COMMIT;
