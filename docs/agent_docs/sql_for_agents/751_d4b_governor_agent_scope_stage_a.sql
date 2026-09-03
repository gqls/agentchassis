-- 751_d4b_governor_agent_scope_stage_a.sql — D4b, STAGE A (DB only, INERT).
--
-- WHY. The D4 spend governor went live 2026-09-03 and was measured the same day: it reaches
-- ~28% of fleet LLM spend, because it sheds WORK ITEMS and can only touch spend descending
-- from a dispatch loop. 69.4% of 24h spend has no dispatch-loop ancestor at all, and
-- council-gate alone is $198.38 — 62% of everything the fleet spends (steady: 69.9% on 09-02,
-- 62.1% on 09-03; 09-01 ran no councils and cost $54.80 all day). So the governor as built
-- cannot defend a budget. OWNER RULING 2026-09-03, verbatim: "extend it, reducing council
-- spend is a fairly easy save if it comes to the crunch."
--
-- WHAT THIS MIGRATION DOES, and what it deliberately does NOT do.
--   DOES: factor the level comparison into ONE function; add an agent-type class map; add
--         `governor_admits_agent(agent_type)`; add the withheld-RUNS record + view.
--   DOES NOT: gate anything. No caller exists until stage B (Go, at the admission point in
--         platform/messaging/processor.go executeWorkflow), which ships opt-in with the unsafe
--         default OFF per the owner's 2026-08-02 seam ruling and takes its own architecture
--         round per AGOV-013's STANDING GATE.
--
-- INERT THREE WAYS, same as D4 stage A: nothing calls governor_admits_agent; the map's only
-- row is seeded at the LAST shed level; and governor_config.enabled already gates every level
-- comparison.
--
-- ⚠ THE ONE RISKY PART, AND HOW IT IS PROVEN. This rewrites the body of `governor_admits`,
-- which is LIVE AND ENFORCING right now (it is called by the dispatch selector, the Go loader
-- and the claim backstop). The signature is unchanged, so no caller text changes and the
-- selector md5 (fcbe8821a2a56512911955735796460e) is untouched. Equivalence is not asserted,
-- it is PROVEN: the verify below keeps a copy of the OLD body as governor_admits_legacy,
-- compares both functions across the FULL cross-product of shed level x class x llm_bearing x
-- unmapped, requires exact agreement on every cell, proves the comparison CAN fail with an
-- induced control, and drops the copy.
--
-- Rollback: 751_..._ROLLBACK.sql (restores the standalone governor_admits body, drops the new
-- objects; refuses while any live agent config or task references governor_admits_agent).

BEGIN;

-- ---------------------------------------------------------------- refusals first
DO $$
BEGIN
  IF EXISTS (SELECT 1 FROM pg_proc WHERE proname='governor_admits_agent') THEN
    RAISE EXCEPTION '751 REFUSED: governor_admits_agent() already exists (replay).';
  END IF;
  IF EXISTS (SELECT 1 FROM pg_proc WHERE proname='governor_admits_class') THEN
    RAISE EXCEPTION '751 REFUSED: governor_admits_class() already exists (replay).';
  END IF;
  IF NOT EXISTS (SELECT 1 FROM pg_proc WHERE proname='governor_admits') THEN
    RAISE EXCEPTION '751 REFUSED: governor_admits() missing — 675 (D4 stage B) not applied.';
  END IF;
  IF NOT EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name='governor_work_class_map') THEN
    RAISE EXCEPTION '751 REFUSED: governor_work_class_map missing — 671 (D4 stage A) not applied.';
  END IF;
  IF EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name='governor_agent_class_map') THEN
    RAISE EXCEPTION '751 REFUSED: governor_agent_class_map already exists (replay).';
  END IF;
  -- The live body must be the 675 text this migration was written against. If someone has
  -- already changed it, the equivalence proof below would prove equivalence to the WRONG thing.
  IF position('LEFT JOIN governor_work_class_map m ON m.item_type = p_item_type'
              in (SELECT pg_get_functiondef(oid) FROM pg_proc WHERE proname='governor_admits')) = 0 THEN
    RAISE EXCEPTION '751 REFUSED: live governor_admits body is not the 675 text — drifted, investigate before rewriting.';
  END IF;
END $$;

