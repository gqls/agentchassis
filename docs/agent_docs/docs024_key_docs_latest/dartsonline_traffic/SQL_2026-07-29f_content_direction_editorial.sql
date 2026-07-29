-- SQL — dartsonline.com content_direction: publication voice + editorial rules
-- (owner decisions D2 + D4, 2026-07-29)
--
-- WHY THIS IS URGENT, and how we found out: the first guide page ever built on this
-- site (barrel-weight, 13:32Z today) came out well-written, on-voice and spec-accurate
-- — and its call-to-action reads "Filter OUR RANGES by weight and tungsten percentage".
-- There are no ranges. The prose was not hallucinated out of nowhere: content_direction
-- instructs the writer to produce shop copy, and the writer obeyed.
--   writing_rules[0]  "…on product listings"
--   writing_rules[4]  "Keep CTAs action-first: 'Add to Bag', 'Pick Your Weight'…"
--   writing_rules[7]  "Price copy should be direct and confident: show savings…"
--   writing_rules[8]  "Brand pages should … not just list SKUs"
--   persuasion_approach.method "Position the store as a trusted guide…"
--   content_depth.thoroughness "Product pages go deep on specs…"
-- So the identity/briefing reset was necessary and NOT sufficient: the voice spec is a
-- third source of the same false premise, and it is the one the writer reads most
-- directly. Fixing only the two obvious aspects would have produced eight more guides
-- all inviting readers to browse a catalogue that does not exist.
--
-- WHAT IS KEPT: the voice itself, which is genuinely good and was never the problem —
-- voice, sentence_style, paragraph_style, heading_style, terminology, example_phrases,
-- things_to_emulate, things_to_avoid are untouched. Second person, contractions, darts
-- vocabulary, spec-first specificity all survive.
-- WHAT IS REPLACED: writing_rules, cta_style, persuasion_approach, content_depth —
-- the four keys that assume a shop.
-- WHAT IS ADDED: `editorial` (D2 — news/analysis rules, which did not exist at all) and
-- `honesty_rails` (D4, mirroring the briefing aspect so the writer sees the constraint
-- in the field it actually reads).
--
-- THE `formatted` FIELD IS LOAD-BEARING AND MUST BE REGENERATED. page-content-writer's
-- prompt reads exactly one field: {{.site_specs.specs.content_direction.formatted}}.
-- `formatted` is normally produced by the `write_site_spec` action
-- (site_spec_actions.go:206-216 -> datahelpers.FormatContentDirection). A hand-written
-- spec that forgets it is INVISIBLE to the writer — the edit would look applied and
-- change nothing. The generator below reproduces FormatContentDirection exactly:
--   string  -> "Humanised key: value"
--   array   -> "Humanised key:\n- item\n- item"
--   object  -> "Humanised key:\n" + same treatment per sub-key, joined by \n
--   blocks joined by "\n\n"; the `formatted` key itself skipped
--   HumaniseKey = underscores to spaces, first character upper-cased (:91-97)
-- Go map iteration order is random, so block ORDER carries no meaning and any order is
-- a faithful reproduction.
--
-- feed-triage needs no change: its prompt iterates every top-level content_direction
-- key ({{range $k,$v := .site_spec.data.content_direction}}), so `editorial` reaches
-- news scoring the moment it exists.

BEGIN;

CREATE TABLE IF NOT EXISTS bak_darts_content_direction_20260729 AS
SELECT * FROM site_specs
WHERE site_id = '5fe8785b-223d-41a3-88ee-c07187622381' AND aspect = 'content_direction';

UPDATE site_specs SET is_current = false, superseded_at = now()
WHERE site_id = '5fe8785b-223d-41a3-88ee-c07187622381'
  AND aspect = 'content_direction' AND is_current = true;

