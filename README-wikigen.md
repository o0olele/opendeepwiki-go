# OpenDeepWiki - Wiki 文档生成工具

OpenDeepWiki 是一个基于 Go 语言和 LangChain 的代码仓库文档生成工具，它可以自动分析 Git 仓库并生成详细的 Wiki 文档。

## 功能特点

- 自动克隆和分析 Git 仓库
- 智能识别仓库结构和重要文件
- 使用 AI 生成项目概述和详细文档
- 支持多种编程语言的代码解析
- 生成提交历史和更新日志
- 并发处理以提高生成速度
- 支持自定义过滤规则和文件大小限制

## 安装

```bash
go get github.com/o0olele/opendeepwiki-go
```

## 使用方法

### 命令行工具

```bash
go run cmd/wikigen/main.go -git https://github.com/username/repo.git -branch main
```

参数说明：
- `-git`: Git 仓库 URL（必需）
- `-branch`: Git 分支（默认：main）
- `-dir`: 仓库存储目录（默认：./repos）
- `-model`: OpenAI 模型（默认：gpt-4）
- `-key`: OpenAI API 密钥（也可通过环境变量 OPENAI_API_KEY 设置）
- `-db`: 数据库路径（默认：./wiki.db）

### 作为库使用

```go
package main

import (
    "context"
    "log"
    
    "github.com/google/uuid"
    "github.com/o0olele/opendeepwiki-go/internal/database"
    "github.com/o0olele/opendeepwiki-go/internal/wikigen"
)

func main() {
    // 初始化数据库
    if err := database.InitDB("./wiki.db"); err != nil {
        log.Fatalf("初始化数据库失败: %v", err)
    }
    
    // 创建选项
    options := &wikigen.WikiOptions{
        EnableSmartFilter: true,
        ExcludedFiles:     wikigen.DefaultExcludedFiles,
        MaxFileSize:       1024 * 1024, // 1MB
        MaxTokens:         8192,
    }
    
    // 创建 Wiki 服务
    service, err := wikigen.NewWikiService(
        "your-openai-api-key",
        "gpt-4",
        "./repos",
        options,
    )
    if err != nil {
        log.Fatalf("创建 Wiki 服务失败: %v", err)
    }
    
    // 生成任务 ID
    taskID := uuid.New().String()
    
    // 生成 Wiki 文档
    if err := service.GenerateWiki(
        context.Background(),
        taskID,
        "https://github.com/username/repo.git",
        "main",
    ); err != nil {
        log.Fatalf("生成 Wiki 文档失败: %v", err)
    }
}
```

## 配置选项

`WikiOptions` 结构体提供了以下配置选项：

- `EnableSmartFilter`: 是否启用智能过滤（对于大型仓库）
- `ExcludedFiles`: 要排除的文件和目录列表
- `MaxFileSize`: 处理的最大文件大小（字节）
- `MaxTokens`: AI 模型使用的最大令牌数

## 生成的文档

生成的文档包括：

1. 项目概述：包含项目名称、简介、主要功能和技术栈
2. 代码结构：包含文件和目录统计、主要模块说明
3. 详细文档：按照仓库结构组织的详细文档
4. 提交历史：重要的提交记录和更新日志

## 依赖项

- Go 1.20 或更高版本
- LangChain Go 库
- OpenAI API 密钥
- Git

## 许可证

MIT 