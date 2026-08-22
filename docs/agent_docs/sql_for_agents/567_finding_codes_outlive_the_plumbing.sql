-- 567 — a deliberate FINDING outlives the error plumbing it shares a table with.
--
-- Owner ruling 2026-08-22, on `bugs_open/358`: keep deliberate findings ~1 year, leave
-- ordinary failure plumbing at 30 days, and stop `resolved` shortening a row's life.
--
-- ── THE DEFECT ──────────────────────────────────────────────────────────────
-- `database-cleanup` arm 1 applies ONE rule to `agent_error_log`:
--
--     WHERE (resolved = true  AND occurred_at < NOW() - INTERVAL '14 days')
--        OR (resolved = false AND occurred_at < NOW() - INTERVAL '30 days')
--
-- That table holds two different kinds of row. A TIMEOUT is worth most on the day it
-- happens. A deliberate FINDING — a detector's record of something it noticed and chose
-- not to fix — is often worth most months later, when the question is whether a pattern
-- is old or new. One clock cannot serve both, and the shorter one wins today.
--
-- This is not hypothetical and it is not a slow leak. Measured 2026-08-22, WHILE the
-- lane was writing the document describing it: `REVIEW_SUPERSEDED_BY_PASSING_SAVE`, 25
-- rows, the oldest in the table, was deleted between two runs of the registry check three
-- hours apart. Its entire output, erased unread. `TRUNCATION_DEGRADED_REVIEW` was three
-- days from the same fate and was extracted by hand (commit 8d15727c8) — by hand is not
-- a mechanism.
--
-- ── WHAT IT COSTS, WHICH IS THE ARGUMENT ────────────────────────────────────
-- Measured 2026-08-22 over the retained 30-day window, 45,553 rows total:
--
--     operational plumbing (14 codes)        33,832
--     instrumented (the 2 RFC_029 resolvers) 10,103
--     consumed                                  360
--     the 32 undecided — 358's whole subject   1,247
--
-- So the material this migration protects is UNDER 3% of the table. Keeping every one of
-- those rows for a full year is ~15,000 rows. That is the entire price, and it is why the
-- one-rule-for-everything clock was never a considered trade — just a rule that predated
-- the distinction.
--
-- ── WHY THE DISCRIMINATOR IS THE CODE, AND NOT `severity` ───────────────────
-- `severity` is the obvious candidate and it was MEASURED AND REJECTED. Findings are
-- written as error, warning AND info; plumbing as error, fatal AND warning; and three
-- codes emit mixed severities (`CONTENT_CLAIMS_FLOOR_DETAIL`, `CONTENT_DATA_ENVELOPE`,
-- `FIX_PLAN_VALIDATION_REFUSED`). Nothing else in the row separates the two kinds.
-- Recorded because it is the idea the next reader will have too.
--
-- ── THE SAFETY PROPERTY: THE DEFAULT IS *KEEP* ──────────────────────────────
-- The list below names what expires EARLY; everything else lives 365 days. This is the
-- whole design. A code that is new, misspelled, or that nobody remembered to add is
-- RETAINED — so drift in this list can only ever over-retain, never delete something
-- unread. That is the opposite failure direction from the rule it replaces, and it is why
-- a hand-written list is acceptable here in a lane that otherwise retired two of them.
-- `error_code` is NOT NULL in practice (0 NULL, 0 empty of 45,553 rows, measured
-- 2026-08-22); a NULL would fall to the keep side anyway, which is the safe side.
--
-- Normalised with `split_part(error_code, ':', 1)`, matching the registry's own key rule —
-- `error_code` is free text and one writer emits colon-suffixed variants
-- (`tool_crosslink_not_emitted:no_related_pages`). Without it a family would dodge its
-- own retention class.
--
-- ── WHY THE TWO RESOLVER CODES EXPIRE WITH THE PLUMBING ─────────────────────
-- `RESOLVER_CONFLICTING_CANDIDATES` and `RESOLVER_MAPPING_BYPASSED` are `instrumented`
-- under RFC_029, with an owner, six dated reads and a ruling of 2026-08-18. They are also
-- 10,103 of the 45,553 rows — ~1,600/day — so a 365-day clock would cost ~585,000 rows on
-- their own account. Their design says frequency is the evidence, not history, and they
-- have 30 days today. Keeping them at 30 days is therefore NO CHANGE for that lane, and
-- the alternative buys them nothing they asked for. Told, not merely measured: this file
-- is named in that lane's registry entries.
--
-- ── WHY `resolved` STOPS SHORTENING A ROW'S LIFE ────────────────────────────
-- The 14-day arm is DELETED outright rather than re-tuned. A resolved row is the finding
-- PLUS its outcome — the most useful version, and the first thing a reader of a pattern
-- wants. Halving its life inverts that. It is free today (48 resolved rows in the table's
-- entire history, all stamped 2026-08-22 by `cmd/content-loss-check`, the first use of
-- that column ever) and it stops being free the moment `358` B1 builds a triage flow,
-- which is the next task in that lane. `resolved` keeps every other meaning it has.
--
-- ── ORDER-INDEPENDENT WITH 566, DELIBERATELY ────────────────────────────────
-- `566_database_cleanup_reaps_every_terminal_status.sql` edits ARM 3 of this same
-- `pre_query` and pins the whole text by md5, before and after. Two migrations editing
-- different arms of one row would normally mean whoever lands second is refused and has
-- to re-derive. So this one accepts EITHER known text — the pre-566 md5 or the post-566
-- md5 — and refuses anything else. That keeps the property 566's header rightly argues
-- for (the input is known EXACTLY, so `replace()` is deterministic) while removing the
-- ordering constraint entirely. It touches nothing 566 touches: different arm, different
-- table, disjoint anchors, and the negative controls below prove arm 3 survives intact
-- whichever form it is in.
--
-- Rollback sidecar: 567_finding_codes_outlive_the_plumbing_ROLLBACK.sql

