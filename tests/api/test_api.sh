#!/bin/bash
# OpenDeepWiki API Test Script

BASE_URL="http://localhost:8080/api"

echo "Testing OpenDeepWiki API..."
echo "============================="

# Submit a repository
echo "1. Submitting a repository..."
RESPONSE=$(curl -s -X POST "${BASE_URL}/warehouse/repos" \
  -H "Content-Type: application/json" \
  -d '{"git_url": "https://github.com/gin-gonic/gin.git"}')

echo "Response: $RESPONSE"

# Extract task ID from response
TASK_ID=$(echo $RESPONSE | grep -o '"task_id":"[^"]*"' | cut -d'"' -f4)

if [ -z "$TASK_ID" ]; then
  echo "Failed to get task ID. Exiting."
  exit 1
fi

echo "Task ID: $TASK_ID"
echo

# Wait a moment for the task to be processed
echo "Waiting for task processing to start..."
sleep 2

# Check task status
echo "2. Checking task status..."
TASK_STATUS=$(curl -s -X GET "${BASE_URL}/warehouse/tasks/$TASK_ID" \
  -H "Content-Type: application/json")

echo "Task Status: $TASK_STATUS"
echo

# Test invalid request
echo "3. Testing invalid request (missing git_url)..."
INVALID_RESPONSE=$(curl -s -X POST "${BASE_URL}/warehouse/repos" \
  -H "Content-Type: application/json" \
  -d '{}')

echo "Response: $INVALID_RESPONSE"
echo

# Test non-existent task
echo "4. Testing non-existent task..."
NON_EXISTENT=$(curl -s -X GET "${BASE_URL}/warehouse/tasks/non_existent_task" \
  -H "Content-Type: application/json")

echo "Response: $NON_EXISTENT"
echo

echo "============================="
echo "API testing complete!" 