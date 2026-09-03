-- 651 (_HOLD): the two agents completing Phase 4's delivery chain — the
-- review-item producer and the delivery-email sender.
--
-- ⚠ _HOLD BECAUSE OF ORDERING, the image-before-seeds rule: delivery-email-sender
-- names the `send_delivery_email` action, which exists only from the roll carrying
-- it — a seed naming an unregistered action fails at runtime. Apply BY HAND after:
--   kubectl -n ai-persona-system logs -l app=agent-chassis --tail=300 | grep -m1 'build provenance'
--   git merge-base --is-ancestor <the commit adding send_delivery_email> <the stamp>
-- (delivery-review-filer only names create_work_item, which is long live — but the
-- two agents are one product chain and apply together.)
--
-- ⚠ ALSO REQUIRED BEFORE THE FIRST REAL SEND, neither carried by this file:
--   1. Migration 650 applied (customer_access_tokens.stored_url).
--   2. DELIVERY_SMTP_* env + secret on the chassis pods:
--        DELIVERY_SMTP_HOST=mail.contactforsales.com  DELIVERY_SMTP_PORT=465
--        DELIVERY_SMTP_USER=webdesign@contactforsales.com
--        DELIVERY_SMTP_FROM=webdesign@contactforsales.com
--        DELIVERY_SMTP_PASS -> secretKeyRef ONLY (the owner holds the password;
--        it is never a value in any file). Port 465 is implicit TLS and
--        mailer.UsesImplicitTLS already handles it. The action constructs the
--        sender BEFORE the once-only claim, so a missing secret fails loudly
--        with nothing stamped.
--   3. DKIM + DMARC enabled at the HOST (cPanel -> Email Deliverability;
--      absent as of 2026-08-26) or customer mail lands in junk.
--
-- ============================ THE FLOW, END TO END ==========================
--
--  (1) Site ready -> operator dispatches delivery-review-filer {site_id, domain,
--      site_url, brief}. It files ONE needs_delivery_review item, parked at
--      needs_human_review with spec.checkpoint=true — exactly what
--      HandleApproveWorkItem demands (it 400s anything else, and its error
--      steers toward the WRONG button: resolve writes resolved_by, which the
--      gate deliberately ignores. APPROVE is the only button that opens it).
--  (2) The OWNER reviews — and since the 2026-08-26 ruling, EDITS — the site,
--      then presses approve on the item (admin.apis.uk). Internal only: the
--      customer-facing position stays one-shot, no approval stage.
--  (3) Operator cuts the ZIP (zip-deliverable-dispatch {domain}, recipes in
--      459's header) and notes presigned_url + expiry_minutes from its output.
--  (4) Operator dispatches delivery-email-sender {site_id, customer_email,
--      live_site_url, zip_presigned_url, zip_presign_minutes}. The action:
--      ⚠ customer_email SOURCE (corrected 2026-08-31, bugs_open/420): take it from
--        SELECT direction->>'customer_email' FROM build_queue WHERE domain=$1
--      — NEVER from sites.email. Since the 420 contract split, sites.email is the
--      PUBLISHED contact only (empty unless the customer explicitly asked to
--      publish one), so on a post-420 site it is legitimately NULL and was never
--      read by this action's code anyway (input_data is the only source; the old
--      sites.email wording in this header was convention, not code).
--      sender first (fail-before-stamp), template-vs-links check (fail-before-
--      stamp), then delivery.Claim = review gate + ONCE-ONLY handover stamp +
--      token minting, then the send. A second dispatch for the same site is
--      REFUSED by the stamp: no retry can double-email a customer.
--
-- Dispatch (both agents; use scripts/kafka-publish-lib.sh — OPP-009 — and CHECK
-- the receipt, don't just print it):
--
-- ⚠⚠ THE HEADER LIST BELOW WAS INCOMPLETE UNTIL 2026-09-03 AND THE OMISSION IS
--    SILENT-BUT-FATAL. The old text named only message_type, action and
--    from_agent_type. `client_id` and `orchestration_id` are ALSO REQUIRED: a
--    message without them is consumed and REFUSED with
--      INCOMING_MESSAGE_REJECTED :: missing required header(s): client_id, orchestration_id
--    and the only symptom you see is NO ORCHESTRATION ROW — which is exactly the
--    signature of ordinary queue latency, the thing CLAUDE.md tells you not to
--    retry on. So the incomplete recipe and the correct do-not-retry guidance
--    COMBINE into a trap: you follow this file, get a refusal, read CLAUDE.md,
--    and wait for a message that will never run. It cost 37 minutes on
--    2026-09-03 and would have cost longer without the library check below.
--
--    THE CHECK, and it is the whole reason kafka-publish-lib exists:
--      kafka_verify_landing "$corr" 30
--      # 0 = landed · 13 = published, not landed (wait) · 12 = CONSUMED AND
--      # REFUSED — the error text names the missing headers. Run it BEFORE you
--      # conclude latency; latency and refusal look identical from the DB.
--
--   . "$REPO_ROOT/scripts/kafka-publish-lib.sh"
--   corr=$(cat /proc/sys/kernel/random/uuid)
--   orch=$(cat /proc/sys/kernel/random/uuid)
--   msg=$(jq -c -n --arg s "<site uuid>" --arg d "<domain>" '{action:"orchestrate",
--        config:{agent_type:"delivery-review-filer"},
--        input_data:{site_id:$s,domain:$d,site_url:("https://"+$d),brief:"<one paragraph>"}}')
--   kafka_publish_checked --topic system.agent.generic.requests --payload "$msg" \
--     --correlation "$corr" \
--     --header "orchestration_id=$orch" \
--     --header "orchestration_name=delivery-review-filer-$(date +%H%M%S)" \
--     --header "client_id=demo_client" \
--     --header "request_id=$(cat /proc/sys/kernel/random/uuid)" \
--     --header "message_id=$(cat /proc/sys/kernel/random/uuid)" \
--     --header "step_name=start" \
--     --header "message_type=request" --header "action=orchestrate" \
--     --header "from_agent_type=user" --header "from_agent_id=cli" \
--     --header "responses_topic=system.agent.generic.responses"
--
--   PROVEN 2026-09-03 on idea.uk: this envelope landed and COMPLETED in under 25
--   seconds — so the "budget 29 minutes of queue latency" figure did not apply to
--   this dispatch at all, and a slow one is a reason to VERIFY, not to assume.
--
-- ====================== RECOVERY: stamped but unemailed =====================
-- If the send fails AFTER the claim (SMTP died mid-conversation; a template
-- typo survived to the post-fill scan), the handover IS stamped and a re-
-- dispatch is REFUSED (ErrAlreadyDelivered) — that refusal is the double-send
-- guard working. Recovery is DELIBERATE and manual:
--   1. Read the failed work item / orchestration error to see what broke.
--   2. Mint fresh tokens by hand (pgcrypto digest matches Go's HashToken):
--        plain=$(head -c32 /dev/urandom | basenc --base64url | tr -d '=')
--        INSERT INTO customer_access_tokens
--               (site_id, purpose, token_hash, expires_at, single_use, created_by)
--        VALUES ('<site>', 'confirm_transfer',
--                encode(digest('<plain>','sha256'),'hex'),
--                (SELECT live_link_expires_at FROM sites WHERE id='<site>'),
--                false, 'operator-recovery');
--      (zip_download rows additionally set stored_url + stored_url_expires_at
--       from a fresh zip-deliverable run.)
--   3. Send the email by hand from the mail client, links composed as
--      https://links.webdesign.uk/c/<plain> and /d/<plain>.
-- ===========================================================================

BEGIN;

INSERT INTO agent_definitions (type, display_name, description, category, agent_category, status, is_active, input_contract, default_config)
SELECT
  'delivery-review-filer',
  'Delivery Review Filer',
  'Files the owner''s pre-delivery review item (needs_delivery_review, parked at needs_human_review with spec.checkpoint=true) for one finished site. The owner''s APPROVE on that item is what the delivery email''s gate reads (DGH-017); resolve does NOT open the gate. Owner ruling 2026-08-21 (review) + 2026-08-26 (review extends to owner edits, internal only).',
  'executor',
  'executor',
  'active',
  true,
  jsonb_build_object(
    'required', jsonb_build_array('site_id', 'domain', 'site_url', 'brief'),
    'notes', 'site_url + brief are what the owner asked to see (DECISION_2026-08-21e: the brief and the render). All four cross the spawn->call boundary via input_data.'
  ),
  jsonb_build_object('workflow', jsonb_build_object(
    'steps', jsonb_build_object(
      'file_review', jsonb_build_object(
        'action', 'create_work_item',
        'config', jsonb_build_object(
          'site_id',        'input_data.site_id',
          'item_type',      'needs_delivery_review',
          'status',         'needs_human_review',
          'item_pipeline',  'delivery',
          'severity',       'high',
          'source',         'delivery-review-filer',
          'summary',        'Pre-delivery review: look at the brief and the rendered site, edit if needed, then APPROVE to release the delivery email. Resolve does NOT release it.',
          'item_key_prefix','delivery_review',
          'spec_literal',   jsonb_build_object('checkpoint', true),
          'spec_paths',     jsonb_build_object(
                              'site_url', 'input_data.site_url',
                              'brief',    'input_data.brief',
                              'domain',   'input_data.domain'),
          'priority',       10
        ),
        'output_field', 'review_item',
        'next_step', 'complete'
      ),
      'complete', jsonb_build_object('action', 'complete_workflow')
    ),
    'start_step', 'file_review'
  ))
WHERE NOT EXISTS (
  SELECT 1 FROM agent_definitions WHERE type = 'delivery-review-filer' AND deleted_at IS NULL
);

INSERT INTO agent_definitions (type, display_name, description, category, agent_category, status, is_active, input_contract, default_config)
SELECT
  'delivery-email-sender',
  'Delivery Email Sender',
  'Sends THE delivery email for one approved site: claims the delivery (review gate + once-only handover stamp + customer link minting, platform/delivery.Claim) and sends through platform/mailer. The body is the body_template in this config — OWNER-EDITABLE HERE, no roll needed; figures trace to the attested register (30 days; GBP 10/mo rent; GBP 200 buy). A second dispatch for a delivered site is refused by the stamp.',
  'executor',
  'executor',
  'active',
  true,
  jsonb_build_object(
    'required', jsonb_build_array('site_id', 'customer_email', 'live_site_url'),
    'optional', jsonb_build_array('zip_presigned_url', 'zip_presign_minutes'),
    'notes', 'zip_presigned_url comes from a zip-deliverable-dispatch run''s output; omit it only with a template that does not name {{zip_link}} — the action refuses the mismatch BEFORE stamping.'
  ),
  jsonb_build_object('workflow', jsonb_build_object(
    'steps', jsonb_build_object(
      'send_email', jsonb_build_object(
        'action', 'send_delivery_email',
        'config', jsonb_build_object(
          'site_id',            'input_data.site_id',
          'customer_email',     'input_data.customer_email',
          'live_site_url',      'input_data.live_site_url',
          'zip_presigned_url',  'input_data.zip_presigned_url',
          'zip_presign_minutes','input_data.zip_presign_minutes',
          'links_host',         'links.webdesign.uk',
          'subject',            'Your website is ready',
          'body_template',      E'Your site is live now at {{live_site}}\nIt stays live there for {{days}} days.\n\nYOUR FILES\nYour finished site as a ZIP, yours to keep:\n{{zip_link}}\nThe ZIP comes with instructions that walk you through putting it on free hosting.\n\nKEEPING IT ONLINE\nTo keep the site up after those {{days}} days, you host it yourself. Free hosting works well.\n\nTHE DOMAIN\nRenting the domain is 10 pounds a month, or buying it outright is a one-off 200 pounds. It is then yours, and you move it to your own registrar; we give you what you need to do that. We supply .co.uk and .uk only. Reply to this email to arrange either.\n\nWHEN YOU HAVE MOVED\nOnce your site is off our hosting, press the button here so we stop reminding you:\n{{confirm_link}}\n\nNo changes are included. You get the site as it is built. The files are yours, and editing a site that already works is a great deal easier than starting from a blank page.\n\nwebdesign.uk'
        ),
        'output_field', 'delivery_email',
        'next_step', 'complete'
      ),
      'complete', jsonb_build_object('action', 'complete_workflow')
    ),
    'start_step', 'send_email'
  ))
WHERE NOT EXISTS (
  SELECT 1 FROM agent_definitions WHERE type = 'delivery-email-sender' AND deleted_at IS NULL
);

-- Verify: both rows present and active, and the sender's template still names
-- the three always-real links (a template edit that drops one is a decision,
-- not a drift — this makes it loud).
DO $$
DECLARE n int; tpl text;
BEGIN
  SELECT count(*) INTO n FROM agent_definitions
   WHERE type IN ('delivery-review-filer','delivery-email-sender')
     AND is_active AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;
  IF n <> 2 THEN
    RAISE EXCEPTION '651 verify failed: expected 2 active agents, found %', n;
  END IF;
  SELECT default_config->'workflow'->'steps'->'send_email'->'config'->>'body_template'
    INTO tpl FROM agent_definitions WHERE type='delivery-email-sender';
  IF tpl NOT LIKE '%{{live_site}}%' OR tpl NOT LIKE '%{{confirm_link}}%' OR tpl NOT LIKE '%{{days}}%' THEN
    RAISE EXCEPTION '651 verify failed: the body_template lost a core placeholder';
  END IF;
END $$;

COMMIT;
