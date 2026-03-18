package checker

import (
	"fmt"

	sitter "github.com/smacker/go-tree-sitter"
	"llm-bouncer/language"
)

func checkNaming(root *sitter.Node, src []byte, lang *language.LanguageConfig) []Violation {
	var violations []Violation

	varTypes := nodeTypeSet(lang.VariableNodeTypes)
	paramTypes := nodeTypeSet(lang.ParameterNodeTypes)
	identTypes := nodeTypeSet(lang.IdentTypes())

	// Collect Go receiver names to exclude them.
	receiverNames := map[string]bool{}
	if lang.ReceiverNodeType != "" {
		walk(root, func(n *sitter.Node) bool {
			if n.Type() == lang.ReceiverNodeType {
				params := n.ChildByFieldName("parameters")
				if params != nil {
					collectReceiverIdents(params, src, receiverNames)
				}
			}
			return true
		})
	}

	walk(root, func(n *sitter.Node) bool {
		if varTypes[n.Type()] {
			checkVarIdents(n, src, lang, identTypes, receiverNames, &violations)
		}
		if paramTypes[n.Type()] {
			checkParamIdents(n, src, lang, identTypes, receiverNames, &violations)
		}
		return true
	})

	return violations
}

func checkVarIdents(n *sitter.Node, src []byte, lang *language.LanguageConfig, identTypes map[string]bool, receivers map[string]bool, violations *[]Violation) {
	// Check the "left" field first (Go short_var_declaration, Python assignment).
	if left := n.ChildByFieldName("left"); left != nil {
		extractAndCheckIdents(left, src, lang, identTypes, receivers, "variable", violations)
		return
	}
	// Scan children for variable_declarator (JS/TS) or variable_declaration (Kotlin) or pattern (Swift).
	for i := 0; i < int(n.ChildCount()); i++ {
		child := n.Child(i)
		switch child.Type() {
		case "variable_declarator":
			if nameNode := child.ChildByFieldName("name"); nameNode != nil && identTypes[nameNode.Type()] {
				checkIdentNode(nameNode, src, lang, receivers, "variable", violations)
			}
		case "variable_declaration", "pattern":
			// Kotlin: variable_declaration → simple_identifier
			// Swift: pattern → simple_identifier
			extractAndCheckIdents(child, src, lang, identTypes, receivers, "variable", violations)
		}
	}
}

func checkParamIdents(n *sitter.Node, src []byte, lang *language.LanguageConfig, identTypes map[string]bool, receivers map[string]bool, violations *[]Violation) {
	// Try "name" field first (Go parameter_declaration, TS required_parameter, etc.).
	if nameNode := n.ChildByFieldName("name"); nameNode != nil {
		checkIdentNode(nameNode, src, lang, receivers, "parameter", violations)
		return
	}
	// Fall back to first identifier-type child.
	for i := 0; i < int(n.ChildCount()); i++ {
		child := n.Child(i)
		if identTypes[child.Type()] {
			checkIdentNode(child, src, lang, receivers, "parameter", violations)
			return
		}
	}
}

func checkIdentNode(n *sitter.Node, src []byte, lang *language.LanguageConfig, receivers map[string]bool, kind string, violations *[]Violation) {
	name := nodeText(n, src)
	if isTooShort(name, lang) && !receivers[name] {
		*violations = append(*violations, Violation{
			Line:    startLine(n),
			Rule:    "naming",
			Message: fmt.Sprintf("%s %q is too short; use a descriptive name", kind, name),
		})
	}
}

func extractAndCheckIdents(n *sitter.Node, src []byte, lang *language.LanguageConfig, identTypes map[string]bool, receivers map[string]bool, kind string, violations *[]Violation) {
	if identTypes[n.Type()] {
		checkIdentNode(n, src, lang, receivers, kind, violations)
		return
	}
	for i := 0; i < int(n.ChildCount()); i++ {
		extractAndCheckIdents(n.Child(i), src, lang, identTypes, receivers, kind, violations)
	}
}

func collectReceiverIdents(paramList *sitter.Node, src []byte, names map[string]bool) {
	for i := 0; i < int(paramList.ChildCount()); i++ {
		child := paramList.Child(i)
		if child.Type() == "parameter_declaration" {
			if nameNode := child.ChildByFieldName("name"); nameNode != nil {
				names[nodeText(nameNode, src)] = true
			}
		}
	}
}

func isTooShort(name string, lang *language.LanguageConfig) bool {
	if lang.AcceptableShortNames[name] {
		return false
	}
	return len(name) == 1
}
