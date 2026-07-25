-- p4_04_guide_list_cta_overridable.sql — stop guide-list overriding its own authored CTA.
--
-- THE DEFECT, observed live on TWO sites (not inferred).
-- p4_02 set the guides-hub CTA to the paid tool:
--     cta_url = /report.html,  cta_button_label = 'Get a verified idea report'
-- The rendered page shows something else entirely:
--     <a class="guide-list-cta-btn" href="/guides/index.html">Browse all guides</a>
-- i.e. a button on the guides hub whose destination is **the guides hub** — a dead control —
-- sitting directly beneath authored copy that promises a £29 report. Same on gamesdesign.co.uk
-- (curl-verified: identical href/label on its own /guides/index.html) and, by inspection,
-- relojistas.com /guias/index.html. Three of the four live instances self-link.
--
-- WHY. `guide-list_pre_037`'s schema declares:
--     cta_url          source query.section_index_for:guide   → resolves to the guides hub
--     cta_button_label source static, fallback 'Browse all guides'
--     eyebrow_label    source static, fallback 'Guides'
-- and resolved_data is merged **LAST** at render time ("resolved data wins on conflicts, by
-- design"), so both kinds of value beat anything authored in content_data. For the two `static`
-- fields this is unconditional: `plan_sections_action.go:1556-1562` writes the fallback into
-- resolvedData whenever one exists and returns — it never looks at content_data. The field's own
-- `llm_guidance` says *"Override if the site tone prefers a different phrasing"*. **You cannot.**
-- That guidance is not merely unhelpful, it is false, which is how the authored value was set here
-- in good faith and silently discarded.
--
-- This is the bugs_open/023 family (a derived CTA that ignores the label/intent beside it), and a
-- SECOND self-link path: 023's own `chooseCTATargets` already drops self-links, but
-- `queryresolve.section_index_for` is a different derivation with no such guard, used *on* the very
-- hub it resolves to. 023 is OWNED by the cta_link_integrity workstream (scripts/who-owns.py), so
-- this is contributed INTO that bug file, not forked into a competing fix.
--
-- THE FIX, and why it is a strict no-op for the other three instances.
-- Stop the schema overriding authored intent on the three fields: drop the query source from
-- `cta_url`, drop the fallbacks from `cta_button_label` and `eyebrow_label`. `items` KEEPS its
-- query source — that list must stay derived, and is the whole point of p4_02.
--
-- Verified before writing: all FOUR live instances already carry these three keys in content_data
-- with exactly the values the resolver produces —
--     gamesdesign /index.html         cta_url=/guides/index.html  label='Browse all guides'  eyebrow='Guides'
--     gamesdesign /guides/index.html  cta_url=/guides/index.html  label='Browse all guides'  eyebrow='Guides'
--     relojistas  /guias/index.html   cta_url=/guias/index.html   label='Browse all guides'  eyebrow='Guides'
--     idea.uk     /guides/index.html  (reset below to the intended paid-tool target)
-- so after this change their stored values render instead of the derived ones, byte-identical.
-- Whether the other two sites *should* keep a self-linking CTA is their workstream's call; this
-- change does not alter what they render today.
--
-- KNOWN CONSEQUENCE, stated rather than discovered later: a NEW guide-list instance created with
-- empty content_data will no longer be auto-filled with a hub URL and a generic label — the
-- template gates on `{{if .cta_url}}`, so it renders the section without a CTA button instead of
-- with a self-link. That is the bugs_open/023 principle applied (do not emit a control with no
-- authored destination), and it is why `items` is deliberately left derived.

\set ON_ERROR_STOP on

BEGIN;

-- Snapshot the schema we are editing.
DROP TABLE IF EXISTS bak_guidelist_schema_20260725;
CREATE TABLE bak_guidelist_schema_20260725 AS
SELECT id, name, input_schema, now() AS snapshotted_at
FROM content_components WHERE id = '9d5e461a-8981-4ecc-b236-05895edfc15d';

-- Guard: refuse unless every existing instance already stores all three keys, or this stops
-- being a no-op and silently drops a button somewhere else in the fleet.
DO $guard$
DECLARE n int;
BEGIN
  SELECT count(*) INTO n
  FROM page_components pc
  WHERE pc.component_id = '9d5e461a-8981-4ecc-b236-05895edfc15d'
    AND NOT (pc.content_data ? 'cta_url'
         AND pc.content_data ? 'cta_button_label'
         AND pc.content_data ? 'eyebrow_label');
  IF n > 0 THEN
    RAISE EXCEPTION 'ABORT: % guide-list instance(s) do not carry all three keys in content_data — this edit would NOT be a no-op for them.', n;
  END IF;
END
$guard$;

-- 1. cta_url: derived → authored.
UPDATE content_components
SET input_schema = jsonb_set(input_schema, '{fields,cta_url}',
      (input_schema->'fields'->'cta_url') - 'source'),
    updated_at = now()
WHERE id = '9d5e461a-8981-4ecc-b236-05895edfc15d';

-- 2 & 3. static fallbacks that beat content_data unconditionally → authored.
UPDATE content_components
SET input_schema = jsonb_set(input_schema, '{fields,cta_button_label}',
      (input_schema->'fields'->'cta_button_label') - 'fallback'),
    updated_at = now()
WHERE id = '9d5e461a-8981-4ecc-b236-05895edfc15d';

UPDATE content_components
SET input_schema = jsonb_set(input_schema, '{fields,eyebrow_label}',
      (input_schema->'fields'->'eyebrow_label') - 'fallback'),
    updated_at = now()
WHERE id = '9d5e461a-8981-4ecc-b236-05895edfc15d';

-- 4. Restore idea.uk's intended CTA (p4_02's values, which the resolver overwrote in content_data
--    on the last render — the override is destructive, not just presentational).
UPDATE page_components pc
SET content_data = pc.content_data || jsonb_build_object(
      'cta_url',          '/report.html',
      'cta_button_label', 'Get a verified idea report'
    ),
    updated_at = now()
FROM pages p
WHERE p.id = pc.page_id
  AND p.site_id = '1244516d-014d-421c-88c6-090bb1e9552a'
  AND p.url = '/guides/index.html'
  AND pc.slot_name = 'guide-list';

DO $guard2$
DECLARE s jsonb;
BEGIN
  SELECT input_schema->'fields' INTO s FROM content_components
   WHERE id = '9d5e461a-8981-4ecc-b236-05895edfc15d';
  IF (s->'cta_url' ? 'source') OR (s->'cta_button_label' ? 'fallback') OR (s->'eyebrow_label' ? 'fallback') THEN
    RAISE EXCEPTION 'ABORT: schema edit did not take: %', s;
  END IF;
  IF NOT (s->'items'->>'source' = 'query.pages_where_type:guide') THEN
    RAISE EXCEPTION 'ABORT: items lost its query source — the listing would stop deriving: %', s->'items';
  END IF;
END
$guard2$;

COMMIT;

SELECT k AS field, v->>'source' AS source, v->>'fallback' AS fallback
FROM content_components c, jsonb_each(c.input_schema->'fields') AS e(k,v)
WHERE c.id = '9d5e461a-8981-4ecc-b236-05895edfc15d'
  AND k IN ('items','cta_url','cta_button_label','eyebrow_label');

SELECT s.domain, p.url, pc.content_data->>'cta_url' AS cta_url, pc.content_data->>'cta_button_label' AS label
FROM page_components pc JOIN pages p ON p.id=pc.page_id JOIN sites s ON s.id=p.site_id
WHERE pc.component_id = '9d5e461a-8981-4ecc-b236-05895edfc15d' ORDER BY s.domain, p.url;
