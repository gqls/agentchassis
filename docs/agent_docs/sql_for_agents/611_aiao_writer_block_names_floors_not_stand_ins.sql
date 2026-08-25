-- 611_aiao_writer_block_names_floors_not_stand_ins.sql
--
-- ai-agent-orchestration.com. Removes the "NNN+" stand-in tokens from the site's
-- writer_block — one of which is being served to the public RIGHT NOW on
-- model-directory.html ("…against the NNN+ agent types already running in
-- production", curled 2026-08-25, regenerated the same morning). bugs_open/387.
--
-- WHAT 557 GOT RIGHT AND WHERE IT BROKE. 557 correctly removed hardcoded live
-- counts (175/174/14 had gone stale within days). But it replaced them with a
-- QUOTED EXEMPLAR containing a stand-in — 'Phrase it as "NNN+ AI agents"' — and an
-- instruction to "take the live value from the fact". Three measurements, all
-- taken 2026-08-25 before this file was written (queries:
-- docs024_key_docs_latest/bugfix_387_deployed_and_404/RUNBOOK_387.md):
--
--   1. THE PROMPT NEVER CONTAINS THE VALUE. On the unscoped path the writer's
--      rendered prompt carries ONLY writer_block — the hero call that shipped the
--      live defect (llm_call_log 9ba94176…) held ZERO occurrences of the fact's
--      value. The instruction could not be followed, whatever it said.
--   2. THE MODEL COPIES THE EXEMPLAR. Since 557 applied: 137 instructed writer
--      calls -> 14 copied "NNN" verbatim into copy (~10%), 0 wrote the agents
--      value. Zero \mNNN\M in ANY writer response before 08-22.
--   3. THE WORKING PATTERN ALREADY EXISTS, LIVE, ON FIVE FACTS FLEET-WIDE
--      (measured by the bugs_open/386 lane, same day): a ROUNDED LITERAL FLOOR in
--      the prose slot — "more than 10 live production sites", "more than 2,000
--      business records" — which stays true as the value rises, so the same
--      verbatim copying that shipped "NNN+" produces correct copy.
--
-- So: floors in the prose slot, no stand-in tokens anywhere, and the ban is
-- DESCRIBED rather than quoted — a brief that quotes a phrase hands it to the
-- writer (cmd/brief-negation-check's measured transfer chain), and a quoted
-- "forbidden" token is still a quoted token.
--
-- DOES NOT CHANGE ANY PUBLISHED PAGE. evidence_base is consumed at write time;
-- the next build of each page reads the new block. model-directory rebuilds
-- ~6-hourly via the "model data refreshed" needs_page.
--
-- FACTS, VALUES AND banned_claims ARE NOT TOUCHED — values are the refresh
-- action's to own (same rule 557 enforced on itself). The floor WORDINGS below
-- are lifted from the site's own writer_lines; the aiao lane owns them
-- (CONTRIB_2026-08-25_from_387… in their dir names this file and the wording as
-- theirs to change).
--
-- THE DURABLE FIX IS NOT THIS FILE: composeWriterBlock's {value} substitution
-- under writer_block_managed:true retires hand-typed numbers entirely; blocked
-- for this site because managed regeneration drops the NEVER-STATE list —
-- proposed to the bugs_open/288 lane (refresh_evidence_base_action.go owner).
--
-- SUPERSEDE, DO NOT MUTATE — same shape as 557/458 (is_current/superseded_at +
-- partial unique index). Council scope: this migration IS the running system.
-- ROLLBACK: 611_..._ROLLBACK.sql (restores the exact 557-era row from migration_backups).

BEGIN;

-- 0. Byte-exact backup of the row being superseded.
INSERT INTO migration_backups (migration_name, target_table, target_id, old_value, notes)
SELECT '611_aiao_writer_block_names_floors_not_stand_ins',
       'site_specs', ss.id::text, jsonb_build_object('data', ss.data),
       'pre-611 evidence_base for ai-agent-orchestration.com (557-era writer_block with NNN+ exemplars)'
FROM site_specs ss
WHERE ss.site_id='2a8ebf9c-20a2-4c39-b191-840b012371da'
  AND ss.aspect='evidence_base' AND ss.is_current;

-- 1. Close the current row.
UPDATE site_specs
SET is_current = false, superseded_at = now()
WHERE site_id='2a8ebf9c-20a2-4c39-b191-840b012371da'
  AND aspect='evidence_base' AND is_current;

-- 2. Insert the successor: new writer_block; facts and banned_claims byte-identical.
INSERT INTO site_specs (site_id, aspect, data, source, source_agent, notes, is_current, created_by)
SELECT
  ss.site_id,
  ss.aspect,
  jsonb_set(
    ss.data,
    '{writer_block}',
    to_jsonb(
      'NUMBERS — the facts list in this document is the ONLY authority for figures about this ' ||
      'business. Never copy a number out of a page, a template, or an older spec. State counts ' ||
      'as the ROUNDED FLOORS written below — a floor stays true as the real value rises, so it ' ||
      'cannot go stale between writing and publication. NEVER write letters as stand-ins for ' ||
      'digits (a letter repeated where a number belongs): a stand-in is a defect that ships, not ' ||
      'a placeholder — one reached the public on this site on 2026-08-25. If a count is not ' ||
      'written in this block, write no count at all.' || E'\n' ||
      '- active agent definitions in the production registry: write "more than 150 active agent ' ||
      'definitions", "150+ AI agents" or "more than 150 agents in the registry" — those wordings ' ||
      'are the ones the evidence checker recognises for this fact. NEVER write "over 70 ' ||
      'specialised AI agents" or "70+ agents": true as a lower bound but understating the fleet ' ||
      'by more than half (owner ruling 2026-07-27).' || E'\n' ||
      '- distinct agent types: write "more than 150 distinct agent types" or "150+ agent types". ' ||
      'NEVER write "30+ agent types" — understates by more than five times (owner ruling ' ||
      '2026-07-27).' || E'\n' ||
      '- 8 departments — Strategy, Research, Content, Design, Development, Quality, Operations, ' ||
      'Data. This is the platform''s OWN organisational taxonomy (a self-description); never ' ||
      'frame it as departments of external clients ("departments served").' || E'\n' ||
      '- live sites in production, built and operated end-to-end by the platform: write "more ' ||
      'than 20 live sites". (History kept from 557: a literal "14" stood in this block from ' ||
      '2026-07-26 until 2026-08-22 while the real figure reached 25 — hence floors, never live ' ||
      'counts, in this block.)' || E'\n' ||
      '- 17 backend services (the service manifests under deployments/kustomize/services; the ' ||
      'directory also holds frontend and infra overlays, so a bare directory count is NOT this ' ||
      'figure).' || E'\n' ||
      '- orchestrations per day: "over a thousand orchestrations a day" is true and may be ' ||
      'stated. DO NOT state an exact daily figure — it is a ROLLING 24-hour window, so any ' ||
      'number you write is stale within hours and will be rejected as unsupported.' || E'\n' ||
      '- automated work items completed: DO NOT state a figure. The ledger is reaped, so this ' ||
      'count FALLS as well as rises (1,051 on 2026-07-26; 6,918 on 2026-08-22); a total that ' ||
      'can go down is misleading however it is phrased.' || E'\n' ||
      '- Architecture: Kubernetes, Kafka, Postgres — true and stated freely.' || E'\n' ||
      'NOT TRACKED / DOES NOT EXIST, NEVER STATE: clients served, "departments served", ' ||
      'satisfaction rates, awards won, concurrent-instance counts ("thousands of concurrent ' ||
      'instances" is not measured), uptime percentages. None of these are measured; every such ' ||
      'figure at any value is an invention.'
    )
  ),
  ss.source, ss.source_agent,
  'superseded by migration 611: stand-in tokens (NNN+) removed from writer_block after one shipped '
    || 'to the public (bugs_open/387); counts now stated as the rounded floors the writer_lines carry',
  true,
  '611_migration'
FROM site_specs ss
WHERE ss.site_id='2a8ebf9c-20a2-4c39-b191-840b012371da'
  AND ss.aspect='evidence_base' AND ss.superseded_at IS NOT NULL
ORDER BY ss.superseded_at DESC
LIMIT 1;

-- 3. Guard the exact post-condition. DO/RAISE, not bare SELECTs — a verify block
--    of SELECTs cannot stop the COMMIT (RFC_006 lesson).
DO $$
DECLARE
  cur         jsonb;
  wb          text;
  old_wb      text;
  n_current   int;
  n_facts_old int;
  n_facts_new int;
  backed_up   int;
BEGIN
  SELECT count(*) INTO backed_up FROM migration_backups
   WHERE migration_name='611_aiao_writer_block_names_floors_not_stand_ins';
  IF backed_up <> 1 THEN
    RAISE EXCEPTION '611: expected 1 backup row, wrote % — refusing without a restore path', backed_up;
  END IF;

  SELECT count(*) INTO n_current FROM site_specs
   WHERE site_id='2a8ebf9c-20a2-4c39-b191-840b012371da' AND aspect='evidence_base' AND is_current;
  IF n_current <> 1 THEN
    RAISE EXCEPTION '611: expected exactly 1 current evidence_base row, found %', n_current;
  END IF;

  SELECT data INTO cur FROM site_specs
   WHERE site_id='2a8ebf9c-20a2-4c39-b191-840b012371da' AND aspect='evidence_base' AND is_current;
  wb := cur->>'writer_block';
  SELECT old_value->'data'->>'writer_block' INTO old_wb FROM migration_backups
   WHERE migration_name='611_aiao_writer_block_names_floors_not_stand_ins';

  -- THE POINT OF THE FILE: no stand-in token of any shape may remain.
  IF wb ~ '\mN{2,}\M' OR wb ~ '\mN{2,}\+' OR wb ~ '\mX{2,}\M' OR position('NNN' in wb) > 0 THEN
    RAISE EXCEPTION '611: writer_block still carries a stand-in token';
  END IF;
  -- And no stale live count may come back.
  IF wb ~ '\m(175|174)\M' THEN
    RAISE EXCEPTION '611: writer_block reintroduced a stale hardcoded registry count';
  END IF;
  IF wb = old_wb THEN
    RAISE EXCEPTION '611: writer_block unchanged — the supersede inserted the old text';
  END IF;

  -- Every protective clause carried forward (the hazard that keeps
  -- writer_block_managed off here is losing one of these silently).
  IF wb !~ 'NEVER write "over 70' THEN
    RAISE EXCEPTION '611: the 70+ agents ban was lost from writer_block';
  END IF;
  IF wb !~ 'NEVER write "30\+ agent types"' THEN
    RAISE EXCEPTION '611: the 30+ agent types ban was lost from writer_block';
  END IF;
  IF wb !~ 'NOT TRACKED / DOES NOT EXIST, NEVER STATE' THEN
    RAISE EXCEPTION '611: the NEVER-STATE list was lost from writer_block';
  END IF;
  IF wb !~ 'ROLLING 24-hour window' THEN
    RAISE EXCEPTION '611: the rolling-window rule was lost from writer_block';
  END IF;
  IF wb !~ 'FALLS as well as rises' THEN
    RAISE EXCEPTION '611: the reaped-ledger rule was lost from writer_block';
  END IF;
  -- And the floors the block now promises.
  IF wb !~ 'more than 150 active agent definitions' OR wb !~ 'more than 150 distinct agent types'
     OR wb !~ 'more than 20 live sites' THEN
    RAISE EXCEPTION '611: a promised floor wording is missing from writer_block';
  END IF;

  -- Facts and values are the refresh action's to own, not this file's.
  SELECT jsonb_array_length(old_value->'data'->'facts') INTO n_facts_old
    FROM migration_backups
   WHERE migration_name='611_aiao_writer_block_names_floors_not_stand_ins';
  n_facts_new := jsonb_array_length(cur->'facts');
  IF n_facts_new <> n_facts_old THEN
    RAISE EXCEPTION '611: fact count changed % -> %', n_facts_old, n_facts_new;
  END IF;
  IF (cur->'facts')::text IS DISTINCT FROM
     (SELECT (old_value->'data'->'facts')::text FROM migration_backups
       WHERE migration_name='611_aiao_writer_block_names_floors_not_stand_ins') THEN
    RAISE EXCEPTION '611: facts changed; this file may only touch writer_block';
  END IF;
  IF (cur->'banned_claims')::text IS DISTINCT FROM
     (SELECT (old_value->'data'->'banned_claims')::text FROM migration_backups
       WHERE migration_name='611_aiao_writer_block_names_floors_not_stand_ins') THEN
    RAISE EXCEPTION '611: banned_claims changed; this file may only touch writer_block';
  END IF;

  RAISE NOTICE '611 OK: writer_block carries rounded floors and no stand-in tokens; every ban kept; facts, values and banned_claims byte-identical. Nothing re-renders — the next page build reads it.';
END $$;

COMMIT;
