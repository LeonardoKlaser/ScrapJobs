resource "aws_security_group" "ec2_sg"{
    name = "${var.project_name}-segredos"
    description = "Permitir acesso HTTP e acesso a Internet"
    vpc_id  = aws_vpc.main.id

    ingress {
        description = "permite acesso a API pela porta 80 (frontend)"
        from_port = 80
        to_port = 80
        protocol = "tcp"
        cidr_blocks = ["0.0.0.0/0"]
    }

    egress{
        from_port = 0
        to_port = 0
        protocol = "-1"
        cidr_blocks = ["0.0.0.0/0"]
    }
}

resource "aws_security_group" "rds_sg"{
    name    ="${var.project_name}-rds-sg"
    description = "Allow connections from EC2 only"
    vpc_id  = aws_vpc.main.id

    ingress{
        description = "Allow Postgresql traffic from EC2"
        from_port   = 5432
        to_port     = 5432
        protocol    = "tcp"
        security_groups = [aws_security_group.ec2_sg.id]
    }

    egress{
        from_port   = 0
        to_port     = 0
        protocol    = "-1"
        cidr_blocks = ["0.0.0.0/0"]
    }

    tags = {
        Name    = "${var.project_name}-rds-sg"
    }
}

resource "aws_security_group" "elasticache_sg" {
    name    = "${var.project_name}-elasticache-sg"
    description = "Allow connections from EC2 only"
    vpc_id  = aws_vpc.main.id

    ingress {
    description     = "Allow Redis traffic from EC2"
    from_port       = 6379
    to_port         = 6379
    protocol        = "tcp"
    security_groups = [aws_security_group.ec2_sg.id]
  }

  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }

   tags = {
    Name = "${var.project_name}-elasticache-sg"
  }

}
