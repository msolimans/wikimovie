package main

import (
	"context"
	"os"
	"os/signal"
	"sync"

	"github.com/msolimans/wikimovie/pkg/appconf"
	"github.com/msolimans/wikimovie/pkg/es"
	"github.com/msolimans/wikimovie/pkg/s3"
	"github.com/msolimans/wikimovie/pkg/sqs"
	"github.com/sirupsen/logrus"
)

func main() {

	ctx := context.Background()

	logger := logrus.New()

	cfg := &appconf.Configuration{}
	if err := appconf.LoadConfig(".", cfg); err != nil {
		logrus.Fatalf("Error loading config %v", err)
	}

	termChan := make(chan os.Signal, 1)
	signal.Notify(termChan, os.Interrupt)

	wg := &sync.WaitGroup{}
	// call cancel in case we want to stop t this worker
	ctx, cancel := context.WithCancel(ctx)

	sqsClient := sqs.NewSQSClient(cfg.Aws)

	logger.Info("Connecting to ES")

	esConf := &es.ESConfig{
		Urls: cfg.ElasticSearch.Urls,
		// Urls:                []string{"http://localhost.localstack.cloud:4571"},
		IdleConnTimeout:     cfg.ElasticSearch.IdleConnTimeout,
		MaxIdleConnsPerHost: cfg.ElasticSearch.MaxIdleConnsPerHost,
		MaxIdleConns:        cfg.ElasticSearch.MaxIdleConns,
	}

	esClient, err := es.NewESClient(esConf)
	if err != nil {
		panic("Can not connect to ES")
	}

	logger.Info("Creating s3 manager")
	s3Manager := s3.NewS3Manager(cfg.Aws)

	workerHandler := &workerHandler{
		s3Manager: s3Manager,
		esClient:  esClient,
	}

	logger.Info("Queue UR ", cfg.Worker.Queue)
	//////////////////////////////////////////////////////////////////////

	go sqs.StartWorker(ctx, wg, sqsClient, workerHandler,
		&sqs.WorkerConfig{
			MaxNumberOfMessages: cfg.Worker.MaxNumberOfMessages,
			WaitTimeSeconds:     cfg.Worker.WaitTimeSeconds, //long polling enabled - maximum = 20
			VisibilityTimeout:   cfg.Worker.VisibilityTimeout,
			RetryIntervals:      cfg.Worker.RetryIntervals, //5 seconds
		}, cfg.Worker.Queue, logger, cancel)

	wg.Add(1)

	// wait for shutdown
	select {
	case <-termChan:
		logger.Info("-----Shutdown Received-------")
	case <-ctx.Done():
		logger.Info("-----Force Shutdown Received--------")
	}

	cancel()
	wg.Wait()
	logger.Info("All done, shutting down")
}
