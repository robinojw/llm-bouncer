package checker

import (
	"context"
	"fmt"
	"os"
	"strings"

	sitter "github.com/smacker/go-tree-sitter"
	"llm-bouncer/language"
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

// CheckFile runs all checks against a source file using tree-sitter.
func CheckFile(filePath string) []Violation {
	lang := language.Detect(filePath)
	if lang == nil {
		return nil
	}

	var violations []Violation

	violations = append(violations, checkFileName(filePath, lang)...)
	violations = append(violations, checkFileSize(filePath)...)

	src, err := os.ReadFile(filePath)
	if err != nil {
		return violations
	}

	parser := sitter.NewParser()
	parser.SetLanguage(lang.Language)

	tree, err := parser.ParseCtx(context.Background(), nil, src)
	if err != nil {
		return violations
	}

	root := tree.RootNode()

	violations = append(violations, checkNaming(root, src, lang)...)
	violations = append(violations, checkNestedIfs(root, src, lang)...)
	violations = append(violations, checkInlineBooleans(root, src, lang)...)
	violations = append(violations, checkInlineComments(root, src, lang)...)
	violations = append(violations, checkRepeatedStrings(root, src, lang)...)
	violations = append(violations, checkMagicNumbers(root, src, lang)...)
	violations = append(violations, checkCyclomaticComplexity(root, src, lang)...)

	return violations
}

// walk calls fn for every node in the tree rooted at n.
// If fn returns false, the children of that node are skipped.
func walk(n *sitter.Node, fn func(*sitter.Node) bool) {
	if n == nil {
		return
	}
	if !fn(n) {
		return
	}
	for i := 0; i < int(n.ChildCount()); i++ {
		walk(n.Child(i), fn)
	}
}

// nodeText returns the source text of a node.
func nodeText(n *sitter.Node, src []byte) string {
	return n.Content(src)
}

// startLine returns the 1-based line number of a node.
func startLine(n *sitter.Node) int {
	return int(n.StartPoint().Row) + 1
}

// nodeTypeSet builds a set from a slice for O(1) lookups.
func nodeTypeSet(types []string) map[string]bool {
	s := make(map[string]bool, len(types))
	for _, t := range types {
		s[t] = true
	}
	return s
}

// isDescendantOf checks if node n has an ancestor whose type is in the given set.
func isDescendantOf(n *sitter.Node, types map[string]bool) bool {
	for p := n.Parent(); p != nil; p = p.Parent() {
		if types[p.Type()] {
			return true
		}
	}
	return false
}

// lineContent returns the text of the line containing the node.
func lineContent(n *sitter.Node, src []byte) string {
	lines := strings.Split(string(src), "\n")
	row := int(n.StartPoint().Row)
	if row < 0 || row >= len(lines) {
		return ""
	}
	return lines[row]
}
