-- bugs_open/391 retirement step 2 (owner decision 5, re-scoped commission): CANARY.
-- One content_rewrite item on finetuning.uk/technical-details, whose hero + call-to-action
-- labels both name the password tool and are therefore label-locked (a nav_order fix cannot
-- reach them — see 391 §THE FEEDBACK LOOP).
--
-- WHY spec.suggestion AND NOT spec.content_guidance: `suggestion` is the key the handler
-- actually reads. `content_guidance` is only ALIASED into it (bugs_open/271,
-- load_work_items_guidance_alias_test.go) and an author-supplied `suggestion` wins over the
-- alias. Writing the read key directly removes the dependency on the alias shipping.
--
-- The FRAMEWORK writes the copy, not this file (owner rule 2026-08-06): the guidance supplies
-- the site's real tool list and the constraint, and the writer chooses and words it.
BEGIN;

INSERT INTO site_work_items
  (site_id, page_id, source, created_by, item_type, handler_agent, status, severity, priority,
   summary, item_key, spec)
VALUES (
  '1368e337-dd1d-4799-bbb3-8221a1b79bcc',
  'a32b8822-db49-4e45-88f8-bda06d73de62',
  'bugfix_391_cta_relevance', 'bugfix_391_cta_relevance',
  'content_rewrite', 'page-build-handler', 'detected', 'medium', 35,
  'Reword the CTA labels on technical-details: both name the Password Strength Physics tool, which is off-topic for this page and is being retired',
  'cta_label_relevance:a32b8822-db49-4e45-88f8-bda06d73de62',
  jsonb_build_object(
    'mode', 'edit_live',
    'source', 'bugfix_391_cta_relevance',
    'page_name', 'technical-details',
    'suggestion',
      'Reword ONLY the call-to-action button LABELS on this page. Leave all other prose exactly as it is. '
      'Both buttons currently invite the reader to try the "Password Strength Physics" tool, which has nothing '
      'to do with this page or with this business, and that tool is being retired from this site. '
      'Choose whichever ONE of the site''s real tools is most relevant to what THIS page is about, and word each '
      'button so it plainly names that tool and says what the reader will get from it. Do not mention passwords, '
      'password strength or entropy anywhere. Do not invent a tool that is not in this list, and do not link to '
      'anything outside it. The site''s tools are: '
      'AI Agent ROI Estimator (/tools/ai-agent-roi-estimator.html); '
      'LLM Provider Cost Comparison Calculator (/tools/llm-cost-calculator.html); '
      'Fine-Tuning vs RAG vs Prompting Decision Guide (/tools/model-approach-selector.html); '
      'AI Data Risk Checker (/tools/tool-ai-data-risk-checker.html); '
      'AI Project Readiness Checker (/tools/ai-readiness-checker/index.html); '
      'AI Readiness Quiz (/tools/tool-ai-readiness-quiz.html); '
      'AI Automation Time Savings Estimator (/tools/automation-savings-estimator/index.html); '
      'GDPR & AI Data Risk Self-Assessment (/tools/gdpr-ai-risk-assessment/index.html). '
      'The button label is what the framework matches to a destination, so the label must name the tool clearly.'
  )
);

\echo '--- the canary item ---'
SELECT id, status, priority, item_key, spec->>'mode' AS mode, left(spec->>'suggestion',60) AS guidance
FROM site_work_items WHERE item_key='cta_label_relevance:a32b8822-db49-4e45-88f8-bda06d73de62';

COMMIT;
