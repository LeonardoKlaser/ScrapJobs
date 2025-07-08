terraform {
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 4.0" 
    }
  }

  # Esta é a parte mais importante.
  # Ela instrui o Terraform a usar o S3 e o DynamoDB.
#   backend "s3" {
#     bucket         = "${var.project_name}-terraform-state-bucket" # Coloque o mesmo nome do bucket definido acima
#     key            = "global/s3/terraform.tfstate"         # O caminho/nome do arquivo de estado dentro do bucket
#     region         = "us-east-1"
#     dynamodb_table = "${var.project_name}-terraform-lock"     # O nome da tabela do DynamoDB definida acima
#     encrypt        = true
#   }
}

provider "aws" {
    region = "us-east-1"
}