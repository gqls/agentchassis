-- 043 treatment — gamesdesign.co.uk/index system-stats (2026-07-24).
--
-- WHAT WAS LIVE (rendered with junk suffixes: 36.6%, 4ms, 10,000+, 6x):
--   stat1 "PRD Accuracy Gap 36.6" — footnote claimed it was "derived from
--     standard binomial probability modelling against PRD at p=0.20". NO tool on
--     the site implements PRD; the figure, the C-constant narrative and the
--     derivation appear nowhere in any deployed artefact. FABRICATED.
--   stat2 "Pity System Variables 4" — description named "floor rate, ceiling
--     threshold, C Const…", parameters that do not exist. The real drop-rate
--     tuner has exactly 4 configurable inputs, but they are drop chance, kills
--     per hour, pity timer, target hours. Value coincidentally right,
--     description invented. REFRAMED to the real four.
--   stat3 "Monte Carlo Trials 10,000" — TRACES: the deployed drop-rate
--     simulator/tuner (and lanchester, xp-curve, bayesian tools) run 10000
--     iterations per query in their shipped JS. TRUE; kept; '+' dropped (it is
--     exactly 10,000).
--
--     >>> CORRECTED 2026-07-31 — bugs_open/161. THIS VERDICT WAS WRONG, on both
--     counts, and it is the most instructive line in this file: three of the four
--     stats here were caught as fabricated or reframed, and the one marked TRUE is
--     the one that was subtly false — the NUMBER was real and the NOUN was not.
--       (a) "Monte Carlo" is false. Neither tool samples anything: Math.random
--           count is 0 in both. The tuner is closed-form Math.pow(1-p,k) plus a
--           CDF ("Cumulative distribution modelled via geometric distribution with
--           optional hard pity cap", its own doc comment); the simulator computes
--           exact binomial probability. A Monte Carlo method IS random sampling, so
--           with no randomness there are no trials to count.
--       (b) "run 10000 iterations per query" is ALSO false. There is no fixed
--           10,000 iteration count anywhere. The tuner's CDF is sized
--           `maxKills = Math.max(1, kph * hours)` — from the USER's inputs. The
--           only real 10000 is `return Math.min(val, 10000)` in the simulator: an
--           input CLAMP, i.e. the MAXIMUM attempts it will model.
--     The trace stopped at finding the digits and took the technique word from the
--     copy it was auditing. Because the register is also the writer whitelist, that
--     one row then instructed writers to repeat it and vouched for it when they
--     did: 10 components published it and cmd/claimscan reported 0 findings,
--     correctly. Corrected wording: "maximum attempts modelled per query".
--   stat4 "Economy Model Types 6" — "faucet-and-sink economy archetypes
--     available in the balance tools": no economy presets exist in any tool.
--     FABRICATED.
--
-- Ground truth (2026-07-24): 11 interactive tool pages, all deployed and
-- serving; 10 written pieces (5 blog posts + 5 guides) deployed; 10000 in the
-- tools' JS (curl + grep, RUNBOOK method). tool-loot-table-balancer was READ
-- ONLY — it is the fix-loop benchmark and is never hand-modified.
--
-- Apply:  kubectl -n ai-persona-system exec -i postgres-clients-0 -- \
--           psql -U clients_user -d clients_db < THIS_FILE

\set site 'e33263f4-74f8-494f-b191-546845dbbddf'
\set comp 'fdd92ad4-521a-4602-89cf-7ee1a66c10f1'

BEGIN;

UPDATE page_components pc
   SET content_data = content_data || jsonb_build_object(
     'footnote_text', 'Figures trace to the deployed tools'' own code and the site''s page register, as of 2026-07-24. Nothing here is projected or modelled — if a number appears, the thing it counts is on this site.',
     'stat1_label', 'Interactive Design Tools',
     'stat1_value', (SELECT count(*)::text FROM pages
                     WHERE site_id = :'site' AND page_type='tool'
                       AND status='active' AND deployed_at IS NOT NULL),
     'stat1_suffix', '',
     'stat1_description', 'Calculators and simulators live on this site — drop rates, TTK, EHP, jump physics, XP curves, loot tables, damage formulas and more. All client-side, all free.',
     'stat2_label', 'Drop-Rate Tuner Inputs',
     'stat2_value', '4',
     'stat2_suffix', '',
     'stat2_description', 'The independently configurable parameters in the drop-rate tuner: drop chance, kills per hour, pity timer and target hours. Enough to model a pity system honestly, few enough to reason about.',
     -- CORRECTED 2026-07-31 (bugs_open/161): was 'Monte Carlo Trials' / 'The number of
     -- simulated runs the drop-rate tools execute per query to produce statistically
     -- stable results'. The tools run NO simulated runs — see the corrected trace note
     -- at the head of this file. Matches the live homepage, which another session
     -- repaired to "Max Attempts Modelled" on 2026-07-30.
     'stat3_label', 'Max Attempts Modelled',
     'stat3_value', '10,000',
     'stat3_suffix', '',
     'stat3_description', 'The largest attempt count the drop-rate tools will model in one query. They compute exact probability — a closed-form binomial/geometric distribution — rather than sampling, so the figure is a bound in the shipped code, not a number of simulated runs.',
     'stat4_label', 'Guides & Articles Published',
     'stat4_value', (SELECT count(*)::text FROM pages
                     WHERE site_id = :'site' AND page_type IN ('blog-post','guide')
                       AND status='active' AND deployed_at IS NOT NULL),
     'stat4_suffix', '',
     'stat4_description', 'Written design analysis currently live alongside the tools — mechanics explained, not just calculated.'
   ),
   updated_at = now()
 WHERE pc.page_id = (SELECT id FROM pages WHERE site_id = :'site' AND name='index')
   AND pc.component_id = :'comp';

\echo '--- after ---'
SELECT e.k, left(e.v,60) FROM page_components pc
JOIN pages p ON p.id=pc.page_id
CROSS JOIN LATERAL jsonb_each_text(pc.content_data) e(k,v)
WHERE p.site_id = :'site' AND p.name='index' AND pc.component_id = :'comp'
  AND e.k ~ 'stat[0-9]_(label|value|suffix)' ORDER BY e.k;

INSERT INTO site_work_items
  (site_id, source, item_type, severity, summary, status, created_by, pipeline,
   priority, triaged_at, handler_agent, item_key, spec)
SELECT p.site_id, 'fabricated-stats-043-gamesdesign', 'page_rerender', 'medium',
  'Rerender index — fabricated PRD/economy stats replaced with traced figures (043)',
  'triaged', 'session-2026-07-24-043-treatment', 'build',
  20, now(), 'page-rerender',
  'page_rerender_' || p.name || '_043stats_' || p.site_id::text,
  jsonb_build_object('domain','gamesdesign.co.uk','reason','cta_links_stale',
                     'page_id',p.id,'page_name',p.name,'filename',ltrim(p.url,'/'))
FROM pages p WHERE p.site_id = :'site' AND p.name='index';

\echo '--- queued ---'
SELECT status, count(*) FROM site_work_items WHERE source='fabricated-stats-043-gamesdesign' GROUP BY 1;

COMMIT;
