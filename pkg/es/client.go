package es

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/elastic/go-elasticsearch/v7"
	"github.com/elastic/go-elasticsearch/v7/esapi"
	"github.com/sirupsen/logrus"
)

// ESClient represents a client for interacting with Elasticsearch
type ESClient struct {
	es *elasticsearch.Client
}
type ESConfig struct {
	Urls                []string
	IdleConnTimeout     int
	MaxIdleConnsPerHost int
	MaxIdleConns        int
}

// NewESClient creates a new instance of ESClient
func NewESClient(config *ESConfig) (*ESClient, error) {
	cfg := elasticsearch.Config{
		Addresses: config.Urls,
		Transport: &http.Transport{
			MaxIdleConnsPerHost: config.MaxIdleConnsPerHost,
			MaxIdleConns:        config.MaxIdleConns,
			IdleConnTimeout:     time.Duration(time.Duration(config.IdleConnTimeout) * time.Second),
			// ExpectContinueTimeout //
		},
	}

	es, err := elasticsearch.NewClient(cfg)
	if err != nil {
		return nil, err
	}

	return &ESClient{es: es}, nil
}

func (c *ESClient) Ping() error {
	res, err := c.es.Ping()

	if err != nil {
		return err
	}

	if res.StatusCode != 200 {
		return fmt.Errorf("Ping status is %v", res.StatusCode)
	}

	return nil
}

func (c *ESClient) CreateIndex(indexName string) error {

	// Create the index request
	res, err := c.es.Indices.Create(indexName)
	if err != nil {
		log.Fatalf("Error creating index: %s", err)
		return err
	}

	if res.IsError() {
		return fmt.Errorf("error creating index: %s", res.String())
	}

	// sess, err := session.NewSession(&aws.Config{
	// 	Region: aws.String("your-aws-region"),
	// })
	// if err != nil {
	// 	log.Fatalf("Error creating AWS session: %v", err)
	// }

	// Create Elasticsearch Service client
	// svc := elasticsearchservice.New(sess)

	// createIndexInput := &elasticsearchservice.CreateIndexInput{
	// 	DomainName: aws.String("your-es-domain-name"),
	// 	IndexName:  aws.String(indexName),
	// }

	// // Create index
	// _, err = svc.CreateIndex(createIndexInput)
	// if err != nil {
	// 	log.Fatalf("Error creating index: %v", err)
	// }

	// fmt.Printf("Index '%s' created successfully.\n", indexName)

	fmt.Println("Index created successfully")
	return nil
}

// IndexDocument indexes a document in Elasticsearch
func (c *ESClient) IndexDocument(index string, documentID string, doc interface{}) error {
	// Serialize document to JSON
	body, err := json.Marshal(doc)
	if err != nil {
		return err
	}

	reader := bytes.NewReader(body)

	// Create index request
	req := esapi.IndexRequest{
		Index:      index,
		DocumentID: documentID,
		Body:       reader,
		Refresh:    "true",
	}

	// Perform index request
	res, err := req.Do(context.Background(), c.es)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.IsError() {
		return fmt.Errorf("failed to index document: %s", res.Status())
	}

	return nil
}

func (c *ESClient) BulkIndex(index string, documents *[]map[string]interface{}) error {

	var reqBody []byte
	for _, doc := range *documents {
		meta := map[string]interface{}{
			"index": map[string]interface{}{
				"_index": index,
				//"_id" //may be titlename + year
			},
		}
		reqBody = append(reqBody, []byte(fmt.Sprintf(`%s\n`, toJSON(meta)))...)
		reqBody = append(reqBody, []byte(fmt.Sprintf(`%s\n`, toJSON(doc)))...)
	}
	reqBody = append(reqBody, '\n')

	res, err := c.es.Bulk(bytes.NewReader(reqBody), c.es.Bulk.WithContext(context.Background()))
	if err != nil {
		logrus.Errorf("Error performing bulk request: %s", err)
		return err
	}
	defer res.Body.Close()

	if res.IsError() {
		logrus.Errorf("Bulk request failed: %s", res.String())
		return errors.New(res.String())
	}

	logrus.Infof("Bulk index response %s", res.String())
	return nil

}

// Helper function to convert map to JSON string
func toJSON(m map[string]interface{}) string {
	b, _ := json.Marshal(m)
	return string(b)
}

func (c *ESClient) QueryDocuments(index string, query string) ([]map[string]interface{}, error) {
	// Create search request
	reader := bytes.NewReader([]byte(query))
	req := esapi.SearchRequest{
		Index: []string{index},
		Body:  reader,
	}

	// Perform search request
	res, err := req.Do(context.Background(), c.es)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if res.IsError() {
		return nil, fmt.Errorf("failed to query documents: %s", res.Status())
	}

	// Deserialize response
	var response map[string]interface{}
	if err := json.NewDecoder(res.Body).Decode(&response); err != nil {
		return nil, err
	}

	// Extract hits from response
	hits, ok := response["hits"].(map[string]interface{})["hits"].([]interface{})
	if !ok {
		return nil, fmt.Errorf("failed to extract hits from response")
	}

	// Extract source documents from hits
	var documents []map[string]interface{}
	for _, hit := range hits {
		source, ok := hit.(map[string]interface{})["_source"].(map[string]interface{})
		if !ok {
			continue
		}
		documents = append(documents, source)
	}

	return documents, nil
}

// func main() {
// 	// Create Elasticsearch client
// 	client, err := NewESClient([]string{"http://localhost:9200"})
// 	if err != nil {
// 		log.Fatalf("Failed to create Elasticsearch client: %v", err)
// 	}

// 	// Index documents
// 	documents := []map[string]interface{}{
// 		{
// 			"title":  "Boarding School Girls' Pajama Parade",
// 			"year":   1900,
// 			"cast":   []string{},
// 			"genres": []string{},
// 			"href":   nil,
// 		},
// 		// Add other documents as needed
// 	}
// 	for i, doc := range documents {
// 		documentID := fmt.Sprintf("doc-%d", i)
// 		if err := client.IndexDocument("movies", documentID, doc); err != nil {
// 			log.Printf("Failed to index document %s: %v", documentID, err)
// 		}
// 	}

// 	// Query documents
// 	query := `
//     {
//         "query": {
//             "match_all": {}
//         }
//     }
//     `
// 	indexedDocuments, err := client.QueryDocuments("movies", query)
// 	if err != nil {
// 		log.Fatalf("Failed to query documents: %v", err)
// 	}
// 	for _, doc := range indexedDocuments {
// 		fmt.Println(doc)
// 	}
// }
