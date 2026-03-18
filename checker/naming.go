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
			checkVarIdents(n, src, lang, receiverNames, &violations)
		}
		if paramTypes[n.Type()] {
			checkParamIdents(n, src, lang, receiverNames, &violations)
		}
		return true
	})

	return violations
}

func checkVarIdents(n *sitter.Node, src []byte, lang *language.LanguageConfig, receivers map[string]bool, violations *[]Violation) {
	// Check the "left" field first (Go short_var_declaration, Python assignment).
	if left := n.ChildByFieldName("left"); left != nil {
		extractAndCheckIdents(left, src, lang, receivers, "variable", violations)
		return
	}
	// Fall back to scanning direct identifier children (JS/TS variable_declarator, etc.).
	for i := 0; i < int(n.ChildCount()); i++ {
		child := n.Child(i)
		if child.Type() == "variable_declarator" {
			if nameNode := child.ChildByFieldName("name"); nameNode != nil && nameNode.Type() == "identifier" {
				name := nodeText(nameNode, src)
				if isTooShort(name, lang) && !receivers[name] {
					*violations = append(*violations, Violation{
						Line:    startLine(nameNode),
						Rule:    "naming",
						Message: fmt.Sprintf("variable %q is too short; use a descriptive name", name),
					})
				}
			}
		}
	}
}

func checkParamIdents(n *sitter.Node, src []byte, lang *language.LanguageConfig, receivers map[string]bool, violations *[]Violation) {
	// Try "name" field first (Go parameter_declaration, TS required_parameter, etc.).
	if nameNode := n.ChildByFieldName("name"); nameNode != nil {
		name := nodeText(nameNode, src)
		if isTooShort(name, lang) && !receivers[name] {
			*violations = append(*violations, Violation{
				Line:    startLine(nameNode),
				Rule:    "naming",
				Message: fmt.Sprintf("parameter %q is too short; use a descriptive name", name),
			})
		}
		return
	}
	// Fall back to first identifier child.
	for i := 0; i < int(n.ChildCount()); i++ {
		child := n.Child(i)
		if child.Type() == "identifier" {
			name := nodeText(child, src)
			if isTooShort(name, lang) && !receivers[name] {
				*violations = append(*violations, Violation{
					Line:    startLine(child),
					Rule:    "naming",
					Message: fmt.Sprintf("parameter %q is too short; use a descriptive name", name),
				})
			}
			return
		}
	}
}

func extractAndCheckIdents(n *sitter.Node, src []byte, lang *language.LanguageConfig, receivers map[string]bool, kind string, violations *[]Violation) {
	if n.Type() == "identifier" {
		name := nodeText(n, src)
		if isTooShort(name, lang) && !receivers[name] {
			*violations = append(*violations, Violation{
				Line:    startLine(n),
				Rule:    "naming",
				Message: fmt.Sprintf("%s %q is too short; use a descriptive name", kind, name),
			})
		}
		return
	}
	for i := 0; i < int(n.ChildCount()); i++ {
		extractAndCheckIdents(n.Child(i), src, lang, receivers, kind, violations)
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
