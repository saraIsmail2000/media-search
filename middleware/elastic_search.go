package middleware

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/elastic/go-elasticsearch/v7"
	elasticsearchapi "github.com/elastic/go-elasticsearch/v7/esapi"
	"github.com/sirupsen/logrus"
	"log"
	"media-search/constants"
	"strconv"
	"time"
)

// AddMovieOrBookToIndex creates or updates a task in an index.
func AddMovieOrBookToIndex(ctx context.Context, client *elasticsearch.Client, body IndexedData, index string) {
	var buf bytes.Buffer

	if err := json.NewEncoder(&buf).Encode(body); err != nil {
		logrus.Infof("failed to encode record %v into reader for indexing it with err: %v", body.ID, err.Error())
	}

	req := elasticsearchapi.IndexRequest{
		Index:      index,
		Body:       &buf,
		DocumentID: strconv.Itoa(body.ID),
		Refresh:    "true",
	}

	resp, err := req.Do(ctx, client)
	if err != nil {
		logrus.Infof("failed to create record %v at elastic search %v index with error: %v", body.ID, index, err.Error())
	}
	defer resp.Body.Close()

	if resp.IsError() {
		logrus.Infof("failed to create record %v at elastic search %v index with error: %v", body.ID, index, err.Error())
	}
}

func applyElasticSearch(ctx context.Context, client *elasticsearch.Client, searchQuery string) ([]interface{}, error) {

	// Define the query
	query := map[string]interface{}{
		"size": 100, // Set the size parameter to 100 to retrieve up to 100 matches
		"query": map[string]interface{}{
			"multi_match": map[string]interface{}{
				"query":  searchQuery,
				"fields": []string{"title", "writers", "directors", "cast", "authors"},
			},
		},
	}

	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(query); err != nil {
		return nil, err
	}

	// get data from searching both indexes
	req := elasticsearchapi.SearchRequest{
		Index: []string{constants.ElasticSearchBooksIndex, constants.ElasticSearchMoviesIndex},
		Body:  &buf,
	}

	resp, err := req.Do(ctx, client)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var hits struct {
		Hits struct {
			Hits []struct {
				Source interface{} `json:"_source"`
			} `json:"hits"`
		} `json:"hits"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&hits); err != nil {
		fmt.Println("Error here", err)
		return nil, err
	}

	// Extract search results and return them
	var results []interface{}
	for _, hit := range hits.Hits.Hits {
		results = append(results, hit.Source)
	}
	return results, nil
}

func DeleteDocumentsFromIndex(client *elasticsearch.Client, index string, ids []int) {
	// Create the delete by query request body
	query := map[string]interface{}{
		"query": map[string]interface{}{
			"terms": map[string]interface{}{
				"_id": ids,
			},
		},
	}

	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(query); err != nil {
		log.Fatalf("Error executing delete by query request: %v", err)
	}

	// Execute the delete by query request
	res, err := client.DeleteByQuery(
		[]string{index},
		&buf,
		client.DeleteByQuery.WithContext(context.Background()),
		client.DeleteByQuery.WithTimeout(60*time.Second),
	)
	if err != nil {
		log.Fatalf("Error executing delete by query request: %v", err)
	}
	defer res.Body.Close()

	// Check the response status
	if res.IsError() {
		log.Fatalf("Delete by query request failed with status code: %d", res.StatusCode)
	}
}
