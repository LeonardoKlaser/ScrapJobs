provider "aws" {
    region = "us-east-1"
}

resource "aws_security_group" "securitygroup"{
    name = "seguritygroup"
    description = "Permitir acesso HTTP e acesso a Internet"

    ingress {
        from_port = 80
        to_port = 80
        protocol = "tcp"
        cidr_blocks = ["0.0.0.0/0"]
    }

    egress{
        from_port = 0
        to_port = 65535
        protocol = "tcp"
        cidr_blocks = ["0.0.0.0/0"]
    }
}

resource "aws_instance" "servidor"{
    ami = "ami-05ffe3c48a9991133"
    instance_type = "t2.micro"
    user_data = file("user_data.sh")
    vpc_security_groups_ids = [aws_security_group.securitygroup.id]
}