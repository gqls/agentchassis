# Auth Service Internal API Documentation

This document describes the internal APIs and message formats used by the auth-service.

## Overview

The auth-service is the API gateway for the AI Persona Platform. It handles:
- User authentication and authorization
- Proxying requests to internal services
- WebSocket connections for real-time features
- User and subscription management

## Internal HTTP Endpoints

These endpoints are only accessible within the Kubernetes cluster and are not exposed externally.

### Internal Health Check
```
GET /internal/health
```

Returns detailed health information including dependency status.

**Response:**
```json
{
  "status": "healthy",
  "service": "auth-service",
  "version": "1.1.0",
  "dependencies": {
    "mysql": "healthy",
    "core-manager": "healthy"
  }
}
```

### Internal Metrics
```
GET /internal/metrics
```

Prometheus metrics endpoint for monitoring.

## Gateway Proxy Routes

The auth-service proxies certain requests to internal services:

### To Core Manager

#### Template Management (Admin Only)
- `GET /api/v1/templates` → `core-manager:8088/api/v1/templates`
- `POST /api/v1/templates` → `core-manager:8088/api/v1/templates`
- `GET /api/v1/templates/{id}` → `core-manager:8088/api/v1/templates/{id}`
- `PUT /api/v1/templates/{id}` → `core-manager:8088/api/v1/templates/{id}`
- `DELETE /api/v1/templates/{id}` → `core-manager:8088/api/v1/templates/{id}`

#### Instance Management
- `GET /api/v1/personas/instances` → `core-manager:8088/api/v1/personas/instances`
- `POST /api/v1/personas/instances` → `core-manager:8088/api/v1/personas/instances`
- `GET /api/v1/personas/instances/{id}` → `core-manager:8088/api/v1/personas/instances/{id}`
- `PUT /api/v1/personas/instances/{id}` → `core-manager:8088/api/v1/personas/instances/{id}`
- `DELETE /api/v1/personas/instances/{id}` → `core-manager:8088/api/v1/personas/instances/{id}`

### Headers Added by Gateway

When proxying requests, the auth-service adds these headers:

```http
X-User-ID: <user_id>
X-Client-ID: <client_id>
X-User-Role: <role>
X-User-Tier: <subscription_tier>
X-User-Email: <email>
X-User-Permissions: <comma-separated-permissions>
```

## WebSocket Protocol

### Connection
```
GET /ws
Authorization: Bearer <token>
```

### Message Format

#### Client to Server
```json
{
  "type": "message_type",
  "id": "unique_message_id",
  "data": {
    // message-specific data
  }
}
```

#### Server to Client
```json
{
  "type": "message_type",
  "id": "message_id",
  "correlation_id": "original_message_id",
  "data": {
    // response data
  },
  "error": {
    "code": "ERROR_CODE",
    "message": "Human readable error"
  }
}
```

### Message Types

#### Start Workflow
```json
{
  "type": "start_workflow",
  "id": "msg-123",
  "data": {
    "workflow_type": "content_generation",
    "agent_instance_id": "instance-uuid",
    "project_id": "project-uuid",
    "parameters": {
      // workflow-specific parameters
    }
  }
}
```

#### Workflow Status Update
```json
{
  "type": "workflow_status",
  "correlation_id": "msg-123",
  "data": {
    "status": "in_progress|completed|failed",
    "step": "current_step_name",
    "progress": 0.75,
    "result": {
      // step results
    }
  }
}
```

## Database Schema

The auth-service manages the following tables in the auth database (MySQL):

### users
```sql
CREATE TABLE users (
    id VARCHAR(36) PRIMARY KEY,
    email VARCHAR(255) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    role VARCHAR(50) DEFAULT 'user',
    client_id VARCHAR(100) NOT NULL,
    subscription_tier VARCHAR(50) DEFAULT 'free',
    is_active BOOLEAN DEFAULT true,
    email_verified BOOLEAN DEFAULT false,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    last_login_at TIMESTAMPTZ
);
```

### user_profiles
```sql
CREATE TABLE user_profiles (
    user_id VARCHAR(36) PRIMARY KEY REFERENCES users(id),
    first_name VARCHAR(100),
    last_name VARCHAR(100),
    company VARCHAR(255),
    phone VARCHAR(50),
    avatar_url VARCHAR(500),
    preferences JSON,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
```

### projects
```sql
CREATE TABLE projects (
    id VARCHAR(36) PRIMARY KEY,
    client_id VARCHAR(100) NOT NULL,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    owner_id VARCHAR(36) NOT NULL REFERENCES users(id),
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
```

### subscriptions
```sql
CREATE TABLE subscriptions (
    id VARCHAR(36) PRIMARY KEY,
    user_id VARCHAR(36) NOT NULL UNIQUE REFERENCES users(id),
    tier VARCHAR(50) NOT NULL,
    status VARCHAR(50) NOT NULL,
    start_date TIMESTAMPTZ NOT NULL,
    end_date TIMESTAMPTZ,
    trial_ends_at TIMESTAMPTZ,
    cancelled_at TIMESTAMPTZ,
    payment_method VARCHAR(100),
    stripe_customer_id VARCHAR(255),
    stripe_subscription_id VARCHAR(255),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
```

## Error Codes

The auth-service uses these standard error codes:

| Code | Description |
|------|-------------|
| AUTH001 | Invalid credentials |
| AUTH002 | Token expired |
| AUTH003 | Token invalid |
| AUTH004 | Insufficient permissions |
| USER001 | User not found |
| USER002 | User already exists |
| USER003 | Invalid password format |
| SUB001 | Subscription not found |
| SUB002 | Quota exceeded |
| GW001 | Upstream service unavailable |
| GW002 | Gateway timeout |

## Environment Variables

Required environment variables:

```bash
# Database
AUTH_DB_PASSWORD=<password>

# JWT
JWT_SECRET_KEY=<secret>

# Service URLs
CORE_MANAGER_URL=http://core-manager:8088

# Allowed Origins (CORS)
ALLOWED_ORIGINS=http://localhost:3000,https://app.personaplatform.com
```

## Integration Notes

### For Internal Services

When receiving requests from the auth-service gateway:
1. Trust the `X-User-*` headers for user context
2. Do NOT re-validate the JWT token
3. Use the provided user information for authorization decisions

### For Frontend Applications

1. Always include the JWT token in the Authorization header
2. Refresh tokens before they expire using the `/api/v1/auth/refresh` endpoint
3. Handle 401 responses by refreshing the token or redirecting to login
4. Use WebSocket for real-time updates on long-running operations