-- Keep the OLD body verbatim, purely so the verify can prove equivalence; dropped at the end.
CREATE FUNCTION governor_admits_legacy(p_item_type text) RETURNS boolean
LANGUAGE sql STABLE AS $FN$
  SELECT NOT COALESCE((
    SELECT gc.enabled
       AND COALESCE(m.llm_bearing, true)
       AND gs.shed_level >= CASE COALESCE(m.class, 'maintenance')
             WHEN 'maintenance' THEN 1
             WHEN 'build'       THEN 2
             ELSE                    3
           END
    FROM governor_config gc
    JOIN governor_state gs ON gs.id = 1
    LEFT JOIN governor_work_class_map m ON m.item_type = p_item_type
    WHERE gc.id = 1
  ), false)
$FN$;

-- ---------------------------------------------------------------- 1. THE one comparison
-- The architecture seat's r1 ruling on corr 8f4bb57d was: the shed predicate is ONE canonical
-- thing and must not be re-spelled per consumer. A second namespace (agent types) is exactly
-- the pressure that would re-spell it, so the comparison is factored out FIRST and both public
-- predicates become one-line callers that differ only in which map they read.
CREATE FUNCTION governor_admits_class(p_class text, p_llm_bearing boolean) RETURNS boolean
LANGUAGE sql STABLE AS $FN$
  SELECT NOT COALESCE((
    SELECT gc.enabled
       AND COALESCE(p_llm_bearing, true)
       AND gs.shed_level >= CASE COALESCE(p_class, 'maintenance')
             WHEN 'maintenance' THEN 1
             WHEN 'build'       THEN 2
             ELSE                    3
           END
    FROM governor_config gc
    JOIN governor_state gs ON gs.id = 1
    WHERE gc.id = 1
  ), false)
$FN$;

COMMENT ON FUNCTION governor_admits_class(text, boolean) IS
'D4/D4b spend governor (AGOV-013): THE level comparison, and the only place it is spelled.
governor_admits(item_type) and governor_admits_agent(agent_type) are one-line callers that
differ only in which map they read. Fail-open: missing config/state rows admit everything.
NULL class = maintenance (sheds earliest); NULL llm_bearing = true.';

-- ---------------------------------------------------------------- 2. the work-item predicate, now a caller
CREATE OR REPLACE FUNCTION governor_admits(p_item_type text) RETURNS boolean
LANGUAGE sql STABLE AS $FN$
  SELECT governor_admits_class(m.class, m.llm_bearing)
  FROM (SELECT 1) d
  LEFT JOIN governor_work_class_map m ON m.item_type = p_item_type
$FN$;

COMMENT ON FUNCTION governor_admits(text) IS
'D4 spend governor (AGOV-013): TRUE unless the governor currently withholds this item_type.
Called by the dispatch selector, the Go loader and the claim backstop — do not change the
signature. The logic lives in governor_admits_class (migration 751); this is the work-item
map lookup and nothing else. Unmapped item types = maintenance + llm_bearing (shed earliest).';

-- ---------------------------------------------------------------- 3. the agent-type namespace
-- Deliberately a SEPARATE table, not more rows in governor_work_class_map: item types and
-- agent types are different namespaces, and a shared primary key would silently collide the
-- day one string appears in both.
CREATE TABLE governor_agent_class_map (
  agent_type  text PRIMARY KEY,
  class       text NOT NULL CHECK (class IN ('maintenance','build','research')),
  llm_bearing boolean NOT NULL,
  note        text,
  updated_at  timestamptz NOT NULL DEFAULT now()
);

COMMENT ON TABLE governor_agent_class_map IS
'D4b (AGOV-013): agent_type -> shed class, for spend that is NOT driven by a site work item.
UNMAPPED AGENT TYPES ARE ADMITTED (see governor_admits_agent) — the OPPOSITE default to
governor_work_class_map, and deliberate: an unmapped work item is a queue row we chose to file,
whereas an unmapped agent type is every agent in the estate, and defaulting those to shed would
turn one un-mapped name into a fleet outage.';

-- Seeded at the LAST level on purpose. ⚠ THE LEVEL IS AN OWNER DECISION AND IT IS ONE UPDATE:
--   UPDATE governor_agent_class_map SET class='maintenance' WHERE agent_type='council-gate';
-- The owner ruled that council spend should be sheddable and called it "a fairly easy save if
-- it comes to the crunch". "Crunch" reads as LATE, so 'research' (L3) is the conservative
-- reading and the safe default; if the intent was the biggest, earliest saving, 'maintenance'
-- (L1) is the other end. Council review is ADVISORY — it cannot block a commit — so what is
-- lost at a shed is review quality and latency, not the ability to ship. Recorded as an open
-- owner question rather than absorbed into a default that looks decided.
INSERT INTO governor_agent_class_map (agent_type, class, llm_bearing, note) VALUES
  ('council-gate', 'research', true,
   'D4b seed 2026-09-03: 62% of fleet LLM spend, measured by orchestration lineage. Level is an OPEN OWNER QUESTION — see migration 751 header.');

