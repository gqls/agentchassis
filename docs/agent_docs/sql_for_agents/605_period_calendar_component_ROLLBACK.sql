-- ROLLBACK for 605 — deactivates the period-calendar component. REFUSES if it is in use.
--
-- Deactivates rather than DELETEs: a page_components row referencing a deleted
-- component_id is a broken page, and is_active=false is the estate's retirement
-- mechanism (the section_type trigger's own message says so).
BEGIN;

DO $$
DECLARE n int;
BEGIN
  IF NOT EXISTS (SELECT 1 FROM content_components WHERE function = 'period-calendar' AND is_active) THEN
    RAISE EXCEPTION '605 ROLLBACK: no active period-calendar component — 605 was not applied, or it has already been rolled back';
  END IF;
  SELECT count(*) INTO n FROM page_components pc
    JOIN content_components cc ON cc.id = pc.component_id
   WHERE cc.function = 'period-calendar' AND cc.is_active;
  IF n > 0 THEN
    RAISE EXCEPTION '605 ROLLBACK: % live page_components use this component — deactivating it would leave those pages referencing an inactive row. Rebuild or repoint those sections first.', n;
  END IF;
END $$;

UPDATE content_components SET is_active = false, updated_at = now()
 WHERE function = 'period-calendar' AND is_active;

DO $$
BEGIN
  IF EXISTS (SELECT 1 FROM content_components WHERE function = 'period-calendar' AND is_active) THEN
    RAISE EXCEPTION '605 ROLLBACK VERIFY: the component is still active';
  END IF;
  RAISE NOTICE '605 ROLLBACK OK: period-calendar deactivated';
END $$;

COMMIT;
