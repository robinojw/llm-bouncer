package language

import (
	"regexp"

	"github.com/smacker/go-tree-sitter/kotlin"
)

func init() {
	Register(&LanguageConfig{
		Name:           "Kotlin",
		Language:       kotlin.GetLanguage(),
		FileExtensions: []string{".kt", ".kts"},
		FileNameRegex:  regexp.MustCompile(`^[a-zA-Z][a-zA-Z0-9]*(\.[a-zA-Z0-9]+)*\.(kt|kts)$`),

		FunctionNodeTypes: []string{"function_declaration"},
		FunctionNameField: "name",

		VariableNodeTypes:   []string{"property_declaration"},
		ParameterNodeTypes:  []string{"parameter", "class_parameter"},
		IdentifierNodeTypes: []string{"simple_identifier"},
		AcceptableShortNames: map[string]bool{
			"i": true, "j": true, "k": true, "_": true,
		},

		IfNodeTypes:     []string{"if_expression"},
		IfBodyNodeTypes: []string{"control_structure_body"},

		ComplexityNodeTypes: []string{
			"if_expression", "for_statement", "while_statement",
			"do_while_statement", "when_entry",
		},
		BooleanOperators:     []string{"&&", "||"},
		BooleanExprNodeTypes: []string{"conjunction_expression", "disjunction_expression"},

		StringNodeTypes:  []string{"string_literal"},
		NumberNodeTypes:  []string{"integer_literal", "real_literal", "long_literal", "hex_literal"},
		CommentNodeTypes: []string{"line_comment", "multiline_comment"},

		ConstStrategy: ConstByBindingKeyword,
		ConstKeyword:  "val",
	})
}
