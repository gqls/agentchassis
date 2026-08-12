-- FILE: SQL_2026-08-12k_lock_cta_components.sql
--
-- Stop webdesign.uk's call-to-action buttons being deleted by every rewrite.
-- Owner, 2026-08-12: "please fix the loss of the button problem on each
-- refresh, somehow."
--
-- THE DEFECT, now measured with a control rather than inferred. A
-- `content_rewrite` drops the resolver-populated URL keys (`cta_url`,
-- `primary_cta_url`, `secondary_cta_url`) from `content_data`. Both templates
-- gate the anchor on the URL rather than the label, so each button renders as
-- NOTHING: no error, no missing text, healthy byte counts.
--   before the 20:2x rewrite: 7 components carried the keys, 28 hrefs site-wide
--   after  it:                all 7 at 0|0|0, 13 hrefs — 15 links gone
--   CONTROL: `contact/hero`, the one page NOT in the rewrite, kept its keys
--            (1|0|1) and both its links.
-- That control is what the earlier 090 lacked: it was asked AFTER the repair,
-- so it sampled restored values and refuted the mechanism on a repaired system.
--
-- WHY A LOCK, AND WHY NOT THE OBVIOUS ALTERNATIVES.
--   * Fixing the shared component's `input_schema` would reach 20 sites and 276
--     hero instances. A hardcoded `/contact.html` fallback is wrong for sites
--     that have no such page, and would ship broken links fleet-wide.
--   * Repairing `content_data` alone does NOT hold: proven twice today. A
--     `page_rerender` dispatched after the repair still rendered no buttons.
--   * A lock is site-scoped, needs no code change, and is the estate's own
--     mechanism — `bugs_open/058`. `save_page_sections` loads locked rows
--     (`loadActiveLockedRows`) and its DELETE excludes anything not
--     agent-writable, where agent-writable is
--     `locked_at IS NULL OR (lock_type='timed' AND lock_expires_at < NOW())`.
--     So `permanent` holds against the REWRITE path, not merely the re-render
--     path the sibling lane proved it on.
--
-- THE COST, STATED PLAINLY: automation can no longer change the copy in these
-- seven components. The hero headline carries the £149 offer, so an offer
-- change now needs a deliberate unlock, edit, relock. That is the trade — a
-- blocked change is RECORDED as a `lock_blocked_change` work item and is
-- therefore visible, whereas the button loss was silent. Visible friction beats
-- silent damage.
--
-- NOT LOCKED, deliberately: `index/call-to-action` (has labels but has never
-- had URLs — locking would freeze a defect in place and block its repair), and
-- the guide's hero/call-to-action (no buttons at all). The guide's `article-body`
-- is left unlocked too: it is prose that must stay editable, and its link set is
-- protected mechanically instead, by `required_links` + `gate_page_links.py`.
--
-- TO CHANGE THIS COPY LATER:
--   UPDATE page_components SET locked_at=NULL, locked_by=NULL, lock_type=NULL
--    WHERE id='<id>';   -- then rewrite, then re-run this file
-- and check the href count either side, because the rewrite will drop the URLs
-- again while the underlying defect is open.

BEGIN;

UPDATE page_components pc
   SET locked_at  = now(),
       locked_by  = 'ai_site_selling_automation lane 2026-08-12: CTA destinations are '
                    || 'dropped by every content_rewrite (fleet defect, 216 components / 19 '
                    || 'sites). Unlock deliberately to change this copy, then re-apply '
                    || 'SQL_2026-08-12d + e and re-run gate_page_links.py.',
       lock_type  = 'permanent',
       updated_at = now()
  FROM pages p
 WHERE pc.page_id = p.id
   AND p.site_id = '1fcfa4f3-ec80-4010-878b-b971cd46711f'
   AND p.status = 'active'
   AND pc.slot_name IN ('hero', 'call-to-action')
   AND (pc.content_data ? 'cta_url' OR pc.content_data ? 'primary_cta_url')  -- only rows worth protecting
   AND pc.locked_at IS NULL;
-- NOTE: scoped by "carries a URL", NOT by a page list. The first draft named
-- the four offer pages and the verify block then caught `contact/hero`, which
-- carries a live button and would have been left exposed — the assertion was
-- broader than the update, which is the right way round for that mistake to
-- surface. contact/hero was not in today's rewrite (its keys survived, which is
-- what made it the control), but any future rewrite of that page would drop it.
-- The sibling lane owns the contact page's chat box and is told in their NOTES.

DO $$
DECLARE n_locked int; n_unprotected int;
BEGIN
  SELECT count(*) INTO n_locked
    FROM page_components pc JOIN pages p ON p.id = pc.page_id
   WHERE p.site_id='1fcfa4f3-ec80-4010-878b-b971cd46711f'
     AND pc.slot_name IN ('hero','call-to-action') AND p.status='active'
     AND pc.locked_at IS NOT NULL AND pc.lock_type='permanent';
  IF n_locked <> 8 THEN RAISE EXCEPTION 'expected 8 locked CTA components, got %', n_locked; END IF;

  -- Every component that currently RENDERS a button must now be protected.
  -- A lock that missed one is the failure that looks like success.
  SELECT count(*) INTO n_unprotected
    FROM page_components pc JOIN pages p ON p.id = pc.page_id
   WHERE p.site_id='1fcfa4f3-ec80-4010-878b-b971cd46711f' AND p.status='active'
     AND pc.slot_name IN ('hero','call-to-action')
     AND (pc.content_data ? 'cta_url' OR pc.content_data ? 'primary_cta_url')
     AND pc.locked_at IS NULL;
  IF n_unprotected <> 0 THEN
    RAISE EXCEPTION '% CTA-bearing components are still unlocked', n_unprotected;
  END IF;
END $$;

COMMIT;
