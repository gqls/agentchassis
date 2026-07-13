-- 089_archetype_page_copy.sql — replace generic boilerplate with canon archetype copy
-- Created 2026-07-12. The content-writer filled all 8 entity pages with generic
-- site copy (zero mentions of the archetype; 'Clients Served'-style stat labels;
-- CTAs pointing at /contact.html and /about). content_data is the source of
-- truth, so this authors the copy directly from the spec's content_context
-- archetypes canon, then the pages re-render via the light no-LLM path
-- (page_rerender, reason section_data_resolved).
-- Reversal: _vonc_archetype_backup_20260712_pc_content.

BEGIN;

CREATE TABLE _vonc_archetype_backup_20260712_pc_content AS
  SELECT pc.id, pc.page_id, pc.content_data
  FROM page_components pc JOIN pages p ON p.id=pc.page_id
  WHERE p.site_id='9ec3b9ee-5b08-461b-b4f8-9e1e03579c74' AND p.page_type='entity-page';

UPDATE page_components pc
SET content_data = pc.content_data || '{"eyebrow_text": "ARCHETYPE PROFILE", "heading": "Cuts to the Core. Every Time.", "body_text": "The Surgeon strips every argument to its essential structure. While the room reacts, you''re already three moves in \u2014 mapping the logic, isolating the load-bearing claim, finding the flaw everyone else walked past. In the Arena your positions read like incisions: short, precise, impossible to unsee. You don''t win by volume. You win because after your take lands, the argument is simply different.", "image_alt": "Stylised icon of The Surgeon archetype \u2014 precision instrument over a dissected argument", "stat_1_value": "Sharpest", "stat_1_label": "Structural clarity in the room", "stat_2_value": "Fastest", "stat_2_label": "On logical frameworks", "stat_3_value": "First", "stat_3_label": "To find the hidden flaw", "cta_label": "Find Your Archetype", "cta_url": "/tools/archetype-taster-quiz/index.html"}'::jsonb, updated_at = NOW()
FROM content_components cc, pages p
WHERE cc.id = pc.component_id AND p.id = pc.page_id
  AND p.site_id = '9ec3b9ee-5b08-461b-b4f8-9e1e03579c74' AND p.name = 'surgeon'
  AND cc.function = 'content-block-about';

UPDATE page_components pc
SET content_data = pc.content_data || '{"headline": "The Surgeon. Precise. Analytical. Cuts to the Core.", "subheadline": "Earned, not chosen. The Archetype that strips arguments to structure and finds what everyone else missed.", "cta_text": "Run the Gauntlet", "cta_url": "/tools/gauntlet/index.html", "secondary_cta": "Take the Archetype Quiz", "secondary_cta_url": "/tools/archetype-taster-quiz/index.html"}'::jsonb, updated_at = NOW()
FROM content_components cc, pages p
WHERE cc.id = pc.component_id AND p.id = pc.page_id
  AND p.site_id = '9ec3b9ee-5b08-461b-b4f8-9e1e03579c74' AND p.name = 'surgeon'
  AND cc.function = 'hero';

UPDATE page_components pc
SET content_data = pc.content_data || '{"primary_cta_url": "/tools/gauntlet/index.html", "secondary_cta_url": "/tools/archetype-taster-quiz/index.html"}'::jsonb, updated_at = NOW()
FROM content_components cc, pages p
WHERE cc.id = pc.component_id AND p.id = pc.page_id
  AND p.site_id = '9ec3b9ee-5b08-461b-b4f8-9e1e03579c74' AND p.name = 'surgeon'
  AND cc.function = 'call-to-action';

UPDATE page_components pc
SET content_data = pc.content_data || '{"eyebrow_text": "ARCHETYPE PROFILE", "heading": "Nobody Sees Your Take Coming.", "body_text": "The Wildcard takes ideas where they weren''t supposed to go. You cross domains without asking permission \u2014 a take on economics lands via game design, a position on privacy arrives through cooking. The Arena tracks remix rate, and yours is the highest: your positions get picked up, bent, and rebuilt more than anyone''s. Nobody can predict where you''ll land. That''s the point.", "image_alt": "Stylised icon of The Wildcard archetype \u2014 trajectories crossing domains in unexpected directions", "stat_1_value": "Highest", "stat_1_label": "Remix rate in the Arena", "stat_2_value": "Freest", "stat_2_label": "Movement across domains", "stat_3_value": "Least", "stat_3_label": "Predictable position filed", "cta_label": "Find Your Archetype", "cta_url": "/tools/archetype-taster-quiz/index.html"}'::jsonb, updated_at = NOW()
FROM content_components cc, pages p
WHERE cc.id = pc.component_id AND p.id = pc.page_id
  AND p.site_id = '9ec3b9ee-5b08-461b-b4f8-9e1e03579c74' AND p.name = 'wildcard'
  AND cc.function = 'content-block-about';

