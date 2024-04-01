#!/bin/bash
set -e;

echo "Deleting s3 buckets and queues"
 
# Wait for LocalStack to be ready
while ! nc -z localhost.localstack.cloud 4566; do
  sleep 1
done
  
function queueExists() {
  name=$1
  aws --region us-east-1 --endpoint-url=http://localhost.localstack.cloud:4566 sqs get-queue-url --queue-name $name >/dev/null 2>&1
}

function cleanSQS() {
  name=$1
  if queueExists $name; then
    aws --region us-east-1 --endpoint-url=http://localhost.localstack.cloud:4566 sqs delete-queue --queue-url http://sqs.us-east-1.localhost.localstack.cloud:4566/000000000000/$1
    aws --region us-east-1 --endpoint-url=http://localhost.localstack.cloud:4566 sqs delete-queue --queue-url http://sqs.us-east-1.localhost.localstack.cloud:4566/000000000000/$1-dlq
  else
    echo "SQS queue $name does not exist"
  fi
}

# SQS
cleanSQS wm-dev-movies-published 


# S3


function bucketExists() {
  name=$1
  aws --region us-east-1 --endpoint-url=http://localhost.localstack.cloud:4566 s3api head-bucket --bucket $name >/dev/null 2>&1
}


function cleanS3() {
  name=$1
  if bucketExists $name; then
    aws --region us-east-1 --endpoint-url=http://localhost.localstack.cloud:4566 s3 rb s3://$1 --force
  else
    echo "S3 bucket $name does not exist"
  fi
   
}

cleanS3  wm-dev-movies-bucket  
echo 'DONE'