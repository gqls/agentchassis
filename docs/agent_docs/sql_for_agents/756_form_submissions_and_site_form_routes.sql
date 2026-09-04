-- 756_form_submissions_and_site_form_routes.sql — the form endpoint's storage. DB ONLY, INERT.
--
-- WHY. Every contact form this estate builds delivers by `mailto:` or not at all. The render
-- seam (deliverableFormAction, platform/orchestration/actions/component_library.go) rewrites a
-- non-delivering action to mailto:<sites.email> — the owner's 2026-07-17 pattern — and where no
-- address resolves it leaves the form dead for check_contact_form_undeliverable to raise.
-- Measured 2026-09-04 at the SERVED layer: of 27 contact-form components, 21 serve a real
-- mailto: and 6 on 6 address-less sites still serve "#contact".
--
-- mailto: cannot carry a structured payload, cannot route to a recipient that changes without a
-- rebuild, and cannot be measured. copyonline.co.uk's lead route needs all three
-- (docs024_key_docs_latest/static_site_form_endpoint/CONTRIB_2026-09-03_...named_first_customer.md),
-- and bugs_open/228 is a live component that tells a visitor "your message has been sent" over a
-- setTimeout with no transport at all. Design of record:
-- docs024_key_docs_latest/static_site_form_endpoint/PLAN_2026-09-04_form_endpoint_build.md
--
-- WHAT THIS MIGRATION DOES, and what it deliberately does NOT do.
--   DOES: create site_form_routes (the opt-in switch AND the movable recipient) and
--         form_submissions (the durable record), with their indexes.
--   DOES NOT: enable anything, seed any row, or change any existing table, column, function or
--         agent config. There is no reader and no writer until the Go half ships
--         (internal/tools-api/{handlers,middleware,store}/, phase 2) and the render seam learns
--         the new branch (phase 3). Both are separate commits and take their own council round.
--
-- INERT TWO WAYS. Nothing in the tree references either table name (verified by grep before
-- writing), and site_form_routes.enabled defaults to FALSE — so even a hand-inserted row does
-- nothing until someone deliberately flips it. That is the owner's 2026-08-02 §2 shape for new
-- authority on a shared seam: opt-in, with the UNSAFE side as the default, visible in the config
-- rather than licensed by a comment.
--
-- WHY THE TWO FOREIGN KEYS DIFFER, since this is the one design choice worth arguing with:
--   form_submissions.site_id  -> sites(id)             ON DELETE RESTRICT
--   form_submissions.route_id -> site_form_routes(id)  ON DELETE SET NULL
-- A submission is a business record — a lead someone may have paid for. CASCADE would make
-- deleting a site silently destroy them, and bugs_open/432 records that site rows DO get deleted
-- on this estate. RESTRICT makes that deletion stop and ask. The route, by contrast, is routing
-- config: re-pointing or retiring it must never touch the submissions it already produced, so it
-- nulls out and intent/payload keep the record self-describing.
--
-- Rollback: 756_form_submissions_and_site_form_routes_ROLLBACK.sql (drops both tables; refuses
-- while form_submissions holds any row, because dropping real leads is not a rollback).

BEGIN;

-- ---------------------------------------------------------------- refusals first
DO $$
BEGIN
  IF to_regclass('public.site_form_routes') IS NOT NULL THEN
    RAISE EXCEPTION 'site_form_routes already exists — 756 has run, or another lane created it. Stop and reconcile.';
  END IF;
  IF to_regclass('public.form_submissions') IS NOT NULL THEN
    RAISE EXCEPTION 'form_submissions already exists — 756 has run, or another lane created it. Stop and reconcile.';
  END IF;
  IF NOT EXISTS (SELECT 1 FROM pg_extension WHERE extname = 'pgcrypto') THEN
    RAISE EXCEPTION 'pgcrypto is required for gen_random_bytes (token minting) and is not installed';
  END IF;
END $$;

-- ---------------------------------------------------------------- routing + opt-in
CREATE TABLE site_form_routes (
  id              uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  site_id         uuid        NOT NULL REFERENCES sites(id) ON DELETE CASCADE,

  -- The extensibility axis. copyonline needs two from day one: a commercial enquiry and a
  -- directory removal/correction request, with different recipients and different urgency.
  -- A single-purpose contact endpoint would not have covered it.
  intent          text        NOT NULL,

  -- What the endpoint resolves the site from. NOT the Origin header: an attacker sets Origin
  -- freely, and the existing tools-api middleware (internal/tools-api/middleware/cors.go ->
  -- store.ActiveSiteByOrigin) does exactly that. For rate-limit buckets that is bounded; for an
  -- endpoint that EMAILS someone it is a spam relay wearing the estate's name. The token is
  -- stamped into the form markup by the render seam and is the site's identity at the receiver.
  token           text        NOT NULL DEFAULT encode(gen_random_bytes(32), 'hex'),

  -- Read at DELIVERY time, never baked into the page — this is what lets the destination move
  -- (site owner today, a third party who buys the leads later) with no rebuild.
  recipient_email text        NOT NULL,
  reply_to        text,

  -- The unsafe side is the default. A row alone changes nothing.
  enabled         boolean     NOT NULL DEFAULT false,

  created_at      timestamptz NOT NULL DEFAULT now(),
  updated_at      timestamptz NOT NULL DEFAULT now(),

  CONSTRAINT site_form_routes_site_intent_uniq UNIQUE (site_id, intent),
  CONSTRAINT site_form_routes_token_uniq       UNIQUE (token),
  -- A token short enough to guess is worse than no token. 32 bytes hex = 64 chars; the floor is
  -- set below that so a hand-written test token is possible, but a one-character one is not.
  CONSTRAINT site_form_routes_token_len        CHECK (length(token) >= 32),
  CONSTRAINT site_form_routes_intent_shape     CHECK (intent ~ '^[a-z][a-z0-9_]{1,39}$'),
  CONSTRAINT site_form_routes_recipient_shape  CHECK (position('@' in recipient_email) > 1)
);

