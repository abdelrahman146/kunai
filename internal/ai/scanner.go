package ai

import (
	"context"
	"github.com/abdelrahman146/kunai/internal/es"
	"github.com/abdelrahman146/kunai/utils"
	"github.com/tmc/langchaingo/documentloaders"
	"github.com/tmc/langchaingo/schema"
	"github.com/tmc/langchaingo/textsplitter"
	"github.com/tmc/langchaingo/vectorstores"
	"io/fs"
	"os"
	"path/filepath"
	"sync"
)

func ScanProject(projectPath string, chunkSize, chunkOverlap int, store vectorstores.VectorStore) error {
	ctx := context.Background()
	var err error
	docsCh := make(chan schema.Document)
	var docsWg sync.WaitGroup
	workers := 5
	docsWg.Add(workers)
	for i := 0; i < workers; i++ {
		go func() {
			defer docsWg.Done()
			var batch []schema.Document
			batchSize := 10
			for doc := range docsCh {
				if len(batch) < batchSize {
					batch = append(batch, doc)
				} else {
					batch = append(batch, doc)
					err = StoreDocuments(ctx, batch, store)
					if err != nil {
						return
					}
					batch = batch[:0]
				}
			}
			// store whatever left before stopping
			if len(batch) > 0 {
				err = StoreDocuments(ctx, batch, store)
			}
		}()
	}
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
		chunks, err := fileToChunks(projectPath, path, chunkSize, chunkOverlap)
		if err != nil {
			return err
		}
		for _, chunk := range chunks {
			docsCh <- chunk
		}
		return nil
	})
	close(docsCh)
	docsWg.Wait()
	return err
}

func fileToChunks(projectPath, filePath string, chunkSize, chunkOverlap int) ([]schema.Document, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	docLoaded := documentloaders.NewText(file)
	split := textsplitter.NewRecursiveCharacter()
	split.ChunkSize = chunkSize
	split.ChunkOverlap = chunkOverlap
	docs, err := docLoaded.LoadAndSplit(context.Background(), split)
	relPath, _ := filepath.Rel(projectPath, filePath)
	for _, doc := range docs {
		doc.Metadata["path"] = relPath
		doc.Metadata["fileName"] = filepath.Base(filePath)
		ext := filepath.Ext(filePath)
		doc.Metadata["ext"] = ext
		doc.Metadata["language"] = es.InferLanguage(ext)
	}
	if err != nil {
		return nil, err
	}
	return docs, nil
}
