package es

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/elastic/go-elasticsearch/v8"
	"log"
	"time"
)

func NewClient(addr string) (*elasticsearch.Client, error) {
	return elasticsearch.NewClient(elasticsearch.Config{
		Addresses: []string{addr},
	})
}

// ReIndex atomically swaps the alias to a fresh index using functional options.
func ReIndex(esURL, alias string, mapping []byte, indexFn func(esClient *elasticsearch.Client, indexName string) error) error {
	// 1. Connect to elastic search
	esClient, err := NewClient(esURL)
	if err != nil {
		return err
	}
	log.Printf("1/5) Connected to %q", esURL)
	// 2. Create a timestamped index
	indexName := fmt.Sprintf("%s-%d", alias, time.Now().Unix())
	if err := CreateIndex(esClient, indexName, mapping); err != nil {
		return err
	}
	log.Printf("2/5) Created new index %s\n", indexName)
	// 3. Index documents
	log.Printf("3/5) Started indexing index %s\n", indexName)
	if err := indexFn(esClient, indexName); err != nil {
		return err
	}
	time.Sleep(2 * time.Second)
	var count int
	if count, err = CountDocs(esClient, indexName); err != nil {
		return err
	}
	log.Printf("3/5) Finished indexing index %q, Total Documents: %d\n", indexName, count)
	// Get the previous indices list
	previousIndices, err := ResolveAlias(esClient, alias)
	if err != nil {
		return err
	}
	// 4. Swap alias to new index
	if err := SwapAlias(esClient, indexName, alias); err != nil {
		return err
	}
	log.Printf("4/5) Alias %q now points to %q\n", alias, indexName)
	// 5. Delete previous indices
	if err := DeleteIndex(esClient, previousIndices); err != nil {
		return err
	}
	log.Printf("5/5) deleted previous indices %+v\n", previousIndices)
	return nil
}

func CreateIndex(esClient *elasticsearch.Client, indexName string, mapping []byte) error {
	if res, err := esClient.Indices.Create(indexName, esClient.Indices.Create.WithBody(bytes.NewReader(mapping))); err != nil {
		return fmt.Errorf("create index error: %w", err)
	} else {
		defer res.Body.Close()
		if res.IsError() {
			return fmt.Errorf("create index API error: %s", res.String())
		}
	}
	return nil
}

// ResolveAlias fetches all indices behind given alias.
func ResolveAlias(es *elasticsearch.Client, alias string) ([]string, error) {
	res, err := es.Indices.GetAlias(es.Indices.GetAlias.WithName(alias))
	if err != nil {
		return nil, fmt.Errorf("alias get error: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		return []string{}, nil
	}

	// Parse response: map[indexName] → alias info
	var data map[string]interface{}
	if err := json.NewDecoder(res.Body).Decode(&data); err != nil {
		return nil, fmt.Errorf("failed to parse alias response: %w", err)
	}

	var indices []string
	for idx := range data {
		indices = append(indices, idx)
	}
	return indices, nil
}

func DeleteIndex(esClient *elasticsearch.Client, indices []string) error {
	if res, err := esClient.Indices.Delete(indices); err != nil {
		return fmt.Errorf("delete index error: %w", err)
	} else {
		defer res.Body.Close()
		if res.IsError() {
			return fmt.Errorf("delete index response error: %s", res.String())
		}
	}
	return nil
}

func CountDocs(es *elasticsearch.Client, index string) (count int, err error) {
	res, err := es.Count(es.Count.WithIndex(index))
	if err != nil {
		return 0, fmt.Errorf("error getting count response: %w", err)
	}
	defer res.Body.Close()
	if res.IsError() {
		return 0, fmt.Errorf("error response: %s", res.String())
	}
	var body struct {
		Count int `json:"count"`
	}
	if err := json.NewDecoder(res.Body).Decode(&body); err != nil {
		return 0, fmt.Errorf("error parsing response: %w", err)
	}
	return body.Count, nil
}

func IndexDocument(esClient *elasticsearch.Client, index string, doc Document) error {
	body, err := json.Marshal(doc)
	if err != nil {
		return err
	}
	if res, err := esClient.Index(index, bytes.NewReader(body)); err != nil {
		return fmt.Errorf("error indexing document: %w", err)
	} else {
		defer res.Body.Close()
		if res.IsError() {
			return fmt.Errorf("%s: %s", res.Status(), res.String())
		}
	}
	return nil
}

func SwapAlias(esClient *elasticsearch.Client, indexName string, alias string) error {
	body := map[string]interface{}{
		"actions": []map[string]map[string]string{
			{
				"remove": {"alias": alias, "index": "*"},
			},
			{
				"add": {"alias": alias, "index": indexName},
			},
		},
	}
	buf := new(bytes.Buffer)
	if err := json.NewEncoder(buf).Encode(body); err != nil {
		return err
	}
	aliasRes, err := esClient.Indices.UpdateAliases(buf)
	if err != nil {
		return fmt.Errorf("update-aliases error: %w", err)
	}
	defer aliasRes.Body.Close()
	if aliasRes.IsError() {
		return fmt.Errorf("update-aliases response error: %s", aliasRes.String())
	}
	return nil
}

type Hit[T any] struct {
	Source T `json:"_source"`
}
type HitsBucket[T any] struct {
	Hits []Hit[T] `json:"hits"`
}
type Response[T any] struct {
	Hits HitsBucket[T] `json:"hits"`
}

func Search[T any](esClient *elasticsearch.Client, alias string, body []byte) (*Response[T], error) {
	res, err := esClient.Search(esClient.Search.WithIndex(alias), esClient.Search.WithBody(bytes.NewReader(body)), esClient.Search.WithTimeout(5*time.Second))
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	var r Response[T]
	if err := json.NewDecoder(res.Body).Decode(&r); err != nil {
		return nil, err
	}
	return &r, nil
}
