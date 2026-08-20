-- 511_handover_state_and_customer_tokens.sql
--
-- Phase 4 of the site-delivery plan: the HANDOVER STATE, and the token table the
-- customer-facing links hang off. Design + owner decisions:
--   docs024_key_docs_latest/site_delivery_and_editor/PLAN_2026-08-14_site_delivery_and_editor.md  (§Phase 4/5 mechanics)
--   docs024_key_docs_latest/site_delivery_and_editor/PLAN_2026-08-17_delivery_architecture_decisions.md  (decision 3)
--   docs024_key_docs_latest/webdesign_uk_build_service/HANDOFF_2026-08-20_continue_here.md  (§2, and the presign ceiling box)
--
-- WHAT THIS ADDS, and nothing else. It is STATE ONLY: no enforcement, no email,
-- no HTTP surface. Those are the next commits, and the register entry says so.
--
--   sites.handed_over_at        when the finished site was handed to the customer.
--                              Phase 5's editor gate is its single reader.
--   sites.live_link_expires_at  when the address WE host it at stops serving.
--                              Owner ruling 2026-08-19: six weeks after handover.
--   sites.transfer_confirmed_at when the customer confirmed they have moved their
--                              files off us. Owner ruling 2026-08-19, scoped to the
--                              simplest possible thing: "their confirmation by
--                              clicking a link that we record." The click IS the
--                              state. No form, no reply parsing.
--   customer_access_tokens      one hashed, expiring, optionally single-use token
--                              per customer-facing link.
--
-- WHY ONE TOKEN TABLE AND NOT TWO. Two customer links are needed immediately and
-- they are the same mechanism: the ZIP download and the confirm-transfer click.
-- A third is coming (Phase 5's editor session exchange). Per the owner ruling of
-- 2026-08-02 §1, converging producers onto one key shape does not need an RFC
-- PROVIDED the producer set is named and the key shape is stated -- so both are
-- stated here and in the register entry. `purpose` is a CLOSED vocabulary,
-- enforced by a CHECK: widening it costs a migration, which is the point. A
-- fourth purpose should be visible in the ledger, not appear in a Go constant.
--
-- ⚠ WHY THE ZIP DOWNLOAD NEEDS A TOKEN AT ALL, rather than a longer presign.
-- `[MEASURED 2026-08-20]` against the live bucket: a presigned URL is capped by
-- the SigV4 signing protocol at 604,800 seconds (7 days), the cap is enforced by
-- B2 and NOT by the SDK, and a longer one mints cleanly and then fails as
-- `SignatureDoesNotMatch` -- which reads as broken credentials. 604800 -> HTTP 404
-- NoSuchKey (signature accepted, the control); 604801 -> 403; 6 weeks -> 403.
-- Exact to the second. So "the link lasts as long as we host the site" (owner,
-- 2026-08-20) is only deliverable by handing out a token of OURS and minting a
-- fresh short presign per click. Full entry in LANDMINES.md.
--
-- THE PLAINTEXT TOKEN IS NEVER STORED. `token_hash` is sha256 hex of the
-- plaintext; the plaintext exists only in the email. A leaked database therefore
-- does not yield working links. Redemption looks the row up BY HASH.
--
-- SINGLE-USE IS ENFORCED TWICE, ON PURPOSE: the redeeming UPDATE carries
-- `used_at IS NULL` so a second click is a clean miss the handler can report, and
-- the CHECK constraint is a backstop that fails loudly if some future writer
-- forgets the predicate. The check alone would turn an ordinary second click into
-- a constraint error, which is why it is not the primary control.
--
-- NOT BUILT HERE, and stated so nobody reads the columns as a working mechanism:
--   * nothing enforces live_link_expires_at. Serving is unbounded today -- sites
--     serve from a git repo synced to B2, with no scheduled retraction, no
--     retention job and no TTL (checked 2026-08-19). The column records the
--     intent; the retraction job is a separate build.
--   * nothing mints or redeems these tokens yet. The Go helper lands in the same
--     commit as this migration; the HTTP surface and the delivery email do not.

BEGIN;

-- Refuse a double apply, loudly, before touching anything.
DO $$
BEGIN
  IF EXISTS (SELECT 1 FROM information_schema.columns
              WHERE table_name='sites' AND column_name='handed_over_at') THEN
    RAISE EXCEPTION '511: already applied - sites.handed_over_at exists';
  END IF;
  IF EXISTS (SELECT 1 FROM information_schema.tables
              WHERE table_schema='public' AND table_name='customer_access_tokens') THEN
    RAISE EXCEPTION '511: already applied - customer_access_tokens exists';
  END IF;
END $$;

ALTER TABLE sites
  ADD COLUMN handed_over_at        timestamptz,
  ADD COLUMN live_link_expires_at  timestamptz,
  ADD COLUMN transfer_confirmed_at timestamptz;

COMMENT ON COLUMN sites.handed_over_at IS
  'When the finished site was handed to the customer (Phase 4). Gates the Phase 5 editor session and nothing else: it does NOT gate deploys, rewrites, locks or reconciliation.';
COMMENT ON COLUMN sites.live_link_expires_at IS
  'When the address WE host the site at stops serving. Owner ruling 2026-08-19: handover + 6 weeks. NOTHING ENFORCES THIS YET - serving is unbounded and the retraction job is unbuilt.';
COMMENT ON COLUMN sites.transfer_confirmed_at IS
  'When the customer confirmed they had moved their files off our hosting, by clicking the tokenised link in a chase email. Owner ruling 2026-08-19: the click IS the state.';

CREATE TABLE customer_access_tokens (
  id          uuid        PRIMARY KEY DEFAULT gen_random_uuid(),
  site_id     uuid        NOT NULL REFERENCES sites(id) ON DELETE CASCADE,
  purpose     text        NOT NULL,
  token_hash  text        NOT NULL,
  issued_at   timestamptz NOT NULL DEFAULT now(),
  expires_at  timestamptz NOT NULL,
  single_use  boolean     NOT NULL DEFAULT false,
  used_at     timestamptz,
  use_count   integer     NOT NULL DEFAULT 0,
  revoked_at  timestamptz,
  created_by  text        NOT NULL DEFAULT '',
  CONSTRAINT customer_access_tokens_hash_uniq UNIQUE (token_hash),
  CONSTRAINT customer_access_tokens_purpose_chk
    CHECK (purpose IN ('zip_download','confirm_transfer')),
  CONSTRAINT customer_access_tokens_single_use_chk
    CHECK (NOT single_use OR use_count <= 1),
  CONSTRAINT customer_access_tokens_window_chk
    CHECK (expires_at > issued_at)
);

COMMENT ON TABLE customer_access_tokens IS
  'One hashed, expiring token per customer-facing link. Purposes are a CLOSED vocabulary (zip_download, confirm_transfer today; Phase 5 editor_session next) so widening costs a migration and stays visible. The plaintext is never stored - token_hash is sha256 hex and redemption looks up BY HASH.';

CREATE INDEX idx_cat_site_purpose ON customer_access_tokens (site_id, purpose)
  WHERE revoked_at IS NULL;
CREATE INDEX idx_cat_expires ON customer_access_tokens (expires_at)
  WHERE used_at IS NULL AND revoked_at IS NULL;

DO $$
DECLARE n int;
BEGIN
  -- The three columns exist, are nullable, and are timestamptz.
  SELECT count(*) INTO n FROM information_schema.columns
   WHERE table_name='sites'
     AND column_name IN ('handed_over_at','live_link_expires_at','transfer_confirmed_at')
     AND data_type='timestamp with time zone' AND is_nullable='YES';
  IF n <> 3 THEN RAISE EXCEPTION '511: expected 3 nullable timestamptz columns on sites, found %', n; END IF;

  -- No site is accidentally marked handed over by this migration.
  SELECT count(*) INTO n FROM sites WHERE handed_over_at IS NOT NULL;
  IF n <> 0 THEN RAISE EXCEPTION '511: % site(s) already stamped handed_over_at - this migration must not stamp any', n; END IF;

  -- The table, its four constraints and its two indexes.
  SELECT count(*) INTO n FROM information_schema.table_constraints
   WHERE table_name='customer_access_tokens'
     AND constraint_name IN ('customer_access_tokens_hash_uniq',
                             'customer_access_tokens_purpose_chk',
                             'customer_access_tokens_single_use_chk',
                             'customer_access_tokens_window_chk');
  IF n <> 4 THEN RAISE EXCEPTION '511: expected 4 named constraints, found %', n; END IF;
  SELECT count(*) INTO n FROM pg_indexes
   WHERE tablename='customer_access_tokens' AND indexname IN ('idx_cat_site_purpose','idx_cat_expires');
  IF n <> 2 THEN RAISE EXCEPTION '511: expected 2 partial indexes, found %', n; END IF;

  -- INDUCE the purpose CHECK rather than trusting that it exists. A constraint
  -- that has never refused anything is a constraint nobody has tested.
  BEGIN
    INSERT INTO customer_access_tokens (site_id, purpose, token_hash, expires_at)
    SELECT id, 'not_a_real_purpose', 'probe-'||id::text, now() + interval '1 day' FROM sites LIMIT 1;
    RAISE EXCEPTION '511: the purpose CHECK did not refuse an unknown purpose';
  EXCEPTION WHEN check_violation THEN
    NULL;  -- refused, as it must
  END;

  -- And induce the window CHECK the same way.
  BEGIN
    INSERT INTO customer_access_tokens (site_id, purpose, token_hash, issued_at, expires_at)
    SELECT id, 'zip_download', 'probe2-'||id::text, now(), now() - interval '1 second' FROM sites LIMIT 1;
    RAISE EXCEPTION '511: the window CHECK allowed expires_at <= issued_at';
  EXCEPTION WHEN check_violation THEN
    NULL;
  END;

  SELECT count(*) INTO n FROM customer_access_tokens;
  IF n <> 0 THEN RAISE EXCEPTION '511: the probes left % row(s) behind', n; END IF;
END $$;

COMMIT;
