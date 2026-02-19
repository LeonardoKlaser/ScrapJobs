resource "aws_elasticache_subnet_group" "default" {
  name       = "${var.project_name}-elasticache-subnet-group"
  subnet_ids = [aws_subnet.private.id, aws_subnet.private_b.id]
}

# Cluster ElastiCache (Redis)
resource "aws_elasticache_cluster" "default" {
  cluster_id           = "scrapjobs-redis-cluster"
  engine               = "redis"
  node_type            = "cache.t2.micro"
  num_cache_nodes      = 1
  parameter_group_name = "default.redis7"
  engine_version       = "7.0"
  port                 = 6379
  subnet_group_name    = aws_elasticache_subnet_group.default.name
  security_group_ids   = [aws_security_group.elasticache_sg.id]

  transit_encryption_enabled  = true
  at_rest_encryption_enabled  = true
}

output "elasticache_endpoint" {
  description = "primary redis endpoint"
  value       = aws_elasticache_cluster.default.cache_nodes[0].address
  sensitive   = true
}

output "elasticache_port" {
  description = "cluster port"
  value       = aws_elasticache_cluster.default.port
}
