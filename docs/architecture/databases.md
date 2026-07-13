Database Architecture Summary for Core Manager Service
Overview
The system uses 3 separate databases with distinct responsibilities:
1. MySQL Database (Auth Database)
   Purpose: User authentication and management
   Tables

users - Main user accounts (using BINARY(16) for UUID storage)
auth_tokens - JWT refresh token storage
user_profiles - Extended user profile information
projects - User projects
subscriptions - User subscription information
subscription_tiers - Available subscription tiers (free, basic, premium, enterprise)
permissions - System permissions
user_permissions - User-permission mappings
user_activity_logs - Activity tracking

Key Notes

Uses MySQL-specific syntax (BINARY(16) for UUIDs, TIMESTAMP instead of TIMESTAMPTZ)
Includes pre-populated subscription tiers and permissions
Handles all authentication and user management concerns

2. PostgreSQL Database #1 (Clients Database)
   Purpose: Multi-tenant agent data with schema isolation
   Global Tables (shared across all clients)

agent_definitions - Global agent type definitions (copywriter, researcher, etc.)
orchestrator_state - Global workflow orchestration state
clients_info - Client metadata (added by Core Manager)

Per-Client Schema Tables (in schema client_{client_id})

agent_instances - Client-specific agent instances
agent_memory - Agent memory with vector embeddings (requires pgvector)
projects - Client-specific projects
workflow_executions - Workflow execution history
usage_analytics - Usage tracking per client

Key Features

Requires pgvector extension for AI embeddings
Uses a create_client_schema() function to create isolated schemas per client
Pre-populated with default agent types
Strong multi-tenant isolation

3. PostgreSQL Database #2 (Templates Database)
   Purpose: Global persona templates
   Tables

persona_templates - Shared templates available to all clients

Migration Order

1. Enable pgvector on Clients database:

CREATE EXTENSION IF NOT EXISTS vector;

2. Create auth schema in MySQL database
3. Create templates schema in Templates PostgreSQL
4. Create global tables in Clients PostgreSQL
5. For each new client, execute:

SELECT create_client_schema('client_id');

Key Architecture Decisions
1. Database Separation

Authentication completely isolated in MySQL
AI/Agent functionality in PostgreSQL (better for JSONB and vector operations)
Templates separated from client data for reusability

2. Multi-tenancy Strategy

Schema-per-client approach in Clients database
Provides strong isolation between clients
Global resources (agent definitions, templates) are shared

3. Core Manager's Role

Primary access to Clients and Templates databases
Read-only access to Auth database (for admin features)
Manages the clients_info table for client metadata
Handles orchestration state across all clients

Benefits

Excellent isolation between authentication and business logic
Strong multi-tenant boundaries with schema separation
Leverages PostgreSQL features (JSONB, pgvector) for AI functionality
Scalable architecture that can grow with client needs
