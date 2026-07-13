# Auth Service API Endpoints Summary

This document provides a quick reference for all available API endpoints in the Auth Service.

## Base URL
- Local Development: `http://localhost:8081`
- Production: `https://api.persona-platform.com`

## Authentication
Most endpoints require a Bearer token in the Authorization header:
```
Authorization: Bearer <your-jwt-token>
```

## Endpoints by Category

### 🔐 Authentication (`/api/v1/auth`)
| Method | Endpoint | Description | Auth Required |
|--------|----------|-------------|---------------|
| POST | `/auth/register` | Register a new user | No |
| POST | `/auth/login` | Login with email/password | No |
| POST | `/auth/refresh` | Refresh access token | No |
| POST | `/auth/validate` | Validate access token | No |
| POST | `/auth/logout` | Logout and invalidate token | Yes |

### 👤 User Management (`/api/v1/user`)
| Method | Endpoint | Description | Auth Required |
|--------|----------|-------------|---------------|
| GET | `/user/profile` | Get current user profile | Yes |
| PUT | `/user/profile` | Update user profile | Yes |
| POST | `/user/password` | Change password | Yes |
| DELETE | `/user/delete` | Delete account | Yes |

### 💳 Subscription (`/api/v1/subscription`)
| Method | Endpoint | Description | Auth Required |
|--------|----------|-------------|---------------|
| GET | `/subscription` | Get current subscription | Yes |
| GET | `/subscription/usage` | Get usage statistics | Yes |
| GET | `/subscription/check-quota` | Check resource quota | Yes |

### 📁 Projects (`/api/v1/projects`)
| Method | Endpoint | Description | Auth Required |
|--------|----------|-------------|---------------|
| GET | `/projects` | List all projects | Yes |
| POST | `/projects` | Create new project | Yes |
| GET | `/projects/{id}` | Get project details | Yes |
| PUT | `/projects/{id}` | Update project | Yes |
| DELETE | `/projects/{id}` | Delete project | Yes |

### 🛡️ Admin - Users (`/api/v1/admin`)
| Method | Endpoint | Description | Auth Required | Role |
|--------|----------|-------------|---------------|------|
| GET | `/admin/users` | List all users | Yes | Admin |
| GET | `/admin/users/{user_id}` | Get user details | Yes | Admin |
| PUT | `/admin/users/{user_id}` | Update user | Yes | Admin |
| DELETE | `/admin/users/{user_id}` | Delete user | Yes | Admin |
| GET | `/admin/users/{user_id}/activity` | Get user activity | Yes | Admin |
| POST | `/admin/users/{user_id}/permissions` | Grant permission | Yes | Admin |
| DELETE | `/admin/users/{user_id}/permissions/{permission_name}` | Revoke permission | Yes | Admin |

### 🛡️ Admin - Subscriptions (`/api/v1/admin`)
| Method | Endpoint | Description | Auth Required | Role |
|--------|----------|-------------|---------------|------|
| GET | `/admin/subscriptions` | List all subscriptions | Yes | Admin |
| POST | `/admin/subscriptions` | Create subscription | Yes | Admin |
| PUT | `/admin/subscriptions/{user_id}` | Update subscription | Yes | Admin |

### 🔧 System
| Method | Endpoint | Description | Auth Required |
|--------|----------|-------------|---------------|
| GET | `/health` | Health check | No |
| GET | `/ws` | WebSocket connection | Yes |

## Common Response Formats

### Success Response
```json
{
  "data": { ... },
  "message": "Operation successful"
}
```

### Error Response
```json
{
  "error": "ERROR_CODE",
  "message": "Human readable error message",
  "details": { ... }
}
```

### Pagination Response
```json
{
  "data": [ ... ],
  "pagination": {
    "page": 1,
    "limit": 20,
    "total": 100,
    "total_pages": 5
  }
}
```

## Status Codes
- `200 OK` - Request successful
- `201 Created` - Resource created successfully
- `400 Bad Request` - Invalid request data
- `401 Unauthorized` - Authentication required or invalid token
- `403 Forbidden` - Access denied (insufficient permissions)
- `404 Not Found` - Resource not found
- `409 Conflict` - Resource already exists
- `500 Internal Server Error` - Server error

## Rate Limiting
API requests are rate-limited based on subscription tier:
- Free: 60 requests/minute
- Basic: 300 requests/minute
- Premium: 1000 requests/minute
- Enterprise: Unlimited

Rate limit headers are included in responses:
- `X-RateLimit-Limit`: Maximum requests allowed
- `X-RateLimit-Remaining`: Requests remaining
- `X-RateLimit-Reset`: Timestamp when limit resets