BEGIN;

-- ── 1. refuse unless the live text is one this migration was written against ──
DO $do$
DECLARE q text; n int; anchor text;
BEGIN
  anchor := 'WHERE (resolved = true AND occurred_at < NOW() - INTERVAL ''14 days'')' || chr(10) ||
            '           OR (resolved = false AND occurred_at < NOW() - INTERVAL ''30 days'')';

  SELECT pre_query INTO q FROM scheduled_tasks WHERE name = 'database-cleanup';
  IF q IS NULL THEN
    RAISE EXCEPTION '567 REFUSED: no scheduled_tasks row named database-cleanup';
  END IF;

  -- Either known text. c26ccf49 = pre-566; b4deb963 = post-566. See the header.
  IF md5(q) NOT IN ('c26ccf49e38f5df181006c1d132f19e4',
                    'b4deb9638404caccb4eecda9795f4eef') THEN
    RAISE EXCEPTION '567 REFUSED: database-cleanup pre_query is neither text this migration knows (md5 %). A third change landed — re-read the LIVE row, re-derive the edit, and add its md5 here only after checking arm 1 is still the text below.', md5(q);
  END IF;

  -- The anchor must be unambiguous: a blind replace() over two occurrences would
  -- rewrite a second arm this migration has no opinion about.
  n := (length(q) - length(replace(q, anchor, ''))) / length(anchor);
  IF n <> 1 THEN
    RAISE EXCEPTION '567 REFUSED: expected exactly 1 occurrence of arm 1''s retention predicate, found %', n;
  END IF;

  RAISE NOTICE '567: live text is a known form (md5 %), anchor is unique — applying.', md5(q);
END
$do$;

