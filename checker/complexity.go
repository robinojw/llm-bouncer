package checker

import (
	"fmt"
	"go/ast"
	"go/token"
)

const maxCyclomaticComplexity = 10

func checkCyclomaticComplexity(fileSet *token.FileSet, file *ast.File) []Violation {
	var violations []Violation

	for _, declaration := range file.Decls {
		funcDecl, ok := declaration.(*ast.FuncDecl)
		if !ok || funcDecl.Body == nil {
			continue
		}

		complexity := calculateComplexity(funcDecl)
		if complexity > maxCyclomaticComplexity {
			violations = append(violations, Violation{
				Line: fileSet.Position(funcDecl.Pos()).Line,
				Rule: "cyclomatic-complexity",
				Message: fmt.Sprintf(
					"function %q has complexity %d (max %d); break into smaller functions",
					funcDecl.Name.Name, complexity, maxCyclomaticComplexity,
				),
			})
		}
	}

	return violations
}

// calculateComplexity counts decision points using McCabe's formula.
func calculateComplexity(funcDecl *ast.FuncDecl) int {
	complexity := 1

	ast.Inspect(funcDecl.Body, func(node ast.Node) bool {
		switch typedNode := node.(type) {
		case *ast.IfStmt, *ast.ForStmt, *ast.RangeStmt:
			complexity++
		case *ast.CaseClause:
			if typedNode.List != nil {
				complexity++
			}
		case *ast.CommClause:
			if typedNode.Comm != nil {
				complexity++
			}
		case *ast.BinaryExpr:
			operator := typedNode.Op.String()
			if operator == "&&" || operator == "||" {
				complexity++
			}
		}
		return true
	})

	return complexity
}
