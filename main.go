package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"llm-bouncer/checker"
	"llm-bouncer/language"
)

type HookInput struct {
	SessionID string          `json:"session_id"`
	CWD       string          `json:"cwd"`
	HookEvent string          `json:"hook_event_name"`
	ToolName  string          `json:"tool_name"`
	ToolInput json.RawMessage `json:"tool_input"`
}

type FileToolInput struct {
	FilePath string `json:"file_path"`
}

type HookOutput struct {
	Decision string `json:"decision"`
	Reason   string `json:"reason,omitempty"`
}

func main() {
	filePath, err := resolveFilePath()
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(0)
	}

	if filePath == "" || !language.Supported(filePath) {
		os.Exit(0)
	}

	violations := checker.CheckFile(filePath)
	if len(violations) == 0 {
		os.Exit(0)
	}

	output := HookOutput{
		Decision: "block",
		Reason:   buildReport(filePath, violations),
	}

	json.NewEncoder(os.Stdout).Encode(output)
}

func resolveFilePath() (string, error) {
	if len(os.Args) > 1 {
		return os.Args[1], nil
	}

	data, err := io.ReadAll(os.Stdin)
	if err != nil {
		return "", fmt.Errorf("failed to read stdin: %w", err)
	}
	if len(data) == 0 {
		return "", nil
	}

	var input HookInput
	if err := json.Unmarshal(data, &input); err != nil {
		return "", fmt.Errorf("failed to parse hook input: %w", err)
	}

	var fileInput FileToolInput
	if err := json.Unmarshal(input.ToolInput, &fileInput); err != nil {
		return "", fmt.Errorf("failed to parse tool input: %w", err)
	}

	return fileInput.FilePath, nil
}

func buildReport(filePath string, violations []checker.Violation) string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Code quality violations in %s:\n\n", filepath.Base(filePath)))
	for _, violation := range violations {
		sb.WriteString(violation.String())
		sb.WriteByte('\n')
	}
	sb.WriteString("\nFix all violations before proceeding.")
	return sb.String()
}
