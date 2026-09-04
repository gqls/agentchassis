-- SQL_2026-09-04_writer_block_admit_hour_prices.sql — finetuning.uk: let the writer STATE the two
-- booked-hour prices, which registering them as facts did NOT do.
--
-- ⚠ THE FINDING, measured at the rendered prompt after the pricing page came back without them:
-- `evidence_base.facts[]` and `evidence_base.writer_block` are DIFFERENT THINGS, and the writer's
-- "Verified Facts (the ONLY numbers you may assert)" block is built from the PROSE in `writer_block`,
-- not from the facts array. Measured on the 14:34Z build: the prompt contained `ft-hour-small`
-- (because this lane's brief named the fact id) and did NOT contain `3.15` anywhere. So the writer
-- could not have stated the price even if it had wanted to, and it correctly wrote the page without it.
-- **Registering a fact is necessary and NOT sufficient. A fact absent from writer_block is invisible
-- to every writer on the site.** Landmine filed.
--
-- This edit adds one clause to the NUMBERS sentence. Nothing else in writer_block changes: the
-- never-state list, the qualitative-outcomes rule and the licence rule are untouched, asserted below.
\set ON_ERROR_STOP on
BEGIN;
DO $$
DECLARE
  v_site uuid := '1368e337-dd1d-4799-bbb3-8221a1b79bcc';
  v_data jsonb; v_source text; v_by text; v_old text; v_new text; n int;
  v_anchor text := 'used only as a market anchor (facts ft-market-anchor).';
  v_add text := ' The two booked playground hour prices may also be stated, exactly as recorded: £3.15 per hour on the standard machine (facts ft-hour-small) and £6.65 per hour on the larger machine (facts ft-hour-large). The free demonstration on /playground.html costs nothing and no figure attaches to it.';
BEGIN
  SELECT data, source, created_by INTO v_data, v_source, v_by FROM site_specs WHERE site_id=v_site AND aspect='evidence_base' AND is_current;
  v_old := v_data->>'writer_block';
  IF v_old IS NULL THEN RAISE EXCEPTION 'no writer_block'; END IF;
  IF position(v_anchor in v_old) = 0 THEN RAISE EXCEPTION 'pre-flight: the market-anchor sentence this edit appends to is not present verbatim'; END IF;
  IF position('ft-hour-small' in v_old) > 0 THEN RAISE EXCEPTION 'pre-flight: writer_block already names the hour prices'; END IF;

  v_new := replace(v_old, v_anchor, v_anchor || v_add);
  IF length(v_new) <> length(v_old) + length(v_add) THEN RAISE EXCEPTION 'pre-flight: the anchor matched more than once'; END IF;

  v_data := jsonb_set(v_data, '{writer_block}', to_jsonb(v_new));
  UPDATE site_specs SET is_current=false, updated_at=NOW() WHERE site_id=v_site AND aspect='evidence_base' AND is_current;
  INSERT INTO site_specs (site_id, aspect, data, source, created_by, is_current, notes)
  VALUES (v_site, 'evidence_base', v_data, v_source, v_by, true,
          'finetuning_uk_service_lane 2026-09-04: writer_block now admits ft-hour-small and ft-hour-large; registering the facts alone left their values absent from every writer prompt');

  SELECT count(*) INTO n FROM site_specs WHERE site_id=v_site AND aspect='evidence_base' AND is_current
    AND data->>'writer_block' LIKE '%ft-hour-small%' AND data->>'writer_block' LIKE '%ft-hour-large%'
    AND data->>'writer_block' LIKE '%£3.15%' AND data->>'writer_block' LIKE '%£6.65%'
    AND data->>'writer_block' LIKE '%NOT TRACKED, NEVER STATE%'
    AND data->>'writer_block' LIKE '%ft-market-anchor%'
    AND jsonb_array_length(data->'facts')=12 AND jsonb_array_length(data->'charts')=1;
  IF n<>1 THEN RAISE EXCEPTION 'post: writer_block or the rest of the evidence base is not as expected'; END IF;
  RAISE NOTICE 'writer_block admits the two hour prices; never-state list, licence rule, 12 facts and 1 chart intact';
END $$;
COMMIT;
