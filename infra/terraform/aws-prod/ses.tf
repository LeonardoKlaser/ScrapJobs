resource "aws_ses_email_identity" "my_email" {
  email = var.notification_email
}
