package language

import (
	"regexp"

	"github.com/smacker/go-tree-sitter/rust"
)

func init() {
	Register(&LanguageConfig{
		Name:           "Rust",
		Language:       rust.GetLanguage(),
		FileExtensions: []string{".rs"},
		FileNameRegex:  regexp.MustCompile(`^[a-z][a-z0-9]*(_[a-z0-9]+)*\.rs$`),

		FunctionNodeTypes: []string{"function_item"},
		FunctionNameField: "name",

		VariableNodeTypes:  []string{"let_declaration"},
		ParameterNodeTypes: []string{"parameter"},
		AcceptableShortNames: map[string]bool{
			"i": true, "j": true, "k": true, "_": true,
		},

		IfNodeTypes: []string{"if_expression"},
		IfBodyField: "consequence",

		ComplexityNodeTypes: []string{
			"if_expression", "for_expression", "while_expression",
			"loop_expression", "match_arm",
		},
		BooleanOperators:   []string{"&&", "||"},
		BinaryExprNodeType: "binary_expression",

		StringNodeTypes:  []string{"string_literal", "raw_string_literal"},
		NumberNodeTypes:  []string{"integer_literal", "float_literal"},
		CommentNodeTypes: []string{"line_comment", "block_comment"},

		ConstStrategy:       ConstByKeyword,
		ConstKeyword:        "const",
		ConstBlockNodeTypes: []string{"const_item"},
	})
}
