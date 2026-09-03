-- 742_evidence_register_absence_check.sql
--
-- Council-Submitted: 0d730d51-a923-4b44-a58f-ab8c898d7e22
-- Creates the scheduled task `evidence-register-absence`: a DAILY runtime check that
-- files one `missing_evidence_register` work item for every `deployed` site holding no
-- current `evidence_base` spec.
--
-- WHY, AND WHY IT CANNOT BE DONE BY WIDENING THE EXISTING SWEEP.
-- `resolveEvidenceSites` (`refresh_evidence_base_action.go:281`, fleet query at :291)
-- builds the daily `evidence-freshness` target list as
--     SELECT site_id FROM site_specs WHERE aspect='evidence_base' AND is_current
-- It selects the sites that HAVE a register. **The target set is defined by the presence
-- of the very thing whose absence is the defect**, so a register-less site is invisible to
-- the freshness sweep, to the fact checks and to the `invalid_banned_claim_pattern`
-- detector alike — permanently, and running any of them more often would never reach one.
-- No `item_type` matching %evidence%/%register%/%claim% has ever been filed for an absence
-- (measured 2026-09-03: only `claims_unverified` 47, `stale_evidence` 10,
-- `spec_supplies_claim` 2, `stale_directory_claim` 2 — every one presupposing a register
-- exists). RFC_060 Q1 requires registers and nothing enforced or even REPORTED the
-- requirement. `loancash.co.uk` was found only because a person went looking.
--
-- AUTHORITY: OWNER RULING 2026-09-03, RFC_060 §3g **D3** — "build the missing check and
-- fill the missing data for the sites", both halves — and **D4**, "a register for each site
-- to avoid AI slop", which is why the check is scoped to EVERY deployed site rather than to
-- finance-tiered ones. §3g(i) sets the two bars the filed item asks the reviewer to choose
-- between.
--
-- WHY A CTE-ONLY SCHEDULED TASK AND NOT GO. Three reasons, in order:
--   1. **Reuse before building** — the mechanism already exists and is documented in
--      `cmd/scheduler/main.go:274`: a task with `fire_message=false` is a "CTE-only task"
--      where the pre_query IS the worker, no Kafka message is sent, and the row the query
--      returns is logged as the tick's report. Ten tasks already run this way.
--   2. **DB config is live immediately; Go is inert until an image is rebuilt and rolled.**
--      A detector that waits for a roll is a detector that is off for an unknown period, and
--      this estate has a documented history of exactly that (a check sitting off for 9 days
--      after its blocker cleared).
--   3. RFC_006's ruled shape for this class: a CI-time check STRUCTURALLY cannot gate live
--      config, so the remedy is a **daily runtime** one. Absence of a `site_specs` row is
--      live config.
--
-- THE REPORT IS DELIBERATELY THREE NUMBERS, NOT A ROW COUNT. The pre_query ALWAYS returns
-- exactly one row carrying `missing_total`, `filed_new` and `already_open`. It would have
-- been shorter to return no rows when there is nothing to file — the scheduler treats that
-- as a successful no-op and logs "found no rows". But that collapses two different states
-- into one line: "no site is missing a register" (healthy) and "sites are missing one but
-- an item is already open for each" (working as intended, nothing new). This lane has spent
-- the day on precisely that failure mode — `omitempty` erasing the clean case so that "ran
-- and found nothing" and "never ran" are indistinguishable (RFC_060 §3f). Three numbers,
-- always, on every tick: a MISSING log line means the task did not run, and
-- `missing_total=0` is a positive statement of health. `bugs_open/083` is why the scheduler
-- logs the pre_query result at all.
--
-- WHY `needs_human_review` AND NOT A DISPATCHABLE HANDLER. The item cannot be auto-resolved
-- and should not pretend it can. Two things a machine must not decide here: (a) the site's
-- POSTURE — Q4 ruled the tier is a RECORD carrying who declared it and on what basis, not a
-- flag a sweep infers; and (b) for a `sourced`/`relied_upon` site, whether each figure is
-- true, which means reading the primary source. So the item is a QUEUE ENTRY for a person or
-- a lane, carrying the evidence needed to choose the rung. Filing it with a handler would
-- manufacture the appearance of an automated fix for work that is irreducibly human.
--
-- DEDUP: by `NOT EXISTS` against the same terminal-status list `idx_swi_dedup` uses, rather
-- than `ON CONFLICT`. The index is PARTIAL (unique on (site_id,item_key) WHERE item_key IS
-- NOT NULL AND status <> ALL(...terminal...)), and an `ON CONFLICT` inference has to restate
-- that predicate exactly or it raises 42P10 — the dedup-index/Go-list lockstep this estate
-- has already been bitten by. `NOT EXISTS` expresses the same intent and cannot drift into a
-- runtime error.
--
-- SCOPE GUARDS BUILT IN: only `status='deployed'` (so the 17 `pool`, 3 `test` and 1 `system`
-- sites are never filed against), and only sites with at least one page (a deployed shell
-- serving nothing is not making claims).
--
-- Register entry: docs/agent_docs/docs026_concept_register/register/claims-verification.md — CLM-033.
-- Lane: docs/agent_docs/docs024_key_docs_latest/loancash_couk_fca_validation/
-- Rollback: 742_..._ROLLBACK.sql

