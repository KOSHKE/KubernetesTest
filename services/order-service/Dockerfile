# Use official Go image based on Debian for better C library support (librdkafka, protobuf, etc.)
FROM golang:1.25-bookworm

# Set working directory for the service
WORKDIR /app/services/order-service

# Install Git and build-essential for compiling C dependencies
RUN apt-get update && \
    apt-get install -y git build-essential && \
    rm -rf /var/lib/apt/lists/*

# Install Air for hot reload during development
RUN go install github.com/air-verse/air@latest

# Copy go.mod and go.sum first to leverage Docker layer caching
COPY services/order-service/go.mod services/order-service/go.sum ./
COPY proto-go/go.mod /app/proto-go/go.mod
COPY pkg/kafka/go.mod /app/pkg/kafka/go.mod

# Download dependencies
RUN go mod download

# Copy the full codebase
COPY . /app

# Set working directory for the service
WORKDIR /app/services/order-service

# Start Air for live reload
CMD ["air", "-c", ".air.toml"]
