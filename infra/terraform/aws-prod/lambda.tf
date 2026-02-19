# IAM Role para a função Lambda
# NOTE: Lambda resources are commented out until a valid build artifact is available.

# resource "aws_iam_role" "lambda_worker_role" {
#   name = "scrapjobs-lambda-worker-role"
#
#   assume_role_policy = jsonencode({
#     Version   = "2012-10-17",
#     Statement = [
#       {
#         Action    = "sts:AssumeRole",
#         Effect    = "Allow",
#         Principal = {
#           Service = "lambda.amazonaws.com"
#         },
#       },
#     ]
#   })
# }
#
# # Política de permissões para a Lambda
# resource "aws_iam_policy" "lambda_worker_policy" {
#   name        = "scrapjobs-lambda-worker-policy"
#   description = "Permissões para o worker Lambda do ScrapJobs"
#
#   policy = jsonencode({
#     Version   = "2012-10-17",
#     Statement = [
#       // Permissões para SQS
#       {
#         Effect   = "Allow",
#         Action   = [
#           "sqs:ReceiveMessage",
#           "sqs:DeleteMessage",
#           "sqs:GetQueueAttributes"
#         ],
#         Resource = aws_sqs_queue.scraping_tasks_queue.arn
#       },
#       // Permissão para logs
#       {
#         Effect   = "Allow",
#         Action   = [
#             "logs:CreateLogGroup",
#             "logs:CreateLogStream",
#             "logs:PutLogEvents"
#         ],
#         Resource = "arn:aws:logs:*:*:*"
#       },
#       // adicionar outras permissoes necessarias (rds, SES...)
#     ]
#   })
# }
#
# resource "aws_iam_role_policy_attachment" "lambda_policy_attach" {
#   role       = aws_iam_role.lambda_worker_role.name
#   policy_arn = aws_iam_policy.lambda_worker_policy.arn
# }
#
# resource "aws_lambda_function" "scraping_worker" {
#   function_name = "scraping-worker-lambda"
#   role          = aws_iam_role.lambda_worker_role.arn
#   handler       = "main"
#   runtime       = "provided.al2"
#   architectures = ["x86_64"]
#   timeout       = 60
#
#   # TODO: Replace with actual Lambda build artifact path
#   source_code_hash = filebase64sha256("path/to/your/lambda_build.zip")
#   filename         = "path/to/your/lambda_build.zip"
#
#   environment {
#     variables = {
#       GIN_MODE = "release"
#     }
#   }
# }
#
# # Gatilho do SQS para o Lambda
# resource "aws_lambda_event_source_mapping" "sqs_trigger" {
#   event_source_arn = aws_sqs_queue.scraping_tasks_queue.arn
#   function_name    = aws_lambda_function.scraping_worker.arn
#   batch_size       = 1 // Processa uma mensagem por vez
# }
