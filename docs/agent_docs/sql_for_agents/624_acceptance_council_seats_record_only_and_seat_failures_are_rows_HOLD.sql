-- 624_acceptance_council_seats_record_only_and_seat_failures_are_rows_HOLD.sql
--
-- _HOLD: NOT for the runner. Phase 3 of the plan in
-- docs/agent_docs/docs024_key_docs_latest/loanzy_uk_example_site/
--   PLAN_2026-08-25_switch_off_the_evolutionary_rewrites_and_switch_the_loop_back_on.md
-- and the config half of RFC_056. Apply BY HAND, and ONLY after BOTH:
--   (1) a chassis roll carrying `filing_mode` on write_audit_findings and the three
--       checks build_prerequisites / heading_promise / structure_floor
--       (prove it the way that WORKS — the CLAUDE.md 'build provenance' log grep is
--        REFUTED [MEASURED 2026-08-25, bugs_open/395 lane, confirmed by the vigilant
--        lane]: that string is emitted nowhere in this repo's Go source. Use:
--          SELECT git_commit FROM service_binary_capabilities
--           WHERE service='agent-chassis' AND kind='build'
--           ORDER BY last_seen_at DESC LIMIT 1;
--        then `git merge-base --is-ancestor c440d5c5e <that sha>` PLUS one
--        must-be-absent and one must-be-present control commit in the same breath.
--        [MEASURED 2026-08-25 19:5xZ] the running chassis a7459a44b… ALREADY carries
--        c440d5c5e with both controls behaving — precondition (1) is met today);
--   (2) the council verdict on RFC_056's submission has been READ.
-- Applied against an older binary, part C makes every model seat's findings
-- dispatch exactly as before (an unknown config key is ignored), part B names
-- three checks the runner warns "Unknown discovery check" about and skips, and
-- the reader seat dispatches rewrites. None of that is destructive, but all of
-- it is the opposite of what this file claims to do - so the probe below
-- refuses to run unless the binary can honour it (see part 0).
--
-- WHAT THIS DOES, in five parts, one transaction:
--   0. refuses unless the running chassis reports a commit that carries filing_mode
--      (a migration cannot ask the binary what it can do - RFC_040 - so it asks the
--      operator: the FILING_MODE_SHIPPED literal below must be set to the stamp).
--   A. creates `acceptance-discovery-agent` - a 3-step discovery agent (the shape of
--      completeness-discovery-agent) whose checks array names exactly the three new
--      mechanical seats, on its own timeout, so two fetch-heavy checks do not ride
--      inside completeness's 120s call.
--   B. creates `reader-experience-auditor` - the fourth seat, an LLM seat in the
--      shape of brief-fidelity-auditor, judging PURPOSE (would a person on this kind
--      of site want this page?) rather than fidelity or tone, with
--      filing_mode: record from birth.
--   C. sets filing_mode: record on the write_audit_findings step of all five
--      existing model seats (visual-design-auditor, content-quality-auditor,
--      site-review-agent, offer-analyser, brief-fidelity-auditor). From here on their
--      findings are VERDICT rows: status deferred, handler '', routing kept in spec.
--      THE SEAT ARITHMETIC, in one place because council round d1342f2a caught the
--      three phrasings drifting: the loop makes FOUR model-seat CALLS
--      (design_audit, site_review, offer_analyser, brief_fidelity); design-audit is
--      a compound seat whose TWO children (visual-design-auditor,
--      content-quality-auditor) do the filing, so FIVE existing write steps get
--      record mode here, and with the reader seat that is SIX write steps — which is
--      what part E verifies. ⚠ STATED GAP: the seat-failure rows of part D are at
--      the LOOP-CALL level; design-audit-agent's own internal error routes
--      (call_visual_auditor error -> spawn_content_auditor, call_content_auditor
--      error -> complete) swallow a CHILD failure without erroring the call, so a
--      child-level failure leaves no row and does not withhold the pass stamp.
--      Closing that needs an edit INSIDE design-audit-agent's workflow — a named
--      follow-up in RFC_056 addendum 2, not smuggled into this file.
--   D. rewires improvement-loop: mechanical seats -> acceptance seats -> the four
--      model seats -> the reader seat -> check_seats_ran -> record_audit_pass; and
--      every seat call plus both enrichment steps gets a record_<seat>_failed step
--      (one deferred capability_gap row per site per seat, recurrence_expected, the
--      record_not_converging shape) that CONTINUES the sweep - one auditor must not
--      strand it - while check_seats_ran withholds the pass stamp when any seat
--      failed this run. Two spellings of error_step exist (step-level and
--      config-level); this file sets the step-level one everywhere it adds a route
--      and REMOVES the config-level twin on the two enrichment steps so there is one.
--   E. verifies: every edge resolves, every model seat carries filing_mode=record,
--      the condition carries no parentheses, and the step count is exactly what
--      this file expects.
--
-- WHAT THIS DOES NOT DO: it does not enable improvement-sweep (Phase 1, the
-- owner's row); it does not touch the render-audit rotation; it does not release
-- any record row (the recipe is on the row).

\set ON_ERROR_STOP on

-- Part 0: the operator asserts the binary. Replace the literal with the stamp
-- read from the running chassis; the DO block refuses the placeholder.
-- Run as:  psql -v FILING_MODE_SHIPPED=<sha> -f 624_...HOLD.sql   (the placeholder below
-- is set ONLY when the operator passed nothing, and the probe refuses it).
\if :{?FILING_MODE_SHIPPED}
\else
\set FILING_MODE_SHIPPED 'REPLACE-WITH-THE-CHASSIS-BUILD-PROVENANCE-SHA'
\endif
-- psql does not interpolate :'vars' inside a dollar-quoted DO body, so the value is
-- staged in a temp table by plain SQL and read from there.
DROP TABLE IF EXISTS _624_operator;
CREATE TEMP TABLE _624_operator AS SELECT :'FILING_MODE_SHIPPED'::text AS sha;

DO $probe$
DECLARE v_sha text;
BEGIN
    IF EXISTS (SELECT 1 FROM agent_definitions WHERE type = 'reader-experience-auditor' AND deleted_at IS NULL) THEN
        RAISE EXCEPTION '624: already applied - reader-experience-auditor exists';
    END IF;
    SELECT sha INTO v_sha FROM _624_operator;
    IF v_sha = 'REPLACE-WITH-THE-CHASSIS-BUILD-PROVENANCE-SHA' OR length(v_sha) < 7 THEN
        RAISE EXCEPTION '624: refuse - set FILING_MODE_SHIPPED (-v FILING_MODE_SHIPPED=<sha>) to the chassis build-provenance sha you have verified carries filing_mode (part 0 of the header)';
    END IF;
    RAISE NOTICE '624: operator asserts filing_mode shipped at chassis %', v_sha;
END $probe$;

BEGIN;

SELECT snapshot_agent('improvement-loop',           '624_acceptance_council_seats_record_only_and_seat_failures_are_rows_HOLD: pre-update');
SELECT snapshot_agent('visual-design-auditor',      '624: pre-update (filing_mode)');
SELECT snapshot_agent('content-quality-auditor',    '624: pre-update (filing_mode)');
SELECT snapshot_agent('site-review-agent',          '624: pre-update (filing_mode)');
SELECT snapshot_agent('offer-analyser',             '624: pre-update (filing_mode)');
SELECT snapshot_agent('brief-fidelity-auditor',     '624: pre-update (filing_mode)');
SELECT snapshot_agent('completeness-discovery-agent','624: shape source for acceptance-discovery-agent (not edited)');

-- ── Part A: acceptance-discovery-agent ───────────────────────────────────────
INSERT INTO agent_definitions
    (type, display_name, description, category, default_config, is_active, capabilities,
     image_repository, image_tag, command, resources, topics, health_config, env_vars, version,
     delegation_preferences, agent_category, status, domain_tags, idle_timeout_seconds)
SELECT 'acceptance-discovery-agent',
       'Acceptance discovery agent (RFC_056 mechanical seats)',
       'Runs the three mechanical seats of the site acceptance council - build_prerequisites, heading_promise, structure_floor - as discovery checks. All flag-only: verdict rows, never dispatched. Own agent so two fetch-heavy checks run on their own timeout rather than inside completeness-discovery-agent''s 120s call. RFC_056, 2026-08-25.',
       category,
       jsonb_set(jsonb_set(default_config,
           '{workflow,steps,run_checks,config,checks}', '["build_prerequisites","heading_promise","structure_floor"]'::jsonb, true),
           '{workflow,steps,run_checks,config,check_pipeline}', '"content"'::jsonb, true),
       true, capabilities, image_repository, image_tag, command, resources, topics, health_config, env_vars, 1,
       delegation_preferences, agent_category, 'experimental', domain_tags, idle_timeout_seconds
  FROM agent_definitions
 WHERE type = 'completeness-discovery-agent'
   AND is_active AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL
 ORDER BY version DESC LIMIT 1;

DO $a$ BEGIN
    IF (SELECT default_config #> '{workflow,steps,run_checks,config,checks}' FROM agent_definitions WHERE type='acceptance-discovery-agent' AND deleted_at IS NULL)
       IS DISTINCT FROM '["build_prerequisites","heading_promise","structure_floor"]'::jsonb THEN
        RAISE EXCEPTION '624 A: acceptance-discovery-agent not created with the three checks (was completeness-discovery-agent''s run_checks step renamed?)';
    END IF;
END $a$;

-- ── Part B: reader-experience-auditor ────────────────────────────────────────
INSERT INTO agent_definitions
    (type, display_name, description, category, default_config, is_active, capabilities,
     image_repository, image_tag, command, resources, topics, health_config, env_vars, version,
     delegation_preferences, agent_category, status, domain_tags, idle_timeout_seconds)
SELECT 'reader-experience-auditor',
       'Reader experience auditor (RFC_056 reader seat)',
       'The site acceptance council''s reader seat: would a person who came to THIS KIND of site want THIS page? Judges purpose - not fidelity to the brief (brief-fidelity-auditor), not tone (content-quality-auditor). Files VERDICT rows only (filing_mode: record): nothing it says is dispatched until a person releases the row. Owner asked for it three times in three words on 2026-08-25 ("user experience agent", "happy user", "adversarially increase the quality").',
       category,
       jsonb_build_object(
         'processing_mode', 'orchestrator',
         'timeout_seconds', 300,
         'workflow', jsonb_build_object(
           'start_step', 'ensure_site_record',
           'steps', jsonb_build_object(
             'ensure_site_record', jsonb_build_object(
               'action', 'ensure_site_record',
               'config', jsonb_build_object('store_brief_in_content_data', false),
               'next_step', 'load_reader_context',
               'output_field', 'site_record'),
             'load_reader_context', jsonb_build_object(
               'action', 'query_database',
               'config', jsonb_build_object(
                 'query', $q$SELECT
  (SELECT LEFT(data::text, 2500) FROM site_specs WHERE site_id=s.id AND aspect='classification' AND is_current) AS classification,
  (SELECT LEFT(data::text, 2500) FROM site_specs WHERE site_id=s.id AND aspect='audience' AND is_current) AS audience,
  (SELECT LEFT(data::text, 1500) FROM site_specs WHERE site_id=s.id AND aspect='identity' AND is_current) AS identity,
  (SELECT LEFT(string_agg(pg.line, E'\n' ORDER BY pg.name), 7000)
     FROM (SELECT p.name,
                  p.name || ' [' || COALESCE(p.url,'?') || ', ' || COALESCE(p.page_type,'?') || '] headings: ' || COALESCE(h.heads, '(no headings)') AS line
             FROM pages p
             LEFT JOIN LATERAL (
               SELECT string_agg(m[1], ' | ') AS heads
                 FROM page_components pc, regexp_matches(pc.rendered_html, '<h[12][^>]*>([^<]{3,90})', 'g') m
                WHERE pc.page_id = p.id AND pc.build_status <> 'removed') h ON true
            WHERE p.site_id = s.id AND p.status = 'active') pg) AS pages_with_headings,
  (SELECT count(*) FROM pages p WHERE p.site_id = s.id AND p.status = 'active') AS page_count
FROM sites s WHERE s.id = $1$q$,
                 'params', jsonb_build_array('site_record.site_id'),
                 'output_format', 'object'),
               'next_step', 'run_reader_audit',
               'output_field', 'reader_context'),
             'run_reader_audit', jsonb_build_object(
               'action', 'execute_llm_prompt',
               'config', jsonb_build_object(
                 'prompt', $p$You are a demanding but fair visitor to a website. You are NOT the site's author, NOT its client, and NOT a copy editor. You judge one thing: does each page give a person who came to THIS KIND of site what they came for?

## What kind of site this is (classification spec)
{{.reader_context.classification}}

## Who it is for (audience spec)
{{.reader_context.audience}}

## Its identity
{{.reader_context.identity}}

## The pages, with the headings each actually serves ({{.reader_context.page_count}} pages)
{{.reader_context.pages_with_headings}}

## How to judge
1. Decide, from the classification and audience, what a typical visitor is trying to DO here (find a price, follow a calendar, compare options, learn a method, contact someone, pick a tool). Name it in one line at the top of your reasoning; do not report it as a finding.
2. For each page, read its headings as the page's own promise and ask: would that visitor want this page, and does it serve their purpose or the site's? Pages about the site's own methodology, process, values or team on a site whose visitors want a practical answer are the classic failure (an about page with 14 of 17 headings about "our approach" was the case that created this seat). A call-to-action that sends a visitor to a contact form on a site where nobody needs to make contact is another.
3. Report ONLY pages where a visitor's purpose is not served, up to 8, most load-bearing first. Every finding must quote the page's own heading(s) as evidence. If the site genuinely serves its visitors, return [].
4. Severity: high = the page a visitor most needs is missing or serves the site instead; medium = a page exists but the visitor's purpose is a footnote on it; low = a smaller mismatch.

## Category - choose by REPAIR SHAPE (it decides who would fix it, if anyone ever does)
- gap -> a page or section the visitor needs and the site lacks
- content / structure / differentiation -> the page exists but its substance or composition serves the wrong reader
- tone -> the page serves the right reader in the wrong register
- cta / nav_restructure -> the calls-to-action or navigation send the visitor the wrong way
Use EXACTLY one of those strings. The page field must be an EXACT page name from the list above, or site-wide.

Your findings are RECORDED, not acted on: they become verdict rows a person reads. Write them for that reader - concrete, quoting the page, no flattery and no padding.

Respond with ONLY a JSON array of UP TO 8 findings, each with ALL of these fields:
{"category":"gap|content|structure|differentiation|tone|cta|nav_restructure","severity":"high|medium|low","description":"what the visitor wanted from this page and what the page gives them instead - quote the headings","current_value":"the page's own headings, verbatim from the list above","suggestion":"the smallest change that would serve the visitor (name the page and the section)","acceptance_test":"a concrete check a different agent could run to confirm the visitor is now served","affected_component":"section or component involved, if any","page":"the exact page name, or site-wide","max_fix_attempts":1}

Do not invent pages or headings: every current_value must come from the list above. If the classification and audience above are both empty, return [].$p$,
                 'ai_service', jsonb_build_object('model', 'claude-sonnet-4-6', 'provider', 'anthropic', 'max_tokens', 4000, 'api_key_env_var', 'ANTHROPIC_API_KEY'),
                 'input_fields', jsonb_build_array('reader_context', 'site_record')),
               'next_step', 'set_audit_source',
               'output_field', 'reader_audit'),
             'set_audit_source', jsonb_build_object(
               'action', 'query_database',
               'config', jsonb_build_object('query', 'SELECT ''reader-experience-audit''::text AS audit_source', 'output_format', 'object'),
               'next_step', 'write_findings',
               'output_field', 'audit_source_literal'),
             'write_findings', jsonb_build_object(
               'action', 'write_audit_findings',
               'config', jsonb_build_object(
                 'site_id', 'site_record.site_id',
                 'audit_source', 'audit_source_literal.audit_source',
                 'findings_field', 'reader_audit.result',
                 'filing_mode', 'record'),
               'next_step', 'complete',
               'output_field', 'findings_written'),
             'complete', jsonb_build_object(
               'action', 'complete_workflow',
               'config', jsonb_build_object('output_fields', jsonb_build_array('reader_context', 'reader_audit', 'findings_written')))
           ))),
       true, capabilities, image_repository, image_tag, command, resources, topics, health_config, env_vars, 1,
       delegation_preferences, agent_category, 'experimental', domain_tags, idle_timeout_seconds
  FROM agent_definitions
 WHERE type = 'brief-fidelity-auditor'
   AND is_active AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL
 ORDER BY version DESC LIMIT 1;

-- ── Part C: filing_mode: record on the five existing model seats ─────────────
DO $c$
DECLARE
    r RECORD;
    v_path text[];
BEGIN
    FOR r IN SELECT * FROM (VALUES
        ('visual-design-auditor',   'write_findings'),
        ('content-quality-auditor', 'write_findings'),
        ('site-review-agent',       'write_strategic_findings'),
        ('offer-analyser',          'write_offer_findings'),
        ('brief-fidelity-auditor',  'write_findings')) AS t(agent, step)
    LOOP
        v_path := ARRAY['workflow','steps', r.step, 'action'];
        IF NOT EXISTS (SELECT 1 FROM agent_definitions
                        WHERE type = r.agent AND is_active AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL
                          AND default_config #>> v_path = 'write_audit_findings') THEN
            RAISE EXCEPTION '624 C drift: %.% is not a write_audit_findings step', r.agent, r.step;
        END IF;
        UPDATE agent_definitions
           SET default_config = jsonb_set(default_config, ARRAY['workflow','steps', r.step, 'config', 'filing_mode'], '"record"'::jsonb, true),
               updated_at = now()
         WHERE type = r.agent AND is_active AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;
    END LOOP;
END $c$;

-- ── Part D: improvement-loop rewiring ────────────────────────────────────────
DO $d$
DECLARE
    v_id    uuid;
    v_steps jsonb;
    v_next  text;
    r       RECORD;
    v_cond  text := '';
BEGIN
    SELECT id, default_config #> '{workflow,steps}' INTO v_id, v_steps
      FROM agent_definitions
     WHERE type = 'improvement-loop' AND is_active AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL
     ORDER BY version DESC LIMIT 1;
    IF v_id IS NULL THEN RAISE EXCEPTION '624 D: no live improvement-loop'; END IF;

    -- Accept EITHER the pre-623 edge or 623's bypass; refuse anything else.
    v_next := v_steps #>> '{call_completeness_discovery,next_step}';
    IF v_next NOT IN ('spawn_design_audit', 'record_audit_pass') THEN
        RAISE EXCEPTION '624 D drift: call_completeness_discovery.next_step is % (expected spawn_design_audit or 623''s record_audit_pass)', v_next;
    END IF;
    IF v_steps #>> '{call_brief_fidelity,next_step}' IS DISTINCT FROM 'record_audit_pass' THEN
        RAISE EXCEPTION '624 D drift: call_brief_fidelity.next_step is %', v_steps #>> '{call_brief_fidelity,next_step}';
    END IF;
    IF v_steps ? 'check_seats_ran' OR v_steps ? 'spawn_reader_audit' OR v_steps ? 'spawn_acceptance_discovery' THEN
        RAISE EXCEPTION '624 D drift: a step this file adds already exists';
    END IF;

    -- D1. The acceptance seats after completeness.
    v_steps := jsonb_set(v_steps, '{call_completeness_discovery,next_step}', '"spawn_acceptance_discovery"'::jsonb, true);
    v_steps := jsonb_set(v_steps, '{call_completeness_discovery,description}',
        '"Run completeness checks (empty sections). 624: -> acceptance seats -> model seats (record-only) -> reader seat -> check_seats_ran"'::jsonb, true);
    v_steps := v_steps || jsonb_build_object(
        'spawn_acceptance_discovery', jsonb_build_object(
            'action', 'spawn_agent',
            'config', jsonb_build_object('role', 'acceptance_checker', 'agent_type', 'acceptance-discovery-agent'),
            'next_step', 'call_acceptance_discovery',
            'description', 'Spawn the acceptance council''s mechanical seats (RFC_056: prerequisites, promise, structure)',
            'output_field', 'acceptance_checker_spawned'),
        'call_acceptance_discovery', jsonb_build_object(
            'action', 'call_agent',
            'config', jsonb_build_object(
                'agent_type', 'acceptance-discovery-agent',
                'target_role', 'acceptance_checker',
                'input_mapping', jsonb_build_object('domain', 'site_record.domain', 'site_id', 'site_record.site_id'),
                'timeout_seconds', 300),
            'next_step', 'spawn_design_audit',
            'error_step', 'record_acceptance_discovery_failed',
            'description', 'Run the three mechanical acceptance seats (flag-only verdict rows; own timeout because two of them fetch every served page)',
            'output_field', 'acceptance_result'));

    -- D2. The reader seat after brief fidelity, then the seats-ran gate.
    v_steps := jsonb_set(v_steps, '{call_brief_fidelity,next_step}', '"spawn_reader_audit"'::jsonb, true);
    v_steps := v_steps || jsonb_build_object(
        'spawn_reader_audit', jsonb_build_object(
            'action', 'spawn_agent',
            'config', jsonb_build_object('role', 'reader_auditor', 'agent_type', 'reader-experience-auditor'),
            'next_step', 'call_reader_audit',
            'description', 'Spawn the reader seat (RFC_056): would a person on this kind of site want this page?',
            'output_field', 'reader_auditor_spawned'),
        'call_reader_audit', jsonb_build_object(
            'action', 'call_agent',
            'config', jsonb_build_object(
                'target_role', 'reader_auditor',
                'input_mapping', jsonb_build_object('domain', 'site_record.domain', 'site_id', 'site_record.site_id'),
                'timeout_seconds', 600),
            'next_step', 'check_seats_ran',
            'error_step', 'record_reader_audit_failed',
            'description', 'Reader seat (LLM, filing_mode record): verdict rows only',
            'output_field', 'reader_audit_result'));

    -- D3. One record step per seat call and per enrichment step. Each CONTINUES to
    --     the seat's original successor (one auditor must not strand the sweep) and
    --     leaves a durable, deduplicated row. The seat list, its successor, and the
    --     field check_seats_ran reads, in one place:
    FOR r IN SELECT * FROM (VALUES
        ('news_feed',              'enrich_news_feed',            'enrich_directory_features', 'config'),
        ('directory_features',     'enrich_directory_features',   'load_audit_state',          'config'),
        ('quality_discovery',      'call_quality_discovery',      'spawn_design_discovery',    'step'),
        ('design_discovery',       'call_design_discovery',       'spawn_completeness_discovery','step'),
        ('completeness_discovery', 'call_completeness_discovery', 'spawn_acceptance_discovery','step'),
        ('acceptance_discovery',   'call_acceptance_discovery',   'spawn_design_audit',        'step'),
        ('design_audit',           'call_design_audit',           'spawn_site_review',         'step'),
        ('site_review',            'call_site_review',            'spawn_offer_analyser',      'step'),
        ('offer_analyser',         'call_offer_analyser',         'spawn_brief_fidelity',      'step'),
        ('brief_fidelity',         'call_brief_fidelity',         'spawn_reader_audit',        'step'),
        ('reader_audit',           'call_reader_audit',           'check_seats_ran',           'step')
      ) AS t(seat, call_step, successor, level)
    LOOP
        IF NOT (v_steps ? r.call_step) THEN
            RAISE EXCEPTION '624 D3 drift: step % missing', r.call_step;
        END IF;
        -- the route: step-level everywhere; the config-level twin removed where it existed
        v_steps := jsonb_set(v_steps, ARRAY[r.call_step, 'error_step'], to_jsonb('record_' || r.seat || '_failed'), true);
        IF r.level = 'config' THEN
            v_steps := v_steps #- ARRAY[r.call_step, 'config', 'error_step'];
        END IF;
        v_steps := v_steps || jsonb_build_object('record_' || r.seat || '_failed', jsonb_build_object(
            'action', 'create_work_item',
            'config', jsonb_build_object(
                'source', 'improvement-loop',
                'status', 'deferred',
                'site_id', 'site_record.site_id',
                'summary', format('Audit seat failed: %s did not run to completion on this sweep (620). The sweep continued; the audit pass was NOT stamped', r.seat),
                'priority', 200,
                'severity', 'low',
                'item_type', 'capability_gap',
                'spec_literal', jsonb_build_object(
                    'capability', 'audit_seat_failed',
                    'seat', r.seat,
                    'reason', 'the seat''s call errored or timed out; orchestration_states is reaped within ~25h so this row is the durable record (RFC_056 section 6)'),
                'handler_agent', '',
                'item_pipeline', 'build',
                'item_key_prefix', 'capability_gap_audit_seat_failed_' || r.seat,
                'recurrence_expected', true),
            'next_step', r.successor,
            'error_step', r.successor,
            'description', format('Record that the %s seat failed, then continue to %s. Deduplicated on item_key: one open row per site per seat', r.seat, r.successor),
            'output_field', 'seat_failed_' || r.seat));
        v_cond := v_cond || CASE WHEN v_cond = '' THEN '' ELSE ' OR ' END
                  || 'seat_failed_' || r.seat || '.item_type == capability_gap';
    END LOOP;

    -- D4. The gate. A record step that RAN leaves seat_failed_<seat>.item_type set
    --     (create_work_item returns item_type whether it inserted or deduped); one that
    --     never ran resolves to nil, which compares unequal. No parentheses: OR-joined
    --     equalities need none, and the evaluator would strip them anyway.
    IF position('(' in v_cond) > 0 THEN RAISE EXCEPTION '624 D4: parentheses in the seats condition'; END IF;
    v_steps := v_steps || jsonb_build_object('check_seats_ran', jsonb_build_object(
        'action', 'conditional',
        'config', jsonb_build_object(
            'condition', v_cond,
            'then_step', 'record_audit_attempt',
            'else_step', 'record_audit_pass'),
        'description', 'RFC_056 rule 4: a seat that did not run must not be recorded as having passed. Any seat failure this run -> record_audit_attempt (cooldown stamped, PASS not counted) instead of record_audit_pass. Evaluator semantics proven by TestSeatsGate_* (absent field != literal; one present field trips the OR chain)'),
    'record_audit_attempt', jsonb_build_object(
        'action', 'query_database',
        'config', jsonb_build_object(
            'query', 'UPDATE sites SET settings = jsonb_set(jsonb_set(COALESCE(settings, ''{}''::jsonb), ''{maintenance_profile}'', COALESCE(settings->''maintenance_profile'', ''{}''::jsonb), true), ''{maintenance_profile,last_audit,at}'', to_jsonb(now()), true) WHERE id = $1',
            'params', jsonb_build_array('site_record.site_id'),
            'output_format', 'object'),
        'next_step', 'triage_findings',
        'description', 'A failed-seat sweep stamps the audit ATTEMPT (last_audit.at -> the 14-day cooldown holds, so a persistently failing seat cannot put a site on an every-sweep audit treadmill - improvement_guardian, round d1342f2a) but touches NEITHER the fingerprint nor passes_at_fingerprint: no pass is counted, rule 4 holds, and not_converging still needs three REAL passes',
        'output_field', 'audit_attempt_recorded'));

    UPDATE agent_definitions
       SET default_config = jsonb_set(default_config, '{workflow,steps}', v_steps, false), updated_at = now()
     WHERE id = v_id;
    RAISE NOTICE '624 D: improvement-loop rewired on % (% steps)', v_id, (SELECT count(*) FROM jsonb_object_keys(v_steps));
END $d$;

-- ── Part E: verify ───────────────────────────────────────────────────────────
DO $e$
DECLARE
    v_steps    jsonb;
    v_dangling int;
    v_n        int;
    r          RECORD;
BEGIN
    SELECT default_config #> '{workflow,steps}' INTO v_steps
      FROM agent_definitions WHERE type = 'improvement-loop' AND is_active AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL
     ORDER BY version DESC LIMIT 1;
    SELECT count(*) INTO v_n FROM jsonb_object_keys(v_steps);
    -- 31 + acceptance(2) + reader(2) + check_seats_ran(1) + record_audit_attempt(1) + 11 record steps = 48
    IF v_n <> 48 THEN RAISE EXCEPTION '624 E: expected 48 steps, got %', v_n; END IF;
    SELECT count(*) INTO v_dangling
      FROM (
        SELECT e.v->>'next_step' AS tgt FROM jsonb_each(v_steps) AS e(k,v) WHERE e.v ? 'next_step'
        UNION ALL SELECT e.v->>'error_step' FROM jsonb_each(v_steps) AS e(k,v) WHERE e.v ? 'error_step'
        UNION ALL SELECT e.v->'config'->>'error_step' FROM jsonb_each(v_steps) AS e(k,v) WHERE e.v->'config' ? 'error_step'
        UNION ALL SELECT e.v->'config'->>'then_step' FROM jsonb_each(v_steps) AS e(k,v) WHERE e.v->'config' ? 'then_step'
        UNION ALL SELECT e.v->'config'->>'else_step' FROM jsonb_each(v_steps) AS e(k,v) WHERE e.v->'config' ? 'else_step'
      ) AS edges WHERE tgt IS NOT NULL AND NOT (v_steps ? tgt);
    IF v_dangling > 0 THEN RAISE EXCEPTION '624 E: % dangling edge(s)', v_dangling; END IF;
    IF v_steps #> '{enrich_news_feed,config}' ? 'error_step' OR v_steps #> '{enrich_directory_features,config}' ? 'error_step' THEN
        RAISE EXCEPTION '624 E: a config-level error_step twin survived on an enrichment step';
    END IF;
    IF v_steps #>> '{check_seats_ran,config,then_step}' <> 'record_audit_attempt' OR v_steps #>> '{check_seats_ran,config,else_step}' <> 'record_audit_pass' THEN
        RAISE EXCEPTION '624 E: check_seats_ran wiring wrong';
    END IF;
    FOR r IN SELECT * FROM (VALUES
        ('visual-design-auditor','write_findings'),('content-quality-auditor','write_findings'),
        ('site-review-agent','write_strategic_findings'),('offer-analyser','write_offer_findings'),
        ('brief-fidelity-auditor','write_findings'),('reader-experience-auditor','write_findings')) AS t(agent, step)
    LOOP
        IF NOT EXISTS (SELECT 1 FROM agent_definitions WHERE type = r.agent AND is_active AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL
                         AND default_config #>> ARRAY['workflow','steps', r.step, 'config', 'filing_mode'] = 'record') THEN
            RAISE EXCEPTION '624 E: %.% does not carry filing_mode=record', r.agent, r.step;
        END IF;
    END LOOP;
    IF NOT EXISTS (SELECT 1 FROM agent_definitions WHERE type='acceptance-discovery-agent' AND is_active AND deleted_at IS NULL) THEN
        RAISE EXCEPTION '624 E: acceptance-discovery-agent missing';
    END IF;
    RAISE NOTICE '624: verified - 48 steps, 0 dangling edges, 6 model seat write steps record-only, 2 agents created, 11 seat-failure records + the audit-attempt stamp wired';
END $e$;

-- The travelling record (tooling_provenance, round d1342f2a): the decision reaches
-- doc_notes under the two subject keys the next session will actually query.
INSERT INTO doc_notes (subject_type, subject_key, body, categories, source, created_by)
VALUES
  ('action', 'write_audit_findings',
   'DECISION (RFC_056, migration 624): filing_mode=record is LIVE on six model-seat write steps (visual-design-auditor, content-quality-auditor, site-review-agent, offer-analyser, brief-fidelity-auditor, reader-experience-auditor). Their findings are VERDICT rows: deferred + handler '''' + routing/provenance/release_recipe in spec; both promoters refuse them; the seat''s own silence-retraction is the revalidator (RFC_056 addendum). Release = the recipe on the row. Do NOT re-point these at dispatch without reading RFC_056.',
   '["decision"]'::jsonb, 'migration-624', 'loanzy_uk_example_site'),
  ('decision', 'improvement-loop',
   'DECISION (RFC_056, migration 624): the loop now runs mechanical seats -> acceptance seats (build_prerequisites, heading_promise, structure_floor; flag-only) -> four model-seat calls (record-only) -> reader seat -> check_seats_ran. Every seat call and both enrichment steps route failure to a record_<seat>_failed step (deferred capability_gap per site per seat) and a failed-seat sweep stamps the audit ATTEMPT but never a PASS. Prior art built on, not re-derived: IMP-054 (premise voided by detected-item-promoter), IMP-006 (approval-mode proposal, shipped as filing_mode), IMP-016 (gated re-enable).',
   '["decision"]'::jsonb, 'migration-624', 'loanzy_uk_example_site');

COMMIT;
