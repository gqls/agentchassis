-- SQL_2026-08-18b: the owner's 2026-08-18 copy directives, applied at the
-- register (the wire), before rewriting what-you-get, faq and how-it-works.
-- Written by the site_delivery_and_editor session at the owner's direction
-- (chat, 2026-08-18); every position below quotes or closely paraphrases the
-- owner's own words from that directive.
--
-- WHAT CHANGES (all attested "owner, 2026-08-18, chat directive"):
--  1. ONE-SHOT: there is no approval stage; the site is built once; that is
--     how the price stays down. Stated bluntly.
--  2. STARTER SITE: what is delivered is a starter site with initial copy
--     included; the customer starts with it and edits it; not a final product.
--  3. DOMAIN/HOSTING: the site arrives hosted by us under a domain the
--     customer can rent (£10/month, link in the delivery email) or buy
--     outright (one-off £200, then free to transfer to their own registrar or
--     host). Supersedes hosting_and_domain_not_included (removed).
--  4. NOT JUST BUSINESSES + EXAMPLES: any sort of site; live examples on the
--     owner's own domains: noted.co.uk, cookly.uk, dartsonline.com,
--     vetcomparison.uk (all four fetched 200 with real titles, 2026-08-18;
--     loancalculator.co.uk deliberately NOT listed while bugfix 146's tool
--     overflow is live on its tool pages).
--  5. NO PRE-SALES SERVICE: no encouragement to email or phone before
--     buying; the copy and the chat box do the answering; bare-bones starter
--     website without frills.
--  6. no_lock_in SCOPED to the site itself so it cannot contradict the £10/mo
--     domain rental (the one optional recurring charge).
--  7. Three banned_claims arming detection for the retired approval/pay-after
--     promises. Patterns deliberately match PROMISE shapes only, never the
--     new denials ("there is no approval stage" does not match any of them)
--     - the refunds ban (\brefunds?\b) shows what happens otherwise: it bans
--     the denial too, and the sentence vanished from index on 2026-08-18.
--
-- writer_block edits are BY ANCHOR with exactly-once guards (the lane's
-- established method). Em dashes deliberately absent from every inserted
-- sentence (the block's own first rule).

BEGIN;

DO $surgery$
DECLARE
  spec_id uuid;
  wb text;
  facts jsonb;
  bans jsonb;
  n_facts_before int; n_bans_before int;
  anchor_a text; anchor_b text; anchor_c text; anchor_d text;
BEGIN
  SELECT id, data->>'writer_block',
         data->'facts', COALESCE(data->'banned_claims','[]'::jsonb),
         jsonb_array_length(data->'facts'),
         jsonb_array_length(COALESCE(data->'banned_claims','[]'::jsonb))
    INTO spec_id, wb, facts, bans, n_facts_before, n_bans_before
    FROM site_specs
   WHERE site_id='1fcfa4f3-ec80-4010-878b-b971cd46711f'
     AND aspect='evidence_base' AND is_current;
  IF spec_id IS NULL THEN RAISE EXCEPTION 'no current evidence_base row'; END IF;

  ---------------------------------------------------------------------------
  -- writer_block, four anchored edits
  ---------------------------------------------------------------------------
  anchor_a := 'Hosting and the domain are not included and stay theirs. Hosting by us and a manual domain transfer are optional paid extras. Never say that we handle the setup, the hosting or the domain: that promise belonged to the old offer and is now false.';
  IF (length(wb) - length(replace(wb, anchor_a, ''))) / length(anchor_a) <> 1 THEN
    RAISE EXCEPTION 'anchor A not found exactly once - writer_block has moved, re-read it';
  END IF;
  wb := replace(wb, anchor_a,
    'The site arrives already hosted by us, under a domain the customer can rent or buy from us if they want to keep it: renting is £10 a month, and the subscription payment link arrives in the delivery email; buying is a one-off £200, and after that they are free to transfer the domain to their own registrar or host. The ZIP means they can equally host it anywhere themselves. Never say that we handle the setup or their own hosting: where they host it themselves, help points outward, per third_party_options.');

  anchor_b := 'the preview link and ZIP, that hosting and the domain are not included, that there is no lock-in, the rough duration, the contact details,';
  IF (length(wb) - length(replace(wb, anchor_b, ''))) / length(anchor_b) <> 1 THEN
    RAISE EXCEPTION 'anchor B not found exactly once';
  END IF;
  wb := replace(wb, anchor_b,
    'the preview link and ZIP, that the site arrives hosted by us under a domain that can be rented for £10 a month or bought outright for a one-off £200 and then transferred freely, that there is no approval stage and the product is one-shot, that what is delivered is a starter site with initial copy the customer is expected to edit, that any sort of site can be asked for with the four example domains (noted.co.uk, cookly.uk, dartsonline.com, vetcomparison.uk), that no pre-sales service is included, that there is no lock-in on the site itself, the rough duration, the contact details,');

  anchor_c := 'Not a long form, not a questionnaire, and nothing that has to be completed before they can talk to us. The point is to open a conversation, not to qualify a lead.';
  IF (length(wb) - length(replace(wb, anchor_c, ''))) / length(anchor_c) <> 1 THEN
    RAISE EXCEPTION 'anchor C not found exactly once';
  END IF;
  wb := replace(wb, anchor_c,
    'Not a long form and not a questionnaire. The conversation happens with the page''s own chat box, which answers questions about the offer. Never encourage the reader to email or phone with questions before buying: there is no pre-sales service, the price stays down because nobody''s time is included, and the copy and the chat box do the answering. Where an answer can help, put the help in the copy itself. Where a page invites a request, say that any sort of site can be asked for, not just a business site, and point at live examples built with this system, linked by name: noted.co.uk, cookly.uk, dartsonline.com, vetcomparison.uk.');

  anchor_d := 'it is the better half of the deal, and it should read like one.';
  IF (length(wb) - length(replace(wb, anchor_d, ''))) / length(anchor_d) <> 1 THEN
    RAISE EXCEPTION 'anchor D not found exactly once';
  END IF;
  wb := replace(wb, anchor_d,
    'it is the better half of the deal, and it should read like one.' || E'\n\n' ||
    'THE PRODUCT IS ONE-SHOT, AND THE PAGES SAY SO BLUNTLY. There is no approval stage: the site is built once, in one go, and that is how the price stays down. Never describe an approval, a sign-off, or any step where the customer confirms the site before delivery; that flow belonged to the old offer and is retired. What is delivered is a starter site with initial copy included: a site the customer starts with and edits, not a final product. Where a page answers whether the customer must write their own text, give both halves: the initial copy is written for them and included, and the site is theirs to edit, and we expect them to edit it as their site grows.');

  ---------------------------------------------------------------------------
  -- facts: remove the superseded hosting fact, scope no_lock_in, add seven
  ---------------------------------------------------------------------------
  IF NOT (facts @> '[{"id":"hosting_and_domain_not_included"}]'::jsonb) THEN
    RAISE EXCEPTION 'hosting_and_domain_not_included absent - previously applied or concurrently edited';
  END IF;
  IF facts @> '[{"id":"one_shot_no_approval"}]'::jsonb THEN
    RAISE EXCEPTION 'one_shot_no_approval already present - previously applied';
  END IF;

  SELECT jsonb_agg(
           CASE WHEN f->>'id' = 'no_lock_in' THEN
             jsonb_set(jsonb_set(f,
               '{claim}', to_jsonb('The £149 is one payment and the site itself carries no monthly fee and no lock-in. The optional domain rental at £10 a month is the only recurring charge, and it is escapable: the customer can buy the domain outright instead, or take the ZIP and host elsewhere.'::text)),
               '{writer_line}', to_jsonb('The site itself has no monthly fee. If you rent the domain from us, that is the one monthly payment, and it is optional.'::text))
           ELSE f END)
    INTO facts
    FROM jsonb_array_elements(facts) f
   WHERE f->>'id' <> 'hosting_and_domain_not_included';

  facts := facts
    || jsonb_build_object('id','one_shot_no_approval','kind','attestation',
         'claim','There is no approval stage. It is a one-shot product: the site is built once, and that is how the price stays down.',
         'source', jsonb_build_object('attested_by','owner, 2026-08-18, chat directive to the site_delivery session'),
         'verified_at','2026-08-18',
         'writer_line','There is no approval stage. The site is built once, in one go. That is how the price stays down.')
    || jsonb_build_object('id','starter_site_initial_copy','kind','capability',
         'claim','What is delivered is a starter site with initial copy included: a site the customer starts with and edits, not a finished final product. The customer is expected to edit and develop it.',
         'source', jsonb_build_object('attested_by','owner, 2026-08-18, chat directive'),
         'verified_at','2026-08-18',
         'writer_line','What you get is a starter site with initial copy included. You start with it and edit it. It is not built as a final product.')
    || jsonb_build_object('id','hosted_under_our_domain','kind','capability',
         'claim','The site is delivered already hosted by us, under a domain the customer can rent or buy from us if they want to keep it.',
         'source', jsonb_build_object('attested_by','owner, 2026-08-18, chat directive'),
         'verified_at','2026-08-18',
         'writer_line','We host the site under a domain you can rent or buy from us if you want to keep it.')
    || jsonb_build_object('id','domain_rent_monthly','kind','metric','value',10,
         'context_terms', jsonb_build_array('domain','rent','month','monthly','keep'),
         'claim','Renting the domain the site is served under costs £10 per month; the subscription payment link arrives in the delivery email.',
         'source', jsonb_build_object('attested_by','owner, 2026-08-18, chat directive'),
         'verified_at','2026-08-18',
         'writer_line','Renting the domain is £10 a month. The payment link comes in your delivery email.')
    || jsonb_build_object('id','domain_buy_once','kind','metric','value',200,
         'context_terms', jsonb_build_array('domain','buy','one-off','transfer','own'),
         'claim','Buying the domain outright is a one-off £200; the customer is then free to transfer it to their own registrar or host.',
         'source', jsonb_build_object('attested_by','owner, 2026-08-18, chat directive'),
         'verified_at','2026-08-18',
         'writer_line','Buying the domain is a one-off £200. After that you are free to transfer it to your own registrar or host.')
    || jsonb_build_object('id','any_site_type_examples','kind','entity',
         'claim','The service builds any sort of site, not just business sites. Live examples built with this system, on the owner''s own domains: noted.co.uk (a note-taking app), cookly.uk (home cooking), dartsonline.com (darts buying guides and news), vetcomparison.uk (a directory of UK veterinary practices). Each may be named and linked; no traffic, popularity or client claims about any of them.',
         'source', jsonb_build_object('attested_by','owner, 2026-08-18, chat directive ("It''s not just business sites... give them examples - links to my domains"); all four fetched HTTP 200 with real titles, 2026-08-18'),
         'verified_at','2026-08-18',
         'writer_line','Not just business sites. Examples built with this system: noted.co.uk, cookly.uk, dartsonline.com, vetcomparison.uk.')
    || jsonb_build_object('id','no_presales_service','kind','attestation',
         'claim','No pre-sales question-answering service and no customer service is included: the price is for a bare-bones starter website without frills, and it stays down because nobody''s time is included. The pages and the chat box do the answering; the copy should be as helpful as it can and recommend where to go, but never offer the owner''s time.',
         'source', jsonb_build_object('attested_by','owner, 2026-08-18, chat directive ("I don''t want to offer customer service... We can be as helpful as we can in the copy and in recommending people but not with my time")'),
         'verified_at','2026-08-18',
         'writer_line','The price is for a bare bones starter website, without frills. These pages and the chat box answer what we can.');

  ---------------------------------------------------------------------------
  -- banned_claims: arm detection for the retired approval/pay-after promises
  ---------------------------------------------------------------------------
  bans := bans
    || jsonb_build_object('reason','RETIRED FLOW (owner directive 2026-08-18): approval-step promises belong to the pay-after-preview offer; the product is one-shot. The denial ("there is no approval stage") deliberately does NOT match this pattern.',
         'pattern','\b(once|when|after|if|before|until) you(''ve| have)? approved?\b')
    || jsonb_build_object('reason','RETIRED FLOW (owner directive 2026-08-18): payment comes first; a "before you pay" promise describes the retired flow.',
         'pattern','\bbefore you pay\b')
    || jsonb_build_object('reason','RETIRED FLOW (owner directive 2026-08-18): the old "you pay after you have seen it" promise.',
         'pattern','\byou pay after\b');

  ---------------------------------------------------------------------------
  -- write + read back
  ---------------------------------------------------------------------------
  UPDATE site_specs
     SET data = jsonb_set(jsonb_set(jsonb_set(data,
                  '{writer_block}', to_jsonb(wb)),
                  '{facts}', facts),
                  '{banned_claims}', bans),
         updated_at = now()
   WHERE id = spec_id;

  SELECT data->'facts', data->'banned_claims', data->>'writer_block'
    INTO facts, bans, wb FROM site_specs WHERE id = spec_id;
  IF jsonb_array_length(facts) <> n_facts_before + 6 THEN
    RAISE EXCEPTION 'fact count % -> %, expected +6 (remove 1, add 7)', n_facts_before, jsonb_array_length(facts);
  END IF;
  IF jsonb_array_length(bans) <> n_bans_before + 3 THEN
    RAISE EXCEPTION 'ban count wrong';
  END IF;
  IF facts @> '[{"id":"hosting_and_domain_not_included"}]'::jsonb THEN
    RAISE EXCEPTION 'superseded fact still present';
  END IF;
  IF NOT (facts @> '[{"id":"domain_buy_once"}]'::jsonb AND facts @> '[{"id":"one_shot_no_approval"}]'::jsonb) THEN
    RAISE EXCEPTION 'new facts missing after write';
  END IF;
  IF position('THE PRODUCT IS ONE-SHOT' in wb) = 0 OR position('£10 a month' in wb) = 0 THEN
    RAISE EXCEPTION 'writer_block edits missing after write';
  END IF;
  IF position(E'—' in wb) > 0 AND position(E'—' in anchor_d) = 0 THEN
    NULL; -- pre-existing em dashes in untouched paragraphs are theirs, not ours
  END IF;
  RAISE NOTICE 'register updated: % facts, % bans, writer_block % chars', jsonb_array_length(facts), jsonb_array_length(bans), length(wb);
END $surgery$;

COMMIT;
