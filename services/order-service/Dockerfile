FROM golang:1.25-alpine

# Base image for development
WORKDIR /app/services/order-service

# Install Git and Air for live reload during development
RUN apk add --no-cache git \
    && go install github.com/air-verse/air@latest

# Copy dependency files first to leverage Docker layer caching
COPY services/order-service/go.mod services/order-service/go.sum ./
COPY proto-go/go.mod /app/proto-go/go.mod
RUN go mod download

# Copy the entire codebase
WORKDIR /app
COPY . /app

# Start Air with the provided configuration
WORKDIR /app/services/order-service
CMD ["air", "-c", ".air.toml"]