-- ── 2. the edit ─────────────────────────────────────────────────────────────
UPDATE scheduled_tasks
   SET pre_query = replace(pre_query,
         'WHERE (resolved = true AND occurred_at < NOW() - INTERVAL ''14 days'')' || chr(10) ||
         '           OR (resolved = false AND occurred_at < NOW() - INTERVAL ''30 days'')',
         'WHERE (occurred_at < NOW() - INTERVAL ''30 days''' || chr(10) ||
         '                AND split_part(error_code, '':'', 1) = ANY (ARRAY[' || chr(10) ||
         '                      ''UNKNOWN'', ''PROCESSING_FAILED'', ''LLM_API_ERROR'', ''TIMEOUT'',' || chr(10) ||
         '                      ''CHILD_ORCHESTRATION_FAILED'', ''PARSE_ERROR'', ''CONTENT_VALIDATION_FAILED'',' || chr(10) ||
         '                      ''TEMPLATE_FIELD_ERROR'', ''CONNECTION_ERROR'', ''MISSING_FIX_TYPE'',' || chr(10) ||
         '                      ''DISPATCH_UNRESOLVABLE'', ''INCOMING_MESSAGE_REJECTED'', ''UNROUTED_IMAGE_KIND'',' || chr(10) ||
         '                      ''DISCOVERY_CHECK_ERROR'',' || chr(10) ||
         '                      ''RESOLVER_CONFLICTING_CANDIDATES'', ''RESOLVER_MAPPING_BYPASSED''' || chr(10) ||
         '                    ]))' || chr(10) ||
         '           OR occurred_at < NOW() - INTERVAL ''365 days'''),
       updated_at = now()
 WHERE name = 'database-cleanup';

