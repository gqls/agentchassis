-- Ofwat PR24 final determination for Thames Water: the bill negotiation.
--
-- SHAPE DECISION AGAIN, and again it is not a time series. Ofwat's own Figure 1.1
-- plots four bars, and three of them are the SAME YEAR (2029-30) under different
-- decision-makers. That is a scenario comparison - what the company asked for,
-- what the regulator first offered, and where it landed - so it belongs in
-- evidence-chart, not evidence-timeseries.
--
-- HOW THE FIGURES WERE VERIFIED, because they come from a chart rather than a
-- sentence. pdftotext -layout preserves column positions, so each data label maps
-- to the axis label beneath it. That mapping is then CORROBORATED ARITHMETICALLY
-- by an independent sentence in the same document: 588 - 436 = 152, and the prose
-- states "increase average household bills by £152 from 2024-2025 to 2029-30".
-- A column mapping that reproduces a separately-stated total is not a guess.
-- Each value appears exactly once in the document, so there is no ambiguity about
-- which label a figure belongs to.
BEGIN;

CREATE TEMP TABLE _cur ON COMMIT DROP AS
  SELECT * FROM site_specs
  WHERE site_id='a0d7f1ae-f37e-4ea5-b30c-9012d1d14f39' AND aspect='evidence_base' AND is_current;

UPDATE site_specs SET is_current=false, superseded_at=now()
WHERE site_id='a0d7f1ae-f37e-4ea5-b30c-9012d1d14f39' AND aspect='evidence_base' AND is_current;

INSERT INTO site_specs (site_id, aspect, data, source, source_agent, created_by, is_current, pinned, notes)
SELECT cur.site_id, cur.aspect,
  jsonb_set(cur.data, '{facts}', (cur.data->'facts') || $new$[
    {"id":"CIT-ofwat-bill-2024-25","kind":"metric","value":436,"unit":"GBP",
     "claim":"Ofwat's final determination shows Thames Water's average household bill at £436 in 2024-25, before inflation, as the baseline in Figure 1.1",
     "context_terms":["average household bill","baseline","thames"],
     "writer_line":"The average household bill stood at £{value} in 2024-25, before inflation",
     "verified_at":"2026-07-28",
     "source":{"citation":{"publisher":"Ofwat","title":"Overview of Thames Water's PR24 final determination","url":"https://www.ofwat.gov.uk/wp-content/uploads/2024/12/Overview-of-Thames-Waters-PR24-final-determination-1.pdf","quote":"Figure 1.1 Average household bills for Thames Water, 2024-25 and 2029-30, before inflation","accessed":"2026-07-28","published":"2024-12"}}},
    {"id":"CIT-ofwat-bill-draft","kind":"metric","value":535,"unit":"GBP",
     "claim":"Ofwat's draft decision would have set Thames Water's average household bill at £535 by 2029-30, before inflation, per Figure 1.1 of the final determination overview",
     "context_terms":["draft decision","average household bill","thames"],
     "writer_line":"Ofwat's draft decision would have allowed £{value} by 2029-30",
     "verified_at":"2026-07-28",
     "source":{"citation":{"publisher":"Ofwat","title":"Overview of Thames Water's PR24 final determination","url":"https://www.ofwat.gov.uk/wp-content/uploads/2024/12/Overview-of-Thames-Waters-PR24-final-determination-1.pdf","quote":"Figure 1.1 Average household bills for Thames Water, 2024-25 and 2029-30, before inflation","accessed":"2026-07-28","published":"2024-12"}}},
    {"id":"CIT-ofwat-bill-proposal","kind":"metric","value":667,"unit":"GBP",
     "claim":"Thames Water proposed an average household bill of £667 by 2029-30, before inflation, per Figure 1.1 of Ofwat's final determination overview",
     "context_terms":["proposal","proposed","average household bill","thames"],
     "writer_line":"Thames Water asked for £{value} by 2029-30",
     "verified_at":"2026-07-28",
     "source":{"citation":{"publisher":"Ofwat","title":"Overview of Thames Water's PR24 final determination","url":"https://www.ofwat.gov.uk/wp-content/uploads/2024/12/Overview-of-Thames-Waters-PR24-final-determination-1.pdf","quote":"average bills will be lower than those proposed by Thames Water in response to our draft decision","accessed":"2026-07-28","published":"2024-12"}}},
    {"id":"CIT-ofwat-bill-final","kind":"metric","value":588,"unit":"GBP",
     "claim":"Ofwat's final decision sets Thames Water's average household bill at £588 by 2029-30, before inflation, per Figure 1.1 of the final determination overview",
     "context_terms":["final decision","final determination","average household bill","thames"],
     "writer_line":"Ofwat's final decision allows £{value} by 2029-30",
     "verified_at":"2026-07-28",
     "source":{"citation":{"publisher":"Ofwat","title":"Overview of Thames Water's PR24 final determination","url":"https://www.ofwat.gov.uk/wp-content/uploads/2024/12/Overview-of-Thames-Waters-PR24-final-determination-1.pdf","quote":"Figure 1.1 Average household bills for Thames Water, 2024-25 and 2029-30, before inflation","accessed":"2026-07-28","published":"2024-12"}}},
    {"id":"CIT-ofwat-bill-increase","kind":"metric","value":152,"unit":"GBP",
     "claim":"Ofwat's final determination increases Thames Water's average household bill by £152 between 2024-25 and 2029-30, before inflation",
     "context_terms":["increase","rise","average household bill","thames"],
     "writer_line":"The determination raises the average household bill by £{value} across the period, before inflation",
     "verified_at":"2026-07-28",
     "source":{"citation":{"publisher":"Ofwat","title":"Overview of Thames Water's PR24 final determination","url":"https://www.ofwat.gov.uk/wp-content/uploads/2024/12/Overview-of-Thames-Waters-PR24-final-determination-1.pdf","quote":"this will increase average household bills by £152 from 2024-2025 to 2029-30 for Thames Water customers, before inflation","accessed":"2026-07-28","published":"2024-12"}}},
    {"id":"CIT-ofwat-revenue-2025-30","kind":"metric","value":16.4,"unit":"GBP_billion",
     "claim":"Ofwat's final determination allows Thames Water to collect £16.4 billion through bills over the 2025-30 period",
     "context_terms":["collect","through bills","revenue","thames"],
     "writer_line":"The determination allows £{value} billion to be collected through bills over 2025-30",
     "verified_at":"2026-07-28",
     "source":{"citation":{"publisher":"Ofwat","title":"Overview of Thames Water's PR24 final determination","url":"https://www.ofwat.gov.uk/wp-content/uploads/2024/12/Overview-of-Thames-Waters-PR24-final-determination-1.pdf","quote":"allows Thames Water to collect £16.4 billion through bills","accessed":"2026-07-28","published":"2024-12"}}}
  ]$new$::jsonb),
  'oufe-workstream','session','oufe-workstream', true, COALESCE(cur.pinned,false),
  'Ofwat PR24 final determination for Thames Water. Figure 1.1 values mapped by pdftotext -layout column position and corroborated arithmetically (588-436=152 matches the separately stated increase). Registered as a scenario comparison, not a series: three of the four are the same year.'
FROM _cur cur;

COMMIT;
SELECT jsonb_array_length(data->'facts')||' facts registered'
FROM site_specs WHERE site_id='a0d7f1ae-f37e-4ea5-b30c-9012d1d14f39' AND aspect='evidence_base' AND is_current;
