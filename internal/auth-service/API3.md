# Auth Service Complete API Documentation

## Overview

The Auth Service provides comprehensive authentication, authorization, user management, and subscription handling for the AI Persona Platform. This document covers all available endpoints including standard user operations and administrative functions.

## API Structure

### Base URLs
- **Local Development**: `http://localhost:8081/api/v1`
- **Staging**: `https://staging-api.persona-platform.com/api/v1`
- **Production**: `https://api.persona-platform.com/api/v1`

### Authentication
Most endpoints require Bearer token authentication:
```
Authorization: Bearer <jwt-token>
```

### Response Format
All responses follow a consistent format:

**Success Response**:
```json
{
  "data": { ... },
  "message": "Operation successful"
}
```

**Error Response**:
```json
{
  "error": "ERROR_CODE",
  "message": "Human-readable error message",
  "details": { ... }
}
```

## Endpoint Categories

### 1. Authentication Endpoints (`/auth`)

Core authentication functionality for user registration, login, and token management.

| Method | Endpoint | Description | Auth | Role |
|--------|----------|-------------|------|------|
| POST | `/auth/register` | Register new user | No | - |
| POST | `/auth/login` | User login | No | - |
| POST | `/auth/refresh` | Refresh access token | No | - |
| POST | `/auth/validate` | Validate access token | No | - |
| POST | `/auth/logout` | Logout user | Yes | Any |

### 2. User Management Endpoints (`/user`)

User profile management and account operations.

| Method | Endpoint | Description | Auth | Role |
|--------|----------|-------------|------|------|
| GET | `/user/profile` | Get current user profile | Yes | Any |
| PUT | `/user/profile` | Update user profile | Yes | Any |
| POST | `/user/password` | Change password | Yes | Any |
| DELETE | `/user/delete` | Delete account | Yes | Any |

### 3. Subscription Endpoints (`/subscription`)

Subscription and usage management for users.

| Method | Endpoint | Description | Auth | Role |
|--------|----------|-------------|------|------|
| GET | `/subscription` | Get current subscription | Yes | Any |
| GET | `/subscription/usage` | Get usage statistics | Yes | Any |
| GET | `/subscription/check-quota` | Check resource quota | Yes | Any |

### 4. Project Management Endpoints (`/projects`)

Project creation and management.

| Method | Endpoint | Description | Auth | Role |
|--------|----------|-------------|------|------|
| GET | `/projects` | List all projects | Yes | Any |
| POST | `/projects` | Create new project | Yes | Any |
| GET | `/projects/{id}` | Get project details | Yes | Any |
| PUT | `/projects/{id}` | Update project | Yes | Any |
| DELETE | `/projects/{id}` | Delete project | Yes | Any |

### 5. Admin - User Management (`/admin/users`)

Administrative endpoints for managing users.

| Method | Endpoint | Description | Auth | Role |
|--------|----------|-------------|------|------|
| GET | `/admin/users` | List all users with filters | Yes | Admin |
| GET | `/admin/users/{user_id}` | Get user details with stats | Yes | Admin |
| PUT | `/admin/users/{user_id}` | Update user (role, tier, status) | Yes | Admin |
| DELETE | `/admin/users/{user_id}` | Delete user account | Yes | Admin |
| GET | `/admin/users/{user_id}/activity` | Get user activity logs | Yes | Admin |
| POST | `/admin/users/{user_id}/permissions` | Grant permission | Yes | Admin |
| DELETE | `/admin/users/{user_id}/permissions/{name}` | Revoke permission | Yes | Admin |

### 6. Admin - Bulk Operations (`/admin/users`)

Bulk user management operations.

| Method | Endpoint | Description | Auth | Role |
|--------|----------|-------------|------|------|
| POST | `/admin/users/bulk` | Bulk user operations | Yes | Admin |
| POST | `/admin/users/export` | Export user data (CSV/JSON) | Yes | Admin |
| POST | `/admin/users/import` | Import users from CSV | Yes | Admin |

### 7. Admin - Session Management (`/admin/users`)

Session and security management.

| Method | Endpoint | Description | Auth | Role |
|--------|----------|-------------|------|------|
| GET | `/admin/users/{user_id}/sessions` | Get user sessions | Yes | Admin |
| DELETE | `/admin/users/{user_id}/sessions` | Terminate all sessions | Yes | Admin |
| POST | `/admin/users/{user_id}/password` | Reset user password | Yes | Admin |
| GET | `/admin/users/{user_id}/audit` | Get audit log | Yes | Admin |

