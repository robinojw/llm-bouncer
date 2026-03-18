package language

import (
	"regexp"

	"github.com/smacker/go-tree-sitter/java"
)

func init() {
	Register(&LanguageConfig{
		Name:           "Java",
		Language:       java.GetLanguage(),
		FileExtensions: []string{".java"},
		FileNameRegex:  regexp.MustCompile(`^[A-Z][a-zA-Z0-9]*\.java$`),

		FunctionNodeTypes: []string{"method_declaration", "constructor_declaration"},
		FunctionNameField: "name",

		VariableNodeTypes:  []string{"local_variable_declaration"},
		ParameterNodeTypes: []string{"formal_parameter"},
		AcceptableShortNames: map[string]bool{
			"i": true, "j": true, "k": true, "_": true,
		},

		IfNodeTypes: []string{"if_statement"},
		IfBodyField: "consequence",

		ComplexityNodeTypes: []string{
			"if_statement", "for_statement", "enhanced_for_statement",
			"while_statement", "do_statement",
			"switch_label",
		},
		BooleanOperators:   []string{"&&", "||"},
		BinaryExprNodeType: "binary_expression",

		StringNodeTypes:  []string{"string_literal"},
		NumberNodeTypes:  []string{"decimal_integer_literal", "decimal_floating_point_literal", "hex_integer_literal"},
		CommentNodeTypes: []string{"line_comment", "block_comment"},

		ConstStrategy: ConstByKeyword,
		ConstKeyword:  "final",
	})
}
