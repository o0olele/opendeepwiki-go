package main

import (
	"fmt"
	"log"
	"os"

	"github.com/o0olele/opendeepwiki-go/internal/analyzer/codemap"
)

func main() {
	// Check command line arguments
	if len(os.Args) < 3 {
		fmt.Println("Usage: example <repo_path> <openai_api_key>")
		os.Exit(1)
	}

	repoPath := os.Args[1]
	apiKey := os.Args[2]

	// Create a new code map service
	service, err := codemap.NewCodeMapService(
		apiKey,                   // OpenAI API key
		"text-embedding-3-small", // OpenAI embedding model
		"",                       // Base URL for OpenAI API (leave empty for default)
		1536,                     // Embedding dimensions
		repoPath,                 // Base path for the repository
	)
	if err != nil {
		log.Fatalf("Failed to create code map service: %v", err)
	}

	// Index the repository
	fmt.Println("Indexing repository...")
	err = service.IndexRepository(repoPath, "example-repo")
	if err != nil {
		log.Fatalf("Failed to index repository: %v", err)
	}
	fmt.Println("Repository indexed successfully!")

	// Search for code
	fmt.Println("\nSearching for 'user authentication'...")
	results, err := service.SearchCode("user authentication", "example-repo", 5)
	if err != nil {
		log.Fatalf("Failed to search code: %v", err)
	}

	// Print search results
	fmt.Printf("Found %d results:\n", len(results))
	for i, result := range results {
		fmt.Printf("\n%d. %s (%.2f relevance)\n", i+1, result.ID, result.Relevance)
		fmt.Printf("Description: %s\n", result.Description)
		fmt.Printf("Code snippet: %s\n", truncateString(result.Code, 200))
	}

	// Analyze file dependencies
	fmt.Println("\nAnalyzing file dependencies...")
	if len(results) > 0 {
		// Get the file path from the first result's metadata
		filePath := results[0].ID
		tree, err := service.AnalyzeFileDependencies(filePath)
		if err != nil {
			log.Fatalf("Failed to analyze file dependencies: %v", err)
		}

		// Print dependency tree
		fmt.Printf("\nDependency tree for %s:\n", filePath)
		printDependencyTree(tree, 0)
	}
}

// printDependencyTree prints a dependency tree with indentation
func printDependencyTree(tree *codemap.DependencyTree, level int) {
	indent := ""
	for i := 0; i < level; i++ {
		indent += "  "
	}

	fmt.Printf("%s- %s\n", indent, tree.Name)

	// Print functions
	for _, function := range tree.Functions {
		fmt.Printf("%s  • Function: %s (line %d)\n", indent, function.Name, function.LineNumber)
	}

	// Print children
	for _, child := range tree.Children {
		printDependencyTree(child, level+1)
	}
}

// truncateString truncates a string to the specified length
func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}
