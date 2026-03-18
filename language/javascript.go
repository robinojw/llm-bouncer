package language

import (
	"regexp"

	"github.com/smacker/go-tree-sitter/javascript"
)

func init() {
	Register(&LanguageConfig{
		Name:           "JavaScript",
		Language:       javascript.GetLanguage(),
		FileExtensions: []string{".js", ".jsx"},
		FileNameRegex:  regexp.MustCompile(`^[a-zA-Z][a-zA-Z0-9]*([.-][a-zA-Z0-9]+)*\.(js|jsx)$`),

		FunctionNodeTypes: []string{"function_declaration", "arrow_function", "method_definition"},
		FunctionNameField: "name",

		VariableNodeTypes:  []string{"lexical_declaration", "variable_declaration"},
		ParameterNodeTypes: []string{"formal_parameters"},
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