UPDATE page_components pc
SET content_data = pc.content_data || '{"headline": "The Wildcard. Unpredictable by Design.", "subheadline": "Earned, not chosen. The Archetype with the highest remix rate in the Arena \u2014 ideas bent across domains until they''re new.", "cta_text": "Run the Gauntlet", "cta_url": "/tools/gauntlet/index.html", "secondary_cta": "Take the Archetype Quiz", "secondary_cta_url": "/tools/archetype-taster-quiz/index.html"}'::jsonb, updated_at = NOW()
FROM content_components cc, pages p
WHERE cc.id = pc.component_id AND p.id = pc.page_id
  AND p.site_id = '9ec3b9ee-5b08-461b-b4f8-9e1e03579c74' AND p.name = 'wildcard'
  AND cc.function = 'hero';

UPDATE page_components pc
SET content_data = pc.content_data || '{"primary_cta_url": "/tools/gauntlet/index.html", "secondary_cta_url": "/tools/archetype-taster-quiz/index.html"}'::jsonb, updated_at = NOW()
FROM content_components cc, pages p
WHERE cc.id = pc.component_id AND p.id = pc.page_id
  AND p.site_id = '9ec3b9ee-5b08-461b-b4f8-9e1e03579c74' AND p.name = 'wildcard'
  AND cc.function = 'call-to-action';

UPDATE page_components pc
SET content_data = pc.content_data || '{"eyebrow_text": "ARCHETYPE PROFILE", "heading": "Speaks Rarely. Lands Precisely.", "body_text": "The Oracle doesn''t file often. But when you do, the room goes quiet. Your record speaks: the best prediction accuracy on the platform, positions timed to the moment they matter most, engagement so selective it reads as verdict. Others take sides every day. You wait \u2014 and when you finally move, the Arena tends to discover you were right.", "image_alt": "Stylised icon of The Oracle archetype \u2014 a single point of light over a field of predictions", "stat_1_value": "Best", "stat_1_label": "Prediction accuracy on record", "stat_2_value": "Rarest", "stat_2_label": "Positions filed", "stat_3_value": "Hardest", "stat_3_label": "Landings to ignore", "cta_label": "Find Your Archetype", "cta_url": "/tools/archetype-taster-quiz/index.html"}'::jsonb, updated_at = NOW()
FROM content_components cc, pages p
WHERE cc.id = pc.component_id AND p.id = pc.page_id
  AND p.site_id = '9ec3b9ee-5b08-461b-b4f8-9e1e03579c74' AND p.name = 'oracle'
  AND cc.function = 'content-block-about';

UPDATE page_components pc
SET content_data = pc.content_data || '{"headline": "The Oracle. Rare. Precise. Right.", "subheadline": "Earned, not chosen. The Archetype that speaks least and lands hardest, with the best prediction accuracy on the platform.", "cta_text": "Run the Gauntlet", "cta_url": "/tools/gauntlet/index.html", "secondary_cta": "Take the Archetype Quiz", "secondary_cta_url": "/tools/archetype-taster-quiz/index.html"}'::jsonb, updated_at = NOW()
FROM content_components cc, pages p
WHERE cc.id = pc.component_id AND p.id = pc.page_id
  AND p.site_id = '9ec3b9ee-5b08-461b-b4f8-9e1e03579c74' AND p.name = 'oracle'
  AND cc.function = 'hero';

UPDATE page_components pc
SET content_data = pc.content_data || '{"primary_cta_url": "/tools/gauntlet/index.html", "secondary_cta_url": "/tools/archetype-taster-quiz/index.html"}'::jsonb, updated_at = NOW()
FROM content_components cc, pages p
WHERE cc.id = pc.component_id AND p.id = pc.page_id
  AND p.site_id = '9ec3b9ee-5b08-461b-b4f8-9e1e03579c74' AND p.name = 'oracle'
  AND cc.function = 'call-to-action';

