FROM golang:1.25-alpine

# Base image for development
WORKDIR /app/services/payment-service

# Install Git and Air for live reload during development
RUN apk add --no-cache git \
    && go install github.com/air-verse/air@latest

# Copy dependency files first to leverage Docker layer caching
COPY services/payment-service/go.mod services/payment-service/go.sum ./
COPY proto-go/go.mod /app/proto-go/go.mod
RUN go mod download

# Copy the entire codebase
WORKDIR /app
COPY . /app

# Start Air with the provided configuration
WORKDIR /app/services/payment-service
CMD ["air", "-c", ".air.toml"]
