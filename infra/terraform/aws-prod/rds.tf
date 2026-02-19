resource "aws_db_subnet_group" "default" {
  name       = "scrapjobs-rds-subnet-group"
  subnet_ids = [aws_subnet.private.id, aws_subnet.private_b.id]
  tags = {
    Name = "${var.project_name}-rds-subnet-group"
  }
}

resource "aws_db_instance" "default" {
  identifier             = "scrapjobs-db"
  allocated_storage      = 20
  engine                 = "postgres"
  engine_version         = "17"
  instance_class         = "db.t3.micro"
  db_name                = var.dbname
  username               = var.dbuser
  password               = var.dbpassword
  db_subnet_group_name   = aws_db_subnet_group.default.name
  vpc_security_group_ids = [aws_security_group.rds_sg.id]
  publicly_accessible    = false

  skip_final_snapshot       = false
  final_snapshot_identifier = "scrapjobs-final-snapshot"
  backup_retention_period   = 7
  backup_window             = "03:00-04:00"
  storage_encrypted         = true
  deletion_protection       = true
}

output "rds_endpoint" {
  description = "endpoint database"
  value       = aws_db_instance.default.endpoint
  sensitive   = true
}

output "rds_port" {
  description = "database port"
  value       = aws_db_instance.default.port
}