UPDATE page_components pc
SET content_data = pc.content_data || '{"eyebrow_text": "ARCHETYPE PROFILE", "heading": "Your Response Starts the Chain.", "body_text": "The Catalyst doesn''t need the best take in the thread \u2014 yours is the one the thread grows from. You open angles other people didn''t know were there, and the longest response chains in the Arena trace back to something you filed. Momentum is your instrument: one position from you and a dead Provocation catches fire.", "image_alt": "Stylised icon of The Catalyst archetype \u2014 a single spark igniting a chain of responses", "stat_1_value": "Longest", "stat_1_label": "Response chains started", "stat_2_value": "Widest", "stat_2_label": "Angles opened per Provocation", "stat_3_value": "Most", "stat_3_label": "Momentum generated in the Arena", "cta_label": "Find Your Archetype", "cta_url": "/tools/archetype-taster-quiz/index.html"}'::jsonb, updated_at = NOW()
FROM content_components cc, pages p
WHERE cc.id = pc.component_id AND p.id = pc.page_id
  AND p.site_id = '9ec3b9ee-5b08-461b-b4f8-9e1e03579c74' AND p.name = 'catalyst'
  AND cc.function = 'content-block-about';

UPDATE page_components pc
SET content_data = pc.content_data || '{"headline": "The Catalyst. Where Chains Begin.", "subheadline": "Earned, not chosen. The Archetype whose responses generate the longest chains in the Arena.", "cta_text": "Run the Gauntlet", "cta_url": "/tools/gauntlet/index.html", "secondary_cta": "Take the Archetype Quiz", "secondary_cta_url": "/tools/archetype-taster-quiz/index.html"}'::jsonb, updated_at = NOW()
FROM content_components cc, pages p
WHERE cc.id = pc.component_id AND p.id = pc.page_id
  AND p.site_id = '9ec3b9ee-5b08-461b-b4f8-9e1e03579c74' AND p.name = 'catalyst'
  AND cc.function = 'hero';

UPDATE page_components pc
SET content_data = pc.content_data || '{"primary_cta_url": "/tools/gauntlet/index.html", "secondary_cta_url": "/tools/archetype-taster-quiz/index.html"}'::jsonb, updated_at = NOW()
FROM content_components cc, pages p
WHERE cc.id = pc.component_id AND p.id = pc.page_id
  AND p.site_id = '9ec3b9ee-5b08-461b-b4f8-9e1e03579c74' AND p.name = 'catalyst'
  AND cc.function = 'call-to-action';

UPDATE page_components pc
SET content_data = pc.content_data || '{"eyebrow_text": "ARCHETYPE PROFILE", "heading": "Watches Everything. Misses Nothing.", "body_text": "The Judge lives in the Gallery. You don''t file first and you don''t file often \u2014 you watch, and your reactions are the most predictive signal on the platform. Where your taste settles, the final outcome tends to follow. Pattern recognition this sharp isn''t spectating. It''s the quiet authority everyone else checks against.", "image_alt": "Stylised icon of The Judge archetype \u2014 an observing eye over settling patterns", "stat_1_value": "Most", "stat_1_label": "Predictive reactions in the Gallery", "stat_2_value": "Sharpest", "stat_2_label": "Pattern recognition", "stat_3_value": "Surest", "stat_3_label": "Taste under pressure", "cta_label": "Find Your Archetype", "cta_url": "/tools/archetype-taster-quiz/index.html"}'::jsonb, updated_at = NOW()
FROM content_components cc, pages p
WHERE cc.id = pc.component_id AND p.id = pc.page_id
  AND p.site_id = '9ec3b9ee-5b08-461b-b4f8-9e1e03579c74' AND p.name = 'judge'
  AND cc.function = 'content-block-about';

UPDATE page_components pc
SET content_data = pc.content_data || '{"headline": "The Judge. The Gallery''s Quiet Authority.", "subheadline": "Earned, not chosen. The Archetype whose reactions predict final outcomes better than anyone''s positions.", "cta_text": "Run the Gauntlet", "cta_url": "/tools/gauntlet/index.html", "secondary_cta": "Take the Archetype Quiz", "secondary_cta_url": "/tools/archetype-taster-quiz/index.html"}'::jsonb, updated_at = NOW()
FROM content_components cc, pages p
WHERE cc.id = pc.component_id AND p.id = pc.page_id
  AND p.site_id = '9ec3b9ee-5b08-461b-b4f8-9e1e03579c74' AND p.name = 'judge'
  AND cc.function = 'hero';

UPDATE page_components pc
SET content_data = pc.content_data || '{"primary_cta_url": "/tools/gauntlet/index.html", "secondary_cta_url": "/tools/archetype-taster-quiz/index.html"}'::jsonb, updated_at = NOW()
FROM content_components cc, pages p
WHERE cc.id = pc.component_id AND p.id = pc.page_id
  AND p.site_id = '9ec3b9ee-5b08-461b-b4f8-9e1e03579c74' AND p.name = 'judge'
  AND cc.function = 'call-to-action';

