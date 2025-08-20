FROM golang:1.25-alpine

# Base image for development
WORKDIR /app/services/api-gateway

# Install Git and Air for live reload during development
RUN apk add --no-cache git \
    && go install github.com/air-verse/air@latest

# Copy dependency files first to leverage Docker layer caching
COPY services/api-gateway/go.mod services/api-gateway/go.sum ./
# Copy proto-go module file for local replace to work during dependency download
COPY proto-go/go.mod /app/proto-go/go.mod
COPY libs/kafka/go.mod /app/libs/kafka/go.mod
RUN go mod download

# Copy the entire codebase
WORKDIR /app
COPY . /app

# Start Air with the provided configuration
WORKDIR /app/services/api-gateway
CMD ["air", "-c", ".air.toml"]