-- 166_vonc_evidence_base.sql — evidence_base for vonc.com (Experience Loop
-- guard rail 4, RUNBOOK T2.4; claims-verification V0 pattern, second site
-- after leopardess).
--
-- vonc is a fictional-brand test site: its product entities (Spark, the
-- Gauntlet, Archetypes) are allowed fiction. Its STATS are not — a number
-- presented as live usage ("12,847 Competitors", a "Live" leaderboard of
-- invented users) is fabricated social proof, the exact class the
-- anti-fabrication rule bans. Facts start EMPTY: no quantitative claim is
-- currently supportable; the experience spec's data contracts (RUNBOOK T3.2
-- §3) are the only path by which numbers become assertable.
--
-- Enforcement lanes: V1a build gate + V1b unverified_claims discovery check
-- activate when the claims image ships (code committed, not yet deployed;
-- unverified_claims is already pre-enabled on quality-discovery-agent).
-- claimscan (operator CLI) works immediately against this row.
--
-- Applied out of band (psql -f + ledger row same sitting per
-- bugs_open/aaa_fails_to_mend/007).

BEGIN;

UPDATE site_specs SET is_current = false, superseded_at = NOW()
WHERE site_id = '9ec3b9ee-5b08-461b-b4f8-9e1e03579c74'
  AND aspect = 'evidence_base' AND is_current;

INSERT INTO site_specs (site_id, aspect, data, source, notes, is_current, pinned, created_by)
VALUES (
  '9ec3b9ee-5b08-461b-b4f8-9e1e03579c74',
  'evidence_base',
  $json${
    "governing_rule": "No quantitative claim, usage stat, user name, or social proof ships unless it traces to a fact below or to a data contract in the current EXPERIENCE_PLAN (doc_plans subject_type=experience, subject_key=vonc-spark-game). A not-yet feature must be ABSENT or labelled coming-soon — never simulated.",
    "audit_doc": "docs/agent_docs/docs024_key_docs_latest/experience_loop/ (PLAN §1 diagnosis; live-page harvest 2026-07-17)",
    "schema_notes": "facts[]: {id, kind: capability|metric, claim, value?, source: {sql|artifact}, tolerance?, verified_at}. banned_claims[]: {pattern (regex over lowercased assertion text), reason}. allowed_entities[]: fictional product nouns that are NOT claims.",
    "facts": [],
    "banned_claims": [
      {"pattern": "12,847", "reason": "2026-07-17 gauntlet harvest: invented competitor count on a live page"},
      {"pattern": "94,210", "reason": "2026-07-17 gauntlet harvest: invented challenges-completed count"},
      {"pattern": "[0-9][0-9,]* ?(competitors|challenges completed|players|participants)", "reason": "usage-count class: no gameplay telemetry exists; no count is assertable until a data contract provides one"},
      {"pattern": "(avg\\.?|average) ?win rate", "reason": "2026-07-17 gauntlet harvest ('38% Avg. Win Rate'): no telemetry, no computable rate"},
      {"pattern": "day streak", "reason": "2026-07-17 gauntlet harvest ('7 Day Streak'): no per-user persistence exists site-wide"},
      {"pattern": "axonfury|zerorush|nexvoid|skorch|proxima", "reason": "2026-07-17 gauntlet harvest: invented leaderboard users presented under a 'Live' label"},
      {"pattern": "leaderboard", "reason": "no leaderboard exists (static site, no server); banned until a data contract defines a real one — coming-soon copy trips this deliberately for review"},
      {"pattern": "everyone else has (already )?filed", "reason": "2026-07-17 gauntlet harvest: fabricated social pressure"},
      {"pattern": "live now", "reason": "'GAUNTLET LIVE NOW' on a non-functional mock; nothing may claim liveness unless the control it labels actually works"}
    ],
    "allowed_entities": [
      "Spark", "vonc", "the Gauntlet", "the Arena", "Provocations", "Positions", "Challenges",
      "Archetype", "Catalyst", "Judge", "Maker", "Mentor", "Oracle", "Scout", "Surgeon", "Wildcard",
      "Elite Badge", "Iron Conviction Trial"
    ]
  }$json$::jsonb,
  'migration',
  'Guard rail 4 (experience loop): fabricated-stats class banned; facts deliberately empty until EXPERIENCE_PLAN data contracts exist',
  true, true, '166_vonc_evidence_base'
);

INSERT INTO doc_notes (id, subject_type, subject_key, body, categories, source, created_by)
VALUES (
  gen_random_uuid(),
  'experience', 'vonc-spark-game',
  '## evidence_base created for vonc (guard rail 4)
Observed: the gauntlet page ships invented usage stats (12,847 competitors, 94,210 challenges, 38% win rate, 7-day streak), five invented leaderboard users under a Live label, and fabricated social pressure copy — with no evidence_base row, every claims lane was blind to vonc.
Fix: evidence_base seeded (migration 166): facts EMPTY (nothing assertable), 9 banned patterns covering the harvested fabrications and their classes, allowed_entities for the fictional product nouns. V1a/V1b activate when the claims image ships; claimscan usable immediately.
Categories: fix, guard-rail',
  '["fix","guard-rail"]'::jsonb,
  'migration', '166_vonc_evidence_base'
);

COMMIT;
