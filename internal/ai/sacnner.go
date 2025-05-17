package ai

import (
	"context"
	"github.com/abdelrahman146/kunai/utils"
	"github.com/tmc/langchaingo/documentloaders"
	"github.com/tmc/langchaingo/schema"
	"github.com/tmc/langchaingo/textsplitter"
	"io/fs"
	"log"
	"os"
	"path/filepath"
)

func ScanProject(projectPath string) ([]schema.Document, error) {
	var docs []schema.Document
	var err error
	utils.RunWithSpinner("Scanning project", func() {
		err = filepath.WalkDir(projectPath, func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if d.IsDir() {
				return nil
			}
			ext := filepath.Ext(path)
			if !utils.CanProcessFile(ext) {
				return nil
			}
			if !utils.CanProcessPath(path) {
				return nil
			}
			log.Printf("Reading %q\n", path)
			chunks, err := fileToChunks(path, 200, 50)
			docs = append(docs, chunks...)
			return nil
		})
	})
	return docs, err
}

func fileToChunks(filePath string, chunkSize, chunkOverlap int) ([]schema.Document, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	docLoaded := documentloaders.NewText(file)
	split := textsplitter.NewRecursiveCharacter()
	split.ChunkSize = chunkSize
	split.ChunkOverlap = chunkOverlap
	docs, err := docLoaded.LoadAndSplit(context.Background(), split)
	if err != nil {
		return nil, err
	}
	return docs, nil
}
