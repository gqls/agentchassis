# Core Manager Internal API Documentation

## Overview

The Core Manager service is responsible for managing persona templates and instances. It serves as the central management hub for all persona-related operations.

## Internal HTTP Endpoints

### Health Check
```
GET /internal/health
```

Returns the health status of the service.

**Response:**
```json
{
  "status": "healthy",
  "service": "core-manager",
  "version": "1.0.0",
  "database": "connected"
}
```

## External HTTP Endpoints (via Auth Service Gateway)

These endpoints are accessed through the auth-service gateway with authentication.

### Persona Templates

#### List Templates
```
GET /api/v1/templates
```

Returns all available persona templates.

**Required Role:** Admin

**Response:**
```json
{
  "templates": [
    {
      "id": "uuid",
      "name": "Copywriter",
      "description": "Creates compelling marketing copy",
      "category": "data-driven",
      "config": {
        "model": "claude-3-sonnet",
        "temperature": 0.7
      },
      "is_active": true,
      "created_at": "2024-01-01T00:00:00Z",
      "updated_at": "2024-01-01T00:00:00Z"
    }
  ],
  "count": 1
}
```

### Persona Instances

#### Create Instance
```
POST /api/v1/personas/instances
```

Creates a new persona instance from a template.

**Request:**
```json
{
  "template_id": "template-uuid",
  "name": "My Marketing Assistant",
  "project_id": "project-uuid",
  "config": {
    "custom_prompt": "Focus on B2B technology companies"
  }
}
```

**Response:**
```json
{
  "id": "instance-uuid",
  "template_id": "template-uuid",
  "name": "My Marketing Assistant",
  "owner_user_id": "user-uuid",
  "config": {
    "model": "claude-3-sonnet",
    "temperature": 0.7,
    "custom_prompt": "Focus on B2B technology companies"
  },
  "is_active": true,
  "created_at": "2024-01-01T00:00:00Z",
  "updated_at": "2024-01-01T00:00:00Z"
}
```

## Kafka Topics

### Consumed Topics

#### system.personas.commands
Commands for persona management operations.

**Message Format:**
```json
{
  "action": "create_instance|update_instance|delete_instance",
  "data": {
    // Action-specific data
  }
}
```

**Required Headers:**
- `correlation_id`: Unique request identifier
- `request_id`: Request tracking ID
- `client_id`: Client identifier
- `user_id`: User performing the action

### Produced Topics

#### system.personas.events
Events related to persona lifecycle.

**Message Format:**
```json
{
  "event": "instance_created|instance_updated|instance_deleted",
  "data": {
    "instance_id": "uuid",
    "template_id": "uuid",
    "user_id": "uuid",
    // Event-specific data
  },
  "timestamp": "2024-01-01T00:00:00Z"
}
```

## Database Schema

The Core Manager uses the `clients` PostgreSQL database with client-specific schemas.

### Global Tables

#### agent_definitions
```sql
CREATE TABLE agent_definitions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    type VARCHAR(100) NOT NULL UNIQUE,
    display_name VARCHAR(255) NOT NULL,
    description TEXT,
    category VARCHAR(50) NOT NULL,
    default_config JSONB NOT NULL DEFAULT '{}',
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
);
```

### Client-Specific Tables (in schema `client_{client_id}`)

#### agent_instances
```sql
CREATE TABLE agent_instances (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    template_id UUID NOT NULL,
    owner_user_id VARCHAR(255) NOT NULL,
    name VARCHAR(255) NOT NULL,
    config JSONB NOT NULL DEFAULT '{}',
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
```

## Environment Variables

```bash
# Database
CLIENTS_DB_HOST=postgres-clients
CLIENTS_DB_PORT=5432
CLIENTS_DB_NAME=clients_db
CLIENTS_DB_USER=clients_user
CLIENTS_DB_PASSWORD=<password>

# Templates Database
TEMPLATES_DB_HOST=postgres-templates
TEMPLATES_DB_PORT=5432
TEMPLATES_DB_NAME=templates_db
TEMPLATES_DB_USER=templates_user
TEMPLATES_DB_PASSWORD=<password>

# Kafka
KAFKA_BROKERS=kafka-0:9092,kafka-1:9092,kafka-2:9092

# Service
SERVICE_PORT=8088
LOG_LEVEL=info
```

## Integration Notes

### For Auth Service
- All requests should include user context headers
- Validate quota limits before creating instances
- Check user permissions for admin operations

### For Agent Services
- Subscribe to `system.personas.events` for instance lifecycle events
- Use instance configuration when processing requests
- Report usage metrics back to core-manager

### Error Handling
The service returns standard HTTP status codes:
- 200: Success
- 201: Created
- 400: Bad Request
- 401: Unauthorized
- 403: Forbidden (quota exceeded or insufficient permissions)
- 404: Not Found
- 500: Internal Server Error

Error responses follow the format:
```json
{
  "error": {
    "code": "PERSONA_001",
    "message": "Human-readable error message",
    "details": {
      // Additional context
    }
  }
}
```