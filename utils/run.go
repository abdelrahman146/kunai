package utils

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"time"
)

func RunCLICommand(name string, args ...string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Minute)
	defer cancel()
	cmd := exec.CommandContext(ctx, name, args...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return "", err
	}
	return string(out), nil
}

func Edit(text string) (string, error) {
	tmp, err := os.CreateTemp("", "edit-*.txt")
	if err != nil {
		return text, fmt.Errorf("failed to create temp file: %w", err)
	}
	defer os.Remove(tmp.Name())
	defer tmp.Close()
	if _, err := io.Copy(tmp, bytes.NewBufferString(text)); err != nil {
		return text, fmt.Errorf("error writing to temp file: %v\n", err)
	}
	// honor $EDITOR, fallback to vim
	editor := os.Getenv("EDITOR")
	if editor == "" {
		editor = "vim"
	}
	cmd := exec.Command(editor, tmp.Name())
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return text, fmt.Errorf("failed to edit: %w", err)
	}
	edited, err := os.ReadFile(tmp.Name())
	if err != nil {
		return text, fmt.Errorf("failed to read edit file: %w", err)
	}
	return string(edited), nil
}
