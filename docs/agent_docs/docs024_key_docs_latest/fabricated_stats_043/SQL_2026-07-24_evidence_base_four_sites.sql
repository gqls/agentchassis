-- 043 candidate 3, part 2 (2026-07-24): declare each fixed site's Verified
-- Facts via the EXISTING evidence_base machinery, so migration 201's rule has
-- figures to allow. Without a writer_block, 201 makes the writer return EMPTY
-- for unsourced numeric fields — honest, but a routine full-writer re-render
-- would then blank the stats we just corrected. With one, the writer is GIVEN
-- the true figures ("Verified Facts — the ONLY numbers you may assert") and
-- keeps them. Shape copied from the live relojistas/leopardess rows.
--
-- Upsert per the (site_id, aspect) WHERE is_current unique index: supersede any
-- current row, insert the new one.
--
-- Apply:  kubectl -n ai-persona-system exec -i postgres-clients-0 -- \
--           psql -U clients_user -d clients_db < THIS_FILE

BEGIN;

UPDATE site_specs SET is_current=false, superseded_at=now()
 WHERE aspect='evidence_base' AND is_current
   AND site_id IN ('00ff3af5-dad8-4770-9f70-3edc267a3c92','9ec3b9ee-5b08-461b-b4f8-9e1e03579c74',
                   'e33263f4-74f8-494f-b191-546845dbbddf','2a8ebf9c-20a2-4c39-b191-840b012371da');

INSERT INTO site_specs (site_id, aspect, data, source, created_by, notes)
VALUES
('00ff3af5-dad8-4770-9f70-3edc267a3c92', 'evidence_base', jsonb_build_object('writer_block',
$rh$NUMBERS (state only these, with their listed meaning; dated snapshots up to a listed live count are fine — recount from the products table before changing one):
- 10 gripper models indexed (products, category='gripper', status='active'; as of 2026-07-24)
- 6 manufacturers covered: Schunk, OnRobot, Robotiq, Zimmer Group, Festo, Schmalz
- 6 actuation technologies represented in the index: electric parallel-jaw (5 models), and one each of pneumatic parallel-jaw, vacuum, magnetic, soft-robotic, adhesive
- 4 parameters MatchMatrix tests an application against: gripping/holding force, jaw travel or grip range, rated payload, IP rating
- 39 published specification figures held across the index (as of 2026-07-24)
NOT TRACKED, NEVER STATE: MatchMatrix query/usage counts, visitor numbers, customer counts, satisfaction rates. Nothing on this site measures them at any value.$rh$),
 'manual', 'session-2026-07-24-043-treatment',
 'bug 043 candidate 3 part 2: verified facts so migration 201 has figures to allow'),

('9ec3b9ee-5b08-461b-b4f8-9e1e03579c74', 'evidence_base', jsonb_build_object('writer_block',
$vc$NUMBERS (state only these, with their listed meaning; recount from the pages register before changing one; as of 2026-07-24):
- 8 archetypes documented: Catalyst, Judge, Maker, Mentor, Oracle, Scout, Surgeon, Wildcard
- 3 interactive tools live: the Archetype Taster Quiz, the Arena, the Gauntlet
- 2 written pieces (guides/provocations) live; 17 pages live in total
- Free to play, no sign-up: true
NOT TRACKED, NEVER STATE: takes filed, positions filed, players scored, gauntlets contested, split percentages, streaks, ratings, customer counts, activity windows or any live-activity figure. The site records NO user activity; every such number at any value is an invention.$vc$),
 'manual', 'session-2026-07-24-043-treatment',
 'bug 043 candidate 3 part 2: verified facts so migration 201 has figures to allow'),

('e33263f4-74f8-494f-b191-546845dbbddf', 'evidence_base', jsonb_build_object('writer_block',
$gd$NUMBERS (state only these, with their listed meaning; as of 2026-07-24):
- 11 interactive design tools live, all client-side and free
- 10,000 maximum attempts modelled per query in the drop-rate tools (the tools compute EXACT probability — closed-form binomial/geometric — they do NOT run Monte Carlo trials or any random sampling; CORRECTED 2026-07-31, bugs_open/161)
- 4 configurable inputs in the drop-rate tuner: drop chance, kills per hour, pity timer, target hours
- 10 guides & articles live (5 blog posts + 5 guides)
NOT TRACKED / DOES NOT EXIST, NEVER STATE: user counts, accuracy-gap percentages, PRD figures (no tool on this site implements PRD), economy model/archetype counts, pity parameters other than the four listed.$gd$),
 'manual', 'session-2026-07-24-043-treatment',
 'bug 043 candidate 3 part 2: verified facts so migration 201 has figures to allow'),

('2a8ebf9c-20a2-4c39-b191-840b012371da', 'evidence_base', jsonb_build_object('writer_block',
$ao$NUMBERS (state only these, with their listed meaning; dated snapshots up to a listed live count are fine — these are live registry counts taken 2026-07-24, recount before restating):
- 170 active agent definitions in the production registry
- 13 live sites in production, built and operated end-to-end by the platform
- 17 backend services (the service manifests under deployments/kustomize/services)
- 1,267 automated work items completed (builds, re-renders, repairs, audits — the platform's work-item ledger)
- Architecture: Kubernetes, Kafka, Postgres — true and stated freely
NOT TRACKED / DOES NOT EXIST, NEVER STATE: clients served, departments served, satisfaction rates, awards won, concurrent-instance counts, uptime percentages. None of these are measured; every such figure at any value is an invention.$ao$),
 'manual', 'session-2026-07-24-043-treatment',
 'bug 043 candidate 3 part 2: verified facts so migration 201 has figures to allow');

\echo '--- verify: four current evidence_base rows ---'
SELECT s.domain, ss.is_current, length(ss.data->>'writer_block') AS block_len
FROM site_specs ss JOIN sites s ON s.id=ss.site_id
WHERE ss.aspect='evidence_base' AND ss.is_current
  AND ss.site_id IN ('00ff3af5-dad8-4770-9f70-3edc267a3c92','9ec3b9ee-5b08-461b-b4f8-9e1e03579c74',
                     'e33263f4-74f8-494f-b191-546845dbbddf','2a8ebf9c-20a2-4c39-b191-840b012371da')
ORDER BY s.domain;

COMMIT;

-- ---------------------------------------------------------------------------
-- CORRECTION, same session, ~15 min later. The blanket supersede above hit a
-- row I had not read: vonc already had a CURRENT structured evidence_base from
-- migration 166 (experience-loop thread) — facts[]/banned_claims[]/
-- allowed_entities, whose 9 banned_claims regexes feed the claims-verification
-- checkers. My writer_block-only row silently removed them from "current".
-- Caught by the UPDATE 1 count on apply (expected 0). Fixed by a MERGE row:
-- 166's structure || this file's writer_block (both machines served; 166 had no
-- writer_block key, which is why the WRITER kept inventing on vonc even with
-- checkers live). Lesson (WRONG_CALLS'd): read before superseding a
-- site_specs aspect row — another thread's machinery may be keyed on it.
-- ---------------------------------------------------------------------------
