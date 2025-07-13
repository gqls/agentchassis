#!/bin/bash
# FILE: docker/scripts/wait-for-services.sh
set -e

# Colors for output
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m'

echo -e "${YELLOW}⏳ Waiting for services to be ready...${NC}"

# Function to wait for PostgreSQL
wait_for_postgres() {
    local host=$1
    local user=$2
    local database=$3
    local password=$4
    local service_name=$5

    echo -e "${YELLOW}🔍 Waiting for PostgreSQL: $service_name...${NC}"

    export PGPASSWORD="$password"

    local max_attempts=60
    local attempt=1

    while [ $attempt -le $max_attempts ]; do
        if pg_isready -h "$host" -U "$user" -d "$database" >/dev/null 2>&1; then
            if psql -h "$host" -U "$user" -d "$database" -c "SELECT 1;" >/dev/null 2>&1; then
                echo -e "${GREEN}✅ $service_name is ready${NC}"
                return 0
            fi
        fi

        if [ $attempt -eq $max_attempts ]; then
            echo -e "${RED}❌ $service_name failed to become ready after $max_attempts attempts${NC}"
            return 1
        fi

        echo -e "${YELLOW}⏳ Attempt $attempt/$max_attempts - $service_name not ready yet...${NC}"
        sleep 5
        ((attempt++))
    done
}

# Function to wait for MySQL
wait_for_mysql() {
    local host=$1
    local user=$2
    local database=$3
    local password=$4
    local service_name=$5

    echo -e "${YELLOW}🔍 Waiting for MySQL: $service_name...${NC}"

    local max_attempts=60
    local attempt=1

    while [ $attempt -le $max_attempts ]; do
        if mysqladmin ping -h "$host" --silent >/dev/null 2>&1; then
            if mysql -h "$host" -u "$user" -p"$password" -e "SELECT 1;" "$database" >/dev/null 2>&1; then
                echo -e "${GREEN}✅ $service_name is ready${NC}"
                return 0
            fi
        fi

        if [ $attempt -eq $max_attempts ]; then
            echo -e "${RED}❌ $service_name failed to become ready after $max_attempts attempts${NC}"
            return 1
        fi

        echo -e "${YELLOW}⏳ Attempt $attempt/$max_attempts - $service_name not ready yet...${NC}"
        sleep 5
        ((attempt++))
    done
}

# Function to wait for Kafka
wait_for_kafka() {
    local host=$1
    local port=$2
    local service_name=$3

    echo -e "${YELLOW}🔍 Waiting for Kafka: $service_name...${NC}"

    local max_attempts=60
    local attempt=1

    while [ $attempt -le $max_attempts ]; do
        if timeout 5 bash -c "</dev/tcp/$host/$port" >/dev/null 2>&1; then
            # Additional check: try to list topics
            if kafka-topics --bootstrap-server "$host:$port" --list >/dev/null 2>&1; then
                echo -e "${GREEN}✅ $service_name is ready${NC}"
                return 0
            fi
        fi

        if [ $attempt -eq $max_attempts ]; then
            echo -e "${RED}❌ $service_name failed to become ready after $max_attempts attempts${NC}"
            return 1
        fi

        echo -e "${YELLOW}⏳ Attempt $attempt/$max_attempts - $service_name not ready yet...${NC}"
        sleep 5
        ((attempt++))
    done
}

# Wait for all required services
echo -e "${YELLOW}🚀 Starting service readiness checks...${NC}"

# Wait for PostgreSQL databases
if [ ! -z "$CLIENTS_DB_PASSWORD" ]; then
    wait_for_postgres "postgres-clients" "clients_user" "clients_db" "$CLIENTS_DB_PASSWORD" "PostgreSQL Clients"
fi

if [ ! -z "$TEMPLATES_DB_PASSWORD" ]; then
    wait_for_postgres "postgres-templates" "templates_user" "templates_db" "$TEMPLATES_DB_PASSWORD" "PostgreSQL Templates"
fi

# Wait for MySQL database
if [ ! -z "$AUTH_DB_PASSWORD" ]; then
    wait_for_mysql "mysql-auth" "auth_user" "auth_db" "$AUTH_DB_PASSWORD" "MySQL Auth"
fi

# Wait for Kafka (optional - only if we need to create topics)
if [ "$WAIT_FOR_KAFKA" = "true" ]; then
    wait_for_kafka "kafka-0.kafka-headless" "9092" "Kafka"
fi

echo -e "${GREEN}🎉 All required services are ready!${NC}"