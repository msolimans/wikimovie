package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/msolimans/wikimovie/pkg/es"
	"github.com/msolimans/wikimovie/pkg/s3"
	"github.com/sirupsen/logrus"
)

// download s3 file
// parse and index docs
// move original s3 file to processed folder with new suffix for filename
func Process(ctx context.Context, s3 s3.ManagerAPI,
	es *es.ESClient, bucket, key string, destinationPath ...string) error {

	result, err := s3.DownloadObject(ctx, bucket, key)
	if err != nil {
		return err
	}

	docs := &[]map[string]interface{}{}
	if err := json.Unmarshal(result, docs); err != nil {
		return err
	}
	//no docs in file
	if len(*docs) == 0 {
		return nil
	}

	processed := 0
	errored := 0

	for i, doc := range *docs {
		fmt.Println(i)
		yearStr := fmt.Sprintf("%.0f", doc["year"])

		// Convert title to string and replace spaces with dashes
		titleStr := strings.ReplaceAll(doc["title"].(string), " ", "-")

		// Combine year and title to create a valid document ID
		docID := fmt.Sprintf("%s-%s", yearStr, titleStr)

		if err := es.IndexDocument("my-index", docID, doc); err != nil {
			errored = errored + 1
			logrus.Errorf("Error processing document %v \n error: %s", doc, err)
		} else {
			processed += 1
		}
	}

	// if err := es.BulkIndex("my-index", docs); err != nil {
	// 	return err
	// }

	if len(destinationPath) > 0 {
		return moveToProcessed(ctx, s3, bucket, key, destinationPath[0])
	}

	return nil

}

// moves processed file with random suffix
func moveToProcessed(ctx context.Context, s3 s3.ManagerAPI, bucket, key, destinationPath string) error {
	_, fileKey := filepath.Split(key)
	f := strings.Split(fileKey, ".")

	//todo: add response from es inside the processed file
	fileName := fmt.Sprintf("%s-%v", f[0], time.Now().UTC().UnixNano())
	extension := strings.Join(f[1:], ".")
	destinationKey := fmt.Sprintf("%s/%s.%s", destinationPath, fileName, extension)
	return s3.MoveObject(ctx, bucket, key, destinationKey)
}
