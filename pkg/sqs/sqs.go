package sqs

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sqs"
)

type SQSClient struct {
	sqsSvc *sqs.SQS
}

// todo: Make a Singleton version (like the one in logrus)
func NewSQSClient(cfg *aws.Config) *SQSClient {

	sess := session.Must(session.NewSession(cfg))

	return &SQSClient{
		sqs.New(sess),
	}

}

func String(str string) *string {
	return &str
}

func (s *SQSClient) getsqsSvc() *sqs.SQS {
	if s == nil {
		s = NewSQSClient(&aws.Config{
			Region: aws.String("us-east-1"),
		}) //default region
	}
	return s.sqsSvc
}

func (s *SQSClient) ReceiveMessage(queueUrl string) (*sqs.Message, error) {

	messages, err := s.ReceiveMessages(queueUrl, 1, 30, 10)

	if len(messages) == 0 {
		return nil, err
	}

	return messages[0], nil
}

func (s *SQSClient) ReceiveMessageWithContext(ctx context.Context, queueUrl string) (*sqs.Message, error) {

	messages, err := s.ReceiveMessagesWithContext(ctx, queueUrl, 1, 30, 10)

	if len(messages) == 0 {
		return nil, err
	}

	return messages[0], nil
}

// waitTimeoutSeconds used for longpolling
func (s *SQSClient) ReceiveMessagesWithContext(ctx aws.Context, queueUrl string, maxNoOfMessages int64, visibilityTimeout int64, waitTimeoutSeconds int64) ([]*sqs.Message, error) {
	result, err := s.getsqsSvc().ReceiveMessageWithContext(ctx, &sqs.ReceiveMessageInput{
		AttributeNames: []*string{
			aws.String(sqs.MessageSystemAttributeNameSentTimestamp),
			aws.String(sqs.MessageSystemAttributeNameApproximateReceiveCount),
		},
		MessageAttributeNames: []*string{
			aws.String(sqs.QueueAttributeNameAll),
		},
		QueueUrl:            aws.String(queueUrl),
		MaxNumberOfMessages: aws.Int64(maxNoOfMessages),
		VisibilityTimeout:   aws.Int64(visibilityTimeout), // 20 seconds
		WaitTimeSeconds:     aws.Int64(waitTimeoutSeconds),
	})

	if err != nil {
		if result != nil {
			return result.Messages, err
		}
		return nil, err
	}

	return result.Messages, nil
}

// waitTimeoutSeconds used for longpolling
func (s *SQSClient) ReceiveMessages(queueUrl string, maxNoOfMessages int64, visibilityTimeout int64, waitTimeoutSeconds int64) ([]*sqs.Message, error) {
	return s.ReceiveMessagesWithContext(context.Background(), queueUrl, maxNoOfMessages, visibilityTimeout, waitTimeoutSeconds)
}

func (s *SQSClient) ChangeMessageVisibilityWithContext(ctx aws.Context, queueUrl *string, receiptHandle *string, visibilityTimeout *int64) error {
	_, err := s.getsqsSvc().ChangeMessageVisibilityWithContext(ctx, &sqs.ChangeMessageVisibilityInput{
		QueueUrl:          queueUrl,
		ReceiptHandle:     receiptHandle,
		VisibilityTimeout: visibilityTimeout,
	})

	return err
}

// returns messageId
func (s *SQSClient) SendMessage(queueUrl string, messageBody *string, attrs ...map[string]string) (string, error) {
	//	"WeeksOn": &sqs.MessageAttributeValue{
	//		DataType:    aws.String("Number"),
	//		StringValue: aws.String("6"),
	//	},
	var messageAttributes map[string]*sqs.MessageAttributeValue = nil

	if len(attrs) > 0 {
		messageAttributes = map[string]*sqs.MessageAttributeValue{}

		for k, v := range attrs[0] {
			messageAttributes[k] = &sqs.MessageAttributeValue{
				DataType:    aws.String("String"),
				StringValue: &v,
			}
		}

	}

	result, err := s.getsqsSvc().SendMessage(&sqs.SendMessageInput{
		DelaySeconds:      aws.Int64(10),
		MessageAttributes: messageAttributes,
		MessageBody:       messageBody,
		QueueUrl:          aws.String(queueUrl),
	})

	if err != nil {
		return "", err
	}

	return *result.MessageId, nil

}

func (s *SQSClient) DeleteMessageWithContext(ctx context.Context, queueUrl string, receiptHandle *string) error {
	_, err := s.getsqsSvc().DeleteMessageWithContext(ctx, &sqs.DeleteMessageInput{
		QueueUrl:      aws.String(queueUrl),
		ReceiptHandle: receiptHandle,
	})

	return err
}

// receiptHandle inside messages[0].ReceiptHandle
func (s *SQSClient) DeleteMessage(queueUrl string, receiptHandle *string) error {
	return s.DeleteMessageWithContext(context.Background(), queueUrl, receiptHandle)
}

func getUrl(output *sqs.CreateQueueOutput) string {
	if output == nil {
		return ""
	}

	return *output.QueueUrl
}