WITH prev AS (
  SELECT data FROM site_specs
  WHERE site_id = '5fe8785b-223d-41a3-88ee-c07187622381'
    AND aspect = 'content_direction' AND is_current = false
  ORDER BY created_at DESC LIMIT 1
),
merged AS (
  SELECT (prev.data - 'formatted') || jsonb_build_object(

    'writing_rules', jsonb_build_array(
      'Always state barrel weight in grams and tungsten percentage when discussing a dart — these are the primary decision variables for any player',
      'Never call a dart ''perfect for beginners'' without explaining why — cite grip style, weight range, or durability rather than vague beginner claims',
      'Use the sport''s own vocabulary: oche, checkout, barrel, shaft, flight, leg, double, bull — do not replace these with generic retail language',
      'Explain what a spec CHANGES about the throw, never just what it is — ''a 25g barrel needs less lift but punishes a loose release'', not ''this barrel weighs 25g''',
      'Keep CTAs action-first and honest about where they lead: ''Read the tungsten guide'', ''Compare flight shapes''. This site sells nothing, so never write ''Add to Bag'', ''Shop the range'', ''Filter our ranges'' or any phrase implying a catalogue',
      'Guide introductions should open with a player scenario or question, not a definition — ''Your flights are doing more work than you think'', not ''Flights are components that...''',
      'Buying guides must acknowledge that equipment preference is personal — recommend, don''t prescribe',
      'Never state a price, discount, stock level or delivery claim. We do not sell, hold stock or ship anything',
      'Write about brands in terms of what makes them distinct in the darts world — players are brand-loyal for real reasons — but never imply we stock, represent or partner with any of them',
      'Avoid copy that reads like it was written for a generic sporting goods store — every line should be recognisable as darts-specific'
    ),

    'cta_style', jsonb_build_object(
      'approach', 'Direct and specific. A CTA points the reader at the next thing worth READING on this site — a guide, a comparison, the news page — because there is nothing to buy here. Use sport-specific language rather than generic retail verbs.',
      'examples', jsonb_build_array(
        'Read the tungsten percentage guide',
        'Compare flight shapes',
        'See how shaft length changes your flight',
        'Catch up on this week''s darts news'
      ),
      'never_use', jsonb_build_array(
        'Add to Bag', 'Shop the range', 'Filter our ranges', 'Browse our catalogue',
        'Buy now', 'Check availability', 'Learn More', 'Proceed'
      )
    ),

    'persuasion_approach', jsonb_build_object(
      'method', 'Help the player make the right choice for their game. Position the SITE as a knowledgeable guide — earn trust through specificity, completeness and admitting trade-offs, never through urgency or sales pressure. We have nothing to sell, which is exactly why the advice can be straight.',
      'evidence', 'Concrete numbers and mechanisms: gram weights, tungsten percentages, barrel profiles, what each one does to flight and grouping. Never testimonials we cannot source, never invented statistics.'
    ),

    'content_depth', jsonb_build_object(
      'thoroughness', 'Guides go deep on the specs that decide a purchase — weight, length, barrel diameter, tungsten percentage, grip style — and always connect the spec to what the player will feel. News items stay short and link out to the original source. Analysis sits between: enough context for a club player, no padding.',
      'assumed_knowledge', 'Mixed audience: casual buyers know little beyond ''steel tip vs soft tip''; club players know barrel weights and flight shapes. Write for both without patronising either.'
    ),

    'editorial', jsonb_build_object(
      'news_scope', 'PDC and major-tour darts: tournament news, results, rankings movements, and equipment or brand releases. UK and world professional darts first; US soft-tip second.',
      'analysis_scope', 'Gear-led analysis for players. Take something that happened — a result, a pro changing setup, a new release — and answer the question a player actually has: what would I change about what I throw, and why? If a piece cannot answer that, it is not for this site.',
      'voice_for_news', 'The same enthusiast-to-enthusiast register as the guides, but reportive. Attribute every fact to its source. Never invent a score, a date, a quote or a ranking. If a detail is not in the source, leave it out.',
      'attribution', 'Aggregated feed items link to the original publication. Analysis pieces name and link the item that prompted them. Never present another outlet''s reporting as our own.',
      'timeliness', 'Say when something happened in plain terms and date it. Do not write ''recently'' or ''just announced'' — a page outlives the phrase.',
      'out_of_scope', jsonb_build_array(
        'Betting odds, tipping, accumulators or any gambling content or promotion (owner decision D1, 2026-07-29)',
        'Player gossip, personal lives, or anything implying insider information',
        'Transfer/contract speculation stated as fact'
      )
    ),

    'honesty_rails', jsonb_build_array(
      'This site holds no stock, sells nothing, ships nothing and has no premises',
      'Never claim a brand relationship, partnership, stockist arrangement or sponsorship',
      'Never state a business address, headquarters or founding history',
      'Never quote a price, discount or delivery time',
      'UK online-only; contact is darts@contactforsales.com / 07934 524 911'
    )
  ) AS data
  FROM prev
),
-- Reproduce datahelpers.FormatContentDirection over the merged document.
blocks AS (
  SELECT
    CASE jsonb_typeof(e.value)
      WHEN 'string' THEN hk.label || ': ' || (e.value #>> '{}')
      WHEN 'array'  THEN hk.label || E':\n' || (
             SELECT string_agg('- ' || a, E'\n') FROM jsonb_array_elements_text(e.value) a)
      WHEN 'object' THEN hk.label || E':\n' || (
             SELECT string_agg(
               CASE jsonb_typeof(o.value)
                 WHEN 'array' THEN (upper(left(replace(o.key,'_',' '),1)) || substr(replace(o.key,'_',' '),2)) || E':\n' || (
                        SELECT string_agg('- ' || a2, E'\n') FROM jsonb_array_elements_text(o.value) a2)
                 ELSE (upper(left(replace(o.key,'_',' '),1)) || substr(replace(o.key,'_',' '),2)) || ': ' || (o.value #>> '{}')
               END, E'\n')
             FROM jsonb_each(e.value) o)
    END AS block
  FROM merged m, jsonb_each(m.data) e
  CROSS JOIN LATERAL (SELECT upper(left(replace(e.key,'_',' '),1)) || substr(replace(e.key,'_',' '),2) AS label) hk
  WHERE e.key <> 'formatted'
    AND jsonb_typeof(e.value) IN ('string','array','object')
    AND e.value <> 'null'::jsonb
    AND NOT (jsonb_typeof(e.value) = 'array'  AND jsonb_array_length(e.value) = 0)
)
INSERT INTO site_specs (site_id, aspect, data, source, source_agent, is_current, created_by, notes)
SELECT '5fe8785b-223d-41a3-88ee-c07187622381', 'content_direction',
       m.data || jsonb_build_object('formatted',
         (SELECT string_agg(block, E'\n\n') FROM blocks WHERE block IS NOT NULL)),
       'authored', 'dartsonline-traffic-workstream', true, 'dartsonline-traffic-workstream',
       'D2 editorial rules + D4 honesty rails. Replaced the four shop-assuming keys '
       || '(writing_rules, cta_style, persuasion_approach, content_depth); kept voice, '
       || 'sentence/paragraph/heading style, terminology, example_phrases, '
       || 'things_to_emulate/avoid untouched. Triggered by the first built guide page '
       || 'writing "Filter our ranges" — traced to writing_rules, not to the writer. '
       || '`formatted` regenerated in SQL to match FormatContentDirection, because '
       || 'page-content-writer reads that field and nothing else. Prior rows in '
       || 'bak_darts_content_direction_20260729.'
FROM merged m;

COMMIT;

-- Verify: formatted exists, is substantial, carries the new rules, and no longer
-- carries the shop instructions. Checked per-phrase, not over the whole blob — the
-- honesty rails legitimately CONTAIN the words they forbid.
SELECT length(data->>'formatted')                              AS formatted_chars,
       (data->>'formatted') LIKE '%Add to Bag%'                AS formatted_still_says_add_to_bag,
       (data->>'formatted') LIKE '%Editorial:%'                AS formatted_has_editorial,
       (data->>'formatted') LIKE '%Honesty rails:%'            AS formatted_has_rails,
       data->'writing_rules' @> '["Always state barrel weight in grams and tungsten percentage when discussing a dart — these are the primary decision variables for any player"]'::jsonb AS rules_replaced,
       data ? 'voice' AND data ? 'terminology' AND data ? 'example_phrases' AS voice_preserved
FROM site_specs
WHERE site_id = '5fe8785b-223d-41a3-88ee-c07187622381'
  AND aspect = 'content_direction' AND is_current = true;
