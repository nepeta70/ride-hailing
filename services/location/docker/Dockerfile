FROM golang:1.26.4-alpine AS builder

RUN adduser -D appuser
USER appuser

WORKDIR /home/appuser

# Since we will build from root, these paths are now simple
COPY go.mod go.sum ./
RUN go mod download

COPY . .

# Build from root context
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -o /home/appuser/location-service ./services/location/cmd/main.go

FROM alpine:latest

RUN adduser -D appuser
USER appuser

WORKDIR /home/appuser

COPY --from=builder /home/appuser/location-service .

RUN chown appuser:appuser ./location-service && \
    chmod +x ./location-service

COPY services/location/cmd/config.json /home/appuser/config/config.json

EXPOSE 50051

ENTRYPOINT ["./location-service"]