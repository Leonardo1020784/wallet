# Dockerfile for transaction-service

FROM golang:1.21-bullseye AS builder

# Install necessary packages
RUN apt-get update && apt-get install -y \
    git gcc make libc-dev librdkafka-dev

# Enable CGO for Kafka client
ENV CGO_ENABLED=1
ENV GOOS=linux
ENV GOARCH=amd64

WORKDIR /wallet/transaction-service

COPY go.mod go.sum ./
RUN go mod download

COPY . .

# Build the transaction-service binary
RUN go build -o transaction-service ./cmd/main.go

EXPOSE 8081

CMD ["./transaction-service"]
