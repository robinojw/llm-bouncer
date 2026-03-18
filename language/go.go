package language

import (
	"regexp"

	sitter "github.com/smacker/go-tree-sitter"
	"github.com/smacker/go-tree-sitter/golang"
)

func init() {
	Register(&LanguageConfig{
		Name:           "Go",
		Language:       golang.GetLanguage(),
		FileExtensions: []string{".go"},
		FileNameRegex:  regexp.MustCompile(`^[a-z][a-z0-9]*(_[a-z0-9]+)*(_test)?\.go$`),

		FunctionNodeTypes: []string{"function_declaration", "method_declaration"},
		FunctionNameField: "name",

		VariableNodeTypes:    []string{"short_var_declaration"},
		ParameterNodeTypes:   []string{"parameter_declaration"},
		AcceptableShortNames: map[string]bool{"i": true, "j": true, "k": true, "_": true},
		ReceiverNodeType:     "method_declaration",

		IfNodeTypes: []string{"if_statement"},
		IfBodyField: "consequence",

		ComplexityNodeTypes: []string{
			"if_statement", "for_statement",
			"expression_case", "type_case",
			"communication_case",
		},
		BooleanOperators:   []string{"&&", "||"},
		BinaryExprNodeType: "binary_expression",

		StringNodeTypes:  []string{"interpreted_string_literal", "raw_string_literal"},
		NumberNodeTypes:  []string{"int_literal", "float_literal"},
		CommentNodeTypes: []string{"comment"},

		ConstStrategy:       ConstByBlock,
		ConstBlockNodeTypes: []string{"const_declaration"},
	})
}

// GoLanguage returns the tree-sitter language for Go.
func GoLanguage() *sitter.Language {
	return golang.GetLanguage()
}
