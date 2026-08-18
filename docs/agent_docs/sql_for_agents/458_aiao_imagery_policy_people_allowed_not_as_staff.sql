-- 458_aiao_imagery_policy_people_allowed_not_as_staff.sql
--
-- OWNER RULING 2026-08-18, ai-agent-orchestration.com imagery policy, verbatim:
--
--   "we don't want fake headshots of people in the about us, but we can use pictures
--    of typical offices or people working as long as we are not pretending that they
--    are part of the company"
--
-- This RELAXES the site's standing instruction and NARROWS it at the same time, and
-- both halves matter:
--
--   RELAXED — `imagery_direction` currently says "Technical illustrations and
--   architectural diagrams ONLY … no stock photography of people in meetings, no
--   handshakes … never staged corporate photography". Under the ruling, people at
--   work and ordinary offices are now permissible.
--
--   NARROWED — the thing actually banned is not "people" but IMPERSONATION: an image
--   that invites a visitor to read a stranger as a member of this company. The old
--   `avoid` line ("Testimonial carousels with headshots of fake people") banned one
--   VEHICLE (a carousel) rather than the deception, so it left the same deception
--   reachable by any other layout — an about-page grid, a team strip, a founder
--   quote. The replacement bans the act.
--
-- WHY THIS IS A SPEC CHANGE AND NOT A COMPONENT CHANGE. `design_intent` is the
-- instruction set the framework writes from (CLAUDE.md: the framework writes the
-- content, not us). `departments-grid` and `leadership-team` — the two about-page
-- components on this site — carry a 120px circular `.member-icon` slot, which is a
-- headshot-shaped hole. The policy has to be right BEFORE anything fills it.
--
-- SUPERSEDE, DO NOT MUTATE. `site_specs` carries `is_current` / `superseded_at` and a
-- partial unique index on (site_id, aspect) WHERE is_current, so the history is the
-- point: the old row is closed and a new current row inserted in one transaction.
-- Mutating in place would destroy the record of what the site was told before.
--
-- SITE-SCOPED. One row, one site. Nothing shared, no other site's policy touched.
--
-- ROLLBACK: 458_aiao_imagery_policy_people_allowed_not_as_staff_ROLLBACK.sql

BEGIN;

-- 1. Close the current row.
UPDATE site_specs
SET is_current = false, superseded_at = now()
WHERE site_id = '2a8ebf9c-20a2-4c39-b191-840b012371da'
  AND aspect = 'design_intent'
  AND is_current;

-- 2. Insert the successor, carrying every other key forward untouched.
INSERT INTO site_specs (site_id, aspect, data, source, source_agent, notes, is_current, created_by)
SELECT
  ss.site_id,
  ss.aspect,
  jsonb_set(
    jsonb_set(
      ss.data,
      '{imagery_direction}',
      to_jsonb(
        'Technical illustrations and architectural diagrams are the default and should carry most ' ||
        'pages: agent topologies, message flows, deployment architectures, monitoring dashboards. ' ||
        'Clean vector work in the dark palette. Photography IS permitted where it earns its place — ' ||
        'ordinary working environments, people at work, offices, desks, screens, server rooms, ' ||
        'whiteboards mid-discussion. ' ||
        'THE ONE HARD RULE: never present a photographed person as a member of this company. No ' ||
        'named or implied staff, no founder or team headshots, no captions, roles or quotes ' ||
        'attributed to a stock face, and no placement that invites the reader to take a stranger ' ||
        'for someone who works here — an about-page team grid being the obvious trap. Generic ' ||
        'people-at-work imagery is illustrative and must read as illustrative. ' ||
        'Still avoid: staged corporate handshakes, boardroom stock clichés, and abstract AI-brain ' ||
        'or neural-network imagery.'
      )
    ),
    '{avoid}',
    (
      -- Drop the old carousel-shaped line, keep every other avoid entry in order,
      -- then append the two that state the ruling.
      SELECT jsonb_agg(v ORDER BY ord)
      FROM (
        SELECT value AS v, ordinality AS ord
        FROM jsonb_array_elements(ss.data->'avoid') WITH ORDINALITY
        WHERE value <> '"Testimonial carousels with headshots of fake people"'::jsonb
        UNION ALL
        SELECT '"Any photographed person presented, captioned or implied as a member of this company"'::jsonb, 1000
        UNION ALL
        SELECT '"Invented team members, founder headshots, or testimonials attributed to a stock face"'::jsonb, 1001
      ) x
    )
  ) AS data,
  ss.source,
  'session:site_ai_agent_orchestration',
  'Owner ruling 2026-08-18: people-at-work and office photography permitted; impersonation of company staff banned. Supersedes the carousel-shaped avoid line, which banned a vehicle rather than the deception. See migration 458.',
  true,
  'session:site_ai_agent_orchestration'
