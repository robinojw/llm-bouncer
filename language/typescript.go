package language

import (
	"regexp"

	"github.com/smacker/go-tree-sitter/typescript/typescript"
)

func init() {
	Register(&LanguageConfig{
		Name:           "TypeScript",
		Language:       typescript.GetLanguage(),
		FileExtensions: []string{".ts", ".tsx"},
		FileNameRegex:  regexp.MustCompile(`^[a-z][a-z0-9]*(-[a-z0-9]+)*\.(ts|tsx)$`),

		FunctionNodeTypes: []string{"function_declaration", "arrow_function", "method_definition"},
		FunctionNameField: "name",

		VariableNodeTypes:  []string{"lexical_declaration", "variable_declaration"},
		ParameterNodeTypes: []string{"required_parameter", "optional_parameter"},
		AcceptableShortNames: map[string]bool{
			"i": true, "j": true, "k": true, "_": true,
		},

		IfNodeTypes: []string{"if_statement"},
		IfBodyField: "consequence",

		ComplexityNodeTypes: []string{
			"if_statement", "for_statement", "for_in_statement",
			"while_statement", "do_statement",
			"switch_case",
		},
		BooleanOperators:   []string{"&&", "||"},
		BinaryExprNodeType: "binary_expression",

		StringNodeTypes:  []string{"string", "template_string"},
		NumberNodeTypes:  []string{"number"},
		CommentNodeTypes: []string{"comment"},

		ConstStrategy:   ConstByKeyword,
		ConstKeyword:    "const",
	})
}