UPDATE page_components pc
SET content_data = pc.content_data || '{"eyebrow_text": "ARCHETYPE PROFILE", "heading": "Shows the Work. Skips the Opinion.", "body_text": "The Maker shows up with creations, not opinions. While the Arena argues, you build the thing that settles it \u2014 and the Stage is where you dominate, filing work with a consistency nobody matches. Skill demonstrated beats skill claimed. Your output is your position, and it''s on the record every day.", "image_alt": "Stylised icon of The Maker archetype \u2014 hands mid-build over a workbench of finished pieces", "stat_1_value": "Most", "stat_1_label": "Consistent output on the Stage", "stat_2_value": "Strongest", "stat_2_label": "Skill shown, not claimed", "stat_3_value": "Steadiest", "stat_3_label": "Creative cadence", "cta_label": "Find Your Archetype", "cta_url": "/tools/archetype-taster-quiz/index.html"}'::jsonb, updated_at = NOW()
FROM content_components cc, pages p
WHERE cc.id = pc.component_id AND p.id = pc.page_id
  AND p.site_id = '9ec3b9ee-5b08-461b-b4f8-9e1e03579c74' AND p.name = 'maker'
  AND cc.function = 'content-block-about';

UPDATE page_components pc
SET content_data = pc.content_data || '{"headline": "The Maker. Creations, Not Opinions.", "subheadline": "Earned, not chosen. The Archetype that dominates the Stage with consistent, demonstrated skill.", "cta_text": "Run the Gauntlet", "cta_url": "/tools/gauntlet/index.html", "secondary_cta": "Take the Archetype Quiz", "secondary_cta_url": "/tools/archetype-taster-quiz/index.html"}'::jsonb, updated_at = NOW()
FROM content_components cc, pages p
WHERE cc.id = pc.component_id AND p.id = pc.page_id
  AND p.site_id = '9ec3b9ee-5b08-461b-b4f8-9e1e03579c74' AND p.name = 'maker'
  AND cc.function = 'hero';

UPDATE page_components pc
SET content_data = pc.content_data || '{"primary_cta_url": "/tools/gauntlet/index.html", "secondary_cta_url": "/tools/archetype-taster-quiz/index.html"}'::jsonb, updated_at = NOW()
FROM content_components cc, pages p
WHERE cc.id = pc.component_id AND p.id = pc.page_id
  AND p.site_id = '9ec3b9ee-5b08-461b-b4f8-9e1e03579c74' AND p.name = 'maker'
  AND cc.function = 'call-to-action';

UPDATE page_components pc
SET content_data = pc.content_data || '{"eyebrow_text": "ARCHETYPE PROFILE", "heading": "Finds It First.", "body_text": "The Scout finds it first. The creator everyone cites next quarter \u2014 you flagged them when the room was empty. Discovery is your discipline: taste sharp enough to trust, applied early enough to matter. The platform''s breakout stories tend to have the same first chapter: you, noticing.", "image_alt": "Stylised icon of The Scout archetype \u2014 a lens catching a distant signal before anyone else", "stat_1_value": "Earliest", "stat_1_label": "To breakout creators", "stat_2_value": "Keenest", "stat_2_label": "Eye for what others miss", "stat_3_value": "First", "stat_3_label": "To surface the next thing", "cta_label": "Find Your Archetype", "cta_url": "/tools/archetype-taster-quiz/index.html"}'::jsonb, updated_at = NOW()
FROM content_components cc, pages p
WHERE cc.id = pc.component_id AND p.id = pc.page_id
  AND p.site_id = '9ec3b9ee-5b08-461b-b4f8-9e1e03579c74' AND p.name = 'scout'
  AND cc.function = 'content-block-about';

UPDATE page_components pc
SET content_data = pc.content_data || '{"headline": "The Scout. First to What''s Next.", "subheadline": "Earned, not chosen. The Archetype that identifies breakout creators before the room notices.", "cta_text": "Run the Gauntlet", "cta_url": "/tools/gauntlet/index.html", "secondary_cta": "Take the Archetype Quiz", "secondary_cta_url": "/tools/archetype-taster-quiz/index.html"}'::jsonb, updated_at = NOW()
FROM content_components cc, pages p
WHERE cc.id = pc.component_id AND p.id = pc.page_id
  AND p.site_id = '9ec3b9ee-5b08-461b-b4f8-9e1e03579c74' AND p.name = 'scout'
  AND cc.function = 'hero';

