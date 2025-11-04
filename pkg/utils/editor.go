package utils

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func EditTextWithEditor(text string, ext string) (string, error) {
	tmpFile, err := os.CreateTemp("./tmp", "pcapstore-edit-*."+ext)
	if err != nil {
		return "", fmt.Errorf("failed to create temporary file: %v", err)
	}

	defer func() {
		if err := os.Remove(tmpFile.Name()); err != nil {
			fmt.Printf("failed to remove temporary file: %v\n", err)
		}
	}()

	if _, err := tmpFile.WriteString(text); err != nil {
		return "", fmt.Errorf("failed to write to temporary file: %v", err)
	}
	_ = tmpFile.Close()

	editor := os.Getenv("EDITOR")
	if editor == "" {
		editor = os.Getenv("VISUAL")
	}
	if editor == "" {
		editor = "vim"
	}

	if _, err := exec.LookPath(editor); err != nil {
		return "", fmt.Errorf("editor '%s' not found in PATH", editor)
	}

	editorName := strings.ToLower(filepath.Base(editor))
	specialEditors := map[string][]string{
		"code":          {"-", "--wait"},
		"code-insiders": {"-", "--wait"},
		"codium":        {"-", "--wait"},
		"atom":          {"-", "--wait"},
		"subl":          {"-", "--wait"},
		"sublime_text":  {"-", "--wait"},
		"zeditor":       {"-"},
	}

	var args []string
	var needsStdinInput bool

	for editorKey, flags := range specialEditors {
		if strings.Contains(editorName, editorKey) {
			args = flags
			needsStdinInput = true
			break
		}
	}

	if needsStdinInput {
		cmd := exec.Command(editor, args...)
		cmd.Stdin = strings.NewReader(text)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			return "", fmt.Errorf("failed to run editor: %v", err)
		}

		editedContent, err := os.ReadFile(tmpFile.Name())
		if err != nil {
			return "", fmt.Errorf("failed to read edited file: %v", err)
		}
		return string(editedContent), nil
	} else {
		cmd := exec.Command(editor, tmpFile.Name())
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			return "", fmt.Errorf("failed to run editor: %v", err)
		}

		editedContent, err := os.ReadFile(tmpFile.Name())
		if err != nil {
			return "", fmt.Errorf("failed to read edited file: %v", err)
		}
		return string(editedContent), nil
	}
}
