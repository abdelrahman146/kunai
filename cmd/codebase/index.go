package codebase

import (
	"encoding/json"
	"github.com/abdelrahman146/kunai/internal/es"
	"github.com/elastic/go-elasticsearch/v8"
	"github.com/spf13/cobra"
	"log"
	"sync"
)

var indexCmd = &cobra.Command{
	Use:   "index",
	Short: `index code base`,
	Long:  `start indexing code base to be available for search and stats commands`,
	RunE:  runIndexCmd,
}

var indexCmdParams struct {
	ElasticSearchURL string
	CodebaseDir      string
}

func init() {
	indexCmd.Flags().StringVarP(&indexCmdParams.CodebaseDir, "dir", "d", "", "codebase directory")
	indexCmd.Flags().StringVar(&indexCmdParams.ElasticSearchURL, "es-url", "http://localhost:9200", "elastic search url")
	if err := indexCmd.MarkFlagRequired("dir"); err != nil {
		log.Fatalln("codebase dir required")
	}
}

func runIndexCmdProjectScanner(id int, projectsChannel <-chan string, documentChannel chan<- es.Document, wg *sync.WaitGroup) {
	defer wg.Done()
	for projectPath := range projectsChannel {
		log.Printf("[Scanner %d]: Started scanning project %q \n", id, projectPath)
		if err := es.ScanProject(projectPath, documentChannel); err != nil {
			log.Printf("[Scanner %d]: Failed to scan project %q: %q\n", id, projectPath, err.Error())
			continue
		}
		log.Printf("[Scanner %d]: Successfully scanned project %q\n", id, projectPath)
	}
}

func runIndexCmd(cmd *cobra.Command, args []string) error {
	log.Printf("indexing started ... \n")
	// get mapping
	mapping, err := json.Marshal(es.IndexMapping)
	if err != nil {
		return err
	}

	projectCh := make(chan string) // project path
	documentCh := make(chan es.Document, 50)

	// start indexing
	if err := es.ReIndex(indexCmdParams.ElasticSearchURL, alias, mapping, func(esClient *elasticsearch.Client, indexName string) error {

		// Prepare ProjectScanners
		workers := 5
		var scannerWg sync.WaitGroup
		scannerWg.Add(workers)
		for i := 0; i < workers; i++ {
			go runIndexCmdProjectScanner(i+1, projectCh, documentCh, &scannerWg)
		}

		// Prepare Indexer (1)
		var indexerWg sync.WaitGroup
		indexerWg.Add(1)
		go func() {
			defer indexerWg.Done()
			for doc := range documentCh {
				if err := es.IndexDocument(esClient, indexName, doc); err != nil {
					log.Printf("[Indexer]: Failed to index document %v\n", err)
				}
			}
		}()

		log.Printf("Gathering projects from %q\n", indexCmdParams.CodebaseDir)
		if err = es.GetProjects(indexCmdParams.CodebaseDir, projectCh); err != nil {
			log.Printf("Failed to get projects: %v\n", err)
		}

		close(projectCh)
		// wait for scanners to finish before closing the document channel
		scannerWg.Wait()
		close(documentCh)
		// wait for indexer to finish
		indexerWg.Wait()
		log.Printf("Indexing completed. new index is: %q\n", indexName)
		return nil
	}); err != nil {
		return err
	}
	return nil
}
