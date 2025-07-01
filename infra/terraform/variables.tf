variable "aws_region" {
    description = "Região da AWS para os recursos"
    type        = string
    default     = "us-east-1"
}

variable "my_ip" {
    description = "IP para acesso SSH"
    type        = string
    default     = "200.201.202.203/32"
}

variable "project_name" {
    description = "Nome base para os recursos"
    type        = string
    default     = "ScrapJobs"
}

variable "ses_domain_name" {
    description = "Dominio verificacao SES"
    type        = string
    default     = "ScrapJobs.com"
}