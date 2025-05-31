# CodeMap Package

The `codemap` package provides tools for analyzing, indexing, and searching code repositories. It's designed to help understand code dependencies, find relevant code snippets, and visualize the structure of a codebase.

## Features

- **Code Dependency Analysis**: Analyze dependencies between files and functions
- **Semantic Code Search**: Search for code using natural language queries
- **Multi-Language Support**: Support for Go, JavaScript, TypeScript, and more
- **Embedding Generation**: Generate embeddings for code snippets using OpenAI
- **In-Memory Storage**: Store embeddings in memory for quick access

## Usage

### Basic Usage

```go
import (
    "github.com/opendeepwiki/opendeepwiki-go/pkg/codemap"
)

// Create a new code map service
service, err := codemap.NewCodeMapService(
    "your-openai-api-key",
    "text-embedding-3-small",
    1536,
    "/path/to/repository",
)
if err != nil {
    // Handle error
}

// Index the repository
err = service.IndexRepository("/path/to/repository", "repo-id")
if err != nil {
    // Handle error
}

// Search for code
results, err := service.SearchCode("user authentication", "repo-id", 5)
if err != nil {
    // Handle error
}

// Analyze file dependencies
tree, err := service.AnalyzeFileDependencies("/path/to/file.go")
if err != nil {
    // Handle error
}
```

### Language Parsers

The package includes parsers for multiple languages:

- Go
- JavaScript
- TypeScript
- Generic (for other languages)

Each parser implements the `LanguageParser` interface:

```go
type LanguageParser interface {
    ExtractImports(fileContent string) []string
    ExtractFunctions(fileContent string) []Function
    ExtractFunctionCalls(functionBody string) []string
    ResolveImportPath(importPath string, currentFilePath string, basePath string) string
    GetFunctionLineNumber(fileContent string, functionName string) int
}
```

### Embedding Generation

The package uses OpenAI embeddings to enable semantic search of code:

```go
// Create an embedder
storage := codemap.NewInMemoryEmbeddingStorage()
embedder, err := codemap.NewOpenAIEmbedder(
    "your-openai-api-key",
    "text-embedding-3-small",
    1536,
    storage,
)
if err != nil {
    // Handle error
}

// Index a document
err = embedder.IndexDocument(
    "document-id",
    "code content",
    map[string]string{
        "language": "go",
        "file_name": "main.go",
    },
)
if err != nil {
    // Handle error
}

// Search for documents
results, err := embedder.Search(
    "user authentication",
    map[string]string{"language": "go"},
    5,
    0.3,
)
if err != nil {
    // Handle error
}
```

## Components

- **CodeMapService**: High-level service for code mapping functionality
- **DependencyAnalyzer**: Analyzes code dependencies
- **CodeIndexer**: Indexes code for searching
- **OpenAIEmbedder**: Generates embeddings using OpenAI
- **InMemoryEmbeddingStorage**: Stores embeddings in memory

## Example

See the `example` directory for a complete example of how to use the package.

## License

This package is part of the OpenDeepWiki project and is licensed under the same license as the project. 