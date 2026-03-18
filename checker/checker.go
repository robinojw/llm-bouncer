package checker

import (
	"fmt"
	"go/parser"
	"go/token"
	"os"
)

// Violation describes a single code quality issue.
type Violation struct {
	Line    int
	Rule    string
	Message string
}

func (v Violation) String() string {
	if v.Line > 0 {
		return fmt.Sprintf("  line %d [%s] %s", v.Line, v.Rule, v.Message)
	}
	return fmt.Sprintf("  [%s] %s", v.Rule, v.Message)
}

// CheckFile runs all checks against a Go source file.
func CheckFile(filePath string) []Violation {
	var violations []Violation

	violations = append(violations, checkFileName(filePath)...)
	violations = append(violations, checkFileSize(filePath)...)

	src, err := os.ReadFile(filePath)
	if err != nil {
		return violations
	}

	fileSet := token.NewFileSet()
	file, err := parser.ParseFile(fileSet, filePath, src, parser.ParseComments)
	if err != nil {
		return violations
	}

	violations = append(violations, checkNaming(fileSet, file)...)
	violations = append(violations, checkNestedIfs(fileSet, file)...)
	violations = append(violations, checkInlineBooleans(fileSet, file)...)
	violations = append(violations, checkInlineComments(fileSet, file, src)...)
	violations = append(violations, checkRepeatedStrings(fileSet, file)...)
	violations = append(violations, checkMagicNumbers(fileSet, file)...)
	violations = append(violations, checkCyclomaticComplexity(fileSet, file)...)

	return violations
}
