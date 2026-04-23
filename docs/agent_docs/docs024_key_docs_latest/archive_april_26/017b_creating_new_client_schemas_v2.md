# 016 — Creating a New Client Schema

Guide for setting up a new client in the system. Each client gets a dedicated PostgreSQL schema for isolation of agent instances, spawn history, and project data.

---

## When to Create a Client

- New customer onboarding
- Internal system identity (e.g. `system` for the kafka-scheduler)
- Testing environments (e.g. `test_client`)

---

## Prerequisites

- Access to `clients_db` (the main database)
- The `create_client_schema` function exists (check: `SELECT proname FROM pg_proc WHERE proname = 'create_client_schema';`)

---

## Method 1: Via the API (preferred for customer clients)

```bash
# Port-forward if calling from outside the cluster
kubectl -n ai-persona-system port-forward svc/core-manager 8088:8088 &

curl -s -X POST http://localhost:8088/api/clients \
  -H "Content-Type: application/json" \
  -d '{
    "client_id": "my_new_client",
    "display_name": "My New Client",
    "settings": {"type": "customer"}
  }'
```

Requirements for `client_id`:
- 3-50 characters
- Alphanumeric and underscores only (`a-z`, `A-Z`, `0-9`, `_`)
- No spaces, hyphens, or special characters

The API calls `create_client_schema()` which creates the schema plus all tables, indexes, and comments.

---

## Method 2: Via SQL (for system/internal clients or when API is unavailable)

### Step 1: Create the schema

```sql
SELECT create_client_schema('my_new_client');
```

This creates:
- Schema `client_my_new_client`
- Tables: `agent_instances`, `agent_spawn_history`, `projects`, `website_projects`, `agent_memory`, `workflow_executions`
- Indexes on all key columns
- Foreign key constraints between tables

### Step 2: Insert the client record

```sql
INSERT INTO clients (external_id, name, settings)
VALUES ('my_new_client', 'My New Client', '{"type": "customer"}'::jsonb)
ON CONFLICT DO NOTHING;
```

### Step 3: Verify

```sql
-- Check schema exists
SELECT schema_name FROM information_schema.schemata
WHERE schema_name = 'client_my_new_client';

-- Check tables created
SELECT table_name FROM information_schema.tables
WHERE table_schema = 'client_my_new_client'
ORDER BY table_name;

-- Expected:
--  agent_instances
--  agent_memory
--  agent_spawn_history
--  projects
--  website_projects
--  workflow_executions

-- Check client record
SELECT * FROM clients WHERE external_id = 'my_new_client';
```

---

## Method 3: Manual table creation (fallback)

If `create_client_schema()` fails (e.g. missing pgvector extension for `agent_memory`), create tables manually.

**The column definitions below must match what `create_client_schema()` produces and what `spawn_agent` Go code expects.** Do not invent column names — the Go code inserts specific columns and will fail at runtime if they don't exist. When in doubt, check the function source:

```sql
SELECT prosrc FROM pg_proc WHERE proname = 'create_client_schema';
```

And check what `spawn_agent` inserts:

```
-- spawn_actions.go inserts into agent_instances:
-- (id, template_id, owner_user_id, name, config, is_active, created_at, updated_at)
```

### Step 1: Create schema and tables

```sql
-- 1. Create schema
CREATE SCHEMA client_my_new_client;

-- 2. Projects (must be created before agent_instances due to FK)
CREATE TABLE client_my_new_client.projects (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL,
    description TEXT,
    owner_user_id VARCHAR(255) NOT NULL DEFAULT 'system',
    settings JSONB DEFAULT '{}'::jsonb,
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
);

-- 3. Agent instances (required for spawn_agent)
--    Column names must match spawn_actions.go INSERT statement exactly:
--    (id, template_id, owner_user_id, name, config, is_active, created_at, updated_at)
CREATE TABLE client_my_new_client.agent_instances (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    template_id UUID REFERENCES agent_definitions(id),
    project_id UUID,
    owner_user_id VARCHAR(255) NOT NULL DEFAULT 'system',
    name VARCHAR(255) NOT NULL,
    config JSONB NOT NULL DEFAULT '{}'::jsonb,
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,
    CONSTRAINT fk_project FOREIGN KEY (project_id)
        REFERENCES client_my_new_client.projects(id) ON DELETE CASCADE
);

-- 4. Agent spawn history (required for spawn_agent audit trail)
CREATE TABLE client_my_new_client.agent_spawn_history (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    parent_agent_id UUID,
    spawned_agent_id UUID NOT NULL,
    spawn_reason TEXT,
    spawn_config JSONB DEFAULT '{}'::jsonb,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

-- 5. Website projects (optional — used by API)
CREATE TABLE client_my_new_client.website_projects (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id UUID NOT NULL REFERENCES client_my_new_client.projects(id) ON DELETE CASCADE,
    domain VARCHAR(255) NOT NULL,
    business_name VARCHAR(255),
    business_type VARCHAR(100),
    status VARCHAR(50) NOT NULL DEFAULT 'planning',
    site_structure JSONB,
    content_data JSONB,
    visual_assets JSONB,
    preview_url VARCHAR(500),
    live_url VARCHAR(500),
    s3_bucket VARCHAR(255),
    cloudfront_distribution_id VARCHAR(255),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    published_at TIMESTAMPTZ,
    CONSTRAINT valid_status CHECK (status IN (
        'planning', 'researching', 'designing',
        'developing', 'reviewing', 'published', 'archived'
    ))
);

-- 6. Indexes
CREATE INDEX IF NOT EXISTS idx_my_new_client_agent_instances_active
    ON client_my_new_client.agent_instances(is_active) WHERE deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS idx_my_new_client_agent_instances_owner
    ON client_my_new_client.agent_instances(owner_user_id);

CREATE INDEX IF NOT EXISTS idx_my_new_client_projects_active
    ON client_my_new_client.projects(is_active) WHERE deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS idx_my_new_client_spawn_history_parent
    ON client_my_new_client.agent_spawn_history(parent_agent_id);

CREATE INDEX IF NOT EXISTS idx_my_new_client_spawn_history_spawned
    ON client_my_new_client.agent_spawn_history(spawned_agent_id);
```

