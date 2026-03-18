package language

import (
	"path/filepath"
	"regexp"
	"strings"

	sitter "github.com/smacker/go-tree-sitter"
)

// ConstStrategy describes how a language marks constants.
type ConstStrategy int

const (
	// ConstByBlock means constants are declared in a block (Go: const (...)).
	ConstByBlock ConstStrategy = iota
	// ConstByKeyword means a keyword marks the binding (JS/TS: const, Rust: const, Java: final).
	ConstByKeyword
	// ConstByConvention means UPPER_SNAKE_CASE names are treated as constants (Python).
	ConstByConvention
	// ConstByBindingKeyword means a binding keyword marks immutability (Kotlin: val, Swift: let).
	ConstByBindingKeyword
)

// LanguageConfig maps language concepts to tree-sitter node types.
type LanguageConfig struct {
	Name           string
	Language       *sitter.Language
	FileExtensions []string
	FileNameRegex  *regexp.Regexp

	FunctionNodeTypes []string
	FunctionNameField string

	VariableNodeTypes  []string
	ParameterNodeTypes []string
	IdentifierNodeTypes []string // Defaults to ["identifier"] if empty; Kotlin/Swift use ["simple_identifier"]
	AcceptableShortNames map[string]bool
	ReceiverNodeType     string // Go-specific; empty for other languages

	IfNodeTypes    []string
	IfBodyField    string   // Named field for if body (e.g., "consequence", "body")
	IfBodyNodeTypes []string // Fallback: child node types that represent the if body (e.g., "control_structure_body")

	ComplexityNodeTypes  []string
	BooleanOperators    []string
	BinaryExprNodeType  string
	BooleanExprNodeTypes []string // Node types that ARE boolean expressions (Kotlin/Swift: conjunction_expression, disjunction_expression)

	StringNodeTypes  []string
	NumberNodeTypes  []string
	CommentNodeTypes []string

	ConstStrategy       ConstStrategy
	ConstBlockNodeTypes []string
	ConstKeyword        string // "const" for JS/TS/Rust, "final" for Java
}

var registry = map[string]*LanguageConfig{}

// Register adds a language config to the registry keyed by each file extension.
func Register(cfg *LanguageConfig) {
	for _, ext := range cfg.FileExtensions {
		registry[ext] = cfg
	}
}

// Detect returns the LanguageConfig for a file path based on its extension, or nil.
func Detect(filePath string) *LanguageConfig {
	ext := strings.ToLower(filepath.Ext(filePath))
	return registry[ext]
}

// Supported returns true if the file extension is supported.
func Supported(filePath string) bool {
	return Detect(filePath) != nil
}

// IdentTypes returns the identifier node types, defaulting to ["identifier"].
func (c *LanguageConfig) IdentTypes() []string {
	if len(c.IdentifierNodeTypes) > 0 {
		return c.IdentifierNodeTypes
	}
	return []string{"identifier"}
}
