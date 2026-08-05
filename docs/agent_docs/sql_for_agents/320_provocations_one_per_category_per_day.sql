-- 320_provocations_one_per_category_per_day.sql — make the daily uniqueness rule
-- per CATEGORY, not per domain.
--
-- WHY THIS EXISTS
-- RFC_013 (RATIFIED 2026-08-05, owner ruling §7: one feed file per category).
-- Migration 282 encoded "one approved provocation per domain per day" as a partial
-- unique index, deliberately, so that an ambiguous daily was unrepresentable rather
-- than merely discouraged. That rule was right when a domain had exactly one live
-- provocation. Under categories it is wrong in a way that would look like a bug in
-- the seeding: "current political opinions" and "pets" are separate dailies with
-- separate audiences, and they must both be publishable on the same date.
--
-- WHAT STAYS THE SAME, AND IT IS THE POINT OF THE SHAPE
-- Uniqueness is still enforced by an INDEX, not by a check in the publisher, so
-- two approved provocations in ONE category on ONE date remain unrepresentable.
-- The ambiguity 282 was protecting against ("a different provocation depending on
-- plan order") is still impossible; it is now scoped to the unit that actually has
-- one daily.
--
-- DEPENDS ON: 282_provocation_pool.sql. Nothing else — `category` already exists,
-- NOT NULL DEFAULT 'general', added early on purpose (PLAN §9.2).
--
-- ROLLBACK NEEDS NO MIGRATION, which is RFC bar 4. The old binary tolerates this
-- schema: today every row is 'general', so (domain, category, publish_on) and
-- (domain, publish_on) admit exactly the same set of rows. The index change is
-- only observable once a second category exists.

BEGIN;

DO $guard$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM information_schema.tables
                   WHERE table_schema = 'public' AND table_name = 'provocations') THEN
        RAISE EXCEPTION '320: the provocations table does not exist — apply 282 first';
    END IF;
    IF EXISTS (SELECT 1 FROM pg_indexes
               WHERE schemaname = 'public' AND indexname = 'idx_provocations_one_per_category_day') THEN
        RAISE EXCEPTION '320: idx_provocations_one_per_category_day already exists — migration already applied';
    END IF;
END
$guard$;

-- Create the replacement BEFORE dropping the incumbent, so there is no window in
-- which a concurrent insert could create the duplicate the old index forbade.
CREATE UNIQUE INDEX idx_provocations_one_per_category_day
    ON provocations (domain, category, publish_on)
    WHERE publish_on IS NOT NULL AND status = 'approved';

DROP INDEX IF EXISTS idx_provocations_one_per_day;

-- The exporter's read path is now filtered by category too, so the covering index
-- should lead with it. Same partial predicate as 282's.
CREATE INDEX IF NOT EXISTS idx_provocations_schedule_by_category
    ON provocations (domain, category, publish_on)
    WHERE status = 'approved' AND publish_on IS NOT NULL;

DROP INDEX IF EXISTS idx_provocations_schedule;

COMMENT ON TABLE provocations IS
    'Pool + schedule for daily provocations. Read by render_provocation_feed, one '
    'feed FILE per category (RFC_013): today = latest approved row for (domain, '
    'category) with publish_on <= current_date; archive = everything approved and '
    'strictly earlier in the same category.';

-- ---------------------------------------------------------------------------
-- Verify the RULE, not merely the object.
--
-- A migration that creates an index and stops has proved that CREATE INDEX works.
-- What matters is the two-sided behaviour: same category same day must still be
-- refused, and different categories same day must now be allowed. Both are
-- INDUCED here inside a savepoint, so a silent failure of either is impossible.
--
-- Written as DO/RAISE rather than a block of SELECTs on purpose: ON_ERROR_STOP
-- ignores a non-empty result set, so a verify block made of SELECTs cannot stop
-- the COMMIT (the fleet has been bitten by exactly that — RFC_006's lane).
-- ---------------------------------------------------------------------------

DO $verify$
DECLARE
    dom       CONSTANT text := '__mig320_probe__';
    d         CONSTANT date := DATE '2000-01-01';
    refused   boolean := false;
    n_allowed int;
BEGIN
    IF EXISTS (SELECT 1 FROM provocations WHERE domain = dom) THEN
        RAISE EXCEPTION '320: probe domain % already present; refusing to test against real data', dom;
    END IF;

    -- (1) Two categories, one date — MUST be allowed now.
    INSERT INTO provocations (domain, slug, category, publish_on, status, title, teaser)
    VALUES (dom, 'p-a', 'general', d, 'approved', 't', 'x'),
           (dom, 'p-b', 'pets',    d, 'approved', 't', 'x');

    SELECT count(*) INTO n_allowed FROM provocations WHERE domain = dom;
    IF n_allowed <> 2 THEN
        RAISE EXCEPTION '320: expected 2 rows across two categories on one date, found %', n_allowed;
    END IF;

    -- (2) Same category, same date — MUST still be refused. If this insert
    --     succeeds, the uniqueness rule has been LOST, which is the failure this
    --     migration could plausibly introduce and the one worth inducing.
    BEGIN
        INSERT INTO provocations (domain, slug, category, publish_on, status, title, teaser)
        VALUES (dom, 'p-c', 'pets', d, 'approved', 't', 'x');
    EXCEPTION WHEN unique_violation THEN
        refused := true;
    END;

    IF NOT refused THEN
        RAISE EXCEPTION '320: a SECOND approved pets provocation on % was accepted — '
                        'the one-per-category-per-day rule is gone', d;
    END IF;

    DELETE FROM provocations WHERE domain = dom;

    -- Leave no probe rows behind under any path.
    IF EXISTS (SELECT 1 FROM provocations WHERE domain = dom) THEN
        RAISE EXCEPTION '320: probe rows survived cleanup';
    END IF;
END
$verify$;

-- Assert the real pool is untouched and still selectable, so a mistake above
-- cannot be mistaken for "no change".
DO $verify_real$
DECLARE
    n int;
BEGIN
    SELECT count(*) INTO n FROM provocations
        WHERE domain = 'vonc.com' AND status = 'approved' AND publish_on IS NOT NULL;
    IF n < 9 THEN
        RAISE EXCEPTION '320: expected at least 9 selectable vonc.com provocations, found %', n;
    END IF;
END
$verify_real$;

COMMIT;
