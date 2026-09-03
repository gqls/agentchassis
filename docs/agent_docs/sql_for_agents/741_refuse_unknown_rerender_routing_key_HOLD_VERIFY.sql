-- 741_refuse_unknown_rerender_routing_key_HOLD_VERIFY.sql
--
-- Read-only. Run AFTER 741 applies. Answers three separate questions, in order,
-- because the answers have different consequences:
--
--   A. Did the flip land, and is the shape the one livespec declares?
--   B. Is the CHECK constraint safe to VALIDATE yet (owner ruling D3)?
--   C. Is anything being refused, and is it being refused CORRECTLY?
--
-- ⚠ This file never runs by itself as proof of the FIX. bugs_open/440 closes on an
-- INDUCED unknown routing key landing in needs_human_review — a census showing zero
-- refusals is equally consistent with "nothing bad has been written yet" and "the
-- refusal branch is unreachable". The induction recipe is in the lane RUNBOOK.

\echo
\echo ==== A. did the flip land ====
SELECT default_config #>> '{workflow,start_step}' AS start_step,
       (default_config #> '{workflow,steps,check_routing_key_known}') IS NOT NULL AS guard_step,
       (default_config #> '{workflow,steps,refuse_unknown_routing_key}') IS NOT NULL AS refusal_step,
       default_config #>> '{workflow,steps,refuse_unknown_routing_key,config,status_override}' AS parks_at,
       (length(c.cond) - length(replace(c.cond, 'input_data.spec.reason ==', '')))
         / length('input_data.spec.reason ==') AS reason_tests,
       (length(c.cond) - length(replace(c.cond, 'input_data.spec.routing_reason ==', '')))
         / length('input_data.spec.routing_reason ==') AS routing_tests
  FROM agent_definitions a
  CROSS JOIN LATERAL (SELECT a.default_config #>> '{workflow,steps,check_rerender_mode,config,condition}' AS cond) c
 WHERE a.type='page-rerender' AND a.is_active
   AND COALESCE(a.is_snapshot,false)=false AND a.deleted_at IS NULL;
-- EXPECT: check_routing_key_known | t | t | needs_human_review | 5 | 5

\echo
\echo ==== B. is the CHECK safe to VALIDATE (D3) ====
-- The constraint went on NOT VALID, so existing rows were never scanned. VALIDATE
-- takes only SHARE UPDATE EXCLUSIVE (concurrent reads and writes continue) but it
-- FAILS if any row violates — and on this table a failed validate is a wasted scan,
-- so count first. ⚠ Count over ALL statuses, not just pending: VALIDATE reads every
-- row in the table, including the terminal ones a work-item census would filter out.
SELECT count(*) AS rows_that_would_fail_validate
  FROM site_work_items
 WHERE item_type='page_rerender'
   AND spec->>'routing_reason' IS NOT NULL
   AND spec->>'routing_reason' NOT IN
       ('image_landed', 'section_data_resolved', 'cta_links_stale', 'template_changed', 'literal_markdown');
-- ZERO -> run:  ALTER TABLE site_work_items VALIDATE CONSTRAINT chk_page_rerender_routing_reason_vocabulary;
-- NON-ZERO -> do NOT validate. Those rows are the pre-741 damage; list them with the
-- query below and decide per row whether the key belongs in the vocabulary or in
-- spec.reason. Validating is the LAST step, not the first.

SELECT spec->>'routing_reason' AS bad_key, count(*), min(created_at)::date AS first_seen, max(created_at)::date AS last_seen
  FROM site_work_items
 WHERE item_type='page_rerender'
   AND spec->>'routing_reason' IS NOT NULL
   AND spec->>'routing_reason' NOT IN
       ('image_landed', 'section_data_resolved', 'cta_links_stale', 'template_changed', 'literal_markdown')
 GROUP BY 1 ORDER BY 2 DESC;

\echo
\echo ==== B2. constraint state ====
SELECT conname, convalidated, pg_get_constraintdef(oid) AS def
  FROM pg_constraint
 WHERE conname='chk_page_rerender_routing_reason_vocabulary'
   AND conrelid='site_work_items'::regclass;
-- EXPECT immediately after 741: convalidated = f

\echo
\echo ==== C. the drain, and what the refusal is actually catching ====
-- The transition clause is load-bearing until this collapses. Narrowing the gate to
-- routing_reason alone is safe only when reason_only reaches 0.
SELECT count(*) FILTER (WHERE spec->>'routing_reason' IS NOT NULL
                          AND spec->>'reason' IS NOT NULL)                       AS both_keys,
       count(*) FILTER (WHERE spec->>'routing_reason' IS NULL
                          AND spec->>'reason' IS NOT NULL)                       AS reason_only,
       count(*) FILTER (WHERE spec->>'routing_reason' IS NOT NULL
                          AND spec->>'reason' IS NULL)                           AS routing_only,
       -- ⚠ THE STATE THE TRANSITION CLAUSE CANNOT HANDLE WELL: the two keys DISAGREE.
       -- The gate routes on either, but rerender_sections is still handed
       -- input_data.spec.REASON, so the single-value readers
       -- (shouldStripLiteralMarkdown, the CTA recompute) would see the annotation
       -- instead of the routing key and silently under-deliver. MEASURED 0 on
       -- 2026-09-03; producers stamp in lockstep, and nothing ENFORCES that.
       count(*) FILTER (WHERE spec->>'routing_reason' IS NOT NULL
                          AND spec->>'reason' IS NOT NULL
                          AND spec->>'routing_reason' <> spec->>'reason')        AS keys_disagree
  FROM site_work_items
 WHERE item_type='page_rerender' AND status NOT IN ('complete','cancelled','rejected');

SELECT spec->>'routing_reason' AS refused_key, count(*), max(updated_at) AS latest
  FROM site_work_items
 WHERE item_type='page_rerender' AND status='needs_human_review'
   AND error LIKE '%not in the sections-rerender vocabulary%'
 GROUP BY 1 ORDER BY 2 DESC;
-- A NULL or empty refused_key here is the ALARM: it means the guard clause lost its
-- `== null` / `== ''` disjunct and the legacy population is being parked. Roll back.
