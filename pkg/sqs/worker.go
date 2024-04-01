package sqs

import (
	"context"
	"fmt"
	"strconv"
	"sync"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/sqs"

	"github.com/sirupsen/logrus"
)

// RetryIntervals:      []int{20, 40, 60},
func StartWorker(ctx context.Context, wg *sync.WaitGroup, sqsClient *SQSClient, handler MessageHandler,
	config *WorkerConfig, queueUrl string, logger *logrus.Logger, cancel context.CancelFunc) {
	defer wg.Done()
	eventWorker := NewWorker(sqsClient, logger, config, queueUrl)
	eventWorker.Start(ctx, handler, cancel)
	logger.WithField("queueURL", queueUrl).Info("Done in worker go routine")
}

// Handler interface
type MessageHandler interface {
	HandleMessage(ctx context.Context, logger *logrus.Entry, msg *sqs.Message) error
}

// InvalidEventError struct
type InvalidEventError struct {
	event string
	msg   string
}

func (e InvalidEventError) Error() string {
	return fmt.Sprintf("[Invalid Event: %s] %s", e.event, e.msg)
}

// Worker struct
type Worker struct {
	Config    *WorkerConfig
	SqsClient *SQSClient
	QueueURL  string
	logger    *logrus.Entry
}

// WorkerConfig struct
type WorkerConfig struct {
	Queue               string
	MaxNumberOfMessages int64
	WaitTimeSeconds     int64
	VisibilityTimeout   int64
	RetryIntervals      []int
}

func (config *WorkerConfig) populateDefaultValues() {
	if config.MaxNumberOfMessages == 0 {
		config.MaxNumberOfMessages = 10
	}

	if config.WaitTimeSeconds == 0 {
		config.WaitTimeSeconds = 20
	}

	if len(config.RetryIntervals) == 0 {
		if config.VisibilityTimeout > 0 {
			config.RetryIntervals = []int{int(config.VisibilityTimeout)}
		} else {
			config.RetryIntervals = []int{30, 60, 120, 360, 480}
		}
	}
}

func (config *WorkerConfig) retryInterval(i int) int {

	if len(config.RetryIntervals) == 0 {
		return int(config.VisibilityTimeout) //default timeout configured in worker config
	}

	if i == 0 {
		return config.RetryIntervals[0]
	}

	size := len(config.RetryIntervals) - 1

	if size >= i {
		return config.RetryIntervals[i] //current index
	}

	return config.RetryIntervals[size] //last index
}

// NewWorker sets up a new Worker
func NewWorker(client *SQSClient, logger *logrus.Logger, config *WorkerConfig, queueURL string) *Worker {
	config.populateDefaultValues()
	return &Worker{
		Config:    config,
		SqsClient: client,
		QueueURL:  queueURL,
		logger:    logger.WithField("queue_url", queueURL),
	}
}

// Start starts the polling and will continue polling till the application is forcibly stopped
func (worker *Worker) Start(ctx context.Context, h MessageHandler, cancel context.CancelFunc) {
	worker.logger.Info("worker: starting poller")
	var startedPolling bool
	for {
		select {
		case <-ctx.Done():
			worker.logger.Info("worker: stopping poller because a context.Done signal was sent")
			return
		default:
			if !startedPolling {
				worker.logger.Info("worker: polling.....")
				startedPolling = true
			}

			nctx := context.Background()
			resp, err := worker.SqsClient.ReceiveMessagesWithContext(nctx, worker.QueueURL, worker.Config.MaxNumberOfMessages, worker.Config.VisibilityTimeout, worker.Config.WaitTimeSeconds)
			if err != nil {
				worker.logger.Error("Error while trying to receive messages from SQS", err)
				// This error was unexpected so shutdown the worker
				// this will allow main to shut down which will cause a restart of the container
				cancel()
				return
			}

			if len(resp) > 0 {
				worker.processMessages(nctx, h, resp)
			}
		}
	}
}

// poll launches goroutine per received message and wait for all message to be processed
func (worker *Worker) processMessages(ctx context.Context, handler MessageHandler, messages []*sqs.Message) {
	numMessages := len(messages)
	logger := worker.logger.WithField("number_of_messages", numMessages)
	logger.Info("worker: received messages")

	var wg sync.WaitGroup
	wg.Add(numMessages)
	for i := range messages {
		go func(m *sqs.Message) {
			// launch goroutine
			defer wg.Done()
			// TODO should I derive a new context her for each handle?
			if err := worker.handleMessage(ctx, m, handler); err != nil {
				logger.Errorf("worker: fatal error processing message: [%v]", err)
			}
		}(messages[i])
	}

	wg.Wait()
}

func (worker *Worker) handleMessage(ctx context.Context, m *sqs.Message, h MessageHandler) error {

	herr := h.HandleMessage(ctx, worker.logger, m)
	logger := worker.logger.WithField("message_id", m.MessageId)
	logger.Infof("message body [%v]", *m.Body)
	if herr != nil {
		logger.Infof("changing visibility timeout due to error while processing message %v", herr.Error())
		return worker.changeMessageVisibility(ctx, m)
	}

	logger.Info("message handled - start deleting ...")
	err := worker.SqsClient.DeleteMessageWithContext(ctx, worker.QueueURL, m.ReceiptHandle)
	if err != nil {
		logger.Errorf("worker: could not delete message from the queue [%v]", err)
		return err
	}
	logger.Info("worker: deleted message from queue")

	return nil
}

func (worker *Worker) changeMessageVisibility(ctx context.Context, m *sqs.Message) error {

	logger := worker.logger.WithField("message_id", m.MessageId)
	// logger.Info("message", m)
	i := 0
	if m.Attributes == nil {
		logger.Errorf("could not retrieve ApproximateReceiveCount attribute - skipping")
	} else {
		var err error
		if m.Attributes["ApproximateReceiveCount"] != nil {
			i, err = strconv.Atoi(*m.Attributes["ApproximateReceiveCount"])
			if err != nil {
				logger.Errorf("worker: could not retrieve ApproximateReceiveCount attribute: [%v]", err)
				// return err
			}
			i -= 1
		}
	}

	visibilityTimeout := int64(worker.Config.retryInterval(i))

	logger = logger.WithFields(logrus.Fields{
		"approximate_receive_count": i,
		"visibility_timeout":        visibilityTimeout,
	})
	logger.Info("setting message visibility")
	err := worker.SqsClient.ChangeMessageVisibilityWithContext(ctx, aws.String(worker.QueueURL), m.ReceiptHandle, aws.Int64(visibilityTimeout))
	if err != nil {
		logger.Errorf("worker: error calling ChangeMessageVisibility: [%v]", err)
	}

	return err
}
