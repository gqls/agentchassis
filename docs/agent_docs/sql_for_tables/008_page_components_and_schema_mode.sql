-- ===========================================================================
-- MIGRATION: Schema Mode Infrastructure for Flexible/Strict Rendering
-- File: 043_schema_mode_infrastructure.sql
-- ===========================================================================
-- Adds support for:
--   - schema_mode on sites (flexible/strict default behavior)
--   - schema_snapshot on page_components (locks schema at approval)
--   - content_snapshot on page_components (stores approved content)
--   - component_version tracking
-- ===========================================================================

BEGIN;

-- ===========================================================================
-- PART 1: SITES TABLE - Default schema mode for new sections
-- ===========================================================================

ALTER TABLE sites ADD COLUMN IF NOT EXISTS
    schema_mode TEXT DEFAULT 'flexible';
COMMENT ON COLUMN sites.schema_mode IS
    'Default rendering mode for new sections: flexible (best-effort, warn on missing) or strict (fail on schema mismatch)';

-- When to transition to strict mode
ALTER TABLE sites ADD COLUMN IF NOT EXISTS
    strict_mode_trigger TEXT DEFAULT 'first_deploy';
COMMENT ON COLUMN sites.strict_mode_trigger IS
    'When to lock sections to strict mode: hitl (on human approval), first_deploy (on first successful deploy), manual (never auto-transition)';

-- ===========================================================================
-- PART 2: PAGE_COMPONENTS TABLE - Per-section schema tracking
-- ===========================================================================

-- The locked schema for this section (set at approval time)
ALTER TABLE page_components ADD COLUMN IF NOT EXISTS
    schema_snapshot JSONB;
COMMENT ON COLUMN page_components.schema_snapshot IS
    'Locked input_schema from component at approval time. Edits must match this schema in strict mode.';

-- The content values that were approved
ALTER TABLE page_components ADD COLUMN IF NOT EXISTS
    content_snapshot JSONB;
COMMENT ON COLUMN page_components.content_snapshot IS
    'The actual content values used when approved. Used for edit comparison, rollback, and form pre-population.';

-- Which component version this was built with
ALTER TABLE page_components ADD COLUMN IF NOT EXISTS
    component_version_id UUID;
COMMENT ON COLUMN page_components.component_version_id IS
    'Reference to specific component version (if versioning enabled). Ensures template consistency.';

-- Per-section schema mode (overrides site default)
ALTER TABLE page_components ADD COLUMN IF NOT EXISTS
    schema_mode TEXT;
COMMENT ON COLUMN page_components.schema_mode IS
    'Section-specific schema mode. NULL = inherit from site. Set to flexible/strict to override.';

-- When the section was locked to strict mode
ALTER TABLE page_components ADD COLUMN IF NOT EXISTS
    locked_at TIMESTAMPTZ;
COMMENT ON COLUMN page_components.locked_at IS
    'Timestamp when section was locked to strict mode';

-- Who/what locked it
ALTER TABLE page_components ADD COLUMN IF NOT EXISTS
    locked_by TEXT;
COMMENT ON COLUMN page_components.locked_by IS
    'What triggered strict mode lock: hitl, auto_eval, first_deploy, manual';

-- ===========================================================================
-- PART 3: COMPONENT VERSIONING TABLE
-- ===========================================================================

CREATE TABLE IF NOT EXISTS component_versions (
                                                  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),

    -- Which component this is a version of
    component_id UUID NOT NULL REFERENCES content_components(id) ON DELETE CASCADE,

    -- Version number (increments on each change)
    version_number INTEGER NOT NULL DEFAULT 1,

    -- Snapshot of the component at this version
    html_template TEXT NOT NULL,
    css_template TEXT,
    input_schema JSONB,

    -- Change tracking
    change_description TEXT,
    changed_by TEXT,

    -- Timestamps
    created_at TIMESTAMPTZ DEFAULT now(),

    -- Unique constraint
    UNIQUE(component_id, version_number)
    );

COMMENT ON TABLE component_versions IS
    'Versioned snapshots of component templates. Allows strict mode pages to use specific versions.';

