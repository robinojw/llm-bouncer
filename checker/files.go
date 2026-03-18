package checker

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
)

const (
	MaxFileLines = 300
)

var snakeCasePattern = regexp.MustCompile(`^[a-z][a-z0-9]*(_[a-z0-9]+)*\.go$`)

func checkFileName(filePath string) []Violation {
	base := filepath.Base(filePath)
	if !snakeCasePattern.MatchString(base) {
		return []Violation{{
			Rule:    "file-naming",
			Message: fmt.Sprintf("%q must use snake_case (e.g. my_handler.go)", base),
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
