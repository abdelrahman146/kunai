package codebase

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/abdelrahman146/kunai/internal/es"
	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
	"os"
)

var statsCmd = &cobra.Command{
	Use:   "stats",
	Short: "Show codebase statistics & aggregations",
	Long:  `Compute counts by project, extension, language and file-size stats across your indexed codebase.`,
	RunE:  runStatsCmd,
}

var statsCmdParams struct {
	ElasticSearchURL string
	ProjectFilter    string
}

func init() {
	statsCmd.Flags().StringVarP(&statsCmdParams.ProjectFilter, "project", "p", "", "only include this project")
	statsCmd.Flags().StringVar(&statsCmdParams.ElasticSearchURL, "es-url", "http://localhost:9200", "elastic search url")
}

func runStatsCmd(cmd *cobra.Command, args []string) error {
	esClient, err := es.NewClient(statsCmdParams.ElasticSearchURL)
	if err != nil {
		return err
	}

	// Build the aggregation query
	body := map[string]interface{}{
		"size": 0, // no hits, just aggs
		"query": map[string]interface{}{
			"bool": map[string]interface{}{
				"filter": []interface{}{},
			},
		},
		"aggs": map[string]interface{}{
			"by_project": map[string]interface{}{
				"terms": map[string]interface{}{
					"field": "project",
					"size":  50,
				},
			},
			"by_extension": map[string]interface{}{
				"terms": map[string]interface{}{
					"field": "extension",
					"size":  20,
				},
			},
			"by_language": map[string]interface{}{
				"terms": map[string]interface{}{
					"field": "language",
					"size":  20,
				},
			},
			"size_stats": map[string]interface{}{
				"stats": map[string]interface{}{
					"field": "size",
				},
			},
		},
	}

	// apply project filter if requested
	if statsCmdParams.ProjectFilter != "" {
		f := body["query"].(map[string]interface{})["bool"].(map[string]interface{})["filter"].([]interface{})
		f = append(f, map[string]interface{}{
			"term": map[string]interface{}{
				"project": statsCmdParams.ProjectFilter,
			},
		})
		body["query"].(map[string]interface{})["bool"].(map[string]interface{})["filter"] = f
	}

	// serialize
	buf, err := json.Marshal(body)
	if err != nil {
		return err
	}

	// execute
	res, err := esClient.Search(
		esClient.Search.WithContext(context.Background()),
		esClient.Search.WithIndex(alias),
		esClient.Search.WithBody(bytes.NewReader(buf)),
		esClient.Search.WithTrackTotalHits(true),
	)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	// decode only aggregations
	var r struct {
		Aggregations struct {
			ByProject struct {
				Buckets []struct {
					Key      string `json:"key"`
					DocCount int    `json:"doc_count"`
				} `json:"buckets"`
			} `json:"by_project"`
			ByExtension struct {
				Buckets []struct {
					Key      string `json:"key"`
					DocCount int    `json:"doc_count"`
				} `json:"buckets"`
			} `json:"by_extension"`
			ByLanguage struct {
				Buckets []struct {
					Key      string `json:"key"`
					DocCount int    `json:"doc_count"`
				} `json:"buckets"`
			} `json:"by_language"`
			SizeStats struct {
				Count int     `json:"count"`
				Min   float64 `json:"min"`
				Max   float64 `json:"max"`
				Avg   float64 `json:"avg"`
				Sum   float64 `json:"sum"`
			} `json:"size_stats"`
		} `json:"aggregations"`
	}
	if err := json.NewDecoder(res.Body).Decode(&r); err != nil {
		return err
	}

	// print project counts
	fmt.Println("\nFiles per Project:")
	t1 := tablewriter.NewWriter(os.Stdout)
	t1.Header([]string{"Project", "Count"})
	for _, b := range r.Aggregations.ByProject.Buckets {
		t1.Append([]string{b.Key, fmt.Sprintf("%d", b.DocCount)})
	}
	t1.Render()

	// print extension counts
	fmt.Println("\nFiles per Extension:")
	t2 := tablewriter.NewWriter(os.Stdout)
	t2.Header([]string{"Extension", "Count"})
	for _, b := range r.Aggregations.ByExtension.Buckets {
		t2.Append([]string{b.Key, fmt.Sprintf("%d", b.DocCount)})
	}
	t2.Render()

	// print language counts
	fmt.Println("\nFiles per Language:")
	t3 := tablewriter.NewWriter(os.Stdout)
	t3.Header([]string{"Language", "Count"})
	for _, b := range r.Aggregations.ByLanguage.Buckets {
		t3.Append([]string{b.Key, fmt.Sprintf("%d", b.DocCount)})
	}
	t3.Render()

	// print size stats
	fmt.Println("\nFile Size Stats (bytes):")
	t4 := tablewriter.NewWriter(os.Stdout)
	t4.Header([]string{"Metric", "Value"})
	stats := r.Aggregations.SizeStats
	t4.Append([]string{"count", fmt.Sprintf("%d", stats.Count)})
	t4.Append([]string{"min", fmt.Sprintf("%.2f KB", stats.Min/1024)})
	t4.Append([]string{"max", fmt.Sprintf("%.2f KB", stats.Max/1024)})
	t4.Append([]string{"avg", fmt.Sprintf("%.2f KB", stats.Avg/1024)})
	t4.Append([]string{"sum", fmt.Sprintf("%.2f KB", stats.Sum/1024)})
	t4.Render()

	return nil
}
