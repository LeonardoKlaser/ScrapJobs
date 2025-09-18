# infra/terraform/aws-prod/oidc.tf

resource "aws_iam_openid_connect_provider" "github" {
  url = "https://token.actions.githubusercontent.com"

  client_id_list = [
    "sts.amazonaws.com"
  ]

  thumbprint_list = ["6938fd4d98bab03faadb97b34396831e3780aea1"] # Thumbprint padrão para GitHub OID
}

data "aws_iam_policy_document" "github_actions_assume_role_policy" {
  statement {
    actions = ["sts:AssumeRoleWithWebIdentity"]
    effect  = "Allow"

    principals {
      type        = "Federated"
      identifiers = [aws_iam_openid_connect_provider.github.arn]
    }

    condition {
      test     = "StringLike"
      variable = "token.actions.githubusercontent.com:sub"
      values   = ["repo:leonardoklaser/scrapjobs:ref:refs/heads/main"] # Restringe à branch main do seu repo
    }
  }
}

resource "aws_iam_role" "github_actions_role" {
  name               = "${var.project_name}-github-actions-role"
  assume_role_policy = data.aws_iam_policy_document.github_actions_assume_role_policy.json
}

# Política para permitir que o GitHub Actions execute comandos na instância EC2 via SSM
resource "aws_iam_policy" "ssm_command_policy" {
  name        = "${var.project_name}-SSMCommandPolicy"
  description = "Allows running commands on EC2 instances via SSM"

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Effect   = "Allow"
        Action   = [
          "ssm:SendCommand",
          "ssm:GetCommandInvocation"
        ]
        Resource = [
          "arn:aws:ssm:${var.aws_region}:${data.aws_caller_identity.current.account_id}:document/AWS-RunShellScript",
          aws_instance.servidor.arn
        ]
      },
    ]
  })
}

# Anexa a política SSM à nova role do GitHub Actions
resource "aws_iam_role_policy_attachment" "ssm_command_attach" {
  role       = aws_iam_role.github_actions_role.name
  policy_arn = aws_iam_policy.ssm_command_policy.arn
}

# Anexa a política do ECR (que já existe) à nova role do GitHub Actions
resource "aws_iam_role_policy_attachment" "ecr_attach_github_actions" {
  role       = aws_iam_role.github_actions_role.name
  policy_arn = aws_iam_policy.ecr_policy.arn
}

data "aws_caller_identity" "current" {}