#!/bin/bash
# FILE: docker/scripts/run-migrations.sh
set -e

echo "🔧 Starting database migrations..."

# Wait for services to be ready
/app/wait-for-services.sh

# Colors for output
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m'

# Function to run PostgreSQL migration
run_postgres_migration() {
    local host=$1
    local user=$2
    local database=$3
    local password=$4
    local migration_file=$5
    local description=$6

    echo -e "${YELLOW}📝 Running $description...${NC}"

    export PGPASSWORD="$password"

    if psql -h "$host" -U "$user" -d "$database" -f "$migration_file"; then
        echo -e "${GREEN}✅ $description completed${NC}"
    else
        echo -e "${RED}❌ $description failed${NC}"
        exit 1
    fi
}

# Function to run MySQL migration
run_mysql_migration() {
    local host=$1
    local user=$2
    local database=$3
    local password=$4
    local migration_file=$5
    local description=$6

    echo -e "${YELLOW}📝 Running $description...${NC}"

    if mysql -h "$host" -u "$user" -p"$password" "$database" < "$migration_file"; then
        echo -e "${GREEN}✅ $description completed${NC}"
    else
        echo -e "${RED}❌ $description failed${NC}"
        exit 1
    fi
}

# 1. Enable pgvector extension on clients database
echo -e "${YELLOW}🔧 Enabling pgvector extension...${NC}"
export PGPASSWORD="$CLIENTS_DB_PASSWORD"
psql -h postgres-clients -U clients_user -d clients_db -c "CREATE EXTENSION IF NOT EXISTS vector;" || {
    echo -e "${RED}❌ Failed to enable pgvector${NC}"
    exit 1
}
echo -e "${GREEN}✅ pgvector enabled${NC}"

# 2. Migrate templates database
run_postgres_migration \
    "postgres-templates" \
    "templates_user" \
    "templates_db" \
    "$TEMPLATES_DB_PASSWORD" \
    "/app/migrations/002_create_templates_schema.sql" \
    "Templates database migration"

# 3. Create base clients database structure
run_postgres_migration \
    "postgres-clients" \
    "clients_user" \
    "clients_db" \
    "$CLIENTS_DB_PASSWORD" \
    "/app/migrations/003_create_client_base.sql" \
    "Clients database base structure"

# 4. Create orchestrator state table
echo -e "${YELLOW}📝 Creating orchestrator state table...${NC}"
export PGPASSWORD="$CLIENTS_DB_PASSWORD"
psql -h postgres-clients -U clients_user -d clients_db -c "
CREATE TABLE IF NOT EXISTS orchestrator_state (
    correlation_id UUID PRIMARY KEY,
    status VARCHAR(50) NOT NULL,
    current_step VARCHAR(255) NOT NULL,
    awaited_steps JSONB DEFAULT '[]',
    collected_data JSONB DEFAULT '{}',
    initial_request_data JSONB,
    final_result JSONB,
    error TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_orchestrator_state_status ON orchestrator_state(status);
CREATE INDEX IF NOT EXISTS idx_orchestrator_state_updated_at ON orchestrator_state(updated_at);
" || {
    echo -e "${RED}❌ Failed to create orchestrator state table${NC}"
    exit 1
}
echo -e "${GREEN}✅ Orchestrator state table created${NC}"

# 5. Migrate auth database (MySQL)
run_mysql_migration \
    "mysql-auth" \
    "auth_user" \
    "auth_db" \
    "$AUTH_DB_PASSWORD" \
    "/app/migrations/004_auth_schema.sql" \
    "Auth database schema migration"

run_mysql_migration \
    "mysql-auth" \
    "auth_user" \
    "auth_db" \
    "$AUTH_DB_PASSWORD" \
    "/app/migrations/005_projects_schema.sql" \
    "Projects and subscriptions schema migration"

echo -e "${GREEN}🎉 All database migrations completed successfully!${NC}"