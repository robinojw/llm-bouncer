package checker

import (
	"fmt"
	"go/ast"
	"go/token"
	"strings"
)

func checkNestedIfs(fileSet *token.FileSet, file *ast.File) []Violation {
	var violations []Violation

	ast.Inspect(file, func(node ast.Node) bool {
		outerIf, ok := node.(*ast.IfStmt)
		if !ok {
			return true
		}

		ast.Inspect(outerIf.Body, func(inner ast.Node) bool {
			if innerIf, ok := inner.(*ast.IfStmt); ok {
				violations = append(violations, Violation{
					Line:    fileSet.Position(innerIf.Pos()).Line,
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

func checkInlineBooleans(fileSet *token.FileSet, file *ast.File) []Violation {
	var violations []Violation

	ast.Inspect(file, func(node ast.Node) bool {
		ifStmt, ok := node.(*ast.IfStmt)
		if !ok {
			return true
		}

		if isCompoundBoolean(ifStmt.Cond) {
			violations = append(violations, Violation{
				Line:    fileSet.Position(ifStmt.Pos()).Line,
				Rule:    "no-inline-booleans",
				Message: "complex boolean used directly in if; assign to a descriptively named variable first",
			})
		}
		return true
	})

	return violations
}

func isCompoundBoolean(expression ast.Expr) bool {
	binary, ok := expression.(*ast.BinaryExpr)
	if !ok {
		return false
	}
	operator := binary.Op.String()
	return operator == "&&" || operator == "||"
}

func checkInlineComments(fileSet *token.FileSet, file *ast.File, src []byte) []Violation {
	var violations []Violation
	lines := strings.Split(string(src), "\n")

	for _, commentGroup := range file.Comments {
		for _, comment := range commentGroup.List {
			position := fileSet.Position(comment.Pos())
			lineIndex := position.Line - 1
			if lineIndex < 0 || lineIndex >= len(lines) {
				continue
			}

			trimmedLine := strings.TrimSpace(lines[lineIndex])
			if !strings.HasPrefix(trimmedLine, "//") && !strings.HasPrefix(trimmedLine, "/*") {
				violations = append(violations, Violation{
					Line:    position.Line,
					Rule:    "no-inline-comments",
					Message: "inline comment found; write self-documenting code instead",
				})
			}
		}
	}

	return violations
}

func checkRepeatedStrings(fileSet *token.FileSet, file *ast.File) []Violation {
	type occurrence struct {
		firstLine int
		count     int
	}

	seen := make(map[string]*occurrence)

	ast.Inspect(file, func(node ast.Node) bool {
		lit, ok := node.(*ast.BasicLit)
		if !ok || lit.Kind != token.STRING || len(lit.Value) <= 3 {
			return true
		}

		if existing, found := seen[lit.Value]; found {
			existing.count++
		} else {
			seen[lit.Value] = &occurrence{firstLine: fileSet.Position(lit.Pos()).Line, count: 1}
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

func checkMagicNumbers(fileSet *token.FileSet, file *ast.File) []Violation {
	constPositions := collectConstPositions(file)

	var violations []Violation
	ast.Inspect(file, func(node ast.Node) bool {
		lit, ok := node.(*ast.BasicLit)
		if !ok {
			return true
		}
		if lit.Kind != token.INT && lit.Kind != token.FLOAT {
			return true
		}
		if lit.Value == "0" || lit.Value == "1" {
			return true
		}
		if constPositions[lit.Pos()] {
			return true
		}

		violations = append(violations, Violation{
			Line:    fileSet.Position(lit.Pos()).Line,
			Rule:    "no-magic-numbers",
			Message: fmt.Sprintf("magic number %s; extract to a named constant", lit.Value),
		})
		return true
	})

	return violations
}

func collectConstPositions(file *ast.File) map[token.Pos]bool {
	positions := make(map[token.Pos]bool)
	for _, declaration := range file.Decls {
		genDecl, ok := declaration.(*ast.GenDecl)
		if !ok || genDecl.Tok != token.CONST {
			continue
		}
		ast.Inspect(genDecl, func(node ast.Node) bool {
			if node != nil {
				positions[node.Pos()] = true
			}
			return true
		})
	}
	return positions
}