UPDATE page_components pc
SET content_data = pc.content_data || '{"primary_cta_url": "/tools/gauntlet/index.html", "secondary_cta_url": "/tools/archetype-taster-quiz/index.html"}'::jsonb, updated_at = NOW()
FROM content_components cc, pages p
WHERE cc.id = pc.component_id AND p.id = pc.page_id
  AND p.site_id = '9ec3b9ee-5b08-461b-b4f8-9e1e03579c74' AND p.name = 'scout'
  AND cc.function = 'call-to-action';

UPDATE page_components pc
SET content_data = pc.content_data || '{"eyebrow_text": "ARCHETYPE PROFILE", "heading": "Turns Admiration Into Learning.", "body_text": "The Mentor turns admiration into learning. Your breakdowns are among the most saved content on the platform \u2014 because you don''t just hold a position, you show the room how it works. Teaching is your edge: when you explain a take, the next hundred positions get sharper. Knowledge shared forward compounds.", "image_alt": "Stylised icon of The Mentor archetype \u2014 one figure''s insight passed forward and multiplied", "stat_1_value": "Most", "stat_1_label": "Saved breakdowns on the platform", "stat_2_value": "Clearest", "stat_2_label": "Explanations in the room", "stat_3_value": "Deepest", "stat_3_label": "Knowledge shared forward", "cta_label": "Find Your Archetype", "cta_url": "/tools/archetype-taster-quiz/index.html"}'::jsonb, updated_at = NOW()
FROM content_components cc, pages p
WHERE cc.id = pc.component_id AND p.id = pc.page_id
  AND p.site_id = '9ec3b9ee-5b08-461b-b4f8-9e1e03579c74' AND p.name = 'mentor'
  AND cc.function = 'content-block-about';

UPDATE page_components pc
SET content_data = pc.content_data || '{"headline": "The Mentor. The Most Saved Voice in the Room.", "subheadline": "Earned, not chosen. The Archetype whose breakdowns turn positions into teaching material.", "cta_text": "Run the Gauntlet", "cta_url": "/tools/gauntlet/index.html", "secondary_cta": "Take the Archetype Quiz", "secondary_cta_url": "/tools/archetype-taster-quiz/index.html"}'::jsonb, updated_at = NOW()
FROM content_components cc, pages p
WHERE cc.id = pc.component_id AND p.id = pc.page_id
  AND p.site_id = '9ec3b9ee-5b08-461b-b4f8-9e1e03579c74' AND p.name = 'mentor'
  AND cc.function = 'hero';

UPDATE page_components pc
SET content_data = pc.content_data || '{"primary_cta_url": "/tools/gauntlet/index.html", "secondary_cta_url": "/tools/archetype-taster-quiz/index.html"}'::jsonb, updated_at = NOW()
FROM content_components cc, pages p
WHERE cc.id = pc.component_id AND p.id = pc.page_id
  AND p.site_id = '9ec3b9ee-5b08-461b-b4f8-9e1e03579c74' AND p.name = 'mentor'
  AND cc.function = 'call-to-action';

DO $$
DECLARE bad INT; wrongcta INT;
BEGIN
  -- every content block must now name its archetype
  SELECT COUNT(*) INTO bad
  FROM page_components pc JOIN content_components cc ON cc.id=pc.component_id
  JOIN pages p ON p.id=pc.page_id
  WHERE p.site_id='9ec3b9ee-5b08-461b-b4f8-9e1e03579c74' AND p.page_type='entity-page'
    AND cc.function='content-block-about'
    AND pc.content_data->>'body_text' NOT ILIKE '%' || p.title || '%';
  -- no CTA may still point at /contact.html or /about
  SELECT COUNT(*) INTO wrongcta
  FROM page_components pc JOIN pages p ON p.id=pc.page_id
  WHERE p.site_id='9ec3b9ee-5b08-461b-b4f8-9e1e03579c74' AND p.page_type='entity-page'
    AND (pc.content_data->>'cta_url' IN ('/contact.html','/about')
      OR pc.content_data->>'primary_cta_url' IN ('/contact.html','/about'));
  IF bad > 0 OR wrongcta > 0 THEN
    RAISE EXCEPTION 'verify failed: % blocks missing archetype name, % wrong ctas', bad, wrongcta;
  END IF;
  RAISE NOTICE 'verified: all 8 content blocks name their archetype; all CTAs corrected';
END $$;

COMMIT;
