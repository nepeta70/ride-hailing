# Application Core
variable "REQUEST_TIMEOUT_SECONDS" {
  type = number
}

variable "MAX_CLOCK_DRIFT_SECONDS" {
  type = number
}

variable "SERVER_HOST" {
  type = string
}

# Redis
variable "REDIS_ADDRESS" {
  type = string
}

variable "REDIS_PASSWORD" {
  type      = string
  sensitive = true
}

variable "REDIS_PORT" {
  type = number
}

# Postgres
variable "POSTGRES_USER" {
  type = string
}

variable "POSTGRES_PASSWORD" {
  type      = string
  sensitive = true
}

variable "POSTGRES_DB" {
  type = string
}

variable "POSTGRES_HOST" {
  type = string
}

variable "POSTGRES_PORT" {
  type = number
}

# MongoDb
variable "MONGO_USER" {
  type = string
}

variable "MONGO_PASSWORD" {
  type      = string
  sensitive = true
}

variable "MONGO_DB" {
  type = string
}

variable "MONGO_HOST" {
  type = string
}

variable "MONGO_PORT" {
  type = number
}

# Google Maps API
variable "GOOGLE_MAPS_API_KEY" {
  type      = string
  sensitive = true
}

# Telemetry
variable "LOG_LEVEL" {
  type = string
}

variable "PROMETHEUS_ADDRESS" {
  type = string
}

variable "OPENTELEMETRY_ADDRESS" {
  type = string
}

variable "TELEMETRY_INTERVAL_SECONDS" {
  type = number
}

# Kafka
variable "KAFKA_ADVERTISE_EXTERNAL" {
  type = string
}

variable "KAFKA_ADVERTISE_INTERNAL" {
  type = string
}

variable "KAFKA_AUTO_CREATE_TOPICS" {
  type = bool
}

variable "KAFKA_BATCH_SIZE" {
  type = number
}

variable "KAFKA_BROKERS" {
  type = string
}

variable "KAFKA_ENABLE_LOGGING" {
  type = bool
}

variable "KAFKA_EXTERNAL_PORT" {
  type = number
}

variable "KAFKA_INTERNAL_PORT" {
  type = number
}

variable "KAFKA_MEMORY" {
  type = string
}

variable "KAFKA_NODE_ID" {
  type = number
}

variable "KAFKA_RESERVE_MEMORY" {
  type = string
}

variable "KAFKA_SMP" {
  type = number
}

variable "KAFKA_UI_BROKERS" {
  type = string
}

# Location
variable "LOCATION_SERVER_PORT" {
  type = number
}

variable "LOCATION_API_KEY" {
  type      = string
  sensitive = true
}

# Ride
variable "RIDE_SERVER_PORT" {
  type = number
}

variable "RIDE_API_KEY" {
  type      = string
  sensitive = true
}

# Matching
variable "MATCHING_SERVER_PORT" {
  type = number
}

variable "MATCHING_API_KEY" {
  type      = string
  sensitive = true
}

variable "MATCHING_SENDER_ID" {
  type = string
}

variable "MAX_MATCHING_MINUTES" {
  type = number
}

variable "MATCHING_RETRY_INTERVAL_SECONDS" {
  type = number
}

# Driver
variable "DRIVER_SERVER_PORT" {
  type = number
}

variable "DRIVER_API_KEY" {
  type      = string
  sensitive = true
}