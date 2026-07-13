ALTER TABLE agent_definitions
ADD CONSTRAINT agent_definitions_category_check
CHECK (category IN ('data-driven', 'code-driven', 'adapter', 'data-collection'));

    "agent_definitions_category_check" CHECK (category::text = ANY (ARRAY['data-driven'::character varying, 'code-driven'::character varying, 'adapter'::character varying]::text[]))

ALTER TABLE agent_definitions
DROP CONSTRAINT agent_definitions_category_check;
