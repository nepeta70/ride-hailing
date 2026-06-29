FROM golang:1.26.4-alpine AS builder

RUN adduser -D appuser
USER appuser

WORKDIR /home/appuser

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -o /home/appuser/gateway-service ./gateway/cmd/main.go

FROM alpine:latest

RUN adduser -D appuser
USER appuser

WORKDIR /home/appuser

COPY --from=builder /home/appuser/gateway-service .

RUN chown appuser:appuser ./gateway-service && \
    chmod +x ./gateway-service

COPY gateway/cmd/config.json /home/appuser/config.json

EXPOSE 5000

ENTRYPOINT ["./gateway-service"]
