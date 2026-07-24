-- 043 wave-2c (2026-07-24) — vonc prose fabrications OUTSIDE stat fields, found
-- chasing the last "4h 12m" grep hit after the stat blocks were cleaned.
--
-- FIXED HERE (unambiguous 043 class; vonc's own evidence_base banned_claims —
-- migration 166 — already bans liveness claims on non-functional features):
--   1. Fabricated countdowns: "Gauntlet closes in 4h 12m" (index),
--      "closes in 3h 44m" (about) — no clock exists.
--   2. Liveness theatre: panel_body "watch the split happen in real time. The
--      clock is live. The takes are stacking."; platform-comparison "Your
--      Archetype updated in real time" — no server, no persistence, nothing is
--      real-time. Also names "the Contrarian, the Realist" — archetypes that
--      do not exist on this site.
--   3. archetype-combinations (archetypes page): ALL THREE cards are built on
--      invented archetypes — Contrarian, Analyst, Idealist, Provocateur,
--      Realist, Sage — none of which are the site's documented eight (Catalyst,
--      Judge, Maker, Mentor, Oracle, Scout, Surgeon, Wildcard). Rewritten to
--      real pairs in the site's voice.
--   4. about differentiators: one invented-name swap ("a Contrarian or a
--      Pragmatist" -> "a Catalyst or a Judge") — surgical, not a rewrite.
--
-- RECORDED, NOT TOUCHED (experience-loop / vonc-spark thread's territory — the
-- present-tense product-VISION copy): the arena guide article ("Every day, a
-- new Provocation drops… watch the distribution shift in real time") and the
-- conceptual differentiators ("The Gauntlet Has a Clock", "The World reads
-- your pattern"). Their 166 banned_claims deliberately routes such copy to
-- review rather than banning the concept; rewriting the concept is their call.
-- Listed in bugs_open/043 Update 2026-07-24.
--
-- Apply:  kubectl -n ai-persona-system exec -i postgres-clients-0 -- \
--           psql -U clients_user -d clients_db < THIS_FILE

\set vonc '9ec3b9ee-5b08-461b-b4f8-9e1e03579c74'

BEGIN;

-- 1+2. index gauntlet-cta: honest panel + no countdown
UPDATE page_components pc
   SET content_data = content_data || jsonb_build_object(
     'panel_body', 'You click in, the Provocation is already waiting. File your position, see the split, and find out whether you read the room or against it — Catalyst, Judge, Wildcard, or something the room didn''t expect. No sign-up, no account. Step in.',
     'urgency_message', 'A Provocation is always open — step in whenever you''re ready.'
   ), updated_at = now()
 WHERE pc.page_id = (SELECT id FROM pages WHERE site_id = :'vonc' AND name='index')
   AND pc.content_data->>'urgency_message' ~ 'closes in';

-- about gauntlet-cta: no countdown
UPDATE page_components pc
   SET content_data = content_data || jsonb_build_object(
     'urgency_message', 'The Gauntlet is always open — no sign-up, straight in.'
   ), updated_at = now()
 WHERE pc.page_id = (SELECT id FROM pages WHERE site_id = :'vonc' AND name='about')
   AND pc.content_data->>'urgency_message' ~ 'closes in';

-- about platform-comparison: no real-time claim
UPDATE page_components pc
   SET content_data = content_data || jsonb_build_object(
     'row4_spark_value', 'Challenged. Defended. Your Archetype read from the positions you take.'
   ), updated_at = now()
 WHERE pc.page_id = (SELECT id FROM pages WHERE site_id = :'vonc' AND name='about')
   AND pc.content_data->>'row4_spark_value' ~* 'real.?time';

-- 4. about differentiators: invented-name swap inside the features JSON
UPDATE page_components pc
   SET content_data = jsonb_set(content_data, '{features}',
        replace(content_data->>'features',
                'a Contrarian or a Pragmatist',
                'a Catalyst or a Judge')::jsonb),
       updated_at = now()
 WHERE pc.page_id = (SELECT id FROM pages WHERE site_id = :'vonc' AND name='about')
   AND pc.content_data->>'features' LIKE '%a Contrarian or a Pragmatist%';

-- 3. archetypes archetype-combinations: three cards on the REAL eight
UPDATE page_components pc
   SET content_data = content_data || jsonb_build_object(
     'card1_badge', 'Catalyst + Judge',
     'card1_title', 'The Calculated Disruptor',
     'card1_icon_a', '⚡', 'card1_icon_b', '⚖️',
     'card1_trait1', 'Spark-First', 'card1_trait2', 'Scored', 'card1_trait3', 'Unsparing',
     'card1_description', 'You don''t just light the fire — you keep score of what it burns. The Catalyst in you starts the argument the room was avoiding; the Judge insists it gets settled on the merits. Takes from this combination open loud and close precise.',
     'card2_badge', 'Oracle + Wildcard',
     'card2_title', 'The Unpredictable Prophet',
     'card2_icon_a', '🦉', 'card2_icon_b', '🃏',
     'card2_trait1', 'Far-Sighted', 'card2_trait2', 'Untamed', 'card2_trait3', 'Never Boring',
     'card2_description', 'You see where it''s heading and you still refuse the script. The Oracle in you reads the pattern early; the Wildcard files the position nobody had priced in. When this combination is right, it''s right before everyone else — and louder about it.',
     'card3_badge', 'Surgeon + Mentor',
     'card3_title', 'The Reluctant Authority',
     'card3_icon_a', '🔬', 'card3_icon_b', '🧭',
     'card3_trait1', 'Precise', 'card3_trait2', 'Patient', 'card3_trait3', 'Rarely Wrong',
     'card3_description', 'You''ve seen this before, and you can prove it. The Surgeon in you cuts the take down to the one claim that matters; the Mentor explains why, so it sticks. Takes from this combination land quietly and stay landed.'
   ), updated_at = now()
 WHERE pc.page_id = (SELECT id FROM pages WHERE site_id = :'vonc' AND name='archetypes')
   AND pc.content_data->>'card1_badge' = 'Contrarian + Analyst';

\echo '--- verify: no invented archetypes/countdowns/real-time left in the touched pages ---'
SELECT p.name, count(*) AS remaining_bad
FROM page_components pc JOIN pages p ON p.id=pc.page_id
CROSS JOIN LATERAL jsonb_each_text(pc.content_data) e(k,v)
WHERE p.site_id = :'vonc' AND p.name IN ('index','about','archetypes')
  AND e.v ~* 'closes in [0-9]|clock is live|takes are stacking|updated in real.?time|\m(Analyst|Idealist|Provocateur, the|Realist|Sage|Pragmatist)\M'
  AND NOT (p.name='about' AND e.k='role_title')  -- "Provocateur, Referee & Architect" is a role phrase, not an archetype claim
GROUP BY p.name;

-- Re-render the three pages
INSERT INTO site_work_items
  (site_id, source, item_type, severity, summary, status, created_by, pipeline,
   priority, triaged_at, handler_agent, item_key, spec)
SELECT p.site_id, 'fabricated-stats-043-wave2c', 'page_rerender', 'medium',
  'Rerender ' || p.name || ' — fabricated countdowns/liveness and invented archetype names corrected (043 wave-2c)',
  'triaged', 'session-2026-07-24-043-treatment', 'build',
  20, now(), 'page-rerender',
  'page_rerender_' || p.name || '_043w2c_' || p.site_id::text,
  jsonb_build_object('domain','vonc.com','reason','cta_links_stale',
                     'page_id',p.id,'page_name',p.name,'filename',ltrim(p.url,'/'))
FROM pages p WHERE p.site_id = :'vonc' AND p.name IN ('index','about','archetypes');

\echo '--- queued ---'
SELECT status, count(*) FROM site_work_items WHERE source='fabricated-stats-043-wave2c' GROUP BY 1;

COMMIT;
