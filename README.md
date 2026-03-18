# Ride-Hailing Platform (Go) — Work in Progress

Hexagonal microservices backend in mono-repo. 3 of 6 planned services implemented, 
each sharing the same architectural foundation (domain core isolated via ports) 
with domain-appropriate adapter implementations.

## Architecture Philosophy

**Universal pattern**: Hexagonal architecture across all services
- `internal/core`: Domain logic, entities, state machines, use cases
- `internal/ports`: Interface definitions (driving and driven)
- `internal/adapters`: Concrete implementations (gRPC, Kafka, Redis, PostgreSQL)

**Domain-specific adapters**: Each service chooses implementations fitting its behavior:

| Service | Core Pattern | Ingress Adapters | Egress Adapters | Storage |
|---------|-------------|------------------|-----------------|-----------|
| **Ride** | Actor model (grains) | gRPC server | Kafka publisher, PostgreSQL persistence | PostgreSQL (state), Redis (cache) |
| **Location** | Simple domain logic | gRPC server | — | Redis geospatial only |
| **Matching** | Event handlers | Kafka consumer (no gRPC server) | gRPC client to Location, Kafka publisher | Kafka (streams), possible Redis cache |

All services: dependencies point inward, domain testable without infrastructure, 
observability via OpenTelemetry.

---

## Core Infrastructure (`internal/pkg`)

Shared primitives used across all services:

| Component | Purpose | Usage |
|-----------|---------|-------|
| `actor` | Grain lifecycle, silo management, identity-based addressing | Ride service uses this; available to others |
| `adapters` | Base adapter patterns, gRPC server/client wrappers, storage abstractions | All services |
| `contracts` | Command and event definitions, topic constants, message envelopes | All services |
| `ctxmgr` | Context propagation for request info, trace IDs | All services |
| `errors` | Domain error types: validation, business, transient, permanent | All services |
| `middleware` | gRPC interceptors: auth, circuit breaker, rate limiting, logging, timeout | All services |
| `ports` | Core interface definitions (EventPublisher, GrainStorage, TelemetryProvider) | All services |
| `resiliency` | Retry factories with exponential backoff, retrier configuration | All services |
| `telemetry` | OpenTelemetry tracer provider, structured logging with slog | All services |
| `validator` | Request validation utilities | All services |

---

## Services Detail

### Ride Service (Complete Core Implementation)

**Purpose**: Manage ride lifecycle from request to completion/cancellation

**Domain Model**:

RideCore (immutable):    RequestID, RiderID, PickupLocation, DropoffLocation,
ServiceType, Fare, Currency
RideState (mutable):     DriverID*, Status, Version

**State Machine**:
NEW → REQUESTED → MATCHED → ACCEPTED → STARTED → COMPLETED
↓
CANCELLED (from REQUESTED, MATCHED, ACCEPTED)

**Command Handlers** (pipeline pattern):
| Command | Validation | Transition | Persist | Publish |
|---------|-----------|------------|---------|---------|
| RequestRide | Message + state | NEW→REQUESTED | Yes | RideRequestedEvent |
| CancelRide | Rider owns ride | →CANCELLED | Yes | RideCanceledEvent |
| AcceptRide | Driver assigned | MATCHED→ACCEPTED | Yes | RideAcceptedEvent |
| RejectRide | Driver assigned | (no persist) | No | RideRejectedEvent |
| StartRide | Driver owns ride | ACCEPTED→STARTED | Yes | RideStartedEvent |
| CompleteRide | Driver owns ride | STARTED→COMPLETED | Yes | RideCompletedEvent |
| RideMatchedEvent | External | REQUESTED→MATCHED | Yes | (re-publish) |

**Adapters**:
- **Driving**: gRPC server (`internal/adapters/grpc`)
- **Driven**: 
  - PostgreSQL persistence (state snapshots)
  - Redis (grain cache, possible future)
  - Kafka event publisher
  - Google Maps API client (directions/fare estimation, with fallback)

### Matching Service (Flow Working, Algorithm Stubbed)

**Purpose**: Match ride requests to available drivers

**Architecture Intentional**: No gRPC server surface
- **Ingress**: Kafka consumer adapter only (`internal/adapters/kafka/consumer`)
  - Subscribes to `TopicRide` (RideRequestedEvent)
- **Processing**: 
  1. Parse ride request
  2. Query Location service via gRPC (driving adapter as client)
  3. Run matching algorithm (currently stubbed — returns predetermined driver)
  4. Build RideMatchedEvent
- **Egress**: Kafka publisher adapter
  - Publishes to `TopicRide` (RideMatchedEvent)

**Core Logic**:
- Event handler per event type
- No state machine (stateless processing)
- No persistence (events are log)

