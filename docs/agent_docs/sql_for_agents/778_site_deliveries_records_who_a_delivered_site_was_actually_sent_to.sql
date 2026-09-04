-- FILE: docs/agent_docs/sql_for_agents/778_site_deliveries_records_who_a_delivered_site_was_actually_sent_to.sql
--
-- bugs_open/477: the estate has no durable record of who a delivered site was
-- delivered to. This is that record, plus a backfill of the ONE delivery in the
-- estate's history — which is time-critical, see the deadline below.
--
-- ⚠⚠ THE OBVIOUS COLUMN IS POPULATED AND WRONG, WHICH IS WORSE THAN EMPTY.
-- `[MEASURED 2026-09-04]` idea.uk — the only site ever handed over — carries
-- `sites.email = 'idea.uk@contactforsales.com'`, a SITE MAILBOX. The delivery
-- actually went to `aaa@designconsultancy.co.uk`. So a person answering a support
-- or refund question finds a well-formed, plausible address and is confidently
-- misled; there is no NULL to warn them. That is the reason this table exists
-- rather than a documented convention about which column to read. (Since
-- bugs_open/420's contract split, `sites.email` is the PUBLISHED contact only —
-- correct by its own definition, and never the customer.)
--
-- ⚠ THE BACKFILL EXPIRES TODAY. The only machine-readable copy of that address is
-- the delivery run's own row in `orchestration_states`, and that table is a QUEUE,
-- not a history. The retention policy is in `sql_for_agents/466`:
--   DELETE ... WHERE status IN ('COMPLETED','FAILED') AND updated_at < now() - '24 hours'
-- The idea.uk delivery row has `updated_at = 2026-09-03 19:30:40Z`, so it is reaped
-- after **2026-09-04 19:30:40Z**. Applied at 14:50Z with 4h40m to spare. Applied
-- tomorrow it captures NOTHING and the address has to be typed in by a human
-- trusting a document.
--
-- > ⚠ CORRECTED before this file was applied. An earlier draft of this header said
-- > "its oldest row was 1 day 02:11 old … roughly SEVEN HOURS left", measured with
-- > `now() - min(created_at)`. **That does not measure retention** — it reports the
-- > oldest SURVIVOR's birthday, while the policy keys on `updated_at`. The true
-- > margin at that moment was 5h31m, not ~7h. For a question about one row
-- > surviving, ask that row:
-- >   SELECT updated_at + interval '24 hours' - now() FROM orchestration_states WHERE ...
--
-- WHY A TABLE AND NOT `sites.delivered_to` — owner ruling, relayed 2026-09-04 by
-- the site_delivery_and_editor lane: a dedicated record, written in the same
-- statement as `handed_over_at`. The reason is the one that made me route the
-- question rather than answer it: `sites` is read by a great many things, and
-- bugs_open/420's whole subject is controlling which address may live where.
-- Customer PII does not belong on the widest-read table in the estate.
--
-- WHAT WRITES IT: platform/delivery.Claim, inside the handover claim, so the
-- record inherits the once-only guarantee for free (the claim can be won by
-- exactly one statement, so a retry or a second operator cannot overwrite the
-- recipient). That code ships in the same commit as this file and is INERT until
-- a chassis image rolls — which is why the backfill below is not optional.
--
-- Council: submitted with this round. Rollback sidecar: 778_..._ROLLBACK.sql.

BEGIN;

CREATE TABLE IF NOT EXISTS site_deliveries (
  site_id       uuid PRIMARY KEY REFERENCES sites(id) ON DELETE CASCADE,
  delivered_to  text        NOT NULL,
  delivered_at  timestamptz NOT NULL,
  -- Provenance, because a backfilled row and a row written by the delivery
  -- itself are not equally trustworthy and a reader must be able to tell.
  recorded_by   text        NOT NULL,
  created_at    timestamptz NOT NULL DEFAULT now(),
  CONSTRAINT site_deliveries_delivered_to_not_blank CHECK (btrim(delivered_to) <> '')
);

-- site_id is the PRIMARY KEY, and that is the once-only property restated: a site
-- is delivered once, so it has at most one recipient. A second delivery to a
-- different address is not an UPDATE, it is a fact that needs a decision.
COMMENT ON TABLE site_deliveries IS
  'bugs_open/477: who a delivered site was actually sent to. Written by platform/delivery.Claim '
  'inside the handover claim, so it inherits that claim''s once-only guarantee. Exists because '
  'sites.email is the PUBLISHED contact (bugs_open/420) and is populated with a DIFFERENT, '
  'plausible, wrong address on the estate''s only delivery — a misleading value, not an absent one.';
