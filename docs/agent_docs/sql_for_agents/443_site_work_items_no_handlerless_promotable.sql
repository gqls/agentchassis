-- 443_site_work_items_no_handlerless_promotable.sql
--
-- bugs_open/284, third and last step: make the bad state UNREPRESENTABLE.
--
-- The promoter guard (`7027a2801`, live in v1.0.1305) stops the mechanism that
-- produced all 58 machine-made rows. It cannot stop the other path: ~20 raw
-- `INSERT INTO site_work_items` sites bypass the shared door, and a human with
-- psql bypasses everything — which is how the 2 hand-inserted rows 442 repaired
-- got there. A row that is handler-less AND in a promotable/claimed state is
-- always a defect: the claim path will find no handler and stamp it `blocked`
-- with a routing error that misdescribes a flag-only finding.
--
-- ORDERING — WHY THIS COULD NOT LAND EARLIER, AND CAN NOW. DB constraints are
-- live immediately; Go is inert until a roll. Had this landed before v1.0.1305,
-- the OLD promoter's blanket `UPDATE ... WHERE status='detected'` would have
-- ERRORED on any site holding a flag-only detected row — breaking
-- `improvement-loop` fleet-wide. The rolled promoter now excludes those rows
-- itself, so the constraint has nothing left to trip on.
--
-- `[MEASURED 2026-08-16]` violators before applying: 0 (checked twice — before
-- migration 442 and again in the guard below). 37 flag-only rows sit at
-- `detected` (head_essentials_missing 36, image_url_404 1) plus the 40 442 just
-- restored: `detected` is DELIBERATELY NOT in the forbidden list — it is the
-- state those findings are filed in and belong in. What is forbidden is being
-- PROMOTED there without a handler.
--
-- NOT VALID + VALIDATE is deliberate: it takes no long table lock on a live
-- table, and VALIDATE then proves the existing 4,000+ rows comply rather than
-- assuming it.
--
-- ROLLBACK: 443_..._ROLLBACK.sql drops the constraint.

BEGIN;

DO $$
DECLARE n_violators int; already int;
BEGIN
    SELECT count(*) INTO already FROM pg_constraint
     WHERE conname = 'swi_no_handlerless_promotable' AND conrelid = 'site_work_items'::regclass;
    IF already > 0 THEN
        RAISE EXCEPTION 'MIGRATION 443: constraint swi_no_handlerless_promotable already exists — already applied';
    END IF;

    SELECT count(*) INTO n_violators FROM site_work_items
     WHERE COALESCE(handler_agent,'')='' AND status IN ('triaged','approved','claimed');
    IF n_violators > 0 THEN
        RAISE EXCEPTION 'MIGRATION 443: % handler-less promotable/claimed row(s) exist — repair them (see migration 442) before constraining', n_violators;
    END IF;
    RAISE NOTICE 'migration 443: 0 violators, adding constraint';
END $$;

ALTER TABLE site_work_items
  ADD CONSTRAINT swi_no_handlerless_promotable
  CHECK (NOT (COALESCE(handler_agent,'') = '' AND status IN ('triaged','approved','claimed')))
  NOT VALID;

ALTER TABLE site_work_items VALIDATE CONSTRAINT swi_no_handlerless_promotable;

DO $$
DECLARE convalidated boolean;
BEGIN
    SELECT c.convalidated INTO convalidated FROM pg_constraint c
     WHERE conname='swi_no_handlerless_promotable' AND conrelid='site_work_items'::regclass;
    IF convalidated IS NOT TRUE THEN
        RAISE EXCEPTION 'MIGRATION 443: constraint present but NOT VALIDATED — existing rows unproven';
    END IF;
    RAISE NOTICE 'migration 443 OK: swi_no_handlerless_promotable added AND validated';
END $$;

COMMIT;
