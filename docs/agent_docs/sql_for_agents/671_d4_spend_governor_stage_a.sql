-- 671_d4_spend_governor_stage_a.sql — D4 LLM spend governor, STAGE A (meter + classes + state).
-- dispatch_throughput lane. Owner ruling 2026-08-31 (verbatim in the lane NOTES): shed order is
-- routine maintenance FIRST, new site builds SECOND, research THIRD (most protected). Supersedes
-- the 08-21 order (RESEARCH §10 corrected visibly). Design: lane PLAN §"D4 — LLM SPEND GOVERNOR".
--
-- WHAT THIS STAGE IS: the DB half only — model prices, the item-class map, config, computed
-- state, the spend view, and a 120s scheduled task that keeps state current. It ENFORCES
-- NOTHING: no work is refused until stage B (a claim-step check in Go) ships AND
-- governor_config.enabled=true AND monthly_budget_usd is set. Inert three separate ways.
--
-- Facts this file rests on, checked 2026-08-31 at the artefact:
--   · llm_call_log retains to 2026-03-25 (~5 months), ~1,800 rows/day — a month-to-date scan
--     every 120s is cheap and complete.
--   · Prices verified at platform.claude.com/docs/en/about-claude/pricing 2026-08-31 (Sonnet 5's
--     introductory $2/$10 made permanent). Cache writes are priced at the 1h-TTL rate (2x base):
--     the log does not record write TTL, and a governor should err HIGH.
--   · Class map seeded from the measured 14d item_type→handler map (lane NOTES 2026-08-31).
--     An item_type ABSENT from the map is treated by stage B as maintenance + llm_bearing
--     (sheds earliest — the safe default for an unknown spender).
--   · doc_notes.subject_type is constrained; 'pipeline' is the fitting value.
--
-- Rerun-safe: a replay RAISEs at the refusal-first guard below.
-- Rollback: 671_d4_spend_governor_stage_a_ROLLBACK.sql (drops the five objects + the task row).

BEGIN;

-- Refusal FIRST (LANDMINES 2026-08-26 replay-decoy ordering: refuse before any side effect).
DO $$
BEGIN
  IF EXISTS (SELECT 1 FROM information_schema.tables WHERE table_schema='public'
               AND table_name='governor_config') THEN
    RAISE EXCEPTION '671 REFUSED: governor_config already exists — this is a replay. Nothing was changed.';
  END IF;
  IF EXISTS (SELECT 1 FROM scheduled_tasks WHERE name='spend-governor-state') THEN
    RAISE EXCEPTION '671 REFUSED: scheduled task spend-governor-state already exists. Nothing was changed.';
  END IF;
END $$;

-- 1. Model prices. NULL rate = unpriced: the view SURFACES those tokens, never drops them
--    (a governor blind to a spender is the classifier-gap failure class).
CREATE TABLE governor_model_prices (
  model                    text PRIMARY KEY,
  input_usd_per_mtok       numeric,
  output_usd_per_mtok      numeric,
  cache_write_usd_per_mtok numeric,   -- 1h-TTL rate (2x base) by policy; see header
  cache_read_usd_per_mtok  numeric,
  note                     text,
  updated_at               timestamptz NOT NULL DEFAULT now()
);

INSERT INTO governor_model_prices (model, input_usd_per_mtok, output_usd_per_mtok,
                                   cache_write_usd_per_mtok, cache_read_usd_per_mtok, note) VALUES
  ('claude-sonnet-5',   2.00, 10.00,  4.00, 0.20, 'verified 2026-08-31; introductory price made permanent'),
  ('claude-sonnet-4-6', 3.00, 15.00,  6.00, 0.30, 'verified 2026-08-31'),
  ('claude-opus-4-8',   5.00, 25.00, 10.00, 0.50, 'verified 2026-08-31'),
  ('claude-opus-4-6',   5.00, 25.00, 10.00, 0.50, 'verified 2026-08-31'),
  ('claude-haiku-4-5',  1.00,  5.00,  2.00, 0.10, 'verified 2026-08-31'),
  ('mistral-small3.1',  0, 0, 0, 0,               'local Ollama — no per-token cost'),
  ('gemini-pro-latest', NULL, NULL, NULL, NULL,   'Google — rate not entered; surfaces as unpriced');

-- 2. Item-class map (owner-adjustable data, not code). class per the 2026-08-31 ruling;
--    llm_bearing=false marks classes whose handlers make no model calls — they are NEVER shed
--    (shedding them saves nothing and stops serving).
CREATE TABLE governor_work_class_map (
  item_type   text PRIMARY KEY,
  class       text NOT NULL CHECK (class IN ('maintenance','build','research')),
  llm_bearing boolean NOT NULL,
  note        text,
  updated_at  timestamptz NOT NULL DEFAULT now()
);

