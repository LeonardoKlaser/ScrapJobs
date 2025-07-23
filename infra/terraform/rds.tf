resource "aws_db_subnet_group" "default" {
  name       = "scrapjobs-rds-subnet-group"
  subnet_ids = [aws_subnet.public.id] 
  tags = {
    Name = "${var.project_name}-rds-subnet-group"
  }
}

resource "aws_db_instance" "default"{
    identifier      = "scrapjobs-db"
    allocated_storage = 20
    engine          = "postgres"
    engine_version  = "14.10"
    instance_class  = "db.t3.micro"
    db_name               = var.dbname   
    username              = var.dbuser
    password              = var.dbpassword
    db_subnet_group_name  = aws_db_subnet_group.default.name
    vpc_security_group_ids = [aws_security_group.rds_sg.id]
    skip_final_snapshot   = true
    publicly_accessible   = true
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