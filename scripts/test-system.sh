#!/bin/bash
# Test script to verify the system is working

set -e

echo "🧪 Testing AI Persona System"
echo "=========================="

# Colors
GREEN='\033[0;32m'
RED='\033[0;31m'
NC='\033[0m'

# Test health endpoints
test_health() {
    local service=$1
    local port=$2

    echo -n "Testing $service health... "

    if curl -s -f http://localhost:$port/health > /dev/null; then
        echo -e "${GREEN}✓${NC}"
        return 0
    else
        echo -e "${RED}✗${NC}"
        return 1
    fi
}

# Test auth flow
test_auth() {
    echo -n "Testing authentication flow... "

    # Register a test user
    REGISTER_RESPONSE=$(curl -s -X POST http://localhost:8081/api/v1/auth/register \
        -H "Content-Type: application/json" \
        -d '{
            "email": "test@example.com",
            "password": "TestPass123!",
            "client_id": "test-client",
            "first_name": "Test",
            "last_name": "User"
        }')

    # Extract token
    TOKEN=$(echo $REGISTER_RESPONSE | jq -r '.access_token')

    if [ "$TOKEN" != "null" ] && [ ! -z "$TOKEN" ]; then
        echo -e "${GREEN}✓${NC}"
        echo "  Token: ${TOKEN:0:20}..."
        return 0
    else
        echo -e "${RED}✗${NC}"
        echo "  Response: $REGISTER_RESPONSE"
        return 1
    fi
}

# Test template listing
test_templates() {
    echo -n "Testing template listing... "

    # Assuming we have a token from previous test
    TEMPLATES=$(curl -s -X GET http://localhost:8088/api/v1/templates \
        -H "Authorization: Bearer $TOKEN")

    if echo $TEMPLATES | jq -e '.templates' > /dev/null; then
        echo -e "${GREEN}✓${NC}"
        TEMPLATE_COUNT=$(echo $TEMPLATES | jq '.templates | length')
        echo "  Found $TEMPLATE_COUNT templates"
        return 0
    else
        echo -e "${RED}✗${NC}"
        return 1
    fi
}

# Main test flow
echo ""
echo "1. Testing service health endpoints:"
test_health "Auth Service" 8081
test_health "Core Manager" 8088

echo ""
echo "2. Testing authentication:"
test_auth

echo ""
echo "3. Testing API endpoints:"
test_templates

echo ""
echo "✅ Basic system tests complete!"
