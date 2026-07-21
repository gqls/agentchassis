-- 183_tool_recreation_no_invented_data.sql — close the prompt half of
-- /bugs_open/020 (tool-recreation invents a dataset when the original tool was
-- data-backed). DB-only; effective immediately, no image roll.
--
-- BUG 020 (vetcomparison.uk, 2026-07-18): tool-recreation-handler rebuilt a real
-- practice directory (2,109 verified practices loaded from /data/vet-full-index.json)
-- as a self-contained widget that GENERATES synthetic practice names + postcodes
-- in the browser (Mulberry32 seeded RNG, TOWNS/PREFIXES/SUFFIXES arrays crossed
-- by makePostcode()/buildData()). It shipped live and served fabricated records
-- to visitors. Every work item reported `complete`.
--
-- Two prompt defects this migration fixes (root cause (a)+(b) in the case file):
--   (a) recreate_tool has NO data-dependency contract. Its prompt demands a
--       "self-contained" tool with all JS/CSS embedded; the model reads that as
--       "embed the data too", and when the tool's whole behaviour IS its data
--       it invents records to make search/filter/pagination demonstrably work.
--       The original fetch() target is visible in the reference source the model
--       already receives, but nothing tells it to preserve it.
--   (b) rule 9 ("No fake data or dummy outputs — calculations must be
--       mathematically correct") is scoped to ARITHMETIC. Read in context
--       (rules 7-10 are all about function completeness), it does not forbid
--       inventing RECORDS, and the model did not read it that way.
--
-- FIX (config only — the Go fabrication gate is the separate belt-and-braces net):
--   recreate_tool:
--     1. Insert a prominent "## Data Integrity" section BEFORE "## Requirements"
--        that forbids inventing/seeding/generating records, clarifies that
--        "self-contained" means CODE not DATA, requires preserving the original
--        data source, and mandates an honest empty state when no source exists.
--     2. Rewrite rule 9 to bind DATA, not just maths, and point at the new section.
--   analyze_tool:
--     3. Add a "data_source" field to the JSON spec schema, capturing the exact
--        fetch/XHR/API/`/data/*.json` target verbatim so the contract flows into
--        recreate_tool (which renders the whole analysis JSON via toJSON).
--     4. Point task item 8 at that field so the analyst captures the source.
--
-- Live models (2026-07-21): analyze_tool = claude-sonnet-5 @ 8000,
--                           recreate_tool = claude-opus-4-8 @ 64000.
-- The two prompt anchors were confirmed unique against the LIVE row before
-- writing this. Standing rule: snapshot_agent opens the transaction. The DO
-- block below fails the whole transaction if any anchor did not match (a no-op
-- replace), so a bad anchor rolls back cleanly rather than applying nothing.

BEGIN;

SELECT snapshot_agent('tool-recreation-handler', '183_tool_recreation_no_invented_data.sql: pre-update');

-- (1) recreate_tool: prepend the Data Integrity section to "## Requirements".
UPDATE agent_definitions
SET default_config = jsonb_set(
      default_config,
      '{workflow,steps,recreate_tool,config,prompt_template}',
      to_jsonb(replace(
        default_config #>> '{workflow,steps,recreate_tool,config,prompt_template}',
        '## Requirements',
        $di$## Data Integrity — do NOT invent data (this overrides every requirement below)
This is the most important rule in this prompt. It has been violated in production: a recreation replaced a real directory of ~2,100 verified records with a client-side generator of fake names and postcodes, and it shipped to a live public site.

- NEVER invent, generate, synthesise, sample, seed or hard-code example records, rows, entries, names, addresses, postcodes, phone numbers, prices or any dataset. Do NOT build a list of realistic-looking items so that search, filtering, sorting or pagination appear to "work".
- "Self-contained" (rules 2 and 3 below) refers to CODE only — embed all JS and CSS. It does NOT permit embedding a DATASET.
- If the original tool loaded its data from a source — a fetch()/XHR call, a /data/*.json file, or an API endpoint (see the "data_source" field in the specification above and the original source code) — the recreation MUST load from that SAME source, unchanged. Keep the fetch and rebuild the UI around it.
- If the tool needs data and you have NO reachable source, render a clear, honest empty state (for example "No data available yet") and stop. An empty widget is correct; an invented one is a serious fault.
- Forbidden as ways to populate real-world content: a seeded pseudo-random generator (Mulberry32, an LCG, Math.imul-based hashing); arrays of name/word fragments crossed together to build labels; a buildData()/generateRecords()/makePostcode()-style function that assembles a list of entities. Genuine game randomness — dice, physics, procedural visuals — is fine; fabricating a directory or catalogue of entities is not.

## Requirements$di$
      ))
    )
WHERE type = 'tool-recreation-handler' AND deleted_at IS NULL;

-- (2) recreate_tool: rewrite rule 9 to bind data, not just arithmetic.
UPDATE agent_definitions
SET default_config = jsonb_set(
      default_config,
      '{workflow,steps,recreate_tool,config,prompt_template}',
      to_jsonb(replace(
        default_config #>> '{workflow,steps,recreate_tool,config,prompt_template}',
        '9. No fake data or dummy outputs — calculations must be mathematically correct',
        $r9$9. No invented data or records — obey the "Data Integrity" section above: preserve the original data source and never fabricate, seed or generate entries. Calculations must still be mathematically correct.$r9$
      ))
    )
WHERE type = 'tool-recreation-handler' AND deleted_at IS NULL;

-- (3) analyze_tool: capture the data source in the JSON spec (before "site_context").
UPDATE agent_definitions
SET default_config = jsonb_set(
      default_config,
      '{workflow,steps,analyze_tool,config,prompt_template}',
      to_jsonb(replace(
        default_config #>> '{workflow,steps,analyze_tool,config,prompt_template}',
        '  "site_context":',
        $ds$  "data_source": {
    "has_external_data": "true if the tool loads its records/data from an external source; false if it is purely computational (a calculator, or a self-contained game with no external dataset)",
    "source": "The EXACT fetch/XHR URL, /data/*.json path, or API endpoint the original loads its data from — copy it verbatim from the source code. Empty string if there is no external dataset.",
    "description": "What the data is and roughly how much (e.g. 'directory of ~2,100 practices loaded from /data/vet-full-index.json')"
  },
  "site_context":$ds$
      ))
    )
WHERE type = 'tool-recreation-handler' AND deleted_at IS NULL;

-- (4) analyze_tool: point task item 8 at the data_source field.
UPDATE agent_definitions
SET default_config = jsonb_set(
      default_config,
      '{workflow,steps,analyze_tool,config,prompt_template}',
      to_jsonb(replace(
        default_config #>> '{workflow,steps,analyze_tool,config,prompt_template}',
        'assets it relies on',
        $i8$assets it relies on. CRITICAL: if the tool loads its records/data from a fetch()/XHR call, a /data/*.json file, or an API, capture the EXACT source URL or path verbatim in the "data_source" field below — the recreation must load from the same place instead of inventing data$i8$
      ))
    )
WHERE type = 'tool-recreation-handler' AND deleted_at IS NULL;

-- Verify: exactly one live row, both prompts patched, old rule 9 gone.
DO $$
DECLARE rt text; at text; n int;
BEGIN
    SELECT count(*) INTO n FROM agent_definitions
      WHERE type='tool-recreation-handler' AND deleted_at IS NULL;
    IF n <> 1 THEN RAISE EXCEPTION 'expected exactly one live row, found %', n; END IF;

    SELECT default_config #>> '{workflow,steps,recreate_tool,config,prompt_template}' INTO rt
      FROM agent_definitions WHERE type='tool-recreation-handler' AND deleted_at IS NULL;
    SELECT default_config #>> '{workflow,steps,analyze_tool,config,prompt_template}' INTO at
      FROM agent_definitions WHERE type='tool-recreation-handler' AND deleted_at IS NULL;

    IF position('## Data Integrity' in rt) = 0 THEN
        RAISE EXCEPTION 'recreate_tool: Data Integrity section missing (anchor "## Requirements" did not match)';
    END IF;
    IF (length(rt)-length(replace(rt,'## Data Integrity','')))/length('## Data Integrity') <> 1 THEN
        RAISE EXCEPTION 'recreate_tool: Data Integrity heading not present exactly once';
    END IF;
    IF position('9. No fake data or dummy outputs' in rt) <> 0 THEN
        RAISE EXCEPTION 'recreate_tool: old rule 9 still present';
    END IF;
    IF position('9. No invented data or records' in rt) = 0 THEN
        RAISE EXCEPTION 'recreate_tool: new rule 9 missing';
    END IF;
    IF position('"data_source"' in at) = 0 THEN
        RAISE EXCEPTION 'analyze_tool: data_source field missing (anchor "site_context" did not match)';
    END IF;
    IF position('instead of inventing data' in at) = 0 THEN
        RAISE EXCEPTION 'analyze_tool: task item 8 data-source instruction missing';
    END IF;

    RAISE NOTICE '183 applied: recreate_tool (Data Integrity + rule 9) and analyze_tool (data_source) patched for bug 020';
END $$;

COMMIT;
