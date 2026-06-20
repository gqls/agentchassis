-- Migration: domain-research-classifier — emit structured design_intent.palette + typography
--
-- WHAT
--   Adds palette.reference_values (8 core slots) and typography.reference_values
--   (font_family + heading_font) to the design_intent output of the classifier's
--   `classify_and_extract` step, plus one MANDATORY-fields instruction bullet.
--
-- WHY
--   The composition pipeline (site-design-planner -> resolve_composition_palette /
--   _typography) and the design renderer read design_intent.palette.reference_values
--   and design_intent.typography.reference_values. They do NOT read the prose
--   colour_mood / typography_mood. Fresh builds (e.g. idea.uk) put their colours in
--   the prose fields, so the cascade found nothing structured and fell through to a
--   default. This makes the classifier emit the structured blocks every consumer reads.
--   `write_design_intent_spec` already persists analysis.result.design_intent wholesale,
--   so the new keys flow through with no write-path change.
--
-- SCOPE / SAFETY
--   * Only static schema + instruction text inside the prompt string is changed.
--     No step names, output fields, config keys, or {{ }} variables are touched.
--   * Backs up the live row first via snapshot_agent() (stores is_snapshot=true at
--     version+1000), inside the same transaction as the change.
--   * The UPDATE is guarded to the LIVE row only (is_active = true and not a snapshot),
--     so the freshly-created backup is not also edited.
--   * Two text anchors are replaced exactly; if either anchor is absent the UPDATE
--     no-ops and the self-check RAISEs, rolling the whole transaction back. Re-runnable:
--     the NOT LIKE guard prevents a second application.
--
-- ROLLBACK
--   See the commented block at the foot of this file (restores default_config from the
--   most recent snapshot for this type).

\set ON_ERROR_STOP on

BEGIN;

-- 1. Back up the current live definition (returns the snapshot row's uuid).
SELECT snapshot_agent(
  'domain-research-classifier',
  'pre-change backup: add structured design_intent.palette + typography reference_values to classify_and_extract'
) AS snapshot_id;

-- 2. Apply the change to the LIVE row only (excludes the snapshot just created).
UPDATE agent_definitions
SET
  default_config = jsonb_set(
    default_config,
    '{workflow,steps,classify_and_extract,config,prompt_template}',
    to_jsonb(
      replace(
        replace(
          default_config #>> '{workflow,steps,classify_and_extract,config,prompt_template}',
          $old1$  "avoid": ["Design elements to avoid"]
}$old1$,
          $new1$  "avoid": ["Design elements to avoid"],
  "palette": {
    "reference_values": {
      "primary": "Hex for the dominant brand colour (not the accent)",
      "secondary": "Hex for a supporting colour",
      "accent": "Hex for the single emphasis colour (CTAs, highlights)",
      "background": "Hex for the page background",
      "surface": "Hex for card/panel backgrounds",
      "text": "Hex for body text",
      "text_muted": "Hex for secondary/muted text",
      "border": "Hex for hairlines/dividers"
    }
  },
  "typography": {
    "reference_values": {
      "font_family": "Body font stack, e.g. 'IBM Plex Sans', system-ui, sans-serif",
      "heading_font": "Heading font stack, e.g. 'Fraunces', Georgia, serif"
    }
  }
}$new1$
        ),
        $old2$- recommended_builder should always be "pageflow-builder" for now.$old2$,
        $new2$- recommended_builder should always be "pageflow-builder" for now.
- design_intent.palette.reference_values (all eight slots, as hex) and design_intent.typography.reference_values (font_family and heading_font, as CSS font stacks) are MANDATORY: the composition pipeline (site-design-planner) and the design renderer read these structured fields, not the prose colour_mood/typography_mood. Populate every slot. Map the dominant brand colour to primary and the single emphasis colour to accent (they are usually different). Never default to generic blue-and-grey; derive the palette from the mission/adoption/research and the industry. style_direction must agree with the palette you emit (a light palette is not "professional-dark").$new2$
      )
    ),
    false
  ),
  updated_at = now()
WHERE type = 'domain-research-classifier'
  AND is_active = true
  AND (is_snapshot IS NULL OR is_snapshot = false)
  AND default_config #>> '{workflow,steps,classify_and_extract,config,prompt_template}'
        LIKE '%"avoid": ["Design elements to avoid"]%'
  AND default_config #>> '{workflow,steps,classify_and_extract,config,prompt_template}'
        NOT LIKE '%design_intent.palette.reference_values%';

-- 3. Self-check: both additions must be present on the live row, else roll back.
DO $check$
DECLARE
  has_schema boolean;
  has_rule   boolean;
BEGIN
  SELECT
    (pt LIKE '%"reference_values"%'),
    (pt LIKE '%design_intent.palette.reference_values%')
  INTO has_schema, has_rule
  FROM (
    SELECT default_config #>> '{workflow,steps,classify_and_extract,config,prompt_template}' AS pt
    FROM agent_definitions
    WHERE type = 'domain-research-classifier'
      AND is_active = true
      AND (is_snapshot IS NULL OR is_snapshot = false)
    LIMIT 1
  ) s;

  IF NOT (COALESCE(has_schema, false) AND COALESCE(has_rule, false)) THEN
    RAISE EXCEPTION
      'domain-research-classifier design_intent blocks not applied (schema=%, rule=%); anchors did not match — rolling back',
      has_schema, has_rule;
  END IF;
END
$check$;

COMMIT;

-- ---------------------------------------------------------------------------
-- VERIFY (run after commit):
--   SELECT default_config #>> '{workflow,steps,classify_and_extract,config,prompt_template}'
--            LIKE '%"reference_values"%' AS has_blocks
--   FROM agent_definitions
--   WHERE type = 'domain-research-classifier'
--     AND is_active = true AND (is_snapshot IS NULL OR is_snapshot = false);
--
-- ROLLBACK (manual — restores the live row's default_config from the latest snapshot):
--   UPDATE agent_definitions live
--   SET default_config = snap.default_config, updated_at = now()
--   FROM (
--     SELECT default_config
--     FROM agent_definitions
--     WHERE type = 'domain-research-classifier' AND is_snapshot = true
--     ORDER BY created_at DESC
--     LIMIT 1
--   ) snap
--   WHERE live.type = 'domain-research-classifier'
--     AND live.is_active = true
--     AND (live.is_snapshot IS NULL OR live.is_snapshot = false);
-- ---------------------------------------------------------------------------
