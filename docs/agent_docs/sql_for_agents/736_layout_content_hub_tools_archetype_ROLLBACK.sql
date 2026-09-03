-- ROLLBACK for 736 — retires the content-hub-tools layout.
--
-- DEACTIVATES rather than deletes: css_themes.layout_id references layouts
-- (ON DELETE SET NULL), so a DELETE would silently null the layout on any
-- site that has since composed onto it, and the matcher only reads
-- is_active=true rows. Sites already on it keep rendering; new compositions
-- stop choosing it. A human decides about those sites, not this file.
BEGIN;
DO $$
DECLARE v_sites int;
BEGIN
    SELECT count(*) INTO v_sites
      FROM sites s JOIN style_collections sc ON sc.id=s.style_collection_id
      JOIN css_themes t ON t.id=sc.css_theme_id JOIN layouts l ON l.id=t.layout_id
     WHERE l.name='content-hub-tools';
    RAISE NOTICE '736 ROLLBACK: % site(s) currently composed onto content-hub-tools will KEEP it (deactivating, not deleting).', v_sites;
    UPDATE layouts SET is_active=false, updated_at=NOW() WHERE name='content-hub-tools' AND origin='seed';
    IF NOT FOUND THEN RAISE EXCEPTION '736 ROLLBACK: no seed row named content-hub-tools.'; END IF;
END $$;
COMMIT;
