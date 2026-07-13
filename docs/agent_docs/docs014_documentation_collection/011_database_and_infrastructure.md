# 011 — Database & Infrastructure

How the system connects to its three databases, the connection pooling architecture, and troubleshooting.

---

## Three Databases

| Database | Engine | Host | Port | Used by |
|----------|--------|------|------|---------|
| `clients_db` | PostgreSQL 16 | `postgres-clients-0` (in-cluster) | 5432 (direct), 6432 (pgbouncer) | core-manager, agent-chassis, kafka-scheduler |
| `templates_db` | PostgreSQL 16 | `postgres-templates-0` (in-cluster) | 5432 (direct), 6432 (pgbouncer) | core-manager |
| `catalogu_vectordb_chassis` | MySQL 8 | `rs17.uk-noc.com` (external, Clook/cPanel) | 3306 | auth-service only |

---

## Connection Architecture

```
                                ┌─────────────────────┐
                                │  rs17.uk-noc.com    │
                                │  MySQL 3306         │
                                │  (external, Clook)  │
                                └──────────▲──────────┘
                                           │ outbound from cluster
                                           │ (requires Remote MySQL whitelist)
┌──────────────┐                ┌──────────┴──────────┐
│ auth-service │───── *sql.DB ──│  egress via node IP │
│ (Go, MySQL)  │                │  134.213.168.37/44/45│
└──────────────┘                └─────────────────────┘

┌──────────────┐     ┌───────────┐     ┌──────────────────┐
│ core-manager │     │ pgbouncer │     │ postgres-clients  │
│ agent-chassis│──►──│ :6432     │──►──│ :5432             │
│ kafka-sched  │     │ tx mode   │     │ clients_db        │
└──────────────┘     │           │     └──────────────────┘
                     │           │     ┌──────────────────┐
                     │           │──►──│ postgres-templates│
                     │           │     │ :5432             │
                     └───────────┘     │ templates_db      │
                                       └──────────────────┘
```

---

## PostgreSQL: Connection Path

### Application → pgbouncer → PostgreSQL

All Go services connect to pgbouncer, not directly to PostgreSQL.

| Layer | Address | Purpose |
|-------|---------|---------|
| Application | `*sql.DB` (pgx stdlib adapter) | Go-side connection pool, max 10 conns per service |
| pgbouncer | `pgbouncer.ai-persona-system.svc.cluster.local:6432` | Connection pooler, transaction mode |
| PostgreSQL | `postgres-clients-0:5432` / `postgres-templates-0:5432` | Actual database |

### pgbouncer Configuration

```ini
pool_mode = transaction          # Connections returned to pool after each transaction
max_client_conn = 200            # Total client connections accepted
default_pool_size = 15           # Actual PG connections per user/db pair
min_pool_size = 2                # Keep warm
reserve_pool_size = 5            # Extra if default exhausted
```

### Why transaction mode matters

In transaction mode, pgbouncer reassigns server connections between transactions. This means:

- **No prepared statements** — they're per-connection, but the connection changes between calls
- **No session-level state** — `SET` commands, temp tables, advisory locks don't persist
- **`*sql.DB` with `simple_protocol`** — the Go connection string includes `default_query_exec_mode=simple_protocol` or `cache_describe` to avoid prepared statement issues

### Connection strings

**Agent chassis and kafka-scheduler** (via env var):
```
postgresql://clients_user:$(CLIENTS_DB_PASSWORD)@pgbouncer.ai-persona-system.svc.cluster.local:6432/clients_db?sslmode=disable&default_query_exec_mode=cache_describe
```

**Core-manager** (via `NewStdlibConnection` in platform/database):
```
postgresql://clients_user:<password>@pgbouncer...:6432/clients_db?sslmode=disable&default_query_exec_mode=simple_protocol
```

### Go connection pool settings

