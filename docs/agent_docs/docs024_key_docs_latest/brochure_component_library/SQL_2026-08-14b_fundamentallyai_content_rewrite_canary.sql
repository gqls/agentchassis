-- Retire the two needs_page canaries and refile as content_rewrite.
--
-- WHY THE SWITCH (measured 2026-08-14 21:35Z, not assumed):
--   needs_page      110 complete / 230 failed  (68% failure) — a full page rebuild,
--                   and the item I cloned had itself errored "20 blockers".
--   content_rewrite  99 complete /  26 failed, and the closest analogue by intent —
--                   source 'voiceh-rollout', a voice rollout across a site — is
--                   32 complete / 0 failed.
-- Same handler (page-build-handler) either way, so this is a cheaper and better-
-- evidenced route to the same end, not a different mechanism.
--
-- The 'suggestion' field carries the INSTRUCTION; an LLM writes the prose. That is
-- the owner ruling of 2026-08-06 (the framework writes the content, not you)
-- honoured rather than worked around — I am not supplying sentences.

BEGIN;

DO $chk$
DECLARE n int;
BEGIN
  SELECT count(*) INTO n FROM site_work_items
   WHERE created_by='brochure_contrast_front_thread' AND item_type='needs_page' AND status='triaged';
  IF n <> 2 THEN RAISE EXCEPTION 'expected my 2 triaged needs_page items, found % — another session may have moved them', n; END IF;

  -- they must still be UNCLAIMED; cancelling a claimed item races the handler
  SELECT count(*) INTO n FROM site_work_items
   WHERE created_by='brochure_contrast_front_thread' AND item_type='needs_page'
     AND (claimed_by IS NOT NULL AND claimed_by <> '');
  IF n <> 0 THEN RAISE EXCEPTION 'a needs_page item is CLAIMED (n=%) — leave it alone, do not cancel mid-flight', n; END IF;

  SELECT count(*) INTO n FROM site_specs
   WHERE site_id='199733a8-ac9c-4c30-b2ce-65ecdac6f3bd' AND aspect='content_direction'
     AND is_current AND data->>'formatted' LIKE '%say what a thing IS%';
  IF n <> 1 THEN RAISE EXCEPTION 'corrected content_direction is not current'; END IF;
END $chk$;

UPDATE site_work_items
   SET status='cancelled',
       resolution_path='Superseded by a content_rewrite item for the same page. needs_page is a full rebuild with a measured 68% failure rate (110 complete / 230 failed, 2026-08-14); content_rewrite mode=edit_live is the proven idiom for a voice/prose pass (99/26 overall, 32/0 for the voiceh-rollout analogue). Nothing was dispatched for these; they sat triaged and unclaimed for 90 minutes.',
       handled_by='brochure_contrast_front_thread'
 WHERE created_by='brochure_contrast_front_thread' AND item_type='needs_page' AND status='triaged';

INSERT INTO site_work_items
  (site_id, page_id, item_type, item_key, summary, spec, handler_agent, pipeline,
   priority, severity, approval_mode, max_attempts, source, created_by, status)
SELECT
  '199733a8-ac9c-4c30-b2ce-65ecdac6f3bd',
  v.page_id::uuid,
  'content_rewrite',
  'positive-definition-rewrite:' || v.page_name,
  'Positive-definition rewrite - ' || v.page_name,
  jsonb_build_object(
    'mode','edit_live',
    'page', v.page_name,
    'page_id', v.page_id,
    'page_name', v.page_name,
    'work_item_type','content_rewrite',
    'max_fix_attempts',1,
    'description','Rewrite this page''s prose to define things positively, per the site Content Direction spec. Preserve all facts and all caveats.',
    'suggestion',
      'Rewrite the prose of this page to remove one specific habit, following the voice defined in this site''s Content Direction spec above. '
      || 'THE HABIT: defining a thing by what it is not - the "X, not Y" and "X rather than Y" construction. This page uses it '
      || v.hits || ' times. Examples on this page: ' || v.examples || ' '
      || 'THE RULE (owner ruling): say what a thing IS, not what it is not. A negative definition makes the reader do subtraction and reads colder, because it withholds. '
      || 'So state the positive fact and let it carry the contrast: "treat the result as a starting conversation" says everything "not a verdict" says, and tells the reader more. '
      || 'CRITICAL - THIS IS NOT AN INSTRUCTION TO REMOVE CAVEATS OR SOFTEN LIMITATIONS. The site''s honesty about what it cannot do, and about not yet having a paying client, is deliberate and must survive in full. '
      || 'Restate each limitation as what IS true rather than deleting it: "this estimate assumes a fixed rate" instead of "this is not a forecast". If a sentence''s only content is a caveat, keep the caveat and change only its grammar. '
      || 'PRESERVE EXACTLY: every factual claim, figure, percentage, date, named institution, product name, and internal link, unchanged. Do not add or remove a fact. '
      || 'Keep the existing heading structure. Do not introduce form controls, ids or scripts. Return HTML with each paragraph wrapped in <p>.',
    'acceptance_test',
      'Reader-visible prose no longer defines things by negation ("X, not Y" / "X rather than Y") except where a genuine comparison is being drawn; every original figure, named entity and internal link is retained; and every limitation or caveat present before is still present, stated positively.'
  ),
  'page-build-handler','build',60,'medium','auto',3,
  'operator:brochure_component_library','brochure_contrast_front_thread','triaged'
FROM (VALUES
  ('model-approach-selector-guide','2f0eb560-04aa-4a7c-b15f-2832dcc46f65','10',
   '"It''s a decision aid, not a verdict." / "a starting recommendation, not a verdict" / "Treat the result as a starting position you can argue with, not a specification handed down from above." / "that''s a normal outcome, not a failure of the tool."'),
  ('multi-agent-review-council','5f855405-1dfd-44a2-9da2-9ba8f8d70bd3','3',
   '"What the council produces is a decision record, not a log entry." / "it writes a decision record - a real artefact we can show you, not a log entry." / "checking the output against a defined criterion rather than a general sense of quality."')
) AS v(page_name, page_id, hits, examples);

DO $post$
DECLARE n int; s text;
BEGIN
  SELECT count(*) INTO n FROM site_work_items
   WHERE created_by='brochure_contrast_front_thread' AND item_type='content_rewrite' AND status='triaged';
  IF n <> 2 THEN RAISE EXCEPTION 'post: expected 2 content_rewrite items, found %', n; END IF;

  SELECT count(*) INTO n FROM site_work_items
   WHERE created_by='brochure_contrast_front_thread' AND item_type='needs_page' AND status='cancelled';
  IF n <> 2 THEN RAISE EXCEPTION 'post: expected 2 cancelled needs_page items, found %', n; END IF;

  -- the caveat-protection clause must be in BOTH suggestions, or the rewrite may
  -- strip the site's honesty and report success
  SELECT count(*) INTO n FROM site_work_items
   WHERE created_by='brochure_contrast_front_thread' AND item_type='content_rewrite'
     AND spec->>'suggestion' LIKE '%must survive in full%';
  IF n <> 2 THEN RAISE EXCEPTION 'post: caveat-protection clause missing from % of 2 suggestions', 2-n; END IF;
END $post$;

SELECT id, spec->>'page_name' AS page, item_type, status, priority,
       length(spec->>'suggestion') AS suggestion_len
  FROM site_work_items WHERE created_by='brochure_contrast_front_thread'
 ORDER BY item_type, created_at;

COMMIT;