### 8. Admin - Subscription Management (`/admin/subscriptions`)

Administrative subscription operations.

| Method | Endpoint | Description | Auth | Role |
|--------|----------|-------------|------|------|
| GET | `/admin/subscriptions` | List all subscriptions | Yes | Admin |
| POST | `/admin/subscriptions` | Create subscription | Yes | Admin |
| PUT | `/admin/subscriptions/{user_id}` | Update subscription | Yes | Admin |

### 9. System Endpoints

System health and real-time communication.

| Method | Endpoint | Description | Auth | Role |
|--------|----------|-------------|------|------|
| GET | `/health` | Health check | No | - |
| GET | `/ws` | WebSocket connection | Yes | Any |

## Detailed Endpoint Documentation

### Authentication Flow

1. **Registration**: POST `/auth/register`
    - Creates new user account
    - Returns JWT tokens
    - Supports multi-tenant via `client_id`

2. **Login**: POST `/auth/login`
    - Authenticates with email/password
    - Returns access and refresh tokens
    - Tokens expire after configured duration

3. **Token Refresh**: POST `/auth/refresh`
    - Uses refresh token to get new access token
    - Maintains user session continuity

### User Roles and Permissions

**Roles**:
- `user`: Standard user access
- `moderator`: Extended permissions
- `admin`: Full administrative access

**Permissions** (examples):
- `read:users`: View user information
- `write:users`: Modify user information
- `manage:subscriptions`: Manage subscriptions
- `system:admin`: System administration

### Subscription Tiers

- **Free**: Basic access, limited resources
- **Basic**: Enhanced limits, standard support
- **Premium**: High limits, priority support
- **Enterprise**: Unlimited resources, dedicated support

### Rate Limiting

API requests are rate-limited based on subscription tier:

| Tier | Requests/Minute | Requests/Hour | Requests/Day |
|------|----------------|---------------|--------------|
| Free | 60 | 1,000 | 10,000 |
| Basic | 300 | 5,000 | 50,000 |
| Premium | 1,000 | 20,000 | 200,000 |
| Enterprise | Unlimited | Unlimited | Unlimited |

### Error Codes

Common error codes returned by the API:

| Code | Description |
|------|-------------|
| `INVALID_CREDENTIALS` | Invalid login credentials |
| `TOKEN_EXPIRED` | JWT token has expired |
| `INSUFFICIENT_PERMISSIONS` | User lacks required permissions |
| `RESOURCE_NOT_FOUND` | Requested resource doesn't exist |
| `QUOTA_EXCEEDED` | Usage quota exceeded |
| `VALIDATION_ERROR` | Request validation failed |
| `INTERNAL_ERROR` | Internal server error |

### Webhook Events

The system can send webhooks for the following events:

- `user.created`: New user registration
- `user.updated`: User profile updated
- `user.deleted`: User account deleted
- `subscription.created`: New subscription
- `subscription.updated`: Subscription changed
- `subscription.cancelled`: Subscription cancelled
- `security.suspicious_activity`: Suspicious activity detected

## Security Considerations

1. **Token Security**:
    - Access tokens expire in 1 hour
    - Refresh tokens expire in 30 days
    - Tokens are invalidated on logout

2. **Password Requirements**:
    - Minimum 8 characters
    - Must contain uppercase, lowercase, number
    - Checked against common passwords

3. **Session Management**:
    - Sessions tracked per device
    - Admins can terminate sessions
    - Automatic timeout after inactivity

4. **Audit Logging**:
    - All admin actions logged
    - User activities tracked
    - IP addresses recorded

## Best Practices

1. **Authentication**:
    - Store tokens securely
    - Refresh tokens before expiry
    - Implement proper logout

2. **Error Handling**:
    - Check response status codes
    - Parse error messages
    - Implement retry logic

3. **Rate Limiting**:
    - Monitor rate limit headers
    - Implement backoff strategies
    - Cache responses when possible

4. **Data Security**:
    - Use HTTPS for all requests
    - Validate input data
    - Sanitize user content

## Support

For additional support or questions:
- Documentation: https://docs.persona-platform.com
- Support: support@persona-platform.com
- Status: https://status.persona-platform.com