FROM golang:1.25-alpine

# Base image for development
WORKDIR /app

# Install Git and Air for live reload during development
RUN apk add --no-cache git \
    && go install github.com/cosmtrek/air@latest

# Copy dependency files first to leverage Docker layer caching
COPY go.mod go.sum ./
RUN go mod download

# Copy the entire codebase
COPY . .

# Start Air with the provided configuration
CMD ["air", "-c", ".air.toml"]
