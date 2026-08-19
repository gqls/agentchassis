-- ✅ HOLD SATISFIED 2026-08-19 — the precondition was MET at the artefact, not assumed.
--    ⚠ AND THE `_HOLD` SUFFIX HAD TO GO, for a reason worth knowing before you use the
--    convention: `--record-only` REFUSES an uppercase-suffixed sidecar ("Sidecars are
--    never applied by the runner, so recording one is meaningless"). So a file that is
--    held, then hand-applied, then left named `_HOLD` is APPLIED WITH NO LEDGER ROW —
--    invisible to every "what is applied" query and to the drift check. Rename it back
--    the moment the hold is satisfied, then record. Applied by hand while still held;
--    renamed and recorded immediately after.
--
--    THE EVIDENCE, in the order CLAUDE.md asks for it:
--      chassis           v1.0.1315
--      pod imageID       sha256:1c422c614bd7f15457401d57abc0f918140388d74f7461283a1c4694c54d56a8
--      local RepoDigests CONTAINS that digest  -> a real build, not a same-tag rebuild
--                                                 serving the node's cached image
--      revision label    590ca3a20cca99e0f6e9c6a2545bd8e94c11b9ae
--      ancestry          aeccfc595 (action+guard), f2dd88f31 (3 more guards),
--                        5b236ccc8, 6cb82f1d7 (copy gates) are ALL ancestors of it
--      control           HEAD is correctly NOT an ancestor, so the check discriminates
--      binary probe      `save_page_meta_description` PRESENT and `voice_gate_unreadable`
--                        PRESENT in /proc/1/exe; positive control `rebuild_blog_listing`
--                        PRESENT, negative control (a fake symbol) ABSENT
--
--    APPLY BY HAND (not `--apply`, which takes every pending file including other
--    sessions'), then record:
--      kubectl -n ai-persona-system exec -i postgres-clients-0 -- \
--        psql -U clients_user -d clients_db -v ON_ERROR_STOP=1 \
--        < docs/agent_docs/sql_for_agents/488_meta_description_backfiller_agent.sql
--      ./scripts/migration/run-migrations.sh --record-only 488_meta_description_backfiller_agent.sql --note "<why>"
--
-- 488 — seed `meta-description-backfiller`, the workflow that DRIVES SEO-004
--       (bugs_open/320; owner chose the full fix 2026-08-19)
--
-- WHY THIS FILE EXISTS AT ALL. Migration 485 stopped new pages being born without
-- a meta description, and the Go half stopped a replan blanking one that exists.
-- Neither fills the pages that are ALREADY empty — 407 of 731 active pages
-- (55.7%) across 26 of 27 sites, measured 2026-08-19. 295 of those are
-- plan-managed and a later replan now reaches them; **112 are not**, and nothing
-- reaches those at all.
--
-- ⚠ AND WITHOUT THIS FILE, `save_page_meta_description` HAS NO CALLER. A helper
-- with no callers looks exactly like a finished refactor — the estate has a
-- standing landmine about precisely that. Registering an action is not shipping a
-- capability; something has to drive it.
--
-- THE SHAPE, AND WHY IT IS THREE EXISTING PARTS PLUS ONE NEW ONE:
--   load_pages_missing_meta  query_database            (existing machinery)
--   write_descriptions       execute_llm_prompt        (existing machinery)
--   backfill_loop            loop -> save_page_meta_description   (SEO-004, new)
-- Finding and writing were never missing; only persistence was. Keeping it this
-- way is also what keeps authorship with the FRAMEWORK rather than with a session
-- (owner ruling 2026-08-06) — the sentence is written by an LLM from the page's
-- own title and rendered content, not by a Go format string and not by me.
--
-- SAFETY, stated rather than left to be inferred:
--   * The action's `overwrite_existing` is NOT set here, so it takes its default
--     of FALSE (owner ruling 2026-08-02 §2 — opt-in, unsafe side OFF). This
--     workflow therefore CANNOT replace copy that already exists; it can only
--     fill a blank. That is the whole of what 320 needs.
--   * The query selects only pages that are blank AND have rendered content to
--     describe. A page with no content would otherwise get a description
--     hallucinated from its title alone, which is worse than none.
--   * `max_iterations` 25 per run, bounded deliberately: this writes public copy,
--     so a first run should be readable in one sitting. It is not a throughput
--     mechanism and should not become one without a look at the output.
--   * The action refuses brief-shaped text (`MetaDescriptionLooksInternal`, the
--     bugs_closed/103 guard) and anything over 320 chars, so a bad LLM answer is
--     dropped with a named reason rather than published.
--
-- ⚠ NOT SCHEDULED, and the reason CHANGED on 2026-08-19 — read this before wiring a cron.
--
-- OWNER RULING 2026-08-19 (recorded as `bugs_open/320` §11): the backfill is authorised
-- FLEET-WIDE and the owner explicitly WAIVED the read-first review pass ("I don't need a
-- review"). So the earlier "one site at a time with someone reading it" caution in this
-- lane's README is SUPERSEDED and must not be quoted back as policy.
--
-- What the owner attached as the CONDITION of that waiver, in the same exchange:
--
--     "please make sure the summaries go through the copy guidance and checks so they
--      don't sound like AI"
--
-- The checks REPLACE the review, so both halves are load-bearing:
--   * the GUIDANCE is in this prompt (the "House style" block above, taken from
--     REVERSE_ENGINEERED_STYLE_PROMPT_v3.md);
--   * the CHECKS run in the ACTION, before the write, and cannot be skipped by editing
--     this workflow: `save_page_meta_description` runs the site's own VoiceGate and the
--     banned-claims sweep over the candidate and refuses with a named reason.
--     ⚠ It deliberately does NOT call `check_voice_tells`, which scans
--     `page_components.rendered_html` and is BLIND to `pages.meta_description` — wiring
--     that would have produced a confident pass over text it never examined.
--
-- ONLY THE STYLE RULES THAT GOVERN A SINGLE SENTENCE WERE CARRIED OVER, and that is a
-- judgement worth stating rather than hiding: v3 is 202 lines written for pitch prose,
-- and its rules 7-13 are about headings, markdown tables, paragraph rhythm and landing
-- beats. Pasting those into a prompt for a 155-character sentence would invite the model
-- to emit a table. Rules 1-6 (one idea per sentence, no em dash, no negative-frame
-- opening, word-weight in both directions, no self-flagging hedges, contractions) are
-- the ones that apply, and they are the ones present.
--
-- STILL NO CRON, for an engineering reason rather than a review one: this is the first
-- execution of brand-new code, so run ONE site, read the `save_result` reasons, and only
-- then widen. That is the bugfix-203 canary lesson (run it on one page against a pinned
-- before-state), not the review the owner waived.
--
-- Dispatch: input_data {"site_id": "<uuid>"}  (or {"domain": "<domain>"})
--
-- VERIFY AT THE SERVED PAGE, not at the row: curl the page and read
-- `<meta name="description" ...>`. `rerender_single_page_action.go` strips an
-- EMPTY description tag rather than serving it, so a page that still has none
-- shows an ABSENT tag, not an empty one — and a DB row updated before the page is
-- rerendered will disagree with what a visitor gets.
--
-- ROLLBACK: 488_meta_description_backfiller_agent_ROLLBACK.sql

BEGIN;

DO $$
DECLARE n int;
BEGIN
  SELECT count(*) INTO n FROM agent_definitions
   WHERE type = 'meta-description-backfiller'
     AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;
  IF n <> 0 THEN
    RAISE EXCEPTION '488: a meta-description-backfiller row already exists (% found) — refusing to double-seed', n;
  END IF;
END $$;

-- display_name and category are NOT NULL with no default (checked with
-- information_schema before writing this, after a first attempt omitted
-- display_name and was refused — the transaction rolled back cleanly, 0 rows).
-- Values follow the house convention: Title Case of the type, category
-- 'specialist' (as build-site-planner, internal-linker and tool-suggester use).
INSERT INTO agent_definitions (type, display_name, category, description, is_active, default_config)
VALUES (
  'meta-description-backfiller',
  'Meta Description Backfiller',
  'specialist',
  'Fills pages.meta_description on pages that have none (bugs_open/320, SEO-004). Cannot overwrite existing copy: save_page_meta_description defaults overwrite_existing=false, and the action runs the site voice gate and banned-claims sweep before any write.',
  true,
  jsonb_build_object(
    'workflow', jsonb_build_object(
      'start_step', 'ensure_site_record',
      'processing_mode', 'sequential',
      'timeout_seconds', 1800,
      'steps', jsonb_build_object(

        'ensure_site_record', jsonb_build_object(
          'action', 'ensure_site_record',
          'next_step', 'load_pages_missing_meta',
          'output_field', 'site_record',
          'description', 'Resolve site_id/domain into site_record'
        ),

        'load_pages_missing_meta', jsonb_build_object(
          'action', 'query_database',
          'config', jsonb_build_object(
            'query',
              'SELECT p.id, p.name, p.url, p.title, p.page_type, ' ||
              'LEFT(string_agg(pc.rendered_html, '' ''), 1200) AS content_sample ' ||
              'FROM pages p ' ||
              'JOIN page_components pc ON pc.page_id = p.id ' ||
              '  AND pc.rendered_html IS NOT NULL ' ||
              '  AND COALESCE(pc.slot_name, '''') NOT IN (''header'',''footer'',''head'') ' ||
              'WHERE p.site_id = $1 ' ||
              '  AND p.status = ''active'' ' ||
              '  AND COALESCE(p.meta_description, '''') = '''' ' ||
              'GROUP BY p.id, p.name, p.url, p.title, p.page_type ' ||
              'HAVING length(string_agg(pc.rendered_html, '' '')) > 400 ' ||
              'ORDER BY p.name LIMIT 25',
            'params', jsonb_build_array('site_record.site_id'),
            'output_format', 'array'
          ),
          'next_step', 'check_has_pages',
          'output_field', 'pages_missing_meta',
          'description', 'Active pages with a BLANK meta_description and enough rendered content to describe. Chrome slots excluded so a page is not judged by its header.'
        ),

        'check_has_pages', jsonb_build_object(
          'action', 'conditional',
          'config', jsonb_build_object(
            'condition', 'pages_missing_meta.count > 0',
            'then_step', 'write_descriptions',
            'else_step', 'complete_nothing_to_do'
          )
        ),

        'complete_nothing_to_do', jsonb_build_object(
          'action', 'complete_workflow',
          'config', jsonb_build_object(
            'result_message', 'No active page on this site has a blank meta_description with content to describe.'
          )
        ),

        'write_descriptions', jsonb_build_object(
          'action', 'execute_llm_prompt',
          'config', jsonb_build_object(
            'ai_service', jsonb_build_object(
              'provider', 'anthropic',
              'model', 'claude-haiku-4-5',
              'max_tokens', 4000,
              'api_key_env_var', 'ANTHROPIC_API_KEY'
            ),
            'input_fields', jsonb_build_array('pages_missing_meta', 'site_record'),
            'output_format', 'json',
            'prompt_template',
              E'You are writing meta descriptions for pages on {{.site_record.domain}}.\n\n' ||
              E'A meta description is the sentence a search engine prints under the page title in its results. It is read by far more people than the page itself, so it is a promise to a visitor, not a summary for a machine.\n\n' ||
              E'## Pages needing a description\n' ||
              E'{{range .pages_missing_meta}}### {{.name}}\n' ||
              E'id: {{.id}}\n' ||
              E'url: {{.url}}\n' ||
              E'title: {{.title}}\n' ||
              E'type: {{.page_type}}\n' ||
              E'content: {{.content_sample}}\n\n' ||
              E'{{end}}\n' ||
              E'## House style (from REVERSE_ENGINEERED_STYLE_PROMPT_v3.md, the rules that govern a single sentence)\n' ||
              E'These were reverse-engineered from copy the owner hand-edited and judged more readable. Follow them.\n' ||
              E'- One idea, one sentence. Do not chain clauses with commas or semicolons to fit more in.\n' ||
              E'- No em dashes. Not one.\n' ||
              E'- Never open with what something is NOT. "Not just a calculator, but a..." and "More than a guide" are the same move: a manufactured twist. Start with the fact, the way a person saying it out loud would.\n' ||
              E'- Match word-weight to the claim, in BOTH directions. "Powerful", "seamless", "comprehensive", "revolutionary" overclaim upward; "simple", "just", "nothing fancy" overclaim downward, which still asks the reader to be impressed. Neither belongs in a description of a web page.\n' ||
              E'- Cut self-flagging commentary and hedges: crucially, genuinely, exactly, deliberately, what matters here is, at its core, in essence, delve, leverage, robust, seamless, furthermore, moreover. Do not tell the reader something is important. State it.\n' ||
              E'- Contractions are fine and usually better: it is -> it''s, does not -> doesn''t.\n' ||
              E'- Do not name the site or brand. The search engine already prints it above this line.\n\n' ||
              E'## Rules\n' ||
              E'- 120-155 characters. Shorter wastes the slot; longer is cut off mid-word.\n' ||
              E'- One sentence, plain English, British spelling.\n' ||
              E'- Say what the visitor GETS from this page. Lead with that, not with the company name.\n' ||
              E'- Do not append the site or brand name; the search engine already shows it.\n' ||
              E'- No build, generator or specification wording. Never describe how the page was made.\n' ||
              E'- No em dashes.\n' ||
              E'- Ground it in the content given. If a page''s content does not support a specific description, omit that page entirely rather than inventing one. Returning fewer entries than you were given is a correct answer.\n\n' ||
              E'Return ONLY valid JSON:\n' ||
              E'{"descriptions": [{"page_id": "<the id exactly as given>", "meta_description": "<the sentence>"}]}\n'
          ),
          'next_step', 'backfill_loop',
          'output_field', 'written'
        ),

        'backfill_loop', jsonb_build_object(
          'action', 'loop',
          'config', jsonb_build_object(
            'items_field', 'written.result.descriptions',
            'item_variable', 'current_description',
            'max_iterations', 25,
            'continue_on_error', true,
            'sub_workflow', jsonb_build_object(
              'start_step', 'save_description',
              'steps', jsonb_build_object(
                'save_description', jsonb_build_object(
                  'action', 'save_page_meta_description',
                  'config', jsonb_build_object(
                    'page_id_field', 'current_description.page_id',
                    'description_field', 'current_description.meta_description'
                    -- overwrite_existing deliberately ABSENT: the action defaults
                    -- it to false, so this workflow can only fill a blank.
                  ),
                  'next_step', 'done',
                  'output_field', 'save_result'
                ),
                'done', jsonb_build_object('action', 'loop_complete')
              )
            )
          ),
          'next_step', 'complete',
          'output_field', 'saved'
        ),

        'complete', jsonb_build_object(
          'action', 'complete_workflow',
          'config', jsonb_build_object(
            'result_message', 'Meta description backfill finished. Read each save_result: "updated" true is a write, false carries a named reason (empty_candidate / candidate_looks_internal / candidate_too_long / already_has_description).'
          )
        )
      )
    )
  )
);

DO $$
DECLARE cfg jsonb;
BEGIN
  SELECT default_config INTO cfg FROM agent_definitions
   WHERE type = 'meta-description-backfiller'
     AND is_active AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;

  IF cfg IS NULL THEN
    RAISE EXCEPTION '488 VERIFY: the row was not inserted';
  END IF;
  IF cfg#>>'{workflow,start_step}' IS DISTINCT FROM 'ensure_site_record' THEN
    RAISE EXCEPTION '488 VERIFY: start_step is %, expected ensure_site_record', cfg#>>'{workflow,start_step}';
  END IF;
  IF cfg#>>'{workflow,steps,backfill_loop,config,sub_workflow,steps,save_description,action}'
       IS DISTINCT FROM 'save_page_meta_description' THEN
    RAISE EXCEPTION '488 VERIFY: the loop does not call save_page_meta_description — the whole point of this agent';
  END IF;
  -- The safety property, asserted rather than trusted: if a later edit ever sets
  -- overwrite_existing true here, this block is what should be updated to say so
  -- deliberately, and until then its ABSENCE is the control.
  IF cfg#>'{workflow,steps,backfill_loop,config,sub_workflow,steps,save_description,config}' ? 'overwrite_existing' THEN
    RAISE EXCEPTION '488 VERIFY: overwrite_existing is set on the save step — this workflow must not be able to replace existing copy';
  END IF;
  IF (cfg#>>'{workflow,steps,backfill_loop,config,max_iterations}')::int > 25 THEN
    RAISE EXCEPTION '488 VERIFY: max_iterations is above the reviewed bound of 25';
  END IF;
  RAISE NOTICE '488 OK: meta-description-backfiller seeded, fill-blanks-only, 25 pages per run, NOT scheduled';
END $$;

COMMIT;
