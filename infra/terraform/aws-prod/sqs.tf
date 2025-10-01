resource "aws_sqs_queue" "scraping_tasks_queue" {
  name                      = "scraping-tasks-queue"
  delay_seconds             = 0
  max_message_size          = 262144
  message_retention_seconds = 86400  
  visibility_timeout_seconds = 90 
}