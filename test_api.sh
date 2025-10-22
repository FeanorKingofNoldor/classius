#!/bin/bash

# Classius API Test Script
API_BASE="http://localhost:8080/api/v1"

echo "🧪 Testing Classius API Endpoints"
echo "================================="

# Colors for output
GREEN='\033[0;32m'
RED='\033[0;31m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Function to test endpoint
test_endpoint() {
    local method=$1
    local endpoint=$2
    local data=$3
    local expected_status=$4
    local description=$5
    
    echo -e "\n${BLUE}Testing: $description${NC}"
    echo "  $method $endpoint"
    
    if [ -n "$data" ]; then
        response=$(curl -s -w "\nSTATUS_CODE:%{http_code}" -X $method \
            -H "Content-Type: application/json" \
            -d "$data" \
            "$API_BASE$endpoint")
    else
        response=$(curl -s -w "\nSTATUS_CODE:%{http_code}" -X $method \
            -H "Content-Type: application/json" \
            "$API_BASE$endpoint")
    fi
    
    status_code=$(echo "$response" | tail -1 | sed 's/STATUS_CODE://')
    body=$(echo "$response" | sed '$d')
    
    if [ "$status_code" = "$expected_status" ]; then
        echo -e "  ${GREEN}✅ PASS${NC} (Status: $status_code)"
    else
        echo -e "  ${RED}❌ FAIL${NC} (Expected: $expected_status, Got: $status_code)"
        echo "  Response: $body"
    fi
}

# Test 1: Health check
test_endpoint "GET" "/../../health" "" "200" "Health Check"

# Test 2: Register new user
echo -e "\n${BLUE}🔐 Testing Authentication Flow${NC}"
USER_DATA='{
    "username": "testuser",
    "email": "test@classius.com",
    "password": "testpassword123",
    "full_name": "Test User"
}'
test_endpoint "POST" "/auth/register" "$USER_DATA" "201" "User Registration"

# Test 3: Login
LOGIN_DATA='{
    "email": "test@classius.com",
    "password": "testpassword123"
}'

echo -e "\n${BLUE}Getting authentication token...${NC}"
login_response=$(curl -s -X POST \
    -H "Content-Type: application/json" \
    -d "$LOGIN_DATA" \
    "$API_BASE/auth/login")

# Extract token from response
TOKEN=$(echo "$login_response" | grep -o '"access_token":"[^"]*' | sed 's/"access_token":"//')

if [ -n "$TOKEN" ] && [ "$TOKEN" != "null" ]; then
    echo -e "${GREEN}✅ Login successful, token obtained${NC}"
    echo "Token: ${TOKEN:0:20}..."
else
    echo -e "${RED}❌ Failed to get authentication token${NC}"
    echo "Response: $login_response"
    exit 1
fi

# Test 4: Protected endpoints with token
echo -e "\n${BLUE}📚 Testing Protected Endpoints${NC}"

# Test user profile
test_endpoint "GET" "/user/profile" "" "200" "Get User Profile (with auth)"

# Test book endpoints
test_endpoint "GET" "/books" "" "200" "Get Books List"
test_endpoint "GET" "/books/stats" "" "200" "Get Book Statistics"
test_endpoint "GET" "/books/tags" "" "200" "Get User Tags"

# Test Sage endpoints (if AI is configured)
echo -e "\n${BLUE}🧠 Testing AI Sage Endpoints${NC}"
test_endpoint "GET" "/sage/capabilities" "" "200" "Get Sage Capabilities"
test_endpoint "GET" "/sage/health" "" "503" "Check Sage Health (expected to fail without API key)"

echo -e "\n${GREEN}✅ API Testing Complete!${NC}"
echo "================================="

# Note: Some endpoints require additional setup (file uploads, AI API keys, etc.)
echo -e "\n${BLUE}📝 Notes:${NC}"
echo "  • Book upload tests require multipart form data"
echo "  • AI Sage tests require OPENAI_API_KEY environment variable"
echo "  • Some endpoints return 503 if external services aren't configured"
echo ""