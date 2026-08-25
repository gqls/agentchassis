-- Retry the ONE page of the eleven whose pair failed: ai-agent-orchestration.com
-- /model-directory.html (page 2c7c836c-98e7-4600-bcbe-8a8f884abcb7).
--
-- WHY THE FIRST ATTEMPT FAILED — and it was NOT the CTA work.
-- content_rewrite 0745e9a4 stopped at step validate_content:
--   "0 blockers, 1 errors" / unregistered_number "150"
-- The refused sentence is the CTA component's own <h2>, and it is ALREADY LIVE on the
-- served page -- my rewrite did not introduce it:
--   "More than 150 agents are listed here. Every one of them still needs a production
--    stack underneath it."
--
-- Two separate things are true about that sentence, both measured 2026-08-25:
--
-- 1. It carries no phrase the register licenses. numberSupported() (claims.go:1256)
--    gates every fact behind its context_terms BEFORE it compares the number. Fact
--    `aao-agent-definitions` (value 200, tolerance gte) would have supported 150, but
--    its terms are "agent definition", "agents in the registry", "ai agents",
--    "agents in production", ... and the +/-70-char claim window round the number
--    (claimWindow, claims.go:1349) contains none of them -- "agents are listed here"
--    matches nothing. So the fact is skipped and the number is reported unregistered.
--    The gate is behaving exactly as designed.
--
-- 2. The claim is also FALSE as written, by the page's OWN data. The listing is
--    rendered client-side by /tools/assets/model-directory-listing.js, which fetches
--    /data/model-directory-full.json. That file (HTTP 200, updated_at
--    2026-08-25T18:26:58Z) carries "count": 30 and exactly 30 entries, and the served
--    HTML holds 30 `class="model-card"` articles. Thirty are listed here, not 150.
--    The 150 appears to have been borrowed from the platform's agent-definition count,
--    which is a different quantity about a different thing.
--    (Routed to the owning lane: docs024_key_docs_latest/model_directory_pipeline/.)
--
-- So this retry asks for BOTH: the CTA labels off the retiring password tool, AND a
-- headline that stops asserting a figure the page contradicts. The writer writes the
-- sentence -- this spec states the constraint and the ground truth, per the 2026-08-04
-- ruling that the framework writes the content.
--
-- New item_key: the original key is still held by the needs_human_review row, and
-- item_keys dedup in ANY status (bugs_open/326).

BEGIN;

WITH rw AS (
  INSERT INTO site_work_items
    (site_id, page_id, source, created_by, item_type, handler_agent, status, severity, priority,
     summary, item_key, spec)
  VALUES (
    '2a8ebf9c-20a2-4c39-b191-840b012371da',
    '2c7c836c-98e7-4600-bcbe-8a8f884abcb7',
    'bugfix_391_cta_relevance', 'bugfix_391_cta_relevance',
    'content_rewrite', 'page-build-handler', 'detected', 'medium', 35,
    'Retry model-directory: reword the CTA labels off the retiring password tool AND drop the unregistered "more than 150 agents" figure that refused the first attempt',
    'cta_label_relevance_retry:2c7c836c-98e7-4600-bcbe-8a8f884abcb7',
    jsonb_build_object(
      'mode','edit_live', 'source','bugfix_391_cta_relevance', 'page_name','model-directory',
      'suggestion',
        'Two changes on this page, and nothing else. Leave all other prose exactly as it is. '
        'FIRST, the call-to-action button LABELS: they currently invite the reader to try the '
        '"Password Strength Physics" tool, which has nothing to do with this page or with this '
        'business, and that tool is being retired from this site. Choose whichever ONE of the '
        'site''s real tools is most relevant to what THIS page is about, and word each button so '
        'it plainly names that tool and says what the reader will get from it. Do not mention '
        'passwords, password strength or entropy anywhere. Do not invent a tool that is not in '
        'this list and do not link to anything outside it. The button label is what the framework '
        'matches to a destination, so the label must name the tool clearly. '
        'SECOND, the call-to-action heading currently reads "More than 150 agents are listed here." '
        'That figure is wrong and it is why the previous attempt at this page was refused. This '
        'directory lists THIRTY models -- that is what the page''s own data file reports and what '
        'the page renders. Rewrite that heading so it does not assert a count at all: say what the '
        'directory is for and what still has to be true to run any of these models in production. '
        'Do not put any new number in it, and do not simply change 150 to 30. '
        'The site''s tools are: Agent Architecture Complexity Estimator '
        '(/tools/agent-complexity-estimator.html); AI Agent ROI Estimator '
        '(/tools/tool-ai-agent-roi-estimator.html); AI Automation Time Savings Estimator '
        '(/tools/automation-savings-estimator/index.html); Multi-Agent Build vs Buy Decision '
        'Analyzer (/tools/build-vs-buy-analyzer/index.html); LLM Provider Cost Comparison '
        'Calculator (/tools/tool-llm-cost-calculator.html).')
  )
  RETURNING id, site_id, page_id
)
INSERT INTO site_work_items
  (site_id, page_id, source, created_by, item_type, handler_agent, status, severity, priority,
   summary, item_key, spec, depends_on)
SELECT rw.site_id, rw.page_id, 'bugfix_391_cta_relevance', 'bugfix_391_cta_relevance',
       'page_rerender', 'page-rerender', 'detected', 'high', 35,
       'Re-resolve CTA hrefs on model-directory after its labels were reworded (depends on the retry rewrite)',
       'cta_relink_retry:2c7c836c-98e7-4600-bcbe-8a8f884abcb7',
       jsonb_build_object(
         'reason','cta_links_stale', 'check','misdirected_cta',
         'page_id', rw.page_id::text, 'page_name','model-directory',
         'fix','The CTA labels now name a different tool; recompute each href so it follows its own copy.',
         'original_pipeline','build'),
       ARRAY[rw.id]
FROM rw;

-- The dependency is the whole reason a single batch is safe: between the rewrite and
-- the relink the page would otherwise serve a button whose text names one tool and
-- whose href points at another (bugs_closed/299). Prove it is armed.
DO $$
DECLARE n_dep int;
BEGIN
  SELECT count(*) INTO n_dep FROM site_work_items
   WHERE item_key = 'cta_relink_retry:2c7c836c-98e7-4600-bcbe-8a8f884abcb7'
     AND depends_on IS NOT NULL AND array_length(depends_on,1) = 1;
  IF n_dep <> 1 THEN
    RAISE EXCEPTION 'relink was created WITHOUT its depends_on (% rows matched)', n_dep;
  END IF;
END $$;

SELECT item_type, status, item_key, depends_on IS NOT NULL AS blocked
FROM site_work_items
WHERE item_key IN ('cta_label_relevance_retry:2c7c836c-98e7-4600-bcbe-8a8f884abcb7',
                   'cta_relink_retry:2c7c836c-98e7-4600-bcbe-8a8f884abcb7')
ORDER BY item_type;

COMMIT;
