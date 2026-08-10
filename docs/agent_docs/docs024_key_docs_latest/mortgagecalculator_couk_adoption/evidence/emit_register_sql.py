#!/usr/bin/env python3
"""Emit the SQL that replaces mortgagecalculator.co.uk's evidence register with
the per-threshold fact set produced by build_facts.py.

Two things this deliberately does NOT do:
  * it never parses the live register through a typed struct (the fleet landmine
    of 2026-08-09: EvidenceBase/EvidenceFact model a subset, so a typed
    round-trip silently deletes every citation, writer_line, unit and
    staleness_days). The governing_rule is carried across as raw text read from
    the DB, and the facts are rebuilt from source.
  * it does not hand-write the writer_block. The block is composed here by the
    same rule as Go's composeWriterBlock — header, "- " per writer_line, {value}
    replaced by formatEvidenceNumber — so the next scheduled sweep recomposes
    byte-identical text instead of showing a spurious change.
"""
import json
import sys

SITE = "62b5978e-4271-4589-8e00-4baebfc0447c"

facts = json.load(open(sys.argv[1], encoding="utf-8"))
governing_rule = open(sys.argv[2], encoding="utf-8").read().strip()


def format_evidence_number(v):
    """Mirror of formatEvidenceNumber: whole numbers get thousands separators."""
    if float(v) != int(v):
        return ("%.4f" % v).rstrip("0").rstrip(".")
    return f"{int(v):,}"


def compose_writer_block(facts):
    numbers = []
    for f in facts:
        line = (f.get("writer_line") or "").strip()
        if not line:
            continue  # Go: a fact without a writer_line is omitted, never auto-phrased
        line = line.replace("{value}", format_evidence_number(f["value"]))
        numbers.append("- " + line)
    if not numbers:
        return ""
    return ("NUMBERS (state only these, with their listed meaning; dated "
            "snapshots up to a listed live count are fine):\n" + "\n".join(numbers))


data = {
    "facts": facts,
    "writer_block": compose_writer_block(facts),
    "governing_rule": governing_rule,
    "writer_block_managed": True,
}

blob = json.dumps(data, ensure_ascii=False, indent=2)

print(f"""-- A1 (PLAN_2026-08-09 Piece 4(i)) -- one SDLT fact per threshold and per rate.
--
-- 4 facts -> {len(facts)}. No schema change: every `value` stays scalar, and each band
-- edge now carries its OWN verbatim GOV.UK quote, so an oracle can read a band
-- table off the register instead of parsing prose out of a `claim` string.
--
-- Every quote was lifted programmatically from the visible text produced by the
-- REAL Go extractor (datahelpers.VisibleTextFromHTML) and re-checked with
-- datahelpers.QuoteFoundInText against the live pages before this file was
-- written; the check was induced red first (a quote reading 126,000 instead of
-- 125,000 reports NOTFOUND and refuses to emit). Nothing here was retyped.
--
-- Retired: `sdlt-standard-bands` (value 12, bands in prose) and
-- `sdlt-ftb-nil-rate` (value 300000, two rules in one claim) -- both are now
-- covered by granular facts. Checked before retiring: zero references in
-- doc_plans, site_work_items or page_components; only this lane's own docs.
-- Kept by id: `sdlt-ftb-relief-cap` and `sdlt-additional-surcharge` (already
-- one number each, and named in the plan's induced-drift test).
-- NEW: `sdlt-additional-surcharge-floor` (40000, cited to the higher-rates
-- guidance) -- registered because the 2026-08-10 rebuild DROPPED the tool's
-- 40,000 floor, which was true law absent from the register.

BEGIN;

UPDATE site_specs SET is_current = false
 WHERE site_id = '{SITE}' AND aspect = 'evidence_base' AND is_current;

INSERT INTO site_specs (site_id, aspect, data, source, created_by, is_current, pinned)
VALUES (
  '{SITE}',
  'evidence_base',
  $spec${blob}$spec$::jsonb,
  'adoption',
  'mortgagecalculator-adoption-lane: A1 per-threshold SDLT facts',
  true,
  true
);

DO $$
DECLARE n_facts int; n_cited int; n_current int; n_ids int;
BEGIN
  SELECT jsonb_array_length(data->'facts'),
         (SELECT count(*) FROM jsonb_array_elements(data->'facts') f
           WHERE f->'source'->'citation'->>'quote' <> ''),
         (SELECT count(DISTINCT f->>'id') FROM jsonb_array_elements(data->'facts') f)
    INTO n_facts, n_cited, n_ids
    FROM site_specs
   WHERE site_id = '{SITE}' AND aspect = 'evidence_base' AND is_current;

  SELECT count(*) INTO n_current FROM site_specs
   WHERE site_id = '{SITE}' AND aspect = 'evidence_base' AND is_current;

  IF n_facts <> {len(facts)} THEN
    RAISE EXCEPTION 'expected {len(facts)} facts, found %', n_facts;
  END IF;
  IF n_cited <> {len(facts)} THEN
    RAISE EXCEPTION 'expected every fact to carry a quote, only % do', n_cited;
  END IF;
  IF n_ids <> {len(facts)} THEN
    RAISE EXCEPTION 'fact ids are not unique: % distinct of {len(facts)}', n_ids;
  END IF;
  IF n_current <> 1 THEN
    RAISE EXCEPTION 'expected exactly one current evidence_base row, found %', n_current;
  END IF;
END $$;

COMMIT;
""")
