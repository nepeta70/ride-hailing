FROM golang:1.26.1-alpine AS builder

RUN adduser -D appuser
USER appuser

WORKDIR /home/appuser

# Since we will build from root, these paths are now simple
COPY go.mod go.sum ./
RUN go mod download

COPY . .

# Build from root context
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -o /home/appuser/ride-service ./services/ride/cmd/main.go

FROM alpine:latest

RUN adduser -D appuser
USER appuser

WORKDIR /home/appuser

COPY --from=builder /home/appuser/ride-service .

RUN chown appuser:appuser ./ride-service && \
    chmod +x ./ride-service

COPY services/ride/cmd/config.json /home/appuser/config/config.json

EXPOSE 50052
ENTRYPOINT ["./ride-service"]