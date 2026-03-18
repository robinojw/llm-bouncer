package checker

import (
	"fmt"

	sitter "github.com/smacker/go-tree-sitter"
	"llm-bouncer/language"
)

const maxCyclomaticComplexity = 10

func checkCyclomaticComplexity(root *sitter.Node, src []byte, lang *language.LanguageConfig) []Violation {
	var violations []Violation

	funcTypes := nodeTypeSet(lang.FunctionNodeTypes)

	walk(root, func(n *sitter.Node) bool {
		if !funcTypes[n.Type()] {
			return true
		}

		// Skip nested functions (they get their own score).
		if isNestedFunction(n, funcTypes) {
			return false
		}

		name := functionName(n, src, lang)
		complexity := calculateComplexity(n, src, lang, funcTypes)

		if complexity > maxCyclomaticComplexity {
			violations = append(violations, Violation{
				Line: startLine(n),
				Rule: "cyclomatic-complexity",
				Message: fmt.Sprintf(
					"function %q has complexity %d (max %d); break into smaller functions",
					name, complexity, maxCyclomaticComplexity,
				),
			})
		}

		// Don't recurse into function body again (nested functions handled separately).
		return false
	})

	return violations
}

func isNestedFunction(n *sitter.Node, funcTypes map[string]bool) bool {
	for p := n.Parent(); p != nil; p = p.Parent() {
		if funcTypes[p.Type()] {
			return true
		}
	}
	return false
}

func functionName(n *sitter.Node, src []byte, lang *language.LanguageConfig) string {
	if nameNode := n.ChildByFieldName(lang.FunctionNameField); nameNode != nil {
		return nodeText(nameNode, src)
	}
	return "<anonymous>"
}

// calculateComplexity counts decision points using McCabe's formula.
func calculateComplexity(funcNode *sitter.Node, src []byte, lang *language.LanguageConfig, funcTypes map[string]bool) int {
	complexity := 1

	complexityTypes := nodeTypeSet(lang.ComplexityNodeTypes)
	boolOps := nodeTypeSet(lang.BooleanOperators)

	walk(funcNode, func(n *sitter.Node) bool {
		// Skip nested functions.
		if n != funcNode && funcTypes[n.Type()] {
			return false
		}

		if complexityTypes[n.Type()] {
			complexity++
		}

		// Count boolean operators in binary expressions.
		if n.Type() == lang.BinaryExprNodeType {
			for i := 0; i < int(n.ChildCount()); i++ {
				child := n.Child(i)
				text := nodeText(child, src)
				if boolOps[text] {
					complexity++
				}
			}
		}

		return true
	})

	return complexity
}