### Step 2: Insert the client record

```sql
INSERT INTO clients (external_id, name, settings)
VALUES ('my_new_client', 'My New Client', '{"type": "customer"}'::jsonb)
ON CONFLICT DO NOTHING;
```

### Step 3: Verify

```sql
-- Check tables
SELECT table_name FROM information_schema.tables
WHERE table_schema = 'client_my_new_client' ORDER BY table_name;

-- Check agent_instances columns match what spawn_agent expects
SELECT column_name, data_type
FROM information_schema.columns
WHERE table_schema = 'client_my_new_client' AND table_name = 'agent_instances'
ORDER BY ordinal_position;

-- Expected columns for agent_instances:
--  id              uuid
--  template_id     uuid
--  project_id      uuid
--  owner_user_id   character varying
--  name            character varying
--  config          jsonb
--  is_active       boolean
--  created_at      timestamp with time zone
--  updated_at      timestamp with time zone
--  deleted_at      timestamp with time zone
```

### Optional tables

The `agent_memory` and `workflow_executions` tables are created by `create_client_schema()` but are only used by RAG and execution tracking features. Skip them for manual creation unless those features are active. If needed later, run `create_client_schema()` which uses `IF NOT EXISTS` and will add the missing tables without affecting existing ones.

---

## What Each Table Does

| Table | Purpose | Used by | Required? |
|-------|---------|---------|-----------|
| `agent_instances` | Active agent records per client | `spawn_agent` action | Yes — spawning fails without it |
| `agent_spawn_history` | Audit log of all spawns | `spawn_agent` action | Yes — spawning fails without it |
| `projects` | Client project groupings | API / project management | Yes — FK from agent_instances |
| `website_projects` | Sites linked to projects | API / project management | No |
| `agent_memory` | Vector embeddings for RAG | RAG actions | No |
| `workflow_executions` | Execution tracking | Execution logging | No |

---

## Using the Client ID

Once created, the client_id goes in Kafka message headers:

```json
{
    "client_id": "my_new_client",
    "message_type": "request",
    "action": "orchestrate"
}
```

The chassis validator requires a non-empty `client_id` header. It does not do a database lookup — it just checks the string is present. The `spawn_agent` action then uses the client_id to resolve the schema name: `client_` + client_id.

---

## Current Clients

| client_id | Schema | Purpose |
|-----------|--------|---------|
| `demo_client` | `client_demo_client` | Main development/production client |
| `vetcomparison` | `client_vetcomparison` | Vet comparison site client |
| `test_client` | `client_test_client` | Testing |
| `system` | `client_system` | Kafka scheduler internal identity |

---

## Troubleshooting

**Error: `relation "client_X.agent_instances" does not exist`**
Schema or tables weren't created. Run Method 2 or 3 above.

**Error: `column "template_id" of relation "agent_instances" does not exist`**
The `agent_instances` table was created with wrong columns. The Go code (`spawn_actions.go`) inserts `(id, template_id, owner_user_id, name, config, is_active, created_at, updated_at)`. Drop the table and recreate using the definitions in Method 3 above. Always check column definitions against `create_client_schema()` function source before creating manually.

**Error: `Client already exists` (409 from API)**
Schema exists. Check: `SELECT schema_name FROM information_schema.schemata WHERE schema_name = 'client_X';`

**Error: `create_client_schema failed` with pgvector**
The function tries to create `agent_memory` with a `vector(1536)` column. If pgvector extension isn't installed, use Method 3 (manual) and skip the `agent_memory` table.

**Scheduler messages rejected: `missing required fields`**
The `client_id` header is empty or missing. Add it to the message producer. The chassis validator checks headers, not the message body.

**Error: `schema "client_X" does not exist`**
The `create_client_schema()` function was never called, or the `INSERT INTO clients` was run without creating the schema first. The schema must be created explicitly — inserting into the `clients` table does not create it.
