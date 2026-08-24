-- 602 — when nobody names `related_pages`, the tool workflows ASK for them.
--       CONFIG ONLY — live on apply. Owner ruling 2026-08-24.
--
--       ⚠ _HOLD. The ordering condition IS a roll this time. Read "ORDERING"
--       below: this migration wires a key that only the new binary reads, and
--       applying it early costs a model call per tool build that changes
--       nothing.
--
-- ============================================================================
-- WHAT THE PROBLEM IS, in plain terms
-- ============================================================================
-- When a tool is built for a site, each page named in the request's
-- `related_pages` gets one CROSS-MENTION: a single sentence woven into that
-- page's existing prose, mentioning the tool and linking to it. Live example on
-- dartsonline.com `barrel-shapes`: "…the tungsten percentage vs barrel diameter
-- visualiser lets you compare percentages against weight and watch the barrel
-- narrow on screen…". Nothing else on the estate produces that.
--
-- `related_pages` has exactly ONE producer. [MEASURED 2026-08-24, add_tool items
-- created since 08-17] `tool-suggester` 11 of 11 carry the key; `owner-request`
-- 0 of 58, `automated_check` 0 of 7, `operator` 0 of 1. It is not a compliance
-- rate — it is a split: the suggester's prompt asks for the field, and every
-- other route writes the spec by hand from a five-key template that never had
-- it (`complexity, description, function, name, priority`).
--
-- Until migration 516 (live 2026-08-21 16:55Z) the omission was masked: the
-- resolver's whole-tree search substituted ANOTHER suggestion's list, which is
-- how nine tools on webdesign.co.uk were all cross-linked to the same two
-- articles (`bugs_open/330`). 516 removed the substitution, correctly. The
-- consequence is that the omission now means NO cross-mentions at all, recorded
-- only as an `info` row while the build succeeds and the page deploys:
-- [MEASURED 2026-08-24] 13 of 13 tool births 08-22→08-24 emitted zero, and
-- `tool-generator` has created no `tool_crosslink:%` item since 08-21.
--
-- ============================================================================
-- WHAT THIS DOES, and why it is not 330 in a new hat
-- ============================================================================
-- Two new steps in each tool workflow, ahead of the step that saves the tool:
--
--   load_site_page_names   action `load_site_pages`  → the site's page names
--   suggest_related_pages  action `execute_llm_prompt` → 1-3 of them, or []
--
-- and one new wire on the saving step:
--
--   "related_pages_fallback?": "suggest_related_pages.result"
--
-- **The requester still wins.** `relatedPagesFromInputs`
-- (`create_tool_cross_link_items.go`) consults `related_pages` first, then
-- `input_data.spec.related_pages`, and only then the fallback. A picker that can
-- overrule an explicit choice makes the field the requester filled in silently
-- advisory, which is worse than no picker at all. Pinned by
-- `TestRelatedPagesPrecedenceAndSource`, mutation-proved by reversing the order.
--
-- **Why this is the opposite of 330's substitution, not a repeat of it.** 330's
-- defect was a SEARCH: an unresolved field fell through to a whole-tree scan
-- that returned a value belonging to a different tool, silently, with no record.
-- This is a DECLARED step with its own name, its own output field, its own
-- prompt and its own row in every census — and it is asked the question about
-- THIS tool. Both keys keep the `?` OPTIONAL-EXPLICIT marker, so neither can
-- ever reach the search again.
--
-- **Absence survives.** The prompt's rule 5 says an empty answer is a correct
-- answer, because a site may genuinely have no page a tool belongs on. Nothing
-- here forces a link to exist.
--
-- **A wrong name cannot become a wrong link.** The emitter resolves every name
-- against `pages` for that site and skips what does not match
-- (`create_tool_cross_link_items.go`, the `pageMap` lookup), and refuses any
-- name beginning with `tool-`. A hallucinated page name therefore produces
-- nothing, not a bad cross-mention.
--
-- ============================================================================
-- COST, stated rather than assumed
-- ============================================================================
-- One extra model call per tool build, INCLUDING the ~14% of builds whose spec
-- already names pages (11 of 77 since 08-17), whose answer is then discarded by
-- the precedence above. That waste is deliberate and it is the cheaper mistake:
-- the alternative is a `conditional` step gating the picker on the absence of a
-- key, and `bugs_open/313` is what that costs when it goes wrong — an
-- unresolvable condition silently evaluated false and skipped an agent's only
-- LLM step on every run for FOUR MONTHS, with the run reporting success
-- throughout. A wasted call is visible in `llm_call_log`; a skipped one is not.
-- Volume makes this affordable: [MEASURED 2026-08-24] tool births run 1-14/day,
-- 4 on the day this was written.
--
-- ============================================================================
-- ORDERING — why this is a _HOLD
-- ============================================================================
-- The reader shipped first and is INERT until this file applies; this file is
-- INERT-BUT-WASTEFUL until the reader is live. `related_pages_fallback` is
-- declared on both actions' input specs in commit `0fb94a7dd`. On a binary
-- without it the extractor DROPS the key before the resolver sees it
-- (`bugs_closed/336`'s shape), so the picker would run, cost a call, and change
-- nothing — invisibly.
--
-- APPLY ONLY WHEN the live chassis carries the reader. Probe the artefact, not
-- git, and run BOTH controls in the same breath:
--   POD=$(kubectl -n ai-persona-system get pods -l app=agent-chassis \
--          -o jsonpath='{.items[0].metadata.name}')
--   kubectl -n ai-persona-system exec $POD -- grep -aq "related_pages_fallback" /proc/1/exe   # must be 0
--   kubectl -n ai-persona-system exec $POD -- grep -aq "tool_page_will_not_go_live" /proc/1/exe # control +, must be 0
--   kubectl -n ai-persona-system exec $POD -- grep -aq "zzz_cannot_exist" /proc/1/exe           # control -, must be 1
-- ⚠ `-l app=agent-chassis` matches 2 pods of the ~68 that run this binary as
-- per-run `agent-*` pods, and `tool-generator` is one of THOSE (it spawns per
-- run; no such pod exists at rest). Probe a per-run pod too, on a different
-- node, before believing the roll is complete.
--
-- The runner refuses `--record-only` on a _HOLD sidecar; record the apply in the
-- `staged_component_build` lane NOTES instead.
--
-- ============================================================================
-- HOW TO VERIFY THE FIX (not the apply — the apply is verified in-transaction)
-- ============================================================================
-- The demand control is a hand-filed `add_tool` WITHOUT the key from a real
-- producer; the `webdesign-tool-rebuilds` lane has offered to supply one. Then:
--
--   SELECT context->>'related_pages_source', context->>'related_pages_n',
--          count(*)
--     FROM agent_error_log
--    WHERE error_code LIKE 'tool_crosslink_not_emitted:%'
--      AND occurred_at > '<apply time>' GROUP BY 1,2;
--   SELECT spec->>'related_pages_source', count(*) FROM site_work_items
--    WHERE item_key LIKE 'tool_crosslink:%' AND created_at > '<apply time>'
--    GROUP BY 1;
--
-- PASS = at least one row of either kind carrying `suggested`. Reading only
-- "cross-mentions resumed" cannot distinguish a working picker from a week in
-- which requesters happened to fill the field in — which is why the reader
-- stamps the source (commit `0fb94a7dd`).
--
-- ROLLBACK: 602_tool_workflows_ask_for_related_pages_HOLD_ROLLBACK.sql
-- ============================================================================

BEGIN;

SELECT snapshot_agent('tool-generator',
                      '602_tool_workflows_ask_for_related_pages: pre-update');
SELECT snapshot_agent('tool-deployer',
                      '602_tool_workflows_ask_for_related_pages: pre-update');

-- ---------------------------------------------------------------------------
-- Guard: refuse if the anchors have moved or the wire already exists.
-- An UPDATE that silently no-ops because someone re-ordered the workflow is the
-- failure this estate keeps paying for; abort instead.
-- ---------------------------------------------------------------------------
DO $$
DECLARE
    tg jsonb;
    td jsonb;
BEGIN
    SELECT default_config->'workflow'->'steps' INTO tg FROM agent_definitions
     WHERE type='tool-generator' AND is_active
       AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;
    SELECT default_config->'workflow'->'steps' INTO td FROM agent_definitions
     WHERE type='tool-deployer' AND is_active
       AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;

    IF tg IS NULL OR td IS NULL THEN
        RAISE EXCEPTION '602: one of the tool agents has no live workflow — refusing to guess';
    END IF;

    IF tg->'generate_tool_html'->>'next_step' IS DISTINCT FROM 'save_tool' THEN
        RAISE EXCEPTION '602: tool-generator generate_tool_html.next_step is %, expected save_tool — the workflow moved, re-read it before inserting steps',
            COALESCE(tg->'generate_tool_html'->>'next_step','<absent>');
    END IF;
    IF td->'ensure_site_record'->>'next_step' IS DISTINCT FROM 'deploy_tool' THEN
        RAISE EXCEPTION '602: tool-deployer ensure_site_record.next_step is %, expected deploy_tool',
            COALESCE(td->'ensure_site_record'->>'next_step','<absent>');
    END IF;

    IF tg ? 'suggest_related_pages' OR td ? 'suggest_related_pages' THEN
        RAISE EXCEPTION '602: a suggest_related_pages step already exists — this migration has run, or another wrote one';
    END IF;

    IF (tg->'save_tool'->'config') ? 'related_pages_fallback?'
       OR (td->'deploy_tool'->'config') ? 'related_pages_fallback?' THEN
        RAISE EXCEPTION '602: the fallback wire already exists — nothing to do';
    END IF;

    -- The wire this file depends on must still be the marked one. If 516's `?`
    -- has been unmarked, the precedence this migration assumes is gone and the
    -- picker could be reached by the whole-tree search instead.
    IF NOT ((tg->'save_tool'->'config') ? 'related_pages?') THEN
        RAISE EXCEPTION '602: tool-generator save_tool no longer carries the OPTIONAL-EXPLICIT related_pages? wire (migration 516). Re-read bugs_open/330 before proceeding';
    END IF;
    IF NOT ((td->'deploy_tool'->'config') ? 'related_pages?') THEN
        RAISE EXCEPTION '602: tool-deployer deploy_tool no longer carries the OPTIONAL-EXPLICIT related_pages? wire (migration 516)';
    END IF;
END $$;

-- ---------------------------------------------------------------------------
-- tool-generator: generate_tool_html → load_site_page_names → suggest_related_pages → save_tool
-- ---------------------------------------------------------------------------
UPDATE agent_definitions SET default_config = jsonb_set(
    jsonb_set(
        jsonb_set(
            jsonb_set(default_config,
                '{workflow,steps,generate_tool_html,next_step}', '"load_site_page_names"'::jsonb),
            '{workflow,steps,load_site_page_names}', jsonb_build_object(
                'action', 'load_site_pages',
                'config', jsonb_build_object(
                    'site_id', 'site_record.site_id',
                    -- Never fail a tool build for a cross-mention: skip straight
                    -- to the save on any error. Absence then propagates and the
                    -- behaviour is exactly what it was before this migration.
                    'error_step', 'save_tool'),
                'next_step', 'suggest_related_pages',
                'description', 'Load the site page names the picker chooses from (602).',
                'output_field', 'site_pages')),
        '{workflow,steps,suggest_related_pages}', jsonb_build_object(
            'action', 'execute_llm_prompt',
            'config', jsonb_build_object(
                'ai_service', jsonb_build_object(
                    'model', 'claude-sonnet-5',
                    'provider', 'anthropic',
                    'max_tokens', 300,
                    'api_key_env_var', 'ANTHROPIC_API_KEY'),
                'error_step', 'save_tool',
                'input_fields', jsonb_build_array('input_data', 'site_pages'),
                'output_format', 'text',
                'prompt_template',
'You are choosing which of a website''s existing pages should carry a one-sentence mention of a new tool.

## The tool
Name: {{.input_data.spec.name}}
Function: {{.input_data.spec.function}}
Description: {{.input_data.spec.description}}

## The site''s existing pages
{{range .site_pages.page_names}}- {{.}}
{{end}}

## Rules
1. Choose between 1 and 3 page names from the list above, copied EXACTLY as written.
2. Never choose a name beginning with "tool-". A tool page must not mention another tool.
3. Never choose "index".
4. Choose by topic: the page should be about something the tool genuinely helps a reader do or understand. A reader of that page should find the tool useful at that moment, not merely be on the same website.
5. If no page is a genuine topical match, output an empty array. An empty answer is a CORRECT answer and is preferred to a weak match: a mention on an unrelated page reads as an advertisement and is worse than no mention at all.
6. Output ONLY a JSON array of strings. No prose, no explanation, no markdown fences.

Example of a good answer: ["barrel-weight","tungsten-guide"]
Example of a correct empty answer: []'),
            'next_step', 'save_tool',
            'description', 'Pick 1-3 related pages when the request named none (602, owner ruling 2026-08-24). Its answer is a FALLBACK: relatedPagesFromInputs consults the request first, always.',
            'output_field', 'suggest_related_pages')),
    '{workflow,steps,save_tool,config,related_pages_fallback?}',
    '"suggest_related_pages.result"'::jsonb)
WHERE type='tool-generator' AND is_active
  AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;

-- ---------------------------------------------------------------------------
-- tool-deployer: ensure_site_record → load_site_page_names → suggest_related_pages → deploy_tool
--
-- Included for the reason 516 stated when it included this same pair: the
-- deployer carries the identical wire to the identical helper, so the identical
-- gap is available to it the moment its spec omits the key. The admin door
-- (`tool_admin_handlers.go`) builds a THREE-key spec, so it omits it always.
-- ---------------------------------------------------------------------------
UPDATE agent_definitions SET default_config = jsonb_set(
    jsonb_set(
        jsonb_set(
            jsonb_set(default_config,
                '{workflow,steps,ensure_site_record,next_step}', '"load_site_page_names"'::jsonb),
            '{workflow,steps,load_site_page_names}', jsonb_build_object(
                'action', 'load_site_pages',
                'config', jsonb_build_object(
                    'site_id', 'site_record.site_id',
                    'error_step', 'deploy_tool'),
                'next_step', 'suggest_related_pages',
                'description', 'Load the site page names the picker chooses from (602).',
                'output_field', 'site_pages')),
        '{workflow,steps,suggest_related_pages}', jsonb_build_object(
            'action', 'execute_llm_prompt',
            'config', jsonb_build_object(
                'ai_service', jsonb_build_object(
                    'model', 'claude-sonnet-5',
                    'provider', 'anthropic',
                    'max_tokens', 300,
                    'api_key_env_var', 'ANTHROPIC_API_KEY'),
                'error_step', 'deploy_tool',
                'input_fields', jsonb_build_array('input_data', 'site_pages'),
                'output_format', 'text',
                'prompt_template',
'You are choosing which of a website''s existing pages should carry a one-sentence mention of a tool being added to the site.

## The tool
Name: {{.input_data.spec.name}}
Function: {{.input_data.spec.function}}
Description: {{.input_data.spec.description}}

## The site''s existing pages
{{range .site_pages.page_names}}- {{.}}
{{end}}

## Rules
1. Choose between 1 and 3 page names from the list above, copied EXACTLY as written.
2. Never choose a name beginning with "tool-". A tool page must not mention another tool.
3. Never choose "index".
4. Choose by topic: the page should be about something the tool genuinely helps a reader do or understand. A reader of that page should find the tool useful at that moment, not merely be on the same website.
5. If no page is a genuine topical match, output an empty array. An empty answer is a CORRECT answer and is preferred to a weak match: a mention on an unrelated page reads as an advertisement and is worse than no mention at all.
6. Output ONLY a JSON array of strings. No prose, no explanation, no markdown fences.

Example of a good answer: ["barrel-weight","tungsten-guide"]
Example of a correct empty answer: []'),
            'next_step', 'deploy_tool',
            'description', 'Pick 1-3 related pages when the request named none (602). FALLBACK only — the request wins.',
            'output_field', 'suggest_related_pages')),
    '{workflow,steps,deploy_tool,config,related_pages_fallback?}',
    '"suggest_related_pages.result"'::jsonb)
WHERE type='tool-deployer' AND is_active
  AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;

-- ---------------------------------------------------------------------------
-- VERIFY, in-transaction and able to ABORT.
-- A block of SELECTs cannot stop the COMMIT — ON_ERROR_STOP ignores a non-empty
-- result set — so every assertion here RAISEs (RFC_006's lesson, LANDMINES).
-- ---------------------------------------------------------------------------
DO $$
DECLARE
    r record;
    n int;
BEGIN
    FOR r IN
        SELECT type,
               default_config->'workflow'->'steps' AS steps,
               CASE type WHEN 'tool-generator' THEN 'save_tool' ELSE 'deploy_tool' END AS saver,
               CASE type WHEN 'tool-generator' THEN 'generate_tool_html' ELSE 'ensure_site_record' END AS anchor
          FROM agent_definitions
         WHERE type IN ('tool-generator','tool-deployer') AND is_active
           AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL
    LOOP
        IF r.steps->r.anchor->>'next_step' <> 'load_site_page_names' THEN
            RAISE EXCEPTION '602 VERIFY: %.%.next_step is %, expected load_site_page_names',
                r.type, r.anchor, COALESCE(r.steps->r.anchor->>'next_step','<absent>');
        END IF;
        IF r.steps->'load_site_page_names'->>'action' <> 'load_site_pages' THEN
            RAISE EXCEPTION '602 VERIFY: %.load_site_page_names runs %, expected load_site_pages',
                r.type, COALESCE(r.steps->'load_site_page_names'->>'action','<absent>');
        END IF;
        IF r.steps->'load_site_page_names'->>'next_step' <> 'suggest_related_pages' THEN
            RAISE EXCEPTION '602 VERIFY: %.load_site_page_names does not lead to the picker', r.type;
        END IF;
        IF r.steps->'suggest_related_pages'->>'next_step' <> r.saver THEN
            RAISE EXCEPTION '602 VERIFY: %.suggest_related_pages does not lead to %', r.type, r.saver;
        END IF;
        -- The saver must be REACHED. A picker that routes nowhere would strand
        -- every tool build, which is the one way this change could do harm.
        IF r.steps->r.saver IS NULL THEN
            RAISE EXCEPTION '602 VERIFY: %.% is gone', r.type, r.saver;
        END IF;
        -- Both error_steps must land on the saver, so a picker failure degrades
        -- to today's behaviour rather than failing the build.
        IF r.steps->'load_site_page_names'->'config'->>'error_step' <> r.saver
           OR r.steps->'suggest_related_pages'->'config'->>'error_step' <> r.saver THEN
            RAISE EXCEPTION '602 VERIFY: %: an error_step does not fall through to %', r.type, r.saver;
        END IF;
        -- The new wire, and the 516 wire it defers to, must BOTH be present and
        -- BOTH be optional-explicit.
        IF NOT ((r.steps->r.saver->'config') ? 'related_pages_fallback?') THEN
            RAISE EXCEPTION '602 VERIFY: %.% has no related_pages_fallback? wire', r.type, r.saver;
        END IF;
        IF NOT ((r.steps->r.saver->'config') ? 'related_pages?') THEN
            RAISE EXCEPTION '602 VERIFY: %.% lost its related_pages? wire', r.type, r.saver;
        END IF;
        IF r.steps->r.saver->'config'->>'related_pages_fallback?' <> 'suggest_related_pages.result' THEN
            RAISE EXCEPTION '602 VERIFY: %.% fallback points at %, expected suggest_related_pages.result',
                r.type, r.saver, r.steps->r.saver->'config'->>'related_pages_fallback?';
        END IF;
        -- An UNMARKED twin would beat nothing but would reintroduce the search
        -- for this field, which is the entire point of 516.
        IF (r.steps->r.saver->'config') ? 'related_pages_fallback'
           OR (r.steps->r.saver->'config') ? 'related_pages' THEN
            RAISE EXCEPTION '602 VERIFY: %.% carries an UNMARKED twin of a related_pages key — that is the whole-tree search back again (bugs_open/330)', r.type, r.saver;
        END IF;
        -- Rule 5 is what keeps absence representable. If it is ever edited out,
        -- the picker can no longer answer "none".
        IF position('An empty answer is a CORRECT answer' in
                    COALESCE(r.steps->'suggest_related_pages'->'config'->>'prompt_template','')) = 0 THEN
            RAISE EXCEPTION '602 VERIFY: %: the picker prompt no longer permits an empty answer', r.type;
        END IF;
    END LOOP;

    SELECT count(*) INTO n FROM agent_definitions
     WHERE type IN ('tool-generator','tool-deployer') AND is_active
       AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;
    IF n <> 2 THEN
        RAISE EXCEPTION '602 VERIFY: expected 2 live tool agents, found %', n;
    END IF;

    RAISE NOTICE '602 VERIFY: both tool workflows ask for related pages, both fall through to the saver on error, and the requester still wins.';
END $$;

COMMIT;