FROM site_specs ss
WHERE ss.site_id = '2a8ebf9c-20a2-4c39-b191-840b012371da'
  AND ss.aspect = 'design_intent'
  AND ss.superseded_at IS NOT NULL
ORDER BY ss.superseded_at DESC
LIMIT 1;

DO $$
DECLARE
  current_rows int;
  has_ban      int;
  has_old      int;
  keys_kept    int;
  keys_before  int;
BEGIN
  SELECT count(*) INTO current_rows FROM site_specs
   WHERE site_id='2a8ebf9c-20a2-4c39-b191-840b012371da' AND aspect='design_intent' AND is_current;
  IF current_rows <> 1 THEN
    RAISE EXCEPTION '458: expected exactly 1 current design_intent row, found %', current_rows;
  END IF;

  SELECT count(*) INTO has_ban FROM site_specs
   WHERE site_id='2a8ebf9c-20a2-4c39-b191-840b012371da' AND aspect='design_intent' AND is_current
     AND data->'avoid' @> '["Any photographed person presented, captioned or implied as a member of this company"]'::jsonb;
  IF has_ban <> 1 THEN
    RAISE EXCEPTION '458: the impersonation ban is not present in the new avoid list';
  END IF;

  SELECT count(*) INTO has_old FROM site_specs
   WHERE site_id='2a8ebf9c-20a2-4c39-b191-840b012371da' AND aspect='design_intent' AND is_current
     AND data->'avoid' @> '["Testimonial carousels with headshots of fake people"]'::jsonb;
  IF has_old <> 0 THEN
    RAISE EXCEPTION '458: the superseded carousel-shaped avoid line is still present';
  END IF;

  -- Nothing else may be lost: the successor must carry every top-level key the
  -- predecessor had. This is the guard that matters — a jsonb_set typo silently
  -- drops a sibling key and the row still looks well-formed.
  -- ⚠ TAKE THE ROW FIRST, THEN EXPAND ITS KEYS. Writing this as
  --   SELECT jsonb_object_keys(data) FROM site_specs WHERE ... LIMIT 1
  -- applies the LIMIT to the EXPANDED KEYS, not to the row, so it returns 1 and the
  -- guard fires a false "a key was dropped". That is exactly what happened on the
  -- first run of this migration; the guard was right to refuse and wrong about why.
  SELECT count(*) INTO keys_kept FROM jsonb_object_keys((
    SELECT data FROM site_specs
     WHERE site_id='2a8ebf9c-20a2-4c39-b191-840b012371da' AND aspect='design_intent'
       AND is_current));
  SELECT count(*) INTO keys_before FROM jsonb_object_keys((
    SELECT data FROM site_specs
     WHERE site_id='2a8ebf9c-20a2-4c39-b191-840b012371da' AND aspect='design_intent'
       AND superseded_at IS NOT NULL ORDER BY superseded_at DESC LIMIT 1));
  IF keys_kept <> keys_before THEN
    RAISE EXCEPTION '458: successor has % top-level keys, predecessor had % — a key was dropped', keys_kept, keys_before;
  END IF;

  RAISE NOTICE '458 OK: imagery policy updated (people-at-work permitted, impersonation banned); % keys carried forward.', keys_kept;
END $$;

COMMIT;
