resource "aws_secretsmanager_secret" "db_credentials" {
  name = "${var.project_name}-db-credentials"
  description = "Database credentials for ScrapJobs"
}

resource "aws_secretsmanager_secret_version" "db_credentials_version" {
  secret_id     = aws_secretsmanager_secret.db_credentials.id
  secret_string = jsonencode({
    host_db      = var.db_host
    port_db      = var.db_port
    db_user      = var.dbuser
    db_password  = var.dbpassword
    db_name      = var.dbname
    redis_host   = var.redis_host
    redis_addr   = var.redis_addr
    redis_port   = var.redis_port
    redis_conf   = var.redis_conf
    gemini_key   = var.gemini_key
    ai_model     = var.ai_model
  })
}