resource "aws_s3_bucket" "main" {
  bucket = var.s3_bucket
}


resource "aws_s3_bucket_notification" "main" {
   
  #   count    = var.sns_topic_arn != "" ? 1 : 0
  bucket = aws_s3_bucket.main.id
  topic {
    topic_arn     = aws_sqs_queue.topic_arn
    events        =   ["s3:ObjectCreated:*"]
    filter_prefix =  "new/"
    filter_suffix = ".json"
  }
}
