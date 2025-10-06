#!/bin/bash

# N8n Parallels Server API Test Script

BASE_URL="http://localhost:8080"

echo "Testing N8n Parallels Server API..."
echo "=================================="

# Test 1: Health Check
echo "1. Testing health check endpoint..."
echo "GET $BASE_URL/health"
curl -s -X GET "$BASE_URL/health" | jq .
echo -e "\n"

# Test 2: Basic parallel execution with httpbin.org
echo "2. Testing basic parallel execution..."
echo "POST $BASE_URL/v1/parallels/execute"
curl -s -X POST "$BASE_URL/v1/parallels/execute" \
  -H "Content-Type: application/json" \
  -d '{
    "webhook_url": "https://httpbin.org/post",
    "payloads": [
      {"id": 1, "name": "Alice", "action": "test1"},
      {"id": 2, "name": "Bob", "action": "test2"},
      {"id": 3, "name": "Charlie", "action": "test3"}
    ],
    "timeout": 30
  }' | jq .
echo -e "\n"

# Test 3: Test with authentication header
echo "3. Testing with authentication header..."
curl -s -X POST "$BASE_URL/v1/parallels/execute" \
  -H "Content-Type: application/json" \
  -d '{
    "webhook_url": "https://httpbin.org/post",
    "auth_header": "Bearer test-token-123",
    "payloads": [
      {"user_id": 100, "operation": "create"},
      {"user_id": 200, "operation": "update"}
    ],
    "timeout": 15
  }' | jq .
echo -e "\n"

# Test 4: Test error handling with invalid URL
echo "4. Testing error handling with invalid URL..."
curl -s -X POST "$BASE_URL/v1/parallels/execute" \
  -H "Content-Type: application/json" \
  -d '{
    "webhook_url": "https://invalid-url-that-does-not-exist.com/webhook",
    "payloads": [
      {"test": "error_handling"}
    ],
    "timeout": 5
  }' | jq .
echo -e "\n"

# Test 5: Test validation with empty payloads
echo "5. Testing validation with empty payloads..."
curl -s -X POST "$BASE_URL/v1/parallels/execute" \
  -H "Content-Type: application/json" \
  -d '{
    "webhook_url": "https://httpbin.org/post",
    "payloads": [],
    "timeout": 30
  }' | jq .
echo -e "\n"

# Test 6: Test validation with missing webhook_url
echo "6. Testing validation with missing webhook_url..."
curl -s -X POST "$BASE_URL/v1/parallels/execute" \
  -H "Content-Type: application/json" \
  -d '{
    "payloads": [
      {"test": "validation"}
    ],
    "timeout": 30
  }' | jq .
echo -e "\n"

echo "API testing complete!"
echo "===================="
echo ""
echo "To run this test:"
echo "1. Start the server: ./n8n-parallels"
echo "2. Run this script in another terminal: ./test_api.sh"
echo "3. Make sure you have 'jq' installed for JSON formatting"