#!/bin/bash

# Set test database environment variables
export TEST_DB_HOST=localhost
export TEST_DB_PORT=5432
export TEST_DB_USER=postgres
export TEST_DB_PASSWORD=your_password
export TEST_DB_NAME=agentchassis_test

# Create test database if it doesn't exist
PGPASSWORD=$TEST_DB_PASSWORD psql -h $TEST_DB_HOST -U $TEST_DB_USER -c "CREATE DATABASE IF NOT EXISTS $TEST_DB_NAME;"

# Run tests
go test -v ./platform/orchestration/...