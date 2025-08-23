# Use official Go image based on Debian for better C library support
FROM golang:1.25-bookworm

# Set working directory for the service
WORKDIR /app/services/payment-service

# Install Git and build-essential for compiling C dependencies
RUN apt-get update && \
    apt-get install -y git build-essential && \
    rm -rf /var/lib/apt/lists/*

# Install Air for hot reload during development
RUN go install github.com/air-verse/air@latest

# Copy go.mod and go.sum first to leverage Docker layer caching
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy the full codebase
COPY . /app

# Set working directory for the service
WORKDIR /app/services/payment-service

# Start Air for live reload
CMD ["air", "-c", ".air.toml"]
