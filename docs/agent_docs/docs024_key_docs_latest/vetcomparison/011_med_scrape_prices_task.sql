-- 011_med_scrape_prices_task.sql — create the med-scrape-prices scheduled task
-- (vetcomparison series; 2026-07-23)
--
-- WHY THIS FILE EXISTS. sql_for_agents/096 §3 tried to repoint an existing
-- 'med-scrape-prices' row at the med-price-scrape-orchestrator with an UPDATE —
-- but no such row was ever created, so the UPDATE was a silent no-op and
-- nothing populates med_price_snapshots. This file owns the row now, as an
-- idempotent INSERT (modelled on 009's directory-export-json task).
--
-- CONTRACT (bugs_closed/054): scheduled_tasks.input_data is the PAYLOAD ONLY.
-- The scheduler's fireTrigger supplies action="orchestrate" and
-- config.agent_type (from target_agent_type) and wraps this column as
-- input_data itself. Never put action/config/input_data keys inside it.
--
-- CADENCE. batch_size 20, stalest-first (vet_med_price_scrape_action.go orders
-- by last_scraped_at). 304 active listings / 20 per run at 21600s (6h) covers
-- the full set in ~4 days — comfortably inside the exporter's hard 14-day
-- freshness window (loadMedPricesForExport: collected_at > NOW() - '14 days').
--
-- SEEDED DISABLED. Enable deliberately, one at a time, per the RUNBOOK:
-- discover → scrape → export, each verified before the next. Prerequisites at
-- enable time: FIRECRAWL_API_KEY present in the business-intel pod env
-- (passthrough to spawned workers), and the med agent_definitions image_tag
-- current (spawned temp pods run agent_definitions.image_tag, NOT the deployed
-- chassis image).

INSERT INTO scheduled_tasks (
    name, description, target_agent_type, target_topic,
    interval_seconds, enabled, input_data, concurrency_group
) VALUES (
    'med-scrape-prices',
    'Scrape medicine prices from active retailers (stalest listings first)',
    'med-price-scrape-orchestrator',
    'system.agent.business-intel.requests',
    21600, false,
    '{"batch_size": 20}'::jsonb,
    'med-collection'
) ON CONFLICT (name) DO UPDATE SET
    target_agent_type = EXCLUDED.target_agent_type,
    target_topic      = EXCLUDED.target_topic,
    input_data        = EXCLUDED.input_data,
    updated_at        = NOW();
-- NOTE: 'enabled' is deliberately NOT in the DO UPDATE set — a re-seed must
-- never flip a task the owner has enabled/disabled by hand.

-- Verify
SELECT name, target_agent_type, enabled, interval_seconds, input_data
FROM scheduled_tasks WHERE name = 'med-scrape-prices';
