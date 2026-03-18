package language

import (
	"regexp"

	"github.com/smacker/go-tree-sitter/python"
)

func init() {
	Register(&LanguageConfig{
		Name:           "Python",
		Language:       python.GetLanguage(),
		FileExtensions: []string{".py"},
		FileNameRegex:  regexp.MustCompile(`^[a-z][a-z0-9]*(_[a-z0-9]+)*\.py$`),

		FunctionNodeTypes: []string{"function_definition"},
		FunctionNameField: "name",

		VariableNodeTypes:  []string{"assignment"},
		ParameterNodeTypes: []string{"typed_parameter", "default_parameter", "typed_default_parameter"},
		AcceptableShortNames: map[string]bool{
			"i": true, "j": true, "k": true, "_": true,
			"self": true, "cls": true,
		},

		IfNodeTypes: []string{"if_statement"},
		IfBodyField: "consequence",

		ComplexityNodeTypes: []string{
			"if_statement", "for_statement", "while_statement",
			"case_clause",
		},
		BooleanOperators:   []string{"and", "or"},
		BinaryExprNodeType: "boolean_operator",

		StringNodeTypes:  []string{"string"},
		NumberNodeTypes:  []string{"integer", "float"},
		CommentNodeTypes: []string{"comment"},

		ConstStrategy: ConstByConvention,
	})
}
