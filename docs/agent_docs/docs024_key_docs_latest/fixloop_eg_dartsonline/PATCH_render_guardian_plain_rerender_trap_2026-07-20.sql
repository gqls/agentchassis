-- PATCH (2026-07-20): render-guardian seat — add the owner's named trap:
-- "page-rerender re-deploys the existing HTML — it does not regenerate it from
-- content_data." Stated in the CORRECTED (bugs_closed/031) mechanics: assemble
-- mode re-embeds each section's EXISTING stored HTML; it never re-renders
-- html_template against content_data; even scoped mode carries stored HTML in
-- its section-level bail-outs. This is the bugs_open/024 false-green class: edit
-- content_data / a template / a tool source, fire a plain rerender, deploy the
-- OLD html, read the completed rerender as proof.
--
-- SURGICAL: two replace() edits on the LIVE prompt (bullet + a new judge clause
-- (e)). The prompt was seen to change UNDER THIS SESSION between a needle count
-- and a dump (the 031 thread's own patch landed in that window), so both anchors
-- are re-asserted HERE, inside the UPDATE's WHERE, and the patch is a 0-row
-- no-op if either anchor is gone or the patch is already present.
-- After applying: run 099_SYNC_gate_roster.py --apply (do NOT hand-patch the gate).

BEGIN;

SELECT snapshot_agent('fix-proposer', 'pre-update: PATCH render-guardian plain-rerender trap (owner, 2026-07-20)')
WHERE EXISTS (SELECT 1 FROM agent_definitions
              WHERE type='fix-proposer' AND is_active
                AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL);

UPDATE agent_definitions d
SET default_config = jsonb_set(
  d.default_config,
  '{workflow,steps,review_render_guardian,config,prompt_template}',
  to_jsonb(
    replace(
      replace(
        d.default_config #>> '{workflow,steps,review_render_guardian,config,prompt_template}',
        'the register is documentation, not the implementation.',
        'the register is documentation, not the implementation.' || chr(10) ||
        '- PLAIN PAGE-RERENDER DOES NOT REGENERATE: assemble mode re-deploys each section''s EXISTING stored HTML unchanged -- it never re-renders html_template against content_data; and even scoped mode carries stored HTML through its section-level bail-outs. A plan that edits content_data, a component template, or a tool source and then fires a plain page-rerender as its proof has deployed the OLD html -- the completed rerender is a FALSE GREEN (the bugs_open/024 class). Regeneration needs the scoped path (or a page build), and verification must check the DEPLOYED page changed, not that a rerender completed.'
      ),
      'or remove a validation layer. If the fix does not touch rendering/assembly/styling, approve.',
      'or remove a validation layer; (e) does any edit rely on a plain page-rerender to surface a content_data/template/tool-source change, or cite a completed rerender as proof the change is live. If the fix does not touch rendering/assembly/styling, approve.'
    )
  )
),
updated_at = now()
WHERE d.type='fix-proposer' AND d.is_active AND COALESCE(d.is_snapshot,false)=false AND d.deleted_at IS NULL
  AND (d.default_config #>> '{workflow,steps,review_render_guardian,config,prompt_template}') LIKE '%the register is documentation, not the implementation.%'
  AND (d.default_config #>> '{workflow,steps,review_render_guardian,config,prompt_template}') LIKE '%or remove a validation layer. If the fix does not touch rendering/assembly/styling, approve.%'
  AND (d.default_config #>> '{workflow,steps,review_render_guardian,config,prompt_template}') NOT LIKE '%PLAIN PAGE-RERENDER DOES NOT REGENERATE%';

COMMIT;

-- Verify (expect t,t):
--   SELECT (default_config #>> '{workflow,steps,review_render_guardian,config,prompt_template}') LIKE '%PLAIN PAGE-RERENDER DOES NOT REGENERATE%' AS bullet_in,
--          (default_config #>> '{workflow,steps,review_render_guardian,config,prompt_template}') LIKE '%(e) does any edit rely on a plain page-rerender%' AS judge_in
--   FROM agent_definitions WHERE type='fix-proposer' AND is_active AND NOT is_snapshot;
-- Rollback: restore the pre-update snapshot from agent_definitions_backup.
