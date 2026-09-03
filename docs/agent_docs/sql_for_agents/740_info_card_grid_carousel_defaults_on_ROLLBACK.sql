-- 740_info_card_grid_carousel_defaults_on_ROLLBACK.sql
--
-- Removes the resolution-time default added by 740. After this, `carousel` is
-- absent from the schema descriptor again, `fallback == nil`, and the static
-- branch in planSection writes nothing — instances revert to the grid on their
-- NEXT render. Already-rendered carousels stay carousels until then, and any
-- instance whose content_data now stores `carousel: true` keeps it (carryStored
-- beats the absent fallback) — this rollback does not touch stored page data.

BEGIN;

DO $$
DECLARE n int;
BEGIN
    SELECT count(*) INTO n FROM content_components
     WHERE is_active AND name = 'info-card-grid'
       AND input_schema->'fields'->'carousel' ? 'fallback';
    IF n <> 1 THEN
        RAISE EXCEPTION 'ABORT: expected 1 active info-card-grid carrying a carousel '
                        'fallback, found % — nothing to roll back, or the schema has moved.', n;
    END IF;
END $$;

UPDATE content_components
   SET input_schema = input_schema #- '{fields,carousel,fallback}',
       updated_at = now()
 WHERE is_active AND name = 'info-card-grid';

DO $$
DECLARE n int;
BEGIN
    SELECT count(*) INTO n FROM content_components
     WHERE is_active AND name = 'info-card-grid'
       AND input_schema->'fields'->'carousel' ? 'fallback';
    IF n <> 0 THEN
        RAISE EXCEPTION 'ABORT: the fallback survived the delete (% rows)', n;
    END IF;
    SELECT count(*) INTO n FROM content_components
     WHERE is_active AND name = 'info-card-grid'
       AND input_schema->'fields'->'carousel'->>'source' = 'static';
    IF n <> 1 THEN
        RAISE EXCEPTION 'ABORT: the carousel field itself is gone — the #- hit the wrong path';
    END IF;
    RAISE NOTICE '740 ROLLBACK: carousel default removed; instances revert to grid on next render';
END $$;

COMMIT;
