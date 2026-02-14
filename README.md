# Ride-Hailing Microservices Platform

## Overview

This project is a modular, microservices-based ride-hailing platform written in Go. It is designed for scalability, reliability, and extensibility, using gRPC, circuit breaker patterns, and robust configuration management. The architecture separates concerns into distinct services (driver, location, matching, notification, ride, user), each with its own domain logic and adapters.

## Features

- **Microservices Architecture:** Each domain (driver, location, matching, notification, ride, user) is implemented as an independent service.
- **gRPC APIs:** Protobuf definitions for all domains, with generated code for efficient communication.
- **Resiliency:** Circuit breaker, retry, and rate limiting middleware for robust service-to-service communication.
- **Configurable:** Centralized configuration and security management.
- **Domain-Driven Design:** Clear separation of domain logic, adapters, ports, and value objects.
- **Testing:** Extensive unit tests for core logic and middleware.
- **Docker Support:** Dockerfiles and docker-compose for easy local development and deployment.

## Directory Structure

- `api/proto/`: Protobuf definitions for all domains.
- `gateway/`: Entry point for API gateway, routing requests to services.
- `gen/proto/`: Generated gRPC code.
- `internal/pkg/`: Shared libraries (actor model, adapters, config, contracts, domain, errors, middleware, ports, resiliency, validator).
- `services/`: Each microservice (driver, location, matching, notification, ride, user) with its own command, config, adapters, and core logic.
- `docker/`: Docker Compose configuration for local orchestration.

## Key Components

- **Actor Model:** Used for scalable state management and concurrency.
- **Resiliency:** Circuit breaker and retry strategies to prevent cascading failures.
- **Adapters:** Abstraction for databases, pub/sub, and external systems.
- **Middleware:** Authentication, circuit breaker, rate limiting, logging, and timeout handling.
- **Contracts:** Command and event definitions for inter-service communication.

## Service Details

### Ride Service

The Ride service manages the lifecycle of ride requests, matching drivers and users, tracking ride status, and handling ride events.

**Features:**
- Handles ride creation, assignment, start, completion, and cancellation.
- Integrates with the Matching service to find available drivers.
- Tracks ride status and updates via gRPC and event-driven contracts.
- Stores ride data using pluggable adapters (e.g., PostgreSQL, Redis).
- Implements domain logic for fare calculation, state transitions, and validation.
- Exposes gRPC endpoints for ride operations.
- Includes Dockerfile for containerized deployment.

**Directory Structure:**
- `services/ride/cmd/`: Service entrypoint and configuration.
- `services/ride/internal/core/`: Business logic and domain models.
- `services/ride/internal/adapters/`: Database and external system adapters.
- `services/ride/internal/ports/`: Interfaces for inbound/outbound communication.

**Operations:**
- Rider requests a fare estimate
- Rider requests a ride (he must choose a service type)
- Rider cancels a ride
- Driver accepts or rejects a ride
- Driver starts ride
- Driver completes ride

### Location Service

The Location service tracks and updates the real-time locations of drivers and users.

**Features:**
- Receives and stores location updates from drivers and users.
- Provides APIs for querying nearby drivers or users.
- Supports geospatial queries for matching and routing.
- Integrates with ride and matching services for location-based operations.
- Uses adapters for persistent storage and pub/sub messaging.
- Exposes gRPC endpoints for location updates and queries.
- Includes Dockerfile for containerized deployment.

**Directory Structure:**
- `services/location/cmd/`: Service entrypoint and configuration.
- `services/location/internal/core/`: Location update logic and geospatial models.
- `services/location/internal/adapters/`: Storage and messaging adapters.
- `services/location/internal/ports/`: Interfaces for communication.

**Example Operations:**
- Update driver/user location (mobile app → location service)
- Query nearby drivers (ride/matching service → location service)
- Broadcast location changes (pub/sub for real-time updates)

## Getting Started

1. **Clone the repository:**
   ```
   git clone https://github.com/nepeta70/ride-hailing.git
   cd ride-hailing
   ```

2. **Build and run with Docker Compose:**
   ```
   docker-compose up --build
   ```

3. **Generate gRPC code:**
   ```
   buf generate
   ```

4. **Run tests:**
   ```
   go test ./...
   ```

## Development

- Each service can be developed and tested independently.
- Shared logic lives in `internal/pkg/`.
- Protobuf changes require regeneration (`buf generate`).

## Contributing

1. Fork the repo and create your branch.
2. Make changes and add tests.
3. Submit a pull request.

## License

MIT License