```go
db.SetMaxOpenConns(10)           // Don't exceed pgbouncer's pool per service
db.SetMaxIdleConns(5)            // Keep some warm
db.SetConnMaxLifetime(30 * min)  // Rotate to avoid stale connections
db.SetConnMaxIdleTime(5 * min)   // Close idle connections promptly
```

### Direct PostgreSQL access (bypassing pgbouncer)

Some operations need direct access — `pg_dump` for backups, `LISTEN/NOTIFY`, advisory locks:

```
postgres-clients-0.ai-persona-system.svc.cluster.local:5432
postgres-templates-0.ai-persona-system.svc.cluster.local:5432
```

The backup cronjob uses direct connections for `pg_dump`.

---

## PostgreSQL: Go Driver Migration

### Old pattern (pgxpool.Pool) — agent-chassis still uses this

```go
import "github.com/jackc/pgx/v5/pgxpool"

pool, err := pgxpool.New(ctx, connString)
pool.QueryRow(ctx, query, args...)   // no Context suffix
pool.Query(ctx, query, args...)
pool.Exec(ctx, query, args...)
pool.Begin(ctx)                      // returns pgx.Tx
tx.Rollback(ctx)                     // takes context
tx.Commit(ctx)                       // takes context
result.RowsAffected()                // returns int64 (single value)
pool.Stat()                          // returns *pgxpool.Stat
stat.AcquiredConns()                 // pgxpool-specific
pgx.ErrNoRows                       // pgx error type
```

### New pattern (*sql.DB) — core-manager uses this

```go
import "database/sql"
import _ "github.com/jackc/pgx/v5/stdlib"

db, err := sql.Open("pgx", connString)
db.QueryRowContext(ctx, query, args...)   // Context suffix required
db.QueryContext(ctx, query, args...)
db.ExecContext(ctx, query, args...)
db.BeginTx(ctx, nil)                     // returns *sql.Tx
tx.Rollback()                            // no context
tx.Commit()                              // no context
affected, err := result.RowsAffected()   // returns (int64, error)
db.Stats()                               // returns sql.DBStats
stats.OpenConnections                     // field, not method
sql.ErrNoRows                            // stdlib error type
db.PingContext(ctx)                       // Context suffix
```

### Conversion cheat sheet

| pgxpool | sql.DB |
|---------|--------|
| `.Query(ctx,` | `.QueryContext(ctx,` |
| `.QueryRow(ctx,` | `.QueryRowContext(ctx,` |
| `.Exec(ctx,` | `.ExecContext(ctx,` |
| `.Begin(ctx)` | `.BeginTx(ctx, nil)` |
| `tx.Rollback(ctx)` | `tx.Rollback()` |
| `tx.Commit(ctx)` | `tx.Commit()` |
| `result.RowsAffected() == 0` | `n, _ := result.RowsAffected(); n == 0` |
| `pool.Stat()` | `db.Stats()` |
| `stat.AcquiredConns()` | `stats.OpenConnections` |
| `pgx.ErrNoRows` | `sql.ErrNoRows` |
| `pool.Ping(ctx)` | `db.PingContext(ctx)` |
| `pool.Close()` | `db.Close()` |

---

## MySQL: Auth Database

### Connection details

| Field | Value |
|-------|-------|
| Host | `rs17.uk-noc.com` |
| Port | 3306 |
| User | `catalogu_personae` |
| Database | `catalogu_vectordb_chassis` |
| Password env var | `AUTH_DB_PASSWORD` |
| Password source | `personae-platform-secrets` K8s secret |

### Remote access whitelist

The MySQL host is on Clook shared hosting (cPanel). Remote connections require the client IP to be registered in cPanel's Remote MySQL interface.

**Current cluster egress IPs** (all three nodes):

```
134.213.168.37  (prod-instance-...1148)
134.213.168.44  (prod-instance-...1149)
134.213.168.45  (prod-instance-...1150)
```

**Whitelist entry in cPanel:** `134.213.168.%` — covers all current nodes and future nodes in the same /24 range.

