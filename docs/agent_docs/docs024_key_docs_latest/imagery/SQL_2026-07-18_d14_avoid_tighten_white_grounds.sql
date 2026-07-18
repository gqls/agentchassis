-- D14 pilot follow-up (2026-07-18): the cycle-time hero came back with a
-- WHITE ground — Banana drifted off "deep charcoal ground" once in three.
-- Tighten the content_hero override's avoid list with explicit light-ground
-- terms (user gate ruling: re-roll cycle-time only; payload's blue-heavy
-- variant accepted). Supersede-row pattern on 361f2ed7 (which superseded
-- the I1 seed 439329c4 earlier in D14).
--
-- Site: robot-hands.com 00ff3af5-dad8-4770-9f70-3edc267a3c92

BEGIN;

UPDATE site_specs
   SET is_current = false,
       superseded_at = now(),
       updated_at = now()
 WHERE id = '361f2ed7-2f5e-4da0-bfd9-b37b46fd9f62'
   AND is_current = true;

INSERT INTO site_specs (site_id, aspect, data, source, source_agent, notes, is_current, created_by)
SELECT site_id,
       aspect,
       jsonb_set(data, '{kinds,content_hero,avoid}',
           to_jsonb((data #>> '{kinds,content_hero,avoid}') ||
                    ', white background, pale background, light background, bright full-bleed colour field')),
       source,
       source_agent,
       'D14 pilot fix 2026-07-18: avoid gains light-ground terms after 1/3 white-ground drift. Supersedes 361f2ed7-2f5e-4da0-bfd9-b37b46fd9f62.',
       true,
       'imagery-i3-d14'
  FROM site_specs
 WHERE id = '361f2ed7-2f5e-4da0-bfd9-b37b46fd9f62';

COMMIT;
