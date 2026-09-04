-- SQL_2026-09-04_pricing_page_rebuild.sql — finetuning.uk /pricing.html rewritten around RENT vs OWN.
-- Owner 2026-09-04: "The pricing page can have a price list on it and it needs to be rewritten e.g.
-- 'What actually drives the cost is scope, not buzzwords.'" and "Maybe the pricing page can make that
-- clearer as to what you can rent and what you can have." Prices confirmed by him the same day and
-- REGISTERED first (ft-hour-small £3.15, ft-hour-large £6.65): this site refuses unregistered numbers,
-- so the price list could not have been written before that.
-- Brief in full: BRIEF_2026-09-04_pricing_page.md. Seven sections (was five; the rent section and the
-- price list are new). Nothing on /your-own-model.html or /technical-details.html is touched.
-- The brief names the fact IDs rather than the figures, so the writer takes each price from the
-- register rather than from this text.
\set ON_ERROR_STOP on
BEGIN;
DO $$
DECLARE
  v_site uuid := '1368e337-dd1d-4799-bbb3-8221a1b79bcc';
  v_page uuid := 'f8baa455-6561-4af3-a6ca-e38e2dc378af';
  n int;
BEGIN
  SELECT count(*) INTO n FROM pages WHERE id=v_page AND site_id=v_site AND url='/pricing.html' AND build_status='deployed' AND status='active';
  IF n <> 1 THEN RAISE EXCEPTION 'pre-flight: /pricing.html is not a deployed active page (%)', n; END IF;
  SELECT count(*) INTO n FROM site_work_items WHERE page_id=v_page AND item_type IN ('needs_content_page','content_rewrite','page_rerender') AND status IN ('triaged','claimed','in_progress');
  IF n <> 0 THEN RAISE EXCEPTION 'pre-flight: % open build item(s) on the pricing page', n; END IF;
  SELECT count(*) INTO n FROM site_specs ss, jsonb_array_elements(ss.data->'facts') f
   WHERE ss.site_id=v_site AND ss.aspect='evidence_base' AND ss.is_current AND f->>'id' IN ('ft-price-99','ft-hour-small','ft-hour-large');
  IF n <> 3 THEN RAISE EXCEPTION 'pre-flight: expected 3 registered prices, found %', n; END IF;

  UPDATE pages SET build_status='planned', section_subjects=NULL, updated_at=NOW() WHERE id=v_page;

  INSERT INTO site_work_items (site_id, page_id, source, pipeline, item_type, severity, summary, priority,
                               handler_agent, status, created_by, spec, item_key, batch_id)
  VALUES (v_site, v_page, 'owner-request', 'build', 'needs_content_page', 'medium',
    'pricing: rewritten around what you can HAVE vs what you can RENT, with the price list the owner asked for',
    40, 'page-build-handler', 'triaged', 'finetuning_uk_service_lane',
    jsonb_build_object(
      'title', 'Pricing | FineTuning',
      'page_name', 'pricing',
      'page_type', 'content',
      'source', 'owner-request',
      'reason', 'owner 2026-09-04: the page must make clear what you can rent and what you can have, and carry a price list; the two hour prices were registered as facts the same day',
      'purpose', 'Show, side by side, the model you buy and keep and the hours you rent, with the three registered prices.',
      'sections', jsonb_build_array('hero','generic-text-block','generic-text-block','generic-text-block','features','faq','call-to-action'),
      'suggestion', 'This is finetuning.uk''s pricing page, rewritten so a reader can see at a glance WHAT THEY CAN HAVE and WHAT THEY CAN RENT. The owner''s own diagnosis, and the reason this rewrite exists: the offer is not unclear, the page is. Someone who lands on the hours first concludes the whole service is rented. Register: plain, specific, calm; every technical term explained in the same breath; state every fact in its positive form (define-by-negation is banned in all its grammatical forms); no em-dash asides; no hype; no competitor comparisons and no jargon being knocked down. British English. SEVEN SECTIONS, one subject each, in this order, and no two sections may open on the same claim:
(1) hero, subject THE TWO WAYS TO PAY US AND WHAT EACH BUYS: one plain statement that there are two different transactions here, a model of your own to keep, or an hour on a machine to try one, with an invitation to read on.
(2) generic-text-block, subject WHAT YOU CAN HAVE: the fine-tune you buy. The trained model file is yours to download and keep; it runs on your own hardware, offline, for as long as you like, and nothing about it expires. Say what is included in the plain terms the technical-details page uses. State the price exactly as the registered fact records it.
(3) generic-text-block, subject WHAT YOU CAN RENT: the playground hour, a booked session on a machine with your own model already loaded, so you can try it on real work before you rely on it. Priced per hour by machine size, because a bigger machine answers faster and costs more to run. Say plainly that the demonstration on /playground.html is free and needs no booking.
(4) generic-text-block, subject WHAT ACTUALLY DRIVES THE COST: open on the idea that what drives the cost is scope rather than buzzwords. Then the honest list of what moves a price: how much of your writing there is, how many rounds of training it takes to sound right, how big a model the job needs, and how quickly you need it to answer.
(5) features, subject THE PRICE LIST ITSELF: the things we sell, side by side, each with what you get and what it costs. Exactly three entries: the fine-tuned model you keep at the registered price; an hour in the playground on the smaller machine; an hour on the larger machine. State each price exactly as the registered facts record them, and say which of the three is a purchase and which two are rentals.
(6) faq, subject THE QUESTIONS THIS PAGE RAISES: do I own the model (yes, the file is yours to keep); what happens when the booked hour ends (the model is still yours, the machine stops); can I buy more hours (yes); do I need an expensive computer to run it (no, that is what the file is for); what if the model is not good enough.
(7) call-to-action: low-pressure, book a discovery call at /contact.html, or read the offer at /your-own-model.html.
NUMBERS DISCIPLINE, and this page is where it matters most: state ONLY the registered prices, which are ft-price-99, ft-hour-small and ft-hour-large. Do not invent a discount, a package, a monthly figure, a turnaround time, a percentage or any other number. Do not state the market-comparison figure on this page; it belongs on the offer page. If a number you want is not in the facts, write the sentence without it.'
    ),
    'gap_plan_pricing_rent_vs_own_' || v_site::text, gen_random_uuid());

  SELECT count(*) INTO n FROM site_work_items WHERE page_id=v_page AND status='triaged' AND item_type='needs_content_page'
    AND NOT (spec ? 'mode') AND NOT (spec ? 'not_dispatchable') AND jsonb_array_length(spec->'sections')=7
    AND spec->>'suggestion' LIKE '%ft-hour-small%' AND spec->>'suggestion' LIKE '%ft-hour-large%' AND spec->>'suggestion' NOT LIKE '%—%';
  IF n <> 1 THEN RAISE EXCEPTION 'post: expected 1 clean 7-section brief naming both hour facts, found %', n; END IF;
  SELECT count(*) INTO n FROM pages WHERE id=v_page AND build_status='planned';
  IF n <> 1 THEN RAISE EXCEPTION 'post: page not planned'; END IF;
  RAISE NOTICE 'pricing page: planned, 7-section rent-vs-own brief queued, prices taken from the register by fact id';
END $$;
COMMIT;
