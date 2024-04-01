package main

import (
	"context"
	"encoding/json"

	"github.com/aws/aws-lambda-go/events"
	. "github.com/aws/aws-sdk-go/service/sqs"
	"github.com/msolimans/wikimovie/pkg/es"
	"github.com/msolimans/wikimovie/pkg/handlers"
	"github.com/msolimans/wikimovie/pkg/s3"
	"github.com/sirupsen/logrus"
)

type NotificationRecord struct {
	EventSource    string
	EventSourceArn string
	AWSRegion      string
	S3             events.S3Entity
}

type NotificationRecordsMessage struct {
	Records []*NotificationRecord `json:"records"`
}
type workerHandler struct {
	s3Manager s3.ManagerAPI
	esClient  *es.ESClient
}

func (w *workerHandler) HandleMessage(ctx context.Context, logger *logrus.Entry, msg *Message) error {
	logger.Info("New message ", msg)
	message := &NotificationRecordsMessage{}

	if err := json.Unmarshal([]byte(*msg.Body), message); err != nil {
		return err
	}
	if len(message.Records) == 0 {
		return nil
	}
	s3Event := message.Records[0].S3
	bucket := s3Event.Bucket.Name
	key := s3Event.Object.Key
	//ignore if either bucket or key is missing
	if bucket == "" || key == "" {
		return nil
	}

	return handlers.Process(ctx, w.s3Manager, w.esClient, bucket, key)

}
