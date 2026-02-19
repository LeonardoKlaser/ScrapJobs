#!/bin/dash

sudo yum update -y
sudo yum install -y docker jq
sudo curl -L "https://github.com/docker/compose/releases/latest/download/docker-compose-$(uname -s)-$(uname -m)" -o /usr/local/bin/docker-compose
sudo chmod +x /usr/local/bin/docker-compose
sudo service docker start
usermod -a -G docker ec2-user
