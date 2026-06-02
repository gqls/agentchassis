#!/bin/bash
# Script to generate and validate API documentation

set -e

# Get the project root directory (assuming script is in scripts/docs/)
SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
PROJECT_ROOT="$( cd "$SCRIPT_DIR/../.." && pwd )"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo -e "${GREEN}🚀 AI Persona Platform - API Documentation Generator${NC}"
echo "=================================================="

# Function to check if command exists
command_exists() {
    command -v "$1" >/dev/null 2>&1
}

# Change to project root
cd "$PROJECT_ROOT"

# Check prerequisites
echo -e "\n${YELLOW}Checking prerequisites...${NC}"

if ! command_exists swag; then
    echo -e "${YELLOW}Installing swag...${NC}"
    go install github.com/swaggo/swag/cmd/swag@latest
fi

if ! command_exists docker; then
    echo -e "${RED}Error: Docker is required but not installed.${NC}"
    exit 1
fi

# Generate Swagger documentation
echo -e "\n${YELLOW}Generating Swagger documentation...${NC}"

# First, let's add swagger annotations to main.go if they don't exist
MAIN_FILE="cmd/auth-service/main.go"
if ! grep -q "@title" "$MAIN_FILE"; then
    echo -e "${YELLOW}Adding Swagger annotations to main.go...${NC}"
    # Create a temporary file with the annotations
    cat > /tmp/swagger_annotations.go << 'EOF'
// @title Auth Service API
// @version 1.0
// @description Authentication and authorization service for the AI Persona Platform
// @termsOfService http://swagger.io/terms/

// @contact.name AI Persona Support
// @contact.email support@persona-platform.com

// @license.name Proprietary

// @host localhost:8081
// @BasePath /api/v1

// @securityDefinitions.apikey Bearer
// @in header
// @name Authorization
// @description Type "Bearer" followed by a space and JWT token.

EOF
    
    # Add annotations before package main
    sed -i '/^package main$/i \
// @title Auth Service API\
// @version 1.0\
// @description Authentication and authorization service for the AI Persona Platform\
// @termsOfService http://swagger.io/terms/\
\
// @contact.name AI Persona Support\
// @contact.email support@persona-platform.com\
\
// @license.name Proprietary\
\
// @host localhost:8081\
// @BasePath /api/v1\
\
// @securityDefinitions.apikey Bearer\
// @in header\
// @name Authorization\
// @description Type "Bearer" followed by a space and JWT token.\
' "$MAIN_FILE"
fi

# Generate swagger docs
swag init -g cmd/auth-service/main.go -o cmd/auth-service/docs --parseDependency --parseInternal

if [ $? -eq 0 ]; then
    echo -e "${GREEN}✓ Swagger documentation generated successfully${NC}"
else
    echo -e "${RED}✗ Failed to generate Swagger documentation${NC}"
    exit 1
fi

# Check if OpenAPI spec exists
OPENAPI_SPEC="internal/auth-service/api/openapi.yaml"
if [ -f "$OPENAPI_SPEC" ]; then
    echo -e "\n${YELLOW}Validating OpenAPI specification...${NC}"
    docker run --rm -v "${PWD}":/spec redocly/cli lint "/spec/${OPENAPI_SPEC}"
    
    if [ $? -eq 0 ]; then
        echo -e "${GREEN}✓ OpenAPI specification is valid${NC}"
    else
        echo -e "${RED}✗ OpenAPI specification validation failed${NC}"
    fi
    
    # Generate HTML documentation
    echo -e "\n${YELLOW}Generating HTML documentation...${NC}"
    mkdir -p docs/api
    docker run --rm -v "${PWD}":/spec redocly/cli build-docs "/spec/${OPENAPI_SPEC}" -o /spec/docs/api/reference.html
    
    if [ $? -eq 0 ]; then
        echo -e "${GREEN}✓ HTML documentation generated at docs/api/reference.html${NC}"
    else
        echo -e "${YELLOW}⚠ Failed to generate HTML documentation${NC}"
    fi
else
    echo -e "${YELLOW}⚠ OpenAPI spec not found at ${OPENAPI_SPEC}${NC}"
    echo -e "${YELLOW}  Using generated Swagger spec instead${NC}"
fi

# Check if docker-compose.swagger.yaml exists
SWAGGER_COMPOSE_FILE="deployments/docker-compose/docker-compose.swagger.yaml"
if [ ! -f "$SWAGGER_COMPOSE_FILE" ]; then
    echo -e "\n${YELLOW}Creating docker-compose.swagger.yaml...${NC}"
    mkdir -p "$(dirname "$SWAGGER_COMPOSE_FILE")"
    cat > "$SWAGGER_COMPOSE_FILE" << 'EOF'
version: '3.8'

services:
  swagger-ui:
    image: swaggerapi/swagger-ui:latest
    container_name: swagger-ui
    ports:
      - "8082:8080"
    environment:
      - SWAGGER_JSON=/api/swagger.json
    volumes:
      - ../../cmd/auth-service/docs:/api
    networks:
      - docs-network

  redoc:
    image: redocly/redoc:latest
    container_name: redoc
    ports:
      - "8083:80"
    environment:
      - SPEC_URL=/api/swagger.json
    volumes:
      - ../../cmd/auth-service/docs:/usr/share/nginx/html/api
    networks:
      - docs-network

  swagger-editor:
    image: swaggerapi/swagger-editor:latest
    container_name: swagger-editor
    ports:
      - "8084:8080"
    environment:
      - SWAGGER_FILE=/api/swagger.json
    volumes:
      - ../../cmd/auth-service/docs:/api
    networks:
      - docs-network

networks:
  docs-network:
    driver: bridge
EOF
    echo -e "${GREEN}✓ Created docker-compose.swagger.yaml${NC}"
fi

# Start documentation servers
echo -e "\n${YELLOW}Do you want to start the documentation servers? (y/n)${NC}"
read -r response

if [[ "$response" =~ ^([yY][eE][sS]|[yY])$ ]]; then
    echo -e "${YELLOW}Starting documentation servers...${NC}"
    docker-compose -f "$SWAGGER_COMPOSE_FILE" up -d

    echo -e "\n${GREEN}Documentation servers started:${NC}"
    echo -e "  • Swagger UI: ${GREEN}http://localhost:8082${NC}"
    echo -e "  • Redoc: ${GREEN}http://localhost:8083${NC}"
    echo -e "  • Swagger Editor: ${GREEN}http://localhost:8084${NC}"
    echo -e "\n${YELLOW}To stop the servers, run: docker-compose -f ${SWAGGER_COMPOSE_FILE} down${NC}"
fi

echo -e "\n${GREEN}✅ Documentation generation complete!${NC}"

# Summary
echo -e "\n${YELLOW}Summary:${NC}"
echo "• Swagger docs: cmd/auth-service/docs/"
if [ -f "$OPENAPI_SPEC" ]; then
    echo "• OpenAPI spec: ${OPENAPI_SPEC}"
fi
echo "• HTML docs: docs/api/reference.html"
echo "• Internal docs: internal/*/API.md"

echo -e "\n${YELLOW}Next steps:${NC}"
echo "1. Review the generated documentation"
echo "2. Update any missing descriptions or examples"
echo "3. Commit the changes to your repository"
echo "4. Run 'make swagger' to regenerate after changes"