COMMENT ON TABLE  site_form_routes IS
  'Per-site, per-intent form routing for the static-site form endpoint. Presence of a row is the opt-in; enabled=false is the default so a row alone does nothing. Read by internal/tools-api (receiver) and by the render seam (which stamps token into the form markup).';
COMMENT ON COLUMN site_form_routes.token IS
  'The site identity at the receiver. Stamped into form markup at render. Never derive the site from the Origin header for a delivering endpoint — Origin is attacker-controlled.';
COMMENT ON COLUMN site_form_routes.recipient_email IS
  'Resolved at delivery time so the destination can move without a rebuild.';

-- ---------------------------------------------------------------- the durable record
CREATE TABLE form_submissions (
  id           uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  site_id      uuid        NOT NULL REFERENCES sites(id)            ON DELETE RESTRICT,
  route_id     uuid                 REFERENCES site_form_routes(id) ON DELETE SET NULL,

  -- Denormalised deliberately: a submission must stay self-describing after its route is
  -- re-pointed or retired, or the record cannot be read back without archaeology.
  intent       text        NOT NULL,

  -- jsonb so a site's field set can grow without a schema change. copyonline's payload is
  -- what-needs-writing / audience / tone / deadline / budget band, not a name-email-message
  -- triple, and the next customer's will differ again.
  payload      jsonb       NOT NULL DEFAULT '{}'::jsonb,

  -- Hashed, never raw: the estate's existing intake handler does the same
  -- (internal/tools-api/handlers/gripper.go, hashIP). Enough to rate-limit and to spot a flood,
  -- not enough to be a personal-data store we did not intend to build.
  ip_hash      text,
  user_agent   text,

  created_at   timestamptz NOT NULL DEFAULT now(),

  -- The notification bookkeeping is the whole reason storage comes FIRST and mail second.
  -- notified_at NULL with attempts > 0 is a send that failed and can be retried; notify_error
  -- keeps why. A channel that drops silently is the defect this lane exists to end, so the
  -- failure has to be a row someone can query, not a log line.
  notified_at    timestamptz,
  notify_attempts integer   NOT NULL DEFAULT 0,
  notify_error    text,

  CONSTRAINT form_submissions_intent_shape CHECK (intent ~ '^[a-z][a-z0-9_]{1,39}$'),
  CONSTRAINT form_submissions_attempts_nonneg CHECK (notify_attempts >= 0)
);

COMMENT ON TABLE form_submissions IS
  'Durable record of a static-site form submission. Written before any notification is attempted; notified_at/notify_attempts/notify_error make a failed send queryable rather than lost.';

CREATE INDEX idx_form_submissions_site_created ON form_submissions (site_id, created_at DESC);
-- Partial: the only interesting scan is "what has not been delivered yet", and it stays small.
CREATE INDEX idx_form_submissions_undelivered  ON form_submissions (created_at)
  WHERE notified_at IS NULL;
CREATE INDEX idx_site_form_routes_enabled      ON site_form_routes (site_id)
  WHERE enabled;

