-- ============================================================
-- Migration : 276_tools_api_round_publication.sql
-- Purpose   : let a completed gauntlet round be PUBLISHED as a
--             public, linkable record — step 2 of the owner's
--             2026-07-31 share-card ruling ("option 3 staged via 1").
--
-- TARGET    : the ISLAND Postgres, NOT clients_db.
--   docker exec -i <island postgres> psql -U tools_api -d tools_api
--   ledger it in island_migrations (see 198's precedent).
--
-- WHY TWO COLUMNS AND NOT ONE BOOLEAN:
--   * published_at (timestamptz, NULL = not published) records WHEN, which a
--     boolean cannot. The public read gates on IS NOT NULL.
--   * public_slug is a short unguessable token used in the URL instead of the
--     row's primary key. Two reasons: the id should not be the public handle of
--     a record, and the slug has to fit legibly on a 1200x630 share card —
--     `?r=k7m2p9qx4t` does, a 36-character uuid does not.
--
-- CONSENT NOTE (owner ruling 2026-07-31): publication is set by the visitor
-- pressing share, so these columns must default to "not published" and no
-- backfill may set them. Every round already stored is OUR OWN traffic and must
-- stay private — the public record starts empty, deliberately.
--
-- Measured on the island at apply time (2026-07-31 12:45Z): 98 rows, 53 with a
-- verdict, and count(DISTINCT client_ip_hash) = 2 — `245c0ffc…` on 95 rows
-- (2026-07-25..30, the docker-gateway constant from before httpguard was
-- adopted) and `9e464fe9…` on 3 rows (2026-07-31: the httpguard adoption proof
-- plus this lane's own live share-card drive). Two keys, both ours. **No
-- stranger has argued on this page yet**, which is exactly why a backfill would
-- manufacture a crowd out of test data.
--   NB an earlier note in this lane quoted "1 distinct across 95 rows". That was
--   true when measured on 2026-07-30 and went stale the next morning when the
--   island was swapped to v1.0.1207 and the identity fix went live. Re-measure
--   rather than quoting either figure.
-- ============================================================

DO $$
BEGIN
  IF NOT EXISTS (SELECT 1 FROM information_schema.tables
                 WHERE table_schema = 'public'
                   AND table_name   = 'gauntlet_rounds') THEN
    RAISE EXCEPTION 'Migration 276: gauntlet_rounds does not exist — apply 198 first (and check you are on the ISLAND, not clients_db)';
  END IF;

  IF EXISTS (SELECT 1 FROM information_schema.columns
             WHERE table_schema = 'public'
               AND table_name   = 'gauntlet_rounds'
               AND column_name  = 'published_at') THEN
    RAISE NOTICE 'Migration 276: published_at already exists — skipping';
  ELSE
    ALTER TABLE gauntlet_rounds
      ADD COLUMN published_at timestamptz,
      ADD COLUMN public_slug  text;

    -- Unique but nullable: every unpublished round holds NULL, and Postgres
    -- does not treat NULLs as equal, so one partial-free UNIQUE index is
    -- correct here and needs no WHERE clause.
    CREATE UNIQUE INDEX gauntlet_rounds_public_slug_key
      ON gauntlet_rounds (public_slug);

    RAISE NOTICE 'Migration 276: published_at + public_slug added';
  END IF;
END $$;

-- Assert the intended end state rather than trusting the block above: a
-- migration that skipped for the wrong reason looks identical to one that
-- worked. Also asserts the consent invariant — nothing is published yet.
DO $$
DECLARE n_cols int; n_published int;
BEGIN
  SELECT count(*) INTO n_cols FROM information_schema.columns
   WHERE table_schema='public' AND table_name='gauntlet_rounds'
     AND column_name IN ('published_at','public_slug');
  IF n_cols <> 2 THEN
    RAISE EXCEPTION 'Migration 276: expected both columns, found %', n_cols;
  END IF;

  IF NOT EXISTS (SELECT 1 FROM pg_indexes
                 WHERE tablename='gauntlet_rounds'
                   AND indexname='gauntlet_rounds_public_slug_key') THEN
    RAISE EXCEPTION 'Migration 276: unique index on public_slug is missing';
  END IF;

  SELECT count(*) INTO n_published FROM gauntlet_rounds WHERE published_at IS NOT NULL;
  IF n_published <> 0 THEN
    RAISE WARNING 'Migration 276: % round(s) are already published — expected 0 on a fresh apply; do NOT backfill', n_published;
  END IF;

  RAISE NOTICE 'Migration 276 verified: 2 columns, unique index present, % published', n_published;
END $$;

-- Ledger entry
-- table  : gauntlet_rounds (altered)
-- cols   : + published_at timestamptz NULL, + public_slug text NULL
-- index  : gauntlet_rounds_public_slug_key UNIQUE (public_slug)
-- reads  : store.GetPublishedRound (by slug, gated on published_at IS NOT NULL)
-- writes : store.PublishRound (idempotent; only a round with a verdict)
