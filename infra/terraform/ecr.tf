resource "aws_ecr_repository" "app_repo"{
    name = "${var.project_name}-repo"
    image_tag_mutability = "MUTABLE"

    image_scanning_configuration {
        scan_on_push = true
    }
}

output "ecr_repository_url" {
  value       = aws_ecr_repository.app_repo.repository_url
  description = "ECR URL repository"
}
