#!/bin/bash
set -e;

echo "INSTALLING s3 buckets and queues"

# Wait for LocalStack to be ready
while ! nc -z localhost 4566; do
  sleep 1
done

#$1 is expected to be queue name  
#$2 is the visibility timeout - defaults to 20
function createSQSQueue() {
  name=$1
  timeout="${2:-20}"
  dlq_arn=arn:aws:sqs:us-east-1:000000000000:$name-dlq
  awslocal sqs create-queue --queue-name $name-dlq
  awslocal sqs create-queue --queue-name $name \
           --attributes '{"RedrivePolicy":"{\"deadLetterTargetArn\":\"'$dlq_arn'\", \"maxReceiveCount\":\"5\"}",
           "VisibilityTimeout":"'$timeout'"}'
}

export LOCALSTACK_HOST=localhost.localstack.cloud

# SQS
createSQSQueue wm-dev-movies-published 30


# S3
function createS3Bucket() {
  name=$1
  queue=$2 
  awslocal s3api create-bucket --bucket $name
  awslocal s3api put-bucket-notification-configuration --bucket $name \
    --notification-configuration '{
      "QueueConfigurations": [
          {
              "QueueArn": "arn:aws:sqs:us-east-1:000000000000:$queue",
              "Events": ["s3:ObjectCreated:*"],
              "Filter": {
                "Key": {
                  "FilterRules": [
                    {
                      "Name": "prefix",
                      "Value": "movies/new/"
                    },
                    {
                      "Name": "suffix",
                      "Value": ".gz"
                    }
                  ]
                }
              }
          }
      ]
    }'
 
}

createS3Bucket  wm-dev-movies-bucket wm-dev-movies-published