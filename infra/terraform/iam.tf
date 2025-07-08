resource "aws_iam_role" "ec2_role"{
    name = "${var.project_name}-ec2-role"

    assume_role_policy = jsonencode({
        Version = "2012-10-17"
        Statement = [
            {
                Action = "sts:AssumeRole"
                Effect = "Allow"
                Principal = {
                    Service = "ec2.amazonaws.com"
                }
            },
        ]
    })
}

resource "aws_iam_policy" "ses_policy" {
    name        = "${var.project_name}-SESPolicy"
    description = "Permite acesso total ao SES"

    policy = jsonencode({
        Version = "2012-10-17"
        Statement = [
            {
                Action = "ses:SendEmail"
                Effect = "Allow"
                Resource = "*"
            },
            {
                Action = "ses:SendRawEmail"
                Effect = "Allow"
                Resource = "*"
            }
        ]
    })
}

resource "aws_iam_role_policy_attachment" "ses_attach" {
    role       = aws_iam_role.ec2_role.name 
    policy_arn = aws_iam_policy.ses_policy.arn
}

resource "aws_iam_policy" "ecr_policy" {
  name   = "${var.project_name}-ECRPolicy"
  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [{
        Effect = "Allow",
        Action = [
          "ecr:GetDownloadUrlForLayer", "ecr:BatchGetImage",
          "ecr:BatchCheckLayerAvailability", "ecr:GetAuthorizationToken"
        ],
        Resource = "*"
    }]
  })
}

resource "aws_iam_role_policy_attachment" "ecr_attach" {
  role       = aws_iam_role.ec2_role.name
  policy_arn = aws_iam_policy.ecr_policy.arn
}


resource "aws_iam_policy" "dynamodb_lock_policy" {
  name        = "${var.project_name}-DynamoDBLockPolicy"
  description = "Permite que o Terraform gerencie a tabela de lock do DynamoDB"
  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Effect = "Allow",
        Action = [
          "dynamodb:CreateTable",
          "dynamodb:DescribeTable",
          "dynamodb:GetItem",
          "dynamodb:PutItem",
          "dynamodb:DeleteItem"
        ],
        
        Resource = "arn:aws:dynamodb:${var.aws_region}:${data.aws_caller_identity.current.account_id}:table/${var.project_name}-terraform-lock"
      },
    ]
  })
}

resource "aws_iam_role_policy_attachment" "dynamodb_attach" {
  role       = aws_iam_role.ec2_role.name
  policy_arn = aws_iam_policy.dynamodb_lock_policy.arn
}



resource "aws_iam_instance_profile" "ec2_profile" {
    name = "${var.project_name}-ec2-profile"
    role = aws_iam_role.ec2_role.name
}