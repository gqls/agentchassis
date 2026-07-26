-- 220_claimed_item_timeout_generic_evidence_ROLLBACK.sql
--
-- Restores the claimed-item-timeout pre_query exactly as it stood before
-- migration 220 (captured from the live row 2026-07-26, md5 27ea3fd389f6843064d421d6a5833e30).
--
-- Applying this reinstates the 3-item-type evidence gap that bugs_open/006 §C
-- describes: 15 of 18 item types go back to a 40-minute reset even when their
-- work succeeded. Use it if the generic branch is completing something it
-- should not, and say what in the bug file.

BEGIN;

UPDATE scheduled_tasks
SET pre_query = $q$

    WITH completed_by_evidence AS (
    -- Items where the handler's work is provably done on the specific
    -- targeted artifact (not just "something on the same site changed").
    -- See debugging guide section 9: "claimed-item-timeout evidence
    -- check produces false-positive completions" for the prior bug.
    UPDATE site_work_items wi
    SET status = 'complete',
        completed_at = NOW(),
        error = 'Auto-completed: work verified done despite lost response'
    WHERE wi.status = 'claimed'
      AND wi.claimed_at < NOW() - INTERVAL '15 minutes'
      AND (
        -- Content pages: the page's OWN content artifact (page_components)
        -- was written by THIS claim's handler. needs_content_page produces
        -- components, not a deploy (deploy is a separate page_rerender item),
        -- so we check the artifact directly rather than build_status='deployed'
        -- (a downstream, separately-set flag that can be true with zero
        -- components — see the gamesdesign homepage, 2026-06-04). The
        -- updated_at > claimed_at guard ensures the components are from THIS
        -- claim's run, not stale rows from a prior plan/generation.
        (wi.item_type = 'needs_content_page' AND wi.page_id IS NOT NULL AND EXISTS (
            SELECT 1 FROM page_components pc
            WHERE pc.page_id = wi.page_id
              AND pc.component_id IS NOT NULL
              AND pc.rendered_html IS NOT NULL
              AND pc.rendered_html <> ''
              AND pc.updated_at > wi.claimed_at
        ))
        OR
        -- Page rerenders: the specific page was deployed after claim.
        -- Note: NOT 'needs_rerender' — that's a site-level orchestrator
        -- with page_id NULL, can't have per-page evidence. Let it
        -- fall through to reset. A page_rerender's job IS to deploy, so
        -- deployed_at is the correct artifact here (the deployed flag is
        -- hardened separately by the UpdatePageStatusAction guard).
        (wi.item_type = 'page_rerender' AND wi.page_id IS NOT NULL AND EXISTS (
            SELECT 1 FROM pages p
            WHERE p.id = wi.page_id
              AND p.build_status = 'deployed'
              AND p.deployed_at IS NOT NULL
              AND p.deployed_at > wi.claimed_at
        ))
        OR
        -- Design items: site-level by nature. CAVEAT — this branch still
        -- has narrow false-positive potential because (a) it only checks
        -- the head slot, not header/footer, and (b) it uses updated_at
        -- rather than a deploy-specific timestamp (site_components has
        -- no deployed_at column). Acceptable for now; needs_design items
        -- are rare and the impact is bounded to design-only work.
        (wi.item_type = 'needs_design' AND EXISTS (
            SELECT 1 FROM site_components sc
            WHERE sc.site_id = wi.site_id
              AND sc.slot_name = 'head'
              AND sc.updated_at > wi.claimed_at
        ))
      )
    RETURNING id, item_type, handler_agent, status
),
reset AS (
    -- Remaining stuck items: no evidence of completion, reset for retry.
    -- needs_rerender items always land here now (previously could be
    -- false-positive auto-completed); that's intended — retry is cheap.
    UPDATE site_work_items
    SET status = CASE
            WHEN attempt_count + 1 >= max_attempts THEN 'failed'
            ELSE 'triaged'
        END,
        claimed_by = NULL,
        claimed_at = NULL,
        attempt_count = attempt_count + 1,
        error = CASE
            WHEN attempt_count + 1 >= max_attempts THEN 'Claim timed out (attempts exhausted)'
            ELSE 'Claim timed out — handler pod likely died'
        END
    WHERE status = 'claimed'
      AND claimed_at < NOW() - INTERVAL '40 minutes'
      AND id NOT IN (SELECT id FROM completed_by_evidence)
    RETURNING id, item_type, handler_agent, status
)
SELECT
    (SELECT COUNT(*) FROM completed_by_evidence) as auto_completed,
    (SELECT COUNT(*) FROM reset) as reset_count
    
$q$,
    updated_at = NOW()
WHERE name = 'claimed-item-timeout';

-- Guard: the new branch is gone and the row still parses.
DO $$
DECLARE
  v_sql text;
BEGIN
  SELECT pre_query INTO v_sql FROM scheduled_tasks WHERE name = 'claimed-item-timeout';
  IF v_sql LIKE '%completed_by_orchestration%' THEN
    RAISE EXCEPTION '220 rollback: the generic branch is still present';
  END IF;
  EXECUTE 'PREPARE claimed_item_timeout_rb_probe AS ' || v_sql;
  EXECUTE 'DEALLOCATE claimed_item_timeout_rb_probe';
END $$;

COMMIT;