-- ── 3. verify (DO/RAISE — ON_ERROR_STOP ignores a non-empty SELECT, so a SELECT
--        cannot stop a COMMIT; only a raised exception can) ───────────────────
DO $do$
DECLARE q text; n int; c text;
BEGIN
  SELECT pre_query INTO q FROM scheduled_tasks WHERE name = 'database-cleanup';

  -- 3a. the old rule is GONE, both halves of it
  IF q LIKE '%resolved = true AND occurred_at%' OR q LIKE '%resolved = false AND occurred_at%' THEN
    RAISE EXCEPTION '567: arm 1 still carries the resolved-shortens-life rule';
  END IF;
  IF q LIKE '%INTERVAL ''14 days''%' THEN
    RAISE EXCEPTION '567: the 14-day clock survives somewhere in the sweep';
  END IF;

  -- 3b. the new rule is present and complete
  IF q NOT LIKE '%INTERVAL ''365 days''%' THEN
    RAISE EXCEPTION '567: the 365-day retention arm is missing';
  END IF;
  IF q NOT LIKE '%split_part(error_code%' THEN
    RAISE EXCEPTION '567: arm 1 does not normalise error_code — a colon family would dodge its class';
  END IF;
  n := (length(q) - length(replace(q, 'RESOLVER_CONFLICTING_CANDIDATES', '')))
       / length('RESOLVER_CONFLICTING_CANDIDATES');
  IF n <> 1 THEN
    RAISE EXCEPTION '567: expected the short-retention list exactly once, found %', n;
  END IF;

  -- 3c. NEGATIVE CONTROLS: every pre-existing arm must survive. Without these a
  --     replace() that ate more than intended would still pass 3a/3b. Arm 3 is
  --     asserted in BOTH its forms because 566 may or may not have landed.
  IF NOT (q LIKE '%deleted_errors AS (%' AND q LIKE '%deleted_audit AS (%'
      AND q LIKE '%deleted_orchestrations AS (%' AND q LIKE '%deleted_stale AS (%'
      AND q LIKE '%deleted_orphan_palettes AS (%' AND q LIKE '%deleted_orphan_typography AS (%') THEN
    RAISE EXCEPTION '567: a pre-existing database-cleanup arm was lost';
  END IF;
  IF NOT (q LIKE '%''COMPLETED'', ''FAILED''%'
       OR q LIKE '%SELECT status FROM orchestration_status_vocabulary WHERE is_terminal)%') THEN
    RAISE EXCEPTION '567: arm 3 is in neither its pre-566 nor its post-566 form — it was damaged';
  END IF;
  IF q NOT LIKE '%DELETE FROM agent_error_log%' THEN
    RAISE EXCEPTION '567: arm 1 no longer deletes from agent_error_log';
  END IF;

  -- 3d. it still PARSES AND EXECUTES. A syntactically valid query that errors at
  --     runtime would stop the whole hourly sweep silently. The RAISE rolls the
  --     block's savepoint back, so the sweep's own deletions are undone here.
  BEGIN
    EXECUTE q;
    RAISE EXCEPTION 'PARSE_CHECK_OK';
  EXCEPTION WHEN OTHERS THEN
    IF SQLERRM <> 'PARSE_CHECK_OK' THEN
      RAISE EXCEPTION '567: database-cleanup pre_query does NOT execute (%)', SQLERRM;
    END IF;
  END;

  -- 3e. the SHIPPED list is what is under test — not a copy of it, and not a
  --     population that cannot answer. The first draft of this block failed BOTH
  --     ways and a mutation test caught it: it compared against its own hard-coded
  --     ARRAY (so editing the list that actually ships changed nothing) and it
  --     filtered on three codes with NO rows older than 30 days (so the EXISTS was
  --     false whatever the list said). It passed a mutant that put
  --     CONTENT_LINK_REPAIR_DETAIL straight into the shipped list. Both halves are
  --     fixed below: every assertion reads `q`, and each one can fail.
  --     Safe as an exact test because none of these names occurs anywhere else in
  --     the sweep (measured 2026-08-22: 0 quoted occurrences for all 16).

  -- 3e-i. no code the registry classes as a FINDING may appear in the short list.
  --       Sampled across every non-operational disposition: consumed, unruled, and
  --       the two loudest of each.
  FOREACH c IN ARRAY ARRAY[
        'CONTENT_LINK_REPAIR_DETAIL','TRUNCATION_DEGRADED_REVIEW','RETRACTION_AUDIT',
        'VALIDATION_ERROR_DROPPED','PLAN_SECTION_NAME_DROPPED','ARCHIVED_PAGE_DEPLOY_REFUSED',
        'component_validation_rejected','CONTENT_CLAIMS_FLOOR_DETAIL','CONTENT_DATA_ENVELOPE',
        'STRUCTURAL_KEY_CARRY_MISS','CONTENT_KEY_LOSS','CONTENT_DATA_REGRESSION',
        'CONTENT_VALIDATION_BLOCKER_DETAIL','DEPLOY_STAMP_REFUSED_ON_SKIP']
  LOOP
    IF q LIKE '%''' || c || '''%' THEN
      RAISE EXCEPTION '567: the short-retention list names %, which the registry classes as a FINDING — a finding must fall through to the 365-day default', c;
    END IF;
  END LOOP;

  -- 3e-ii. and every code that SHOULD expire early is actually there. Omitting one
  --        is the opposite error and is silent: that code would quietly gain a
  --        365-day clock and the table would grow without anyone deciding it.
  FOREACH c IN ARRAY ARRAY[
        'UNKNOWN','PROCESSING_FAILED','LLM_API_ERROR','TIMEOUT','CHILD_ORCHESTRATION_FAILED',
        'PARSE_ERROR','CONTENT_VALIDATION_FAILED','TEMPLATE_FIELD_ERROR','CONNECTION_ERROR',
        'MISSING_FIX_TYPE','DISPATCH_UNRESOLVABLE','INCOMING_MESSAGE_REJECTED',
        'UNROUTED_IMAGE_KIND','DISCOVERY_CHECK_ERROR',
        'RESOLVER_CONFLICTING_CANDIDATES','RESOLVER_MAPPING_BYPASSED']
  LOOP
    IF q NOT LIKE '%''' || c || '''%' THEN
      RAISE EXCEPTION '567: the short-retention list is MISSING % — that code would silently gain a 365-day clock', c;
    END IF;
  END LOOP;

  RAISE NOTICE '567: applied. Findings now live 365 days; plumbing and the RFC_029 resolvers stay at 30; resolved no longer shortens a row.';
END
$do$;

COMMIT;
