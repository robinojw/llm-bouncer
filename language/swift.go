package language

import (
	"regexp"

	"github.com/smacker/go-tree-sitter/swift"
)

func init() {
	Register(&LanguageConfig{
		Name:           "Swift",
		Language:       swift.GetLanguage(),
		FileExtensions: []string{".swift"},
		FileNameRegex:  regexp.MustCompile(`^[a-zA-Z][a-zA-Z0-9]*(\+[a-zA-Z0-9]+)?\.swift$`),

		FunctionNodeTypes: []string{"function_declaration", "init_declaration"},
		FunctionNameField: "name",

		VariableNodeTypes:   []string{"property_declaration"},
		ParameterNodeTypes:  []string{"parameter"},
		IdentifierNodeTypes: []string{"simple_identifier"},
		AcceptableShortNames: map[string]bool{
			"i": true, "j": true, "k": true, "_": true,
		},

		IfNodeTypes:     []string{"if_statement"},
		IfBodyNodeTypes: []string{"statements"},

		ComplexityNodeTypes: []string{
			"if_statement", "guard_statement",
			"for_statement", "while_statement", "repeat_while_statement",
			"switch_entry",
		},
		BooleanOperators:     []string{"&&", "||"},
		BooleanExprNodeTypes: []string{"conjunction_expression", "disjunction_expression"},

		StringNodeTypes:  []string{"line_string_literal", "multi_line_string_literal", "raw_string_literal"},
		NumberNodeTypes:  []string{"integer_literal", "real_literal", "hex_literal"},
		CommentNodeTypes: []string{"comment", "multiline_comment"},

		ConstStrategy: ConstByBindingKeyword,
		ConstKeyword:  "let",
	})
}
