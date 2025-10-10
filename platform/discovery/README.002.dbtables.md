public | agent_capabilities      | table             | clients_user | permanent   | heap          | 16 kB      |
public | agent_default_configs   | table             | clients_user | permanent   | heap          | 16 kB      |
public | agent_definitions       | table             | clients_user | permanent   | heap          | 208 kB     | Core agent definitions including content-creator with memory and generic orchestrator - updated by migration 080
public | agent_dependencies      | table             | clients_user | permanent   | heap          | 16 kB      |
public | agent_group_definitions | table             | clients_user | permanent   | heap          | 16 kB      |
public | agent_group_members     | table             | clients_user | permanent   | heap          | 16 kB      |
public | agent_groups            | table             | clients_user | permanent   | heap          | 8192 bytes |
public | agent_metrics           | table             | clients_user | permanent   | heap          | 16 kB      |
public | agent_metrics_config    | table             | clients_user | permanent   | heap          | 16 kB      |
public | event_statistics        | materialized view | clients_user | permanent   | heap          | 0 bytes    |
public | orchestration_requests  | table             | clients_user | permanent   | heap          | 0 bytes    |
public | orchestration_states    | table             | clients_user | permanent   | heap          | 200 kB     |
public | pending_requests        | table             | clients_user | permanent   | heap          | 0 bytes    |
public | processed_messages      | table             | clients_user | permanent   | heap          | 16 kB      |
public | recent_agent_events     | view              | clients_user | permanent   |               | 0 bytes    |
public | system_events           | table             | clients_user | permanent   | heap          | 544 kB     | Audit log for all system events including agent bootstrapping, workflows, and API calls
public | workflow_templates      | table             | clients_user | permanent   | heap          | 16 kB      |


--


Table "public.agent_capabilities"
Column   |           Type           | Collation | Nullable |      Default      | Storage  | Compression | Stats target | Description
------------+--------------------------+-----------+----------+-------------------+----------+-------------+--------------+-------------
id         | uuid                     |           | not null | gen_random_uuid() | plain    |             |              |
agent_type | character varying(100)   |           | not null |                   | extended |             |              |
capability | character varying(255)   |           | not null |                   | extended |             |              |
strength   | numeric(3,2)             |           |          | 1.0               | main     |             |              |
metadata   | jsonb                    |           |          | '{}'::jsonb       | extended |             |              |
created_at | timestamp with time zone |           |          | now()             | plain    |             |              |
updated_at | timestamp with time zone |           |          | now()             | plain    |             |              |
Indexes:
"agent_capabilities_pkey" PRIMARY KEY, btree (id)
"agent_capabilities_agent_type_capability_key" UNIQUE CONSTRAINT, btree (agent_type, capability)
"idx_agent_capabilities_agent" btree (agent_type)
Check constraints:
"agent_capabilities_strength_check" CHECK (strength >= 0::numeric AND strength <= 1::numeric)
Foreign-key constraints:
"agent_capabilities_agent_type_fkey" FOREIGN KEY (agent_type) REFERENCES agent_definitions(type)
Access method: heap

                                                         Table "public.agent_default_configs"
Column    |           Type           | Collation | Nullable |             Default             | Storage  | Compression | Stats target | Description
-------------+--------------------------+-----------+----------+---------------------------------+----------+-------------+--------------+-------------
id          | uuid                     |           | not null | gen_random_uuid()               | plain    |             |              |
config_name | character varying(255)   |           | not null |                                 | extended |             |              |
agent_type  | character varying(100)   |           | not null |                                 | extended |             |              |
environment | character varying(50)    |           |          | 'production'::character varying | extended |             |              |
config      | jsonb                    |           | not null |                                 | extended |             |              |
created_at  | timestamp with time zone |           |          | now()                           | plain    |             |              |
updated_at  | timestamp with time zone |           |          | now()                           | plain    |             |              |
Indexes:
"agent_default_configs_pkey" PRIMARY KEY, btree (id)
"agent_default_configs_config_name_key" UNIQUE CONSTRAINT, btree (config_name)
"idx_agent_default_configs_agent" btree (agent_type)
"idx_agent_default_configs_name" btree (config_name)
Foreign-key constraints:
"agent_default_configs_agent_type_fkey" FOREIGN KEY (agent_type) REFERENCES agent_definitions(type)
Access method: heap

                                                      Table "public.agent_dependencies"
     Column      |           Type           | Collation | Nullable |      Default      | Storage  | Compression | Stats target | Description 
-----------------+--------------------------+-----------+----------+-------------------+----------+-------------+--------------+-------------
id              | uuid                     |           | not null | gen_random_uuid() | plain    |             |              |
agent_type      | character varying(100)   |           | not null |                   | extended |             |              |
depends_on      | character varying(100)   |           | not null |                   | extended |             |              |
dependency_type | character varying(50)    |           | not null |                   | extended |             |              |
config          | jsonb                    |           |          | '{}'::jsonb       | extended |             |              |
created_at      | timestamp with time zone |           |          | now()             | plain    |             |              |
updated_at      | timestamp with time zone |           |          | now()             | plain    |             |              |
Indexes:
"agent_dependencies_pkey" PRIMARY KEY, btree (id)
"agent_dependencies_agent_type_depends_on_key" UNIQUE CONSTRAINT, btree (agent_type, depends_on)
"idx_agent_dependencies_agent" btree (agent_type)
Check constraints:
"agent_dependencies_dependency_type_check" CHECK (dependency_type::text = ANY (ARRAY['data'::character varying, 'optional'::character varying, 'orchestration'::character varying]::text[]))
Foreign-key constraints:
"agent_dependencies_agent_type_fkey" FOREIGN KEY (agent_type) REFERENCES agent_definitions(type)
Access method: heap

                                                        Table "public.agent_group_definitions"
         Column         |            Type             | Collation | Nullable |      Default      | Storage  | Compression | Stats target | Description 