CREATE FUNCTION governor_admits_agent(p_agent_type text) RETURNS boolean
LANGUAGE sql STABLE AS $FN$
  SELECT CASE
           WHEN m.agent_type IS NULL THEN true          -- unmapped agent types are ADMITTED
           ELSE governor_admits_class(m.class, m.llm_bearing)
         END
  FROM (SELECT 1) d LEFT JOIN governor_agent_class_map m ON m.agent_type = p_agent_type
$FN$;

COMMENT ON FUNCTION governor_admits_agent(text) IS
'D4b spend governor (AGOV-013): TRUE unless the governor currently withholds runs of this
agent_type. For spend with no site work item behind it (councils, verifiers, auditors).
UNMAPPED = ADMITTED, unlike the work-item predicate — mapping is opt-in per agent type, and a
typo must not shed the fleet. No caller until stage B, which is opt-in per agent config.';

-- ---------------------------------------------------------------- 4. the observable
-- D4 stage B's r1 round was gated by exactly this objection (bug_historian): shedding that
-- leaves no trace makes a withheld thing indistinguishable from a stuck one. For work items
-- the answer was a VIEW, because the predicate is computable from rows that still exist. A
-- refused RUN leaves NO row anywhere — the request is consumed and nothing is created — so
-- here the observable has to be WRITTEN. Without this table a shed council submission is
-- indistinguishable from the documented ~29-minute dispatch latency, which CLAUDE.md
-- explicitly warns sessions not to retry on.
CREATE TABLE governor_withheld_runs (
  id             uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  agent_type     text NOT NULL,
  correlation_id text,
  shed_level     int  NOT NULL,
  class          text,
  llm_bearing    boolean,
  request_topic  text,
  withheld_at    timestamptz NOT NULL DEFAULT now()
);
CREATE INDEX idx_governor_withheld_runs_corr ON governor_withheld_runs (correlation_id)
  WHERE correlation_id IS NOT NULL;
CREATE INDEX idx_governor_withheld_runs_at ON governor_withheld_runs (withheld_at DESC);

COMMENT ON TABLE governor_withheld_runs IS
'D4b (AGOV-013): one row per RUN the spend governor refused to start. This is the answer to
"my council submission never ran — is it queued or was it withheld?", which is otherwise
indistinguishable from the normal ~29 minute dispatch latency. Only refusals are written, so
it stays small; it is a log of a deliberate decision, not telemetry.';

CREATE VIEW governor_withheld_runs_recent AS
SELECT r.*, gs.shed_level AS current_shed_level, gc.enabled AS governor_enabled
FROM governor_withheld_runs r, governor_state gs, governor_config gc
WHERE gs.id = 1 AND gc.id = 1 AND r.withheld_at > now() - interval '7 days';

-- ---------------------------------------------------------------- verify (DO/RAISE; a SELECT cannot stop a COMMIT)
DO $$
DECLARE
  lvl int; cls text; bear boolean; want boolean; got boolean; n int;
  saved_level int; saved_enabled boolean;