### If the connection breaks

1. Check if node IPs have changed: `kubectl get nodes -o wide`
2. Check egress IP: `kubectl run ip-check --rm -it --restart=Never --image=alpine -- sh -c "apk add curl && curl -s ifconfig.me"`
3. Update the cPanel Remote MySQL entry if the IP range has changed
4. Test: `kubectl run mysql-test --rm -it --restart=Never --image=mysql:8.0 -- mysql -h rs17.uk-noc.com -P 3306 -u catalogu_personae -p --connect-timeout=10 catalogu_vectordb_chassis`

Error `2003 (HY000): Can't connect to MySQL server (110)` means the IP is not whitelisted — error 110 is "connection timed out" at the TCP level.

### Tables (tiny, mostly schema-only)

| Table | Rows | Size |
|-------|------|------|
| users | 0 | 32 KiB |
| auth_tokens | 0 | 32 KiB |
| permissions | 6 | 32 KiB |
| projects | 0 | 32 KiB |
| subscriptions | 0 | 48 KiB |
| subscription_tiers | 4 | 32 KiB |
| user_permissions | 0 | 32 KiB |
| user_profiles | 0 | 16 KiB |

### MySQL syntax in Go code

The auth-service and dashboard_handlers.go use MySQL syntax for queries against `authDB`:
- `CURDATE()` (not `CURRENT_DATE`)
- `DATE_SUB(NOW(), INTERVAL 1 MONTH)` (not `NOW() - INTERVAL '1 month'`)
- `INTERVAL 30 DAY` (not `INTERVAL '30 days'`)

Do not convert these to PostgreSQL syntax unless the auth database is migrated to PostgreSQL.

### Future: Migrate auth to PostgreSQL?

The auth database is small enough to migrate trivially. Benefits:
- Eliminate MySQL dependency entirely
- Use pgbouncer for connection pooling
- Simplify backup (one pg_dump covers everything)
- Remove the external hosting dependency and cPanel IP whitelist maintenance

Migration would involve:
1. Create `auth` schema in `clients_db` (or a separate `auth_db` in the same PostgreSQL cluster)
2. Convert MySQL DDL to PostgreSQL (mostly `AUTO_INCREMENT` → `SERIAL`, `DATETIME` → `TIMESTAMPTZ`)
3. Update auth-service Go code: `database/sql` + MySQL driver → `database/sql` + pgx driver
4. Convert MySQL syntax in queries (`CURDATE()` → `CURRENT_DATE` etc.)
5. Port the data (trivial at current size)

Not urgent — the current setup works, it just requires IP whitelist maintenance.

---

## Credentials

All database passwords are stored in the `personae-platform-secrets` K8s secret.

| Secret key | Used for |
|------------|----------|
| `CLIENTS_DB_PASSWORD` | PostgreSQL clients_db user password |
| `TEMPLATES_DB_PASSWORD` | PostgreSQL templates_db user password |
| `AUTH_DB_PASSWORD` | MySQL auth database password |
| `JWT_SECRET_KEY` | JWT signing (auth-service + core-manager shared) |
| `B2_APPLICATION_KEY_ID` | Backblaze B2 access key (for backups and asset storage) |
| `B2_APPLICATION_KEY` | Backblaze B2 secret key |
| `AGENT_BOOTSTRAP_KEY` | Agent bootstrap authentication |

Managed by Terraform in `047-base-configs` with values in `terraform.tfvars.secret` (not committed to git).

---

## Troubleshooting

### "prepared statement does not exist"

pgbouncer transaction mode reassigned the connection. Fix: ensure the connection string includes `default_query_exec_mode=simple_protocol` or `cache_describe`.

### "too many connections"

Check pgbouncer stats:
```bash
kubectl -n ai-persona-system exec deploy/pgbouncer -- psql -p 6432 pgbouncer -c "SHOW POOLS;"
kubectl -n ai-persona-system exec deploy/pgbouncer -- psql -p 6432 pgbouncer -c "SHOW CLIENTS;"
```