COMMENT ON COLUMN site_deliveries.recorded_by IS
  'Provenance: ''delivery-email'' when written by the delivery claim itself; '
  '''backfill-orchestration-states-778'' for the one row recovered from the run log before it aged out.';

-- ── THE BACKFILL ────────────────────────────────────────────────────────────
--
-- ⚠ THIS MIGRATION TAKES ~2 MINUTES, and the reason is worth knowing before you
-- "fix" it: `collected_data->'input_data' ? 'customer_email'` cannot use an index
-- and every row of `orchestration_states` is a large jsonb document.
-- `[MEASURED 2026-09-04]` ONE such scan over 6,662 rows = **102 seconds**.
--
-- So the scan happens EXACTLY ONCE, into a temp table, and both the insert and
-- the verify read that. The first draft of this file used a correlated LATERAL
-- per site and an independent EXISTS per site — three scans' worth against 60
-- sites — and it did not finish; it read as a hung psql rather than as a slow
-- query, which is its own small trap.
CREATE TEMP TABLE _778_delivery_runs ON COMMIT DROP AS
SELECT DISTINCT ON (collected_data->'input_data'->>'site_id')
       (collected_data->'input_data'->>'site_id')                     AS site_id,
       btrim(collected_data->'input_data'->>'customer_email')         AS customer_email,
       created_at                                                     AS run_at
  FROM orchestration_states
 WHERE collected_data->'input_data' ? 'customer_email'
 ORDER BY 1, created_at DESC;   -- newest run per site wins

-- Idempotent: ON CONFLICT DO NOTHING, so re-applying can never overwrite a row
-- the delivery itself wrote with a backfilled one.
INSERT INTO site_deliveries (site_id, delivered_to, delivered_at, recorded_by)
SELECT s.id, r.customer_email, s.handed_over_at, 'backfill-orchestration-states-778'
  FROM sites s
  JOIN _778_delivery_runs r ON r.site_id = s.id::text
 WHERE s.handed_over_at IS NOT NULL
   AND r.customer_email <> ''
ON CONFLICT (site_id) DO NOTHING;

-- ── VERIFY ──────────────────────────────────────────────────────────────────
-- A DO block, not SELECTs: ON_ERROR_STOP ignores a non-empty result set, so a
-- SELECT-based verify reports and commits anyway. RAISE is what aborts.
--
-- The design point: a backfill capturing ZERO must be DISTINGUISHABLE from one
-- that worked. Silently capturing nothing is the exact failure this whole bug is
-- about — a mechanism that looks healthy and selects nobody — so this asks
-- whether the SOURCE ROW STILL EXISTS and reacts differently to the two cases. A
-- zero because the window aged out is a (loud) fact about the world; a zero while
-- the source is sitting there is MY BUG and aborts.
DO $$
DECLARE
  handed      int;
  captured    int;
  recoverable int;
BEGIN
  SELECT count(*) INTO handed FROM sites WHERE handed_over_at IS NOT NULL;
  SELECT count(*) INTO captured FROM site_deliveries;

  SELECT count(*) INTO recoverable
    FROM sites s JOIN _778_delivery_runs r ON r.site_id = s.id::text
   WHERE s.handed_over_at IS NOT NULL AND r.customer_email <> '';

  IF captured < recoverable THEN
    RAISE EXCEPTION '778 FAILED: % handed-over site(s) still have a recoverable recipient in the run log but only % row(s) exist. The backfill is wrong — do NOT retry blind, the source window is closing.', recoverable, captured;
  END IF;

  IF handed > 0 AND captured = 0 THEN
    -- Not an abort: the table is worth having regardless, and this is a true
    -- statement about the world rather than a defect. But it must be LOUD,
    -- because it means the addresses now exist only in prose.
    RAISE WARNING '778: % handed-over site(s) and ZERO recipients captured. The orchestration_states window has aged out — each recipient is now only in bugs_open/477 and this file, and must be entered by hand after confirming it against those documents.', handed;
  END IF;

  RAISE NOTICE '778 OK: site_deliveries created. handed over: %, recoverable from the run log: %, captured: %.', handed, recoverable, captured;
END $$;

COMMIT;
