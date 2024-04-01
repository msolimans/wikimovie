

resource "aws_sqs_queue" "main" {
  name                       = var.queue_name
  fifo_queue                 = false
  visibility_timeout_seconds = 90 #up to 12 hours 
  max_message_size           = 262144 #256kB
  message_retention_seconds  = 604800 #up to 14 days 

  receive_wait_time_seconds = 10 #long polling 

#   kms_master_key_id = ""

  redrive_policy       = jsonencode({
    deadLetterTargetArn = aws_sqs_queue.dlq.arn
    maxReceiveCount     = 5 #max receive_time
  })

  redrive_allow_policy = jsonencode({
    redrivePermission = "byQueue",
    sourceQueueArns   = [aws_sqs_queue.dlq.arn] #allow redrive from dlq
  }) 

  tags   = {} # add tags here
}

#DLQ queue  
resource "aws_sqs_queue" "dlq" {
  name                      = "${var.queue_name}-dlq"
  message_retention_seconds = 1209600 #retention 
} 

