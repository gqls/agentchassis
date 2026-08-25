-- bugs_open/391 step 2, the remaining 11 label-locked pages (canary already done).
--
-- ONE PAIR PER PAGE, and the ORDER IS ENFORCED BY THE PLATFORM, not by a human:
-- each page_rerender carries depends_on = ARRAY[<its content_rewrite id>], and
-- load_work_item_actions.go's eligibility query refuses a row whose depends_on is not
-- yet 'complete'/'verified'. So a page's relink cannot be served before its rewrite
-- lands, and the pages progress independently instead of waiting on one another.
-- This is why all 11 can be dispatched at once without putting every page into the
-- text-says-one-thing/href-says-another state simultaneously.
--
-- spec.suggestion (NOT content_guidance) is the key the handler reads — bugs_open/271.
-- The FRAMEWORK writes the copy; the guidance only supplies each site's real tool list
-- and the constraint (owner rule 2026-08-06).
BEGIN;

CREATE TEMP TABLE tgt(domain text, site_id uuid, page_id uuid, page_name text) ON COMMIT DROP;
INSERT INTO tgt VALUES
 ('ai-agent-orchestration.com','2a8ebf9c-20a2-4c39-b191-840b012371da','9baea9f9-4d4c-449d-949c-e5753d4faf67','index'),
 ('ai-agent-orchestration.com','2a8ebf9c-20a2-4c39-b191-840b012371da','2c7c836c-98e7-4600-bcbe-8a8f884abcb7','model-directory'),
 ('ai-agent-orchestration.com','2a8ebf9c-20a2-4c39-b191-840b012371da','35c46cc0-fec3-4c22-91fd-10369acd0a6c','news-index'),
 ('ai-agent-orchestration.com','2a8ebf9c-20a2-4c39-b191-840b012371da','356a8f3f-62e3-4610-b22a-ff0f0017c0fe','pricing'),
 ('ai-agent-orchestration.com','2a8ebf9c-20a2-4c39-b191-840b012371da','f801e96d-f92a-45f8-b050-6394a8378259','tool-automation-savings-estimator-guide'),
 ('ai-agent-orchestration.com','2a8ebf9c-20a2-4c39-b191-840b012371da','7d2b1baf-39cc-41f4-8d64-fb3679db04de','tools'),
 ('finetuning.uk','1368e337-dd1d-4799-bbb3-8221a1b79bcc','1b419ed4-6ef7-4faf-8be0-647d5400947a','chatgpt-has-your-data-does-that-matter'),
 ('finetuning.uk','1368e337-dd1d-4799-bbb3-8221a1b79bcc','1e0f6525-ea9d-44d7-9353-1ba96c3f0123','services'),
 ('finetuning.uk','1368e337-dd1d-4799-bbb3-8221a1b79bcc','a8909fc1-f1ff-43fe-842c-5ce364b8b182','your-own-model'),
 ('leopardessconsulting.co.uk','4851f6fc-71cf-4160-a270-e03d6d3e0732','ebc2c413-61e2-465e-b22b-9aab0167abc9','services'),
 ('leopardessconsulting.co.uk','4851f6fc-71cf-4160-a270-e03d6d3e0732','ffc7d214-a8c5-445c-89cf-846e782846f8','tools');