**Future adapters**:
- Redis cache for `ride_id → pending_request` (deduplication, timeout handling)
- Possible gRPC server if external API needed (currently not)

**Why this pattern**: Matching is pure reaction to streams. No client calls it directly; it observes and acts. gRPC server would be unnecessary complexity.

---

## Planned Services (Not Started)

| Service | Domain Core | Likely Ingress | Likely Egress | Rationale |
|---------|-------------|--------------|-------------|-----------|
| **Notification** | Event-to-push mapping | Kafka consumer | Push adapters (FCM, APNS, email, SMS) | Stateless reaction to ride events |
| **User** | Profile aggregate, preferences | gRPC server | PostgreSQL, possibly Redis | Simple CRUD, no complex state |
| **Driver** | Driver aggregate, document verification, shift status | gRPC server, Kafka consumer | PostgreSQL, Kafka publisher | Similar complexity to Ride — likely grains |

---

## Infrastructure Status

**Current**: Docker Compose for local development
- PostgreSQL (Ride persistence)
- Redis (Location cache, possible future Matching cache)
- Kafka (event bus)
- Services: Ride, Location, Matching

**Deferred**: Kubernetes + Terraform
- **Rationale**: API contracts still evolving as services are added
- **Trigger**: All 6 services have basic implementations
- **Avoiding**: Updating code, Docker Compose, and K8s manifests simultaneously during rapid iteration

---

## Directory Structure
ride-hailing/
├── api/proto/                    # Protobuf definitions
│   ├── location/v1/
│   ├── matching/v1/             # (definitions only, no gRPC server)
│   └── ride/v1/
├── gen/proto/                    # Generated Go gRPC code
├── gateway/                      # Planned: API Gateway (empty)
├── internal/pkg/                 # Shared hexagonal infrastructure
│   ├── actor/
│   │   ├── grain/               # Grain identity, lifecycle interfaces
│   │   └── silo/                # Grain activation, message routing
│   ├── adapters/
│   │   ├── grpc/                # gRPC server/client wrappers
│   │   ├── kafka/               # Publisher and consumer
│   │   ├── pgstore/             # PostgreSQL persistence
│   │   ├── pubsub/              # Event publisher interface
│   │   ├── rdstore/             # Redis client
│   │   └── telemetry/           # OpenTelemetry provider
│   ├── contracts/               # Commands, events, topics
│   ├── ctxmgr/                  # Context propagation
│   ├── errors/                  # Domain error types
│   ├── middleware/              # gRPC interceptors
│   ├── ports/                   # Core interfaces
│   ├── resiliency/
│   │   └── retry/               # Retrier factories
│   ├── validator/               # Request validation
│   └── ...
├── services/
│   ├── location/
│   │   ├── cmd/                 # main.go
│   │   ├── internal/
│   │   │   ├── core/
│   │   │   │   └── service/     # Domain logic
│   │   │   ├── adapters/        # gRPC server, Redis
│   │   │   └── ports/           # Interface definitions
│   │   └── config/
│   ├── matching/
│   │   ├── cmd/
│   │   ├── internal/
│   │   │   ├── core/
│   │   │   │   └── handlers/    # Event handlers
│   │   │   ├── adapters/        # Kafka consumer, gRPC client
│   │   │   └── ports/
│   │   └── config/
│   └── ride/
│       ├── cmd/
│       ├── internal/
│       │   ├── core/
│       │   │   ├── app/         # Application layer, grain system
│       │   │   │   ├── grains/  # Ride grain implementation
│       │   │   │   └── service/ # Use cases
│       │   │   └── domain/      # Entities, value objects, state machine
│       │   ├── adapters/        # gRPC, PostgreSQL, Kafka, Google Maps
│       │   └── ports/           # Driven and driving interfaces
│       └── config/
├── docker/
│   └── docker-compose.yml       # Local infrastructure
└── infra/                        # Planned: K8s, Terraform (empty)

---

## Development Status
`
| Component | Status | Notes |
|-----------|--------|-------|
| Core infrastructure | Stable | Used by all 3 services |
| Ride service | Feature-complete core | State machine working, events flowing |
| Location service | Functional | Geospatial queries working |
| Matching service | Integration working | Algorithm stubbed for replacement |
| Notification service | Not started | — |
| User service | Not started | — |
| Driver service | Not started | — |
| API Gateway | Not started | — |
| Production infra | Deferred | Waiting for service API stabilization |
`
---

## Getting Started

```bash
# Clone
git clone https://github.com/nepeta70/ride-hailing.git
cd ride-hailing

# Start infrastructure + services
docker-compose up --build

# Services available:
# - Ride: gRPC on :50051
# - Location: gRPC on :50052  
# - Matching: Kafka consumer only (no exposed port)

# Generate protobuf (if changing api/proto/)
buf generate

# Run tests
go test ./...