-- Index for lookups
CREATE INDEX IF NOT EXISTS idx_component_versions_component
    ON component_versions(component_id, version_number DESC);

-- ===========================================================================
-- PART 4: FUNCTION TO LOCK SECTION TO STRICT MODE
-- ===========================================================================

CREATE OR REPLACE FUNCTION lock_section_to_strict(
    p_page_component_id UUID,
    p_content_data JSONB,
    p_locked_by TEXT DEFAULT 'manual'
) RETURNS VOID AS $$
DECLARE
v_component_id UUID;
    v_input_schema JSONB;
BEGIN
    -- Get the component's current input_schema
SELECT pc.component_id, cc.input_schema
INTO v_component_id, v_input_schema
FROM page_components pc
         JOIN content_components cc ON pc.component_id = cc.id
WHERE pc.id = p_page_component_id;

IF NOT FOUND THEN
        RAISE EXCEPTION 'Page component not found: %', p_page_component_id;
END IF;

    -- Lock the section
UPDATE page_components SET
                           schema_mode = 'strict',
                           schema_snapshot = v_input_schema,
                           content_snapshot = p_content_data,
                           locked_at = now(),
                           locked_by = p_locked_by
WHERE id = p_page_component_id;
END;
$$ LANGUAGE plpgsql;

COMMENT ON FUNCTION lock_section_to_strict IS
    'Locks a page section to strict schema mode, capturing the current schema and content.';

-- ===========================================================================
-- PART 5: FUNCTION TO UNLOCK SECTION (for redesign)
-- ===========================================================================

CREATE OR REPLACE FUNCTION unlock_section_for_redesign(
    p_page_component_id UUID,
    p_preserve_content BOOLEAN DEFAULT true
) RETURNS VOID AS $$
BEGIN
UPDATE page_components SET
                           schema_mode = 'flexible',
                           -- Optionally preserve snapshots for reference
                           schema_snapshot = CASE WHEN p_preserve_content THEN schema_snapshot ELSE NULL END,
                           content_snapshot = CASE WHEN p_preserve_content THEN content_snapshot ELSE NULL END,
                           locked_at = NULL,
                           locked_by = NULL
WHERE id = p_page_component_id;
END;
$$ LANGUAGE plpgsql;

COMMENT ON FUNCTION unlock_section_for_redesign IS
    'Unlocks a section from strict mode, allowing flexible rendering during redesign.';

-- ===========================================================================
-- PART 6: VIEW FOR SECTION SCHEMA STATUS
-- ===========================================================================

CREATE OR REPLACE VIEW v_section_schema_status AS
SELECT
    pc.id AS page_component_id,
    pc.page_id,
    p.name AS page_name,
    s.domain,
    cc.name AS component_name,
    cc.function AS component_function,
    COALESCE(pc.schema_mode, s.schema_mode, 'flexible') AS effective_schema_mode,
    pc.schema_snapshot IS NOT NULL AS has_schema_snapshot,
    pc.content_snapshot IS NOT NULL AS has_content_snapshot,
    pc.locked_at,
    pc.locked_by,
    pc.build_status,
    pc.reviewed_at
FROM page_components pc
         JOIN pages p ON pc.page_id = p.id
         JOIN sites s ON p.site_id = s.id
         LEFT JOIN content_components cc ON pc.component_id = cc.id;

COMMENT ON VIEW v_section_schema_status IS
    'Shows effective schema mode and lock status for all page sections.';

-- ===========================================================================
-- PART 7: TRIGGER TO AUTO-LOCK ON FIRST DEPLOY (if site configured)
-- ===========================================================================

CREATE OR REPLACE FUNCTION auto_lock_on_deploy() RETURNS TRIGGER AS $$
BEGIN
    -- Only act if changing to 'deployed' status
    IF NEW.build_status = 'deployed' AND OLD.build_status != 'deployed' THEN
        -- Check if site is configured for first_deploy locking
        IF EXISTS (
            SELECT 1 FROM pages p
            JOIN sites s ON p.site_id = s.id
            WHERE p.id = NEW.page_id
            AND s.strict_mode_trigger = 'first_deploy'
        ) THEN
            -- Lock to strict mode if not already locked
            IF NEW.schema_mode IS NULL OR NEW.schema_mode = 'flexible' THEN
                NEW.schema_mode := 'strict';
                NEW.locked_at := now();
                NEW.locked_by := 'first_deploy';
                -- Note: schema_snapshot and content_snapshot should be set before deploy
