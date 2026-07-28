-- Register the Thames Water capital structure as at 31 March 2024.
--
-- SHAPE DECISION, and it is the interesting part. The task was "register a real
-- Thames series". The verified figures do NOT form one:
--   * every debt figure is at ONE date (31 March 2024) - that is a capital
--     structure snapshot, which is evidence-chart's job (magnitudes), not
--     evidence-timeseries' (one measure over time);
--   * the two percentage figures measure DIFFERENT THINGS - 23% is "average
--     yearly bills", 35% is "above inflation" - so plotting them as one series
--     would compare unlike quantities and invent a trend.
-- Forcing these into a time series would have produced a chart that looked
-- authoritative and was analytically false. Registered as what they are.
--
-- Every quote below was verified LITERALLY PRESENT in the fetched page on
-- 2026-07-28 (fetch, strip tags, substring test) rather than trusted from a
-- model's reading of it. Source was already in this site's register, V5-verified.
BEGIN;

-- Supersede FIRST, as its own statement. Doing it in a CTE alongside the INSERT
-- fails: every CTE branch sees the same snapshot, so the partial unique index
-- (site_id, aspect) WHERE is_current still sees the old row and raises 23505.
CREATE TEMP TABLE _cur ON COMMIT DROP AS
  SELECT * FROM site_specs
  WHERE site_id='a0d7f1ae-f37e-4ea5-b30c-9012d1d14f39' AND aspect='evidence_base' AND is_current;

UPDATE site_specs SET is_current=false, superseded_at=now()
WHERE site_id='a0d7f1ae-f37e-4ea5-b30c-9012d1d14f39' AND aspect='evidence_base' AND is_current;

INSERT INTO site_specs (site_id, aspect, data, source, source_agent, created_by, is_current, pinned, notes)
SELECT cur.site_id, cur.aspect,
  jsonb_set(cur.data, '{facts}', (cur.data->'facts') || $new$[
    {
      "id": "CIT-tw-classa-2024",
      "kind": "metric",
      "claim": "Thames Water's Class A debt totalled around £14.7 billion as at 31 March 2024, as stated in the Paul Weiss client memorandum on the restructuring plan",
      "value": 14.7,
      "unit": "GBP_billion",
      "context_terms": ["class a", "thames"],
      "writer_line": "Class A debt stood at around £{value} billion as at 31 March 2024",
      "verified_at": "2026-07-28",
      "source": {"citation": {
        "publisher": "Paul, Weiss, Rifkind, Wharton & Garrison LLP",
        "title": "Three's a Crowd: The Thames Water Restructuring Plan",
        "url": "https://www.paulweiss.com/insights/client-memos/three-s-a-crowd-the-thames-water-restructuring-plan-s",
        "quote": "the Class A debt totalling around £14.7 billion",
        "accessed": "2026-07-28"}}
    },
    {
      "id": "CIT-tw-classb-2024",
      "kind": "metric",
      "claim": "Thames Water's Class B debt totalled £1.4 billion as at 31 March 2024, as stated in the Paul Weiss client memorandum on the restructuring plan",
      "value": 1.4,
      "unit": "GBP_billion",
      "context_terms": ["class b", "thames"],
      "writer_line": "Class B debt stood at £{value} billion as at 31 March 2024",
      "verified_at": "2026-07-28",
      "source": {"citation": {
        "publisher": "Paul, Weiss, Rifkind, Wharton & Garrison LLP",
        "title": "Three's a Crowd: The Thames Water Restructuring Plan",
        "url": "https://www.paulweiss.com/insights/client-memos/three-s-a-crowd-the-thames-water-restructuring-plan-s",
        "quote": "the Class B debt totalling £1.4 billion",
        "accessed": "2026-07-28"}}
    },
    {
      "id": "CIT-tw-wbs-total-2024",
      "kind": "metric",
      "claim": "The drawn facilities within Thames Water's whole-business securitisation totalled approximately £16.3 billion as at 31 March 2024, as stated in the Paul Weiss client memorandum",
      "value": 16.3,
      "unit": "GBP_billion",
      "context_terms": ["drawn facilities", "securitisation", "wbs", "thames"],
      "writer_line": "Drawn facilities within the securitisation totalled about £{value} billion as at 31 March 2024",
      "verified_at": "2026-07-28",
      "source": {"citation": {
        "publisher": "Paul, Weiss, Rifkind, Wharton & Garrison LLP",
        "title": "Three's a Crowd: The Thames Water Restructuring Plan",
        "url": "https://www.paulweiss.com/insights/client-memos/three-s-a-crowd-the-thames-water-restructuring-plan-s",
        "quote": "the various drawn facilities within the WBS totalled c.£16.3 billion",
        "accessed": "2026-07-28"}}
    },
    {
      "id": "CIT-tw-mtm-2024",
      "kind": "metric",
      "claim": "Thames Water carried an additional mark-to-market exposure of approximately £1.7 billion as at 31 March 2024, as stated in the Paul Weiss client memorandum",
      "value": 1.7,
      "unit": "GBP_billion",
      "context_terms": ["mark-to-market", "hedging", "thames"],
      "writer_line": "A further mark-to-market exposure of about £{value} billion sat alongside the drawn debt as at 31 March 2024",
      "verified_at": "2026-07-28",
      "source": {"citation": {
        "publisher": "Paul, Weiss, Rifkind, Wharton & Garrison LLP",
        "title": "Three's a Crowd: The Thames Water Restructuring Plan",
        "url": "https://www.paulweiss.com/insights/client-memos/three-s-a-crowd-the-thames-water-restructuring-plan-s",
        "quote": "an additional mark-to-market exposure of c.£1.7 billion as at 31 March 2024",
        "accessed": "2026-07-28"}}
    }
  ]$new$::jsonb),
  'oufe-workstream', 'session', 'oufe-workstream', true, COALESCE(cur.pinned,false),
  'Adds the 31 March 2024 capital structure (Class A, Class B, WBS drawn total, MTM). Quotes verified literally present in the fetched page, not taken from a model summary. Registered as dated metrics rather than a series because every figure shares one date.'
FROM _cur cur;

COMMIT;

SELECT jsonb_array_length(data->'facts')||' facts now registered'
FROM site_specs
WHERE site_id='a0d7f1ae-f37e-4ea5-b30c-9012d1d14f39' AND aspect='evidence_base' AND is_current;