Each Go service should use `MaxOpenConns(10)`. With 5 chassis pods + core-manager + scheduler = ~70 client connections, well within pgbouncer's 200 max.

### "connection refused" to PostgreSQL

Check pod is running: `kubectl -n ai-persona-system get pods | grep postgres`
Check pgbouncer: `kubectl -n ai-persona-system get pods | grep pgbouncer`

### MySQL "Can't connect" (error 110)

IP not whitelisted in cPanel. Check egress IP and update Remote MySQL. See the MySQL section above.

### Backup pg_dump failures

Backups must connect directly to PostgreSQL (not pgbouncer). pg_dump uses extended query protocol with prepared statements, which breaks in transaction mode. The backup cronjob connects to `postgres-clients-0:5432` directly.


Testing connection:

That's progress — it connected. The error changed from `2003 (connection timed out)` to `1045 (access denied)`. The IP whitelist is working. The problem is `-p` with no value tries to prompt for a password interactively, which doesn't work in this kubectl context.

Pass the password inline:

```bash
# Get the password from the secret first
kubectl -n ai-persona-system get secret personae-platform-secrets \
  -o jsonpath='{.data.AUTH_DB_PASSWORD}' | base64 -d && echo

# Then test with it (no space between -p and the password)
kubectl -n ai-persona-system run mysql-test --rm -it --restart=Never \
  --image=mysql:8.0 -- mysql -h rs17.uk-noc.com -P 3306 \
  -u catalogu_personae -p"PpC47410423123!" --connect-timeout=10 catalogu_vectordb_chassis
```

If that connects, the auth-service should work too. The `(using password: NO)` in the error confirms no password was sent — it's not an IP issue.


----


database backup

# Deploy the CronJob
make deploy-database-backup ENVIRONMENT=production REGION=uk001

# Run one immediately to test
make backup-now ENVIRONMENT=production REGION=uk001

# Watch it
make backup-logs ENVIRONMENT=production REGION=uk001-e 

---

## Creating a New Client Schema

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

---

## Admin User Setup Runbook

### Register user via auth-service

```bash
kubectl -n ai-persona-system exec -it $(kubectl -n ai-persona-system get pod -l app=auth-service -o jsonpath='{.items[0].metadata.name}') -- wget -qO- \
http://localhost:8081/api/v1/auth/register \
--post-data='{"email":"uk@websy.uk","password":"AdminPass2026xyz","client_id":"demo_client"}' \
--header='Content-Type: application/json' 2>&1
```

### Promote to admin in MySQL

```bash
# Start a temporary pod with mysql client
kubectl -n ai-persona-system run mysql-check --rm -it --image=postgres:16-alpine -- /bin/sh
# Inside: apk add --no-cache mysql-client

mysql -h rs17.uk-noc.com -u catalogu_personae -p"<password>" --skip-ssl catalogu_vectordb_chassis \
-e "UPDATE users SET role = 'admin', subscription_tier = 'enterprise' WHERE email = 'uk@websy.uk';"
```

### Get fresh admin token

```bash
TOKEN=$(kubectl -n ai-persona-system exec -it $(kubectl -n ai-persona-system get pod -l app=auth-service -o jsonpath='{.items[0].metadata.name}') -- wget -qO- \
http://localhost:8081/api/v1/auth/login \
--post-data='{"email":"uk@websy.uk","password":"AdminPass2026xyz"}' \
--header='Content-Type: application/json' 2>/dev/null | python3 -c "import sys,json; print(json.load(sys.stdin)['access_token'])")
```

### Test admin API

```bash
kubectl -n ai-persona-system port-forward svc/core-manager 8088:8088 &
curl -s http://localhost:8088/api/v1/admin/sites -H "Authorization: Bearer $TOKEN" | python3 -m json.tool | head -40
```
