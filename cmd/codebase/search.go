package codebase

import (
	"encoding/json"
	"fmt"
	"github.com/abdelrahman146/kunai/internal/es"
	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
	"log"
	"os"
)

var searchCmd = &cobra.Command{
	Use:   "search",
	Short: `Search within codebase`,
	Long:  `Search for anything within indexed codebase. you need to run "search index" before using this command`,
	RunE:  runSearchCmd,
}

var searchCmdParams struct {
	ElasticSearchURL string
	Query            string
	ProjectFilter    string
}

func init() {
	searchCmd.Flags().StringVarP(&searchCmdParams.Query, "query", "q", "", "search query")
	if err := searchCmd.MarkFlagRequired("query"); err != nil {
		log.Fatalln("search query required")
	}
	searchCmd.Flags().StringVarP(&searchCmdParams.ProjectFilter, "project", "p", "", "filter by project")
}

func runSearchCmd(cmd *cobra.Command, args []string) error {
	searchCmdParams.ElasticSearchURL = "http://localhost:9200"
	esClient, _ := es.NewClient(searchCmdParams.ElasticSearchURL)
	// build search query
	body := map[string]map[string]interface{}{
		"query": {
			"bool": map[string]interface{}{
				"must": []interface{}{
					map[string]interface{}{
						"multi_match": map[string]interface{}{
							"query":     searchCmdParams.Query,
							"fields":    []string{"content^4", "content.ngram^2", "name^2"},
							"type":      "best_fields",
							"operator":  "and",
							"fuzziness": "AUTO",
						},
					},
				},
				"filter": []interface{}{},
			},
		},
	}
	if searchCmdParams.ProjectFilter != "" {
		boolQ := body["query"]
		boolQ["filter"] = []interface{}{
			map[string]interface{}{
				"term": map[string]interface{}{
					"project": map[string]interface{}{
						"value": searchCmdParams.ProjectFilter,
					},
				},
			},
		}
	}
	buf, _ := json.Marshal(body)
	r, err := es.Search[es.Document](esClient, alias, buf)
	if err != nil {
		return err
	}
	// Pretty-print table
	table := tablewriter.NewWriter(os.Stdout)
	table.Header([]string{"Project", "Name", "Path", "Ext", "Language", "Size (KB)", "Updated At"})
	for _, hit := range r.Hits.Hits {
		size := fmt.Sprintf("%.2f", float64(hit.Source.Size)/1024)
		err := table.Append([]string{hit.Source.Project, hit.Source.Name, hit.Source.RelPath, hit.Source.Extension, hit.Source.Language, size, hit.Source.UpdatedAt.Format("2006-01-02 15:04")})
		if err != nil {
			return err
		}
	}
	return table.Render()
}
