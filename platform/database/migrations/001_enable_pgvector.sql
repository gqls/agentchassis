-- FILE: platform/database/migrations/001_enable_pgvector.sql
-- Run this on the clients database as superuser
CREATE EXTENSION IF NOT EXISTS vector;
