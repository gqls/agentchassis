-- FILE: SQL_2026-08-15_unlock_cta_components.sql
--
-- OWNER DIRECTIVE 2026-08-15 (bugs_closed/268 closure follow-up): "take off
-- webdesign.uk's 8 emergency locks". Reverses SQL_2026-08-12k (same
-- directory), which locked every webdesign.uk hero/call-to-action row that
-- carried a CTA destination, as a tourniquet while bugs_open/268 (a
-- content_rewrite deletes renderer-sourced CTA url keys) was open.
--
-- WHY IT IS NOW SAFE: 268 is CLOSED (bugs_closed/268 §12). The carry fix
-- (`8f899cc8d`, live since v1.0.1298, re-verified on v1.0.1300 stamp
-- a2a691213) makes a regeneration re-supply stored renderer/static-sourced
-- keys; proven by canary AND by a permanence rewrite on a repaired row —
-- the exact operation these locks were guarding against no longer deletes
-- the keys.
--
-- SCOPE: exactly the 8 hero/call-to-action permanent locks (contact/hero,
-- faq/both, how-it-works/both, index/hero, what-you-get/both).
-- NOT TOUCHED: contact/chat-input-box (locked 2026-08-11 by the sibling
-- webdesign_uk_build_service lane for its own reasons — not one of the 8,
-- not this directive's to lift).

BEGIN;

UPDATE page_components pc
   SET locked_at  = NULL,
       locked_by  = NULL,
       lock_type  = NULL,
       updated_at = now()
  FROM pages p
 WHERE pc.page_id = p.id
   AND p.site_id = '1fcfa4f3-ec80-4010-878b-b971cd46711f'
   AND p.status = 'active'
   AND pc.slot_name IN ('hero', 'call-to-action')
   AND pc.lock_type = 'permanent';

DO $$
DECLARE n_left int; n_chat int; n_keys int;
BEGIN
  -- All 8 hero/call-to-action locks must be gone.
  SELECT count(*) INTO n_left
    FROM page_components pc JOIN pages p ON p.id = pc.page_id
   WHERE p.site_id='1fcfa4f3-ec80-4010-878b-b971cd46711f'
     AND pc.slot_name IN ('hero','call-to-action')
     AND pc.locked_at IS NOT NULL;
  IF n_left <> 0 THEN RAISE EXCEPTION 'expected 0 hero/cta locks left, got %', n_left; END IF;

  -- The sibling lane's chat-input-box lock must be UNTOUCHED.
  SELECT count(*) INTO n_chat
    FROM page_components pc JOIN pages p ON p.id = pc.page_id
   WHERE p.site_id='1fcfa4f3-ec80-4010-878b-b971cd46711f'
     AND pc.slot_name = 'chat-input-box'
     AND pc.lock_type = 'permanent';
  IF n_chat <> 1 THEN RAISE EXCEPTION 'chat-input-box lock disturbed: % (expected 1)', n_chat; END IF;

  -- The rows keep their destinations (the unlock must not itself lose data).
  SELECT count(*) INTO n_keys
    FROM page_components pc JOIN pages p ON p.id = pc.page_id
   WHERE p.site_id='1fcfa4f3-ec80-4010-878b-b971cd46711f'
     AND pc.slot_name IN ('hero','call-to-action') AND p.status='active'
     AND (pc.content_data ? 'cta_url' OR pc.content_data ? 'primary_cta_url');
  IF n_keys <> 9 THEN RAISE EXCEPTION 'expected 9 rows carrying a CTA destination (8 previously locked + index/call-to-action restored by 268), got %', n_keys; END IF;
END $$;

COMMIT;