BEGIN;

-- GUARD 1: do not create a second copy of this task.
DO $$
DECLARE n int;
BEGIN
  SELECT count(*) INTO n FROM scheduled_tasks WHERE name = 'evidence-register-absence';
  IF n <> 0 THEN
    RAISE EXCEPTION '742 ABORT: scheduled task evidence-register-absence already exists (% row(s)) - read it before writing', n;
  END IF;
END $$;

-- GUARD 2: the terminal-status list below must still match idx_swi_dedup's predicate, or
-- this task's NOT EXISTS and the index disagree about what "already open" means.
DO $$
DECLARE d text;
BEGIN
  SELECT pg_get_indexdef(i.indexrelid) INTO d
    FROM pg_index i JOIN pg_class c ON c.oid = i.indexrelid
   WHERE c.relname = 'idx_swi_dedup';
  IF d IS NULL THEN
    RAISE EXCEPTION '742 ABORT: idx_swi_dedup not found - the dedup contract this task relies on is gone';
  END IF;
  IF d NOT LIKE '%complete%' OR d NOT LIKE '%verified%' OR d NOT LIKE '%rejected%'
     OR d NOT LIKE '%wont_fix%' OR d NOT LIKE '%failed%' OR d NOT LIKE '%unresolved%'
     OR d NOT LIKE '%cancelled%' THEN
    RAISE EXCEPTION '742 ABORT: idx_swi_dedup predicate no longer lists the seven terminal statuses this task hardcodes: %', d;
  END IF;
END $$;

INSERT INTO scheduled_tasks
  (name, description, target_agent_type, target_topic, fire_message,
   interval_seconds, timeout_seconds, enabled, input_data, pre_query)
