-- ============================================================================
-- 349_bugfix_136_residue_repairs_ROLLBACK.sql
--
-- Reverses 349: restores the deprecated key spellings, re-adds the dead
-- plan_sections.domain key, and re-labels the three completeness-discovery
-- rows back to design.
--
-- Both directions are behaviour-neutral today: the live binary honours old
-- and new spellings alike (DeprecatedConfigKeys/ResolveConfigSetting), and no
-- live consumer distinguishes pipeline 'design' from 'content'
-- (bugs_open/136 §6). This exists for completeness of the sidecar, not
-- because a revert is expected to be needed.
--
-- The 19 paths are ENUMERATED, not discovered by walking for the new
-- spelling. Two live steps carried `item_pipeline` before 349 ever ran —
-- claims-auditor.request_claims_review and
-- site-work-orchestrator.load_work_items — and a generic descent would
-- "revert" those two into a spelling they never had.
-- ============================================================================

BEGIN;

-- C reversed: new spelling -> old spelling, at exactly the paths 349 touched
DO $$
DECLARE
  t   record;
  hit int;
  n   int := 0;
BEGIN
  FOR t IN
    SELECT agent_type, parent_path, new_key, old_key FROM (VALUES
      ('build-briefing-agent',         ARRAY['workflow','steps','create_next_item','config'],                                                          'item_pipeline',   'item_domain'),
      ('completeness-discovery-agent', ARRAY['workflow','steps','run_checks','config'],                                                                'check_pipeline',  'check_domain'),
      ('component-quality-auditor',    ARRAY['workflow','steps','create_regen_items','config','sub_workflow','steps','create_work_item','config'],      'item_pipeline',   'item_domain'),
      ('deduplicate-sections',         ARRAY['workflow','steps','queue_rerender','config'],                                                            'item_pipeline',   'item_domain'),
      ('design-discovery-agent',       ARRAY['workflow','steps','run_checks','config'],                                                                'check_pipeline',  'check_domain'),
      ('domain-research-classifier',   ARRAY['workflow','steps','create_next_item','config'],                                                          'item_pipeline',   'item_domain'),
      ('domain-strategist',            ARRAY['workflow','steps','create_next_item','config'],                                                          'item_pipeline',   'item_domain'),
      ('domain-submitter',             ARRAY['workflow','steps','create_research_item','config'],                                                      'item_pipeline',   'item_domain'),
      ('improvement-loop',             ARRAY['workflow','steps','insert_rerender_item','config'],                                                      'item_pipeline',   'item_domain'),
      ('improvement-loop',             ARRAY['workflow','steps','record_not_converging','config'],                                                     'item_pipeline',   'item_domain'),
      ('improvement-loop',             ARRAY['workflow','steps','triage_findings','config'],                                                           'target_pipeline', 'target_domain'),
      ('internal-linker',              ARRAY['workflow','steps','create_items_loop','config','sub_workflow','steps','create_rewrite_item','config'],    'item_pipeline',   'item_domain'),
      ('quality-discovery-agent',      ARRAY['workflow','steps','run_checks','config'],                                                                'check_pipeline',  'check_domain'),
      ('tool-auditor',                 ARRAY['workflow','steps','create_items_loop','config','sub_workflow','steps','create_improve_item','config'],    'item_pipeline',   'item_domain'),
      ('tool-auditor',                 ARRAY['workflow','steps','create_items_loop','config','sub_workflow','steps','create_review_item','config'],     'item_pipeline',   'item_domain'),
      ('tool-improver',                ARRAY['workflow','steps','create_rerender_item','config'],                                                      'item_pipeline',   'item_domain'),
      ('tool-suggester',               ARRAY['workflow','steps','create_items_loop','config','sub_workflow','steps','create_library_item','config'],    'item_pipeline',   'item_domain'),
      ('tool-suggester',               ARRAY['workflow','steps','create_items_loop','config','sub_workflow','steps','create_novel_item','config'],      'item_pipeline',   'item_domain'),
      ('vertical-exemplar-researcher', ARRAY['workflow','steps','create_next_item','config'],                                                          'item_pipeline',   'item_domain')
    ) AS m(agent_type, parent_path, new_key, old_key)
  LOOP
    UPDATE agent_definitions ad
       SET default_config = jsonb_set(
             ad.default_config,
             t.parent_path || t.old_key,
             ad.default_config #> (t.parent_path || t.new_key)
           ) #- (t.parent_path || t.new_key),
           updated_at = now()
     WHERE ad.type = t.agent_type
       AND ad.is_active AND COALESCE(ad.is_snapshot,false)=false
       AND ad.deleted_at IS NULL
       AND ad.default_config #> t.parent_path ? t.new_key;
    GET DIAGNOSTICS hit = ROW_COUNT;
    IF hit <> 1 THEN
      RAISE EXCEPTION '349 ROLLBACK: %.% (% -> %) touched % rows, expected 1',
        t.agent_type, array_to_string(t.parent_path,'.'), t.new_key, t.old_key, hit;
    END IF;
    n := n + 1;
  END LOOP;
  RAISE NOTICE '349 ROLLBACK: restored % old-spelling keys', n;
END $$;

-- B reversed: re-add the (dead) domain key to plan_sections
UPDATE agent_definitions
SET default_config = jsonb_set(
      default_config,
      '{workflow,steps,plan_sections,config,domain}',
      '"site_record.domain"'::jsonb
    ),
    updated_at = now()
WHERE type='page-build-handler' AND is_active
  AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;

-- A reversed: by id, NOT by (created_by, pipeline) — the agent files
-- legitimately-content rows since the code fix, and a predicate-keyed
-- reversal would sweep those too. These are the three rows 349 repaired.
UPDATE site_work_items
   SET pipeline='design', updated_at=now()
 WHERE id IN ('b6edce72-c53c-481f-bcaa-d57dc2214a63',
              'a833555d-3263-491d-93f7-870961998fda',
              '74bb48ff-8d17-4e2d-8239-27c299fe264b')
   AND pipeline='content';

COMMIT;