END IF;
END IF;
END IF;

RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Create trigger if it doesn't exist
DROP TRIGGER IF EXISTS trigger_auto_lock_on_deploy ON page_components;
CREATE TRIGGER trigger_auto_lock_on_deploy
    BEFORE UPDATE ON page_components
    FOR EACH ROW
    EXECUTE FUNCTION auto_lock_on_deploy();

-- ===========================================================================
-- PART 8: INDEXES
-- ===========================================================================

CREATE INDEX IF NOT EXISTS idx_page_components_schema_mode
    ON page_components(schema_mode) WHERE schema_mode IS NOT NULL;

CREATE INDEX IF NOT EXISTS idx_page_components_locked
    ON page_components(locked_at) WHERE locked_at IS NOT NULL;

CREATE INDEX IF NOT EXISTS idx_sites_schema_mode
    ON sites(schema_mode);

COMMIT;

-- ===========================================================================
-- VERIFICATION QUERIES
-- ===========================================================================

-- Check new columns on sites
-- SELECT column_name, data_type, column_default
-- FROM information_schema.columns
-- WHERE table_name = 'sites' AND column_name IN ('schema_mode', 'strict_mode_trigger');

-- Check new columns on page_components
-- SELECT column_name, data_type
-- FROM information_schema.columns
-- WHERE table_name = 'page_components'
-- AND column_name IN ('schema_mode', 'schema_snapshot', 'content_snapshot', 'locked_at', 'locked_by');

-- Check component_versions table
-- SELECT * FROM component_versions LIMIT 5;

-- Check section schema status view
-- SELECT * FROM v_section_schema_status LIMIT 10;



UPDATE pages
SET nav_label = 'Case Studies'
WHERE name = 'case-studies'
  AND site_id = (SELECT id FROM sites WHERE domain = 'leopardessconsulting.co.uk');

---

-- Content brief: records the instructions that generated each component's content.
-- Enables admins to see, edit, and regenerate content with modified instructions.
ALTER TABLE page_components ADD COLUMN IF NOT EXISTS content_brief JSONB;

---

-- blog pages
-- 1. Delete empty shell page_components (featured_article, category_section, article_grid, ad_zone_inline)
DELETE FROM page_components
WHERE page_id = 'ff56bcaf-cf3c-40bd-a6ee-18703bd3d656'
  AND slot_name IN ('featured-article', 'category-section', 'article-grid', 'ad-zone-inline');

-- 2. Update hero to be blog-specific
UPDATE page_components
SET rendered_html = '<section class="hero" data-component="hero" style="background: linear-gradient(135deg, var(--primary-color, #1a1a2e) 0%, var(--secondary-color, #16213e) 50%, var(--accent-color, #0f3460) 100%);">
    <div class="hero-content">
        <h1>Engineering Blog</h1>
        <p class="hero-subheadline">Deep dives into building, deploying, and operating multi-agent systems in production — from architecture decisions to the things that actually break.</p>
    </div>
</section>
<style>
.hero {
    min-height: 40vh;
    display: flex;
    align-items: center;
    justify-content: center;
    text-align: center;
    padding: 3rem 2rem;
    position: relative;
    --section-text: rgba(255,255,255,0.95);
    --section-text-muted: rgba(255,255,255,0.8);
    --section-heading: #ffffff;
}
.hero-content {
    max-width: 800px;
    margin: 0 auto;
    color: #fff;
    z-index: 1;
}
.hero h1 {
    font-size: clamp(2rem, 5vw, 3rem);
    font-weight: 700;
    margin-bottom: 1rem;
    line-height: 1.2;
    text-shadow: 0 2px 4px rgba(0,0,0,0.3);
}
.hero-subheadline {
    font-size: clamp(1rem, 2vw, 1.25rem);
    line-height: 1.6;
    opacity: 0.9;
}
@media (max-width: 768px) {
    .hero { min-height: 30vh; padding: 2rem 1.5rem; }
}
</style>'
WHERE id = '4a3e0db6-b06c-422a-8d5d-8699b1778194';

