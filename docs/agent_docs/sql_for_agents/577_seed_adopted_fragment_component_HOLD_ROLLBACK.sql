-- ROLLBACK for 577 — remove the seeded `adopted-fragment` component.
--
-- ⚠ SAFE ONLY WHILE NOTHING IS BOUND TO IT. Once `adopt_unidentified_fragments`
-- has been armed, live `page_components` rows carry this component's id, and
-- deleting it would orphan them — the row would then point at nothing, which is a
-- WORSE state than the one bugs_open/357 describes (a row that claims the wrong
-- component can at least be repaired; a row pointing at a deleted component
-- cannot be re-rendered at all).
--
-- So this refuses rather than cascades. If it refuses, the correct order is:
-- disarm the flag, repair or re-type the bound rows, then run this.

BEGIN;

DO $$
DECLARE bound int;
BEGIN
    SELECT count(*) INTO bound
      FROM page_components pc
      JOIN content_components cc ON cc.id = pc.component_id
     WHERE cc.function = 'adopted-fragment';
    IF bound > 0 THEN
        RAISE EXCEPTION
          'refusing to remove adopted-fragment: % page_components row(s) are bound to it. '
          'Disarm adopt_unidentified_fragments, re-type those rows, then re-run.', bound;
    END IF;
END $$;

DELETE FROM content_components WHERE function = 'adopted-fragment';

COMMIT;
