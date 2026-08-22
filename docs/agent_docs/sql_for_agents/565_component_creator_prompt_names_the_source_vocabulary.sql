-- 565_component_creator_prompt_names_the_source_vocabulary.sql
--
-- bugs_open/337 — the prompt half. Tells component-creator's writer which
-- `site_specs` sources it may declare, from the vocabulary the BIRTH GATE
-- itself enforces.
--
-- WHY. The prompt's TIER D already enumerates the query vocabulary exactly
-- ("Known query names (use these EXACTLY — do not invent new ones)"). TIER C,
-- for site data, says only `source: "site_specs.{path}"` with three prose
-- examples and NO aspect list — and the live prompt_template renders no part of
-- `site_specs` anywhere, even though it sits in generate_template's
-- input_fields. So the writer has never been shown a single aspect name and
-- guesses. The live failure is one character: it declared
-- `site_specs.ctas.primary_url` / `ctas.secondary_url` when the aspect is
-- `cta`, which carries EXACTLY primary_url and secondary_url. The gate refuses,
-- the item regenerates unchanged, and the page parks.
--
-- DORMANT UNTIL THE CHASSIS ROLLS, and safe in either order. The block is
-- guarded on {{if .existing_component.aspect_paths}}, a key that only the Go
-- half (commit e1951c24b) emits. Applied before the roll it renders nothing;
-- the Go half without this migration simply writes a key nothing reads. No
-- ordering constraint is claimed (owner ruling 2026-07-29 §2).
--
-- GUARDED ON THE RICHEST KEY (aspect_paths) rather than the shorter
-- `known_aspects`, so the block can never half-render: if leaf-path coverage
-- was unavailable the writer gets today's behaviour, not a heading with an
-- empty list under it.
--
-- THE "NOT LISTED" SENTENCE IS LOAD-BEARING, NOT DECORATION. The sibling
-- failure in the same window wanted `currency_symbol`, which NO site carries in
-- any resolvable form — so a vocabulary list alone would not have saved it. The
-- only correct answers there were `static` with a `£` fallback, or `llm`, and
-- the prompt has to say so.
--
-- COVERAGE NUMBERS ARE THERE TO BE ACTED ON. The gate validates only the FIRST
-- path segment, so a listed aspect with an invented key still passes and then
-- renders blank (bugs_open/309's damage, arriving quietly). Showing real leaf
-- keys narrows that; showing how many sites carry each is what lets the writer
-- decide whether the field may be `required` at all — `cta` is on 4 of 26 sites
-- [MEASURED 2026-08-22].
--
-- Migration 561 is the bugs_open/345 lane's (site_work_items.retry_feedback);
-- agreed with that session directly. 562, 563 and 564 were taken by other lanes
-- while this file was being written, hence 565.
--
-- Rollback: 565_component_creator_prompt_names_the_source_vocabulary_ROLLBACK.sql

SELECT snapshot_agent('component-creator',
  'migration 565: pre-update (bugs_open/337 — source vocabulary into the TIER C prompt block)');

BEGIN;

-- ── Pre-state gates ─────────────────────────────────────────────────────────
-- DO/RAISE, not a verify block of SELECTs: ON_ERROR_STOP ignores a non-empty
-- result set, so only an exception can stop the COMMIT.
DO $$
DECLARE
  live_rows int;
  tpl       text;
  anchor    text := '   TIER C — SITE DATA (source: "site_specs.{path}" or "site_assets.{type}")';
  hits      int;
BEGIN
  SELECT count(*) INTO live_rows
  FROM agent_definitions
  WHERE type = 'component-creator' AND is_active
    AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;
  IF live_rows <> 1 THEN
    RAISE EXCEPTION 'MIGRATION 565: expected exactly 1 live component-creator row, found %', live_rows;
  END IF;

  SELECT default_config ->> 'prompt_template' INTO tpl
  FROM agent_definitions
  WHERE type = 'component-creator' AND is_active
    AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;

  -- The prompt lives at the TOP level of default_config, not under a `prompts`
  -- key (component-creator has none). Absence here means the shape moved.
  IF tpl IS NULL OR length(tpl) = 0 THEN
    RAISE EXCEPTION 'MIGRATION 565: default_config->>''prompt_template'' is absent or empty';
  END IF;

  -- Anchor read from the LIVE row, never from the 093 on-disk seed: the
  -- function-pin block proves this prompt has been edited outside the migration
  -- trail, so a seed-derived anchor cannot be trusted.
  hits := (length(tpl) - length(replace(tpl, anchor, ''))) / length(anchor);
  IF hits <> 1 THEN
    RAISE EXCEPTION 'MIGRATION 565: TIER C anchor found % times, expected exactly 1 — the prompt has drifted; re-read it before applying', hits;
  END IF;

  -- Double-apply refusal.
  IF position('{{if .existing_component.aspect_paths}}' in tpl) > 0 THEN
    RAISE EXCEPTION 'MIGRATION 565: the source-vocabulary block is already present — refusing to insert a second copy';
  END IF;
END $$;

-- ── The change ──────────────────────────────────────────────────────────────
UPDATE agent_definitions
SET default_config = jsonb_set(
      default_config,
      '{prompt_template}',
      to_jsonb(replace(
        default_config ->> 'prompt_template',
        '   TIER C — SITE DATA (source: "site_specs.{path}" or "site_assets.{type}")',
        '   TIER C — SITE DATA (source: "site_specs.{path}" or "site_assets.{type}")' || E'\n' ||
'{{if .existing_component.aspect_paths}}' || E'\n' ||
'     VALID site_specs SOURCES — the first path segment after "site_specs." MUST' || E'\n' ||
'     be one of these EXACTLY. Anything else is refused at store time and the' || E'\n' ||
'     whole component is thrown away, so do not invent one.' || E'\n' ||
'     "(N sites)" is how many sites carry that path. A shared component renders' || E'\n' ||
'     on all of them, so anything short of every site MUST be "required": false' || E'\n' ||
'     with "on_missing": "skip_field", and the template must gate its markup on' || E'\n' ||
'     the field.' || E'\n' ||
'{{.existing_component.aspect_paths}}' || E'\n' ||
'     Only the ASPECT (the first segment) is validated. A listed aspect with a' || E'\n' ||
'     key that is not listed above will pass validation and then resolve to' || E'\n' ||
'     nothing, rendering the section blank with no error — which is worse than' || E'\n' ||
'     being refused. So use a path exactly as listed, or do not use site_specs.' || E'\n' ||
'     IF THE VALUE YOU NEED IS NOT IN THE LIST, DO NOT INVENT A PATH. Use' || E'\n' ||
'     source "static" with a sensible fallback, or source "llm". Some values a' || E'\n' ||
'     section wants (a currency symbol, a CTA destination) exist nowhere in' || E'\n' ||
'     site_specs on any site, and static-with-a-fallback is the correct answer.' || E'\n' ||
'{{end}}'
      )),
      false
    ),
    version    = version + 1,
    updated_at = now()
WHERE type = 'component-creator' AND is_active
  AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;

-- ── Post-state verification ─────────────────────────────────────────────────
DO $$
DECLARE
  tpl        text;
  anchor     text := '   TIER C — SITE DATA (source: "site_specs.{path}" or "site_assets.{type}")';
  open_hits  int;
  path_hits  int;
BEGIN
  SELECT default_config ->> 'prompt_template' INTO tpl
  FROM agent_definitions
  WHERE type = 'component-creator' AND is_active
    AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;

  open_hits := (length(tpl) - length(replace(tpl, '{{if .existing_component.aspect_paths}}', '')))
               / length('{{if .existing_component.aspect_paths}}');
  IF open_hits <> 1 THEN
    RAISE EXCEPTION 'MIGRATION 565: guard opener present % times, expected 1', open_hits;
  END IF;

  path_hits := (length(tpl) - length(replace(tpl, '{{.existing_component.aspect_paths}}', '')))
               / length('{{.existing_component.aspect_paths}}');
  IF path_hits <> 1 THEN
    RAISE EXCEPTION 'MIGRATION 565: the aspect_paths placeholder is present % times, expected 1', path_hits;
  END IF;

  -- The anchor must SURVIVE: the replace appends after it, never over it.
  IF (length(tpl) - length(replace(tpl, anchor, ''))) / length(anchor) <> 1 THEN
    RAISE EXCEPTION 'MIGRATION 565: the TIER C anchor did not survive the edit';
  END IF;

  -- The guard must close, or every line after it is swallowed by the template.
  IF position('{{end}}' in substring(tpl from position('{{if .existing_component.aspect_paths}}' in tpl))) = 0 THEN
    RAISE EXCEPTION 'MIGRATION 565: the inserted {{if}} block has no {{end}}';
  END IF;
END $$;

COMMIT;