INSERT INTO governor_work_class_map (item_type, class, llm_bearing, note) VALUES
  -- maintenance, LLM-bearing (shed at L1)
  ('content_rewrite',         'maintenance', true,  'page-build-handler / rewrite path'),
  ('section_edit',            'maintenance', true,  'section-editor'),
  ('improve_tool',            'maintenance', true,  'tool-improver'),
  ('head_essentials_missing', 'maintenance', true,  'seed judgment — adjust if handler is LLM-free'),
  ('required_fields_missing', 'maintenance', true,  'required-fields-missing-handler'),
  ('acceptance_run',          'maintenance', true,  'tool-acceptance-agent'),
  ('audit_tool',              'maintenance', true,  'tool-auditor'),
  -- maintenance, LLM-free (never shed — llm_bearing=false)
  ('page_rerender',           'maintenance', false, 'renders + deploys, no model calls'),
  ('undeployed_asset',        'maintenance', false, 'asset-deployer'),
  ('needs_rerender',          'maintenance', false, 'rerender-pages, site-level'),
  ('deactivated_component',   'maintenance', false, 'rerender-pages'),
  ('orphan_blog_posts',       'maintenance', false, 'rerender-pages'),
  ('needs_content_image',     'maintenance', false, 'asset-deployer (deploy, not generation)'),
  -- build (shed at L2)
  ('needs_page',              'build',       true,  'page-build-handler'),
  ('needs_content_page',      'build',       true,  'page-build-handler'),
  ('needs_content_planning',  'build',       true,  'planning'),
  ('unbuilt_internal_link',   'build',       true,  'page-build-handler'),
  ('needs_imagery',           'build',       true,  'image-build-handler'),
  ('capability_gap',          'build',       true,  'seed judgment — tool building'),
  -- research (shed at L3, most protected)
  ('needs_vertical_research', 'research',    true,  'vertical research');

-- 3. Config. Single row. Three independent OFF switches: enabled=false (stage B reads it),
--    monthly_budget_usd NULL (state task computes level 0), and stage B not yet shipped.
CREATE TABLE governor_config (
  id                 smallint PRIMARY KEY DEFAULT 1 CHECK (id = 1),
  enabled            boolean NOT NULL DEFAULT false,
  monthly_budget_usd numeric,
  l1_pct             numeric NOT NULL DEFAULT 70,
  l2_pct             numeric NOT NULL DEFAULT 85,
  l3_pct             numeric NOT NULL DEFAULT 95,
  updated_at         timestamptz NOT NULL DEFAULT now(),
  CHECK (l1_pct < l2_pct AND l2_pct < l3_pct)
);
INSERT INTO governor_config (id) VALUES (1);

-- 4. Computed state. computed_at is the task's HEARTBEAT: a stale computed_at means the task
--    is not running, and must never read as "level 0, all fine".
CREATE TABLE governor_state (
  id                 smallint PRIMARY KEY DEFAULT 1 CHECK (id = 1),
  shed_level         int NOT NULL DEFAULT 0 CHECK (shed_level BETWEEN 0 AND 3),
  month              text,
  mtd_usd            numeric,
  unpriced_io_tokens bigint,
  computed_at        timestamptz
);
INSERT INTO governor_state (id) VALUES (1);

-- 5. The meter, single-sourced. COALESCE everywhere: token columns are NULL on some rows.
CREATE VIEW governor_spend_mtd AS
SELECT to_char(now(), 'YYYY-MM') AS month,
       round(sum( (COALESCE(l.input_tokens,0)                * COALESCE(p.input_usd_per_mtok,0)
                 + COALESCE(l.output_tokens,0)               * COALESCE(p.output_usd_per_mtok,0)
                 + COALESCE(l.cache_creation_input_tokens,0) * COALESCE(p.cache_write_usd_per_mtok,0)
                 + COALESCE(l.cache_read_input_tokens,0)     * COALESCE(p.cache_read_usd_per_mtok,0)
                  ) / 1e6 )::numeric, 2) AS mtd_usd,
       COALESCE(sum(COALESCE(l.input_tokens,0) + COALESCE(l.output_tokens,0))
                FILTER (WHERE p.model IS NULL OR p.input_usd_per_mtok IS NULL), 0) AS unpriced_io_tokens,
       count(*) FILTER (WHERE p.model IS NULL) AS calls_from_unlisted_models
FROM llm_call_log l
LEFT JOIN governor_model_prices p ON p.model = l.model
WHERE l.created_at >= date_trunc('month', now());

-- 6. The state task: recompute every 120s; note to doc_notes ONLY on a level change
--    (at 120s a note-per-run would be spam; the heartbeat is governor_state.computed_at).
INSERT INTO scheduled_tasks (name, description, interval_seconds, target_agent_type, target_topic,
                             pre_query, enabled, timeout_seconds, fire_message)
