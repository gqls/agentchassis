-- Migration: domain-research-classifier — add the universal "build standard" (best-in-class quality bar)
--
-- WHAT
--   Inserts a "Build standard" framing block at the top of the classify_and_extract
--   prompt, applied to every site regardless of inputs. It raises the QUALITY/FIT
--   bar (design, writing, genuine usefulness, fit-to-vertical) and is deliberately
--   scoped NOT to expand scope — it forbids inventing services/pages/features/facts
--   beyond the evidence and defers aspiration to the fidelity dial. This keeps it
--   compatible with doc 028's current state (no per-item status yet -> the build
--   pipeline builds whatever the spec says; adopted sites must stay faithful).
--
-- WHY
--   Encodes the platform's standing ambition ("best-in-class results") in the agent
--   that reasons out every site's direction. Its outputs are written to the spec, so
--   the bar propagates to strategist / planner / design downstream without
--   duplicating the text per site.
--
-- SCOPE / SAFETY
--   * Only static text inserted at the top of the prompt; no step names, output
--     fields, config keys, or {{ }} variables changed.
--   * snapshot_agent() backup inside the transaction; UPDATE guarded to the live,
--     non-snapshot row; self-check RAISEs (rolls back) if the block didn't land.
--   * Re-runnable: NOT LIKE guard prevents a second application.
--
-- ROLLBACK: footer comment (restore default_config from the latest snapshot).

\set ON_ERROR_STOP on

BEGIN;

SELECT snapshot_agent(
  'domain-research-classifier',
  'pre-change backup: add universal best-in-class build standard to classify_and_extract'
) AS snapshot_id;

UPDATE agent_definitions
SET
  default_config = jsonb_set(
    default_config,
    '{workflow,steps,classify_and_extract,config,prompt_template}',
    to_jsonb(
      replace(
        default_config #>> '{workflow,steps,classify_and_extract,config,prompt_template}',
        $old$You are analyzing a domain for website creation.$old$,
        $new$You are analyzing a domain for website creation.

## Build standard (applies to every site, regardless of inputs)
Aim for best-in-class quality in this site's field. The bar is not "competent template" but "stands comparison with the strongest sites in this vertical" — in the quality of the design, the clarity and usefulness of the writing, and the genuine value of any tools or content to the people who will actually use the site. When forming direction, consider what the best sites in this space do — how they position, what their design signals, how their copy reads, what earns return visits — and build on the reasoning behind those choices rather than copying them. Choose design and content that fit this specific industry and these objectives, not a generic house style. Favour fewer things done genuinely well over filler, and prefer interactive or visual elements where they aid understanding. Do what is most useful and interesting for the site's visitors.

This standard governs QUALITY and FIT, not scope. Do not invent services, pages, features, or facts beyond what the evidence supports; where research is thin, say so honestly in the confidence fields rather than fabricating detail. Treat aspirational ideas as direction to be realised at the pace the site's fidelity allows — adopted sites stay faithful to their source at first — not as things to force into the first build.$new$
      )
    ),
    false
  ),
  updated_at = now()
WHERE type = 'domain-research-classifier'
  AND is_active = true
  AND (is_snapshot IS NULL OR is_snapshot = false)
  AND default_config #>> '{workflow,steps,classify_and_extract,config,prompt_template}'
        LIKE '%You are analyzing a domain for website creation.%'
  AND default_config #>> '{workflow,steps,classify_and_extract,config,prompt_template}'
        NOT LIKE '%Build standard (applies to every site%';

DO $check$
DECLARE has_block boolean;
BEGIN
  SELECT (pt LIKE '%Build standard (applies to every site%')
  INTO has_block
  FROM (
    SELECT default_config #>> '{workflow,steps,classify_and_extract,config,prompt_template}' AS pt
    FROM agent_definitions
    WHERE type = 'domain-research-classifier'
      AND is_active = true AND (is_snapshot IS NULL OR is_snapshot = false)
    LIMIT 1
  ) s;
  IF NOT COALESCE(has_block, false) THEN
    RAISE EXCEPTION 'build standard block not applied; anchor did not match — rolling back';
  END IF;
END
$check$;

COMMIT;

-- ---------------------------------------------------------------------------
-- VERIFY:
--   SELECT default_config #>> '{workflow,steps,classify_and_extract,config,prompt_template}'
--            LIKE '%Build standard%' AS has_block
--   FROM agent_definitions WHERE type='domain-research-classifier'
--     AND is_active = true AND (is_snapshot IS NULL OR is_snapshot = false);
--
-- ROLLBACK (restore default_config from the latest snapshot):
--   UPDATE agent_definitions live SET default_config = snap.default_config, updated_at = now()
--   FROM (SELECT default_config FROM agent_definitions
--         WHERE type='domain-research-classifier' AND is_snapshot = true
--         ORDER BY created_at DESC LIMIT 1) snap
--   WHERE live.type='domain-research-classifier' AND live.is_active = true
--     AND (live.is_snapshot IS NULL OR live.is_snapshot = false);
-- ---------------------------------------------------------------------------
