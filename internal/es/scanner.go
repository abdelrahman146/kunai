package es

import (
	"github.com/abdelrahman146/kunai/utils"
	"io/fs"
	"os"
	"path/filepath"
	"time"
)

// Document holds metadata per file.
type Document struct {
	Name      string    `json:"name"`
	Project   string    `json:"project"`
	RelPath   string    `json:"relPath"`
	Extension string    `json:"extension"`
	Content   string    `json:"content"`
	Language  string    `json:"language"`
	Size      int64     `json:"size"`
	UpdatedAt time.Time `json:"updatedAt"`
}

func GetProjects(rootDir string, projectCh chan<- string) error {
	return filepath.WalkDir(rootDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			if utils.CanProcessPath(path) && isProjectRoot(path) {
				projectCh <- path
			}
		}
		return nil
	})
}

func ScanProject(projectPath string, documentCh chan<- Document) error {
	projName := filepath.Base(projectPath)
	return filepath.WalkDir(projectPath, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		// Only process acceptable source extension
		ext := filepath.Ext(path)
		if !utils.CanProcessFile(ext) {
			return nil
		}
		if !utils.CanProcessPath(path) {
			return nil
		}

		// Compute the path *within* that project
		relToProj, _ := filepath.Rel(projectPath, path)

		// Gather file info & content
		info, err := os.Stat(path)
		if err != nil {
			return err
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}

		doc := Document{
			Name:      info.Name(),
			Project:   projName,
			RelPath:   relToProj,
			Extension: ext,
			Language:  inferLanguage(ext),
			Content:   string(data),
			Size:      info.Size(),
			UpdatedAt: info.ModTime(),
		}
		documentCh <- doc
		return nil
	})
}

// isProjectRoot returns true if the directory contains one of your markers.
func isProjectRoot(dirPath string) bool {
	if fileExists(filepath.Join(dirPath, "package.json")) {
		return !fileExists(filepath.Join(dirPath, "turbo.json")) && !fileExists(filepath.Join(dirPath, "nx.json")) && !fileExists(filepath.Join(dirPath, "lerna.json"))
	} else if fileExists(filepath.Join(dirPath, "go.mod")) {
		return !fileExists(filepath.Join(dirPath, "go.work"))
	}
	return false
}

// fileExists is a small helper to check for existence.
func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func inferLanguage(ext string) string {
	switch ext {
	case ".go":
		return "go"
	case ".js", ".jsx", ".mjs":
		return "javascript"
	case ".ts", ".tsx":
		return "typescript"
	case ".md":
		return "markdown"
	case ".env":
		return "env"
	case ".yaml", ".yml":
		return "yaml"
	case ".json":
		return "json"
	default:
		return "unknown"
	}
}
