package ai

import (
	"context"
	"fmt"
	"github.com/abdelrahman146/kunai/internal/es"
	"github.com/abdelrahman146/kunai/utils"
	"github.com/tmc/langchaingo/documentloaders"
	"github.com/tmc/langchaingo/schema"
	"github.com/tmc/langchaingo/textsplitter"
	"github.com/tmc/langchaingo/vectorstores"
	"io/fs"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
)

func ScanProject(projectPath string, chunkSize, chunkOverlap int, store vectorstores.VectorStore) error {
	ctx := context.Background()
	var err error
	var docsWg sync.WaitGroup
	workers := runtime.NumCPU() * 2
	batchSize := 100
	docsCh := make(chan schema.Document, workers*batchSize)
	docsWg.Add(workers)
	for i := 0; i < workers; i++ {
		go func() {
			defer docsWg.Done()
			var batch []schema.Document
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
	var paths []string
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
		rel, _ := filepath.Rel(projectPath, path)
		paths = append(paths, rel)
		chunks, err := fileToDocuments(projectPath, path, chunkSize, chunkOverlap)
		if err != nil {
			return err
		}

		for _, chunk := range chunks {
			docsCh <- chunk
		}
		return nil
	})
	// Add table of content
	toc := schema.Document{
		PageContent: "// PROJECT TOC\n" + strings.Join(paths, "\n"),
		Metadata:    map[string]any{"type": "toc"},
	}
	docsCh <- toc
	close(docsCh)
	docsWg.Wait()
	return err
}

func fileToDocuments(projectPath, filePath string, chunkSize, chunkOverlap int) ([]schema.Document, error) {
	file, err := os.Open(filePath)
	defer file.Close()
	if err != nil {
		return nil, err
	}
	docLoaded := documentloaders.NewText(file)
	fileInfo, err := os.Stat(filePath)
	if err != nil {
		return nil, err
	}
	size := fileInfo.Size()
	relPath, _ := filepath.Rel(projectPath, filePath)
	ext := filepath.Ext(filePath)
	meta := map[string]any{
		"path":     relPath,
		"dir":      filepath.Dir(relPath),
		"fileName": filepath.Base(relPath),
		"ext":      filepath.Ext(relPath),
		"language": es.InferLanguage(ext),
		"isTest":   strings.Contains(strings.ToLower(relPath), "test"),
	}
	var docs []schema.Document
	if size <= 50000 {
		docs, err = docLoaded.Load(context.Background())
	} else {
		split := textsplitter.NewRecursiveCharacter()
		split.ChunkSize = chunkSize
		split.ChunkOverlap = chunkOverlap
		docs, err = docLoaded.LoadAndSplit(context.Background(), split)
	}
	if err != nil {
		return nil, err
	}
	var enhancedDocs []schema.Document
	for _, doc := range docs {
		var enhancedDoc schema.Document
		hdr := fmt.Sprintf(
			"// FILE: %s\n// DIR: %s\n// LANG: %s\n// EXTENSION: %s\n// TEST: %v\n\n",
			meta["fileName"], meta["dir"], meta["language"], meta["ext"], meta["isTest"],
		)
		enhancedDoc.PageContent = hdr + doc.PageContent
		enhancedDoc.Metadata = meta
		enhancedDoc.Score = doc.Score
		enhancedDocs = append(enhancedDocs, enhancedDoc)
	}
	return enhancedDocs, nil
}