VALUES (
  'evidence-register-absence',
  'DAILY: files one missing_evidence_register work item per deployed site with no current evidence_base. '
  'CTE-only (fire_message=false) - the pre_query is the worker. Exists because the evidence-freshness sweep '
  'draws its target list FROM the registers, so an absent register is invisible to it for ever. '
  'RFC_060 §3g D3/D4 (owner, 2026-09-03). Report is three numbers on EVERY tick: missing_total, filed_new, '
  'already_open - a missing log line means the task did not run, and missing_total=0 is a positive statement of health.',
  'generic',
  'system.agent.generic.requests',
  false,
  86400,
  300,
  true,
  '{}'::jsonb,
$PQ$
WITH candidates AS (
    SELECT s.id AS site_id, s.domain
      FROM sites s
      LEFT JOIN site_specs eb
             ON eb.site_id = s.id
            AND eb.aspect  = 'evidence_base'
            AND eb.is_current
     WHERE s.status = 'deployed'
       AND eb.id IS NULL
       AND EXISTS (SELECT 1 FROM pages p WHERE p.site_id = s.id)
), fresh AS (
    SELECT c.site_id, c.domain,
           (SELECT count(*) FROM pages p WHERE p.site_id = c.site_id) AS page_count,
           (SELECT string_agg(DISTINCT p.page_type, ', ' ORDER BY p.page_type)
              FROM pages p WHERE p.site_id = c.site_id)                AS page_types
      FROM candidates c
     WHERE NOT EXISTS (
             SELECT 1 FROM site_work_items w
              WHERE w.site_id  = c.site_id
                AND w.item_key = 'missing_evidence_register_' || c.site_id::text
                AND w.status NOT IN ('complete','verified','rejected','wont_fix',
                                     'failed','unresolved','cancelled'))
), ins AS (
    INSERT INTO site_work_items
      (site_id, source, pipeline, item_type, severity, summary, spec,
       priority, handler_agent, status, created_by, item_key, approval_mode)
    SELECT f.site_id,
           'evidence-register-absence',
           'build',
           'missing_evidence_register',
           'medium',
           'No evidence register on deployed site ' || f.domain || ' (' || f.page_count ||
             ' pages) - RFC_060 Q1/D4 require one; choose the posture rung and populate',
           jsonb_build_object(
             'domain',        f.domain,
             'page_count',    f.page_count,
             'page_types',    f.page_types,
             'why',           'This site is status=deployed and has NO current evidence_base spec - an ABSENT '
                              'register, not an empty one. RFC_060 Q1 requires a register on compliance-tiered '
                              'sites; owner ruling D4 (2026-09-03) extends that to EVERY site, at a lower bar for '
                              'ordinary ones. Until a register exists, ScanUnregisteredNumbers is disarmed on this '
                              'site entirely (claims.go: `if eb == nil ... return nil`), so any figure can be '
                              'invented into the copy and nothing notices.',
             'decide_first',  'The POSTURE RUNG, which is a Q4 RECORD carrying who declared it and on what basis - '
                              'not something this sweep may infer. standard = the site''s claims are about its own '
                              'offering. sourced = it asserts EXTERNAL facts, rules or figures as true. '
                              'relied_upon = a reader may act on those assertions to their financial, legal, '
                              'medical or safety detriment. Read what the site actually asserts, at the SERVED '
                              'pages, before choosing. The page_types above are evidence for that decision, not '
                              'the decision.',
             'bar_standard',  'ATTESTED register (RFC_060 §3g(i)): facts carry value + context_terms (+ tolerance) '
                              'and NO citation. This arms ScanUnregisteredNumbers exactly as fully as a cited one - '
                              'numberSupported never reads f.Source - and carries no citation_lost drift risk '
                              'because there is nothing to fetch nightly. Cost: hours. Correct for a site whose '
                              'claims are about itself, where there IS no external authority to cite.',
             'bar_sourced',   'CITED register: everything in the attested bar PLUS '
                              'source.citation{url,quote,title,publisher,accessed} per fact, every quote verified '
                              'through the PRODUCTION matcher (go run ./cmd/fcaquotecheck <url> "<quote>" '
                              '"absent control") so it cannot become daily false drift. Cost: about half a day. '
                              'Method: lendzy_co_uk/RUNBOOK_lendzy_co_uk.md §8, and read §8b, §8c and §8e first.',
             'known_bound',   'ScanUnregisteredNumbers is gated on ProseNumbersAreClaims(), which is FALSE for '
                              'editorialPageTypes (guide, blog-post, tool, game, news-index) and for '
                              'thirdPartyDataComponents. So a register does NOT arm the numeric scan over those '
                              'bodies. That exclusion is measured and deliberate, not a defect to route around - '
                              'see the doctrine block above editorialPageTypes. RFC_060 §3g(ii).',
             'expect_errors', 'Every lane that has run this method has found real errors in its OWN site''s live '
                              'copy: lendzy 2, loanzy 1, loancalculator 2, loancash 3. Run it expecting to find '
                              'some, not to confirm correctness. Record each as corrects_site_citation on the fact; '
                              'do NOT rewrite served copy without the owner - that hold was lifted once, for three '
                              'named findings on one site (D2), and is not general.',
             'acceptance_test','A current site_specs row exists for this site with aspect=evidence_base carrying at '
                              'least one fact with a value, and the posture rung is recorded with who declared it '
                              'and when. Verify by reading the row back through a JOIN on sites - never by the '
                              'site_id you typed (RUNBOOK §8e).',
             'authority',     'OWNER RULING 2026-09-03, RFC_060 §3g D3 and D4.',
             'filed_by',      'scheduled task evidence-register-absence (migration 742)'),
           60,
           '',
           'needs_human_review',
           'evidence-register-absence',
           'missing_evidence_register_' || f.site_id::text,
           'manual'
      FROM fresh f
    RETURNING id
)
SELECT (SELECT count(*) FROM candidates)::text AS missing_total,
       (SELECT count(*) FROM ins)::text        AS filed_new,
       (SELECT (count(*) - (SELECT count(*) FROM fresh))::text FROM candidates) AS already_open
$PQ$
);

-- VERIFY as DO/RAISE. A verify block of bare SELECTs cannot stop the COMMIT.
DO $$
DECLARE n int; fm boolean; iv int; en boolean; pq text;
BEGIN
  SELECT count(*) INTO n FROM scheduled_tasks WHERE name = 'evidence-register-absence';
  IF n <> 1 THEN
    RAISE EXCEPTION '742 VERIFY: expected exactly 1 task row, found %', n;
  END IF;

  SELECT fire_message, interval_seconds, enabled, pre_query
    INTO fm, iv, en, pq
    FROM scheduled_tasks WHERE name = 'evidence-register-absence';

  IF fm <> false THEN
    RAISE EXCEPTION '742 VERIFY: fire_message must be false (CTE-only task) - it is %', fm;
  END IF;
  IF iv <> 86400 THEN
    RAISE EXCEPTION '742 VERIFY: interval_seconds must be 86400 (daily) - it is %', iv;
  END IF;
  IF en <> true THEN
    RAISE EXCEPTION '742 VERIFY: task must be enabled - a detector that ships disabled is a detector that is off';
  END IF;
  IF pq NOT LIKE '%missing_evidence_register%' OR pq NOT LIKE '%missing_total%'
     OR pq NOT LIKE '%filed_new%' OR pq NOT LIKE '%already_open%' THEN
    RAISE EXCEPTION '742 VERIFY: pre_query must file missing_evidence_register and report all three numbers';
  END IF;
  IF pq LIKE '%ON CONFLICT%' THEN
    RAISE EXCEPTION '742 VERIFY: pre_query must dedup by NOT EXISTS, not ON CONFLICT (idx_swi_dedup is partial - 42P10 risk)';
  END IF;

  RAISE NOTICE '742 OK: evidence-register-absence created - daily, CTE-only, enabled, reports missing_total/filed_new/already_open on every tick';
END $$;

COMMIT;
