variable "aws_region" {
    description = "Região da AWS para os recursos"
    type        = string
    default     = "us-east-1"
}

variable "my_ip" {
    description = "IP para acesso SSH"
    type        = string
    sensitive   = true
}

variable "project_name" {
    description = "Nome base para os recursos"
    type        = string
    default     = "ScrapJobs"
}

variable "ses_domain_name" {
    description = "Dominio verificacao SES"
    type        = string
    default     = "scrapjobs.com.br"
}

variable "dbname" {
    description = "databse name"
    type        = string
    sensitive   = true
}

variable "dbuser" {
    description = "database user"
    type        = string
    sensitive   = true
}

variable "dbpassword" {
    description = "database password"
    type        = string
    sensitive   = true
}

variable "db_host" {
  description = "Database endpoint"
  type        = string
  sensitive   = true
}

variable "db_port" {
  description = "Database port"
  type        = string
  sensitive   = true
}

variable "ai_model" {
  description = "AI Model"
  type        = string
  sensitive   = true
}

variable "gemini_key" {
  description = "gemini_key"
  type        = string
  sensitive   = true
}

variable "redis_addr" {
  description = "redis_addr"
  type        = string
  sensitive   = true
}

variable "redis_conf" {
  description = "redis_conf"
  type        = string
  sensitive   = true
}

variable "redis_host" {
  description = "redis_host"
  type        = string
  sensitive   = true
}

variable "redis_port" {
  description = "redis_port"
  type        = string
  sensitive   = true
}

variable "notification_email" {
  description = "Email for notifications and SES"
  type        = string
}