VALUES (
  'spend-governor-state',
  'D4 LLM spend governor: recompute month-to-date spend and shed level (671, stage A). Enforcement is stage B.',
  120, 'generic', 'system.agent.scheduled.requests',
  $PRE$
WITH cfg AS (SELECT * FROM governor_config WHERE id = 1),
spend AS (SELECT * FROM governor_spend_mtd),
new AS (
  SELECT CASE
           WHEN cfg.monthly_budget_usd IS NULL THEN 0
           WHEN spend.mtd_usd >= cfg.monthly_budget_usd * cfg.l3_pct/100 THEN 3
           WHEN spend.mtd_usd >= cfg.monthly_budget_usd * cfg.l2_pct/100 THEN 2
           WHEN spend.mtd_usd >= cfg.monthly_budget_usd * cfg.l1_pct/100 THEN 1
           ELSE 0
         END AS lvl,
         spend.mtd_usd, spend.unpriced_io_tokens, spend.month
  FROM cfg, spend),
old AS (SELECT shed_level FROM governor_state WHERE id = 1),
noted AS (
  INSERT INTO doc_notes (subject_type, subject_key, body, categories, source)
  SELECT 'pipeline', 'spend-governor',
         format('spend-governor: shed level %s -> %s (month-to-date $%s of budget $%s; unpriced io tokens %s)',
                old.shed_level, new.lvl, new.mtd_usd,
                COALESCE((SELECT monthly_budget_usd::text FROM cfg), 'UNSET'), new.unpriced_io_tokens),
         '["spend-governor"]'::jsonb, 'scheduled_tasks:spend-governor-state'
  FROM old, new WHERE old.shed_level <> new.lvl
  RETURNING 1),
upd AS (
  UPDATE governor_state s
     SET shed_level = new.lvl, month = new.month, mtd_usd = new.mtd_usd,
         unpriced_io_tokens = new.unpriced_io_tokens, computed_at = now()
    FROM new WHERE s.id = 1
  RETURNING s.shed_level)
SELECT (SELECT shed_level FROM upd)  AS shed_level,
       (SELECT count(*) FROM noted)  AS level_changed
$PRE$,
  true, 60, false);

-- Verify block: DO/RAISE (a SELECT-only verify cannot stop the COMMIT).
DO $$
DECLARE n int; lvl int; v_mtd numeric;
BEGIN
  SELECT count(*) INTO n FROM governor_model_prices WHERE input_usd_per_mtok IS NOT NULL;
  IF n < 5 THEN RAISE EXCEPTION '671 VERIFY: expected >=5 priced models, found %', n; END IF;

  SELECT count(*) INTO n FROM governor_work_class_map;
  IF n < 20 THEN RAISE EXCEPTION '671 VERIFY: class map has % rows, expected >=20', n; END IF;

  SELECT count(*) INTO n FROM governor_work_class_map WHERE NOT llm_bearing AND class <> 'maintenance';
  IF n <> 0 THEN RAISE EXCEPTION '671 VERIFY: % llm-free rows outside maintenance — seed error', n; END IF;

  PERFORM 1 FROM governor_config WHERE id=1 AND enabled=false AND monthly_budget_usd IS NULL;
  IF NOT FOUND THEN RAISE EXCEPTION '671 VERIFY: config row not in the shipped-inert shape'; END IF;

  -- The meter must run and return one row with a non-negative figure.
  SELECT mtd_usd INTO v_mtd FROM governor_spend_mtd;
  IF v_mtd IS NULL OR v_mtd < 0 THEN RAISE EXCEPTION '671 VERIFY: meter returned %', v_mtd; END IF;

  -- Execute the task's own computation once (proves the stored pre_query text is runnable):
  EXECUTE (SELECT pre_query FROM scheduled_tasks WHERE name='spend-governor-state');
  SELECT shed_level INTO lvl FROM governor_state WHERE id=1;
  IF lvl <> 0 THEN RAISE EXCEPTION '671 VERIFY: shed level % on a NULL budget — must be 0', lvl; END IF;
  PERFORM 1 FROM governor_state WHERE id=1 AND computed_at > now() - interval '1 minute';
  IF NOT FOUND THEN RAISE EXCEPTION '671 VERIFY: state computation did not stamp computed_at'; END IF;

  PERFORM 1 FROM scheduled_tasks WHERE name='spend-governor-state'
    AND enabled AND interval_seconds=120 AND fire_message=false;
  IF NOT FOUND THEN RAISE EXCEPTION '671 VERIFY: task row not in expected shape'; END IF;

  RAISE NOTICE '671 OK: governor stage A in place — meter $% MTD, level %, enforcement INERT (no budget, not enabled, no stage B)', v_mtd, lvl;
END $$;

COMMIT;
