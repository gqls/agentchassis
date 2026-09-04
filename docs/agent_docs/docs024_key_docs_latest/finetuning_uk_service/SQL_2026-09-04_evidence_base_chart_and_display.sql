-- SQL_2026-09-04_evidence_base_chart_and_display.sql — finetuning.uk: the three data edits the
-- evidence-chart component needs before the £99 vs ~$5,000 comparison can be drawn (infographics lane,
-- 2026-09-04, verified here at the live row):
--   (1) evidence_base has NO `charts` key (keys today: facts, audit_doc, writer_block, banned_claims,
--       governing_rule) and evidence-chart REQUIRES it with on_missing=skip_section — mount it without
--       one and the section silently renders nothing;
--   (2) neither fact carries `display`, and the template renders `{{if $f.display}}` — without it the
--       bare value shows (99 / 5000);
--   (3) `tolerance` is stored (exact / approximate) and the template NEVER READS IT — grepped, zero
--       matches — so "the approximate side must read as approximate" is carried by `display` or by
--       nothing.
--
-- ⚠ AND ONE THING NEITHER LANE RAISED, which is why the caption below is not optional: THE TWO FACTS
-- ARE IN DIFFERENT CURRENCIES. £99 and $5,000 cannot share a bar axis honestly without a conversion,
-- and no exchange rate is a registered fact here. The bars are therefore INDICATIVE of magnitude, the
-- display strings carry their own currency symbols, and the caption says so in words. At ~0.79 £/$ the
-- true ratio is about 40:1 rather than the 50:1 the raw numbers imply — the visual point survives
-- either way, but the page must not imply a like-for-like axis.
--
-- Reversible: the previous evidence_base row is superseded, not deleted (is_current flips), and this
-- file writes a NEW current row the same way the estate's other evidence edits do.
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
  IF v_data ? 'charts' THEN RAISE EXCEPTION 'pre-flight: charts already exists — read it before overwriting'; END IF;

  -- (2)+(3): display strings on the two price facts, carrying the currency AND the tolerance
  v_data := jsonb_set(v_data, '{facts}', (
    SELECT jsonb_agg(CASE
      WHEN f->>'id' = 'ft-price-99'      THEN f || '{"display":"£99"}'::jsonb
      WHEN f->>'id' = 'ft-market-anchor' THEN f || '{"display":"~$5,000"}'::jsonb
      ELSE f END)
    FROM jsonb_array_elements(v_data->'facts') f));

  -- (1): one chart, two points, scaled by the larger fact; the caption states the currency mismatch
  v_data := v_data || jsonb_build_object('charts', jsonb_build_array(jsonb_build_object(
    'id', 'chart-price-vs-market',
    'title', 'What a fine-tuned model costs',
    'caption', 'Our price is in pounds; the consultancy figure is as quoted, in US dollars. The bars show the difference in scale, not a like-for-like conversion.',
    'max_fact_id', 'ft-market-anchor',
    'tone', 'neutral',
    'source_note', 'Both figures are registered facts: our own price, and a 2026-08-18 survey of done-for-you fine-tuning consultancies.',
    'points', jsonb_build_array(
      jsonb_build_object('fact_id','ft-price-99','label','A fine-tune from us'),
      jsonb_build_object('fact_id','ft-market-anchor','label','Done-for-you consultancy, typical start')
    ))));

  UPDATE site_specs SET is_current=false, updated_at=NOW() WHERE site_id=v_site AND aspect='evidence_base' AND is_current;
  INSERT INTO site_specs (site_id, aspect, data, source, created_by, is_current, notes)
  VALUES (v_site, 'evidence_base', v_data, v_source, v_by, true,
          'finetuning_uk_service_lane 2026-09-04: added charts[1] (chart-price-vs-market) and display strings on ft-price-99 / ft-market-anchor so evidence-chart can render the comparison; facts unchanged at 10');

  SELECT count(*) INTO n FROM site_specs WHERE site_id=v_site AND aspect='evidence_base' AND is_current;
  IF n<>1 THEN RAISE EXCEPTION 'post: % current evidence_base rows', n; END IF;
  SELECT count(*) INTO n FROM site_specs ss, jsonb_array_elements(ss.data->'facts') f WHERE ss.site_id=v_site AND ss.aspect='evidence_base' AND ss.is_current AND f ? 'display';
  IF n<>2 THEN RAISE EXCEPTION 'post: expected 2 facts with display, found %', n; END IF;
  SELECT count(*) INTO n FROM site_specs WHERE site_id=v_site AND aspect='evidence_base' AND is_current AND jsonb_array_length(data->'charts')=1 AND jsonb_array_length(data->'facts')=10;
  IF n<>1 THEN RAISE EXCEPTION 'post: charts/facts shape wrong'; END IF;
  RAISE NOTICE 'evidence_base: 1 chart added, display strings on the two price facts, 10 facts intact';
END $$;
COMMIT;
