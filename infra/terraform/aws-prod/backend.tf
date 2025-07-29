# resource "aws_s3_bucket" "terraform_state" {
#   bucket = "scrapjobs-terraform-state-bucket"

#   lifecycle {
#     prevent_destroy = true
#   }
# }

# resource "aws_s3_bucket_versioning" "terraform_state_versioning" {
#     bucket = aws_s3_bucket.terraform_state.id
#     versioning_configuration {
#         status = "Enabled"
#     }
# }

# resource "aws_dynamodb_table" "terraform_state_lock" {
#   name           = "scrapjobs-terraform-lock"
#   billing_mode   = "PAY_PER_REQUEST" 
#   hash_key       = "LockID"

#   attribute {
#     name = "LockID"
#     type = "S"
#   }a
# }