BEGIN
  SELECT shed_level INTO saved_level FROM governor_state WHERE id=1;
  SELECT enabled INTO saved_enabled FROM governor_config WHERE id=1;
  -- Drive with the governor ENABLED so the comparison is not vacuous under a disabled config.
  UPDATE governor_config SET enabled = true WHERE id=1;

  -- (a) EQUIVALENCE, the whole point: new governor_admits must agree with the old body on
  --     every cell of level x mapped-class x llm_bearing, plus the unmapped default.
  FOR lvl IN 0..3 LOOP
    UPDATE governor_state SET shed_level = lvl WHERE id=1;
    FOREACH cls IN ARRAY ARRAY['maintenance','build','research'] LOOP
      FOREACH bear IN ARRAY ARRAY[true,false] LOOP
        INSERT INTO governor_work_class_map (item_type, class, llm_bearing, note)
        VALUES ('__751_probe__', cls, bear, 'transient verify probe')
        ON CONFLICT (item_type) DO UPDATE SET class=EXCLUDED.class, llm_bearing=EXCLUDED.llm_bearing;
        IF governor_admits('__751_probe__') IS DISTINCT FROM governor_admits_legacy('__751_probe__') THEN
          RAISE EXCEPTION '751 VERIFY: governor_admits DIVERGED from the old body at level=% class=% bearing=%',
            lvl, cls, bear;
        END IF;
      END LOOP;
    END LOOP;
    IF governor_admits('__751_never_mapped__') IS DISTINCT FROM governor_admits_legacy('__751_never_mapped__') THEN
      RAISE EXCEPTION '751 VERIFY: governor_admits DIVERGED on an UNMAPPED item_type at level=%', lvl;
    END IF;
  END LOOP;
  DELETE FROM governor_work_class_map WHERE item_type='__751_probe__';

  -- (b) the comparison must be CAPABLE of discriminating — a known-shed cell must read false.
  --     Without this arm, two functions that both returned TRUE everywhere would "agree".
  UPDATE governor_state SET shed_level = 1 WHERE id=1;
  INSERT INTO governor_work_class_map (item_type, class, llm_bearing, note)
  VALUES ('__751_probe__', 'maintenance', true, 'transient')
  ON CONFLICT (item_type) DO UPDATE SET class='maintenance', llm_bearing=true;
  IF governor_admits('__751_probe__') IS DISTINCT FROM false THEN
    RAISE EXCEPTION '751 VERIFY: control failed — a maintenance/bearing item is not shed at L1, so the agreement above proves nothing';
  END IF;
  IF governor_admits_legacy('__751_probe__') IS DISTINCT FROM false THEN
    RAISE EXCEPTION '751 VERIFY: control failed on the LEGACY body — the copy is not the live text';
  END IF;
  DELETE FROM governor_work_class_map WHERE item_type='__751_probe__';

  -- (c) the AGENT predicate, driven at every level (execution probes, not string assertions).
  FOR lvl IN 0..3 LOOP
    UPDATE governor_state SET shed_level = lvl WHERE id=1;
    IF governor_admits_agent('__751_no_such_agent__') IS DISTINCT FROM true THEN
      RAISE EXCEPTION '751 VERIFY: an unmapped agent_type was SHED at level % — that default would take out the fleet', lvl;
    END IF;
    want := NOT (lvl >= 3);                       -- seeded as research: admitted below L3, shed at L3
    got  := governor_admits_agent('council-gate');
    IF got IS DISTINCT FROM want THEN
      RAISE EXCEPTION '751 VERIFY: council-gate admitted=% at level %, expected %', got, lvl, want;
    END IF;
  END LOOP;

  -- (d) fail-open, driven not asserted: governor disabled => everything admitted at L3.
  UPDATE governor_state SET shed_level = 3 WHERE id=1;
  UPDATE governor_config SET enabled = false WHERE id=1;
  IF governor_admits_agent('council-gate') IS DISTINCT FROM true THEN
    RAISE EXCEPTION '751 VERIFY: disabled governor still shed an agent run — fail-open is broken';
  END IF;
  IF governor_admits('__751_never_mapped__') IS DISTINCT FROM true THEN
    RAISE EXCEPTION '751 VERIFY: disabled governor still shed a work item — fail-open is broken';
  END IF;

  -- restore the live posture exactly as found
  UPDATE governor_config SET enabled = saved_enabled WHERE id=1;
  UPDATE governor_state  SET shed_level = saved_level WHERE id=1;

  -- (e) INERT: nothing in the estate calls the new predicate yet.
  SELECT count(*) INTO n FROM agent_definitions
   WHERE is_active AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL
     AND default_config::text LIKE '%governor_admits_agent%';
  IF n <> 0 THEN RAISE EXCEPTION '751 VERIFY: % agent rows already reference governor_admits_agent — not inert', n; END IF;
  SELECT count(*) INTO n FROM scheduled_tasks WHERE pre_query LIKE '%governor_admits_agent%';
  IF n <> 0 THEN RAISE EXCEPTION '751 VERIFY: % scheduled tasks reference governor_admits_agent — not inert', n; END IF;

  -- (f) the selector text is byte-identical: the rewrite must not have touched any caller.
  IF (SELECT md5(default_config#>>'{workflow,steps,find_dispatchable_site,config,query}')
        FROM agent_definitions WHERE type='build-pipeline-trigger' AND is_active
         AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL ORDER BY version DESC LIMIT 1)
     <> 'fcbe8821a2a56512911955735796460e' THEN
    RAISE EXCEPTION '751 VERIFY: selector md5 changed — this migration must not touch callers';
  END IF;

  RAISE NOTICE '751 OK: governor_admits rewritten and PROVEN equivalent across 4 levels x 3 classes x 2 bearings + unmapped (with a discriminating control); governor_admits_agent driven at every level; unmapped agents admitted; fail-open proven; selector untouched; inert (0 callers).';
END $$;

DROP FUNCTION governor_admits_legacy(text);

COMMIT;
