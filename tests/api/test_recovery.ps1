# OpenDeepWiki 任务恢复测试脚本 (PowerShell)

Write-Host "===== 测试任务恢复功能 =====" -ForegroundColor Green
Write-Host "1. 提交一个仓库任务" -ForegroundColor Cyan

# 提交仓库任务
$repoPayload = @{
    git_url = "https://github.com/gin-gonic/gin.git"
} | ConvertTo-Json

$response = Invoke-RestMethod -Uri "http://localhost:8080/api/warehouse/repos" -Method Post -ContentType "application/json" -Body $repoPayload
Write-Host "响应:" -ForegroundColor Yellow
$response | ConvertTo-Json

# 获取任务ID
$taskId = $response.task_id
Write-Host "任务ID: $taskId" -ForegroundColor Cyan

Write-Host "2. 等待几秒钟让任务开始处理" -ForegroundColor Cyan
Start-Sleep -Seconds 3

Write-Host "3. 检查任务状态" -ForegroundColor Cyan
$taskStatus = Invoke-RestMethod -Uri "http://localhost:8080/api/warehouse/tasks/$taskId" -Method Get -ContentType "application/json"
Write-Host "任务状态:" -ForegroundColor Yellow
$taskStatus | ConvertTo-Json

Write-Host "4. 现在终止服务器进程 (Ctrl+C)" -ForegroundColor Red
Write-Host "   然后重新启动服务器..." -ForegroundColor Cyan
Write-Host "   观察日志中是否有任务恢复的信息" -ForegroundColor Cyan
Write-Host "   例如: 'Recovering pending tasks from database...'" -ForegroundColor Yellow
Write-Host "        'Found X pending tasks to recover'" -ForegroundColor Yellow
Write-Host "        'Recovered task XXX for repository XXX with status XXX'" -ForegroundColor Yellow

Write-Host "5. 重启后，再次检查任务状态，应该会继续处理或已完成" -ForegroundColor Cyan
Write-Host "   Invoke-RestMethod -Uri 'http://localhost:8080/api/warehouse/tasks/$taskId' -Method Get -ContentType 'application/json' | ConvertTo-Json" -ForegroundColor Yellow

Write-Host "===== 测试完成 =====" -ForegroundColor Green 