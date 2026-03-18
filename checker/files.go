package checker

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"

	"llm-bouncer/language"
)

const (
	MaxFileLines = 300
)

func checkFileName(filePath string, lang *language.LanguageConfig) []Violation {
	base := filepath.Base(filePath)
	if lang.FileNameRegex != nil && !lang.FileNameRegex.MatchString(base) {
		return []Violation{{
			Rule:    "file-naming",
			Message: fmt.Sprintf("%q does not match %s naming convention", base, lang.Name),
		}}
	}
	return nil
}

func checkFileSize(filePath string) []Violation {
	file, err := os.Open(filePath)
	if err != nil {
		return nil
	}
	defer file.Close()

	lineCount := 0
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lineCount++
	}

	if lineCount > MaxFileLines {
		return []Violation{{
			Rule:    "file-size",
			Message: fmt.Sprintf("%d lines exceeds the %d line limit; split into smaller files", lineCount, MaxFileLines),
		}}
	}
	return nil
}