CREATE TEMP TABLE tools(domain text, list text) ON COMMIT DROP;
INSERT INTO tools VALUES
 ('ai-agent-orchestration.com',
  'Agent Architecture Complexity Estimator (/tools/agent-complexity-estimator.html); AI Agent ROI Estimator (/tools/tool-ai-agent-roi-estimator.html); AI Automation Time Savings Estimator (/tools/automation-savings-estimator/index.html); Multi-Agent Build vs Buy Decision Analyzer (/tools/build-vs-buy-analyzer/index.html); LLM Provider Cost Comparison Calculator (/tools/tool-llm-cost-calculator.html).'),
 ('finetuning.uk',
  'AI Agent ROI Estimator (/tools/ai-agent-roi-estimator.html); LLM Provider Cost Comparison Calculator (/tools/llm-cost-calculator.html); Fine-Tuning vs RAG vs Prompting Decision Guide (/tools/model-approach-selector.html); AI Data Risk Checker (/tools/tool-ai-data-risk-checker.html); AI Project Readiness Checker (/tools/ai-readiness-checker/index.html); AI Readiness Quiz (/tools/tool-ai-readiness-quiz.html); AI Automation Time Savings Estimator (/tools/automation-savings-estimator/index.html); GDPR & AI Data Risk Self-Assessment (/tools/gdpr-ai-risk-assessment/index.html).'),
 ('leopardessconsulting.co.uk',
  'AI Agent ROI Estimator (/tools/ai-agent-roi-estimator.html); LLM Provider Cost Comparison Calculator (/tools/llm-cost-calculator.html); Agent Architecture Complexity Estimator (/tools/tool-agent-complexity-estimator.html); AI Vendor Trust Checklist (/tools/ai-vendor-trust-checklist.html); AI Automation Time Savings Estimator (/tools/automation-savings-estimator/index.html); Process Automation Suitability Scorer (/tools/process-automation-scorer/index.html).');

WITH rw AS (
  INSERT INTO site_work_items
    (site_id, page_id, source, created_by, item_type, handler_agent, status, severity, priority,
     summary, item_key, spec)
  SELECT t.site_id, t.page_id, 'bugfix_391_cta_relevance', 'bugfix_391_cta_relevance',
         'content_rewrite', 'page-build-handler', 'detected', 'medium', 35,
         'Reword the CTA labels on ' || t.page_name || ': they name the Password Strength Physics tool, which is off-topic here and is being retired',
         'cta_label_relevance:' || t.page_id::text,
         jsonb_build_object(
           'mode','edit_live', 'source','bugfix_391_cta_relevance', 'page_name', t.page_name,
           'suggestion',
             'Reword ONLY the call-to-action button LABELS on this page. Leave all other prose exactly as it is. '
             'The buttons currently invite the reader to try the "Password Strength Physics" tool, which has nothing to do '
             'with this page or with this business, and that tool is being retired from this site. Choose whichever ONE of '
             'the site''s real tools is most relevant to what THIS page is about, and word each button so it plainly names '
             'that tool and says what the reader will get from it. Do not mention passwords, password strength or entropy '
             'anywhere. Do not invent a tool that is not in this list and do not link to anything outside it. '
             'The button label is what the framework matches to a destination, so the label must name the tool clearly. '
             'The site''s tools are: ' || tl.list)
  FROM tgt t JOIN tools tl ON tl.domain = t.domain
  RETURNING id, site_id, page_id
)
INSERT INTO site_work_items
  (site_id, page_id, source, created_by, item_type, handler_agent, status, severity, priority,
   summary, item_key, spec, depends_on)
SELECT rw.site_id, rw.page_id, 'bugfix_391_cta_relevance', 'bugfix_391_cta_relevance',
       'page_rerender', 'page-rerender', 'detected', 'high', 35,
       'Re-resolve CTA hrefs on ' || t.page_name || ' after its labels were reworded (depends on the rewrite)',
       'cta_relink:' || rw.page_id::text,
       jsonb_build_object(
         'reason','cta_links_stale', 'check','misdirected_cta',
         'page_id', rw.page_id::text, 'page_name', t.page_name,
         'fix','The CTA labels now name a different tool; recompute each href so it follows its own copy.',
         'original_pipeline','build'),
       ARRAY[rw.id]
FROM rw JOIN tgt t ON t.page_id = rw.page_id;

\echo '--- 22 items: 11 rewrites + 11 relinks, each relink blocked on its rewrite ---'
SELECT item_type, count(*), count(*) FILTER (WHERE depends_on IS NOT NULL) AS with_dependency
FROM site_work_items WHERE created_by='bugfix_391_cta_relevance' AND status='detected'
GROUP BY 1 ORDER BY 1;
COMMIT;
