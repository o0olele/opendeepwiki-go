# OpenDeepWiki API 测试工具

本目录包含用于测试 OpenDeepWiki API 的各种工具和脚本。

## 测试文件说明

### api_tests.http

用于 VS Code REST Client 或 JetBrains IDE 的 HTTP 请求文件，可以直接在编辑器中测试 API。

使用方法：
1. 在 VS Code 中安装 REST Client 扩展
2. 打开 api_tests.http 文件
3. 点击 "Send Request" 链接执行请求

### test_api.sh

用于 Linux/macOS 的 Shell 脚本，使用 curl 命令测试 API。

使用方法：
```bash
chmod +x test_api.sh
./test_api.sh
```

### test_api.ps1

用于 Windows 的 PowerShell 脚本，使用 Invoke-RestMethod 测试 API。

使用方法：
```powershell
.\test_api.ps1
```

### api_test.html

基于浏览器的 API 测试工具，提供简单的 Web 界面。

使用方法：
1. 在浏览器中打开 api_test.html 文件
2. 输入 Git 仓库 URL 并提交
3. 查看 API 响应

### test_recovery.sh / test_recovery.ps1

用于测试服务重启后任务恢复功能的脚本。

使用方法：
```bash
# Linux/macOS
chmod +x test_recovery.sh
./test_recovery.sh

# Windows
.\test_recovery.ps1
```

## 注意事项

1. 所有测试脚本默认连接到 `http://localhost:8080/api`
2. 确保 OpenDeepWiki 服务器已运行
3. 某些测试可能需要手动修改任务 ID 