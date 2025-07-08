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

    ingress {
        description = "permite acesso a SSH somente pelo meu IP e gitHubActions"
        from_port = 22
        to_port = 22
        protocol = "tcp"
        cidr_blocks = [var.my_ip]
    }

    egress{
        from_port = 0
        to_port = 0
        protocol = "-1"
        cidr_blocks = ["0.0.0.0/0"]
    }
}

