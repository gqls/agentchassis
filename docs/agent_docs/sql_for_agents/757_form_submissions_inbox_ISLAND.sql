-- 757 — the form endpoint's INBOX on the ISLAND: what the browser posted, before anyone believes it.
--
-- WHY THE `_ISLAND` SUFFIX. scripts/migration/run-migrations.sh sweeps this directory against
-- clients_db and treats any UPPERCASE-suffixed file as a sidecar it reports and never applies
-- (SIDECAR_RE='_[A-Z][A-Z0-9_]*\.sql$'). A plain-named island file IS picked up and applied to the
-- wrong database — 198_tools_api_gauntlet_rounds.sql is in clients_db's schema_migrations and
-- `gauntlet_rounds` exists there as a result. This CREATE TABLE would have succeeded in clients_db
-- too, leaving a second, permanently empty inbox that no process writes and that reads exactly like
-- the real one. The suffix is the only thing that keeps the runner off. (737's header, same trap.)
--
-- TARGET    : the ISLAND Postgres (toolsapisuk.vs.mythic-beasts.com), NOT clients_db.
--             Same target and ledger as 198 / 276 / 436 / 737:
--   cd /opt/island && docker compose exec -T postgres psql -U tools_api -d tools_api -v ON_ERROR_STOP=1 < this_file.sql
--   then ledger it: INSERT INTO island_migrations (filename, note) VALUES ('757_form_submissions_inbox', '<what you checked>');
--   (columns are filename/note/applied_at, and rows store the name WITHOUT the .sql suffix —
--   the RUNBOOK_island.md correction of 2026-08-25.)
--
-- WHY THIS TABLE EXISTS, AND WHY IT IS DELIBERATELY CREDULOUS.
--
-- The island runs its own Postgres. It cannot see clients_db, so it cannot resolve which site a
-- submission belongs to, cannot read a recipient, and cannot send anything. That is not a
-- limitation this table works around — it is the security property the design now rests on:
--
--   * the RECEIVER records what it was handed, including the token EXACTLY AS PRESENTED;
--   * the CLUSTER pulls, and resolves that token against site_form_routes (migration 756) —
--     a table the island has no access to — and only then does anything reach a mailbox.
--
-- So a forged token is stored here and then discarded at ingest. It can never be delivered. Had
-- identity been resolved at the edge (the shape this lane submitted to council before finding the
-- database split), a forged Origin or a stolen token would have been believed by the process that
-- also does the sending.
--
-- It is also the estate's proven shape rather than a new one: gripper_report_requests + the
-- cluster's poller, and the shopfront's /internal/orders + order-intake-collector, are both
-- "receiver stores, cluster pulls". The publish-seam reviewer asked for exactly this property to be
-- carried across whatever D1 decided — the cluster exposes no inbound surface, and receipt survives
-- cluster downtime, because a submission waits here until the collector comes back.
--
-- NOT VALIDATED HERE, ON PURPOSE: token authenticity, whether the intent exists, whether the site
-- is enabled, whether the payload has the fields the site's form asks for. Every one of those is a
-- clients_db question. Validating any of them here would either be wrong (no data) or would move
-- the trust decision back to the machine that must not make it.
--
-- Rollback: 757_form_submissions_inbox_ISLAND_ROLLBACK.sql (drops the table; refuses while any row
-- is still 'pending', because an unpulled row is a submission nobody has received yet).

BEGIN;

DO $$
BEGIN
  IF to_regclass('public.form_submissions_inbox') IS NOT NULL THEN
    RAISE EXCEPTION '757: form_submissions_inbox already exists — already applied, or another lane created it. Stop and reconcile.';
  END IF;
  -- Guard against exactly the accident the _ISLAND suffix exists to prevent. site_form_routes is
  -- a clients_db table (756); if it is visible, this is clients_db and NOT the island.
  IF to_regclass('public.site_form_routes') IS NOT NULL THEN
    RAISE EXCEPTION '757: site_form_routes is visible, so this is clients_db, NOT the island. This file targets the ISLAND Postgres — see the header.';
  END IF;
END $$;

CREATE TABLE form_submissions_inbox (
  id          uuid PRIMARY KEY DEFAULT gen_random_uuid(),

  -- The token AS PRESENTED by the browser. Untrusted, unresolved, unindexed for lookup — this
  -- process has nothing to check it against. The cluster resolves it at ingest.
  token       text        NOT NULL,

  -- The intent AS PRESENTED, likewise. Shape-checked only so a hostile value cannot make the
  -- cluster's logs or queries awkward; whether the intent EXISTS is a clients_db question.
  intent      text        NOT NULL,

  -- What the visitor typed. jsonb because each site's form asks for different things.
  payload     jsonb       NOT NULL DEFAULT '{}'::jsonb,

  -- What CORSMiddleware resolved from the Origin header against the island's own minimal `sites`
  -- mirror. Recorded as a CROSS-CHECK and a debugging aid, never as identity: Origin is
  -- attacker-controlled, which is the whole reason the token exists. The collector compares the
  -- two and can flag a disagreement; it trusts only the token.
  site_id     uuid,
  site_domain text,

  ip_hash     text,
  user_agent  text,
  created_at  timestamptz NOT NULL DEFAULT now(),

  -- pending -> pulled, moved by a guarded UPDATE from GET /forms/requests, the same way
  -- gripper_report_requests moves. The row is kept after pulling so a re-pull is detectable and
  -- the island retains a short receipt of what it accepted.
  status      text        NOT NULL DEFAULT 'pending',
  pulled_at   timestamptz,

  CONSTRAINT form_submissions_inbox_status_known CHECK (status IN ('pending','pulled')),
  CONSTRAINT form_submissions_inbox_intent_shape CHECK (intent ~ '^[a-z][a-z0-9_]{1,39}$'),
  CONSTRAINT form_submissions_inbox_token_len    CHECK (length(token) BETWEEN 32 AND 256),
  CONSTRAINT form_submissions_inbox_pulled_time  CHECK ((status = 'pulled') = (pulled_at IS NOT NULL))
);

COMMENT ON TABLE form_submissions_inbox IS
  'Island-side inbox for static-site form submissions. Deliberately credulous: it records the token as presented and resolves nothing. The cluster pulls and resolves against site_form_routes (clients_db, migration 756), so a forged token is stored here and discarded at ingest, never delivered.';
COMMENT ON COLUMN form_submissions_inbox.site_id IS
  'Origin-derived, from the island sites mirror. A cross-check and a debugging aid — NEVER identity. Origin is attacker-controlled; the token is what the cluster trusts.';

-- The only hot query: "what has not been pulled yet", oldest first.
CREATE INDEX idx_form_submissions_inbox_pending ON form_submissions_inbox (created_at)
  WHERE status = 'pending';

-- Verify. DO/RAISE, not SELECTs: ON_ERROR_STOP ignores a non-empty result set, so a SELECT-based
-- verify cannot stop the COMMIT. Each assertion is paired with an induced failure so that none of
-- them can pass vacuously.
DO $$
DECLARE
  v_id uuid;
  v_ok boolean;
BEGIN
  IF to_regclass('public.form_submissions_inbox') IS NULL THEN
    RAISE EXCEPTION 'verify: table missing after creation';
  END IF;

  -- (1) A row lands pending, with no pulled_at.
  INSERT INTO form_submissions_inbox (token, intent, payload)
       VALUES (repeat('v', 64), 'verify_probe', '{"probe":true}'::jsonb)
    RETURNING id INTO v_id;

  IF (SELECT status FROM form_submissions_inbox WHERE id = v_id) <> 'pending' THEN
    RAISE EXCEPTION 'verify: a new row did not default to pending';
  END IF;

  -- (2) INDUCED: status and pulled_at cannot disagree. Without this the collector could mark a
  --     row pulled and leave no timestamp, and a re-pull would be undetectable.
  v_ok := false;
  BEGIN
    UPDATE form_submissions_inbox SET status = 'pulled' WHERE id = v_id;   -- pulled_at still NULL
  EXCEPTION WHEN check_violation THEN
    v_ok := true;
  END;
  IF NOT v_ok THEN
    RAISE EXCEPTION 'verify: status=pulled was accepted with a NULL pulled_at — the pair CHECK is inert';
  END IF;

  -- (3) INDUCED: an unknown status is refused.
  v_ok := false;
  BEGIN
    UPDATE form_submissions_inbox SET status = 'delivered', pulled_at = now() WHERE id = v_id;
  EXCEPTION WHEN check_violation THEN
    v_ok := true;
  END;
  IF NOT v_ok THEN
    RAISE EXCEPTION 'verify: an unknown status was accepted — this table must not grow a vocabulary by accident';
  END IF;

  -- (4) INDUCED: a token too short to be one of ours is refused at the door.
  v_ok := false;
  BEGIN
    INSERT INTO form_submissions_inbox (token, intent) VALUES ('short', 'verify_probe');
  EXCEPTION WHEN check_violation THEN
    v_ok := true;
  END;
  IF NOT v_ok THEN
    RAISE EXCEPTION 'verify: the token length CHECK did not fire — constraint is inert';
  END IF;

  -- (5) The legitimate transition works, so (2) and (3) are refusing the wrong thing, not everything.
  UPDATE form_submissions_inbox SET status = 'pulled', pulled_at = now() WHERE id = v_id;
  IF (SELECT pulled_at FROM form_submissions_inbox WHERE id = v_id) IS NULL THEN
    RAISE EXCEPTION 'verify: the legitimate pending->pulled transition did not apply';
  END IF;

  DELETE FROM form_submissions_inbox WHERE intent = 'verify_probe';
  IF EXISTS (SELECT 1 FROM form_submissions_inbox WHERE intent = 'verify_probe') THEN
    RAISE EXCEPTION 'verify: probe fixtures were left behind';
  END IF;

  RAISE NOTICE '757 verify: OK — inbox created, rows land pending, status/pulled_at pair enforced, unknown status refused, short token refused, legitimate transition applies.';
END $$;

COMMIT;
