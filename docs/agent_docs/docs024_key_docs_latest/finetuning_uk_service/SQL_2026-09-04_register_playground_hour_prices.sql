-- SQL_2026-09-04_register_playground_hour_prices.sql — finetuning.uk: register the two booked-hour
-- prices as evidence_base facts, so the pricing page is ALLOWED to state them.
--
-- Owner, 2026-09-04: "I confirm those prices" — the table put to him was £3.15/hour on the small GPU and
-- £6.65/hour on the big one, derived as HIS rule £1.50 fixed + 6× the vendor's invoiced cost.
--
-- HOW THE NUMBERS WERE REACHED, recorded in the fact's own source so no future reader has to re-derive:
--   vendor invoice 2026-08-18, billed per minute: a6000 $0.35/hr, a100xl $1.09/hr
--   FX £0.787/$ [ASSUMED 2026-09-04, not a registered fact] -> £0.276 and £0.858 cost
--   ×6 -> £1.65 and £5.15; +£1.50 fixed -> £3.15 and £6.65
-- The PRICE is exact because it is a decision. The DERIVATION used an assumed exchange rate, which is
-- why the source says so: if the rate moves, the price does not follow automatically — it is re-decided.
\set ON_ERROR_STOP on
BEGIN;
DO $$
DECLARE
  v_site uuid := '1368e337-dd1d-4799-bbb3-8221a1b79bcc';
  v_data jsonb; v_source text; v_by text; n int;
BEGIN
  SELECT data, source, created_by INTO v_data, v_source, v_by FROM site_specs WHERE site_id=v_site AND aspect='evidence_base' AND is_current;
  IF v_data IS NULL THEN RAISE EXCEPTION 'no current evidence_base'; END IF;
  IF jsonb_array_length(v_data->'facts') <> 10 THEN RAISE EXCEPTION 'pre-flight: expected 10 facts, found %', jsonb_array_length(v_data->'facts'); END IF;
  IF EXISTS (SELECT 1 FROM jsonb_array_elements(v_data->'facts') f WHERE f->>'id' LIKE 'ft-hour-%') THEN RAISE EXCEPTION 'pre-flight: an ft-hour-* fact already exists'; END IF;

  v_data := jsonb_set(v_data, '{facts}', (v_data->'facts') || jsonb_build_array(
    jsonb_build_object(
      'id','ft-hour-small','kind','price','value',3.15,'display','£3.15',
      'claim','price per booked hour in the playground on the smaller GPU',
      'tolerance','exact','verified_at','2026-09-04',
      'source', jsonb_build_object('attested_by','owner decision 2026-09-04 ("I confirm those prices"), applying his rule of £1.50 fixed + 6× cost to the vendor-invoiced a6000 rate of $0.35/hr (invoice 2026-08-18, billed per minute) at an assumed £0.787/$'),
      'writer_line','an hour in the playground costs £{value} on the smaller machine',
      'context_terms', jsonb_build_array('£3.15','3.15','hour','hourly','playground','per hour')),
    jsonb_build_object(
      'id','ft-hour-large','kind','price','value',6.65,'display','£6.65',
      'claim','price per booked hour in the playground on the larger GPU',
      'tolerance','exact','verified_at','2026-09-04',
      'source', jsonb_build_object('attested_by','owner decision 2026-09-04 ("I confirm those prices"), applying his rule of £1.50 fixed + 6× cost to the vendor-invoiced a100xl rate of $1.09/hr (invoice 2026-08-18, billed per minute) at an assumed £0.787/$'),
      'writer_line','an hour on the larger machine, which answers faster, costs £{value}',
      'context_terms', jsonb_build_array('£6.65','6.65','hour','hourly','playground','larger','faster'))
  ));

  UPDATE site_specs SET is_current=false, updated_at=NOW() WHERE site_id=v_site AND aspect='evidence_base' AND is_current;
  INSERT INTO site_specs (site_id, aspect, data, source, created_by, is_current, notes)
  VALUES (v_site, 'evidence_base', v_data, v_source, v_by, true,
          'finetuning_uk_service_lane 2026-09-04: registered ft-hour-small (£3.15) and ft-hour-large (£6.65) on the owner''s confirmation, so the pricing page may state them; charts[1] and the 10 prior facts unchanged');

  SELECT count(*) INTO n FROM site_specs WHERE site_id=v_site AND aspect='evidence_base' AND is_current AND jsonb_array_length(data->'facts')=12 AND jsonb_array_length(data->'charts')=1;
  IF n<>1 THEN RAISE EXCEPTION 'post: expected 12 facts and 1 chart'; END IF;
  SELECT count(*) INTO n FROM site_specs ss, jsonb_array_elements(ss.data->'facts') f WHERE ss.site_id=v_site AND ss.aspect='evidence_base' AND ss.is_current AND f->>'id' IN ('ft-hour-small','ft-hour-large') AND (f->>'value')::numeric > 0 AND f ? 'display';
  IF n<>2 THEN RAISE EXCEPTION 'post: the two hour facts are not both present with a display'; END IF;
  RAISE NOTICE 'registered ft-hour-small £3.15 and ft-hour-large £6.65; evidence_base now 12 facts, 1 chart';
END $$;
COMMIT;
