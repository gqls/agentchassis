-- 734_meta_description_repair_agent.sql
--
-- Seeds `meta-description-repair`: the ACTOR that a copy-gate refusal of a meta
-- description is now filed at (bugs_open/442, owner ruling 2026-09-03 "make them
-- loud").
--
-- ══ WHY AN ACTOR AND NOT A REVIEW QUEUE — the number that decided it ═════════
-- The obvious "make it loud" is a `needs_human_review` row with no handler. That
-- is how the estate's other flag-only findings are filed, and it is not loud.
-- [MEASURED 2026-09-03, site_work_items UNION site_work_items_archive]:
--
--     items WITH a handler_agent    56,315    83% complete
--     items with NO handler          6,699    17% complete,  989 parked
--
-- and `voice_tells` — the queue this refusal would naturally have joined — is 69
-- rows, EVERY ONE handler-less: 3 complete, 66 parked, nothing filed since
-- 2026-08-27. The graveyard is not a fact about busy humans; it is what filing
-- without an actor looks like.
--
-- ══ WHAT IT DOES ════════════════════════════════════════════════════════════
-- Re-asks for the sentence with the refusal QUOTED BACK, then saves through
-- `save_page_meta_description` so the SAME gates judge the retry. It cannot
-- smuggle bad copy past them: the gate is inside the action, not in this
-- workflow (the 2026-08-02 §2 reason — a gate a workflow author can forget to
-- wire is a comment).
--
--   ensure_site_record -> load_page -> load_voice_rules -> rewrite_description
--   -> save_description -> check_saved -> {mark_repaired | mark_needs_human}
--
-- ⚠ ON A SECOND REFUSAL IT PARKS AT `needs_human_review`, AND THAT IS DELIBERATE
-- despite the 17% above. The alternative is `fail_work_item`'s ordinary attempt
-- ladder, which after two strikes brands the row `unresolved` — terminal, and
-- silent again, which is the defect this whole change exists to end. Parking is
-- the terminal state AFTER a genuine automated attempt, not INSTEAD of one, and
-- the row then carries the original candidate, the rule that refused it, and the
-- rewrite that also failed. That is a far better thing to hand a person than
-- "this page has no description".
--
-- ══ ORDERING, AND WHY THIS CONFIG LEADS THE CODE ════════════════════════════
-- The Go half (save_page_meta_description_refusal_item.go) files at
-- `meta-description-repair` and is INERT until the chassis rolls. This seed must
-- be live FIRST: a refusal filed at a handler that does not exist is demoted to
-- `deferred` by writeWorkItem's registration probe — safe, never a dispatcher
-- livelock (bugs_open/078), but a parked row nobody asked for.
-- Every action named below is registered TODAY (registry.go: ensure_site_record,
-- query_database, execute_llm_prompt, save_page_meta_description,
-- conditional_branch, complete_work_item, fail_work_item), so this seed cannot
-- fail at runtime for naming an unregistered action — the hazard CLAUDE.md warns
-- about when it says "image first, then seeds".
--
-- ⚠ NO VERIFIER IS REGISTERED for `meta_description_refused`, and that is stated
-- rather than omitted. `complete_work_item` consults the item type's verifier
-- when one exists and completes without it when none does — the same position as
-- `content_rewrite` and most types (23 RegisterVerifier calls fleet-wide).
-- Registering one is NOT a one-line change: it fails five build guards and needs
-- a live migration amending the claimed-item-timeout sweep's pre_query, merged
-- with any other lane's pending amendment
-- (discovery_checks/verify_required_fields_missing.go documents the exact
-- sequence). Recorded as a named follow-up in bugs_open/442, not smuggled in here.
--
-- Reversible: 734_..._ROLLBACK.sql soft-deletes the row.

BEGIN;

DO $$
DECLARE n int;
BEGIN
    SELECT count(*) INTO n FROM agent_definitions
     WHERE type = 'meta-description-repair' AND deleted_at IS NULL;
    IF n <> 0 THEN
        RAISE EXCEPTION
            'ABORT: % meta-description-repair row(s) already exist — this migration has '
            'ALREADY applied, or another session has seeded the agent. Re-read before re-running.', n;
    END IF;

    -- The agent this one repairs FOR must still exist and still gate its writes,
    -- or the repair path is pointing at a mechanism that has moved.
    SELECT count(*) INTO n FROM agent_definitions
     WHERE type = 'meta-description-backfiller' AND is_active
       AND COALESCE(is_snapshot,false) = false AND deleted_at IS NULL;
    IF n <> 1 THEN
        RAISE EXCEPTION 'ABORT: expected exactly 1 live meta-description-backfiller, found %', n;
    END IF;
END $$;

INSERT INTO agent_definitions (type, display_name, description, category, status, is_active, default_config)
SELECT
  'meta-description-repair',
  'Meta Description Repair',
  'Repairs a meta description the site copy gates refused. Dispatched from a meta_description_refused work item filed by save_page_meta_description (bugs_open/442). Re-asks for the sentence with the refusal reason quoted back, saves through the same gated action, and parks at needs_human_review only if the rewrite is refused too.',
  'content',
  'experimental',
  true,
  jsonb_build_object('workflow', jsonb_build_object(
    'start_step', 'ensure_site_record',
    'processing_mode', 'sequential',
    'timeout_seconds', 600,
    'steps', jsonb_build_object(

      'ensure_site_record', jsonb_build_object(
        'action', 'ensure_site_record',
        'next_step', 'load_page',
        'description', 'Resolve site_id/domain into site_record',
        'output_field', 'site_record'),

      'load_page', jsonb_build_object(
        'action', 'query_database',
        'config', jsonb_build_object(
          'query', 'SELECT p.id, p.name, p.url, p.title, p.page_type, LEFT(string_agg(regexp_replace(regexp_replace(regexp_replace(regexp_replace(pc.rendered_html, ''<(style|script)[^>]*>.*?</\1>'', '' '', ''gis''), ''<[^>]+>'', '' '', ''g''), ''&nbsp;|&amp;|&quot;|&#39;|&lt;|&gt;'', '' '', ''g''), ''\s+'', '' '', ''g''), '' '' ORDER BY pc.slot_name), 1200) AS content_sample FROM pages p JOIN page_components pc ON pc.page_id = p.id AND pc.rendered_html IS NOT NULL AND COALESCE(pc.slot_name, '''') NOT IN (''header'',''footer'',''head'') WHERE p.id = $1 GROUP BY p.id, p.name, p.url, p.title, p.page_type',
          'params', jsonb_build_array('input_data.spec.page_id'),
          'output_format', 'object'),
        'next_step', 'load_voice_rules',
        'description', 'The one page this item is about. Chrome slots excluded so the page is not described by its header — the same exclusion the backfiller''s selection query makes.',
        'output_field', 'page'),

      'load_voice_rules', jsonb_build_object(
        'action', 'query_database',
        'config', jsonb_build_object(
          'query', 'SELECT ph->>''pattern'' AS pattern, ph->>''reason'' AS reason FROM site_specs ss CROSS JOIN LATERAL jsonb_array_elements(COALESCE(ss.data#>''{voice_gate,banned_phrases}'', ''[]''::jsonb)) ph WHERE ss.site_id = $1 AND ss.aspect = ''voice'' AND ss.is_current = true AND COALESCE((ss.data#>>''{voice_gate,enabled}'')::boolean, false) = true',
          'params', jsonb_build_array('site_record.site_id'),
          'output_format', 'object'),
        'next_step', 'rewrite_description',
        'description', 'The site''s own banned phrases, told to the writer BEFORE it writes. Verbatim from the backfiller so the two cannot drift apart in what they show the model.',
        'output_field', 'voice_rules'),

      'rewrite_description', jsonb_build_object(
        'action', 'execute_llm_prompt',
        'config', jsonb_build_object(
          'ai_service', jsonb_build_object(
            'model', 'claude-haiku-4-5', 'provider', 'anthropic',
            'max_tokens', 1000, 'api_key_env_var', 'ANTHROPIC_API_KEY'),
          'input_fields', jsonb_build_array('page', 'site_record', 'voice_rules', 'input_data'),
          'output_format', 'json',
          'prompt_template', E'You are rewriting ONE meta description for a page on {{.site_record.domain}}.\n\nA meta description is the sentence a search engine prints under the page title in its results. It is read by far more people than the page itself, so it is a promise to a visitor, not a summary for a machine.\n\n## THIS PAGE ALREADY HAD A DESCRIPTION REFUSED\nThe site''s copy rules rejected the previous attempt, so the page is currently BLANK. Here is exactly what was refused and why. Do not repeat the mistake, and do not work around it with a synonym that makes the same claim.\n\nRefused sentence: {{.input_data.spec.refused_candidate}}\nRule that refused it: {{.input_data.spec.reason}}\nDetail: {{.input_data.spec.detail}}\n\n## The page\n{{range .page.rows}}name: {{.name}}\nurl: {{.url}}\ntitle: {{.title}}\ntype: {{.page_type}}\ncontent: {{.content_sample}}\n{{end}}\n{{if .voice_rules.count}}## THIS SITE BANS THESE PHRASES — a description containing one is REFUSED and the page stays blank\nEach line is a regular expression the site checks, with the reason it exists.\n{{range .voice_rules.rows}}- {{.pattern}}  ({{.reason}})\n{{end}}\n{{end}}## House style (the rules that govern a single sentence)\n- One idea, one sentence. Do not chain clauses with commas or semicolons to fit more in.\n- No em dashes. Not one.\n- Never open with what something is NOT. "Not just a calculator, but a..." and "More than a guide" are the same move: a manufactured twist. Start with the fact, the way a person saying it out loud would.\n- Match word-weight to the claim, in BOTH directions. "Powerful", "seamless", "comprehensive", "revolutionary" overclaim upward; "simple", "just", "nothing fancy" overclaim downward, which still asks the reader to be impressed.\n- Cut self-flagging commentary and hedges: crucially, genuinely, exactly, deliberately, what matters here is, at its core, in essence, delve, leverage, robust, seamless, furthermore, moreover. Do not tell the reader something is important. State it.\n- Contractions are fine and usually better: it is -> it''s, does not -> doesn''t.\n- Do not name the site or brand. The search engine already prints it above this line.\n\n## Rules\n- 110-150 characters AND AT MOST 20 words. Both limits, not either.\n- One sentence, plain English, British spelling.\n- Say what the visitor GETS from this page. Lead with that.\n- No build, generator or specification wording. Never describe how the page was made.\n- Ground it in the content given. If the content genuinely cannot support a specific description, return an empty string rather than inventing one — a blank is better than a false promise, and the page will be looked at by a person.\n\nReturn ONLY valid JSON:\n{"meta_description": "<the sentence, or an empty string>"}\n'),
        'next_step', 'save_description',
        'description', 'Rewrite with the refusal quoted back. The reason the first attempt failed is the one piece of information the hourly backfiller never had.',
        'output_field', 'rewritten'),

      'save_description', jsonb_build_object(
        'action', 'save_page_meta_description',
        'config', jsonb_build_object(
          'page_id_field', 'input_data.spec.page_id',
          'description_field', 'rewritten.result.meta_description'),
        'next_step', 'check_saved',
        'description', 'The SAME gated action the backfiller uses, so the same voice gate and banned-claims sweep judge the rewrite. A refusal here is the second one and routes to a human.',
        'output_field', 'save_result'),

      'check_saved', jsonb_build_object(
        'action', 'conditional_branch',
        'config', jsonb_build_object(
          'condition_field', 'save_result.updated',
          'conditions', jsonb_build_object('true', 'mark_repaired', 'false', 'mark_needs_human'),
          'default', 'mark_needs_human'),
        'description', 'save_result.updated is the action''s own honest boolean. Anything that is not an unambiguous true — including a missing field — routes to a human, which is the fail-safe direction.',
        'output_field', 'branch'),

      'mark_repaired', jsonb_build_object(
        'action', 'complete_work_item',
        'config', jsonb_build_object('work_item_id', 'input_data.work_item_id'),
        'next_step', 'complete',
        'description', 'The page now has a description that passed the same gates that refused the first one.',
        'output_field', 'complete_result'),

      'mark_needs_human', jsonb_build_object(
        'action', 'fail_work_item',
        'config', jsonb_build_object(
          'work_item_id', 'input_data.work_item_id',
          'status_override', 'needs_human_review',
          'error_message', 'The rewrite was refused by the copy gates too, so this page still has no meta description and a person has to judge the copy. The item carries the original refused sentence, the rule that refused it, and this run''s save_result. Do NOT raise the site''s voice thresholds to make it pass: they are per-site and would stop gating that site''s PAGES as well (bugs_open/338).'),
        'next_step', 'complete',
        'description', 'Parked only AFTER a genuine automated attempt. fail_work_item''s ordinary ladder would brand the row unresolved after two strikes, which is silent again — the defect this whole change exists to end (bugs_open/442).',
        'output_field', 'fail_result'),

      'complete', jsonb_build_object(
        'action', 'complete_workflow',
        'config', jsonb_build_object(
          'result_message', 'Meta description repair finished. Read save_result: updated true means the rewrite passed the same gates that refused the original and the work item is complete; false means it was refused again and the item is parked at needs_human_review with both attempts on it.'))
    )))
WHERE NOT EXISTS (SELECT 1 FROM agent_definitions WHERE type='meta-description-repair' AND deleted_at IS NULL);

DO $$
DECLARE
    steps      jsonb;
    missing    text[] := '{}';
    s          text;
    unreg      int;
BEGIN
    SELECT default_config->'workflow'->'steps' INTO steps
      FROM agent_definitions
     WHERE type='meta-description-repair' AND is_active AND deleted_at IS NULL;

    IF steps IS NULL THEN
        RAISE EXCEPTION 'ABORT: meta-description-repair was not inserted';
    END IF;

    FOREACH s IN ARRAY ARRAY['ensure_site_record','load_page','load_voice_rules',
                             'rewrite_description','save_description','check_saved',
                             'mark_repaired','mark_needs_human','complete']
    LOOP
        IF NOT (steps ? s) THEN missing := missing || s; END IF;
    END LOOP;
    IF array_length(missing,1) > 0 THEN
        RAISE EXCEPTION 'ABORT: workflow is missing step(s) %', missing;
    END IF;

    -- The load-bearing assertion: the retry must go through the GATED action.
    -- A workflow that wrote the column directly would repair the page and defeat
    -- the gates in the same move, and every "the page has a description" check
    -- would go on passing — the exact shape bugs_open/442 is about.
    IF steps->'save_description'->>'action' <> 'save_page_meta_description' THEN
        RAISE EXCEPTION 'ABORT: the save step does not use the gated action (found %)',
            steps->'save_description'->>'action';
    END IF;

    -- Both branches must exist and must differ, or the conditional is decorative.
    IF steps->'check_saved'->'config'->'conditions'->>'true'
       = steps->'check_saved'->'config'->'conditions'->>'false' THEN
        RAISE EXCEPTION 'ABORT: both branches of check_saved route to the same step';
    END IF;

    RAISE NOTICE '734: meta-description-repair seeded — 9 steps, retry goes through the gated action, both branches distinct.';
END $$;

COMMIT;