-- ---------------------------------------------------------------- verify, and prove it can fail
-- DO/RAISE, not a SELECT: ON_ERROR_STOP ignores a non-empty result set, so a verify block made
-- of SELECTs cannot stop the COMMIT (CLAUDE.md, RFC_006's landmine). Each assertion below is
-- followed by an INDUCED failure proving the assertion is live and not vacuous.
DO $$
DECLARE
  v_site uuid;
  v_route uuid;
  v_tok  text;
  v_ok   boolean;
BEGIN
  -- Both tables exist.
  IF to_regclass('public.site_form_routes') IS NULL OR to_regclass('public.form_submissions') IS NULL THEN
    RAISE EXCEPTION 'verify: a table is missing after creation';
  END IF;

  -- A real site to hang the fixtures off. Any one will do; nothing is left behind.
  SELECT id INTO v_site FROM sites ORDER BY created_at LIMIT 1;
  IF v_site IS NULL THEN
    RAISE EXCEPTION 'verify: no sites row to test against';
  END IF;

  -- (1) The default is OFF, and the token is minted and long.
  INSERT INTO site_form_routes (site_id, intent, recipient_email)
       VALUES (v_site, 'verify_probe', 'verify@example.com')
    RETURNING id, token INTO v_route, v_tok;

  IF (SELECT enabled FROM site_form_routes WHERE id = v_route) IS DISTINCT FROM false THEN
    RAISE EXCEPTION 'verify: enabled did not default to false — the opt-in default is the wrong way round';
  END IF;
  IF v_tok IS NULL OR length(v_tok) < 64 THEN
    RAISE EXCEPTION 'verify: token was not minted at full length (got %)', coalesce(length(v_tok), -1);
  END IF;

  -- INDUCED CONTROL for (1): the length constraint must actually refuse a short token.
  -- If this UPDATE succeeds, the CHECK is not doing anything and the assertion above proves
  -- nothing about any row a human writes by hand.
  v_ok := false;
  BEGIN
    UPDATE site_form_routes SET token = 'short' WHERE id = v_route;
  EXCEPTION WHEN check_violation THEN
    v_ok := true;
  END;
  IF NOT v_ok THEN
    RAISE EXCEPTION 'verify: the token length CHECK did not fire — constraint is inert';
  END IF;

  -- (2) One route per site+intent.
  v_ok := false;
  BEGIN
    INSERT INTO site_form_routes (site_id, intent, recipient_email)
         VALUES (v_site, 'verify_probe', 'other@example.com');
  EXCEPTION WHEN unique_violation THEN
    v_ok := true;
  END;
  IF NOT v_ok THEN
    RAISE EXCEPTION 'verify: (site_id, intent) is not unique — two routes could claim one form';
  END IF;

  -- (3) A submission survives its route being deleted, and keeps saying what it was.
  INSERT INTO form_submissions (site_id, route_id, intent, payload)
       VALUES (v_site, v_route, 'verify_probe', '{"probe":true}'::jsonb);

  DELETE FROM site_form_routes WHERE id = v_route;

  IF NOT EXISTS (SELECT 1 FROM form_submissions
                  WHERE site_id = v_site AND intent = 'verify_probe' AND route_id IS NULL) THEN
    RAISE EXCEPTION 'verify: deleting a route did not leave its submissions intact and self-describing';
  END IF;

  -- (4) The site FK is live, and its delete action is RESTRICT rather than CASCADE.
  --
  -- Split deliberately into a behavioural half and a declarative half, because the obvious
  -- behavioural test — DELETE the site and require the failure — is not worth its blast radius.
  -- `sites` carries a BEFORE INSERT trigger and a fan of dependent tables, several CASCADEing;
  -- a delete that did NOT refuse would cascade across a live site's rows and fire their triggers
  -- before this block's RAISE rolled it back. Correct, and far more of the schema than a storage
  -- migration has any business exercising.
  --
  -- 4a, BEHAVIOURAL: the FK fires. Induced with a site_id that cannot exist, so no real row is
  -- touched. Without this, 4b would be asserting a constraint that might not be enforced at all.
  v_ok := false;
  BEGIN
    INSERT INTO form_submissions (site_id, intent, payload)
         VALUES ('00000000-0000-0000-0000-000000000000'::uuid, 'verify_probe', '{}'::jsonb);
  EXCEPTION WHEN foreign_key_violation THEN
    v_ok := true;
  END;
  IF NOT v_ok THEN
    RAISE EXCEPTION 'verify: form_submissions.site_id accepted a non-existent site — the FK is not enforced';
  END IF;

  -- 4b, DECLARATIVE: and its ON DELETE action is RESTRICT ('r'), not CASCADE ('c') or SET NULL.
  -- Stated as a declaration check, which is what it is — it reads the catalogue rather than
  -- observing a refusal.
  IF NOT EXISTS (
    SELECT 1 FROM pg_constraint
     WHERE conrelid = 'public.form_submissions'::regclass
       AND confrelid = 'public.sites'::regclass
       AND contype = 'f'
       AND confdeltype = 'r'
  ) THEN
    RAISE EXCEPTION 'verify: form_submissions -> sites is not ON DELETE RESTRICT — deleting a site could destroy leads';
  END IF;

  -- Clean up the fixtures. Nothing this block created may survive.
  DELETE FROM form_submissions WHERE site_id = v_site AND intent = 'verify_probe';
  IF EXISTS (SELECT 1 FROM site_form_routes WHERE intent = 'verify_probe')
     OR EXISTS (SELECT 1 FROM form_submissions WHERE intent = 'verify_probe') THEN
    RAISE EXCEPTION 'verify: probe fixtures were left behind';
  END IF;

  RAISE NOTICE '756 verify: OK — tables created, opt-in defaults OFF, token minted and constrained, route deletion preserves submissions, site deletion refused while submissions exist.';
END $$;

COMMIT;
