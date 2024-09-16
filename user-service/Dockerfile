# Dockerfile for user-service

FROM golang:1.21-bullseye AS builder

# Install necessary packages
RUN apt-get update && apt-get install -y \
    git gcc make libc-dev librdkafka-dev

# Enable CGO for Kafka client
ENV CGO_ENABLED=1
ENV GOOS=linux
ENV GOARCH=amd64

WORKDIR /wallet/user-service

# Copy go.mod and go.sum and download dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy the rest of the project files
COPY . .

# Build the user-service binary
RUN go build -o user-service ./cmd/main.go

EXPOSE 8080

CMD ["./user-service"]
