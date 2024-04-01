package sqs

import (
	"errors"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/stretchr/testify/assert"
)

const (
	SKIP_TEST = false
)

// var TEST_SQS_CLIENT = NewSQSClient(aws.String(TEST_SQS_REGION), aws.String(TEST_SQS_ENDPOINT))
func getTestAwsConf() *aws.Config {

	return &aws.Config{
		Region:   aws.String("us-east-1"),
		Endpoint: aws.String("http://localhost.localstack.cloud:4566"),
	}
}
func TestNewSQSClient(t *testing.T) {
	if SKIP_TEST {
		return
	}
	sqsClient := NewSQSClient(getTestAwsConf())
	assert.NotNil(t, sqsClient)
}

var queue = "http://sqs.us-east-1.localhost.localstack.cloud:4566/000000000000/wm-dev-movies-published"

// var queue = "http://sqs.us-east-1.goaws.com:4575/100010001000/local-queue3"

func TestSQSClient_SendMessage(t *testing.T) {
	if SKIP_TEST {
		return
	}
	assert.NotEmpty(t, queue)
	sqsClient := NewSQSClient(getTestAwsConf())
	messageId, err := sqsClient.SendMessage(queue, aws.String("test"))

	assert.Nil(t, err)
	assert.NotNil(t, messageId)
	assert.NotEmpty(t, messageId)
}

var recepeintHandle = ""

func TestSQSClient_ReceiveMessage(t *testing.T) {
	if SKIP_TEST {
		return
	}

	sqsClient := NewSQSClient(getTestAwsConf())
	message, err := sqsClient.ReceiveMessage(queue)

	assert.Nil(t, err)
	assert.NotNil(t, message)
	if message.Body == nil { //fault
		assert.Error(t, errors.New("message body is nil"))
		err = sqsClient.DeleteMessage(queue, message.ReceiptHandle)
		assert.Nil(t, err)
		return
	}

	assert.NotEmpty(t, message.Body)

	assert.True(t, *message.Body == "test")
	recepeintHandle = *message.ReceiptHandle
}

func TestSQSClient_ReceiveMessages(t *testing.T) {
	if SKIP_TEST {
		return
	}

	sqsClient := NewSQSClient(getTestAwsConf())
	messageId, err := sqsClient.SendMessage(queue, aws.String("test"))

	assert.Nil(t, err)
	assert.NotNil(t, messageId)
	assert.NotEmpty(t, messageId)

	message, err := sqsClient.ReceiveMessages(queue, 10, 30, 1) //
	// http://sqs.us-east-1.goaws.com:4575/100010001000/local-queue3
	// http://sqs.us-east-1.goaws.com:4575/100010001000/local-queue1
	assert.Nil(t, err)

	if len(message) > 0 {
		assert.NotNil(t, message[0])
		assert.NotEmpty(t, message[0].Body)
		assert.True(t, *message[0].Body == "test")
	}
}

func TestSQSClient_DeleteMessage(t *testing.T) {
	if SKIP_TEST || recepeintHandle == "" {
		return
	}
	sqsClient := NewSQSClient(getTestAwsConf())
	err := sqsClient.DeleteMessage(queue, &recepeintHandle)

	assert.Nil(t, err)
}
