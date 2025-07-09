terraform {
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 4.0" 
    }
  }

  #instrui o Terraform a usar o S3 e o DynamoDB.
  # backend "s3" {
  #   bucket         = "scrapjobs-terraform-state-bucket" # Coloque o mesmo nome do bucket definido acima
  #   key            = "global/s3/terraform.tfstate"         # O caminho/nome do arquivo de estado dentro do bucket
  #   region         = "us-east-1"
  #   dynamodb_table = "scrapjobs-terraform-lock"     # O nome da tabela do DynamoDB definida acima
  #   encrypt        = true
  # }
}

provider "aws" {
    region = "us-east-1"
}