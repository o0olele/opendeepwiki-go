# OpenDeepWiki API Test Script (PowerShell)

$BaseUrl = "http://localhost:8080/api"

Write-Host "Testing OpenDeepWiki API..."
Write-Host "============================="

# Submit a repository
Write-Host "1. Submitting a repository..."
$repoPayload = @{
    git_url = "https://github.com/gin-gonic/gin.git"
} | ConvertTo-Json

$response = Invoke-RestMethod -Uri "$BaseUrl/warehouse/repos" -Method Post -ContentType "application/json" -Body $repoPayload -ErrorAction SilentlyContinue
if ($response) {
    Write-Host "Response:" ($response | ConvertTo-Json)
    $taskId = $response.task_id
    Write-Host "Task ID: $taskId"
} else {
    Write-Host "Failed to submit repository."
    exit 1
}

Write-Host ""

# Wait a moment for the task to be processed
Write-Host "Waiting for task processing to start..."
Start-Sleep -Seconds 2

# Check task status
Write-Host "2. Checking task status..."
try {
    $taskStatus = Invoke-RestMethod -Uri "$BaseUrl/warehouse/tasks/$taskId" -Method Get -ContentType "application/json" -ErrorAction SilentlyContinue
    Write-Host "Task Status:" ($taskStatus | ConvertTo-Json)
} catch {
    Write-Host "Error checking task status: $_"
}

Write-Host ""

# Test invalid request
Write-Host "3. Testing invalid request (missing git_url)..."
$invalidPayload = "{}"
try {
    $invalidResponse = Invoke-RestMethod -Uri "$BaseUrl/warehouse/repos" -Method Post -ContentType "application/json" -Body $invalidPayload -ErrorAction SilentlyContinue
    Write-Host "Response:" ($invalidResponse | ConvertTo-Json)
} catch {
    Write-Host "Expected error response: $($_.Exception.Response.StatusCode)"
    if ($_.ErrorDetails.Message) {
        Write-Host $_.ErrorDetails.Message
    }
}

Write-Host ""

# Test non-existent task
Write-Host "4. Testing non-existent task..."
try {
    $nonExistent = Invoke-RestMethod -Uri "$BaseUrl/warehouse/tasks/non_existent_task" -Method Get -ContentType "application/json" -ErrorAction SilentlyContinue
    Write-Host "Response:" ($nonExistent | ConvertTo-Json)
} catch {
    Write-Host "Expected error response: $($_.Exception.Response.StatusCode)"
    if ($_.ErrorDetails.Message) {
        Write-Host $_.ErrorDetails.Message
    }
}

Write-Host ""
Write-Host "============================="
Write-Host "API testing complete!" 