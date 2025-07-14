#!/bin/bash
# Fix UUID generation in migration files

# Update auth_schema.sql
sed -i 's/gen_random_uuid()/gen_random_uuid()/g' ../platform/database/migrations/004_auth_schema.sql

# Update client_schema.sql - ensure we're using gen_random_uuid() consistently
sed -i 's/uuid_generate_v4()/gen_random_uuid()/g' ../platform/database/migrations/003_create_client_schema.sql

# Remove the CREATE EXTENSION IF NOT EXISTS "uuid-ossp" line as it's not needed with gen_random_uuid()
sed -i '/CREATE EXTENSION IF NOT EXISTS "uuid-ossp";/d' ../platform/database/migrations/003_create_client_schema.sql

echo "UUID generation standardized to use gen_random_uuid() across all migrations"