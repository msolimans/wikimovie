package es

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func getTestClientConfig() *ESConfig {
	return &ESConfig{
		Urls: []string{"http://my-domain.us-east-1.es.localhost.localstack.cloud:4566"},
		// Urls:                []string{"http://localhost.localstack.cloud:4571"},
		IdleConnTimeout:     10,
		MaxIdleConnsPerHost: 10,
		MaxIdleConns:        10,
	}
}

func TestNewESClient(t *testing.T) {
	// Create Elasticsearch client
	client, err := NewESClient(getTestClientConfig())
	assert.Nil(t, err)
	assert.NotNil(t, client)
}

func TestESOps(t *testing.T) {
	client, err := NewESClient(getTestClientConfig())
	assert.Nil(t, err)
	assert.NotNil(t, client)

	// err = client.CreateIndex("movies")
	// assert.Nil(t, err)

	// Index documents
	documents := []map[string]interface{}{
		{
			"title":  "Boarding School Girls' Pajama Parade",
			"year":   1900,
			"cast":   []string{},
			"genres": []string{},
			"href":   nil,
		},
		// Add other documents as needed
	}
	for i, doc := range documents {
		documentID := fmt.Sprintf("doc-%d", i)
		err := client.IndexDocument("my-index", documentID, doc)
		assert.Nil(t, err)
	}

	// Query documents
	query := `
		{
			"query": {
				"match_all": {}
			}
		}
		`
	docs, err := client.QueryDocuments("my-index", query)
	assert.Nil(t, err)
	fmt.Println(len(docs))
	assert.Greater(t, len(docs), 0)

}
