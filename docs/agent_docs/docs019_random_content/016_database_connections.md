# 017 — Database Connections

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