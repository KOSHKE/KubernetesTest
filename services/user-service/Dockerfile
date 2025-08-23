FROM golang:1.25-alpine

# Base image for development
WORKDIR /app/services/user-service

# Install Git and Air for live reload during development
RUN apk add --no-cache git \
    && go install github.com/air-verse/air@latest

# Copy dependency files first to leverage Docker layer caching
COPY services/user-service/go.mod services/user-service/go.sum ./
COPY proto-go/ /app/proto-go/
COPY pkg/ /app/pkg/
RUN go mod download

# Copy the entire codebase
WORKDIR /app
COPY . /app

# Start Air with the provided configuration
WORKDIR /app/services/user-service
CMD ["air", "-c", ".air.toml"]