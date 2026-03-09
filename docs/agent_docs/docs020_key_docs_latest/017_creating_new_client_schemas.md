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

If `create_client_schema()` fails (e.g. missing pgvector extension for `agent_memory`), create tables manually. This is what was done for `client_system`:

```sql
-- 1. Create schema
CREATE SCHEMA client_my_new_client;

-- 2. Core tables (minimum required for agent spawning)
CREATE TABLE client_my_new_client.agent_instances (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    agent_type VARCHAR(100) NOT NULL,
    agent_id UUID NOT NULL,
    agent_name VARCHAR(255),
    client_id VARCHAR(100),
    status VARCHAR(50) DEFAULT 'active',
    metadata JSONB DEFAULT '{}',
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE TABLE client_my_new_client.agent_spawn_history (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    agent_type VARCHAR(100) NOT NULL,
    agent_id UUID NOT NULL,
    agent_name VARCHAR(255),
    parent_agent_id UUID,
    orchestration_id UUID,
    correlation_id UUID,
    client_id VARCHAR(100),
    status VARCHAR(50) DEFAULT 'spawned',
    metadata JSONB DEFAULT '{}',
    created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE TABLE client_my_new_client.projects (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL,
    description TEXT,
    settings JSONB DEFAULT '{}',
    status VARCHAR(50) DEFAULT 'active',
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE TABLE client_my_new_client.website_projects (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id UUID REFERENCES client_my_new_client.projects(id),
    site_id UUID,
    domain VARCHAR(255),
    settings JSONB DEFAULT '{}',
    status VARCHAR(50) DEFAULT 'active',
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

-- 3. Client record
INSERT INTO clients (external_id, name, settings)
VALUES ('my_new_client', 'My New Client', '{"type": "customer"}'::jsonb)
ON CONFLICT DO NOTHING;
```

The `agent_memory` and `workflow_executions` tables are optional — they're used by RAG and execution tracking features which may not be active yet.

---

## What Each Table Does

| Table | Purpose | Used by |
|-------|---------|---------|
| `agent_instances` | Active agent records per client | `spawn_agent` action (required) |
| `agent_spawn_history` | Audit log of all spawns | `spawn_agent` action (required) |
| `projects` | Client project groupings | API / future project management |
| `website_projects` | Sites linked to projects | API / future project management |
| `agent_memory` | Vector embeddings for RAG | RAG actions (optional) |
| `workflow_executions` | Execution tracking | Execution logging (optional) |

The minimum tables needed for agent spawning are `agent_instances` and `agent_spawn_history`. Without these, `spawn_agent` fails with `relation "client_X.agent_instances" does not exist`.

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

The chassis validator requires a non-empty `client_id` header. It does not do a database lookup — it just checks the string is present. The spawn action then uses it to resolve the schema (`client_` + client_id).

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

**Error: `Client already exists` (409 from API)**
Schema exists. Check: `SELECT schema_name FROM information_schema.schemata WHERE schema_name = 'client_X';`

**Error: `create_client_schema failed` with pgvector**
The function tries to create `agent_memory` with a `vector(1536)` column. If pgvector extension isn't installed, use Method 3 (manual) and skip the `agent_memory` table.

**Scheduler messages rejected: `missing required fields`**
The `client_id` header is empty or missing. Add it to the message producer. The chassis validator checks headers, not the message body.
