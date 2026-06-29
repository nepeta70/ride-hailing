FROM golang:1.26.4-alpine AS builder

RUN adduser -D appuser
USER appuser

WORKDIR /home/appuser

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -o /home/appuser/matching-service ./services/matching/cmd/main.go

FROM alpine:latest

RUN adduser -D appuser
USER appuser

WORKDIR /home/appuser

COPY --from=builder /home/appuser/matching-service .

RUN chown appuser:appuser ./matching-service && \
    chmod +x ./matching-service

COPY services/matching/cmd/config.json /home/appuser/config/config.json

EXPOSE 50053

ENTRYPOINT ["./matching-service"]