------------------------+-----------------------------+-----------+----------+-------------------+----------+-------------+--------------+-------------
id                     | uuid                        |           | not null | gen_random_uuid() | plain    |             |              |
name                   | character varying(255)      |           | not null |                   | extended |             |              |
group_type             | character varying(100)      |           | not null |                   | extended |             |              |
agent_configs          | jsonb                       |           | not null |                   | extended |             |              |
orchestration_workflow | jsonb                       |           |          |                   | extended |             |              |
usage_count            | integer                     |           |          | 0                 | plain    |             |              |
version                | integer                     |           |          | 1                 | plain    |             |              |
created_at             | timestamp without time zone |           |          | now()             | plain    |             |              |
updated_at             | timestamp without time zone |           |          | now()             | plain    |             |              |
Indexes:
"agent_group_definitions_pkey" PRIMARY KEY, btree (id)
Access method: heap

                                                   Table "public.agent_group_members"
Column   |           Type           | Collation | Nullable |      Default      | Storage  | Compression | Stats target | Description
------------+--------------------------+-----------+----------+-------------------+----------+-------------+--------------+-------------
id         | uuid                     |           | not null | gen_random_uuid() | plain    |             |              |
group_name | character varying(255)   |           | not null |                   | extended |             |              |
agent_type | character varying(100)   |           | not null |                   | extended |             |              |
role       | character varying(100)   |           | not null |                   | extended |             |              |
required   | boolean                  |           |          | true              | plain    |             |              |
config     | jsonb                    |           |          | '{}'::jsonb       | extended |             |              |
created_at | timestamp with time zone |           |          | now()             | plain    |             |              |
updated_at | timestamp with time zone |           |          | now()             | plain    |             |              |
Indexes:
"agent_group_members_pkey" PRIMARY KEY, btree (id)
"agent_group_members_group_name_agent_type_key" UNIQUE CONSTRAINT, btree (group_name, agent_type)
"idx_agent_group_members_agent" btree (agent_type)
"idx_agent_group_members_group" btree (group_name)
Foreign-key constraints:
"agent_group_members_agent_type_fkey" FOREIGN KEY (agent_type) REFERENCES agent_definitions(type)
Access method: heap

                                                     Table "public.agent_groups"
     Column     |            Type             | Collation | Nullable | Default | Storage  | Compression | Stats target | Description 
----------------+-----------------------------+-----------+----------+---------+----------+-------------+--------------+-------------
group_id       | uuid                        |           | not null |         | plain    |             |              |
parent_orch_id | uuid                        |           | not null |         | plain    |             |              |
group_type     | character varying(100)      |           |          |         | extended |             |              |
members        | jsonb                       |           | not null |         | extended |             |              |
created_at     | timestamp without time zone |           |          | now()   | plain    |             |              |
Indexes:
"agent_groups_pkey" PRIMARY KEY, btree (group_id)
Foreign-key constraints:
"fk_agent_groups_parent" FOREIGN KEY (parent_orch_id) REFERENCES orchestration_states(orchestration_id) ON DELETE CASCADE
Access method: heap

                                                       Table "public.agent_metrics_config"
       Column        |           Type           | Collation | Nullable |      Default      | Storage  | Compression | Stats target | Description 
---------------------+--------------------------+-----------+----------+-------------------+----------+-------------+--------------+-------------
id                  | uuid                     |           | not null | gen_random_uuid() | plain    |             |              |
agent_type          | character varying(100)   |           | not null |                   | extended |             |              |
metrics_enabled     | boolean                  |           |          | true              | plain    |             |              |
collection_interval | integer                  |           |          | 60                | plain    |             |              |
retention_days      | integer                  |           |          | 30                | plain    |             |              |
metrics_config      | jsonb                    |           |          | '{}'::jsonb       | extended |             |              |
created_at          | timestamp with time zone |           |          | now()             | plain    |             |              |
updated_at          | timestamp with time zone |           |          | now()             | plain    |             |              |
Indexes:
"agent_metrics_config_pkey" PRIMARY KEY, btree (id)
"agent_metrics_config_agent_type_key" UNIQUE CONSTRAINT, btree (agent_type)
"idx_agent_metrics_config_agent" btree (agent_type)
Foreign-key constraints:
"agent_metrics_config_agent_type_fkey" FOREIGN KEY (agent_type) REFERENCES agent_definitions(type)
Access method: heap

                                                        Table "public.workflow_templates"
       Column        |           Type           | Collation | Nullable |      Default      | Storage  | Compression | Stats target | Description 
---------------------+--------------------------+-----------+----------+-------------------+----------+-------------+--------------+-------------
id                  | uuid                     |           | not null | gen_random_uuid() | plain    |             |              |
name                | character varying(255)   |           | not null |                   | extended |             |              |
description         | text                     |           |          |                   | extended |             |              |
category            | character varying(100)   |           |          |                   | extended |             |              |
workflow_definition | jsonb                    |           | not null |                   | extended |             |              |
default_config      | jsonb                    |           |          | '{}'::jsonb       | extended |             |              |
created_at          | timestamp with time zone |           |          | now()             | plain    |             |              |
updated_at          | timestamp with time zone |           |          | now()             | plain    |             |              |
Indexes:
"workflow_templates_pkey" PRIMARY KEY, btree (id)
"idx_workflow_templates_category" btree (category)
"idx_workflow_templates_name" btree (name)
"workflow_templates_name_key" UNIQUE CONSTRAINT, btree (name)
Access method: heap
