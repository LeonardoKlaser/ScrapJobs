resource "aws_instance" "servidor" {
    ami = "ami-05ffe3c48a9991133"
    instance_type = "t2.micro"
    subnet_id  = aws_subnet.public.id
    iam_instance_profile = aws_iam_instance_profile.ec2_profile.name
    key_name  = var.key_name
    user_data = file("user_data.sh")
    vpc_security_groups_ids = [aws_security_group.ec2_sg.id]

    tags = {
        Name = "${var.project_name}-server"
    }
}

output "ip_publico_servidor" {
  description = "IP Público da instância EC2. Use este IP no seu frontend e nos segredos do GitHub."
  value       = aws_instance.app_server.public_ip
}