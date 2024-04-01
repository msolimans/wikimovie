resource "aws_s3_bucket" "main" {
  bucket = var.s3_bucket
}


resource "aws_s3_bucket_notification" "main" {
  for_each = { for v in var.notifications : v => v }
  #   count    = var.sns_topic_arn != "" ? 1 : 0
  bucket = aws_s3_bucket.main.id
  topic {
    topic_arn     = each.value.topic_arn
    events        = try(each.value.events, ["s3:ObjectCreated:*"])
    filter_prefix = each.value.filter_prefix
    filter_suffix = each.value.filter_suffix
  }
}
