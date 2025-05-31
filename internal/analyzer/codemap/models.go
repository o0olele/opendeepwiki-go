package codemap

// CodeSegment represents a segment of code with metadata
type CodeSegment struct {
	Type          string   `json:"type"`          // Type of code segment (class, method, function, etc.)
	Name          string   `json:"name"`          // Name of the code segment
	Code          string   `json:"code"`          // Content of the code
	StartLine     int      `json:"start_line"`    // Starting line number
	EndLine       int      `json:"end_line"`      // Ending line number
	Namespace     string   `json:"namespace"`     // Namespace
	ClassName     string   `json:"class_name"`    // Class name (for methods)
	Documentation string   `json:"documentation"` // Documentation comments
	ReturnType    string   `json:"return_type"`   // Return type
	Parameters    string   `json:"parameters"`    // Parameter list
	Dependencies  []string `json:"dependencies"`  // Dependencies
	Modifiers     string   `json:"modifiers"`     // Modifiers (public, private, etc.)
}

// FunctionInfo represents information about a function
type FunctionInfo struct {
	Name       string   `json:"name"`        // Function name
	FullName   string   `json:"full_name"`   // Full function name with path
	Body       string   `json:"body"`        // Function body
	FilePath   string   `json:"file_path"`   // File path containing the function
	LineNumber int      `json:"line_number"` // Line number where the function starts
	Calls      []string `json:"calls"`       // List of function calls made by this function
}

// Function represents a simple function definition
type Function struct {
	Name string `json:"name"` // Function name
	Body string `json:"body"` // Function body
}

// DependencyNodeType represents the type of a dependency node
type DependencyNodeType string

const (
	FileNodeType     DependencyNodeType = "file"
	FunctionNodeType DependencyNodeType = "function"
)

// DependencyTree represents a tree of dependencies
type DependencyTree struct {
	NodeType   DependencyNodeType    `json:"node_type"`   // Type of node (file or function)
	Name       string                `json:"name"`        // Name of the node
	FullPath   string                `json:"full_path"`   // Full path of the node
	LineNumber int                   `json:"line_number"` // Line number (for functions)
	IsCyclic   bool                  `json:"is_cyclic"`   // Whether this node creates a cycle
	Children   []*DependencyTree     `json:"children"`    // Child dependencies
	Functions  []*DependencyFunction `json:"functions"`   // Functions contained in a file
}

// DependencyFunction represents a function in a dependency tree
type DependencyFunction struct {
	Name       string `json:"name"`        // Function name
	LineNumber int    `json:"line_number"` // Line number where the function starts
}

// SearchResult represents a search result with code and metadata
type SearchResult struct {
	ID          string          `json:"id"`          // Unique identifier
	Code        string          `json:"code"`        // Code content
	Description string          `json:"description"` // Description of the code
	Relevance   float64         `json:"relevance"`   // Relevance score
	References  *DependencyTree `json:"references"`  // References to other code
}
