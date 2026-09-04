-- FILE: docs/agent_docs/sql_for_agents/775_delivery_followup_sender_seeded_disabled_until_the_owner_rules.sql
--
-- bugs_open/477 step B, part 2 of 2: the follow-up sender, SEEDED DISABLED.
--
-- ⚠⚠ THIS MIGRATION SENDS NOTHING, AND THAT IS ITS DESIGN, NOT ITS LIMITATION.
-- The scheduled task is inserted with enabled = FALSE. Applying this file is
-- safe at any hour. Turning it on is one UPDATE, and it must not be run until
-- three things are true — see "BEFORE YOU ENABLE" below. The check at the end
-- REFUSES to apply if anything here would already be live.
--
-- WHY DISABLED. Not because the interval is unknown — that is now settled — but
-- because enabling it emails a real person and that should be a deliberate act by
-- someone who has read the warning below.
--
-- ⚠ THE INTERVAL IS RULED: **THREE DAYS** (owner, 2026-09-04, verbatim: "I think
-- the follow up should be 3 days"), relayed by the site_delivery_and_editor lane.
-- It supersedes the "a week or so" in his original suggestion, which is what
-- caused 477 to be filed. The Go action still refuses to run without an explicit
-- `followup_after_days` rather than defaulting, so this value is a decision
-- recorded in config, never a fallback.
--
-- ⚠ AND THE FIRST REAL RUN WILL EMAIL THE OWNER. idea.uk is the only site in the
-- estate this can select (handed_over_at set, transfer_confirmed_at NULL — 1 of
-- 60, measured 2026-09-04), and its delivery address is aaa@designconsultancy.co.uk,
-- his own. He must be told before, not after.
--
-- REQUIRES: migration 774 (sites.followup_sent_at) AND an image carrying the
-- send_followup_email action. A seed naming an unregistered action fails at
-- runtime, so this file checks 774 and refuses without it; the image it cannot
-- check, which is why the task ships disabled.
--
-- THE RECIPIENT comes from build_queue.direction->>'customer_email' and NEVER
-- from sites.email (corrected 2026-08-31 in 651's header, bugs_open/420: since
-- the contract split, sites.email is the PUBLISHED contact only and is
-- legitimately NULL on a post-420 site). The pre_query below takes it from the
-- right place and REFUSES to emit a row without one — a follow-up with no
-- recipient must be a site that is skipped, not a dispatch that fails.
--
-- ⚠⚠ AND TODAY THAT SKIPS EVERY SITE THERE IS. This is the honest headline of
-- this migration and it must not be discovered later by a silence.
--
-- `[MEASURED 2026-09-04]` build_queue has **ZERO** rows for idea.uk — the only
-- site in the estate that has ever been handed over. It was our own rehearsal
-- site, delivered on an address typed into the dispatch by hand, so it never
-- went through the order pipeline that writes build_queue. The pre_query
-- therefore returns nothing for it, for ever, and would go on doing so silently:
-- a scheduled task selecting zero rows looks exactly like "nothing is due".
--
-- Found by running the pre_query with a DEMAND CONTROL (the same query at
-- interval '0 days', which MUST have returned idea.uk and returned nothing).
-- Without that control the zero above reads as correct.
--
-- The fallback that looks obvious does not work either, and this was measured
-- rather than assumed: the delivery run DOES record the address it used
-- (orchestration_states, input_data.customer_email = the address idea.uk was
-- delivered to), but `[MEASURED 2026-09-04]` the OLDEST row in that table is
-- less than **24 hours** old (6,662 rows, oldest 2026-09-03 11:47Z — the
-- stale-orchestration-reaper runs every 180s). A follow-up due in seven days
-- would look for that row six days after it was reaped.
--
-- SO THE REAL GAP IS THIS: the estate has no DURABLE record of who a delivered
-- site was delivered to. build_queue holds it only for order-originated sites;
-- orchestration_states holds it for about a day. The structural fix is to stamp
-- the recipient onto the site row at delivery time, in the same statement that
-- stamps handed_over_at — which is a change to the delivery path
-- (platform/delivery/prepare.go) owned by the site_delivery_and_editor lane and
-- routed to them, NOT made here. Until it exists, this sender can only reach
-- sites that came through an order, and the verify below says how many that is.
--
-- BEFORE YOU ENABLE (all three, in this order):
--   1. ~~The owner has given an interval~~ DONE — three days, ruled 2026-09-04
--      and carried in this agent's config below.
--   2. The owner has been told, in words, that enabling it emails him about
--      idea.uk.
--   3. An image carrying send_followup_email has rolled. Check per SERVICE:
--      SELECT pod_name, git_commit, started_at FROM service_binary_capabilities
--       WHERE kind='build' AND pod_name LIKE 'agent-chassis-%' ORDER BY started_at DESC;
--      then: git merge-base --is-ancestor 0949244e8 <stamp>
--      ⚠ 0949244e8 (the {{instructions_link}} RENAME), not f89dfa31d (the commit
--        that added the action). Necessary-and-sufficient matters here: a binary
--        between the two carries send_followup_email but does NOT know
--        {{instructions_link}}, which this file's template uses. The literal then
--        survives the fill and trips the post-fill {{ scan — which runs AFTER
--        ClaimFollowup has stamped followup_sent_at, so the customer's single
--        follow-up is consumed and no email is sent. Config naming a token must
--        never go live ahead of the binary that can fill it.
--   Then: UPDATE scheduled_tasks SET enabled = true WHERE name = 'delivery-followup-send';
--
-- ⚠ THE LETTER QUOTES NO PRICE, AND THAT IS DELIBERATE — DO NOT "COMPLETE" IT.
-- Its domain paragraph is one sentence: "If you have not decided yet, reply to
-- this email and we will sort it out." The delivery letter quotes the figures
-- (10 pounds a month, 59.99 to buy); this one does not repeat them, and the
-- reason is the same rule that keeps the hosting steps out of it (bugs_open/475):
-- anything that can go out of date must live somewhere correctable, never in a
-- second copy the customer already holds. TWO letters quoting one price is two
-- places to update and nothing joining them — the buy-out has already moved from
-- 200 to 59.99 once, and a follow-up seeded with the old figure would have been
-- wrong the day it was written, silently, in a customer's inbox.
-- Owner of the prices is the `stripe` lane (2026-09-04): if this letter ever
-- genuinely needs to state a figure, ask them rather than copying one.
--
-- ⚠ THE COPY SAYS "A FEW DAYS AGO", NOT "THREE DAYS AGO", DELIBERATELY. The
-- interval is CONFIG (`followup_after_days`) and prose that restates a config
-- value is a lie waiting for the config to change — set it to 5 and the letter
-- is wrong, silently, with nothing to catch it. Never write a configurable number
-- into copy that the same file makes configurable.
--
-- TWO THINGS THE COPY DOES CARRY, both from the owner performing the hosting
-- steps himself on 2026-09-04 and neither of them guesses:
--   * "about forty minutes, and most of that is waiting" — his measured elapsed
--     time, most of it a signup confirmation, a rejected password and an
--     invisible security check. At three days a customer who has not confirmed
--     may be stuck in a signup fight rather than ignoring us, so the letter reads
--     "here is what you need", never "you have not done this yet".
--   * ⚠ THE PRIVATE-BY-DEFAULT PARAGRAPH IS THE MOST USEFUL LINE IN THE EMAIL.
--     A Netlify Drop site is PRIVATE by default and looks perfectly public to the
--     person who uploaded it — he opened his own URL, saw his site, and it was
--     rendering only because he was signed in; a private window said "This site
--     is private". So a customer can believe in good faith that they are
--     finished, press the confirm button, and have a site nobody can reach — and
--     transfer_confirmed_at would then suppress the one message that might have
--     told them. We cannot check somebody else's host, so the email carries the
--     check instead. It costs nothing and it is the one nobody thinks to run.
--
-- Council: submitted with the step B round; the interval and copy changes go with
-- the 778 round. Rollback sidecar: 775_..._ROLLBACK.sql.

BEGIN;

-- Refuse without 774: the agent would claim on a column that does not exist and
-- every run would fail at the first statement.
DO $$
BEGIN
  IF NOT EXISTS (
    SELECT 1 FROM information_schema.columns
     WHERE table_schema='public' AND table_name='sites' AND column_name='followup_sent_at'
  ) THEN
    RAISE EXCEPTION '775 REFUSED: sites.followup_sent_at is missing. Apply 774 first — without it the follow-up has no at-most-once claim at all.';
  END IF;
END $$;

INSERT INTO agent_definitions (type, display_name, description, category, agent_category, status, is_active, input_contract, default_config)
SELECT
  'delivery-followup-sender',
  'Delivery Follow-up Sender',
  'Sends the ONE post-delivery follow-up email for a site: claims it once-only on sites.followup_sent_at, refuses any site whose customer has already pressed the confirm button (transfer_confirmed_at), mints a fresh confirm link and sends through platform/mailer. The body is the body_template in this config — OWNER-EDITABLE HERE, no roll needed. It carries the instructions URL and never the instructions themselves, because a copy the customer already holds cannot be corrected (bugs_open/475). It cannot be a config variant of delivery-email-sender: that agent claims through the handover stamp, which refuses every site already handed over, i.e. exactly this population (bugs_open/477).',
  'executor',
  'executor',
  'active',
  true,
  jsonb_build_object(
    'required', jsonb_build_array('site_id', 'customer_email', 'live_site_url'),
    'optional', jsonb_build_array(),
    'notes', 'Dispatched per site by the delivery-followup-send scheduled task, whose pre_query selects the due, unconfirmed, not-yet-followed-up population and supplies all three fields. customer_email comes from build_queue.direction->>''customer_email'', never sites.email.'
  ),
  jsonb_build_object('workflow', jsonb_build_object(
    'steps', jsonb_build_object(
      'send_followup', jsonb_build_object(
        'action', 'send_followup_email',
        'config', jsonb_build_object(
          'site_id',             'input_data.site_id',
          'customer_email',      'input_data.customer_email',
          'live_site_url',       'input_data.live_site_url',
          'links_host',          'links.webdesign.uk',
          'instructions_url',    'https://webdesign.uk/your-site',
          'followup_after_days', 3,
          'subject',             'Your website, and where the instructions live',
          'body_template',       E'A few days ago we sent you your finished site and your files.\n\nThis is the one follow-up you will get from us.\n\nYOUR SITE\nIt is still live at {{live_site}} for the rest of the {{days}} days.\n\nTHE INSTRUCTIONS\nEverything about putting the site on your own hosting is kept here, and it is kept up to date:\n{{instructions_link}}\nIf that page and anything you downloaded ever disagree, believe the page.\n\nPutting it up takes about forty minutes, and most of that is waiting for the host to email you back. It feeling slow does not mean anything has gone wrong.\n\nIF YOU HAVE ALREADY PUT IT UP\nOpen your new address in a private browsing window. Some hosts show a brand new site only to the person who uploaded it, and it looks completely normal to you while nobody else can reach it. If you see a page saying the site is private, your host will have a setting to make it public.\n\nTHE DOMAIN\nIf you have not decided yet, reply to this email and we will sort it out.\n\nWHEN YOU HAVE MOVED\nOnce your site is off our hosting, press the button here to tell us:\n{{confirm_link}}\n\nwebdesign.uk'
        ),
        'output_field', 'followup_email',
        'next_step', 'complete'
      ),
      'complete', jsonb_build_object('action', 'complete_workflow')
    ),
    'start_step', 'send_followup'
  ))
WHERE NOT EXISTS (
  SELECT 1 FROM agent_definitions WHERE type = 'delivery-followup-sender' AND deleted_at IS NULL
);

-- The schedule. DISABLED. Modelled on zip-link-refresh, which is the estate's
-- only other site-selecting scheduled sender.
--
-- THE PRE_QUERY IS A CANDIDATE FILTER, NOT THE GUARD. Every predicate here is
-- repeated inside delivery.ClaimFollowup's UPDATE, and that repetition is the
-- point: a customer who presses the confirm button between this SELECT and the
-- dispatch landing must not be emailed, and only the UPDATE can promise that.
-- If you ever find these two out of step, the UPDATE is right.
INSERT INTO scheduled_tasks
  (name, description, interval_seconds, target_agent_type, input_data, concurrency_group, max_concurrent, pre_query, enabled, timeout_seconds)
SELECT
  'delivery-followup-send',
  'One post-delivery follow-up per site, three days after handover (owner ruling 2026-09-04), suppressed by the confirm button. SEEDED DISABLED (bugs_open/477): enabling it emails a real customer, and the only selectable site today is idea.uk, whose delivery address is the owner''s own.',
  21600, -- six hours: the follow-up is due on a day, not a minute
  'delivery-followup-sender',
  '{}'::jsonb,
  'delivery-followup-send',
  1,
  $q$
    SELECT s.id::text                          AS site_id,
           bq.direction->>'customer_email'     AS customer_email,
           -- ⚠ CORRECTED 2026-09-04 BEFORE THIS FILE WAS EVER APPLIED, and the
           -- version it replaces was in the same class as the bug this whole
           -- lane exists to fix: it would have mailed a customer an address
           -- that does not resolve.
           --
           -- It said `'https://' || s.domain`, copied from the operator recipe
           -- the idea.uk rehearsal used. That is right ONLY for a domain
           -- already pointed at us. For a customer who has bought a build and
           -- not yet transferred their domain — the NORMAL case — it composes a
           -- URL that 404s or does not resolve at all. It looked correct
           -- because idea.uk is ours and was already pointed, so the defect
           -- could not show itself on the only delivery this estate has made.
           --
           -- OWNER RULING (2026-09-04, relayed by the site_delivery_and_editor
           -- lane): "the delivery email should say it is on the ugg2 subdomain
           -- not paper-cups.com. I will transfer it to our system later."
           --
           -- publish_project is the SERVING hostname (boxingonline.ugg2.com);
           -- publish_target is the worker name. b2worker.go:63-70 refuses an
           -- empty publish_project AND refuses one equal to the site domain,
           -- which is the code's own statement that these are different things.
           --
           -- ⚠ NO FALLBACK TO s.domain. A COALESCE onto the domain is exactly
           -- the defect above wearing a guard: it would be correct for the two
           -- pointed sites and quietly wrong for every customer who has not
           -- transferred. A site with no serving host is EXCLUDED here and
           -- COUNTED in the verify block, so the silence is loud at apply time
           -- rather than looking like "nothing due".
           'https://' || s.publish_project AS live_site_url,
           s.domain
      FROM sites s
      JOIN LATERAL (
             SELECT direction FROM build_queue
              WHERE lower(domain) = lower(s.domain)
              ORDER BY created_at DESC LIMIT 1
           ) bq ON true
     WHERE s.handed_over_at IS NOT NULL
       AND s.handed_over_at <= now() - interval '3 days'
       AND s.live_link_expires_at > now()
       AND s.transfer_confirmed_at IS NULL
       AND s.followup_sent_at IS NULL
       AND COALESCE(bq.direction->>'customer_email','') <> ''
       -- No serving host, no email: a follow-up naming an address that does not
       -- resolve is worse than no follow-up. Counted in the verify block.
       AND COALESCE(s.publish_project,'') <> ''
  $q$,
  false, -- ⚠ DISABLED. See "BEFORE YOU ENABLE" in this file's header.
  600
WHERE NOT EXISTS (
  SELECT 1 FROM scheduled_tasks WHERE name = 'delivery-followup-send'
);

-- VERIFY — a DO block, not a SELECT. A verify made of SELECTs cannot stop the
-- COMMIT: ON_ERROR_STOP ignores a non-empty result set, so the migration would
-- report and commit anyway. RAISE is what aborts.
DO $$
DECLARE
  agents int;
  enabled_now boolean;
  tpl text;
  selectable int;
  unaddressable int;
  handed int;
  no_host int;
BEGIN
  SELECT count(*) INTO agents FROM agent_definitions
   WHERE type = 'delivery-followup-sender'
     AND is_active AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;
  IF agents <> 1 THEN
    RAISE EXCEPTION '775 FAILED: expected exactly 1 active delivery-followup-sender, found %', agents;
  END IF;

  -- THE LOAD-BEARING ASSERTION. If this file ever ships enabled by accident, it
  -- emails a customer within six hours and nobody decided to.
  SELECT enabled INTO enabled_now FROM scheduled_tasks WHERE name = 'delivery-followup-send';
  IF enabled_now IS DISTINCT FROM false THEN
    RAISE EXCEPTION '775 FAILED: delivery-followup-send is enabled=%, and this migration must leave it FALSE. Enabling is a separate, deliberate act after the owner has ruled and been warned.', enabled_now;
  END IF;

  -- The template must still name the three links the action always produces. A
  -- template edit that drops one is a decision, not a drift, and this makes it
  -- loud rather than silent.
  SELECT default_config->'workflow'->'steps'->'send_followup'->'config'->>'body_template'
    INTO tpl FROM agent_definitions WHERE type='delivery-followup-sender' AND deleted_at IS NULL;
  IF tpl NOT LIKE '%{{live_site}}%' OR tpl NOT LIKE '%{{confirm_link}}%' OR tpl NOT LIKE '%{{instructions_link}}%' THEN
    RAISE EXCEPTION '775 FAILED: the follow-up template is missing one of {{live_site}} / {{confirm_link}} / {{instructions_link}}';
  END IF;
  -- And it must NOT promise the reminders stop, because this IS the reminder.
  IF tpl ILIKE '%stop reminding%' OR tpl ILIKE '%no more reminders%' THEN
    RAISE EXCEPTION '775 FAILED: the follow-up template promises reminders will stop. This email IS the reminder, and it is the last one; the promise belongs on the confirm page once there is something to suppress (bugs_open/477).';
  END IF;

  -- Report what this WOULD select on the day it is enabled, so whoever enables
  -- it knows who gets an email. Not an assertion: the population changes.
  SELECT count(*) INTO selectable FROM sites s
   WHERE s.handed_over_at IS NOT NULL
     AND s.handed_over_at <= now() - interval '3 days'
     AND s.live_link_expires_at > now()
     AND s.transfer_confirmed_at IS NULL
     AND s.followup_sent_at IS NULL;

  -- AND how many handed-over sites this sender can NEVER reach because nothing
  -- records who they were delivered to. This is the number that must not be
  -- discovered by a silence: if it equals the number of handed-over sites, the
  -- sender is a mechanism with no population, however healthy it looks.
  SELECT count(*) INTO unaddressable FROM sites s
   WHERE s.handed_over_at IS NOT NULL
     AND NOT EXISTS (
       SELECT 1 FROM build_queue bq
        WHERE lower(bq.domain) = lower(s.domain)
          AND COALESCE(bq.direction->>'customer_email','') <> ''
     );
  SELECT count(*) INTO handed FROM sites WHERE handed_over_at IS NOT NULL;

  RAISE NOTICE '775 OK: agent seeded, schedule seeded DISABLED. Sites this would select TODAY if enabled: % (check WHO before enabling — the estate''s only handed-over site is idea.uk, addressed to the owner).', selectable;
  RAISE NOTICE '775 GAP: % of % handed-over site(s) have NO recorded recipient (no build_queue row with a customer_email) and can never be selected. If those two numbers are equal, this sender has no reachable population at all — see the header.', unaddressable, handed;

  -- THE SECOND GAP, and today it is the binding one. publish_project is set BY
  -- HAND, per site (nothing in the codebase writes it — the only non-test
  -- reference is b2worker.go's refusal), so a freshly built site does not have
  -- one and this sender will skip it rather than mail a dead address.
  SELECT count(*) INTO no_host FROM sites
   WHERE handed_over_at IS NOT NULL AND COALESCE(publish_project,'') <> '' IS NOT TRUE;
  RAISE NOTICE '775 GAP 2: % of % handed-over site(s) have NO publish_project (serving hostname) and are skipped, because a follow-up naming an address that does not resolve is worse than none. publish_project is set BY HAND per site; nothing writes it.', no_host, handed;
END $$;

COMMIT;
