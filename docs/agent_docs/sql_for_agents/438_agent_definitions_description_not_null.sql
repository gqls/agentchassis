-- 438_agent_definitions_description_not_null.sql
--
-- OWNER DECISION 2026-08-16 (bugs_open/287): agent_definitions.description becomes
-- NOT NULL DEFAULT ''. This closes the door the 287 code fix only guards: a seed
-- that omits `description` now gets '' instead of NULL, so it is spawnable and
-- resolvable on EVERY binary — including the ones running today, which still scan
-- the column into a plain Go string and die on NULL ("converting NULL to string is
-- unsupported", the exact failure that hid brief-fidelity-auditor from spawn_agent
-- until 2026-08-16). DDL is live immediately; the COALESCE code half (same bug)
-- rides the next roll and stays as belt-and-braces.
--
-- WHY IT IS SAFE TO ADD NOW. [MEASURED 2026-08-16, whole table, no filters]
-- 209 rows, 3 NULL — all three inactive `diagnose-wiring-probe-*` scratch rows
-- (the only live NULL was fixed by 420). They are backfilled to '' below in the
-- same transaction, BEFORE the constraint, so ALTER cannot fail on them. Writers
-- checked: both snapshot_agent overloads copy description (a non-null source
-- yields a non-null copy); the admin create handler binds a Go string (never
-- NULL); discovery_actions.go's variant clone writes `description || ' [Variant…]'`,
-- which is NULL only when the SOURCE is NULL — and after this constraint no
-- source can be. Seeds that omit the column now take the DEFAULT.
--
-- ROLLBACK: 438_..._ROLLBACK.sql drops the DEFAULT and the NOT NULL. It does NOT
-- restore the three '' values to NULL — there is nothing to restore; '' and NULL
-- meant the same thing to every reader except the one that crashed.

BEGIN;

DO $$
DECLARE n_null integer; n_live_null integer; already text;
BEGIN
    SELECT is_nullable INTO already FROM information_schema.columns
     WHERE table_schema='public' AND table_name='agent_definitions' AND column_name='description';
    IF already = 'NO' THEN
        RAISE EXCEPTION 'MIGRATION 438: agent_definitions.description is already NOT NULL — already applied';
    END IF;

    SELECT count(*) INTO n_null FROM agent_definitions WHERE description IS NULL;
    SELECT count(*) INTO n_live_null FROM agent_definitions
     WHERE description IS NULL AND is_active AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;
    -- A LIVE NULL means an agent is currently unspawnable; backfilling it to ''
    -- here would silently "fix" it without anyone writing the description it
    -- deserves (that is what 420 did deliberately, with a real description).
    -- Refuse so the row gets a proper description in its own migration first.
    IF n_live_null > 0 THEN
        RAISE EXCEPTION 'MIGRATION 438: % LIVE definition(s) have NULL description — give each a real description (as 420 did) before applying this constraint', n_live_null;
    END IF;
    RAISE NOTICE 'migration 438: backfilling % inactive NULL description(s) to '''' before the constraint', n_null;
END $$;

UPDATE agent_definitions SET description = '' WHERE description IS NULL;

ALTER TABLE agent_definitions
    ALTER COLUMN description SET DEFAULT '',
    ALTER COLUMN description SET NOT NULL;

DO $$
DECLARE nullable text; dflt text; n_null integer;
BEGIN
    SELECT is_nullable, column_default INTO nullable, dflt FROM information_schema.columns
     WHERE table_schema='public' AND table_name='agent_definitions' AND column_name='description';
    SELECT count(*) INTO n_null FROM agent_definitions WHERE description IS NULL;
    IF nullable <> 'NO' OR dflt IS NULL OR n_null <> 0 THEN
        RAISE EXCEPTION 'MIGRATION 438: post-condition failed (is_nullable=%, default=%, nulls=%)', nullable, dflt, n_null;
    END IF;
    RAISE NOTICE 'migration 438 OK: description NOT NULL DEFAULT '''' (default=%)', dflt;
END $$;

COMMIT;
