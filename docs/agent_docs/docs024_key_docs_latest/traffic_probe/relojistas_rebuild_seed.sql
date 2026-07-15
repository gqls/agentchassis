-- ============================================================================
-- relojistas_rebuild_seed.sql — P1 + P2 artifacts (apply in order)
-- ----------------------------------------------------------------------------
-- P1a  sites onboarding UPDATE  (VM target + engine wiring)
-- P1b  classification news_feed recommendation (site_specs, versioned)
-- P2   content_sources seed     (5 VERIFIED Spanish RSS + Spanish Grok api_news)
--
-- Schema basis (verified against the repo, 2026-07-15):
--   * content_sources  — sql_for_tables/027_content_sources_table.sql
--       cols: site_id, source_type, name, entity_type, config jsonb,
--             fetch_interval, next_fetch_at, is_active, ...
--       dedup: UNIQUE (site_id, name)  → ON CONFLICT (site_id, name)
--   * site_specs       — actions/feed_news_recommendation_action.go
--       cols used: site_id, aspect, data jsonb, source, source_agent,
--                  is_current, created_by, superseded_at
--       versioning: supersede is_current=true row, insert new is_current=true
--   * sites onboarding — traffic_probe/intent_events_migration(1).sql (the
--                        documented deploy_config.engine UPDATE for relojistas)
--
-- IMPORTANT — the auto-seeder (SeedContentSourcesAction) SKIPS source_type='rss'
-- and 'scrape' ("requires manual URL config"); it only auto-creates news_search
-- (one per keyword) + one api_news named 'LLM News: <domain>'. So:
--   - the 5 RSS rows below are the ONLY way those feeds get in → insert them here;
--   - the Grok row below is pre-inserted under the canonical name
--     'LLM News: relojistas.com' with a SPANISH prompt, so the auto-seeder's
--     ON CONFLICT (site_id,name) DO NOTHING no-ops and OUR Spanish config wins.
-- ============================================================================

-- Pre-flight: confirm the site row exists (a rebuild presumes/creates one via the
-- normal site-creation path). If this returns 0 rows, create the site first.
--   SELECT id, domain, status, github_repo FROM sites WHERE domain = 'relojistas.com';

