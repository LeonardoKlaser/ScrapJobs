
resource "aws_sns_topic" "alarms_topic_local" {
  name = "scrapjobs-alarms-topic-local"
}


resource "aws_sns_topic_subscription" "email_subscription_local" {
  topic_arn = aws_sns_topic.alarms_topic_local.arn
  protocol  = "email"
  endpoint  = "leobkklaser@gmail.com"
}


resource "aws_cloudwatch_metric_alarm" "asynq_archived_queue_depth_local" {
  alarm_name          = "ScrapJobs-Asynq-Archived-Tasks-Local"
  comparison_operator = "GreaterThanThreshold"
  evaluation_periods  = "2"
  metric_name         = "AsynqArchivedQueueDepth"
  namespace           = "ScrapJobs/Application"
  period              = "60" # Período mais curto para testes
  statistic           = "Maximum"
  threshold           = "0"
  alarm_description   = "Alerta local quando tarefas do Asynq são arquivadas."

  dimensions = {
    QueueName = "default" 
  }
  
  alarm_actions = [aws_sns_topic.alarms_topic_local.arn]
  ok_actions    = [aws_sns_topic.alarms_topic_local.arn]
}