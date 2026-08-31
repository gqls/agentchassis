-- ROLLBACK for 667: return the privacy policy to unprotected (generic, no component lock).
BEGIN;
UPDATE pages p SET rebuild_policy = 'generic', updated_at = now()
  FROM sites s WHERE s.id=p.site_id AND s.domain='finetuning.uk' AND p.name='privacy-policy';
UPDATE page_components pc
   SET locked_at=NULL, locked_by=NULL, lock_type=NULL, lock_expires_at=NULL
  FROM pages p, sites s
 WHERE p.id=pc.page_id AND s.id=p.site_id AND s.domain='finetuning.uk' AND p.name='privacy-policy';
COMMIT;
