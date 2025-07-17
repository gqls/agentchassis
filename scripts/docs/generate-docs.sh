#!/bin/bash
# Script to generate and validate API documentation

set -e

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
swag init -g cmd/auth-service/main.go -o cmd/auth-service/docs --parseDependency --parseInternal

if [ $? -eq 0 ]; then
    echo -e "${GREEN}✓ Swagger documentation generated successfully${NC}"
else
    echo -e "${RED}✗ Failed to generate Swagger documentation${NC}"
    exit 1
fi

# Validate OpenAPI specification
echo -e "\n${YELLOW}Validating OpenAPI specification...${NC}"
docker run --rm -v "${PWD}":/spec redocly/cli lint /spec/internal/auth-service/api/openapi.yaml

if [ $? -eq 0 ]; then
    echo -e "${GREEN}✓ OpenAPI specification is valid${NC}"
else
    echo -e "${RED}✗ OpenAPI specification validation failed${NC}"
    exit 1
fi

# Generate HTML documentation
echo -e "\n${YELLOW}Generating HTML documentation...${NC}"
mkdir -p docs/api
docker run --rm -v "${PWD}":/spec redocly/cli build-docs /spec/internal/auth-service/api/openapi.yaml -o /spec/docs/api/reference.html

if [ $? -eq 0 ]; then
    echo -e "${GREEN}✓ HTML documentation generated at docs/api/reference.html${NC}"
else
    echo -e "${YELLOW}⚠ Failed to generate HTML documentation${NC}"
fi

# Start documentation servers
echo -e "\n${YELLOW}Do you want to start the documentation servers? (y/n)${NC}"
read -r response

if [[ "$response" =~ ^([yY][eE][sS]|[yY])$ ]]; then
    echo -e "${YELLOW}Starting documentation servers...${NC}"
    docker-compose -f docker-compose.swagger.yml up -d

    echo -e "\n${GREEN}Documentation servers started:${NC}"
    echo -e "  • Swagger UI: ${GREEN}http://localhost:8082${NC}"
    echo -e "  • Redoc: ${GREEN}http://localhost:8083${NC}"
    echo -e "  • Swagger Editor: ${GREEN}http://localhost:8084${NC}"
    echo -e "\n${YELLOW}To stop the servers, run: docker-compose -f docker-compose.swagger.yml down${NC}"
fi

echo -e "\n${GREEN}✅ Documentation generation complete!${NC}"

# Summary
echo -e "\n${YELLOW}Summary:${NC}"
echo "• Swagger docs: cmd/auth-service/docs/"
echo "• OpenAPI spec: internal/auth-service/api/openapi.yaml"
echo "• HTML docs: docs/api/reference.html"
echo "• Internal docs: internal/*/API.md"

echo -e "\n${YELLOW}Next steps:${NC}"
echo "1. Review the generated documentation"
echo "2. Update any missing descriptions or examples"
echo "3. Commit the changes to your repository"
echo "4. Run 'make swagger' to regenerate after changes"