-- 3. Insert blog-listing component at position 3
INSERT INTO page_components (page_id, slot_name, position, rendered_html, build_status)
VALUES ('ff56bcaf-cf3c-40bd-a6ee-18703bd3d656', 'blog-listing', 3,
        '<section class="blog-listing" data-component="blog-listing">
            <div class="blog-container">
                <div class="blog-grid">

                    <a href="/blog/the-enterprise-ai-agent-adoption-gap-2025.html" class="blog-card">
                        <span class="blog-card__tag">Strategy</span>
                        <h3>The Enterprise AI Agent Adoption Gap</h3>
                        <p>Why pilots succeed and production deployments stall — and what engineering teams can do about it.</p>
                    </a>

                    <a href="/blog/orchestrating-ai-agents-in-production-what-actually-breaks.html" class="blog-card">
                        <span class="blog-card__tag">Architecture</span>
                        <h3>Orchestrating AI Agents in Production: What Actually Breaks</h3>
                        <p>Timeout handling, state recovery, cascading failures — the problems you hit after the demo works.</p>
                    </a>

                    <a href="/blog/building-a-hierarchical-agent-system-with-kafka-and-postgres.html" class="blog-card">
                        <span class="blog-card__tag">Tutorial</span>
                        <h3>Building a Hierarchical Agent System with Kafka and Postgres</h3>
                        <p>A practical walkthrough of the message-driven architecture behind multi-agent coordination.</p>
                    </a>

                    <a href="/blog/deploying-ai-agents-kubernetes-practical-guide.html" class="blog-card">
                        <span class="blog-card__tag">DevOps</span>
                        <h3>Deploying AI Agents on Kubernetes</h3>
                        <p>Configuration patterns, resource management, and health checks for agent workloads on K8s.</p>
                    </a>

                    <a href="/blog/multi-agent-state-management-distributed-systems.html" class="blog-card">
                        <span class="blog-card__tag">Architecture</span>
                        <h3>State Management for Multi-Agent Systems</h3>
                        <p>Patterns that hold up in production — orchestration state, checkpoints, and recovery strategies.</p>
                    </a>

                    <a href="/blog/why-most-ai-agent-frameworks-fail-at-the-orchestration-layer.html" class="blog-card">
                        <span class="blog-card__tag">Analysis</span>
                        <h3>Why Most AI Agent Frameworks Fail at the Orchestration Layer</h3>
                        <p>The gap between single-agent toolkits and production multi-agent systems.</p>
                    </a>

                    <a href="/blog/llm-provider-abstraction-production-agent-systems.html" class="blog-card">
                        <span class="blog-card__tag">Engineering</span>
                        <h3>Why You Should Abstract Your LLM Provider from Day One</h3>
                        <p>Provider lock-in, fallback strategies, and the abstraction layer that saves you later.</p>
                    </a>

                    <a href="/blog/ai-agent-observability-2025-what-teams-are-actually-monitoring.html" class="blog-card">
                        <span class="blog-card__tag">Observability</span>
                        <h3>AI Agent Observability in 2025</h3>
                        <p>What engineering teams are actually monitoring — token spend, latency chains, and failure attribution.</p>
                    </a>

                </div>
            </div>
        </section>
        <style>
        .blog-listing {
            padding: 4rem 2rem;
            background: var(--background-color, #0a0a1a);
        }
        .blog-container {
            max-width: 1200px;
            margin: 0 auto;
        }
        .blog-grid {
            display: grid;
            grid-template-columns: repeat(auto-fill, minmax(340px, 1fr));
            gap: 2rem;
        }
        .blog-card {
            display: block;
            background: rgba(255,255,255,0.04);
            border: 1px solid rgba(255,255,255,0.08);
            border-radius: 8px;
            padding: 2rem;
            text-decoration: none;
            color: rgba(255,255,255,0.9);
            transition: all 0.2s ease;
        }
        .blog-card:hover {
            background: rgba(255,255,255,0.08);
            border-color: rgba(255,255,255,0.15);
            transform: translateY(-2px);
        }
        .blog-card__tag {
            display: inline-block;
            font-size: 0.75rem;
            font-weight: 600;
            text-transform: uppercase;
            letter-spacing: 0.05em;
            color: var(--accent-color, #4fc3f7);
            margin-bottom: 0.75rem;
        }
        .blog-card h3 {
            font-size: 1.25rem;
            font-weight: 600;
            line-height: 1.3;
            margin-bottom: 0.75rem;
            color: #fff;
        }
        .blog-card p {
            font-size: 0.95rem;
            line-height: 1.6;
            color: rgba(255,255,255,0.65);
            margin: 0;
        }
        @media (max-width: 768px) {
            .blog-listing { padding: 2rem 1rem; }
            .blog-grid { grid-template-columns: 1fr; gap: 1.5rem; }
        }
        </style>', 'deployed');

-- 4. Fix call-to-action position (was 7, now should be 4)
UPDATE page_components
SET position = 4
WHERE id = 'd5f73b45-f068-4292-bb68-d3906fa9705c';

-- 5. Update page record
UPDATE pages
SET page_type = 'blog-index',
    sections = '["hero", "blog-listing", "call-to-action"]'::jsonb
WHERE id = 'ff56bcaf-cf3c-40bd-a6ee-18703bd3d656';

---
-- news feed and titles js
UPDATE page_components
SET rendered_html = '<!-- latest-news component -->
<section data-component="latest-news" class="latest-news-section section-padding">
  <div class="container">
    <h2 class="section-heading" id="news-headline">Energy Market News</h2>
    <p class="section-subheadline" id="news-subheadline">Latest developments in wholesale gas and energy markets</p>
    <div class="news-grid" id="news-container">
      <noscript>
        <p class="news-empty">Enable JavaScript to see the latest news.</p>
      </noscript>
    </div>
    <div id="news-footer"></div>
  </div>
</section>
<script>
(function() {
  fetch("/data/latest-news.json")
    .then(function(r) { return r.json(); })
    .then(function(data) {
      if (data.headline)
        document.getElementById("news-headline").textContent = data.headline;
      var sub = document.getElementById("news-subheadline");
      if (sub && data.subheadline)
        sub.textContent = data.subheadline;
      var container = document.getElementById("news-container");
      if (data.items && data.items.length > 0) {
        container.innerHTML = data.items.map(function(item) {
          var html = "<article class=\"news-card\"><div class=\"news-card-content\">";
          html += "<h3 class=\"news-card-title\"><a href=\"" + item.url + "\" target=\"_blank\" rel=\"noopener noreferrer\">" + item.title + "</a></h3>";
          if (item.summary) {
            html += "<p class=\"news-card-summary\">" + item.summary + "</p>";
          }
          html += "<div class=\"news-card-meta\">";
          if (item.source) {
            html += "<span class=\"news-source\">" + item.source + "</span>";
          }
          if (item.date) {
            html += "<time class=\"news-date\">" + item.date + "</time>";
          }
          html += "</div></div></article>";
          return html;
        }).join("");
      }
      if (data.insights_url) {
        document.getElementById("news-footer").innerHTML =
          "<div class=\"news-section-footer\"><a href=\"" + data.insights_url +
          "\" class=\"news-more-link\">" + (data.insights_label || "More insights &rarr;") +
          "</a></div>";
      }
    })
    .catch(function() {});
})();
</script>',
    updated_at = NOW()
WHERE id = (
    SELECT pc.id FROM page_components pc
                          JOIN pages p ON p.id = pc.page_id
    WHERE p.site_id = '5fe15466-4e2e-4ff2-981e-98c1b7074002'
      AND p.name = 'index'
      AND pc.slot_name = 'latest-news'
    LIMIT 1
    );