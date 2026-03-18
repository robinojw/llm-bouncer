package checker

import (
	"fmt"
	"go/ast"
	"go/token"
)

var acceptableShortIdents = map[string]bool{
	"i": true, "j": true, "k": true, "_": true,
}

func checkNaming(fileSet *token.FileSet, file *ast.File) []Violation {
	var violations []Violation
	receiverNames := collectReceiverNames(file)

	ast.Inspect(file, func(node ast.Node) bool {
		switch declaration := node.(type) {
		case *ast.AssignStmt:
			for _, expression := range declaration.Lhs {
				if ident, ok := expression.(*ast.Ident); ok {
					if isTooShort(ident.Name) && !receiverNames[ident.Name] {
						violations = append(violations, Violation{
							Line:    fileSet.Position(ident.Pos()).Line,
							Rule:    "naming",
							Message: fmt.Sprintf("variable %q is too short; use a descriptive name", ident.Name),
						})
					}
				}
			}
		case *ast.Field:
			for _, name := range declaration.Names {
				if isTooShort(name.Name) && !receiverNames[name.Name] {
					violations = append(violations, Violation{
						Line:    fileSet.Position(name.Pos()).Line,
						Rule:    "naming",
						Message: fmt.Sprintf("parameter %q is too short; use a descriptive name", name.Name),
					})
				}
			}
		}
		return true
	})

	return violations
}

func collectReceiverNames(file *ast.File) map[string]bool {
	names := make(map[string]bool)
	for _, declaration := range file.Decls {
		funcDecl, ok := declaration.(*ast.FuncDecl)
		if !ok || funcDecl.Recv == nil {
			continue
		}
		for _, field := range funcDecl.Recv.List {
			for _, name := range field.Names {
				names[name.Name] = true
			}
		}
	}
	return names
}

func isTooShort(name string) bool {
	return len(name) == 1 && !acceptableShortIdents[name]
}
