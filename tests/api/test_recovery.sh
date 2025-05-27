#!/bin/bash
# OpenDeepWiki 任务恢复测试脚本

echo "===== 测试任务恢复功能 ====="
echo "1. 提交一个仓库任务"

# 提交仓库任务
curl -s -X POST "http://localhost:8080/api/warehouse/repos" \
  -H "Content-Type: application/json" \
  -d '{"git_url": "https://github.com/gin-gonic/gin.git"}' | jq

echo "2. 等待几秒钟让任务开始处理"
sleep 3

echo "3. 检查任务状态"
# 获取最新任务ID (这里需要手动替换为实际的任务ID)
TASK_ID="task_1234567890"
curl -s -X GET "http://localhost:8080/api/warehouse/tasks/$TASK_ID" \
  -H "Content-Type: application/json" | jq

echo "4. 现在终止服务器进程 (Ctrl+C)"
echo "   然后重新启动服务器..."
echo "   观察日志中是否有任务恢复的信息"
echo "   例如: 'Recovering pending tasks from database...'"
echo "        'Found X pending tasks to recover'"
echo "        'Recovered task XXX for repository XXX with status XXX'"

echo "5. 重启后，再次检查任务状态，应该会继续处理或已完成"
echo "   curl -s -X GET \"http://localhost:8080/api/warehouse/tasks/$TASK_ID\" -H \"Content-Type: application/json\" | jq"

echo "===== 测试完成 =====" 