-- ----------------------------------------------------------------------------
-- P1a. Onboarding UPDATE — VM host + engine wiring (additive; safe to re-run)
--   Replace <INTERNAL_API_KEY> with the value in the box env
--   (/etc/site-engine/site-engine.env — read the FILE, never echo $KEY; debug #26).
--   Do NOT commit the real key; it is low-sensitivity (read-only capture-export)
--   but still a secret.
-- ----------------------------------------------------------------------------
UPDATE sites SET
    github_repo   = 'vm-sites',
    deploy_config = COALESCE(deploy_config, '{}'::jsonb) || jsonb_build_object(
        'target',       'vm',
        'capabilities', jsonb_build_array('backend'),
        'engine', jsonb_build_object(
            'base_url',  'https://relojistas.com',
            'stats_key', '<INTERNAL_API_KEY>'
        ))
WHERE domain = 'relojistas.com';

-- ----------------------------------------------------------------------------
-- P1b. Classification news_feed recommendation (site_specs, versioned merge).
--   Mirrors EvaluateNewsFeedAction's write, in SQL, and is safe whether or not a
--   classification spec already exists: it deep-merges content_features.news_feed
--   into the current spec (or starts from a minimal watch/news-portal spec).
--   Apply AFTER the build pipeline's own classification if you let it classify,
--   so this enrichment lands on top (evaluate_news_feed can also do this step).
-- ----------------------------------------------------------------------------
CREATE OR REPLACE FUNCTION relojistas_set_news_feed(p_site_id uuid)
RETURNS void AS $$
DECLARE
    v_cur   jsonb;
    v_base  jsonb;
    v_next  jsonb;
BEGIN
    SELECT data INTO v_cur
    FROM site_specs
    WHERE site_id = p_site_id AND aspect = 'classification' AND is_current = true;

    -- Minimal base spec if none exists yet (rebuild-from-scratch case).
    v_base := COALESCE(v_cur, jsonb_build_object(
        'industry',  'horology',
        'site_type', 'news_portal',
        'category',  'watches',
        'language',  'es'
    ));

    -- content_features.news_feed block (source_types drives the auto-seeder;
    -- rss is served by our explicit rows in P2, api_news/news_search auto-seed).
    v_next := jsonb_set(
        jsonb_set(v_base, '{content_features}',
            COALESCE(v_base->'content_features', '{}'::jsonb), true),
        '{content_features,news_feed}',
        jsonb_build_object(
            'recommended',       true,
            'reason',            'Watch/horology vertical with a live legacy RSS audience; reactivate the feed with current Spanish watch news.',
            'separate_page',     true,
            'source_types',      jsonb_build_array('rss', 'api_news', 'news_search'),
            'vertical_keywords', jsonb_build_array(
                'alta relojería', 'relojes de lujo', 'novedades relojes',
                'Rolex Omega Patek Philippe noticias', 'ferias relojería Watches and Wonders',
                'relojes vintage subastas', 'reparación de relojes'
            )
        ), true);

    -- Versioned write: supersede the current row, insert the merged one.
    UPDATE site_specs
       SET is_current = false, superseded_at = NOW()
     WHERE site_id = p_site_id AND aspect = 'classification' AND is_current = true;

    INSERT INTO site_specs (site_id, aspect, data, source, source_agent, is_current, created_by)
    VALUES (p_site_id, 'classification', v_next, 'enrichment', 'relojistas-rebuild', true, 'relojistas-rebuild');
END;
$$ LANGUAGE plpgsql;

-- ----------------------------------------------------------------------------
-- P2. content_sources seed — VERIFIED live Spanish feeds (checked 2026-07-15) +
--     canonical Spanish Grok api_news. Re-check feed liveness at apply time.
-- ----------------------------------------------------------------------------
CREATE OR REPLACE FUNCTION seed_relojistas_sources(p_site_id uuid)
RETURNS void AS $$
BEGIN
    -- 1. RSS — 5 verified live Spanish watch magazines (auto-seeder never adds rss).
    INSERT INTO content_sources (site_id, source_type, name, config, fetch_interval)
    VALUES
        (p_site_id, 'rss', 'Debajo del Reloj',
         '{"feed_url": "https://www.debajodelreloj.com/feed/", "max_items": 15}'::jsonb, '3 hours'),
        (p_site_id, 'rss', 'Tiempo de Relojes',
         '{"feed_url": "https://tiempoderelojes.com/feed/", "max_items": 15}'::jsonb, '3 hours'),
        (p_site_id, 'rss', 'TR Magazine',
         '{"feed_url": "https://trmagazine.es/feed/", "max_items": 15}'::jsonb, '4 hours'),
        (p_site_id, 'rss', 'Máquinas del Tiempo',
         '{"feed_url": "https://www.maquinasdeltiempo.com/feed/", "max_items": 15}'::jsonb, '4 hours'),
        (p_site_id, 'rss', 'Relojes y Estilo',
         '{"feed_url": "https://relojesyestilo.es/feed/", "max_items": 10}'::jsonb, '6 hours')
    ON CONFLICT (site_id, name) DO NOTHING;

    -- 2. api_news — pre-seed the CANONICAL name with a SPANISH prompt so the
    --    auto-seeder no-ops (its generic English row never overwrites this).
    INSERT INTO content_sources (site_id, source_type, name, config, fetch_interval)
    VALUES
        (p_site_id, 'api_news', 'LLM News: relojistas.com',
         jsonb_build_object(
            'provider',        'xai',
            'model',           'grok-4-1-fast',
            'prompt_template', 'Busca las noticias de relojería más importantes de las últimas {{.hours}} horas en español: novedades y lanzamientos de marcas, ferias (Watches and Wonders, Geneva), subastas, alta relojería, reparación y relojes vintage. Para cada noticia devuelve: title, summary (2-3 frases en español), source_url, source_name, published_at (ISO). Devuelve un array JSON.',
            'hours_lookback',  24,
            'max_items',       10,
            'keywords',        jsonb_build_array('alta relojería','relojes de lujo','novedades relojes','ferias relojería','subastas relojes','reparación de relojes'),
            'search_tools',    jsonb_build_array('web_search', 'x_search')
         ), '6 hours')
    ON CONFLICT (site_id, name) DO NOTHING;

    -- 3. Gemini as a SECOND api_news provider — LATER (operator request).
    --    BLOCKED until the feed-ingester's api_news provider routing supports
    --    Gemini (today it routes xai / openai / perplexity — verify before enabling).
    -- INSERT INTO content_sources (site_id, source_type, name, config, fetch_interval)
    -- VALUES (p_site_id, 'api_news', 'LLM News (Gemini): relojistas.com',
    --   '{"provider":"gemini","model":"gemini-...","prompt_template":"…(same ES prompt)…",
    --     "hours_lookback":24,"max_items":10,"search_tools":["web_search"]}'::jsonb, '6 hours')
    -- ON CONFLICT (site_id, name) DO NOTHING;

    -- news_search rows (one per vertical keyword) are created by the auto-seeder
    -- from the classification vertical_keywords — no need to insert them here.

    -- Make every new source due immediately for the first pull.
    UPDATE content_sources SET next_fetch_at = now()
     WHERE site_id = p_site_id AND next_fetch_at IS NULL;
END;
$$ LANGUAGE plpgsql;

-- ----------------------------------------------------------------------------
-- Apply (fill the site id once, from the pre-flight SELECT):
--   SELECT relojistas_set_news_feed('<site-uuid>');
--   SELECT seed_relojistas_sources('<site-uuid>');
-- Verify:
--   SELECT source_type, name, config->>'feed_url' AS feed, next_fetch_at
--     FROM content_sources WHERE site_id = '<site-uuid>' ORDER BY source_type, name;
--   SELECT data->'content_features'->'news_feed'
--     FROM site_specs WHERE site_id='<site-uuid>' AND aspect='classification' AND is_current;
-- After the first orchestrator pass (or the 6h heartbeat):
--   SELECT status, count(*) FROM content_feed_items WHERE site_id='<site-uuid>' GROUP BY status;
--   -- expect ingested>0, then relevant>0 after the next triage pass (two-pass by design).
-- ============================================================================
