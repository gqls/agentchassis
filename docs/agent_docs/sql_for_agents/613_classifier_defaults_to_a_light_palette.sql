-- ============================================================================
-- 613 — domain-research-classifier: DEFAULT to a light palette with dark text
--
-- OWNER INSTRUCTION, 2026-08-25: "I would prefer a lighter palette (dark text)
-- as the default on all domains."
--
-- WHAT IS TRUE TODAY. The classifier's prompt gives the palette SHAPE and no
-- guidance either way — every field reads like "Hex for the page background".
-- So light-versus-dark is a free per-model choice, per site, with nothing
-- steering it. Measured 2026-08-25 across current design_intent rows:
--
--     light background   22
--     DARK background     9
--
-- Dark was therefore never a default; it is an unsteered coin-flip that lands
-- dark about a third of the time. This migration makes the owner's preference
-- the stated default and leaves a named escape, rather than banning dark.
--
-- WHY A DEFAULT AND NOT A RULE. A hard ban would be wrong for a vertical whose
-- subject genuinely reads dark, and this estate has sites like that today
-- (dartsonline, gamesdesign). A default with a stated exception puts the burden
-- in the right place: light unless there is a reason, and the reason has to be
-- written down where a reviewer of the SITE can see it.
--
-- SURGICAL, ANCHORED ON A VERBATIM LINE, per the standing rule about not
-- regenerating prompts wholesale. It inserts one instruction immediately before
-- the palette block's opening and touches nothing else. The guard aborts if the
-- anchor is absent or appears more than once — a prompt that has moved must not
-- be patched blind.
--
-- BLAST RADIUS: every FUTURE classification. Existing sites keep their palettes;
-- this changes no live design_intent row. A site already dark stays dark until
-- someone changes it deliberately, which is the conservative direction.
-- ============================================================================

\set ON_ERROR_STOP on

BEGIN;

DO $g$
DECLARE
  anchor text := '"palette": {';
  hits   int;
  step   text := 'classify_and_extract';
  tmpl   text;
BEGIN
  SELECT v->'config'->>'prompt_template' INTO tmpl
  FROM agent_definitions ad, LATERAL jsonb_each(ad.default_config->'workflow'->'steps') AS e(k,v)
  WHERE ad.type='domain-research-classifier' AND k=step
    AND ad.is_active AND COALESCE(ad.is_snapshot,false)=false AND ad.deleted_at IS NULL;

  IF tmpl IS NULL THEN
    RAISE EXCEPTION 'classifier step % not found or has no prompt_template', step;
  END IF;

  hits := (length(tmpl) - length(replace(tmpl, anchor, ''))) / length(anchor);
  IF hits <> 1 THEN
    RAISE EXCEPTION 'anchor %L appears % times, expected exactly 1 - the prompt has moved, refusing to patch blind', anchor, hits;
  END IF;

  IF position('DEFAULT TO A LIGHT SCHEME' in tmpl) > 0 THEN
    RAISE EXCEPTION 'the light-scheme instruction is already present - refusing to add it twice';
  END IF;
END $g$;

-- snapshot before writing (the standing rule for any agent_definitions edit)
CREATE TABLE IF NOT EXISTS bak_classifier_palette_20260825 AS
SELECT * FROM agent_definitions
WHERE type='domain-research-classifier' AND is_active
  AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;

UPDATE agent_definitions ad
SET default_config = jsonb_set(
      ad.default_config,
      '{workflow,steps,classify_and_extract,config,prompt_template}',
      to_jsonb(replace(
        ad.default_config #>> '{workflow,steps,classify_and_extract,config,prompt_template}',
        '"palette": {',
        'DEFAULT TO A LIGHT SCHEME: a light page background with dark body text, unless this specific vertical genuinely reads better dark. If you choose dark, say why in colour_mood - "the subject is X, which is conventionally presented dark" - so the choice is visible to a reviewer rather than silent. Light is the default because it is what most readers expect and what most of this estate already uses; dark is a deliberate exception, not a coin-flip.' || E'\n' ||
        '  "palette": {'
      )),
      false),
    updated_at = now()
WHERE ad.type='domain-research-classifier' AND ad.is_active
  AND COALESCE(ad.is_snapshot,false)=false AND ad.deleted_at IS NULL;

DO $v$
DECLARE tmpl text;
BEGIN
  SELECT default_config #>> '{workflow,steps,classify_and_extract,config,prompt_template}' INTO tmpl
  FROM agent_definitions WHERE type='domain-research-classifier' AND is_active
    AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;
  IF position('DEFAULT TO A LIGHT SCHEME' in tmpl) = 0 THEN
    RAISE EXCEPTION 'post-check FAILED: the instruction is not in the stored prompt';
  END IF;
  IF position('"palette": {' in tmpl) = 0 THEN
    RAISE EXCEPTION 'post-check FAILED: the palette block was damaged by the edit';
  END IF;
  RAISE NOTICE 'post-check OK: instruction present, palette block intact';
END $v$;

COMMIT;
