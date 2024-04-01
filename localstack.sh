#!/bin/bash
set -e;

echo "INSTALLING s3 buckets and queues"
 
# Wait for LocalStack to be ready
while ! nc -z localhost.localstack.cloud 4566; do
  sleep 1
done

# for some reason it didn't work 
# alias awslocal="aws --region us-east-1 --endpoint-url=http://localhost.localstack.cloud:4566"
function queueExists() {
  name=$1
  aws --region us-east-1 --endpoint-url=http://localhost.localstack.cloud:4566 sqs get-queue-url --queue-name $name >/dev/null 2>&1
}


#$1 is expected to be queue name  
#$2 is the visibility timeout (20 by default)
function createSQSQueue() {
  name=$1
  timeout="${2:-20}" #20 if not provided

  if queueExists $name; then
    echo 'queue exists'
  else 
    dlq_arn=arn:aws:sqs:us-east-1:000000000000:$name-dlq
    aws --region us-east-1 --endpoint-url=http://localhost.localstack.cloud:4566 sqs create-queue --queue-name $name-dlq
    aws --region us-east-1 --endpoint-url=http://localhost.localstack.cloud:4566 sqs create-queue --queue-name $name \
           --attributes '{"RedrivePolicy":"{\"deadLetterTargetArn\":\"'$dlq_arn'\", \"maxReceiveCount\":\"5\"}",
           "VisibilityTimeout":"'$timeout'"}'
  fi 
}

export LOCALSTACK_HOST=localhost.localstack.cloud

# SQS
createSQSQueue wm-dev-movies-published 30


# S3


function bucketExists() {
  name=$1
  aws --region us-east-1 --endpoint-url=http://localhost.localstack.cloud:4566 s3api head-bucket --bucket $name >/dev/null 2>&1
}

function createS3Bucket() {
  name=$1
  queue=$2 
  if bucketExists $name; then
    echo 's3 bucket exists'
  else 
    aws --region us-east-1 --endpoint-url=http://localhost.localstack.cloud:4566 s3api create-bucket --bucket $name
    # todo: queuarn is not accepting $queue (this needs to get fixed - hardcoding for now)
    aws --region us-east-1 --endpoint-url=http://localhost.localstack.cloud:4566 s3api put-bucket-notification-configuration --bucket $name \
    --notification-configuration '{
      "QueueConfigurations": [
          {
              "QueueArn": "arn:aws:sqs:us-east-1:000000000000:wm-dev-movies-published",
              "Events": ["s3:ObjectCreated:*"],
              "Filter": {
                "Key": {
                  "FilterRules": [
                    {
                      "Name": "prefix",
                      "Value": "new/"
                    },
                    {
                      "Name": "suffix",
                      "Value": ".json"
                    }
                  ]
                }
              }
          }
      ]
    }'
  fi 
}

createS3Bucket  wm-dev-movies-bucket wm-dev-movies-published