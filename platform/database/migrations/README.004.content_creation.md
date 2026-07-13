CREATE TABLE IF NOT EXISTS content_components (
id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
name TEXT NOT NULL,
description TEXT,
html_template TEXT NOT NULL,
input_schema JSONB,
"function" TEXT NOT NULL DEFAULT 'generic-text-block',
created_at TIMESTAMPTZ DEFAULT NOW(),
updated_at TIMESTAMPTZ DEFAULT NOW()
);

-- ---
-- Create the index for our new "function" column for fast lookups
-- ---
CREATE INDEX IF NOT EXISTS idx_content_components_function ON content_components ("function");


-- ---
-- Now, let's add our "Base Fallback" components for the MVP
We've added our "functional" tags (e.g., problem_statement).

clearly labeled bits is implemented. The html_template uses Go template placeholders (e.g., {{.title}}), and the input_schema tells the Content Creator agent exactly what content it needs to generate.

-- ---

-- 1. The 'generic' fallback component (Priority 3)
INSERT INTO content_components (name, description, html_template, input_schema, "function")
VALUES (
'Generic Text Block',
'A simple, generic text block. Used as a fallback.',
'<div class="generic-block" data-function="{{.Function}}" data-component-id="{{.ComponentID}}"><h3>{{.title}}</h3><p>{{.body}}</p></div>',
'{"title": "string", "body": "string"}'::jsonb,
'generic-text-block'
);

-- 2. The 'PAS' model components (Priority 1)
INSERT INTO content_components (name, description, html_template, input_schema, "function")
VALUES (
'Problem Headline',
'A component for stating the PAS "Problem".',
'<div class="problem-section" data-function="problem_statement" data-component-id="{{.ComponentID}}"><h1 class="problem-headline">{{.headline}}</h1></div>',
'{"headline": "string"}'::jsonb,
'problem_statement'
),
(
'Agitation Block',
'A component for the PAS "Agitate" step.',
'<div class="agitate-section" data-function="agitation" data-component-id="{{.ComponentID}}"><h2>{{.subheading}}</h2><ul class="agitation-list"><li>{{.point1}}</li><li>{{.point2}}</li></ul></div>',
'{"subheading": "string", "point1": "string", "point2": "string"}'::jsonb,
'agitation'
),
(
'Solution Block (CTA)',
'A component for the PAS "Solution" step with a CTA.',
'<div class="solution-section" data-function="solution_provider" data-component-id="{{.ComponentID}}"><h2>{{.heading}}</h2><p>{{.description}}</p><a href="{{.cta_url}}" class="cta-button">{{.cta_text}}</a></div>',
'{"heading": "string", "description": "string", "cta_url": "string", "cta_text": "string"}'::jsonb,
'solution_provider'
);