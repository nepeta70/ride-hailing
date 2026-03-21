FROM golang:1.26-alpine AS builder

RUN adduser -D appuser
USER appuser

WORKDIR /home/appuser

# Since we will build from root, these paths are now simple
COPY go.mod go.sum ./
RUN go mod download

COPY . .

# Build from root context
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -o /home/appuser/user-service ./services/user/cmd/main.go

FROM alpine:latest

RUN adduser -D appuser
USER appuser

WORKDIR /home/appuser

COPY --from=builder /home/appuser/user-service .

RUN chown appuser:appuser ./user-service && \
    chmod +x ./user-service

COPY services/user/cmd/config.json /home/appuser/config/config.json

EXPOSE 50055
ENTRYPOINT ["./user-service"]