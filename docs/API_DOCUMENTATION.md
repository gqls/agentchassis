# API Documentation Guide

This guide explains how to work with and maintain the API documentation for the AI Persona Platform.

## Overview

We use two types of API documentation:

1. **External API Documentation** (OpenAPI 3.0) - For customer-facing APIs
2. **Internal API Documentation** (Markdown) - For internal service communication

## External API Documentation

### Location
- OpenAPI Spec: `internal/auth-service/api/openapi.yaml`
- Swagger Annotations: Throughout handler files (`*_swagger.go`)

### Viewing Documentation

#### Option 1: Run Swagger UI Locally
```bash
# Start the documentation servers
docker-compose -f docker-compose.swagger.yml up -d

# Access the documentation:
# - Swagger UI: http://localhost:8082
# - Redoc: http://localhost:8083
# - Swagger Editor: http://localhost:8084
```

#### Option 2: Build and Run Auth Service
```bash
# Generate swagger docs
make swagger

# Run the auth service
go run cmd/auth-service/main.go

# Access swagger UI at http://localhost:8081/swagger/index.html
```

### Updating API Documentation

1. **Update OpenAPI Spec** (`internal/auth-service/api/openapi.yaml`):
    - Add new endpoints
    - Update request/response schemas
    - Add examples

2. **Update Swagger Annotations** in handler files:
   ```go
   // HandleNewEndpoint godoc
   // @Summary      Brief description
   // @Description  Detailed description
   // @Tags         Category
   // @Accept       json
   // @Produce      json
   // @Param        request body RequestType true "Description"
   // @Success      200 {object} ResponseType
   // @Failure      400 {object} ErrorResponse
   // @Router       /api/v1/endpoint [post]
   // @Security     BearerAuth
   ```

3. **Regenerate Documentation**:
   ```bash
   make swagger
   ```

4. **Validate OpenAPI Spec**:
   ```bash
   make validate-openapi
   ```

## Internal API Documentation

### Location
Each service has its own `API.md` file:
- `internal/auth-service/API.md`
- `internal/core-manager/API.md`
- `internal/agents/*/API.md`
- `internal/adapters/*/API.md`

### Format
Internal API docs should include:
- Service overview
- Internal endpoints (if HTTP service)
- Kafka topics and message formats (if Kafka service)
- Database schemas
- Environment variables
- Integration notes

### Template
```markdown
# [Service Name] Internal API Documentation

## Overview
Brief description of the service's purpose and responsibilities.

## Internal HTTP Endpoints (if applicable)
Document any HTTP endpoints not exposed through the gateway.

## Kafka Topics (if applicable)
### Consumed Topics
- Topic name: Description
  - Message format
  - Required headers

### Produced Topics
- Topic name: Description
  - Message format
  - Headers added

## Database Schema (if applicable)
SQL schemas for tables managed by this service.

## Environment Variables
List of required environment variables.

## Integration Notes
Important information for services integrating with this one.
```

## Best Practices

1. **Keep Documentation in Sync**
    - Update docs when changing APIs
    - Include documentation updates in PRs
    - Run validation in CI/CD

2. **Use Examples**
    - Include request/response examples
    - Show common use cases
    - Document error scenarios

3. **Version Your APIs**
    - Use `/api/v1/` prefix
    - Document breaking changes
    - Maintain backward compatibility

4. **Security Documentation**
    - Clearly mark which endpoints require authentication
    - Document required permissions/roles
    - Include rate limiting information

## CI/CD Integration

Add these steps to your CI/CD pipeline:

```yaml
# .github/workflows/api-docs.yml
name: API Documentation

on:
  pull_request:
    paths:
      - '**.go'
      - '**.yaml'
      - '**.yml'

jobs:
  validate:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      
      - name: Validate OpenAPI Spec
        run: |
          docker run --rm -v $PWD:/spec redocly/cli \
            lint /spec/internal/auth-service/api/openapi.yaml
      
      - name: Generate Swagger Docs
        run: |
          go install github.com/swaggo/swag/cmd/swag@latest
          make swagger
      
      - name: Check for uncommitted changes
        run: |
          git diff --exit-code
```

## Troubleshooting

### Common Issues

1. **Swagger UI not showing endpoints**
    - Ensure you've run `make swagger`
    - Check that annotations are properly formatted
    - Verify the import in `main.go`: `_ "github.com/gqls/agentchassis/cmd/auth-service/docs"`

2. **OpenAPI validation errors**
    - Use the Swagger Editor to debug
    - Check for missing required fields
    - Ensure all `$ref` paths are correct

3. **Annotations not picked up**
    - Annotations must be directly above the handler function
    - Use the exact format shown in examples
    - Run `swag init` with `--parseDependency` flag

## Tools and Resources

- [OpenAPI 3.0 Specification](https://swagger.io/specification/)
- [Swagger Editor](https://editor.swagger.io/)
- [Redocly CLI](https://redocly.com/docs/cli/)
- [swaggo/swag Documentation](https://github.com/swaggo/swag)
- [gin-swagger Documentation](https://github.com/swaggo/gin-swagger)

## API Documentation Checklist

Use this checklist when adding new endpoints:

- [ ] Add endpoint to `openapi.yaml`
- [ ] Add request/response schemas
- [ ] Add swagger annotations to handler
- [ ] Include authentication requirements
- [ ] Document error responses
- [ ] Add examples for complex requests
- [ ] Update internal `API.md` if needed
- [ ] Run `make swagger` to regenerate
- [ ] Run `make validate-openapi`
- [ ] Test in local Swagger UI