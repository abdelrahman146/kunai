package utils

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func FindRepoRoot() (string, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return "", err
	}
	for dir := cwd; dir != string(filepath.Separator); dir = filepath.Dir(dir) {
		if _, err := os.Stat(filepath.Join(dir, ".git")); err == nil {
			return dir, nil
		}
	}
	return "", fmt.Errorf(".git directory not found")
}

func GetAbsPath(relPath string) (string, error) {
	basePath, _ := os.Getwd()
	path, err := filepath.Rel(basePath, relPath)
	return path, err
}

// CanProcessPath checks if the dirname is blocklisted
func CanProcessPath(path string) bool {
	ignored := []string{"node_modules", ".git", ".idea", "dist", "build", "out", ".next"}
	for _, ignore := range ignored {
		if strings.Contains(path, ignore) {
			return false
		}
	}
	return true
}

// CanProcessFile checks if file ext is valid
func CanProcessFile(ext string) bool {
	switch ext {
	case ".go", ".js", ".jsx", ".mjs", ".ts", ".tsx", ".md", ".yaml", ".yml", ".env", ".example", ".json":
		return true
	default:
		return false
	}
}
