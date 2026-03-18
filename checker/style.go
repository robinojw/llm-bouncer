package checker

import (
	"fmt"
	"strings"
	"unicode"

	sitter "github.com/smacker/go-tree-sitter"
	"llm-bouncer/language"
)

func checkNestedIfs(root *sitter.Node, src []byte, lang *language.LanguageConfig) []Violation {
	var violations []Violation
	ifTypes := nodeTypeSet(lang.IfNodeTypes)

	walk(root, func(n *sitter.Node) bool {
		if !ifTypes[n.Type()] {
			return true
		}

		body := findIfBody(n, lang)
		if body == nil {
			return true
		}

		walk(body, func(inner *sitter.Node) bool {
			if inner == body {
				return true
			}
			if ifTypes[inner.Type()] {
				violations = append(violations, Violation{
					Line:    startLine(inner),
					Rule:    "no-nested-ifs",
					Message: "nested if statement; use early returns or extract a helper function",
				})
				return false
			}
			return true
		})

		return true
	})

	return violations
}

// findIfBody returns the body node of an if statement/expression.
func findIfBody(n *sitter.Node, lang *language.LanguageConfig) *sitter.Node {
	// Try named field first.
	if lang.IfBodyField != "" {
		if body := n.ChildByFieldName(lang.IfBodyField); body != nil {
			return body
		}
	}
	// Fallback: find first child matching IfBodyNodeTypes (Kotlin: control_structure_body, Swift: statements).
	if len(lang.IfBodyNodeTypes) > 0 {
		bodyTypes := nodeTypeSet(lang.IfBodyNodeTypes)
		for i := 0; i < int(n.ChildCount()); i++ {
			child := n.Child(i)
			if bodyTypes[child.Type()] {
				return child
			}
		}
	}
	// Last fallback: try "body" field.
	return n.ChildByFieldName("body")
}

func checkInlineBooleans(root *sitter.Node, src []byte, lang *language.LanguageConfig) []Violation {
	var violations []Violation
	ifTypes := nodeTypeSet(lang.IfNodeTypes)
	boolOps := nodeTypeSet(lang.BooleanOperators)

	walk(root, func(n *sitter.Node) bool {
		if !ifTypes[n.Type()] {
			return true
		}

		cond := findIfCondition(n, lang)
		if cond == nil {
			return true
		}

		if hasCompoundBoolean(cond, src, lang, boolOps) {
			violations = append(violations, Violation{
				Line:    startLine(n),
				Rule:    "no-inline-booleans",
				Message: "complex boolean used directly in if; assign to a descriptively named variable first",
			})
		}

		return true
	})

	return violations
}

// findIfCondition returns the condition expression of an if statement/expression.
func findIfCondition(n *sitter.Node, lang *language.LanguageConfig) *sitter.Node {
	// Try "condition" field first (Go, Python, JS/TS, Java, Rust).
	if cond := n.ChildByFieldName("condition"); cond != nil {
		return cond
	}
	// Fallback for Kotlin/Swift: the condition is the first expression child
	// (skip keyword tokens like "if", "(", ")").
	keywords := map[string]bool{"if": true, "(": true, ")": true, "{": true}
	bodyTypes := nodeTypeSet(lang.IfBodyNodeTypes)
	for i := 0; i < int(n.ChildCount()); i++ {
		child := n.Child(i)
		if keywords[child.Type()] {
			continue
		}
		if bodyTypes[child.Type()] {
			break
		}
		// First non-keyword, non-body child is the condition.
		return child
	}
	return nil
}

func hasCompoundBoolean(n *sitter.Node, src []byte, lang *language.LanguageConfig, boolOps map[string]bool) bool {
	boolExprTypes := nodeTypeSet(lang.BooleanExprNodeTypes)
	found := false
	walk(n, func(child *sitter.Node) bool {
		if found {
			return false
		}
		// Kotlin/Swift: node type itself IS the boolean expression.
		if boolExprTypes[child.Type()] {
			found = true
			return false
		}
		// Go/JS/TS/Java/Rust/Python: binary_expression with operator children.
		if lang.BinaryExprNodeType != "" && child.Type() == lang.BinaryExprNodeType {
			for i := 0; i < int(child.ChildCount()); i++ {
				text := nodeText(child.Child(i), src)
				if boolOps[text] {
					found = true
					return false
				}
			}
		}
		return true
	})
	return found
}

func checkInlineComments(root *sitter.Node, src []byte, lang *language.LanguageConfig) []Violation {
	var violations []Violation
	commentTypes := nodeTypeSet(lang.CommentNodeTypes)
	lines := strings.Split(string(src), "\n")

	walk(root, func(n *sitter.Node) bool {
		if !commentTypes[n.Type()] {
			return true
		}

		row := int(n.StartPoint().Row)
		if row < 0 || row >= len(lines) {
			return true
		}

		trimmedLine := strings.TrimSpace(lines[row])
		if !isStandaloneComment(trimmedLine, lang) {
			violations = append(violations, Violation{
				Line:    startLine(n),
				Rule:    "no-inline-comments",
				Message: "inline comment found; write self-documenting code instead",
			})
		}

		return true
	})

	return violations
}

func isStandaloneComment(trimmedLine string, lang *language.LanguageConfig) bool {
	// A standalone comment means the line starts with a comment marker.
	commentPrefixes := []string{"//", "/*", "#", "--"}
	for _, prefix := range commentPrefixes {
		if strings.HasPrefix(trimmedLine, prefix) {
			return true
		}
	}
	return false
}

