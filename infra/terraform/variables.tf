variable "aws_region" {
    description = "Região da AWS para os recursos"
    type        = string
    default     = "us-east-1"
}

variable "my_ip" {
    description = "IP para acesso SSH"
    type        = string
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