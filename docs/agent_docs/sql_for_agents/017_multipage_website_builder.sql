UPDATE agent_definitions
SET default_config = jsonb_set(
        default_config,
        '{workflow,steps,build_pages_loop,next_step}',
        '"trigger_site_deploy"'
                     ),
    updated_at = NOW()
WHERE agent_type = 'multipage-website-builder';

--

- Update hero template to use background image (even though we're not using it')
-- ============================================================

UPDATE content_components
SET html_template = '<section class="hero" data-component="hero"{{if .hero_home_url}} style="background: linear-gradient(135deg, rgba(26,26,46,0.8) 0%, rgba(22,33,62,0.75) 50%, rgba(15,52,96,0.7) 100%), url(''{{.hero_home_url}}'') center/cover no-repeat;"{{end}}>
        <div class="hero-content">
            <h1>{{.headline}}</h1>
            <p class="hero-subheadline">{{.subheadline}}</p>
            {{if .primary_cta_text}}<a href="{{.primary_cta_url | default "/contact.html"}}" class="btn btn-primary">{{.primary_cta_text}}</a>{{end}}
            {{if .secondary_cta_text}}<a href="{{.secondary_cta_url | default "/services.html"}}" class="btn btn-secondary">{{.secondary_cta_text}}</a>{{end}}
        </div>
    </section>',
    updated_at = NOW()
WHERE function = 'hero' AND (category IS NULL OR category = '');

-- Verify
SELECT name, function,
       CASE WHEN html_template LIKE '%hero_home_url%' THEN 'YES' ELSE 'NO' END as has_image_support
FROM content_components WHERE function = 'hero';

-- nav changes
-- =========================================================================
-- 2. multipage-website-builder: change sync_pages_to_db.next_step and add populate_nav
-- =========================================================================

-- Change sync_pages_to_db next_step from generate_pages_loop to populate_nav
UPDATE agent_definitions
SET workflow = jsonb_set(
        workflow,
        '{workflow,steps,sync_pages_to_db,next_step}',
        '"populate_nav"'
               )
WHERE type = 'multipage-website-builder' AND is_active = true;

-- Add the populate_nav step
UPDATE agent_definitions
SET workflow = jsonb_set(
        workflow,
        '{workflow,steps,populate_nav}',
        '{
            "action": "populate_nav_tables",
            "config": {
                "site_id_field": "site_record.site_id",
                "max_header_items": 6
            },
            "next_step": "generate_pages_loop",
            "description": "Populate navigation tables from page plan",
            "output_field": "nav_data"
        }'::jsonb
               )
WHERE type = 'multipage-website-builder' AND is_active = true;

-- Update max_header_items from 6 to 8 in deployed workflows
-- Run this on the live DB since the original 002 SQL already ran with 6.

UPDATE agent_definitions
SET default_config = jsonb_set(
        default_config,
        '{workflow,steps,populate_nav,config,max_header_items}',
        '8'
                     )
WHERE type = 'multipage-website-builder' AND is_active = true;