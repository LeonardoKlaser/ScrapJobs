resource "aws_sns_topic" "alarms_topic" {
  name = "${var.project_name}-alarms-topic"
}

resource "aws_sns_topic_subscription" "email_subscription" {
  topic_arn = aws_sns_topic.alarms_topic.arn
  protocol  = "email"
  endpoint  = var.notification_email
}


resource "aws_cloudwatch_metric_alarm" "ec2_cpu_credit_balance" {
  alarm_name          = "${var.project_name}-EC2-Low-CPUCreditBalance"
  comparison_operator = "LessThanOrEqualToThreshold"
  evaluation_periods  = "3"
  metric_name         = "CPUCreditBalance"
  namespace           = "AWS/EC2"
  period              = "300"
  statistic           = "Average"
  threshold           = "20"
  alarm_description   = "Alerta quando o saldo de créditos de CPU da instância EC2 está perigosamente baixo."

  dimensions = {
    InstanceId = aws_instance.servidor.id // Vincula dinamicamente ao ID da instância EC2
  }

  alarm_actions = [aws_sns_topic.alarms_topic.arn]
  ok_actions    = [aws_sns_topic.alarms_topic.arn]
}

resource "aws_cloudwatch_metric_alarm" "asynq_archived_queue_depth" {
  alarm_name          = "${var.project_name}-Asynq-Archived-Tasks"
  comparison_operator = "GreaterThanThreshold"
  evaluation_periods  = "1"
  metric_name         = "AsynqArchivedQueueDepth"
  namespace           = "ScrapJobs/Application"
  period              = "300"
  statistic           = "Maximum"
  threshold           = "0"
  alarm_description   = "Alerta quando uma ou mais tarefas do Asynq falham permanentemente e são arquivadas."

  dimensions = {
    QueueName = "default"
  }

  alarm_actions = [aws_sns_topic.alarms_topic.arn]
  ok_actions    = [aws_sns_topic.alarms_topic.arn]
}
