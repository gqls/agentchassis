#!/usr/bin/env python3
"""Generate register.sql from template.html + input_schema.json so the SQL can
never drift from the version-controlled source it is supposed to install."""
import json
import pathlib
import sys

comp = pathlib.Path(sys.argv[1])
html = (comp / "template.html").read_text()
schema = json.dumps(json.loads((comp / "input_schema.json").read_text()), indent=2)

for tag in ("$HTML$", "$SCHEMA$"):
    if tag in html or tag in schema:
        sys.exit(f"refusing to generate: source contains the dollar-quote tag {tag}")

sql = f"""\\set ON_ERROR_STOP on
-- evidence-chart — shared, evidence-sourced chart component.
-- GENERATED from components/evidence-chart/{{template.html,input_schema.json}}
-- by scripts/gen_component_register_sql.py. Edit those files and regenerate;
-- do not hand-edit this file.
--
-- The guarantee this component exists to make: the ONLY place a figure lives is
-- site_specs.evidence_base.facts. A chart definition names fact ids and never
-- restates a value, and both fields are system-resolved — resolved data beats
-- LLM content at render time, so the writer cannot supply a number even if it
-- tries. No evidence_base charts => the section is skipped, not invented.
BEGIN;
INSERT INTO content_components
  (id, name, function, display_name, description, category, semantic_tags,
   section_type, component_level, render_mode, is_dark_section, is_active,
   suitable_site_types, suitable_page_types, html_template, input_schema)
VALUES (
  gen_random_uuid(),
  'evidence-chart','evidence-chart','Evidence Chart',
  'Code-rendered bar charts whose values come from the site''s evidence_base register, never from the model. Chart definitions name fact ids; the figures, units and verified dates are read from the audited fact rows. Bars are drawn in CSS from the real value; the label and figure are real selectable text, so screen readers and the claims gate both see the number. Skipped entirely on a site with no audited series.',
  'data','["chart","data","evidence","code-rendered","brochure","infographic"]'::jsonb,
  'evidence-chart','section','agent',false,true,
  '["brochure","consultancy","professional-services","b2b"]'::jsonb,
  '["index","home","about","capabilities","landing","content"]'::jsonb,
  $HTML${html}$HTML$,
  $SCHEMA${schema}$SCHEMA$::jsonb
);
COMMIT;

SELECT function, section_type, component_level, is_active,
       length(html_template) AS template_bytes,
       jsonb_object_keys(input_schema->'fields') AS field
  FROM content_components WHERE function = 'evidence-chart';
"""

(comp / "register.sql").write_text(sql)
print(f"wrote {comp / 'register.sql'} ({len(sql)} bytes)")
