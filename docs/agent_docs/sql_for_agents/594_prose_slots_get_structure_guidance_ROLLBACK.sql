-- ROLLBACK for 594 — surgical inverse; REFUSES rather than mis-splicing.
--
-- Asymmetric by construction, because the pre-states differ:
--   generic-text-block.content  had NO llm_guidance key  -> the key is REMOVED
--   about-content.content       had NO llm_guidance key  -> the key is REMOVED
--   illustrated-text-block      had specific wording     -> restored verbatim
--   article-body                had proven wording       -> the appended sentence
--                                                          is stripped, the rest kept
-- All four go back to type `text`.

BEGIN;

DO $$
DECLARE r record; t text;
BEGIN
  FOR r IN SELECT * FROM (VALUES
      ('generic-text-block'), ('about-content'), ('illustrated-text-block'), ('article-body')
    ) AS v(fn)
  LOOP
    SELECT input_schema->'fields'->'content'->>'type' INTO t
      FROM content_components WHERE function = r.fn AND is_active;
    IF t IS DISTINCT FROM 'html' THEN
      RAISE EXCEPTION '594 ROLLBACK: %.content is typed % not html — 594 was not applied, or something has changed it since. Refusing.', r.fn, COALESCE(t, '<null>');
    END IF;
  END LOOP;

  IF position($chk$ Never emit <img>, <figure>, <iframe>, form controls$chk$ in
      (SELECT input_schema->'fields'->'content'->>'llm_guidance' FROM content_components WHERE function = 'article-body' AND is_active)) = 0 THEN
    RAISE EXCEPTION '594 ROLLBACK: article-body does not carry the appended sentence verbatim — refusing rather than half-reverting';
  END IF;
END $$;

UPDATE content_components
   SET input_schema = jsonb_set(input_schema, '{fields,content,type}', '"text"'::jsonb)
                        #- '{fields,content,llm_guidance}',
       updated_at = now()
 WHERE function IN ('generic-text-block', 'about-content') AND is_active;

UPDATE content_components
   SET input_schema = jsonb_set(
         jsonb_set(input_schema, '{fields,content,type}', '"text"'::jsonb),
         '{fields,content,llm_guidance}',
         to_jsonb($g$One or more HTML <p> paragraphs of prose for this section's own subject.$g$::text)
       ),
       updated_at = now()
 WHERE function = 'illustrated-text-block' AND is_active;

UPDATE content_components
   SET input_schema = jsonb_set(
         jsonb_set(input_schema, '{fields,content,type}', '"text"'::jsonb),
         '{fields,content,llm_guidance}',
         to_jsonb(replace(
           input_schema->'fields'->'content'->>'llm_guidance',
           $g$ Never emit <img>, <figure>, <iframe>, form controls, inputs, buttons, element ids, class attributes, inline styles or <script>: imagery and visual treatment belong to the component system, not inside this text.$g$,
           $e$$e$))
       ),
       updated_at = now()
 WHERE function = 'article-body' AND is_active;

DO $$
DECLARE r record; t text; g text;
BEGIN
  FOR r IN SELECT * FROM (VALUES
      ('generic-text-block'), ('about-content'), ('illustrated-text-block'), ('article-body')
    ) AS v(fn)
  LOOP
    SELECT input_schema->'fields'->'content'->>'type',
           input_schema->'fields'->'content'->>'llm_guidance'
      INTO t, g FROM content_components WHERE function = r.fn AND is_active;
    IF t <> 'text' THEN
      RAISE EXCEPTION '594 ROLLBACK VERIFY: %.content did not return to text', r.fn;
    END IF;
    IF r.fn IN ('generic-text-block', 'about-content') AND g IS NOT NULL THEN
      RAISE EXCEPTION '594 ROLLBACK VERIFY: %.content still carries llm_guidance — it had none before 594', r.fn;
    END IF;
  END LOOP;

  SELECT input_schema->'fields'->'content'->>'llm_guidance' INTO g
    FROM content_components WHERE function = 'illustrated-text-block' AND is_active;
  IF g <> $g$One or more HTML <p> paragraphs of prose for this section's own subject.$g$ THEN
    RAISE EXCEPTION '594 ROLLBACK VERIFY: illustrated-text-block guidance not restored verbatim';
  END IF;

  SELECT input_schema->'fields'->'content'->>'llm_guidance' INTO g
    FROM content_components WHERE function = 'article-body' AND is_active;
  IF position('Use h2 for main sections' in g) = 0 OR position('<img>' in g) > 0 THEN
    RAISE EXCEPTION '594 ROLLBACK VERIFY: article-body guidance not correctly restored';
  END IF;

  IF (SELECT count(*) FROM content_components cc
        CROSS JOIN LATERAL jsonb_each(cc.input_schema->'fields')
       WHERE cc.function = 'illustrated-text-block' AND cc.is_active) <> 5 THEN
    RAISE EXCEPTION '594 ROLLBACK VERIFY: illustrated-text-block lost sibling fields';
  END IF;

  RAISE NOTICE '594 ROLLBACK OK: four prose slots restored';
END $$;

COMMIT;
