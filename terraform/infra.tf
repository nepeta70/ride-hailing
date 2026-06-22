# PostgreSQL - Using 17 (Latest actual RDS support)
resource "aws_db_instance" "postgres_db" {
  identifier           = "postgres-db"
  engine               = "postgres"
  engine_version       = "17.1" 
  instance_class       = "db.t3.micro" # Free Tier
  allocated_storage    = 20
  db_name              = var.POSTGRES_DB
  username             = var.POSTGRES_USER
  password             = var.POSTGRES_PASSWORD
  skip_final_snapshot  = true
  publicly_accessible  = false
  vpc_security_group_ids = [aws_security_group.db_sg.id]
}

# Redis - Standard Cluster (cache.t3.micro is Free Tier)
resource "aws_elasticache_cluster" "redis" {
  cluster_id           = "ride-redis"
  engine               = "redis"
  node_type            = "cache.t3.micro" # Free Tier
  num_cache_nodes      = 1
  parameter_group_name = "default.redis7"
  port                 = 6379
  subnet_group_name    = aws_elasticache_subnet_group.redis_subnets.name
  security_group_ids   = [aws_security_group.redis_sg.id]
}

resource "aws_elasticache_subnet_group" "redis_subnets" {
  name       = "redis-subnets"
  subnet_ids = [aws_subnet.private_1.id, aws_subnet.private_2.id]
}

# MongoDB (DocumentDB)
resource "aws_docdb_cluster" "mongodb" {
  cluster_identifier      = "mongodb-cluster"
  engine                  = "docdb"
  master_username         = var.MONGO_USER
  master_password         = var.MONGO_PASSWORD
  skip_final_snapshot     = true
  vpc_security_group_ids  = [aws_security_group.db_sg.id]
}

resource "aws_docdb_cluster_instance" "mongodb_instances" {
  count              = 1
  identifier         = "mongodb-instance"
  cluster_identifier = aws_docdb_cluster.mongodb.id
  instance_class     = "db.t3.medium" 
}

# SQS (Queue for Event Messaging)
resource "aws_sqs_queue" "kafka_replacement" {
  name = "ride-hailing-queue"
}