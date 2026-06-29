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
    go build -o /home/appuser/driver-service ./services/driver/cmd/main.go

FROM alpine:latest

RUN adduser -D appuser
USER appuser

WORKDIR /home/appuser

COPY --from=builder /home/appuser/driver-service .

RUN chown appuser:appuser ./driver-service && \
    chmod +x ./driver-service

COPY services/driver/cmd/config.json /home/appuser/config/config.json

EXPOSE 50054
ENTRYPOINT ["./driver-service"]