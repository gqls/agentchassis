# Gateway Service Documentation

## Overview

The Gateway service in the Auth Service acts as a reverse proxy, routing authenticated requests to the core-manager service. It enriches requests with user context and handles both HTTP and WebSocket connections.

## Architecture

```
Client → Auth Service (Gateway) → Core Manager
         ↓
    Authentication &
    Context Enrichment
```

## Key Features

### 1. Request Proxying
- Routes authenticated requests to core-manager
- Maintains request/response integrity
- Handles all HTTP methods (GET, POST, PUT, DELETE, PATCH)

### 2. Context Enrichment
All proxied requests are enriched with user context headers:
- `X-User-ID`: Authenticated user's ID
- `X-Client-ID`: Multi-tenant client identifier
- `X-User-Role`: User's role (user, admin, moderator)
- `X-User-Tier`: Subscription tier
- `X-User-Email`: User's email address
- `X-User-Permissions`: Comma-separated permissions list

### 3. WebSocket Proxy
- Upgrades HTTP connections to WebSocket
- Maintains bidirectional communication
- Forwards user context to core-manager

## Proxied Endpoints

### Template Management (Admin Only)
Templates define persona configurations and behaviors.

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/api/v1/templates` | List all templates |
| POST | `/api/v1/templates` | Create new template |
| GET | `/api/v1/templates/{id}` | Get template details |
| PUT | `/api/v1/templates/{id}` | Update template |
| DELETE | `/api/v1/templates/{id}` | Delete template |
| POST | `/api/v1/templates/{id}/clone` | Clone template |
| POST | `/api/v1/templates/{id}/validate` | Validate template configuration |

### Persona Instances
User-created instances based on templates.

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/api/v1/personas/instances` | List user's instances |
| POST | `/api/v1/personas/instances` | Create new instance |
| GET | `/api/v1/personas/instances/{id}` | Get instance details |
| PUT | `/api/v1/personas/instances/{id}` | Update instance |
| DELETE | `/api/v1/personas/instances/{id}` | Delete instance |
| POST | `/api/v1/personas/instances/{id}/execute` | Execute instance |
| GET | `/api/v1/personas/instances/{id}/history` | Get execution history |
| POST | `/api/v1/personas/instances/{id}/stop` | Stop running instance |
| GET | `/api/v1/personas/instances/{id}/logs` | Get instance logs |

### Admin Routes
Various administrative endpoints for system management.

#### Client Management
| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/api/v1/admin/clients` | List all clients |
| POST | `/api/v1/admin/clients` | Create new client |
| GET | `/api/v1/admin/clients/{id}` | Get client details |
| PUT | `/api/v1/admin/clients/{id}` | Update client |
| DELETE | `/api/v1/admin/clients/{id}` | Delete client |
| GET | `/api/v1/admin/clients/{id}/stats` | Get client statistics |

#### System Management
| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/api/v1/admin/system/health` | System health check |
| GET | `/api/v1/admin/system/metrics` | System metrics |
| GET | `/api/v1/admin/system/config` | Get configuration |
| PUT | `/api/v1/admin/system/config` | Update configuration |
| POST | `/api/v1/admin/system/maintenance` | Toggle maintenance mode |
| GET | `/api/v1/admin/system/logs` | Get system logs |

#### Workflow Management
| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/api/v1/admin/workflows` | List workflows |
| POST | `/api/v1/admin/workflows` | Create workflow |
| GET | `/api/v1/admin/workflows/{id}` | Get workflow |
| PUT | `/api/v1/admin/workflows/{id}` | Update workflow |
| DELETE | `/api/v1/admin/workflows/{id}` | Delete workflow |
| GET | `/api/v1/admin/workflows/{id}/runs` | Get workflow runs |

#### Agent Definitions
| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/api/v1/admin/agent-definitions` | List agent types |
| POST | `/api/v1/admin/agent-definitions` | Create agent type |
| GET | `/api/v1/admin/agent-definitions/{id}` | Get agent type |
| PUT | `/api/v1/admin/agent-definitions/{id}` | Update agent type |
| DELETE | `/api/v1/admin/agent-definitions/{id}` | Delete agent type |

## WebSocket Protocol

### Connection
```
GET /api/v1/ws
Upgrade: websocket
Authorization: Bearer <token>
```

### Message Format
```json
{
  "type": "command|event|response|error",
  "event": "event.name",
  "data": { "...": "..." },
  "id": "unique-id",
  "timestamp": "2024-07-17T14:30:00Z"
}
```

### Commands

#### Subscribe to Events
```json
{
  "type": "command",
  "event": "subscribe",
  "data": {
    "events": ["instance.status.*", "execution.complete"]
  },
  "id": "cmd_123"
}
```

#### Unsubscribe from Events
```json
{
  "type": "command",
  "event": "unsubscribe",
  "data": {
    "events": ["instance.status.*"]
  },
  "id": "cmd_124"
}
```

#### Keep-Alive Ping
```json
{
  "type": "command",
  "event": "ping",
  "id": "cmd_125"
}
```

### Event Types

#### Instance Events
- `instance.status.changed`: Status update
- `instance.created`: New instance created
- `instance.deleted`: Instance deleted
- `instance.updated`: Instance configuration updated

#### Execution Events
- `execution.started`: Execution began
- `execution.completed`: Execution finished successfully
- `execution.failed`: Execution failed
- `execution.progress`: Progress update

#### System Events
- `system.notification`: System-wide notification
- `system.maintenance`: Maintenance mode change
- `system.alert`: System alert

### Example Event
```json
{
  "type": "event",
  "event": "instance.status.changed",
  "data": {
    "instance_id": "inst_123",
    "previous_status": "running",
    "new_status": "completed",
    "reason": "Execution completed successfully"
  },
  "timestamp": "2024-07-17T14:30:00Z"
}
```

## Error Handling

### Gateway Errors
When the gateway itself encounters errors:

```json
{
  "error": "BAD_GATEWAY",
  "message": "Service temporarily unavailable",
  "status_code": 502,
  "service": "core-manager"
}
```

### Proxied Errors
Errors from core-manager are passed through unchanged.

### WebSocket Errors
```json
{
  "type": "error",
  "error": "Invalid command",
  "data": {
    "details": "Command 'foo' not recognized",
    "command": "foo"
  }
}
```

## Rate Limiting

The gateway respects rate limits from core-manager and adds headers:
- `X-RateLimit-Limit`: Request limit
- `X-RateLimit-Remaining`: Remaining requests
- `X-RateLimit-Reset`: Reset timestamp

## Security Considerations

1. **Authentication**: All requests must include valid JWT token
2. **Authorization**: Role-based access enforced before proxying
3. **Context Isolation**: User context passed via headers, not body
4. **Request Validation**: Basic validation before proxying
5. **Timeout Protection**: 30-second timeout on all proxied requests

## Performance

- **Connection Pooling**: Reuses HTTP connections to core-manager
- **Buffering**: Minimal buffering for streaming responses
- **WebSocket**: 1024-byte read/write buffers
- **Timeout**: 30-second default timeout

## Monitoring

The gateway logs:
- All proxied requests with latency
- Failed proxy attempts
- WebSocket connection lifecycle
- Authentication failures

## Configuration

Gateway configuration via environment/config:
```yaml
custom:
  core_manager_url: "http://core-manager:8080"
  request_timeout: 30
  enable_request_logging: true
  websocket_buffer_size: 1024
```