func checkRepeatedStrings(root *sitter.Node, src []byte, lang *language.LanguageConfig) []Violation {
	type occurrence struct {
		firstLine int
		count     int
	}

	stringTypes := nodeTypeSet(lang.StringNodeTypes)
	seen := make(map[string]*occurrence)

	walk(root, func(n *sitter.Node) bool {
		if !stringTypes[n.Type()] {
			return true
		}

		value := nodeText(n, src)
		if len(value) <= 3 {
			return true
		}

		if existing, found := seen[value]; found {
			existing.count++
		} else {
			seen[value] = &occurrence{firstLine: startLine(n), count: 1}
		}
		return true
	})

	var violations []Violation
	for value, occ := range seen {
		if occ.count > 1 {
			violations = append(violations, Violation{
				Line:    occ.firstLine,
				Rule:    "no-repeated-strings",
				Message: fmt.Sprintf("string literal %s appears %d times; assign to a named constant", value, occ.count),
			})
		}
	}
	return violations
}

func checkMagicNumbers(root *sitter.Node, src []byte, lang *language.LanguageConfig) []Violation {
	numberTypes := nodeTypeSet(lang.NumberNodeTypes)

	var violations []Violation
	walk(root, func(n *sitter.Node) bool {
		if !numberTypes[n.Type()] {
			return true
		}

		value := nodeText(n, src)

		// Skip non-numeric text (e.g., TypeScript "number" type annotations).
		if !looksNumeric(value) {
			return true
		}

		if value == "0" || value == "1" {
			return true
		}

		if isInConstContext(n, src, lang) {
			return true
		}

		violations = append(violations, Violation{
			Line:    startLine(n),
			Rule:    "no-magic-numbers",
			Message: fmt.Sprintf("magic number %s; extract to a named constant", value),
		})
		return true
	})

	return violations
}

func isInConstContext(n *sitter.Node, src []byte, lang *language.LanguageConfig) bool {
	switch lang.ConstStrategy {
	case language.ConstByBlock:
		constTypes := nodeTypeSet(lang.ConstBlockNodeTypes)
		return isDescendantOf(n, constTypes)

	case language.ConstByKeyword:
		// Walk up to find a declaration with the const keyword.
		for p := n.Parent(); p != nil; p = p.Parent() {
			pType := p.Type()
			// JS/TS: lexical_declaration with "const" keyword child
			if pType == "lexical_declaration" || pType == "variable_declaration" {
				for i := 0; i < int(p.ChildCount()); i++ {
					child := p.Child(i)
					if nodeText(child, src) == lang.ConstKeyword {
						return true
					}
				}
			}
			// Rust: const_item
			if lang.ConstBlockNodeTypes != nil {
				constTypes := nodeTypeSet(lang.ConstBlockNodeTypes)
				if constTypes[pType] {
					return true
				}
			}
			// Java: field_declaration or local_variable_declaration with "final" modifier
			if pType == "field_declaration" || pType == "local_variable_declaration" || pType == "constant_declaration" {
				if hasModifier(p, src, lang.ConstKeyword) {
					return true
				}
			}
		}
		return false

	case language.ConstByConvention:
		// Python: check if the variable name is UPPER_SNAKE_CASE.
		for p := n.Parent(); p != nil; p = p.Parent() {
			if p.Type() == "assignment" {
				if left := p.ChildByFieldName("left"); left != nil {
					name := nodeText(left, src)
					if isUpperSnakeCase(name) {
						return true
					}
				}
				return false
			}
		}
		return false

	case language.ConstByBindingKeyword:
		// Kotlin (val) / Swift (let): walk up to property_declaration,
		// check if a child contains the binding keyword.
		for p := n.Parent(); p != nil; p = p.Parent() {
			if p.Type() == "property_declaration" {
				for i := 0; i < int(p.ChildCount()); i++ {
					child := p.Child(i)
					text := nodeText(child, src)
					if text == lang.ConstKeyword {
						return true
					}
					// Kotlin: binding_pattern_kind contains "val"
					if child.Type() == "binding_pattern_kind" {
						if nodeText(child, src) == lang.ConstKeyword {
							return true
						}
					}
					// Swift: value_binding_pattern contains "let"
					if child.Type() == "value_binding_pattern" {
						for j := 0; j < int(child.ChildCount()); j++ {
							if nodeText(child.Child(j), src) == lang.ConstKeyword {
								return true
							}
						}
					}
				}
				return false
			}
		}
		return false
	}
	return false
}

func hasModifier(n *sitter.Node, src []byte, keyword string) bool {
	for i := 0; i < int(n.ChildCount()); i++ {
		child := n.Child(i)
		if child.Type() == "modifiers" {
			for j := 0; j < int(child.ChildCount()); j++ {
				if nodeText(child.Child(j), src) == keyword {
					return true
				}
			}
		}
		if nodeText(child, src) == keyword {
			return true
		}
	}
	return false
}

func isUpperSnakeCase(name string) bool {
	if len(name) == 0 {
		return false
	}
	for _, r := range name {
		if r != '_' && !unicode.IsUpper(r) && !unicode.IsDigit(r) {
			return false
		}
	}
	return true
}

func looksNumeric(s string) bool {
	if len(s) == 0 {
		return false
	}
	// Starts with a digit, or a dot followed by a digit (e.g., ".5"), or 0x/0b/0o prefix.
	first := s[0]
	return (first >= '0' && first <= '9') || (first == '.' && len(s) > 1 && s[1] >= '0' && s[1] <= '9')
}
