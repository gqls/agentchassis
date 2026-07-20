-- PATCH_render_guardian_031_content_hash.sql (2026-07-20)
--
-- bugs_closed/031: the render_guardian seat's prompt asserted scoped page-rerender
-- "SKIPS pages whose content hash is unchanged". No such code exists and never did
-- (git log -S over the rerender actions: no commits). The claim came from a wrong
-- concept-register entry (STY-048, since corrected) and produced a HIGH-severity
-- block on a correct plan (submission 7ef4de4e round 3, refuted round 4).
--
-- SURGICAL: replaces two substrings inside review_render_guardian's prompt_template
-- on the LIVE fix-proposer row only. The council-gate mirror is NOT patched here —
-- run 099_SYNC_gate_roster.py --apply afterwards (do not hand-patch the gate).
-- Idempotent: the LIKE guard makes a re-run (or a concurrent-change race) a 0-row
-- no-op. v16 seed file corrected in the same commit so a replay cannot resurrect it.

BEGIN;

SELECT snapshot_agent('fix-proposer', 'pre-update: PATCH 031 — render_guardian content-hash claim was never true')
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
        'REGENERATES section HTML from content_data but SKIPS pages whose content hash is unchanged -- silently wrong for header/footer/chrome-only changes; assemble mode (page_id, no reason) re-embeds chrome unconditionally. A fix routing chrome changes through scoped mode will silently miss pages.',
        're-renders each stored section from its html_template + stored content_data. There is NO content-hash skip (bugs_closed/031: that claim was never true of the code). Its real bail-outs (rerender_page_sections_action.go): page-level -- skipped when the page has no stored components, escalated-to-writer when any section''s stored content_data is absent or missing a required llm field; neither writes nor deploys. Section-level -- stored HTML carried when the component cannot be loaded, its plan is not ready, or html_template is empty. Assemble mode (page_id, no reason) re-embeds chrome unconditionally. A fix routing chrome-only changes through scoped mode can silently miss exactly the pages that bail out. When judging a rerender claim, cite the code path, not a register entry -- the register is documentation, not the implementation.'
      ),
      '(hash-skip)',
      '(whose page-level bail-outs can leave pages undeployed)'
    )
  )
)
WHERE d.type='fix-proposer' AND d.is_active
  AND COALESCE(d.is_snapshot,false)=false AND d.deleted_at IS NULL
  AND d.default_config #>> '{workflow,steps,review_render_guardian,config,prompt_template}'
      LIKE '%content hash is unchanged%';

-- Post-conditions: false claim gone, corrected text present, on the row just updated.
SELECT type,
       position('content hash is unchanged' in default_config #>> '{workflow,steps,review_render_guardian,config,prompt_template}') AS false_claim_pos,   -- expect 0
       position('NO content-hash skip' in default_config #>> '{workflow,steps,review_render_guardian,config,prompt_template}') AS corrected_pos          -- expect > 0
FROM agent_definitions
WHERE type='fix-proposer' AND is_active AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;

COMMIT;
