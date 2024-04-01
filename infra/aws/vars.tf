variable "queue_name" {
    default = "wm-dev-movies-published"  
}

variable "s3_bucket" {
    default = "wm-dev-movies-bucket"  
}

variable "notifications" {
  description = "notifications"
  type = list(object({
    topic_arn     = string
    events        = list(string)
    filter_prefix = string
    filter_suffix = string
  }))
  default = []
}