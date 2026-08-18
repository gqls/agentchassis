-- SQL_2026-08-18c: three owner corrections to this morning's 18b register
-- surgery (owner chat, 2026-08-18, second directive; applied by the
-- site_delivery_and_editor session):
--  1. NO EXAMPLE SITES YET. The four linked domains were not produced by
--     this one-shot route, so naming them as examples overclaims. Examples
--     return once the owner has produced sites through webdesign.uk itself.
--  2. THE WORD "PREVIEW" IS FOR BEFORE PAYMENT ONLY, where its job is to say
--     there is none. The post-payment link stops being called a preview
--     (the home page currently says "no preview beforehand" and then "a
--     preview link", which reads as a contradiction). It becomes "a link to
--     your site, already live at an address we provide".
--  3. HOSTING, SAID PLAINLY: keeping the site online beyond the included
--     month means hosting it yourself; free hosting is recommended by name
--     and the ZIP's included instructions walk through the set-up. Never
--     "we do the set-up".

BEGIN;

DO $fix$
DECLARE
  spec_id uuid;
  wb text;
  facts jsonb;
  a1 text; a2 text; a3 text; a4 text;
BEGIN
  SELECT id, data->>'writer_block', data->'facts'
    INTO spec_id, wb, facts
    FROM site_specs
   WHERE site_id='1fcfa4f3-ec80-4010-878b-b971cd46711f'
     AND aspect='evidence_base' AND is_current;
  IF spec_id IS NULL THEN RAISE EXCEPTION 'no current evidence_base row'; END IF;
  IF NOT (facts @> '[{"id":"any_site_type_examples"}]'::jsonb) THEN
    RAISE EXCEPTION '18b not applied or 18c already applied - read the row';
  END IF;

  -- writer_block edit 1: drop the example links from the conversation paragraph
  a1 := ', and point at live examples built with this system, linked by name: noted.co.uk, cookly.uk, dartsonline.com, vetcomparison.uk.';
  IF (length(wb) - length(replace(wb, a1, ''))) / length(a1) <> 1 THEN
    RAISE EXCEPTION 'anchor 1 not found exactly once';
  END IF;
  wb := replace(wb, a1,
    '. Do not name or link example sites: none built by this one-shot route exist yet, and examples will be added once the owner has produced some through it.');

  -- writer_block edit 2: fact enumeration - no examples, and live link not preview
  a2 := 'the preview link and ZIP, that the site arrives hosted by us under a domain that can be rented for £10 a month or bought outright for a one-off £200 and then transferred freely, that there is no approval stage and the product is one-shot, that what is delivered is a starter site with initial copy the customer is expected to edit, that any sort of site can be asked for with the four example domains (noted.co.uk, cookly.uk, dartsonline.com, vetcomparison.uk), that no pre-sales service is included';
  IF (length(wb) - length(replace(wb, a2, ''))) / length(a2) <> 1 THEN
    RAISE EXCEPTION 'anchor 2 not found exactly once';
  END IF;
  wb := replace(wb, a2,
    'the live link and ZIP, that the site arrives hosted by us under a domain that can be rented for £10 a month or bought outright for a one-off £200 and then transferred freely, that there is no approval stage and the product is one-shot, that what is delivered is a starter site with initial copy the customer is expected to edit, that any sort of site can be asked for (no example sites are named yet), that keeping the site online beyond the included month means hosting it yourself with free options recommended and instructions included, that no pre-sales service is included');

  -- writer_block edit 3: the offer paragraph stops calling it a preview
  a3 := 'and a preview link so they can see it working straight away; the preview stays up for about a month and the ZIP is permanent.';
  IF (length(wb) - length(replace(wb, a3, ''))) / length(a3) <> 1 THEN
    RAISE EXCEPTION 'anchor 3 not found exactly once';
  END IF;
  wb := replace(wb, a3,
    'and a link to their site, already live at a web address we provide; it stays live for about a month and the ZIP is permanent.');

  -- writer_block edit 4: the gets paragraph, same renaming
  a4 := 'plus a preview link, live for about a month, so they can see it working before they host it.';
  IF (length(wb) - length(replace(wb, a4, ''))) / length(a4) <> 1 THEN
    RAISE EXCEPTION 'anchor 4 not found exactly once';
  END IF;
  wb := replace(wb, a4,
    'plus a link to their site, already live at a web address we provide, staying live for about a month.');

  -- writer_block edit 5: append the naming rule + hosting clarity paragraph
  wb := wb || E'\n\n' ||
    'THE WORD PREVIEW IS FOR BEFORE PAYMENT ONLY, where its whole job is to say there is none. Never call the post-payment link a preview: the same page says there is no preview before paying, and one word doing both jobs reads as a contradiction (the owner flagged exactly this on the home page, 2026-08-18). After payment the site is simply live, at an address we provide, for about a month. Call it that: the site, live, or a link to your live site. And say what follows plainly: keeping it online beyond that month means hosting it yourself. Recommend the free options by name, per third_party_options, and say that the ZIP''s included instructions walk through the set-up step by step. Never say that we do the set-up, and never offer our time with it: the help is the instructions and the recommendations.';

  -- facts: replace the examples fact, rename the delivery artefact, add keep-it-online
  SELECT jsonb_agg(
    CASE
      WHEN f->>'id' = 'any_site_type_examples' THEN
        jsonb_build_object('id','any_site_type','kind','attestation',
          'claim','The service builds any sort of site, not just business sites. No example sites are named yet: the sites on the owner''s other domains were not produced by this one-shot route, and examples will be added once the owner has produced sites through it.',
          'source', jsonb_build_object('attested_by','owner, 2026-08-18, second chat directive (examples deferred until the system has been used via the webdesign.uk route)'),
          'verified_at','2026-08-18',
          'writer_line','Not just business sites. Any sort of site can be asked for.')
      WHEN f->>'id' = 'delivery_preview_and_zip' THEN
        jsonb_build_object('id','delivery_live_link_and_zip','kind','capability',
          'claim','After payment the customer receives a ZIP of the finished site to keep, with instructions for putting it online, and a link to their site already live at a web address we provide; it stays live for about a month, and the ZIP is theirs permanently. The post-payment link is never called a preview: there is no preview before payment, and one word doing both jobs reads as a contradiction.',
          'source', jsonb_build_object('attested_by','owner, 2026-08-17 (the month and the permanent ZIP); owner, 2026-08-18, second chat directive (stop calling the post-payment link a preview)'),
          'verified_at','2026-08-18',
          'writer_line','You get the finished site as a ZIP to keep, plus a link to your site, already live at an address we provide. It stays live for about a month; the ZIP is yours for good.')
      ELSE f
    END)
    INTO facts FROM jsonb_array_elements(facts) f;

  facts := facts || jsonb_build_object('id','keep_it_online','kind','capability',
    'claim','Keeping the site online beyond the included month means the customer hosts it themselves. Free hosting is recommended by name (Cloudflare Pages or Netlify, per third_party_options) and the ZIP includes instructions that walk through setting it up. The help offered is the instructions and the recommendations, never our time or a done-for-you set-up.',
    'source', jsonb_build_object('attested_by','owner, 2026-08-18, second chat directive ("they need to host elsewhere and we can recommend and help them set up free hosting"); the written setup instructions are the owner''s 2026-08-11 attestation'),
    'verified_at','2026-08-18',
    'writer_line','To keep the site up after that month, you host it yourself. Free hosting works well, and the ZIP comes with instructions that walk you through setting it up.');

  UPDATE site_specs
     SET data = jsonb_set(jsonb_set(data, '{writer_block}', to_jsonb(wb)), '{facts}', facts),
         updated_at = now()
   WHERE id = spec_id;

  SELECT data->'facts', data->>'writer_block' INTO facts, wb FROM site_specs WHERE id = spec_id;
  IF facts @> '[{"id":"any_site_type_examples"}]'::jsonb OR facts @> '[{"id":"delivery_preview_and_zip"}]'::jsonb THEN
    RAISE EXCEPTION 'old facts still present after write';
  END IF;
  IF NOT (facts @> '[{"id":"keep_it_online"}]'::jsonb AND facts @> '[{"id":"delivery_live_link_and_zip"}]'::jsonb) THEN
    RAISE EXCEPTION 'new facts missing after write';
  END IF;
  IF position('noted.co.uk' in wb) > 0 THEN
    RAISE EXCEPTION 'example domains still present in writer_block';
  END IF;
  IF position('preview link' in wb) > 0 THEN
    RAISE EXCEPTION 'the phrase preview link survives in writer_block';
  END IF;
  RAISE NOTICE '18c applied: % facts, writer_block % chars, no example domains, no post-payment preview naming', jsonb_array_length(facts), length(wb);
END $fix$;

COMMIT;
